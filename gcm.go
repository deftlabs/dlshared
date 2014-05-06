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
	"sync"
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
//				"maxWaitOnStopInMs": 30000
//          }
//		}
//
// In the example above, the configuration path would be "gcm" (passed in New function). This assumes
// that "gcm" is located as child of the root of the document.
//
type GoogleCloudMessagingSvc struct {
	Logger
	consumer *Consumer
	configPath string

	consumerChannel chan interface{}
	requestChannel chan interface{}
	responseChannel chan interface{}

	authKey string
	postUrl string

	updateStatsTicker *time.Ticker

	httpClient HttpRequestClient

	stats *GoogleCloudMsgSendStats
	sync.WaitGroup

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
	MulticastId float64 `json:"multicast_id"`
	Success float64 `json:"success"`
	Failure float64 `json:"failure"`
	CanonicalIds float64 `json:"canonical_ids"`
	Results []*GoogleCloudMsgResponseResult `json:"results"`
	OrigMsg *GoogleCloudMsg `json:"-"`
	Err error `json:"-"` // If there is an error (usually io) returned by send POST operation
	HttpStatusCode int `json:"-"`
}

type GoogleCloudMsgResponseResult struct {
	MessageId string `json:"message_id"`
	RegistrationId string `json:"registration_id"`
	Error string `json:"error"`
}

func NewGoogleCloudMessagingSvc(configPath string, httpClient HttpRequestClient, requestChannel, responseChannel chan interface{}) *GoogleCloudMessagingSvc {
	return &GoogleCloudMessagingSvc{	configPath: configPath,
										requestChannel: requestChannel,
										responseChannel: responseChannel,
										consumerChannel: make(chan interface{}),
										updateStatsTicker: time.NewTicker(1 * time.Second),
										httpClient: httpClient,
										responseHandlers: make([]GoogleCloudMessagingMsgResponseHandler, 0),
										stats: &GoogleCloudMsgSendStats{},
	}
}

func (self *GoogleCloudMessagingSvc) sendResponse(response *GoogleCloudMsgResponse, msgCount int, err error) {
	self.Add(1)
	defer self.Done()

	if err != nil {
		self.stats.updateFailureCount(msgCount)
		response.Err = err
	} else {
		self.stats.updateSuccessCount(msgCount)
	}

	self.responseChannel <- response
}


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

		self.sendResponse(&response, msgCount, NewStackError("Problem with gcm post - err: %v", err))

	} else if response.HttpStatusCode >= 500 {

		self.sendResponse(&response, msgCount, NewStackError("Problem with gcm servers - http status error code: %d", response.HttpStatusCode))

	} else if response.HttpStatusCode == 400 {

		self.sendResponse(&response, msgCount, NewStackError("Problem with gcm post - http status: 400 - problem with JSON sent"))

	} else if response.HttpStatusCode == 401 {

		self.sendResponse(&response, msgCount, NewStackError("Problem with gcm post - http status: 401 - unable to authenticate - check your auth key"))

	} else if response.HttpStatusCode != 200 {

		self.sendResponse(&response, msgCount, NewStackError("Problem with gcm post - http status: %d", response.HttpStatusCode))

	} else {

		if err = json.Unmarshal(result, &response); err != nil {

			self.sendResponse(&response, msgCount, NewStackError("Unable to parse gcm response - err: %v", err))

		} else {

			self.sendResponse(&response, msgCount, nil)

		}
	}
}

func (self *GoogleCloudMessagingSvc) listenForRequests() {
	self.Add(1)
	defer self.Done()
	for msg := range self.requestChannel {
		backoff := self.stats.backoffTime()
		if backoff > 0 { time.Sleep(time.Duration(backoff) * time.Millisecond) }
		self.consumerChannel <- msg
	}
}

func (self *GoogleCloudMessagingSvc) updateStats() { for now := range self.updateStatsTicker.C { self.stats.update(&now) } }

func (self *GoogleCloudMessagingSvc) Start(kernel *Kernel) error {

	self.authKey = strings.TrimSpace(kernel.Configuration.StringWithPath(self.configPath, "authKey", ""))
	self.postUrl = strings.TrimSpace(kernel.Configuration.StringWithPath(self.configPath, "postUrl", "https://android.googleapis.com/gcm/send"))

	if len(self.authKey) == 0 { return NewStackError("Unable to create GoogleCloudMessagingSvc - no \"authKey\" field in config file - path: %s", self.configPath) }

	self.stats.acceptableGcmFailurePercent = kernel.Configuration.IntWithPath(self.configPath, "acceptableGoogleCloudMsgFailurePercent", 10)
	self.stats.initialGcmBackoffInMs = kernel.Configuration.IntWithPath(self.configPath, "initialGoogleCloudMsgBackoffInMs", 100)
	self.stats.maxGcmBackoffInMs = kernel.Configuration.IntWithPath(self.configPath, "maxGoogleCloudMsgBackoffInMs", 10000)

	self.consumer = NewConsumer("GoogleCloudMessagingSvcConsumer",
								self.consumerChannel,
								self.processMsg,
								kernel.Configuration.IntWithPath(self.configPath, "consumer.maxGoroutines", 1000),
								kernel.Configuration.IntWithPath(self.configPath, "consumer.maxWaitOnStopInMs", 30000),
								kernel.Logger)


	if err := self.consumer.Start(); err != nil { return err }

	go self.updateStats()

	go self.listenForRequests()

	return nil
}

func (self *GoogleCloudMessagingSvc) Stop(kernel *Kernel) error {

	close(self.consumerChannel)
	if self.consumer != nil { if err := self.consumer.Stop(); err != nil { return err } }

	self.Wait()

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
	sync.Mutex
}

func (self *GoogleCloudMsgSendStats) update(now *time.Time) {
	self.Lock()
	defer self.Unlock()

	if self.currentFailurePercent() <= self.acceptableGcmFailurePercent { self.reduceBackoff() } else { self.increaseBackoff() }

	// Update the current and previous values.
	self.updateCurrent()
}

func (self *GoogleCloudMsgSendStats) updateCurrent() {
	self.PreviousSuccessCount = self.CurrentSuccessCount
	self.PreviousFailureCount = self.CurrentFailureCount
	self.CurrentSuccessCount = 0
	self.CurrentFailureCount = 0
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

func (self *GoogleCloudMsgSendStats) backoffTime() int {
	self.Lock()
	defer self.Unlock()
	return self.BackoffTimeInMs
}

func (self *GoogleCloudMsgSendStats) updateFailureCount(count int) {
	self.Lock()
	defer self.Unlock()
	self.CurrentFailureCount += count
}

func (self *GoogleCloudMsgSendStats) updateSuccessCount(count int) {
	self.Lock()
	defer self.Unlock()
	self.CurrentSuccessCount += count
}

