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
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util/polymorphic"
)

// BaseInterface provides methods for retrieving the latest version of a service.
type BaseInterface interface {
	// Init initialises the Lookup, assigning Defaults and initialising child structs.
	Init(options *opt.Options, status *status.Status, cfg DefaultsConfig)
	// InitMetrics initialises the parentLookup's metrics.
	InitMetrics(parentLookup BaseInterface)
	// DeleteMetrics deletes the parentLookup's metrics.
	DeleteMetrics(parentLookup BaseInterface)
	// CheckValues validates the fields of the receiver.
	CheckValues() (errs error)

	// Query the Lookup for the latest version.
	Query(metrics bool, logFrom logx.LogFrom) (newVersion bool, err error)

	// InheritSecrets will inherit secrets from `otherLookup` if the values should query the same data.
	InheritSecrets(otherLookup BaseInterface, secretRefs *shared.VSecretRef)

	// ServiceURL returns the Service URL for the Lookup.
	ServiceURL() string

	// Helpers:

	// GetType returns the receiver type.
	GetType() string
	// GetStatus returns the receiver's status.
	GetStatus() *status.Status
	// SetStatus sets the receiver's status.
	SetStatus(status *status.Status)
	// GetServiceID returns the receiver's service ID.
	GetServiceID() string
	// GetOptions returns the receiver's options.
	GetOptions() *opt.Options
	// GetRequire returns the receiver's require options.
	GetRequire() *filter.Require
	// SetRequire sets the receiver's require options.
	SetRequire(require *filter.Require)
	// GetDefaults returns the receiver's defaults.
	GetDefaults() *Defaults
	// GetHardDefaults returns the receiver's hard defaults.
	GetHardDefaults() *Defaults
}

// Interface provides methods for retrieving the latest version of a service.
type Interface interface {
	BaseInterface
	polymorphic.Inheritable

	// Copy returns a deep copy of the receiver, with the given status.
	Copy(svcStatus *status.Status) Interface
	// String returns a string representation of the receiver with any given prefix.
	String(prefix string) string
}
