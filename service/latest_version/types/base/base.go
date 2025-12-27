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

// Package base provides the base struct for latest_version lookups.
package base

import (
	"errors"

	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// Lookup is the base struct for a Lookup.
type Lookup struct {
	Type        string             `json:"type,omitempty" yaml:"type,omitempty"`                 // "github" | "url".
	URL         string             `json:"url,omitempty" yaml:"url,omitempty"`                   // "owner/repo" or "https://github.com/owner/repo".
	URLCommands filter.URLCommands `json:"url_commands,omitempty" yaml:"url_commands,omitempty"` // Commands to filter the release from the URL request.
	Require     *filter.Require    `json:"require,omitempty" yaml:"require,omitempty"`           // Options to require before considering a release valid.

	Options *opt.Options   `json:"-" yaml:"-"` // Options.
	Status  *status.Status `json:"-" yaml:"-"` // Service Status.

	Defaults     *Defaults `json:"-" yaml:"-"` // Defaults.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hard Defaults.
}

// String returns a string representation of the Lookup.
func (l *Lookup) String(parentLookup Interface, prefix string) string {
	return util.ToYAMLString(parentLookup, prefix)
}

// Init will initialise the Lookup.
func (l *Lookup) Init(
	options *opt.Options,
	status *status.Status,
	defaults, hardDefaults *Defaults,
) {
	l.HardDefaults = hardDefaults
	l.Defaults = defaults
	l.Status = status
	l.Options = options

	if l.Require != nil && defaults != nil {
		l.Require.Init(status,
			&defaults.Require)

		// If the require is empty, set it to nil.
		if l.Require.String() == "{}\n" {
			l.Require = nil
		}
	}
}

// GetServiceID returns the service ID of the Lookup.
func (l *Lookup) GetServiceID() string {
	return l.Status.ServiceInfo.ID
}

// GetType returns the type of the Lookup.
func (l *Lookup) GetType() string {
	return l.Type
}

// GetOptions returns the Lookup's options.
func (l *Lookup) GetOptions() *opt.Options {
	return l.Options
}

// GetRequire returns the Lookup's require options.
func (l *Lookup) GetRequire() *filter.Require {
	return l.Require
}

// GetStatus returns the Lookup's status.
func (l *Lookup) GetStatus() *status.Status {
	return l.Status
}

// GetDefaults returns the Lookup's defaults.
func (l *Lookup) GetDefaults() *Defaults {
	return l.Defaults
}

// GetHardDefaults returns the Lookup's hard defaults.
func (l *Lookup) GetHardDefaults() *Defaults {
	return l.HardDefaults
}

// ServiceURL returns the service's URL.
func (l *Lookup) ServiceURL() string {
	return l.URL
}

// CheckValues validates the fields of the Lookup struct.
func (l *Lookup) CheckValues(prefix string) error {
	var errs []error

	// url_commands
	util.AppendCheckError(&errs, prefix, "url_commands",
		l.URLCommands.CheckValues(prefix+"  "))
	// require
	util.AppendCheckError(&errs, prefix, "require",
		l.Require.CheckValues(prefix+"  "))

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// inheritRequireTokens will inherit the `require` tokens from `fromLookup`.
func (l *Lookup) inheritRequireTokens(fromLookup Interface) {
	require := l.GetRequire()
	fromRequire := fromLookup.GetRequire()
	require.Inherit(fromRequire)
}

// InheritSecrets will inherit secrets from `fromLookup` if the values should query the same data.
func (l *Lookup) InheritSecrets(fromLookup Interface, secretRefs *shared.VSecretRef) {
	l.inheritRequireTokens(fromLookup)
}

// Query will query the service for the latest version.
func (l *Lookup) Query(_ bool, _ logutil.LogFrom) (bool, error) {
	return false, errors.New("not implemented")
}
