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

// Test the max mind location service. Note: If you uprade the test database,
// it will likely break components of this test.
func TestMaxMindLocationSvc(t *testing.T) {

	locationSvc := NewMaxMindLocationSvc("test/GeoLiteCity.dat")

	err := locationSvc.Start()

	if err != nil { t.Errorf("TestMaxMindLocationSvc is broken - unable to load test db: %v", err); return }

	loc := locationSvc.LocationByIp("64.27.101.155")

	if loc == nil { t.Errorf("TestMaxMindLocationSvc is broken - unable to find address"); return }

	if loc.CountryCode != "US" { t.Errorf("TestMaxMindLocationSvc is broken - bad country code") }
	if loc.CountryName != "United States" { t.Errorf("TestMaxMindLocationSvc is broken - bad country name") }
	if loc.Region != "NJ" { t.Errorf("TestMaxMindLocationSvc is broken - bad region") }
	if loc.City != "Jersey City" { t.Errorf("TestMaxMindLocationSvc is broken - bad city") }
	if loc.PostalCode != "07302" { t.Errorf("TestMaxMindLocationSvc is broken - bad postal code") }
	if loc.Latitude != 40.7209  { t.Errorf("TestMaxMindLocationSvc is broken - bad latitude") }
	if loc.Longitude != -74.0468  { t.Errorf("TestMaxMindLocationSvc is broken - bad longitude") }
}


