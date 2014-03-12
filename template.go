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
	texttemplate "text/template"
	htmltemplate "html/template"
)

type TemplateSvc struct {
	baseTemplateSvcDir string
}

// Create a new template service struct. This simplifies
// accessing and rendering templates. End your base template directory with a slash (/).
func NewTemplateSvc(baseTemplateSvcDir string) *TemplateSvc {
	return &TemplateSvc{ baseTemplateSvcDir: baseTemplateSvcDir }
}

// The html template output is the first param and the text is the second param.
func (self *TemplateSvc) RenderHtmlAndText(htmlTemplateSvcFileName, textTemplateSvcFileName string, params interface{}) (bodyHtml, bodyText string, err error) {

	bodyHtml, err = self.RenderHtml(htmlTemplateSvcFileName, params)
	if err != nil { return }

	bodyText, err = self.RenderText(textTemplateSvcFileName, params)
	if err != nil { return }

	return
}

// The template file name should not start with a slash and be under the baseTemplateSvcDir
func (self *TemplateSvc) RenderHtml(templateFileName string, params interface{}) (string, error) {

	template, err := htmltemplate.ParseFiles(self.baseTemplateSvcDir + templateFileName)
	if err != nil { return nadaStr, err }

	var out bytes.Buffer
	if err := template.ExecuteTemplate(&out, templateFileName, params); err != nil { return nadaStr, err }
	return out.String(), nil
}

// The template file name should not start with a slash and be under the baseTemplateSvcDir
func (self *TemplateSvc) RenderText(templateFileName string, params interface{}) (string, error) {

	template, err := texttemplate.ParseFiles(self.baseTemplateSvcDir + templateFileName)
	if err != nil { return nadaStr, err }

	var out bytes.Buffer
	if err := template.ExecuteTemplate(&out, templateFileName, params); err != nil { return nadaStr, err }
	return out.String(), nil
}

