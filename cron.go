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
	"sync"
	"github.com/robfig/cron"
	"labix.org/v2/mgo/bson"
)

type CronJob interface { Run() }
type CronSchedule interface { Next(time.Time) time.Time }

type cronJobDefinition struct {
	Id string  `bson:"_id"`
	ComponentId string `bson:"componentId"`
	MethodName string `bson:"methodName"`
	Schedule string `bson:"schedule"`
	RequiresDistributedLock bool `bson:"requiresDistributedLock"`
	Audit bool `bson:"audit"`
	Enabled bool `bson:"enabled"`
	Created *time.Time `bson:"created"`
}

// The cron service is a wrapper around robfig's cron library that adds
// audit/tracking data that is stored in MongoDB. For more information on
// the core cron library, see:
//        http://godoc.org/github.com/robfig/cron
//
// The cron service can also be used in conjunction with the DistributedLock
// to ensure that crons only execute in one process in the cluster.
//
// Similar to the wwy the mongo component works, the cron service is configured
// via the configuration json file.
//
//    "cron": {
//        "scheduled": {
//            "mongoComponentId": "MongoDbData",
//
//   		  "definitionDbName": "cron",
//            "definitionCollectionName": "cron.definitions",
//
//   		  "auditDbName": "cron",
//            "auditCollectionName": "cron.audit",
//            "auditTimeoutInSec": 126144000,
//
//            "distributedLockComponentId": "MyDistributedLock",
//
//            "scheduledFunctions": [
//                { "jobId": "testCronJob-Run", "componentId": "testComponentId", "methodName": "Run", "schedule": "0 30 * * * *", "requiresDistributedLock": true, "audit": true, enabled: true },
//                { "jobId": "testCronJob-Test", "componentId": "testComponentId", "methodName": "Test", "schedule": "0 30 * * * *", "requiresDistributedLock": true, "audit": true, enabled: false }
//            ]
//        }
//    }
//
// The configPath for this component would be "cron.scheduled". The path can be any arbitrary set of nested
// json documents (json path). If the path is incorrect, the Start() method will panic when called by the kernel.
// The configuration file currently only supports scheduling methods by component id. You need to register your
// component in the kernel and define the method name as a member of that struct. The method cannot not take
// any params and cannot return any values. The method must also be declared public (i.e., the first character
// must be uppercase). The method name should not have a bracket/parentheses. When defining scheduled methods,
// the job ids must be unique or the service will error on Start. You can disable the db audit for jobs by setting "audit"
// to false in the scheduled method. If you set the auditTimeoutInSec to zero, then it will never timeout the audit/history.
// When the auditTimeoutInSec is greater than zero, it will remove the audit/history from the database after the configured
// number of seconds. Ten years in seconds is ~: 315360000 (leap year etc. not included simply 86,400 * 365 * 10).
//
// If you wish to use other definition options, add the items to the component in your Start method. The cron
// service must be added to the kernel after all required components have been added.
//
// For supported cron expression format options, see: http://godoc.org/github.com/robfig/cron
//
// If you change "enabled" for a scheduled function in the database directly, the app will update after a bit. The component
// polls the db for changes.
type CronSvc struct {
	cron *cron.Cron
	configPath string
	definitionDs *cronDefinitionDs
	auditDs *cronAuditDs
	distributedLock DistributedLock
	Logger
	lock *sync.RWMutex
	cronJobDefinitions map[string]*cronJobDefinition
	stopChannel chan bool
	stopWaitGroup *sync.WaitGroup
	cronJobDefMonitorTicker *time.Ticker
}

func NewCronSvc(configPath string) *CronSvc {
	return &CronSvc{
		cron: cron.New(),
		configPath: configPath,
		definitionDs: &cronDefinitionDs{},
		auditDs: &cronAuditDs{},
		lock : &sync.RWMutex{},
		cronJobDefinitions: make(map[string]*cronJobDefinition),
		stopChannel: make(chan bool),
		stopWaitGroup: new(sync.WaitGroup),
	}
}

func (self *CronSvc) cronJobEnabled(jobId string) bool {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.cronJobDefinitions[jobId].Enabled
}

// Returns true if the cron job has audit enabled. If audit is enabled, a job
// run id is returned. The job run id is used the audit table to link a job start
// stop to the same process/call.
func (self *CronSvc) cronJobAuditEnabled(jobId string) (enabled bool, jobRunId *bson.ObjectId) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	enabled = self.cronJobDefinitions[jobId].Audit
	if enabled { jobRunId = self.auditDs.NewObjectId() }
	return
}

func (self *CronSvc) cronJobRequiresDistributedLock(jobId string) bool {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.cronJobDefinitions[jobId].RequiresDistributedLock
}

// Update the cron job definition. Currently, you can only update enabled/disabled, audit and requires distributed lock.
func (self *CronSvc) updateCronJobDefintion(def *cronJobDefinition) {
	self.lock.Lock()
	defer self.lock.Unlock()

	currentDef, found := self.cronJobDefinitions[def.Id]

	if !found { return }

	if currentDef.Enabled != def.Enabled || currentDef.Audit != def.Audit  || currentDef.RequiresDistributedLock != def.RequiresDistributedLock {
		self.Logf(	Info,
					"Changing cron: %s to enabled: %t - audit: %t - requires distributed lock: %t",
					currentDef.Id,
					def.Enabled,
					def.Audit,
					def.RequiresDistributedLock)
	}

	currentDef.Enabled = def.Enabled
	currentDef.Audit = def.Audit
	currentDef.RequiresDistributedLock = def.RequiresDistributedLock
}

func (self *CronSvc) lookupCronJobDef(jobId string) *cronJobDefinition {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.cronJobDefinitions[jobId]
}

// Add a function. This call adds a wrapper around the function which handles locking
// and auditing (both if enabled).
func (self *CronSvc) addFunc(jobId, schedule string, cmd func()) error {
	return self.cron.AddFunc(schedule, func() {

		if !self.cronJobEnabled(jobId) { return }

		if self.cronJobRequiresDistributedLock(jobId) && !self.distributedLock.HasLock() { return }

		auditEnabled, jobRunId := self.cronJobAuditEnabled(jobId)

		if auditEnabled { self.auditDs.start(self.lookupCronJobDef(jobId), jobRunId, time.Now()) }

		cmd()

		if auditEnabled { self.auditDs.end(self.lookupCronJobDef(jobId), jobRunId, time.Now()) }
	})
}

// Load the cron job configuration from the config file and add the jobs to the cron struct.
func (self *CronSvc) initJobsFromConfig(kernel *Kernel) error {

	self.lock.Lock()
	defer self.lock.Unlock()

	mongoComponentId := kernel.Configuration.StringWithPath(self.configPath, "mongoComponentId", "")

	auditTimeoutInSec := kernel.Configuration.IntWithPath(self.configPath, "auditTimeoutInSec", 31536000) // one year in seconds is the default

	distributedLockComponentId := kernel.Configuration.StringWithPath(self.configPath, "distributedLockComponentId", "")

	self.distributedLock = kernel.GetComponent(distributedLockComponentId).(DistributedLock)

	self.cronJobDefMonitorTicker = time.NewTicker(time.Duration(kernel.Configuration.IntWithPath(self.configPath, "monitorScheduledFreqInSec", 5)) * time.Second)

	if self.distributedLock.TryLock() {
		self.Logf(Info, "Acquired the distributed cron lock: %s", self.distributedLock.LockId())
	}

	self.definitionDs.MongoDataSource = MongoDataSource{
		DbName: kernel.Configuration.StringWithPath(self.configPath, "definitionDbName", ""),
		CollectionName: kernel.Configuration.StringWithPath(self.configPath, "definitionCollectionName", ""),
		Mongo: kernel.GetComponent(mongoComponentId).(*Mongo),
	}

	self.auditDs.MongoDataSource = MongoDataSource{
		DbName: kernel.Configuration.StringWithPath(self.configPath, "auditDbName", ""),
		CollectionName: kernel.Configuration.StringWithPath(self.configPath, "auditCollectionName", ""),
		Mongo: kernel.GetComponent(mongoComponentId).(*Mongo),
	}

	// If the audit timeout is eanbled, ensure the index on the created field.
	if auditTimeoutInSec > 0 { if err := self.auditDs.EnsureTtlIndex("created", auditTimeoutInSec); err != nil { return err } }

	// Add some indexes on the audit table.
	if err := self.auditDs.EnsureIndex([]string{ "jobId", "created" }); err != nil { return err }
	if err := self.auditDs.EnsureIndex([]string{ "jobId", "jobRunId", "created" }); err != nil { return err }
	if err := self.auditDs.EnsureIndex([]string{ "jobRunId" }); err != nil { return err }
	if err := self.auditDs.EnsureIndex([]string{ "created"  }); err != nil { return err }

	scheduledInterface := kernel.Configuration.ListWithPath(self.configPath, "scheduledFunctions", nil)

	if scheduledInterface == nil { return NewStackError("Unable to init configuration - path: %s.%s", self.configPath, "scheduledFunctions") }

	seenJobIds := make(map[string]bool)

	for i := range scheduledInterface {
		if err := self.initJobFromConfig(kernel, seenJobIds, scheduledInterface[i].(map[string]interface{})); err != nil { return err }
	}

	return nil
}

func (self *CronSvc) initJobFromConfig(kernel *Kernel, seenJobIds map[string]bool, scheduledEntry map[string]interface{}) error {

	cronJobDefinition := &cronJobDefinition{
		Id: scheduledEntry["jobId"].(string),
		ComponentId: scheduledEntry["componentId"].(string),
		MethodName: scheduledEntry["methodName"].(string),
		Schedule : scheduledEntry["schedule"].(string),
		RequiresDistributedLock : scheduledEntry["requiresDistributedLock"].(bool),
		Audit: scheduledEntry["audit"].(bool),
		Enabled: scheduledEntry["enabled"].(bool),
	}

	if _, found := seenJobIds[cronJobDefinition.Id]; found { return NewStackError("Duplicate cron job id - jobId: %s", cronJobDefinition.Id)
	} else if !found { seenJobIds[cronJobDefinition.Id] = true }

	// Ensure the component is valid.
	if len(cronJobDefinition.ComponentId) == 0 || !kernel.HasComponent(cronJobDefinition.ComponentId) {
		return NewStackError("Invalid cron component id - jobId: %s - componentId: %s", cronJobDefinition.Id, cronJobDefinition.ComponentId)
	}

	component := kernel.GetComponent(cronJobDefinition.ComponentId)

	// Load the method and verify.
	err, methodValue := GetMethodValueByName(component, cronJobDefinition.MethodName, 0, 0)
	if err != nil { return NewStackError("Invalid method: %s on component: %s", cronJobDefinition.MethodName, cronJobDefinition.ComponentId) }

	// Ensure the definition is in the database
	if err := self.definitionDs.ensure(cronJobDefinition); err != nil { return NewStackError("Unable to persist cron job def: %v", err) }

	self.cronJobDefinitions[cronJobDefinition.Id] = cronJobDefinition

	// Create the method call that can recover from a panic.
	methodCall := func() {
		defer func() {
			if r := recover(); r != nil {
				self.Logf(Error, "WTF - a panic calling cron method: %s - component: %s - problem: %v", cronJobDefinition.MethodName, cronJobDefinition.ComponentId, r)
			}
		}()

		CallNoParamNoReturnValueMethod(component, methodValue)
	}

	// Add the fuction to the cron.
	if err := self.addFunc(cronJobDefinition.Id, cronJobDefinition.Schedule, methodCall); err != nil {
		return NewStackError(	"Problem adding cron function - likely a problem with schedule - cron: %s - method: %s - schedule: %s",
								cronJobDefinition.Id,
								cronJobDefinition.MethodName,
								cronJobDefinition.Schedule)
	}

	return nil
}

func (self *CronSvc) Start(kernel *Kernel) error {

	self.Logger = kernel.Logger

	self.auditDs.Logger = kernel.Logger

	if err := self.initJobsFromConfig(kernel); err != nil { return err }

	self.cron.Start()

	self.stopWaitGroup.Add(1)
	go self.monitorCronJobDefinitions()

	return nil
}

func (self *CronSvc) monitorCronJobDefinitions() {
	defer self.stopWaitGroup.Done()
	for {
		select {
			case <- self.cronJobDefMonitorTicker.C: {
				cronJobDefinitions, err := self.definitionDs.loadAll()
				if err != nil { self.Logf(Error, "Unable to load cron job definitions - err: %v", err); continue }
				for _, cronJobDefinition := range cronJobDefinitions { self.updateCronJobDefintion(cronJobDefinition) }
			}

			case <- self.stopChannel: return
		}
	}
}

func (self *CronSvc) Stop(kernel *Kernel) error {
	self.stopChannel <- true
	self.cron.Stop()
	self.stopWaitGroup.Wait()
	return nil
}

type cronDefinitionDs struct { MongoDataSource }

func (self *cronDefinitionDs) loadAll() ([]*cronJobDefinition, error) {
	var results []*cronJobDefinition
	if err := self.Collection().Find(nil).All(&results); err != nil { return nil, err }
	return results, nil
}

// Ensure the cron job definition is stored in the database.
func (self *cronDefinitionDs) ensure(def *cronJobDefinition) error {
	change := &bson.M{
		"$setOnInsert": &bson.M{ "created": self.Now() },
		"$set": &bson.M{
			"componentId": def.ComponentId,
			"methodName": def.MethodName,
			"schedule": def.Schedule,
			"requiresDistributedLock": def.RequiresDistributedLock,
			"audit": def.Audit,
			"enabled": def.Enabled,
		},
	}

	return self.UpsertSafe(&bson.M{ "_id": def.Id }, change)
}

type cronAuditDs struct {
	MongoDataSource
	Logger
}

func (self *cronAuditDs) start(def *cronJobDefinition, jobRunId *bson.ObjectId, now time.Time) {
	go func() {
		if err := self.InsertSafe(&bson.M{
			"_id": self.NewObjectId(),
			"jobId": def.Id,
			"jobRunId": jobRunId,
			"componentId": def.ComponentId,
			"methodName": def.MethodName,
			"schedule": def.Schedule,
			"requiresDistributedLock": def.RequiresDistributedLock,
			"audit": def.Audit,
			"enabled": def.Enabled,
			"created": now,
			"start": true,
		}); err != nil {
			self.Logf(Error, "Unable to insert cron start audit - id: %s", def.Id)
		}
	}()
}

func (self *cronAuditDs) end(def *cronJobDefinition, jobRunId *bson.ObjectId, now time.Time) {
	go func() {
		if err := self.InsertSafe(&bson.M{
			"_id": self.NewObjectId(),
			"jobId": def.Id,
			"jobRunId": jobRunId,
			"componentId": def.ComponentId,
			"methodName": def.MethodName,
			"schedule": def.Schedule,
			"requiresDistributedLock": def.RequiresDistributedLock,
			"audit": def.Audit,
			"enabled": def.Enabled,
			"created": now,
			"start": false,
		}); err != nil {
			self.Logf(Error, "Unable to insert cron end audit - id: %s", def.Id)
		}
	}()
}

