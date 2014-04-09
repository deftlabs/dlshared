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

// This is a thin wrapper around libgeo, which is built on top of MaxMind's geo databases. Currently
// this only supports IPV4 addresses.
package dlshared

import "github.com/nranchev/go-libGeoIP"

// A wrapper around the libgeo struct. This adds json tags and provides
// abstraction against geo impls.
type GeoLocation struct {
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName"`
	Region string `json:"region"`
	City string `json:"city"`
	PostalCode string `json:"postalCode"`
	Latitude float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

func convertLibGeoLocation(loc *libgeo.Location) *GeoLocation {
	if loc == nil { return nil }
	return &GeoLocation{
		CountryCode: loc.CountryCode,
		CountryName: loc.CountryName,
		Region: loc.Region,
		City: loc.City,
		PostalCode: loc.PostalCode,
		Latitude: loc.Latitude,
		Longitude: loc.Longitude,
	}
}

type GeoLocationSvc interface {
	LocationByIp(ipAddress string) *GeoLocation
	Start() error
	Stop() error
}

type MaxMindLocationSvc struct {
	geoIp *libgeo.GeoIP
	dbFile string
}

func NewMaxMindLocationSvc(dbFile string) GeoLocationSvc { return &MaxMindLocationSvc{ dbFile: dbFile } }

// This method returns nil if not found.
func (self *MaxMindLocationSvc) LocationByIp(ipAddress string) *GeoLocation { return convertLibGeoLocation(self.geoIp.GetLocationByIP(ipAddress)) }

func (self *MaxMindLocationSvc) Start() error {
	var err error
	self.geoIp, err = libgeo.Load(self.dbFile)
	if err != nil { return err }

	return nil
}

func (self *MaxMindLocationSvc) Stop() error { return nil }

