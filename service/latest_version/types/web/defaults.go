// Copyright [2026] [Argus]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

// Defaults are the URL-specific default values for a Lookup.
type Defaults struct {
	AllowInvalidCerts *bool `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
}

// IsZero implements the yaml.IsZeroer interface.
func (d Defaults) IsZero() bool {
	return d.AllowInvalidCerts == nil
}

// Default sets the values of the receiver to their default values.
func (d *Defaults) Default() {
	allowInvalidCerts := false
	d.AllowInvalidCerts = &allowInvalidCerts
}
