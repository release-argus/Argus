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

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
)

// filterGitHubReleases filters the cached releases against the URLCommands,
// semantic-versioning and pre-release settings of the receiver.
//
// See [forge.FilterReleases].
func (l *Lookup) filterGitHubReleases(logFrom logx.LogFrom) []forgetypes.Release {
	return forge.FilterReleases(
		l.data.Releases(),
		l.URLCommands,
		l.Options.GetSemanticVersioning(),
		l.usePreRelease(),
		logFrom,
	)
}

// unmarshalGitHubReleasesBody validates that the response body conforms to the JSON formatting.
func (l *Lookup) unmarshalGitHubReleasesBody(body []byte) ([]forgetypes.Release, error) {
	releases, err := forge.UnmarshalReleases(body)
	if err != nil {
		return nil, fmt.Errorf("unmarshal of GitHub API data failed: %w", err)
	}

	return releases, nil
}
