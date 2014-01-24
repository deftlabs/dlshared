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
	"os"
)

// CreateDir will create a new directory with all parent directories created. If the
// directory already exists, nothing is done. This sets the file permissions to 0750. This
// will use the default group for the user. To override, change at the OS level.
func CreateDir(path string) error {
	return os.MkdirAll(path, 0750)
}

// Delete a directory. This function checks to see if the directory exists and if it does,
// nothing is done. If the directory does exists, it is deleted.
func DeleteDir(path string) error {
	if exists, err := FileOrDirExists(path); err != nil {
		return err
	} else if !exists {
		return nil
	}

	return os.Remove(path)
}

// Check to see if a file or directory exists and the user has access. The original
// source came from: http://bit.ly/18GDn5Q
func FileOrDirExists(fileOrDir string) (bool, error) {
	if _, err := os.Stat(fileOrDir); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

