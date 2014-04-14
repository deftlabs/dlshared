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
	testCalledCount int
}

func (self *testCronTestComponent) Run() {
	self.waitGroup.Add(1)
	self.runCalledCount++
	self.waitGroup.Done()
}

func (self *testCronTestComponent) Test() {
	self.waitGroup.Add(1)
	self.testCalledCount++
	self.waitGroup.Done()
}

// Test the cron services.
func TestCron(t *testing.T) {

	waitGroup := new(sync.WaitGroup)

	testComponent := &testCronTestComponent{ waitGroup: waitGroup }

	kernel, err := baseTestStartKernel("cronTest", func(kernel *Kernel) {

		kernel.AddComponent("testCronTestComponent", testComponent)

		kernel.AddComponentWithStartStopMethods("CronSvc", NewCronSvc("cron.scheduled"), "Start", "Stop")
	})

	if err != nil { t.Errorf("TestCron start kernel is broken:", err); return }

	waitGroup.Wait()

	time.Sleep(2*time.Second)

	if testComponent.runCalledCount == 0 { t.Errorf("TestCron Run() not called") }

	if testComponent.testCalledCount == 0 { t.Errorf("TestCron Test() not called") }

	if err := kernel.Stop(); err != nil { t.Errorf("TestCron stop kernel is broken:", err) }
}

