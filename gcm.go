/**
 * (C) Copyright 2014, Deft Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dlshared

import(
	"fmt"
	"time"
	"strings"
	"encoding/json"
)

// The google cloud messaging service. This component dispatches JSON messages to
// Google via HTTP and implements exponential backoff to ensure that if there is a
// failure, everything slows down as to not overload the consuming service. This
// component wraps a consumer component. The component is configured by passing in
// the configuration path.
// 		"gcm": {
//			"postUrl": "https://android.googleapis.com/gcm/send",
//			"authKey": "WHATEVER_IT_IS_FOR_YOUR_API_ACCESS",
//			"acceptableGoogleCloudMsgFailurePercent": 10,
//			"initialGoogleCloudMsgBackoffInMs": 100,
//			"maxGoogleCloudMsgBackoffInMs": 10000,
//			"consumer": {
//				"maxGoroutines": 1000,
//				"channelBufferSize": 1000,
//				"maxWaitOnStopInMs": 30000
//          }
//		}
//
// In the example above, the configuration path would be "gcm" (passed in New function). This assumes
// that "gcm" is located as child of the root of the document. Messages are processed asynchronously,
// so there is no response sent to the caller. If the consumer message queue is full,
// passing new messages will result in dropped messages with an error message in the logs.
//
// If you wish to wish to take action based on response from Google you may register a response handler.
// You must register the response handler before you start sending messages. You usually want to handle
// the responses to deal with canonical ids and unregistered devices. Response handlers are called in the order
// they were added.
//
type GoogleCloudMessagingSvc struct {
	Logger
	consumer *Consumer
	configPath string
	requestChannel chan interface{}
	msgChannel chan *GoogleCloudMsg
	authKey string
	postUrl string
	shutdownChannel chan bool
	successChannel chan int
	failureChannel chan int
	msgRequestChannel chan bool
	msgResponseChannel chan int
	backoffCheckTicker *time.Ticker

	acceptableGcmFailurePercent int
	initialGcmBackoffInMs int
	maxGcmBackoffInMs int

	httpClient HttpRequestClient

	responseHandlers []GoogleCloudMessagingMsgResponseHandler
}

type GoogleCloudMessagingMsgResponseHandler func(*GoogleCloudMsgResponse)

const(
	gcmAuthHeader = "Authorization"
	gcmAuthHeaderKeyPrefix = "key="
)

// The user must set the required fields in the message.
type GoogleCloudMsg struct {
	RegistrationIds []string `json:"registration_ids"`
	NotificationKey string `json:"notification_key,omitempty"`
	CollapseKey string `json:"collapse_key,omitempty"`
	DelayWhileIdle bool `json:"delay_while_idle,omitempty"`
	TimeToLive int `json:"time_to_live,omitempty"`
	RestrictedPackageName string `json:"restricted_package_name,omitempty"`
	DryRun bool `json:"dry_run,omitempty"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// The response object returned by google.
type GoogleCloudMsgResponse struct {
	MulticastId []string `json:"multicast_id"`
	Success float64 `json:"success"`
	Failure float64 `json:"failure"`
	CanonicalIds float64 `json:"canonical_ids"`
	Results []*GoogleCloudMsgResponseResult `json:"results"`
	OrigMsg *GoogleCloudMsg `json:"-"`
	LocalErr error `json:"-"` // If there is an error (usually io) returned by send POST operation
	HttpStatusCode int `json:"-"`
	ConsumerOverwhelmed bool `json:"-"`
}

type GoogleCloudMsgResponseResult struct {
	MessageId string `json:"message_id"`
	RegistrationId string `json:"registration_id"`
	Error string `json:"error"`
}

func NewGoogleCloudMessagingSvc(configPath string, httpClient HttpRequestClient) *GoogleCloudMessagingSvc {
	return &GoogleCloudMessagingSvc{	configPath: configPath,
										requestChannel: make(chan interface{}),
										shutdownChannel: make(chan bool),
										msgChannel: make(chan *GoogleCloudMsg),
										successChannel: make(chan int),
										failureChannel: make(chan int),
										msgRequestChannel: make(chan bool),
										msgResponseChannel: make(chan int),
										backoffCheckTicker: time.NewTicker(1 * time.Second),
										httpClient: httpClient,
										responseHandlers: make([]GoogleCloudMessagingMsgResponseHandler, 0),

	}
}

// Called by the user of the service to start the send process. There is no guarantee this
// message will be sent. The best way to implement this is to set a time of send and then
// when the device updates, set the last update time. This call can block if there are problems
// sending data to google. Google requires exponential backoff in calls, so if there are errors
// sending to google, this method will temporarily blocks.
func (self *GoogleCloudMessagingSvc) Send(msg *GoogleCloudMsg) {
	self.msgRequestChannel <- true
	if backoff := <- self.msgResponseChannel; backoff > 0 { time.Sleep(time.Duration(backoff) * time.Millisecond) }
	self.msgChannel <- msg
}

func (self *GoogleCloudMessagingSvc) callResponseHandlers(response *GoogleCloudMsgResponse) { for _, handler := range self.responseHandlers { handler(response) } }

// This is the internal method called by the consumer to actually send the message.
func (self *GoogleCloudMessagingSvc) processMsg(msg interface{}) {

	gcmMsg := msg.(*GoogleCloudMsg)

	response := GoogleCloudMsgResponse{ OrigMsg: gcmMsg }

	msgCount := len(gcmMsg.RegistrationIds)

	var result []byte
	var err error

	response.HttpStatusCode,
	result,
	err = self.httpClient.PostJson(self.postUrl, gcmMsg, map[string]string { gcmAuthHeader:  fmt.Sprintf(gcmAuthHeaderKeyPrefix, self.authKey) })

	if err != nil {
		self.Logf(Warn, "Problem with gcm post - err: %v", err)
		self.failureChannel <- msgCount
		response.LocalErr = err
		self.callResponseHandlers(&response)
		return
	}

	if response.HttpStatusCode >= 500 {
		self.Logf(Warn, "Problem with gcm post - http status error code: %d", response.HttpStatusCode)
		self.failureChannel <- msgCount
		self.callResponseHandlers(&response)
		return
	}

	if response.HttpStatusCode == 400 {
		self.Logf(Error, "Problem with gcm post - http status: 400 - problem with JSON sent")
		self.failureChannel <- msgCount
		self.callResponseHandlers(&response)
		return
	}

	if response.HttpStatusCode == 401 {
		self.Logf(Error, "Problem with gcm post - http status: 401 - unable to authenticate - check your auth key")
		self.failureChannel <- msgCount
		self.callResponseHandlers(&response)
		return
	}

	if response.HttpStatusCode != 200 {
		self.Logf(Warn, "Problem with gcm post - http status: %d", response.HttpStatusCode)
		self.failureChannel <- msgCount
		self.callResponseHandlers(&response)
		return
	}

	err = json.Unmarshal(result, &response)

	if err != nil {
		self.Logf(Warn, "Unable to handle gcm response - err: %v", err)
		self.failureChannel <- msgCount
		response.LocalErr = err
		self.callResponseHandlers(&response)
		return
	}

	self.successChannel <- msgCount
	self.callResponseHandlers(&response)
}

func (self *GoogleCloudMessagingSvc) AddResponseHandler(handler GoogleCloudMessagingMsgResponseHandler) {
	self.responseHandlers = append(self.responseHandlers, func(response *GoogleCloudMsgResponse) {
		defer func() {
			if r := recover(); r != nil {
				self.Logf(Error, "Panic in a google cloud message response handler - name: %s - err: %v", GetFunctionName(handler), r)
			}
		}()

		handler(response)
	})
}

func (self *GoogleCloudMessagingSvc) listenForEvents() {
	stats := &GoogleCloudMsgSendStats{
		acceptableGcmFailurePercent: self.acceptableGcmFailurePercent,
		initialGcmBackoffInMs: self.initialGcmBackoffInMs,
		maxGcmBackoffInMs: self.maxGcmBackoffInMs,
	}

	var successCount int
	var failureCount int

	for {
		select {
			case msg := <- self.msgChannel: self.requestChannel <- msg
			case <- self.msgRequestChannel: self.msgResponseChannel <- stats.BackoffTimeInMs
			case successCount = <- self.successChannel: stats.CurrentSuccessCount += successCount
			case failureCount = <- self.failureChannel: stats.CurrentFailureCount += failureCount
			case <- self.backoffCheckTicker.C: stats.update()
			case <- self.shutdownChannel: return
		}
	}
}

func (self *GoogleCloudMessagingSvc) spilloverHandler(msg interface{}) {
	self.callResponseHandlers(&GoogleCloudMsgResponse{ OrigMsg: msg.(*GoogleCloudMsg), ConsumerOverwhelmed: true })
}

func (self *GoogleCloudMessagingSvc) Start(kernel *Kernel) error {

	self.authKey = strings.TrimSpace(kernel.Configuration.StringWithPath(self.configPath, "authKey", ""))
	self.postUrl = strings.TrimSpace(kernel.Configuration.StringWithPath(self.configPath, "postUrl", "https://android.googleapis.com/gcm/send"))

	if len(self.authKey) == 0 { return NewStackError("Unable to create GoogleCloudMessagingSvc - no \"authKey\" field in config file - path: %s", self.configPath) }

	self.acceptableGcmFailurePercent = kernel.Configuration.IntWithPath(self.configPath, "acceptableGoogleCloudMsgFailurePercent", 10)
	self.initialGcmBackoffInMs = kernel.Configuration.IntWithPath(self.configPath, "initialGoogleCloudMsgBackoffInMs", 100)
	self.maxGcmBackoffInMs = kernel.Configuration.IntWithPath(self.configPath, "maxGoogleCloudMsgBackoffInMs", 10000)

	self.consumer = NewConsumer("GoogleCloudMessagingSvcConsumer",
								self.requestChannel,
								self.processMsg,
								self.spilloverHandler,
								kernel.Configuration.IntWithPath(self.configPath, "consumer.maxGoroutines", 1000),
								kernel.Configuration.IntWithPath(self.configPath, "consumer.channelBufferSize", 1000),
								kernel.Configuration.IntWithPath(self.configPath, "consumer.maxWaitOnStopInMs", 30000),
								kernel.Logger)

	if err := self.consumer.Start(); err != nil { return err }
	go self.listenForEvents()

	return nil
}

func (self *GoogleCloudMessagingSvc) Stop(kernel *Kernel) error {
	self.shutdownChannel <- true
	if self.consumer != nil { if err := self.consumer.Stop(); err != nil { return err } }
	return nil
}

// The cloud message stats.
type GoogleCloudMsgSendStats struct {
	BackoffTimeInMs int
	CurrentSuccessCount int
	CurrentFailureCount int
	PreviousSuccessCount int
	PreviousFailureCount int

	acceptableGcmFailurePercent int
	initialGcmBackoffInMs int
	maxGcmBackoffInMs int
}

func (self *GoogleCloudMsgSendStats) update() {

	if self.currentFailurePercent() <= self.acceptableGcmFailurePercent { self.reduceBackoff()
	} else { self.increaseBackoff() }

	// Update the current and previous values.
	self.updateCurrent()
}

func (self *GoogleCloudMsgSendStats) updateCurrent() {
	self.PreviousSuccessCount = self.CurrentSuccessCount
	self.PreviousFailureCount = self.CurrentFailureCount
	self.CurrentSuccessCount = 0
	self.CurrentFailureCount = 0
}

func (self *GoogleCloudMsgSendStats) clearBackoff() {
	self.BackoffTimeInMs = 0
}

func (self *GoogleCloudMsgSendStats) reduceBackoff() {
	if self.BackoffTimeInMs == 0 { return }
	self.BackoffTimeInMs /= 2
	if self.BackoffTimeInMs <= self.initialGcmBackoffInMs { self.BackoffTimeInMs = 0 }
}

func (self *GoogleCloudMsgSendStats) increaseBackoff() {
	if self.BackoffTimeInMs == 0 { self.BackoffTimeInMs = self.initialGcmBackoffInMs
	} else { self.BackoffTimeInMs *= 2 }
	if self.BackoffTimeInMs > self.maxGcmBackoffInMs { self.BackoffTimeInMs = self.maxGcmBackoffInMs  }
}

func (self *GoogleCloudMsgSendStats) currentFailurePercent() int {
	totalMsgs := self.CurrentSuccessCount + self.CurrentFailureCount
	if totalMsgs == 0 { return 0 } else { return (self.CurrentFailureCount / totalMsgs) * 100 }
}

