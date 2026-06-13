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

// Package github provides a github-based lookup type.
package github

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	ghtypes "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/util"
)

// filterGitHubReleases filters releases based on the following:
//   - URLCommands.
//   - Non-semantic versions (if semantic versions are required).
//   - Pre-releases (if not allowed).
//
// -
//
//	Returns the filtered list, sorted in descending order (if semantic-versioning wanted).
func (l *Lookup) filterGitHubReleases(logFrom logx.LogFrom) []ghtypes.Release {
	semanticVersioning := l.Options.GetSemanticVersioning()
	usePreReleases := l.usePreRelease()

	releases := l.data.Releases()
	// Make a slice with the same capacity as releases.
	filteredReleases := make([]ghtypes.Release, 0, len(releases))

	for _, release := range releases {
		// Skip prereleases if not wanted.
		if release.PreRelease && !usePreReleases {
			continue
		}

		// Check that TagName matches URLCommands.
		tag := util.FirstNonDefault(release.TagName, release.Name)
		tagName, err := l.URLCommands.Run(tag, logFrom)
		if err != nil || len(tagName) == 0 {
			continue
		}

		release.TagName = tagName[0]

		// Parse semver if enabled.
		if semanticVersioning {
			semVer, err := semver.NewVersion(tagName[0])
			if err != nil {
				continue
			}
			release.SemanticVersion = semVer
		}

		filteredReleases = append(filteredReleases, release)
	}

	if semanticVersioning {
		sort.Slice(filteredReleases, func(i, j int) bool {
			return filteredReleases[i].SemanticVersion.GreaterThan(
				filteredReleases[j].SemanticVersion,
			)
		})
	}

	return filteredReleases
}

// unmarshalGitHubReleasesBody validates that the response body conforms to the JSON formatting.
func (l *Lookup) unmarshalGitHubReleasesBody(body []byte) ([]ghtypes.Release, error) {
	var releases []ghtypes.Release
	if err := decode.Unmarshal("json", body, &releases); err != nil {
		return nil, fmt.Errorf("unmarshal of GitHub API data failed: %w", err)
	}

	return releases, nil
}
