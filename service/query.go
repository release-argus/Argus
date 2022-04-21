// Copyright [2022] [Hymenaios]
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

package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/hymenaios-io/Hymenaios/utils"
	"gopkg.in/yaml.v3"
)

// Query queries the Service source, updating Service.LatestVersion
// and returning true if it has changed (is a new release),
// otherwise returns false.
func (s *Service) Query() (bool, error) {
	logFrom := utils.LogFrom{Primary: *s.ID}
	customTransport := &http.Transport{}
	// HTTPS insecure skip verify.
	if s.GetAllowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	req, err := http.NewRequest(http.MethodGet, GetURL(*s.URL, *s.Type), nil)
	if err != nil {
		jLog.Error(err, logFrom, true)
		return false, err
	}

	if s.GetAccessToken() != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", s.GetAccessToken()))
	}

	client := &http.Client{Transport: customTransport}
	resp, err := client.Do(req)

	if err != nil {
		// Don't crash on invalid certs.
		if strings.Contains(err.Error(), "x509") {
			err = fmt.Errorf("x509 (Cert invalid)")
			jLog.Warn(err, logFrom, true)
			return false, err
		}
		jLog.Error(err, logFrom, true)
		return false, err
	}

	// Read the response body.
	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		jLog.Error(err, logFrom, true)
		return false, err
	}
	// Convert the body to string.
	body := string(rawBody)
	version := body

	var versions []string
	var releases []GitHubRelease
	var filteredReleases []GitHubRelease
	// GitHub service.
	if *s.Type == "github" {
		// Check for rate limit.
		if len(body) < 500 {
			if strings.Contains(body, "rate limit") {
				err = errors.New("rate limit reached for GitHub")
				jLog.Warn(err, logFrom, true)
				return false, err
			}
			if !strings.Contains(body, `"tag_name"`) {
				err = errors.New("github access token is invalid")
				jLog.Fatal(err, logFrom, strings.Contains(body, "Bad credentials"))

				err = fmt.Errorf("tag_name not found at %s\n%s", *s.URL, body)
				jLog.Error(err, logFrom, true)
				return false, err
			}
		}

		err = yaml.Unmarshal(rawBody, &releases)
		if err != nil {
			jLog.Error(err, logFrom, true)
			msg := fmt.Sprintf("Unmarshal of GitHub API data failed\n%s", err)
			jLog.Error(msg, logFrom, true)
		}

		for i := range releases {
			// If it isn't a prerelease, or it is and they're wanted
			if !releases[i].PreRelease || (releases[i].PreRelease && s.GetUsePreRelease()) {
				versions = append(versions, releases[i].TagName)
				filteredReleases = append(filteredReleases, releases[i])
			}
		}

		// Web service
	} else {
		versions = append(versions, version)
	}

	for i := range versions {
		// Iterate through the URLCommands to filter out the version.
		version, err = s.URLCommands.run(versions[i], logFrom)
		// If URLCommands failed
		if err != nil {
			// If this is the last version, return
			if i == len(versions)-1 {
				// Don't log here as the `run` will already have logged
				return false, err
			}
			// Try another version
			continue
		}

		// Break if version passed the regex check
		if err := s.regexCheckVersion(
			version,
			logFrom,
		); err == nil {
			// regexCheckContent if it's a newer version
			if version != utils.DefaultIfNil(s.Status.LatestVersion) {
				if *s.Type == "github" {
					// GitHub service
					if err := s.regexCheckContent(
						version,
						filteredReleases[i].Assets,
						logFrom,
					); err != nil {
						if i == len(versions)-1 {
							return false, err
						}
						continue
					}
					break
					// Web service
				} else {
					if err := s.regexCheckContent(
						version,
						body,
						logFrom,
					); err != nil {
						return false, err
					}
				}

				// Ignore tags older than the current latest.
			} else {
				break
			}
		}
	}

	s.Status.SetLastQueried()
	// If this version is different (new).
	if version != utils.DefaultIfNil(s.Status.LatestVersion) {
		wantSemanticVersioning := s.GetSemanticVersioning()
		// Check for a progressive change in version.
		if wantSemanticVersioning && s.Status.LatestVersion != nil {
			oldVersion, err := semver.NewVersion(*s.Status.LatestVersion)
			if err != nil {
				err := fmt.Errorf("failed converting %q to a semantic version (This is the old version, so you've probably just enabled `semantic_versioning`. Update/remove this latest_version from the config)", *s.Status.LatestVersion)
				jLog.Error(err, logFrom, true)
				return false, err
			}
			newVersion, err := semver.NewVersion(version)
			if err != nil {
				err := fmt.Errorf("failed converting %q to a semantic version", version)
				jLog.Error(err, logFrom, true)
				return false, err
			}

			// e.g.
			// newVersion = 1.2.9
			// oldVersion = 1.2.10
			// return false (don't notify anything. Stay on oldVersion)
			if newVersion.LessThan(*oldVersion) {
				err := fmt.Errorf("queried version %q is less than the current version %q", version, *s.Status.LatestVersion)
				jLog.Error(err, logFrom, true)
				return false, err
			}
		}

		// Found new version, so reset regex misses.
		s.Status.RegexMissesContent = 0
		s.Status.RegexMissesVersion = 0

		// First version found.
		if s.Status.LatestVersion == nil {
			if wantSemanticVersioning {
				if _, err := semver.NewVersion(version); err != nil {
					err = fmt.Errorf("failed converting %q to a semantic version. If all versions are in this style, consider adding url_commands to get the version into the style of 'MAJOR.MINOR.PATCH' (https://semver.org/), or disabling semantic versioning (globally with defaults.service.semantic_versioning or just for this service with the semantic_versioning var)", version)
					jLog.Error(err, logFrom, true)
					return false, err
				}
			}

			s.Status.SetLatestVersion(version)
			if s.Status.CurrentVersion == nil && s.DeployedVersionLookup == nil {
				s.Status.SetCurrentVersion(version)
			}
			msg := fmt.Sprintf("Latest Release - %q", version)
			jLog.Info(msg, logFrom, true)

			if s.SaveChannel != nil {
				*s.SaveChannel <- true
			}

			s.AnnounceFirstVersion()

			// Don't notify on first version.
			return false, nil
		}

		// New version found.
		s.Status.SetLatestVersion(version)
		msg := fmt.Sprintf("New Release - %q", version)
		jLog.Info(msg, logFrom, true)
		*s.SaveChannel <- true
		return true, nil
	}

	// Announce `LastQueried`
	s.AnnounceQuery()
	// No version change.
	return false, nil
}
