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

type MongoDataSource struct {
	DbName string
	CollectionName string
	Mongo *Mongo
	Logger
}

func ObjectIdHex(objectIdHex string) *bson.ObjectId {
	id := bson.ObjectIdHex(objectIdHex)
	return &id
}

func (self *MongoDataSource) NewObjectId() *bson.ObjectId {
	id := bson.NewObjectId()
	return &id
}

// Insert a document into a collection with the base configured write concern.
func (self *MongoDataSource) Insert(doc interface{}) error {
	if err := self.Mongo.Collection(self.DbName, self.CollectionName).Insert(doc); err != nil {
		return NewStackError("Unable to Insert - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Upsert a document in a collection with the base configured write concern.
func (self *MongoDataSource) Upsert(selector interface{}, change interface{}) error {
	if _, err := self.Mongo.Collection(self.DbName, self.CollectionName).Upsert(selector, change); err != nil {
		return NewStackError("Unable to Upsert - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Upsert a document into a collection with the passed write concern.
func (self *MongoDataSource) UpsertSafe(selector interface{}, change interface{}) error {
	session := self.SessionClone()
	defer session.Close()

	session.SetSafe(self.Mongo.DefaultSafe)

	if _, err := self.Mongo.Collection(self.DbName, self.CollectionName).Upsert(selector, change); err != nil {
		return NewStackError("Unable to Upsert - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Insert a document into a collection with the passed write concern.
func (self *MongoDataSource) InsertSafe(doc interface{}) error {
	session := self.SessionClone()

	defer session.Close()

	session.SetSafe(self.Mongo.DefaultSafe)
	if err := session.DB(self.DbName).C(self.CollectionName).Insert(doc); err != nil {
		return NewStackError("Unable to InsertSafe - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Unset a field with the default safe enabled - uses $unset
func (self *MongoDataSource) UnsetFieldSafe(query interface{}, field string) error {
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
func (self *MongoDataSource) SetFieldsSafe(query interface{}, fieldsDoc interface{}) error {
	session := self.SessionClone()
	defer session.Close()

	session.SetSafe(self.Mongo.DefaultSafe)

	update := &bson.M{ "$set": fieldsDoc }

	if err := self.RemoveNotFoundErr(session.DB(self.DbName).C(self.CollectionName).Update(query, update)); err != nil {
		return NewStackError("Unable to SetFieldsSafe - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Pull from an array call using a "safe" operation. If this is a standalone mongo or a mongos, it will use: WMode: "majority".
// If this is a standalone mongo, it will use: w: 1
func (self *MongoDataSource) PullSafe(query interface{}, fieldsDoc interface{}) error {
	session := self.SessionClone()
	defer session.Close()

	session.SetSafe(self.Mongo.DefaultSafe)

	update := &bson.M{ "$pull": fieldsDoc }

	if err := self.RemoveNotFoundErr(session.DB(self.DbName).C(self.CollectionName).Update(query, update)); err != nil {
		return NewStackError("Unable to PullSafe - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Find the distinct string fields. Do not use this on datasets with a large amount of distinct values or
// you will blow out memory. The selector can be nil.
func (self *MongoDataSource) FindDistinctStrs(selector interface{}, fieldName string) ([]string, error) {
	var result []string
	if err := self.Mongo.Collection(self.DbName, self.CollectionName).Find(selector).Distinct(fieldName, &result); err != nil { return nil, err }

	return result, nil
}

// The caller must close the cursor when done. Use: defer cursor.Clse()
func (self *MongoDataSource) FindManyWithBatchSize(selector interface{}, batchSize int) *mgo.Iter {
	return self.Mongo.Collection(self.DbName, self.CollectionName).Find(selector).Batch(batchSize).Iter()
}

// Push to an array call using a "safe" operation. If this is a standalone mongo or a mongos, it will use: WMode: "majority".
// If this is a standalone mongo, it will use: w: 1
func (self *MongoDataSource) PushSafe(query interface{}, fieldsDoc interface{}) error {
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
func (self *MongoDataSource) SetFieldSafe(query interface{}, field string, value interface{}) error {
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
func (self *MongoDataSource) FindById(id interface{}, result interface{}) error {
	return self.FindOne(&bson.M{ "_id": id }, result)
}

// Returns nil if this is a NOT a document not found error.
func (self *MongoDataSource) RemoveNotFoundErr(err error) error {
	if self.NotFoundErr(err) {
		return nil
	}
	return err
}

// Returns true if this is a document not found error.
func (self *MongoDataSource) NotFoundErr(err error) (bool) {
	return err == mgo.ErrNotFound
}

// Finds one document or returns false.
func (self *MongoDataSource) FindOne(query *bson.M, result interface{}) error {
	return self.Collection().Find(query).One(result)
}

// Delete one document from the collection. If the document is not found, no error is returned.
func (self *MongoDataSource) DeleteOne(selector interface{}) error {
	if err := self.RemoveNotFoundErr(self.Collection().Remove(selector)); err != nil {
		return NewStackError("Unable to DeleteOne - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Delete one or more documents from the collection. If the document(s) is/are not found, no error
// is returned.
func (self *MongoDataSource) Delete(selector interface{}) error {
	if _, err := self.Collection().RemoveAll(selector); err != nil {
		return NewStackError("Unable to Delete - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Ensure a unique, non-sparse index is created. This does not create in the background. This does
// NOT drop duplicates if they exist. Duplicates will cause an error.
func (self *MongoDataSource) EnsureUniqueIndex(fields []string) error {
	if err := self.Collection().EnsureIndex(mgo.Index{
		Key: fields,
		Unique: true,
		DropDups: false,
		Background: false,
		Sparse: false,
	}); err != nil {
		return NewStackError("Unable to EnsureUniqueIndex - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Ensure a non-unique, non-sparse index is created. This does not create in the background.
func (self *MongoDataSource) EnsureIndex(fields []string) error {
	if err := self.Collection().EnsureIndex(mgo.Index{
		Key: fields,
		Unique: false,
		DropDups: false,
		Background: false,
		Sparse: false,
	}); err != nil {
		return NewStackError("Unable to EnsureIndex - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Ensure a non-unique, sparse index is created. This does not create in the background.
func (self *MongoDataSource) EnsureSparseIndex(fields []string) error {
	if err := self.Collection().EnsureIndex(mgo.Index{
		Key: fields,
		Unique: false,
		DropDups: false,
		Background: false,
		Sparse: true,
	}); err != nil {
		return NewStackError("Unable to EnsureSparseIndex - db: %s - collection: %s - error: %v", self.DbName, self.CollectionName, err)
	}

	return nil
}

// Create a capped collection.
func (self *MongoDataSource) CreateCappedCollection(sizeInBytes int) error {
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
func (self *MongoDataSource) EnsureTtlIndex(field string, expireAfterSeconds int) error {
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

func (self *MongoDataSource) Now() *time.Time {
	now := time.Now()
	return &now
}

// Ensure a unique, sparse index is created. This does not create in the background. This does
// NOT drop duplicates if they exist. Duplicates will cause an error.
func (self *MongoDataSource) EnsureUniqueSparseIndex(fields []string) error {
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
func (self *MongoDataSource) Collection() *mgo.Collection { return self.Mongo.Collection(self.DbName, self.CollectionName) }

// Returns the database from the session.
func (self *MongoDataSource) Db() *mgo.Database { return self.Mongo.Db(self.DbName) }

// Returns the session struct.
func (self *MongoDataSource) Session() *mgo.Session { return self.Mongo.session }

// Returns a clone of the session struct.
func (self *MongoDataSource) SessionClone() *mgo.Session { return self.Mongo.session.Clone() }

// Returns a copy of the session struct.
func (self *MongoDataSource) SessionCopy() *mgo.Session { return self.Mongo.session.Clone() }

