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

package github

// Defaults are the GitHub-specific default values for a Lookup.
type Defaults struct {
	AccessToken   string `json:"access_token,omitzero" yaml:"access_token,omitzero"`     // Access token to use.
	UsePreRelease *bool  `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"` // Whether releases with prerelease tag are considered.
}

// IsZero implements the yaml.IsZeroer interface.
func (d Defaults) IsZero() bool {
	return d.AccessToken == "" &&
		d.UsePreRelease == nil
}

// Default sets the values of the receiver to their default values.
func (d *Defaults) Default() {
	usePreRelease := false
	d.UsePreRelease = &usePreRelease
}
