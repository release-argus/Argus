// Copyright [2023] [Argus]
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

package deployedver

import (
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

var (
	jLog *util.JLog
)

// LookupBase is the base struct for the Lookup struct.
type LookupBase struct {
	AllowInvalidCerts *bool `yaml:"allow_invalid_certs,omitempty" json:"allow_invalid_certs,omitempty"` // default - false = Disallows invalid HTTPS certificates
}

// LookupDefaults are the default values for the Lookup struct.
type LookupDefaults struct {
	LookupBase `yaml:",inline" json:",inline"`
}

// NewDefaults returns a new LookupDefaults struct.
func NewDefaults(
	allowInvalidCerts *bool,
) *LookupDefaults {
	return &LookupDefaults{
		LookupBase: LookupBase{
			AllowInvalidCerts: allowInvalidCerts}}
}

// Lookup the deployed version of the service.
type Lookup struct {
	URL        string `yaml:"url,omitempty" json:"url,omitempty"` // URL to query.
	LookupBase `yaml:",inline" json:",inline"`
	BasicAuth  *BasicAuth `yaml:"basic_auth,omitempty" json:"basic_auth,omitempty"` // Basic Auth for the HTTP(S) request.
	Headers    []Header   `yaml:"headers,omitempty" json:"headers,omitempty"`       // Headers for the HTTP(S) request.
	JSON       string     `yaml:"json,omitempty" json:"json,omitempty"`             // JSON key to use e.g. version_current.
	Regex      string     `yaml:"regex,omitempty" json:"regex,omitempty"`           // Regex to get the DeployedVersion

	Options *opt.Options      `yaml:"-" json:"-"` // Options for the lookups
	Status  *svcstatus.Status `yaml:"-" json:"-"` // Service Status

	Defaults     *LookupDefaults `yaml:"-" json:"-"` // Default values.
	HardDefaults *LookupDefaults `yaml:"-" json:"-"` // Hardcoded default values.
}

// New returns a new Lookup struct.
func New(
	allowInvalidCerts *bool,
	basicAuth *BasicAuth,
	headers *[]Header,
	json string,
	options *opt.Options,
	regex string,
	status *svcstatus.Status,
	url string,
	defaults *LookupDefaults,
	hardDefaults *LookupDefaults,
) (lookup *Lookup) {
	lookup = &Lookup{
		LookupBase: LookupBase{
			AllowInvalidCerts: allowInvalidCerts},
		BasicAuth:    basicAuth,
		JSON:         json,
		Options:      options,
		Regex:        regex,
		Status:       status,
		URL:          url,
		Defaults:     defaults,
		HardDefaults: hardDefaults}
	if headers != nil {
		lookup.Headers = *headers
	}
	return
}

// String returns a string representation of the Lookup.
func (l *Lookup) String(prefix string) (str string) {
	if l != nil {
		str = util.ToYAMLString(l, prefix)
	}
	return
}

// BasicAuth to use on the HTTP(s) request.
type BasicAuth struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// Header to use in the HTTP request.
type Header struct {
	Key   string `yaml:"key" json:"key"`     // Header key, e.g. X-Sig
	Value string `yaml:"value" json:"value"` // Value to give the key
}

// isEqual will return a bool of whether this lookup is the same as `other` (excluding status).
func (l *Lookup) IsEqual(other *Lookup) bool {
	return l.String("") == other.String("")
}
