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
	"fmt"
	"time"
	"labix.org/v2/mgo"
)

// Create a new Mongo component from a configuration path. The path passed must be in the following format.
//
// mongodb: {
//     configDb: {
//         mongoUrl: "mongodb://localhost:27017/test",
//         mode: 0,
//         dialTimeoutInMs: 3000,
//         socketTimeoutInMs: 3000,
//         syncTimeoutInMs: 3000,
//         cursorTimeoutInMs: 30000,
//     }
// }
//
// The configPath for this component would be "mongodb.configDb". The path can be any arbitrary set of nested
// json documents (json path). If the path is incorrect, the Start() method will panic when called by the kernel.
//
// All of the params above must be present or
// the Start method will panic. If the componentId or configPath param is nil or empty,
// this method will panic.
func NewMongoFromConfigPath(componentId, configPath string) *Mongo {

	if len(componentId) == 0 {
		panic("When calling NewMongoFromConfigPath you must pass in a non-empty componentId param")
	}

	if len(configPath) == 0 {
		panic("When calling NewMongoFromConfigPath you must pass in a non-empty configPath param")
	}

	return &Mongo{ componentId : componentId, configPath : configPath }
}

// Create a new Mongo component. This method will panic if either of the params are nil or len == 0.
func NewMongo(componentId, mongoUrl string, mode, dialTimeoutInMs, socketTimeoutInMs, syncTimeoutInMs, cursorTimeoutInMs int) *Mongo {

	if len(componentId) == 0 {
		panic("When calling NewMongo you must pass in a non-empty component id")
	}

	if len(mongoUrl) == 0 {
		panic("When calling NewMongo you must pass in a non-empty Mongo url")
	}

	return &Mongo{
		componentId : componentId,
		mongoUrl: mongoUrl,
		mode : mode,
		dialTimeoutInMs : dialTimeoutInMs,
		socketTimeoutInMs : socketTimeoutInMs,
		syncTimeoutInMs : syncTimeoutInMs,
		cursorTimeoutInMs : cursorTimeoutInMs,
	}
}

type Mongo struct {
	kernel *Kernel
	Logger

	configPath string

	componentId string
	mongoUrl string
	mode int
	dialTimeoutInMs int
	socketTimeoutInMs int
	syncTimeoutInMs int
	cursorTimeoutInMs int
	session *mgo.Session
}


// Returns the collection from the session.
func (self *Mongo) Collection(dbName, collectionName string) *mgo.Collection { return self.Db(dbName).C(collectionName) }

// Returns the database from the session.
func (self *Mongo) Db(name string) *mgo.Database { return self.session.DB(name) }

// Returns the session struct.
func (self *Mongo) Session() *mgo.Session { return self.session }

// Returns a clone of the session struct.
func (self *Mongo) SessionClone() *mgo.Session { return self.session.Clone() }

// Returns a copy of the session struct.
func (self *Mongo) SessionCopy() *mgo.Session { return self.session.Clone() }

func (self *Mongo) Start(kernel *Kernel) error {
	self.kernel = kernel
	self.Logger = kernel.Logger

	var err error

	// This is a configuration based creation. Load the config data first.
	if len(self.configPath) > 0 {
		self.mongoUrl = self.kernel.Configuration.String(fmt.Sprintf("%s.%s", self.configPath, "mongoUrl"), "")
		self.mode = self.kernel.Configuration.Int(fmt.Sprintf("%s.%s", self.configPath, "mode"), -1)
		self.dialTimeoutInMs = self.kernel.Configuration.Int(fmt.Sprintf("%s.%s", self.configPath, "dialTimeoutInMs"), -1)
		self.socketTimeoutInMs = self.kernel.Configuration.Int(fmt.Sprintf("%s.%s", self.configPath, "socketTimeoutInMs"), -1)
		self.syncTimeoutInMs = self.kernel.Configuration.Int(fmt.Sprintf("%s.%s", self.configPath, "syncTimeoutInMs"), -1)
		self.cursorTimeoutInMs = self.kernel.Configuration.Int(fmt.Sprintf("%s.%s", self.configPath, "cursorTimeoutInMs"), -1)
	}

	// Validate the params
	if len(self.mongoUrl) == 0 {
		panic(fmt.Sprintf("In Mongo - mongoUrl is not set - componentId: %s", self.componentId))
	}

	if self.dialTimeoutInMs < 0 {
		panic(fmt.Sprintf("In Mongo - dialTimeoutInMs is invalid - value: %d - componentId: %s", self.dialTimeoutInMs, self.componentId))
	}

	if self.socketTimeoutInMs < 0 {
		panic(fmt.Sprintf("In Mongo - socketTimeoutInMs is invalid - value: %d - componentId: %s", self.socketTimeoutInMs, self.componentId))
	}

	if self.syncTimeoutInMs < 0 {
		panic(fmt.Sprintf("In Mongo - syncTimeoutInMs is invalid - value: %d - componentId: %s", self.syncTimeoutInMs, self.componentId))
	}

	if self.cursorTimeoutInMs < 0 {
		panic(fmt.Sprintf("In Mongo - cursorTimeoutInMs is invalid - value: %d - componentId: %s", self.cursorTimeoutInMs, self.componentId))
	}

	if self.mode < 0  || self.mode > 2 {
		panic(fmt.Sprintf("In Mongo - mode is invalid - value: %d - componentId: %s", self.mode, self.componentId))
	}

	// Create the session.
	if self.session, err = mgo.DialWithTimeout(self.mongoUrl, time.Duration(self.dialTimeoutInMs) * time.Millisecond); err != nil {
		return NewStackError("Unable to init Mongo session - component: %s - mongodbUrl: %s", self.componentId, self.mongoUrl)
	}

	// This is annoying, but mgo defines these constants as the restricted "mode" type.
	switch self.mode {
		case 0: self.session.SetMode(mgo.Eventual, true)
		case 1: self.session.SetMode(mgo.Monotonic, true)
		case 2: self.session.SetMode(mgo.Strong, true)
	}

	self.session.SetSocketTimeout(time.Duration(self.socketTimeoutInMs) * time.Millisecond)
	self.session.SetSyncTimeout(time.Duration(self.syncTimeoutInMs) * time.Millisecond)

	return nil
}

// Stop the component. This will close the base session.
func (self *Mongo) Stop(kernel *Kernel) error {

	if self.session != nil {
		self.session.Close()
	}

	return nil
}

