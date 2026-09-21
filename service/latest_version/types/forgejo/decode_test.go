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
	"github.com/release-argus/Argus/service/latest_version/filter"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_DecodeSelf(t *testing.T) {
	dvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	// GIVEN: data in a given format to Decode into an existing Lookup.
	tests := []struct {
		name         string
		format, data string
		errRegex     string
		want         string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     ``,
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     `{}`,
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     ``,
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid payload decode error",
			format:   "json",
			data:     `{`,
			errRegex: `unexpected`,
		},
		{
			name:   "JSON/valid payload, no require",
			format: "json",
			data: test.TrimJSON(`{
				"host": "example.com",
				"url": "owner/repo",
				"allow_invalid_certs": true,
				"use_prerelease": false
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				host: example.com
				url: owner/repo
				allow_invalid_certs: true
				use_prerelease: false
			`),
		},
		{
			name:     "JSON/invalid data types",
			format:   "json",
			data:     `{"allow_invalid_certs": "true"}`,
			errRegex: `^json: .*unmarshal.*$`,
			want:     "type: url\n",
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				host: example.com
				url: owner/repo
				access_token: foo
				allow_invalid_certs: true
				use_prerelease: false
				url_commands:
					- type: regex
						regex: '.*'
				require:
					regex_content: '.*'
			`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				host: example.com
				url: owner/repo
				url_commands:
					- type: regex
						regex: .*
				require:
					regex_content: .*
				access_token: foo
				allow_invalid_certs: true
				use_prerelease: false
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Lookup.
			options := opttest.PlainOptions(t, optCfg)
			svcStatus := &status.Status{}
			lookup := &Lookup{}
			lookup.Init(
				options,
				svcStatus,
				dvCfg,
			)

			// WHEN: DecodeSelf is called.
			lookup, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					err := lookup.DecodeSelf(format, data)
					return lookup, err
				},
				tc.format, tc.data,
				func(v *Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"Lookup.DecodeSelf",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nLookup.DecodeSelf(format=%q, data=%q)",
				packageName, tc.format, tc.data,
			)

			// AND: Pointers are handed out to it correctly.
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: dvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: dvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	// GIVEN: a Lookup in YAML/JSON.
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:     "valid/YAML/empty",
			format:   "yaml",
			data:     "",
			errRegex: `^$`,
		},
		{
			name:   "valid/YAML/host and repository",
			format: "yaml",
			data: test.TrimYAML(`
				host: https://codeberg.org
				url: owner/repo`,
			),
			want: test.TrimYAML(`
				host: https://codeberg.org
				url: owner/repo
			`),
			errRegex: `^$`,
		},
		{
			name:   "valid/YAML/use_prerelease",
			format: "yaml",
			data: test.TrimYAML(`
				host: https://codeberg.org
				url: owner/repo
				use_prerelease: true
			`),
			want: test.TrimYAML(`
				host: https://codeberg.org
				url: owner/repo
				use_prerelease: true
			`),
			errRegex: `^$`,
		},
		{
			name:     "valid/YAML/no host",
			format:   "yaml",
			data:     `url: owner/repo`,
			want:     "url: owner/repo\n",
			errRegex: `^$`,
		},
		{
			name:     "invalid/YAML/not a mapping",
			format:   "yaml",
			data:     `- host: https://codeberg.org`,
			errRegex: `sequence was used where mapping is expected`,
		},
		{
			name:   "invalid/YAML/type mismatch",
			format: "yaml",
			data: test.TrimYAML(`
				host: https://codeberg.org
				url: owner/repo
				use_prerelease: [true]
			`),
			errRegex: `cannot unmarshal .*UsePreRelease`,
		},
		{
			name:   "valid/JSON",
			format: "json",
			data: test.TrimJSON(`{
				"host": "https://codeberg.org",
				"url": "owner/repo",
				"use_prerelease": true
			}`),
			want: test.TrimYAML(`
				host: https://codeberg.org
				url: owner/repo
				use_prerelease: true
			`),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					return Decode(
						format, data,
						nil, nil,
						plainDefaultsConfig(t),
					)
				},
				tc.format, tc.data,
				func(v *Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"Decode",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestLookup_ApplyOverrides(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	type Args struct {
		format, data string
		target       *Lookup
	}
	// GIVEN: a decoded Lookup, and overrides to apply to it.
	tests := []struct {
		name     string
		args     Args
		errRegex string
		want     string
	}{
		{
			name: "empty data returns previous",
			args: Args{
				format: "json",
				data:   "",
				target: &Lookup{},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid payload causes decode error",
			args: Args{
				format: "json",
				data:   `{`,
				target: &Lookup{},
			},
			errRegex: test.TrimYAML(`
				^extract "require":
					[^\s]+ unexpected EOF`,
			),
		},
		{
			name: "override error/base.Lookup",
			args: Args{
				format: "json",
				data:   `{"url": []}`,
				target: &Lookup{},
			},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "override error/Lookup",
			args: Args{
				format: "json",
				data:   `{"use_prerelease": "true"}`,
				target: &Lookup{},
			},
			errRegex: `^json: .*unmarshal.* string.*$`,
		},
		{
			name: "require removed",
			args: Args{
				format: "json",
				data:   `{"require": null}`,
				target: &Lookup{
					URL: "owner/repo",
					Require: &filter.Require{
						RegexContent: "v?",
					},
				},
			},
			want:     "url: owner/repo\n",
			errRegex: `^$`,
		},
		{
			name: "valid require block",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
				"type": "forgejo",
				"require": {
					"regex_content": "v?"
				}
			}`),
				target: &Lookup{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: forgejo
				require:
					regex_content: v?
			`),
		},
		{
			name: "previous require inherited",
			args: Args{
				format: "json",
				data:   `{"type": "-"}`,
				target: &Lookup{
					Require: &filter.Require{
						RegexContent: "v?",
					},
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: '-'
				require:
					regex_content: v?
			`),
		},
		{
			name: "AccessToken added",
			args: Args{
				format: "json",
				data:   `{"access_token": "def"}`,
				target: &Lookup{
					Type: "forgejo",
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: forgejo
				access_token: def
			`),
		},
		{
			name: "AccessToken changed",
			args: Args{
				format: "json",
				data:   `{"access_token": "def"}`,
				target: &Lookup{
					Type:        "forgejo",
					AccessToken: "abc",
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: forgejo
				access_token: def
			`),
		},
		{
			name: "AccessToken removed",
			args: Args{
				format: "json",
				data:   `{"access_token": ""}`,
				target: &Lookup{
					Type:        "forgejo",
					AccessToken: "abc",
				},
			},
			errRegex: `^$`,
			want:     "type: forgejo\n",
		},
		{
			name: "Host added",
			args: Args{
				format: "json",
				data:   `{"host": "https://codeberg.org"}`,
				target: &Lookup{
					Type: "forgejo",
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: forgejo
				host: https://codeberg.org
			`),
		},
		{
			name: "Host changed",
			args: Args{
				format: "json",
				data:   `{"host": "https://gitea.com"}`,
				target: &Lookup{
					Type:          "forgejo",
					Host:          "https://codeberg.org",
					URL:           "owner/repo",
					UsePreRelease: new(true),
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: forgejo
				host: https://gitea.com
				url: owner/repo
				use_prerelease: true
			`),
		},
		{
			name: "Host removed",
			args: Args{
				format: "json",
				data:   `{"host": ""}`,
				target: &Lookup{
					Type: "forgejo",
					Host: "https://codeberg.org",
				},
			},
			errRegex: `^$`,
			want:     "type: forgejo\n",
		},
		{
			name: "filled",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "forgejo",
					"host": "https://gitea.com",
					"url": "other/repo",
					"access_token": "def",
					"allow_invalid_certs": true,
					"use_prerelease": false
				}`),
				target: &Lookup{
					Type:              "forgejo",
					Host:              "https://codeberg.org",
					URL:               "owner/repo",
					AccessToken:       "abc",
					AllowInvalidCerts: new(false),
					UsePreRelease:     new(true),
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: forgejo
				host: https://gitea.com
				url: other/repo
				access_token: def
				allow_invalid_certs: true
				use_prerelease: false
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: Options + Status.
			options := opttest.PlainOptions(t, optCfg)
			svcStatus := &status.Status{}
			tc.args.target.Init(options, svcStatus, lvCfg)
			// Default want to the stringified struct.
			if tc.want == "" {
				tc.want = decode.ToYAMLString(tc.args.target, "")
			}

			// WHEN: ApplyOverrides is called.
			lookup, err, testErr := test.AssertApplyOverrides(
				t,
				tc.args.target,
				func(format string, data []byte, v *Lookup) (*Lookup, error) {
					err := v.ApplyOverrides(format, data)
					return v, err
				},
				tc.args.format, tc.args.data,
				func(v *Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				true,
				packageName,
				"ApplyOverrides",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nApplyOverrides(format=%q, data=%q)",
				packageName, tc.args.format, tc.args.data,
			)

			// AND: pointers are handed out as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: lvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: lvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
