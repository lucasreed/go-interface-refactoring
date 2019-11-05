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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

// Route53 represents a connection to route53 API
type Route53 struct {
	API            *route53.Route53
	Zone           string
	RegistryConfig *RegistrySettings
}

// NewRoute53API returns a Route53 object
func NewRoute53API(zone string, registry *RegistrySettings) *Route53 {
	api := route53.New(session.New())
	r53 := Route53{
		API:            api,
		Zone:           zone,
		RegistryConfig: registry,
	}
	return &r53
}

// ValidateRegistry compares A records with TXT registry
func (r *Route53) ValidateRegistry() (*DNS, error) {
	var ret DNS
	var registered []Record
	var missingRegistry []Record
	var otherRegistry = make(map[string]RegistryRecord)
	registry := r.getRegistry()
	records := r.getRecords()
	for _, record := range records {
		if _, exists := registry[record.Name]; exists {
			re, err := regexp.Compile(r.RegistryConfig.Prefix + `.*`)
			if err != nil {
				return nil, fmt.Errorf("Error compiling regex for registry matching: %v", err)
			}
			if !re.Match([]byte(registry[record.Name].WithPrefix)) {
				missingRegistry = append(missingRegistry, record)
			} else {
				record.Registered = true
				registered = append(registered, record)
			}
		} else {
			missingRegistry = append(missingRegistry, record)
		}
	}
	for _, content := range registry {
		re, err := regexp.Compile(r.RegistryConfig.Prefix + `.*`)
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

func (r *Route53) getRecords() []Record {
	var ret []Record
	hostedZoneByNameInput := route53.ListHostedZonesByNameInput{
		DNSName: aws.String(r.Zone),
	}
	hostedZonesOutput, err := r.API.ListHostedZonesByName(&hostedZoneByNameInput)
	if err != nil {
		log.Fatal(err)
	}

	var id *string
	for _, zone := range hostedZonesOutput.HostedZones {
		if aws.StringValue(zone.Name) == r.Zone+"." {
			id = zone.Id
		}
	}

	resourceRecordSetsInput := route53.ListResourceRecordSetsInput{
		HostedZoneId: id,
	}

	var records []*route53.ResourceRecordSet

	lastPage := false
	for {
		if !lastPage {
			listRecordsResponse, err := r.API.ListResourceRecordSets(&resourceRecordSetsInput)
			if err != nil {
				log.Fatal(err)
			}
			records = append(records, listRecordsResponse.ResourceRecordSets...)
			if !aws.BoolValue(listRecordsResponse.IsTruncated) {
				lastPage = true
			} else {
				resourceRecordSetsInput.SetStartRecordName(aws.StringValue(listRecordsResponse.NextRecordName))
				resourceRecordSetsInput.SetStartRecordType(aws.StringValue(listRecordsResponse.NextRecordType))
			}
		} else {
			break
		}
	}

	for _, item := range records {
		if aws.StringValue(item.Type) != "A" {
			continue
		}
		var target string
		if item.AliasTarget != nil && aws.StringValue(item.AliasTarget.DNSName) != "" {
			t := aws.StringValue(item.AliasTarget.DNSName)
			target = t[:len(t)-1]
		}
		if len(item.ResourceRecords) == 1 {
			target = aws.StringValue(item.ResourceRecords[0].Value)
		}
		name := aws.StringValue(item.Name)
		if name[len(name)-1:] == "." {
			name = name[:len(name)-1]
		}
		ret = append(ret, Record{
			Name:   name,
			Type:   aws.StringValue(item.Type),
			Target: target,
		})
	}
	return ret
}

func (r *Route53) getRegistry() map[string]RegistryRecord {
	ret := make(map[string]RegistryRecord)
	hostedZoneByNameInput := route53.ListHostedZonesByNameInput{
		DNSName: aws.String(r.Zone),
	}
	hostedZonesOutput, err := r.API.ListHostedZonesByName(&hostedZoneByNameInput)
	if err != nil {
		log.Fatal(err)
	}

	var id *string
	for _, zone := range hostedZonesOutput.HostedZones {
		if aws.StringValue(zone.Name) == r.Zone+"." {
			id = zone.Id
		}
	}
	resourceRecordSetsInput := route53.ListResourceRecordSetsInput{
		HostedZoneId: id,
	}

	var records []*route53.ResourceRecordSet

	lastPage := false
	for {
		if !lastPage {
			listRecordsResponse, err := r.API.ListResourceRecordSets(&resourceRecordSetsInput)
			if err != nil {
				log.Fatal(err)
			}
			records = append(records, listRecordsResponse.ResourceRecordSets...)
			if !aws.BoolValue(listRecordsResponse.IsTruncated) {
				lastPage = true
			} else {
				resourceRecordSetsInput.SetStartRecordName(aws.StringValue(listRecordsResponse.NextRecordName))
				resourceRecordSetsInput.SetStartRecordType(aws.StringValue(listRecordsResponse.NextRecordType))
			}
		} else {
			break
		}
	}

	for _, item := range records {
		if aws.StringValue(item.Type) != "TXT" {
			continue
		}
		if len(item.ResourceRecords) != 1 {
			continue
		}
		name := aws.StringValue(item.Name)
		if name[len(name)-1:] == "." {
			name = name[:len(name)-1]
		}
		record, err := newRegistryRecord(aws.StringValue(item.ResourceRecords[0].Value))
		if err != nil {
			continue
		}
		record.WithPrefix = name
		if r.RegistryConfig.Prefix != "" {
			splitName := strings.Split(name, r.RegistryConfig.Prefix)
			name = strings.Join(splitName, "")
		}
		record.WithoutPrefix = name
		ret[name] = record
	}

	return ret
}
