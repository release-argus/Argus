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

// Package github provides a github-based lookup type.
package github

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/status"
)

// #############
// # CONSTANTS #
// #############

// Type is the lookup type identifier for GitHub latest version lookups.
var Type = "github"

// #########
// # TYPES #
// #########

// Lookup provides a GitHub-based lookup type.
type Lookup struct {
	base.Lookup `json:",embed" yaml:",inline"`

	AccessToken   string `json:"access_token,omitzero" yaml:"access_token,omitzero"`     // GitHub access token to use.
	UsePreRelease *bool  `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"` // Whether releases with the prerelease tag should be considered.

	data Data // GitHub Conditional Request vars / Releases.

	typeDefaults     *Defaults // GitHub-specific Defaults.
	typeHardDefaults *Defaults // GitHub-specific Hard Defaults.
}

// LookupDecode is an unmarshal-only helper for [Lookup].
type LookupDecode struct {
	AccessToken   string `json:"access_token,omitzero" yaml:"access_token,omitzero"`
	UsePreRelease *bool  `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"`
}

// ############
// # DECODING #
// ############

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [Decode] for a complete Lookup.
func (l *Lookup) UnmarshalJSON(data []byte) error {
	return l.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [Decode] for a complete Lookup.
func (l *Lookup) UnmarshalYAML(data []byte) error {
	return l.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (l *Lookup) unmarshal(format string, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	aux := LookupDecode{
		AccessToken:   l.AccessToken,
		UsePreRelease: l.UsePreRelease,
	}

	// Unmarshal in the given format.
	if err := decode.Unmarshal(format, data, &aux); err != nil {
		return err //nolint:wrapcheck
	}
	l.AccessToken = aux.AccessToken
	l.UsePreRelease = aux.UsePreRelease
	l.data.SetETag(getEmptyListETag())

	// Require.
	if l.Defaults != nil && l.HardDefaults != nil {
		if err := base.UnmarshalRequire(
			format, data,
			l,
			l.Status,
			&l.Defaults.Require,
		); err != nil {
			return err //nolint:wrapcheck
		}
	}

	return nil
}

// #############
// # STRINGIFY #
// #############

// String returns a string representation of the receiver.
func (l *Lookup) String(prefix string) string {
	return decode.ToYAMLString(l, prefix)
}

// #########
// # STATE #
// #########

// Clone returns a deep copy of the receiver.
func (l *Lookup) Clone(svcStatus *status.Status) *Lookup {
	if l == nil {
		return nil
	}

	var usePreRelease *bool
	if l.UsePreRelease != nil {
		value := *l.UsePreRelease
		usePreRelease = &value
	}

	return &Lookup{
		Lookup:           *l.Lookup.Clone(svcStatus), //nolint:staticcheck
		AccessToken:      l.AccessToken,
		UsePreRelease:    usePreRelease,
		data:             *l.data.Copy(),
		typeDefaults:     l.typeDefaults,
		typeHardDefaults: l.typeHardDefaults,
	}
}

// Copy returns a deep copy of the receiver as a [base.Interface].
func (l *Lookup) Copy(svcStatus *status.Status) base.Interface {
	if got := l.Clone(svcStatus); got != nil {
		return got
	}
	return nil
}

// ############
// # DEFAULTS #
// ############

// SetTypeDefaults assigns the GitHub-specific Defaults/HardDefaults to the receiver.
func (l *Lookup) SetTypeDefaults(defaults, hardDefaults *Defaults) {
	l.typeDefaults = defaults
	l.typeHardDefaults = hardDefaults
}

// GetTypeDefaults returns the receiver's GitHub-specific Defaults/HardDefaults.
func (l *Lookup) GetTypeDefaults() (defaults, hardDefaults *Defaults) {
	return l.typeDefaults, l.typeHardDefaults
}
