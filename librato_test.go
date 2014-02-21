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

func TestLibrato(t *testing.T) {

	testUrl := assembleLibratoUrl(LibratoMetricsPostUrl, "bill@microsoft.com", "testtoken")

	if testUrl != "https://bill%40microsoft.com:testtoken@metrics-api.librato.com/v1/metrics" {
		t.Errorf("TestLibrato assembleLibratoUrl is broken - received: %s", testUrl)
	}

	librato := NewLibrato("bill@microsoft.com", "testtoken", Logger{})
}

