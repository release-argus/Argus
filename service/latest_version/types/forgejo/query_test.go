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
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/forge"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_Query(t *testing.T) {
	// GIVEN: an instance responding in a particular way, and a service configured against it.
	tests := []struct {
		name        string
		endpoints   map[string]forgeEndpoint
		lookupYAML  string
		wantVersion string
		errRegex    string
	}{
		{
			name: "valid/the highest non-prerelease version is reported",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: releasesBody},
			},
			lookupYAML:  `url: Release-Argus/Argus`,
			wantVersion: "0.17.4",
			errRegex:    `^$`,
		},
		{
			name: "valid/the highest semantic version wins, not the most recently published",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					body: test.TrimJSON(`[
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
				endpointReleases: {body: emptyListNewline},
				endpointTags:     {body: tagsBody},
			},
			lookupYAML:  `url: owner/repo`,
			wantVersion: "0.12.1",
			errRegex:    `^$`,
		},
		{
			name: "valid/tags fallback, on the compact empty-list encoding",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: emptyListCompact},
				endpointTags:     {body: tagsBody},
			},
			lookupYAML:  `url: owner/repo`,
			wantVersion: "0.12.1",
			errRegex:    `^$`,
		},
		{
			name: "valid/releases disabled (404), tags fallback",
			endpoints: map[string]forgeEndpoint{
				endpointTags: {body: tagsBody},
			},
			lookupYAML:  `url: owner/repo`,
			wantVersion: "0.12.1",
			errRegex:    `^$`,
		},
		{
			name: "valid/url_commands filter the tag",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: test.TrimJSON(`[
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
				endpointReleases: {body: test.TrimJSON(`[
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
				endpointReleases: {body: emptyListNewline},
				endpointTags:     {body: emptyListNewline},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^no releases or tags were found$`,
		},
		{
			name: "invalid/releases empty, tags 404",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: emptyListNewline},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^no releases were found, and the repository's tags are disabled$`,
		},
		{
			name: "invalid/releases 404, tags empty",
			endpoints: map[string]forgeEndpoint{
				endpointTags: {body: emptyListCompact},
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
			name:      "invalid/releases 404, no tags fallback (use_prereleases)",
			endpoints: map[string]forgeEndpoint{},
			lookupYAML: `
				url: owner/repo
				use_prerelease: true`,
			errRegex: `^repository not found, or its releases are disabled$`,
		},
		{
			name: "invalid/401",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					status: http.StatusUnauthorized,
					body:   `{"message":"token is required"}`,
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^authentication failed for 127\.0\.0\.1:\d+ \(401\)$`,
		},
		{
			name: "invalid/403 without rate-limit headers",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					status: http.StatusForbidden,
					body:   `{"message":"forbidden"}`,
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^authentication failed for 127\.0\.0\.1:\d+ \(403\)$`,
		},
		{
			name: "invalid/403 with rate-limit headers",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					status:  http.StatusForbidden,
					headers: map[string]string{"RateLimit": `"baseline";r=0;t=600`},
					body:    `{"message":"forbidden"}`,
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^rate limit reached for 127\.0\.0\.1:\d+$`,
		},
		{
			name: "invalid/429",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					status: http.StatusTooManyRequests,
					body:   `<!DOCTYPE html><html class="codeberg-design"><body>Too many requests</body></html>`,
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^rate limit reached for 127\.0\.0\.1:\d+$`,
		},
		{
			name: "invalid/429 with a retry window",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					status:  http.StatusTooManyRequests,
					headers: map[string]string{"Retry-After": "120"},
					body:    `<!DOCTYPE html><html><body>Too many requests</body></html>`,
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^rate limit reached for 127\.0\.0\.1:\d+ - retry after 120$`,
		},
		{
			name: "invalid/an unmapped status code is reported with its body",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					status: http.StatusInternalServerError,
					body:   "proxy error",
				},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^unknown status code 500\nproxy error$`,
		},
		{
			name: "invalid/a body that is not a release list",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: `<!DOCTYPE html><html><body>not the API</body></html>`},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `release data failed to parse`,
		},
		{
			name: "invalid/no release matches the url_commands",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: test.TrimJSON(`[
					{"tag_name":"not-a-version","prerelease":false,"published_at":"2026-01-01T00:00:00Z"}
				]`)},
			},
			lookupYAML: `url: owner/repo`,
			errRegex:   `^no releases were found matching the url_commands on page 1 of the API response$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newForgeServer(t, tc.endpoints)
			lookupYAML := "host: " + server.URL + "\n" + test.TrimYAML(tc.lookupYAML)
			lookup := testLookup(t, lookupYAML)

			// WHEN: it is queried.
			_, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()})

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
				t.Fatalf(
					"%s version mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantVersion,
				)
			}
		})
	}
}

func TestLookup_Query__Paginates(t *testing.T) {
	// GIVEN: an instance whose first page of the endpoint queried holds no usable version.
	tests := []struct {
		name        string
		endpoints   map[string]forgeEndpoint
		wantVersion string
		wantOn      string
	}{
		{
			name: "releases",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {pages: []string{
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
				endpointReleases: {body: emptyListNewline},
				endpointTags: {pages: []string{
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

			// WHEN: it is queried.
			if _, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()}); err != nil {
				t.Fatalf(
					"%s\nLookup.Query() unexpected error: %v",
					packageName, err,
				)
			}

			prefix := fmt.Sprintf("%s\nLookup.Query()", packageName)

			// THEN: the version from the second page is reported.
			if got := lookup.Status.LatestVersion(); got != tc.wantVersion {
				t.Fatalf(
					"%s version mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantVersion,
				)
			}

			// AND: page 2 was asked for by number, from the endpoint the version was on.
			requests := server.requests()
			if !slices.ContainsFunc(requests, func(r recordedRequest) bool {
				return r.query.Get("page") == "2" && strings.HasSuffix(r.path, "/"+tc.wantOn)
			}) {
				t.Fatalf(
					"%s never requested page 2 of /%s\ngot:  %+v",
					prefix, tc.wantOn, requests,
				)
			}

			// AND: the page size used the forge's own parameter name, at its cap.
			for _, request := range requests {
				if got, want := request.query.Get("limit"), fmt.Sprint(apiPageSize); got != want {
					t.Fatalf(
						"%s limit mismatch on %s\ngot:  %q\nwant: %q",
						prefix, request.path, got, want,
					)
				}
			}
		})
	}
}

func TestLookup_Query__CachesNothingBetweenQueries(t *testing.T) {
	// GIVEN: an instance serving releases.
	server := newForgeServer(t, map[string]forgeEndpoint{
		endpointReleases: {body: releasesBody},
	})
	lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")
	logFrom := logx.LogFrom{Primary: t.Name()}

	// WHEN: it is queried twice.
	if _, err := lookup.Query(false, logFrom); err != nil {
		t.Fatalf(
			"%s\nLookup.Query() unexpected error on the first query: %v",
			packageName, err,
		)
	}
	afterFirst := len(server.requests())

	if _, err := lookup.Query(false, logFrom); err != nil {
		t.Fatalf(
			"%s\nLookup.Query() unexpected error on the second query: %v",
			packageName, err,
		)
	}

	// THEN: the second query went to the instance rather than reusing a cached response.
	if got := len(server.requests()); got <= afterFirst {
		t.Fatalf(
			"%s\nLookup.Query() reused a cached response\ngot:  %d requests\nwant: more than %d",
			packageName, got, afterFirst,
		)
	}
}

func TestLookup_Query__SubPathInstance(t *testing.T) {
	// GIVEN: an instance served under a sub-path.
	server := newForgeServer(t, map[string]forgeEndpoint{
		endpointReleases: {body: releasesBody},
	})
	lookup := testLookup(t, "host: "+server.URL+"/git\nurl: owner/repo")

	// WHEN: it is queried.
	if _, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()}); err != nil {
		t.Fatalf(
			"%s\nLookup.Query() unexpected error: %v",
			packageName, err,
		)
	}

	// THEN: the version is reported.
	if got, want := lookup.Status.LatestVersion(), "0.17.4"; got != want {
		t.Fatalf(
			"%s\nLookup.Query() version mismatch\ngot:  %q\nwant: %q",
			packageName, got, want,
		)
	}

	// AND: the sub-path prefixed the API path.
	const want = "/git/api/v1/repos/owner/repo/" + endpointReleases
	for _, request := range server.requests() {
		if request.path != want {
			t.Fatalf(
				"%s\nLookup.Query() path mismatch\ngot:  %q\nwant: %q",
				packageName, request.path, want,
			)
		}
	}
}

func TestLookup_Query__NoTagsFallback(t *testing.T) {
	// GIVEN: an instance with no releases, and a service filtering on something tags cannot carry.
	tests := []struct {
		name       string
		lookupYAML string
	}{
		{
			name: "use_prerelease/tags have no prerelease labelling",
			lookupYAML: `
				url: owner/repo
				use_prerelease: true`,
		},
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
				endpointReleases: {body: emptyListNewline},
				endpointTags:     {body: tagsBody},
			})
			lookupYAML := "host: " + server.URL + "\n" + test.TrimYAML(tc.lookupYAML)
			lookup := testLookup(t, lookupYAML)

			// WHEN: it is queried.
			_, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()})

			prefix := fmt.Sprintf("%s\nLookup.Query()", packageName)

			// THEN: it reports that no releases were found, without mentioning tags.
			e := errfmt.FormatError(err)
			if wantRe := `^no releases were found$`; !util.RegexCheck(wantRe, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, wantRe,
				)
			}

			// AND: /tags was never requested.
			for _, request := range server.requests() {
				if strings.HasSuffix(request.path, "/"+endpointTags) {
					t.Fatalf(
						"%s requested /%s, which cannot satisfy the filters\ngot:  %q",
						prefix, endpointTags, request.path,
					)
				}
			}
		})
	}
}

func TestIsEmptyJSONList(t *testing.T) {
	// GIVEN: a response body.
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

			// WHEN: isEmptyJSONList is called on it.
			got := isEmptyJSONList([]byte(tc.body))

			// THEN: emptiness is judged on content, not on length.
			if got != tc.want {
				t.Fatalf(
					"%s\nisEmptyJSONList(%q) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.body,
					got, tc.want,
				)
			}
		})
	}
}

func TestGetNextPage(t *testing.T) {
	// GIVEN: a Link header.
	tests := []struct {
		name string
		link string
		want int
	}{
		{
			name: "next and last",
			link: `<https://codeberg.org/api/v1/repos/o/r/releases?limit=50&page=2>; rel="next",` +
				`<https://codeberg.org/api/v1/repos/o/r/releases?limit=50&page=4>; rel="last"`,
			want: 2,
		},
		{
			name: "prev, next, last and first",
			link: `<https://codeberg.org/api/v1/repos/o/r/releases?page=2>; rel="prev",` +
				`<https://codeberg.org/api/v1/repos/o/r/releases?page=4>; rel="next",` +
				`<https://codeberg.org/api/v1/repos/o/r/releases?page=9>; rel="last",` +
				`<https://codeberg.org/api/v1/repos/o/r/releases?page=1>; rel="first"`,
			want: 4,
		},
		{
			name: "last page, no next link",
			link: `<https://codeberg.org/api/v1/repos/o/r/releases?page=1>; rel="first",` +
				`<https://codeberg.org/api/v1/repos/o/r/releases?page=3>; rel="prev"`,
			want: 0,
		},
		{
			name: "no header",
			link: "",
			want: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: getNextPage is called on it.
			got := getNextPage(tc.link)

			// THEN: the next page number is as expected.
			if got != tc.want {
				t.Fatalf(
					"%s\ngetNextPage(%q) mismatch\ngot:  %d\nwant: %d",
					packageName, tc.link,
					got, tc.want,
				)
			}
		})
	}
}

func TestLookup_Query__UnvalidatedHost(t *testing.T) {
	// GIVEN: a Lookup whose host wouldn't pass CheckValues.
	lookup := testLookup(t, `url: owner/repo`)

	// WHEN: it is queried.
	_, err := lookup.Query(true, logx.LogFrom{Primary: t.Name()})

	// THEN: the request errors.
	e := errfmt.FormatError(err)
	if wantRe := `no Host in request URL`; !util.RegexCheck(wantRe, e) {
		t.Fatalf(
			"%s\nLookup.Query() error mismatch\ngot:  %q\nwant: %q",
			packageName, e, wantRe,
		)
	}
}

func TestLookup_Query__UnreachableInstance(t *testing.T) {
	// GIVEN: an instance that is not listening.
	server := newForgeServer(t, map[string]forgeEndpoint{
		endpointReleases: {body: releasesBody},
	})
	host := server.URL
	server.Close()

	lookup := testLookup(t, "host: "+host+"\nurl: owner/repo")

	// WHEN: it is queried.
	_, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()})

	// THEN: the transport error is reported.
	e := errfmt.FormatError(err)
	if !util.RegexCheck(`connection refused|connect: `, e) {
		t.Fatalf(
			"%s\nLookup.Query() error mismatch\ngot:  %q\nwant: a connection error",
			packageName, e,
		)
	}
}

func TestLookup_Query__RejectsAnUntrustedCertificate(t *testing.T) {
	// GIVEN: an instance presenting a certificate no client trusts.
	server := newForgeServerTLS(t, map[string]forgeEndpoint{
		endpointReleases: {body: releasesBody},
	})
	lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")

	// WHEN: it is queried.
	_, err := lookup.Query(false, logx.LogFrom{Primary: t.Name()})

	// THEN: the query fails on the certificate, and no version is reported.
	e := errfmt.FormatError(err)
	if !util.RegexCheck(`certificate|x509`, e) {
		t.Fatalf(
			"%s\nLookup.Query() error mismatch\ngot:  %q\nwant: a certificate error",
			packageName, e,
		)
	}
	if got := lookup.Status.LatestVersion(); got != "" {
		t.Fatalf(
			"%s\nLookup.Query() reported %q over an untrusted connection",
			packageName, got,
		)
	}
}

func TestEndpointYieldedNothing(t *testing.T) {
	// GIVEN: an error from an endpoint.
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

			// WHEN: endpointYieldedNothing is called on it.
			got := endpointYieldedNothing(tc.err)

			// THEN: only an empty list or a 404 counts as nothing to filter.
			if got != tc.want {
				t.Fatalf(
					"%s\nendpointYieldedNothing(%v) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.err,
					got, tc.want,
				)
			}
		})
	}
}

func TestNothingFound(t *testing.T) {
	// GIVEN: what each endpoint tried came back with.
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

			// WHEN: nothingFound is called.
			got := nothingFound(tc.releasesErr, tc.tagsErr, tc.triedTags)

			// THEN: the message names only the causes still possible.
			if got == nil || got.Error() != tc.want {
				t.Fatalf(
					"%s\nnothingFound(%v, %v, %t) mismatch\ngot:  %v\nwant: %q",
					packageName, tc.releasesErr, tc.tagsErr, tc.triedTags, got, tc.want,
				)
			}
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	// GIVEN: an API response.
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
			name:    "limited/403 with a RateLimit-Policy header",
			status:  http.StatusForbidden,
			headers: map[string]string{"RateLimit-Policy": `"baseline";q=2000;w=600`},
			want:    true,
		},
		{
			name:    "limited/403 with GitHub's X-RateLimit-Remaining",
			status:  http.StatusForbidden,
			headers: map[string]string{"X-RateLimit-Remaining": "0"},
			want:    true,
		},
		{
			name:   "not limited/403 with no rate-limit headers",
			status: http.StatusForbidden,
			want:   false,
		},
		{
			name:    "not limited/403 with only Retry-After (possible maintenance window)",
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

			response := newResponse(t, "https://codeberg.org", tc.status, tc.headers, "")

			// WHEN: isRateLimited is called on it.
			got := isRateLimited(response)

			// THEN: only a 429, or a 403 announcing a limiter, counts as rate limited.
			if got != tc.want {
				t.Fatalf(
					"%s\nisRateLimited(status=%d, headers=%v) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.status, tc.headers,
					got, tc.want,
				)
			}
		})
	}
}

func TestHandleResponse(t *testing.T) {
	// GIVEN: a response with a given status and headers.
	tests := []struct {
		name         string
		status       int
		headers      map[string]string
		body         string
		page         int
		wantBody     string
		wantNextPage int
		errRegex     string
	}{
		{
			name:     "200 returns the body",
			status:   http.StatusOK,
			body:     releasesBody,
			page:     1,
			wantBody: releasesBody,
			errRegex: `^$`,
		},
		{
			name:     "200 with an empty first page is nothing to filter",
			status:   http.StatusOK,
			body:     emptyListNewline,
			page:     1,
			errRegex: `^empty list$`,
		},
		{
			name:     "200 on page 0",
			status:   http.StatusOK,
			body:     emptyListCompact,
			page:     0,
			errRegex: `^empty list$`,
		},
		{
			name:     "200 with an empty later page is returned, not reported as nothing",
			status:   http.StatusOK,
			body:     emptyListNewline,
			page:     2,
			wantBody: emptyListNewline,
			errRegex: `^$`,
		},
		{
			name:   "200 takes the next page from the Link header",
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
			name:     "404 is the ambiguous not-found",
			status:   http.StatusNotFound,
			body:     `{"message":"The target couldn't be found."}`,
			page:     1,
			errRegex: `^not found$`,
		},
		{
			name:     "401 is an authentication failure",
			status:   http.StatusUnauthorized,
			page:     1,
			errRegex: `^authentication failed for codeberg.org \(401\)$`,
		},
		{
			name:     "403 with no limiter headers is an authentication failure",
			status:   http.StatusForbidden,
			page:     1,
			errRegex: `^authentication failed for codeberg.org \(403\)$`,
		},
		{
			name:     "403 announcing a limiter is a rate limit",
			status:   http.StatusForbidden,
			headers:  map[string]string{"RateLimit": `"baseline";r=0;t=600`},
			page:     1,
			errRegex: `^rate limit reached for codeberg.org$`,
		},
		{
			name:     "429 is a rate limit",
			status:   http.StatusTooManyRequests,
			body:     `<!DOCTYPE html>`,
			page:     1,
			errRegex: `^rate limit reached for codeberg.org$`,
		},
		{
			name:     "429 with a retry window names it",
			status:   http.StatusTooManyRequests,
			headers:  map[string]string{"Retry-After": "120"},
			body:     `<!DOCTYPE html>`,
			page:     1,
			errRegex: `^rate limit reached for codeberg.org - retry after 120$`,
		},
		{
			name:     "an unmapped status is reported with its body",
			status:   http.StatusInternalServerError,
			body:     "proxy error",
			page:     1,
			errRegex: `^unknown status code 500\nproxy error$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			response := newResponse(t,
				"https://codeberg.org/api/v1/repos/o/r/releases",
				tc.status, tc.headers, tc.body)

			// WHEN: handleResponse is called.
			body, nextPage, err := handleResponse(
				response, []byte(tc.body), tc.page,
				logx.LogFrom{Primary: t.Name()},
			)

			prefix := fmt.Sprintf(
				"%s\nhandleResponse(%d)",
				packageName, tc.status,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the body and next page are as expected.
			if string(body) != tc.wantBody {
				t.Fatalf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, string(body), tc.wantBody,
				)
			}
			if nextPage != tc.wantNextPage {
				t.Fatalf(
					"%s next page mismatch\ngot:  %d\nwant: %d",
					prefix, nextPage, tc.wantNextPage,
				)
			}
		})
	}
}

func TestLookup_FilterReleases(t *testing.T) {
	// GIVEN: releases, and a Lookup whose settings filter them.
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
				t.Fatalf(
					"%s\nfailed to unmarshal the body: %v",
					packageName, err,
				)
			}

			// WHEN: filterReleases is called on them.
			got := lookup.filterReleases(releases, logx.LogFrom{Primary: t.Name()})

			// THEN: only the expected releases are kept, in the expected order.
			if err := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a forgetypes.Release, b string) bool { return a.TagName == b },
				fmt.Sprintf("%s\nLookup.filterReleases()", packageName),
				"",
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLookup_GetVersion(t *testing.T) {
	// GIVEN: a response body, and a Lookup whose filters must be met.
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
			name: "a tag has no published_at",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo`,
			body: tagsBody,
			page: 1,

			wantVersion: "0.12.1",
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

			// WHEN: getVersion is called on the body.
			version, releaseDate, err := lookup.getVersion(
				[]byte(tc.body), tc.page,
				logx.LogFrom{Primary: t.Name()},
			)

			prefix := fmt.Sprintf(
				"%s\nLookup.getVersion(page=%d)",
				packageName, tc.page,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the version and its release date are as expected.
			if version != tc.wantVersion {
				t.Fatalf(
					"%s version mismatch\ngot:  %q\nwant: %q",
					prefix, version, tc.wantVersion,
				)
			}
			if releaseDate != tc.wantReleaseDate {
				t.Fatalf(
					"%s release date mismatch\ngot:  %q\nwant: %q",
					prefix, releaseDate, tc.wantReleaseDate,
				)
			}
		})
	}
}

func TestLookup_CreateRequest(t *testing.T) {
	// GIVEN: a Lookup, an endpoint and a page.
	tests := []struct {
		name     string
		host     string
		endpoint string
		page     int
		wantURL  string
		errRegex string
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
			name:     "an unvalidated host still builds an address", // CheckValues validates.
			host:     "",
			endpoint: endpointReleases,
			page:     1,
			wantURL:  "https:///api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := Lookup{Host: tc.host}
			lookup.URL = "owner/repo"

			// WHEN: createRequest is called.
			request, err := lookup.createRequest(tc.endpoint, tc.page, logx.LogFrom{Primary: t.Name()})

			prefix := fmt.Sprintf(
				"%s\nLookup.createRequest(%q, %d)",
				packageName, tc.endpoint, tc.page,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if tc.errRegex != "^$" {
				return
			}

			// AND: it is a GET for the expected address.
			if request.Method != http.MethodGet {
				t.Fatalf(
					"%s method mismatch\ngot:  %q\nwant: %q",
					prefix, request.Method, http.MethodGet,
				)
			}
			if got := request.URL.String(); got != tc.wantURL {
				t.Fatalf(
					"%s address mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantURL,
				)
			}

			// AND: it carries no conditional-request header.
			for _, header := range []string{"If-None-Match", "If-Modified-Since"} {
				if got := request.Header.Get(header); got != "" {
					t.Fatalf(
						"%s set %s: %q",
						prefix, header, got,
					)
				}
			}
		})
	}
}

func TestGetResponse(t *testing.T) {
	// GIVEN: an instance, and a request for one of its endpoints.
	t.Run("valid/the response and its body come back", func(t *testing.T) {
		t.Parallel()

		server := newForgeServer(t, map[string]forgeEndpoint{
			endpointReleases: {body: releasesBody},
		})
		lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")
		logFrom := logx.LogFrom{Primary: t.Name()}

		request, err := lookup.createRequest(endpointReleases, 1, logFrom)
		if err != nil {
			t.Fatalf(
				"%s\nLookup.createRequest() unexpected error: %v",
				packageName, err,
			)
		}

		// WHEN: getResponse is called.
		response, body, err := getResponse(request, logFrom)
		if err != nil {
			t.Fatalf(
				"%s\ngetResponse() unexpected error: %v",
				packageName, err,
			)
		}

		// THEN: the status and body are as served.
		if response.StatusCode != http.StatusOK {
			t.Fatalf(
				"%s\ngetResponse() status mismatch\ngot:  %d\nwant: %d",
				packageName, response.StatusCode, http.StatusOK,
			)
		}
		if string(body) != releasesBody {
			t.Fatalf(
				"%s\ngetResponse() body mismatch\ngot:  %q\nwant: %q",
				packageName, string(body), releasesBody,
			)
		}
	})

	t.Run("invalid/an instance that is not listening", func(t *testing.T) {
		t.Parallel()

		server := newForgeServer(t, map[string]forgeEndpoint{})
		host := server.URL
		server.Close()

		lookup := testLookup(t, "host: "+host+"\nurl: owner/repo")
		logFrom := logx.LogFrom{Primary: t.Name()}

		request, err := lookup.createRequest(endpointReleases, 1, logFrom)
		if err != nil {
			t.Fatalf(
				"%s\nLookup.createRequest() unexpected error: %v",
				packageName, err,
			)
		}

		// WHEN: getResponse is called.
		response, body, err := getResponse(request, logFrom)

		// THEN: the transport error comes back, with nothing else.
		e := errfmt.FormatError(err)
		if !util.RegexCheck(`connection refused|connect: `, e) {
			t.Fatalf(
				"%s\ngetResponse() error mismatch\ngot:  %q\nwant: a connection error",
				packageName, e,
			)
		}
		if response != nil || body != nil {
			t.Fatalf("%s\ngetResponse() returned a response or body alongside its error", packageName)
		}
	})
}

func TestLookup_RequestFor(t *testing.T) {
	// GIVEN: an address, which [Lookup.createRequest] has already assembled.
	tests := []struct {
		name     string
		address  string
		errRegex string
	}{
		{
			name:     "valid address",
			address:  "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "an address no request can be built for",
			address:  "invalid://\ttest",
			errRegex: `^failed creating http request for .*$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lookup Lookup

			// WHEN: requestFor is called.
			request, err := lookup.requestFor(tc.address, logx.LogFrom{Primary: t.Name()})

			prefix := fmt.Sprintf(
				"%s\nLookup.requestFor(%q)",
				packageName, tc.address,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if tc.errRegex != `^$` {
				return
			}

			// AND: it is a GET for that address.
			if request.Method != http.MethodGet {
				t.Fatalf(
					"%s method mismatch\ngot:  %q\nwant: %q",
					prefix, request.Method, http.MethodGet,
				)
			}
			if got := request.URL.String(); got != tc.address {
				t.Fatalf(
					"%s address mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.address,
				)
			}
		})
	}
}

func TestLookup_HTTPRequest(t *testing.T) {
	// GIVEN: an instance responding in a particular way.
	tests := []struct {
		name         string
		endpoints    map[string]forgeEndpoint
		endpoint     string
		page         int
		wantBody     string
		wantNextPage int
		errRegex     string
	}{
		{
			name: "valid/a page of releases",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: releasesBody},
			},
			endpoint: endpointReleases,
			page:     1,
			wantBody: releasesBody,
			errRegex: `^$`,
		},
		{
			name: "valid/a paginated page advertises the next",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {pages: []string{releasesBody, tagsBody}},
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
				endpointReleases: {body: emptyListNewline},
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newForgeServer(t, tc.endpoints)
			lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")

			// WHEN: httpRequest is called.
			body, nextPage, err := lookup.httpRequest(
				tc.endpoint, tc.page,
				logx.LogFrom{Primary: t.Name()},
			)

			prefix := fmt.Sprintf(
				"%s\nLookup.httpRequest(%q, %d)",
				packageName, tc.endpoint, tc.page,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the body and next page are as expected.
			if string(body) != tc.wantBody {
				t.Fatalf(
					"%s body mismatch\ngot:  %q\nwant: %q",
					prefix, string(body), tc.wantBody,
				)
			}
			if nextPage != tc.wantNextPage {
				t.Fatalf(
					"%s next page mismatch\ngot:  %d\nwant: %d",
					prefix, nextPage, tc.wantNextPage,
				)
			}
		})
	}
}

func TestLookup_QueryPage(t *testing.T) {
	// GIVEN: an instance, and a page to ask it for.
	tests := []struct {
		name            string
		endpoints       map[string]forgeEndpoint
		page            int
		latestVersion   string
		wantNewVersion  bool
		wantNextPage    int
		wantLatest      string
		errRegex        string
		wantRequestPath string
	}{
		{
			name: "a first version is recorded, but is not announced as new",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: releasesBody},
			},
			page:       1,
			wantLatest: "0.17.4",
			errRegex:   `^$`,
		},
		{
			name: "a version replacing an older one is new",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: releasesBody},
			},
			page:           1,
			latestVersion:  "0.1.0",
			wantNewVersion: true,
			wantLatest:     "0.17.4",
			errRegex:       `^$`,
		},
		{
			name: "an unchanged version is neither new nor an error",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: releasesBody},
			},
			page:          1,
			latestVersion: "0.17.4",
			wantLatest:    "0.17.4",
			errRegex:      `^$`,
		},
		{
			name: "no version on this page, but another page to try",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {pages: []string{
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
				endpointReleases: {body: `[{"tag_name":"not-a-version","prerelease":false}]`},
			},
			page:     1,
			errRegex: `^no releases were found matching the url_commands on page 1 of the API response$`,
		},
		{
			name:      "a transport-level failure stops the walk",
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

			// WHEN: queryPage is called for the second check, so it commits rather than re-checking.
			newVersion, nextPage, err := lookup.queryPage(
				1, tc.page, endpointReleases,
				logx.LogFrom{Primary: t.Name()},
			)

			prefix := fmt.Sprintf(
				"%s\nLookup.queryPage(page=%d)",
				packageName, tc.page,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: whether a new version was found, and the next page, are as expected.
			if newVersion != tc.wantNewVersion {
				t.Fatalf(
					"%s new version mismatch\ngot:  %t\nwant: %t",
					prefix, newVersion, tc.wantNewVersion,
				)
			}
			if nextPage != tc.wantNextPage {
				t.Fatalf(
					"%s next page mismatch\ngot:  %d\nwant: %d",
					prefix, nextPage, tc.wantNextPage,
				)
			}
			if got := lookup.Status.LatestVersion(); got != tc.wantLatest {
				t.Fatalf(
					"%s latest version mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantLatest,
				)
			}
		})
	}
}

func TestLookup_QueryEndpoint(t *testing.T) {
	// GIVEN: an instance whose pages must be walked.
	tests := []struct {
		name       string
		endpoints  map[string]forgeEndpoint
		endpoint   string
		wantLatest string
		errRegex   string
	}{
		{
			name: "a version on the first page",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: releasesBody},
			},
			endpoint:   endpointReleases,
			wantLatest: "0.17.4",
			errRegex:   `^$`,
		},
		{
			name: "a version only on a later page",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {pages: []string{
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
			name: "the tags endpoint is walked",
			endpoints: map[string]forgeEndpoint{
				endpointTags: {body: tagsBody},
			},
			endpoint:   endpointTags,
			wantLatest: "0.12.1",
			errRegex:   `^$`,
		},
		{
			name: "an empty list stops the walk with nothing to filter",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: emptyListCompact},
			},
			endpoint: endpointReleases,
			errRegex: `^empty list$`,
		},
		{
			name: "pages that never yield a version end in an error",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {pages: []string{
					`[{"tag_name":"not-a-version","prerelease":false}]`,
					`[{"tag_name":"nor-this","prerelease":false}]`,
				}},
			},
			endpoint: endpointReleases,
			errRegex: `^no releases were found matching the url_commands on page 2 of the API response$`,
		},
		{
			name: "a next page that does not advance is refused",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					body: `[{"tag_name":"not-a-version","prerelease":false}]`,
					headers: map[string]string{
						"Link": `<https://root-url.example.com/api/v1/repos/o/r/releases?page=1>; rel="next"`,
					},
				},
			},
			endpoint: endpointReleases,
			errRegex: `^/releases did not advance past page 1$`,
		},
		{
			name: "an endless walk stops at the page cap",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {
					body:    `[{"tag_name":"not-a-version","prerelease":false}]`,
					endless: true,
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

			// WHEN: queryEndpoint is called.
			_, err := lookup.queryEndpoint(tc.endpoint, logx.LogFrom{Primary: t.Name()})

			prefix := fmt.Sprintf(
				"%s\nLookup.queryEndpoint(%q)",
				packageName, tc.endpoint,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the version found is as expected.
			if got := lookup.Status.LatestVersion(); got != tc.wantLatest {
				t.Fatalf(
					"%s latest version mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantLatest,
				)
			}
		})
	}
}

func TestLookup_Query__EndpointFallback(t *testing.T) {
	// GIVEN: an instance, and a service whose filters may or may not allow the tags fallback.
	tests := []struct {
		name          string
		endpoints     map[string]forgeEndpoint
		lookupYAML    string
		wantRequested []string
		wantLatest    string
		errRegex      string
	}{
		{
			name: "releases answer, so tags are never asked for",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: releasesBody},
				endpointTags:     {body: tagsBody},
			},
			lookupYAML:    `url: owner/repo`,
			wantRequested: []string{endpointReleases},
			wantLatest:    "0.17.4",
			errRegex:      `^$`,
		},
		{
			name: "an empty releases list falls through to tags",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: emptyListNewline},
				endpointTags:     {body: tagsBody},
			},
			lookupYAML:    `url: owner/repo`,
			wantRequested: []string{endpointReleases, endpointTags},
			wantLatest:    "0.12.1",
			errRegex:      `^$`,
		},
		{
			name: "filters that tags cannot satisfy stop the fallback",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {body: emptyListNewline},
				endpointTags:     {body: tagsBody},
			},
			lookupYAML: `
				url: owner/repo
				use_prerelease: true`,
			wantRequested: []string{endpointReleases},
			errRegex:      `^no releases were found$`,
		},
		{
			name: "a rate limit stops the fallback",
			endpoints: map[string]forgeEndpoint{
				endpointReleases: {status: http.StatusTooManyRequests, body: "<!DOCTYPE html>"},
				endpointTags:     {body: tagsBody},
			},
			lookupYAML:    `url: owner/repo`,
			wantRequested: []string{endpointReleases},
			errRegex:      `^rate limit reached for 127\.0\.0\.1:\d+$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := newForgeServer(t, tc.endpoints)
			lookupYAML := "host: " + server.URL + "\n" + test.TrimYAML(tc.lookupYAML)
			lookup := testLookup(t, lookupYAML)

			// WHEN: query is called.
			_, err := lookup.query(logx.LogFrom{Primary: t.Name()})

			prefix := fmt.Sprintf("%s\nLookup.query()", packageName)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the version found is as expected.
			if got := lookup.Status.LatestVersion(); got != tc.wantLatest {
				t.Fatalf(
					"%s latest version mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wantLatest,
				)
			}

			// AND: only the expected endpoints were asked for.
			var requested []string
			for _, request := range server.requests() {
				endpoint := path.Base(request.path)
				if len(requested) == 0 || requested[len(requested)-1] != endpoint {
					requested = append(requested, endpoint)
				}
			}
			if !slices.Equal(requested, tc.wantRequested) {
				t.Fatalf(
					"%s endpoints requested mismatch\ngot:  %v\nwant: %v",
					prefix, requested, tc.wantRequested,
				)
			}
		})
	}
}

func TestLookup_HandleNewVersion(t *testing.T) {
	// GIVEN: an instance serving a version, and a service that has not seen it.
	t.Run("the first check re-queries before committing", func(t *testing.T) {
		t.Parallel()

		server := newForgeServer(t, map[string]forgeEndpoint{
			endpointReleases: {body: releasesBody},
		})
		lookup := testLookup(t, "host: "+server.URL+"\nurl: owner/repo")

		// WHEN: handleNewVersion is called as the first check.
		_, err := lookup.handleNewVersion(
			0,
			"0.17.4", "2026-08-01T10:50:00Z", "",
			endpointReleases, 1,
			logx.LogFrom{Primary: t.Name()},
		)
		if err != nil {
			t.Fatalf(
				"%s\nLookup.handleNewVersion() unexpected error: %v",
				packageName, err,
			)
		}

		// THEN: it asked the instance again before recording the version.
		if got := len(server.requests()); got != 1 {
			t.Fatalf(
				"%s\nLookup.handleNewVersion() request count mismatch\ngot:  %d\nwant: 1",
				packageName, got,
			)
		}
		if got, want := lookup.Status.LatestVersion(), "0.17.4"; got != want {
			t.Fatalf(
				"%s\nLookup.handleNewVersion() latest version mismatch\ngot:  %q\nwant: %q",
				packageName, got, want,
			)
		}
	})

	t.Run("the first version found is recorded, but not announced as new", func(t *testing.T) {
		t.Parallel()

		lookup := testLookup(t, test.TrimYAML(`
			host: https://codeberg.org
			url: owner/repo`))

		// WHEN: handleNewVersion is called as the second check, with no version known.
		newVersion, err := lookup.handleNewVersion(
			1,
			"1.2.3", "2026-01-01T00:00:00Z", "",
			endpointReleases, 1,
			logx.LogFrom{Primary: t.Name()},
		)
		if err != nil {
			t.Fatalf(
				"%s\nLookup.handleNewVersion() unexpected error: %v",
				packageName, err,
			)
		}

		// THEN: it is recorded without being reported as an update.
		if newVersion {
			t.Fatalf("%s\nLookup.handleNewVersion() reported the first version as new", packageName)
		}
		if got, want := lookup.Status.LatestVersion(), "1.2.3"; got != want {
			t.Fatalf(
				"%s\nLookup.handleNewVersion() latest version mismatch\ngot:  %q\nwant: %q",
				packageName, got, want,
			)
		}
	})

	t.Run("a version replacing an older one is announced as new", func(t *testing.T) {
		t.Parallel()

		lookup := testLookup(t, test.TrimYAML(`
			host: https://codeberg.org
			url: owner/repo`))
		lookup.Status.SetLatestVersion("1.0.0", "", false)

		// WHEN: handleNewVersion is called as the second check.
		newVersion, err := lookup.handleNewVersion(
			1,
			"1.2.3", "2026-01-01T00:00:00Z", "1.0.0",
			endpointReleases, 1,
			logx.LogFrom{Primary: t.Name()},
		)
		if err != nil {
			t.Fatalf(
				"%s\nLookup.handleNewVersion() unexpected error: %v",
				packageName, err,
			)
		}

		// THEN: it is reported as an update.
		if !newVersion {
			t.Fatalf("%s\nLookup.handleNewVersion() did not report the version as new", packageName)
		}
		if got, want := lookup.Status.LatestVersion(), "1.2.3"; got != want {
			t.Fatalf(
				"%s\nLookup.handleNewVersion() latest version mismatch\ngot:  %q\nwant: %q",
				packageName, got, want,
			)
		}
	})
}

func TestLookup_HandleNoVersionChange(t *testing.T) {
	// GIVEN: a service already on the version just found.
	tests := []struct {
		name        string
		checkNumber int
	}{
		{
			name:        "the first check",
			checkNumber: 0,
		},
		{
			name:        "the second check, which also logs that it is staying put",
			checkNumber: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: https://codeberg.org
				url: owner/repo`))
			lookup.Status.SetLatestVersion("1.2.3", "", false)
			// Drain what setting the version announced.
			for len(lookup.Status.AnnounceChannel) > 0 {
				<-lookup.Status.AnnounceChannel
			}

			// WHEN: handleNoVersionChange is called.
			lookup.handleNoVersionChange(tc.checkNumber, "1.2.3", logx.LogFrom{Primary: t.Name()})

			prefix := fmt.Sprintf(
				"%s\nLookup.handleNoVersionChange(%d)",
				packageName, tc.checkNumber,
			)

			// THEN: the query is announced, and the version is left alone.
			if got := len(lookup.Status.AnnounceChannel); got != 1 {
				t.Fatalf(
					"%s announcement count mismatch\ngot:  %d\nwant: 1",
					prefix, got,
				)
			}
			if got, want := lookup.Status.LatestVersion(), "1.2.3"; got != want {
				t.Fatalf(
					"%s latest version mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}
		})
	}
}
