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
	"bytes"
	"testing"
	"net/http"
	"encoding/json"
	"net/http/httptest"
)

func TestNewHttpContextBasic(t *testing.T) {

	response := httptest.NewRecorder()
	request := &http.Request{ Method: "GET" }

	ctx := NewHttpContext(response, request)

	if !ctx.ParamsAreValid() { t.Errorf("TestNewHttpContextBasic is broken - empty params are not valid") }

	if len(ctx.ErrorCodes) != 0 { t.Errorf("TestNewHttpContextBasic is broken - there are error codes") }
}

func TestNewHttpContextJsonPostParams(t *testing.T) {
	response := httptest.NewRecorder()

	doc := make(map[string]interface{})
	doc["testObjectId0"] = "52e29b18eee7d580e9bb1544"
	doc["testObjectId1"] = ""
	doc["testInt0"] = 100
	doc["testBool0"] = true
	doc["testFloat0"] = 99.9999999999
	doc["testString0"] = "hello!"

	nestedDoc := make(map[string]interface{})
	nestedDoc["testStr0"] = "ok"
	nestedDoc["testStr1"] = "what"
	doc["testJson0"] = nestedDoc

	nestedArray := make([]string, 0, 0)
	nestedArray = append(nestedArray, "one")
	nestedArray = append(nestedArray, "two")
	doc["testJsonArray0"] = nestedArray

	data, err := json.Marshal(doc)
	if err != nil { t.Errorf("TestNewHttpContextJsonPostParams is broken - cannot encode json: %v", err); return }

	request, err := http.NewRequest(
		"POST",
		"/foo",
		bytes.NewBuffer(data),
	)

	if err != nil { t.Errorf("TestNewHttpContextJsonPostParams is broken - NewRequest failed"); return }

	ctx := NewHttpContext(response, request)

	defineParams(ctx, HttpParamJsonPost)

	validateParamOutput("jsonpost", ctx, t)
}

func TestNewHttpContextPostParams(t *testing.T) {
	response := httptest.NewRecorder()

	request, err := http.NewRequest(
		"POST",
		"/foo",
		bytes.NewBuffer([]byte("testInt0=100&testInt1=&testBool0=true&testBoo1=&testFloat0=99.9999999999&testFloat1=&testString0=hello%21&testString1=&testObjectId0=52e29b18eee7d580e9bb1544&testObjectId1=")),
	)

	if err != nil {
		t.Errorf("TestNewHttpContextPostParams is broken - NewRequest failed")
	}

	// Set the content type.
	request.Header["Content-Type"] = []string{ "application/x-www-form-urlencoded" }

	if err := request.ParseForm(); err != nil {
		t.Errorf("TestNewHttpContextPostParams is broken - ParseForm failed")
	}

	ctx := NewHttpContext(response, request)

	defineParams(ctx, HttpParamPost)

	validateParamOutput("post", ctx, t)
}

func TestNewHttpContextQueryParams(t *testing.T) {

	response := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/foo?testInt0=100&testBool0=true&testBoo1=&testFloat0=99.9999999999&testString0=hello%21&testObjectId0=52e29b18eee7d580e9bb1544", nil)
	if err != nil {
		t.Errorf("TestNewHttpContextQueryParams is broken - NewRequest failed")
	}

	ctx := NewHttpContext(response, request)

	defineParams(ctx, HttpParamQuery)

	validateParamOutput("query", ctx, t)
}

func TestNewHttpContextHeaderParams(t *testing.T) {

	response := httptest.NewRecorder()
	request := &http.Request{ Method: "GET", Header: make(map[string][]string) }

	// Add some values to the headers.
	request.Header.Set("testInt0", "100")
	request.Header.Set("testInt1", "")

	request.Header.Set("testBool0", "true")
	request.Header.Set("testBool1", "")

	request.Header.Set("testFloat0", "99.9999999999")
	request.Header.Set("testFloat1", "")

	request.Header.Set("testString0", "hello!")
	request.Header.Set("testString1", "")

	request.Header.Set("testObjectId0", "52e29b18eee7d580e9bb1544")
	request.Header.Set("testObjectId1", "")

	ctx := NewHttpContext(response, request)

	defineParams(ctx, HttpParamHeader)

	validateParamOutput("header", ctx, t)
}

func defineParams(ctx *HttpContext, httpParamType HttpParamType) {
	// Define some params
	ctx.DefineIntParam("testInt0", "invalid_int0", httpParamType, true)
	ctx.DefineIntParam("testInt1", "invalid_int1", httpParamType, false)

	ctx.DefineBoolParam("testBool0", "invalid_bool0", httpParamType, true)
	ctx.DefineBoolParam("testBool1", "invalid_bool1", httpParamType, false)

	ctx.DefineFloatParam("testFloat0", "invalid_float0", httpParamType, true)
	ctx.DefineFloatParam("testFloat1", "invalid_float1", httpParamType, false)

	ctx.DefineStringParam("testString0", "invalid_string0", httpParamType, true, 0, 10)
	ctx.DefineStringParam("testString1", "invalid_string1", httpParamType, false, 1, 20)

	ctx.DefineObjectIdParam("testObjectId0", "invalid_objectId0", httpParamType, true)
	ctx.DefineObjectIdParam("testObjectId1", "invalid_objectId1", httpParamType, false)

	if httpParamType == HttpParamJsonPost {
		ctx.DefineJsonParam("testJson0", "invalid_json0", httpParamType, true)
		ctx.DefineJsonParam("testJson1", "invalid_json1", httpParamType, false)

		ctx.DefineJsonParam("testJsonArray0", "invalid_json_array_1", httpParamType, true)
	}
}

func validateParamOutput(paramTypeName string, ctx *HttpContext, t *testing.T) {

	if !ctx.ParamsAreValid() {
		t.Errorf("%s is broken - params are not valid", paramTypeName)
		for i := range ctx.ErrorCodes { fmt.Println(ctx.ErrorCodes[i]) }
	}

	if ctx.HasRawErrors() {
		for i := range ctx.Errors { t.Errorf("TestNewHttpContextJsonPostParams is broken - errors: %v", ctx.Errors[i]) }
	}

	if len(ctx.ErrorCodes) != 0 { t.Errorf("%s is broken - there are error codes", paramTypeName) }

	// Verify the ints

	if !ctx.Params["testInt0"].Present { t.Errorf("%s is broken - testInt0 is not present", paramTypeName) }

	if !ctx.Params["testInt0"].Valid { t.Errorf("%s is broken - testInt0 is not valid", paramTypeName) }

	if ctx.Params["testInt0"].Int() != 100 { t.Errorf("%s is broken - testInt0 is not 100", paramTypeName) }

	if ctx.Params["testInt1"].Present { t.Errorf("%s is broken - testInt1 is present", paramTypeName) }

	if !ctx.Params["testInt1"].Valid { t.Errorf("%s is broken - testInt1 is not valid", paramTypeName) }

	// Verify the bools

	if !ctx.Params["testBool0"].Present { t.Errorf("%s is broken - testBool0 is not present", paramTypeName) }

	if !ctx.Params["testBool0"].Valid { t.Errorf("%s is broken - testBool0 is not valid", paramTypeName) }

	if ctx.Params["testBool0"].Bool() != true { t.Errorf("%s is broken - testBool0 is not true", paramTypeName) }

	if ctx.Params["testBool1"].Present { t.Errorf("%s is broken - testBool1 is present", paramTypeName) }

	if !ctx.Params["testBool1"].Valid { t.Errorf("%s is broken - testBool1 is not valid", paramTypeName) }

	// Verify the floats

	if !ctx.Params["testFloat0"].Present { t.Errorf("%s is broken - testFloat0 is not present", paramTypeName) }

	if !ctx.Params["testFloat0"].Valid { t.Errorf("%s is broken - testFloat0 is not valid", paramTypeName) }

	if ctx.Params["testFloat0"].Float() != 99.9999999999 { t.Errorf("testFloat0 is not 99.9999999999") }

	if ctx.Params["testFloat1"].Present { t.Errorf("%s is broken - testFloat1 is present", paramTypeName) }

	if !ctx.Params["testFloat1"].Valid { t.Errorf("%s is broken - testFloat1 is not valid", paramTypeName) }

	// Verify the strings

	if !ctx.Params["testString0"].Present { t.Errorf("%s is broken - testString0 is not present", paramTypeName) }

	if !ctx.Params["testString0"].Valid { t.Errorf("%s is broken - testString0 is not valid", paramTypeName) }

	if ctx.Params["testString0"].String() != "hello!" { t.Errorf("%s is broken - testString0 is not hello!", paramTypeName) }

	if !ctx.Params["testString1"].Present { t.Errorf("%s is broken - testString1 is present", paramTypeName) }

	if !ctx.Params["testString1"].Valid { t.Errorf("%s is broken - testString1 is not valid", paramTypeName) }

	// Verify the object ids

	if !ctx.Params["testObjectId0"].Present { t.Errorf("%s is broken - testObjectId0 is not present", paramTypeName) }

	if !ctx.Params["testObjectId0"].Valid { t.Errorf("%s is broken - testObjectId0 is not valid", paramTypeName) }

	if ctx.Params["testObjectId0"].ObjectId().Hex() != "52e29b18eee7d580e9bb1544" { t.Errorf("%s is broken - testObjectId0 is not 52e29b18eee7d580e9bb1544", paramTypeName) }

	if ctx.Params["testObjectId1"].Present { t.Errorf("%s is broken - testObjectId1 is present", paramTypeName) }

	if paramTypeName == "jsonpost" {
		if !ctx.Params["testJson0"].Present { t.Errorf("%s is broken - testJson0 is not present", paramTypeName) }
		if ctx.Params["testJson1"].Present { t.Errorf("%s is broken - testJson1 is present", paramTypeName) }

		nestedDoc := ctx.Params["testJson0"].Value.(map[string]interface{})

		if nestedDoc == nil { t.Errorf("%s is broken - testJson0 is nil", paramTypeName); return }

		if nestedDoc["testStr0"] != "ok" { t.Errorf("%s is broken - testJson0.testJson0.testStr0 does not equal ok", paramTypeName) }
		if nestedDoc["testStr1"] != "what" { t.Errorf("%s is broken - testJson0.testJson0.testStr1 does not equal what", paramTypeName) }

		if !ctx.Params["testJsonArray0"].Present { t.Errorf("%s is broken - testJsonArray0 is not present", paramTypeName) }
		nestedArray := ctx.Params["testJsonArray0"].Value.([]interface{})
		if nestedArray == nil || len(nestedArray) == 0 { t.Errorf("%s is broken - testJsonArray0 is nil", paramTypeName); return }
	}
}

