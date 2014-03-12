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
	"bytes"
	"testing"
)

func TestCompressUncompressBytes(t *testing.T) {

	var testBytes []byte
	for i := 0; i < 10001; i++ { testBytes = append(testBytes, []byte("this is a random test of this and that as a test that is random that is another")...) }

	compressedBytes, err := CompressBytes(testBytes)

	if err != nil {
		t.Errorf("TestCompressUncompressBytes is broken - unable to compress bytes: %v", err)
		return
	}

	if len(compressedBytes) == 0 {
		t.Errorf("TestCompressUncompressBytes is broken - byte slice length is zero")
		return
	}

	uncompressedBytes, err := UncompressBytes(compressedBytes)
	if err != nil {
		t.Errorf("TestCompressUncompressBytes is broken - unable to uncompress bytes: %v", err)
		return
	}

	if len(uncompressedBytes) != len(testBytes) {
		t.Errorf("TestCompressUncompressBytes is broken - uncompressed bytes length: %d - orig length: %d", len(uncompressedBytes), len(testBytes))
	}

	if !bytes.Equal(uncompressedBytes, testBytes) {
		t.Errorf("TestCompressUncompressBytes is broken - before and after slices are not equal")
	}
}

