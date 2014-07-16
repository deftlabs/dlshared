/**
 * (C) Copyright 2013, Deft Labs
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
	"math"
	"testing"
)

func TestTimeStartOfCurrentDay(t *testing.T) {
	start := TimeStartOfCurrentDay()
	if start == nil { t.Errorf("TestTimeStartOfCurrentDay is broken - nil response") }
	if start.Hour() != 0 { t.Errorf("TestTimeStartOfCurrentDay is broken - hour not zero") }
	if start.Minute() != 0 { t.Errorf("TestTimeStartOfCurrentDay is broken - minute not zero") }
	if start.Second() != 0 { t.Errorf("TestTimeStartOfCurrentDay is broken - second not zero") }
	if start.Nanosecond() != 0 { t.Errorf("TestTimeStartOfCurrentDay is broken - nanosecond not zero") }
}

func TestTimeStartOfNextDay(t *testing.T) {

	today := NowInUtc()
	start := TimeStartOfNextDay()

	if start.Day() == today.Day() { t.Errorf("TestTimeStartOfNextDay is broken - today and tomorrow day is the same") }
	if start == nil { t.Errorf("TestTimeStartOfNextDay is broken - nil response") }
	if start.Hour() != 0 { t.Errorf("TestTimeStartOfNextDay is broken - hour not zero") }
	if start.Minute() != 0 { t.Errorf("TestTimeStartOfNextDay is broken - minute not zero") }
	if start.Second() != 0 { t.Errorf("TestTimeStartOfNextDay is broken - second not zero") }
	if start.Nanosecond() != 0 { t.Errorf("TestTimeStartOfNextDay is broken - nanosecond not zero") }
}

func TestTimeFromMillis(t *testing.T) {
	if time := TimeFromMillis(1402102305274); time == nil { t.Errorf("TestTimeFromMillis is broken - nil response?") }
}

func TestDurationToMillis(t *testing.T) {

	start := time.Now()
	time.Sleep(time.Duration(20) * time.Millisecond)
	elapsed := time.Since(start)

	if DurationToMillis(&elapsed) > 22 { t.Errorf("TestDurationToMillis is broken - received out of range response") }

	longDuration := time.Duration(math.MinInt32) * time.Millisecond
	DurationToMillis(&longDuration)
}

func TestTimeToMillis(t *testing.T) {
	location, err := time.LoadLocation("UTC")

	if err != nil { t.Errorf("TestTimeToMillis is broken - load location error: %v", err) }

	tv := time.Date(2014, 1, 29, 14, 14, 32, 0, location)
	millis := TimeToMillis(&tv)

	if millis !=  1391004872000 { t.Errorf("TestTimeToMillis is broken - expected millis: 1391004872000 - received:  %d", millis) }
}

func TestNowTimeUnixStr(t *testing.T) { if str := NowTimeUnixStr(); len(str) == 0 { t.Errorf("NowTimeUnixStr is broken - result is empty") } }

func TestCurrentTimeInMillis(t *testing.T) { if millis := CurrentTimeInMillis(); millis <= 0 { t.Errorf("CurrentTimeInMillis is broken") } }

func TestCurrentTimeInSeconds(t *testing.T) { if seconds := CurrentTimeInSeconds(); seconds <= 0 { t.Errorf("CurrentTimeInSeconds is broken") } }

