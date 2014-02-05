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
	"testing"
	"net/http"
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

	writeOkResponseStringEmptyMsgPanic(response, t)

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

	// Verify the panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("WriteOkResponseString is broken - it did not panic on nil response")
		}
	}()

	// This will panic
	WriteOkResponseString(nil, "")
}

func writeOkResponseStringEmptyMsgPanic(response http.ResponseWriter, t *testing.T) {
	// Verify the panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("WriteOkResponseString is broken - it did not panic on an empty message")
		}
	}()

	WriteOkResponseString(response, "")
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




