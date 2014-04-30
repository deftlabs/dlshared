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
// It also has optional support for spillover processing if the goroutines cannot keep pace with
// the incoming events. If your spillover method blocks (i.e., cannot keep pace), it will cause further
// issues with incoming message processing. If you do not provide a spillover method, data will be dropped
// and a log statement will be generated for each discarded message.
// The contact is that you must call the NewConsumer method to create the struct.
// After that, you must call the Start/Stop methods before attempting to pass in a msg. If the passed
// consume function panics, the consumer will lose a goroutine each time, so don't panic!
type Consumer struct {

	logger Logger

	name string

	quitChannel chan bool

	consumeFunc func(msg interface{})

	spilloverFunc func(msg interface{})

	processorChannel chan interface{}

	receiveChannel chan interface{}

	processedChannel chan bool

	maxGoroutines int

	maxWaitOnStopInMs int64

	waitGroup *sync.WaitGroup
}

// Create the consumer. You must call this method to create a consumer. If you set the maxGoroutines
// to zero (or less) then it will spawn a new goroutine for each request and there is no limit. If maxGoroutines
// is set, these goroutines are created and added to a pool when Start is called. If you set channelBufferSize to something
// greater than zero then a buffered channel is created. The maxWaitOnStopMs is the amount of time the Stop
// method will wait for the goroutines to finish clearing out the processor channel. If the processor channel unbuffered,
// this does not apply. A maxWaitOnStopMs value of zero indicates an unlimited wait.
func NewConsumer(	name string,
					receiveChannel chan interface{},
					consumeFunc func(msg interface{}),
					spilloverFunc func(msg interface{}),
					maxGoroutines, channelBufferSize, maxWaitOnStopInMs int,
					logger Logger) *Consumer {

	if channelBufferSize < 0 { channelBufferSize = 0 }

	if maxWaitOnStopInMs < 0 { maxWaitOnStopInMs = 0 }

	if maxGoroutines <= 0 { maxGoroutines = 1 }

	consumer := &Consumer{
		logger: logger,
		name: name,
		consumeFunc: consumeFunc,
		spilloverFunc: spilloverFunc,
		receiveChannel: receiveChannel,
		processorChannel: make(chan interface{}, channelBufferSize),
		quitChannel: make(chan bool),
		processedChannel: make(chan bool, maxGoroutines),
		maxGoroutines: maxGoroutines,
		maxWaitOnStopInMs: int64(maxWaitOnStopInMs),
		waitGroup: new(sync.WaitGroup),
	}

	return consumer
}

// This method listens to the processor channel and then
// calls the consume function passed. The passed consume
// function should never panic. Don't Panic!
func (self *Consumer) msgProcessor() {
	defer self.waitGroup.Done()
	for msg := range self.processorChannel {
		self.consumeFunc(msg)
		self.processedChannel <- true
	}
}

func (self *Consumer) listenForMsgs() {
	defer self.waitGroup.Done()
	msgsBeingProcessed := 0

    for {
        select {
			case msg := <- self.receiveChannel:
				if msgsBeingProcessed <= self.maxGoroutines {

					self.processorChannel <- msg
					msgsBeingProcessed++

				} else {
					if self.spilloverFunc == nil {
						self.logger.Logf(Warn, "Max concurrent messages (%d) reached in consumer: %s - dropping data (no spillover func set)", self.maxGoroutines, self.name)
					} else {
						// If this blocks, it will further compound problems with the consumer
						self.spilloverFunc(msg)
					}
				}
			case <- self.processedChannel: msgsBeingProcessed--
			case <- self.quitChannel: return
        }
    }
}

func (self *Consumer) Start() error {

	// Create the goroutine pool
	for idx := 0; idx < self.maxGoroutines; idx++ { self.waitGroup.Add(1); go self.msgProcessor() }

	// Create the message listener
	self.waitGroup.Add(1)
	go self.listenForMsgs()

	return nil
}

// This method will block until all goroutines exit. It is up to the
// caller to clear the receiveChannel.
func (self *Consumer) Stop() error {

	// Exit the listen for events select
	self.quitChannel <- true

	// If this is an unbuffered channel, close, wait for the goroutines
	// to exit and then get out of here.
	if cap(self.processorChannel) == 0 {
		close(self.processorChannel)
		self.waitGroup.Wait()
		return nil
	}

	// This is a buffered channel so we need to wait until the channel
	// is emptied (if configured to do so).
	if self.maxWaitOnStopInMs > 0 {
		start := CurrentTimeInMillis()

		for {
			if len(self.processorChannel) == 0 {
				break
			}

		 	if CurrentTimeInMillis() - start >= self.maxWaitOnStopInMs {
				break
			}

			time.Sleep(10 * time.Millisecond)
		}
	}

	// Close the processor channel and wait for all of the goroutines to return
	close(self.processorChannel)
	self.waitGroup.Wait()

	return nil
}

