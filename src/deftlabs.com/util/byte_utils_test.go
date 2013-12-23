/**
 * (C) Copyright 2013 Deft Labs
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


import (
	"testing"
)

func TestInt32 (t *testing.T) {

	data := make([]byte, 4)
	WriteInt32(data, 100, 0)

	if ExtractInt32(data, 0) != 100 {
		t.Errorf("TestInt32 - write/extract int32 is broken")
	}
}

func TestInt64 (t *testing.T) {

	data := make([]byte, 8)
	WriteInt64(data, 1000000, 0)

	if ExtractInt64(data, 0) != 1000000 {
		t.Errorf("TestInt64 - write/extract int64 is broken")
	}
}

