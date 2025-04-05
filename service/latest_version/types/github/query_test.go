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

//go:build unit

// Package github provides a github-based lookup type.
package github

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/latest_version/filter"
	github_types "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestQuery(t *testing.T) {
	tLookup := testLookup(false)
	tLookup.URL = "release-argus/.github"
	tLookup.Query(false, logutil.LogFrom{})
	emptyReleasesETag := tLookup.data.eTag

	type statusVars struct {
		deployedVersion   string
		latestVersion     string
		wantLatestVersion *string
	}
	type want struct {
		stdoutRegex string
		errRegex    string
	}
	// GIVEN a Lookup.
	tests := map[string]struct {
		overrides             string
		overrideETag          *string
		nonSemanticVersioning bool
		status                statusVars
		want                  want
	}{
		"invalid url": {
			overrides: test.TrimYAML(`
				url: "release-argus	Argus"
			`),
			want: want{
				errRegex: `invalid control character in URL`},
		},
		"query that gets a non-semantic version - not allowed": {
			overrides: test.TrimYAML(`
				url: release-argus/test
				url_commands:
					- type: regex
						regex: 'ver[0-9.]+'
			`),
			want: want{
				errRegex: `no releases were found matching the url_commands`},
		},
		"query that gets a non-semantic version - allowed": {
			overrides: test.TrimYAML(`
				url: release-argus/test
				url_commands:
					- type: regex
						regex: 'ver[0-9.]+'
			`),
			nonSemanticVersioning: true,
			status: statusVars{
				wantLatestVersion: test.StringPtr("ver1.1.1")},
			want: want{
				errRegex: `^$`},
		},
		"changed to semantic_versioning but had a non-semantic deployed_version": {
			status: statusVars{
				deployedVersion: "1.2.3.4"},
			want: want{
				errRegex: `^$`},
		},
		"regex_content mismatch": {
			overrides: test.TrimYAML(`
				require:
					regex_content: "argus[0-9]+.exe"
			`),
			want: want{
				stdoutRegex: `regex "[^"]+" not matched on content for version`,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
					regex "[^"]+" not matched on content for version "[^"]+"$`)},
		},
		"regex_content match": {
			overrides: test.TrimYAML(`
				require:
					regex_content: "{{ version }}.linux-amd64"
			`),
			nonSemanticVersioning: true,
			want: want{
				errRegex: `^$`},
		},
		"command fail": {
			overrides: test.TrimYAML(`
				require:
					command: ["false"]
			`),
			want: want{
				stdoutRegex: `exit status 1`,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
					command failed: .*`)},
		},
		"command pass": {
			overrides: test.TrimYAML(`
				require:
					command: ["true"]
			`),
			want: want{
				errRegex: `^$`},
		},
		"docker tag mismatch": {
			overrides: test.TrimYAML(`
				require:
					docker:
						type: ghcr
						image: release-argus/argus
						tag: 0.9.0-beta
						token: ` + os.Getenv("GH_TOKEN") + `
			`),
			want: want{
				stdoutRegex: `manifest unknown`,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
					release-argus/argus:0.9.0-beta .*`)},
		},
		"docker tag match": {
			overrides: test.TrimYAML(`
				require:
					docker:
						type: ghcr
						image: release-argus/argus
						tag: 0.9.0
						token: ` + os.Getenv("GH_TOKEN") + `
			`),
			want: want{
				errRegex: `^$`},
		},
		"regex_version mismatch": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "ver([0-9.]+)"
			`),
			want: want{
				stdoutRegex: `regex "[^"]+" not matched on version`,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
					regex "[^"]+" not matched on version "[^"]+"$`)},
		},
		"urlCommand regex mismatch": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: 'x([0-9.]+)'
			`),
			want: want{
				errRegex:    `no releases were found matching the url_commands`,
				stdoutRegex: `regex "[^"]+" didn't return any matches`},
		},
		"valid semantic version query": {
			overrides: test.TrimYAML(`
				url: release-argus/test
				url_commands:
					- type: regex
						regex: 'ver([0-9.]+)'
			`),
			want: want{
				errRegex: `^$`},
		},
		"older version found": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '([0-9.]+)'
			`),
			status: statusVars{
				latestVersion:   "0.0.0",
				deployedVersion: "9.9.9"},
			want: want{
				errRegex: `queried version .* is less than the deployed version`},
		},
		"newer version found": {
			overrides: test.TrimYAML(`
				url: release-argus/test
				url_commands:
					- type: regex
						regex: '([0-9.]+)'
			`),
			status: statusVars{
				latestVersion:     "1.1.0",
				wantLatestVersion: test.StringPtr("1.1.1")},
			want: want{
				stdoutRegex: `New Release - "[^"]+"`,
				errRegex:    `^$`},
		},
		"same version found": {
			overrides: test.TrimYAML(`
				url: release-argus/test
				url_commands:
					- type: regex
						regex: '([0-9.]+)'
			`),
			status: statusVars{
				latestVersion:     "1.1.1",
				wantLatestVersion: test.StringPtr("1.1.1")},
			want: want{
				errRegex: `^$`},
		},
		"repo that uses tags, not releases - has tags": {
			overrides: test.TrimYAML(`
				url: "release-argus/test"
			`),
			status: statusVars{
				wantLatestVersion: test.StringPtr("1.1.1")},
			want: want{
				errRegex:    `^$`,
				stdoutRegex: `no tags found on /releases, trying /tags`},
		},
		"repo that uses tags, not releases - no tags": {
			overrides: test.TrimYAML(`
				url: "release-argus/.github"
			`),
			overrideETag: &emptyReleasesETag,
			want: want{
				errRegex:    `no releases were found matching the url_commands`,
				stdoutRegex: `no tags found on /releases, trying /tags`},
		},
		"repo that uses tags, not releases - no tags - emptyListETag changed": {
			overrides: test.TrimYAML(`
				url: "release-argus/.github"
			`),
			overrideETag: test.StringPtr(""),
			want: want{
				errRegex:    `no releases were found`,
				stdoutRegex: `/releases gave .*, trying /tags`},
		},
		"no access token": {
			overrides: test.TrimYAML(`
				access_token: null
			`),
			want: want{
				errRegex: `^$`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			try := 0
			temporaryFailureInNameResolution := true
			for temporaryFailureInNameResolution != false {
				releaseStdout := test.CaptureStdout()
				try++
				temporaryFailureInNameResolution = false
				lookup := testLookup(false)
				if strings.Contains(tc.overrides, "access_token: null") {
					lookup.HardDefaults.AccessToken = ""
				}
				lookup.Status.ServiceID = &name
				err := yaml.Unmarshal([]byte(tc.overrides), lookup)
				if err != nil {
					t.Fatalf("%s\nfailed to unmarshal overrides: %v",
						packageName, err)
				}
				*lookup.Options.SemanticVersioning = !tc.nonSemanticVersioning
				lookup.Status.SetLatestVersion(tc.status.latestVersion, "", false)
				lookup.Status.SetDeployedVersion(tc.status.deployedVersion, "", false)
				// In case require is non-nil, Init to give it Status.
				lookup.Init(
					lookup.Options,
					lookup.Status,
					lookup.Defaults, lookup.HardDefaults)
				// Clear ETag/Releases if URL changed.
				if tc.overrideETag != nil {
					lookup.data.eTag = *tc.overrideETag
				}

				// WHEN Query is called on it.
				_, err = lookup.Query(true, logutil.LogFrom{})

				// THEN any err is expected.
				stdout := releaseStdout()
				e := util.ErrorToString(err)
				if !util.RegexCheck(tc.want.errRegex, e) {
					if strings.Contains(e, "context deadline exceeded") {
						temporaryFailureInNameResolution = true
						if try != 3 {
							time.Sleep(time.Second)
							continue
						}
					}
					t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.want.errRegex, e)
				}
				// AND the stdout contains the expected strings.
				if !util.RegexCheck(tc.want.stdoutRegex, stdout) {
					t.Fatalf("%s\nstdout mismatch\n%q\nnot found in:\n%q",
						packageName, tc.want.stdoutRegex, stdout)
				}
				// AND the LatestVersion is as expected.
				if tc.status.wantLatestVersion != nil &&
					*tc.status.wantLatestVersion != lookup.Status.LatestVersion() {
					t.Fatalf("%s\nLatestVersion mismatch\nwant: %q\ngot:  %q",
						packageName, *tc.status.wantLatestVersion, lookup.Status.LatestVersion())
				}
			}
		})
	}
}

func TestHTTPRequest(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		failing  bool
		url      string
		eTag     *string
		errRegex string
	}{
		"invalid url": {
			url:      "invalid://	test",
			errRegex: `invalid control character in URL`},
		"unknown url": {
			url:      "https://release-argus.invalid-tld",
			errRegex: `no such host`},
		"valid url": {
			url:      "release-argus/Argus",
			errRegex: `^$`},
		"repo that uses tags, not releases - has tags": {
			url:      "release-argus/test",
			errRegex: `^$`},
		"repo that uses tags, not releases - no tags": {
			url:      "release-argus/.github",
			errRegex: `^$`},
		"repo that uses tags, not releases - update EmptyListETag if 200 on empty list": {
			url:      "release-argus/.github",
			eTag:     test.StringPtr(""),
			errRegex: `^$`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.URL = tc.url
			if tc.eTag != nil {
				lookup.data.eTag = *tc.eTag
			}

			// WHEN httpRequest is called on it.
			_, err := lookup.httpRequest(logutil.LogFrom{})

			// THEN any err is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestGetResponse_ReadError(t *testing.T) {
	// GIVEN a server that closes the connection immediately to simulate a read error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10") // Set Content-Length without sending data.
		w.WriteHeader(http.StatusOK)
		// Immediately close the connection without writing any body.
		conn, _, _ := w.(http.Hijacker).Hijack()
		conn.Close()
	}))
	t.Cleanup(func() { server.Close() })

	// AND a request to the mock server's URL.
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("%s\ncould not create request: %v",
			packageName, err)
	}

	// WHEN getResponse is called on that URL.
	l := Lookup{}
	_, _, err = l.getResponse(req, logutil.LogFrom{})

	// THEN an error is expected from the read error.
	if err == nil {
		t.Fatalf("%s\nexpected an error when reading response body, got none",
			packageName)
	}
}

func TestHandleResponse(t *testing.T) {
	type wants struct {
		nilBody          bool
		errRegex         string
		setEmptyListETag bool
	}
	type conditions struct {
		hadReleases           bool
		hadDefaultAccessToken bool
	}

	// GIVEN a HTTP Response and an accompanying body.
	tests := map[string]struct {
		conditions conditions
		statusCode int
		body       []byte
		want       wants
	}{
		"200 OK - EmptyListETag set if default access_token": {
			conditions: conditions{
				hadDefaultAccessToken: true},
			statusCode: http.StatusOK,
			body:       []byte(`[]`),
			want: wants{
				setEmptyListETag: true},
		},
		"200 OK - EmptyListETag not set if non-default access_token": {
			conditions: conditions{
				hadDefaultAccessToken: false},
			statusCode: http.StatusOK,
			body:       []byte(`[]`),
			want: wants{
				setEmptyListETag: false},
		},
		"200 OK - Get releases": {
			statusCode: http.StatusOK,
			body:       []byte(`[{"tag_name":"v1.0.0"}]`),
		},
		"304 Not Modified - Tag fallback request": {
			statusCode: http.StatusNotModified,
			want: wants{
				nilBody: false},
		},
		"304 Not Modified - Use old releases": {
			conditions: conditions{
				hadReleases: true},
			statusCode: http.StatusNotModified,
			want: wants{
				nilBody: true},
		},
		"401 Unauthorized - Bad credentials": {
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"message":"Bad credentials"}`),
			want: wants{
				errRegex: `github access token is invalid`,
				nilBody:  true},
		},
		"401 Unauthorized - Unknown": {
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"message":"Something else"}`),
			want: wants{
				errRegex: `unknown 401 response`,
				nilBody:  true},
		},
		"403 Forbidden - rate limit": {
			statusCode: http.StatusForbidden,
			body:       []byte(`{"message":"API rate limit exceeded"}`),
			want: wants{
				errRegex: `rate limit reached for GitHub`,
				nilBody:  true},
		},
		"403 Forbidden - missing tag_name": {
			statusCode: http.StatusForbidden,
			body:       []byte(`{"message":"some other error"}`),
			want: wants{
				errRegex: `tag_name not found at`,
				nilBody:  true},
		},
		"403 Forbidden - unknown": {
			statusCode: http.StatusForbidden,
			body:       []byte(`[{"tag_name":"v1.0.0"}]`),
			want: wants{
				errRegex: `unknown 403 response`,
				nilBody:  true},
		},
		"429 Too Many Requests - rate limit": {
			statusCode: http.StatusTooManyRequests,
			body:       []byte(`{"message":"something from GitHub"}`),
			want: wants{
				errRegex: `^too many requests made to GitHub - "something from GitHub"$`,
				nilBody:  true},
		},
		"429 Too Many Requests - unmarshal fail": {
			statusCode: http.StatusTooManyRequests,
			body:       []byte(`{"message":}`),
			want: wants{
				errRegex: `^too many requests made to GitHub$`,
				nilBody:  true},
		},
		"Unknown status code": {
			statusCode: http.StatusTeapot,
			body:       []byte(`{"message":"I'm a teapot"}`),
			want: wants{
				errRegex: `unknown status code 418`,
				nilBody:  true},
		},
	}

	// Ensure other tests that modify global state don't interfere.
	releaseStdout := test.CaptureStdout()
	defer releaseStdout()
	hadEmptyListETag := getEmptyListETag()
	t.Cleanup(func() { setEmptyListETag(hadEmptyListETag) })

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're modifying global state.

			lookup := testLookup(false)
			resp := &http.Response{
				StatusCode: tc.statusCode,
				Header:     http.Header{},
			}
			hadETag := name
			resp.Header.Add("ETag", hadETag)
			if tc.conditions.hadReleases {
				lookup.data.releases = testBodyObject
			}
			lookup.AccessToken = lookup.accessToken()
			if !tc.conditions.hadDefaultAccessToken {
				lookup.Defaults.AccessToken = ""
				lookup.HardDefaults.AccessToken = "Something"
			}

			logFrom := logutil.LogFrom{Primary: "TestHandleResponse", Secondary: name}

			// WHEN handleResponse is called on it.
			gotBody, err := lookup.handleResponse(resp, tc.body, logFrom)

			// THEN any err is expected.
			if tc.want.errRegex == "" && err != nil {
				t.Errorf("%s\nunexpected error: %v",
					packageName, err)
			} else if !util.RegexCheck(tc.want.errRegex, util.ErrorToString(err)) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want.errRegex, util.ErrorToString(err))
			}
			// AND the body returned is as expected.
			if tc.want.nilBody && len(gotBody) != 0 {
				t.Errorf("%s\nbody mismatch\nwant: nil\ngot:  %q",
					packageName, string(tc.body))
			} else if !tc.want.nilBody && len(gotBody) == 0 {
				t.Errorf("%s\nbody mismatch\nwant: non-nil\ngot:  nil",
					packageName)
			}
			// AND the new EmptyListETag is as expected.
			emptyListETag := getEmptyListETag()
			if tc.want.setEmptyListETag && emptyListETag != hadETag {
				t.Errorf("%s\nempty list ETag not set\nwant: %q\ngot:  %q",
					packageName, hadETag, emptyListETag)
			} else if !tc.want.setEmptyListETag && emptyListETag == hadETag {
				t.Errorf("%s\nempty list ETag should not have been set",
					packageName)
			}
		})
	}
}

func TestReleaseMeetsRequirements(t *testing.T) {
	type wants struct {
		version     string
		releaseDate string
		errRegex    string
	}

	defaultRelease := testBodyObject[0]
	// GIVEN a Lookup with different requirements.
	tests := map[string]struct {
		overrides        string
		releaseOverrides *github_types.Release
		want             wants
	}{
		"no requirements": {
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.PublishedAt},
		},
		"no requirements - use semantic version": {
			releaseOverrides: &github_types.Release{
				TagName:         "v1.0.0",
				SemanticVersion: semver.MustParse("v1.0.0"),
				PublishedAt:     "2021-01-01T00:00:00Z"},
			want: wants{
				version:     "1.0.0",
				releaseDate: "2021-01-01T00:00:00Z",
				errRegex:    `^$`},
		},
		"invalid timestamp": {
			releaseOverrides: &github_types.Release{
				TagName:     "v1.0.0",
				PublishedAt: "invalid"},
			want: wants{
				version:     "v1.0.0",
				releaseDate: "",
				errRegex:    `^$`},
		},
		"require.regex_version - match": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
			`),
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.PublishedAt,
				errRegex:    `^$`},
		},
		"require.regex_version - match but timestamp of asset invalid": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
			`),
			releaseOverrides: &github_types.Release{
				TagName:     "v1.0.0",
				PublishedAt: "invalid"},
			want: wants{
				version:     "v1.0.0",
				releaseDate: "",
				errRegex:    `^$`},
		},
		"require.regex_version - no match": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "x[0-9.]+"
			`),
			want: wants{
				errRegex: `^regex "[^"]+" not matched on version "[^"]+"$`},
		},
		"require.regex_content - match": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "(?i)argus.*amd64"
			`),
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.Assets[0].CreatedAt,
				errRegex:    `^$`},
		},
		"require.regex_content - no match": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "aArgus"
			`),
			want: wants{
				errRegex: `^regex "[^"]+" not matched on content for version "[^"]+"$`},
		},
		"command - pass": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "(?i)argus.*amd64"
					command: ["true"]
			`),
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.Assets[0].CreatedAt,
				errRegex:    `^$`},
		},
		"command - fail": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "(?i)argus.*amd64"
					command: ["false"]
			`),
			want: wants{
				errRegex: `^command failed: .*$`},
		},
		"docker tag - found": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "(?i)argus.*amd64"
					command: ["true"]
					docker:
						type: ghcr
						image: "release-argus/argus"
						tag: "{{ version }}"
						token: ` + os.Getenv("GH_TOKEN") + `
			`),
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.Assets[0].CreatedAt,
				errRegex:    `^$`},
		},
		"docker tag - not found": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "(?i)argus.*amd64"
					command: ["true"]
					docker:
						type: ghcr
						image: "release-argus/argus"
						tag: "x{{ version }}"
						token: ` + os.Getenv("GH_TOKEN") + `
			`),
			want: wants{
				errRegex: `release-argus\/argus:x[0-9.]+ - .*manifest unknown`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			err := yaml.Unmarshal([]byte(tc.overrides), lookup)
			if err != nil {
				t.Fatalf("%s\nfailed to unmarshal overrides: %v",
					packageName, err)
			}
			lookup.Init(
				lookup.Options,
				lookup.Status,
				lookup.Defaults, lookup.HardDefaults)
			testRelease := defaultRelease
			if tc.releaseOverrides != nil {
				testRelease = *tc.releaseOverrides
			}
			logFrom := logutil.LogFrom{Primary: "TestReleaseMeetsRequirements", Secondary: name}

			// WHEN releaseMeetsRequirements is called on it.
			version, releaseDate, err := lookup.releaseMeetsRequirements(testRelease, logFrom)

			// THEN the version is as expected.
			if version != tc.want.version {
				t.Errorf("%s\nersion mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want.version, version)
			}
			// AND the releaseDate is as expected.
			if releaseDate != tc.want.releaseDate {
				t.Errorf("%s\nrelease date mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want.releaseDate, releaseDate)
			}
			// AND any err is as expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want.errRegex, e)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	type want struct {
		version     string
		releaseDate string
		errRegex    string
	}

	// GIVEN a body from the GitHub API and a Lookup.
	body := testBody
	bodyObject := testBodyObject
	tests := map[string]struct {
		body        *string
		overrides   string
		hadReleases []github_types.Release
		want        want
	}{
		"no releases": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: 'ver[0-9.]+'
			`),
			want: want{
				errRegex: `^no releases were found matching the url_commands$`},
		},
		"invalid JSON": {
			body: test.StringPtr("invalid"),
			want: want{
				errRegex: `unmarshal of GitHub API data failed`},
		},
		"cached releases, used when empty body": {
			body: test.StringPtr(""),
			overrides: test.TrimYAML(`
				use_prerelease: false
			`),
			hadReleases: bodyObject,
			want: want{
				version:     bodyObject[1].TagName,
				releaseDate: bodyObject[1].PublishedAt,
				errRegex:    `^$`},
		},
		"cached releases, var changes affect result": {
			body: test.StringPtr(""),
			overrides: test.TrimYAML(`
				use_prerelease: true
			`),
			hadReleases: bodyObject,
			want: want{
				version:     bodyObject[0].TagName,
				releaseDate: bodyObject[0].PublishedAt,
				errRegex:    `^$`},
		},
		"cached releases, ignored if body present": {
			body: test.StringPtr(
				test.TrimJSON(`
				[{"tag_name":"v1.2.3","published_at":"2021-01-01T00:00:00Z"}]
			`)),
			hadReleases: bodyObject,
			want: want{
				version:     "1.2.3",
				releaseDate: "2021-01-01T00:00:00Z",
				errRegex:    `^$`},
		},
		"no releases that meet requirements": {
			overrides: test.TrimYAML(`
				require:
					regex_version: "x[0-9.]+"
			`),
			want: want{
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
					regex "[^"]+" not matched on version "[^"]+"$`)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			err := yaml.Unmarshal([]byte(tc.overrides), lookup)
			if err != nil {
				t.Fatalf("%s\nfailed to unmarshal overrides: %v",
					packageName, err)
			}
			lookup.data.releases = tc.hadReleases
			// Ensure the Status has been handed out.
			lookup.Init(
				lookup.Options,
				lookup.Status,
				lookup.Defaults, lookup.HardDefaults)
			logFrom := logutil.LogFrom{Primary: "TestGetVersion", Secondary: name}
			testBody := body
			if tc.body != nil {
				testBody = []byte(*tc.body)
			}

			// WHEN getVersion is called on it.
			version, releaseDate, err := lookup.getVersion(testBody, logFrom)

			// THEN any err is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want.errRegex, e)
			}
			if tc.want.errRegex != "^$" {
				return
			}
			// AND the version is as expected.
			if version != tc.want.version {
				t.Errorf("%s\nversion mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want.version, version)
			}
			// AND the releaseDate is as expected.
			if releaseDate != tc.want.releaseDate {
				t.Errorf("%s\nrelease date mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want.releaseDate, releaseDate)
			}
		})
	}
}

func TestSetReleases(t *testing.T) {
	// GIVEN a body from the GitHub API and a Lookup.
	body := testBody
	tests := map[string]struct {
		overrides    string
		body         string
		wantReleases bool
		errRegex     string
	}{
		"no pre-releases": {
			overrides: test.TrimYAML(`
				use_prerelease: false
			`),
			wantReleases: true,
			errRegex:     `^$`,
		},
		"want pre-releases": {
			overrides: test.TrimYAML(`
				use_prerelease: true
			`),
			wantReleases: true,
			errRegex:     `^$`,
		},
		"release body that's not valid JSON": {
			body: test.TrimJSON(`
				{"tag_name":"v1.2.3","published_at":"2021-01-01T00:00:00Z"}
			`),
			wantReleases: false,
			errRegex:     `unmarshal of GitHub API data failed`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			err := yaml.Unmarshal([]byte(tc.overrides), lookup)
			if err != nil {
				t.Fatalf("%s\nailed to unmarshal overrides: %v",
					packageName, err)
			}
			logFrom := logutil.LogFrom{Primary: "TestGetVersions", Secondary: name}
			testBody := body
			if tc.body != "" {
				testBody = []byte(tc.body)
			}

			// WHEN setReleases is called on it.
			err = lookup.setReleases(testBody, logFrom)

			// THEN any err is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nrror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
			if tc.errRegex != "^$" {
				return
			}
			// AND the versions are as expected.
			gotReleases := lookup.data.Releases()
			if !tc.wantReleases {
				if len(gotReleases) != 0 {
					t.Errorf("%s\nwant: no releases\ngot:  %v",
						packageName, gotReleases)
				}
				return
			} else if len(gotReleases) == 0 {
				t.Errorf("%s\nwant: releases\ngot:  none",
					packageName)
				return
			}
			if len(gotReleases) != len(testBodyObject) {
				t.Errorf("%s\nwant: %d releases\ngot:  %d",
					packageName, len(testBodyObject), len(gotReleases))
			}
			for i, release := range testBodyObject {
				for j, asset := range release.Assets {
					// Asset counts match.
					if len(gotReleases[i].Assets) != len(release.Assets) {
						t.Errorf("%s\nwant: %d assets for release %d\ngot:  %d",
							packageName, len(release.Assets), i, len(gotReleases[i].Assets))
					}
					// non-matching asset.
					if gotReleases[i].Assets[j].Name != asset.Name {
						t.Errorf("%s\nmismatch at asset [%d][%d]\nwant: %q\ngot:  %v",
							packageName, i, j,
							asset.Name, gotReleases[i].Assets[j])
					}
				}
			}
		})
	}
}

func TestQueryGitHubETag(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		attempts                           int
		eTagChanged, eTagUnchangedUseCache int
		initialRequireRegexVersion         string
		urlCommands                        filter.URLCommandSlice
		errRegex                           string
	}{
		// Keeps .Releases in case filters change.
		"three requests only uses 1 api limit": {
			attempts:              3,
			eTagChanged:           1,
			eTagUnchangedUseCache: 3, // 2 attempts + 1 recheck.
			errRegex:              `^$`},
		"if initial request fails filter, cached results will be used": {
			attempts:                   3,
			eTagChanged:                1,
			eTagUnchangedUseCache:      3, // 2 attempts + 1 recheck.
			initialRequireRegexVersion: `^FOO$`,
			errRegex: test.TrimYAML(`
				^no releases were found matching the require field.*
				regex "[^"]+" not matched on version "[^"]+"$`)},
		"invalid url_commands will catch no versions": {
			attempts:              2,
			eTagChanged:           1,
			eTagUnchangedUseCache: 1,
			urlCommands: filter.URLCommandSlice{
				{Type: "regex", Regex: (`^FOO$`)}},
			errRegex: test.TrimYAML(`
				^no releases were found matching the url_commands
				no releases were found matching the url_commands$`)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			lookup := testLookup(false)
			lookup.GetGitHubData().SetETag("foo")
			lookup.Status.ServiceID = &name
			lookup.Require = &filter.Require{
				RegexVersion: tc.initialRequireRegexVersion,
				Status:       lookup.Status}
			lookup.URLCommands = tc.urlCommands

			attempt := 0
			// WHEN Query is called on it attempts number of times.
			var errs []error
			for tc.attempts != attempt {
				attempt++
				if attempt == 2 {
					lookup.Require = &filter.Require{}
				}

				_, err := lookup.Query(true, logutil.LogFrom{})
				if err != nil {
					errs = append(errs, err)
				}
				t.Logf("%s - attempt %d, ETag: %s",
					packageName, attempt, lookup.GetGitHubData().ETag())
			}

			// THEN any err is expected.
			stdout := releaseStdout()
			e := util.ErrorToString(errors.Join(errs...))
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
			gotETagChanged := strings.Count(stdout, "new ETag")
			if gotETagChanged != tc.eTagChanged {
				t.Errorf("%s\nnew ETag\nwant: %d\ngot:  %d\n\nstdout: %q",
					packageName, tc.eTagChanged, gotETagChanged, stdout)
			}
			gotETagUnchangedUseCache := strings.Count(stdout, "Using cached releases")
			if gotETagUnchangedUseCache != tc.eTagUnchangedUseCache {
				t.Errorf("%s\nETag unchanged use cache\nwant: %d\ngot:  %d\n\nstdout: %q",
					packageName, tc.eTagUnchangedUseCache, gotETagUnchangedUseCache, stdout)
			}
		})
	}
}

func TestHandleNoVersionChange(t *testing.T) {
	// GIVEN a Lookup that got an unchanged version on check X.
	type args struct {
	}
	tests := map[string]struct {
		version     string
		checkNumber int
		doesPrint   bool
	}{
		"first check": {
			version:     "a.b.c",
			checkNumber: 0,
			doesPrint:   false,
		},
		"second check": {
			version:     "x.y.z",
			checkNumber: 1,
			doesPrint:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			lookup := testLookup(false)

			// WHEN handleNoVersionChange is called on it.
			lookup.handleNoVersionChange(tc.checkNumber, tc.version, logutil.LogFrom{})

			// THEN a message is printed when expected.
			stdout := releaseStdout()
			gotMessage := util.RegexCheck(
				fmt.Sprintf(`Staying on "%s" as that's the latest version in the second check`, tc.version),
				stdout)
			if gotMessage != tc.doesPrint {
				if gotMessage {
					t.Errorf("%s\nprinted message when not expected %s",
						packageName, stdout)
				} else {
					t.Errorf("%s\ndid not print message when expected %s",
						packageName, stdout)
				}
			}
		})
	}
}
