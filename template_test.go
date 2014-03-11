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

import "testing"

type testTemplateEntry struct {
	Link string
	Text string
}

func initTemplateTestParams() map[string]interface{} {
	params := make(map[string]interface{})

	params["Test"] = "Hello"

	var entries []*testTemplateEntry

    for i := 0; i < 5; i++ {
		entries = append(entries, &testTemplateEntry{ Link: "http://deftlabs.com", Text: "Deft Labs" })
	}

	params["Entries"] = entries

	return params
}

func TestRenderHtml(t *testing.T) {

	template := NewTemplate("test/templates/")

	response, err := template.RenderHtml("test.html", initTemplateTestParams())

	if err != nil {
		t.Errorf("TestRenderHtml is broken: %v", err)
		return
	}

	if len(response) != 390 {
		t.Errorf("TestRenderHtml is broken - expected response length: 390 - actual: %d - response: %s", len(response), response)
	}
}

func TestRenderText(t *testing.T) {

	template := NewTemplate("test/templates/")

	response, err := template.RenderText("test.txt", initTemplateTestParams())

	if err != nil {
		t.Errorf("TestRenderText is broken: %v", err)
		return
	}

	if len(response) != 285 {
		t.Errorf("TestRenderText is broken - expected response length: 285 - actual: %d - response: %s", len(response), response)
	}
}

