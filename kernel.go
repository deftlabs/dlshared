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
	"os"
	"fmt"
	"time"
	"reflect"
	"strconv"
	"syscall"
	"os/signal"
	"math/rand"
)

type Kernel struct {
	Configuration *Configuration
	Components map[string]Component
	components []Component
	Id string
	Logger
	Pid int
}

type Component struct {
	componentId string
	singleton interface{}
	startMethodName string
	stopMethodName string
}

// Access another component. This method will panic if you attempt to reference a
// non-existent component. If the component id has a length of zero, it is also panics.
func (self *Kernel) GetComponent(componentId string) interface{} {

	if len(componentId) == 0 {
		panic("kernel.GetComponent called with an empty component id")
	}

	if _, found := self.Components[componentId]; !found {
		panic(fmt.Sprintf("kernel.GetComponent called with an invalid component id: %s", componentId))
	}

	return self.Components[componentId].singleton.(interface{})
}

// Register a component with a start and stop methods.
func (self *Kernel) AddComponentWithStartStopMethods(componentId string, singleton interface{}, startMethodName, stopMethodName string) {

	component := Component{ componentId : componentId, singleton : singleton, startMethodName : startMethodName, stopMethodName : stopMethodName }

	self.components = append(self.components , component)
	self.Components[componentId] = component
}


// Register a component with a start method.
func (self *Kernel) AddComponentWithStartMethod(componentId string, singleton interface{}, startMethodName string) {
	self.AddComponentWithStartStopMethods(componentId, singleton, startMethodName, "")
}

// Register a component with a stop method.
func (self *Kernel) AddComponentWithStopMethod(componentId string, singleton interface{}, stopMethodName string) {
	self.AddComponentWithStartStopMethods(componentId, singleton, "", stopMethodName)
}

// Register a component without a start or stop method.
func (self *Kernel) AddComponent(componentId string, singleton interface{}) {
	self.AddComponentWithStartStopMethods(componentId, singleton, "", "")
}

// Called by the kernel during Start/Stop.
func callStartStopMethod(methodTypeName, methodName string, singleton interface{}, kernel *Kernel) error {

	value := reflect.ValueOf(singleton)

	methodValue := value.MethodByName(methodName)

	if !methodValue.IsValid() {
		return fmt.Errorf("Start method: %s is NOT found on struct: %s", methodName, value.Type())
	}

	methodType := methodValue.Type()

	if methodType.NumOut() > 1 {
		panic(fmt.Sprintf("The %s method: %s on struct: %s has more than one return value - you can only return error or nothing", methodTypeName, methodName, value.Type()))
	}

	if methodType.NumIn() > 1 {
		return fmt.Errorf("The %s method: %s on struct: %s has more than one parameter - you can only accept Kernel or nothing", methodTypeName, methodName, value.Type())
	}

	// Verify the return type is error
	if methodType.NumOut() == 1 && methodType.Out(0).Name() != "error"  {

		return fmt.Errorf("The %s method: %s on struct: %s has an invalid return type - you can return nothing or error", methodTypeName, methodName, value.Type())
	}

	methodInputs := make([]reflect.Value, 0)
	if methodType.NumIn() == 1 {
		methodInputs = append(methodInputs, reflect.ValueOf(kernel))
	}

	returnValues := methodValue.Call(methodInputs)

	// Check to see if there was an error
	if len(returnValues) == 1 {
		err := returnValues[0].Interface()
		if err != nil {
			return err.(error)
		}
	}

	return nil
}

// Call this after the kernel has been created and components registered.
func (self *Kernel) Start() error {

	rand.Seed(time.Now().UTC().UnixNano())

	self.Logf(Info, "Starting %s - version: %s - config file: %s", self.Id, self.Configuration.Version, self.Configuration.FileName)

	for i := range self.components {
		if len(self.components[i].startMethodName) > 0 {
			if err := callStartStopMethod("start", self.components[i].startMethodName, self.components[i].singleton, self); err != nil {
				return err
			}
		}
	}

	self.Logf(Info, "Started %s - version: %s - config file: %s ", self.Id, self.Configuration.Version, self.Configuration.FileName)

	return nil
}

// Stop the kernel. Call this before exiting.
func (self *Kernel) Stop() error {

	self.Logf(Info, "Stopping %s - version: %s - config file %s", self.Id, self.Configuration.Version, self.Configuration.FileName)

	for i := len(self.components)-1 ; i >= 0 ; i-- {

		if len(self.components[i].stopMethodName) > 0 {
			if err := callStartStopMethod("stop", self.components[i].stopMethodName, self.components[i].singleton, self); err != nil {
				return err
			}
		}
	}

	self.Logf(Info, "Stopped %s - version: %s - config file: %s", self.Id, self.Configuration.Version, self.Configuration.FileName)

	return nil
}

func newKernel(id, configFileName string) (*Kernel, error) {

	// Init the application configuration
	conf, err := NewConfiguration(configFileName)

	if err != nil {
		return nil, err
	}

	// TODO: Add a logging structure to the configuration file and configure. Make
	// sure this supports configuring syslog.

	syslogAppender, err := NewSyslogAppender("udp", "127.0.0.1:514", id)
	if err != nil {
		return nil, err
	}

	logger := Logger {
		Prefix: id,
		Appenders: [] Appender{
			LevelFilter(Debug, StdErrAppender()),
			LevelFilter(Debug, syslogAppender),
		},
	}

	kernel := &Kernel{ Components : make(map[string]Component), Configuration : conf }
	kernel.Logger = logger
	kernel.Id = id

	if err = writePidFile(kernel); err != nil {
		return nil, err
	}

	return kernel, nil
}

func writePidFile(kernel *Kernel) error {
	kernel.Pid = os.Getpid()
	pidFile, err := os.Create(kernel.Configuration.PidFile)
	if err != nil {
		return NewStackError("Unable to start kernel - problem creating pid file %s - error: %v", kernel.Configuration.PidFile, err)
	}
	defer pidFile.Close()

	if _, err := pidFile.Write([]byte(strconv.Itoa(kernel.Pid))); err != nil {
		return NewStackError("Unable to start kernel - problem writing pid file %s - error: %v", kernel.Configuration.PidFile, err)
	}

	return nil
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

