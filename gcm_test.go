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

import (
	"sync"
	"testing"
	"encoding/json"
)

// Test the google cloud messaging service.
func TestGcm(t *testing.T) {

	mockResponse := testCreateGoogleCloudMsgResponse(100, 1, 0, 0)
	testAddGoogleCloudMsgResponseResult(mockResponse, "someMessageId", nadaStr, nadaStr)

	data, err := json.Marshal(mockResponse)
	if err != nil { t.Errorf("TestGcm json encode mock response broken - err: %v", err); return }

	httpClient := NewHttpRequestClientMock()
	httpClient.(*HttpRequestClientMock).AddMock("https://android.googleapis.com/gcm/send", &HttpRequestClientMockResponse{
		HttpStatusCode: 200,
		Data: data,
	})

	requestChannel := make(chan interface{})
	responseChannel := make(chan interface{})

	svc := NewGoogleCloudMessagingSvc("gcm", httpClient, requestChannel, responseChannel)

	kernel, err := baseTestStartKernel("gcmTest", func(kernel *Kernel) {
		kernel.AddComponentWithStartStopMethods("GoogleCloudMessagingSvc", svc, "Start", "Stop")
	})

	if err != nil { t.Errorf("TestGcm start kernel is broken: %v", err); return }

	msgSendCount := 100000
	msgReceivedCount := 0

	var waitGroup sync.WaitGroup

	go func() {
		waitGroup.Add(1)
		defer waitGroup.Done()
		for {
			msg := <- responseChannel
			if msg == nil { t.Errorf("TestGcm is broken - response message is nil") }
			msgReceivedCount++
			if msgReceivedCount == msgSendCount { return }
		}
	}()

	for idx := 0; idx < msgSendCount; idx++ {
		requestChannel <- &GoogleCloudMsg{
			RegistrationIds: []string { "someRegistrationId" },
			CollapseKey: "someCollapseKey",
			DelayWhileIdle: true,
			TimeToLive: 300,
			RestrictedPackageName: "somePackageName",
			DryRun: false,
			Data: map[string]interface{} { "someKey": "someValue" },
		}
	}

	waitGroup.Wait()

	close(requestChannel)

	if err := kernel.Stop(); err != nil { t.Errorf("TestGcm stop kernel is broken:", err) }
}

func testCreateGoogleCloudMsgResponse(multicastId, success, failure, canonicalIds float64) *GoogleCloudMsgResponse {
	return &GoogleCloudMsgResponse {
		MulticastId: multicastId,
		Success: success,
		Failure: failure,
		CanonicalIds: canonicalIds,
	}
}

func testAddGoogleCloudMsgResponseResult(response *GoogleCloudMsgResponse, messageId, registrationId, err string) {
	response.Results = append(response.Results, &GoogleCloudMsgResponseResult{
		MessageId: messageId,
		RegistrationId: registrationId,
		Error: err,
	})
}

