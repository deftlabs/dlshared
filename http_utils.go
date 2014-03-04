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
	"net/url"
	"net/http"
	"io/ioutil"
	"time"
	"bytes"
	"strings"
	"labix.org/v2/mgo/bson"
	"encoding/json"
	"github.com/mreiferson/go-httpclient"
)

const (
	SocketTimeout = 40
	HttpPostMethod = "POST"
	HttpGetMethod = "GET"

	ContentTypeHeader = "Content-Type"

	ContentTypeTextPlain = "text/plain; charset=utf-8"
	ContentTypePostForm = "application/x-www-form-urlencoded"
	ContentTypeJson = "application/json"
)

func HttpPost(url string, values url.Values) ([]byte, error) {

	httpClient, httpTransport := getDefaultHttpClient()
	defer httpTransport.Close()

	response, err := httpClient.PostForm(url, values)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if data, err := ioutil.ReadAll(response.Body); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func HttpPostStr(url string, value string) ([]byte, error) {

	httpClient, httpTransport := getDefaultHttpClient()
	defer httpTransport.Close()

	request, err := http.NewRequest(HttpPostMethod, url, bytes.NewReader([]byte(value)))
	if err != nil {
		return nil, err
	}
	request.Header.Set(ContentTypeHeader, ContentTypePostForm)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if data, err := ioutil.ReadAll(response.Body); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func HttpPostJson(url string, value interface{}) ([]byte, error) {

	rawJson, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	httpClient, httpTransport := getDefaultHttpClient()
	defer httpTransport.Close()

	request, err := http.NewRequest(HttpPostMethod, url, bytes.NewReader(rawJson))
	if err != nil {
		return nil, err
	}
	request.Header.Set(ContentTypeHeader, ContentTypeJson)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if data, err := ioutil.ReadAll(response.Body); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func HttpPostBson(url string, bsonDoc interface{}) ([]byte, error) {

	rawBson, err := bson.Marshal(bsonDoc)
	if err != nil {
		return nil, err
	}

	httpClient, httpTransport := getDefaultHttpClient()
	defer httpTransport.Close()

	request, err := http.NewRequest("POST", url, bytes.NewReader(rawBson))
	if err != nil {
		return nil, err
	}
	request.Header.Set(ContentTypeHeader, ContentTypePostForm)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	// We do not return the response so don't report if there is an error.
	if data, err := ioutil.ReadAll(response.Body); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func getDefaultHttpClient() (*http.Client, *httpclient.Transport) {
	transport := getDefaultHttpTransport()
	return &http.Client{ Transport: transport }, transport
}

func getDefaultHttpTransport() *httpclient.Transport {
	return &httpclient.Transport {
		ConnectTimeout:        SocketTimeout * time.Second,
		RequestTimeout:        SocketTimeout * time.Second,
		ResponseHeaderTimeout: SocketTimeout * time.Second,
	}
}

func HttpGetBson(url string) (bson.M, error) {

	httpClient, httpTransport := getDefaultHttpClient()
	defer httpTransport.Close()

	request, requestErr := http.NewRequest(HttpGetMethod, url, nil)
	if requestErr != nil {
		return nil, requestErr
	}

	response, err := httpClient.Do(request)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	rawBson, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var bsonDoc bson.M
	if err := bson.Unmarshal(rawBson, &bsonDoc); err != nil {
		return nil, err
	}

	return bsonDoc, nil
}

// This method returns true if the http request method is a HTTP post. If the
// field missing or incorrect, false is returned. This method will panic if
// the request is nil.
func IsHttpMethodPost(request *http.Request) bool {
	if request == nil {
		panic("request param is nil")
	}
	return len(request.Method) > 0 && strings.ToUpper(request.Method) == HttpPostMethod
}

// Encode and write a json response. If there is a problem encoding an http 500 is sent and an
// error is returned. If there are problems writting the response an error is returned.
func JsonEncodeAndWriteResponse(response http.ResponseWriter, value interface{}) error {

	if value == nil {
		return NewStackError("Nil value passed")
	}

	rawJson, err := json.Marshal(value)
	if err != nil {
		http.Error(response, "Error", 500)
		return NewStackError("Unable to marshal json: %v", err)
	}

	response.Header().Set(ContentTypeHeader, ContentTypeJson)

	written, err := response.Write(rawJson)
	if err != nil {
		return NewStackError("Unable to write response: %v", err)
	}

	if written != len(rawJson) {
		return NewStackError("Unable to write full response - wrote: %d - expected: %d", written, len(rawJson))
	}

	return nil
}

// Write an http ok response string. The content type is text/plain.
func WriteOkResponseString(response http.ResponseWriter, msg string) error {
	if response == nil {
		return NewStackError("response param is nil")
	}

	msgLength := len(msg)

	if msgLength == 0 {
		return NewStackError("Response message is an empty string")
	}

	response.Header().Set(ContentTypeHeader, ContentTypeTextPlain)

	written, err := response.Write([]byte(msg))

	if err != nil {
		return err
	}

	if written != msgLength {
		return NewStackError("Did not write full message - bytes written %d - expected %d", written, msgLength)
	}

	return nil
}

