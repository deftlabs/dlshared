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
	"errors"
	"testing"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type structPointerTest struct {
	name string
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

	if nil == err {
		t.Errorf("TestErrors is broken - nil == error")
	}

	if err != err {
		t.Errorf("TestErrors is broken - error == error")
	}
}

// Confirm the way slices behave.
func TestSlices(t *testing.T) {

	var slice []byte

	if slice != nil {
		t.Errorf("TestSlice is broken - slice is not nil")
	}

	if len(slice) != 0 {
		t.Errorf("TestSlice is broken - slice length is not zero")
	}
}

// Confirm the way data types behave.
func TestDataTypes(t *testing.T) {

	val := fmt.Sprintf("%t", true)
	if val != "true" {
		t.Errorf("TestDataTypes is broken - true != true")
	}

	val = fmt.Sprintf("%t", false)
	if val != "false" {
		t.Errorf("TestDataTypes is broken - false != false")
	}
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

	if _, found := mapTest["four"]; found {
		t.Errorf("TestMaps is broken - something found that does not exist")
	}

	if _, found := mapTest["one"]; !found {
		t.Errorf("TestMaps is broken - something not found that should be found")
	}

	// A missing key, should not panic.
	emptyStr := mapTest["four"]

	if len(emptyStr) != 0 {
		t.Errorf("TestMaps is broken - the empty value has something")
	}

	var testMap map[string]string

	// Make sure a var map is nil
	if testMap != nil {
		t.Errorf("TestMaps is broken - uninitiated map should be nil")
	}
}

