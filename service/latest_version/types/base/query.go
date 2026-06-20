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
	"fmt"

	"github.com/release-argus/Argus/internal/logx"
)

// HandleNewVersion handles a new version, updating the status, and logging the event.
func (l *Lookup) HandleNewVersion(version, releaseDate string, logFrom logx.LogFrom) (bool, error) {
	// Found a new version, so reset regex misses.
	l.Status.ResetRegexMisses()

	// First version found.
	if l.Status.LatestVersion() == "" {
		l.Status.SetLatestVersion(version, releaseDate, true)
		if l.Status.DeployedVersion() == "" {
			l.Status.SetDeployedVersion(version, "", true)
		}
		msg := fmt.Sprintf("Latest Release - %q", version)
		logx.Info(msg, logFrom, true)
		l.Status.AnnounceFirstVersion()

		// Don't notify on first version.
		return false, nil
	}

	// New version found.
	l.Status.SetLatestVersion(version, "", true)
	msg := fmt.Sprintf("New Release - %q", version)
	logx.Info(msg, logFrom, true)
	return true, nil
}
