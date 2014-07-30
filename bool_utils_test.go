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

func TestStrIsTrue(t *testing.T) {

	if !StrIsTrue("true") { t.Errorf("TestStrIsTrue is broken - expecting true - string param: true") }

	if !StrIsTrue("yes") {
		t.Errorf("TestStrIsTrue is broken - expecting true - string param: yes")
	}

	if !StrIsTrue("y") {
		t.Errorf("TestStrIsTrue is broken - expecting true - string param: y")
	}

	if !StrIsTrue("1") {
		t.Errorf("TestStrIsTrue is broken - expecting true - string param: 1")
	}

	if StrIsTrue("g") {
		t.Errorf("TestStrIsTrue is broken - expecting false - string param: g")
	}

	if StrIsTrue("false") {
		t.Errorf("TestStrIsTrue is broken - expecting false - string param: false")
	}

	if StrIsTrue("no") {
		t.Errorf("TestStrIsTrue is broken - expecting false - string param: no")
	}

	if StrIsTrue("n") {
		t.Errorf("TestStrIsTrue is broken - expecting false - string param: n")
	}

	if StrIsTrue("0") {
		t.Errorf("TestStrIsTrue is broken - expecting false - string param: 0")
	}
}

func TestStrIsFalse(t *testing.T) {

	if StrIsFalse("true") {
		t.Errorf("TestStrIsFalse is broken - expecting false - string param: true")
	}

	if StrIsFalse("yes") {
		t.Errorf("TestStrIsFalse is broken - expecting false - string param: yes")
	}

	if StrIsFalse("y") {
		t.Errorf("TestStrIsFalse is broken - expecting false - string param: y")
	}

	if StrIsFalse("1") {
		t.Errorf("TestStrIsFalse is broken - expecting false - string param: 1")
	}

	if StrIsFalse("g") {
		t.Errorf("TestStrIsFalse is broken - expecting false - string param: g")
	}

	if !StrIsFalse("false") {
		t.Errorf("TestStrIsFalse is broken - expecting true - string param: false")
	}

	if !StrIsFalse("no") {
		t.Errorf("TestStrIsFalse is broken - expecting true - string param: no")
	}

	if !StrIsFalse("n") {
		t.Errorf("TestStrIsFalse is broken - expecting true - string param: n")
	}

	if !StrIsFalse("0") {
		t.Errorf("TestStrIsFalse is broken - expecting true - string param: 0")
	}
}

