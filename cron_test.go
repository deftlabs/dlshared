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

type testCronTestComponent struct {
	waitGroup *sync.WaitGroup
	runCalledCount int
	interruptCalledCount int
	interruptReceived bool
	lock *sync.Mutex
}

func (self *testCronTestComponent) Run(interruptChannel chan bool) {
	self.waitGroup.Add(1)
	defer self.waitGroup.Done()

	self.lock.Lock()
	self.runCalledCount++
	self.lock.Unlock()

	time.Sleep(1* time.Millisecond)
}

func (self *testCronTestComponent) Interrupt(interruptChannel chan bool) {
	self.waitGroup.Add(1)
	defer self.waitGroup.Done()

	self.lock.Lock()
	self.interruptCalledCount++
	self.lock.Unlock()

	<- interruptChannel // wait for the signal

	self.lock.Lock()
	self.interruptReceived = true
	self.lock.Unlock()

}

// Test the cron services.
func TestCron(t *testing.T) {

	waitGroup := new(sync.WaitGroup)

	testComponent := &testCronTestComponent{ waitGroup: waitGroup, lock: new(sync.Mutex) }

	kernel, err := baseTestStartKernel("cronTest", func(kernel *Kernel) {
		kernel.AddComponent("testCronTestComponent", testComponent)
		kernel.AddComponentWithStartStopMethods("CronSvc", NewCronSvc("cron.scheduled"), "Start", "Stop")
	})

	if err != nil { t.Errorf("TestCron start kernel is broken:", err); return }

	waitGroup.Wait()

	time.Sleep(4*time.Second)

	// We have multiple goroutines modifying the
	testComponent.lock.Lock()
	if testComponent.runCalledCount == 0 { t.Errorf("TestCron Run(chan) not called") }
	if testComponent.interruptCalledCount == 0 { t.Errorf("TestCron Interrupt(chan) not called") }
	if !testComponent.interruptReceived { t.Errorf("TestCron Interrupt(chan) was not interrupted") }
	testComponent.lock.Unlock()

	if err := kernel.Stop(); err != nil { t.Errorf("TestCron stop kernel is broken: %v", err) }
}

