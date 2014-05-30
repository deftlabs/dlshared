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
	"sync"
	"time"
	"errors"
	"testing"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

func TestTime(t *testing.T) {

	now := time.Now()

	// This should not panic or return true.
	var toCheck time.Time
	if now.Before(toCheck) { t.Errorf("TestTime is broken - nothing is before Now") }
	if !now.After(toCheck) { t.Errorf("TestTime is broken - nothing is after Now") }
}

func TestChannels1(t *testing.T) {

	channel := make(chan bool, 1)

	channel <- true

	value, ok := <- channel

	if !value { t.Errorf("TestChannels1 is broken - value should be true") }

	if !ok { t.Errorf("TestChannels1 is broken - ok is false") }

	var waitGroup sync.WaitGroup

	// We are going to block here.
	go func() {
		waitGroup.Add(1)
		value, ok = <- channel
		if ok { t.Errorf("TestChannels1 is broken - ok should be false") }
		waitGroup.Done()
	}()

	close(channel)

	waitGroup.Wait()

	structChannel := make(chan *bson.M, 1)
	close(structChannel)
	result := <- structChannel

	if result != nil { t.Errorf("TestChannels1 is broken - result should be nil") }

}

func TestChannels(t *testing.T) {
	test := &testPanicStruct{}
	channel := make(chan bool)
	close(channel)
	test.call2(channel)
	if test.called0 != 1 { t.Errorf("TestChannels is broken - called0 should be 1 but, is %d", test.called0) }

	test.called0 = 0

	channel1 := make(chan bool)

	go test.call3(channel1)

	close(channel1)

	time.Sleep(400*time.Millisecond)

	if test.called0 != 0 { t.Errorf("TestChannels is broken - called0 should be 0 but, is %d", test.called0) }
}

type testPanicStruct struct {
	called0 int
	called1 int
}

func (self *testPanicStruct) call3(channel chan bool) {
	defer func() { if r := recover(); r != nil { self.called0++ } }()
	for v := range channel { if v { } }
}

func (self *testPanicStruct) call2(channel chan bool) {
	defer func() { if r := recover(); r != nil { self.called0++ } }()
	channel <- true
}

func (self *testPanicStruct) call0() {
	defer func() { self.called0++ }()
	defer func() {
		if r := recover(); r != nil { }
		self.called0++
	}()
	panic("This is a panic: 0")
}

func (self *testPanicStruct) call1() {
	defer func() { self.called0++ }()
	panic("This is a panic: 1")
}

func TestPanic(t *testing.T) {

	test := &testPanicStruct{}

	test.call0()

	if test.called0 != 2 { t.Errorf("TestPanic is broken - called0 should be 2 but, is %d", test.called0) }

	defer func() {
		recover()
		if test.called1 != 0 { t.Errorf("TestPanic is broken - called1 should be zero but, is %d", test.called1) }
	}()

	test.call1()
}

func TestDivisionEquals(t *testing.T) {

	value := 10
	value /= 1
	if value != 10 { t.Errorf("TestDivisionEquals is broken - value off") }

	value = 20
	value /= 2
	if value != 10 { t.Errorf("TestDivisionEquals is broken - value off") }
}

type structPointerTest struct { name string }

func TestClosures(t *testing.T) {

	counter := testClosureMethod()

	for i := 0; i < 9; i++ { counter() }

	if counter() != 10 { t.Errorf("TestClosures is broken - count off") }
}

func testClosureMethod() func() int {
	count := 0
	return func() int { count++; return count }
}

func TestDeleteMissingKeyInMap(t *testing.T) { delete(make(map[string]string, 0), "missing") }

func TestDeletingKeysInMapWhileInRange(t *testing.T) {

	test := map[string]string { "test0": "value", "test1": "value", "test2": "value" }

	count := 0

	for key, _ := range test { delete(test, key); count++ }

	if count != 3 { t.Errorf("TestDeletingKeysInMapWhileInRange is broken - count off") }

	if len(test) != 0 { t.Errorf("TestDeletingKeysInMapWhileInRange is broken - length not zero") }

	for _, _ = range test {  }
}

func TestNilErr(t *testing.T) {
	if nil == mgo.ErrNotFound { t.Errorf("TestNilErr is broken - mgo not found matches nil") }
	if mgo.IsDup(nil) { t.Errorf("TestNilErr is broken - is dup matches nil") }
}

// Confirm the way structs/pointers works.
func TestStructs(t *testing.T) {

	id1 := structPointerTest{ name: "test" }
	id2 := structPointerTest{ name: "test" }

	if id1 != id2 { t.Errorf("TestStructs is broken - no match") }

	p1 := &id1
	p2 := &id2

	if p1 == p2 { t.Errorf("TestStructs is broken - pointer match") }
	if *p1 != *p2 { t.Errorf("TestStructs is broken - deferenced pointer - no match") }

	id1 = structPointerTest{ name: "test0" }
	id2 = structPointerTest{ name: "test1" }

	if id1 == id2 { t.Errorf("TestStructs is broken - match") }

	p1 = &id1
	p2 = &id2

	if p1 == p2 { t.Errorf("TestStructs is broken - pointer match") }
	if *p1 == *p2 { t.Errorf("TestStructs is broken - deferenced pointer match") }
}

func TestObjectId(t *testing.T) {

	id1 := bson.ObjectIdHex("532b19b784a8f7f139f3e338")
	id2 := bson.ObjectIdHex("532b19b784a8f7f139f3e338")

	if id1 != id2 { t.Errorf("TestObjectId is broken - no match") }

	p1 := &id1
	p2 := &id2

	if p1 == p2 { t.Errorf("TestObjectId is broken - pointer match") }
	if *p1 != *p2 { t.Errorf("TestObjectId is broken - deferenced pointer - no match") }

	id1 = bson.ObjectIdHex("532b19b784a8f7f139f3e338")
	id2 = bson.ObjectIdHex("532b19b884a8f7f139f3e339")

	if id1 == id2 { t.Errorf("TestStructs is broken - match") }

	p1 = &id1
	p2 = &id2

	if p1 == p2 { t.Errorf("TestStructs is broken - pointer match") }
	if *p1 == *p2 { t.Errorf("TestStructs is broken - deferenced pointer match") }
}

// Confirm the way errors behave.
func TestErrors(t *testing.T) {

	err := errors.New("test")

	if nil == err { t.Errorf("TestErrors is broken - nil == error") }

	if err != err { t.Errorf("TestErrors is broken - error == error") }
}

// Confirm the way slices behave.
func TestSlices(t *testing.T) {

	var slice []byte

	if slice != nil { t.Errorf("TestSlice is broken - slice is not nil") }

	if len(slice) != 0 { t.Errorf("TestSlice is broken - slice length is not zero") }
}

// Confirm the way data types behave.
func TestDataTypes(t *testing.T) {

	val := fmt.Sprintf("%t", true)
	if val != "true" { t.Errorf("TestDataTypes is broken - true != true") }

	val = fmt.Sprintf("%t", false)
	if val != "false" { t.Errorf("TestDataTypes is broken - false != false") }
}

// Confirm the way range behaves.
func TestRange(t *testing.T) {

	var test map[string]string

	// This should not panic
	for _, _ = range test { }

	// Just to be clear again
	test = nil

	// This should not panic
	for _, _ = range test { }
}

// Confirm the way maps behave.
func TestMaps(t *testing.T) {

	mapTest := make(map[string]string)
	mapTest["one"] = "one"
	mapTest["two"] = "two"
	mapTest["three"] = "three"

	if _, found := mapTest["four"]; found { t.Errorf("TestMaps is broken - something found that does not exist") }

	if _, found := mapTest["one"]; !found { t.Errorf("TestMaps is broken - something not found that should be found") }

	// A missing key, should not panic.
	emptyStr := mapTest["four"]

	if len(emptyStr) != 0 { t.Errorf("TestMaps is broken - the empty value has something") }

	var testMap map[string]string

	// Make sure a var map is nil
	if testMap != nil { t.Errorf("TestMaps is broken - uninitiated map should be nil") }

	// This should not panic.
	for _, v := range testMap { if v == "" { } }
}

