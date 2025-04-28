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
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
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

	// Track the deployed version (DeployedVersion) of the `parent`.
	Track()
	// Query the Lookup for the latest version.
	Query(metrics bool, logFrom logutil.LogFrom) (err error)

	// InheritSecrets will inherit secrets from the `otherLookup`.
	InheritSecrets(otherLookup Interface, secretRefs *shared.DVSecretRef)

	// Helpers:

	// GetType returns the Lookup type.
	GetType() string
	// GetStatus returns the Lookup's status.
	GetStatus() *status.Status
	// GetServiceID returns the Lookup's service ID.
	GetServiceID() string
	// GetOptions returns the Lookup's options.
	GetOptions() *opt.Options
	// GetDefaults returns the Lookup's defaults.
	GetDefaults() *Defaults
	// GetHardDefaults returns the Lookup's hard defaults.
	GetHardDefaults() *Defaults
}
