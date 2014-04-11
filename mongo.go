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

type MongoConnectionType string

const (
	MongosConnectionType = MongoConnectionType("mongos")
	StandaloneConnectionType = MongoConnectionType("standalone")
	ReplicaSetConnectionType = MongoConnectionType("replicaSet")
)

// Create a new Mongo component from a configuration path. The path passed must be in the following format.
//
//    "mongodb": {
//        "configDb": {
//            "mongoUrl": "mongodb://localhost:27017/test",
//            "mode": "strong",
//            "dialTimeoutInMs": 3000,
//            "socketTimeoutInMs": 3000,
//            "syncTimeoutInMs": 3000,
//            "cursorTimeoutInMs": 30000,
//            "type": "standalone",
//        }
//    }
//
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
func NewMongo(	componentId,
				mongoUrl string,
				connectionType MongoConnectionType,
				mode string,
				dialTimeoutInMs,
				socketTimeoutInMs,
				syncTimeoutInMs,
				cursorTimeoutInMs int) *Mongo {

	if len(componentId) == 0 {
		panic("When calling NewMongo you must pass in a non-empty component id")
	}

	if len(mongoUrl) == 0 {
		panic("When calling NewMongo you must pass in a non-empty Mongo url")
	}

	if len(connectionType) == 0 {
		panic("When calling NewMongo you must pass in a non-empty connection type (standalone | mongos | replicaSet)")
	}

	return &Mongo{
		componentId: componentId,
		mongoUrl: mongoUrl,
		connectionType: MongoConnectionType(connectionType),
		mode: mode,
		dialTimeoutInMs: dialTimeoutInMs,
		socketTimeoutInMs: socketTimeoutInMs,
		syncTimeoutInMs: syncTimeoutInMs,
		cursorTimeoutInMs: cursorTimeoutInMs,
		DefaultSafe: defaultSafe(MongoConnectionType(connectionType)),
	}
}

type Mongo struct {
	kernel *Kernel
	Logger

	configPath string

	componentId string
	mongoUrl string
	mode string
	dialTimeoutInMs int
	socketTimeoutInMs int
	syncTimeoutInMs int
	cursorTimeoutInMs int
	session *mgo.Session
	connectionType MongoConnectionType

	DefaultSafe *mgo.Safe
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

// Returns a default safest mode. If this is a mongos or replica set, WMode: "majority" - if this is
// a standalone instance, w: 1
func defaultSafe(connectionType MongoConnectionType) *mgo.Safe {
	switch connectionType {
		case MongosConnectionType: return &mgo.Safe{ WMode: "majority" }
		case ReplicaSetConnectionType: return &mgo.Safe{ WMode: "majority" }
		case StandaloneConnectionType: return &mgo.Safe{ W: 1 }
		default: panic("Unknown connection type: " + connectionType)
	}

	return nil
}

func (self *Mongo) Start(kernel *Kernel) error {
	self.kernel = kernel
	self.Logger = kernel.Logger

	var err error

	// This is a configuration based creation. Load the config data first.
	if len(self.configPath) > 0 {
		self.mongoUrl = kernel.Configuration.StringWithPath(self.configPath, "mongoUrl", "")

		self.connectionType = MongoConnectionType(kernel.Configuration.StringWithPath(self.configPath, "type", ""))
		self.mode = kernel.Configuration.StringWithPath(self.configPath, "mode", "")
		self.dialTimeoutInMs = kernel.Configuration.IntWithPath(self.configPath, "dialTimeoutInMs", -1)
		self.socketTimeoutInMs = kernel.Configuration.IntWithPath(self.configPath, "socketTimeoutInMs", -1)
		self.syncTimeoutInMs = kernel.Configuration.IntWithPath(self.configPath, "syncTimeoutInMs", -1)
		self.cursorTimeoutInMs = kernel.Configuration.IntWithPath(self.configPath, "cursorTimeoutInMs", -1)
	}

	// Validate the params
	if len(self.mongoUrl) == 0 { panic(fmt.Sprintf("In Mongo - mongoUrl is not set - componentId: %s", self.componentId)) }

	if len(self.connectionType) == 0 { panic(fmt.Sprintf("In Mongo - type is not set - componentId: %s", self.componentId)) }

	if self.dialTimeoutInMs < 0 { panic(fmt.Sprintf("In Mongo - dialTimeoutInMs is invalid - value: %d - componentId: %s", self.dialTimeoutInMs, self.componentId)) }

	if self.socketTimeoutInMs < 0 { panic(fmt.Sprintf("In Mongo - socketTimeoutInMs is invalid - value: %d - componentId: %s", self.socketTimeoutInMs, self.componentId)) }

	if self.syncTimeoutInMs < 0 { panic(fmt.Sprintf("In Mongo - syncTimeoutInMs is invalid - value: %d - componentId: %s", self.syncTimeoutInMs, self.componentId)) }

	if self.cursorTimeoutInMs < 0 { panic(fmt.Sprintf("In Mongo - cursorTimeoutInMs is invalid - value: %d - componentId: %s", self.cursorTimeoutInMs, self.componentId)) }

	if len(self.mode) == 0 { panic(fmt.Sprintf("In Mongo - mode is invalid - value: %s - componentId: %s", self.mode, self.componentId)) }

	if self.mode != "strong" && self.mode != "eventual" && self.mode != "montonic" { panic(fmt.Sprintf("In Mongo - mode is invalid - value: %s - componentId: %s", self.mode, self.componentId)) }

	// Create the session.
	if self.session, err = mgo.DialWithTimeout(self.mongoUrl, time.Duration(self.dialTimeoutInMs) * time.Millisecond); err != nil {
		return NewStackError("Unable to init Mongo session - component: %s - mongodbUrl: %s", self.componentId, self.mongoUrl)
	}

	// This is annoying, but mgo defines these constants as the restricted "mode" type.
	switch self.mode {
		case "eventual": self.session.SetMode(mgo.Eventual, true)
		case "monotonic": self.session.SetMode(mgo.Monotonic, true)
		case "strong": self.session.SetMode(mgo.Strong, true)
	}

	self.DefaultSafe = defaultSafe(self.connectionType)

	self.session.SetSocketTimeout(time.Duration(self.socketTimeoutInMs) * time.Millisecond)
	self.session.SetSyncTimeout(time.Duration(self.syncTimeoutInMs) * time.Millisecond)
	self.session.SetSafe(nil)

	return nil
}

// Stop the component. This will close the base session.
func (self *Mongo) Stop(kernel *Kernel) error {

	if self.session != nil {
		self.session.Close()
	}

	return nil
}

