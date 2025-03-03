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
	"fmt"

	"github.com/Masterminds/semver/v3"

	logutil "github.com/release-argus/Argus/util/log"
)

// VerifySemanticVersioning checks whether `newVersion` is a valid semantic version,
// and compares it with `currentVersion`.
//
// It returns an error if `newVersion` is not a valid semantic version,
// or if it is older than `currentVersion`.
func (l *Lookup) VerifySemanticVersioning(newVersion, currentVersion string, logFrom logutil.LogFrom) error {
	// Check it is a valid semantic version.
	_, err := semver.NewVersion(newVersion)
	if err != nil {
		err = fmt.Errorf(
			"failed converting %q to a semantic version. "+
				"If all versions are in this style, consider adding regex templating to get the version into the style of 'MAJOR.MINOR.PATCH' "+
				"(https://semver.org/), or disabling semantic versioning "+
				"(globally with defaults.service.semantic_versioning or just for this service with the options.semantic_versioning var)",
			newVersion,
		)
		logutil.Log.Error(err, logFrom, logFrom.Primary != "" || logFrom.Secondary != "")
		return err
	}

	// Passed.
	return nil
}

// HandleNewVersion handles a new version, updating the status, and logging the event.
func (l *Lookup) HandleNewVersion(version, releaseDate string, writeToDB bool, logFrom logutil.LogFrom) {
	// If the new version is empty, or unchanged, return.
	if version == "" || version == l.Status.DeployedVersion() {
		return
	}

	// Set the new Deployed version.
	l.Status.SetDeployedVersion(version, "", writeToDB)

	// Announce version change to WebSocket clients.
	logutil.Log.Info(
		fmt.Sprintf("Updated to %q", version),
		logFrom,
		true)
	l.Status.AnnounceUpdate()
}
