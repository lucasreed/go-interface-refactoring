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

package ednsctl

import (
	"fmt"

	"github.com/lucasreed/go-interface-refactoring/after-ednsctl/pkg/internal/dns"
	"github.com/lucasreed/go-interface-refactoring/after-ednsctl/pkg/internal/dns/clouddns"
	"github.com/lucasreed/go-interface-refactoring/after-ednsctl/pkg/internal/dns/cloudflare"
	"github.com/lucasreed/go-interface-refactoring/after-ednsctl/pkg/internal/dns/route53"
)

// Config represents everything we need to know about a DNS Provider
type Config struct {
	API                    dns.API
	Provider               string
	ProviderSpecificConfig map[string]string
	RegistryPrefix         string
	RegistryOwner          string
}

// Run executes the main logic of the application
func Run(conf *Config) error {
	err := conf.configureAPI()
	if err != nil {
		return err
	}
	return nil
}

func (conf *Config) configureAPI() error {
	switch conf.Provider {
	case "clouddns":
		{
			conf.API = clouddns.NewAPI()
			return nil
		}
	case "cloudflare":
		{
			conf.API = cloudflare.NewAPI()
			return nil
		}
	case "route53":
		{
			conf.API = route53.NewAPI()
			return nil
		}
	default:
		{
			return fmt.Errorf("This DNS provider is not supported: %s", conf.Provider)
		}
	}
}
