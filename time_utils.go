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
	"syscall"
)

const (
	nanosecondsPerMillisecond float64 = 1000000.0
)

// Returns the minute of the day (0 - 1439) for the time passed. This does not look at seconds or nanoseconds.
func MinuteOfTheDay(checkTime *time.Time) int16 { return int16((checkTime.Hour() * 60) + checkTime.Minute()) }

// Returns the hour of the day for the minute of the day passed. The minute of the day is a value within 0-1439.
// The hour of the day is a value between 0-23. This expects your data to be in the proper range, it will return
// a bad result if bad data is passed.
func MinuteHourOfTheDay(minuteOfTheDay int16) int8 { return int8(minuteOfTheDay / 60) }

// Returns the time configured to the start of the current day (00:00:00 etc). The current
// day is defined using UTC.
func TimeStartOfCurrentDay() *time.Time {
	location, _ := time.LoadLocation("UTC")
	now := NowInUtc()
	startOfCurrentDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	return &startOfCurrentDay
}

// Returns the time configured to the start of the next day (00:00:00 etc). The current
// day is defined using UTC.
func TimeStartOfNextDay() *time.Time {
	location, _ := time.LoadLocation("UTC")
	tomorrow := NowInUtc().Add(time.Duration(24) * time.Hour)
	startOfNextDay := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, location)
	return &startOfNextDay
}

// Returns the time from the milliseconds since epoch. This returns the time in UTC.
func TimeFromMillis(timeInMillis int64) *time.Time {
	theTime := time.Unix(timeInMillis/1000, 0)
	location, _ := time.LoadLocation("UTC")
	theTime = theTime.In(location)
	return &theTime
}

// Convert a time struct to milliseconds since epoch.
func TimeToMillis(tv *time.Time) int64 { return tv.UnixNano() / 1e6 }

func NowInUtc() *time.Time {
	location, _ := time.LoadLocation("UTC")
	time := time.Now().In(location)
	return &time
}

func NowInUtcMinusSeconds(seconds int) *time.Time {
	now := NowInUtc()
	adjusted := now.Add((time.Duration(seconds)*time.Second)*-1)
	return &adjusted
}

// Convert a duration to milliseconds.
func DurationToMillis(dur *time.Duration) int64 { return int64(float64(dur.Nanoseconds()) / nanosecondsPerMillisecond) }

// Get the current time in millis since epoch. Source from stackoverflow:
// http://stackoverflow.com/questions/6161839/go-time-milliseconds
func CurrentTimeInMillis() int64 {
	tv := new(syscall.Timeval)
	syscall.Gettimeofday(tv)
	return (int64(tv.Sec)*1e3 + int64(tv.Usec)/1e3)
}

// Returns the current time in seconds since epoch (i.e., a unix timestamp). Source from stackoverflow:
// http://stackoverflow.com/questions/9539108/obtaining-a-unix-timestamp-in-go-language-current-time-in-seconds-since-epoch
func CurrentTimeInSeconds() int32 { return int32(time.Now().Unix()) }

// NowTimeUnixStr returns the date in unix date string format e.g., Wed Dec 11 19:03:18 EST 2013
func NowTimeUnixStr() string { return time.Now().Format(time.UnixDate) }

