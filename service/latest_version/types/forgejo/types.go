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

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// #############
// # CONSTANTS #
// #############

// Type is the lookup type identifier for Forgejo latest version lookups.
var Type = "forgejo"

// #########
// # TYPES #
// #########

// Lookup provides a Forgejo-based lookup type.
type Lookup struct {
	base.Lookup `json:",embed" yaml:",inline"`

	Host              string `json:"host,omitzero" yaml:"host,omitzero"`                               // Instance to query, e.g. "https://codeberg.org".
	AccessToken       string `json:"access_token,omitzero" yaml:"access_token,omitzero"`               // Access token to send to Host.
	AllowInvalidCerts *bool  `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"` // Default - false = Disallows invalid HTTPS certificates.
	UsePreRelease     *bool  `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"`           // Whether releases with the prerelease tag should be considered.

	typeDefaults     *Defaults // Forgejo-specific Defaults.
	typeHardDefaults *Defaults // Forgejo-specific Hard Defaults.
}

// lookupMarshal is a marshal-only helper for [Lookup].
type lookupMarshal struct {
	Type        string             `json:"type,omitzero" yaml:"type,omitzero"`
	Host        string             `json:"host,omitzero" yaml:"host,omitzero"`
	URL         string             `json:"url,omitzero" yaml:"url,omitzero"`
	URLCommands filter.URLCommands `json:"url_commands,omitempty" yaml:"url_commands,omitempty"`
	Require     *filter.Require    `json:"require,omitzero" yaml:"require,omitzero"`

	AccessToken       string `json:"access_token,omitzero" yaml:"access_token,omitzero"`
	AllowInvalidCerts *bool  `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"`
	UsePreRelease     *bool  `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"`
}

// LookupDecode is an unmarshal-only helper for [Lookup].
type LookupDecode struct {
	Host              string `json:"host,omitzero" yaml:"host,omitzero"`
	AccessToken       string `json:"access_token,omitzero" yaml:"access_token,omitzero"`
	AllowInvalidCerts *bool  `json:"allow_invalid_certs,omitzero" yaml:"allow_invalid_certs,omitzero"`
	UsePreRelease     *bool  `json:"use_prerelease,omitzero" yaml:"use_prerelease,omitzero"`
}

// ############
// # DECODING #
// ############

// MarshalJSON implements the json.Marshaler interface.
func (l *Lookup) MarshalJSON() ([]byte, error) {
	return decode.Marshal("json", l.marshalAux()) //nolint:wrapcheck
}

// MarshalYAML implements the yaml.InterfaceMarshaler interface.
func (l *Lookup) MarshalYAML() (any, error) {
	return l.marshalAux(), nil
}

// marshalAux converts the Lookup to its marshal-only helper representation.
func (l *Lookup) marshalAux() lookupMarshal {
	return lookupMarshal{
		Type:              l.Type,
		Host:              l.Host,
		URL:               l.URL,
		URLCommands:       l.URLCommands,
		Require:           l.Require,
		AccessToken:       l.AccessToken,
		AllowInvalidCerts: l.AllowInvalidCerts,
		UsePreRelease:     l.UsePreRelease,
	}
}

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
		Host:              l.Host,
		AccessToken:       l.AccessToken,
		AllowInvalidCerts: l.AllowInvalidCerts,
		UsePreRelease:     l.UsePreRelease,
	}

	// Unmarshal in the given format.
	if err := decode.Unmarshal(format, data, &aux); err != nil {
		return err //nolint:wrapcheck
	}
	l.Host = aux.Host
	l.AccessToken = aux.AccessToken
	l.AllowInvalidCerts = aux.AllowInvalidCerts
	l.UsePreRelease = aux.UsePreRelease

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

	return &Lookup{
		Lookup:            *l.Lookup.Clone(svcStatus), //nolint:staticcheck
		Host:              l.Host,
		AccessToken:       l.AccessToken,
		AllowInvalidCerts: util.ClonePtr(l.AllowInvalidCerts),
		UsePreRelease:     util.ClonePtr(l.UsePreRelease),
		typeDefaults:      l.typeDefaults,
		typeHardDefaults:  l.typeHardDefaults,
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

// SetTypeDefaults assigns the Forgejo-specific Defaults/HardDefaults to the receiver.
func (l *Lookup) SetTypeDefaults(defaults, hardDefaults *Defaults) {
	l.typeDefaults = defaults
	l.typeHardDefaults = hardDefaults
}

// GetTypeDefaults returns the receiver's Forgejo-specific Defaults/HardDefaults.
func (l *Lookup) GetTypeDefaults() (defaults, hardDefaults *Defaults) {
	return l.typeDefaults, l.typeHardDefaults
}
