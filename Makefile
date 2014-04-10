#
# Copyright 2013, Deft Labs
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at:
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

SHELL := /bin/bash

compile:
	@go build

init.test:
	@mongo --quiet --host 127.0.0.1 --port 28000 test/init_db.js

test: init.test
	@rm -Rf /tmp/dlshared_test.pid
	@go test

init.libs:
	@go get -u github.com/mreiferson/go-httpclient
	@go get -u labix.org/v2/mgo
	@go get -u github.com/daviddengcn/go-ljson-conf
	@go get -u github.com/gorilla/mux
	@go get -u code.google.com/p/go.crypto/bcrypt
	@go get -u github.com/nranchev/go-libGeoIP
	@go get -u github.com/robfig/cron

