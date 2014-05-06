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
)

// The consumer is a generic component that is meant to be embedded into your applications. It
// allows you to create a specific number of goroutines to handle the realtime data/msg processing.
// The contact is that you must call the NewConsumer method to create the struct. After that, you
// must call the Start method before attempting to pass in a msg. To stop the consumer, close the
// receive channel and THEN call the Stop method.
type Consumer struct {

	Logger

	name string

	consumeFunc func(msg interface{})

	receiveChannel chan interface{}

	maxGoroutines int

	maxWaitOnStopInMs int64

	waitGroup *sync.WaitGroup
}

// Create the consumer. You must call this method to create a consumer. If you set the maxGoroutines
// to zero (or less) then it will spawn a new goroutine for each request and there is no limit. If maxGoroutines
// is set, these goroutines are created and added to a pool when Start is called. The maxWaitOnStopInMs is the amount of time the Stop
// method will wait for the goroutines to finish clearing out what they are processing. A maxWaitOnStopInMs value of zero
// indicates an unlimited wait. You MUST close the receive channel before calling stop.
func NewConsumer(	name string,
					receiveChannel chan interface{},
					consumeFunc func(msg interface{}),
					maxGoroutines,
					maxWaitOnStopInMs int,
					logger Logger) *Consumer {

	if maxWaitOnStopInMs < 0 { maxWaitOnStopInMs = 0 }
	if maxGoroutines <= 0 { maxGoroutines = 1 }

	consumer := &Consumer{
		Logger: logger,
		name: name,
		receiveChannel: receiveChannel,
		maxGoroutines: maxGoroutines,
		maxWaitOnStopInMs: int64(maxWaitOnStopInMs),
		waitGroup: new(sync.WaitGroup),
	}

	consumer.consumeFunc = func(msg interface{}) {
		defer func() {
			if r := recover(); r != nil {
				consumer.Logf(Warn, "Consume func in consumer: %s - panicked - err: %v", consumer.name, r)
			}
		}()
		consumeFunc(msg)
	}

	return consumer
}

// This method listens to the receive channel and then
// calls the consume function passed. The passed consume
// function should never panic. Don't Panic!
func (self *Consumer) msgProcessor() {
	defer self.waitGroup.Done()
	for msg := range self.receiveChannel { self.consumeFunc(msg) }
}

func (self *Consumer) Start() error {

	// Create the goroutine pool
	for idx := 0; idx < self.maxGoroutines; idx++ { self.waitGroup.Add(1); go self.msgProcessor() }

	return nil
}

// This method will block until all goroutines exit. It is up to the
// caller to clear the receiveChannel.
func (self *Consumer) Stop() error {

	// If the max wait on stop in ms equals zero, we will wait indefinitely before
	// stopping.
	if self.maxWaitOnStopInMs == 0 {
		self.waitGroup.Wait()
	} else {
		stopNotification := make(chan bool, 1)
		var stopWaitGroup sync.WaitGroup
		stopWaitGroup.Add(1)

		go func() {
			defer func() { stopWaitGroup.Done() }()
			self.waitGroup.Wait()
			stopNotification <- true
		}()

		select {
			case <- stopNotification: // This is a clean shutdown, do nothing.
			case <- time.After(time.Duration(self.maxWaitOnStopInMs) * time.Millisecond):
				self.Logf(Warn, "Cunsumer : %s - unabled to shutdown cleanly - stopping", self.name)
		}
	}

	return nil
}

