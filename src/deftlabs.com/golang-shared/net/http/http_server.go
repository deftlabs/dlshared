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

package deftlabsnet

import (
	"time"
	"net/http"
	"github.com/gorilla/mux"
	"deftlabs.com/golang-shared/log"
	"deftlabs.com/golang-shared/kernel"
	"deftlabs.com/golang-shared/util"
)

type HttpServerHandlerDef struct {
	Path string
	HandlerFunc http.HandlerFunc
}

type HttpServer struct {
	router *mux.Router
	server *http.Server
	handlerDefs []*HttpServerHandlerDef
	kernel *deftlabskernel.Kernel
	slogger.Logger
	staticFileDir string
	bindAddress string
	port int16
}

func (self *HttpServer) Id() string {
	return "httpServer"
}

func (self *HttpServer) Stop(kernel *deftlabskernel.Kernel) error {
	return nil
}

func (self *HttpServer) Start(kernel *deftlabskernel.Kernel) error {

	self.Logger = kernel.Logger

	self.staticFileDir = kernel.Configuration.String("server.http.staticFileDir", "./static/")
	self.bindAddress = kernel.Configuration.String("server.http.bindAddress", "127.0.0.1")
	self.port = int16(kernel.Configuration.Int("server.http.port", 8080))

	self.kernel = kernel
	self.router = mux.NewRouter()

	if self.handlerDefs != nil {
		for _, handlerDef := range self.handlerDefs {
			self.router.HandleFunc(handlerDef.Path, handlerDef.HandlerFunc)
		}
	}

	self.router.PathPrefix("/").Handler(http.FileServer(http.Dir(self.staticFileDir)))

	http.Handle("/", self.router)

	self.server = &http.Server{
		Addr: deftlabsutil.AssembleHostnameAndPort(self.bindAddress, self.port),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := self.server.ListenAndServe(); err != nil {
			self.Logf(slogger.Error, "Error in listen and serve call - server unpredictable: %v", err)
		}
	}()

	return nil
}

func NewHttpServer(handlerDefs ...*HttpServerHandlerDef) *HttpServer {

	server := &HttpServer{ }

	if handlerDefs != nil {
		for _, def := range handlerDefs {
			server.handlerDefs = append(server.handlerDefs, def)
		}
	}

	return server
}

