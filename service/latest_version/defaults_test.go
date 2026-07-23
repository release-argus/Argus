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

package latestver

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
)

func TestDecodeDefaults(t *testing.T) {
	// GIVEN: data in a given format to Decode into Defaults.
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:   "JSON/empty",
			format: "json",
			data:   "",
			want:   "",
			errRegex: test.TrimYAML(`
				latest_version:
					jsontext:
						unexpected EOF$`,
			),
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "YAML/invalid data types",
			format: "yaml",
			data:   `type: ['github']`,
			errRegex: test.TrimYAML(`
					^latest_version:
						[^\s]+ cannot unmarshal .* into Go struct field Defaults.Type of type string
						[^\s]+.*
						\s+\^$`,
			),
		},
		{
			name:   "JSON/full",
			format: "json",
			data: test.TrimJSON(`{
				"type": "github",
				"common": {
					"require": {
						"docker": {
							"type": "hub",
							"tag": "t"
						}
					}
				},
				"github": {
					"access_token": "foo",
					"use_prerelease": true
				},
				"url": {
					"allow_invalid_certs": true
				}
			}`),
			want: test.TrimYAML(`
				type: github
				common:
					require:
						docker:
							type: hub
							tag: t
				github:
					access_token: foo
					use_prerelease: true
				url:
					allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/full",
			format: "yaml",
			data: test.TrimYAML(`
				type: github
				common:
					require:
						docker:
							type: hub
							tag: t
				github:
					access_token: foo
					use_prerelease: true
				url:
					allow_invalid_certs: true
			`),
			want: test.TrimYAML(`
				type: github
				common:
					require:
						docker:
							type: hub
							tag: t
				github:
					access_token: foo
					use_prerelease: true
				url:
					allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/deprecated fields decode",
			format: "yaml",
			data: test.TrimYAML(`
				access_token: foo
				use_prerelease: true
				allow_invalid_certs: true
				require:
					docker:
						type: hub
						tag: t
			`),
			want: test.TrimYAML(`
				access_token: foo
				use_prerelease: true
				allow_invalid_certs: true
				require:
					docker:
						type: hub
						tag: t
			`),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				DecodeDefaults,
				tc.format, tc.data,
				func(v *Defaults) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name string
		data *Defaults
		want bool
	}{
		{
			name: "empty",
			data: &Defaults{},
			want: true,
		},
		{
			name: "non-empty/Type",
			data: &Defaults{Type: "github"},
			want: false,
		},
		{
			name: "non-empty/GitHub",
			data: &Defaults{GitHub: github.Defaults{AccessToken: "foo"}},
			want: false,
		},
		{
			name: "non-empty/URL",
			data: &Defaults{URL: web.Defaults{AllowInvalidCerts: new(true)}},
			want: false,
		},
		{
			name: "non-empty/AccessTokenDeprecated",
			data: &Defaults{AccessTokenDeprecated: "foo"},
			want: false,
		},
		{
			name: "non-empty/UsePreReleaseDeprecated",
			data: &Defaults{UsePreReleaseDeprecated: new(true)},
			want: false,
		},
		{
			name: "non-empty/AllowInvalidCertsDeprecated",
			data: &Defaults{AllowInvalidCertsDeprecated: new(true)},
			want: false,
		},
		{
			name: "non-empty/RequireDeprecated",
			data: &Defaults{RequireDeprecated: &filter.RequireDefaults{
				Docker: docker.Defaults{Type: "hub"},
			}},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &Defaults{
				Type:                        "github",
				GitHub:                      github.Defaults{AccessToken: "foo"},
				URL:                         web.Defaults{AllowInvalidCerts: new(true)},
				AccessTokenDeprecated:       "foo",
				UsePreReleaseDeprecated:     new(true),
				AllowInvalidCertsDeprecated: new(true),
				RequireDeprecated: &filter.RequireDefaults{
					Docker: docker.Defaults{Type: "hub"},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called.
			got := tc.data.IsZero()

			// THEN: the result is expected.
			if got != tc.want {
				t.Errorf(
					"%s\nDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDefaults_Default(t *testing.T) {
	// GIVEN: a Defaults.
	defaults := Defaults{}
	want := Defaults{
		Type: "github",
		Common: base.Defaults{
			Require: filter.RequireDefaults{
				Docker: docker.Defaults{
					Type: "hub",
					ContainerDetailDefaults: docker.ContainerDetailDefaults{
						Tag: "{{ version }}",
					},
				},
			},
		},
		GitHub: github.Defaults{
			UsePreRelease: new(false),
		},
		URL: web.Defaults{
			AllowInvalidCerts: new(false),
		},
	}

	// WHEN: Default is called.
	defaults.Default()

	wantStr := decode.ToYAMLString(want, "")
	gotStr := decode.ToYAMLString(defaults, "")
	// THEN: it should set the defaults as expected - common (Require), and every
	// registered type's own defaults.
	if gotStr != wantStr {
		t.Errorf(
			"%s\nDefaults.Default() value mismatch\ngot:  %q\nwant: %q",
			packageName, gotStr, wantStr,
		)
	}
}

func TestDefaults_SetDefaults(t *testing.T) {
	// GIVEN: Two sets of Defaults.
	d := &Defaults{}
	hd := &Defaults{}
	hd.Default()

	// WHEN: SetDefaults is called on defaults with these other defaults.
	d.SetDefaults(hd)

	// THEN: the common Require defaults chain to the hard defaults' Require.
	if d.Common.Require.Docker.Defaults != &hd.Common.Require.Docker {
		t.Errorf(
			"%s\nDefaults.SetDefaults() pointer mismatch\ngot:  %p\nwant: %p",
			packageName, d.Common.Require.Docker.Defaults, &hd.Common.Require.Docker,
		)
	}
}

func TestDefaults_MigrateDeprecated(t *testing.T) {
	// GIVEN: a Defaults, possibly with deprecated fields set.
	tests := []struct {
		name                 string
		input                *Defaults
		wantAccessToken      string
		wantUsePreRelease    *bool
		wantAllowInvalidCert *bool
		wantRequire          *filter.RequireDefaults
	}{
		{
			name: "access_token migrated to github.access_token",
			input: &Defaults{
				AccessTokenDeprecated: "foo",
			},
			wantAccessToken: "foo",
		},
		{
			name: "access_token does not override an explicit github.access_token",
			input: &Defaults{
				AccessTokenDeprecated: "old",
				GitHub:                github.Defaults{AccessToken: "new"},
			},
			wantAccessToken: "new",
		},
		{
			name: "use_prerelease migrated to github.use_prerelease",
			input: &Defaults{
				UsePreReleaseDeprecated: new(true),
			},
			wantUsePreRelease: new(true),
		},
		{
			name: "use_prerelease does not override an explicit github.use_prerelease",
			input: &Defaults{
				UsePreReleaseDeprecated: new(true),
				GitHub: github.Defaults{
					UsePreRelease: new(false),
				},
			},
			wantUsePreRelease: new(false),
		},
		{
			name: "allow_invalid_certs migrated to url.allow_invalid_certs",
			input: &Defaults{
				AllowInvalidCertsDeprecated: new(true),
			},
			wantAllowInvalidCert: new(true),
		},
		{
			name: "allow_invalid_certs does not override an explicit url.allow_invalid_certs",
			input: &Defaults{
				AllowInvalidCertsDeprecated: new(true),
				URL:                         web.Defaults{AllowInvalidCerts: new(false)},
			},
			wantAllowInvalidCert: new(false),
		},
		{
			name: "require migrated to common.require",
			input: &Defaults{
				RequireDeprecated: &filter.RequireDefaults{
					Docker: docker.Defaults{
						Type: "hub",
					},
				},
			},
			wantRequire: &filter.RequireDefaults{
				Docker: docker.Defaults{
					Type: "hub",
				},
			},
		},
		{
			name: "require does not override an explicit common.require",
			input: &Defaults{
				RequireDeprecated: &filter.RequireDefaults{
					Docker: docker.Defaults{
						Type: "hub",
					},
				},
				Common: base.Defaults{
					Require: filter.RequireDefaults{
						Docker: docker.Defaults{
							Type: "ghcr",
						},
					},
				},
			},
			wantRequire: &filter.RequireDefaults{
				Docker: docker.Defaults{
					Type: "ghcr",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf("%s\nDefaults.MigrateDeprecated()", packageName)

			// WHEN: MigrateDeprecated is called.
			tc.input.MigrateDeprecated()

			// THEN: the deprecated fields were migrated to their new home.
			if tc.wantAccessToken != "" && tc.input.GitHub.AccessToken != tc.wantAccessToken {
				t.Errorf(
					"%s GitHub.AccessToken mismatch\ngot:  %q\nwant: %q",
					prefix, tc.input.GitHub.AccessToken, tc.wantAccessToken,
				)
			}
			if tc.wantUsePreRelease != nil {
				if got, want := test.StringifyPtr(tc.input.GitHub.UsePreRelease), test.StringifyPtr(tc.wantUsePreRelease); got != want {
					t.Errorf(
						"%s GitHub.UsePreRelease mismatch\ngot:  %v\nwant: %v",
						prefix, got, want,
					)
				}
			}
			if tc.wantAllowInvalidCert != nil {
				if got, want := test.StringifyPtr(tc.input.URL.AllowInvalidCerts), test.StringifyPtr(tc.wantAllowInvalidCert); got != want {
					t.Errorf(
						"%s URL.AllowInvalidCerts mismatch\ngot:  %v\nwant: %v",
						prefix, got, want,
					)
				}
			}
			if tc.wantRequire != nil {
				if got, want := tc.input.Common.Require.Docker.String(""), tc.wantRequire.Docker.String(""); got != want {
					t.Errorf(
						"%s Require mismatch\ngot:  %v\nwant: %v",
						prefix, got, want,
					)
				}
			}

			// AND: the deprecated fields were cleared.
			if tc.input.AccessTokenDeprecated != "" {
				t.Errorf(
					"%s AccessTokenDeprecated not cleared: %q",
					prefix, tc.input.AccessTokenDeprecated)
			}
			if tc.input.UsePreReleaseDeprecated != nil {
				t.Errorf(
					"%s UsePreReleaseDeprecated not cleared: %v",
					prefix, tc.input.UsePreReleaseDeprecated)
			}
			if tc.input.AllowInvalidCertsDeprecated != nil {
				t.Errorf(
					"%s AllowInvalidCertsDeprecated not cleared: %v",
					prefix, tc.input.AllowInvalidCertsDeprecated)
			}
			if tc.input.RequireDeprecated != nil {
				t.Errorf(
					"%s RequireDeprecated not cleared: %v",
					prefix, tc.input.RequireDeprecated)
			}
		})
	}
}

func TestDefaults_MigrateDeprecated__PreservesDefaultsChainForSetDefaults(t *testing.T) {
	// GIVEN: HardDefaults with the Docker tag template set, and the deprecated
	// 'latest_version.require' key still in its old (pre-'common.require')
	// spot, as a user upgrading from an old config would have:
	hd := &Defaults{}
	hd.Default()

	d := &Defaults{
		RequireDeprecated: &filter.RequireDefaults{
			Docker: docker.Defaults{
				Type: "quay",
			},
		},
	}

	// WHEN: MigrateDeprecated runs on it and SetDefaults wires the (now-migrated) chain.
	d.MigrateDeprecated()
	d.SetDefaults(hd)

	prefix := fmt.Sprintf("%s\nDefaults.MigrateDeprecated()+SetDefaults()", packageName)

	// THEN: the migrated Common.Require.Docker must chain to the hard defaults.
	if d.Common.Require.Docker.Defaults != &hd.Common.Require.Docker {
		t.Errorf(
			"%s broke the Require defaults chain\ngot:  %p\nwant: %p",
			prefix, d.Common.Require.Docker.Defaults, &hd.Common.Require.Docker,
		)
	}
	want := hd.Common.Require.Docker.Tag
	if want == "" {
		t.Errorf("%s HardDefaults.Common.Require.Docker.Tag should be set! got an empty string", prefix)
	}
	if got := d.Common.Require.Docker.GetTag(); got != want {
		t.Errorf(
			"%s lost the hard-default Docker tag template\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}

	// AND: the migrated Type ('quay', not the hard default 'hub') must be
	// what a per-service require.docker inherits when it omits its own type.
	if got, want := d.Common.Require.Docker.GetType(), "quay"; got != want {
		t.Errorf(
			"%s lost the migrated Docker type\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}
}

func TestApplyTypeDefaults(t *testing.T) {
	// GIVEN: a DefaultsConfig.
	cfg := DefaultsConfig{
		Soft: &Defaults{
			GitHub: github.Defaults{AccessToken: "soft-token"},
			URL:    web.Defaults{AllowInvalidCerts: new(true)},
		},
		Hard: &Defaults{
			GitHub: github.Defaults{AccessToken: "hard-token"},
			URL:    web.Defaults{AllowInvalidCerts: new(false)},
		},
	}

	t.Run("github.Lookup gets the GitHub-specific defaults", func(t *testing.T) {
		// AND: a Lookup of type github.Lookup.
		lookup := &github.Lookup{}

		// WHEN: applyTypeDefaults is called.
		applyTypeDefaults(lookup, cfg)

		// THEN: the receiver's type defaults point at the given GitHub defaults.
		gotSoft, gotHard := lookup.GetTypeDefaults()
		if gotSoft != &cfg.Soft.GitHub {
			t.Errorf(
				"%s\napplyTypeDefaults() Soft pointer mismatch\ngot:  %p\nwant: %p",
				packageName, gotSoft, &cfg.Soft.GitHub,
			)
		}
		if gotHard != &cfg.Hard.GitHub {
			t.Errorf(
				"%s\napplyTypeDefaults() Hard pointer mismatch\ngot:  %p\nwant: %p",
				packageName, gotHard, &cfg.Hard.GitHub,
			)
		}
	})

	t.Run("web.Lookup gets the URL-specific defaults", func(t *testing.T) {
		// AND: a Lookup of type web.Lookup.
		lookup := &web.Lookup{}

		// WHEN: applyTypeDefaults is called.
		applyTypeDefaults(lookup, cfg)

		// THEN: the receiver's type defaults point at the given URL defaults.
		gotSoft, gotHard := lookup.GetTypeDefaults()
		if gotSoft != &cfg.Soft.URL {
			t.Errorf(
				"%s\napplyTypeDefaults() Soft pointer mismatch\ngot:  %p\nwant: %p",
				packageName, gotSoft, &cfg.Soft.URL,
			)
		}
		if gotHard != &cfg.Hard.URL {
			t.Errorf(
				"%s\napplyTypeDefaults() Hard pointer mismatch\ngot:  %p\nwant: %p",
				packageName, gotHard, &cfg.Hard.URL,
			)
		}
	})

	t.Run("unregistered type is a no-op", func(t *testing.T) {
		// AND: an unregistered Lookup type.
		lookup := &mockLookup{}

		// WHEN: applyTypeDefaults is called.
		// THEN: it must not panic for a type with no SetTypeDefaults dispatch.
		applyTypeDefaults(lookup, cfg)
	})
}
