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

package route53

// API represents a connection to route53
type API struct {
}

// NewAPI configures and returns a valid API object
func NewAPI() *API {
	var ret *API
	return ret
}

// GetRegistry represents the external-dns TXT registry in route53
func (*API) GetRegistry() map[string]map[string]string {
	var ret = make(map[string]map[string]string)
	return ret
}

// GetRecords represents the external-dns records in route53
func (*API) GetRecords() map[string]map[string]string {
	var ret = make(map[string]map[string]string)
	return ret
}
