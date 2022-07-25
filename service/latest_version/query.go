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

package latest_version

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/coreos/go-semver/semver"
	github_types "github.com/release-argus/Argus/service/latest_version/api_types"
	"github.com/release-argus/Argus/utils"
)

// Query queries the Service source, updating Service.LatestVersion
// and returning true if it has changed (is a new release),
// otherwise returns false.
func (l *Lookup) Query() (bool, error) {
	logFrom := utils.LogFrom{Primary: *(*l.Status).ServiceID}
	rawBody, err := l.httpRequest(logFrom)
	if err != nil {
		return false, err
	}

	version, err := l.GetVersion(rawBody, logFrom)
	if err != nil {
		return false, err
	}

	(*l.Status).SetLastQueried()
	wantSemanticVersioning := l.Options.GetSemanticVersioning()

	// If this version is different (new).
	if version != (*l.Status).LatestVersion {
		if wantSemanticVersioning {
			// Check it's a valid smenatic version
			newVersion, err := semver.NewVersion(version)
			if err != nil {
				err = fmt.Errorf("failed converting %q to a semantic version. If all versions are in this style, consider adding url_commands to get the version into the style of 'MAJOR.MINOR.PATCH' (https://semver.org/), or disabling semantic versioning (globally with defaults.service.semantic_versioning or just for this service with the semantic_versioning var)",
					version)
				jLog.Error(err, logFrom, true)
				return false, err
			}

			// Check for a progressive change in version.
			if (*l.Status).LatestVersion != "" {
				oldVersion, err := semver.NewVersion((*l.Status).LatestVersion)
				if err != nil {
					err := fmt.Errorf("failed converting %q to a semantic version (This is the old version, so you've probably just enabled `semantic_versioning`. Update/remove this latest_version from the config)",
						(*l.Status).LatestVersion)
					jLog.Error(err, logFrom, true)
					return false, err
				}

				// e.g.
				// newVersion = 1.2.9
				// oldVersion = 1.2.10
				// return false (don't notify anything. Stay on oldVersion)
				if newVersion.LessThan(*oldVersion) {
					err := fmt.Errorf("queried version %q is less than the deployed version %q",
						version, (*l.Status).LatestVersion)
					jLog.Warn(err, logFrom, true)
					return false, err
				}
			}
		}

		// Found new version, so reset regex misses.
		(*l.Status).RegexMissesContent = 0
		(*l.Status).RegexMissesVersion = 0

		// First version found.
		if (*l.Status).LatestVersion == "" {
			(*l.Status).SetLatestVersion(version)
			if (*l.Status).DeployedVersion == "" {
				(*l.Status).SetDeployedVersion(version)
			}
			msg := fmt.Sprintf("Latest Release - %q", version)
			jLog.Info(msg, logFrom, true)

			(*l.Status).AnnounceFirstVersion()

			// Don't notify on first version.
			return false, nil
		}

		// New version found.
		(*l.Status).SetLatestVersion(version)
		msg := fmt.Sprintf("New Release - %q", version)
		jLog.Info(msg, logFrom, true)
		return true, nil
	}

	// Announce `LastQueried`
	(*l.Status).AnnounceQuery()
	// No version change.
	return false, nil
}

func (l *Lookup) httpRequest(logFrom utils.LogFrom) (rawBody []byte, err error) {
	customTransport := &http.Transport{}
	// HTTPS insecure skip verify.
	if l.GetAllowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	req, err := http.NewRequest(http.MethodGet, GetURL(l.URL, l.Type), nil)
	if err != nil {
		jLog.Error(err, logFrom, true)
		return
	}

	if l.Type == "github" && utils.DefaultIfNil(l.GetAccessToken()) != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", *l.GetAccessToken()))
	}

	client := &http.Client{Transport: customTransport}
	resp, err := client.Do(req)

	if err != nil {
		// Don't crash on invalid certs.
		if strings.Contains(err.Error(), "x509") {
			err = fmt.Errorf("x509 (certificate invalid)")
			jLog.Warn(err, logFrom, true)
			return
		}
		jLog.Error(err, logFrom, true)
		return
	}

	// Read the response body.
	rawBody, err = ioutil.ReadAll(resp.Body)
	jLog.Error(err, logFrom, err != nil)
	return
}

// GetVersions will filter out releases from rawBody that are preReleases (if not wanted) and will sort releases if
// semantic versioning is wanted
func (l *Lookup) GetVersions(rawBody []byte, logFrom utils.LogFrom) (filteredReleases []github_types.Release, err error) {
	var releases []github_types.Release
	body := string(rawBody)
	// GitHub service.
	if l.Type == "github" {
		releases, err = l.checkGitHubReleasesBody(&rawBody, logFrom)
		if err != nil {
			return
		}
		filteredReleases = l.filterGitHubReleases(
			releases,
			logFrom,
		)

		// url service
	} else {
		version, err := l.URLCommands.Run(body, logFrom)
		if err != nil {
			return filteredReleases, err
		}
		filteredReleases = append(filteredReleases, github_types.Release{TagName: version})
	}

	if len(filteredReleases) == 0 {
		err = fmt.Errorf("no releases were found matching the url_commands")
		jLog.Warn(err, logFrom, true)
		return
	}
	return filteredReleases, nil
}

// GetVersion will return the latest version from rawBody matching the URLCommands and Regex requirements
func (l *Lookup) GetVersion(rawBody []byte, logFrom utils.LogFrom) (version string, err error) {
	filteredReleases, err := l.GetVersions(rawBody, logFrom)
	if err != nil {
		return
	}

	wantSemanticVersioning := l.Options.GetSemanticVersioning()
	for i := range filteredReleases {
		version = filteredReleases[i].TagName
		if wantSemanticVersioning && l.Type != "url" {
			version = filteredReleases[i].SemanticVersion.String()
		}

		// Break if version passed the regex check
		if err = l.Require.RegexCheckVersion(
			version,
			jLog,
			logFrom,
		); err == nil {
			// regexCheckContent if it's a newer version
			if version != (*l.Status).LatestVersion {
				if l.Type == "github" {
					// GitHub service
					if err = l.Require.RegexCheckContent(
						version,
						filteredReleases[i].Assets,
						jLog,
						logFrom,
					); err != nil {
						if i == len(filteredReleases)-1 {
							return
						}
						continue
					}
					break
					// Web service
				} else {
					if err = l.Require.RegexCheckContent(
						version,
						string(rawBody),
						jLog,
						logFrom,
					); err != nil {
						return
					}
				}

				// Ignore tags older than the deployed latest.
			} else {
				// return LatestVersion
				return
			}
		}
	}
	return
}
