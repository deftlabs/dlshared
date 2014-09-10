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
	//"fmt"
	//"time"
	"testing"
)

// Test the apple push notification (apn) service.
func TestApn(t *testing.T) {
	/*
	TODO: Add a mock service - this requires abstracting out the TcpSocketProcessor a bit.

	testDeviceToken := "SOME_DEVICE_TOKEN"

	requestChannel := make(chan *ApnMsg)
	responseChannel := make(chan *ApnMsg)

	svc := NewApplePushNotificationSvc("apn", requestChannel, responseChannel)

	kernel, err := baseTestStartKernel("apnTest", func(kernel *Kernel) {
		kernel.AddComponentWithStartStopMethods("ApplePushNotificationSvc", svc, "Start", "Stop")
	})

	if err != nil { t.Errorf("TestApn start kernel is broken: %v", err); return }

	msg := NewApnMsg(testDeviceToken, 10, 8, &ApnAps{ ContentAvailable: 1 })

	requestChannel <- msg

	fmt.Println("Message sent")

	go func() {

		for responseMsg := range responseChannel {
			fmt.Println("response received")
			fmt.Println(responseMsg.DeviceToken)
			fmt.Println(responseMsg.ApnErrorCode)
		}
	}()

	time.Sleep(300 * time.Second)

	if err := kernel.Stop(); err != nil { t.Errorf("TestGcm stop kernel is broken:", err) }
	*/
}

