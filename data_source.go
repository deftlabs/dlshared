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
	"time"
	"strings"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type DataSource struct {
	DbName string
	CollectionName string
	Mongo *Mongo
	Logger
}

func (self *DataSource) NewObjectId() *bson.ObjectId {
	id := bson.NewObjectId()
	return &id
}

// Insert a document into a collection with the base configured write concern.
func (self *DataSource) Insert(doc interface{}) error {
	if err := self.Mongo.Collection(self.DbName, self.CollectionName).Insert(doc); err != nil {
		return NewStackError("Unable to Insert - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Upsert a document in a collection with the base configured write concern.
func (self *DataSource) Upsert(selector interface{}, change interface{}) error {
	if _, err := self.Mongo.Collection(self.DbName, self.CollectionName).Upsert(selector, change); err != nil {
		return NewStackError("Unable to Upsert - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Insert a document into a collection with the passed write concern.
func (self *DataSource) InsertSafe(doc interface{}) error {
	session := self.SessionClone()

	defer session.Close()

	session.SetSafe(self.Mongo.DefaultSafe)
	if err := session.DB(self.DbName).C(self.CollectionName).Insert(doc); err != nil {
		return NewStackError("Unable to InsertSafe - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Unset a field with the default safe enabled - uses $unset
func (self *DataSource) UnsetFieldSafe(query interface{}, field string) error {
	session := self.SessionClone()
	defer session.Close()

	session.SetSafe(self.Mongo.DefaultSafe)

	update := &bson.M{ "$unset": &bson.M{ field: nadaStr } }

	if err := self.RemoveNotFoundErr(session.DB(self.DbName).C(self.CollectionName).Update(query, update)); err != nil {
		return NewStackError("Unable to UnsetFieldSafe - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Set fields using a "safe" operation. If this is a standalone mongo or a mongos, it will use: WMode: "majority".
// If this is a standalone mongo, it will use: w: 1
func (self *DataSource) SetFieldsSafe(query interface{}, fieldsDoc interface{}) error {
	session := self.SessionClone()
	defer session.Close()

	session.SetSafe(self.Mongo.DefaultSafe)

	update := &bson.M{ "$set": fieldsDoc }

	if err := self.RemoveNotFoundErr(session.DB(self.DbName).C(self.CollectionName).Update(query, update)); err != nil {
		return NewStackError("Unable to SetFieldsSafe - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Add to set call using a "safe" operation. If this is a standalone mongo or a mongos, it will use: WMode: "majority".
// If this is a standalone mongo, it will use: w: 1
func (self *DataSource) PushSafe(query interface{}, fieldsDoc interface{}) error {
	session := self.SessionClone()
	defer session.Close()

	session.SetSafe(self.Mongo.DefaultSafe)

	update := &bson.M{ "$push": fieldsDoc }

	if err := self.RemoveNotFoundErr(session.DB(self.DbName).C(self.CollectionName).Update(query, update)); err != nil {
		return NewStackError("Unable to PushSafe - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Set a property using a "safe" operation. If this is a standalone mongo or a mongos, it will use: WMode: "majority".
// If this is a standalone mongo, it will use: w: 1
func (self *DataSource) SetFieldSafe(query interface{}, field string, value interface{}) error {
	session := self.SessionClone()
	defer session.Close()

	session.SetSafe(self.Mongo.DefaultSafe)

	update := &bson.M{ "$set": &bson.M{ field: value } }

	if err := self.RemoveNotFoundErr(session.DB(self.DbName).C(self.CollectionName).Update(query, update)); err != nil {
		return NewStackError("Unable to SetFieldSafe - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Find by the _id. Returns false if not found.
func (self *DataSource) FindById(id interface{}, result interface{}) error {
	return self.FindOne(&bson.M{ "_id": id }, result)
}

// Returns nil if this is a NOT a document not found error.
func (self *DataSource) RemoveNotFoundErr(err error) error {
	if self.NotFoundErr(err) {
		return nil
	}
	return err
}

// Returns true if this is a document not found error.
func (self *DataSource) NotFoundErr(err error) (bool) {
	return err == mgo.ErrNotFound
}

// Finds one document or returns false.
func (self *DataSource) FindOne(query *bson.M, result interface{}) error {
	return self.Collection().Find(query).One(result)
}

// Delete one document from the collection. If the document is not found, no error is returned.
func (self *DataSource) DeleteOne(selector interface{}) error {
	if err := self.RemoveNotFoundErr(self.Collection().Remove(selector)); err != nil {
		return NewStackError("Unable to DeleteOne - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Delete one or more documents from the collection. If the document(s) is/are not found, no error
// is returned.
func (self *DataSource) Delete(selector interface{}) error {
	if _, err := self.Collection().RemoveAll(selector); err != nil {
		return NewStackError("Unable to Delete - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Ensure a unique, non-sparse index is created. This does not create in the background. This does
// NOT drop duplicates if they exist. Duplicates will cause an error.
func (self *DataSource) EnsureUniqueIndex(fields []string) error {
	if err := self.Collection().EnsureIndex(mgo.Index{
		Key: fields,
		Unique: true,
		DropDups: true,
		Background: false,
		Sparse: false,
	}); err != nil {
		return NewStackError("Unable to EnsureUniqueIndex - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Ensure a non-unique, non-sparse index is created. This does not create in the background.
func (self *DataSource) EnsureIndex(fields []string) error {
	if err := self.Collection().EnsureIndex(mgo.Index{
		Key: fields,
		Unique: false,
		DropDups: true,
		Background: false,
		Sparse: false,
	}); err != nil {
		return NewStackError("Unable to EnsureIndex - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Ensure a non-unique, sparse index is created. This does not create in the background.
func (self *DataSource) EnsureSparseIndex(fields []string) error {
	if err := self.Collection().EnsureIndex(mgo.Index{
		Key: fields,
		Unique: false,
		DropDups: true,
		Background: false,
		Sparse: true,
	}); err != nil {
		return NewStackError("Unable to EnsureSparseIndex - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Create a capped collection.
func (self *DataSource) CreateCappedCollection(sizeInBytes int) error {
	if err := self.Collection().Create(&mgo.CollectionInfo{ DisableIdIndex: false, ForceIdIndex: true, Capped: true, MaxBytes: sizeInBytes }); err != nil {

		// This is a bit of a hack, but the error returned does not provide any codes.
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "already") || strings.Contains(msg, "exists") { return nil }

		return NewStackError("Unable to CreateCappedCollection - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)

	} else {
		return nil
	}
}

// Create a ttl index.
func (self *DataSource) EnsureTtlIndex(field string, expireAfterSeconds int) error {
	if err := self.Collection().EnsureIndex(mgo.Index{
		Key: []string{ field },
		Unique: false,
		DropDups: false,
		Background: false,
		Sparse: false,
		ExpireAfter: time.Duration(expireAfterSeconds) * time.Second,
	}); err != nil {
		return NewStackError("Unable to EnsureTtlIndex - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

func (self *DataSource) Now() *time.Time {
	now := time.Now()
	return &now
}

// Ensure a unique, sparse index is created. This does not create in the background. This does
// NOT drop duplicates if they exist. Duplicates will cause an error.
func (self *DataSource) EnsureUniqueSparseIndex(fields []string) error {
	if err := self.Collection().EnsureIndex(mgo.Index{
		Key: fields,
		Unique: true,
		DropDups: false,
		Background: false,
		Sparse: true,
	}); err != nil {
		return NewStackError("Unable to EnsureUniqueSparseIndex - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
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

