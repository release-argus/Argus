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
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// #############
// # CONSTANTS #
// #############

// Type is the lookup type identifier for URL deployed version lookups.
var Type = "url"

// #########
// # TYPES #
// #########

// Lookup is a web-based lookup type.
type Lookup struct {
	base.Lookup `json:",inline" yaml:",inline"`

	Method            string `json:"method,omitempty" yaml:"method,omitempty"`                           // REQUIRED: HTTP method.
	URL               string `json:"url,omitempty" yaml:"url,omitempty"`                                 // REQUIRED: url to query.
	AllowInvalidCerts *bool  `json:"allow_invalid_certs,omitempty" yaml:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
	TargetHeader      string `json:"target_header,omitempty" yaml:"target_header,omitempty"`             // OPTIONAL: header to target for the version.

	BasicAuth     *BasicAuth     `json:"basic_auth,omitempty" yaml:"basic_auth,omitempty"`         // OPTIONAL: basic auth credentials.
	Headers       shared.Headers `json:"headers,omitempty" yaml:"headers,omitempty"`               // OPTIONAL: request headers.
	Body          string         `json:"body,omitempty" yaml:"body,omitempty"`                     // OPTIONAL: request body.
	JSON          string         `json:"json,omitempty" yaml:"json,omitempty"`                     // OPTIONAL: JSON key to use e.g. version_current.
	Regex         string         `json:"regex,omitempty" yaml:"regex,omitempty"`                   // OPTIONAL: regex for the version.
	RegexTemplate string         `json:"regex_template,omitempty" yaml:"regex_template,omitempty"` // OPTIONAL: template to apply to the RegEx match.
}

// BasicAuth to use on the HTTP(s) request.
type BasicAuth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
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

	// Alias to avoid recursion.
	type Alias Lookup
	aux := (*Alias)(l)

	// Unmarshal in the given format.
	if err := decode.Unmarshal(format, data, aux); err != nil {
		return err //nolint:wrapcheck
	}

	// Normalise Type.
	if l.Type == "web" {
		l.Type = Type
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

// Copy returns a deep copy of the receiver.
func (l *Lookup) Copy(svcStatus *status.Status) base.Interface {
	if l == nil {
		return nil
	}

	return &Lookup{
		Lookup:            *l.Lookup.Clone(svcStatus), //nolint:staticcheck
		Method:            l.Method,
		URL:               l.URL,
		AllowInvalidCerts: util.ClonePtr(l.AllowInvalidCerts),
		TargetHeader:      l.TargetHeader,
		BasicAuth:         l.BasicAuth.Copy(),
		Headers:           l.Headers.Copy(),
		Body:              l.Body,
		JSON:              l.JSON,
		Regex:             l.Regex,
		RegexTemplate:     l.RegexTemplate,
	}
}

// Copy returns a deep copy of the receiver.
func (b *BasicAuth) Copy() *BasicAuth {
	if b == nil {
		return nil
	}

	return &BasicAuth{
		Username: b.Username,
		Password: b.Password,
	}
}

// InheritSecrets will inherit secrets from the `otherLookup`.
func (l *Lookup) InheritSecrets(otherLookup base.BaseInterface, secretRefs *shared.VSecretRef) {
	if otherL, ok := otherLookup.(*Lookup); ok {
		if l.BasicAuth != nil &&
			l.BasicAuth.Password == util.SecretValue &&
			otherL.BasicAuth != nil {
			l.BasicAuth.Password = otherL.BasicAuth.Password
		}

		if secretRefs != nil {
			l.Headers.InheritSecrets(otherL.Headers, secretRefs.Headers)
		}
	}
}
