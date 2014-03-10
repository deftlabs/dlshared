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

// Return the function name. This returns packageName.functionName (if available).
// Initial code came from:
// http://stackoverflow.com/questions/7052693/how-to-get-the-name-of-a-function-in-go
//
// This method does not work if you have [ '(' || ')' || '*' ] in your file path. If you use this method
// add a unit test to confirm it works for you needs. If it does not work for your
// application, please reach out and let us know: https://github.com/deftlabs/dlshared/issues
//
// This will likely not work on Window$
func GetFunctionName(i interface{}) string {

	val := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	startFunc := strings.Index(val, "(")
	endFunc := strings.Index(val, ")")

	if startFunc == -1 {
		return removeFilePathIfPresent(val)
	}

	funcStr := removeFilePathIfPresent(val[startFunc+1:endFunc])

	if i := strings.LastIndex(funcStr, "."); i > -1 {
		funcStr = funcStr[i+1:len(funcStr)]
	}

	structStr := removeFilePathIfPresent(val[0:startFunc-1])

	return strings.Replace(fmt.Sprintf("%s.%s", structStr, funcStr), "*", "", -1)
}

func removeFilePathIfPresent(val string) string {
	if i := strings.LastIndex(val, "/"); i > -1 { val = val[i+1:len(val)] }

	return val
}

