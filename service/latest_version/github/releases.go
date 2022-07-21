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

package github

import (
	"fmt"

	github_types "github.com/release-argus/Argus/service/latest_version/github/api_types"
	"github.com/release-argus/Argus/utils"
)

// GetVersion will return the latest version from rawBody matching the URLCommands and Regex requirements
func (l *LatestVersion) GetVersion(rawBody []byte, logFrom utils.LogFrom) (version string, err error) {
	filteredReleases, err := l.GetVersions(rawBody, logFrom)
	if err != nil {
		return
	}

	wantSemanticVersioning := l.Options.GetSemanticVersioning()
	for i := range filteredReleases {
		version = filteredReleases[i].TagName
		if wantSemanticVersioning {
			version = filteredReleases[i].SemanticVersion.String()
		}

		// Break if version passed the regex check
		if err = l.Require.RegexCheckVersion(
			version,
			jLog,
			logFrom,
		); err == nil {
			// regexCheckContent if it's a newer version
			if version != l.Status.LatestVersion {
				// GitHub service
				if err = l.Require.RegexCheckContent(
					version,
					filteredReleases[i].Assets,
					jLog,
					logFrom,
				); err != nil {
					if i == len(filteredReleases)-1 {
						return
					}
					continue
				}
				break

				// Ignore tags older than the deployed latest.
			} else {
				// return LatestVersion.
				return
			}
		}
	}
	return
}

// GetVersions will filter out releases from rawBody that are preReleases (if not wanted) and will sort releases if
// semantic versioning is wanted
func (l *LatestVersion) GetVersions(rawBody []byte, logFrom utils.LogFrom) (filteredReleases []github_types.Release, err error) {
	var releases []github_types.Release
	// GitHub service.
	releases, err = l.checkGitHubReleasesBody(&rawBody, logFrom)
	if err != nil {
		return
	}
	filteredReleases = l.filterGitHubReleases(
		releases,
		logFrom,
	)

	if len(filteredReleases) == 0 {
		err = fmt.Errorf("no releases were found matching the url_commands")
		jLog.Warn(err, logFrom, true)
		return
	}
	return filteredReleases, nil
}
