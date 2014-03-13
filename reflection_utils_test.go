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


	if GetFunctionName(testFunction1) != "testFunction1" {
		t.Errorf("TestGetFunctionName failed - expected: testFunction1 - received: %s", GetFunctionName(testFunction1))
	}

	if GetFunctionName(testFunction2) != "testFunction2" {
		t.Errorf("TestGetFunctionName failed - expected: testFunction2 - received: %s", GetFunctionName(testFunction2))
	}

	bar := Bar{}
	if GetFunctionName(bar.test) != "Bar.test" {
		t.Errorf("TestGetFunctionName failed - expected: Bar.test - received: %s", GetFunctionName(bar.test))
	}

	bar1 := &Bar{}
	if GetFunctionName(bar1.test) != "Bar.test" {
		t.Errorf("TestGetFunctionName failed - expected: Bar.test - received: %s", GetFunctionName(bar1.test))
	}

	foo := Foo{}
	if GetFunctionName(foo.test) != "Foo.test" {
		t.Errorf("TestGetFunctionName failed - expected: Foo.test - received: %s", GetFunctionName(foo.test))
	}

	foo1 := Foo{}
	if GetFunctionName(foo1.test) != "Foo.test" {
		t.Errorf("TestGetFunctionName failed - expected: Foo.test - received: %s", GetFunctionName(foo1.test))
	}

	if GetFunctionName(foo1.test1) != "Foo.test1" {
		t.Errorf("TestGetFunctionName failed - expected: Foo.test1 - received: %s", GetFunctionName(foo.test1))
	}

	testFunc := func() { }
	if GetFunctionName(testFunc) != "func·010" {
		t.Errorf("TestGetFunctionName failed - expected: func·010 - received: %s", GetFunctionName(testFunc))
	}
}

// Note: if you add more tests, you must update the last test (dlshared.func·010).
type Foo struct {

}

func (self *Foo) test() {

}

func (self *Foo) test1(withParam string) {

}

type Bar struct {

}

func (self Bar) test() {

}

func testFunction2(hello string) {

}

func testFunction1() {

}
