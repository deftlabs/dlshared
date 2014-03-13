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
	"bytes"
	"net/url"
	"net/http"
)

func NewTestHttpPostRequest() *http.Request {
	request := &http.Request{ Method : "POST" }
	request.URL, _ = url.Parse("http://www.google.com/search?q=foo&q=bar")
	return request
}

type TestReadCloserBuffer struct { *bytes.Buffer }
func (self *TestReadCloserBuffer) Close() error { return nil }

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

