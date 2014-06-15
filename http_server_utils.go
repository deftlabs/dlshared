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
	"strings"
	"reflect"
	"strconv"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"labix.org/v2/mgo/bson"
	"github.com/gorilla/mux"
)

type HttpParamDataType int8
type HttpParamType int8

const(
	HttpIntParam = HttpParamDataType(0)
	HttpStringParam = HttpParamDataType(1)
	HttpFloatParam = HttpParamDataType(2)
	HttpBoolParam = HttpParamDataType(3) // Boolean types include: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false
	HttpObjectIdParam = HttpParamDataType(4)
	HttpJsonParam = HttpParamDataType(5)
	HttpJsonArrayParam = HttpParamDataType(6)

	// All of the param types only support single values (i.e., no slices). If multiple values are present, the
	// first is taken.
	HttpParamPost = HttpParamType(0)
	HttpParamJsonPost = HttpParamType(1) // When the content body is posted in json format. This only supports one level
	HttpParamQuery = HttpParamType(2)
	HttpParamHeader = HttpParamType(3)
	HttpParamPath = HttpParamType(4) // This must be declared as {someName} in the path mapping
)

type HttpContext struct {
	Response http.ResponseWriter
	Request *http.Request
	Params map[string]*HttpParam
	ErrorCodes []string
	Errors []error

	postJson map[string]interface{}

	body []byte
}

type HttpParam struct {
	Name string
	InvalidErrorCode string
	DataType HttpParamDataType
	Type HttpParamType
	Required bool
	MinLength int
	MaxLength int
	Post bool
	Value interface{}
	Raw string
	Valid bool
	Present bool // If value is present and parsed properly
}

// Make sure your params are present and valid before trying to access.
func (self *HttpParam) Int() int { return self.Value.(int) }

func (self *HttpParam) Float() float64 { return self.Value.(float64) }

func (self *HttpParam) String() string { return self.Value.(string) }

func (self *HttpParam) Bool() bool { return self.Value.(bool) }

func (self *HttpParam) ObjectId() *bson.ObjectId { if self.Value != nil { return self.Value.(*bson.ObjectId) } else { return nil } }

func (self *HttpParam) Json() map[string]interface{} { if self.Value == nil { return nil } else { return self.Value.(map[string]interface{}) } }

func (self *HttpParam) JsonArray() []interface{} { if self.Value == nil { return nil } else { return self.Value.([]interface{}) } }

// Set a valid value for a param. Missing can be valid, but not present.
func (self *HttpParam) setPresentValue(value interface{}) {
	self.Present = true
	self.Value = value
}

// Validate the params. If any of the params are invalid, false is returned. You must call
// this first before calling the ErrorCodes []string. If not params are defined, this always
// returns "true". If there are raw data extraction errors, this is always false (e.g., body missing or incorrect).
func (self *HttpContext) ParamsAreValid() bool {

	if len(self.Errors) != 0 { return false }

	if len(self.Params) == 0 { return true }

	for _, param := range self.Params {
		switch param.DataType {
			case HttpIntParam: validateIntParam(self, param)
			case HttpStringParam: validateStringParam(self, param)
			case HttpFloatParam: validateFloatParam(self, param)
			case HttpBoolParam: validateBoolParam(self, param)
			case HttpObjectIdParam: validateObjectIdParam(self, param)
			case HttpJsonParam: validateJsonParam(self, param)
			case HttpJsonArrayParam: validateJsonParam(self, param)
		}
	}

	return len(self.ErrorCodes) == 0
}

func (self *HttpContext) ParamInt(name string) int { return self.Params[name].Int() }

func (self *HttpContext) ParamFloat(name string) float64 { return self.Params[name].Float() }

func (self *HttpContext) ParamString(name string) string { return self.Params[name].String() }

func (self *HttpContext) HasParam(name string) bool {
	if val, found := self.Params[name]; !found || val == nil || val.Value == nil { return false }
	return true
}

func (self *HttpContext) ParamBool(name string) bool { return self.Params[name].Bool() }

func (self *HttpContext) ParamObjectId(name string) *bson.ObjectId { return self.Params[name].ObjectId() }

func (self *HttpContext) ParamJson(name string) map[string]interface{} { return self.Params[name].Json() }

func (self *HttpContext) ParamJsonArray(name string) []interface{} { return self.Params[name].JsonArray() }

func (self *HttpContext) HasRawErrors() bool { return len(self.Errors) > 0 }

// This returns the param value as a string. If the param is missing or empty,
// the string will be len == 0.
func retrieveParamValue(ctx *HttpContext, param *HttpParam) interface{} {
	switch param.Type {
		case HttpParamPost: return strings.TrimSpace(ctx.Request.PostFormValue(param.Name))
		case HttpParamJsonPost: return retrieveJsonParamValue(ctx, param)
		case HttpParamQuery: return strings.TrimSpace(ctx.Request.FormValue(param.Name))
		case HttpParamHeader: return strings.TrimSpace(ctx.Request.Header.Get(param.Name))
		case HttpParamPath: return strings.TrimSpace(mux.Vars(ctx.Request)[param.Name])
	}
	return nadaStr
}

func retrieveJsonParamValue(ctx *HttpContext, param *HttpParam) interface{} {

	var noData interface{}

	if param.DataType != HttpJsonParam { noData = nadaStr }

	if len(ctx.Errors) > 0 { return noData }

	// If this is the first access, read the body
	if len(ctx.body) == 0 {
		var err error
		ctx.body, err = ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.Errors = append(ctx.Errors, NewStackError("Error in raw data extraction - error: %v", err))
			return noData
		}
	}

	if ctx.postJson == nil {
		var genJson interface{}
		err := json.Unmarshal(ctx.body, &genJson)
		if err != nil {
			ctx.Errors = append(ctx.Errors, NewStackError("Error in raw json data extraction - error: %v", err))
			return noData
		}

		ctx.postJson = genJson.(map[string]interface{})
	}

	// Look for the value in the json. The json may hold the data in a variety
	// of formats. Convert back to a string to deal with the other data types :-(
	val, found := ctx.postJson[param.Name]
	if !found { return noData }

	// If this is json, return the value.
	if param.DataType == HttpJsonParam || param.DataType == HttpJsonArrayParam { return val }

	valType := reflect.TypeOf(val)

	if valType == nil { return noData }

	switch valType.Kind() {
		case reflect.Invalid: return noData
		case reflect.Bool: return fmt.Sprintf("%t", val.(bool))
		case reflect.Float64: return fmt.Sprintf("%g", val.(float64))
		case reflect.String: return val.(string)
		default: return noData
	}

	return noData
}

func appendInvalidErrorCode(ctx *HttpContext, param *HttpParam) {
	// Do not dupplicate error codes.
	for i := range ctx.ErrorCodes { if ctx.ErrorCodes[i] == param.InvalidErrorCode { return } }

	if len(param.InvalidErrorCode) == 0 {
		panic(fmt.Sprintf("We do not have an error code defined for param: %s - required: %t", param.Name, param.Required))
	}

	ctx.ErrorCodes = append(ctx.ErrorCodes, param.InvalidErrorCode)
	param.Valid = false
}

func validateIntParam(ctx *HttpContext, param *HttpParam) {

	param.Raw = retrieveParamValue(ctx, param).(string)

	if len(param.Raw) == 0 && param.Required { appendInvalidErrorCode(ctx, param); return }

	if len(param.Raw) == 0 { return }

	if val, err := strconv.Atoi(param.Raw); err != nil { appendInvalidErrorCode(ctx, param)
	} else { param.setPresentValue(val) }
}

func validateStringParam(ctx *HttpContext, param *HttpParam) {

	param.Raw = retrieveParamValue(ctx, param).(string)

	if len(param.Raw) == 0 && param.Required {
		appendInvalidErrorCode(ctx, param)
		return
	}

	if param.Required && param.MinLength > 0 && len(param.Raw) < param.MinLength {
		appendInvalidErrorCode(ctx, param)
		return
	}

	if param.Required && param.MaxLength > 0 && len(param.Raw) > param.MaxLength {
		appendInvalidErrorCode(ctx, param)
		return
	}

	param.setPresentValue(param.Raw)
}

func validateFloatParam(ctx *HttpContext, param *HttpParam) {
	param.Raw = retrieveParamValue(ctx, param).(string)

	if len(param.Raw) == 0 && param.Required { appendInvalidErrorCode(ctx, param); return }
	if len(param.Raw) == 0 { return }

	if val, err := strconv.ParseFloat(param.Raw, 64); err != nil { appendInvalidErrorCode(ctx, param)
	} else { param.setPresentValue(val) }
}

func validateObjectIdParam(ctx *HttpContext, param *HttpParam) {

	param.Raw = retrieveParamValue(ctx, param).(string)

	if len(param.Raw) == 0 && param.Required { appendInvalidErrorCode(ctx, param); return }
	if len(param.Raw) == 0 { return }

	if !bson.IsObjectIdHex(param.Raw) { appendInvalidErrorCode(ctx, param); return }

	value := bson.ObjectIdHex(param.Raw)
	param.setPresentValue(&value)
}

func validateJsonParam(ctx *HttpContext, param *HttpParam) {

	val := retrieveParamValue(ctx, param)

	if val == nil && param.Required { appendInvalidErrorCode(ctx, param); return }
	if val == nil { return }

	param.setPresentValue(val)
}

// Boolean types include: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false
func validateBoolParam(ctx *HttpContext, param *HttpParam) {

	param.Raw = retrieveParamValue(ctx, param).(string)

	if len(param.Raw) == 0 && param.Required {
		appendInvalidErrorCode(ctx, param)
		return
	}

	if len(param.Raw) == 0 { return }

	if val, err := strconv.ParseBool(param.Raw); err != nil {
		appendInvalidErrorCode(ctx, param)
	} else {
		param.setPresentValue(val)
	}
}

func (self *HttpContext) DefineIntParam(name, invalidErrorCode string, paramType HttpParamType, required bool) {
	self.Params[name] = &HttpParam{ Name: name, InvalidErrorCode: invalidErrorCode, DataType: HttpIntParam, Required: required, Type: paramType, Valid: true }
}

// Boolean types include: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false
func (self *HttpContext) DefineBoolParam(name, invalidErrorCode string, paramType HttpParamType, required bool) {
	self.Params[name] = &HttpParam{ Name: name, InvalidErrorCode: invalidErrorCode, DataType: HttpBoolParam, Required: required, Type: paramType, Valid: true }
}

func (self *HttpContext) DefineFloatParam(name, invalidErrorCode string, paramType HttpParamType, required bool) {
	self.Params[name] = &HttpParam{ Name: name, InvalidErrorCode: invalidErrorCode, DataType: HttpFloatParam, Required: required, Type: paramType, Valid: true }
}

func (self *HttpContext) DefineObjectIdParam(name, invalidErrorCode string, paramType HttpParamType, required bool) {
	self.Params[name] = &HttpParam{ Name: name, InvalidErrorCode: invalidErrorCode, DataType: HttpObjectIdParam, Required: required, Type: paramType, Valid: true }
}

func (self *HttpContext) DefineJsonParam(name, invalidErrorCode string, paramType HttpParamType, required bool) {
	self.Params[name] = &HttpParam{ Name: name, InvalidErrorCode: invalidErrorCode, DataType: HttpJsonParam, Required: required, Type: paramType, Valid: true }
}

func (self *HttpContext) DefineJsonArrayParam(name, invalidErrorCode string, paramType HttpParamType, required bool) {
	self.Params[name] = &HttpParam{ Name: name, InvalidErrorCode: invalidErrorCode, DataType: HttpJsonArrayParam, Required: required, Type: paramType, Valid: true }
}

func (self *HttpContext) DefineStringParam(name, invalidErrorCode string, paramType HttpParamType, required bool, minLength, maxLength int) {
	self.Params[name] = &HttpParam{ Name: name, InvalidErrorCode: invalidErrorCode, DataType: HttpStringParam, Required: required, Type: paramType, Valid: true }
}

// Call this method to init the http context struct.
func NewHttpContext(response http.ResponseWriter, request *http.Request) *HttpContext {
	return &HttpContext{ Response: response, Request: request, Params: make(map[string]*HttpParam) }
}

// This method returns true if the http request method is a HTTP post. If the
// field missing or incorrect, false is returned. This method will panic if
// the request is nil.
func IsHttpMethodPost(request *http.Request) bool {
	if request == nil {
		panic("request param is nil")
	}
	return len(request.Method) > 0 && strings.ToUpper(request.Method) == HttpPostMethod
}

// Write an http ok response string. The content type is text/plain.
func WriteOkResponseString(response http.ResponseWriter, msg string) error {
	if response == nil {
		return NewStackError("response param is nil")
	}

	msgLength := len(msg)

	if msgLength == 0 {
		return NewStackError("Response message is an empty string")
	}

	response.Header().Set(ContentTypeHeader, ContentTypeTextPlain)

	written, err := response.Write([]byte(msg))

	if err != nil {
		return err
	}

	if written != msgLength {
		return NewStackError("Did not write full message - bytes written %d - expected %d", written, msgLength)
	}

	return nil
}

// Encode and write a json response. If there is a problem encoding an http 500 is sent and an
// error is returned. If there are problems writting the response an error is returned.
func JsonEncodeAndWriteResponse(response http.ResponseWriter, value interface{}) error {

	if value == nil {
		return NewStackError("Nil value passed")
	}

	rawJson, err := json.Marshal(value)
	if err != nil {
		http.Error(response, "Error", 500)
		return NewStackError("Unable to marshal json: %v", err)
	}

	response.Header().Set(ContentTypeHeader, ContentTypeJson)

	written, err := response.Write(rawJson)
	if err != nil {
		return NewStackError("Unable to write response: %v", err)
	}

	if written != len(rawJson) {
		return NewStackError("Unable to write full response - wrote: %d - expected: %d", written, len(rawJson))
	}

	return nil
}

