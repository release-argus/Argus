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

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"errors"

	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// Lookup is the base struct for a Lookup.
type Lookup struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"` // "url".

	Options *opt.Options   `json:"-" yaml:"-"` // Options.
	Status  *status.Status `json:"-" yaml:"-"` // Service Status.

	Defaults     *Defaults `json:"-" yaml:"-"` // Defaults.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hard Defaults.
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
}

// String returns a string representation of the Lookup.
func (l *Lookup) String(parentLookup Interface, prefix string) string {
	return util.ToYAMLString(parentLookup, prefix)
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

// CheckValues validates the fields of the Lookup struct.
func (l *Lookup) CheckValues(prefix string) error {
	// Nothing to check.
	return nil
}

// Track will query the service's deployed version at the given Options.Interval.
func (l *Lookup) Track() {
	// Nothing to track.
}

// Query will query the service for the deployed version.
func (l *Lookup) Query(_ bool, _ logutil.LogFrom) error {
	return errors.New("not implemented")
}

// InheritSecrets will inherit secrets from the `otherLookup`.
func (l *Lookup) InheritSecrets(otherLookup Interface, secretRefs *shared.DVSecretRef) {
	// Nothing to inherit.
}
