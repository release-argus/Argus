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
	semNewVersion, err := semver.NewVersion(newVersion)
	if err != nil {
		err = fmt.Errorf("failed converting %q to a semantic version. If all versions are in this style, consider adding url_commands to get the version into the style of 'MAJOR.MINOR.PATCH' (https://semver.org/), or disabling semantic versioning (globally with defaults.service.semantic_versioning or just for this service with the semantic_versioning var)",
			newVersion)
		logutil.Log.Error(err, logFrom, true)
		return err
	}

	// Check for a progressive change in version.
	if currentVersion != "" {
		deployedVersion := l.Status.DeployedVersion()
		semDeployedVersion, err := semver.NewVersion(deployedVersion)
		// If the old version is not a semantic version, we can't compare it.
		// (if we switched to semantic versioning with non-semantic versions tracked).
		if err == nil && semNewVersion.LessThan(semDeployedVersion) {
			// e.g.
			// newVersion = 1.2.9
			// oldVersion = 1.2.10
			err := fmt.Errorf("queried version %q is less than the deployed version %q",
				newVersion, deployedVersion)
			logutil.Log.Warn(err, logFrom, true)
			return err
		}
	}

	// Passed.
	return nil
}

// HandleNewVersion handles a new version, updating the status, and logging the event.
func (l *Lookup) HandleNewVersion(version, releaseDate string, logFrom logutil.LogFrom) (bool, error) {
	// Found a new version, so reset regex misses.
	l.Status.ResetRegexMisses()

	// First version found.
	if l.Status.LatestVersion() == "" {
		l.Status.SetLatestVersion(version, releaseDate, true)
		if l.Status.DeployedVersion() == "" {
			l.Status.SetDeployedVersion(version, "", true)
		}
		msg := fmt.Sprintf("Latest Release - %q", version)
		logutil.Log.Info(msg, logFrom, true)
		l.Status.AnnounceFirstVersion()

		// Don't notify on first version.
		return false, nil
	}

	// New version found.
	l.Status.SetLatestVersion(version, "", true)
	msg := fmt.Sprintf("New Release - %q", version)
	logutil.Log.Info(msg, logFrom, true)
	return true, nil
}
