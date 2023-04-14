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

package latestver

import (
	"fmt"
	"strings"

	"github.com/release-argus/Argus/util"
)

func (l *Lookup) GetAccessToken() *string {
	return util.GetFirstNonNilPtr(
		l.AccessToken,
		l.Defaults.AccessToken,
		l.HardDefaults.AccessToken)
}

func (l *Lookup) GetAllowInvalidCerts() bool {
	return *util.GetFirstNonNilPtr(
		l.AllowInvalidCerts,
		l.Defaults.AllowInvalidCerts,
		l.HardDefaults.AllowInvalidCerts)
}

// GetServiceURL returns the service's URL (handles the github type where the URL
// may be `owner/repo`, adding the github.com prefix in that case).
func (l *Lookup) GetServiceURL(ignoreWebURL bool) string {
	if !ignoreWebURL && *l.Status.WebURL != "" {
		// Don't use this template if `LatestVersion` hasn't been found and is used in `WebURL`.
		latestVersion := l.Status.GetLatestVersion()
		if !(latestVersion == "" && strings.Contains(*l.Status.WebURL, "version")) {
			return util.TemplateString(
				*l.Status.WebURL,
				util.ServiceInfo{LatestVersion: latestVersion})
		}
	}

	serviceURL := l.URL
	// GitHub service. Get the non-API URL.
	if l.Type == "github" {
		// If it's "owner/repo" rather than a full path.
		if strings.Count(serviceURL, "/") == 1 {
			serviceURL = fmt.Sprintf("https://github.com/%s", serviceURL)
		}
	}
	return serviceURL
}

// Get UsePreRelease will return whether GitHub PreReleases are considered valid for new versions.
func (l *Lookup) GetUsePreRelease() bool {
	return *util.GetFirstNonDefault(
		l.UsePreRelease,
		l.Defaults.UsePreRelease,
		l.HardDefaults.UsePreRelease)
}

// GetURL will ensure `url` is a valid GitHub API URL if `urlType` is 'github'
func GetURL(url string, urlType string) string {
	if urlType == "github" {
		// Convert "owner/repo" to the API path.
		if strings.Count(url, "/") == 1 {
			url = fmt.Sprintf("https://api.github.com/repos/%s/releases", url)
		}
	}
	return url
}
