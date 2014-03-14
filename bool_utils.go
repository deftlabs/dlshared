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

import "strings"

const (
	TrueStrTrue = "true"
	TrueStrYes = "yes"
	TrueStrY = "y"
	TrueStr1 = "1"

	FalseStrFalse = "false"
	FalseStrNo = "no"
	FalseStrN = "n"
	FalseStr0 = "0"
)

// Returns true if the string is (case insensitive): (true | yes | y | 1)
func StrIsTrue(val string) bool {
	if len(val) == 0 { return false }
	s := strings.ToLower(val)
	return s == TrueStrTrue || s == TrueStrYes || s == TrueStrY || s == TrueStr1
}

// Returns false if the string is (case insensitive): (false | no | n | 0)
func StrIsFalse(val string) bool {
	if len(val) == 0 { return false }
	s := strings.ToLower(val)
	return s == FalseStrFalse || s == FalseStrNo || s == FalseStrN || s == FalseStr0
}

