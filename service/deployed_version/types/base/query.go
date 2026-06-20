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
	"fmt"

	"github.com/release-argus/Argus/internal/logx"
)

// HandleNewVersion handles a new version, updating the status, and logging the event.
func (l *Lookup) HandleNewVersion(version, releaseDate string, writeToDB bool, logFrom logx.LogFrom) {
	// If the new version is empty, or unchanged, return.
	if version == "" || version == l.Status.DeployedVersion() {
		return
	}

	// Set the new Deployed version.
	l.Status.SetDeployedVersion(version, "", writeToDB)

	// Announce version change to WebSocket clients.
	logx.Info(
		fmt.Sprintf("Updated to %q", version),
		logFrom,
		true,
	)
	l.Status.AnnounceUpdate()
}
