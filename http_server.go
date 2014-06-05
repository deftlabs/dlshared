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
	"fmt"
	"sync"
	"time"
	"net"
	"strings"
	"net/http"
	"github.com/gorilla/mux"
)

type HttpServerHandlerDef struct {
	Path string
	HandlerFunc http.HandlerFunc
}

type HttpServer struct {
	router *mux.Router
	server *http.Server
	handlerDefs []*HttpServerHandlerDef
	kernel *Kernel
	Logger
	listener net.Listener
}

func (self *HttpServer) Id() string { return "httpServer" }

func (self *HttpServer) Stop(kernel *Kernel) error {

	if self.listener != nil { if err := self.listener.Close(); err != nil { return err } }
	return nil
}

func (self *HttpServer) Start(kernel *Kernel) error {

	// TODO: Add access logging

	self.Logger = kernel.Logger

	staticFileDir := kernel.Configuration.String("server.http.staticFileDir", "./static/")
	bindAddress := kernel.Configuration.String("server.http.bindAddress", "127.0.0.1")
	port := int16(kernel.Configuration.Int("server.http.port", 8080))

	self.kernel = kernel
	self.router = mux.NewRouter()

	if self.handlerDefs != nil { for _, handlerDef := range self.handlerDefs { self.router.HandleFunc(handlerDef.Path, handlerDef.HandlerFunc) } }

	self.router.PathPrefix("/").Handler(http.FileServer(http.Dir(staticFileDir)))

	var err error

	self.listener, err = net.Listen("tcp", AssembleHostnameAndPort(bindAddress, port))
	if err != nil { return NewStackError(fmt.Sprintf("Unable to bind listener - address: %s - port: %d - err: %v", bindAddress, port, err)) }

	self.server = &http.Server{
		Addr: AssembleHostnameAndPort(bindAddress, port),
		Handler: self.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	var startWaitGroup sync.WaitGroup
	startWaitGroup.Add(1)

	go func() {
		startWaitGroup.Done()
		if err = self.server.Serve(self.listener); err != nil {
			if !strings.Contains(err.Error(), "closed") {
				panic(fmt.Sprintf("Error in serve call - server unpredictable: %v", err))
			}
		}
	}()

	// Wait for the goroutine to be allocated before moving on. This is a hack that does
	// no really solve the problem. Ideally, listen and serve would have a notification/callback
	// of some sort so that we know the server is initialized and running.
	startWaitGroup.Wait()

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

