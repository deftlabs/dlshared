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
	//"fmt"
	"testing"
)

type testKernelInjectStruct1 struct {
	Logger
	Configuration *Configuration
	Struct2 *testKernelInjectStruct2 `dlinject:"testKernelInjectStruct2"`
}

type testKernelInjectStruct2 struct { methodCalled bool }
func (self *testKernelInjectStruct2) call() { self.methodCalled = true }

func TestKernelInject(t *testing.T) {

	kernel, err := StartKernel("kernelInject", testConfigFileName, func(kernel *Kernel) {
		kernel.AddComponent("testKernelInjectStruct1", &testKernelInjectStruct1{})
		kernel.AddComponent("testKernelInjectStruct2", &testKernelInjectStruct2{})
	})

	if err != nil { t.Errorf("TestKernelInject start kernel is broken:", err); return }

	struct1 := kernel.GetComponent("testKernelInjectStruct1").(*testKernelInjectStruct1)

	if struct1.Struct2 == nil { t.Errorf("TestKernelInject is broken - component not injected"); return }

	struct1.Struct2.call()

	if !struct1.Struct2.methodCalled { t.Errorf("TestKernelInject is broken - method not called on struct"); return }

	if struct1.Logf == nil { t.Errorf("TestKernelInject is broken - logger not injected"); return }

	struct1.Logf(Debug, "This is a test - don't panic")

	if struct1.Configuration == nil { t.Errorf("TestKernelInject is broken - configuration not injected"); return }

	if val := struct1.Configuration.String("environment", ""); len(val) == 0 { t.Errorf("TestKernelInject is broken - configuration is not working"); }

	if err := kernel.Stop(); err != nil { t.Errorf("TestKernelInject stop kernel is broken:", err) }

}

