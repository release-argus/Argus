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

// Package base provides the base struct for latest_version lookups.
package base

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/option"
)

// DefaultsConfig pairs soft and hard latest version defaults for decoding.
type DefaultsConfig struct {
	Soft *Defaults
	Hard *Defaults
}

// Defaults are the default values for a Lookup.
type Defaults struct {
	Type              string `json:"type,omitempty" yaml:"type,omitempty"`                               // "github" | "url".
	AccessToken       string `json:"access_token,omitempty" yaml:"access_token,omitempty"`               // GitHub access token to use.
	AllowInvalidCerts *bool  `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	UsePreRelease     *bool  `json:"use_prerelease,omitempty" yaml:"use_prerelease,omitempty"`           // Whether releases with prerelease tag are considered.

	Options *opt.Defaults          `json:"-" yaml:"-"`                               // Options for the Lookup.
	Require filter.RequireDefaults `json:"require,omitzero" yaml:"require,omitzero"` // Requirements before release considered valid.
}

// IsZero implements the yaml.IsZeroer interface.
func (d Defaults) IsZero() bool {
	return d.Type == "" &&
		d.AccessToken == "" &&
		d.AllowInvalidCerts == nil &&
		d.UsePreRelease == nil &&
		d.Require.IsZero()
}

// DefaultsDecode is an unmarshal-only helper for [Defaults].
type DefaultsDecode struct {
	Type              string `json:"type,omitempty" yaml:"type,omitempty"`
	AccessToken       string `json:"access_token,omitempty" yaml:"access_token,omitempty"`
	AllowInvalidCerts *bool  `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"`
	UsePreRelease     *bool  `json:"use_prerelease,omitempty" yaml:"use_prerelease,omitempty"`
}

// DecodeDefaults creates and returns new [Defaults] from format-encoded data.
func DecodeDefaults(format string, data []byte) (*Defaults, error) {
	var field Defaults

	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "latest_version",
			Err: err,
		}
	}

	return &field, nil
}

// Default sets the values of the receiver to their default values.
func (d *Defaults) Default() {
	// allow_invalid_certs.
	allowInvalidCerts := false
	d.AllowInvalidCerts = &allowInvalidCerts
	// use_prerelease.
	usePreRelease := false
	d.UsePreRelease = &usePreRelease

	// type.
	d.Type = "github"

	d.Require.Default()
}

// SetDefaults assigns defaults to the receiver.
func (d *Defaults) SetDefaults(dflts *Defaults) {
	d.Require.SetDefaults(&dflts.Require)
}

// CheckValues validates the fields of the receiver.
func (d *Defaults) CheckValues() error {
	if requireErrs := d.Require.CheckValues(); requireErrs != nil {
		return &decode.KeyFieldError{
			Key: "require",
			Err: requireErrs,
		}
	}

	return nil
}
