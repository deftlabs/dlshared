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

func TestContentSvc(t *testing.T) {

	contentSvc, err := NewContentSvc("test/content_test.json")

	if err != nil {
		t.Errorf("TestContentSvc is broken: %v", err)
		return
	}

	// Do a simple lookup
	val := contentSvc.Lookup("greetings.generic.hello.en_US", "")

	if len(val) == 0 {
		t.Errorf("TestContentSvc is broken - path not found: greetings.generic.hello.en_US")
	}

	if val != "Hello" {
		t.Errorf("TestContentSvc is broken - path: greetings.generic.hello.en_US - returned %s - expecting: Hello", val)
	}

	// Do a locale lookup
	val = contentSvc.LookupWithLocale("greetings.generic.hello", "pt_BR", "", "MISSING")

	if val != "Olá" {
		t.Errorf("TestContentSvc is broken - path: greetings.generic.hello - locale: pt_BR - returned %s - expecting: Olá", val)
	}

	// Do a missing locale lookup that falls back to the default locale.
	val = contentSvc.LookupWithLocale("greetings.generic.hello", "ru_RU", "en_US", "MISSING")

	if val != "Hello" {
		t.Errorf("TestContentSvc is broken - path: greetings.generic.hello - locale: ru_RU - default locale: en_US - returned %s - expecting: Hello", val)
	}

	// Do a lookup with a missing locale and default locale.
	val = contentSvc.LookupWithLocale("greetings.generic.hello", "ru_RU", "zh_CN", "MISSING")

	if val != "MISSING" {
		t.Errorf("TestContentSvc is broken - path: greetings.generic.hello - locale: ru_RU - default locale: ah_CN - returned %s - expecting: MISSING", val)
	}
}

