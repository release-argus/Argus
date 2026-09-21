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

// Package forgejo provides a Forgejo-based lookup type.
package forgejo

// CommonDefaults are the Forgejo default values that apply to every instance.
type CommonDefaults struct {
	UsePreRelease *bool `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"` // Whether releases with prerelease tag are considered.
}

// IsZero implements the yaml.IsZeroer interface.
func (d CommonDefaults) IsZero() bool {
	return d.UsePreRelease == nil
}

// Defaults are the Forgejo-specific default values for a Lookup.
type Defaults struct {
	Common CommonDefaults `json:"common,omitzero" yaml:"common,omitzero"`
}

// IsZero implements the yaml.IsZeroer interface.
func (d Defaults) IsZero() bool {
	return d.Common.IsZero()
}

// Default sets the values of the receiver to their default values.
func (d *Defaults) Default() {
	usePreRelease := false
	d.Common.UsePreRelease = &usePreRelease
}
