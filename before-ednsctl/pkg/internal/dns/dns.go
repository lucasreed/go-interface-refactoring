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
	"fmt"
	"strings"
)

// RegistrySettings represents the values external-dns may add to registry TXT records
type RegistrySettings struct {
	Owner  string
	Prefix string
}

// RegistryRecord represents a single TXT registry record
type RegistryRecord struct {
	Heritage      string
	Owner         string
	Resource      string
	WithPrefix    string
	WithoutPrefix string
}

// Record represents an A record in the zone
type Record struct {
	Name       string
	Type       string
	Registered bool
	Target     string
}

// DNS represents every record and registry item for the given provider
// As well as items that belong to a different registry
type DNS struct {
	ExternalRegistry  map[string]RegistryRecord
	MissingRegistry   []Record
	RegisteredRecords []Record
	Registry          map[string]RegistryRecord
}

func newRegistryRecord(content string) (RegistryRecord, error) {
	record := RegistryRecord{}
	s := strings.Split(content[1:len(content)-1], ",")
	if len(s) == 0 {
		return RegistryRecord{}, fmt.Errorf("This record does not appear to be a TXT registry record. Content: %s", content)
	}
	for _, item := range s {
		i := strings.Split(item, "=")
		switch i[0] {
		case "heritage":
			{
				record.Heritage = i[1]
			}
		case "external-dns/owner":
			{
				record.Owner = i[1]
			}
		case "external-dns/resource":
			{
				record.Resource = i[1]
			}
		default:
			{
				return RegistryRecord{}, fmt.Errorf("This record does not appear to be a TXT registry record. Content: %s", content)
			}
		}
	}
	return record, nil
}
