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
	"time"
	"bytes"
	"strings"
	"math/rand"
	"encoding/json"
	"encoding/hex"
	"encoding/binary"
	lrucache "github.com/hashicorp/golang-lru"
)

// Constants used for by Apple for their messages/responses.
const (
	apnDeviceTokenItemId = 1
	apnPayloadItemId = 2
	apnNotificationIdentifierItemId = 3
	apnExpirationDateItemId = 4
	apnPriorityItemId = 5
	apnDeviceTokenLength = 32
	apnNotificationIdentifierLength = 4
	apnExpirationDateLength = 4
	apnPriorityLength = 1

	apnPushCommandValue = 2
	apnMaxPayloadSize = 256

	apnErrorResponseCommand = 8
	apnErrorResponseMsgSize = 6

	apnFeedbackMsgSize = 38

	apnResponseErrorNone = 0
	apnResponseErrorProcessing = 1
	apnResponseErrorNoDeviceToken = 2
	apnResponseErrorMissingTopic = 3
	apnResponseErrorMissingPayload = 4
	apnResponseErrorInvalidTokenSize = 5
	apnResponseErrorInvalidTopicSize = 6
	apnResponseErrorInvalidPayloadSize = 7
	apnResponseErrorInvalidToken = 8
	apnResponseErrorShutdown = 10
	apnResponseErrorUnknown = 255
)

// The apple push notification (apn) service. This component dispatches JSON messages to
// Apple via a custom Apple binary protocol. The component is configured by passing in the
// configuration path.
//
// 		"apn": {
//			"gateway": "gateway.sandbox.push.apple.com:2195",
// 			"feedback": "feedback.sandbox.push.apple.com:2196",
//			"certificateFile": "WHATEVER_YOUR_PEM_FILE",
//			"keyFile": "WHATEVER_YOUR_KEY_PEM_FILE",
//			"socketTimeoutInMs": "4000",
//			"msgCacheElementCount": "2000"
//		}
//
// In the example above, the configuration path would be "apn" (passed in New function). This assumes
// that "apn" is located as child of the root of the document.
//
// This uses a little code from Alan Haris open source (Copyright 2013) apns project.
//
// https://github.com/anachronistic/apns
//
// Alan's libraries were released under the MIT license.
//
type ApplePushNotificationSvc struct {
	Logger
	configPath string

	requestChannel chan *ApnMsg
	responseChannel chan *ApnMsg

	shutdownChannel chan bool

	gatewayProcessor *TcpSocketProcessor
	feedbackProcessor *TcpSocketProcessor

	gatewayWriteChannel chan TcpSocketProcessorWrite
	gatewayReadChannel chan TcpSocketProcessorRead

	feedbackWriteChannel chan TcpSocketProcessorWrite
	feedbackReadChannel chan TcpSocketProcessorRead

	updateStatsTicker *time.Ticker

	lruCache *lrucache.Cache
}

type ApnAlert struct {
	Body string `json:"body,omitempty"`
	ActionLocKey string `json:"action-loc-key,omitempty"`
	LocKey string `json:"loc-key,omitempty"`
	LocArgs []string `json:"loc-args,omitempty"`
	LaunchImage string `json:"launch-image,omitempty"`
}

// The apn feedback message is a response from Apple's feedback gateway. If the
// app does not exist on the device or the token has changed, they will send this
// information back.
type ApnFeedbackMsg struct {
	Timestamp   uint32
	DeviceToken string
}

// The apn aps struct. These values are mostly ints, but Go json considers
// a value of zero to to be empty causing 'omitempty' to be true and will
// not serialize the json with a zero. If you need to set the badge to
// zero, set it to -1. This is ugly, but it is either this or defining Badge as an
// interface. The alert field is either a nested json structure or a string. The nested json
// struct for the alert is defined by using the ApnAlert struct.
type ApnAps struct {
	Alert interface{} `json:"alert,omitempty"`
	Badge int `json:"badge,omitempty"`
	Sound string `json:"sound,omitempty"`
	ContentAvailable int `json:"content-available,omitempty"`
}

func NewApnAps() *ApnAps { return &ApnAps{} }

type ApnResponseMsg struct {
	DeviceToken string
	Feedback *ApnFeedbackMsg
}

// The apple push notification message. For more information, see: http://bit.ly/1kFnCnj
type ApnMsg struct {
	deviceToken string
	Payload map[string]interface{}
	identifier int32
	expiration uint32
	priority uint8

	// These fields are used by the feedback service.
	Timestamp uint32
	DeviceToken string

	// This field is used for error conditions. Anything other than a zero
	// means a send failure - see: http://bit.ly/1kNgAwn for more info.
	ApnErrorCode int
}

// The bytes method encodes the apn message so it can be sent to the apple.
func (self *ApnMsg) bytes() ([]byte, error) {

	tokenBytes, err := hex.DecodeString(self.deviceToken)
	if err != nil { return nil, NewStackError("Unable to convert string to bytes - err: %v", err) }

	payloadJson, err := json.Marshal(self.Payload)
	if err != nil { return nil, NewStackError("Unable to convert payload to json - err: %v", err) }

	if len(payloadJson) > apnMaxPayloadSize {
		return nil, NewStackError("Apple's max payload size is 256 bytes - you passed: %d", len(payloadJson))
	}

	var buffer bytes.Buffer

	// Device token
	binary.Write(&buffer, binary.BigEndian, uint8(apnDeviceTokenItemId))
	binary.Write(&buffer, binary.BigEndian, uint16(apnDeviceTokenLength))
	binary.Write(&buffer, binary.BigEndian, tokenBytes)

	// Payload
	binary.Write(&buffer, binary.BigEndian, uint8(apnPayloadItemId))
	binary.Write(&buffer, binary.BigEndian, uint16(len(payloadJson)))
	binary.Write(&buffer, binary.BigEndian, payloadJson)

	// Identifier
	binary.Write(&buffer, binary.BigEndian, uint8(apnNotificationIdentifierItemId))
	binary.Write(&buffer, binary.BigEndian, uint16(apnNotificationIdentifierLength))
	binary.Write(&buffer, binary.BigEndian, self.identifier)

	// Expiration
	binary.Write(&buffer, binary.BigEndian, uint8(apnExpirationDateItemId))
	binary.Write(&buffer, binary.BigEndian, uint16(apnExpirationDateLength))
	binary.Write(&buffer, binary.BigEndian, self.expiration)

	// Priority
	binary.Write(&buffer, binary.BigEndian, uint8(apnPriorityItemId))
	binary.Write(&buffer, binary.BigEndian, uint16(apnPriorityLength))
	binary.Write(&buffer, binary.BigEndian, self.priority)

	// Create the final message
	var msg bytes.Buffer
	binary.Write(&msg, binary.BigEndian, uint8(apnPushCommandValue))
	binary.Write(&msg, binary.BigEndian, uint32(buffer.Len()))
	binary.Write(&msg, binary.BigEndian, buffer.Bytes())

	return msg.Bytes(), nil
}

// Create a new apn message. You must specify the device token, the priority, expiration time
// and aps. The expiration time is a Unix epoch expressed in seconds. Set this to zero
// to have the message expire immediately. You must also specify the priority. Ten (10) is
// the highest priority, but cannot be used with the content-available key (error will result). You
// can add additional values to the payload by saying: msg.Payload["key"] = something. Do *not*
// use the key of "aps" because it is reserved by the apple for the apn aps struct.
func NewApnMsg(deviceToken string, expiration uint32, priority uint8, apnAps *ApnAps) *ApnMsg {
	return &ApnMsg {
		deviceToken: deviceToken,
		Payload: map[string]interface{} { "aps": apnAps },
		identifier: rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(9999),
		expiration: expiration,
		priority: priority,
	}
}

// Create the apple push notification service. Do not close the response channel. This will be closed by this
// component because this is the component that writes to it. Make sure you close the request channel before
// calling Stop() or you will leak goroutines.
func NewApplePushNotificationSvc(configPath string, requestChannel, responseChannel chan *ApnMsg) *ApplePushNotificationSvc {

	return &ApplePushNotificationSvc{	configPath: configPath,
										requestChannel: requestChannel,
										responseChannel: responseChannel,
										gatewayWriteChannel: make(chan TcpSocketProcessorWrite),
										gatewayReadChannel: make(chan TcpSocketProcessorRead),
										feedbackWriteChannel: make(chan TcpSocketProcessorWrite),
										feedbackReadChannel: make(chan TcpSocketProcessorRead),
										shutdownChannel: make(chan bool),
										updateStatsTicker: time.NewTicker(1 * time.Second),
	}
}

func (self *ApplePushNotificationSvc) processFeedbackReads() {

	for read := range self.feedbackReadChannel {

		if read.Error != nil { self.Logf(Error, "Unable to read apn data - err: %v", read.Error); continue }
		if read.BytesRead != apnFeedbackMsgSize { self.Logf(Error, "Bad apn feedback data - expected 38 bytes - read: %d", read.BytesRead); continue }
		if len(read.Data) != apnFeedbackMsgSize { self.Logf(Error, "Bad apn feedback data - expected 38 bytes - have: %d", len(read.Data)); continue }

		reader := bytes.NewReader(read.Data)

		apnMsg := &ApnMsg{}
		binary.Read(reader, binary.BigEndian, &apnMsg.Timestamp)

		var tokenLength uint16
		binary.Read(reader, binary.BigEndian, &tokenLength)

		deviceTokenRaw := make([]byte, 32, 32)
		binary.Read(reader, binary.BigEndian, &deviceTokenRaw)
		if tokenLength != 32 { self.Logf(Error, "Invalid feedback token length - expecting: 32 - receved: %d", tokenLength); continue }

		apnMsg.DeviceToken = hex.EncodeToString(deviceTokenRaw)

		self.responseChannel <- apnMsg
	}
}

func (self *ApplePushNotificationSvc) processGatwayReads() {

	for read := range self.gatewayReadChannel {

		if read.Error != nil { self.Logf(Error, "Unable to read apn data - err: %v", read.Error); continue }
		if read.BytesRead != apnErrorResponseMsgSize { self.Logf(Error, "Bad apn response data - expected 6 bytes - read: %d", read.BytesRead); continue }
		if len(read.Data) != apnErrorResponseMsgSize { self.Logf(Error, "Bad apn response data - expected 6 bytes - have: %d", len(read.Data)); continue }

		reader := bytes.NewReader(read.Data)

		command , err := reader.ReadByte()
		if err != nil { self.Logf(Error, "Unable to read apn command - err: %v", err); continue }

		if command != apnErrorResponseCommand { self.Logf(Error, "Bad response command - want: %d - received: %d", apnErrorResponseCommand, command); continue }

		status, err := reader.ReadByte()
		if err != nil { self.Logf(Error, "Unable to read apn status - err: %v", err); continue }

		if status == apnResponseErrorNone { continue }

		identifierRaw := make([]byte, 4, 4)
		read, err := reader.Read(identifierRaw)
		if err != nil { self.Logf(Error, "Unable to read apn identifier - err: %v", err); continue }
		if read != len(identifierRaw) { self.Logf(Error, "Unable to read apn identifier bytes - err: %v", err); continue }

		var identifier int32
		buf := bytes.NewBuffer(identifierRaw)
		binary.Read(buf, binary.BigEndian, &identifier)

		apnMsgRaw, found := self.lruCache.Get(identifier)
		if !found || apnMsgRaw == nil {
			self.Logf(Warn, "Unable to find msg for error condition - identifier: %d", identifier)
			continue
		}

		apnMsg := apnMsgRaw.(*ApnMsg)
		apnMsg.ApnErrorCode = int(status)

		self.responseChannel <- apnMsg
	}
}

// This is called to write the apn message to the socket.
func (self *ApplePushNotificationSvc) processMsg(msg *ApnMsg) {

	responseChannel := make(chan TcpSocketProcessorWrite)

	data, err := msg.bytes()

	if err != nil { self.Logf(Error, "Unable to convert message to bytes - err: %v", err); return }

	self.gatewayWriteChannel <- TcpSocketProcessorWrite{ Data: data, ResponseChannel: responseChannel }

	response := <- responseChannel

	if response.Error != nil { self.Logf(Error, "Unable to write message - err: %v", response.Error)
	} else if len(data) != response.BytesWritten { self.Logf(Error, "Bad socket write - wrote: %d - expected: %d", response.BytesWritten, len(data))
	} else { self.lruCache.Add(msg.identifier, msg) }
}

func (self *ApplePushNotificationSvc) listenForRequests() {
	for msg := range self.requestChannel { self.processMsg(msg) }
}

func (self *ApplePushNotificationSvc) Start(kernel *Kernel) error {

	gateway := strings.TrimSpace(kernel.Configuration.StringWithPath(self.configPath, "gateway", ""))
	if len(gateway) == 0 { return NewStackError("Did not set the apn.gateway config") }

	feedback := strings.TrimSpace(kernel.Configuration.StringWithPath(self.configPath, "feedback", ""))
	if len(feedback) == 0 { return NewStackError("Did not set the apn.feedback config") }

	certificateFile := strings.TrimSpace(kernel.Configuration.StringWithPath(self.configPath, "certificateFile", ""))
	if len(certificateFile) == 0 { return NewStackError("Did not set the apn.certificateFile config") }

	keyFile := strings.TrimSpace(kernel.Configuration.StringWithPath(self.configPath, "keyFile", ""))
	if len(keyFile) == 0 { return NewStackError("Did not set the apn.keyFile config") }

	socketTimeoutInMs := int64(kernel.Configuration.IntWithPath(self.configPath, "socketTimeoutInMs", 3000))

	var err error
	self.lruCache, err = lrucache.New(kernel.Configuration.IntWithPath(self.configPath, "msgCacheElementCount", 2000))
	if err != nil { return NewStackError("Unable to init lru cache - err: %v", err) }

	self.gatewayProcessor = NewTlsTcpSocketProcessor(	gateway,
														socketTimeoutInMs,
														0,
														socketTimeoutInMs,
														apnErrorResponseMsgSize,
														self.gatewayWriteChannel,
														self.gatewayReadChannel,
														kernel.Logger,
														certificateFile,
														keyFile)

	if err = self.gatewayProcessor.Start(); err != nil { return NewStackError("Unable to start gateway - err: %v", err) }

	self.feedbackProcessor = NewTlsTcpSocketProcessor(	feedback,
														socketTimeoutInMs,
														0,
														socketTimeoutInMs,
														apnFeedbackMsgSize,
														self.feedbackWriteChannel,
														self.feedbackReadChannel,
														kernel.Logger,
														certificateFile,
														keyFile)

	if err = self.feedbackProcessor.Start(); err != nil { return NewStackError("Unable to start feedback - err: %v", err) }

	go self.processGatwayReads()

	go self.processFeedbackReads()

	go self.listenForRequests()

	return nil
}

func (self *ApplePushNotificationSvc) Stop(kernel *Kernel) error {

	if self.feedbackProcessor != nil { self.feedbackProcessor.Stop() }
	if self.gatewayProcessor != nil { self.gatewayProcessor.Stop() }

	close(self.responseChannel)

	if self.lruCache != nil { self.lruCache.Purge() }

	return nil
}

