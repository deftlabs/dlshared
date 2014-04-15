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

type reflectionTestStruct struct {
	helloCalled bool
	channelTestCalled bool
}

func (self *reflectionTestStruct) Hello() { self.helloCalled = true }
func (self *reflectionTestStruct) Method0(value string) { }
func (self *reflectionTestStruct) Method1() error { return nil }
func (self *reflectionTestStruct) Method2(value1, value2 string) { }
func (self *reflectionTestStruct) Method3() (error, error) { return nil, nil }
func (self *reflectionTestStruct) ChannelTest(channel chan bool) { self.channelTestCalled = true }

func TestCallBoolChanParamNoReturnValueMethod(t *testing.T) {

	val := &reflectionTestStruct{}

	err, methodValue := GetMethodValueByName(val, "ChannelTest", 1, 0)
	if err != nil { t.Errorf("TestCallBoolChanParamNoReturnValueMethod failed with %v", err) }

	CallBoolChanParamNoReturnValueMethod(val, methodValue, make(chan bool))

	if !val.channelTestCalled { t.Errorf("TestCallBoolChanParamNoReturnValueMethod failed - ChannelTest method not called") }
}

func TestCallNoParamNoReturnValueMethod(t *testing.T) {

	val := &reflectionTestStruct{}

	err, methodValue := GetMethodValueByName(val, "Hello", 0, 0)
	if err != nil { t.Errorf("TestCallNoParamNoReturnValueMethod failed with %v", err) }

	CallNoParamNoReturnValueMethod(val, methodValue)

	if !val.helloCalled { t.Errorf("TestCallNoParamNoReturnValueMethodfailed - Hello method not called") }
}

func TestGetMethodValueByName(t *testing.T) {

	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Hello", 0, 0); err != nil { t.Errorf("TestGetMethodValueByName failed Hello 0 0 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Hello", 1, 1); err == nil { t.Errorf("TestGetMethodValueByName failed Hello 1 1") }

	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method0", 1, 0); err != nil { t.Errorf("TestGetMethodValueByName failed Method0 1 0 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method0", 0, 1); err == nil { t.Errorf("TestGetMethodValueByName failed Method0 0 1 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method0", 1, 1); err == nil { t.Errorf("TestGetMethodValueByName failed Method0 1 1 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method0", 0, 0); err == nil { t.Errorf("TestGetMethodValueByName failed Method0 0 0 - with %v", err) }

	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method1", 0, 1); err != nil { t.Errorf("TestGetMethodValueByName failed Method1 0 1 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method1", 1, 0); err == nil { t.Errorf("TestGetMethodValueByName failed Method1 1 0 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method1", 0, 0); err == nil { t.Errorf("TestGetMethodValueByName failed Method1 0 0 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method1", 1, 0); err == nil { t.Errorf("TestGetMethodValueByName failed Method1 1 0 - with %v", err) }

	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method2", 2, 0); err != nil { t.Errorf("TestGetMethodValueByName failed Method2 2 0 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method2", 2, 1); err == nil { t.Errorf("TestGetMethodValueByName failed Method2 2 1 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method2", 0, 0); err == nil { t.Errorf("TestGetMethodValueByName failed Method2 0 0 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method2", 0, 1); err == nil { t.Errorf("TestGetMethodValueByName failed Method2 0 1 - with %v", err) }

	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method3", 0, 2); err != nil { t.Errorf("TestGetMethodValueByName failed Method3 0 2 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method3", 1, 0); err == nil { t.Errorf("TestGetMethodValueByName failed Method3 1 0 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method3", 2, 0); err == nil { t.Errorf("TestGetMethodValueByName failed Method3 2 0 - with %v", err) }
	if err, _ := GetMethodValueByName(&reflectionTestStruct{}, "Method3", 0, 0); err == nil { t.Errorf("TestGetMethodValueByName failed Method3 0 0 - with %v", err) }
}

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
	if len(GetFunctionName(testFunc)) == 0 { t.Errorf("TestGetFunctionName failed - nested function name not returned") }
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
