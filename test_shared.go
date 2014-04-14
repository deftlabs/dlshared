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

const testConfigFileName = "test/configuration.json"

// Create the test kernel. This creates and adds a mongo and distributed lock component.
func baseTestStartKernel(testName string, addComponentsFunc func(kernel *Kernel)) (*Kernel, error) {

	lock := NewMongoDistributedLock("testLockId", "MongoTestDb", "test", "locks", 1, 1, 2, 86400)

	kernel, err := StartKernel(testName, testConfigFileName, func(kernel *Kernel) {
		kernel.AddComponentWithStartStopMethods("MongoTestDb", NewMongoFromConfigPath("MongoConfigDb", "mongoDb.testDb"), "Start", "Stop")
		kernel.AddComponentWithStartStopMethods("DistributedLock", lock, "Start", "Stop")
		addComponentsFunc(kernel)
	})

	if err != nil { return nil, err }

	return kernel, nil
}


