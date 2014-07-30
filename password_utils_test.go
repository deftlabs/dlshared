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
	"time"
	"math/rand"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())

	seenPasswords := make(map[string]bool)

	for idx := 0; idx < 10000; idx++ {
		password := GeneratePassword(8, 10, 8, 2, 1, 1)
		if len(password) < 8 { t.Errorf("TestGeneratePassword is broken - expecting min length: 8 - recieved: %d", len(password)) }
		if len(password) > 10 { t.Errorf("TestGeneratePassword is broken - expecting max length: 10 - recieved: %d", len(password)) }
		if _, found := seenPasswords[password]; found { t.Errorf("TestGeneratePassword is broken - saw same password") }
		seenPasswords[password] = true
	}
}

