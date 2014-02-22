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
	"testing"
	"encoding/json"
)

func TestLibrato(t *testing.T) {

	testUrl := assembleLibratoUrl(LibratoMetricsPostUrl, "bill@microsoft.com", "testtoken")

	if testUrl != "https://bill%40microsoft.com:testtoken@metrics-api.librato.com/v1/metrics" {
		t.Errorf("TestLibrato assembleLibratoUrl is broken - received: %s", testUrl)
	}

	librato := NewLibrato("bill@microsoft.com", "testtoken", Logger{})

	if librato == nil {
		t.Errorf("TestLibrato NewLibrato is broken")
	}

	testUrl = assembleLibratoUrl(LibratoMetricsPostUrl, "", "")
}

func TestLibratoMsgJson(t *testing.T) {
	msg := libratoMsg{}

	msg.Counters = append(msg.Counters, libratoMetric{ Name: "test", Value: float64(100), Source: "source" })

	msg.Counters = append(msg.Counters, libratoMetric{ Name: "test1", Value: float64(200), Source: "source" })

	rawJson, err := json.Marshal(msg)
	if err != nil {
		t.Errorf("TestLibratoMsgJson json marshal is broken - error %v", err)
	}

	if len(rawJson) == 0 {
		t.Errorf("TestLibratoMsgJson json marshal is broken - empty json")
	}

}


