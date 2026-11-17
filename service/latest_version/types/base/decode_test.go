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
	lvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig()

	type Args struct {
		format string
		data   string
	}

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
				data:   "",
			},
			want:     "",
			errRegex: `^$`,
		},
		{
			name: "JSON/empty object",
			args: Args{
				format: "json",
				data:   "{}",
			},
			want:     "{}\n",
			errRegex: `^$`,
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
			name: "valid payload/invalid type",
			args: Args{
				format: "json",
				data:   `{"type": ["url"]}`,
			},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "valid payload/no require",
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

			options := opttest.PlainOptions(optCfg)
			svcStatus := &status.Status{}

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
				"Decode",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: lvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: lvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestApplyOverrides(t *testing.T) {
	lvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig()

	type Args struct {
		format string
		data   string
		target *Lookup
	}

	tests := []struct {
		name              string
		args              Args
		want              string
		reqPointerChanged bool
		errRegex          string
	}{
		{
			name: "JSON/new/no data",
			args: Args{
				format: "json",
				data:   "",
				target: nil,
			},
			want:     "",
			errRegex: `^$`,
		},
		{
			name: "JSON/new/empty object",
			args: Args{
				format: "json",
				data:   "{}",
				target: nil,
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "YAML/new/no data",
			args: Args{
				format: "yaml",
				data:   "",
				target: nil,
			},
			want:     "",
			errRegex: `^$`,
		},
		{
			name: "JSON/new/fail to extract require",
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
			name: "YAML/new/fail to extract require",
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
			name: "JSON/existing/no data",
			args: Args{
				format: "json",
				data:   "",
				target: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "JSON/existing/empty object",
			args: Args{
				format: "json",
				data:   "{}",
				target: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "YAML/existing/no data",
			args: Args{
				format: "yaml",
				data:   "",
				target: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "JSON/existing/fail to extract require",
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
			name: "YAML/existing/fail to extract require",
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
			name: "JSON/existing/no require",
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
			name: "YAML/existing/no require",
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
			name: "JSON/existing/remove Require",
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
			want:              "url: https://example.com/123\n",
			reqPointerChanged: true,
			errRegex:          `^$`,
		},
		{
			name: "YAML/existing/remove Require",
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
			want:              "url: https://example.com/456\n",
			reqPointerChanged: true,
			errRegex:          `^$`,
		},
		{
			name: "JSON/existing/remove Require.Docker",
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
			name: "YAML/existing/remove Require.Docker",
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
			name: "JSON/existing/change Require.x and Require.Docker.x",
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
			name: "YAML/existing/change Require.x and Require.Docker.x",
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
			name: "JSON/existing/change Require.x/type error",
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
			name: "YAML/existing/change Require.x/type error",
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
			name: "JSON/new/filled Require",
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
			name: "YAML/new/filled Require",
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
			name: "JSON/new/type error",
			args: Args{
				format: "json",
				data:   `{"url": ["https://example.com"]}`,
				target: nil,
			},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "YAML/new/type error",
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
			name: "JSON/existing/type error",
			args: Args{
				format: "json",
				data:   `{"url": ["https://example.com"]}`,
				target: &Lookup{},
			},
			errRegex: `^json: .*unmarshal.*`,
		},
		{
			name: "YAML/new/filled/type error",
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

			options := opttest.PlainOptions(optCfg)
			svcStatus := &status.Status{}
			var hadRequire *filter.Require
			if tc.args.target != nil {
				tc.args.target.Init(
					options,
					svcStatus,
					lvCfg,
				)
				hadRequire = tc.args.target.Require
			}

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
				"ApplyOverrides",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: lvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: lvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			// An existing Require is patched in place, and replaced only by a null override.
			if hadRequire != nil {
				reqPointerChanged := lookup.Require != hadRequire
				if reqPointerChanged != tc.reqPointerChanged {
					t.Fatalf(
						"Lookup.Require mismatch\ngot:  changed=%t\nwant: changed=%t",
						reqPointerChanged, tc.reqPointerChanged,
					)
				}
			}
		})
	}
}

func TestUnmarshalRequire(t *testing.T) {
	lvCfg := plainDefaultsConfig()

	type Args struct {
		format string
		data   string
		lookup *Lookup
	}

	tests := []struct {
		name     string
		args     Args
		want     string
		errRegex string
	}{
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
			name: "JSON/empty object",
			args: Args{
				format: "json",
				data:   "{}",
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
			name: "JSON/existing/change Require.x and Require.Docker.x",
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
			name: "YAML/existing/change Require.x and Require.Docker.x",
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
			name: "JSON/existing/change Require.x/type error",
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
			name: "YAML/existing/change Require.x/type error",
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
			name: "JSON/new, filled/bare",
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
			name: "YAML/new, filled/bare",
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
			name: "JSON/new, filled/type error",
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
			name: "YAML/new, filled/type error",
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

			err := UnmarshalRequire(
				tc.args.format, []byte(tc.args.data),
				tc.args.lookup,
				test.Must(t, func() (*status.Status, error) {
					return statustest.New("yaml", nil)
				}),
				&lvCfg.Soft.Require,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if got := decode.ToYAMLString(tc.args.lookup, ""); got != tc.want {
				t.Fatalf("Lookup mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
