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

// Package forge provides the release handling shared by git forge lookup types.
package forge

import (
	"sort"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/filter"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
)

// UnmarshalReleases validates that the response body conforms to the JSON formatting.
func UnmarshalReleases(body []byte) ([]forgetypes.Release, error) {
	var releases []forgetypes.Release
	if err := decode.Unmarshal("json", body, &releases); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return releases, nil
}

// FilterReleases filters releases based on the following:
//   - urlCommands.
//   - Non-semantic versions (if semanticVersioning).
//   - Pre-releases (if not usePreReleases).
//
// Returns the filtered list, sorted in descending order (if semanticVersioning).
func FilterReleases(
	releases []forgetypes.Release,
	urlCommands filter.URLCommands,
	semanticVersioning bool,
	usePreReleases bool,
	logFrom logx.LogFrom,
) []forgetypes.Release {
	filteredReleases := make([]forgetypes.Release, 0, len(releases))

	for _, release := range releases {
		if release.PreRelease && !usePreReleases {
			continue
		}

		tag := util.FirstNonDefault(release.TagName, release.Name)
		tagName, err := urlCommands.Run(tag, logFrom)
		if err != nil || len(tagName) == 0 {
			continue
		}

		release.TagName = tagName[0]

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
			return releaseSortsBefore(filteredReleases[i], filteredReleases[j])
		})
	}

	return filteredReleases
}

// releaseSortsBefore reports whether release `a` should sort ahead of release `b` when ordering
// releases from newest to oldest.
//
// Both releases must have a parsed SemanticVersion.
func releaseSortsBefore(a, b forgetypes.Release) bool {
	return a.SemanticVersion.GreaterThan(b.SemanticVersion)
}
