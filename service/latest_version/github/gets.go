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

	"github.com/release-argus/Argus/utils"
)

// GetURL will get the non-API URL.
func (l LatestVersion) GetFriendlyURL() string {
	// Convert "owner/repo" to the non-API path.
	return fmt.Sprintf("https://github.com/%s", l.URL)
}

// GetURL will ensure `url` is a valid GitHub API URL if `urlType` is 'github'
func (l LatestVersion) GetLookupURL() string {
	// Convert "owner/repo" to the API path.
	return fmt.Sprintf("https://api.github.com/repos/%s/releases", l.URL)
}

// Get UsePreRelease will return whether GitHub PreReleases are considered valid for new versions.
func (l *LatestVersion) GetUsePreRelease() bool {
	return *utils.GetFirstNonDefault(l.UsePreRelease, l.Defaults.UsePreRelease, l.HardDefaults.UsePreRelease)
}
