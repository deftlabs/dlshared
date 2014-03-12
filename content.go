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
	"fmt"
	"github.com/daviddengcn/go-ljson-conf"
)

// The content service is a very simple content lookup based on passed locale. ContentSvc is stored
// in a json file. The path/structure of your content is basically up to you. However, to support locales,
// the following structure is required:
//
// 	{
// 		"something": {
// 			"else": {
//				"en_US": "Hello",
//				"de_DE": "Hallo",
// 				"pt_BR": "Olá"
//			}
// 		}
//	}
//
// With this structure, the path you would pass is: "something.else". The locale param would be "end_US" or
// another locale. You can store the locale keys in any format you like (e.g., you could use "en-US" or just
// "English", or whatever you like).
type ContentSvc struct {
	data *ljconf.Conf
}

func NewContentSvc(fileName string) (*ContentSvc, error) {
	content := &ContentSvc{}

	var err error
	if content.data, err = ljconf.Load(fileName); err != nil {
		return nil, err
	}

	return content, nil
}

// Lookup content. If the content is not found, the default is returned. This method
// requires the full path. For the example above, you would have to pass the path of:
// something.else.en_US to pull some content.
func (self *ContentSvc) Lookup(path, def string) string { return self.data.String(path, def) }

// Lookup content. If the content is not found, the default locale string is returned. If
// the default locale is missing too, the default string is returned. See above for information about
// content structure.
func (self *ContentSvc) LookupWithLocale(path, locale, defaultLocale, def string) string {
	val := self.data.String(fmt.Sprintf("%s.%s", path, locale), nadaStr)
	if len(val) > 0 { return val }

	if len(defaultLocale) == 0 { return def }

	return self.data.String(fmt.Sprintf("%s.%s", path, defaultLocale), def)
}

