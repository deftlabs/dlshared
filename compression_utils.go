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
	"io"
	"fmt"
	"bytes"
	"errors"
	"compress/zlib"
)

// Compress the byte slice using zlib.
func CompressBytes(val []byte) ([]byte, error) {

	var b bytes.Buffer
	writer := zlib.NewWriter(&b)

	written, err := writer.Write(val);
	if err != nil { return nil, err }

	if err := writer.Close(); err != nil { return nil, err }

	if len(val) != written { return nil, errors.New(fmt.Sprintf("Unable to copy bytes in UncompressBytes - wrote: %d - expected: %d", written, len(val))) }

	return b.Bytes(), nil
}


// Uncompress the byte slice using zlib
func UncompressBytes(val []byte) ([]byte, error) {

	reader, err := zlib.NewReader(bytes.NewBuffer(val))
	if err != nil { return nil, err }

	var out bytes.Buffer
	_, err = io.Copy(&out, reader)
	if err != nil { return nil, err }

	if err := reader.Close(); err != nil { return nil, err }

	return out.Bytes(), nil
}


