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

// Package filter provides filtering for latest_version queries.
package filter

import (
	"fmt"

	logutil "github.com/release-argus/Argus/util/log"
)

// ExecCommand will run Command.
func (r *Require) ExecCommand(logFrom logutil.LogFrom) error {
	if r == nil || len(r.Command) == 0 {
		return nil
	}

	// Apply the template vars to the command.
	cmd := r.Command.ApplyTemplate(r.Status.GetServiceInfo())

	// Execute the command.
	if err := cmd.Exec(logFrom); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}
