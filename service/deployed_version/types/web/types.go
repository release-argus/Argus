// Copyright [2025] [Argus]
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
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var supportedMethods = []string{"GET", "POST"}

// Lookup is a web-based lookup type.
type Lookup struct {
	base.Lookup `yaml:",inline" json:",inline"` // Base struct for a Lookup.

	Method            string `yaml:"method,omitempty" json:"method,omitempty"`                           // REQUIRED: HTTP method.
	URL               string `yaml:"url,omitempty" json:"url,omitempty"`                                 // REQUIRED: URL to query.
	AllowInvalidCerts *bool  `yaml:"allow_invalid_certs,omitempty" json:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.

	BasicAuth     *BasicAuth `yaml:"basic_auth,omitempty" json:"basic_auth,omitempty"`         // OPTIONAL: Basic Auth credentials.
	Headers       []Header   `yaml:"headers,omitempty" json:"headers,omitempty"`               // OPTIONAL: Request Headers.
	Body          string     `yaml:"body,omitempty" json:"body,omitempty"`                     // OPTIONAL: Request Body.
	JSON          string     `yaml:"json,omitempty" json:"json,omitempty"`                     // OPTIONAL: JSON key to use e.g. version_current.
	Regex         string     `yaml:"regex,omitempty" json:"regex,omitempty"`                   // OPTIONAL: RegEx for the version.
	RegexTemplate string     `yaml:"regex_template,omitempty" json:"regex_template,omitempty"` // OPTIONAL: Template to apply to the RegEx match.
}

// New returns a new Lookup from a string in a given format (json/yaml).
func New(
	configFormat string,
	configData interface{}, // []byte | string | *yaml.Node | json.RawMessage.
	options *opt.Options,
	status *status.Status,
	defaults, hardDefaults *base.Defaults,
) (*Lookup, error) {
	lookup := &Lookup{}

	// Unmarshal.
	if err := util.UnmarshalConfig(configFormat, configData, lookup); err != nil {
		return nil, fmt.Errorf("failed to unmarshal web.Lookup:\n%w", err)
	}

	lookup.Init(
		options,
		status,
		defaults, hardDefaults)

	return lookup, nil
}

// UnmarshalJSON will unmarshal the Lookup.
func (l *Lookup) UnmarshalJSON(data []byte) error {
	// Alias to avoid recursion.
	type Alias Lookup
	aux := &struct {
		*Alias `json:",inline"`
	}{Alias: (*Alias)(l)}

	// Unmarshal.
	if err := json.Unmarshal(data, aux); err != nil {
		return err //nolint:wrapcheck
	}
	l.Type = "url"

	return nil
}

// UnmarshalYAML will unmarshal the Lookup.
func (l *Lookup) UnmarshalYAML(value *yaml.Node) error {
	// Alias to avoid recursion.
	type Alias Lookup
	aux := &struct {
		*Alias `yaml:",inline"`
	}{
		Alias: (*Alias)(l),
	}

	// Decode the YAML node into the struct.
	if err := value.Decode(aux); err != nil {
		return err //nolint:wrapcheck
	}

	l.Type = "url"
	return nil
}

// BasicAuth to use on the HTTP(s) request.
type BasicAuth struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// Header to use in the HTTP request.
type Header struct {
	Key   string `yaml:"key" json:"key"`     // Header key, e.g. X-Sig.
	Value string `yaml:"value" json:"value"` // Value to give the key.
}

// inheritSecrets from the `oldLookup`.
func (l *Lookup) InheritSecrets(otherLookup base.Interface, secretRefs *shared.DVSecretRef) {
	if otherL, ok := otherLookup.(*Lookup); ok {
		if l.BasicAuth != nil &&
			l.BasicAuth.Password == util.SecretValue &&
			otherL.BasicAuth != nil {
			l.BasicAuth.Password = otherL.BasicAuth.Password
		}

		// If we have headers in old and new.
		if len(l.Headers) != 0 &&
			len(otherL.Headers) != 0 {
			for i := range l.Headers {
				// If referencing a secret of an existing header.
				if l.Headers[i].Value == util.SecretValue {
					// Don't have a secretRef for this header.
					if i >= len(secretRefs.Headers) {
						break
					}
					oldIndex := secretRefs.Headers[i].OldIndex
					// Not a reference to an old Header.
					if oldIndex == nil {
						continue
					}

					if *oldIndex < len(otherL.Headers) {
						l.Headers[i].Value = otherL.Headers[*oldIndex].Value
					}
				}
			}
		}
	}
}
