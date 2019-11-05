// Copyright © 2019Luke Reed
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dns

import (
	"strings"
)

// RegistryRecord represents a single TXT registry record
type RegistryRecord struct {
	Heritage string
	Name     string
	Owner    string
	Prefix   string
	Resource string
	// RegisteredRecord Record
}

// Record represents an A record in the zone
type Record struct {
	Name       string
	Registered bool
	Target     string
	Type       string
}

// API abstracts the functions that must be present in a DNS Provider
type API interface {
	GetRegistry() map[string]map[string]string
	GetRecords() map[string]map[string]string
}

// ParseRegistry takes registry data from a provider and returns
// a map of RegistryRecords
func ParseRegistry(api API) map[string]RegistryRecord {
	ret := make(map[string]RegistryRecord)
	rawRegistry := api.GetRegistry()
	for hostname, dataMap := range rawRegistry {
		name := removeTrailingDot(hostname)
		ret[name] = createRegistryRecordFromMap(dataMap)
	}
	return ret
}

// ParseRecords takes dns data from a provider and returns
// a map of Records
func ParseRecords(api API) map[string]Record {
	ret := make(map[string]Record)
	rawRecords := api.GetRecords()
	for hostname, dataMap := range rawRecords {
		name := removeTrailingDot(hostname)
		ret[name] = createRecordFromMap(dataMap)
	}
	return ret
}

func createRegistryRecordFromMap(regMap map[string]string) RegistryRecord {
	var ret RegistryRecord
	for key, value := range regMap {
		switch strings.ToLower(key) {
		case "heritage":
			{
				ret.Heritage = value
			}
		case "name":
			{
				ret.Name = removeTrailingDot(value)
			}
		case "owner":
			{
				ret.Owner = value
			}
		case "prefix":
			{
				ret.Prefix = value
			}
		case "resource":
			{
				ret.Resource = value
			}
		default:
			{
				continue
			}
		}
	}
	return ret
}

func createRecordFromMap(regMap map[string]string) Record {
	var ret Record
	for key, value := range regMap {
		switch strings.ToLower(key) {
		case "name":
			{
				ret.Name = removeTrailingDot(value)
			}
		case "target":
			{
				ret.Target = value
			}
		case "type":
			{
				ret.Type = value
			}
		default:
			{
				continue
			}
		}
	}
	return ret
}

func removeTrailingDot(name string) string {
	if name[len(name)-1:] == "." {
		name = name[:len(name)-1]
	}
	return name
}
