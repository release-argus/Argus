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

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"errors"

	"github.com/release-argus/Argus/internal/logx"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
)

// #########
// # TYPES #
// #########

// Lookup is the base struct for an [Interface].
type Lookup struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"` // "url".

	Options *opt.Options   `json:"-" yaml:"-"` // Options.
	Status  *status.Status `json:"-" yaml:"-"` // Service Status.

	Defaults     *Defaults `json:"-" yaml:"-"` // Defaults.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hard Defaults.
}

// ########
// # INIT #
// ########

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
}

// #############
// # ACCESSORS #
// #############

// GetServiceID returns the service ID of the receiver.
func (l *Lookup) GetServiceID() string {
	return l.Status.ServiceInfo.ID
}

// GetOptions returns the receiver's options.
func (l *Lookup) GetOptions() *opt.Options {
	return l.Options
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

// #############
// # INTERFACE #
// #############

// CheckValues validates the fields of the receiver.
func (l *Lookup) CheckValues() error {
	// Nothing to check.
	return nil
}

// Query will query the service for the deployed version.
func (l *Lookup) Query(_ bool, _ logx.LogFrom) error {
	return errors.New("not implemented")
}

// InheritSecrets is a no-op for the Lookup interface.
func (l *Lookup) InheritSecrets(otherLookup BaseInterface, secretRefs *shared.VSecretRef) {
	// Nothing to inherit.
}
