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

	"github.com/release-argus/Argus/config/decode"
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
			name: "valid/two instances addressing the same URL, each with its own token",
			hosts: map[string]HostDefaults{
				"Codeberg":      {URL: "https://codeberg.org", AccessToken: "cb-personal"},
				"Codeberg-Work": {URL: "CODEBERG.org:443/", AccessToken: "cb-work"},
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
