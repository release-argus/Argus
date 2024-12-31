// Copyright [2024] [Argus]
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

// Package base provides the base struct for latest_version lookups.
package base

import (
	"fmt"

	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/option"
)

// Defaults are the default values for a Lookup.
type Defaults struct {
	AccessToken       string `yaml:"access_token,omitempty" json:"access_token,omitempty"`               // GitHub access token to use.
	AllowInvalidCerts *bool  `yaml:"allow_invalid_certs,omitempty" json:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	UsePreRelease     *bool  `yaml:"use_prerelease,omitempty" json:"use_prerelease,omitempty"`           // Whether releases with prerelease tag are considered.

	Options *opt.Defaults          `yaml:"-" json:"-"`             // Options for the Lookup.
	Require filter.RequireDefaults `yaml:"require" json:"require"` // Requirements before release considered valid.
}

// Default sets this Defaults to the default values.
func (d *Defaults) Default() {
	// allow_invalid_certs
	allowInvalidCerts := false
	d.AllowInvalidCerts = &allowInvalidCerts
	// use_prerelease
	usePreRelease := false
	d.UsePreRelease = &usePreRelease

	d.Require.Default()
}

// CheckValues validates the fields of the Defaults struct.
func (d *Defaults) CheckValues(prefix string) error {
	if requireErrs := d.Require.CheckValues(prefix + "  "); requireErrs != nil {
		return fmt.Errorf("%srequire:\n%w",
			prefix, requireErrs)
	}

	return nil
}
