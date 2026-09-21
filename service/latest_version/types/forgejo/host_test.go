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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_APIURL(t *testing.T) {
	// GIVEN: a host, a repository, and a page to request.
	tests := []struct {
		name     string
		host     string
		url      string
		page     int
		want     string
		errRegex string
	}{
		{
			name:     "valid/scheme, host",
			host:     "https://codeberg.org",
			url:      "owner/repo",
			page:     1,
			want:     "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "valid/no scheme defaults to https",
			host:     "codeberg.org",
			url:      "owner/repo",
			page:     1,
			want:     "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "valid/http scheme kept",
			host:     "http://forge.example.com",
			url:      "owner/repo",
			page:     1,
			want:     "http://forge.example.com/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "valid/port honoured",
			host:     "https://forge.example.com:3000",
			url:      "owner/repo",
			page:     1,
			want:     "https://forge.example.com:3000/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "valid/port with no scheme",
			host:     "forge.example.com:3000",
			url:      "owner/repo",
			page:     1,
			want:     "https://forge.example.com:3000/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "valid/sub-path honoured",
			host:     "https://example.com/git",
			url:      "owner/repo",
			page:     1,
			want:     "https://example.com/git/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "valid/scheme, host, port and sub-path together",
			host:     "http://example.com:8080/forge",
			url:      "owner/repo",
			page:     1,
			want:     "http://example.com:8080/forge/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "valid/page 0 treated as the first page",
			host:     "https://codeberg.org",
			url:      "owner/repo",
			page:     0,
			want:     "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "valid/page 1 treated as the first page",
			host:     "https://codeberg.org",
			url:      "owner/repo",
			page:     0,
			want:     "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50",
			errRegex: `^$`,
		},
		{
			name:     "valid/page 2 requested by number",
			host:     "https://codeberg.org",
			url:      "owner/repo",
			page:     2,
			want:     "https://codeberg.org/api/v1/repos/owner/repo/releases?limit=50&page=2",
			errRegex: `^$`,
		},
		{
			name:     "invalid/unparsable",
			host:     "https://exa mple.com",
			url:      "owner/repo",
			page:     1,
			errRegex: `^invalid host "https://exa mple.com":`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := Lookup{
				Host: tc.host,
				URL:  tc.url,
			}

			// WHEN: apiURL is called for the releases endpoint.
			got, err := lookup.apiURL(endpointReleases, tc.page)

			prefix := fmt.Sprintf(
				"%s\nLookup{host: %q, url: %q}.apiURL(endpoint: %q, page: %d)",
				packageName, tc.host, tc.url, endpointReleases, tc.page,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the address is as expected.
			if got != tc.want {
				t.Fatalf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the Host was not modified.
			if lookup.Host != tc.host {
				t.Fatalf(
					"%s mutated the Host var\ngot:  %q\nwant: %q",
					prefix, lookup.Host, tc.host,
				)
			}

			// AND: the URL was not modified.
			if lookup.URL != tc.url {
				t.Fatalf(
					"%s mutated the URL var\ngot:  %q\nwant: %q",
					prefix, lookup.URL, tc.host,
				)
			}
		})
	}
}

func TestLookup_ParseHost(t *testing.T) {
	// GIVEN: a host spelling.
	tests := []struct {
		name     string
		host     string
		want     string
		errRegex string
	}{
		{
			name:     "valid/scheme and host",
			host:     "https://codeberg.org",
			want:     "https://codeberg.org",
			errRegex: `^$`,
		},
		{
			name:     "valid/scheme defaults to https",
			host:     "codeberg.org",
			want:     "https://codeberg.org",
			errRegex: `^$`,
		},
		{
			name:     "valid/http kept",
			host:     "http://forge.example.com",
			want:     "http://forge.example.com",
			errRegex: `^$`,
		},
		{
			name:     "valid/port kept",
			host:     "https://forge.example.com:3000",
			want:     "https://forge.example.com:3000",
			errRegex: `^$`,
		},
		{
			name:     "valid/sub-path kept",
			host:     "https://example.com/git",
			want:     "https://example.com/git",
			errRegex: `^$`,
		},
		{
			name:     "valid/trailing slash passed through",
			host:     "https://example.com/git/",
			want:     "https://example.com/git/",
			errRegex: `^$`,
		},
		{
			name:     "valid/query param dropped",
			host:     "https://example.com/git?a=b#c",
			want:     "https://example.com/git",
			errRegex: `^$`,
		},
		{
			name:     "valid/environment variables expanded",
			host:     "https://${ARGUS_TEST_FORGEJO_HOST}",
			want:     "https://forge.from-env.example.com",
			errRegex: `^$`,
		},
		{
			name:     "invalid/unparsable",
			host:     "https://exa mple.com",
			errRegex: `^invalid host "https://exa mple.com":`,
		},
	}

	t.Setenv("ARGUS_TEST_FORGEJO_HOST", "forge.from-env.example.com")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := Lookup{Host: tc.host}

			// WHEN: parseHost is called.
			got, err := lookup.parseHost()

			prefix := fmt.Sprintf(
				"%s\nLookup{host: %q}.parseHost()",
				packageName, tc.host,
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

			// AND: the instance URL is as expected.
			if got.String() != tc.want {
				t.Fatalf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}
