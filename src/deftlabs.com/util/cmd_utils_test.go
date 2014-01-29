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
	"bytes"
	"testing"
)

func TestCmdExecWithMaxTimeNoWait(t *testing.T) {

	if killed, err := CmdExecWithMaxTime("uptime", 2000, nil, nil); err != nil || killed {
		t.Errorf("CmdExecWithMaxTime is broken:", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if killed, err := CmdExecWithMaxTime("uptime", 1000, &stdout, &stderr); err != nil || killed {
		t.Errorf("CmdExecWithMaxTime is broken: %v", err)
	}

	if stdout.Len() <= 0 {
		t.Errorf("CmdExecWithMaxTime is broken - no data in stdout")
	}

	if stderr.Len() > 0 {
		t.Errorf("CmdExecWithMaxTime is broken - data in stderr")
	}
}

func TestCmdExecWithMaxTimeWithWait(t *testing.T) {

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if killed, err := CmdExecWithMaxTime("../../../test/exec_test.sh", 250, &stdout, &stderr); err != nil || !killed {
		t.Errorf("CmdExecWithMaxTime is broken: %v", err)
	}

	if stdout.Len() <= 0 {
		t.Errorf("CmdExecWithMaxTime is broken - no data in stdout")
	}

	if stderr.Len() > 0 {
		t.Errorf("CmdExecWithMaxTime is broken - data in stderr")
	}
}

func TestCmdExecWithMaxTimeWithBwmNg(t *testing.T) {

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if killed, err := CmdExecWithMaxTime("bwm-ng", 250, &stdout, &stderr); err != nil || !killed {
		t.Errorf("CmdExecWithMaxTime is broken %v", err)
	}

	if stdout.Len() <= 0 {
		t.Errorf("CmdExecWithMaxTime is broken - no data in stdout")
	}

	if stderr.Len() > 0 {
		t.Errorf("CmdExecWithMaxTime is broken - data in stderr")
	}
}

func TestCmdExecWithMaxTimeWithCat(t *testing.T) {

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if killed, err := CmdExecWithMaxTime("cat", 250, &stdout, &stderr, "/dev/urandom"); err != nil || !killed {
		t.Errorf("CmdExecWithMaxTime is broken: %v", err)
	}

	if stdout.Len() <= 0 {
		t.Errorf("CmdExecWithMaxTime is broken - no data in stdout")
	}

	if stderr.Len() > 0 {
		t.Errorf("CmdExecWithMaxTime is broken - data in stderr")
	}

}

func TestCmdExecWithMissingCmd(t *testing.T) {
	if killed, err := CmdExecWithMaxTime("garbage123", 1000, nil, nil); err == nil || killed {
		t.Errorf("CmdExecWithMaxTime is broken - missing command does not break")
	}
}

func TestCmdExecWithMissingBadParams(t *testing.T) {
	if killed, err := CmdExecWithMaxTime("top", 1000, nil, nil, "-skdfjaskdf"); err == nil || killed {
		t.Errorf("CmdExecWithMaxTime is broken - bad command param does not break")
	}
}

func TestCmdExecWithMaxTimeBadMaxTimeMs(t *testing.T) {
	// Verify the panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("CmdExecWithMaxTime is broken - it did not panic on bad maxTimeMs")
		}
	}()
	CmdExecWithMaxTime("uptime", -1, nil, nil)
}

func TestCmdExecWithMaxTimeBadCmdName(t *testing.T) {
	// Verify the panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("CmdExecWithMaxTime is broken - it did not panic on bad cmdName")
		}
	}()
	CmdExecWithMaxTime("", 1, nil, nil)
}

