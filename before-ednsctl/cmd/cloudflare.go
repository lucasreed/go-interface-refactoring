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
	"os"

	"github.com/lucasreed/go-interface-refactoring/before-ednsctl/pkg/app"
	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
)

// cloudflareCmd represents the cloudflare command
var (
	cloudflareCmd = &cobra.Command{
		Use:   "cloudflare",
		Short: "Cloudflare DNS",
		Long: dedent.Dedent(`
			ednsctl will grab all ingresses and services from a kube cluster and compare
			them with the given dns-provider to validate external-dns A records as well
			as the TXT registry
	   `),
		Run: func(cmd *cobra.Command, args []string) {
			if apiKey == "" {
				apiKey = os.Getenv("EDNS_API_KEY")
			}
			if apiUser == "" {
				apiUser = os.Getenv("EDNS_API_USER")
			}
			app.RunCF(apiKey, apiUser, dnsZone, txtOwner, txtPrefix, ignoredSubdomains)
		},
	}
)

func init() {
	rootCmd.AddCommand(cloudflareCmd)
}
