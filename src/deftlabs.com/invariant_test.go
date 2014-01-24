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

package deftlabsutil

import "testing"

// Confirm the way slices behave.
func TestSlices(t *testing.T) {

	var slice []byte

	if slice != nil {
		t.Errorf("TestSlice is broken - slice is not nil")
	}

	if len(slice) != 0 {
		t.Errorf("TestSlice is broken - slice length is not zero")
	}
}
