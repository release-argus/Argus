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
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	logutil "github.com/release-argus/Argus/util/log"
)

// Lookup provides methods for retrieving the latest version of a service.
type Interface interface {
	// String returns a string representation of the parentLookup with any given prefix.
	String(parentLookup Interface, prefix string) string
	// Init initialises the Lookup, assigning Defaults and initialising child structs.
	Init(options *opt.Options, status *status.Status, defaults *Defaults, hardDefaults *Defaults)
	// InitMetrics initialises the parentLookup's metrics.
	InitMetrics(parentLookup Interface)
	// DeleteMetrics deletes the parentLookup's metrics.
	DeleteMetrics(parentLookup Interface)
	// CheckValues validates the fields of the Lookup struct.
	CheckValues(errPrefix string) (errs error)

	// Query the Lookup for the latest version.
	Query(metrics bool, logFrom logutil.LogFrom) (newVersion bool, err error)

	// Inherit state values from `fromLookup` if the values should query the same data.
	Inherit(fromLookup Interface)
	// IsEqual will return a bool of whether `this` Lookup is the same as `other` (excluding status).
	IsEqual(this, other Interface) bool

	// ServiceURL returns the Service URL for the Lookup.
	ServiceURL(ignoreWebURL bool) string

	// Helpers:

	// GetType returns the Lookup type.
	GetType() string
	// GetStatus returns the Lookup's status.
	GetStatus() *status.Status
	// GetOptions returns the Lookup's options.
	GetOptions() *opt.Options
	// GetRequire returns the Lookup's require options.
	GetRequire() *filter.Require
	// GetDefaults returns the Lookup's defaults.
	GetDefaults() *Defaults
	// GetHardDefaults returns the Lookup's hard defaults.
	GetHardDefaults() *Defaults
}
