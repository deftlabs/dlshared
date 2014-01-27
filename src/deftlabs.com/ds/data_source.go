/**
 * (C) Copyright 2014, Deft Labs
 */

package deftlabsds

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"deftlabs.com/log"
)

type DataSource struct {
	DbName string
	CollectionName string
	Mongo *Mongo
	slogger.Logger
}

// Insert a document into a collection with the base configured write concern.
func (self *DataSource) Insert(doc interface{}) error { return self.Mongo.Collection(self.DbName, self.CollectionName).Insert(doc) }

// Insert a document into a collection with the passed write concern.
func (self *DataSource) InsertSafe(doc interface{}, safeMode *mgo.Safe) error {
	session := self.SessionClone()

	defer session.Close()

	session.SetSafe(safeMode)
	return session.DB(self.DbName).C(self.CollectionName).Insert(doc)
}

// Finds one document or returns nil. If the document is not found an error of type mgo.ErrNotFound is
// returned. The result must be a pointer.
func (self *DataSource) FindOne(query *bson.M, result interface{}) error {
	return self.Collection().Find(query).One(result)
}

// Delete one or more documents from the collection. If the document(s) is/are not found, no error
// is returned.
func (self *DataSource) Delete(selector interface{}) error {
	_, err := self.Collection().RemoveAll(selector)
	return err
}

// Returns the collection from the session.
func (self *DataSource) Collection() *mgo.Collection { return self.Mongo.Collection(self.DbName, self.CollectionName) }

// Returns the database from the session.
func (self *DataSource) Db() *mgo.Database { return self.Mongo.Db(self.DbName) }

// Returns the session struct.
func (self *DataSource) Session() *mgo.Session { return self.Mongo.session }

// Returns a clone of the session struct.
func (self *DataSource) SessionClone() *mgo.Session { return self.Mongo.session.Clone() }

// Returns a copy of the session struct.
func (self *DataSource) SessionCopy() *mgo.Session { return self.Mongo.session.Clone() }

