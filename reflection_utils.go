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
	"runtime"
)

// Return the function name.
//
// Initial code came from:
// http://stackoverflow.com/questions/7052693/how-to-get-the-name-of-a-function-in-go
//
// This does not include parameter names/types or package, simply the name. If the function
// is attached to a struct, it returns the struct name.
//
// This method does not work if you have [ '(' || ')' || '*' ] in your file path. If you use this method
// add a unit test to confirm it works for you needs. If it does not work for your
// application, please reach out and let us know: https://github.com/deftlabs/dlshared/issues
func GetFunctionName(i interface{}) string {

	val := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	// Strip the path if present.
	if pathIdx := strings.Index(val, "."); pathIdx != -1 { val = val[pathIdx+1:len(val)] }

	// Remove the pointer if present.
	if strings.HasPrefix(val, "*") { val = val[1:len(val)] }

	startFunc := strings.Index(val, "(")
	endFunc := strings.Index(val, ")")

	if startFunc == -1 { return reflectRemoveFilePathIfPresent(val) }

	funcStr := reflectRemoveFilePathIfPresent(val[startFunc+1:endFunc])

	if i := strings.LastIndex(funcStr, "."); i > -1 { funcStr = funcStr[i+1:len(funcStr)] }

	structStr := reflectRemoveFilePathIfPresent(val[0:startFunc-1])

	return strings.Replace(fmt.Sprintf("%s.%s", structStr, funcStr), "*", "", -1)
}

func GetMethodValueByName(val interface{}, methodName string, allowedArgs, allowedReturn int) (error, reflect.Value) {

	value := reflect.ValueOf(val)

	methodValue := value.MethodByName(methodName)

	if !methodValue.IsValid() { return NewStackError("Method: %s is NOT found on struct: %s", methodName, value.Type()), reflect.Value{} }

	methodType := methodValue.Type()

	if methodType.NumOut() != allowedReturn {
		return NewStackError(	"The method: %s on struct: %s has: %d return value(s) - you must  pass: %d",
								methodName,
								value.Type(),
								methodType.NumOut(),
								allowedReturn),
								reflect.Value{}
	}

	if methodType.NumIn() !=  allowedArgs {
		return NewStackError(	"The method: %s on struct: %s has: %d parameter(s) - you must pass: %d",
								methodName,
								value.Type(),
								methodType.NumIn(),
								allowedArgs),
								reflect.Value{}
	}

	return nil, methodValue
}

func CallNoParamNoReturnValueMethod(val interface{}, methodValue reflect.Value) { methodValue.Call(make([]reflect.Value, 0)) }

func reflectRemoveFilePathIfPresent(val string) string {
	if i := strings.LastIndex(val, "/"); i > -1 { val = val[i+1:len(val)] }
	return val
}

