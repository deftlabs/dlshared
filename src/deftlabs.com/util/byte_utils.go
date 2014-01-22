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

import(
	"deftlabs.com/log"
	"labix.org/v2/mgo/bson"
)

const NullCharacter = '\x00'

func ExtractUInt32(data []byte, offset int) uint32 {
	return (uint32(data[offset]) << 0) | (uint32(data[offset+1]) << 8) | (uint32(data[offset+2]) << 16) | (uint32(data[offset+3]) << 24)
}

func ExtractUInt16(data []byte, offset int) uint16 {
	return (uint16(data[offset]) << 0) | (uint16(data[offset+1]) << 8)
}

func ExtractInt32(data []byte, offset int) int32 {
	return int32((uint32(data[offset]) << 0) | (uint32(data[offset+1]) << 8) | (uint32(data[offset+2]) << 16) | (uint32(data[offset+3]) << 24))
}

func ExtractInt64(data []byte, offset int) int64 {
	return 	int64((uint64(data[offset]) << 0) |
			(uint64(data[offset+1]) << 8) |
			(uint64(data[offset+2]) << 16) |
			(uint64(data[offset+3]) << 24) |
			(uint64(data[offset+4]) << 32) |
			(uint64(data[offset+5]) << 40) |
			(uint64(data[offset+6]) << 48) |
			(uint64(data[offset+7]) << 56))
}

func ExtractBytes(data []byte, offset int, length int) ([]byte, error) {

	if len(data) < int(offset+length) {
		return nil, slogger.NewStackError("Byte array is not long enough to extract data")
	}

	return data[offset:offset+length], nil
}

func ExtractBsonDoc(data []byte, offset int) (*bson.M, int, error) {

	bsonLength := int(ExtractInt32(data, offset))

	if bsonBytes, err := ExtractBytes(data, offset, bsonLength); err != nil {
		return nil, bsonLength, err
	} else {
		bsonDoc := &bson.M{}
		if err = bson.Unmarshal(bsonBytes, bsonDoc); err != nil {
			return nil, bsonLength, err
		}

		return bsonDoc, bsonLength, nil
	}
}

// Extracts a cstring from the data.
func ExtractCStr(data []byte, offset int) (string, error) {

	end := offset
	maxLength := len(data)

	for ; end != maxLength; end++ {
		if data[end] == NullCharacter {
			break
		}
	}

	// Make sure the last character is the null char.
	if data[end] != NullCharacter {
		return "", slogger.NewStackError("Corrupt string - did not end with null char")
	}

	return string(data[offset:end]), nil
}

func WriteInt32(data []byte, v int32, offset int) {
	u := uint32(v)
	data[offset] = byte(u)
	data[offset+1] = byte(u>>8)
	data[offset+2] = byte(u>>16)
	data[offset+3] = byte(u>>24)
}

func WriteInt64(data []byte, v int64, offset int) {
	u := uint64(v)
	data[offset] = byte(u)
	data[offset+1] = byte(u>>8)
	data[offset+2] = byte(u>>16)
	data[offset+3] = byte(u>>24)
	data[offset+4] = byte(u>>32)
	data[offset+5] = byte(u>>40)
	data[offset+6] = byte(u>>48)
	data[offset+7] = byte(u>>56)
}

func AppendInt32(data []byte, v int32) []byte {
	u := uint32(v)
	return AppendByte(data, byte(u), byte(u>>8), byte(u>>16), byte(u>>24))
}

func AppendInt64(data []byte, v int64) []byte {
	u := uint64(v)
	return AppendByte(data, byte(u), byte(u>>8), byte(u>>16), byte(u>>24), byte(u>>32), byte(u>>40), byte(u>>48), byte(u>>56))
}

func AppendBson(data []byte, v interface{}) ([]byte, error) {

	if bytes, err := bson.Marshal(v); err != nil {
		return nil, err
	} else {
		updated := AppendByte(data, bytes...)
		return updated, nil
	}
}

func AppendCStr(data []byte, v string) []byte {
	data = AppendByte(data, []byte(v)...)
	data = AppendByte(data, 0)
	return data
}

// Append N bytes to the slice. This function came from to golang blog post on slices
// and internals.
func AppendByte(slice []byte, data ...byte) []byte {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newSlice := make([]byte, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}

	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

