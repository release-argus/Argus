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

// Package manual provides a manually set version lookup.
package manual

import (
	"sync"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/status"
)

// #############
// # CONSTANTS #
// #############

// Type is the lookup type identifier for manual deployed version lookups.
var Type = "manual"

// #########
// # TYPES #
// #########

// Lookup is a web-based lookup type.
type Lookup struct {
	base.Lookup `json:",inline" yaml:",inline"`

	mu      sync.RWMutex // Lock for the Lookup.
	Version string       `json:"version,omitempty" yaml:"version,omitempty"` // OPTIONAL: Version to initialise with/set to.
}

// #########
// # STATE #
// #########

// Copy returns a deep copy of the receiver as a [base.Interface].
func (l *Lookup) Copy(svcStatus *status.Status) base.Interface {
	if l == nil {
		return nil
	}

	return &Lookup{
		Lookup:  *l.Lookup.Clone(svcStatus), //nolint:staticcheck
		Version: l.Version,
	}
}

// #############
// # STRINGIFY #
// #############

// String returns a string representation of the receiver.
func (l *Lookup) String(prefix string) string {
	return decode.ToYAMLString(l, prefix)
}
