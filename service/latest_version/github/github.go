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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
	github_types "github.com/release-argus/Argus/service/latest_version/github/api_types"
	"github.com/release-argus/Argus/utils"
)

// filterGitHubReleases will filter releases that fail the URLCommands, aren't semantic (if wanted),
// or are pre_release's (when they're not wanted). This list will be returned and be sorted descending.
func (l *LatestVersion) filterGitHubReleases(
	releases []github_types.Release,
	logFrom utils.LogFrom,
) (filteredReleases []github_types.Release) {
	semanticVerioning := l.Options.GetSemanticVersioning()
	usePreReleases := l.GetUsePreRelease()
	for i := range releases {
		// If it isn't a prerelease, or it is and they're wanted
		if !releases[i].PreRelease || (releases[i].PreRelease && usePreReleases) {
			var err error
			// Check that TagName matches URLCommands
			if releases[i].TagName, err = l.URLCommands.Run(releases[i].TagName, logFrom); err != nil {
				continue
			}

			// If SemVer isn't wanted, add all
			if !semanticVerioning {
				filteredReleases = append(filteredReleases, releases[i])
				continue
			}

			// Else, sort the versions
			semVer, err := semver.NewVersion(releases[i].TagName)
			if err != nil {
				continue
			}
			releases[i].SemanticVersion = semVer
			if len(filteredReleases) == 0 {
				filteredReleases = append(filteredReleases, releases[i])
				continue
			}
			// Insertion Sort
			insertionSort(releases[i], &filteredReleases)
		}
	}
	return
}

// insertionSort will do an insertion sort of release on filteredReleases.
//
// Every GitHubRelease must be follow SemanticVersioning for this insertion
func insertionSort(release github_types.Release, filteredReleases *[]github_types.Release) {
	index := len(*filteredReleases)
	for index != 0 {
		index--
		// semVer @current is less than @index
		if release.SemanticVersion.LessThan(*(*filteredReleases)[index].SemanticVersion) {
			if index == len(*filteredReleases)-1 {
				*filteredReleases = append(*filteredReleases, release)
				return
			}
			*filteredReleases = append((*filteredReleases)[:index+1], (*filteredReleases)[index:]...)
			(*filteredReleases)[index+1] = release
			return
		} else if index == 0 {
			// releases[i] is newer than all filteredReleases. Prepend
			*filteredReleases = append([]github_types.Release{release}, *filteredReleases...)
		}
	}
}

// checkGitHubReleasesBody will check that the body is of the expected API format for a successful query
func (l *LatestVersion) checkGitHubReleasesBody(body *[]byte, logFrom utils.LogFrom) (releases []github_types.Release, err error) {
	// Check for rate lirmRDrit.
	if len(string(*body)) < 500 {
		if strings.Contains(string(*body), "rate limit") {
			err = errors.New("rate limit reached for GitHub")
			jLog.Warn(err, logFrom, true)
			return
		}
		if !strings.Contains(string(*body), `"tag_name"`) {
			err = errors.New("github access token is invalid")
			jLog.Fatal(err, logFrom, strings.Contains(string(*body), "Bad credentials"))

			err = fmt.Errorf("tag_name not found at %s\n%s",
				l.URL, string(*body))
			jLog.Error(err, logFrom, true)
			return
		}
	}

	if err = json.Unmarshal(*body, &releases); err != nil {
		jLog.Error(err, logFrom, true)
		err = fmt.Errorf("unmarshal of GitHub API data failed\n%s",
			err)
		jLog.Error(err, logFrom, true)
	}
	return
}
