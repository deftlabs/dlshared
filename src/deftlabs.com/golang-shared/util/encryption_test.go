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

import "testing"

func TestPasswordMatchesHash(t *testing.T) {

	hashedPassword, err := HashPassword("test", 4)

	if err != nil {
		t.Errorf("PasswordMatchesHash is broken - error from HashPassword: %v", err)
	}

	matches, err := PasswordMatchesHash(hashedPassword, "test")

	if err != nil {
		t.Errorf("PasswordMatchesHash is broken - error returned", err)
	}

	if !matches {
		t.Errorf("PasswordMatchesHash is broken - stored value did not match")
	}

	testHashedPassword := []byte("$2a$04$BBN.CGtEMXLF/S.YRN/Qiuanib4nFztcfF9xlHDYpjcLn4RjHf6x2")

	matches, err = PasswordMatchesHash(testHashedPassword, "test")

	if err != nil {
		t.Errorf("PasswordMatchesHash is broken - error returned: %v", err)
	}

	if !matches {
		t.Errorf("PasswordMatchesHash is broken - stored value did not match")
	}
}

func TestHashPassword(t *testing.T) {

	if _, err := HashPassword("test", 4); err != nil {
		t.Errorf("EncryptPassword is broken - basic encrypt: %v", err)
	}

	if _, err := HashPassword("test", 99); err == nil {
		t.Errorf("EncryptPassword is broken - max cost")
	}

	if _, err := HashPassword("test", -1); err != nil {
		t.Errorf("EncryptPassword is broken: %v", err)
	}

	hashedPassword, err := HashPassword("testamuchlongerpasswordsomethingpeoplewoulduse", 4)

	if err != nil {
		t.Errorf("EncryptPassword is broken - encrypt: %v", err)
	}

	if hashedPassword == nil || len(hashedPassword) == 0 {
		t.Errorf("EncryptPassword is broken - invalid value retruned")
	}
}

