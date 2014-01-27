/**
 * (C) Copyright 2014, Deft Labs
 */

package deftlabsds

import (
	"fmt"
	"time"
	"labix.org/v2/mgo"
	"deftlabs.com/log"
	"deftlabs.com/kernel"
)

// Create a new MongoDs component from a configuration path. The path passed must be in the following format.
//
// mongodb: {
//     configDb: {
//         mongoDsUrl: "mongodb://localhost:27017/test",
//         safeMode: 0,
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
func NewMongoDsFromConfigPath(componentId, configPath string) *MongoDs {

	if len(componentId) == 0 {
		panic("When calling NewMongoDsFromConfigPath you must pass in a non-empty componentId param")
	}

	if len(configPath) == 0 {
		panic("When calling NewMongoDsFromConfigPath you must pass in a non-empty configPath param")
	}

	return &MongoDs{ componentId : componentId, configPath : configPath }
}

// Create a new MongoDs component. This method will panic if either of the params are nil or len == 0.
func NewMongoDs(componentId, mongoDsUrl string, safeMode, dialTimeoutInMs, socketTimeoutInMs, syncTimeoutInMs, cursorTimeoutInMs int) *MongoDs {

	if len(componentId) == 0 {
		panic("When calling NewMongoDs you must pass in a non-empty component id")
	}

	if len(mongoDsUrl) == 0 {
		panic("When calling NewMongoDs you must pass in a non-empty MongoDs url")
	}

	return &MongoDs{
		componentId : componentId,
		mongoDsUrl: mongoDsUrl,
		safeMode : safeMode,
		dialTimeoutInMs : dialTimeoutInMs,
		socketTimeoutInMs : socketTimeoutInMs,
		syncTimeoutInMs : syncTimeoutInMs,
		cursorTimeoutInMs : cursorTimeoutInMs,
	}
}

type MongoDs struct {
	kernel *deftlabskernel.Kernel
	slogger.Logger

	configPath string

	componentId string
	mongoDsUrl string
	safeMode int
	dialTimeoutInMs int
	socketTimeoutInMs int
	syncTimeoutInMs int
	cursorTimeoutInMs int
	session *mgo.Session
}

// Returns the collection from the session.
func (self *MongoDs) Collection(dbName, collectionName string) *mgo.Collection { return self.Db(dbName).C(collectionName) }

// Returns the database from the session.
func (self *MongoDs) Db(name string) *mgo.Database { return self.session.DB(name) }

// Returns the session struct.
func (self *MongoDs) Session() *mgo.Session { return self.session }

// Returns a clone of the session struct.
func (self *MongoDs) SessionClone() *mgo.Session { return self.session.Clone() }

// Returns a copy of the session struct.
func (self *MongoDs) SessionCopy() *mgo.Session { return self.session.Clone() }

func (self *MongoDs) Start(kernel *deftlabskernel.Kernel) error {
	self.kernel = kernel
	self.Logger = kernel.Logger

	var err error

	// This is a configuration based creation. Load the config data first.
	if len(self.configPath) > 0 {
		self.mongoDsUrl = self.kernel.Configuration.String(fmt.Sprintf("%s.%s", self.configPath, "mongoDsUrl"), "")
		self.safeMode = self.kernel.Configuration.Int(fmt.Sprintf("%s.%s", self.configPath, "safeMode"), -1)
		self.dialTimeoutInMs = self.kernel.Configuration.Int(fmt.Sprintf("%s.%s", self.configPath, "dialTimeoutInMs"), -1)
		self.socketTimeoutInMs = self.kernel.Configuration.Int(fmt.Sprintf("%s.%s", self.configPath, "socketTimeoutInMs"), -1)
		self.syncTimeoutInMs = self.kernel.Configuration.Int(fmt.Sprintf("%s.%s", self.configPath, "syncTimeoutInMs"), -1)
		self.cursorTimeoutInMs = self.kernel.Configuration.Int(fmt.Sprintf("%s.%s", self.configPath, "cursorTimeoutInMs"), -1)
	}

	// Validate the params
	if len(self.mongoDsUrl) == 0 {
		panic(fmt.Sprintf("In MongoDs - mongoDsUrl is not set - componentId: %s", self.componentId))
	}

	if self.dialTimeoutInMs < 0 {
		panic(fmt.Sprintf("In MongoDs - dialTimeoutInMs is invalid - value: %d - componentId: %s", self.dialTimeoutInMs, self.componentId))
	}

	if self.socketTimeoutInMs < 0 {
		panic(fmt.Sprintf("In MongoDs - socketTimeoutInMs is invalid - value: %d - componentId: %s", self.socketTimeoutInMs, self.componentId))
	}

	if self.syncTimeoutInMs < 0 {
		panic(fmt.Sprintf("In MongoDs - syncTimeoutInMs is invalid - value: %d - componentId: %s", self.syncTimeoutInMs, self.componentId))
	}

	if self.cursorTimeoutInMs < 0 {
		panic(fmt.Sprintf("In MongoDs - cursorTimeoutInMs is invalid - value: %d - componentId: %s", self.cursorTimeoutInMs, self.componentId))
	}

	if self.safeMode < 0  || self.safeMode > 2 {
		panic(fmt.Sprintf("In MongoDs - safeMode is invalid - value: %d - componentId: %s", self.safeMode, self.componentId))
	}

	// Create the session.
	if self.session, err = mgo.DialWithTimeout(self.mongoDsUrl, time.Duration(self.dialTimeoutInMs) * time.Millisecond); err != nil {
		return slogger.NewStackError("Unable to init MongoDs session - component: %s - mongodbUrl: %s", self.componentId, self.mongoDsUrl)
	}

	// This is annoying, but mgo defines these constants as the restricted "mode" type.
	switch self.safeMode {
		case 0: self.session.SetMode(mgo.Eventual, true)
		case 1: self.session.SetMode(mgo.Monotonic, true)
		case 2: self.session.SetMode(mgo.Strong, true)
	}

	self.session.SetSocketTimeout(time.Duration(self.socketTimeoutInMs) * time.Millisecond)
	self.session.SetSyncTimeout(time.Duration(self.syncTimeoutInMs) * time.Millisecond)

	return nil
}

// Stop the component. This will close the base session.
func (self *MongoDs) Stop(kernel *deftlabskernel.Kernel) error {

	if self.session != nil {
		self.session.Close()
	}

	return nil
}

func (self *MongoDs) Id() string { return self.componentId }

