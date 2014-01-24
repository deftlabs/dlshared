/**
 * (C) Copyright 2014, Deft Labs
 */

package deftlabsds

import (
	"time"
	"labix.org/v2/mgo"
	"deftlabs.com/log"
	"deftlabs.com/kernel"
)

// Create a new MongoDb component. This method will panic if either of the params are nil or len == 0.
func NewMongoDb(componentId, mongoDbUrl string, safeMode, dialTimeoutInMs, socketTimeoutInMs, syncTimeoutInMs, cursorTimeoutInMs  int) *MongoDb {

	if len(componentId) == 0 {
		panic("When calling NewMongoDb you must pass in a non-empty component id")
	}

	if len(mongoDbUrl) == 0 {
		panic("When calling NewMongoDb you must pass in a valid MongoDb url")
	}

	return &MongoDb{
		componentId : componentId,
		mongoDbUrl: mongoDbUrl,
		safeMode : safeMode,
		dialTimeoutInMs : dialTimeoutInMs,
		socketTimeoutInMs : socketTimeoutInMs,
		syncTimeoutInMs : syncTimeoutInMs,
		cursorTimeoutInMs : cursorTimeoutInMs,
	}
}

type MongoDb struct {
	slogger.Logger
	componentId string
	mongoDbUrl string
	safeMode int
	dialTimeoutInMs int
	socketTimeoutInMs int
	syncTimeoutInMs int
	cursorTimeoutInMs int
	session *mgo.Session
}

// Returns the collection from the session.
func (self *MongoDb) Collection(dbName, collectionName string) *mgo.Collection { return self.Db(dbName).C(collectionName) }

// Returns the database from the session.
func (self *MongoDb) Db(name string) *mgo.Database { return self.session.DB(name) }

// Returns the session struct.
func (self *MongoDb) Session() *mgo.Session { return self.session }

// Returns a clone of the session struct.
func (self *MongoDb) SessionClone() *mgo.Session { return self.session.Clone() }

// Returns a copy of the session struct.
func (self *MongoDb) SessionCopy() *mgo.Session { return self.session.Clone() }

func (self *MongoDb) Start(kernel *deftlabskernel.Kernel) error {
	self.Logger = kernel.Logger

	var err error

	self.session, err = mgo.DialWithTimeout(self.mongoDbUrl, time.Duration(self.dialTimeoutInMs) * time.Millisecond)

	if err != nil {
		return slogger.NewStackError("Unable to init MongoDb session - component: %s - mongodbUrl: %s", self.componentId, self.mongoDbUrl)
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

func (self *MongoDb) Stop(kernel *deftlabskernel.Kernel) error {

	if self.session != nil {
		self.session.Close()
	}

	return nil
}

func (self *MongoDb) Id() string { return self.componentId }

