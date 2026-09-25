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
	"errors"
	"regexp"

	"github.com/release-argus/Argus/config/decode"
)

// CheckValues validates the fields of the receiver.
func (l *Lookup) CheckValues() error {
	var errs []error

	l.mu.RLock()
	commandSlice := l.Command
	regex := l.Regex
	l.mu.RUnlock()

	// Command.
	if len(commandSlice) == 0 {
		errs = append(
			errs,
			&decode.ErrField{
				Key:         "command",
				Description: "command to run to get the deployed_version from its stdout",
			},
		)
	}

	// RegEx.
	if regex != "" {
		if _, err := regexp.Compile(regex); err != nil {
			errs = append(
				errs,
				&decode.ErrField{
					Key:         "regex",
					Value:       regex,
					Description: "RegEx to extract the version from the command output",
				},
			)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}