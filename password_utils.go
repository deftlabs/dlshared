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

import "math/rand"

var passwordGenerateLowerCase = []byte("abcdefgijkmnpqrstwxyz")
var passwordGenerateUpperCase = []byte("ABCDEFGHJKLMNPQRSTWXYZ")
var passwordGenerateNumeric = []byte("23456789")
var passwordGenerateSpecial = []byte("*$+?&=!%{}/")

// Generate a random password based on the inputs. Make sure you seed random on server/process startup
// E.g., rand.Seed( time.Now().UTC().UnixNano())
// This logic is based on the example on StackOverflow seen here:
// http://stackoverflow.com/questions/4090021/need-a-secure-password-generator-recommendation
func GeneratePassword(	minPasswordLength,
						maxPasswordLength,
						minLowerCaseCount,
						minUpperCaseCount,
						minNumberCount,
						minSpecialCount int32) string {

    groupsUsed := map[string]int32 {
		"lower": minLowerCaseCount,
		"upper": minUpperCaseCount,
		"num": minNumberCount,
		"special": minSpecialCount,
	}

	if minPasswordLength > maxPasswordLength {
		maxPasswordLength = minPasswordLength
	}

	passwordLength := minPasswordLength + rand.Int31n(maxPasswordLength - minPasswordLength)

	password := make([]byte, passwordLength, passwordLength)

	remainingBytes := minLowerCaseCount + minUpperCaseCount + minNumberCount + minSpecialCount

	for idx := int32(0); idx < passwordLength; idx++ {

		selectable := make([]byte, 0)

		if (remainingBytes < (int32(len(password)) - idx)) {
			selectable = append(selectable, passwordGenerateLowerCase...)
			selectable = append(selectable, passwordGenerateUpperCase...)
			selectable = append(selectable, passwordGenerateNumeric...)
			selectable = append(selectable, passwordGenerateSpecial...)
		} else {
			for key, value := range groupsUsed {
				if value <= 0 { continue }
				switch key {
					case "lower": selectable = append(selectable, passwordGenerateLowerCase...)
					case "upper": selectable = append(selectable, passwordGenerateUpperCase...)
					case "num": selectable = append(selectable, passwordGenerateNumeric...)
					case "special": selectable = append(selectable, passwordGenerateSpecial...)
				}
			}
		}

		next := selectable[rand.Int31n(int32(len(selectable)-1))]
		password[idx] = next

		var groupUsedKey string
		if byteSliceContainsByte(passwordGenerateLowerCase, next) { groupUsedKey = "lower"
		} else if byteSliceContainsByte(passwordGenerateUpperCase, next) { groupUsedKey = "upper"
		} else if byteSliceContainsByte(passwordGenerateNumeric, next) { groupUsedKey = "num"
		} else if byteSliceContainsByte(passwordGenerateSpecial, next) { groupUsedKey = "special" }

		groupsUsed[groupUsedKey] = groupsUsed[groupUsedKey] - 1
		if groupsUsed[groupUsedKey] >= 0 { remainingBytes-- }
	}

	return string(password)
}

func byteSliceContainsByte(slice []byte, c byte) bool {
	for _, v := range slice { if v == c { return true } }
	return false
}


