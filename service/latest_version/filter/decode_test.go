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

package filter

import (
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

func TestDecode(t *testing.T) {
	defaults, _ := plainDefaults()

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
			data:     "",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/null",
			format:   "json",
			data:     "null",
			errRegex: `^$`,
			want:     "",
		},
		{
			name:     "YAML/null",
			format:   "yaml",
			data:     "null",
			errRegex: `^$`,
			want:     "",
		},
		{
			name:   "JSON/invalid payload decode error",
			format: "json",
			data:   `{`,
			errRegex: test.TrimYAML(`
				^require:
					extract "docker":
						[^\s]+ unexpected EOF`,
			),
		},
		{
			name:   "JSON/docker extraction, invalid data type",
			format: "json",
			data:   `{"docker": 123}`,
			errRegex: test.TrimYAML(`
				^require:
					docker:
						json: .*unmarshal.* number.*$`,
			),
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"regex_content": "content-v?",
				"regex_version": "v?",
				"command": ["ls", "-lah"],
				"docker": {
					"type": "hub",
					"image": "test/app",
					"tag": "{{ version }}"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				regex_content: content-v?
				regex_version: v?
				command:
					- ls
					- -lah
				docker:
					type: hub
					image: test/app
					tag: '{{ version }}'
			`),
		},
		{
			name:   "JSON/static fields unmarshal fail",
			format: "json",
			data:   `{"regex_version": ["-"]}`,
			errRegex: test.TrimYAML(`
				^require:
					json: .*unmarshal.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Require, error) {
					return Decode(
						format, data,
						test.Must(t, func() (*status.Status, error) {
							return statustest.New("yaml", nil)
						}),
						defaults,
					)
				},
				tc.format, tc.data,
				func(v *Require) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"Decode",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestApplyOverrides(t *testing.T) {
	defaults, _ := plainDefaults()

	type Args struct {
		format, data string
		target       *Require
	}
	tests := []struct {
		name     string
		args     Args
		errRegex string
		want     *string
	}{
		{
			name: "New Require",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"docker": {
						"type": "ghcr",
						"image": "` + test.ArgusDockerGHCRRepo + `"
					}
				}`),
				target: nil,
			},
			errRegex: `^$`,
			want: new(
				test.TrimYAML(`
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
				`),
			),
		},
		{
			name: "empty data returns previous",
			args: Args{
				format: "json",
				data:   "",
				target: &Require{},
			},
			errRegex: `^$`,
		},
		{
			name: "null data returns nil",
			args: Args{
				format: "json",
				data:   `null`,
				target: &Require{},
			},
			errRegex: `^$`,
			want:     new(""),
		},
		{
			name: "invalid payload, decode error",
			args: Args{
				format: "json",
				data:   `{"`,
				target: &Require{},
			},
			errRegex: test.TrimYAML(`
				^require:
					extract "docker":
						[^\s]+ unexpected EOF`,
			),
		},
		{
			name: "valid payload, no docker",
			args: Args{
				format: "json",
				data:   `{"regex_content": "v?\\d+"}`,
				target: &Require{},
			},
			errRegex: `^$`,
			want:     new("regex_content: 'v?\\d+'\n"),
		},
		{
			name: "docker extraction, invalid data type",
			args: Args{
				format: "json",
				data:   `{"docker": 123}`,
				target: &Require{},
			},
			errRegex: test.TrimYAML(`
				^require:
					docker:
						json: .*unmarshal.* number .*`,
			),
		},
		{
			name: "filled",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"regex_content": "content-v?",
					"regex_version": "v?",
					"command": ["ls", "-lah"],
					"docker": {
						"type": "hub",
						"image": "` + test.ArgusDockerHubRepo + `",
						"tag": "{{ version }}"
					}
				}`),
				target: &Require{},
			},
			errRegex: `^$`,
			want: new(
				test.TrimYAML(`
					regex_content: content-v?
					regex_version: v?
					command:
						- ls
						- -lah
					docker:
						type: hub
						image: ` + test.ArgusDockerHubRepo + `
						tag: '{{ version }}'
				`),
			),
		},
		{
			name: "previous require inherited",
			args: Args{
				format: "json",
				data:   `{"regex_version": "-"}`,
				target: &Require{
					RegexContent: "v?",
				},
			},
			errRegex: `^$`,
			want: new(
				test.TrimYAML(`
					regex_content: v?
					regex_version: '-'
				`),
			),
		},
		{
			name: "Docker: removed",
			args: Args{
				format: "json",
				data:   `{"docker": null}`,
				target: &Require{
					RegexContent: "content-v?",
					RegexVersion: "v?",
					Docker: &docker.HubRegistry{
						Type:  "hub",
						Image: test.ArgusDockerHubRepo,
						Tag:   "{{ version }}",
					},
				},
			},
			errRegex: `^$`,
			want: new(
				test.TrimYAML(`
					regex_content: content-v?
					regex_version: v?
				`),
			),
		},
		{
			name: "Docker: hub -> ghcr",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"docker": {
						"type": "ghcr",
						"image": "` + test.ArgusDockerGHCRRepo + `"
					}
				}`),
				target: &Require{
					RegexContent: "content-v?",
					RegexVersion: "v?",
					Docker: &docker.HubRegistry{
						Type:  "hub",
						Image: test.ArgusDockerHubRepo,
						Tag:   "{{ version }}",
						Auth:  &docker.HubAuth{},
					},
				},
			},
			errRegex: `^$`,
			want: new(
				test.TrimYAML(`
					regex_content: content-v?
					regex_version: v?
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
				`),
			),
		},
		{
			name: "static fields unmarshal fail",
			args: Args{
				format: "json",
				data:   `{"regex_version": ["-"]}`,
				target: &Require{},
			},
			errRegex: test.TrimYAML(`
				^require:
					json: .*unmarshal.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.args.target != nil {
				tc.args.target.defaults = defaults
			}
			// Without a want, the target is expected unchanged.
			want := tc.args.target.String("")
			if tc.want != nil {
				want = *tc.want
			}

			if _, _, testErr := test.AssertApplyOverrides(
				t,
				tc.args.target,
				func(format string, data []byte, v *Require) (*Require, error) {
					return v.ApplyOverrides(
						format, data,
						test.Must(t, func() (*status.Status, error) {
							return statustest.New("yaml", nil)
						}),
						defaults,
					)
				},
				tc.args.format, tc.args.data,
				func(v *Require) string { return v.String("") },
				want,
				tc.errRegex,
				true,
				"Require.ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestDecodeDefaults(t *testing.T) {
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
			data:     "",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/null",
			format:   "json",
			data:     "null",
			errRegex: `^$`,
			want:     "",
		},
		{
			name:     "YAML/null",
			format:   "yaml",
			data:     "null",
			errRegex: `^$`,
			want:     "",
		},
		{
			name:   "JSON/invalid payload decode error",
			format: "json",
			data:   `{`,
			errRegex: test.TrimYAML(`
				^require:
					extract "docker":
						[^\s]+ unexpected EOF`,
			),
		},
		{
			name:   "JSON/docker extraction, invalid data type",
			format: "json",
			data:   `{"docker": 123}`,
			errRegex: test.TrimYAML(`
				^require:
					docker:
						json: .*unmarshal.* number.*$`,
			),
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"regex_content": "content-v?",
				"regex_version": "v?",
				"command": ["ls", "-lah"],
				"docker": {
					"type": "hub",
					"tag": "{{ version }}"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				docker:
					type: hub
					tag: '{{ version }}'
			`),
		},
		{
			name:   "JSON/docker unmarshal fail",
			format: "json",
			data:   `{"docker": ["-"]}`,
			errRegex: test.TrimYAML(`
				^require:
					docker:
						json: .*unmarshal.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				DecodeDefaults,
				tc.format, tc.data,
				func(v *RequireDefaults) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
