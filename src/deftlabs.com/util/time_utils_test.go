/**
 * (C) Copyright 2013, Deft Labs
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

package deftlabsutil

import "testing"

func TestNowTimeUnixStr(t *testing.T) {
	if str := NowTimeUnixStr(); len(str) == 0 {
		t.Errorf("NowTimeUnixStr is broken")
	}
}

func TestCurrentTimeInMillis(t *testing.T) {
	if millis := CurrentTimeInMillis(); millis <= 0 {
		t.Errorf("CurrentTimeInMillis is broken")
	}
}

func TestCurrentTimeInSeconds(t *testing.T) {
	if seconds := CurrentTimeInSeconds(); seconds <= 0 {
		t.Errorf("CurrentTimeInSeconds is broken")
	}
}

