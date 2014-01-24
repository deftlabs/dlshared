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

import "testing"

func TestCreateDeleteDir(t *testing.T) {

	dirName := "/tmp/systemunittest/one"

	if err := DeleteDir(dirName); err != nil {
		t.Errorf("DeleteDir is broken", err)
	}

	if err := CreateDir(dirName); err != nil {
		t.Errorf("CreateDir is broken", err)
	}

	if err := DeleteDir(dirName); err != nil {
		t.Errorf("DeleteDir is broken", err)
	}
}

func TestFileOrDirExists(t *testing.T) {
	if exists, err := FileOrDirExists("/tmp"); err != nil || !exists {
		t.Errorf("FileOrDirExists is broken")
	}

	if exists, err := FileOrDirExists("/tmpsdksjdfkajdfksajfdkf8888"); err != nil || exists {
		t.Errorf("FileOrDirExists is broken")
	}
}

