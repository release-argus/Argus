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

package forgejo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/release-argus/Argus/internal/httpx"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
)

// API endpoints queried for versions, relative to a repository.
const (
	endpointReleases = "releases"
	endpointTags     = "tags"
)

var (
	// errNoneFound signals that an endpoint responded with an empty list.
	errNoneFound = errors.New("empty list")
	// errNotFound signals a 404.
	errNotFound = errors.New("not found")
)

// Query queries the forge for releases, sets Prometheus metrics if requested, and returns whether a
// new version was found.
func (l *Lookup) Query(metrics bool, logFrom logx.LogFrom) (bool, error) {
	isNewVersion, err := l.query(logFrom)

	if metrics {
		l.QueryMetrics(l, err)
	}

	return isNewVersion, err
}

// query queries [endpointReleases], falling back to [endpointTags] for
// repositories that publish no releases.
func (l *Lookup) query(logFrom logx.LogFrom) (bool, error) {
	newVersion, releasesErr := l.queryEndpoint(endpointReleases, logFrom)
	if !endpointYieldedNothing(releasesErr) {
		return newVersion, releasesErr
	}

	if !l.useTagsAPI() {
		return false, nothingFound(releasesErr, nil, false)
	}

	logx.Verbose("/releases gave nothing, trying /tags", logFrom, true)
	newVersion, tagsErr := l.queryEndpoint(endpointTags, logFrom)
	if !endpointYieldedNothing(tagsErr) {
		return newVersion, tagsErr
	}

	return false, nothingFound(releasesErr, tagsErr, true)
}

// endpointYieldedNothing reports whether err means the endpoint held nothing to filter.
func endpointYieldedNothing(err error) bool {
	return errors.Is(err, errNoneFound) || errors.Is(err, errNotFound)
}

// nothingFound describes a repository that yielded no version.
//
// A 404 from every endpoint tried is ambiguous on a forge.
// The repository may not exist, or the feature may be disabled.
// A 200 from any of them proves the repository does exist.
func nothingFound(releasesErr, tagsErr error, triedTags bool) error {
	releasesMissing := errors.Is(releasesErr, errNotFound)
	tagsMissing := triedTags && errors.Is(tagsErr, errNotFound)

	switch {
	case releasesMissing && !triedTags:
		return errors.New("repository not found, or its releases are disabled")
	case releasesMissing && tagsMissing:
		return errors.New("repository not found, or its releases and tags are disabled")
	case releasesMissing:
		return errors.New("no tags were found, and the repository's releases are disabled")
	case tagsMissing:
		return errors.New("no releases were found, and the repository's tags are disabled")
	case triedTags:
		return errors.New("no releases or tags were found")
	}
	return errors.New("no releases were found")
}

// maxPages bounds a single endpoint walk.
const maxPages = 100

// queryEndpoint iterates pages of `endpoint` and returns whether a new version was found.
func (l *Lookup) queryEndpoint(endpoint string, logFrom logx.LogFrom) (bool, error) {
	page := 1
	var newVersion bool
	var err error

	// Query until we find a version, or run out of pages.
	for pagesRead := 0; page > 0; pagesRead++ {
		if pagesRead == maxPages {
			return false, fmt.Errorf(
				"gave up walking /%s after %d pages",
				endpoint, maxPages,
			)
		}

		var nextPage int
		newVersion, nextPage, err = l.queryPage(0, page, endpoint, logFrom)
		if newVersion {
			return true, nil
		}

		// The forge names the next page, so only follow one that moves forward.
		if nextPage > 0 && nextPage <= page {
			return false, fmt.Errorf(
				"/%s did not advance past page %d",
				endpoint, page,
			)
		}
		page = nextPage
	}

	return false, err
}

// queryPage requests `page` of `endpoint` and returns whether a new version was found,
// the next page to request, and any error.
func (l *Lookup) queryPage(
	checkNumber int,
	page int,
	endpoint string,
	logFrom logx.LogFrom,
) (bool, int, error) {
	body, nextPage, err := l.httpRequest(endpoint, page, logFrom)
	if err != nil {
		return false, 0, err
	}

	version, releaseDate, err := l.getVersion(body, page, logFrom)
	if err != nil {
		logx.Error(err, logFrom, true)
		if nextPage == 0 {
			return false, 0, err
		}
	}
	if version == "" {
		return false, nextPage, nil
	}

	// Only set on the first check.
	if checkNumber == 0 {
		l.Status.SetLastQueried("")
	}

	// If this version differs (new?).
	previousLatestVersion := l.Status.LatestVersion()
	if version != previousLatestVersion {
		newVersion, err := l.handleNewVersion(
			checkNumber,
			version, releaseDate, previousLatestVersion,
			endpoint, page,
			logFrom,
		)
		return newVersion, 0, err
	}

	// Inherit version.
	l.handleNoVersionChange(checkNumber, version, logFrom)
	return false, 0, nil
}

// httpRequest makes a HTTP GET request for `page` of `endpoint` and returns the body retrieved,
// along with the next page to request.
func (l *Lookup) httpRequest(endpoint string, page int, logFrom logx.LogFrom) ([]byte, int, error) {
	req, err := l.createRequest(endpoint, page, logFrom)
	if err != nil {
		return nil, 0, err
	}

	resp, body, err := getResponse(req, logFrom)
	if err != nil {
		return nil, 0, err
	}

	return handleResponse(resp, body, page, logFrom)
}

// createRequest returns a HTTP GET request for `page` of `endpoint`.
func (l *Lookup) createRequest(endpoint string, page int, logFrom logx.LogFrom) (*http.Request, error) {
	address, err := l.apiURL(endpoint, page)
	if err != nil {
		logx.Error(err, logFrom, true)
		return nil, err
	}

	return l.requestFor(address, logFrom)
}

// requestFor returns a HTTP GET request for `address`.
func (l *Lookup) requestFor(address string, logFrom logx.LogFrom) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		err = fmt.Errorf(
			"failed creating http request for %q: %w",
			address, err,
		)
		logx.Error(err, logFrom, true)
		return nil, err
	}

	return req, nil
}

// getResponse makes the request and returns the response, response body, and any errors encountered.
func getResponse(req *http.Request, logFrom logx.LogFrom) (*http.Response, []byte, error) {
	// Make the request.
	resp, err := httpx.Client.Do(req)
	if err != nil {
		logx.Error(err, logFrom, true)
		return nil, nil, err //nolint:wrapcheck
	}
	logx.Debug("GET "+req.URL.String(), logFrom, true)

	// Read the response body.
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20)) // Limit to 50 MiB.
	if err != nil {
		logx.Error(err, logFrom, true)
		return nil, nil, err //nolint:wrapcheck
	}
	return resp, body, nil
}

// handleResponse maps the HTTP response to a body, the next page to request, and any error.
//   - 200 OK, returns the body, treating an empty first page as [errNoneFound].
//   - 401 Unauthorized/403 Forbidden, an authentication failure.
//   - 403 Forbidden/429 Too Many Requests, Rate-limited - detected by status code/header presence.
//   - 404 Not Found, [errNotFound] - the repository is missing, or the feature is disabled on it.
//   - Unknown status code, it logs the error and returns a nil body along with an error.
func handleResponse(
	resp *http.Response,
	body []byte,
	page int,
	logFrom logx.LogFrom,
) ([]byte, int, error) {
	// 403/429 - Possibly rate-limited.
	if isRateLimited(resp) {
		return handleRateLimited(resp, logFrom)
	}

	switch resp.StatusCode {
	// 200 - Have releases/tags, or an empty list.
	case http.StatusOK:
		return handleStatusOK(resp, body, page)

	// 401/403 - Instance or repository requires authentication.
	case http.StatusUnauthorized, http.StatusForbidden:
		return handleStatusUnauthorized(resp, logFrom)

	// 404 - No such repository, or the feature is disabled on it.
	case http.StatusNotFound:
		logx.Verbose(resp.Request.URL.Path+" gave 404", logFrom, true)
		return nil, 0, errNotFound
	}

	// Unknown status code.
	err := fmt.Errorf("unknown status code %d\n%s", resp.StatusCode, string(body))
	logx.Error(err, logFrom, true)
	return nil, 0, err
}

// handleStatusOK processes a 200 status code response.
func handleStatusOK(resp *http.Response, body []byte, page int) ([]byte, int, error) {
	if page <= 1 && isEmptyJSONList(body) {
		return nil, 0, errNoneFound
	}

	return body, getNextPage(resp.Header.Get("Link")), nil
}

// isEmptyJSONList reports whether body holds an empty JSON list.
func isEmptyJSONList(body []byte) bool {
	return bytes.Equal(bytes.TrimSpace(body), []byte("[]"))
}

// rateLimitHeaders are set by a limiter in front of the forge.
var rateLimitHeaders = []string{
	"RateLimit",
	"RateLimit-Policy",
	"X-RateLimit-Remaining",
}

// isRateLimited reports whether resp is a rate-limited response.
func isRateLimited(resp *http.Response) bool {
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}

	if resp.StatusCode == http.StatusForbidden {
		for _, header := range rateLimitHeaders {
			if resp.Header.Get(header) != "" {
				return true
			}
		}
	}

	return false
}

// handleRateLimited processes a rate-limited response.
func handleRateLimited(resp *http.Response, logFrom logx.LogFrom) ([]byte, int, error) {
	err := fmt.Errorf("rate limit reached for %s", resp.Request.URL.Host)
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		err = fmt.Errorf(
			"rate limit reached for %s - retry after %s",
			resp.Request.URL.Host, retryAfter,
		)
	}
	logx.Warn(err, logFrom, true)

	return nil, 0, err
}

// handleStatusUnauthorized processes a 401/403 status code response.
func handleStatusUnauthorized(resp *http.Response, logFrom logx.LogFrom) ([]byte, int, error) {
	err := fmt.Errorf(
		"authentication failed for %s (%d)",
		resp.Request.URL.Host, resp.StatusCode,
	)
	logx.Error(err, logFrom, true)

	return nil, 0, err
}

// getNextPage returns the next page number from the Link header.
//
// Example:
//
//	<https://codeberg.org/api/v1/repos/OWNER/REPO/releases?limit=50&page=2>; rel="next",
//	<https://codeberg.org/api/v1/repos/OWNER/REPO/releases?limit=50&page=3>; rel="last"
//
// Output:
//
//	2
//
// If the Link header does not include a next page link, it returns 0.
func getNextPage(linkHeader string) int {
	re := regexp.MustCompile(`<[^>]+page=(\d+)[^>]*>;\s*rel="next"`)

	if matches := re.FindStringSubmatch(linkHeader); matches != nil {
		pageNum, _ := strconv.Atoi(matches[1])
		return pageNum
	}

	return 0 // No next page found.
}

// getVersion returns the version and date of the matching asset/release from `body`
// that matches the URLCommands, and RegEx requirements.
func (l *Lookup) getVersion(body []byte, page int, logFrom logx.LogFrom) (string, string, error) {
	releases, err := forge.UnmarshalReleases(body)
	if err != nil {
		return "", "", fmt.Errorf("release data failed to parse: %w", err)
	}

	filteredReleases := l.filterReleases(releases, logFrom)
	if len(filteredReleases) == 0 {
		return "", "", fmt.Errorf(
			"no releases were found matching the url_commands on page %d of the API response",
			page,
		)
	}

	var firstErr error
	for _, release := range filteredReleases {
		v, rd, err := forge.ReleaseMeetsRequirements(release, l.Require, l.GetServiceID(), logFrom)
		if err == nil {
			return v, rd, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	return "", "", fmt.Errorf("no releases were found matching the require fields %w", firstErr)
}

// filterReleases filters `releases` against the URLCommands, semantic-versioning and
// pre-release settings of the receiver.
//
// See [forge.FilterReleases].
func (l *Lookup) filterReleases(
	releases []forgetypes.Release,
	logFrom logx.LogFrom,
) []forgetypes.Release {
	return forge.FilterReleases(
		releases,
		forge.FilterOptions{
			URLCommands:        l.URLCommands,
			SemanticVersioning: l.Options.GetSemanticVersioning(),
			UsePreReleases:     l.usePreRelease(),
		},
		logFrom,
	)
}

// handleNewVersion processes the case of a new version find,
// and re-checks validity if first run.
func (l *Lookup) handleNewVersion(
	checkNumber int,
	version, releaseDate, latestVersion string,
	endpoint string,
	page int,
	logFrom logx.LogFrom,
) (bool, error) {
	// Confirm that the version has changed.
	if checkNumber == 0 {
		msg := fmt.Sprintf(
			"Possibly found a new version (From %q to %q). Checking again",
			latestVersion, version,
		)
		logx.Verbose(msg, logFrom, latestVersion != "")
		time.Sleep(time.Second)

		newVersion, _, err := l.queryPage(1, page, endpoint, logFrom)
		return newVersion, err
	}

	return l.HandleNewVersion(version, releaseDate, logFrom) //nolint:wrapcheck
}

// handleNoVersionChange processes the case of no new versions found on a re-check.
func (l *Lookup) handleNoVersionChange(checkNumber int, version string, logFrom logx.LogFrom) {
	if checkNumber == 1 {
		logx.Verbose(
			fmt.Sprintf("Staying on %q as that's the latest version in the second check", version),
			logFrom,
			true,
		)
	}

	l.Status.AnnounceQuery()
}
