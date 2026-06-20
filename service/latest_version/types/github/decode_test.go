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

package github

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_DecodeSelf(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)
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
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
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
			data:     ``,
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:   "invalid payload causes decode error",
			format: "json",
			data:   `{`,
			errRegex: test.TrimYAML(`
				^extract "require":
					[^\s]+ unexpected EOF`,
			),
		},
		{
			name:   "valid payload, no require",
			format: "json",
			data: test.TrimJSON(`{
				"type": "github",
				"allow_invalid_certs": true,
				"access_token": "abc",
				"use_prerelease": false
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: github
				access_token: abc
				use_prerelease: false
			`),
		},
		{
			name:     "invalid data types",
			format:   "json",
			data:     `{"use_prerelease": "true"}`,
			errRegex: `^json: .*unmarshal.*$`,
			want:     "type: url\n",
		},
		{
			name:   "require extraction, invalid type",
			format: "json",
			data: test.TrimJSON(`{
				"type": "github",
				"require": 123
			}`),
			errRegex: test.TrimYAML(`
				^require:
					extract "docker":
						json: .*unmarshal.* number.*$`,
			),
		},
		{
			name:   "valid require block",
			format: "json",
			data: test.TrimJSON(`{
				"type":"github",
				"require": {
					"regex_content": "v?"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: github
				require:
					regex_content: v?
			`),
		},
		{
			name:   "filled",
			format: "json",
			data: test.TrimJSON(`{
				"type": "github",
				"url": "https://example.com",
				"access_token": "abc",
				"use_prerelease": false,
				"require": {
					"regex_version": "v?"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: github
				url: https://example.com
				require:
					regex_version: v?
				access_token: abc
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
				lvCfg,
			)

			// WHEN: DecodeSelf is called.
			decoded, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					err := lookup.DecodeSelf(format, data)
					return lookup, err
				},
				tc.format, tc.data,
				func(lv *Lookup) string { return decode.ToYAMLString(lv, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"Lookup.DecodeSelf",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if decoded == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nLookup.DecodeSelf(format=%q, data=%q)",
				tc.format, tc.format, tc.data,
			)

			// THEN: pointers are set as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: decoded.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: decoded.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: decoded.Defaults, Want: lvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: decoded.HardDefaults, Want: lvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
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
					Lookup: base.Lookup{
						URL: "https://example.com",
						Require: &filter.Require{
							RegexContent: "v?",
						},
					},
				},
			},
			want:     "url: https://example.com\n",
			errRegex: `^$`,
		},
		{
			name: "valid require block",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
				"type":"url",
				"require": {
					"regex_content": "v?"
				}
			}`),
				target: &Lookup{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				require:
					regex_content: v?
			`),
		},
		{
			name: "previous require inherited",
			args: Args{
				format: "json",
				data: `{
				"type": "-"
			}`,
				target: &Lookup{
					Lookup: base.Lookup{
						Require: &filter.Require{
							RegexContent: "v?",
						},
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
					Lookup: base.Lookup{
						Type: "github",
					},
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: github
				access_token: def
			`),
		},
		{
			name: "AccessToken changed",
			args: Args{
				format: "json",
				data:   `{"access_token": "def"}`,
				target: &Lookup{
					Lookup: base.Lookup{
						Type: "github",
					},
					AccessToken: "abc",
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: github
				access_token: def
			`),
		},
		{
			name: "AccessToken removed",
			args: Args{
				format: "json",
				data:   `{"access_token": ""}`,
				target: &Lookup{
					Lookup: base.Lookup{
						Type: "github",
					},
					AccessToken: "abc",
				},
			},
			want: "type: github\n",
		},
		{
			name: "filled",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
				"type": "github",
				"url": "https://release-argus",
				"access_token": "def",
				"use_prerelease": false
			}`),
				target: &Lookup{
					Lookup: base.Lookup{
						Type: "github",
						URL:  "https://example.com",
					},
					AccessToken:   "abc",
					UsePreRelease: test.Ptr(true),
				},
			},
			want: test.TrimYAML(`
				type: github
				url: https://release-argus
				access_token: def
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
			// Default want to the unchanged stringified struct.
			if tc.want == "" {
				tc.want = decode.ToYAMLString(tc.args.target, "")
			}
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
