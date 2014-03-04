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
	"strings"
	"strconv"
	"net/http"
	"github.com/gorilla/mux"
)

type HttpParamDataType int8
type HttpParamType int8

const(
	HttpIntParam = HttpParamDataType(0)
	HttpStringParam = HttpParamDataType(1)
	HttpFloatParam = HttpParamDataType(2)
	HttpBoolParam = HttpParamDataType(3) // Boolean types include: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false

	// All of the param types only support single values (i.e., no slices). If multiple values are present, the
	// first is taken.
	HttpParamPost = HttpParamType(0)
	HttpParamQuery = HttpParamType(1)
	HttpParamHeader = HttpParamType(2)
	HttpParamPath = HttpParamType(3) // This must be declared as {someName} in the path mapping
)

type HttpContext struct {
	Response http.ResponseWriter
	Request *http.Request
	Params map[string]*HttpParam
	InvalidErrorCodes []string
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

// Set a valid value for a param. Missing can be valid, but not present.
func (self *HttpParam) setPresentValue(value interface{}) {
	self.Present = true
	self.Value = value
}

// Validate the params. If any of the params are invalid, false is returned. You must call
// this first before calling the InvalidErrorCodes []string. If not params are defined, this always
// returns "true".
func (self *HttpContext) ParamsAreValid() bool {
	if len(self.Params) == 0 {
		return true
	}

	for _, param := range self.Params {
		switch param.DataType {
			case HttpIntParam: validateIntParam(self, param)
			case HttpStringParam: validateStringParam(self, param)
			case HttpFloatParam: validateFloatParam(self, param)
			case HttpBoolParam: validateBoolParam(self, param)
		}
	}

	return len(self.InvalidErrorCodes) == 0
}

// This returns the param value as a string. If the param is missing or empty,
// the string will be len == 0.
func retrieveParamValue(ctx *HttpContext, param *HttpParam) string {
	switch param.Type {
		case HttpParamPost: return strings.TrimSpace(ctx.Request.PostFormValue(param.Name))
		case HttpParamQuery: return strings.TrimSpace(ctx.Request.FormValue(param.Name))
		case HttpParamHeader: return strings.TrimSpace(ctx.Request.Header.Get(param.Name))
		case HttpParamPath: return strings.TrimSpace(mux.Vars(ctx.Request)[param.Name])
	}
	return nadaStr
}

func appendInvalidErrorCode(ctx *HttpContext, param *HttpParam) {
	ctx.InvalidErrorCodes = append(ctx.InvalidErrorCodes, param.InvalidErrorCode)
	param.Valid = false
}

func validateIntParam(ctx *HttpContext, param *HttpParam) {
	param.Raw = retrieveParamValue(ctx, param)

	if len(param.Raw) == 0 && param.Required {
		appendInvalidErrorCode(ctx, param)
		return
	}

	if len(param.Raw) == 0 {
		return
	}

	if val, err := strconv.Atoi(param.Raw); err != nil {
		appendInvalidErrorCode(ctx, param)
	} else {
		param.setPresentValue(val)
	}
}

func validateStringParam(ctx *HttpContext, param *HttpParam) {

	param.Raw = retrieveParamValue(ctx, param)
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

	param.Raw = retrieveParamValue(ctx, param)
	if len(param.Raw) == 0 && param.Required {
		appendInvalidErrorCode(ctx, param)
		return
	}

	if len(param.Raw) == 0 {
		return
	}

	if val, err := strconv.ParseFloat(param.Raw, 64); err != nil {
		appendInvalidErrorCode(ctx, param)
	} else {
		param.setPresentValue(val)
	}
}

// Boolean types include: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false
func validateBoolParam(ctx *HttpContext, param *HttpParam) {
	param.Raw = retrieveParamValue(ctx, param)
	if len(param.Raw) == 0 && param.Required {
		appendInvalidErrorCode(ctx, param)
		return
	}

	if len(param.Raw) == 0 {
		return
	}

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

func (self *HttpContext) DefineStringParam(name, invalidErrorCode string, paramType HttpParamType, required bool, minLength, maxLength int) {
	self.Params[name] = &HttpParam{ Name: name, InvalidErrorCode: invalidErrorCode, DataType: HttpStringParam, Required: required, Type: paramType, Valid: true }
}

// Call this method to init the http context struct.
func NewHttpContext(response http.ResponseWriter, request *http.Request) *HttpContext {
	return &HttpContext{ Response: response, Request: request, Params: make(map[string]*HttpParam) }
}

