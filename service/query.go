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

package service

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/release-argus/Argus/utils"
)

// Query queries the Service source, updating Service.LatestVersion
// and returning true if it has changed (is a new release),
// otherwise returns false.
func (s *Service) Query() (bool, error) {
	logFrom := utils.LogFrom{Primary: *s.ID}
	rawBody, err := s.httpRequest(logFrom)
	if err != nil {
		return false, err
	}

	version, err := s.GetVersion(rawBody, logFrom)
	if err != nil {
		return false, err
	}

	s.Status.SetLastQueried()
	wantSemanticVersioning := s.GetSemanticVersioning()
	// If this version is different (new).
	if version != s.Status.LatestVersion {
		// Check for a progressive change in version.
		if wantSemanticVersioning && s.Status.LatestVersion != "" {
			oldVersion, err := semver.NewVersion(s.Status.LatestVersion)
			if err != nil {
				err := fmt.Errorf("failed converting %q to a semantic version (This is the old version, so you've probably just enabled `semantic_versioning`. Update/remove this latest_version from the config)", s.Status.LatestVersion)
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
				err := fmt.Errorf("queried version %q is less than the deployed version %q", version, s.Status.LatestVersion)
				jLog.Warn(err, logFrom, true)
				return false, err
			}
		}

		// Found new version, so reset regex misses.
		s.Status.RegexMissesContent = 0
		s.Status.RegexMissesVersion = 0

		// First version found.
		if s.Status.LatestVersion == "" {
			if wantSemanticVersioning {
				if _, err := semver.NewVersion(version); err != nil {
					err = fmt.Errorf("failed converting %q to a semantic version. If all versions are in this style, consider adding url_commands to get the version into the style of 'MAJOR.MINOR.PATCH' (https://semver.org/), or disabling semantic versioning (globally with defaults.service.semantic_versioning or just for this service with the semantic_versioning var)", version)
					jLog.Error(err, logFrom, true)
					return false, err
				}
			}

			s.Status.SetLatestVersion(version)
			if s.Status.DeployedVersion == "" && s.DeployedVersionLookup == nil {
				s.SetDeployedVersion(version)
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

func (s *Service) httpRequest(logFrom utils.LogFrom) (rawBody []byte, err error) {
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
		return
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

func (s *Service) GetVersions(rawBody []byte, logFrom utils.LogFrom) (filteredReleases []GitHubRelease, err error) {
	var releases []GitHubRelease
	body := string(rawBody)
	// GitHub service.
	if *s.Type == "github" {
		// Check for rate limit.
		if len(body) < 500 {
			if strings.Contains(body, "rate limit") {
				err = errors.New("rate limit reached for GitHub")
				jLog.Warn(err, logFrom, true)
				return
			}
			if !strings.Contains(body, `"tag_name"`) {
				err = errors.New("github access token is invalid")
				jLog.Fatal(err, logFrom, strings.Contains(body, "Bad credentials"))

				err = fmt.Errorf("tag_name not found at %s\n%s", *s.URL, body)
				jLog.Error(err, logFrom, true)
				return
			}
		}

		if err = json.Unmarshal(rawBody, &releases); err != nil {
			jLog.Error(err, logFrom, true)
			msg := fmt.Errorf("unmarshal of GitHub API data failed\n%s", err)
			jLog.Error(msg, logFrom, true)
		}

		semanticVerioning := s.GetSemanticVersioning()
		for i := range releases {
			// If it isn't a prerelease, or it is and they're wanted
			if !releases[i].PreRelease || (releases[i].PreRelease && s.GetUsePreRelease()) {
				// Check that TagName matches URLCommands
				if releases[i].TagName, err = s.URLCommands.run(releases[i].TagName, logFrom); err != nil {
					continue
				}

				// If SemVer isn't wanted, add all
				if !semanticVerioning {
					filteredReleases = append(filteredReleases, releases[i])
					continue
				}

				// Else, sort the versions
				semVer, err := semver.NewVersion(releases[i].TagName)
				if err != nil {
					continue
				}
				releases[i].SemanticVersion = semVer
				if len(filteredReleases) == 0 {
					filteredReleases = append(filteredReleases, releases[i])
					continue
				}
				// Insertion Sort
				index := len(filteredReleases)
				for index != 0 {
					index--
					// semVer @ current is less than @ index
					if releases[i].SemanticVersion.LessThan(*filteredReleases[index].SemanticVersion) {
						if index == len(filteredReleases)-1 {
							filteredReleases = append(filteredReleases, releases[i])
							break
						}
						filteredReleases = append(filteredReleases[:index+1], filteredReleases[index:]...)
						filteredReleases[index+1] = releases[i]
						break
					} else if index == 0 {
						// releases[i] is newer than all filteredReleases. Prepend
						filteredReleases = append([]GitHubRelease{releases[i]}, filteredReleases...)
					}
				}
			}
		}

		// url service
	} else {
		version, err := s.URLCommands.run(body, logFrom)
		if err != nil {
			return filteredReleases, err
		}
		filteredReleases = append(filteredReleases, GitHubRelease{TagName: version})
	}

	if len(filteredReleases) == 0 {
		err = fmt.Errorf("no releases were found matching the url_commands")
		jLog.Warn(err, logFrom, true)
		return
	}
	return filteredReleases, nil
}

func (s *Service) GetVersion(rawBody []byte, logFrom utils.LogFrom) (version string, err error) {
	filteredReleases, err := s.GetVersions(rawBody, logFrom)
	if err != nil {
		return
	}

	wantSemanticVersioning := s.GetSemanticVersioning()
	for i := range filteredReleases {
		version = filteredReleases[i].TagName
		if wantSemanticVersioning && *s.Type != "url" {
			version = filteredReleases[i].SemanticVersion.String()
		}

		// Break if version passed the regex check
		if err = s.regexCheckVersion(
			version,
			logFrom,
		); err == nil {
			// regexCheckContent if it's a newer version
			if version != s.Status.LatestVersion {
				if *s.Type == "github" {
					// GitHub service
					if err = s.regexCheckContent(
						version,
						filteredReleases[i].Assets,
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
					if err = s.regexCheckContent(
						version,
						string(rawBody),
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
