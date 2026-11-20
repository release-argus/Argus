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
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/Masterminds/semver/v3"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	lvbasetest "github.com/release-argus/Argus/service/latest_version/types/base/test"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/util/polymorphic"
)

// nonSemanticReleasesBody is a JSON array of releases that do not have semantic versions.
var nonSemanticReleasesBody = test.TrimJSON(`[
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
	tests := []struct {
		name        string
		endpoints   map[string]githubEndpoint
		requireAuth string // Authorization the API required.
		overrides   string
		semVer      bool
		hadVersion  string
		wantVersion string
		wantNew     bool
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
			name: "valid/a release newer than the recorded version is new",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: string(testBody)},
			},
			semVer:      true,
			hadVersion:  "0.17.3",
			wantVersion: "0.17.4",
			wantNew:     true,
			errRegex:    `^$`,
		},
		{
			name: "invalid/a non-semantic version is rejected when semver is required",
			endpoints: map[string]githubEndpoint{
				"releases": {Body: nonSemanticReleasesBody},
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
				"releases": {Body: nonSemanticReleasesBody},
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

			lookup := testLookup(t)
			lookup.Status.ServiceInfo.ID = tc.name
			if err := lookup.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf("Lookup.ApplyOverrides() failed: %v", err)
			}
			lookup.Options.SemanticVersioning = &tc.semVer
			lookup.Status.SetLatestVersion(tc.hadVersion, "", false)

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

			newVersion, err := lookup.Query(true, logx.LogFrom{Primary: t.Name()})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if newVersion != tc.wantNew {
				t.Errorf("newVersion mismatch\ngot:  %t\nwant: %t", newVersion, tc.wantNew)
			}
			if got := lookup.Status.LatestVersion(); got != tc.wantVersion {
				t.Errorf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, tc.wantVersion)
			}
		})
	}
}

func TestLookup_Query__eTagReusesTheCache(t *testing.T) {
	hadETag := getEmptyListETag()
	t.Cleanup(func() { setEmptyListETag(hadETag) })
	// No empty-list ETag learnt yet, so the first request is unconditional.
	setEmptyListETag("")

	server := newGitHubServer(t, map[string]githubEndpoint{
		"releases": {Body: string(testBody), ETag: `W/"unchanged"`},
	})

	lookup := testLookup(t)
	lookup.apiRoot = server.URL
	lookup.Status.ServiceInfo.ID = t.Name()

	for attempt := 1; attempt <= 3; attempt++ {
		_, err := lookup.Query(true, logx.LogFrom{Primary: t.Name()})
		if err != nil {
			t.Fatalf("attempt %d: Lookup.Query() failed: %v", attempt, errfmt.FormatError(err))
		}

		if got, want := lookup.Status.LatestVersion(), "0.17.4"; got != want {
			t.Fatalf(
				"attempt %d: Status.LatestVersion() mismatch\ngot:  %q\nwant: %q",
				attempt, got, want,
			)
		}
	}

	requests := server.Requests()
	if len(requests) < 3 {
		t.Fatalf(
			"Server.Requests() length mismatch\ngot:  %d %v\nwant: >= 3",
			len(requests), requests,
		)
	}
	if got, want := server.NotModifiedCount(), len(requests)-1; got != want {
		t.Errorf(
			"Server.NotModifiedCount() mismatch\ngot:  %d of %d request(s)\nwant: %d",
			got, len(requests), want,
		)
	}
	for i, request := range requests {
		want := `"unchanged"`
		if i == 0 {
			want = ""
		}
		if got := request.Header.Get("If-None-Match"); got != want {
			t.Errorf(
				"request %d: RecordedRequest.Header[If-None-Match] mismatch\ngot:  %q\nwant: %q",
				i, got, want,
			)
		}
	}
}

func TestLookup_Query__recheckUsesTheFoundPage(t *testing.T) {
	prereleaseOnlyBody := test.TrimJSON(`[
		{"tag_name":"0.18.0","name":"0.18.0","prerelease":true,"published_at":"2024-05-07T13:10:29Z"}
	]`)
	server := newGitHubServer(t, map[string]githubEndpoint{
		"releases": {
			Pages:     []string{prereleaseOnlyBody, string(testBody)},
			PageETags: []string{`"p1"`, `"p2"`},
		},
	})

	lookup := testLookup(t)
	lookup.apiRoot = server.URL
	lookup.Status.ServiceInfo.ID = t.Name()

	_, err := lookup.Query(true, logx.LogFrom{Primary: t.Name()})
	if err != nil {
		t.Fatalf("unexpected error: %v", errfmt.FormatError(err))
	}

	if got, want := lookup.Status.LatestVersion(), "0.17.4"; got != want {
		t.Errorf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, want)
	}

	pages := requestedPages(server)
	if want := []string{"", "2", "2"}; !slices.Equal(pages, want) {
		t.Errorf("Server.Requests() pages mismatch\ngot:  %q\nwant: %q", pages, want)
	}
	// The re-check asks with page 2's ETag, which only page 2 answers with 304.
	if got, want := server.NotModifiedCount(), 1; got != want {
		t.Errorf("Server.NotModifiedCount() mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestLookup_Query__cachedLastPageEndsTheWalk(t *testing.T) {
	server := newGitHubServer(t, map[string]githubEndpoint{
		"releases": {Pages: []string{string(testBody)}, PageETags: []string{`"p1"`}},
	})

	lookup := testLookup(t)
	lookup.apiRoot = server.URL
	lookup.Status.ServiceInfo.ID = t.Name()
	lookup.data.SetResponse(`"p1"`, 0, testBodyObject)
	lookup.URLCommands = filter.URLCommands{{Type: "regex", Regex: `^FOO$`}}

	_, err := lookup.Query(true, logx.LogFrom{Primary: t.Name()})
	e := errfmt.FormatError(err)
	if want := `^no releases were found matching the url_commands on page 1 `; !util.RegexCheck(want, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, want)
	}

	pages := requestedPages(server)
	// Page 1 is the last page, so a filter miss on its cached releases requests no page 2.
	if want := []string{""}; !slices.Equal(pages, want) {
		t.Errorf("Server.Requests() pages mismatch\ngot:  %q\nwant: %q", pages, want)
	}
	if got, want := server.NotModifiedCount(), 1; got != want {
		t.Errorf("Server.NotModifiedCount() mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestLookup_HTTPRequest(t *testing.T) {
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
			// Serial: shares the empty-list ETag.

			lookup := testLookup(t)
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

			nextPage, err := lookup.httpRequest(1, logx.LogFrom{Primary: t.Name()})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if nextPage != tc.nextPage {
				t.Errorf("nextPage mismatch\ngot:  %d\nwant: %d", nextPage, tc.nextPage)
			}

			if tc.wantPaths != nil {
				var paths []string
				for _, request := range server.Requests() {
					if len(paths) == 0 || paths[len(paths)-1] != request.Path {
						paths = append(paths, request.Path)
					}
				}
				if !slices.Equal(paths, tc.wantPaths) {
					t.Errorf(
						"Server.Requests() paths mismatch\ngot:  %q\nwant: %q",
						paths, tc.wantPaths,
					)
				}
			}

			if tc.wantETagSet != "" {
				if got := getEmptyListETag(); got != tc.wantETagSet {
					t.Errorf("getEmptyListETag() mismatch\ngot:  %q\nwant: %q", got, tc.wantETagSet)
				}
			}
		})
	}
}

func TestLookup_GetResponse__readError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The Content-Length promises data that is never sent.
			w.Header().Set("Content-Length", "10")
			w.WriteHeader(http.StatusOK)
			// Immediately close the connection without writing any body.
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Errorf("Hijacker.Hijack() failed: %v", err)
				return
			}
			conn.Close()
		}),
	)
	t.Cleanup(server.Close)

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("http.NewRequest() failed: %v", err)
	}

	l := Lookup{}
	wantRe := `unexpected EOF|connection reset|body length`
	response, body, err := l.getResponse(req, logx.LogFrom{})
	e := errfmt.FormatError(err)
	if !util.RegexCheck(wantRe, e) {
		t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, wantRe)
	}
	if response != nil || body != nil {
		t.Fatalf(
			"response, body mismatch\n"+
				"got:  response nil=%t, body nil=%t\nwant: response nil=true, body nil=true",
			response == nil, body == nil,
		)
	}
}

func TestLookup_HandleResponse(t *testing.T) {
	const (
		servedETag = "<served>" // Stands in for the ETag of the response under test.
		tagsETag   = "tags-etag"
	)

	type wants struct {
		hasReleases      bool
		nextPage         int
		setEmptyListETag bool
		etag             string // ETag stored after the response.
		tagFallback      bool
		errRegex         string
	}
	type conditions struct {
		hadReleases bool
		nextPage    int    // Next page the cached response named.
		accessToken string // "unset" | "default", or anything else for a non-default token.
		page        int    // Page requested; 0 is a /tags fallback, so also the first.
		link        string
	}

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
				etag:             tagsETag,
				hasReleases:      true,
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
				etag:             tagsETag,
				hasReleases:      true,
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
				etag:             tagsETag,
				hasReleases:      true,
				nextPage:         2,
				setEmptyListETag: false,
				tagFallback:      true,
				errRegex:         `^$`,
			},
		},
		{
			name: "200 OK/an empty later page ends the list without tag fallback",
			conditions: conditions{
				accessToken: "unset",
				page:        2,
			},
			statusCode: http.StatusOK,
			body:       []byte(`[]`),
			want: wants{
				etag:             servedETag,
				nextPage:         0,
				setEmptyListETag: false,
				tagFallback:      false,
				errRegex:         `^$`,
			},
		},
		{
			name:       "200 OK/a body that fails to parse",
			statusCode: http.StatusOK,
			body:       []byte(`{"tag_name":"v1.0.0"}`),
			want: wants{
				etag:     "old",
				nextPage: 0,
				errRegex: `^release data failed to parse`,
			},
		},
		{
			name:       "200 OK/Get releases",
			statusCode: http.StatusOK,
			body:       []byte(`[{"tag_name":"v1.0.0"}]`),
			want: wants{
				etag:        servedETag,
				hasReleases: true,
				nextPage:    0,
				tagFallback: false,
				errRegex:    `^$`,
			},
		},
		{
			name: "200 OK/Get releases with a next page",
			conditions: conditions{
				link: `</repos/o/r/releases?page=2>; rel="next"`,
			},
			statusCode: http.StatusOK,
			body:       []byte(`[{"tag_name":"v1.0.0"}]`),
			want: wants{
				etag:        servedETag,
				hasReleases: true,
				nextPage:    2,
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
				etag:        tagsETag,
				hasReleases: true,
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
				etag:        servedETag,
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
				etag:        tagsETag,
				hasReleases: true,
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
				etag:        "old",
				nextPage:    0,
				tagFallback: false,
				errRegex:    `^$`,
			},
		},
		{
			name:       "304 Not Modified/Tag fallback request",
			statusCode: http.StatusNotModified,
			want: wants{
				etag:        tagsETag,
				hasReleases: true,
				nextPage:    2,
				tagFallback: true,
				errRegex:    `^$`,
			},
		},
		{
			name: "304 Not Modified/Use old releases and their next page",
			conditions: conditions{
				hadReleases: true,
				nextPage:    3,
			},
			statusCode: http.StatusNotModified,
			want: wants{
				etag:        "old",
				hasReleases: true,
				nextPage:    3,
				tagFallback: false,
				errRegex:    `^$`,
			},
		},
		{
			name: "304 Not Modified/old releases from the last page end the walk",
			conditions: conditions{
				hadReleases: true,
			},
			statusCode: http.StatusNotModified,
			want: wants{
				etag:        "old",
				hasReleases: true,
				nextPage:    0,
				tagFallback: false,
				errRegex:    `^$`,
			},
		},
		{
			name:       "401 Unauthorized/bad credentials",
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"message":"Bad credentials"}`),
			want: wants{
				etag:     "old",
				errRegex: `github access token is invalid`,
			},
		},
		{
			name:       "401 Unauthorized/unknown",
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"message":"Something else"}`),
			want: wants{
				etag:     "old",
				errRegex: `unknown 401 response`,
			},
		},
		{
			name:       "403 Forbidden/rate limit",
			statusCode: http.StatusForbidden,
			body:       []byte(`{"message":"API rate limit exceeded"}`),
			want: wants{
				etag:     "old",
				errRegex: `rate limit reached for GitHub`,
			},
		},
		{
			name:       "403 Forbidden/missing tag_name",
			statusCode: http.StatusForbidden,
			body:       []byte(`{"message":"some other error"}`),
			want: wants{
				etag:     "old",
				errRegex: `tag_name not found at`,
			},
		},
		{
			name:       "403 Forbidden/unknown",
			statusCode: http.StatusForbidden,
			body:       []byte(`[{"tag_name":"v1.0.0"}]`),
			want: wants{
				etag:     "old",
				errRegex: `unknown 403 response`,
			},
		},
		{
			name:       "429 Too Many Requests/rate limit",
			statusCode: http.StatusTooManyRequests,
			body:       []byte(`{"message":"something from GitHub"}`),
			want: wants{
				etag:     "old",
				errRegex: `^too many requests made to GitHub - "something from GitHub"$`,
			},
		},
		{
			name:       "429 Too Many Requests/unmarshal fail",
			statusCode: http.StatusTooManyRequests,
			body:       []byte(`{"message":}`),
			want: wants{
				etag:     "old",
				errRegex: `^too many requests made to GitHub$`,
			},
		},
		{
			name:       "Unknown status code",
			statusCode: http.StatusTeapot,
			body:       []byte(`{"message":"I'm a teapot"}`),
			want: wants{
				etag:     "old",
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
			// Serial: shares the empty-list ETag.

			server := newGitHubServer(t, map[string]githubEndpoint{
				"tags": {Pages: []string{string(testBody), string(testBody)}, ETag: tagsETag},
			})

			lookup := testLookup(t)
			lookup.apiRoot = server.URL
			resp := &http.Response{
				StatusCode: tc.statusCode,
				Header:     http.Header{},
			}
			etag := tc.name
			resp.Header.Add("ETag", etag)
			if tc.conditions.link != "" {
				resp.Header.Set("Link", tc.conditions.link)
			}
			lookup.data.etag = "old"
			if tc.conditions.hadReleases {
				lookup.data.releases = testBodyObject
			}
			lookup.data.nextPage = tc.conditions.nextPage
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

			nextPage, err := lookup.handleResponse(resp, tc.body, tc.conditions.page, logFrom)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.want.errRegex)
			}

			wantETag := tc.want.etag
			if wantETag == servedETag {
				wantETag = etag
			}
			if got := lookup.data.ETag(); got != wantETag {
				t.Errorf("Data.ETag() mismatch\ngot:  %q\nwant: %q", got, wantETag)
			}
			if got := lookup.data.hasReleases(); got != tc.want.hasReleases {
				t.Errorf("Data.hasReleases() mismatch\ngot:  %t\nwant: %t", got, tc.want.hasReleases)
			}

			emptyListETag := getEmptyListETag()
			if tc.want.setEmptyListETag && emptyListETag != etag {
				t.Errorf("getEmptyListETag() mismatch\ngot:  %q\nwant: %q", emptyListETag, etag)
			} else if !tc.want.setEmptyListETag && emptyListETag == etag {
				t.Errorf("getEmptyListETag() mismatch\ngot:  %q\nwant: NOT %q", emptyListETag, etag)
			}

			if nextPage != tc.want.nextPage {
				t.Errorf("nextPage mismatch\ngot:  %d\nwant: %d", nextPage, tc.want.nextPage)
			}
			if got := lookup.data.NextPage(); got != tc.want.nextPage {
				t.Errorf("Data.NextPage() mismatch\ngot:  %d\nwant: %d", got, tc.want.nextPage)
			}

			if got := lookup.data.TagFallback(); got != tc.want.tagFallback {
				t.Errorf(
					"Data.TagFallback() mismatch\ngot:  %t\nwant: %t",
					got, tc.want.tagFallback,
				)
			}
		})
	}
}

func TestLookup_ReleaseMeetsRequirements(t *testing.T) {
	lvCfg := lvbasetest.PlainDefaultsConfig()

	type wants struct {
		version     string
		releaseDate string
		errRegex    string
	}

	defaultRelease := testBodyObject[0]

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
				errRegex:    `^$`,
			},
		},
		{
			name: "no requirements/use semantic version",
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

			lookup := testLookup(t)
			if err := lookup.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf("Lookup.ApplyOverrides() failed: %v", err)
			}
			requireData, err := polymorphic.Extract("yaml", []byte(tc.overrides), "require")
			if err != nil {
				t.Fatalf("polymorphic.Extract() failed: %v", err)
			}
			req, err := filter.Decode(
				"yaml", requireData,
				lookup.Status,
				&lvCfg.Soft.Require,
			)
			if err != nil {
				t.Fatalf("filter.Decode() failed: %v", err)
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

			version, releaseDate, err := forge.ReleaseMeetsRequirements(
				testRelease, lookup.Require, lookup.GetServiceID(), logFrom,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.want.errRegex)
			}
			if err != nil {
				return
			}

			if version != tc.want.version {
				t.Errorf("version mismatch\ngot:  %q\nwant: %q", version, tc.want.version)
			}

			if releaseDate != tc.want.releaseDate {
				t.Errorf(
					"releaseDate mismatch\ngot:  %q\nwant: %q",
					releaseDate, tc.want.releaseDate,
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

	bodyObject := testBodyObject
	tests := []struct {
		name            string
		lookupOverrides string
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
			name:            "use_prerelease false skips the pre-release",
			lookupOverrides: `use_prerelease: false`,
			want: want{
				version:     bodyObject[1].TagName,
				releaseDate: bodyObject[1].PublishedAt,
				errRegex:    `^$`,
			},
		},
		{
			name:            "use_prerelease true takes the pre-release",
			lookupOverrides: `use_prerelease: true`,
			want: want{
				version:     bodyObject[0].TagName,
				releaseDate: bodyObject[0].PublishedAt,
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

			lookup := testLookup(t)
			if err := lookup.ApplyOverrides("yaml", []byte(tc.lookupOverrides)); err != nil {
				t.Fatalf("Lookup.ApplyOverrides() failed: %v", err)
			}
			lookup.data.releases = bodyObject
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

			version, releaseDate, err := lookup.getVersion(1, logFrom)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.want.errRegex)
			}
			if err != nil {
				return
			}

			if version != tc.want.version {
				t.Errorf("version mismatch\ngot:  %q\nwant: %q", version, tc.want.version)
			}

			if releaseDate != tc.want.releaseDate {
				t.Errorf(
					"releaseDate mismatch\ngot:  %q\nwant: %q",
					releaseDate, tc.want.releaseDate,
				)
			}
		})
	}
}

func TestLookup_SetReleases(t *testing.T) {
	tests := []struct {
		name      string
		overrides string
		body      string
		errRegex  string
	}{
		{
			name:      "use_prerelease false still stores pre-releases",
			overrides: `use_prerelease: false`,
			errRegex:  `^$`,
		},
		{
			name:      "use_prerelease true stores pre-releases",
			overrides: `use_prerelease: true`,
			errRegex:  `^$`,
		},
		{
			name: "release body that's an object, not a list",
			body: `{"tag_name":"v1.2.3","published_at":"2021-01-01T00:00:00Z"}`,
			errRegex: test.TrimYAML(`
				^release data failed to parse:
					unmarshal of GitHub API data failed`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t)
			if err := lookup.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf("Lookup.ApplyOverrides() failed: %v", err)
			}
			lookup.data.SetResponse("old", 1, nil)
			body := testBody
			if tc.body != "" {
				body = []byte(tc.body)
			}

			err := lookup.setReleases(body, "new", 2)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				gotETag, gotNextPage := lookup.data.ETag(), lookup.data.NextPage()
				gotReleases := len(lookup.data.Releases())
				if gotETag != "old" || gotNextPage != 1 || gotReleases != 0 {
					t.Errorf(
						"Data mismatch\ngot:  etag=%q, nextPage=%d, releases=%d\n"+
							"want: etag=\"old\", nextPage=1, releases=0",
						gotETag, gotNextPage, gotReleases,
					)
				}
				return
			}

			if got, want := lookup.data.ETag(), "new"; got != want {
				t.Errorf("Data.ETag() mismatch\ngot:  %q\nwant: %q", got, want)
			}
			if got, want := lookup.data.NextPage(), 2; got != want {
				t.Errorf("Data.NextPage() mismatch\ngot:  %d\nwant: %d", got, want)
			}

			if testErr := test.AssertSlicesEqualFunc(
				t,
				lookup.data.Releases(),
				testBodyObject,
				func(gotRelease forgetypes.Release, wantRelease forgetypes.Release) bool {
					if gotRelease.TagName != wantRelease.TagName ||
						gotRelease.PublishedAt != wantRelease.PublishedAt ||
						gotRelease.PreRelease != wantRelease.PreRelease {
						return false
					}
					if len(gotRelease.Assets) != len(wantRelease.Assets) {
						return false
					}
					for i := range gotRelease.Assets {
						if gotRelease.Assets[i].Name != wantRelease.Assets[i].Name {
							return false
						}
					}
					return true
				},
				"Data.Releases()",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
