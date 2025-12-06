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

// Package manual provides a manually set version lookup.
package manual

import (
	"errors"
	"time"

	logutil "github.com/release-argus/Argus/util/log"
)

// Track the deployed version (DeployedVersion) of the `parent`.
func (l *Lookup) Track() {
	// Do nothing.
}

// Query queries the source
// and returns whether a new release was found and updates LatestVersion if so.
//
// Parameters:
//
//	_metrics: ignored
func (l *Lookup) Query(_metrics bool, logFrom logutil.LogFrom) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.Version != "" {
		defer func() { l.Version = "" }()

		lastQueriedAt, _ := time.Parse(time.RFC3339, l.Status.DeployedVersionTimestamp())
		if time.Since(lastQueriedAt) < time.Second {
			return errors.New("manual version updates are rate-limited. Please try again in 1 second")
		}
		// If semantic versioning enabled, check version formatting.
		if l.Options.GetSemanticVersioning() {
			if _, err := l.Options.VerifySemanticVersioning(l.Version, logFrom); err != nil {
				return err //nolint:wrapcheck
			}
		}

		l.HandleNewVersion(l.Version, "", true, logFrom)
	}
	return nil
}
