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
	"errors"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
)

// Lookup is the base struct for an [Interface].
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

// LookupDecode is an unmarshal-only helper for [Lookup].
type LookupDecode struct {
	Type        string             `json:"type,omitempty" yaml:"type,omitempty"`
	URL         string             `json:"url,omitempty" yaml:"url,omitempty"`
	URLCommands filter.URLCommands `json:"url_commands,omitempty" yaml:"url_commands,omitempty"`
}

// IsZero implements the yaml.IsZeroer interface.
func (l *Lookup) IsZero() bool {
	return l.Type == "" &&
		l.URL == "" &&
		len(l.URLCommands) == 0 &&
		l.Require.IsZero()
}

// Init wires the dependencies into the receiver as pointers.
func (l *Lookup) Init(
	options *opt.Options,
	status *status.Status,
	cfg DefaultsConfig,
) {
	l.HardDefaults = cfg.Hard
	l.Defaults = cfg.Soft
	l.Status = status
	l.Options = options

	if l.Require != nil && cfg.Soft != nil {
		l.Require.Init(status, &cfg.Soft.Require)

		// If the 'Require' is empty, set it to nil.
		if l.Require.IsZero() {
			l.Require = nil
		}
	}
}

// GetServiceID returns the service ID of the receiver.
func (l *Lookup) GetServiceID() string {
	return l.Status.ServiceInfo.ID
}

// GetType returns the type of the receiver.
func (l *Lookup) GetType() string {
	return "-"
}

// GetOptions returns the receiver's options.
func (l *Lookup) GetOptions() *opt.Options {
	return l.Options
}

// GetRequire returns the receiver's require options.
func (l *Lookup) GetRequire() *filter.Require {
	return l.Require
}

// SetRequire sets the receiver's require options.
func (l *Lookup) SetRequire(require *filter.Require) {
	l.Require = require
}

// GetStatus returns the receiver's status.
func (l *Lookup) GetStatus() *status.Status {
	return l.Status
}

// SetStatus sets the receiver's status.
func (l *Lookup) SetStatus(status *status.Status) {
	l.Status = status
}

// GetDefaults returns the receiver's defaults.
func (l *Lookup) GetDefaults() *Defaults {
	return l.Defaults
}

// GetHardDefaults returns the receiver's hard defaults.
func (l *Lookup) GetHardDefaults() *Defaults {
	return l.HardDefaults
}

// ServiceURL returns the service's URL.
func (l *Lookup) ServiceURL() string {
	return l.URL
}

// CheckValues validates the fields of the receiver.
func (l *Lookup) CheckValues() error {
	var errs []error

	// url_commands
	if err := l.URLCommands.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "url_commands",
				Err: err,
			},
		)
	}
	// require
	if err := l.Require.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "require",
				Err: err,
			},
		)
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// InheritSecrets copies secrets from fromLookup if both query the same data.
func (l *Lookup) InheritSecrets(fromLookup BaseInterface, secretRefs *shared.VSecretRef) {
	l.inheritRequireTokens(fromLookup)
}

// Query queries the service for the latest version.
func (l *Lookup) Query(_ bool, _ logx.LogFrom) (bool, error) {
	return false, errors.New("not implemented")
}

// inheritRequireTokens copies require tokens from fromLookup.
func (l *Lookup) inheritRequireTokens(fromLookup BaseInterface) {
	require := l.GetRequire()
	fromRequire := fromLookup.GetRequire()
	require.Inherit(fromRequire)
}
