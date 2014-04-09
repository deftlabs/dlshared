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
 *
 *
 * Some of the initial SES code came from: https://github.com/sourcegraph/go-ses/blob/master/ses.go
 * Copyright 2013 SourceGraph, Inc.
 * Copyright 2011-2013 Numrotron Inc.
 * Use of this source code is governed by an MIT-style license
 */

package dlshared

import (
	"fmt"
	"time"
	"errors"
	"net/url"
	"net/http"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/xml"
	"encoding/base64"
	"labix.org/v2/mgo/bson"
)

// -----------------------------------------------------------------------------

// The email ds interface
type EmailDs interface {
	SendTextEmailToOneAddress(from, to, subject, body string) (interface{}, error)
	SendHtmlEmailToOneAddress(from, to, subject, bodyHtml, bodyText string) (interface{}, error)
}

// -----------------------------------------------------------------------------

// The AWS SES email ds.
const (
	AwsSesEndpoint = "https://email.us-east-1.amazonaws.com"
)

type AwsEmailDs struct {
	Logger
	accessKeyId string
	secretAccessKey string
	httpClient *HttpRequestClient
}

// Create a new aws email ds.
func NewAwsEmailDs(awsAccessKeyId, awsSecretKey string, httpClient  *HttpRequestClient, logger Logger) *AwsEmailDs {
	return &AwsEmailDs{ accessKeyId: awsAccessKeyId, secretAccessKey: awsSecretKey, httpClient: httpClient, Logger: logger }
}

// Send a text email to one address.
func (self *AwsEmailDs) SendTextEmailToOneAddress(from, to, subject, body string) (interface{}, error) {
	data := make(url.Values)
	data.Add("Action", "SendEmail")
	data.Add("Source", from)
	data.Add("Destination.ToAddresses.member.1", to)
	data.Add("Message.Subject.Data", subject)
	data.Add("Message.Body.Text.Data", body)
	data.Add("AWSAccessKeyId", self.accessKeyId)
	return self.postEmailToSes(data)
}

// Send an html email to one address.
func (self *AwsEmailDs) SendHtmlEmailToOneAddress(from, to, subject, bodyHtml, bodyText string) (interface{}, error) {
	data := make(url.Values)
	data.Add("Action", "SendEmail")
	data.Add("Source", from)
	data.Add("Destination.ToAddresses.member.1", to)
	data.Add("Message.Subject.Data", subject)
	data.Add("Message.Body.Text.Data", bodyText)
	data.Add("Message.Body.Html.Data", bodyHtml)
	data.Add("AWSAccessKeyId", self.accessKeyId)
	return self.postEmailToSes(data)
}

type AwsEmailResponse struct {
	MessageId string `xml:"SendEmailResult>MessageId"`
	RequestId string `xml:"ResponseMetadata>RequestId"`
	HttpStatusCode int
}

// Post the email to SES. The raw response is:
// 	<SendEmailResponse xmlns="http://ses.amazonaws.com/doc/2010-12-01/">
// 		<SendEmailResult>
// 			<MessageId>00000144b1cc0395-e655ff4d-6c5f-4aed-919b-b4f0472f4491-000000</MessageId>
// 		</SendEmailResult>
// 		<ResponseMetadata>
// 			<RequestId>44c89891-a933-11e3-bec7-3f7c55b51d3e</RequestId>
// 		</ResponseMetadata>
// </SendEmailResponse>
//
// Documentation for the API is @: http://docs.aws.amazon.com/ses/2010-12-01/APIReference/Welcome.html
//
func (self *AwsEmailDs) postEmailToSes(data url.Values) (interface{}, error) {

	headers := make(map[string]string)

	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 -0700")

	headers["Date"] = date

	h := hmac.New(sha256.New, []uint8(self.secretAccessKey))
	h.Write([]uint8(date))
	headers["X-Amzn-Authorization"] = fmt.Sprintf("AWS3-HTTPS AWSAccessKeyId=%s, Algorithm=HmacSHA256, Signature=%s", self.accessKeyId, base64.StdEncoding.EncodeToString(h.Sum(nil)))

	httpStatusCode, response, err := self.httpClient.Post(AwsSesEndpoint, data, headers)
	if err != nil { return nil, err }

	awsResponse := &AwsEmailResponse{ HttpStatusCode: httpStatusCode }

	// Parse the response.
	err = xml.Unmarshal(response, &awsResponse)
	if err != nil {
		return awsResponse, errors.New(fmt.Sprintf("Unable to parse Aws SES response - raw: %s - error: %v", string(response), err))
	}

	// Create an error if the status is not ok/200.
	if httpStatusCode != http.StatusOK {
		return awsResponse, errors.New(fmt.Sprintf("Non-200 http error code returned from Aws SES post - status code: %d - raw: %s", httpStatusCode, string(response)))
	}

	return awsResponse, nil
}

// -----------------------------------------------------------------------------

// The email service

type EmailSvc interface {
	SendHtmlEmailToOneAddress(from, to, subject, htmlTemplateFileName, textTemplateFileName string, params map[string]interface{}) error
	Start(kernel *Kernel) error
	Stop(kernel *Kernel) error
}

type EmailDoc struct {

	Id bson.ObjectId `bson:"_id"`

	HtmlTemplateName string `bson:"htmlTemplateName,omitempty"`
	TextTemplateName string `bson:"textTemplateName,omitempty"`

	SendMethod string `bson:"sendMethod"`

	EmailType string `bson:"type"`

	MessageId string `bson:"messageId,omitempty"`
	RequestId string `bson:"requestId,omitempty"`
	HttpStatusCode int `bson:"httpStatusCode,omitempty"`

	ToAddrs []string `bson:"toAddrs"`
	FromAddr string `bson:"fromAddr"`
	Sent *time.Time `bson:"sent"`

	Subject []byte `bson:"subject"`
	BodyHtml []byte `bson:"bodyHtml"`
	BodyText []byte `bson:"bodyText"`

	Error string `bson:"error,omitempty"`
}

// Create a new email service. Set cappedCollectionSizeInBytes to less than one to create a permanent collection
// (i.e., otherwise it creates a capped collection). Currently, the email service only supports sending via
// AWS SES.
func NewAwsEmailSvc(dbComponentName,
					templateComponentName,
					dbName, collectionName,
					awsAccessKeyId,
					awsSecretKey string,
					cappedCollectionSizeInBytes int) EmailSvc {
	return &AwsEmailSvc{
		Logger: Logger{},
		MongoDataSource: MongoDataSource{ DbName: dbName, CollectionName: collectionName },
		dbComponentName: dbComponentName,
		templateComponentName: templateComponentName,
		cappedCollectionSizeInBytes: cappedCollectionSizeInBytes,
		awsEmailDs: NewAwsEmailDs(awsAccessKeyId, awsSecretKey, NewDefaultHttpRequestClient(), Logger{}),
	}
}

type AwsEmailSvc struct {
	Logger
	MongoDataSource
	dbComponentName string
	templateComponentName string
	cappedCollectionSizeInBytes int
	awsEmailDs *AwsEmailDs
	templateSvc *TemplateSvc
}

func (self *AwsEmailSvc) storeEmail(from, to, subject, htmlTemplateFileName, textTemplateFileName, bodyHtml, bodyText string, response interface{}, err error) {
	now := time.Now()

	compressedBodyHtml, err := CompressBytes([]byte(bodyHtml))
	if err != nil { self.Logf(Warn, "Unable to compress html body: %v", err) }

	compressedBodyTxt, err := CompressBytes([]byte(bodyText))
	if err != nil { self.Logf(Warn, "Unable to compress text body: %v", err) }

	compressedSubject, err := CompressBytes([]byte(subject))
	if err != nil { self.Logf(Warn, "Unable to compress subject: %v", err) }

	doc := &EmailDoc{	Id: bson.NewObjectId(),
						HtmlTemplateName: htmlTemplateFileName,
						TextTemplateName: textTemplateFileName,
						SendMethod: "AWS-SES",
						EmailType: "html",
						ToAddrs: []string{ to },
						FromAddr: from,
						Sent: &now,
						Subject: compressedSubject,
						BodyHtml: compressedBodyHtml,
						BodyText: compressedBodyTxt,
	}

	if err != nil { doc.Error = err.Error() }

	if response != nil {
		r := response.(*AwsEmailResponse)
		doc.MessageId = r.MessageId
		doc.RequestId = r.RequestId
		doc.HttpStatusCode = r.HttpStatusCode
	}

	if err := self.Insert(doc); err != nil { self.Logf(Warn, "Unable to store email doc: %v", err) }
}

// Send an html email.
func (self *AwsEmailSvc) SendHtmlEmailToOneAddress(from, to, subject, htmlTemplateFileName, textTemplateFileName string, params map[string]interface{}) error {

	// Render the templates.
	bodyHtml, bodyText, err := self.templateSvc.RenderHtmlAndText(htmlTemplateFileName, textTemplateFileName, params)

	// Send the email to the user.
	response, err := self.awsEmailDs.SendHtmlEmailToOneAddress(from, to, subject, bodyHtml, bodyText)
	if err != nil {
		self.storeEmail(from, to, subject, htmlTemplateFileName, textTemplateFileName, bodyHtml, bodyText, response, err)
		return err
	}

	self.storeEmail(from, to, subject, htmlTemplateFileName, textTemplateFileName, bodyHtml, bodyText, response, nil)

	return nil
}

func (self *AwsEmailSvc) Start(kernel *Kernel) error {

	self.Logger = kernel.Logger
	self.awsEmailDs.Logger = self.Logger

	self.Mongo = kernel.GetComponent(self.dbComponentName).(*Mongo)
	self.templateSvc = kernel.GetComponent(self.templateComponentName).(*TemplateSvc)

	// Create the capped collection to store the emails.
	if self.cappedCollectionSizeInBytes > 0 {
		if err := self.CreateCappedCollection(self.cappedCollectionSizeInBytes); err != nil { return err }
	}

	if err := self.EnsureIndex([]string{ "sent" }); err != nil { return err }
	if err := self.EnsureIndex([]string{ "toAddrs" }); err != nil { return err }

	if err := self.EnsureIndex([]string{ "toAddrs", "sent"  }); err != nil { return err }

	if err := self.EnsureSparseIndex([]string{ "messageId" }); err != nil { return err }
	if err := self.EnsureSparseIndex([]string{ "requestId" }); err != nil { return err }

	return nil
}

func (self *AwsEmailSvc) Stop(kernel *Kernel) error { return nil }

