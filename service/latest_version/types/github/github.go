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

// Package github provides a github-based lookup type.
package github

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	github_types "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/util"
)

// insertionSort performs an 'insertion sort' of the given `release` into the `filteredReleases` slice.
// This operation assumes that semantic versioning is enabled, and sorts the releases accordingly.
func insertionSort(release github_types.Release, filteredReleases *[]github_types.Release) {
	n := len(*filteredReleases)
	// find the insertion point.
	i := sort.Search(n, func(index int) bool {
		return (*filteredReleases)[index].SemanticVersion.LessThan(release.SemanticVersion)
	})

	// append an empty release to the end of the slice.
	*filteredReleases = append(*filteredReleases, github_types.Release{})

	// insert the release at the insertion point.
	if i < n {
		// shift elements to the right to make room for this release.
		copy((*filteredReleases)[i+1:], (*filteredReleases)[i:])
		// overwrite the element at the insertion point.
		(*filteredReleases)[i] = release
	} else {
		// append the release to the end of the slice.
		(*filteredReleases)[n] = release
	}
}

// filterGitHubReleases filters releases based on the following:
//   - URLCommands.
//   - Non-semantic versions (if semantic versions are required).
//   - Pre-releases (if not allowed).
//
// -
//
//	Returns the filtered list, sorted in descending order (if semantic-versioning wanted).
func (l *Lookup) filterGitHubReleases(logFrom util.LogFrom) []github_types.Release {
	semanticVersioning := l.Options.GetSemanticVersioning()
	usePreReleases := l.usePreRelease()

	releases := l.data.Releases()
	// Make a slice with the same capacity as releases.
	filteredReleases := make([]github_types.Release, 0, len(releases))

	for i := range releases {
		// Skip prereleases if not wanted.
		if releases[i].PreRelease && !usePreReleases {
			continue
		}

		// Check that TagName matches URLCommands.
		tag := releases[i].TagName
		if tag == "" {
			tag = releases[i].Name
		}
		tagName, err := l.URLCommands.Run(tag, logFrom)
		if err != nil || len(tagName) == 0 {
			continue
		}

		// Copy the release with the filtered TagName.
		release := releases[i]
		release.TagName = tagName[0]

		// If SemVer not required, add without sorting.
		if !semanticVersioning {
			filteredReleases = append(filteredReleases, release)
			continue
		}

		// Else, sort the versions.
		semVer, err := semver.NewVersion(tagName[0])
		if err != nil {
			continue
		}
		release.SemanticVersion = semVer
		// If first version, add without sorting.
		if len(filteredReleases) == 0 {
			filteredReleases = append(filteredReleases, release)
			continue
		}
		// else, insertion sort the release.
		insertionSort(release, &filteredReleases)
	}
	return filteredReleases
}

// checkGitHubReleasesBody validates that the response body conforms to the JSON formatting.
func (l *Lookup) checkGitHubReleasesBody(body []byte, logFrom util.LogFrom) ([]github_types.Release, error) {
	var releases []github_types.Release
	if err := json.Unmarshal(body, &releases); err != nil {
		jLog.Error(err, logFrom, true)
		err = fmt.Errorf("unmarshal of GitHub API data failed\n%w",
			err)
		jLog.Error(err, logFrom, true)
		return nil, err
	}

	return releases, nil
}
