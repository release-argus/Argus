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
	"fmt"
	"strings"

	"github.com/release-argus/Argus/util"
)

func (l *Lookup) GetAccessToken() *string {
	return util.FirstNonNilPtrWithEnv(
		l.AccessToken,
		l.Defaults.AccessToken,
		l.HardDefaults.AccessToken)
}

func (l *Lookup) GetAllowInvalidCerts() bool {
	return *util.FirstNonNilPtr(
		l.AllowInvalidCerts,
		l.Defaults.AllowInvalidCerts,
		l.HardDefaults.AllowInvalidCerts)
}

// ServiceURL (handles the github type where the URL may be `owner/repo`
// and adds the github.com/ prefix in that case).
func (l *Lookup) ServiceURL(ignoreWebURL bool) (serviceURL string) {
	if !ignoreWebURL && *l.Status.WebURL != "" {
		// Don't use this template if `LatestVersion` hasn't been found and is used in `WebURL`.
		latestVersion := l.Status.LatestVersion()
		if !(latestVersion == "" && strings.Contains(*l.Status.WebURL, "version")) {
			serviceURL = util.TemplateString(
				*l.Status.WebURL,
				util.ServiceInfo{LatestVersion: latestVersion})
			return
		}
	}

	serviceURL = l.URL
	// GitHub service. Get the non-API URL.
	if l.Type == "github" {
		// If it's "owner/repo" rather than a full path.
		if strings.Count(serviceURL, "/") == 1 {
			serviceURL = fmt.Sprintf("https://github.com/%s", serviceURL)
		}
	}
	return
}

// Get UsePreRelease will return whether GitHub PreReleases are considered valid for new versions.
func (l *Lookup) GetUsePreRelease() bool {
	return *util.FirstNonDefault(
		l.UsePreRelease,
		l.Defaults.UsePreRelease,
		l.HardDefaults.UsePreRelease)
}

// GetURL will ensure `url` is a valid GitHub API URL if `urlType` is 'github'
func (l *Lookup) GetURL() string {
	url := util.EvalEnvVars(l.URL)
	if l.Type == "github" {
		// Convert "owner/repo" to the API path.
		if strings.Count(url, "/") == 1 {
			apiTarget := "releases"
			if l.GitHubData.TagFallback() {
				apiTarget = "tags"
			}
			url = fmt.Sprintf("https://api.github.com/repos/%s/%s",
				url, apiTarget)
		}
	}
	return url
}
