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
	"sync"
	"time"
	"testing"
)

// Test the distributed lock. This requires mongo running on localhost, port 28000
func TestDistributedLock(t *testing.T) {

	lock := NewMongoDistributedLock("testLockId", "MongoTestDb", "test", "locks", 1, 1, 2, 86400)

	kernel, err := StartKernel("disritbutedLockTest", "test/configuration.json", func(kernel *Kernel) {
	 	kernel.AddComponentWithStartStopMethods("MongoTestDb", NewMongoFromConfigPath("MongoConfigDb", "mongoDb.testDb"), "Start", "Stop")
	 	kernel.AddComponentWithStartStopMethods("DistributedLock", lock, "Start", "Stop")
	})

	if err != nil { t.Errorf("TestDistributedLock start kernel is broken:", err); return }

	testLockUnlock(lock)

	waitGroup := new(sync.WaitGroup)

	// Make sure the lock is held
	lock.Lock()

	waitGroup.Add(10000)
	for i := 0; i < 10000; i++ {
		go func() {
			if lock.TryLock() { t.Errorf("TestDistributedLock is broken - try lock was able to get the lock") }
			waitGroup.Done()
		}()
	}

	go func() {
		lock.Lock()
		t.Errorf("TestDistributedLock is broken - we should not be able to obtain the lock.")
	}()

	if !lock.HasLock() { t.Errorf("TestDistributedLock is broken - we should have the lock.") }

	time.Sleep(1*time.Second)

	waitGroup.Wait()

	lock.Unlock()

	if err := kernel.Stop(); err != nil { t.Errorf("TestDistributedLock stop kernel is broken:", err) }
}

func testLockUnlock(lock DistributedLock) {
	lock.Lock()
	lock.Unlock()
}


