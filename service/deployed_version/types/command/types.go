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

// Package command provides a command-based version lookup.
package command

import (
	"slices"
	"sync"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/status"
)

// #############
// # CONSTANTS #
// #############

// Type is the lookup type identifier for command-based deployed version lookups.
var Type = "command"

// defaultCommandTimeout is the maximum time a version query command may run for.
const defaultCommandTimeout = 30 * time.Second

// #########
// # TYPES #
// #########

// Lookup is a command-based lookup type.
//
// The command is run on the service's configured interval, and its stdout is
// matched against the RegEx (if any) to determine the deployed version.
type Lookup struct {
	base.Lookup `json:",embed" yaml:",inline"`

	mu      sync.RWMutex // Lock for the Lookup.
	Command []string     `json:"command,omitempty" yaml:"command,omitempty"` // REQUIRED: command to run (argv) to get the version from its stdout.
	Regex   string       `json:"regex,omitzero" yaml:"regex,omitzero"`       // OPTIONAL: regex for the version.
}

// #########
// # STATE #
// #########

// Copy returns a deep copy of the receiver as a [base.Interface].
func (l *Lookup) Copy(svcStatus *status.Status) base.Interface {
	if l == nil {
		return nil
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	return &Lookup{
		Lookup:  *l.Lookup.Clone(svcStatus), //nolint:staticcheck
		Command: slices.Clone(l.Command),
		Regex:   l.Regex,
	}
}

// #########
// # STRINGIFY #
// #########

// String returns a string representation of the receiver.
func (l *Lookup) String(prefix string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return decode.ToYAMLString(l, prefix)
}