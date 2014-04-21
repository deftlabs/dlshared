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

import "testing"

// Test the CoR service.
func TestCoRSvcSimple(t *testing.T) {

	svc := NewCoRSvc()
	count := 0

	for i := 0; i < 10; i++ { svc.AddNextFunction("testChainId", func(ctx *CoRContext) error { count++; return nil }) }

	if err := svc.RunChain("testChainId"); err != nil { t.Errorf("TestCoRSvc is broken - error returned: %v", err) }
	if count != 10 { t.Errorf("TestCoRSvc is broken - count != 10") }
}

func TestCoRSvcWithError(t *testing.T) {

	svc := NewCoRSvc()
	count := 0

	for i := 0; i < 10; i++ {
		svc.AddNextFunction("testChainId", func(ctx *CoRContext) error {
			count++
			if count == 4 { return NewStackError("simulate a problem") }
			return nil
		})
	}

	if err := svc.RunChain("testChainId"); err == nil { t.Errorf("TestCoRSvcWithError is broken - no error returned") }
	if count != 4 { t.Errorf("TestCoRSvcWithError is broken - count with error check != 5 ") }
}

func TestCoRSvcWithPanic(t *testing.T) {

	svc := NewCoRSvc()
	count := 0

	for i := 0; i < 10; i++ {
		svc.AddNextFunction("testChainId", func(ctx *CoRContext) error {
			count++
			if count == 4 { panic("simulate a panic") }
			return nil
		})
	}

	if err := svc.RunChain("testChainId"); err == nil { t.Errorf("TestCoRSvcWithPanic is broken - no error returned") }
	if count != 4 { t.Errorf("TestCoRSvcWithPanic is broken - count with error check != 5 ") }
}

