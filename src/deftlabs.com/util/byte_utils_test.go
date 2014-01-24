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


import (
	//"fmt"
	"bytes"
	"testing"
	"encoding/binary"
)

func TestExtractUInt32(t *testing.T) {

	buf := new(bytes.Buffer)

	var value uint32 = 4294967295

	if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
		t.Errorf("binary.Write failed:", err)
	}

	raw := buf.Bytes()

	if len(raw) != 4 {
		t.Errorf("TestExtractUInt32 is broken result buffer length is not four")
	}

	result := ExtractUInt32(raw, 0)

	if result != value {
		t.Errorf("TestExtractUInt32 is broken result does not equal value")
	}
}

func TestExtractUInt16(t *testing.T) {

	buf := new(bytes.Buffer)

	var value uint16 = 65535

	if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
		t.Errorf("binary.Write failed:", err)
	}

	raw := buf.Bytes()

	if len(raw) != 2 {
		t.Errorf("TestExtractUInt32 is broken result buffer length is not two")
	}

	result := ExtractUInt16(raw, 0)

	if result != value {
		t.Errorf("TestExtractUInt32 is broken result does not equal value")
	}
}

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

