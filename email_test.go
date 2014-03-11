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
	// "fmt"
	"testing"
)

const (
	EmailTestAccessKeyId = "HUH?"
	EmailTestSecretAccessKey = "WHAT?"
	EmailTestFrom = "Me <noreply@someplace.com>"
	EmailTestTo = "you@someplace.com"
)

// These tests are difficult to run in an open source repo because they require aws credentials.
func TestAwsTextEmail(t *testing.T) {
	/*
	logger := Logger { Prefix: "test", Appenders: []Appender{ LevelFilter(Debug, StdErrAppender()) } }

	emailDs := NewAwsEmailDs(EmailTestAccessKeyId, EmailTestSecretAccessKey, NewDefaultHttpRequestClient(), logger)

	response, err := emailDs.SendTextEmailToOneAddress(EmailTestFrom, EmailTestTo, "test subject", "this is a test text body")

	if err != nil {
		t.Errorf("TestAwsTextEmail is broken: %v")
		return
	}

	fmt.Println("message id:", response.(*AwsEmailResponse).MessageId)
	fmt.Println("request id:", response.(*AwsEmailResponse).RequestId)
	fmt.Println("status code:", response.(*AwsEmailResponse).HttpStatusCode)
	*/
}

func TestAwsHtmlEmail(t *testing.T) {
	/*
	bodyHtml := `<table width="100%" border="0" cellspacing="0" cellpadding="0">
	<tr><td><a href="http://deftlabs.com">Test Link</a></td></tr>
	</table>
	`
	logger := Logger { Prefix: "test", Appenders: []Appender{ LevelFilter(Debug, StdErrAppender()) } }

	emailDs := NewAwsEmailDs(EmailTestAccessKeyId, EmailTestSecretAccessKey, NewDefaultHttpRequestClient(), logger)

	response, err := emailDs.SendHtmlEmailToOneAddress(EmailTestFrom, EmailTestTo, "test subject", bodyHtml, "Test Link: http://deftlabs.com")

	if err != nil {
		t.Errorf("TestAwsHtmlEmail is broken: %v")
		return
	}

	fmt.Println("message id:", response.(*AwsEmailResponse).MessageId)
	fmt.Println("request id:", response.(*AwsEmailResponse).RequestId)
	fmt.Println("status code:", response.(*AwsEmailResponse).HttpStatusCode)
	*/
}

