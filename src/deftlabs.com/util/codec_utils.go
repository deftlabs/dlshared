/**
 * (C) Copyright 2013 Deft Labs
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
	"crypto/md5"
	"fmt"
	"errors"
	"io"
)

func Md5Hex(v string) (string, error) {

	if v == "" {
		return "", errors.New("Value cannot be nil/empty")
	}

	h := md5.New()

	written, err := io.WriteString(h, v)

	if err != nil {
		return "", err
	}

	if written != len(v) {
		return "", errors.New("Written does not equal length")
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil

}

