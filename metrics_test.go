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
	"time"
	"testing"
)

func TestMetrics(t *testing.T) {

	relayFuncCallCount := 0

	relayFunc := func(sourceName string, metrics map[string]Metric) {
		relayFuncCallCount++
	}

	metrics := NewMetrics("TestSource", relayFunc, 1, 100)

	if err := metrics.Start(); err != nil {
		t.Errorf("TestMetrics Start is broken: %v", err)
	}

	startTime := CurrentTimeInMillis()

	for idx := 0; idx < 100000; idx++ {
		metrics.Gauge("test.guage", 100)
		metrics.Count("test.count.by.one")
		metrics.CountWithValue("test.count.with.value", 10000000)
	}

	execTime := CurrentTimeInMillis() - startTime

	if execTime > 500 {
		t.Errorf("TestMetrics is too slow - time in ms: %d", execTime)
	}

	time.Sleep(1 * time.Second)

	if err := metrics.Stop(); err != nil {
		t.Errorf("TestMetrics Stop is broken: %v", err)
	}

	if relayFuncCallCount == 0 {
		t.Errorf("TestMetrics relay function call is broken - not called")
	}
}

