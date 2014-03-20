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

// Returns true if all the values are found in the toCheck slice. If the
// values slice is empty, false is returned. This is intended to be used
// on very small slices.
func StrSliceValuesInSlice(values []string, toCheck []string) bool {

	needToFind := len(values)
	if needToFind == 0 || len(toCheck) == 0 { return false }
	found := 0

	for _, value := range values {
		if found == needToFind { break }
		for _, check := range toCheck { if value == check { found++ } }
	}

	return needToFind == found
}

