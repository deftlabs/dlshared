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

import "sync"

type CoRFunction func(ctx *CoRContext) error

type CoRContext struct {
	Params map[string]interface{}
	Kernel *Kernel
	Logger
}

// A chain-of-responsibility service implementation in Go. To use, define
// your chain, add your function calls in order and then execute. You
// can pass anything you want into context object. Init this struct
// using the kernel to set the global logger and kernel struct. The
// kernel will be nil if you do not init as component. The logger will
// be a new (i.e., not configured) struct if you do not use the kernel
// to init the component.
type CoRSvc struct {
	chains map[string]*cor
	waitGroup *sync.WaitGroup
	Logger
	Kernel *Kernel
}

func NewCoRSvc() *CoRSvc {
	return &CoRSvc{
		chains: make(map[string]*cor),
		waitGroup: &sync.WaitGroup{},
		Logger: Logger{},
	}
}

func (self *CoRSvc) RunChainWithParams(chainId string, params map[string]interface{}) error {
	return self.RunChainWithContext(chainId, &CoRContext{ Params: params, Kernel: self.Kernel, Logger: self.Logger })
}

func (self *CoRSvc) RunChain(chainId string) error {
	return self.RunChainWithContext(chainId, &CoRContext{ Params: make(map[string]interface{}), Kernel: self.Kernel, Logger: self.Logger })
}

func (self *CoRSvc) RunChainWithContext(chainId string, ctx *CoRContext) error {

	if ctx == nil { ctx = &CoRContext{ Params: make(map[string]interface{}), Kernel: self.Kernel, Logger: self.Logger } }

	cor, found := self.chains[chainId]
	if !found { return NewStackError("Chain not found: %s", chainId) }

	self.waitGroup.Add(1)
	defer self.waitGroup.Done()

	return cor.run(ctx)
}

func (self *CoRSvc) AddNextFunction(chainId string, next CoRFunction) {
	cor, found := self.chains[chainId];
	if !found { cor = self.createChain(chainId) }
	cor.addNextFunction(next)
}

func (self *CoRSvc) createChain(chainId string) *cor {
	cor := &cor{ chainId: chainId }
	self.chains[chainId] = cor
	return cor
}

func (self *CoRSvc) Start() error { return nil }

func (self *CoRSvc) Stop() error {
	// Wait for any running CoRs to exit.
	self.waitGroup.Wait()
	return nil
}

// The chain-of-responsibility struct.
type cor struct {
	chainId string
	functions []CoRFunction
}

// Execute the functions. If a function returns an error,
// the next function is not executed and the error is returned.
func (self *cor) run(ctx *CoRContext) error {
	var panicError interface{}
	for idx, function := range self.functions {
		err := func() error {
			defer func() { if r := recover(); r != nil { panicError = r } }()
			return function(ctx)
		}()

		if err != nil { return err }
		if panicError != nil { return NewStackError("CoR panic - chain: %s - index: %d - err: %v", self.chainId, idx, panicError) }
	}
	return nil
}

// Add functions into the cor.
func (self *cor) addNextFunction(next CoRFunction) *cor {
	self.functions = append(self.functions, next)
	return self
}

