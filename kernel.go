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
	"flag"
	"strings"
	"reflect"
	"strconv"
	"syscall"
	"os/signal"
	"math/rand"
)

const (
	nadaStr = "" // This is internal to dlshared.
	commaStr = "," // This is internal to dlshared.

	injectLoggerName = "dlshared.Logger"
	injectLoggerFieldName = "Logger"

	injectConfigurationName = "*dlshared.Configuration"
	injectConfigurationFieldName = "Configuration"

	injectKernelName = "*dlshared.Kernel"
	injectKernelFieldName = "Kernel"

	injectMongoDataSourceName = "dlshared.MongoDataSource"
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

// Returns true if the component is present. This panics if the component id is empty.
func (self *Kernel) HasComponent(componentId string) (found bool) {
	panicIfComponentIdNotSet(componentId)
	_, found = self.Components[componentId]
	return
}

// Access another component. This method will panic if you attempt to reference a
// non-existent component. If the component id has a length of zero, it is also panics.
func (self *Kernel) GetComponent(componentId string) interface{} {

	panicIfComponentIdNotSet(componentId)

	if _, found := self.Components[componentId]; !found {
		panic(fmt.Sprintf("kernel.GetComponent called with an invalid component id: %s", componentId))
	}

	return self.Components[componentId].singleton.(interface{})
}

func panicIfComponentIdNotSet(componentId string) { if len(componentId) == 0 { panic("kernel.GetComponent called with an empty component id") } }

// Register a component with a start and stop methods. This method will panic if a nil component is passed.
func (self *Kernel) AddComponentWithStartStopMethods(componentId string, singleton interface{}, startMethodName, stopMethodName string) {

	if singleton == nil { panic(fmt.Sprintf("Nil component passed to kernel for id: %s", componentId)) }

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
		return fmt.Errorf("The %s method: %s on struct: %s has more than one parameter - you can only pass the Kernel or nothing", methodTypeName, methodName, value.Type())
	}

	// Verify the return type is error
	if methodType.NumOut() == 1 && methodType.Out(0).Name() != "error" {
		return fmt.Errorf("The %s method: %s on struct: %s has an invalid return type - you can return nothing or error", methodTypeName, methodName, value.Type())
	}

	methodInputs := make([]reflect.Value, 0)
	if methodType.NumIn() == 1 { methodInputs = append(methodInputs, reflect.ValueOf(kernel)) }

	returnValues := methodValue.Call(methodInputs)

	// Check to see if there was an error
	if len(returnValues) == 1 {
		err := returnValues[0].Interface()
		if err != nil { return err.(error) }
	}

	return nil
}

// Call this after the kernel has been created and components registered.
func (self *Kernel) Start() error {

	rand.Seed(time.Now().UTC().UnixNano())

	self.Logf(Info, "Starting: %s - version: %s - config file: %s", self.Id, self.Configuration.Version, self.Configuration.FileName)

	if err := self.injectComponents(); err != nil { return err }

	for i := range self.components {
		if len(self.components[i].startMethodName) > 0 {
			if err := callStartStopMethod("start", self.components[i].startMethodName, self.components[i].singleton, self); err != nil {
				return err
			}
		}
	}

	self.Logf(Info, "Started: %s - version: %s - config file: %s ", self.Id, self.Configuration.Version, self.Configuration.FileName)

	return nil
}

func (self *Kernel) injectComponents() error {

	// Loop through the components, look at the variables for tags and automatically do the injection
	// if the tag is set on a field.
	for componentId, component := range self.Components {

		// Get the value of the component and cast.
		componentValue := reflect.ValueOf(component.singleton).Elem()
		componentType := reflect.TypeOf(componentValue.Interface())

		// Loop through the fields.
		fieldCount := componentType.NumField()
		for i := 0; i < fieldCount; i++ {

			structField := componentType.Field(i)
			fieldValue := componentValue.Field(i)

			var mongoDbName string
			var mongoCollectionName string

			// Check to see if the tag is set.
			injectComponentId := structField.Tag.Get("dlinject")

			if len(injectComponentId) == 0 {
				if structField.Type.String() == injectLoggerName && structField.Name == injectLoggerFieldName {
					fieldValue.Set(reflect.ValueOf(self.Logger))
					continue
				}

				if structField.Type.String() == injectConfigurationName && structField.Name == injectConfigurationFieldName {
					fieldValue.Set(reflect.ValueOf(self.Configuration))
					continue
				}

				if structField.Type.String() == injectKernelName && structField.Name == injectKernelFieldName {
					fieldValue.Set(reflect.ValueOf(self))
					continue
				}

				continue
			}

			if structField.Type.String() == injectMongoDataSourceName {
				dataSourceConfig := strings.Split(injectComponentId, commaStr)

				if len(dataSourceConfig) != 3 {
					return NewStackError(	"Unable to inject component: %s - into component: %s - reason: config must be componentId,dbName,collectionName",
											injectComponentId,
											componentId)
				}

				injectComponentId = strings.TrimSpace(dataSourceConfig[0])
				mongoDbName = strings.TrimSpace(dataSourceConfig[1])
				mongoCollectionName = strings.TrimSpace(dataSourceConfig[2])
			}

			// Make sure the component is present.
			injectComponent, found := self.Components[injectComponentId]
			if !found {
				return NewStackError(	"Unable to inject component: %s - into component: %s - reason: %s not found",
										injectComponentId,
										componentId,
										injectComponentId)
			}

			if !fieldValue.CanSet() {
				return NewStackError(	"Unable to inject component: %s - into component: %s - on field: %s - reason: field not exported",
										injectComponentId,
										componentId,
										structField.Name)
			}

			// Check to see if this is a mongo data source component
			if structField.Type.String() == injectMongoDataSourceName {
				fieldValue.Set(reflect.ValueOf(MongoDataSource{	DbName: mongoDbName,
																CollectionName: mongoCollectionName,
																Mongo: injectComponent.singleton.(*Mongo),
																Logger: self.Logger,
				}))

			} else { fieldValue.Set(reflect.ValueOf(injectComponent.singleton)) }
		}
	}

	return nil
}

// Stop the kernel. Call this before exiting.
func (self *Kernel) Stop() error {

	self.Logf(Info, "Stopping: %s - version: %s - config file %s", self.Id, self.Configuration.Version, self.Configuration.FileName)

	for i := len(self.components)-1 ; i >= 0 ; i-- {

		if len(self.components[i].stopMethodName) > 0 {
			if err := callStartStopMethod("stop", self.components[i].stopMethodName, self.components[i].singleton, self); err != nil {
				return err
			}
		}
	}

	self.Logf(Info, "Stopped: %s - version: %s - config file: %s", self.Id, self.Configuration.Version, self.Configuration.FileName)

	return nil
}

func newKernel(id, configFileName string) (*Kernel, error) {

	// Init the application configuration
	conf, err := NewConfiguration(configFileName)

	if err != nil {
		return nil, err
	}

	// Init the logger
	logAppenders, err := configureLogger(id, conf)
	if err != nil {
		return nil, err
	}

	logger := Logger{ Prefix: id, Appenders: logAppenders }

	// Create the kernel
	kernel := &Kernel{ Components : make(map[string]Component), Configuration : conf }
	kernel.Logger = logger
	kernel.Id = id

	if err = writePidFile(kernel); err != nil {
		return nil, err
	}

	return kernel, nil
}

// TODO: Add a logging structure to the configuration file and configure. Make
// sure this supports configuring syslog.
func configureLogger(id string, conf *Configuration) ([]Appender, error) {

	var appenders []Appender
	appenders = append(appenders, LevelFilter(Debug, StdErrAppender()))

	if conf.EnvironmentIs("prod") {
		syslogAppender, err := NewSyslogAppender("", "", id)
		if err != nil {
			return nil, err
		}

		appenders = append(appenders, LevelFilter(Debug, syslogAppender))
	}

	return appenders, nil
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

// This method will load the configuration file, start the kernel and then
// listen for the interrupt.
func RunKernelAndListenForInterrupt(id string, addComponentsFunction func(kernel *Kernel)) error {

	var configFileName string

	flag.StringVar(&configFileName, "config", "configuration.json", "You must pass in a configuration file")
	flag.Parse()

	var kernel *Kernel
	var err error

	if kernel, err = StartKernel(id, configFileName, addComponentsFunction); err != nil {
		return NewStackError("Error starting: %s - exiting - err: %v", id, err)
	}

	if err := kernel.ListenForInterrupt(); err != nil { return NewStackError("Error stopping %s - err: %v", id, err) }

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

