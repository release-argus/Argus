// Copyright [2022] [Argus]
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

package url

import (
	"fmt"

	"github.com/release-argus/Argus/utils"
)

// GetVersion will return the latest version from rawBody matching the URLCommands and Regex requirements
func (l *LatestVersion) GetVersion(rawBody []byte, logFrom utils.LogFrom) (version string, err error) {
	version, err = l.filterReleases(rawBody, logFrom)
	if err != nil {
		return
	}

	// Break if version passed the regex check
	if err = l.Require.RegexCheckVersion(
		version,
		logFrom,
	); err == nil {
		// regexCheckContent if it's a newer version
		if version != l.Status.LatestVersion {
			if err = l.Require.RegexCheckContent(
				version,
				string(rawBody),
				logFrom,
			); err != nil {
				return
			}

			// Ignore tags older than the deployed latest.
		} else {
			// return LatestVersion
			return
		}
	}
	return
}

// GetVersions will filter out releases from rawBody with URLCommands
func (l *LatestVersion) filterReleases(rawBody []byte, logFrom utils.LogFrom) (filteredRelease string, err error) {
	body := string(rawBody)
	version, err := l.URLCommands.Run(body, logFrom)

	if err != nil {
		return version, err
	}
	if version == "" {
		err = fmt.Errorf("no releases were found matching the url_commands")
		jLog.Warn(err, logFrom, true)
		return
	}

	return version, nil
}
