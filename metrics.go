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

import "time"

type metricType int8

const (
	Counter metricType = 0
	Gauge metricType = 1
)

type Metric struct {
	Name string
	Type metricType
	Value float64
}

type Metrics struct {
	sourceName string
	quitChannel chan bool
	relayFunc func(string, []Metric)
	relayPeriodInSecs int
	metricChannel chan *Metric
	ticker *time.Ticker
}

func NewMetrics(	sourceName string,
					relayFunc func(string, []Metric),
					relayPeriodInSecs int,
					metricQueueLength int) *Metrics {

	return &Metrics{
		sourceName : sourceName,
		relayFunc : relayFunc,
		relayPeriodInSecs: relayPeriodInSecs,
		quitChannel : make(chan bool),
		metricChannel : make(chan *Metric, metricQueueLength),
	}
}

// Update the gauge value
func (self *Metrics) Gauge(metricName string, value float64) {
	self.metricChannel <- &Metric{
		Name: metricName,
		Type: Gauge,
		Value: value,
	}
}

// Increases the counter by one.
func (self *Metrics) Count(metricName string) {
	self.metricChannel <- &Metric{
		Name: metricName,
		Type: Counter,
		Value: 1,
	}
}

// Increase the counter
func (self *Metrics) CountWithValue(metricName string, value float64) {
	self.metricChannel <- &Metric{
		Name: metricName,
		Type: Counter,
		Value: value,
	}
}

func (self *Metrics) listenForEvents() {

	metrics := make(map[string]*Metric)

    for {
        select {
			case metric := <- self.metricChannel:
				current, found := metrics[metric.Name]
				if !found {
					metrics[metric.Name] = metric
					continue
				}

				if metric.Type == Counter {
					current.Value = current.Value + metric.Value
				} else {
					// This is a gague
					current.Value = metric.Value
				}

			case <- self.ticker.C:
				toRelay := make([]Metric, len(metrics))
				for _, v := range metrics {
					toRelay = append(toRelay, Metric{ Name: v.Name, Type: v.Type, Value: v.Value })
				}

				go self.relayFunc(self.sourceName, toRelay)

			case <- self.quitChannel:
				self.ticker.Stop()
				return
        }
    }
}


func (self *Metrics) Start() error {

	self.ticker = time.NewTicker(time.Duration(self.relayPeriodInSecs) * time.Second)

	go self.listenForEvents()

	return nil
}

func (self *Metrics) Stop() error {
	self.quitChannel <- true
	return nil
}

