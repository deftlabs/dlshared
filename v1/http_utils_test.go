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

package dlshared

import (
	"math"
	"testing"
	"net/http"
	"encoding/json"
)

type RecordingResponseWriter struct {
	header http.Header
	HeaderCode int
	Data []byte
}

func (self *RecordingResponseWriter) reset() {
	self.header = make(map[string][]string)
	self.HeaderCode = 0
	self.Data = nil
}

func (self *RecordingResponseWriter) Header() http.Header {
	return self.header
}

func (self *RecordingResponseWriter) Write(data []byte) (int, error) {
	self.Data = append(self.Data, data...)
	return len(data), nil
}

func NewRecordingResponseWriter() *RecordingResponseWriter {
	return &RecordingResponseWriter{ header : make(map[string][]string) }
}

func (self *RecordingResponseWriter) WriteHeader(code int) {
	self.HeaderCode = code
}

type testJsonStruct struct {
	String string
	Boolean bool
	Number float64
}

func TestJsonEncodeAndWriteResponse(t *testing.T) {

	response := NewRecordingResponseWriter()

	test := &testJsonStruct{ String: "test", Boolean: true, Number: math.MaxFloat64 }

	// Write the data
	if err := JsonEncodeAndWriteResponse(response, test); err != nil {
		t.Errorf("JsonEncodeAndWriteResponse is broken - %v", err)
	}

	// Ensure the response
	decoded := &testJsonStruct{}
	if err := json.Unmarshal(response.Data, decoded); err != nil {
		t.Errorf("JsonEncodeAndWriteResponse unmarshal data is broken - %v", err)
	}

	if test.String != decoded.String {
		t.Errorf("JsonEncodeAndWriteResponse is broken - expected string: %s - received: %s", test.String, decoded.String)
	}

	if test.Boolean != decoded.Boolean {
		t.Errorf("JsonEncodeAndWriteResponse is broken - expected bool : %s - received: %s", test.Boolean, decoded.Boolean)
	}

	if test.Number != decoded.Number {
		t.Errorf("JsonEncodeAndWriteResponse is broken - expected number : %s - received: %s", test.Number, decoded.Number)
	}

}

func TestWriteOkResponseString(t *testing.T) {

	response := NewRecordingResponseWriter()

	if err := WriteOkResponseString(response, "test"); err != nil {
		t.Errorf("IsHttpMethodPost is broken - %v", err)
	}

	if string(response.Data) != "test" {
		t.Errorf("IsHttpMethodPost is broken")
	}

	if response.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("IsHttpMethodPost Content-Type is broken")
	}

	response.reset()

	if response.Header().Get("Content-Type") != "" {
		t.Errorf("IsHttpMethodPost reset is broken")
	}

	err := WriteOkResponseString(response, "")
	if err == nil {
		t.Errorf("WriteOkResponseString is broken - no error on empty message")
	}

	//writeOkResponseStringEmptyMsgPanic(response, t)

	if err := WriteOkResponseString(response, "t"); err != nil {
		t.Errorf("IsHttpMethodPost is broken - %v", err)
	}

	if string(response.Data) != "t" {
		t.Errorf("IsHttpMethodPost is broken")
	}

	response.reset()
	if err := WriteOkResponseString(response, "tttttttttttttttttttttttttttttttttttttttttttttttt"); err != nil {
		t.Errorf("IsHttpMethodPost is broken - %v", err)
	}

	// This will panic
	err = WriteOkResponseString(nil, "")
	if err == nil {
		t.Errorf("WriteOkResponseString is broken  - no error on nil response param")
	}
}

func TestIsHttpMethodPost(t *testing.T) {

	if IsHttpMethodPost(&http.Request{ Method : "" }) {
		t.Errorf("IsHttpMethodPost is broken")
	}

	if IsHttpMethodPost(&http.Request{ Method : "wrong" }) {
		t.Errorf("IsHttpMethodPost is broken")
	}

	if !IsHttpMethodPost(&http.Request{ Method : "post" }) {
		t.Errorf("IsHttpMethodPost is broken")
	}

	if !IsHttpMethodPost(&http.Request{ Method : "Post" }) {
		t.Errorf("IsHttpMethodPost is broken")
	}

	if !IsHttpMethodPost(&http.Request{ Method : "POST" }) {
		t.Errorf("IsHttpMethodPost is broken")
	}

	if !IsHttpMethodPost(&http.Request{ Method : "PosT" }) {
		t.Errorf("IsHttpMethodPost is broken")
	}

	// Verify the panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("IsHttpMethodPost is broken - it did not panic on nil request")
		}
	}()

	// This method will panic.
	IsHttpMethodPost(nil)
}

