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
	"crypto/tls"
	"labix.org/v2/mgo/bson"
	"encoding/json"
	"github.com/mreiferson/go-httpclient"
)

const (
	DefaultSocketTimeout = 40000
	HttpPostMethod = "POST"
	HttpGetMethod = "GET"

	ContentTypeHeader = "Content-Type"

	NoStatusCode = -1 // This is the value used if an error is generated before the status code is available

	ContentTypeTextPlain = "text/plain; charset=utf-8"
	ContentTypePostForm = "application/x-www-form-urlencoded"
	ContentTypeJson = "application/json"
)

// The http request client interface allows you to configure global values for timeout/connect etc. It also
// handles closing the transport and provides convenience methods for posting some types of data.
type HttpRequestClient interface {
	Post(url string, values url.Values, headers map[string]string) (int, []byte, error)
	PostStr(url string, value string, headers map[string]string) (int, []byte, error)
	PostJson(url string, value interface{}, headers map[string]string) (int, []byte, error)
	PostBson(url string, bsonDoc interface{}, headers map[string]string) (int, []byte, error)
	GetBson(url string, headers map[string]string) (int, bson.M, error)
	Clone() HttpRequestClient
	GetClientAndTransport() (*http.Client, *httpclient.Transport)
	GetTransport() *httpclient.Transport
}

type HttpRequestClientImpl struct {
	disableKeepAlives bool
	disableCompression bool
	skipSslVerify bool
	maxIdleConnsPerHost int
	connectTimeout time.Duration
	responseHeaderTimeout time.Duration
	requestTimeout time.Duration
	readWriteTimeout time.Duration
}

func NewDefaultHttpRequestClient() HttpRequestClient {
	return NewHttpRequestClient(true, false, false, 0, DefaultSocketTimeout, DefaultSocketTimeout, DefaultSocketTimeout, DefaultSocketTimeout)
}

func NewHttpRequestClient(	disableKeepAlives,
							disableCompression,
							skipSslVerify bool,
							maxIdleConnsPerHost,
							connectTimeoutInMs,
							responseHeaderTimeoutInMs,
							requestTimeoutInMs,
							readWriteTimeoutInMs int) HttpRequestClient {

	return &HttpRequestClientImpl{ 	disableKeepAlives: disableKeepAlives,
									disableCompression: disableCompression,
									skipSslVerify: skipSslVerify,
									maxIdleConnsPerHost: maxIdleConnsPerHost,
									connectTimeout: time.Duration(connectTimeoutInMs) * time.Millisecond,
									responseHeaderTimeout: time.Duration(responseHeaderTimeoutInMs) * time.Millisecond,
									requestTimeout: time.Duration(requestTimeoutInMs) * time.Millisecond,
									readWriteTimeout: time.Duration(readWriteTimeoutInMs) * time.Millisecond,
	}
}

// Post the values to the url.
func (self *HttpRequestClientImpl) Post(url string, values url.Values, headers map[string]string) (int, []byte, error) {

	client, transport := self.GetClientAndTransport()
	defer transport.Close()

	request, err := http.NewRequest(HttpPostMethod, url,  strings.NewReader(values.Encode()))
	if err != nil { return NoStatusCode, nil, err }

	request.Header.Set(ContentTypeHeader, ContentTypePostForm)
	for key, value := range headers { request.Header.Set(key, value) }

	response, err := client.Do(request)
	if err != nil { return NoStatusCode, nil, err }

	statusCode := response.StatusCode

	defer response.Body.Close()

	if data, err := ioutil.ReadAll(response.Body); err != nil { return statusCode, nil, err
	} else { return statusCode, data, nil }
}

// Post the raw string to the url.
func (self *HttpRequestClientImpl) PostStr(url string, value string, headers map[string]string) (int, []byte, error) {

	client, transport := self.GetClientAndTransport()
	defer transport.Close()

	request, err := http.NewRequest(HttpPostMethod, url, bytes.NewReader([]byte(value)))
	if err != nil { return NoStatusCode, nil, err }

	request.Header.Set(ContentTypeHeader, ContentTypePostForm)
	for key, value := range headers { request.Header.Set(key, value) }

	response, err := client.Do(request)
	if err != nil { return NoStatusCode, nil, err }

	statusCode := response.StatusCode

	defer response.Body.Close()

	if data, err := ioutil.ReadAll(response.Body); err != nil { return statusCode, nil, err
	} else { return statusCode, data, nil }
}

// Post the json struct to the url.
func (self *HttpRequestClientImpl) PostJson(url string, value interface{}, headers map[string]string) (int, []byte, error) {

	rawJson, err := json.Marshal(value)
	if err != nil { return NoStatusCode, nil, err }

	client, transport := self.GetClientAndTransport()
	defer transport.Close()

	request, err := http.NewRequest(HttpPostMethod, url, bytes.NewReader(rawJson))
	if err != nil { return NoStatusCode, nil, err }

	request.Header.Set(ContentTypeHeader, ContentTypeJson)
	for key, value := range headers { request.Header.Set(key, value) }

	response, err := client.Do(request)
	if err != nil { return NoStatusCode, nil, err }

	statusCode := response.StatusCode

	defer response.Body.Close()

	if data, err := ioutil.ReadAll(response.Body); err != nil { return statusCode, nil, err
	} else { return statusCode, data, nil }
}

// Post the bson doc to the url.
func (self *HttpRequestClientImpl) PostBson(url string, bsonDoc interface{}, headers map[string]string) (int, []byte, error) {

	rawBson, err := bson.Marshal(bsonDoc)
	if err != nil { return NoStatusCode, nil, err }

	client, transport := self.GetClientAndTransport()
	defer transport.Close()

	request, err := http.NewRequest(HttpPostMethod, url, bytes.NewReader(rawBson))
	if err != nil { return NoStatusCode, nil, err }

	request.Header.Set(ContentTypeHeader, ContentTypePostForm)
	for key, value := range headers { request.Header.Set(key, value) }

	response, err := client.Do(request)
	if err != nil { return NoStatusCode, nil, err }

	statusCode := response.StatusCode

	defer response.Body.Close()

	// We do not return the response so don't report if there is an error.
	if data, err := ioutil.ReadAll(response.Body); err != nil { return statusCode, nil, err
	} else { return statusCode, data, nil }
}

// Issue a GET to retrieve a bson doc.
func (self *HttpRequestClientImpl) GetBson(url string, headers map[string]string) (int, bson.M, error) {

	client, transport := self.GetClientAndTransport()
	defer transport.Close()

	request, requestErr := http.NewRequest(HttpGetMethod, url, nil)
	if requestErr != nil { return NoStatusCode, nil, requestErr }

	for key, value := range headers { request.Header.Set(key, value) }

	response, err := client.Do(request)
	if err != nil { return NoStatusCode, nil, err }

	statusCode := response.StatusCode

	defer response.Body.Close()

	rawBson, err := ioutil.ReadAll(response.Body)
	if err != nil { return statusCode, nil, err }

	var bsonDoc bson.M
	if err := bson.Unmarshal(rawBson, &bsonDoc); err != nil { return statusCode, nil, err }

	return statusCode, bsonDoc, nil
}

// Use the clone method if you need to override some/all of the configured values.
func (self *HttpRequestClientImpl) Clone() HttpRequestClient {
	val := *self
	return &val
}

func (self *HttpRequestClientImpl) GetClientAndTransport() (*http.Client, *httpclient.Transport) {
	transport := self.GetTransport()
	return &http.Client{ Transport: transport }, transport
}

func (self *HttpRequestClientImpl) GetTransport() *httpclient.Transport {
	return &httpclient.Transport {

		DisableKeepAlives: self.disableKeepAlives,
		DisableCompression: self.disableCompression,

		MaxIdleConnsPerHost: self.maxIdleConnsPerHost,

		ConnectTimeout: self.connectTimeout,
		ResponseHeaderTimeout: self.responseHeaderTimeout,
		RequestTimeout: self.requestTimeout,
		ReadWriteTimeout: self.readWriteTimeout,

		TLSClientConfig: &tls.Config{ InsecureSkipVerify: self.skipSslVerify },
	}
}

func HttpPost(url string, values url.Values, headers map[string]string) (int, []byte, error) { return NewDefaultHttpRequestClient().Post(url, values, headers) }

func HttpPostStr(url string, value string, headers map[string]string) (int, []byte, error) { return NewDefaultHttpRequestClient().PostStr(url, value, headers) }

func HttpPostJson(url string, value interface{}, headers map[string]string) (int, []byte, error) { return NewDefaultHttpRequestClient().PostJson(url, value, headers) }

func HttpPostBson(url string, bsonDoc interface{}, headers map[string]string) (int, []byte, error) { return NewDefaultHttpRequestClient().PostBson(url, bsonDoc, headers) }

func HttpGetBson(url string, headers map[string]string) (int, bson.M, error) { return NewDefaultHttpRequestClient().GetBson(url, headers) }

