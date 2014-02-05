/**
 * (C) Copyright 2013, Deft Labs
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
//import "fmt"

func TestEncodeStrToBase64(t *testing.T) {

	test := "try this example"

	result := EncodeStrToBase64(test)

	if len(result) == 0 {
		t.Errorf("EncodeStrToBase64 is broken - empty/nil result")
	}

	if result != "dHJ5IHRoaXMgZXhhbXBsZQ==" {
		t.Errorf("EncodeStrToBase64 is broken - result is not what is expected")
	}

	decoded, err := DecodeBase64ToStr("dHJ5IHRoaXMgZXhhbXBsZQ==")

	if err != nil {
		t.Errorf("EncodeStrToBase64 is broken - DecodeBase64ToStr return an error: %v", err)
	}

	if decoded != test {
		t.Errorf("EncodeStrToBase64 is broken - decoded value does not equal original")
	}
}

func TestMd5Hex(t *testing.T) {

	if result, err := Md5Hex("try this example"); err != nil || result != "479491a4d572898b1b2e31d3417dc6e6" {
		t.Errorf("Md5Hex is broken - expected: 479491a4d572898b1b2e31d3417dc6e6")
	}

	if result, err := Md5Hex("try this"); err != nil || result != "d5becebdd30256e08493fac40accd284" {
		t.Errorf("Md5Hex is broken - expected: d5becebdd30256e08493fac40accd284")
	}

	if result, err := Md5Hex("4d0a4c09daca8103cfc0639f:localhost:27017"); err != nil || result != "ea47c1fe0870fddff39c89c898a13364" {
		t.Errorf("Md5Hex is broken - expected: ea47c1fe0870fddff39c89c898a13364")
	}

	if result, err := Md5Hex(""); err == nil || result != "" {
		t.Errorf("Md5Hex is broken - expected an error")
	}
}

