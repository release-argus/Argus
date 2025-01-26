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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	github_types "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// Query queries the source,
// and returns whether a new release was found, and updates LatestVersion if so.
//
// Parameters:
//   - metrics: if true, set Prometheus metrics based on the query.
func (l *Lookup) Query(metrics bool, logFrom logutil.LogFrom) (bool, error) {
	newVersion, err := l.query(logFrom, 0)

	if metrics {
		l.QueryMetrics(l, err)
	}

	return newVersion, err
}

// Query queries the source,
// and returns whether a new release was found, and updates LatestVersion if so.
//
// Parameters:
//
//	checkNumber: 0 for first check, 1 for second check (if the first check found a new version).
func (l *Lookup) query(logFrom logutil.LogFrom, checkNumber int) (bool, error) {
	body, err := l.httpRequest(logFrom)
	if err != nil {
		return false, err
	}

	// Get the latest version, and its release date from the body.
	version, releaseDate, err := l.getVersion(body, logFrom)
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		return false, err
	}

	// Only set on the first check.
	if checkNumber == 0 {
		l.Status.SetLastQueried("")
	}

	// If this version differs (new?).
	previousLatestVersion := l.Status.LatestVersion()
	if version != previousLatestVersion {
		if l.Options.GetSemanticVersioning() {
			if err := l.VerifySemanticVersioning(version, previousLatestVersion, logFrom); err != nil {
				return false, err //nolint:wrapcheck
			}
		}

		return l.handleNewVersion(checkNumber,
			version, releaseDate,
			previousLatestVersion,
			logFrom)
	}

	l.handleNoVersionChange(checkNumber,
		version, logFrom)
	return false, nil
}

// httpRequest makes a HTTP GET request to the URL of this Lookup, and returns the body retrieved.
func (l *Lookup) httpRequest(logFrom logutil.LogFrom) ([]byte, error) {
	req, err := l.createRequest(logFrom)
	if err != nil {
		return nil, err
	}

	resp, body, err := l.getResponse(req, logFrom)
	if err != nil {
		return nil, err
	}

	return l.handleResponse(resp, body, logFrom)
}

// createRequest returns a HTTP GET request to the URL of this Lookup.
func (l *Lookup) createRequest(logFrom logutil.LogFrom) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, l.url(), nil)
	if err != nil {
		err = fmt.Errorf("failed creating http request for %q: %w", l.URL, err)
		logutil.Log.Error(err, logFrom, true)
		return nil, err
	}

	// Set headers.
	req.Header.Set("Connection", "close")
	// Access Token.
	accessToken := l.accessToken()
	if accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	}
	// Conditional requests - https://docs.github.com/en/rest/overview/resources-in-the-rest-api#conditional-requests.
	eTag := l.data.ETag()
	if eTag != "" {
		req.Header.Set("If-None-Match", eTag)
	}

	return req, nil
}

// getResponse will make the request, and return the response, response body, and any errors encountered.
func (l *Lookup) getResponse(req *http.Request, logFrom logutil.LogFrom) (*http.Response, []byte, error) {
	// Make the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		return nil, nil, err //nolint:wrapcheck
	}

	// Read the response body.
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // Limit to 10 MB.
	if err != nil {
		logutil.Log.Error(err, logFrom, true)
		return nil, nil, err //nolint:wrapcheck
	}
	return resp, body, nil
}

// handleResponse processes the HTTP response based on the status code, and takes appropriate action.
//   - 200 OK, if the response body contains one or more releases, it returns the body.
//     Otherwise, it flips the tags flag, and performs a query on the "/tags" endpoint, and returns that body.
//   - 304 Not Modified, it flips the tags flag, performs a query on the "/tags" endpoint, and returns that body.
//   - 401 Unauthorized, 403 Forbidden, and 429 Too Many Requests, it logs the error, and returns a nil body.
//   - unknown status code, it logs the error, and returns a nil body along with an error.
func (l *Lookup) handleResponse(resp *http.Response, body []byte, logFrom logutil.LogFrom) ([]byte, error) {
	switch resp.StatusCode {
	// 200 - Resource has changed.
	case http.StatusOK:
		return l.handleStatusOK(resp, body, logFrom)

	// 304 - Resource has not changed.
	case http.StatusNotModified:
		return l.handleStatusNotModified(logFrom)

	// 401 - Invalid access token.
	case http.StatusUnauthorized:
		return l.handleStatusUnauthorized(body, logFrom)

	// 403 - Rate limit exceeded.
	case http.StatusForbidden:
		return l.handleStatusForbidden(body, logFrom)

	// 429 - Too many requests.
	case http.StatusTooManyRequests:
		return l.handleStatusTooManyRequests(body, logFrom)
	}

	// Unknown status code.
	err := fmt.Errorf("unknown status code %d\n%s", resp.StatusCode, string(body))
	logutil.Log.Error(err, logFrom, true)
	return nil, err
}

// handleStatusOK will handle a 200 status code response,
// and return any errors from the possible tag fallback request.
//
// 200 when the ETag changed.
func (l *Lookup) handleStatusOK(resp *http.Response, body []byte, logFrom logutil.LogFrom) ([]byte, error) {
	newETag := strings.TrimPrefix(resp.Header.Get("etag"), "W/")
	l.data.SetETag(newETag)

	// []byte{91, 93} == []byte("[]") == empty JSON array.
	if len(body) == 2 && bytes.Equal(body, []byte{91, 93}) {
		defaultAccessToken := util.FirstNonDefaultWithEnv(l.Defaults.AccessToken, l.HardDefaults.AccessToken)
		// Update the default empty list ETag if we used the default access_token.
		if l.AccessToken == "" || l.accessToken() == defaultAccessToken {
			setEmptyListETag(newETag)
		}
		// Flip the fallback flag.
		l.data.SetTagFallback()
		if l.data.TagFallback() {
			logutil.Log.Verbose(
				fmt.Sprintf("/releases gave %s, trying /tags", body),
				logFrom, true)
			body, err := l.httpRequest(logFrom)
			return body, err
		}

		// Has tags/releases.
	} else {
		msg := fmt.Sprintf("Potentially found new releases (new ETag %s)", newETag)
		logutil.Log.Verbose(msg, logFrom, true)
	}
	return body, nil
}

// handleStatusNotModified will handle a 304 status code response,
// and return any errors from the possible tag fallback request.
//
// 304 when the ETag is unchanged and the response body is empty.
func (l *Lookup) handleStatusNotModified(logFrom logutil.LogFrom) ([]byte, error) {
	// Didn't find any releases before and nothing has changed?
	if !l.data.hasReleases() {
		// Flip the fallback flag for next time.
		l.data.SetTagFallback()
		if l.data.TagFallback() {
			logutil.Log.Verbose("no tags found on /releases, trying /tags", logFrom, true)
			body, err := l.httpRequest(logFrom)
			return body, err
		}
	}

	// Had releases, use them.
	return nil, nil
}

// handleStatusUnauthorized will handle a 401 status code response.
//
// 401 when the access token is invalid.
func (l *Lookup) handleStatusUnauthorized(body []byte, logFrom logutil.LogFrom) ([]byte, error) {
	bodyStr := string(body)
	var err error

	// Check for invalid access token.
	if strings.Contains(bodyStr, "Bad credentials") {
		err = errors.New("github access token is invalid")
	} else {
		// Unknown error.
		err = fmt.Errorf("unknown 401 response\n%s", bodyStr)
	}
	logutil.Log.Error(err, logFrom, true)

	return nil, err
}

// handleStatusForbidden will handle a 403 status code response.
//
// 403 when the rate limit is exceeded.
func (l *Lookup) handleStatusForbidden(body []byte, logFrom logutil.LogFrom) ([]byte, error) {
	bodyStr := string(body)
	var err error

	switch {
	// Check for rate limit.
	case strings.Contains(bodyStr, "rate limit"):
		err = errors.New("rate limit reached for GitHub")
		logutil.Log.Warn(err, logFrom, true)

		// Missing tag_name.
	case !strings.Contains(bodyStr, `"tag_name"`):
		err = fmt.Errorf("tag_name not found at %s\n%s", l.URL, bodyStr)
		logutil.Log.Error(err, logFrom, true)

		// Other.
	default:
		err = fmt.Errorf("unknown 403 response\n%s", bodyStr)
		logutil.Log.Error(err, logFrom, true)
	}

	return nil, err
}

// handleStatusTooManyRequests handles a 429 status code response.
//
// 429 when too many requests made within a short period.
func (l *Lookup) handleStatusTooManyRequests(body []byte, logFrom logutil.LogFrom) ([]byte, error) {
	var message github_types.Message
	if err := json.Unmarshal(body, &message); err != nil {
		err = fmt.Errorf("unmarshal of GitHub API data failed\n%w", err)
		logutil.Log.Error(err, logFrom, true)
		return nil, errors.New("too many requests made to GitHub")
	}

	return nil, fmt.Errorf("too many requests made to GitHub - %q", message.Message)
}

// releaseMeetsRequirements verifies that the `release` meets the requirements of the Lookup
// and returns the version, and its release date if it does.
func (l *Lookup) releaseMeetsRequirements(release github_types.Release, logFrom logutil.LogFrom) (string, string, error) {
	version := release.TagName
	if release.SemanticVersion != nil {
		version = release.SemanticVersion.String()
	}
	releaseDate := release.PublishedAt

	// No `Require` filters, return the version, and release date.
	if l.Require == nil {
		// Verify that date is in RFC3339 format.
		if _, err := time.Parse(time.RFC3339, releaseDate); err != nil {
			logutil.Log.Warn(
				fmt.Errorf("ignoring release date of %q for version %q on %q as it's not in RFC3339 format\n%w",
					releaseDate, version, l.GetServiceID(), err),
				logFrom, true)
			releaseDate = ""
		}

		return version, releaseDate, nil
	}

	// Check all `Require` filters for this version.
	// Version RegEx.
	if err := l.Require.RegexCheckVersion(version, logFrom); err != nil {
		return "", "", err //nolint: wrapcheck
	}

	// Content RegEx (on assets of release).
	if assetReleaseDate, err := l.Require.RegexCheckContentGitHub(version, release.Assets, logFrom); err != nil {
		return "", "", err //nolint: wrapcheck
	} else if assetReleaseDate != "" {
		releaseDate = assetReleaseDate
	}

	// If the Command didn't return successfully.
	if err := l.Require.ExecCommand(logFrom); err != nil {
		return "", "", err //nolint: wrapcheck
	}

	// If the Docker tag doesn't exist.
	if err := l.Require.DockerTagCheck(version); err != nil {
		if strings.HasSuffix(err.Error(), "\n") {
			err = errors.New(strings.TrimSuffix(err.Error(), "\n"))
		}
		logutil.Log.Warn(err, logFrom, true)
		return "", "", err
		// else if the tag does exist (and we did search for one).
	} else if l.Require.Docker != nil {
		logutil.Log.Info(
			fmt.Sprintf(`found %s container "%s:%s"`,
				l.Require.Docker.GetType(), l.Require.Docker.Image, l.Require.Docker.GetTag(version)),
			logFrom, true)
	}

	// Verify date is in RFC3339 format.
	if _, err := time.Parse(time.RFC3339, releaseDate); err != nil {
		logutil.Log.Warn(
			fmt.Errorf("ignoring release date of %q for version %q on %q as it's not in RFC3339 format\n%w",
				releaseDate, version, l.GetServiceID(), err),
			logFrom, true)
		releaseDate = ""
	}

	return version, releaseDate, nil
}

// getVersion returns the version, and date of the matching asset/release from `body`
// that matches the URLCommands, and Regex requirements.
func (l *Lookup) getVersion(body []byte, logFrom logutil.LogFrom) (string, string, error) {
	// body length = 0 if GitHub ETag unchanged.
	if len(body) != 0 {
		if err := l.setReleases(body, logFrom); err != nil {
			return "", "", fmt.Errorf("release data failed to parse\n%w", err)
		}
	} else {
		// Recheck this ETag's filteredReleases in case filters/releases changed.
		logutil.Log.Verbose("Using cached releases (ETag unchanged)", logFrom, true)
	}
	filteredReleases := l.filterGitHubReleases(logFrom)
	if len(filteredReleases) == 0 {
		return "", "", errors.New("no releases were found matching the url_commands")
	}

	// Check all releases for the one meeting requirements.
	var firstErr error
	for _, release := range filteredReleases {
		if v, rd, err := l.releaseMeetsRequirements(release, logFrom); err == nil {
			return v, rd, nil
		} else if firstErr == nil {
			firstErr = err
		}
	}

	return "", "", fmt.Errorf("no releases were found matching the require field(s)\n%w", firstErr)
}

// setReleases processes, and stores the provided GitHub releases data.
func (l *Lookup) setReleases(body []byte, logFrom logutil.LogFrom) error {
	releases, err := l.checkGitHubReleasesBody(body, logFrom)
	if err != nil {
		return err
	}
	// Store unfiltered releases to support filter changes without a refetch.
	l.data.SetReleases(releases)
	return nil
}

// handleNewVersion handles the case of a new version find,
// and re-checks if first run.
func (l *Lookup) handleNewVersion(
	checkNumber int,
	version,
	releaseDate,
	latestVersion string,
	logFrom logutil.LogFrom,
) (bool, error) {
	// Verify that the version has changed. (GitHub may have just omitted the tag for some reason).
	if checkNumber == 0 {
		msg := fmt.Sprintf("Possibly found a new version (From %q to %q). Checking again", latestVersion, version)
		logutil.Log.Verbose(msg, logFrom, latestVersion != "")
		time.Sleep(time.Second)
		return l.query(logFrom, 1)
	}

	return l.HandleNewVersion(version, releaseDate, logFrom) //nolint: wrapcheck
}

// handleNoVersionChange handles the case of no version(s) being found.
func (l *Lookup) handleNoVersionChange(checkNumber int, version string, logFrom logutil.LogFrom) {
	if checkNumber == 1 {
		logutil.Log.Verbose(
			fmt.Sprintf("Staying on %q as that's the latest version in the second check", version),
			logFrom, true)
	}

	l.Status.AnnounceQuery()
}
