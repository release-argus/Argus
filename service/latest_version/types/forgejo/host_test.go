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
			name:     "valid/trailing slash trimmed",
			host:     "https://example.com/git/",
			want:     "https://example.com/git",
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
			name:     "invalid/unparseable",
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

func TestParseInstanceURL(t *testing.T) {
	// GIVEN: an instance address.
	tests := []struct {
		name     string
		address  string
		want     string
		errRegex string
	}{
		{
			name:     "scheme and host",
			address:  "https://codeberg.org",
			want:     "https://codeberg.org",
			errRegex: `^$`,
		},
		{
			name:     "scheme defaults to https",
			address:  "codeberg.org",
			want:     "https://codeberg.org",
			errRegex: `^$`,
		},
		{
			name:     "http kept",
			address:  "http://forge.example.com",
			want:     "http://forge.example.com",
			errRegex: `^$`,
		},
		{
			name:     "port and sub-path kept",
			address:  "https://forge.example.com:3000/git",
			want:     "https://forge.example.com:3000/git",
			errRegex: `^$`,
		},
		{
			name:     "trailing slash trimmed, so a request path gains no '//'",
			address:  "https://example.com/git/",
			want:     "https://example.com/git",
			errRegex: `^$`,
		},
		{
			name:     "query and fragment dropped",
			address:  "https://example.com/git?a=b#c",
			want:     "https://example.com/git",
			errRegex: `^$`,
		},
		{
			name:     "an empty address gives a URL with no host",
			address:  "",
			want:     "https:",
			errRegex: `^$`,
		},
		{
			name:     "unparseable",
			address:  "https://exa mple.com",
			errRegex: `^parse "https://exa mple.com":`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: parseInstanceURL is called with it.
			got, err := parseInstanceURL(tc.address)

			prefix := fmt.Sprintf(
				"%s\nparseInstanceURL(%q)",
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
			if err != nil {
				return
			}

			// AND: the root URL of the instance is returned.
			if got.String() != tc.want {
				t.Fatalf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got.String(), tc.want,
				)
			}
		})
	}
}

func TestURLProblem(t *testing.T) {
	// GIVEN: an instance URL.
	tests := []struct {
		name    string
		address string
		want    string
	}{
		{
			name:    "ok/scheme,host",
			address: "https://codeberg.org",
			want:    "",
		},
		{
			name:    "ok/scheme-less",
			address: "codeberg.org",
			want:    "",
		},
		{
			name:    "ok/http",
			address: "http://forge.example.com",
			want:    "",
		},
		{
			name:    "ok/scheme. host. port, sub-path",
			address: "https://forge.example.com:3000/git",
			want:    "",
		},
		{
			name:    "problem/unsupported scheme",
			address: "ftp://codeberg.org",
			want:    "scheme must be http or https",
		},
		{
			name:    "problem/scheme with no hostname",
			address: "https://",
			want:    "no hostname",
		},
		{
			name:    "problem/empty",
			address: "",
			want:    "no hostname",
		},
		{
			name:    "ok/trailing slash, trimmed rather than rejected",
			address: "https://codeberg.org/",
			want:    "",
		},
		{
			name:    "ok/trailing slash on a sub-path",
			address: "https://example.com/git/",
			want:    "",
		},
		{
			name:    "problem/unparseable",
			address: "https://exa mple.com",
			want:    "not a valid URL",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: urlProblem is called with it.
			got := urlProblem(tc.address)

			// THEN: it names what is wrong, or nothing at all.
			if got != tc.want {
				t.Fatalf(
					"%s\nurlProblem(%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.address, got, tc.want,
				)
			}
		})
	}
}

func TestCanonicalHost(t *testing.T) {
	// GIVEN: a host spelling.
	tests := []struct {
		name string
		host string
		want string
	}{
		{
			name: "scheme, host, default port",
			host: "https://codeberg.org",
			want: "https://codeberg.org",
		},
		{
			name: "scheme defaults to https",
			host: "codeberg.org",
			want: "https://codeberg.org",
		},
		{
			name: "scheme and hostname lowercased",
			host: "HTTPS://Codeberg.ORG",
			want: "https://codeberg.org",
		},
		{
			name: "trailing slash stripped",
			host: "https://codeberg.org/",
			want: "https://codeberg.org",
		},
		{
			name: "surrounding whitespace stripped",
			host: "  https://codeberg.org  ",
			want: "https://codeberg.org",
		},
		{
			name: "default https port trimmed",
			host: "https://codeberg.org:443",
			want: "https://codeberg.org",
		},
		{
			name: "default http port trimmed",
			host: "http://forge.example.com:80",
			want: "http://forge.example.com",
		},
		{
			name: "non-default port kept",
			host: "https://forge.example.com:3000",
			want: "https://forge.example.com:3000",
		},
		{
			name: "https port on http kept",
			host: "http://forge.example.com:443",
			want: "http://forge.example.com:443",
		},
		{
			name: "sub-path kept, case preserved",
			host: "https://example.com/Git",
			want: "https://example.com/Git",
		},
		{
			name: "bare host with a sub-path",
			host: "example.com/git",
			want: "https://example.com/git",
		},
		{
			name: "sub-path trailing slash stripped",
			host: "https://example.com/git/",
			want: "https://example.com/git",
		},
		{
			name: "casing, scheme, port, sub-path",
			host: "HTTP://Forge.Example.COM:8080/Git/",
			want: "http://forge.example.com:8080/Git",
		},
		{
			name: "IPv6 literal maintains brackets",
			host: "https://[2001:DB8::1]:3000",
			want: "https://[2001:db8::1]:3000",
		},
		{
			name: "IPv6 literal on the default port",
			host: "[2001:db8::1]:443",
			want: "https://[2001:db8::1]",
		},
		{
			name: "uncompressed IPv6 left as typed",
			host: "https://[0:0:0:0:0:0:0:1]",
			want: "https://[0:0:0:0:0:0:0:1]",
		},
		{
			name: "userinfo dropped",
			host: "https://user@git.example.com",
			want: "https://git.example.com",
		},
		{
			name: "query and fragment dropped",
			host: "https://example.com/git?a=b#c",
			want: "https://example.com/git",
		},
		{
			name: "environment variables expanded",
			host: "https://${ARGUS_TEST_FORGEJO_CANONICAL_HOST}",
			want: "https://forge.from-env.example.com",
		},
		{
			name: "unparseable falls back to the lowercased address",
			host: "https://EXAMPLE.com/%zz",
			want: "https://example.com/%zz",
		},
		{
			name: "empty stays empty",
			host: "",
			want: "",
		},
		{
			name: "whitespace-only cleared",
			host: "   ",
			want: "",
		},
	}

	t.Setenv("ARGUS_TEST_FORGEJO_CANONICAL_HOST", "forge.from-env.example.com")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: canonicalHost is called with it.
			got := canonicalHost(tc.host)

			// THEN: the canonical form is as expected.
			if got != tc.want {
				t.Fatalf(
					"%s\ncanonicalHost(%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.host,
					got, tc.want,
				)
			}
		})
	}
}

func TestIsDefaultPort(t *testing.T) {
	// GIVEN: a scheme and a port.
	tests := []struct {
		name   string
		scheme string
		port   string
		want   bool
	}{
		{
			name:   "443/http",
			scheme: "http",
			port:   "443",
		},
		{
			name:   "443/https",
			scheme: "https",
			port:   "443",
			want:   true,
		},
		{
			name:   "443/unknown scheme",
			scheme: "ftp",
			port:   "443",
		},
		{
			name:   "80/http",
			scheme: "http",
			port:   "80",
			want:   true,
		},
		{
			name:   "80/https",
			scheme: "https",
			port:   "80",
		},
		{
			name:   "80/unknown scheme",
			scheme: "ftp",
			port:   "80",
		},
		{
			name:   "non-default port/http",
			scheme: "http",
			port:   "3000",
		},
		{
			name:   "non-default port/https",
			scheme: "https",
			port:   "3000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: isDefaultPort is called with them.
			got := isDefaultPort(tc.scheme, tc.port)

			// THEN: it reports whether the scheme implies that port.
			if got != tc.want {
				t.Fatalf(
					"%s\nisDefaultPort(%q, %q) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.scheme, tc.port,
					got, tc.want,
				)
			}
		})
	}
}

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
			name:     "invalid/unparseable",
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
