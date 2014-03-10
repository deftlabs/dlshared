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

func TestGetFunctionName(t *testing.T) {


	if GetFunctionName(testFunction1) != "dlshared.testFunction1" {
		t.Errorf("TestGetFunctionName failed - expected: dlshared.testFunction1 - received: %s", GetFunctionName(testFunction1))
	}

	if GetFunctionName(testFunction2) != "dlshared.testFunction2" {
		t.Errorf("TestGetFunctionName failed - expected: dlshared.testFunction2 - received: %s", GetFunctionName(testFunction2))
	}

	bar := Bar{}
	if GetFunctionName(bar.test) != "dlshared.Bar.test" {
		t.Errorf("TestGetFunctionName failed - expected: dlshared.Bar.test - received: %s", GetFunctionName(bar.test))
	}

	bar1 := &Bar{}
	if GetFunctionName(bar1.test) != "dlshared.Bar.test" {
		t.Errorf("TestGetFunctionName failed - expected: dlshared.Bar.test - received: %s", GetFunctionName(bar1.test))
	}

	foo := Foo{}
	if GetFunctionName(foo.test) != "dlshared.Foo.test" {
		t.Errorf("TestGetFunctionName failed - expected: dlshared.Foo.test - received: %s", GetFunctionName(foo.test))
	}

	foo1 := Foo{}
	if GetFunctionName(foo1.test) != "dlshared.Foo.test" {
		t.Errorf("TestGetFunctionName failed - expected: dlshared.Foo.test - received: %s", GetFunctionName(foo1.test))
	}

	testFunc := func() { }
	if GetFunctionName(testFunc) != "dlshared.func·013" {
		t.Errorf("TestGetFunctionName failed - expected: dlshared.func·013 - received: %s", GetFunctionName(testFunc))
	}
}

// Note: if you add more tests, you must update the last test (dlshared.func·013).
type Foo struct {

}

func (self *Foo) test() {

}

type Bar struct {

}

func (self Bar) test() {

}

func testFunction2(hello string) {

}

func testFunction1() {

}
