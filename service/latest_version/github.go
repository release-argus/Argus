// Copyright [2023] [Argus]
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

package latestver

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/coreos/go-semver/semver"
	github_types "github.com/release-argus/Argus/service/latest_version/api_type"
	"github.com/release-argus/Argus/util"
)

// filterGitHubReleases will filter releases that fail the URLCommands, aren't semantic (if wanted),
// or are pre_release's (when they're not wanted). This list will be returned and be sorted descending.
func (l *Lookup) filterGitHubReleases(
	logFrom *util.LogFrom,
) (filteredReleases []github_types.Release) {
	semanticVerioning := l.Options.GetSemanticVersioning()
	usePreReleases := l.GetUsePreRelease()

	releases := l.GitHubData.Releases()
	// Make a slice with the same capacity as releases
	filteredReleases = make([]github_types.Release, 0, len(releases))

	for i := range releases {
		// If it's a prerelease, and they're not wanted, skip
		if releases[i].PreRelease && !usePreReleases {
			continue
		}

		var tagName string
		var err error

		// Check that TagName matches URLCommands
		if tagName, err = l.URLCommands.Run(releases[i].TagName, *logFrom); err != nil {
			continue
		}

		// Copy the release with the filtered TagName
		release := releases[i]
		release.TagName = tagName

		// If SemVer isn't wanted, add without any sorting
		if !semanticVerioning {
			filteredReleases = append(filteredReleases, release)
			continue
		}

		// Else, sort the versions
		semVer, err := semver.NewVersion(tagName)
		if err != nil {
			continue
		}
		release.SemanticVersion = semVer
		// If there's no other versions, just add it without insertion sort
		if len(filteredReleases) == 0 {
			filteredReleases = append(filteredReleases, release)
			continue
		}
		// Insertion Sort
		insertionSort(release, &filteredReleases)
	}
	return
}

// insertionSort will do an insertion sort of release on filteredReleases.
//
// Every GitHubRelease must be follow SemanticVersioning for this insertion
func insertionSort(release github_types.Release, filteredReleases *[]github_types.Release) {
	n := len(*filteredReleases)
	// find the insertion point
	i := sort.Search(n, func(index int) bool {
		return (*filteredReleases)[index].SemanticVersion.LessThan(*release.SemanticVersion)
	})

	// append an empty release to the end of the slice
	*filteredReleases = append(*filteredReleases, github_types.Release{})

	// insert the release at the insertion point
	if i < n {
		// shift elements to the right to make room for this release
		copy((*filteredReleases)[i+1:], (*filteredReleases)[i:])
		// overwrite the element at the insertion point
		(*filteredReleases)[i] = release
	} else {
		// append the release to the end of the slice
		(*filteredReleases)[n] = release
	}
}

// checkGitHubReleasesBody will check that the body is of the expected API format for a successful query
func (l *Lookup) checkGitHubReleasesBody(body *[]byte, logFrom *util.LogFrom) (releases []github_types.Release, err error) {
	// Check for rate lirmRDrit.
	if len(string(*body)) < 500 {
		if strings.Contains(string(*body), "rate limit") {
			err = errors.New("rate limit reached for GitHub")
			jLog.Warn(err, *logFrom, true)
			return
		}
		if !strings.Contains(string(*body), `"tag_name"`) {
			err = errors.New("github access token is invalid")
			jLog.Error(err, *logFrom, strings.Contains(string(*body), "Bad credentials"))

			err = fmt.Errorf("tag_name not found at %s\n%s",
				l.URL, string(*body))
			jLog.Error(err, *logFrom, true)
			return
		}
	}

	if err = json.Unmarshal(*body, &releases); err != nil {
		jLog.Error(err, *logFrom, true)
		err = fmt.Errorf("unmarshal of GitHub API data failed\n%w",
			err)
		jLog.Error(err, *logFrom, true)
	}
	return
}
