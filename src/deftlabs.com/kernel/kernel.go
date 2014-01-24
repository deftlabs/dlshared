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

package deftlabskernel

import (
	"os"
	"os/signal"
	"syscall"
	"math/rand"
	"time"
	"deftlabs.com/log"
)

type Kernel struct {
	Configuration *Configuration
	Components map[string]interface{}
	components []Component
	Id string
	slogger.Logger
}

type Component struct {
	componentId string
	singleton interface{}
	startFunction StartStopFunction
	stopFunction StartStopFunction
}

type StartStopFunction func(kernel *Kernel) error

// Register a component with a start and stop functions.
func (self *Kernel) AddComponentWithStartStopFunctions(componentId string, singleton interface{}, startFunction StartStopFunction, stopFunction StartStopFunction) {

	component := Component{ componentId : componentId, singleton : singleton, startFunction : startFunction, stopFunction : stopFunction }

	self.components = append(self.components , component)
	self.Components[componentId] = singleton
}

// Register a component with a start function.
func (self *Kernel) AddComponentWithStartFunction(componentId string, singleton interface{}, startFunction StartStopFunction) {
	self.AddComponentWithStartStopFunctions(componentId, singleton, startFunction, nil)
}

// Register a component with a stop function.
func (self *Kernel) AddComponentWithStopFunction(componentId string, singleton interface{}, stopFunction StartStopFunction) {
	self.AddComponentWithStartStopFunctions(componentId, singleton, nil, stopFunction)
}

// Register a component without a start or stop function..
func (self *Kernel) AddComponent(componentId string, singleton interface{}) {
	self.AddComponentWithStartStopFunctions(componentId, singleton, nil, nil)
}

// Stop the kernel. Call this before exiting.
func (self *Kernel) Stop() error {

	self.Logf(slogger.Info, "Stopping %s server - version: %s - config file %s", self.Id, self.Configuration.Version, self.Configuration.FileName)

	for i := len(self.components)-1 ; i >= 0 ; i-- {
		if self.components[i].stopFunction != nil {
			if  err := self.components[i].stopFunction(self); err != nil {
				return err
			}
		}
	}

	self.Logf(slogger.Info, "Stopped %s server - version: %s - config file: %s", self.Id, self.Configuration.Version, self.Configuration.FileName)

	return nil
}

// Call this after the kernel has been created and components registered.
func (self *Kernel) Start() error {

	rand.Seed(time.Now().UTC().UnixNano())

	self.Logf(slogger.Info, "Starting %s server - version: %s - config file: %s", self.Id, self.Configuration.Version, self.Configuration.FileName)

	for i := range self.components {
		if self.components[i].startFunction != nil {
			if  err := self.components[i].startFunction(self); err != nil {
				return err
			}
		}
	}

	self.Logf(slogger.Info, "Started %s server - version: %s - config file: %s ", self.Id, self.Configuration.Version, self.Configuration.FileName)

	return nil
}

func newKernel(id, configFileName string) (*Kernel, error) {

	// Init the application configuration
	conf, err := NewConfiguration(configFileName)

	if err != nil {
		return nil, err
	}

	logger := slogger.Logger {
		Prefix: id,
		Appenders: [] slogger.Appender{
			slogger.LevelFilter(slogger.Debug, slogger.StdErrAppender()),
		},
	}

	kernel := &Kernel{ Components : make(map[string]interface{}), Configuration : conf }
	kernel.Logger = logger
	kernel.Id = id

	return kernel, nil
}

// Call this from your main to create the kernel. After init kernel is called you must add
// your components and then call kernel.Start()
func StartKernel(id string, configFileName string, addComponentsFunction func(kernel *Kernel)) (*Kernel, error) {

	kernel, err := newKernel(id, configFileName)
	if err != nil {
		return nil, err
	}

	addComponentsFunction(kernel)

    if err = kernel.Start(); err != nil {
		return nil, err
	}

	return kernel, nil
}

// ListenForInterrupt blocks until an interrupt signal is detected.
func (self *Kernel) ListenForInterrupt() error {
	quitChannel := make(chan bool)

	// Register the interrupt listener.
	interruptSignalChannel := make(chan os.Signal, 1)
	signal.Notify(interruptSignalChannel, os.Interrupt)
	go func() {
		for sig := range interruptSignalChannel {
			if sig == syscall.SIGINT {
				quitChannel <- true
			}
		}
	}()

	// Block until we receive the stop notification.
	select {
		case <- quitChannel: {
			return self.Stop()
		}
	}

	// Should never happen
	panic("How did we end up here?")
}

