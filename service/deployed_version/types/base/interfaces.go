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
	"github.com/release-argus/Argus/internal/logx"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util/polymorphic"
)

// BaseInterface provides methods for retrieving the deployed version of a service.
type BaseInterface interface {
	// Init initialises the receiver, assigning Defaults and initialising child structs.
	Init(options *opt.Options, status *status.Status, cfg DefaultsConfig)
	// InitMetrics initialises the parentLookup's metrics.
	InitMetrics(parentLookup Interface)
	// DeleteMetrics deletes the parentLookup's metrics.
	DeleteMetrics(parentLookup Interface)
	// CheckValues validates the fields of the receiver.
	CheckValues() (errs error)

	// Query the Lookup for the deployed version.
	Query(metrics bool, logFrom logx.LogFrom) (err error)

	// InheritSecrets will inherit secrets from the `otherLookup`.
	InheritSecrets(otherLookup BaseInterface, secretRefs *shared.VSecretRef)

	// Helpers:

	// GetStatus returns the receiver's status.
	GetStatus() *status.Status
	// SetStatus sets the receiver's status.
	SetStatus(status *status.Status)
	// GetServiceID returns the receiver's service ID.
	GetServiceID() string
	// GetOptions returns the receiver's options.
	GetOptions() *opt.Options
	// GetDefaults returns the receiver's defaults.
	GetDefaults() *Defaults
	// GetHardDefaults returns the receiver's hard defaults.
	GetHardDefaults() *Defaults
}

// Interface provides methods for retrieving the deployed version of a service.
type Interface interface {
	BaseInterface
	polymorphic.Inheritable

	// String returns a string representation of the receiver with any given prefix.
	String(prefix string) string
	// Track the deployed version of the `parent`.
	Track()
	// Copy returns a deep copy of the receiver, with the given status.
	Copy(svcStatus *status.Status) Interface
}
