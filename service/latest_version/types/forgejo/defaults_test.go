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

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestHostDefaults_IsZero(t *testing.T) {
	// GIVEN: a HostDefaults with some combination of its fields set.
	tests := []struct {
		name  string
		entry HostDefaults
		want  bool
	}{
		{
			name: "empty",
			want: true,
		},
		{
			name:  "URL only",
			entry: HostDefaults{URL: "https://codeberg.org"},
		},
		{
			name:  "Access Token only",
			entry: HostDefaults{AccessToken: "cb-token"},
		},
		{
			name:  "Certificate trust only",
			entry: HostDefaults{AllowInvalidCerts: new(false)},
		},
		{
			name: "every field",
			entry: HostDefaults{
				URL:               "https://codeberg.org",
				AccessToken:       "cb-token",
				AllowInvalidCerts: new(true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called.
			got := tc.entry.IsZero()

			// THEN: it reports whether anything is set.
			if got != tc.want {
				t.Fatalf(
					"%s\nHostDefaults.IsZero() mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestHostDefaults_Inherit(t *testing.T) {
	// GIVEN: an entry, and the entry it falls back on.
	tests := []struct {
		name  string
		host  HostDefaults
		entry HostDefaults
		want  HostDefaults
	}{
		{
			name:  "an empty receiver takes every field",
			entry: HostDefaults{URL: "https://codeberg.org", AccessToken: "bar", AllowInvalidCerts: new(true)},
			want:  HostDefaults{URL: "https://codeberg.org", AccessToken: "bar", AllowInvalidCerts: new(true)},
		},
		{
			name:  "set fields are kept",
			host:  HostDefaults{URL: "https://own.example.com", AccessToken: "own", AllowInvalidCerts: new(false)},
			entry: HostDefaults{URL: "https://codeberg.org", AccessToken: "bar", AllowInvalidCerts: new(true)},
			want:  HostDefaults{URL: "https://own.example.com", AccessToken: "own", AllowInvalidCerts: new(false)},
		},
		{
			name:  "set fields aren't changed, others fallback",
			host:  HostDefaults{AccessToken: "own"},
			entry: HostDefaults{URL: "https://codeberg.org", AccessToken: "bar", AllowInvalidCerts: new(true)},
			want:  HostDefaults{URL: "https://codeberg.org", AccessToken: "own", AllowInvalidCerts: new(true)},
		},
		{
			name:  "allow_invalid_certs=false is set, so it is kept",
			host:  HostDefaults{AllowInvalidCerts: new(false)},
			entry: HostDefaults{AllowInvalidCerts: new(true)},
			want:  HostDefaults{AllowInvalidCerts: new(false)},
		},
		{
			name: "an empty entry changes nothing",
			host: HostDefaults{URL: "https://own.example.com", AccessToken: "own"},
			want: HostDefaults{URL: "https://own.example.com", AccessToken: "own"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.host

			// WHEN: inherit is called with the entry to fall back on.
			got.inherit(tc.entry)

			// THEN: only the unset fields were filled.
			for _, check := range []struct {
				name, got, want string
			}{
				{
					name: "url",
					got:  got.URL,
					want: tc.want.URL,
				},
				{
					name: "access_token",
					got:  got.AccessToken,
					want: tc.want.AccessToken,
				},
				{
					name: "allow_invalid_certs",
					got:  test.StringifyPtr(got.AllowInvalidCerts),
					want: test.StringifyPtr(tc.want.AllowInvalidCerts),
				},
			} {
				if check.got != check.want {
					t.Errorf(
						"%s\nHostDefaults.inherit() %s mismatch\ngot:  %q\nwant: %q",
						packageName, check.name,
						check.got, check.want,
					)
				}
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
	// GIVEN: a Defaults with some combination of its fields set.
	tests := []struct {
		name     string
		defaults Defaults
		want     bool
	}{
		{
			name: "empty",
			want: true,
		},
		{
			name:     "common only",
			defaults: Defaults{Common: CommonDefaults{UsePreRelease: new(false)}},
		},
		{
			name: "one instance",
			defaults: Defaults{Host: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org"},
			}},
		},
		{
			name: "an empty instance still counts",
			defaults: Defaults{Host: map[string]HostDefaults{
				"Codeberg": {},
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called.
			got := tc.defaults.IsZero()

			// THEN: it reports whether anything is set.
			if got != tc.want {
				t.Fatalf(
					"%s\nDefaults.IsZero() mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDefaults_Default(t *testing.T) {
	// GIVEN: a Defaults.
	defaults := Defaults{}

	// WHEN: Default is called.
	defaults.Default()

	// THEN: prereleases are excluded.
	if defaults.Common.UsePreRelease == nil || *defaults.Common.UsePreRelease {
		t.Fatalf(
			"%s\nDefaults.Default() common.use_prerelease mismatch\ngot:  %v\nwant: false",
			packageName, defaults.Common.UsePreRelease,
		)
	}
}

func TestDefaults_SetDefaults(t *testing.T) {
	host := "Codeberg"
	// GIVEN: a Defaults naming an instance it gives no URL.
	defaults := Defaults{Host: map[string]HostDefaults{
		host: {AccessToken: "cb-token"},
	}}
	hardDefaults := Defaults{Host: map[string]HostDefaults{
		host: {URL: "https://codeberg.org"},
	}}

	// WHEN: nothing has been set to fall back on.
	resolved, _ := defaults.resolvedHost(host)
	if g, w := resolved.URL, ""; g != w {
		t.Fatalf(
			"%s\nDefaults.resolvedHost().URL before SetDefaults mismatch\ngot:  %q\nwant: %q",
			packageName, g, w,
		)
	}

	// WHEN: SetDefaults gives it a layer to fall back on.
	defaults.SetDefaults(&hardDefaults)

	// THEN: that layer is consulted.
	resolved, _ = defaults.resolvedHost(host)
	if g, w := resolved.URL, hardDefaults.Host[host].URL; g != w {
		t.Fatalf(
			"%s\nDefaults.resolvedHost().URL after SetDefaults mismatch\ngot:  %q\nwant: %q",
			packageName, g, w,
		)
	}

	// AND: a nil Defaults takes a layer without panicking.
	var nilDefaults *Defaults
	nilDefaults.SetDefaults(&hardDefaults)
	if resolved, named := nilDefaults.resolvedHost(host); named || resolved.URL != "" {
		t.Errorf(
			"%s\nnil Defaults.resolvedHost() mismatch\ngot:  url=%q named=%t\nwant: url=\"\" named=false",
			packageName, resolved.URL, named,
		)
	}
}

func TestDefaults_ResolvedHost(t *testing.T) {
	// GIVEN: a host, and instances layered under it.
	tests := []struct {
		name      string
		host      string
		hosts     map[string]HostDefaults
		defaults  *map[string]HostDefaults
		want      HostDefaults
		wantNamed bool
	}{
		{
			name:      "the receiver's own fields",
			host:      "Codeberg",
			hosts:     map[string]HostDefaults{"Codeberg": {URL: "https://codeberg.org"}},
			want:      HostDefaults{URL: "https://codeberg.org"},
			wantNamed: true,
		},
		{
			name:      "the receiver's own URL wins over the defaults",
			host:      "Codeberg",
			hosts:     map[string]HostDefaults{"Codeberg": {URL: "https://codeberg.org"}},
			defaults:  &map[string]HostDefaults{"Codeberg": {URL: "https://git.example.com"}},
			want:      HostDefaults{URL: "https://codeberg.org"},
			wantNamed: true,
		},
		{
			name:      "the receiver's own token wins over the defaults",
			host:      "Codeberg",
			hosts:     map[string]HostDefaults{"Codeberg": {AccessToken: "own"}},
			defaults:  &map[string]HostDefaults{"Codeberg": {AccessToken: "inherited"}},
			want:      HostDefaults{AccessToken: "own"},
			wantNamed: true,
		},
		{
			name:      "unset field comes from the defaults",
			host:      "Codeberg",
			hosts:     map[string]HostDefaults{"Codeberg": {AccessToken: "cb-token"}},
			defaults:  &map[string]HostDefaults{"Codeberg": {URL: "https://codeberg.org", AllowInvalidCerts: new(true)}},
			want:      HostDefaults{URL: "https://codeberg.org", AccessToken: "cb-token", AllowInvalidCerts: new(true)},
			wantNamed: true,
		},
		{
			name:      "allow_invalid_certs=false, so the defaults do not apply",
			host:      "Codeberg",
			hosts:     map[string]HostDefaults{"Codeberg": {AllowInvalidCerts: new(false)}},
			defaults:  &map[string]HostDefaults{"Codeberg": {AllowInvalidCerts: new(true)}},
			want:      HostDefaults{AllowInvalidCerts: new(false)},
			wantNamed: true,
		},
		{
			name:     "no layer names the instance",
			host:     "Codeberg",
			hosts:    map[string]HostDefaults{"Elsewhere": {URL: "https://forge.example.com"}},
			defaults: &map[string]HostDefaults{"Nowhere": {URL: "https://git.example.com"}},
		},
		{
			name:      "an entry matched in both layers, neither giving a URL",
			host:      "Codeberg",
			hosts:     map[string]HostDefaults{"Codeberg": {AccessToken: "cb-token"}},
			defaults:  &map[string]HostDefaults{"Codeberg": {AllowInvalidCerts: new(true)}},
			want:      HostDefaults{AccessToken: "cb-token", AllowInvalidCerts: new(true)},
			wantNamed: true,
		},
		{
			name:      "no defaults",
			host:      "Codeberg",
			hosts:     map[string]HostDefaults{"Codeberg": {AccessToken: "cb-token"}},
			want:      HostDefaults{AccessToken: "cb-token"},
			wantNamed: true,
		},
		{
			name:      "names are matched case-insensitively across layers",
			host:      "cODEBERG",
			hosts:     map[string]HostDefaults{"Codeberg": {AccessToken: "cb-token"}},
			defaults:  &map[string]HostDefaults{"CODEBERG": {URL: "https://codeberg.org"}},
			want:      HostDefaults{URL: "https://codeberg.org", AccessToken: "cb-token"},
			wantNamed: true,
		},
		{
			name:     "an empty host names nothing",
			hosts:    map[string]HostDefaults{"Codeberg": {URL: "https://codeberg.org"}},
			defaults: &map[string]HostDefaults{"Codeberg": {URL: "https://git.example.com"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defaults := Defaults{Host: tc.hosts}
			if tc.defaults != nil {
				defaults.SetDefaults(&Defaults{Host: *tc.defaults})
			}

			// WHEN: resolvedHost is called.
			got, gotNamed := defaults.resolvedHost(tc.host)

			prefix := fmt.Sprintf(
				"%s\nDefaults{host: %q}.resolvedHost()",
				packageName, tc.host,
			)

			// THEN: every field resolves across the layers.
			for _, check := range []struct {
				name, got, want string
			}{
				{
					name: "url",
					got:  got.URL,
					want: tc.want.URL,
				},
				{
					name: "access_token",
					got:  got.AccessToken,
					want: tc.want.AccessToken,
				},
				{
					name: "allow_invalid_certs",
					got:  test.StringifyPtr(got.AllowInvalidCerts),
					want: test.StringifyPtr(tc.want.AllowInvalidCerts),
				},
			} {
				if check.got != check.want {
					t.Errorf(
						"%s %s mismatch\ngot:  %q\nwant: %q",
						prefix, check.name,
						check.got, check.want,
					)
				}
			}

			// AND: whether any layer names the instance is reported correctly.
			if gotNamed != tc.wantNamed {
				t.Errorf(
					"%s named mismatch\ngot:  %t\nwant: %t",
					prefix, gotNamed, tc.wantNamed,
				)
			}
		})
	}
}

func TestDefaults_HostEntry(t *testing.T) {
	// GIVEN: named instances, and a host naming one of them.
	hosts := map[string]HostDefaults{
		"Codeberg": {URL: "https://codeberg.org", AccessToken: "cb-token"},
		"Work":     {URL: "https://git.example.com:8443/forge", AccessToken: "work-token"},
	}
	tests := []struct {
		name  string
		hosts map[string]HostDefaults
		host  string
		want  string
		found bool
	}{
		{
			name:  "by name",
			host:  "Codeberg",
			want:  "cb-token",
			found: true,
		},
		{
			name:  "by name, mixed-case match",
			host:  "cODEBERG",
			want:  "cb-token",
			found: true,
		},
		{
			name: "a URL identifies no entry, even its own",
			host: "https://codeberg.org",
		},
		{
			name: "an unknown name misses",
			host: "Elsewhere",
		},
		{
			name: "an empty host misses",
			host: "",
		},
		{
			name: "the shared URL of several entries identifies none of them",
			hosts: map[string]HostDefaults{
				"Codeberg-Work":     {URL: "https://codeberg.org", AccessToken: "work-token"},
				"Codeberg-Personal": {URL: "https://codeberg.org", AccessToken: "personal-token"},
			},
			host: "https://codeberg.org",
		},
		{
			name: "a URL-named entry, missed by a scheme-less spelling of it",
			hosts: map[string]HostDefaults{
				"https://forgejo.example.com": {URL: "https://forgejo.example.com", AccessToken: "fj-token"},
			},
			host: "forgejo.example.com",
		},
		{
			name: "one of several entries addressing the instance, by name",
			hosts: map[string]HostDefaults{
				"Codeberg-Work":     {URL: "https://codeberg.org", AccessToken: "work-token"},
				"Codeberg-Personal": {URL: "codeberg.org", AccessToken: "personal-token"},
			},
			host:  "Codeberg-Work",
			want:  "work-token",
			found: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			entries := tc.hosts
			if entries == nil {
				entries = hosts
			}
			defaults := Defaults{Host: entries}

			// WHEN: hostEntry is called.
			got, found := defaults.hostEntry(tc.host)

			// THEN: the entry naming that instance is returned, if any.
			if found != tc.found {
				t.Fatalf(
					"%s\nDefaults.hostEntry(%q) found mismatch\ngot:  %t\nwant: %t",
					packageName, tc.host, found, tc.found,
				)
			}
			if got.AccessToken != tc.want {
				t.Fatalf(
					"%s\nDefaults.hostEntry(%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.host, got.AccessToken, tc.want,
				)
			}
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN: named instances.
	tests := []struct {
		name     string
		hosts    map[string]HostDefaults
		defaults map[string]HostDefaults
		errRegex string
	}{
		{
			name:     "no instances",
			errRegex: `^$`,
		},
		{
			name: "valid/multiple",
			hosts: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org", AccessToken: "cb-token"},
				"Work":     {URL: "git.example.com:8443/forge", AllowInvalidCerts: new(true)},
			},
			errRegex: `^$`,
		},
		{
			name: "valid/the URL may come from the environment",
			hosts: map[string]HostDefaults{
				"Work": {URL: "${ARGUS_TEST_FORGEJO_INSTANCE_URL}"},
			},
			errRegex: `^$`,
		},
		{
			name: "valid/the URL may come from the environment and be extended",
			hosts: map[string]HostDefaults{
				"Work": {URL: "${ARGUS_TEST_FORGEJO_INSTANCE_URL}/sub-path"},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid/an unnamed instance",
			hosts: map[string]HostDefaults{
				"  ": {URL: "https://codeberg.org"},
			},
			errRegex: test.TrimYAML(`
				^host:
					an instance must be named$`,
			),
		},
		{
			name: "invalid/a name surrounded by whitespace",
			hosts: map[string]HostDefaults{
				"  Codeberg  ": {URL: "https://codeberg.org"},
			},
			errRegex: `  Codeberg  : <invalid> \(surrounded by whitespace\)`,
		},
		{
			name: "valid/a URL-shaped name, with a URL naming another instance",
			hosts: map[string]HostDefaults{
				"https://foo.com": {URL: "https://bar.com"},
			},
			errRegex: `^$`,
		},
		{
			name: "valid/a name is only a label, however oddly it is spelt",
			hosts: map[string]HostDefaults{
				"https://codeberg.org/": {URL: "https://codeberg.org"},
				"ftp://codeberg.org":    {URL: "https://codeberg.org"},
				"https://":              {URL: "https://codeberg.org"},
				"https://exa mple.com":  {URL: "https://codeberg.org"},
				"forgejo/test":          {URL: "https://codeberg.org"},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid/a URL that is unparseable",
			hosts: map[string]HostDefaults{
				"foo": {URL: "https://exa mple.com"},
			},
			errRegex: test.TrimYAML(`
				^host:
					foo:
						url: "https://exa mple\.com" <invalid> \(not a valid URL\)$`,
			),
		},
		{
			name: "invalid/a URL with an unsupported scheme",
			hosts: map[string]HostDefaults{
				"foo": {URL: "ftp://codeberg.org"},
			},
			errRegex: test.TrimYAML(`
				^host:
					foo:
						url: "ftp://codeberg\.org" <invalid> \(scheme must be http or https\)$`,
			),
		},
		{
			name: "invalid/a URL with no hostname",
			hosts: map[string]HostDefaults{
				"foo": {URL: "https://"},
			},
			errRegex: test.TrimYAML(`
				^host:
					foo:
						url: "https://" <invalid> \(no hostname\)$`,
			),
		},
		{
			name: "valid/a URL may be spelt with a trailing '/'",
			hosts: map[string]HostDefaults{
				"foo": {URL: "FORGE.example.com:443/"},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid/two names differing only by case",
			hosts: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org"},
				"codeberg": {URL: "https://git.example.com"},
			},
			errRegex: test.TrimYAML(`
				^host:
					Codeberg: <invalid> \(already used, names are case-insensitive\)
					codeberg: <invalid> \(already used, names are case-insensitive\)`,
			),
		},
		{
			name:     "valid/an instance with no URL, which the other layer may name",
			hosts:    map[string]HostDefaults{"Codeberg": {AccessToken: "cb-token"}},
			errRegex: `^$`,
		},
		{
			name:     "valid/the URL comes from the defaults",
			hosts:    map[string]HostDefaults{"Codeberg": {AccessToken: "cb-token"}},
			defaults: map[string]HostDefaults{"Codeberg": {URL: "https://codeberg.org"}},
			errRegex: `^$`,
		},
		{
			name:     "invalid/no layer names a URL for the instance",
			hosts:    map[string]HostDefaults{"Codeberg": {AccessToken: "cb-token"}},
			defaults: map[string]HostDefaults{"Elsewhere": {URL: "https://forge.example.com"}},
			errRegex: test.TrimYAML(`
				^host:
					Codeberg:
						url: <required> \(e\.g\. https://codeberg\.org\)$`,
			),
		},
		{
			name:     "valid/fallback URL is not validated by the parent",
			hosts:    map[string]HostDefaults{"Codeberg": {AccessToken: "cb-token"}},
			defaults: map[string]HostDefaults{"Codeberg": {URL: "ftp://codeberg.org"}},
			errRegex: `^$`,
		},
		{
			name:     "invalid/this layer's URL fails validation even if fallback is valid",
			hosts:    map[string]HostDefaults{"Codeberg": {URL: "ftp://codeberg.org"}},
			defaults: map[string]HostDefaults{"Codeberg": {URL: "https://codeberg.org"}},
			errRegex: test.TrimYAML(`
				^host:
					Codeberg:
						url: "ftp://codeberg\.org" <invalid> \(scheme must be http or https\)$`,
			),
		},
		{
			name: "valid/two instances addressing the same URL, each with its own token",
			hosts: map[string]HostDefaults{
				"Codeberg":      {URL: "https://codeberg.org", AccessToken: "cb-personal"},
				"Codeberg-Work": {URL: "CODEBERG.org:443/", AccessToken: "cb-work"},
			},
			errRegex: `^$`,
		},
		{
			name: "valid/a token on an http instance, which only warns",
			hosts: map[string]HostDefaults{
				"Work": {URL: "http://git.example.com", AccessToken: "work-token"},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid/multiple errors",
			hosts: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org"},
				"codeberg": {URL: "https://git.example.com"},
				"  padded": {URL: "https://codeberg.org"},
				"foo":      {URL: "https://exa mple.com"},
			},
			errRegex: test.TrimYAML(`
				^host:
					  padded: <invalid> \(surrounded by whitespace\)
					Codeberg: <invalid> \(already used, names are case-insensitive\)
					codeberg: <invalid> \(already used, names are case-insensitive\)
					foo:
						url: "https://exa mple\.com" <invalid> \(not a valid URL\)`,
			),
		},
	}

	t.Setenv("ARGUS_TEST_FORGEJO_INSTANCE_URL", "https://git.example.com")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defaults := Defaults{Host: tc.hosts}
			if tc.defaults != nil {
				defaults.SetDefaults(&Defaults{Host: tc.defaults})
			}

			// WHEN: CheckValues is called.
			err := defaults.CheckValues()

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s\nDefaults.CheckValues() error mismatch\ngot:  %q\nwant: %q",
					packageName, e, tc.errRegex,
				)
			}
		})
	}
}

func TestDefaults_CheckValues__warnsOnPlaintextToken(t *testing.T) {
	// GIVEN: named instances, each with a token, on http and on https.
	tests := []struct {
		name     string
		hosts    map[string]HostDefaults
		defaults map[string]HostDefaults
		wantWarn bool
	}{
		{
			name: "url=http, token warns",
			hosts: map[string]HostDefaults{
				"Work": {URL: "http://git.example.com", AccessToken: "work-token"},
			},
			wantWarn: true,
		},
		{
			name: "url=https, quiet with token",
			hosts: map[string]HostDefaults{
				"Work": {URL: "https://git.example.com", AccessToken: "work-token"},
			},
			wantWarn: false,
		},
		{
			name: "url=http, no token is quiet",
			hosts: map[string]HostDefaults{
				"Work": {URL: "http://git.example.com"},
			},
			wantWarn: false,
		},
		{
			name:     "url=http here, token in the fallback layer",
			hosts:    map[string]HostDefaults{"Work": {URL: "http://git.example.com"}},
			defaults: map[string]HostDefaults{"Work": {AccessToken: "work-token"}},
			wantWarn: true,
		},
		{
			name:     "token here, url=http in the fallback layer",
			hosts:    map[string]HostDefaults{"Work": {AccessToken: "work-token"}},
			defaults: map[string]HostDefaults{"Work": {URL: "http://git.example.com"}},
			wantWarn: true,
		},
		{
			name:     "url=https in the fallback layer is quiet",
			hosts:    map[string]HostDefaults{"Work": {AccessToken: "work-token"}},
			defaults: map[string]HostDefaults{"Work": {URL: "https://git.example.com"}},
			wantWarn: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			defaults := Defaults{Host: tc.hosts}
			if tc.defaults != nil {
				defaults.SetDefaults(&Defaults{Host: tc.defaults})
			}

			// WHEN: CheckValues is called.
			err := defaults.CheckValues()

			// THEN: the plaintext-credential warning is logged only when expected.
			logged := releaseStdout()
			if gotWarn := strings.Contains(
				logged, "access_token will be sent unencrypted",
			); gotWarn != tc.wantWarn {
				t.Fatalf(
					"%s\nDefaults.CheckValues() warn mismatch\ngot:  %t\nwant: %t\nlog:  %q",
					packageName, gotWarn, tc.wantWarn, logged,
				)
			}

			// AND: no error is returned.
			if err != nil {
				t.Fatalf(
					"%s\nDefaults.CheckValues() unexpected error: %v",
					packageName, err,
				)
			}
		})
	}
}

func TestDefaults_Unmarshal(t *testing.T) {
	// GIVEN: a defaults document holding common values and named instances.
	tests := []struct {
		name     string
		yaml     string
		json     string
		want     Defaults
		errRegex string
	}{
		{
			name:     "empty",
			yaml:     ``,
			json:     `{}`,
			errRegex: `^$`,
		},
		{
			name: "common only",
			yaml: `
				common:
					use_prerelease: true`,
			json:     `{"common": {"use_prerelease": true}}`,
			want:     Defaults{Common: CommonDefaults{UsePreRelease: new(true)}},
			errRegex: `^$`,
		},
		{
			name: "instances only",
			yaml: `
				host:
					Codeberg:
						url: https://codeberg.org
						access_token: cb-token`,
			json: `{
				"host": {
					"Codeberg": {"url": "https://codeberg.org", "access_token": "cb-token"}
				}
			}`,
			want: Defaults{Host: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org", AccessToken: "cb-token"},
			}},
			errRegex: `^$`,
		},
		{
			name: "common and instances together",
			yaml: `
				common:
					use_prerelease: false
				host:
					Codeberg:
						url: https://codeberg.org
					Work:
						url: https://git.example.com
						allow_invalid_certs: true`,
			json: `{
				"common": {"use_prerelease": false},
				"host": {
					"Codeberg": {"url": "https://codeberg.org"},
					"Work": {"url": "https://git.example.com", "allow_invalid_certs": true}
				}
			}`,
			want: Defaults{
				Common: CommonDefaults{UsePreRelease: new(false)},
				Host: map[string]HostDefaults{
					"Codeberg": {URL: "https://codeberg.org"},
					"Work":     {URL: "https://git.example.com", AllowInvalidCerts: new(true)},
				},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid/use_prerelease type",
			yaml: `
				common:
					use_prerelease: [1]`,
			json:     `{"common": {"use_prerelease": [1]}}`,
			errRegex: `.* unmarshal`,
		},
		{
			name:     "invalid/list rather than map",
			yaml:     `- common: {}`,
			json:     `[{"common": {}}]`,
			errRegex: `(sequence was used where mapping is expected|.* unmarshal)`,
		},
	}

	for _, tc := range tests {
		for _, format := range []string{"json", "yaml"} {
			t.Run(tc.name+"/"+format, func(t *testing.T) {
				t.Parallel()

				var data string
				if format == "yaml" {
					data = test.TrimYAML(tc.yaml)
				} else {
					data = test.TrimJSON(tc.json)
				}

				// WHEN: it is unmarshalled.
				var got Defaults
				err := decode.Unmarshal(format, []byte(data), &got)

				prefix := fmt.Sprintf(
					"%s\nDefaults.Unmarshal(format=%q, data=%q)",
					packageName, format, data,
				)

				// THEN: any error is as expected.
				e := errfmt.FormatError(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Fatalf("%s error mismatch\ngot:  %q\nwant: %q", prefix, e, tc.errRegex)
				}
				if tc.errRegex != `^$` {
					return
				}

				// AND: the fields are as expected.
				gotStr := decode.ToYAMLString(got, "")
				wantStr := decode.ToYAMLString(tc.want, "")
				if gotStr != wantStr {
					t.Fatalf("%s mismatch\ngot:  %q\nwant: %q", prefix, gotStr, wantStr)
				}
			})
		}
	}
}

func TestDefaults_Marshal(t *testing.T) {
	// GIVEN: a Defaults holding common values and named instances.
	defaults := Defaults{
		Common: CommonDefaults{UsePreRelease: new(true)},
		Host: map[string]HostDefaults{
			"Codeberg": {URL: "https://codeberg.org", AccessToken: "cb-token"},
			"Work":     {URL: "https://git.example.com", AllowInvalidCerts: new(true)},
		},
	}

	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			// WHEN: it is marshalled and unmarshalled again.
			data, err := decode.Marshal(format, defaults)
			if err != nil {
				t.Fatalf("%s\nDefaults.Marshal(%q) failed: %v", packageName, format, err)
			}
			var got Defaults
			if err := decode.Unmarshal(format, data, &got); err != nil {
				t.Fatalf("%s\nDefaults.Unmarshal(%q, %q) failed: %v",
					packageName, format, string(data), err)
			}

			// THEN: nothing is lost on the way round.
			gotStr := decode.ToYAMLString(got, "")
			wantStr := decode.ToYAMLString(defaults, "")
			if gotStr != wantStr {
				t.Fatalf(
					"%s\nDefaults round-trip (%s) mismatch\ngot:  %q\nwant: %q",
					packageName, format, gotStr, wantStr,
				)
			}
		})
	}
}
