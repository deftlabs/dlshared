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

package deftlabsutil

import (
	"fmt"
	"crypto/md5"
	"encoding/base64"
	"github.com/deftlabs/golang-shared/log"
)

// Encodes the string to base64. This method panics if the value passed is nil
// or length is zero.
func EncodeStrToBase64(v string) string {

	if len(v) == 0 {
		panic("EncodeStrToBase64 - value passed is nil or empty")
	}

	return base64.StdEncoding.EncodeToString([]byte(v))
}

// Encodes the byte array to base64. This method panics if the value passed is nil
// or length is zero.
func EncodeToBase64(v []byte) string {
	if len(v) == 0 {
		panic("EncodeStrToBase64 - value passed is nil or empty")
	}

	return base64.StdEncoding.EncodeToString(v)
}

// Decodes the base64 string. This method panics if the value passed is nil
// or length is zero.
func DecodeBase64(v string) ([]byte, error) {
	if len(v) == 0 {
		panic("DecodeBase64 - value passed is nil or empty")
	}

	return base64.StdEncoding.DecodeString(v)
}

// Decodes the base64 string to a string. This method panics if the value passed is nil
// or length is zero.
func DecodeBase64ToStr(v string) (string, error) {
	if len(v) == 0 {
		panic("DecodeBase64 - value passed is nil or empty")
	}

	data, err := base64.StdEncoding.DecodeString(v)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func Md5HexFromBytes(v []byte) (string, error) {

	if len(v) == 0 {
		return "", slogger.NewStackError("Value cannot be nil/empty")
	}

	h := md5.New()

	written, err := h.Write(v)

	if err != nil {
		return "", err
	}

	if written != len(v) {
		return "", slogger.NewStackError("Written does not equal length - written: %d - len: %d", written, len(v))
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func Md5Hex(v string) (string, error) {
	return Md5HexFromBytes([]byte(v))
}

