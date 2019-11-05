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

package cmd

import (
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
)

var (
	apiKey            string
	apiUser           string
	dnsProvider       string
	dnsZone           string
	ignoredSubdomains []string
	txtPrefix         string
	txtOwner          string
	rootCmd           = &cobra.Command{
		Use:   "ednsctl",
		Short: "Verify external-dns TXT registry and created records are in sync",
		Long: dedent.Dedent(`
				ednsctl will grab all ingresses and services from a kube cluster and compare
				them with the given dns-provider to validate external-dns A records as well
				as the TXT registry
	   `),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Required Flags
	rootCmd.PersistentFlags().StringVarP(&dnsZone, "dns-zone", "z", "", "DNS Zone name e.g. example.com (required)")
	rootCmd.MarkPersistentFlagRequired("dns-zone")

	// Optional Flags
	rootCmd.PersistentFlags().StringVarP(&apiKey, "api-key", "k", "", "API key for the DNS provider, overwrites EDNS_API_KEY env var")
	rootCmd.PersistentFlags().StringVarP(&apiUser, "api-user", "u", "", "API user for the DNS provider, overwrites EDNS_API_USER env var")
	rootCmd.PersistentFlags().StringVarP(&txtPrefix, "prefix", "p", "", "TXT registry prefix setting in external-dns; default is none")
	rootCmd.PersistentFlags().StringVarP(&txtOwner, "owner", "o", "default", "TXT registry owner setting in external-dns")
	rootCmd.PersistentFlags().StringSliceVarP(&ignoredSubdomains, "ignored-subdomains", "i", make([]string, 0), "subdomains to ignore if necessary (comma separated list)")
}
