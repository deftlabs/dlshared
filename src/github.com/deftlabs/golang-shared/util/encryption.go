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

package deftlabsutil

import (
	"github.com/deftlabs/golang-shared/log"
	"code.google.com/p/go.crypto/bcrypt"
)

// Encrypt a password. The cost option is 4 - 31. If the cost is above 31,
// then an error is displayed. If the cost is below four, then four is used. If a
// nil or empty password is passed, this method returns an error.
func HashPassword(password string, cost int) ([]byte, error) {
	if len(password) == 0 {
		return nil, slogger.NewStackError("You must pass in a non-empty password")
	}

	return bcrypt.GenerateFromPassword([]byte(password), cost)
}

// Check to see if the password is the same as the hashed value. If the values match,
// true is returned. If a nil or empty password or hash is passed, this method returns an error.
func PasswordMatchesHash(hashedPassword []byte, password string) (bool, error) {

	if len(password) == 0 {
		return false, slogger.NewStackError("You must pass in a non-empty password")
	}

	if len(hashedPassword) == 0 {
		return false, slogger.NewStackError("You must pass in a non-empty hashed password")
	}

	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	return err == nil, nil
}

