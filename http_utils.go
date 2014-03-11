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
	DefaultSocketTimeout = 40
	HttpPostMethod = "POST"
	HttpGetMethod = "GET"

	ContentTypeHeader = "Content-Type"

	ContentTypeTextPlain = "text/plain; charset=utf-8"
	ContentTypePostForm = "application/x-www-form-urlencoded"
	ContentTypeJson = "application/json"
)

// The http request client allows you to configure global values for timeout/connect etc. It also
// handles closing the transport and provides convenience methods for posting some types of data.
type HttpRequestClient struct {
	DisableKeepAlives bool
	DisableCompression bool
	SkipSslVerify bool
	MaxIdleConnsPerHost int
	ConnectTimeout time.Duration
	ResponseHeaderTimeout time.Duration
	RequestTimeout time.Duration
	ReadWriteTimeout time.Duration
}

func NewDefaultHttpRequestClient() *HttpRequestClient {
	return NewHttpRequestClient(true, false, false, 0, DefaultSocketTimeout, DefaultSocketTimeout, DefaultSocketTimeout, DefaultSocketTimeout)
}

func NewHttpRequestClient(	disableKeepAlives,
							disableCompression,
							skipSslVerify bool,
							maxIdleConnsPerHost,
							connectTimeoutInMs,
							responseHeaderTimeoutInMs,
							requestTimeoutInMs,
							readWriteTimeoutInMs int) *HttpRequestClient {

	return &HttpRequestClient{ 	DisableKeepAlives: disableKeepAlives,
								DisableCompression: disableCompression,
								SkipSslVerify: skipSslVerify,
								MaxIdleConnsPerHost: maxIdleConnsPerHost,
								ConnectTimeout: time.Duration(connectTimeoutInMs) * time.Millisecond,
								ResponseHeaderTimeout: time.Duration(responseHeaderTimeoutInMs) * time.Millisecond,
								RequestTimeout: time.Duration(requestTimeoutInMs) * time.Millisecond,
								ReadWriteTimeout: time.Duration(readWriteTimeoutInMs) * time.Millisecond,
	}
}

// Post the values to the url.
func (self *HttpRequestClient) Post(url string, values url.Values, headers map[string]string) ([]byte, error) {

	client, transport := self.getClientAndTransport()
	defer transport.Close()

	request, err := http.NewRequest(HttpPostMethod, url,  strings.NewReader(values.Encode()))
	if err != nil { return nil, err }

	request.Header.Set(ContentTypeHeader, ContentTypePostForm)
	for key, value := range headers { request.Header.Set(key, value) }

	response, err := client.Do(request)
	if err != nil { return nil, err }

	defer response.Body.Close()

	if data, err := ioutil.ReadAll(response.Body); err != nil { return nil, err
	} else { return data, nil }
}

// Post the raw string to the url.
func (self *HttpRequestClient) PostStr(url string, value string, headers map[string]string) ([]byte, error) {

	client, transport := self.getClientAndTransport()
	defer transport.Close()

	request, err := http.NewRequest(HttpPostMethod, url, bytes.NewReader([]byte(value)))
	if err != nil { return nil, err }

	request.Header.Set(ContentTypeHeader, ContentTypePostForm)
	for key, value := range headers { request.Header.Set(key, value) }

	response, err := client.Do(request)
	if err != nil { return nil, err }

	defer response.Body.Close()

	if data, err := ioutil.ReadAll(response.Body); err != nil { return nil, err
	} else { return data, nil }
}

// Post the json struct to the url.
func (self *HttpRequestClient) PostJson(url string, value interface{}, headers map[string]string) ([]byte, error) {

	rawJson, err := json.Marshal(value)
	if err != nil { return nil, err }

	client, transport := self.getClientAndTransport()
	defer transport.Close()

	request, err := http.NewRequest(HttpPostMethod, url, bytes.NewReader(rawJson))
	if err != nil { return nil, err }

	request.Header.Set(ContentTypeHeader, ContentTypeJson)
	for key, value := range headers { request.Header.Set(key, value) }

	response, err := client.Do(request)
	if err != nil { return nil, err }

	defer response.Body.Close()

	if data, err := ioutil.ReadAll(response.Body); err != nil { return nil, err
	} else { return data, nil }
}

// Post the bson doc to the url.
func (self *HttpRequestClient) PostBson(url string, bsonDoc interface{}, headers map[string]string) ([]byte, error) {

	rawBson, err := bson.Marshal(bsonDoc)
	if err != nil { return nil, err }

	client, transport := self.getClientAndTransport()
	defer transport.Close()

	request, err := http.NewRequest(HttpPostMethod, url, bytes.NewReader(rawBson))
	if err != nil { return nil, err }

	request.Header.Set(ContentTypeHeader, ContentTypePostForm)
	for key, value := range headers { request.Header.Set(key, value) }

	response, err := client.Do(request)
	if err != nil { return nil, err }

	defer response.Body.Close()

	// We do not return the response so don't report if there is an error.
	if data, err := ioutil.ReadAll(response.Body); err != nil { return nil, err
	} else { return data, nil }
}

// Issue a GET to retrieve a bson doc.
func (self *HttpRequestClient) GetBson(url string, headers map[string]string) (bson.M, error) {

	client, transport := self.getClientAndTransport()
	defer transport.Close()

	request, requestErr := http.NewRequest(HttpGetMethod, url, nil)
	if requestErr != nil { return nil, requestErr }

	for key, value := range headers { request.Header.Set(key, value) }

	response, err := client.Do(request)

	if err != nil { return nil, err }

	defer response.Body.Close()

	rawBson, err := ioutil.ReadAll(response.Body)
	if err != nil { return nil, err }

	var bsonDoc bson.M
	if err := bson.Unmarshal(rawBson, &bsonDoc); err != nil { return nil, err }

	return bsonDoc, nil
}

// Use the clone method if you need to override some/all of the configured values.
func (self *HttpRequestClient) Clone() *HttpRequestClient {
	val := *self
	return &val
}

func (self *HttpRequestClient) getClientAndTransport() (*http.Client, *httpclient.Transport) {
	transport := self.getTransport()
	return &http.Client{ Transport: transport }, transport
}

func (self *HttpRequestClient) getTransport() *httpclient.Transport {
	return &httpclient.Transport {

		DisableKeepAlives: self.DisableKeepAlives,
		DisableCompression: self.DisableCompression,

		MaxIdleConnsPerHost: self.MaxIdleConnsPerHost,

		ConnectTimeout: self.ConnectTimeout,
		ResponseHeaderTimeout: self.ResponseHeaderTimeout,
		RequestTimeout: self.RequestTimeout,
		ReadWriteTimeout: self.ReadWriteTimeout,

		TLSClientConfig: &tls.Config{ InsecureSkipVerify: self.SkipSslVerify },
	}
}

func HttpPost(url string, values url.Values, headers map[string]string) ([]byte, error) { return NewDefaultHttpRequestClient().Post(url, values, headers) }

func HttpPostStr(url string, value string, headers map[string]string) ([]byte, error) { return NewDefaultHttpRequestClient().PostStr(url, value, headers) }

func HttpPostJson(url string, value interface{}, headers map[string]string) ([]byte, error) { return NewDefaultHttpRequestClient().PostJson(url, value, headers) }

func HttpPostBson(url string, bsonDoc interface{}, headers map[string]string) ([]byte, error) { return NewDefaultHttpRequestClient().PostBson(url, bsonDoc, headers) }

func HttpGetBson(url string, headers map[string]string) (bson.M, error) { return NewDefaultHttpRequestClient().GetBson(url, headers) }

