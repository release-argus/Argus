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

// Package forgejo provides a Forgejo-based lookup type.
package forgejo

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	lvbasetest "github.com/release-argus/Argus/service/latest_version/types/base/test"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	forgetest "github.com/release-argus/Argus/service/latest_version/types/forge/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_Query(t *testing.T) {
	tests := []struct {
		name         string
		endpoints    map[string]forgeEndpoint
		tls          bool
		requiredAuth string // Answer 401 to any request not carrying this.
		defaults     map[string]HostDefaults
		hostSuffix   string
		hostNamed    bool // Address the instance by its defaults name.
		noHost       bool // Give the Lookup no host.
		closeServer  bool // Stop the instance before querying.
		metrics      bool // Record Prometheus metrics for the query.
		lookupYAML   string
		wantVersion  string
		wantPaths    []string // Distinct API paths requested.
		errRegex     string
	}{
		{
			name: "valid/the highest non-prerelease version is reported",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			lookupYAML:  `url: Release-Argus/Argus`,
			metrics:     true,
			wantVersion: "0.17.4",
			errRegex:    `^$`,
		},
		{
			name: "valid/the highest semantic version wins, not the most recently published",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					Body: test.TrimJSON(`[
						{"tag_name":"1.0.0","prerelease":false,"published_at":"2000-01-01T00:00:00Z"},
						{"tag_name":"0.9.9","prerelease":false,"published_at":"2000-01-02T00:00:00Z"}
					]`),
				},
			},
			lookupYAML:  `url: owner/repo`,
			wantVersion: "1.0.0",
			errRegex:    `^$`,
		},
		{
			name: "valid/tags fallback, on the newline-terminated empty-list encoding",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
				endpointTags:     {Body: tagsBody},
			},
			lookupYAML:  `url: owner/repo`,
			wantVersion: "0.12.1",
			errRegex:    `^$`,
		},
		{
			name: "valid/tags fallback, on the compact empty-list encoding",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListCompact},
				endpointTags:     {Body: tagsBody},
			},
			lookupYAML:  `url: owner/repo`,
			wantVersion: "0.12.1",
			errRegex:    `^$`,
		},
		{
			name: "valid/releases disabled (404), tags fallback",
			endpoints: map[string]forgeEndpoint{
				endpointTags: {Body: tagsBody},
			},
			lookupYAML:  `url: owner/repo`,
			wantVersion: "0.12.1",
			errRegex:    `^$`,
		},
		{
			name: "valid/url_commands filter the tag",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: test.TrimJSON(`[
					{"tag_name":"release-1.2.3","prerelease":false,"published_at":"2026-01-01T00:00:00Z"}
				]`)},
			},
			lookupYAML: `
				url: owner/repo
				url_commands:
					- type: regex
						regex: '([0-9.]+)$'`,
			wantVersion: "1.2.3",
			errRegex:    `^$`,
		},
		{
			name: "valid/require regex_version narrows the release taken",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: test.TrimJSON(`[
					{"tag_name":"2.0.0","prerelease":false,"published_at":"2026-02-01T00:00:00Z"},
					{"tag_name":"1.9.0","prerelease":false,"published_at":"2026-01-01T00:00:00Z"}
				]`)},
			},
			lookupYAML: `
				url: owner/repo
				require:
					regex_version: '^1\.'`,
			wantVersion: "1.9.0",
			errRegex:    `^$`,
		},
		{
			name: "invalid/no releases nor tags",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
				endpointTags:     {Body: emptyListNewline},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^no releases or tags were found$`,
		},
		{
			name: "invalid/releases empty, tags 404",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^no releases were found, and the repository's tags are disabled$`,
		},
		{
			name: "invalid/releases 404, tags empty",
			endpoints: map[string]forgeEndpoint{
				endpointTags: {Body: emptyListCompact},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^no tags were found, and the repository's releases are disabled$`,
		},
		{
			name:       "invalid/no such repository names both possible causes",
			endpoints:  map[string]forgeEndpoint{},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^repository not found, or its releases and tags are disabled$`,
		},
		{
			name:      "invalid/releases 404, tags 404, with use_prereleases",
			endpoints: map[string]forgeEndpoint{},
			lookupYAML: `
				url: owner/repo
				use_prerelease: true`,
			errRegex: `^repository not found, or its releases and tags are disabled$`,
		},
		{
			name: "invalid/401",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					Status: http.StatusUnauthorized,
					Body:   `{"message":"token is required"}`,
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^authentication failed for 127\.0\.0\.1:\d+ \(401\) - no access_token was sent$`,
		},
		{
			name: "invalid/403 without rate-limit headers",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					Status: http.StatusForbidden,
					Body:   `{"message":"forbidden"}`,
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^authentication failed for 127\.0\.0\.1:\d+ \(403\) - no access_token was sent$`,
		},
		{
			name: "invalid/403 with rate-limit headers",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					Status:  http.StatusForbidden,
					Headers: map[string]string{"RateLimit": `"baseline";r=0;t=600`},
					Body:    `{"message":"forbidden"}`,
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^rate limit reached for 127\.0\.0\.1:\d+ - no access_token was sent$`,
		},
		{
			name: "invalid/429",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					Status: http.StatusTooManyRequests,
					Body:   `<!DOCTYPE html><html class="codeberg-design"><body>Too many requests</body></html>`,
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^rate limit reached for 127\.0\.0\.1:\d+ - no access_token was sent$`,
		},
		{
			name: "invalid/429 with a retry window",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					Status:  http.StatusTooManyRequests,
					Headers: map[string]string{"Retry-After": "120"},
					Body:    `<!DOCTYPE html><html><body>Too many requests</body></html>`,
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^rate limit reached for 127\.0\.0\.1:\d+ - retry after 120 - no access_token was sent$`,
		},
		{
			name: "invalid/unmapped status code is reported with its body",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					Status: http.StatusInternalServerError,
					Body:   "proxy error",
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^unknown status code 500\nproxy error$`,
		},
		{
			name: "invalid/body that is not a release list",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: `<!DOCTYPE html><html><body>not the API</body></html>`},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `release data failed to parse`,
		},
		{
			name: "valid/service's own token passes",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			requiredAuth: "Bearer host-token",
			lookupYAML: `
				url: owner/repo
				access_token: host-token`,
			wantVersion: "0.17.4",
			errRegex:    `^$`,
		},
		{
			name: "valid/token from the named host entry passes",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			requiredAuth: "Bearer host-token",
			defaults:     map[string]HostDefaults{"Fixture": {URL: hostPlaceholder, AccessToken: "host-token"}},
			hostNamed:    true,
			lookupYAML:   `url: owner/repo`,
			wantVersion:  "0.17.4",
			errRegex:     `^$`,
		},
		{
			name: "invalid/token bound to another host is not sent",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			requiredAuth: "Bearer host-token",
			defaults: map[string]HostDefaults{
				"https://forge.example.com": {URL: "https://forge.example.com", AccessToken: "host-token"},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^authentication failed for 127\.0\.0\.1:\d+ \(401\) - no access_token was sent$`,
		},
		{
			name: "valid/host entry relaxes certificate trust for this instance",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			tls:         true,
			defaults:    map[string]HostDefaults{"Fixture": {URL: hostPlaceholder, AllowInvalidCerts: new(true)}},
			hostNamed:   true,
			lookupYAML:  `url: owner/repo`,
			wantVersion: "0.17.4",
			errRegex:    `^$`,
		},
		{
			name: "invalid/untrusted certificate is rejected",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			tls:        true,
			lookupYAML: `url: owner/repo`,
			errRegex:   `certificate|x509`,
		},
		{
			name: "invalid/certificate trust relaxed for another host does not apply here",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			tls: true,
			defaults: map[string]HostDefaults{
				"https://forge.example.com": {AllowInvalidCerts: new(true)},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `certificate|x509`,
		},
		{
			name: "invalid/no release matches the url_commands",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: test.TrimJSON(`[
					{"tag_name":"not-a-version","prerelease":false,"published_at":"2026-01-01T00:00:00Z"}
				]`)},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^no releases were found matching the url_commands on page 1 of the API response$`,
		},
		{
			name: "valid/an instance served under a sub-path prefixes the API path",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			hostSuffix:  "/git",
			lookupYAML:  `url: owner/repo`,
			wantVersion: "0.17.4",
			wantPaths:   []string{"/git/api/v1/repos/owner/repo/" + endpointReleases},
			errRegex:    `^$`,
		},
		{
			name:       "invalid/no host to build a request URL from",
			noHost:     true,
			lookupYAML: `url: owner/repo`,
			errRegex:   `no Host in request URL`,
		},
		{
			name: "invalid/an instance that is not listening",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			closeServer: true,
			lookupYAML:  `url: owner/repo`,
			errRegex:    `connection refused|connect: `,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			start := httptest.NewServer
			if tc.tls {
				start = httptest.NewTLSServer
			}
			server := startForgeServer(
				t,
				&forgeServer{Endpoints: tc.endpoints, RequireAuth: tc.requiredAuth},
				start,
			)
			if tc.closeServer {
				server.Close()
			}

			lookupYAML := test.TrimYAML(tc.lookupYAML)
			if tc.hostNamed {
				lookupYAML = "host: Fixture\n" + lookupYAML
			} else if !tc.noHost {
				lookupYAML = "host: " + server.URL + tc.hostSuffix + "\n" + lookupYAML
			}
			lookup := testLookup(t, lookupYAML)
			setHostDefaults(lookup, hostKeyedAt(tc.defaults, server.URL), nil)

			_, err := lookup.Query(tc.metrics, logx.LogFrom{Primary: t.Name()})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if got := lookup.Status.LatestVersion(); got != tc.wantVersion {
				t.Fatalf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, tc.wantVersion)
			}

			if tc.wantPaths != nil {
				var paths []string
				for _, request := range server.Requests() {
					if len(paths) == 0 || paths[len(paths)-1] != request.Path {
						paths = append(paths, request.Path)
					}
				}
				if !slices.Equal(paths, tc.wantPaths) {
					t.Fatalf(
						"Server.Requests() paths mismatch\ngot:  %q\nwant: %q",
						paths, tc.wantPaths,
					)
				}
			}
		})
	}
}

func TestLookup_Query__Paginates(t *testing.T) {
	tests := []struct {
		name        string
		endpoints   map[string]forgeEndpoint
		wantVersion string
		wantOn      string
	}{
		{
			name: "releases",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Pages: []string{
					test.TrimJSON(`[{"tag_name":"not-a-version","prerelease":false}]`),
					test.TrimJSON(`[{"tag_name":"1.2.3","prerelease":false}]`),
				}},
			},
			wantVersion: "1.2.3",
			wantOn:      endpointReleases,
		},
		{
			name: "tags, after falling back to them",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
				endpointTags: {Pages: []string{
					test.TrimJSON(`[{"name":"not-a-version"}]`),
					test.TrimJSON(`[{"name":"3.2.1"}]`),
				}},
			},
			wantVersion: "3.2.1",
			wantOn:      endpointTags,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newForgeServer(t, tc.endpoints)
			lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")

			_, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()})
			if err != nil {
				t.Fatalf("Lookup.Query() failed: %v", err)
			}

			if got := lookup.Status.LatestVersion(); got != tc.wantVersion {
				t.Fatalf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, tc.wantVersion)
			}

			requests := server.Requests()
			if !slices.ContainsFunc(requests, func(r forgetest.RecordedRequest) bool {
				return r.Query.Get("page") == "2" && strings.HasSuffix(r.Path, "/"+tc.wantOn)
			}) {
				t.Fatalf(
					"Server.Requests() mismatch\ngot:  %v\nwant: page=2 of /%s",
					requests, tc.wantOn,
				)
			}

			for i, request := range requests {
				if got, want := request.Query.Get("limit"), strconv.Itoa(apiPageSize); got != want {
					t.Fatalf(
						"request %d: RecordedRequest.Query[limit] mismatch\n"+
							"got:  %q (path=%q)\nwant: %q",
						i, got, request.Path, want,
					)
				}
			}
		})
	}
}

func TestLookup_Query__CachesNothingBetweenQueries(t *testing.T) {
	server := newForgeServer(t, map[string]forgeEndpoint{
		endpointReleases: {Body: releasesBody},
	})
	lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")
	logFrom := logx.LogFrom{Primary: t.Name()}

	_, err := lookup.Query(false, logFrom)
	if err != nil {
		t.Fatalf("first: Lookup.Query() failed: %v", err)
	}
	afterFirst := len(server.Requests())

	_, err = lookup.Query(false, logFrom)
	if err != nil {
		t.Fatalf("second: Lookup.Query() failed: %v", err)
	}

	if got := len(server.Requests()); got <= afterFirst {
		t.Fatalf("Server.Requests() length mismatch\ngot:  %d\nwant: > %d", got, afterFirst)
	}
}

func TestLookup_Query__NoTagsFallback(t *testing.T) {
	tests := []struct {
		name       string
		lookupYAML string
	}{
		{
			name: "require.regex_content/tags have no release assets",
			lookupYAML: `
				url: owner/repo
				require:
					regex_content: 'Argus-{{ version }}.linux-amd64'`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newForgeServer(t, map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
				endpointTags:     {Body: tagsBody},
			})
			lookupYAML := "host: " + server.URL + "\n" + test.TrimYAML(tc.lookupYAML)
			lookup := testLookup(t, lookupYAML)

			_, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()})
			e := errfmt.FormatError(err)
			if wantRe := `^no releases were found$`; !util.RegexCheck(wantRe, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, wantRe)
			}

			for _, request := range server.Requests() {
				if strings.HasSuffix(request.Path, "/"+endpointTags) {
					t.Fatalf(
						"RecordedRequest.Path mismatch\ngot:  %q\nwant: NOT %q",
						request.Path, "/"+endpointTags,
					)
				}
			}
		})
	}
}

func TestLookup_Query__EndpointFallback(t *testing.T) {
	tests := []struct {
		name          string
		endpoints     map[string]forgeEndpoint
		lookupYAML    string
		wantRequested []string
		wantLatest    string
		errRegex      string
	}{
		{
			name: "releases answers, so tags skipped",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
				endpointTags:     {Body: tagsBody},
			},
			lookupYAML:    `url: owner/repo`,
			wantRequested: []string{endpointReleases, endpointReleases},
			wantLatest:    "0.17.4",
			errRegex:      `^$`,
		},
		{
			name: "empty releases list falls through to tags",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
				endpointTags:     {Body: tagsBody},
			},
			lookupYAML:    `url: owner/repo`,
			wantRequested: []string{endpointReleases, endpointTags, endpointTags},
			wantLatest:    "0.12.1",
			errRegex:      `^$`,
		},
		{
			name: "filters that tags cannot satisfy (regex_content) stop the fallback",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
				endpointTags:     {Body: tagsBody},
			},
			lookupYAML: `
				url: owner/repo
				require:
					regex_content: 'Argus-{{ version }}.linux-amd64'`,
			wantRequested: []string{endpointReleases},
			errRegex:      `^no releases were found$`,
		},
		{
			name: "wanting prereleases doesn't stop the tags fallback",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
				endpointTags:     {Body: tagsBody},
			},
			lookupYAML: `
				url: owner/repo
				use_prerelease: true`,
			wantRequested: []string{endpointReleases, endpointTags, endpointTags},
			wantLatest:    "0.12.1",
			errRegex:      `^$`,
		},
		{
			name: "pre-release tag is excluded by default as its name labels it",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
				endpointTags:     {Body: preReleaseTagsBody},
			},
			lookupYAML:    `url: owner/repo`,
			wantRequested: []string{endpointReleases, endpointTags, endpointTags},
			wantLatest:    "0.12.1",
			errRegex:      `^$`,
		},
		{
			name: "pre-release tag is taken when prereleases are wanted",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
				endpointTags:     {Body: preReleaseTagsBody},
			},
			lookupYAML: `
				url: owner/repo
				use_prerelease: true`,
			wantRequested: []string{endpointReleases, endpointTags, endpointTags},
			wantLatest:    "0.13.0-rc1",
			errRegex:      `^$`,
		},
		{
			name: "rate-limit response stops the fallback",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Status: http.StatusTooManyRequests, Body: "<!DOCTYPE html>"},
				endpointTags:     {Body: tagsBody},
			},
			lookupYAML:    `url: owner/repo`,
			wantRequested: []string{endpointReleases},
			errRegex:      `^rate limit reached for 127\.0\.0\.1:\d+ - no access_token was sent$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newForgeServer(t, tc.endpoints)
			lookupYAML := "host: " + server.URL + "\n" + test.TrimYAML(tc.lookupYAML)
			lookup := testLookup(t, lookupYAML)

			_, err := lookup.query(true, logx.LogFrom{Primary: t.Name()})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if got := lookup.Status.LatestVersion(); got != tc.wantLatest {
				t.Fatalf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, tc.wantLatest)
			}

			requested := test.Attributes(server.Requests(), func(r forgetest.RecordedRequest) string {
				return path.Base(r.Path)
			})
			if !slices.Equal(requested, tc.wantRequested) {
				t.Fatalf(
					"Server.Requests() endpoints mismatch\ngot:  %q\nwant: %q",
					requested, tc.wantRequested,
				)
			}
		})
	}
}

func TestLookup_Query__noSideEffects(t *testing.T) {
	tests := []struct {
		name               string
		startLatestVersion string
	}{
		{name: "first version", startLatestVersion: ""},
		{name: "newer version", startLatestVersion: "0.0.1"},
		{name: "same version", startLatestVersion: "0.17.4"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newForgeServer(t, map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			})
			lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")

			lvbasetest.AssertQueryNoSideEffects(t, lookup, tc.startLatestVersion, "0.17.4")
		})
	}
}

func TestEndpointYieldedNothing(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nothing/an empty list",
			err:  errNoneFound,
			want: true,
		},
		{
			name: "nothing/a 404",
			err:  errNotFound,
			want: true,
		},
		{
			name: "nothing/a wrapped empty list",
			err:  fmt.Errorf("page 1: %w", errNoneFound),
			want: true,
		},
		{
			name: "something/no error",
			err:  nil,
			want: false,
		},
		{
			name: "something/any other error",
			err:  errors.New("connection refused"),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := endpointYieldedNothing(tc.err)

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestNothingFound(t *testing.T) {
	tests := []struct {
		name                 string
		releasesErr, tagsErr error
		triedTags            bool
		want                 string
	}{
		{
			name:        "releases empty, tags not tried",
			releasesErr: errNoneFound,
			want:        "no releases were found",
		},
		{
			name:        "releases 404, tags not tried",
			releasesErr: errNotFound,
			want:        "repository not found, or its releases are disabled",
		},
		{
			name:        "both empty",
			releasesErr: errNoneFound,
			tagsErr:     errNoneFound,
			triedTags:   true,
			want:        "no releases or tags were found",
		},
		{
			name:        "both 404",
			releasesErr: errNotFound,
			tagsErr:     errNotFound,
			triedTags:   true,
			want:        "repository not found, or its releases and tags are disabled",
		},
		{
			name:        "releases 404, tags empty",
			releasesErr: errNotFound,
			tagsErr:     errNoneFound,
			triedTags:   true,
			want:        "no tags were found, and the repository's releases are disabled",
		},
		{
			name:        "releases empty, tags 404",
			releasesErr: errNoneFound,
			tagsErr:     errNotFound,
			triedTags:   true,
			want:        "no releases were found, and the repository's tags are disabled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := nothingFound(tc.releasesErr, tc.tagsErr, tc.triedTags)

			if got == nil || got.Error() != tc.want {
				t.Fatalf("mismatch\ngot:  %v\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestLookup_QueryEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		endpoints      map[string]forgeEndpoint
		latestVersion  string // Already known to the service.
		endpoint       string
		wantNewVersion bool
		wantLatest     string
		errRegex       string
	}{
		{
			name: "version on the first page",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			endpoint:   endpointReleases,
			wantLatest: "0.17.4",
			errRegex:   `^$`,
		},
		{
			name: "newer version on the first page",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			latestVersion:  "0.0.1",
			endpoint:       endpointReleases,
			wantNewVersion: true,
			wantLatest:     "0.17.4",
			errRegex:       `^$`,
		},
		{
			name: "version only on a later page",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Pages: []string{
					`[{"tag_name":"not-a-version","prerelease":false}]`,
					`[{"tag_name":"nor-this","prerelease":false}]`,
					`[{"tag_name":"3.2.1","prerelease":false,"published_at":"2026-01-01T00:00:00Z"}]`,
				}},
			},
			endpoint:   endpointReleases,
			wantLatest: "3.2.1",
			errRegex:   `^$`,
		},
		{
			name: "tags endpoint is walked",
			endpoints: map[string]forgeEndpoint{
				endpointTags: {Body: tagsBody},
			},
			endpoint:   endpointTags,
			wantLatest: "0.12.1",
			errRegex:   `^$`,
		},
		{
			name: "an empty list stops the walk with nothing to filter",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListCompact},
			},
			endpoint: endpointReleases,
			errRegex: `^empty list$`,
		},
		{
			name: "pages that never yield a version end in an error",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Pages: []string{
					`[{"tag_name":"not-a-version","prerelease":false}]`,
					`[{"tag_name":"nor-this","prerelease":false}]`,
				}},
			},
			endpoint: endpointReleases,
			errRegex: `^no releases were found matching the url_commands on page 2 of the API response$`,
		},
		{
			name: "next page that does not advance is refused",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					Body: `[{"tag_name":"not-a-version","prerelease":false}]`,
					Headers: map[string]string{
						"Link": `<https://root-url.example.com/api/v1/repos/o/r/releases?page=1>; rel="next"`,
					},
				},
			},
			endpoint: endpointReleases,
			errRegex: `^/releases did not advance past page 1$`,
		},
		{
			name: "endless walk stops at the page cap",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					Body:    `[{"tag_name":"not-a-version","prerelease":false}]`,
					Endless: true,
				},
			},
			endpoint: endpointReleases,
			errRegex: fmt.Sprintf(`^gave up walking /releases after %d pages$`, maxPages),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newForgeServer(t, tc.endpoints)
			lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")
			if tc.latestVersion != "" {
				lookup.Status.SetLatestVersion(tc.latestVersion, "", false)
			}

			newVersion, err := lookup.queryEndpoint(tc.endpoint, true, logx.LogFrom{Primary: t.Name()})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if newVersion != tc.wantNewVersion {
				t.Fatalf("newVersion mismatch\ngot:  %t\nwant: %t", newVersion, tc.wantNewVersion)
			}

			if got := lookup.Status.LatestVersion(); got != tc.wantLatest {
				t.Fatalf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, tc.wantLatest)
			}
		})
	}
}

func TestLookup_QueryPage(t *testing.T) {
	tests := []struct {
		name           string
		endpoints      map[string]forgeEndpoint
		page           int
		latestVersion  string
		wantNewVersion bool
		wantNextPage   int
		wantLatest     string
		errRegex       string
	}{
		{
			name: "first version is recorded, but is not announced as new",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			page:       1,
			wantLatest: "0.17.4",
			errRegex:   `^$`,
		},
		{
			name: "version replacing an older one is new",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			page:           1,
			latestVersion:  "0.1.0",
			wantNewVersion: true,
			wantLatest:     "0.17.4",
			errRegex:       `^$`,
		},
		{
			name: "unchanged version is neither new nor an error",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			page:          1,
			latestVersion: "0.17.4",
			wantLatest:    "0.17.4",
			errRegex:      `^$`,
		},
		{
			name: "no version on this page, but another page to try",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Pages: []string{
					`[{"tag_name":"not-a-version","prerelease":false}]`,
					releasesBody,
				}},
			},
			page:         1,
			wantNextPage: 2,
			errRegex:     `^$`,
		},
		{
			name: "no version and no further page is an error",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: `[{"tag_name":"not-a-version","prerelease":false}]`},
			},
			page:     1,
			errRegex: `^no releases were found matching the url_commands on page 1 of the API response$`,
		},
		{
			name:      "404 stops the walk",
			endpoints: map[string]forgeEndpoint{},
			page:      1,
			errRegex:  `^not found$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newForgeServer(t, tc.endpoints)
			lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")
			if tc.latestVersion != "" {
				lookup.Status.SetLatestVersion(tc.latestVersion, "", false)
			}

			// The second check, so it commits rather than re-checking.
			newVersion, nextPage, err := lookup.queryPage(
				1, tc.page, endpointReleases, true,
				logx.LogFrom{Primary: t.Name()},
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if newVersion != tc.wantNewVersion {
				t.Fatalf("newVersion mismatch\ngot:  %t\nwant: %t", newVersion, tc.wantNewVersion)
			}
			if nextPage != tc.wantNextPage {
				t.Fatalf("nextPage mismatch\ngot:  %d\nwant: %d", nextPage, tc.wantNextPage)
			}
			if got := lookup.Status.LatestVersion(); got != tc.wantLatest {
				t.Fatalf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, tc.wantLatest)
			}

			requests := server.Requests()
			if len(requests) == 0 {
				t.Fatal("Server.Requests() length mismatch\ngot:  0\nwant: > 0")
			}
			wantPath := "/api/v1/repos/owner/repo/" + endpointReleases
			if got := requests[0].Path; got != wantPath {
				t.Errorf("RecordedRequest.Path mismatch\ngot:  %q\nwant: %q", got, wantPath)
			}
		})
	}
}

func TestLookup_HTTPRequest(t *testing.T) {
	tests := []struct {
		name         string
		endpoints    map[string]forgeEndpoint
		host         string // Fixture override.
		endpoint     string
		page         int
		wantBody     string
		wantNextPage int
		errRegex     string
	}{
		{
			name: "valid/a page of releases",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			},
			endpoint: endpointReleases,
			page:     1,
			wantBody: releasesBody,
			errRegex: `^$`,
		},
		{
			name: "valid/a paginated page advertises the next",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Pages: []string{releasesBody, tagsBody}},
			},
			endpoint:     endpointReleases,
			page:         1,
			wantBody:     releasesBody,
			wantNextPage: 2,
			errRegex:     `^$`,
		},
		{
			name: "invalid/an empty first page",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {Body: emptyListNewline},
			},
			endpoint: endpointReleases,
			page:     1,
			errRegex: `^empty list$`,
		},
		{
			name:      "invalid/an endpoint the instance does not serve",
			endpoints: map[string]forgeEndpoint{},
			endpoint:  endpointReleases,
			page:      1,
			errRegex:  `^not found$`,
		},
		{
			name:      "invalid/a host no request can be built for",
			endpoints: map[string]forgeEndpoint{},
			host:      "https://codeberg.org/%zz",
			endpoint:  endpointReleases,
			page:      1,
			errRegex: test.TrimYAML(`
				^invalid host "[^"]+":
					parse "[^"]+":
						invalid URL escape "%zz"$`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newForgeServer(t, tc.endpoints)
			host := util.ValueOr(tc.host, server.URL)
			lookup := testLookup(t, "host: "+host+"\nurl: owner/repo")

			body, nextPage, err := lookup.httpRequest(
				tc.endpoint, tc.page,
				logx.LogFrom{Primary: t.Name()},
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if string(body) != tc.wantBody {
				t.Fatalf("body mismatch\ngot:  %q\nwant: %q", string(body), tc.wantBody)
			}
			if nextPage != tc.wantNextPage {
				t.Fatalf("nextPage mismatch\ngot:  %d\nwant: %d", nextPage, tc.wantNextPage)
			}
		})
	}
}

func TestLookup_CreateRequest(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		serviceToken string
		defaults     map[string]HostDefaults
		endpoint     string
		page         int
		wantURL      string
		wantAuth     string
		errRegex     string
	}{
		{
			name:     "releases, first page",
			host:     "https://codeberg.org",
			endpoint: endpointReleases,
			page:     1,
			wantURL:  "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "tags, second page",
			host:     "https://codeberg.org",
			endpoint: endpointTags,
			page:     2,
			wantURL:  "https://codeberg.org/api/v1/repos/owner/repo/tags?limit=50&page=2",
			errRegex: `^$`,
		},
		{
			name:     "unvalidated host still builds an address",
			host:     "", // CheckValues validates the host.
			endpoint: endpointReleases,
			page:     1,
			wantURL:  "https:///api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:         "service's own token is sent",
			host:         "https://codeberg.org",
			serviceToken: "service-token",
			endpoint:     endpointReleases,
			page:         1,
			wantURL:      "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			wantAuth:     "Bearer service-token",
			errRegex:     `^$`,
		},
		{
			name: "token from the matching host entry is sent",
			host: "https://codeberg.org",
			defaults: map[string]HostDefaults{
				"https://codeberg.org": {URL: "https://codeberg.org", AccessToken: "defaults-token"},
			},
			endpoint: endpointReleases,
			page:     1,
			wantURL:  "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			wantAuth: "Bearer defaults-token",
			errRegex: `^$`,
		},
		{
			name: "host matching no entry is sent no credential",
			host: "https://codeberg.org",
			defaults: map[string]HostDefaults{
				"https://forge.example.com": {URL: "https://forge.example.com", AccessToken: "defaults-token"},
			},
			endpoint: endpointReleases,
			page:     1,
			wantURL:  "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "a host no address can be built from",
			host:     "https://codeberg.org/%zz", // CheckValues rejects it first.
			endpoint: endpointReleases,
			page:     1,
			errRegex: test.TrimYAML(`
				^invalid host "[^"]+":
					parse "[^"]+":
						invalid URL escape "%zz"$`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := Lookup{
				Host:        tc.host,
				AccessToken: tc.serviceToken,
			}
			lookup.URL = "owner/repo"
			lookup.SetTypeDefaults(
				&Defaults{Host: tc.defaults},
				&Defaults{},
			)

			request, err := lookup.createRequest(
				tc.endpoint,
				tc.page,
				logx.LogFrom{Primary: t.Name()},
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if request.Method != http.MethodGet {
				t.Fatalf(
					"Request.Method mismatch\ngot:  %q\nwant: %q",
					request.Method, http.MethodGet,
				)
			}
			if got := request.URL.String(); got != tc.wantURL {
				t.Fatalf("Request.URL.String() mismatch\ngot:  %q\nwant: %q", got, tc.wantURL)
			}

			if got := request.Header.Get("Authorization"); got != tc.wantAuth {
				t.Fatalf(
					"Request.Header[Authorization] mismatch\ngot:  %q\nwant: %q",
					got, tc.wantAuth,
				)
			}

			for _, header := range []string{"If-None-Match", "If-Modified-Since"} {
				if got := request.Header.Get(header); got != "" {
					t.Fatalf("Request.Header[%s] mismatch\ngot:  %q\nwant: \"\"", header, got)
				}
			}
		})
	}
}

func TestLookup_RequestFor(t *testing.T) {
	tests := []struct {
		name         string
		address      string
		serviceToken string
		wantAuth     string
		errRegex     string
	}{
		{
			name:     "valid address",
			address:  "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:         "the access token is sent",
			address:      "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			serviceToken: "service-token",
			wantAuth:     "Bearer service-token",
			errRegex:     `^$`,
		},
		{
			name:    "an address no request can be built for",
			address: "invalid://\ttest",
			errRegex: test.TrimYAML(`
				^failed creating http request for .*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := Lookup{AccessToken: tc.serviceToken}
			lookup.SetTypeDefaults(&Defaults{}, &Defaults{})

			request, err := lookup.requestFor(
				tc.address,
				logx.LogFrom{Primary: t.Name()},
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if request.Method != http.MethodGet {
				t.Fatalf(
					"Request.Method mismatch\ngot:  %q\nwant: %q",
					request.Method, http.MethodGet,
				)
			}
			if got := request.URL.String(); got != tc.address {
				t.Fatalf("Request.URL.String() mismatch\ngot:  %q\nwant: %q", got, tc.address)
			}
			if got := request.Header.Get("Authorization"); got != tc.wantAuth {
				t.Fatalf(
					"Request.Header[Authorization] mismatch\ngot:  %q\nwant: %q",
					got, tc.wantAuth,
				)
			}
		})
	}
}

func TestLookup_GetResponse(t *testing.T) {
	t.Run("valid/the response and its body come back from /releases", func(t *testing.T) {
		t.Parallel()

		server := newForgeServer(t, map[string]forgeEndpoint{
			endpointReleases: {Body: releasesBody},
		})
		lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")
		logFrom := logx.LogFrom{Primary: t.Name()}

		req, err := lookup.createRequest(endpointReleases, 1, logFrom)
		if err != nil {
			t.Fatalf("Lookup.createRequest() failed: %v", err)
		}

		response, body, err := lookup.getResponse(req, logFrom)
		if err != nil {
			t.Fatalf("Lookup.getResponse() failed: %v", err)
		}

		if got, want := response.StatusCode, http.StatusOK; got != want {
			t.Fatalf("status code mismatch\ngot:  %d - %q\nwant: %d", got, string(body), want)
		}
		if string(body) != releasesBody {
			t.Fatalf("body mismatch\ngot:  %q\nwant: %q", string(body), releasesBody)
		}
	})

	t.Run("invalid/an instance that is not listening", func(t *testing.T) {
		t.Parallel()

		server := newForgeServer(t, map[string]forgeEndpoint{})
		host := server.URL
		server.Close()

		lookup := testLookup(t, "host: "+host+"\nurl: owner/repo")
		logFrom := logx.LogFrom{Primary: t.Name()}

		req, err := lookup.createRequest(endpointReleases, 1, logFrom)
		if err != nil {
			t.Fatalf("Lookup.createRequest() failed: %v", err)
		}

		response, body, err := lookup.getResponse(req, logFrom)
		e := errfmt.FormatError(err)
		if wantRe := `connection refused|connect: `; !util.RegexCheck(wantRe, e) {
			t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, wantRe)
		}
		if response != nil || body != nil {
			t.Fatalf(
				"response, body mismatch\n"+
					"got:  response nil=%t, body nil=%t\nwant: response nil=true, body nil=true",
				response == nil, body == nil,
			)
		}
	})
}

func TestLookup_GetResponse__readError(t *testing.T) {
	// Promises a body, then closes the connection without sending it.
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "10")
			w.WriteHeader(http.StatusOK)
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

	lookup := Lookup{}
	response, body, err := lookup.getResponse(req, logx.LogFrom{Primary: t.Name()})

	e := errfmt.FormatError(err)
	if wantRe := `unexpected EOF|connection reset|body length`; !util.RegexCheck(wantRe, e) {
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

func TestHandleResponse(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		headers      map[string]string
		body         string
		accessToken  string
		page         int
		wantBody     string
		wantNextPage int
		errRegex     string
	}{
		{
			name:     "200/returns the body",
			status:   http.StatusOK,
			body:     releasesBody,
			page:     1,
			wantBody: releasesBody,
			errRegex: `^$`,
		},
		{
			name:     "200/empty first page is nothing to filter",
			status:   http.StatusOK,
			body:     emptyListNewline,
			page:     1,
			errRegex: `^empty list$`,
		},
		{
			name:     "200/page 0",
			status:   http.StatusOK,
			body:     emptyListCompact,
			page:     0,
			errRegex: `^empty list$`,
		},
		{
			name:     "200/empty later page is returned",
			status:   http.StatusOK,
			body:     emptyListNewline,
			page:     2,
			wantBody: emptyListNewline,
			errRegex: `^$`,
		},
		{
			name:   "200/takes the next page from the Link header",
			status: http.StatusOK,
			headers: map[string]string{
				"Link": `<https://codeberg.org/api/v1/repos/o/r/releases?limit=50&page=2>; rel="next"`,
			},
			body:         releasesBody,
			page:         1,
			wantBody:     releasesBody,
			wantNextPage: 2,
			errRegex:     `^$`,
		},
		{
			name:     "404/ambiguous not-found",
			status:   http.StatusNotFound,
			body:     `{"message":"The target couldn't be found."}`,
			page:     1,
			errRegex: `^not found$`,
		},
		{
			name:        "401/authentication failure",
			status:      http.StatusUnauthorized,
			accessToken: "a-token",
			page:        1,
			errRegex:    `^authentication failed for codeberg.org \(401\)$`,
		},
		{
			name:     "401/unauthenticated query names the absent credential",
			status:   http.StatusUnauthorized,
			page:     1,
			errRegex: `^authentication failed for codeberg.org \(401\) - no access_token was sent$`,
		},
		{
			name:        "403/no limiter headers is an authentication failure",
			status:      http.StatusForbidden,
			accessToken: "a-token",
			page:        1,
			errRegex:    `^authentication failed for codeberg.org \(403\)$`,
		},
		{
			name:        "403/announcing a limiter is a rate limit",
			status:      http.StatusForbidden,
			headers:     map[string]string{"RateLimit": `"baseline";r=0;t=600`},
			accessToken: "a-token",
			page:        1,
			errRegex:    `^rate limit reached for codeberg.org$`,
		},
		{
			name:        "429/rate limit",
			status:      http.StatusTooManyRequests,
			body:        `<!DOCTYPE html>`,
			accessToken: "a-token",
			page:        1,
			errRegex:    `^rate limit reached for codeberg.org$`,
		},
		{
			name:        "429/retry window named",
			status:      http.StatusTooManyRequests,
			headers:     map[string]string{"Retry-After": "120"},
			body:        `<!DOCTYPE html>`,
			accessToken: "a-token",
			page:        1,
			errRegex:    `^rate limit reached for codeberg.org - retry after 120$`,
		},
		{
			name:     "429/unauthenticated query names the absent credential",
			status:   http.StatusTooManyRequests,
			body:     `<!DOCTYPE html>`,
			page:     1,
			errRegex: `^rate limit reached for codeberg.org - no access_token was sent$`,
		},
		{
			name:     "429/retry window and the absent credential both named",
			status:   http.StatusTooManyRequests,
			headers:  map[string]string{"Retry-After": "120"},
			body:     `<!DOCTYPE html>`,
			page:     1,
			errRegex: `^rate limit reached for codeberg.org - retry after 120 - no access_token was sent$`,
		},
		{
			name:     "unmapped/status reported with an excerpt of its body",
			status:   http.StatusInternalServerError,
			body:     "proxy error",
			page:     1,
			errRegex: `^unknown status code 500\nproxy error$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			response := newResponse(
				t,
				"https://codeberg.org/api/v1/repos/o/r/releases",
				tc.status,
				tc.headers,
				tc.body,
				tc.accessToken,
			)

			body, nextPage, err := handleResponse(
				response, []byte(tc.body), tc.page,
				logx.LogFrom{Primary: t.Name()},
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if string(body) != tc.wantBody {
				t.Fatalf("body mismatch\ngot:  %q\nwant: %q", string(body), tc.wantBody)
			}
			if nextPage != tc.wantNextPage {
				t.Fatalf("nextPage mismatch\ngot:  %d\nwant: %d", nextPage, tc.wantNextPage)
			}
		})
	}
}

func TestIsEmptyJSONList(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "empty/compact",
			body: emptyListCompact,
			want: true,
		},
		{
			name: "empty/LF",
			body: emptyListNewline,
			want: true,
		},
		{
			name: "empty/CRLF",
			body: "[]\r\n",
			want: true,
		},
		{
			name: "empty/surrounded by whitespace",
			body: "  []  ",
			want: true,
		},
		{
			name: "not empty/a release list",
			body: releasesBody,
			want: false,
		},
		{
			name: "not empty/a list holding one release",
			body: `[{"tag_name":"1.0.0"}]`,
			want: false,
		},
		{
			name: "not empty/an object",
			body: `{}`,
			want: false,
		},
		{
			name: "not empty/no body",
			body: "",
			want: false,
		},
		{
			name: "not empty/a non-empty list",
			body: `[1]`,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isEmptyJSONList([]byte(tc.body))

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestRequestsRemaining(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		want     int
		wantSaid bool
	}{
		{
			name:     "silent/no headers",
			want:     0,
			wantSaid: false,
		},
		{
			name:     "silent/only a policy describing the budget",
			headers:  map[string]string{"RateLimit-Policy": `"baseline";q=2000;w=600`},
			want:     0,
			wantSaid: false,
		},
		{
			name:     "silent/only 'Retry-After', which any outage may set",
			headers:  map[string]string{"Retry-After": "120"},
			want:     0,
			wantSaid: false,
		},
		{
			name:     "silent/remaining count that is not a number",
			headers:  map[string]string{"X-RateLimit-Remaining": "many"},
			want:     0,
			wantSaid: false,
		},
		{
			name:     "silent/IETF header with no 'r=' parameter",
			headers:  map[string]string{"RateLimit": `"baseline";t=600`},
			want:     0,
			wantSaid: false,
		},
		{
			name:     "said/GitHub's X-RateLimit-Remaining, spent",
			headers:  map[string]string{"X-RateLimit-Remaining": "0"},
			want:     0,
			wantSaid: true,
		},
		{
			name:     "said/GitHub's X-RateLimit-Remaining, with budget left",
			headers:  map[string]string{"X-RateLimit-Remaining": "1"},
			want:     1,
			wantSaid: true,
		},
		{
			name:     "said/bare RateLimit-Remaining spelling",
			headers:  map[string]string{"RateLimit-Remaining": "7"},
			want:     7,
			wantSaid: true,
		},
		{
			name:     "said/IETF RateLimit header, spent",
			headers:  map[string]string{"RateLimit": `"baseline";r=0;t=600`},
			want:     0,
			wantSaid: true,
		},
		{
			name:     "said/IETF RateLimit header, with budget left",
			headers:  map[string]string{"RateLimit": `"baseline";r=1960;t=600`},
			want:     1960,
			wantSaid: true,
		},
		{
			name: "said/a bare count wins over the IETF header",
			headers: map[string]string{
				"RateLimit":             `"baseline";r=1960;t=600`,
				"X-RateLimit-Remaining": "0",
			},
			want:     0,
			wantSaid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			header := http.Header{}
			for key, value := range tc.headers {
				header.Set(key, value)
			}

			got, gotSaid := requestsRemaining(header)

			if gotSaid != tc.wantSaid {
				t.Fatalf("said mismatch\ngot:  %t\nwant: %t", gotSaid, tc.wantSaid)
			}
			if got != tc.want {
				t.Fatalf("remaining count mismatch\ngot:  %d\nwant: %d", got, tc.want)
			}
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		headers map[string]string
		want    bool
	}{
		{
			name:   "limited/429 alone",
			status: http.StatusTooManyRequests,
			want:   true,
		},
		{
			name:    "limited/403 with the IETF RateLimit header",
			status:  http.StatusForbidden,
			headers: map[string]string{"RateLimit": `"baseline";r=0;t=600`},
			want:    true,
		},
		{
			name:    "not limited/403 with only a RateLimit-Policy header",
			status:  http.StatusForbidden,
			headers: map[string]string{"RateLimit-Policy": `"baseline";q=2000;w=600`},
			want:    false,
		},
		{
			name:    "not limited/403 with the IETF RateLimit header showing requests remaining",
			status:  http.StatusForbidden,
			headers: map[string]string{"RateLimit": `"baseline";r=1960;t=600`},
			want:    false,
		},
		{
			name:    "not limited/403 with GitHub's X-RateLimit-Remaining above zero",
			status:  http.StatusForbidden,
			headers: map[string]string{"X-RateLimit-Remaining": "59"},
			want:    false,
		},
		{
			name:    "limited/403 with RateLimit-Remaining at zero",
			status:  http.StatusForbidden,
			headers: map[string]string{"RateLimit-Remaining": "0"},
			want:    true,
		},
		{
			name:    "limited/403 with GitHub's X-RateLimit-Remaining at zero",
			status:  http.StatusForbidden,
			headers: map[string]string{"X-RateLimit-Remaining": "0"},
			want:    true,
		},
		{
			name:    "limited/IETF RateLimit header, spent",
			status:  http.StatusForbidden,
			headers: map[string]string{"RateLimit": `"baseline";r=0;t=600`},
			want:    true,
		},
		{
			name:   "not limited/403 with no rate-limit headers",
			status: http.StatusForbidden,
			want:   false,
		},
		{
			name:    "not limited/403 with only 'Retry-After' (possible maintenance window)",
			status:  http.StatusForbidden,
			headers: map[string]string{"Retry-After": "120"},
			want:    false,
		},
		{
			name:   "not limited/401",
			status: http.StatusUnauthorized,
			want:   false,
		},
		{
			name:    "not limited/200 carrying the limiter's running count",
			status:  http.StatusOK,
			headers: map[string]string{"RateLimit": `"baseline";r=1960;t=600`},
			want:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			response := newResponse(
				t,
				"https://codeberg.org",
				tc.status,
				tc.headers,
				"",
				"",
			)

			got := isRateLimited(response)

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestLookup_GetVersion(t *testing.T) {
	tests := []struct {
		name            string
		lookupYAML      string
		body            string
		page            int
		wantVersion     string
		wantReleaseDate string
		errRegex        string
	}{
		{
			name: "the highest semantic version, with its release date",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo`,
			body:            releasesBody,
			page:            1,
			wantVersion:     "0.17.4",
			wantReleaseDate: "2000-01-02T00:00:00Z",
			errRegex:        `^$`,
		},
		{
			name: "require narrows which release is taken",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo
				require:
					regex_version: '^0\.16\.'`,
			body:            releasesBody,
			page:            1,
			wantVersion:     "0.16.9",
			wantReleaseDate: "2000-01-01T00:00:00Z",
			errRegex:        `^$`,
		},
		{
			name: "tags/tag has no published_at",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo`,
			body: tagsBody,
			page: 1,

			wantVersion: "0.12.1",
			errRegex:    `^$`,
		},
		{
			name: "tags/semver-prerelease tag name is excluded",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo`,
			body: preReleaseTagsBody,
			page: 1,

			wantVersion: "0.12.1",
			errRegex:    `^$`,
		},
		{
			name: "tags/semver-prerelease tag name is kept when opted into",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo
				use_prerelease: true`,
			body: preReleaseTagsBody,
			page: 1,

			wantVersion: "0.13.0-rc1",
			errRegex:    `^$`,
		},
		{
			name: "nothing matches the url_commands",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo`,
			body:     `[{"tag_name":"not-a-version","prerelease":false}]`,
			page:     3,
			errRegex: `^no releases were found matching the url_commands on page 3 of the API response$`,
		},
		{
			name: "nothing meets the require fields",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo
				require:
					regex_version: '^9\.'`,
			body:     releasesBody,
			page:     1,
			errRegex: `^no releases were found matching the require fields`,
		},
		{
			name: "a body that is not a release list",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo`,
			body:     `<!DOCTYPE html>`,
			page:     1,
			errRegex: `^release data failed to parse`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, tc.lookupYAML)

			version, releaseDate, err := lookup.getVersion(
				[]byte(tc.body), tc.page,
				logx.LogFrom{Primary: t.Name()},
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if version != tc.wantVersion {
				t.Fatalf("version mismatch\ngot:  %q\nwant: %q", version, tc.wantVersion)
			}
			if releaseDate != tc.wantReleaseDate {
				t.Fatalf(
					"releaseDate mismatch\ngot:  %q\nwant: %q",
					releaseDate, tc.wantReleaseDate,
				)
			}
		})
	}
}

func TestLookup_FilterReleases(t *testing.T) {
	tests := []struct {
		name       string
		lookupYAML string
		body       string
		want       []string
	}{
		{
			name: "prereleases excluded, and the highest semantic version sorts first",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo`,
			body: releasesBody,
			want: []string{"0.17.4", "0.16.9"},
		},
		{
			name: "prereleases kept when opted into",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo
				use_prerelease: true`,
			body: releasesBody,
			want: []string{"0.18.0-rc1", "0.17.4", "0.16.9"},
		},
		{
			name: "url_commands applied before the version is parsed",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo
				url_commands:
					- type: regex
						regex: '([0-9.]+)$'`,
			body: test.TrimJSON(`[
				{"tag_name":"release-1.2.3","prerelease":false},
				{"tag_name":"nothing-here","prerelease":false}
			]`),
			want: []string{"1.2.3"},
		},
		{
			name: "tags endpoint",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo`,
			body: tagsBody,
			want: []string{"0.12.1", "0.12.0"},
		},
		{
			name: "no releases",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo`,
			body: emptyListNewline,
			want: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, tc.lookupYAML)
			releases, err := forge.UnmarshalReleases([]byte(tc.body))
			if err != nil {
				t.Fatalf("forge.UnmarshalReleases() failed: %v", err)
			}

			got := lookup.filterReleases(releases, logx.LogFrom{Primary: t.Name()})

			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a forgetypes.Release, b string) bool { return a.TagName == b },
				"",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestLookup_HandleNewVersion(t *testing.T) {
	for _, tc := range []struct {
		name      string
		writeToDB bool
		wantLog   bool
	}{
		{
			name:      "the first check logs that it checks again",
			writeToDB: true,
			wantLog:   true,
		},
		{
			name: "no DB write, so no log",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.
			releaseLog := test.CaptureLog(t, logx.Default())

			server := newForgeServer(t, map[string]forgeEndpoint{
				endpointReleases: {Body: releasesBody},
			})
			lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")
			lookup.Status.SetLatestVersion("0.0.1", "", false)

			if _, err := lookup.handleNewVersion(
				0,
				"0.17.4", "2026-08-01T10:50:00Z", "0.0.1",
				endpointReleases, 1,
				tc.writeToDB,
				logx.LogFrom{Primary: t.Name()},
			); err != nil {
				t.Fatalf("Lookup.handleNewVersion() failed: %v", err)
			}

			stdout := releaseLog()
			if got := strings.Contains(stdout, `Possibly found a new version (From "0.0.1" to "0.17.4")`); got != tc.wantLog {
				t.Errorf("stdout 'Possibly found' log mismatch\ngot:  %t (%q)\nwant: %t", got, stdout, tc.wantLog)
			}
		})
	}

	t.Run("the first check re-queries before committing", func(t *testing.T) {
		t.Parallel()

		server := newForgeServer(t, map[string]forgeEndpoint{
			endpointReleases: {Body: releasesBody},
		})
		lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")

		version := "0.17.4"
		_, err := lookup.handleNewVersion(
			0,
			version, "2026-08-01T10:50:00Z", "",
			endpointReleases, 1,
			true,
			logx.LogFrom{Primary: t.Name()},
		)
		if err != nil {
			t.Fatalf("Lookup.handleNewVersion() failed: %v", err)
		}

		if got, want := len(server.Requests()), 1; got != want {
			t.Fatalf("Server.Requests() length mismatch\ngot:  %d\nwant: %d", got, want)
		}
		if got, want := lookup.Status.LatestVersion(), version; got != want {
			t.Fatalf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("the first version found is recorded, but not announced as new", func(t *testing.T) {
		t.Parallel()

		lookup := testLookup(t, test.TrimYAML(`
			host: https://codeberg.org
			url: owner/repo`))

		version := "1.2.3"
		// The second check, so it commits rather than re-checking.
		newVersion, err := lookup.handleNewVersion(
			1,
			version, "2026-01-01T00:00:00Z", "",
			endpointReleases, 1,
			true,
			logx.LogFrom{Primary: t.Name()},
		)
		if err != nil {
			t.Fatalf("Lookup.handleNewVersion() failed: %v", err)
		}

		if newVersion {
			t.Fatal("newVersion mismatch\ngot:  true\nwant: false")
		}
		if got, want := lookup.Status.LatestVersion(), version; got != want {
			t.Fatalf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, want)
		}
	})

	t.Run("a version replacing an older one is announced as new", func(t *testing.T) {
		t.Parallel()

		lookup := testLookup(t, "host: https://codeberg.org\nurl: owner/repo")
		lookup.Status.SetLatestVersion("1.0.0", "", false)

		version := "1.2.3"
		// The second check, so it commits rather than re-checking.
		newVersion, err := lookup.handleNewVersion(
			1,
			version, "2026-01-01T00:00:00Z", "1.0.0",
			endpointReleases, 1,
			true,
			logx.LogFrom{Primary: t.Name()},
		)
		if err != nil {
			t.Fatalf("Lookup.handleNewVersion() failed: %v", err)
		}

		if !newVersion {
			t.Fatal("newVersion mismatch\ngot:  false\nwant: true")
		}
		if got, want := lookup.Status.LatestVersion(), version; got != want {
			t.Fatalf("Status.LatestVersion() mismatch\ngot:  %q\nwant: %q", got, want)
		}
	})
}
