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
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util/polymorphic"
)

func TestDecode(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	// GIVEN: data in a given format to Decode into a Lookup.
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
			errRegex: `^$`,
			want:     "",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			errRegex: `^$`,
			want:     "",
		},
		{
			name:   "JSON/invalid payload decode error",
			format: "json",
			data:   "{",
			errRegex: test.TrimYAML(`
				^latest_version:
					[^\s]+ unexpected EOF`,
			),
		},
		{
			name:   "JSON/valid payload",
			format: "json",
			data: test.TrimJSON(`{
				"type": "url",
				"url": "https://example.com"
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				url: https://example.com
			`),
		},
		{
			name:   "JSON/invalid Lookup type",
			format: "json",
			data:   `{"type":"url-5"}`,
			errRegex: test.TrimYAML(`
				^latest_version:
					type: "url-5" <invalid> \(supported values = \['github', 'url'\]\)$`,
			),
		},
		{
			name:   "JSON/Require extraction, invalid type",
			format: "json",
			data: test.TrimJSON(`{
				"type":"url",
				"require": 123
			}`),
			errRegex: test.TrimYAML(`
				^latest_version:
					require:
						extract "docker":
							json: .*unmarshal.* number.*$`,
			),
		},
		{
			name:   "JSON/valid Require block",
			format: "json",
			data: test.TrimJSON(`{
				"type":"url",
				"url": "https://example.com",
				"require": {
					"regex_content": "v?"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				url: https://example.com
				require:
					regex_content: v?
			`),
		},
		{
			name:   "YAML/github - bare",
			format: "yaml",
			data: test.TrimYAML(`
				type: github
				url: ` + test.ArgusGitHubRepo + `
			`),
			want: test.TrimYAML(`
				type: github
				url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/github - full",
			format: "yaml",
			data: test.TrimYAML(`
				type: github
				url: ` + test.ArgusGitHubRepo + `
				access_token: token
				url_commands:
					- type: split
						text: v
				require:
					regex_version: 'v[\d.]+'
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: '{{ version }}'
						auth:
							token: 123
				allow_invalid_certs: true
				use_prerelease: true
			`),
			want: test.TrimYAML(`
				type: github
				url: ` + test.ArgusGitHubRepo + `
				url_commands:
					- type: split
						text: v
				require:
					regex_version: 'v[\d.]+'
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: '{{ version }}'
						auth:
							token: '123'
				access_token: token
				use_prerelease: true
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/url - bare",
			format: "yaml",
			data: test.TrimYAML(`
				type: url
				url: https://example.com
			`),
			want: test.TrimYAML(`
				type: url
				url: https://example.com
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/url - full",
			format: "yaml",
			data: test.TrimYAML(`
				type: url
				url: ` + test.ArgusGitHubRepo + `
				access_token: token
				url_commands:
				- type: split
					text: v
				require:
					regex_version: v[\d.]+
					docker:
						type: hub
						image: releaseargus/argus
						tag: '{{ version }}'
						auth:
							username: me
							token: 123!
				allow_invalid_certs: false
				use_prerelease: true
			`),
			want: test.TrimYAML(`
				type: url
				url: ` + test.ArgusGitHubRepo + `
				url_commands:
					- type: split
						text: v
				require:
					regex_version: 'v[\d.]+'
					docker:
						type: hub
						image: releaseargus/argus
						tag: '{{ version }}'
						auth:
							username: me
							token: 123!
				allow_invalid_certs: false
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/unknown/invalid type",
			format: "yaml",
			data: test.TrimYAML(`
				type: foo
				url: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: test.TrimYAML(`
				^latest_version:
					type: "foo" <invalid> .*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: Options + Status.
			options, _ := opt.Decode(
				"yaml", nil,
				optCfg,
			)
			svcStatus, _ := statustest.New(
				"yaml", nil,
			)

			// WHEN: Decode is called with the args.
			lookup, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (Lookup, error) {
					return Decode(
						format, data,
						options,
						svcStatus,
						lvCfg,
					)
				},
				tc.format, tc.data,
				func(v Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"Decode",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nDecode(format=%q, data=%q)",
				packageName, tc.format, tc.data,
			)

			// AND: Pointers are handed out to it correctly.
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.GetOptions(), Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.GetStatus(), Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.GetDefaults(), Want: lvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.GetHardDefaults(), Want: lvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type BadInheritable struct{}

func (b *BadInheritable) GetType() string                                 { return "bad" }
func (b *BadInheritable) DecodeSelf(format string, data []byte) error     { return nil }
func (f *BadInheritable) ApplyOverrides(format string, data []byte) error { return nil }

func TestDecode_TypeAssertionFailure(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	orig := ServiceMapInheritable
	t.Cleanup(func() { ServiceMapInheritable = orig })

	// GIVEN: ServiceMap with a struct that does not implement Lookup.
	bad := map[string]func() polymorphic.Inheritable{
		"bad": func() polymorphic.Inheritable {
			return &BadInheritable{}
		},
	}
	ServiceMapInheritable = polymorphic.ToInheritableMap(bad)

	data := `{"type":"bad"}`

	// WHEN: We decode a payload that maps to that type.
	_, err := Decode(
		"json", []byte(data),
		nil,
		nil,
		lvCfg,
	)

	prefix := fmt.Sprintf(
		"%s\nDecode(%q)",
		packageName, data,
	)

	// THEN: We get an error because the type assertion failed.
	if err == nil {
		t.Fatalf("%s expected type assertion error", prefix)
	}
	if !strings.Contains(err.Error(), "expected latestver.Lookup") {
		t.Fatalf("%s unexpected error: %v", prefix, err)
	}
}

func TestApplyOverrides(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	type Args struct {
		format, data string
		target       Lookup
	}
	tests := []struct {
		name        string
		args        Args
		sameAddress bool
		errRegex    string
		want        string
	}{
		{
			name: "empty data returns previous",
			args: Args{
				format: "json",
				data:   "",
				target: &mockLookup{},
			},
			sameAddress: true,
			errRegex:    `^$`,
			want:        "lookup: {}\n",
		},
		{
			name: "null data returns nil",
			args: Args{
				format: "json",
				data:   `null`,
				target: &mockLookup{},
			},
			errRegex: `^$`,
			want:     "",
		},
		{
			name: "invalid payload causes decode error",
			args: Args{
				format: "json",
				data:   `{`,
				target: &mockLookup{},
			},
			errRegex: test.TrimYAML(`
				^latest_version:
					[^\s]+ unexpected EOF`,
			),
		},
		{
			name: "override error",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "-",
					"OverrideErr": "yes"
				}`),
				target: &mockLookup{
					OverrideErr: "yes",
				},
			},
			errRegex: test.TrimYAML(`
				^latest_version:
					yes$`,
			),
		},
		{
			name: "no previous, valid payload, no require",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "url",
					"url": "https://example.com"
				}`),
				target: nil,
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				url: https://example.com
			`),
		},
		{
			name: "have previous, valid payload, no require",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "url",
					"url": "https://example.com"
				}`),
				target: &mockLookup{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				url: https://example.com
			`),
		},
		{
			name: "require extraction, invalid data type",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type":["url"],
					"require": 123
				}`),
				target: &mockLookup{},
			},
			errRegex: test.TrimYAML(`
				^latest_version:
					json: .*unmarshal.*$`,
			),
		},
		{
			name: "require extraction, invalid type",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type":"url",
					"require": 123
				}`),
				target: &mockLookup{},
			},
			errRegex: test.TrimYAML(`
				^latest_version:
					require:
						extract "docker":
							json: .*unmarshal.* number .*`,
			),
		},
		{
			name: "valid require block",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type":"url",
					"url": "https://example.com",
					"require": {
						"regex_content": "v?"
					}
				}`),
				target: &mockLookup{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				url: https://example.com
				require:
					regex_content: v?
			`),
		},
		{
			name: "previous require inherited",
			args: Args{
				format: "json",
				data:   `{"type": "-"}`,
				target: &mockLookup{
					Lookup: base.Lookup{
						Require: &filter.Require{
							RegexContent: "v?",
						},
					},
				},
			},
			sameAddress: true,
			errRegex:    `^$`,
			want: test.TrimYAML(`
				lookup:
					require:
						regex_content: v?
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: Options + Status.
			options, _ := opt.Decode(
				"yaml", nil,
				optCfg,
			)
			svcStatus, _ := statustest.New(
				"yaml", nil,
			)

			// WHEN: ApplyOverrides is called.
			if _, _, testErr := test.AssertApplyOverrides(
				t,
				tc.args.target,
				func(format string, data []byte, v Lookup) (Lookup, error) {
					return ApplyOverrides(
						format, data,
						v,
						options,
						svcStatus,
						lvCfg,
					)
				},
				tc.args.format, tc.args.data,
				func(v Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				tc.sameAddress,
				packageName,
				"ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
