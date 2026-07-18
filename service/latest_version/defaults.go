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

// Package latestver provides the latest_version lookup service to for a service.
package latestver

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
)

// DefaultsConfig pairs soft and hard latest version defaults.
type DefaultsConfig struct {
	Soft *Defaults
	Hard *Defaults
}

// Defaults are the default values for a Lookup - fields common to every registered
// type live under 'common', and fields specific to a single type live under that
// type's own key.
type Defaults struct {
	Type string `json:"type,omitzero" yaml:"type,omitzero"` // "github" | "url".

	Common base.Defaults   `json:"common,omitzero" yaml:"common,omitzero"`
	GitHub github.Defaults `json:"github,omitzero" yaml:"github,omitzero"`
	URL    web.Defaults    `json:"url,omitzero" yaml:"url,omitzero"`

	// Deprecated: moved to 'github.access_token'.
	AccessTokenDeprecated string `json:"access_token,omitzero" yaml:"access_token,omitzero"`
	// Deprecated: moved to 'github.use_prerelease'.
	UsePreReleaseDeprecated *bool `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"`
	// Deprecated: moved to 'url.allow_invalid_certs'.
	AllowInvalidCertsDeprecated *bool `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"`
	// Deprecated: moved to 'common.require'.
	RequireDeprecated *filter.RequireDefaults `json:"require,omitzero" yaml:"require,omitzero"`
}

// DecodeDefaults creates and returns new [Defaults] from format-encoded data.
func DecodeDefaults(format string, data []byte) (*Defaults, error) {
	var field Defaults

	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, &decode.ErrKeyField{
			Key: "latest_version",
			Err: err,
		}
	}

	return &field, nil
}

// IsZero implements the yaml.IsZeroer interface.
func (d Defaults) IsZero() bool {
	return d.Type == "" &&
		d.Common.IsZero() &&
		d.GitHub.IsZero() &&
		d.URL.IsZero() &&
		d.AccessTokenDeprecated == "" &&
		d.UsePreReleaseDeprecated == nil &&
		d.AllowInvalidCertsDeprecated == nil &&
		(d.RequireDeprecated == nil || d.RequireDeprecated.IsZero())
}

// Default sets the values of the receiver to their default values.
func (d *Defaults) Default() {
	d.Type = "github"
	d.Common.Default()
	d.GitHub.Default()
	d.URL.Default()
}

// SetDefaults assigns defaults to the receiver.
func (d *Defaults) SetDefaults(dflts *Defaults) {
	d.Common.SetDefaults(&dflts.Common)
}

// MigrateDeprecated renames deprecated fields to their new locations.
func (d *Defaults) MigrateDeprecated() {
	// Deprecated: access_token -> github.access_token.
	if d.AccessTokenDeprecated != "" && d.GitHub.AccessToken == "" {
		d.GitHub.AccessToken = d.AccessTokenDeprecated
		logx.Deprecated("Renaming 'defaults.service.latest_version.access_token' to 'defaults.service.latest_version.github.access_token'")
	}
	d.AccessTokenDeprecated = ""

	// Deprecated: use_prerelease -> github.use_prerelease.
	if d.UsePreReleaseDeprecated != nil && d.GitHub.UsePreRelease == nil {
		d.GitHub.UsePreRelease = d.UsePreReleaseDeprecated
		logx.Deprecated("Renaming 'defaults.service.latest_version.use_prerelease' to 'defaults.service.latest_version.github.use_prerelease'")
	}
	d.UsePreReleaseDeprecated = nil

	// Deprecated: allow_invalid_certs -> url.allow_invalid_certs.
	if d.AllowInvalidCertsDeprecated != nil && d.URL.AllowInvalidCerts == nil {
		d.URL.AllowInvalidCerts = d.AllowInvalidCertsDeprecated
		logx.Deprecated("Renaming 'defaults.service.latest_version.allow_invalid_certs' to 'defaults.service.latest_version.url.allow_invalid_certs'")
	}
	d.AllowInvalidCertsDeprecated = nil

	// Deprecated: require -> common.require.
	if d.RequireDeprecated != nil && !d.RequireDeprecated.IsZero() && d.Common.Require.IsZero() {
		d.Common.Require = *d.RequireDeprecated
		logx.Deprecated("Renaming 'defaults.service.latest_version.require' to 'defaults.service.latest_version.common.require'")
	}
	d.RequireDeprecated = nil
}

// CheckValues validates the fields of the receiver.
func (d *Defaults) CheckValues() error {
	return d.Common.CheckValues() //nolint:wrapcheck
}

// applyTypeDefaults assigns the cfg's per-type Soft/Hard defaults onto
// lookup, based on its concrete type. It is a no-op for unregistered types.
func applyTypeDefaults(lookup Lookup, cfg DefaultsConfig) {
	switch v := lookup.(type) {
	case *github.Lookup:
		v.SetTypeDefaults(&cfg.Soft.GitHub, &cfg.Hard.GitHub)
	case *web.Lookup:
		v.SetTypeDefaults(&cfg.Soft.URL, &cfg.Hard.URL)
	}
}
