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
	"log"
	"regexp"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
)

// Cloudflare represents a connection to CF API
type Cloudflare struct {
	API            *cf.API
	Zone           string
	RegistryConfig *RegistrySettings
}

// NewCloudFlareAPI returns a Cloudflare object
func NewCloudFlareAPI(apikey, apiuser, zone string, registry *RegistrySettings) (*Cloudflare, error) {
	api, err := cf.New(apikey, apiuser)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to cloudflare: %v", err)
	}
	cf := Cloudflare{
		API:            api,
		Zone:           zone,
		RegistryConfig: registry,
	}
	return &cf, nil
}

// ValidateRegistry compares A records with TXT registry
func (c *Cloudflare) ValidateRegistry() (*DNS, error) {
	var ret DNS
	var registered []Record
	var missingRegistry []Record
	var otherRegistry = make(map[string]RegistryRecord)
	registry := c.getRegistry()
	records := c.getRecords()
	for _, record := range records {
		if _, exists := registry[record.Name]; !exists {
			missingRegistry = append(missingRegistry, record)
		} else {
			record.Registered = true
			registered = append(registered, record)
		}
	}
	for _, content := range registry {
		re, err := regexp.Compile(c.RegistryConfig.Prefix + `.*`)
		if err != nil {
			return nil, fmt.Errorf("Error compiling regex for registry matching: %v", err)
		}
		if !re.Match([]byte(content.WithPrefix)) {
			otherRegistry[content.WithoutPrefix] = content
		}
	}
	ret = DNS{
		ExternalRegistry:  otherRegistry,
		MissingRegistry:   missingRegistry,
		RegisteredRecords: registered,
		Registry:          registry,
	}
	return &ret, nil
}

func (c *Cloudflare) getRecords() []Record {
	var ret []Record
	id, err := c.API.ZoneIDByName(c.Zone)
	if err != nil {
		log.Fatal(err)
	}

	recs, err := c.API.DNSRecords(id, cf.DNSRecord{
		Type: "A",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range recs {
		ret = append(ret, Record{
			Name:   item.Name,
			Type:   item.Type,
			Target: item.Content,
		})
	}
	return ret
}

func (c *Cloudflare) getRegistry() map[string]RegistryRecord {
	ret := make(map[string]RegistryRecord)
	id, err := c.API.ZoneIDByName(c.Zone)
	if err != nil {
		log.Fatal(err)
	}
	recs, err := c.API.DNSRecords(id, cf.DNSRecord{
		Type: "TXT",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range recs {
		record, err := newRegistryRecord(item.Content)
		if err != nil {
			continue
		}
		name := item.Name
		if c.RegistryConfig.Prefix != "" {
			splitName := strings.Split(item.Name, c.RegistryConfig.Prefix)
			name = strings.Join(splitName, "")
			record.WithPrefix = item.Name
		}
		record.WithoutPrefix = name
		ret[name] = record
	}
	return ret
}
