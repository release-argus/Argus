// Copyright [2026] [Argus]
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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/httpx"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
)

// Query queries GitHub releases, recording the result (metrics, DB, announces, logs) only if metrics,
// and returns whether a new version was found.
func (l *Lookup) Query(metrics bool, logFrom logx.LogFrom) (bool, error) {
	isNewVersion, err := l.query(metrics, logFrom)

	if metrics {
		l.QueryMetrics(l, err)
	}

	return isNewVersion, err
}

// query iterates pages of the GitHub releases API and returns whether a new version was found.
func (l *Lookup) query(writeToDB bool, logFrom logx.LogFrom) (bool, error) {
	page := 1
	var newVersion bool
	var err error

	// Query until we find a version or run out of pages.
	for page > 0 {
		newVersion, page, err = l.queryPage(0, page, writeToDB, logFrom)
		if newVersion {
			return true, nil
		}
	}

	return false, err
}

// queryPage requests page of the GitHub releases API and returns whether a new version was found,
// the next page to request, and any error.
func (l *Lookup) queryPage(
	checkNumber int,
	page int,
	writeToDB bool,
	logFrom logx.LogFrom,
) (bool, int, error) {
	nextPage, err := l.httpRequest(page, logFrom)
	if err != nil {
		return false, 0, err
	}

	version, releaseDate, err := l.getVersion(page, logFrom)
	if err != nil {
		logx.Error(err, logFrom, true)
		if nextPage == 0 {
			return false, 0, err
		}
	}
	if version == "" {
		return false, nextPage, nil
	}

	// Found a version.
	// Only set on the first check.
	if checkNumber == 0 {
		l.Status.SetLastQueried("")
	}

	// If this version differs (new?).
	previousLatestVersion := l.Status.LatestVersion()
	if version != previousLatestVersion {
		newVersion, err := l.handleNewVersion(
			checkNumber, page,
			version, releaseDate, l.Status.LatestVersion(),
			writeToDB,
			logFrom,
		)
		return newVersion, 0, err
	}

	// Inherit version.
	l.HandleNoVersionChange(checkNumber, version, writeToDB, logFrom)
	return false, 0, nil
}

// httpRequest makes a HTTP GET request to the address of this Lookup, stores any releases
// retrieved, and returns the next page to request.
func (l *Lookup) httpRequest(page int, logFrom logx.LogFrom) (int, error) {
	req, err := l.createRequest(page, logFrom)
	if err != nil {
		return 0, err
	}

	resp, body, err := l.getResponse(req, logFrom)
	if err != nil {
		return 0, err
	}

	return l.handleResponse(resp, body, page, logFrom)
}

// createRequest returns a HTTP GET request to the address of this Lookup.
func (l *Lookup) createRequest(page int, logFrom logx.LogFrom) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, l.url(page), nil)
	if err != nil {
		err = fmt.Errorf(
			"failed creating http request for %q: %w",
			l.URL, err,
		)
		logx.Error(err, logFrom, true)
		return nil, err
	}

	// Access Token.
	if accessToken := l.accessToken(); accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	// Conditional requests - https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api?apiVersion=2022-11-28#use-conditional-requests-if-appropriate.
	if etag := l.data.ETag(); etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	return req, nil
}

// getResponse makes the request and returns the response, response body, and any errors encountered.
func (l *Lookup) getResponse(req *http.Request, logFrom logx.LogFrom) (*http.Response, []byte, error) {
	// Make the request.
	resp, err := httpx.Client.Do(req)
	if err != nil {
		logx.Error(err, logFrom, true)
		return nil, nil, err //nolint:wrapcheck
	}
	logx.Debug("GET "+req.URL.String(), logFrom, true)

	// Read the response body.
	defer resp.Body.Close()
	body, err := httpx.ReadBody(resp.Body)
	if err != nil {
		logx.Error(err, logFrom, true)
		return nil, nil, err //nolint:wrapcheck
	}
	return resp, body, nil
}

// handleResponse processes the HTTP response based on the status code and takes appropriate action.
//   - 200 OK, it stores the releases with the ETag and next page of the response.
//     An empty first page conditionally falls back to the "/tags" endpoint (see [Lookup.useTagsAPI]).
//   - 304 Not Modified, conditionally falls back to "/tags" when there are no cached releases,
//     and otherwise returns the cached response's next page.
//   - 401 Unauthorized, 403 Forbidden, and 429 Too Many Requests, it logs and returns the error.
//   - Unknown status code, it logs and returns the error.
func (l *Lookup) handleResponse(resp *http.Response, body []byte, page int, logFrom logx.LogFrom) (int, error) {
	switch resp.StatusCode {
	// 200 - Resource has changed.
	case http.StatusOK:
		return l.handleStatusOK(resp, body, page, logFrom)

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
	err := fmt.Errorf(
		"unknown status code %d%s",
		resp.StatusCode, forge.BodyExcerpt(body),
	)
	logx.Error(err, logFrom, true)
	return 0, err
}

// handleStatusOK processes a 200 status code response, storing its releases,
// and returns any errors from parsing them or from the possible tag fallback request.
//
// 200 when the ETag changed.
func (l *Lookup) handleStatusOK(resp *http.Response, body []byte, page int, logFrom logx.LogFrom) (int, error) {
	newETag := strings.TrimPrefix(resp.Header.Get("etag"), "W/")

	// []byte{91, 93} == []byte("[]") == empty JSON array.
	if len(body) == 2 && bytes.Equal(body, []byte{91, 93}) {
		defaultAccessToken := util.FirstNonDefaultWithEnv(l.typeDefaults.AccessToken, l.typeHardDefaults.AccessToken)
		firstPage := page <= 1
		// Update the default empty list ETag if we used the default access_token.
		if firstPage && (l.AccessToken == "" || l.AccessToken == defaultAccessToken) {
			setEmptyListETag(newETag)
		}

		if firstPage && l.useTagsAPI() {
			l.data.SetTagFallback()
			if l.data.TagFallback() {
				logx.Verbose(
					fmt.Sprintf("/releases gave %s, trying /tags", body),
					logFrom,
					true,
				)
				return l.httpRequest(0, logFrom)
			}
		}

		// Has tags/releases.
	} else {
		msg := fmt.Sprintf("Potentially found new releases (new ETag %s)", newETag)
		logx.Verbose(msg, logFrom, true)
	}

	nextPage := forge.NextPage(resp.Header.Get("Link"))
	if err := l.setReleases(body, newETag, nextPage); err != nil {
		logx.Error(err, logFrom, true)
		return 0, err
	}
	return nextPage, nil
}

// handleStatusNotModified processes a 304 status code response
// and returns any errors from the possible tag fallback request.
//
// 304 when the ETag is unchanged, and the response body is empty.
func (l *Lookup) handleStatusNotModified(logFrom logx.LogFrom) (int, error) {
	// Didn't find any releases before and nothing has changed?
	if !l.data.hasReleases() {
		if l.useTagsAPI() {
			// Flip the fallback flag for next time.
			l.data.SetTagFallback()
			if l.data.TagFallback() {
				logx.Verbose("no tags found on /releases, trying /tags", logFrom, true)
				return l.httpRequest(0, logFrom)
			}
		}
		return 0, nil
	}

	logx.Verbose("Using cached releases (ETag unchanged)", logFrom, true)
	return l.data.NextPage(), nil
}

// handleStatusUnauthorized processes a 401 status code response.
//
// 401 when the access token is invalid.
func (l *Lookup) handleStatusUnauthorized(body []byte, logFrom logx.LogFrom) (int, error) {
	bodyStr := string(body)
	var err error

	// Check for invalid access token.
	if strings.Contains(bodyStr, "Bad credentials") {
		err = errors.New("github access token is invalid")
	} else {
		// Unknown error.
		err = errors.New("unknown 401 response\n" + bodyStr)
	}
	logx.Error(err, logFrom, true)

	return 0, err
}

// handleStatusForbidden processes a 403 status code response.
//
// 403 when the rate limit is exceeded.
func (l *Lookup) handleStatusForbidden(body []byte, logFrom logx.LogFrom) (int, error) {
	bodyStr := string(body)
	var err error

	switch {
	// Check for rate limit.
	case strings.Contains(bodyStr, "rate limit"):
		err = errors.New("rate limit reached for GitHub")
		logx.Warn(err, logFrom, true)

		// Missing tag_name.
	case !strings.Contains(bodyStr, `"tag_name"`):
		err = fmt.Errorf(
			"tag_name not found at %s\n%s",
			l.URL, bodyStr,
		)
		logx.Error(err, logFrom, true)

		// Other.
	default:
		err = errors.New("unknown 403 response\n" + bodyStr)
		logx.Error(err, logFrom, true)
	}

	return 0, err
}

// handleStatusTooManyRequests processes a 429 status code response.
//
// 429 when too many requests made within a short period.
func (l *Lookup) handleStatusTooManyRequests(body []byte, logFrom logx.LogFrom) (int, error) {
	var message forgetypes.Message
	if err := decode.Unmarshal("json", body, &message); err != nil {
		err = fmt.Errorf("unmarshal of GitHub API data failed: %w", err)
		logx.Error(err, logFrom, true)
		return 0, errors.New("too many requests made to GitHub")
	}

	return 0, fmt.Errorf("too many requests made to GitHub - %q", message.Message)
}

// getVersion returns the version and date of the first stored release
// that matches the URLCommands and Require.
func (l *Lookup) getVersion(page int, logFrom logx.LogFrom) (string, string, error) {
	filteredReleases := l.filterGitHubReleases(logFrom)
	if len(filteredReleases) == 0 {
		return "", "", fmt.Errorf(
			"no releases were found matching the url_commands on page %d of the API response",
			page,
		)
	}

	// Check all releases for the one meeting requirements.
	return forge.FirstReleaseMeetingRequirements( //nolint:wrapcheck
		filteredReleases, l.Require, l.GetServiceID(), logFrom,
	)
}

// setReleases parses body and stores its releases with the etag and nextPage of its response.
func (l *Lookup) setReleases(body []byte, etag string, nextPage int) error {
	releases, err := l.unmarshalGitHubReleasesBody(body)
	if err != nil {
		return fmt.Errorf("release data failed to parse: %w", err)
	}

	// Store unfiltered releases to support filter changes without a refetch.
	l.data.SetResponse(etag, nextPage, releases)
	return nil
}

// handleNewVersion processes the case of a new version find,
// and re-checks page if first run.
func (l *Lookup) handleNewVersion(
	checkNumber, page int,
	version, releaseDate, latestVersion string,
	writeToDB bool,
	logFrom logx.LogFrom,
) (bool, error) {
	// Confirm that the version has changed.
	if checkNumber == 0 {
		msg := fmt.Sprintf(
			"Possibly found a new version (From %q to %q). Checking again",
			latestVersion, version,
		)
		logx.Verbose(msg, logFrom, writeToDB && latestVersion != "")
		time.Sleep(time.Second)

		newVersion, _, err := l.queryPage(1, page, writeToDB, logFrom)
		return newVersion, err
	}

	return l.HandleNewVersion(version, releaseDate, writeToDB, logFrom) //nolint:wrapcheck
}
