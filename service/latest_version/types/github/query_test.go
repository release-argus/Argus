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
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/util/polymorphic"
)

var nonSemanticBody = test.TrimJSON(`[
	{
		"tag_name":"ver1.1.1",
		"name":"ver1.1.1",
		"prerelease":false,
		"published_at":"2024-04-27T10:50:00Z",
		"assets":[
			{"id": 3,"name":"Argus-ver1.1.1.linux-amd64","created_at":"2024-04-27T10:50:53Z","browser_download_url":"https://example.com/Argus-ver1.1.1.linux-amd64"}
		]
	}
]`)

func TestLookup_Query__hermetic(t *testing.T) {
	// GIVEN: a Lookup, and the releases its API serves.
	tests := []struct {
		name        string
		endpoints   map[string]githubEndpoint
		requireAuth string // Authorization the API required.
		overrides   string
		semVer      bool
		wantVersion string
		errRegex    string
	}{
		{
			name:      "invalid/a URL that cannot be parsed",
			overrides: "url: 'release-argus\tArgus'",
			errRegex:  `invalid control character in URL`,
		},
		{
			name: "valid/the newest stable release is reported",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: string(testBody)},
			},
			semVer:      true,
			wantVersion: "0.17.4",
			errRegex:    `^$`,
		},
		{
			name: "invalid/a non-semantic version is rejected when semver is required",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: nonSemanticBody},
			},
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: 'ver[0-9.]+'
			`),
			semVer:   true,
			errRegex: `no releases were found matching the url_commands`,
		},
		{
			name: "valid/a non-semantic version is kept when semver is not required",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: nonSemanticBody},
			},
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: 'ver[0-9.]+'
			`),
			semVer:      false,
			wantVersion: "ver1.1.1",
			errRegex:    `^$`,
		},
		{
			name: "invalid/require.regex_content matches no asset",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: string(testBody)},
			},
			overrides: test.TrimYAML(`
				require:
					regex_content: "argus[0-9]+.exe"
			`),
			semVer: true,
			errRegex: test.TrimYAML(`
				^no releases were found matching the require field.*
					regex "[^"]+" not matched on content for version "[^"]+"$`,
			),
		},
		{
			name: "valid/require.regex_content matches an asset of the version",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: string(testBody)},
			},
			overrides: test.TrimYAML(`
				require:
					regex_content: "{{ version }}.linux-amd64"
			`),
			semVer:      true,
			wantVersion: "0.17.4",
			errRegex:    `^$`,
		},
		{
			name: "valid/the access_token is sent as a Bearer token",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: string(testBody)},
			},
			requireAuth: "Bearer foo",
			overrides:   "access_token: foo",
			semVer:      true,
			wantVersion: "0.17.4",
			errRegex:    `^$`,
		},
		{
			name: "invalid/a credential the API will not accept",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: string(testBody)},
			},
			requireAuth: "Bearer wanted",
			overrides:   "access_token: foo",
			semVer:      true,
			errRegex:    `github access token is invalid`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.Status.ServiceInfo.ID = tc.name
			if err := lookup.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Lookup overrides: %v",
					packageName, err,
				)
			}
			lookup.Options.SemanticVersioning = &tc.semVer

			server := startGitHubServer(t, &githubServer{
				Endpoints:   tc.endpoints,
				RequireAuth: tc.requireAuth,
			})
			lookup.apiRoot = server.URL
			lookup.Init(
				lookup.Options,
				lookup.Status,
				base.DefaultsConfig{
					Soft: lookup.Defaults,
					Hard: lookup.HardDefaults,
				},
			)

			// WHEN: Query is called on it.
			_, err := lookup.Query(true, logx.LogFrom{Primary: t.Name()})

			prefix := fmt.Sprintf("%s\nLookup.Query()", packageName)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the version reported is as expected.
			if got := lookup.Status.LatestVersion(); got != tc.wantVersion {
				t.Errorf(
					"%s version mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantVersion,
				)
			}
		})
	}
}

func TestLookup_Query__eTagReusesTheCache(t *testing.T) {
	// GIVEN: no empty-list ETag has been learnt yet, so the first request below is
	// unconditional.
	hadETag := getEmptyListETag()
	t.Cleanup(func() { setEmptyListETag(hadETag) })
	setEmptyListETag("")

	// AND: an API that answers matching ETag requests with 304.
	server := newGitHubServer(t, map[string]githubEndpoint{
		"releases": {Body: string(testBody), ETag: `W/"unchanged"`},
	})

	lookup := testLookup(t, false)
	lookup.apiRoot = server.URL
	lookup.Status.ServiceInfo.ID = t.Name()

	prefix := fmt.Sprintf("%s\nLookup.Query()", packageName)

	// WHEN: it is queried three times.
	for attempt := 1; attempt <= 3; attempt++ {
		if _, err := lookup.Query(true, logx.LogFrom{Primary: t.Name()}); err != nil {
			t.Fatalf(
				"%s attempt %d gave an unexpected error: %v",
				prefix, attempt, errfmt.FormatError(err),
			)
		}

		// THEN: the version is reported every time, cached or not.
		if got, want := lookup.Status.LatestVersion(), "0.17.4"; got != want {
			t.Fatalf(
				"%s attempt %d version mismatch\ngot:  %q\nwant: %q",
				prefix, attempt,
				got, want,
			)
		}
	}

	// AND: only the first request was unconditional, so the body was served once. Every
	// later request carried the ETag the API gave, and was answered 304.
	requests := server.Requests()
	if len(requests) < 3 {
		t.Fatalf(
			"%s made %d request(s)\n%v\nwant: at least one per query (3)",
			prefix, len(requests), requests,
		)
	}
	if got, want := server.NotModifiedCount(), len(requests)-1; got != want {
		t.Errorf(
			"%s the body should be served once, every later request answered 304\ngot:  %d 304(s) of %d request(s)\nwant: %d",
			prefix, got, len(requests), want,
		)
	}
	if got, want := requests[0].Header.Get("If-None-Match"), ""; got != want {
		t.Errorf(
			"%s the first request should carry no ETag to ask against\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}
	for _, request := range requests[1:] {
		if got, want := request.Header.Get("If-None-Match"), `"unchanged"`; got != want {
			t.Errorf(
				"%s a later request was not conditional\ngot:  %q\nwant: %q",
				prefix, got, want,
			)
		}
	}
}

func TestLookup_HTTPRequest(t *testing.T) {
	// GIVEN: a Lookup, and the API it queries.
	//
	// An empty list on page one teaches the package-wide empty-list ETag, so these
	// cases share state and cannot run in parallel.
	hadETag := getEmptyListETag()
	t.Cleanup(func() { setEmptyListETag(hadETag) })

	tests := []struct {
		name         string
		endpoints    map[string]githubEndpoint
		closedAPI    bool // Address an API that is not listening.
		url          string
		etag         *string
		regexContent string // Blocks the tag fallback.
		nextPage     int
		wantPaths    []string
		wantETagSet  string
		errRegex     string
	}{
		{
			name:     "invalid/a URL that cannot be parsed",
			url:      "invalid://\ttest",
			errRegex: `invalid control character in URL`,
		},
		{
			name:      "invalid/an API that is not listening",
			closedAPI: true,
			url:       "release-argus/Argus",
			errRegex:  `connection refused|connect: `,
		},
		{
			name:      "invalid/a repository the API cannot see",
			endpoints: map[string]githubEndpoint{},
			url:       "release-argus/no-such-repo",
			errRegex:  `Not Found`,
		},
		{
			name: "valid/releases, with another page to come",
			endpoints: map[string]githubEndpoint{
				"releases": {Pages: []string{string(testBody), string(testBody)}},
			},
			url:       "release-argus/Argus",
			nextPage:  2,
			wantPaths: []string{"/repos/release-argus/Argus/releases"},
			errRegex:  `^$`,
		},
		{
			name: "valid/no releases falls back to tags",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: "[]"},
				"tags":     {Body: string(testBody)},
			},
			url: "release-argus/test",
			wantPaths: []string{
				"/repos/release-argus/test/releases",
				"/repos/release-argus/test/tags",
			},
			errRegex: `^$`,
		},
		{
			name: "valid/neither releases nor tags published",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: "[]"},
				"tags":     {Body: "[]"},
			},
			url: "release-argus/.github",
			wantPaths: []string{
				"/repos/release-argus/.github/releases",
				"/repos/release-argus/.github/tags",
			},
			errRegex: `^$`,
		},
		{
			name: "valid/an empty list on page one teaches its ETag",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: "[]", ETag: `W/"learnt-etag"`},
			},
			url:          "release-argus/.github",
			etag:         new(""),
			regexContent: "argus",
			wantETagSet:  `"learnt-etag"`,
			errRegex:     `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - the empty-list ETag is package-wide state.

			lookup := testLookup(t, false)
			lookup.URL = tc.url
			if tc.regexContent != "" {
				lookup.Require = &filter.Require{RegexContent: tc.regexContent}
			}
			if tc.etag != nil {
				lookup.data.etag = *tc.etag
			}

			server := newGitHubServer(t, tc.endpoints)
			lookup.apiRoot = server.URL
			if tc.closedAPI {
				server.Close()
			}

			// WHEN: httpRequest is called on it.
			_, nextPage, err := lookup.httpRequest(1, logx.LogFrom{Primary: t.Name()})

			prefix := fmt.Sprintf(
				"%s\nLookup.httpRequest(%q)",
				packageName, tc.url,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the nextPage is as expected.
			if nextPage != tc.nextPage {
				t.Errorf(
					"%s nextPage mismatch\ngot:  %d\nwant: %d",
					prefix, nextPage, tc.nextPage,
				)
			}

			// AND: the paths requested are as expected.
			if tc.wantPaths != nil {
				var paths []string
				for _, request := range server.Requests() {
					if len(paths) == 0 || paths[len(paths)-1] != request.Path {
						paths = append(paths, request.Path)
					}
				}
				if !slices.Equal(paths, tc.wantPaths) {
					t.Errorf(
						"%s requested paths mismatch\ngot:  %v\nwant: %v",
						prefix, paths, tc.wantPaths,
					)
				}
			}

			// AND: an empty list's ETag is remembered for next time.
			if tc.wantETagSet != "" {
				if got := getEmptyListETag(); got != tc.wantETagSet {
					t.Errorf(
						"%s empty-list ETag mismatch\ngot:  %q\nwant: %q",
						prefix, got, tc.wantETagSet,
					)
				}
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
	type wants struct {
		nilBody          bool
		nextPage         int
		setEmptyListETag bool
		tagFallback      bool
		errRegex         string
	}
	type conditions struct {
		hadReleases bool
		accessToken string // "unset" | "default", or anything else for a non-default token.
	}

	// GIVEN: a HTTP Response and an accompanying body.
	tests := []struct {
		name        string
		conditions  conditions
		lookupSetup func(*Lookup)
		statusCode  int
		body        []byte
		want        wants
	}{
		{
			name: "200 OK/EmptyListETag set if no access_token",
			conditions: conditions{
				accessToken: "unset",
			},
			statusCode: http.StatusOK,
			body:       []byte(`[]`),
			want: wants{
				nextPage:         2,
				setEmptyListETag: true,
				tagFallback:      true,
				errRegex:         `^$`,
			},
		},
		{
			name: "200 OK/EmptyListETag set if the access_token spells the default",
			conditions: conditions{
				accessToken: "default",
			},
			statusCode: http.StatusOK,
			body:       []byte(`[]`),
			want: wants{
				nextPage:         2,
				setEmptyListETag: true,
				tagFallback:      true,
				errRegex:         `^$`,
			},
		},
		{
			name: "200 OK/EmptyListETag not set if non-default access_token",
			conditions: conditions{
				accessToken: "",
			},
			statusCode: http.StatusOK,
			body:       []byte(`[]`),
			want: wants{
				nextPage:         2,
				setEmptyListETag: false,
				tagFallback:      true,
				errRegex:         `^$`,
			},
		},
		{
			name:       "200 OK/Get releases",
			statusCode: http.StatusOK,
			body:       []byte(`[{"tag_name":"v1.0.0"}]`),
			want: wants{
				nextPage:    0,
				tagFallback: false,
				errRegex:    `^$`,
			},
		},
		{
			name:       "200 OK/use_prerelease doesn't block tag fallback on empty list",
			statusCode: http.StatusOK,
			body:       []byte(`[]`),
			lookupSetup: func(l *Lookup) {
				l.UsePreRelease = new(true)
			},
			want: wants{
				nilBody:     false,
				nextPage:    2,
				tagFallback: true,
				errRegex:    `^$`,
			},
		},
		{
			name:       "200 OK/require.regex_content blocks tag fallback on empty list",
			statusCode: http.StatusOK,
			body:       []byte(`[]`),
			lookupSetup: func(l *Lookup) {
				l.Require = &filter.Require{RegexContent: "some-pattern"}
			},
			want: wants{
				nilBody:     false,
				nextPage:    0,
				tagFallback: false,
				errRegex:    `^$`,
			},
		},
		{
			name:       "304 Not Modified/use_prerelease doesn't block tag fallback",
			statusCode: http.StatusNotModified,
			lookupSetup: func(l *Lookup) {
				l.UsePreRelease = new(true)
			},
			want: wants{
				nilBody:     false,
				nextPage:    2,
				tagFallback: true,
				errRegex:    `^$`,
			},
		},
		{
			name:       "304 Not Modified/require.regex_content blocks tag fallback",
			statusCode: http.StatusNotModified,
			lookupSetup: func(l *Lookup) {
				l.Require = &filter.Require{RegexContent: "some-pattern"}
			},
			want: wants{
				nilBody:     true,
				nextPage:    0,
				tagFallback: false,
				errRegex:    `^$`,
			},
		},
		{
			name:       "304 Not Modified/Tag fallback request",
			statusCode: http.StatusNotModified,
			want: wants{
				nilBody:     false,
				nextPage:    2,
				tagFallback: true,
				errRegex:    `^$`,
			},
		},
		{
			name: "304 Not Modified/Use old releases",
			conditions: conditions{
				hadReleases: true,
			},
			statusCode: http.StatusNotModified,
			want: wants{
				nilBody:     true,
				nextPage:    2,
				tagFallback: false,
				errRegex:    `^$`,
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
				nilBody:  true,
				errRegex: `^too many requests made to GitHub$`,
			},
		},
		{
			name:       "Unknown status code",
			statusCode: http.StatusTeapot,
			body:       []byte(`{"message":"I'm a teapot"}`),
			want: wants{
				nilBody:  true,
				errRegex: `unknown status code 418`,
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

			server := newGitHubServer(t, map[string]githubEndpoint{
				"tags": {Pages: []string{string(testBody), string(testBody)}},
			})

			lookup := testLookup(t, false)
			lookup.apiRoot = server.URL
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
			lookup.typeDefaults.AccessToken = ""
			lookup.typeHardDefaults.AccessToken = "default-token"
			switch tc.conditions.accessToken {
			case "unset":
				lookup.AccessToken = ""
			case "default":
				lookup.AccessToken = "default-token"
			default:
				lookup.AccessToken = "service-token"
			}
			if tc.lookupSetup != nil {
				tc.lookupSetup(lookup)
			}

			logFrom := logx.LogFrom{Primary: "TestHandleResponse", Secondary: tc.name}

			// WHEN: handleResponse is called on it.
			gotBody, nextPage, err := lookup.handleResponse(resp, tc.body, logFrom)

			e := errfmt.FormatError(err)

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

			// AND: TagFallback is as expected.
			if got := lookup.data.TagFallback(); got != tc.want.tagFallback {
				t.Errorf(
					"%s TagFallback mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.want.tagFallback,
				)
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
		releaseOverrides *forgetypes.Release
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
			releaseOverrides: &forgetypes.Release{
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
			releaseOverrides: &forgetypes.Release{
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
			releaseOverrides: &forgetypes.Release{
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

			// WHEN: forge.ReleaseMeetsRequirements is called on it.
			version, releaseDate, err := forge.ReleaseMeetsRequirements(
				testRelease, lookup.Require, lookup.GetServiceID(), logFrom,
			)

			prefix := fmt.Sprintf(
				"%s\nforge.ReleaseMeetsRequirements(%+v)",
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
		hadReleases     []forgetypes.Release
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
			bodyOverride: new("invalid"),
			want: want{
				errRegex: `unmarshal of GitHub API data failed`,
			},
		},
		{
			name:            "cached releases, used when empty body",
			bodyOverride:    new(""),
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
			bodyOverride:    new(""),
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
			bodyOverride: new(`[{"tag_name":"v1.2.3","published_at":"2021-01-01T00:00:00Z"}]`),
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
			if err != nil {
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
				func(gotRelease forgetypes.Release, wantRelease forgetypes.Release) bool {
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

// nonSemanticBody is a release whose tag only yields a version once
// url_commands have run.
