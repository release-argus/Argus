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
	"fmt"
	"strings"

	"github.com/release-argus/Argus/util"
)

// accessToken returns the GitHub API access token.
func (l *Lookup) accessToken() string {
	return util.FirstNonDefaultWithEnv(
		l.AccessToken,
		l.Defaults.AccessToken,
		l.HardDefaults.AccessToken)
}

// url returns a GitHub API URL for the repository.
func (l *Lookup) url(page int) string {
	url := util.EvalEnvVars(l.URL)
	// Convert "owner/repo" to the API path.
	if strings.Count(url, "/") == 1 {
		apiTarget := "releases"
		if l.data.TagFallback() {
			apiTarget = "tags"
		}
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s",
			url, apiTarget)

		// Query params
		params := make([]string, 0, 2)
		if page > 1 {
			params = append(params, fmt.Sprintf("page=%d", page))
		}
		if perPage := l.data.PerPage(); perPage != 0 {
			params = append(params, fmt.Sprintf("per_page=%d", perPage))
		}
		if len(params) > 0 {
			url += "?" + strings.Join(params, "&")
		}
	}

	return url
}

// usePreRelease returns we want to consider GitHub PreReleases for new versions.
func (l *Lookup) usePreRelease() bool {
	return *util.FirstNonDefault(
		l.UsePreRelease,
		l.Defaults.UsePreRelease,
		l.HardDefaults.UsePreRelease)
}

// ServiceURL translates possible `owner/repo` URLs, adding the github.com/ prefix.
func (l *Lookup) ServiceURL() string {
	serviceURL := l.URL
	// GitHub service. Get the non-API URL.
	// If "owner/repo" rather than a full path.
	if strings.Count(serviceURL, "/") == 1 {
		serviceURL = "https://github.com/" + serviceURL
	}

	return serviceURL
}

// GetGitHubData returns the GitHub data. (For tests).
func (l *Lookup) GetGitHubData() *Data {
	return &l.data
}
