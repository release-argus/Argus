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
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/coreos/go-semver/semver"
	github_types "github.com/release-argus/Argus/service/latest_version/api_type"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

// Query queries the Service source, updating Service.LatestVersion
// and returning true if it has changed (is a new release),
// otherwise returns false.
func (l *Lookup) query(logFrom *util.LogFrom) (bool, error) {
	rawBody, err := l.httpRequest(logFrom)
	if err != nil {
		return false, err
	}

	version, err := l.GetVersion(rawBody, logFrom)
	if err != nil {
		return false, err
	}

	l.Status.SetLastQueried("")
	wantSemanticVersioning := l.Options.GetSemanticVersioning()

	// If this version is different (new?).
	latestVersion := l.Status.GetLatestVersion()
	if version != latestVersion {
		if wantSemanticVersioning {
			// Check it's a valid semnatic version
			newVersion, err := semver.NewVersion(version)
			if err != nil {
				err = fmt.Errorf("failed converting %q to a semantic version. If all versions are in this style, consider adding url_commands to get the version into the style of 'MAJOR.MINOR.PATCH' (https://semver.org/), or disabling semantic versioning (globally with defaults.service.semantic_versioning or just for this service with the semantic_versioning var)",
					version)
				jLog.Error(err, *logFrom, true)
				return false, err
			}

			// Check for a progressive change in version.
			if latestVersion != "" {
				oldVersion, err := semver.NewVersion(l.Status.GetDeployedVersion())
				// If the old version is not a semantic version, then we can't compare it.
				// (if we switched to semantic versioning with non-semantic versions tracked)
				if err == nil {
					// e.g.
					// newVersion = 1.2.9
					// oldVersion = 1.2.10
					// return false (don't notify anything and stay on oldVersion)
					if newVersion.LessThan(*oldVersion) {
						err := fmt.Errorf("queried version %q is less than the deployed version %q",
							version, l.Status.GetLatestVersion())
						jLog.Warn(err, *logFrom, true)
						return false, err
					}
				}
			}
		}

		// Found new version, so reset regex misses.
		l.Status.RegexMissesContent = 0
		l.Status.RegexMissesVersion = 0

		// First version found.
		if l.Status.GetLatestVersion() == "" {
			l.Status.SetLatestVersion(version, true)
			if l.Status.GetDeployedVersion() == "" {
				l.Status.SetDeployedVersion(version, true)
			}
			msg := fmt.Sprintf("Latest Release - %q", version)
			jLog.Info(msg, *logFrom, true)

			l.Status.AnnounceFirstVersion()

			// Don't notify on first version.
			return false, nil
		}

		// New version found.
		l.Status.SetLatestVersion(version, true)
		msg := fmt.Sprintf("New Release - %q", version)
		jLog.Info(msg, *logFrom, true)
		return true, nil
	}

	// Announce `LastQueried`
	l.Status.AnnounceQuery()
	// No version change.
	return false, nil
}

// Query the Lookup, updating Service.Status.LatestVersion
// and returning true if a new release was found.
//
// metrics - if true, set Prometheus metrics based on the query
func (l *Lookup) Query(metrics bool, logFrom *util.LogFrom) (newVersion bool, err error) {
	newVersion, err = l.query(logFrom)

	if metrics {
		l.queryMetrics(err)
	}

	return
}

// queryMetrics sets the Prometheus metrics for the LatestVersion query.
func (l *Lookup) queryMetrics(err error) {
	// If it failed
	if err != nil {
		switch e := err.Error(); {
		case strings.HasPrefix(e, "regex "):
			metric.SetPrometheusGauge(metric.LatestVersionQueryLiveness,
				*l.Status.ServiceID,
				2)
		case strings.HasPrefix(e, "failed converting") && strings.Contains(e, " semantic version."):
			metric.SetPrometheusGauge(metric.LatestVersionQueryLiveness,
				*l.Status.ServiceID,
				3)
		case strings.HasPrefix(e, "queried version") && strings.Contains(e, " less than "):
			metric.SetPrometheusGauge(metric.LatestVersionQueryLiveness,
				*l.Status.ServiceID,
				4)
		default:
			metric.IncreasePrometheusCounter(metric.LatestVersionQueryMetric,
				*l.Status.ServiceID,
				"",
				"",
				"FAIL")
			metric.SetPrometheusGauge(metric.LatestVersionQueryLiveness,
				*l.Status.ServiceID,
				0)
		}
	} else {
		metric.IncreasePrometheusCounter(metric.LatestVersionQueryMetric,
			*l.Status.ServiceID,
			"",
			"",
			"SUCCESS")
		metric.SetPrometheusGauge(metric.LatestVersionQueryLiveness,
			*l.Status.ServiceID,
			1)
	}
}

func (l *Lookup) httpRequest(logFrom *util.LogFrom) (rawBody []byte, err error) {
	customTransport := &http.Transport{}
	// HTTPS insecure skip verify.
	if l.GetAllowInvalidCerts() {
		customTransport = http.DefaultTransport.(*http.Transport).Clone()
		//#nosec G402 -- explicitly wanted InsecureSkipVerify
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	req, err := http.NewRequest(http.MethodGet, GetURL(l.URL, l.Type), nil)
	if err != nil {
		jLog.Error(err, *logFrom, true)
		return
	}

	// Set headers
	req.Header.Set("Connection", "close")
	if l.Type == "github" {
		// Access Token
		if util.DefaultIfNil(l.GetAccessToken()) != "" {
			req.Header.Set("Authorization", fmt.Sprintf("token %s", *l.GetAccessToken()))
		}
		// Conditional requests - https://docs.github.com/en/rest/overview/resources-in-the-rest-api#conditional-requests
		if l.GitHubData.ETag != "" {
			req.Header.Set("If-None-Match", l.GitHubData.ETag)
		}
	}

	client := &http.Client{Transport: customTransport}
	resp, err := client.Do(req)
	if err != nil {
		// Don't crash on invalid certs.
		if strings.Contains(err.Error(), "x509") {
			err = fmt.Errorf("x509 (certificate invalid)")
			jLog.Warn(err, *logFrom, true)
			return
		}
		jLog.Error(err, *logFrom, true)
		return
	}

	// Read the response body.
	defer resp.Body.Close()
	rawBody, err = io.ReadAll(resp.Body)
	jLog.Error(err, *logFrom, err != nil)
	if l.Type == "github" && err == nil {
		newETag := strings.TrimPrefix(resp.Header.Get("etag"), "W/")
		if l.GitHubData.ETag != newETag {
			jLog.Verbose("Potentially found new releases (ETag changed)", *logFrom, true)
		}
		l.GitHubData.ETag = newETag
	}
	return
}

// GetVersions will filter out releases from rawBody that are preReleases (if not wanted) and will sort releases if
// semantic versioning is wanted
func (l *Lookup) GetVersions(
	rawBody []byte,
	logFrom *util.LogFrom,
) (filteredReleases []github_types.Release, err error) {
	var releases []github_types.Release
	body := string(rawBody)
	// GitHub service.
	if l.Type == "github" {
		releases, err = l.checkGitHubReleasesBody(&rawBody, logFrom)
		if err != nil {
			return
		}
		// Store the unfiltered releases to support filter changes without a refetch
		// Copy as we don't want this slice to be modified by the filter
		l.GitHubData.Releases = releases
		// Filter releases
		filteredReleases = l.filterGitHubReleases(
			l.GitHubData.Releases,
			logFrom,
		)
		if len(filteredReleases) == 0 {
			err = fmt.Errorf("no releases were found matching the url_commands")
			jLog.Warn(err, *logFrom, true)
			return
		}

		// url service
	} else {
		var version string
		version, err = l.URLCommands.Run(body, *logFrom)
		if err != nil {
			//nolint:wrapcheck
			return
		}
		filteredReleases = []github_types.Release{{TagName: version}}
	}
	return
}

// GetVersion will return the latest version from rawBody matching the URLCommands and Regex requirements
func (l *Lookup) GetVersion(rawBody []byte, logFrom *util.LogFrom) (version string, err error) {
	var filteredReleases []github_types.Release
	// rawBody length = 0 if GitHub ETag is unchanged
	if len(rawBody) != 0 {
		filteredReleases, err = l.GetVersions(rawBody, logFrom)
		if err != nil {
			return
		}
	} else if l.Type == "github" {
		// ReCheck this ETag's filteredReleases incase filters/releases changed
		jLog.Verbose("Using cached releases (ETag unchanged)", *logFrom, true)
		filteredReleases = l.filterGitHubReleases(
			l.GitHubData.Releases,
			logFrom,
		)
	}

	wantSemanticVersioning := l.Options.GetSemanticVersioning()
	for i := range filteredReleases {
		version = filteredReleases[i].TagName
		if wantSemanticVersioning && l.Type != "url" {
			version = filteredReleases[i].SemanticVersion.String()
		}

		if l.Require == nil {
			break
		}

		// Check all `Require` filters for this version
		// Version RegEx
		if err = l.Require.RegexCheckVersion(version, logFrom); err != nil {
			continue
		}

		// Content RegEx
		var body interface{}
		if l.Type == "github" {
			// GitHub service
			body = filteredReleases[i].Assets
			// Web service
		} else {
			body = string(rawBody)
		}
		// If the Content doesn't match the provided RegEx
		if err = l.Require.RegexCheckContent(version, body, logFrom); err != nil {
			continue
		}

		// If the Command didn't return successfully
		if err = l.Require.ExecCommand(logFrom); err != nil {
			continue
		}

		// If the Docker tag doesn't exist
		if err = l.Require.DockerTagCheck(version); err != nil {
			if strings.HasSuffix(err.Error(), "\n") {
				err = fmt.Errorf(strings.TrimSuffix(err.Error(), "\n"))
			}
			jLog.Warn(err, *logFrom, true)
			continue
			// else if the tag does exist (and we did search for one)
		} else if l.Require.Docker != nil {
			jLog.Info(
				fmt.Sprintf(`found %s container "%s:%s"`,
					l.Require.Docker.Type, l.Require.Docker.Image, l.Require.Docker.GetTag(version)),
				*logFrom,
				true)
		}
		break
	}
	if version == "" {
		err = fmt.Errorf("no releases were found matching the url_commands and/or require")
		jLog.Warn(err, *logFrom, true)
	}
	return
}
