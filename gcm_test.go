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

import "testing"

// Test the google cloud messaging service.
func TestGcm(t *testing.T) {

	svc := NewGoogleCloudMessagingSvc("gcm", nil)

	kernel, err := baseTestStartKernel("gcmTest", func(kernel *Kernel) {
		kernel.AddComponentWithStartStopMethods("GoogleCloudMessagingSvc", svc, "Start", "Stop")
	})

	if err != nil { t.Errorf("TestGcm start kernel is broken: %v", err); return }

	// TODO: Simulate

	svc.AddResponseHandler(func(response *GoogleCloudMsgResponse) {

	})

	if err := kernel.Stop(); err != nil { t.Errorf("TestGcm stop kernel is broken:", err) }
}

