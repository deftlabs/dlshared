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
	"sync"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type DistributedLock interface {
	Start(kernel *Kernel) error
	Stop(kernel *Kernel) error
	Lock()
	TryLock() bool
	Unlock()
	HasLock() bool
	LockId() string
}

const (
	DistributedLockLocked = 2
	DistributedLockUnlocked = 1

	// These are error log timeout codes:
	DistributedLockErrTimeoutWillOccur = "DISTRIBUTED_LOCK_TIMEOUT_WILL_OCCUR"
	DistributedLockErrNoLockDoc = "DISTRIBUTED_LOCK_NO_LOCK_DOC"
)

// The purpose of the distributed lock is to provide a lock that is available across multiple
// processes/servers. The distributed lock requires a central synchronization point. In
// this impl, MongoDB is the central synchronization server. The attempt to lock method allows
// to try and obtain a lock. If the try is successful, true is returned (and no error). You
// must call Start(kernel) and Stop(kernel) to start/stop the lock service. Given that Go does
// garbage collection, it is theoretically possible that the entire app could lock up for periods
// of time. Extended time in garbage collection could cause a distributed lock to time out. A timeout
// occurs when the Go process is unable to update the database lock heartbeat within a specific period of time.
// When a timeout occurs, there is an election to see who is granted the lock. For this
// reason, we added a HasLock method which returns true if the process still has the distributed lock. Network
// issues and database issues can also lead to lock timeouts. If a lock times out, it does not automatically
// allow callers access to the lock. Nothing happens until there is an election amongst the various processes.
//
// This lock is somewhat confusing in the sense that once a process acquires the global lock, it is pinned to that
// process until the process is killed (or the heartbeat/lock times out/fails). Pinning the lock to a process allows problems
// to be debugged faster.
//
// The MongoDB backed distributed lock uses a similar lock schema as the MongoDB config.locks collection. It looks like:
//
// 	{
// 		"_id" : "balancer",
// 		"process" : "example.net:40000:1350402818:16807",
// 		"state" : 2,
// 		"ts" : ObjectId("507daeedf40e1879df62e5f3"),
// 		"when" : ISODate("2012-10-16T19:01:01.593Z"),
// 		"who" : "example.net:40000:1350402818:16807:Balancer:282475249",
// 	}
//
// The lockId field maps to the _id field. The "ts" field is the "heartbeat" field. The "why" field was removed to support
// the Go Locker interface. When the lock is active, the "state" field is set to 2.
//
type MongoDistributedLock struct {
	lockId string
	hostId string

	currentProcessHasLock bool

	stopWaitGroup *sync.WaitGroup

	heartbeatTicker *time.Ticker
	expireInactiveLockTicker *time.Ticker
	acquireLockTicker *time.Ticker

	localLock *sync.Cond

	lockInUseChannel chan bool

	lockRequestChannel chan bool
	unlockRequestChannel chan bool

	lockAcquiredChannel chan bool
	lockReleasedChannel chan bool

	hasLockRequestChannel chan bool
	hasLockResponseChannel chan bool

	tryLockRequestChannel chan bool
	tryLockResponseChannel chan bool

	acquireLockCompletedChannel chan bool
	heartbeatCompletedChannel chan bool
	expireInactiveLockCompletedChannel chan bool

	stopChannel chan bool

	ds *mongoDistributedLockDs
	historyDs *mongoDistributedLockHistoryDs // May be nil if history timeout is zero or less.

	Logger
}

// Create the distributed lock. If you want to enable history tracking, set the historyTimeoutInSec param
// something greater than zero. Locks and unlocks will be stored for this amount of time in a separate
// collection by appending the suffix "History" to your collection name.
func NewMongoDistributedLock(	lockId,
								mongoComponentId,
								dbName,
								collectionName string,
								heartbeatFreqInSec,
								lockCheckFreqInSec,
								lockTimeoutInSec int64,
								historyTimeoutInSec int) DistributedLock {

	var historyDs *mongoDistributedLockHistoryDs
	if historyTimeoutInSec > 0 {
		historyDs = &mongoDistributedLockHistoryDs{
			lockId: lockId,
			MongoDataSource: MongoDataSource{ DbName: dbName, CollectionName: (collectionName + "History") },
			historyTimeoutInSec: historyTimeoutInSec,
			mongoComponentId: mongoComponentId,
			lockTimeoutInSec: lockTimeoutInSec,
		}
	}

	return &MongoDistributedLock{
		Logger: Logger{},
		ds: &mongoDistributedLockDs{
			lockId: lockId,
			MongoDataSource: MongoDataSource{ DbName: dbName, CollectionName: collectionName },
			historyTimeoutInSec: historyTimeoutInSec,
			mongoComponentId: mongoComponentId,
			lockTimeoutInSec: lockTimeoutInSec,
		},
		historyDs: historyDs,
		lockId: lockId,
		stopWaitGroup: new(sync.WaitGroup),
		stopChannel: make(chan bool),

		heartbeatTicker: time.NewTicker(time.Duration(heartbeatFreqInSec) * time.Second),
		acquireLockTicker: time.NewTicker(time.Duration(lockCheckFreqInSec) * time.Second),
		expireInactiveLockTicker: time.NewTicker(time.Duration(lockCheckFreqInSec) * time.Second),

		localLock: sync.NewCond(new(sync.Mutex)),

		lockRequestChannel: make(chan bool),
		unlockRequestChannel: make(chan bool),

		tryLockRequestChannel: make(chan bool),
		tryLockResponseChannel: make(chan bool),

		lockInUseChannel: make(chan bool),
		lockAcquiredChannel: make(chan bool),
		lockReleasedChannel: make(chan bool),

		acquireLockCompletedChannel: make(chan bool),
		heartbeatCompletedChannel: make(chan bool),
		expireInactiveLockCompletedChannel: make(chan bool),

		hasLockRequestChannel: make(chan bool),
		hasLockResponseChannel: make(chan bool),
	}
}

// Call the lock. This method will block until (if ever) the lock is available.
func (self *MongoDistributedLock) Lock() {
	self.lockRequestChannel <- true
	self.localLock.Wait()
	self.lockInUseChannel <- true
}

func (self *MongoDistributedLock) LockId() string { return self.lockId }

func (self *MongoDistributedLock) TryLock() bool {
	self.tryLockRequestChannel <- true
	return <- self.tryLockResponseChannel
}

// Release the distributed lock. There is no guarantee that the same process
// will be granted access to the distributed lock when it is released.
func (self *MongoDistributedLock) Unlock() { self.unlockRequestChannel <- true }

func (self *MongoDistributedLock) HasLock() bool {
	self.hasLockRequestChannel <- true
	return <- self.hasLockResponseChannel
}

// Try to expire inactive locks.
func (self *MongoDistributedLock) expireInactiveLock() {
	defer func() { self.stopWaitGroup.Done(); self.expireInactiveLockCompletedChannel <- true }()

	if expired, err := self.ds.expireInactiveLock(); err != nil {
		self.Logf(Error, "Problem expiring inactive lock: %v", err)
	} else if expired {
		self.Logf(Error, "Unlocked lock becase heartbeat is expired - lock %s - host: %s", self.lockId, self.hostId)
	}
}

func (self *MongoDistributedLock) acquireLock() {
	defer func() { self.stopWaitGroup.Done(); self.acquireLockCompletedChannel <- true }()

	acquired, err, newLock := self.ds.acquireLock()

	if err != nil { self.Logf(Error, "Problem trying to acquire lock: %s - err: %v", self.lockId, err); return }

	if acquired {
		// Store the history, if configured to do so.
		if self.ds.historyTimeoutInSec > 0 {
			go func() {
				if err := self.historyDs.lockAcquired(newLock); err != nil {
					self.Logf(Error, "Problem adding lock acquired history for lock: %s - err: %v", self.lockId, err)
				}
			}()
		}

		self.lockAcquiredChannel <- true
	}
}

// Send the heartbeat, unless the current process holds the lock, this is a nop.
func (self *MongoDistributedLock) sendHeartbeat() {
	defer func() { self.stopWaitGroup.Done(); self.heartbeatCompletedChannel <- true }()

	if found, err := self.ds.heartbeat(); err != nil { self.Logf(Error, "Problem trying to send heartbeat - lock: %s - err: %v - timeout may occur if failure continues", self.lockId, err)
	} else if !found {
		// Update the status to make sure we release the local lock.
		self.lockReleasedChannel <- true
		self.Logf(Error, "Problem trying to send heartbeat - lock: %s - err-code: %s - lock no longer held locally", self.lockId, DistributedLockErrNoLockDoc)
	}
}

func (self *MongoDistributedLock) listenForEvents() {
	defer self.stopWaitGroup.Done()

	var acquireLockRunning bool
	var heartbeatRunning bool
	var expireInactiveLockRunning bool

	var lockInUse bool
	var haveDistributedLock bool

    for {
        select {
			case <- self.lockRequestChannel: { if haveDistributedLock && !lockInUse { self.localLock.Signal() } }

			case <- self.unlockRequestChannel: { if haveDistributedLock && lockInUse { lockInUse = false; self.localLock.Signal() } }

			case <- self.lockInUseChannel: { lockInUse = true }

			case <- self.lockAcquiredChannel: { haveDistributedLock = true; lockInUse = false; self.localLock.Signal() }

			case <- self.lockReleasedChannel: { haveDistributedLock = false; lockInUse = false }
			case <- self.hasLockRequestChannel: { self.hasLockResponseChannel <- haveDistributedLock }

			case <- self.acquireLockCompletedChannel: { acquireLockRunning = false }
			case <- self.heartbeatCompletedChannel: { heartbeatRunning = false }
			case <- self.expireInactiveLockCompletedChannel: { expireInactiveLockRunning = false }

			case <- self.tryLockRequestChannel: {
				if !haveDistributedLock || lockInUse { self.tryLockResponseChannel <- false
				} else { lockInUse = true; self.tryLockResponseChannel <- true }
			}

			case <- self.acquireLockTicker.C: {
				if !acquireLockRunning && !haveDistributedLock {
					acquireLockRunning = true;
					self.stopWaitGroup.Add(1)
					go self.acquireLock()
				}
			}

			case <- self.heartbeatTicker.C: {
				if !heartbeatRunning && haveDistributedLock {
					heartbeatRunning = true
					self.stopWaitGroup.Add(1)
					go self.sendHeartbeat()
				}
			}

			case <- self.expireInactiveLockTicker.C: {
				if !expireInactiveLockRunning {
					expireInactiveLockRunning = true
					self.stopWaitGroup.Add(1)
					go self.expireInactiveLock()
				}
			}

			case <- self.stopChannel: { return }
        }
    }
}

// This is only called when the process is stopped.
func (self *MongoDistributedLock) releaseLock() {

	found, err := self.ds.releaseLock()

	if err != nil {
		self.Logf(Error, "Unable to release lock: %s - hostId: %s - err: %s - err-code: %s", self.lockId, self.hostId, err, DistributedLockErrTimeoutWillOccur)
	}

	if self.ds.historyTimeoutInSec > 0 && err != nil && found {
		if err := self.historyDs.lockReleased(); err != nil {
			self.Logf(Error, "Unable to log history for release - lock: %s - host id: %s - err: %v", self.lockId, self.hostId, err)
		}
	}
}

func (self *MongoDistributedLock) Start(kernel *Kernel) error {

	hostId := fmt.Sprintf("%s-%s-%d-%s", kernel.Configuration.Hostname, kernel.Id, kernel.Configuration.Pid, kernel.Configuration.Version)
	self.hostId = hostId

	self.ds.hostId = hostId
	self.historyDs.hostId = hostId

	self.localLock.L.Lock()
	self.Logger = kernel.Logger
	self.ds.Mongo = kernel.GetComponent(self.ds.mongoComponentId).(*Mongo)
	self.historyDs.Mongo = kernel.GetComponent(self.historyDs.mongoComponentId).(*Mongo)

	// Ensure the lock definition is in the database.
	if err := self.ds.ensureLockDefinition(); err != nil { return err }

	// If the history table is in use, add indices.
	if self.ds.historyTimeoutInSec > 0 {
		if err := self.historyDs.EnsureTtlIndex("when", self.ds.historyTimeoutInSec); err != nil { return err }
		if err := self.historyDs.EnsureIndex([]string{ "lockId", }); err != nil { return err }
		if err := self.historyDs.EnsureIndex([]string{ "lockId", "when" }); err != nil { return err }
		if err := self.historyDs.EnsureIndex([]string{ "lockId", "when", "state" }); err != nil { return err }
	}

	if err := self.ds.EnsureIndex([]string{ "_id", "state" }); err != nil { return err }
	if err := self.ds.EnsureIndex([]string{ "_id", "state", "who" }); err != nil { return err }
	if err := self.ds.EnsureIndex([]string{ "_id", "ts" }); err != nil { return err }

	go self.listenForEvents()
	self.stopWaitGroup.Add(1)

	return nil
}

func (self *MongoDistributedLock) Stop(kernel *Kernel) error {

	self.stopChannel <- true

	self.heartbeatTicker.Stop()
	self.acquireLockTicker.Stop()
	self.expireInactiveLockTicker.Stop()

	self.releaseLock() // Only if held by the current process
	self.localLock.L.Unlock()

	// Wait for any cleanup
	self.stopWaitGroup.Wait()
	return nil
}

type mongoDistributedLockDs struct {
	MongoDataSource
	historyTimeoutInSec int
	lockId string
	mongoComponentId string
	lockTimeoutInSec int64
	hostId string
}

type mongoDistributedLockHistoryDs struct {
	MongoDataSource
	historyTimeoutInSec int
	lockId string
	mongoComponentId string
	lockTimeoutInSec int64
	hostId string
}

// Try to release the lock. This happens when stop is called on the distributed lock.
func (self *mongoDistributedLockDs) releaseLock() (bool, error) {

	session := self.SessionClone()
	defer session.Close()
	session.SetSafe(self.Mongo.DefaultSafe)

	query := &bson.M{ "_id": self.lockId, "state": DistributedLockLocked, "process": self.hostId }

	toSet := bson.M{"$set": bson.M{ "process": nil, "state": DistributedLockUnlocked, "ts": nil, "when": nil, "who": nil } }

	_, err := self.Mongo.Collection(self.DbName, self.CollectionName).Find(query).Apply(mgo.Change{ Update: toSet, ReturnNew: true }, &bson.M{})

	if err != nil && self.NotFoundErr(err) { return false, nil
	} else if err != nil { return false, err }

	return true, nil
}

// Send the lock heartbeat. This is only called if the lock is held. This returns
// false if the doc is not found.
func (self *mongoDistributedLockDs) heartbeat() (bool, error) {
	session := self.SessionClone()
	defer session.Close()
	session.SetSafe(self.Mongo.DefaultSafe)

	query := &bson.M{ "_id": self.lockId, "state": DistributedLockLocked, "who": self.hostId }
	toSet := &bson.M{"$set": bson.M{ "ts": self.NewObjectId() } }

	_, err := self.Mongo.Collection(self.DbName, self.CollectionName).Find(query).Apply(mgo.Change{ Update: toSet, ReturnNew: true }, &bson.M{})

	if err != nil && self.NotFoundErr(err) { return false, nil }
	if err != nil { return true, nil }

	return true, nil
}

func (self *mongoDistributedLockDs) expireInactiveLock() (bool, error) {
	session := self.SessionClone()
	defer session.Close()
	session.SetSafe(self.Mongo.DefaultSafe)

	tsCheck := bson.NewObjectIdWithTime(time.Now().Add(time.Duration(-1 * self.lockTimeoutInSec) * time.Second))

	query := &bson.M{ "_id": self.lockId, "state": DistributedLockLocked, "ts": &bson.M{ "$lte": tsCheck }}
	toSet := bson.M{ "$set": bson.M{ "process": nil, "state": DistributedLockUnlocked, "ts": nil, "when": nil, "who": nil } }

	_, err := self.Mongo.Collection(self.DbName, self.CollectionName).Find(query).Apply(mgo.Change{ Update: toSet, ReturnNew: true }, &bson.M{})

	if err != nil && self.NotFoundErr(err) { return false, nil }

	if err != nil { return false, err }

	// We expired the lock.
	return true, nil
}

func (self *mongoDistributedLockDs) ensureLockDefinition() error {
	return self.UpsertSafe(	&bson.M{ "_id": self.lockId },
							&bson.M{ "$setOnInsert": &bson.M{ "process": nil, "state": DistributedLockUnlocked, "ts": nil, "when": nil, "who": nil } })
}

// Try to acquire the specific lock.
func (self *mongoDistributedLockDs) acquireLock() (bool, error, *bson.M) {

	session := self.SessionClone()
	defer session.Close()
	session.SetSafe(self.Mongo.DefaultSafe)

	find := &bson.M{ "_id": self.lockId, "state": DistributedLockUnlocked }

	newLockDoc := &bson.M{}

	change := mgo.Change{
		Update: bson.M{"$set": bson.M{ "process": self.hostId, "state": DistributedLockLocked, "ts": self.NewObjectId(), "when": self.Now(), "who": self.hostId }},
		ReturnNew: true,
	}

	_, err := self.Mongo.Collection(self.DbName, self.CollectionName).Find(find).Apply(change, newLockDoc)

	if err != nil && self.NotFoundErr(err) { return false, nil, nil }
	if err != nil { return false, err, nil }

	// Yay! We have the lock.
	return true, nil, newLockDoc
}

// Store the lock acquired info.
func (self *mongoDistributedLockHistoryDs) lockAcquired(doc *bson.M) error {
	return self.Insert(bson.M{
		"_id": self.NewObjectId(),
		"lockId": self.lockId,
		"process": self.hostId,
		"state": DistributedLockLocked,
		"when": (*doc)["when"],
		"who": self.hostId,
	})
}

// Store the lock released info/state.
func (self *mongoDistributedLockHistoryDs) lockReleased() error {
	return self.Insert(&bson.M{
		"_id": self.NewObjectId(),
		"lockId": self.lockId,
		"process": self.hostId,
		"state": DistributedLockUnlocked,
		"when": self.Now(),
		"who": self.hostId,
	})
}

