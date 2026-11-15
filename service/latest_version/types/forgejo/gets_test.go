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
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
)

func TestLookup_GetType(t *testing.T) {
	l := &Lookup{}
	l.Type = "other"

	got := l.GetType()

	if got != Type {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, Type)
	}
}

func TestLookup_ServiceURL(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		url      string
		defaults map[string]HostDefaults
		want     string
	}{
		{
			name: "scheme+host and repository",
			host: "https://codeberg.org",
			url:  "owner/repo",
			want: "https://codeberg.org/owner/repo",
		},
		{
			name: "host and repository",
			host: "codeberg.org",
			url:  "owner/repo",
			want: "https://codeberg.org/owner/repo",
		},
		{
			name: "scheme+host+sub-path and repository",
			host: "https://example.com/git",
			url:  "owner/repo",
			want: "https://example.com/git/owner/repo",
		},
		{
			name: "scheme+host+port and repository",
			host: "http://forge.example.com:3000",
			url:  "owner/repo",
			want: "http://forge.example.com:3000/owner/repo",
		},
		{
			name: "scheme+host+port+sub-path and repository",
			host: "http://forge.example.com:3000/sub/path",
			url:  "owner/repo",
			want: "http://forge.example.com:3000/sub/path/owner/repo",
		},
		{
			name: "no host",
			host: "",
			url:  "owner/repo",
			want: "",
		},
		{
			name: "a host naming an instance resolves to its URL",
			host: "Codeberg",
			url:  "owner/repo",
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org"},
			},
			want: "https://codeberg.org/owner/repo",
		},
		{
			name: "a name matched case-insensitively still resolves",
			host: "cODEBERG",
			url:  "owner/repo",
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org"},
			},
			want: "https://codeberg.org/owner/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := Lookup{Host: tc.host}
			lookup.URL = tc.url
			lookup.SetTypeDefaults(
				&Defaults{Host: tc.defaults},
				&Defaults{},
			)

			got := lookup.ServiceURL()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestLookup_HostDefaults(t *testing.T) {
	tests := []struct {
		name                   string
		host                   string
		defaults, hardDefaults map[string]HostDefaults
		want                   HostDefaults
		wantNamed              bool
	}{
		{
			name: "AccessToken=defaults, AllowInvalidCerts=defaults",
			host: "Codeberg",
			defaults: map[string]HostDefaults{
				"Codeberg": {
					URL:               "https://codeberg.org",
					AccessToken:       "defaults",
					AllowInvalidCerts: new(false),
				},
			},
			hardDefaults: map[string]HostDefaults{
				"Codeberg": {
					URL:               "CODEBERG.org:443",
					AccessToken:       "hardDefaults",
					AllowInvalidCerts: new(true),
				},
			},
			want: HostDefaults{
				URL:               "https://codeberg.org",
				AccessToken:       "defaults",
				AllowInvalidCerts: new(false),
			},
			wantNamed: true,
		},
		{
			name: "AccessToken=defaults, AllowInvalidCerts=hardDefaults",
			host: "Codeberg",
			defaults: map[string]HostDefaults{
				"Codeberg": {
					URL:         "https://codeberg.org",
					AccessToken: "defaults",
				},
			},
			hardDefaults: map[string]HostDefaults{
				"Codeberg": {
					URL:               "CODEBERG.org:443",
					AccessToken:       "hardDefaults",
					AllowInvalidCerts: new(true),
				},
			},
			want: HostDefaults{
				URL:               "https://codeberg.org",
				AccessToken:       "defaults",
				AllowInvalidCerts: new(true),
			},
			wantNamed: true,
		},
		{
			name: "AccessToken=hardDefaults, AllowInvalidCerts=defaults",
			host: "Codeberg",
			defaults: map[string]HostDefaults{
				"Codeberg": {
					URL:               "https://codeberg.org",
					AllowInvalidCerts: new(false),
				},
			},
			hardDefaults: map[string]HostDefaults{
				"Codeberg": {
					URL:               "CODEBERG.org:443",
					AccessToken:       "hardDefaults",
					AllowInvalidCerts: new(true),
				},
			},
			want: HostDefaults{
				URL:               "https://codeberg.org",
				AccessToken:       "hardDefaults",
				AllowInvalidCerts: new(false),
			},
			wantNamed: true,
		},
		{
			name: "AccessToken=hardDefaults, AllowInvalidCerts=hardDefaults",
			host: "Codeberg",
			defaults: map[string]HostDefaults{
				"Elsewhere": {
					URL:         "https://forge.example.com",
					AccessToken: "defaults",
				},
			},
			hardDefaults: map[string]HostDefaults{
				"Codeberg": {
					URL:               "https://codeberg.org",
					AccessToken:       "hardDefaults",
					AllowInvalidCerts: new(true),
				},
			},
			want: HostDefaults{
				URL:               "https://codeberg.org",
				AccessToken:       "hardDefaults",
				AllowInvalidCerts: new(true),
			},
			wantNamed: true,
		},
		{
			name: "no matches",
			host: "https://git.internal.corp",
			defaults: map[string]HostDefaults{
				"https://codeberg.org": {
					AccessToken: "defaults"},
			},
			hardDefaults: map[string]HostDefaults{
				"https://forge.example.com": {
					AccessToken: "hardDefaults"},
			},
		},
		{
			name: "no entries",
			host: "Codeberg",
		},
		{
			name: "a URL identifies no entry, even its own",
			host: "https://codeberg.org",
			defaults: map[string]HostDefaults{
				"Codeberg-Work":     {URL: "https://codeberg.org", AccessToken: "work-token"},
				"Codeberg-Personal": {URL: "https://codeberg.org", AccessToken: "personal-token"},
			},
		},
		{
			name: "a name in another case finds the instance",
			host: "cODEBERG",
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org", AccessToken: "cb-token"},
			},
			want:      HostDefaults{URL: "https://codeberg.org", AccessToken: "cb-token"},
			wantNamed: true,
		},
		{
			name: "a matched entry with no fields is still named",
			host: "Codeberg",
			defaults: map[string]HostDefaults{
				"Codeberg": {},
			},
			wantNamed: true,
		},
		{
			name: "an empty host matches nothing",
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org", AccessToken: "cb-token"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{Host: tc.host}
			typeDefaults := &Defaults{Host: tc.defaults}
			typeHardDefaults := &Defaults{Host: tc.hardDefaults}
			typeDefaults.SetDefaults(typeHardDefaults)
			lookup.SetTypeDefaults(typeDefaults, typeHardDefaults)

			got, gotNamed := lookup.hostDefaults()

			if gotNamed != tc.wantNamed {
				t.Fatalf("named mismatch\ngot:  %t\nwant: %t", gotNamed, tc.wantNamed)
			}

			gotStr := decode.ToYAMLString(got, "")
			wantStr := decode.ToYAMLString(tc.want, "")
			if gotStr != wantStr {
				t.Fatalf("HostDefaults mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}
		})
	}
}

func TestLookup_HostDefaults__EnvVarName(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		envValue  string
		wantToken string
		wantNamed bool
	}{
		{
			name:      "an expanded name addresses the entry",
			host:      "${ARGUS_TEST_FORGEJO_HOST_NAME}",
			envValue:  "Codeberg",
			wantToken: "cb-token",
			wantNamed: true,
		},
		{
			name:     "an expanded name that addresses nothing",
			host:     "${ARGUS_TEST_FORGEJO_HOST_NAME}",
			envValue: "Elsewhere",
		},
		{
			name: "an unset variable stays literal, and names nothing",
			host: "${ARGUS_TEST_FORGEJO_HOST_NAME_UNSET}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars.

			if tc.envValue != "" {
				t.Setenv("ARGUS_TEST_FORGEJO_HOST_NAME", tc.envValue)
			}

			lookup := &Lookup{Host: tc.host}
			lookup.SetTypeDefaults(
				&Defaults{Host: map[string]HostDefaults{
					"Codeberg": {URL: "https://codeberg.org", AccessToken: "cb-token"},
				}},
				&Defaults{},
			)

			got, gotNamed := lookup.hostDefaults()

			if gotNamed != tc.wantNamed {
				t.Fatalf("named mismatch\ngot:  %t\nwant: %t", gotNamed, tc.wantNamed)
			}
			if got.AccessToken != tc.wantToken {
				t.Fatalf(
					"HostDefaults.AccessToken mismatch\ngot:  %q\nwant: %q",
					got.AccessToken, tc.wantToken,
				)
			}
		})
	}
}

func TestLookup_AccessToken(t *testing.T) {
	tests := []struct {
		name                   string
		host                   string
		serviceValue           string
		defaults, hardDefaults map[string]HostDefaults
		want                   string
	}{
		{
			name:         "root overrides all",
			host:         "Codeberg",
			serviceValue: "service-token",
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org", AccessToken: "defaults-token"},
			},
			want: "service-token",
		},
		{
			name: "the named host entry is used",
			host: "Codeberg",
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org", AccessToken: "defaults-token"},
			},
			want: "defaults-token",
		},
		{
			name: "a token held only by the hardDefaults layer",
			host: "Codeberg",
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org"},
			},
			hardDefaults: map[string]HostDefaults{
				"Codeberg": {AccessToken: "hard-token"},
			},
			want: "hard-token",
		},
		{
			name: "unknown host gets no token",
			host: "https://forge.example.com",
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org", AccessToken: "defaults-token"},
			},
			hardDefaults: map[string]HostDefaults{
				"Work": {URL: "https://git.example.com", AccessToken: "hard-defaults-token"},
			},
			want: "",
		},
		{
			name: "empty host gets no token",
			host: "",
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org", AccessToken: "defaults-token"},
			},
			want: "",
		},
		{
			name: "environment variables are expanded",
			host: "Codeberg",
			defaults: map[string]HostDefaults{
				"Codeberg": {URL: "https://codeberg.org", AccessToken: "${ARGUS_TEST_FORGEJO_TOKEN_FROM_ENV}"},
			},
			want: "token-from-env",
		},
		{
			name: "a scheme-less spelling of a URL-named entry gets no token",
			host: "forgejo.example.com",
			defaults: map[string]HostDefaults{
				"https://forgejo.example.com": {URL: "https://forgejo.example.com", AccessToken: "fj-token"},
			},
			want: "",
		},
	}

	t.Setenv("ARGUS_TEST_FORGEJO_TOKEN_FROM_ENV", "token-from-env")

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{
				Host:        tc.host,
				AccessToken: tc.serviceValue,
			}
			typeDefaults := &Defaults{Host: tc.defaults}
			typeHardDefaults := &Defaults{Host: tc.hardDefaults}
			typeDefaults.SetDefaults(typeHardDefaults)
			lookup.SetTypeDefaults(typeDefaults, typeHardDefaults)

			got := lookup.accessToken()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestLookup_AllowInvalidCerts(t *testing.T) {
	tests := []struct {
		name                   string
		host                   string
		serviceValue           *bool
		defaults, hardDefaults map[string]HostDefaults
		want                   bool
	}{
		{
			name:         "root overrides all",
			host:         "https://git.example.com",
			serviceValue: new(false),
			defaults: map[string]HostDefaults{
				"https://git.example.com": {AllowInvalidCerts: new(true)},
			},
			want: false,
		},
		{
			name: "matching host entry is used",
			host: "https://git.example.com",
			defaults: map[string]HostDefaults{
				"https://git.example.com": {AllowInvalidCerts: new(true)},
			},
			want: true,
		},
		{
			name: "unknown host has certificates enforced",
			host: "https://codeberg.org",
			defaults: map[string]HostDefaults{
				"https://git.example.com": {AllowInvalidCerts: new(true)},
			},
			want: false,
		},
		{
			name: "matching host entry held only by the hardDefaults",
			host: "https://git.example.com",
			hardDefaults: map[string]HostDefaults{
				"https://git.example.com": {AllowInvalidCerts: new(true)},
			},
			want: true,
		},
		{
			name: "no defaults",
			host: "https://git.example.com",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{
				Host:              tc.host,
				AllowInvalidCerts: tc.serviceValue,
			}
			typeDefaults := &Defaults{Host: tc.defaults}
			typeHardDefaults := &Defaults{Host: tc.hardDefaults}
			typeDefaults.SetDefaults(typeHardDefaults)
			lookup.SetTypeDefaults(typeDefaults, typeHardDefaults)

			got := lookup.allowInvalidCerts()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestLookup_UsePreRelease(t *testing.T) {
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      bool
	}{
		{
			name:             "root overrides all",
			rootValue:        new(false),
			defaultValue:     new(true),
			hardDefaultValue: new(true),
			want:             false,
		},
		{
			name:             "default overrides hardDefault",
			defaultValue:     new(true),
			hardDefaultValue: new(false),
			want:             true,
		},
		{
			name:             "hardDefault is last resort",
			hardDefaultValue: new(false),
			want:             false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{UsePreRelease: tc.rootValue}
			lookup.SetTypeDefaults(
				&Defaults{Common: CommonDefaults{UsePreRelease: tc.defaultValue}},
				&Defaults{Common: CommonDefaults{UsePreRelease: tc.hardDefaultValue}},
			)

			got := lookup.usePreRelease()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestLookup_UseTagsAPI(t *testing.T) {
	tests := []struct {
		name        string
		requireYAML string
		want        bool
	}{
		{
			name: "usable/Require=nil",
			want: true,
		},
		{
			name:        "unusable/regex_content set",
			requireYAML: "regex_content: 'Argus-{{ version }}.linux-amd64'",
			want:        false, // regex_content needs release assets.
		},
		{
			name:        "usable/Require without regex_content",
			requireYAML: `regex_version: '^[0-9.]+$'`,
			want:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookupYAML := test.TrimYAML(`
				host: https://codeberg.org
				url: owner/repo`,
			)
			if tc.requireYAML != "" {
				lookupYAML += "\nrequire:\n  " + tc.requireYAML
			}
			lookup := testLookup(t, lookupYAML)

			got := lookup.useTagsAPI()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}
