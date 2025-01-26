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

// Package deployedver provides the deployed_version lookup.
package deployedver

import (
	"encoding/json"
	"fmt"

	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	supportedTypes = []string{"GET", "POST"}
)

// Base is the base struct for the Lookup struct.
type Base struct {
	AllowInvalidCerts *bool `yaml:"allow_invalid_certs,omitempty" json:"allow_invalid_certs,omitempty"` // Default - false = Disallows invalid HTTPS certificates.
}

// Defaults are the default values for the Lookup struct.
type Defaults struct {
	Base `yaml:",inline" json:",inline"`

	Options *opt.Defaults `yaml:"-" json:"-"` // Options for the lookup.
}

// NewDefaults returns a new Defaults struct.
func NewDefaults(
	allowInvalidCerts *bool,
) *Defaults {
	return &Defaults{
		Base: Base{
			AllowInvalidCerts: allowInvalidCerts}}
}

// Default sets this Defaults to the default values.
func (ld *Defaults) Default() {
	allowInvalidCerts := false
	ld.AllowInvalidCerts = &allowInvalidCerts
}

// Lookup the deployed version of the service.
type Lookup struct {
	Method        string `yaml:"method,omitempty" json:"method,omitempty"` // REQUIRED: HTTP method.
	URL           string `yaml:"url,omitempty" json:"url,omitempty"`       // REQUIRED: URL to query.
	Base          `yaml:",inline" json:",inline"`
	BasicAuth     *BasicAuth `yaml:"basic_auth,omitempty" json:"basic_auth,omitempty"`         // OPTIONAL: Basic Auth credentials.
	Headers       []Header   `yaml:"headers,omitempty" json:"headers,omitempty"`               // OPTIONAL: Request Headers.
	Body          string     `yaml:"body,omitempty" json:"body,omitempty"`                     // OPTIONAL: Request Body.
	JSON          string     `yaml:"json,omitempty" json:"json,omitempty"`                     // OPTIONAL: JSON key to use e.g. version_current.
	Regex         string     `yaml:"regex,omitempty" json:"regex,omitempty"`                   // OPTIONAL: RegEx for the version.
	RegexTemplate string     `yaml:"regex_template,omitempty" json:"regex_template,omitempty"` // OPTIONAL: Template to apply to the RegEx match.

	Options *opt.Options   `yaml:"-" json:"-"` // Options for the lookups.
	Status  *status.Status `yaml:"-" json:"-"` // Service Status.

	Defaults     *Defaults `yaml:"-" json:"-"` // Default values.
	HardDefaults *Defaults `yaml:"-" json:"-"` // Hardcoded default values.
}

// New returns a new instance of Lookup from the provided configuration data
// in either JSON or YAML format, and initialises it with the provided
// options, status, defaults, and hardDefaults.
func New(
	configFormat string,
	configData interface{}, // []byte | string | *yaml.Node.
	options *opt.Options,
	status *status.Status,
	defaults, hardDefaults *Defaults,
) (*Lookup, error) {
	lookup := &Lookup{}

	// Unmarshal.
	if err := util.UnmarshalConfig(configFormat, configData, lookup); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployedver.Lookup:\n%w", err)
	}

	lookup.Init(
		options,
		status,
		defaults, hardDefaults)

	return lookup, nil
}

// Copy returns a copy of the Lookup.
func Copy(
	lookup *Lookup,
) *Lookup {
	if lookup == nil {
		return nil
	}

	// JSON of existing lookup.
	lookupJSON, _ := json.Marshal(lookup)

	// Create a new lookup.
	newLookup, _ := New(
		"json", lookupJSON,
		lookup.Options.Copy(),
		&status.Status{},
		lookup.Defaults,
		lookup.HardDefaults)

	return newLookup
}

// String returns a string representation of the Lookup.
func (l *Lookup) String(prefix string) string {
	if l == nil {
		return ""
	}
	return util.ToYAMLString(l, prefix)
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

// IsEqual will return a bool of whether this lookup is the same as `other` (excluding status).
func (l *Lookup) IsEqual(other *Lookup) bool {
	// If one/both nil.
	if other == nil || l == nil {
		// Equal if both nil.
		return other == nil && l != nil
	}

	// Equal if Options and Lookup marshal to the same strings.
	return l.Options.String() == other.Options.String() &&
		l.String("") == other.String("")
}
