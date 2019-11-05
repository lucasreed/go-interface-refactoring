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
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"google.golang.org/api/dns/v1"
	clouddns "google.golang.org/api/dns/v1"
)

// CloudDNS represents a connection to CloudDNS API
type CloudDNS struct {
	API             *clouddns.Service
	ManagedZoneName string
	Zone            string
	RegistryConfig  *RegistrySettings
	Project         string
}

// NewCloudDNSAPI returns a CloudDNS object
func NewCloudDNSAPI(project, managedZone, zone string, registry *RegistrySettings) (*CloudDNS, error) {
	ctx := context.Background()
	dnsService, err := dns.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to clouddns: %v", err)
	}
	cdns := CloudDNS{
		API:             dnsService,
		ManagedZoneName: managedZone,
		Zone:            zone,
		RegistryConfig:  registry,
		Project:         project,
	}
	return &cdns, nil
}

// ValidateRegistry compares A records with TXT registry
func (c *CloudDNS) ValidateRegistry() (*DNS, error) {
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

func (c *CloudDNS) getRecords() []Record {
	var ret []Record
	recs, err := c.API.ResourceRecordSets.List(c.Project, c.ManagedZoneName).Do()
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range recs.Rrsets {
		if item.Type != "A" {
			continue
		}
		ret = append(ret, Record{
			Name:   item.Name,
			Type:   item.Type,
			Target: item.Rrdatas[0],
		})
	}
	return ret
}

func (c *CloudDNS) getRegistry() map[string]RegistryRecord {
	ret := make(map[string]RegistryRecord)
	recs, err := c.API.ResourceRecordSets.List(c.Project, c.ManagedZoneName).Do()
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range recs.Rrsets {
		if item.Type != "TXT" {
			continue
		}
		record, err := newRegistryRecord(item.Rrdatas[0])
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
