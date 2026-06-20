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

//go:build unit

// Package github provides a github-based lookup type.
package github

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/util/errfmt"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	ghtypes "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
)

func TestLookup_HTTPRequest(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name     string
		failing  bool
		url      string
		eTag     *string
		nextPage int
		errRegex string
	}{
		{
			name:     "invalid url",
			url:      "invalid://	test",
			errRegex: `invalid control character in URL`,
		},
		{
			name:     "unknown url",
			url:      "https://release-argus.invalid-tld",
			errRegex: `no such host`,
		},
		{
			name:     "valid url",
			url:      test.ArgusGitHubRepo,
			errRegex: `^$`,
			nextPage: 2,
		},
		{
			name:     "repo that uses tags, not releases/has tags",
			url:      "release-argus/test",
			errRegex: `^$`,
		},
		{
			name:     "repo that uses tags, not releases/no tags",
			url:      "release-argus/.github",
			errRegex: `^$`,
		},
		{
			name:     "repo that uses tags, not releases/update EmptyListETag if 200 on empty list",
			url:      "release-argus/.github",
			eTag:     test.Ptr(""),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.URL = tc.url
			if tc.eTag != nil {
				lookup.data.eTag = *tc.eTag
			}

			// WHEN: httpRequest is called on it.
			_, nextPage, err := lookup.httpRequest(1, logx.LogFrom{})

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nLookup.httpRequest(%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.url,
					e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the nextPage is as expected.
			if nextPage != tc.nextPage {
				t.Errorf(
					"%s\nLookup.httpRequest(%q) nextPage value mismatch\ngot:  %d\nwant: %d",
					packageName, tc.url,
					nextPage, tc.nextPage,
				)
			}
		})
	}
}

func TestGetResponse_ReadError(t *testing.T) {
	// GIVEN: a server that closes the connection immediately to simulate a read error.
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "10") // Set Content-Length without sending data.
			w.WriteHeader(http.StatusOK)
			// Immediately close the connection without writing any body.
			conn, _, _ := w.(http.Hijacker).Hijack()
			conn.Close()
		}),
	)
	t.Cleanup(server.Close)

	// AND: a request to the mock server's URL.
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf(
			"%s\ncould not create request: %v",
			packageName, err,
		)
	}

	// WHEN: getResponse is called on that URL.
	l := Lookup{}
	_, _, err = l.getResponse(req, logx.LogFrom{})

	// THEN: an error is expected from the read error.
	if err == nil {
		t.Fatalf("%s\nexpected an error when reading response body, got none", packageName)
	}
}

func TestLookup_HandleResponse(t *testing.T) {
	githubClientConnErr := `\/tags": http2: client conn could not be established`
	type wants struct {
		nilBody          bool
		nextPage         int
		setEmptyListETag bool
		errRegex         string
	}
	type conditions struct {
		hadReleases           bool
		hadDefaultAccessToken bool
	}

	// GIVEN: a HTTP Response and an accompanying body.
	tests := []struct {
		name       string
		conditions conditions
		statusCode int
		body       []byte
		want       wants
	}{
		{
			name: "200 OK/EmptyListETag set if default access_token",
			conditions: conditions{
				hadDefaultAccessToken: true,
			},
			statusCode: http.StatusOK,
			body:       []byte(`[]`),
			want: wants{
				nextPage:         2,
				setEmptyListETag: true,
				errRegex:         `^$`,
			},
		},
		{
			name: "200 OK/EmptyListETag not set if non-default access_token",
			conditions: conditions{
				hadDefaultAccessToken: false,
			},
			statusCode: http.StatusOK,
			body:       []byte(`[]`),
			want: wants{
				nextPage:         2,
				setEmptyListETag: false,
				errRegex:         `^$`,
			},
		},
		{
			name:       "200 OK/Get releases",
			statusCode: http.StatusOK,
			body:       []byte(`[{"tag_name":"v1.0.0"}]`),
			want: wants{
				nextPage: 0,
				errRegex: `^$`,
			},
		},
		{
			name:       "304 Not Modified/Tag fallback request",
			statusCode: http.StatusNotModified,
			want: wants{
				nilBody:  false,
				nextPage: 2,
				errRegex: `^$`,
			},
		},
		{
			name: "304 Not Modified/Use old releases",
			conditions: conditions{
				hadReleases: true,
			},
			statusCode: http.StatusNotModified,
			want: wants{
				nilBody:  true,
				nextPage: 2,
				errRegex: `^$`,
			},
		},
		{
			name:       "401 Unauthorized/bad credentials",
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"message":"Bad credentials"}`),
			want: wants{
				errRegex: `github access token is invalid`,
				nilBody:  true,
			},
		},
		{
			name:       "401 Unauthorized/unknown",
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"message":"Something else"}`),
			want: wants{
				errRegex: `unknown 401 response`,
				nilBody:  true,
			},
		},
		{
			name:       "403 Forbidden/rate limit",
			statusCode: http.StatusForbidden,
			body:       []byte(`{"message":"API rate limit exceeded"}`),
			want: wants{
				errRegex: `rate limit reached for GitHub`,
				nilBody:  true,
			},
		},
		{
			name:       "403 Forbidden/missing tag_name",
			statusCode: http.StatusForbidden,
			body:       []byte(`{"message":"some other error"}`),
			want: wants{
				errRegex: `tag_name not found at`,
				nilBody:  true,
			},
		},
		{
			name:       "403 Forbidden/unknown",
			statusCode: http.StatusForbidden,
			body:       []byte(`[{"tag_name":"v1.0.0"}]`),
			want: wants{
				errRegex: `unknown 403 response`,
				nilBody:  true,
			},
		},
		{
			name:       "429 Too Many Requests/rate limit",
			statusCode: http.StatusTooManyRequests,
			body:       []byte(`{"message":"something from GitHub"}`),
			want: wants{
				errRegex: `^too many requests made to GitHub - "something from GitHub"$`,
				nilBody:  true,
			},
		},
		{
			name:       "429 Too Many Requests/unmarshal fail",
			statusCode: http.StatusTooManyRequests,
			body:       []byte(`{"message":}`),
			want: wants{
				errRegex: `^too many requests made to GitHub$`,
				nilBody:  true,
			},
		},
		{
			name:       "Unknown status code",
			statusCode: http.StatusTeapot,
			body:       []byte(`{"message":"I'm a teapot"}`),
			want: wants{
				errRegex: `unknown status code 418`,
				nilBody:  true,
			},
		},
	}

	// Ensure other tests that modify global state don't interfere.
	releaseStdout := test.CaptureLog(t, logx.Default())
	defer releaseStdout()
	hadEmptyListETag := getEmptyListETag()
	t.Cleanup(func() { setEmptyListETag(hadEmptyListETag) })

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're modifying global state.

			prefix := fmt.Sprintf("%s\nLookup.handleResponse()", packageName)

			// Retry up-to 3 times if we get a client conn error on GitHub requests.
			for try := range 3 {
				t.Logf(
					"%s - attempt %d\n",
					prefix, try+1,
				)

				lookup := testLookup(t, false)
				resp := &http.Response{
					StatusCode: tc.statusCode,
					Header:     http.Header{},
					Request: &http.Request{
						URL: &url.URL{},
					},
				}
				hadETag := tc.name
				resp.Header.Add("ETag", hadETag)
				if tc.conditions.hadReleases {
					lookup.data.releases = testBodyObject
				}
				lookup.AccessToken = lookup.accessToken()
				if !tc.conditions.hadDefaultAccessToken {
					lookup.Defaults.AccessToken = ""
					lookup.HardDefaults.AccessToken = "Something"
				}

				logFrom := logx.LogFrom{Primary: "TestHandleResponse", Secondary: tc.name}

				// WHEN: handleResponse is called on it.
				gotBody, nextPage, err := lookup.handleResponse(resp, tc.body, logFrom)

				// GitHub actions regularly fail /tags with:
				//   'Get "https://api.github.com/repos/.../tags": http2: client conn could not be established'
				e := errfmt.FormatError(err)
				if util.RegexCheck(githubClientConnErr, e) {
					t.Logf(
						"%s retrying... %q\n",
						prefix, e,
					)
					time.Sleep(time.Duration(rand.Intn(25)) * time.Millisecond)
					continue
				}

				// THEN: any error is as expected.
				if !util.RegexCheck(tc.want.errRegex, e) {
					t.Errorf(
						"%s error mismatch\ngot:  %q\nwant: %q",
						prefix, tc.want.errRegex, e,
					)
				}

				// AND: the body returned is as expected.
				if tc.want.nilBody && len(gotBody) != 0 {
					t.Errorf(
						"%s body mismatch\ngot:  %q\nwant: nil",
						prefix, string(tc.body),
					)
				} else if !tc.want.nilBody && len(gotBody) == 0 {
					t.Errorf("%s body mismatch\ngot:  nil\nwant: non-nil", prefix)
				}

				// AND: the new EmptyListETag is as expected.
				emptyListETag := getEmptyListETag()
				if tc.want.setEmptyListETag && emptyListETag != hadETag {
					t.Errorf(
						"%s didn't set empty list ETag\ngot:  %q\nwant: %q",
						prefix, hadETag, emptyListETag,
					)
				} else if !tc.want.setEmptyListETag && emptyListETag == hadETag {
					t.Errorf("%s empty list ETag should not have been set", prefix)
				}

				// AND: the nextPage is as expected.
				if nextPage != tc.want.nextPage {
					t.Errorf(
						"%s nextPage mismatch\ngot:  %d\nwant: %d",
						prefix, nextPage, tc.want.nextPage,
					)
				}
				break
			}
		})
	}
}

func TestLookup_ReleaseMeetsRequirements(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	type wants struct {
		version     string
		releaseDate string
		errRegex    string
	}

	defaultRelease := testBodyObject[0]
	// GIVEN: a Lookup with different requirements.
	tests := []struct {
		name             string
		overrides        string
		releaseOverrides *ghtypes.Release
		want             wants
	}{
		{
			name: "no requirements",
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.PublishedAt,
			},
		},
		{
			name: "no requirements - use semantic version",
			releaseOverrides: &ghtypes.Release{
				TagName:         "v1.0.0",
				SemanticVersion: semver.MustParse("v1.0.0"),
				PublishedAt:     "2021-01-01T00:00:00Z",
			},
			want: wants{
				version:     "1.0.0",
				releaseDate: "2021-01-01T00:00:00Z",
				errRegex:    `^$`,
			},
		},
		{
			name: "invalid timestamp",
			releaseOverrides: &ghtypes.Release{
				TagName:     "v1.0.0",
				PublishedAt: "invalid",
			},
			want: wants{
				version:     "v1.0.0",
				releaseDate: "",
				errRegex:    `^$`,
			},
		},
		{
			name: "require.regex_version/match",
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
			`),
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.PublishedAt,
				errRegex:    `^$`,
			},
		},
		{
			name: "require.regex_version/match, but timestamp of asset invalid",
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
			`),
			releaseOverrides: &ghtypes.Release{
				TagName:     "v1.0.0",
				PublishedAt: "invalid",
			},
			want: wants{
				version:     "v1.0.0",
				releaseDate: "",
				errRegex:    `^$`,
			},
		},
		{
			name: "require.regex_version/no match",
			overrides: test.TrimYAML(`
				require:
					regex_version: "x[0-9.]+"
			`),
			want: wants{
				errRegex: `^regex "[^"]+" not matched on version "[^"]+"$`,
			},
		},
		{
			name: "require.regex_content/match",
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "(?i)argus.*amd64"
			`),
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.Assets[0].CreatedAt,
				errRegex:    `^$`,
			},
		},
		{
			name: "require.regex_content/no match",
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "aArgus"
			`),
			want: wants{
				errRegex: `^regex "[^"]+" not matched on content for version "[^"]+"$`,
			},
		},
		{
			name: "command/pass",
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "(?i)argus.*amd64"
					command: ["true"]
			`),
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.Assets[0].CreatedAt,
				errRegex:    `^$`,
			},
		},
		{
			name: "command/fail",
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "(?i)argus.*amd64"
					command: ["false"]
			`),
			want: wants{
				errRegex: test.TrimYAML(`
					^command failed:
						exit status 1$`,
				),
			},
		},
		{
			name: "docker tag/found",
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "(?i)argus.*amd64"
					command: ["true"]
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: "{{ version }}"
						token: ` + test.DockerHubToken(t) + `
			`),
			want: wants{
				version:     defaultRelease.TagName,
				releaseDate: defaultRelease.Assets[0].CreatedAt,
				errRegex:    `^$`,
			},
		},
		{
			name: "docker tag/not found",
			overrides: test.TrimYAML(`
				require:
					regex_version: "[0-9.]+"
					regex_content: "(?i)argus.*amd64"
					command: ["true"]
					docker:
						type: ghcr
						image: "` + test.ArgusDockerGHCRRepo + `"
						tag: "x{{ version }}"
						token: ` + test.DockerHubToken(t) + `
			`),
			want: wants{
				errRegex: `release-argus\/argus:x[0-9.]+ - .*tag not found`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			// overrides.
			if err := lookup.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Lookup overrides: %v",
					packageName, err,
				)
			}
			requireData, _ := polymorphic.Extract("yaml", []byte(tc.overrides), "require")
			req, err := filter.Decode(
				"yaml", requireData,
				lookup.Status,
				&lvCfg.Soft.Require,
			)
			if err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Require overrides: %v",
					packageName, err,
				)
			}
			lookup.SetRequire(req)
			lookup.Init(
				lookup.Options,
				lookup.Status,
				base.DefaultsConfig{
					Soft: lookup.Defaults,
					Hard: lookup.HardDefaults,
				},
			)
			testRelease := defaultRelease
			if tc.releaseOverrides != nil {
				testRelease = *tc.releaseOverrides
			}
			logFrom := logx.LogFrom{Primary: "TestReleaseMeetsRequirements", Secondary: tc.name}

			// WHEN: releaseMeetsRequirements is called on it.
			version, releaseDate, err := lookup.releaseMeetsRequirements(testRelease, logFrom)

			prefix := fmt.Sprintf(
				"%s\nLookup.releaseMeetsRequirements(%+v)",
				packageName, testRelease,
			)

			// THEN: any decode is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, tc.want.errRegex, e,
				)
			}
			if e != "" {
				return
			}

			// AND: the version is as expected.
			if version != tc.want.version {
				t.Errorf(
					"%s version mismatch\ngot:  %q\nwant: %q",
					prefix, version, tc.want.version,
				)
			}

			// AND: the releaseDate is as expected.
			if releaseDate != tc.want.releaseDate {
				t.Errorf(
					"%s release date mismatch\ngot:  %q\nwant: %q",
					prefix, releaseDate, tc.want.releaseDate,
				)
			}
		})
	}
}

func TestLookup_GetVersion(t *testing.T) {
	type want struct {
		version     string
		releaseDate string
		errRegex    string
	}

	// GIVEN: a body from the GitHub API and a Lookup.
	body := testBody
	bodyObject := testBodyObject
	tests := []struct {
		name            string
		bodyOverride    *string
		lookupOverrides string
		hadReleases     []ghtypes.Release
		want            want
	}{
		{
			name: "no releases",
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: 'ver[0-9.]+'
			`),
			want: want{
				errRegex: `^no releases were found matching the url_commands on page \d+ of the API response$`,
			},
		},
		{
			name:         "invalid JSON",
			bodyOverride: test.Ptr("invalid"),
			want: want{
				errRegex: `unmarshal of GitHub API data failed`,
			},
		},
		{
			name:            "cached releases, used when empty body",
			bodyOverride:    test.Ptr(""),
			lookupOverrides: `use_prerelease: false`,
			hadReleases:     bodyObject,
			want: want{
				version:     bodyObject[1].TagName,
				releaseDate: bodyObject[1].PublishedAt,
				errRegex:    `^$`,
			},
		},
		{
			name:            "cached releases, var changes affect result",
			bodyOverride:    test.Ptr(""),
			lookupOverrides: `use_prerelease: true`,
			hadReleases:     bodyObject,
			want: want{
				version:     bodyObject[0].TagName,
				releaseDate: bodyObject[0].PublishedAt,
				errRegex:    `^$`,
			},
		},
		{
			name:         "cached releases, ignored if body present",
			bodyOverride: test.Ptr(`[{"tag_name":"v1.2.3","published_at":"2021-01-01T00:00:00Z"}]`),
			hadReleases:  bodyObject,
			want: want{
				version:     "1.2.3",
				releaseDate: "2021-01-01T00:00:00Z",
				errRegex:    `^$`,
			},
		},
		{
			name: "no releases that meet requirements",
			lookupOverrides: test.TrimYAML(`
				require:
					regex_version: "x[0-9.]+"
			`),
			want: want{
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
						regex "[^"]+" not matched on version "[^"]+"$`,
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			if err := lookup.ApplyOverrides("yaml", []byte(tc.lookupOverrides)); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Lookup overrides: %v",
					packageName, err,
				)
			}
			lookup.data.releases = tc.hadReleases
			// Ensure the Status has been handed out.
			lookup.Init(
				lookup.Options,
				lookup.Status,
				base.DefaultsConfig{
					Soft: lookup.Defaults,
					Hard: lookup.HardDefaults,
				},
			)
			logFrom := logx.LogFrom{Primary: "TestGetVersion", Secondary: tc.name}
			testBody := body
			if tc.bodyOverride != nil {
				testBody = []byte(*tc.bodyOverride)
			}

			// WHEN: getVersion is called on it.
			version, releaseDate, err := lookup.getVersion(testBody, 1, logFrom)

			prefix := fmt.Sprintf(
				"%s\nLookup.getVersion(%q)",
				packageName, testBody,
			)

			// THEN: any decode is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.want.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the version is as expected.
			if version != tc.want.version {
				t.Errorf(
					"%s unexpected version returned\ngot:  %q\nwant: %q",
					prefix, version, tc.want.version,
				)
			}

			// AND: the releaseDate is as expected.
			if releaseDate != tc.want.releaseDate {
				t.Errorf(
					"%s unexpected release date returned\ngot:  %q\nwant: %q",
					prefix, releaseDate, tc.want.releaseDate,
				)
			}
		})
	}
}

func TestLookup_SetReleases(t *testing.T) {
	// GIVEN: a body from the GitHub API and a Lookup.
	body := testBody
	tests := []struct {
		name         string
		overrides    string
		body         string
		wantReleases bool
		errRegex     string
	}{
		{
			name:         "no pre-releases",
			overrides:    `use_prerelease: false`,
			wantReleases: true,
			errRegex:     `^$`,
		},
		{
			name:         "want pre-releases",
			overrides:    `use_prerelease: true`,
			wantReleases: true,
			errRegex:     `^$`,
		},
		{
			name:         "release body that's not valid JSON",
			body:         `{"tag_name":"v1.2.3","published_at":"2021-01-01T00:00:00Z"}`,
			wantReleases: false,
			errRegex:     `unmarshal of GitHub API data failed`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			if err := lookup.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Lookup overrides: %v",
					packageName, err,
				)
			}
			testBody := body
			if tc.body != "" {
				testBody = []byte(tc.body)
			}

			// WHEN: setReleases is called on it.
			err := lookup.setReleases(testBody)

			prefix := fmt.Sprintf(
				"%s\nLookup setReleases(%q)",
				packageName, testBody,
			)

			// THEN: any decode is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if tc.errRegex != "^$" {
				return
			}

			// AND: the number of Releases is as expected.
			gotReleases := lookup.data.Releases()
			if len(gotReleases) == 0 {
				t.Errorf("%s Data.Releases() result mismatch\ngot:  no releases\nwant: releases", prefix)
				return
			}
			if gotLen, wantLen := len(gotReleases), len(testBodyObject); gotLen != wantLen {
				t.Errorf(
					"%s Release count mismatch\ngot:  %d\nwant: %d",
					prefix,
					gotLen, wantLen,
				)
			}

			// AND: the assets attached to each Release is as expected.
			if err := test.AssertSlicesEqualFunc(
				t,
				gotReleases,
				testBodyObject,
				func(gotRelease ghtypes.Release, wantRelease ghtypes.Release) bool {
					// Asset counts match.
					if len(gotRelease.Assets) != len(wantRelease.Assets) {
						return false
					}
					// Asset names match
					for i := range gotRelease.Assets {
						if gotRelease.Assets[i].Name != wantRelease.Assets[i].Name {
							return false
						}
					}
					return true
				},
				prefix,
				"Release",
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLookup_HandleNoVersionChange(t *testing.T) {
	// GIVEN: a Lookup that got an unchanged version on check X.
	tests := []struct {
		name      string
		version   string
		doesPrint bool
	}{
		{
			name:      "first check",
			version:   "a.b.c",
			doesPrint: false,
		},
		{
			name:      "second check",
			version:   "x.y.z",
			doesPrint: true,
		},
	}

	for checkNumber, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			lookup := testLookup(t, false)

			// WHEN: handleNoVersionChange is called on it.
			lookup.handleNoVersionChange(checkNumber, tc.version, logx.LogFrom{})

			prefix := fmt.Sprintf(
				"%s\nLookup.handleNoVersionChange(checkNumber=%d, version=%s)",
				packageName, checkNumber, tc.version,
			)

			// THEN: a message is printed when expected.
			stdout := releaseStdout()
			wantRe := fmt.Sprintf(`Staying on %q as that's the latest version in the second check`, tc.version)
			gotMessage := util.RegexCheck(
				wantRe,
				stdout,
			)
			if gotMessage != tc.doesPrint {
				format := "%s printed message when not expected\ngot:  %q\nwant: %q"
				if gotMessage {
					format = "%s printed message when not expected\ngot:  %q\nwant: NOT %q"
				}
				t.Errorf(
					format,
					prefix, stdout, wantRe,
				)
			}
		})
	}
}
