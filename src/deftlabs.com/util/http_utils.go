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
	"fmt"
	"github.com/mreiferson/go-httpclient"
	"net/http"
	"io/ioutil"
	"time"
	"bytes"
	"strings"
	"errors"
	"labix.org/v2/mgo/bson"
)

const (
	SocketTimeout = 40
	HttpPostMethod = "POST"
	ContentTypeHeader = "Content-Type"
	ContentTypeTextPlain = "text/plain; charset=utf-8"
)

func HttpPostBson(url string, bsonDoc interface{}) error {

	rawBson, err := bson.Marshal(bsonDoc)
	if (err != nil) {
		return err
	}

	httpClient, httpTransport := getDefaultHttpClient()
	defer httpTransport.Close()

	request, err := http.NewRequest("POST", url, bytes.NewReader(rawBson))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	// We do not return the response so don't report if there is an error.
	ioutil.ReadAll(response.Body)

	return nil
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

	request, requestErr := http.NewRequest("GET", url, nil)
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
// field missing or incorrect, false is returned. This method will panic if the request
// is nil.
func IsHttpMethodPost(request *http.Request) bool {
	if request == nil {
		panic("request param is nil")
	}
	return len(request.Method) > 0 && strings.ToUpper(request.Method) == HttpPostMethod
}

// Write an http ok response string. The content type is text/plain.
func WriteOkResponseString(response http.ResponseWriter, msg string) error {
	if response == nil {
		panic("response param is nil")
	}

	msgLength := len(msg)

	if msgLength == 0 {
		panic("do not write an empty string to the response")
	}

	response.Header().Set(ContentTypeHeader, ContentTypeTextPlain)

	written, err := response.Write([]byte(msg))

	if err != nil {
		return err
	}

	if written != msgLength {
		return errors.New(fmt.Sprintf("Did not write full message - bytes written %d - expected %d", written, msgLength))
	}

	return nil
}

