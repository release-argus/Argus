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

package base

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestDecode(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	type Args struct {
		format string
		data   string
		lookup *Lookup
	}
	// GIVEN: data in a given format to Decode into a Lookup.
	tests := []struct {
		name     string
		args     Args
		errRegex string
		want     string
	}{
		{
			name: "JSON/empty",
			args: Args{
				format: "json",
				data:   `{}`,
			},
			want: "{}\n",
		},
		{
			name: "YAML/empty",
			args: Args{
				format: "yaml",
				data:   "",
			},
			errRegex: `^$`,
			want:     "",
		},
		{
			name: "invalid payload causes decode error",
			args: Args{
				format: "json",
				data:   `{`,
			},
			errRegex: test.TrimYAML(`
				^extract "require":
					[^\s]+ unexpected EOF`,
			),
		},
		{
			name: "valid payload, invalid type",
			args: Args{
				format: "json",
				data:   `{"type": ["url"]}`,
			},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "valid payload, no require",
			args: Args{
				format: "json",
				data:   `{"type": "url"}`,
			},
			errRegex: `^$`,
			want:     "type: url\n",
		},
		{
			name: "require extraction, invalid type",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type":"url",
					"require": 123
				}`),
			},
			errRegex: test.TrimYAML(`
				^require:
					extract "docker":
						json: .*unmarshal.* number.*$`,
			),
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
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				require:
					regex_content: v?
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := opttest.PlainOptions(t, optCfg)
			svcStatus := &status.Status{}

			// WHEN: Decode is called.
			lookup, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					return Decode(
						format, data,
						options,
						svcStatus,
						lvCfg,
					)
				},
				tc.args.format,
				tc.args.data,
				func(v *Lookup) string { return decode.ToYAMLString(v, "") },
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

func TestApplyOverrides(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	type Args struct {
		format string
		data   string
		target *Lookup
	}
	// GIVEN: data in a given format to Decode into a Lookup.
	tests := []struct {
		name              string
		args              Args
		want              string
		reqPointerChanged bool
		errRegex          string
	}{
		{
			name: "JSON/new, fail to extract require",
			args: Args{
				format: "json",
				data: test.TrimYAML(`
					[
						{
							"require": {
								"regex_content": "a",
								"regex_version": "b",
								"command": ["ls", "-lah"],
								"docker": {
									"type": "hub",
									"image": "foo",
									"tag": "bar"
								}
							}
						}
					]`,
				),
				target: nil,
			},
			errRegex: test.TrimYAML(`
				^extract "require":
					json: .*unmarshal.*$`,
			),
		},
		{
			name: "YAML/new, fail to extract require",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					[
						require:
							regex_content: a
							regex_version: b
							command: ["ls", "-lah"]
							docker:
								type: hub
								image: foo
								tag: bar
					]`,
				),
				target: nil,
			},
			errRegex: test.TrimYAML(`
				^extract "require":
					[^\s]+ sequence was used.*`,
			),
		},
		{
			name: "JSON/existing, fail to extract require",
			args: Args{
				format: "json",
				data: test.TrimYAML(`
					[
						{
							"require": {
								"regex_content": "a",
								"regex_version": "b",
								"command": ["ls", "-lah"],
								"docker": {
									"type": "hub",
									"image": "foo",
									"tag": "bar"
								}
							}
						}
					]`,
				),
				target: &Lookup{},
			},
			errRegex: test.TrimYAML(`
				^extract "require":
					json: .*unmarshal.*$`,
			),
		},
		{
			name: "YAML/existing, fail to extract require",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					[
						require:
							regex_content: a
							regex_version: b
							command: ["ls", "-lah"]
							docker:
								type: hub
								image: foo
								tag: bar
					]`,
				),
				target: &Lookup{},
			},
			errRegex: test.TrimYAML(`
				^extract "require":
					[^\s]+ sequence was used.*`,
			),
		},
		{
			name: "JSON/existing, no require",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}`),
				target: &Lookup{
					URL: "release-argus/Test",
					URLCommands: filter.URLCommands{
						{Type: "regex", Regex: "v?x"},
					},
				},
			},
			want: test.TrimYAML(`
				type: github
				url: ` + test.ArgusGitHubRepo + `
				url_commands:
					- type: regex
						regex: v?x
			`),
			errRegex: `^$`,
		},
		{
			name: "YAML/existing, no require",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					type: github
					url: ` + test.ArgusGitHubRepo + `
				`),
				target: &Lookup{
					URL: "release-argus/Test",
					URLCommands: filter.URLCommands{
						{Type: "regex", Regex: "v?x"},
					},
				},
			},
			want: test.TrimYAML(`
				type: github
				url: ` + test.ArgusGitHubRepo + `
				url_commands:
					- type: regex
						regex: v?x
			`),
			errRegex: `^$`,
		},
		{
			name: "JSON/no data",
			args: Args{
				format: "json",
				data:   "",
				target: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "YAML/no data",
			args: Args{
				format: "yaml",
				data:   "",
				target: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "JSON/existing, remove Require",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"url": "https://example.com/123",
					"require": null
				}`),
				target: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			want:     "url: https://example.com/123\n",
			errRegex: `^$`,
		},
		{
			name: "YAML/existing, remove Require",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					url: https://example.com/456
					require: null
				`),
				target: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			want:     "url: https://example.com/456\n",
			errRegex: `^$`,
		},
		{
			name: "JSON/existing, remove Require.Docker",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"url": "https://example.com/123",
					"require": {
						"regex_content": "c",
						"regex_version": "b",
						"command": ["ls", "-lah"],
						"docker": null
					}
				}`),
				target: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			want: test.TrimYAML(`
				url: https://example.com/123
				require:
					regex_content: c
					regex_version: b
					command:
						- ls
						- -lah
			`),
			errRegex: `^$`,
		},
		{
			name: "YAML/existing, remove Require.Docker",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					url: https://example.com/456
					require:
						regex_content: c
						regex_version: b
						command: ['ls', '-lah']
						docker: null
				`),
				target: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			want: test.TrimYAML(`
				url: https://example.com/456
				require:
					regex_content: c
					regex_version: b
					command:
						- ls
						- -lah
			`),
			errRegex: `^$`,
		},
		{
			name: "JSON/existing, change Require.x and Require.Docker.x",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"url": "https://example.com/123",
					"require": {
						"regex_content": "c",
						"regex_version": "b",
						"command": ["ls", "-lah"],
						"docker": {
							"type": "hub",
							"image": "foo",
							"tag": "something-else"
						}
					}
				}`),
				target: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			want: test.TrimYAML(`
				url: https://example.com/123
				require:
					regex_content: c
					regex_version: b
					command:
						- ls
						- -lah
					docker:
						type: hub
						image: foo
						tag: something-else
			`),
			errRegex: `^$`,
		},
		{
			name: "YAML/existing, change Require.x and Require.Docker.x",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					url: https://example.com/456
					require:
						regex_content: c
						regex_version: b
						command: ['ls', '-lah']
						docker:
							type: hub
							image: foo
							tag: something-else
				`),
				target: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			want: test.TrimYAML(`
				url: https://example.com/456
				require:
					regex_content: c
					regex_version: b
					command:
						- ls
						- -lah
					docker:
						type: hub
						image: foo
						tag: something-else
			`),
			errRegex: `^$`,
		},
		{
			name: "JSON/existing, change Require.x - type error",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"require": {
						"regex_content": ["c"]
					}
				}`),
				target: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			errRegex: test.TrimYAML(`
				^require:
					json: .*unmarshal.*$`,
			),
		},
		{
			name: "YAML/existing, change Require.x - type error",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					require:
						regex_content: [c]
				`),
				target: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			errRegex: test.TrimYAML(`
				^require:
					[^\s]+ .*unmarshal.*`,
			),
		},
		{
			name: "JSON/new, filled Require",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"require": {
						"regex_content": "a",
						"regex_version": "b",
						"command": ["ls", "-lah"],
						"docker": {
							"type": "hub",
							"image": "foo",
							"tag": "bar"
						}
					}
				}`),
				target: nil,
			},
			want: test.TrimYAML(`
				require:
					regex_content: a
					regex_version: b
					command:
						- ls
						- -lah
					docker:
						type: hub
						image: foo
						tag: bar
			`),
			errRegex: `^$`,
		},
		{
			name: "YAML/new, filled Require",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					require:
						regex_content: a
						regex_version: b
						command: ['ls', '-lah']
						docker:
							type: hub
							image: foo
							tag: bar
				`),
				target: nil,
			},
			want: test.TrimYAML(`
				require:
					regex_content: a
					regex_version: b
					command:
						- ls
						- -lah
					docker:
						type: hub
						image: foo
						tag: bar
			`),
			errRegex: `^$`,
		},
		{
			name: "JSON/new - type error",
			args: Args{
				format: "json",
				data:   `{"url": ["https://example.com"]}`,
				target: nil,
			},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "YAML/new - type error",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					url:
						- https://example.com
				`),
				target: nil,
			},
			errRegex: `^[^\s]+ .*unmarshal.*`,
		},
		{
			name: "JSON/existing - type error",
			args: Args{
				format: "json",
				data:   `{"url": ["https://example.com"]}`,
				target: &Lookup{},
			},
			errRegex: `^json: .*unmarshal.*`,
		},
		{
			name: "YAML/new, filled - type error",
			args: Args{
				format: "yaml",
				data:   `url: [https://example.com]`,
				target: &Lookup{},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := opttest.PlainOptions(t, optCfg)
			svcStatus := &status.Status{}
			if tc.args.target != nil {
				tc.args.target.Init(
					options,
					svcStatus,
					lvCfg,
				)
			}

			// WHEN: ApplyOverrides is called.
			lookup, err, testErr := test.AssertApplyOverrides(
				t,
				tc.args.target,
				func(format string, data []byte, v *Lookup) (*Lookup, error) {
					return ApplyOverrides(
						format, data,
						v,
						options,
						svcStatus,
						lvCfg,
					)
				},
				tc.args.format, tc.args.data,
				func(v *Lookup) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				true,
				packageName,
				"Lookup.ApplyOverrides",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nApplyOverrides(format=%q, data=%q, previous=%p)",
				packageName, tc.args.format, tc.args.data, tc.args.target,
			)

			wantOptions := options
			wantStatus := svcStatus
			wantDefaults := lvCfg.Soft
			wantHardDefaults := lvCfg.Hard
			// Fields unchanged when no data provided.
			if len(tc.args.data) == 0 {
				wantOptions = tc.args.target.Options
				wantStatus = tc.args.target.Status
				wantDefaults = tc.args.target.Defaults
				wantHardDefaults = tc.args.target.HardDefaults
			}
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: wantOptions, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: wantStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: wantDefaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: wantHardDefaults, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}

			// AND: the Require pointer changes only when the root pointer changes.
			if tc.args.target != nil && tc.args.target.Require != nil {
				reqPointerChanged := lookup.Require != tc.args.target.Require
				if reqPointerChanged != tc.reqPointerChanged {
					t.Fatalf(
						"%s .Require pointer changed unexpectedly\ngot:  changed=%v\nwant: changed=%v",
						prefix, reqPointerChanged, tc.reqPointerChanged,
					)
				}
			}
		})
	}
}

func TestUnmarshalRequire(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	type Args struct {
		format string
		data   string
		lookup *Lookup
	}
	// GIVEN: data to unmarshal in a format onto a Require.
	tests := []struct {
		name     string
		args     Args
		want     string
		errRegex string
	}{
		{
			name: "JSON/fail to extract require",
			args: Args{
				format: "json",
				data: test.TrimYAML(`
					[
						{
							"require": {
								"regex_content": "a",
								"regex_version": "b",
								"command": ["ls", "-lah"],
								"docker": {
									"type": "hub",
									"image": "foo",
									"tag": "bar"
								}
							}
						}
					]`,
				),
				lookup: &Lookup{},
			},
			errRegex: test.TrimYAML(`
				^extract "require":
					json: .*unmarshal.*$`,
			),
		},
		{
			name: "YAML/fail to extract require",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					[
						require:
							regex_content: a
							regex_version: b
							command: ["ls", "-lah"]
							docker:
								type: hub
								image: foo
								tag: bar
					]`,
				),
				lookup: &Lookup{},
			},
			errRegex: test.TrimYAML(`
				^extract "require":
					[^\s]+ sequence was used.*`,
			),
		},
		{
			name: "JSON/no require",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}`),
				lookup: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "YAML/no require",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					type: github
					url: ` + test.ArgusGitHubRepo + `
				`),
				lookup: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "JSON/no data",
			args: Args{
				format: "json",
				data:   "",
				lookup: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "YAML/no data",
			args: Args{
				format: "yaml",
				data:   "",
				lookup: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "JSON/existing, change Require.x and Require.Docker.x",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"require": {
						"regex_content": "c",
						"regex_version": "b",
						"command": ["ls", "-lah"],
						"docker": {
							"type": "hub",
							"image": "foo",
							"tag": "something-else"
						}
					}
				}`),
				lookup: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			want: test.TrimYAML(`
				url: https://example.com
				require:
					regex_content: c
					regex_version: b
					command:
						- ls
						- -lah
					docker:
						type: hub
						image: foo
						tag: something-else
			`),
			errRegex: `^$`,
		},
		{
			name: "YAML/existing, change Require.x and Require.Docker.x",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					url: https://example.com
					require:
						regex_content: c
						regex_version: b
						command: ['ls', '-lah']
						docker:
							type: hub
							image: foo
							tag: something-else
				`),
				lookup: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			want: test.TrimYAML(`
				url: https://example.com
				require:
					regex_content: c
					regex_version: b
					command:
						- ls
						- -lah
					docker:
						type: hub
						image: foo
						tag: something-else
			`),
			errRegex: `^$`,
		},
		{
			name: "JSON/existing, change Require.x - type error",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"require": {
						"regex_content": ["c"]
					}
				}`),
				lookup: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			errRegex: test.TrimYAML(`
				^require:
					json: .*unmarshal.*$`,
			),
		},
		{
			name: "YAML/existing, change Require.x - type error",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					require:
						regex_content: [c]
				`),
				lookup: &Lookup{
					URL: "https://example.com",
					Require: test.Must(t, func() (*filter.Require, error) {
						return filter.Decode(
							"yaml", []byte(test.TrimYAML(`
								regex_content: a
								regex_version: b
								command: ['ls', '-lah']
								docker:
									type: hub
									image: foo
									tag: bar
							`)),
							test.Must(t, func() (*status.Status, error) {
								return statustest.New("yaml", nil)
							}),
							&lvCfg.Soft.Require,
						)
					}),
				},
			},
			errRegex: test.TrimYAML(`
				^require:
					[^\s]+ .*unmarshal.*`,
			),
		},
		{
			name: "JSON/new, filled",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"require": {
						"regex_content": "a",
						"regex_version": "b",
						"command": ["ls", "-lah"],
						"docker": {
							"type": "hub",
							"image": "foo",
							"tag": "bar"
						}
					}
				}`),
				lookup: &Lookup{},
			},
			want: test.TrimYAML(`
				require:
					regex_content: a
					regex_version: b
					command:
						- ls
						- -lah
					docker:
						type: hub
						image: foo
						tag: bar
			`),
			errRegex: `^$`,
		},
		{
			name: "YAML/new, filled",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					require:
						regex_content: a
						regex_version: b
						command: ['ls', '-lah']
						docker:
							type: hub
							image: foo
							tag: bar
				`),
				lookup: &Lookup{},
			},
			want: test.TrimYAML(`
				require:
					regex_content: a
					regex_version: b
					command:
						- ls
						- -lah
					docker:
						type: hub
						image: foo
						tag: bar
			`),
			errRegex: `^$`,
		},
		{
			name: "JSON/new, filled - type error",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"require": {
						"regex_content": ["a"]
					}
				}`),
				lookup: &Lookup{},
			},
			errRegex: test.TrimYAML(`
				^require:
					json: .*unmarshal.*$`,
			),
		},
		{
			name: "YAML/new, filled - type error",
			args: Args{
				format: "yaml",
				data: test.TrimYAML(`
					require:
						regex_content: [a]
				`),
				lookup: &Lookup{},
			},
			errRegex: test.TrimYAML(`
				^require:
					[^\s]+ .*unmarshal.*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: UnmarshalRequire is called.
			err := UnmarshalRequire(
				tc.args.format, []byte(tc.args.data),
				tc.args.lookup,
				test.Must(t, func() (*status.Status, error) {
					return statustest.New("yaml", nil)
				}),
				&lvCfg.Soft.Require,
			)

			prefix := fmt.Sprintf(
				"%s\nUnmarshalRequire(format=%q, data=%q)",
				packageName, tc.args.format, tc.args.data,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the new struct stringifies as expected
			if got := decode.ToYAMLString(&tc.args.lookup, ""); got != tc.want {
				t.Fatalf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}
