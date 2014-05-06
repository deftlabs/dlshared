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
	"testing"
)

func TestConsumer1(t *testing.T) {

	receiveChannel := make(chan interface{})

	consumer := NewConsumer("TestConsumer1",
							receiveChannel,
							func(msg interface{}) { },
							10,
							0,
							Logger{})


	if err := consumer.Start(); err != nil { t.Errorf("TestConsumer1 Start is broken: %v", err) }

	for idx := 0; idx < 10000; idx++ { receiveChannel <- "test" }


	close(receiveChannel)
	if err := consumer.Stop(); err != nil { t.Errorf("TestConsumer1 Stop is broken: %v", err) }
}

func TestConsumer2(t *testing.T) {

	receiveChannel := make(chan interface{})

	consumer := NewConsumer("TestConsumer2",
							receiveChannel,
							func(msg interface{}) { },
							10,
							0,
							Logger{})

	if err := consumer.Start(); err != nil { t.Errorf("TestConsumer2 Start is broken: %v", err) }

	for idx := 0; idx < 10000; idx++ { receiveChannel <- "test" }

	close(receiveChannel)
	if err := consumer.Stop(); err != nil { t.Errorf("TestConsumer2 Stop is broken: %v", err) }
}

