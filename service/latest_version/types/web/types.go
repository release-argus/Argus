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

// Package web provides a web-based lookup type.
package web

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
)

// #############
// # CONSTANTS #
// #############

// Type is the lookup type identifier for URL latest version lookups.
var Type = "url"

// #########
// # TYPES #
// #########

// Lookup is a web-based lookup type.
type Lookup struct {
	base.Lookup `json:",inline" yaml:",inline"`

	AllowInvalidCerts *bool          `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Allow invalid SSL certificates.
	Headers           shared.Headers `json:"headers,omitempty" yaml:"headers,omitempty"`                         // OPTIONAL: request headers.
}

// LookupDecode is an unmarshal-only helper for [Lookup].
type LookupDecode struct {
	AllowInvalidCerts *bool          `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"`
	Headers           shared.Headers `json:"headers,omitempty" yaml:"headers,omitempty"`
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
		AllowInvalidCerts: l.AllowInvalidCerts,
		Headers:           l.Headers,
	}

	// Unmarshal in the given format.
	if err := decode.Unmarshal(format, data, &aux); err != nil {
		return err //nolint:wrapcheck
	}
	l.AllowInvalidCerts = aux.AllowInvalidCerts
	l.Headers = aux.Headers

	// Normalise Type.
	if l.Type == "web" {
		l.Type = Type
	}

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
		AllowInvalidCerts: l.AllowInvalidCerts,
		Headers:           l.Headers.Copy(),
	}
}

// Copy returns a deep copy of the receiver as a [base.Interface].
func (l *Lookup) Copy(svcStatus *status.Status) base.Interface {
	if got := l.Clone(svcStatus); got != nil {
		return got
	}
	return nil
}

// InheritSecrets will inherit secrets from the `otherLookup`.
func (l *Lookup) InheritSecrets(otherLookup base.BaseInterface, secretRefs *shared.VSecretRef) {
	if otherL, ok := otherLookup.(*Lookup); ok && secretRefs != nil {
		l.Headers.InheritSecrets(otherL.Headers, secretRefs.Headers)
	}

	l.Lookup.InheritSecrets(otherLookup, secretRefs)
}
