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
	"labix.org/v2/mgo/bson"
)

// A metrics relay function that stores count and guage values in mongo. This
// does not store historical values, simply current ones.
type MetricsMongo struct {
	Logger
	DataSource
	mongoComponentName string
	fireAndForget bool
}

type PersistedMetric struct {
	Id string `bson:"_id" json:"id"`

	Name string `bson:"name" json:"name"`
	Source string `bson:"source" json:"source"`
	Type string `bson:"type" json:"type"`

	Value float64 `bson:"value" json:"value"`
	Previous float64 `bson:"previous" json:"previous"`

	Updated *time.Time `bson:"updated" json:"updated"`
	Created *time.Time `bson:"created" json:"created"`
}

// This method returns the change from the current - previous. Usually, only for counters.
func (self *PersistedMetric) Change() float64 { return self.Value - self.Previous }

func NewMetricsMongo(dbName, collectionName, mongoComponentName string, fireAndForget bool) *MetricsMongo {
	return &MetricsMongo{ Logger: Logger{}, DataSource: DataSource{ DbName: dbName, CollectionName: collectionName }, mongoComponentName: mongoComponentName, fireAndForget: fireAndForget }
}

// Assemble the doc id. If there is an error, it is logged here.
func (self *MetricsMongo) assembleDocId(metricName, sourceName string) (string, error) {
	id, err := Md5Hex(fmt.Sprintf("%s-%s-metrics", metricName, sourceName))
	if err != nil {
		self.Logf(Error, "Unable to assemble doc id - metric: %s - source: %s - error: %v", metricName, sourceName, err)
		return nadaStr, err
	}

	return id, nil
}

func (self *MetricsMongo) persistCounter(sourceName string, metric *Metric) {

	docId, err := self.assembleDocId(metric.Name, sourceName)
	if err != nil { return }

	previous, err := self.loadPrevious(docId, metric.Value)
	if err != nil { return }

	selector := &bson.M{ "_id": docId }

	now := self.Now()

	upsert := &bson.M{
		"$setOnInsert": &bson.M{ "name": metric.Name, "source": sourceName, "type": CounterStr, "created": now },
		"$set": &bson.M{ "updated": now, "previous": previous },
		"$inc": &bson.M{ "value": metric.Value },
	}

	if self.fireAndForget { err = self.Upsert(selector, upsert)
	} else { err = self.UpsertSafe(selector, upsert) }

	if err != nil { self.Logf(Error, "Unable to persist counter - source: %s - metric: %s - error: %v", sourceName, metric.Name, err) }
}

// Returns all of the distinct metric names
func (self *MetricsMongo) FindDistinctMetricNames() ([]string, error) { return self.FindDistinctStrs(nil, "name") }

// Returns the cursor for metrics by metric name. The caller must close the cursor when done.
func (self *MetricsMongo) FindMetricsByName(metricName string, batchSize int) *mgo.Iter { return self.FindManyWithBatchSize(&bson.M{ "name": metricName }, batchSize) }

func (self *MetricsMongo) persistGauge(sourceName string, metric *Metric) {

	docId, err := self.assembleDocId(metric.Name, sourceName)
	if err != nil { return }

	previous, err := self.loadPrevious(docId, metric.Value)
	if err != nil { return }

	selector := &bson.M{ "_id": docId }

	now := self.Now()

	upsert := &bson.M{
		"$setOnInsert": &bson.M{ "name": metric.Name, "source": sourceName, "type": CounterStr, "created": now },
		"$set": &bson.M{ "updated": now, "value": metric.Value, "previous": previous },
	}

	if self.fireAndForget { err = self.Upsert(selector, upsert)
	} else { err = self.UpsertSafe(selector, upsert) }

	if err != nil { self.Logf(Error, "Unable to persist counter - source: %s - metric: %s - error: %v", sourceName, metric.Name, err) }
}

// This returns the previous value or the current metric if not found.
func (self *MetricsMongo) loadPrevious(docId string, current float64) (float64, error) {

	previous := current

	previousDoc, err := self.FindById(docId)
	if err != nil {
		self.Logf(Error, "Unable to find previous - error: %v", err)
		return previous, err
	}

	if previousDoc != nil { previous = (*previousDoc)["value"].(float64) }

	return previous, nil
}

func (self *MetricsMongo) FindById(id string) (*bson.M, error) { return self.findOneBy(&bson.M{ "_id": id }) }

func (self *MetricsMongo) findOneBy(query *bson.M) (*bson.M, error) {
	doc := &bson.M{}
	if err := self.FindOne(query, doc); err != nil { return nil, self.RemoveNotFoundErr(err) }
	return doc, nil
}

// This method can be used as the Metrics relay function.
func (self *MetricsMongo) StoreMetricsInMongo(sourceName string, metrics []Metric) {

	for i := range metrics {
		switch metrics[i].Type {
			case Counter: self.persistCounter(sourceName, &metrics[i])
			case Gauge: self.persistGauge(sourceName, &metrics[i])
		}
	}
}

func (self *MetricsMongo) Start(kernel *Kernel) error {

	self.Logger = kernel.Logger
	self.Mongo = kernel.GetComponent(self.mongoComponentName).(*Mongo)

	if err := self.EnsureIndex([]string{ "name" }); err != nil { return err }

	if err := self.EnsureIndex([]string{ "source" }); err != nil { return err }

	if err := self.EnsureIndex([]string{ "name", "source", "value" }); err != nil { return err }

	if err := self.EnsureIndex([]string{ "source", "updated" }); err != nil { return err }

	if err := self.EnsureUniqueIndex([]string{ "name", "source" }); err != nil { return err }

	return nil
}

func (self *MetricsMongo) Stop(kernel *Kernel) error { return nil }

