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

package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lucasreed/go-interface-refactoring/before-ednsctl/pkg/internal/dns"
	"github.com/lucasreed/go-interface-refactoring/before-ednsctl/pkg/internal/kube"
)

// TXTRecord represents a record to be added to the DNS Zone
type TXTRecord struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

type TXTRecords struct {
	Records []TXTRecord `json:"records"`
}

// RunCF starts the main logic in cloudflare
func RunCF(apiKey, apiUser, dnsZone, txtOwner, txtPrefix string, ignoredSubdomains []string) {
	k := kube.New(dnsZone, ignoredSubdomains)
	registry := dns.RegistrySettings{
		Owner:  txtOwner,
		Prefix: txtPrefix,
	}
	client, err := dns.NewCloudFlareAPI(apiKey, apiUser, dnsZone, &registry)
	if err != nil {
		log.Fatal(err)
	}
	dns, _ := client.ValidateRegistry()
	hosts := k.GetHosts()
	compareAndOutput(dns, hosts, k.ValidTargets, txtOwner, txtPrefix)
}

// RunGCP starts the main logic in gcp
func RunGCP(project, managedZone, dnsZone, txtOwner, txtPrefix string, ignoredSubdomains []string) {
	k := kube.New(dnsZone, ignoredSubdomains)
	registry := dns.RegistrySettings{
		Owner:  txtOwner,
		Prefix: txtPrefix,
	}
	client, err := dns.NewCloudDNSAPI(project, managedZone, dnsZone, &registry)
	if err != nil {
		log.Fatal(err)
	}
	dns, err := client.ValidateRegistry()
	if err != nil {
		log.Fatal(err)
	}
	hosts := k.GetHosts()
	compareAndOutput(dns, hosts, k.ValidTargets, txtOwner, txtPrefix)
}

// RunR53 starts the main logic in route 53
func RunR53(dnsZone, txtOwner, txtPrefix string, ignoredSubdomains []string) {
	k := kube.New(dnsZone, ignoredSubdomains)
	registry := dns.RegistrySettings{
		Owner:  txtOwner,
		Prefix: txtPrefix,
	}
	client := dns.NewRoute53API(dnsZone, &registry)
	dns, err := client.ValidateRegistry()
	if err != nil {
		log.Fatal(err)
	}
	hosts := k.GetHosts()
	compareAndOutput(dns, hosts, k.ValidTargets, txtOwner, txtPrefix)
}

func compareAndOutput(d *dns.DNS, h kube.Hostnames, validTargets map[string][]string, owner, prefix string) {
	var noRegistry = make(kube.Hostnames)
	var noKubeResource []string
	var wrongRegistry = make(map[string]string)
	for hostname, resources := range h {
		if external, exists := d.ExternalRegistry[string(hostname)]; exists {
			wrongRegistry[string(hostname)] = external.WithPrefix
			continue
		}
		if _, exists := d.Registry[string(hostname)]; !exists {
			noRegistry[hostname] = resources
			continue
		}
	}
	for _, record := range d.MissingRegistry {
		if _, exists := h[kube.Hostname(record.Name)]; !exists {
			if _, exists := validTargets[record.Target]; exists {
				noKubeResource = append(noKubeResource, record.Name)
			}
		}
	}
	fmt.Printf("The following records can be deleted (%d items)\n", len(noKubeResource))
	for _, record := range noKubeResource {
		fmt.Println(record)
	}

	fmt.Printf("\nThe following records belong to the incorrect TXT registry (%d items) \n", len(wrongRegistry))
	for hostname, registry := range wrongRegistry {
		fmt.Printf("Record: %s\n", hostname)
		fmt.Printf("External TXT Record: %s\n", registry)
	}

	fmt.Printf("\nThe following records need TXT registry records added (%d items)\n", len(noRegistry))
	var recordValue string
	var toBeAdded TXTRecords
	for host, resource := range noRegistry {
		recordValue = fmt.Sprintf("heritage=external-dns,external-dns/owner=%s,external-dns/resource=%s/%s/%s", owner, resource[0].Kind, resource[0].Namespace, resource[0].Name)
		fmt.Printf("Record: %s\n", prefix+string(host))
		fmt.Printf("TXT Record Value: %s\n", recordValue)
		record := TXTRecord{
			Name:  prefix + string(host),
			Value: recordValue,
			TTL:   300,
		}
		toBeAdded.Records = append(toBeAdded.Records, record)
	}
	jsonOut, err := json.MarshalIndent(toBeAdded, "", "  ")
	if err != nil {
		panic("problem encoding json")
	}
	t := time.Now()
	f, err := os.Create("./tmp-" + t.Format("20060102150405") + ".json")
	if err != nil {
		panic("problem creating output file")
	}
	defer f.Close()
	f.Write(jsonOut)
}
