// Copyright [2024] [Argus]
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

// accessToken will return the GitHub API access token.
func (l *Lookup) accessToken() string {
	return util.FirstNonDefaultWithEnv(
		l.AccessToken,
		l.Defaults.AccessToken,
		l.HardDefaults.AccessToken)
}

// url will return a GitHub API URL for the repository.
func (l *Lookup) url() string {
	url := util.EvalEnvVars(l.URL)
	// Convert "owner/repo" to the API path.
	if strings.Count(url, "/") == 1 {
		apiTarget := "releases"
		if l.data.TagFallback() {
			apiTarget = "tags"
		}
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s",
			url, apiTarget)
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
func (l *Lookup) ServiceURL(ignoreWebURL bool) (serviceURL string) {
	if !ignoreWebURL && *l.Status.WebURL != "" {
		// Don't use this template if `LatestVersion` hasn't been found and is used in `WebURL`.
		latestVersion := l.Status.LatestVersion()
		if latestVersion != "" && strings.Contains(*l.Status.WebURL, "{{") {
			serviceURL = util.TemplateString(
				*l.Status.WebURL,
				util.ServiceInfo{LatestVersion: latestVersion})
			return
		}
	}

	serviceURL = l.URL
	// GitHub service. Get the non-API URL.
	// If "owner/repo" rather than a full path.
	if strings.Count(serviceURL, "/") == 1 {
		serviceURL = "https://github.com/" + serviceURL
	}
	return
}

// GetGitHubData will return the GitHub data. (For tests).
func (l *Lookup) GetGitHubData() *Data {
	return &l.data
}
