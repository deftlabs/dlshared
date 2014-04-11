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

import (
	"testing"
)

func TestLoadConfiguration(t *testing.T) {

	configuration, err := NewConfiguration("test/configuration.json")
	if err != nil { t.Errorf("NewConfiguration is broken: %v", err); return }

	intVal := configuration.Int("server.http.port", 0)
	if intVal  != 9999 { t.Errorf("Configuration server.http.port is broken - expected 9999 - received: %d", intVal) }

	intVal = configuration.IntWithPath("mongoDb.testDb", "syncTimeoutInMs", 0)
	if intVal != 30000 { t.Errorf("Configuration mongoDb.testDb.syncTimeoutInMs is broken - expected 30000 - received: %d", intVal) }

	strVal := configuration.StringWithPath("mongoDb.testDb", "type", "")
	if strVal != "standalone" { t.Errorf("Configuration mongoDb.testDb.type is broken - expected standalone - received: %s", strVal) }

	interfaceVal := configuration.InterfaceWithPath("mongoDb", "testDb", nil)
	if interfaceVal == nil { t.Errorf("Configuration mongoDb.testDb is broken - expected interface - received: nil"); return }

	// Make sure the cast is correct
	_ = interfaceVal.(map[string]interface{})

	// Load the cron example
	interfaceVal = configuration.InterfaceWithPath("cron", "scheduled", nil)
	if interfaceVal == nil { t.Errorf("Configuration cron.scheduled is broken - expected interface - received: nil"); return }

	mapVal := interfaceVal.(map[string]interface{})

	// Pull out the array.
	nestedSlice, found := mapVal["scheduledFunctions"]

	if !found || nestedSlice == nil { t.Errorf("Configuration cron.scheduled.scheduledFunctions is missing"); return }
}

