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
	"os"
	"github.com/daviddengcn/go-ljson-conf"
)

type Configuration struct {

	Version string
	PidFile string

	Pid int

	data *ljconf.Conf

	Hostname string
	FileName string
}

func (self *Configuration) String(key string, def string) string {
	return self.data.String(key, def)
}

func (self *Configuration) Int(key string, def int) int {
	return self.data.Int(key, def)
}

func (self *Configuration) Bool(key string, def bool) bool {
	return self.data.Bool(key, def)
}

func (self *Configuration) Float(key string, def float64) float64 {
	return self.data.Float(key, def)
}

func (self *Configuration) StrList(key string, def [] string) []string {
	return self.data.StringList(key, def)
}

func (self *Configuration) IntList(key string, def []int) []int {
	return self.data.IntList(key, def)
}

func NewConfiguration(fileName string) (*Configuration, error) {

	conf := &Configuration{ FileName : fileName }

	var err error
	if conf.data, err = ljconf.Load(fileName); err != nil {
		return nil, err
	}

	conf.PidFile = conf.data.String("pidFile", "")

	if len(conf.PidFile) == 0 {
		return nil, NewStackError("Configuration file error - pidFile not set")
	}

	conf.Version = conf.data.String("version", "@VERSION@")

	conf.Pid = os.Getpid()

	conf.Hostname, err = os.Hostname()
	if err != nil {
		return nil, err
	}

	if len(conf.Version) == 0 || conf.Version == "@VERSION@" {
		return nil, NewStackError("Configuration file error - version not set")
	}

	return conf, nil
}

