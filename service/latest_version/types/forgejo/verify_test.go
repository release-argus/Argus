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
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN: a Lookup in YAML.
	tests := []struct {
		name       string
		lookupYAML string
		defaults   map[string]HostDefaults
		errRegex   string
	}{
		{
			name: "valid/host and url",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo`,
			errRegex: `^$`,
		},
		{
			name: "valid/bare host",
			lookupYAML: `
				host: codeberg.org
				url: owner/repo`,
			errRegex: `^$`,
		},
		{
			name: "valid/sub-path host",
			lookupYAML: `
				host: https://example.com:8080/git
				url: owner/repo`,
			errRegex: `^$`,
		},
		{
			name: "invalid/no host",
			lookupYAML: `
				url: owner/repo`,
			errRegex: `^host: <required> \(e\.g\. https://codeberg\.org\)$`,
		},
		{
			name: "invalid/unsupported scheme",
			lookupYAML: `
				host: ftp://codeberg.org
				url: owner/repo`,
			errRegex: `^host: "ftp://codeberg.org" <invalid> \(scheme must be http or https\)$`,
		},
		{
			name: "invalid/host surrounded by whitespace",
			lookupYAML: `
				host: "  https://codeberg.org  "
				url: owner/repo`,
			errRegex: `^host: "  https://codeberg.org  " <invalid> \(surrounded by whitespace\)$`,
		},
		{
			name: "valid/trailing slash is trimmed, not rejected",
			lookupYAML: `
				host: https://codeberg.org/
				url: owner/repo`,
			errRegex: `^$`,
		},
		{
			name: "valid/trailing slash on a sub-path host",
			lookupYAML: `
				host: https://example.com/git/
				url: owner/repo`,
			errRegex: `^$`,
		},
		{
			name: "invalid/no url",
			lookupYAML: `
				host: https://codeberg.org`,
			errRegex: `^url: <required> \(e\.g\. owner/repo\)$`,
		},
		{
			name: "invalid/url is a full address, not owner-repo",
			lookupYAML: `
				host: https://codeberg.org
				url: https://codeberg.org/owner/repo`,
			errRegex: `^url: "https://codeberg.org/owner/repo" <invalid> \(e\.g\. owner/repo\)$`,
		},
		{
			name: "invalid/url has no owner",
			lookupYAML: `
				host: https://codeberg.org
				url: Argus`,
			errRegex: `^url: "Argus" <invalid> \(e\.g\. owner/repo\)$`,
		},
		{
			name: "invalid/url has a leading slash",
			lookupYAML: `
				host: https://codeberg.org
				url: /owner/repo`,
			errRegex: `^url: "/owner/repo" <invalid> \(e\.g\. owner/repo\)$`,
		},
		{
			name: "invalid/url has a trailing slash",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo/`,
			errRegex: `^url: "owner/repo/" <invalid> \(e\.g\. owner/repo\)$`,
		},
		{
			name: "invalid/url has leading and trailing slashes",
			lookupYAML: `
				host: https://codeberg.org
				url: /owner/repo/`,
			errRegex: `^url: "/owner/repo/" <invalid> \(e\.g\. owner/repo\)$`,
		},
		{
			name: "invalid/url has an empty repo",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/`,
			errRegex: `^url: "owner/" <invalid> \(e\.g\. owner/repo\)$`,
		},
		{
			name: "invalid/url is padded with whitespace",
			lookupYAML: `
				host: https://codeberg.org
				url: "  owner/repo  "`,
			errRegex: `^url: "  owner/repo  " <invalid> \(e\.g\. owner/repo\)$`,
		},
		{
			name: "invalid/require from the base Lookup",
			lookupYAML: `
				host: https://codeberg.org
				url: owner/repo
				require:
					regex_content: "[0-"`,
			errRegex: test.TrimYAML(`
				^require:
					regex_content: "\[0-" <invalid> .*$`,
			),
		},
		{
			name: "invalid/host and url both unusable",
			lookupYAML: `
				url: Argus`,
			errRegex: `^host: <required> .*\nurl: "Argus" <invalid> .*$`,
		},
		{
			name: "valid/host names an instance, the defaults give a URL",
			lookupYAML: `
				host: Codeberg
				url: owner/repo`,
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org"},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid/host names an instance with no URL, however the name is spelt",
			lookupYAML: `
				host: https://forgejo.example.com
				url: owner/repo`,
			defaults: map[string]HostDefaults{
				"https://forgejo.example.com": {AccessToken: "fj-token"},
			},
			errRegex: `^host: "https://forgejo.example.com" <invalid> \(names an instance with no url\)$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup, err := Decode(
				"yaml", []byte(test.TrimYAML(tc.lookupYAML)),
				nil, nil,
				plainDefaultsConfig(t),
			)
			if err != nil {
				t.Fatalf(
					"%s\nfailed to decode Lookup: %v",
					packageName, err,
				)
			}
			lookup.SetTypeDefaults(
				&Defaults{Host: tc.defaults},
				&Defaults{},
			)
			hadHost, hadURL := lookup.Host, lookup.URL

			// WHEN: CheckValues is called.
			err = lookup.CheckValues()

			prefix := fmt.Sprintf(
				"%s\nLookup.CheckValues() on %q",
				packageName, tc.lookupYAML,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: nothing beyond a trailing '/' on the host was touched.
			hadHost = strings.TrimRight(hadHost, "/")
			if lookup.Host != hadHost || lookup.URL != hadURL {
				t.Fatalf(
					"%s\nLookup.CheckValues() rewrote what the user wrote\ngot:  host=%q, url=%q\nwant: host=%q, url=%q",
					packageName, lookup.Host, lookup.URL, hadHost, hadURL,
				)
			}
		})
	}
}

func TestLookup_CheckValues__TrimsTrailingSlash(t *testing.T) {
	// GIVEN: hosts written with trailing slashes.
	tests := []struct {
		name string
		host string
		want string
	}{
		{
			name: "bare host",
			host: "https://codeberg.org/",
			want: "https://codeberg.org",
		},
		{
			name: "sub-path host",
			host: "https://example.com/git/",
			want: "https://example.com/git",
		},
		{
			name: "repeated slashes",
			host: "https://example.com/git//",
			want: "https://example.com/git",
		},
		{
			name: "nothing to trim",
			host: "https://codeberg.org",
			want: "https://codeberg.org",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: `+tc.host+`
				url: owner/repo`))

			// WHEN: CheckValues is called.
			if err := lookup.CheckValues(); err != nil {
				t.Fatalf(
					"%s\nLookup.CheckValues() unexpected error: %v",
					packageName, err,
				)
			}

			// THEN: the trailing slash is trimmed from the receiver.
			if lookup.Host != tc.want {
				t.Fatalf(
					"%s\nLookup.CheckValues() host mismatch\ngot:  %q\nwant: %q",
					packageName, lookup.Host, tc.want,
				)
			}
		})
	}
}

func TestLookup_HostProblem(t *testing.T) {
	// GIVEN: a host spelling.
	tests := []struct {
		name  string
		host  string
		hosts map[string]HostDefaults
		want  string
	}{
		{
			name: "ok/scheme and host",
			host: "https://codeberg.org",
			want: "",
		},
		{
			name: "ok/scheme defaulted",
			host: "codeberg.org",
			want: "",
		},
		{
			name: "ok/http",
			host: "http://forge.example.com",
			want: "",
		},
		{
			name: "ok/https",
			host: "https://forge.example.com",
			want: "",
		},
		{
			name: "ok/schema, host, port, sub-path",
			host: "http://forge.example.com:3000/git",
			want: "",
		},
		{
			name: "problem/no host",
			host: "",
			want: "e.g. https://codeberg.org",
		},
		{
			name: "problem/whitespace only",
			host: "   ",
			want: "e.g. https://codeberg.org",
		},
		{
			name: "problem/surrounded by whitespace",
			host: "  https://codeberg.org  ",
			want: "surrounded by whitespace",
		},
		{
			name: "problem/unsupported scheme",
			host: "ftp://codeberg.org",
			want: "scheme must be http or https",
		},
		{
			name: "problem/scheme with no hostname",
			host: "https://",
			want: "no hostname",
		},
		{
			name: "ok/trailing slash from an env var, which is trimmed at request time",
			host: "${ARGUS_TEST_FORGEJO_SLASHED}",
			want: "",
		},
		{
			name: "problem/unparseable",
			host: "https://exa mple.com",
			want: "not a valid URL",
		},
		{
			name: "problem/names an instance with no URL, however the name is spelt",
			host: "https://forgejo.example.com",
			hosts: map[string]HostDefaults{
				"https://forgejo.example.com": {AccessToken: "fj-token"},
			},
			want: "names an instance with no url",
		},
		{
			name: "ok/names an instance, and the entry gives the URL",
			host: "Codeberg",
			hosts: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org"},
			},
			want: "",
		},
		{
			name: "problem/the URL the named entry gives is unusable",
			host: "Codeberg",
			hosts: map[string]HostDefaults{
				"Codeberg": {URL: "ftp://codeberg.org"},
			},
			want: "scheme must be http or https",
		},
	}

	t.Setenv("ARGUS_TEST_FORGEJO_SLASHED", "https://example.com/git/")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lookup := Lookup{Host: tc.host}
			lookup.SetTypeDefaults(
				&Defaults{Host: tc.hosts},
				&Defaults{},
			)

			// WHEN: hostProblem is called.
			got := lookup.hostProblem()

			// THEN: it names what is wrong, or nothing at all.
			if got != tc.want {
				t.Fatalf(
					"%s\nLookup{host: %q}.hostProblem() mismatch\ngot:  %q\nwant: %q",
					packageName, tc.host, got, tc.want,
				)
			}
		})
	}
}
