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

func TestStrSliceValuesInSlice(t *testing.T) {

	values := []string{ "one", "two" }
	toCheck := []string{ "one", "two" }
	if !StrSliceValuesInSlice(values, toCheck) { t.Errorf("StrSliceValuesInSlice is broken - exact match") }

	values = []string{ "one" }
	toCheck = []string{ "one" }
	if !StrSliceValuesInSlice(values, toCheck) { t.Errorf("StrSliceValuesInSlice is broken - exact match") }

	values = []string{ "one" }
	toCheck = []string{ "one", "two" }
	if !StrSliceValuesInSlice(values, toCheck) { t.Errorf("StrSliceValuesInSlice is broken - one match") }

	values = []string{ "three" }
	toCheck = []string{ "one", "two" }

	if StrSliceValuesInSlice(values, toCheck) { t.Errorf("StrSliceValuesInSlice is broken - no match") }

	values = []string{ "three" }
	toCheck = []string{ "one" }

	if StrSliceValuesInSlice(values, toCheck) { t.Errorf("StrSliceValuesInSlice is broken - no match") }

	if StrSliceValuesInSlice(nil, toCheck) { t.Errorf("StrSliceValuesInSlice is broken - nil") }
	if StrSliceValuesInSlice(values, nil) { t.Errorf("StrSliceValuesInSlice is broken - nil") }
	if StrSliceValuesInSlice(nil, nil) { t.Errorf("StrSliceValuesInSlice is broken - nil") }



}
