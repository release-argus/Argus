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

// Package base provides the base struct for latest_version lookups.
package base

import (
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
)

func TestDefaults_IsZero(t *testing.T) {
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
			name: "non-empty/Require",
			data: &Defaults{
				Require: filter.RequireDefaults{
					Docker: *test.Must(t, func() (*docker.Defaults, error) {
						return docker.DecodeDefaults(
							"yaml", []byte(`type: ghcr`),
							nil,
						)
					}),
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestDecodeDefaults(t *testing.T) {
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
			data:   `require: ['docker']`,
			errRegex: test.TrimYAML(`
					^latest_version:
						[^\s]+ sequence was used where mapping is expected
						[^\s]+.*
						\s+\^$`,
			),
		},
		{
			name:   "YAML/full",
			format: "yaml",
			data: test.TrimYAML(`
				foo: bar
				require:
					docker:
						tag: t
			`),
			want: test.TrimYAML(`
				require:
					docker:
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
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestDefaults_Default(t *testing.T) {
	defaults := Defaults{}
	want := Defaults{
		Require: filter.RequireDefaults{
			Docker: docker.Defaults{
				Type: "hub",
				Tag:  "{{ version }}",
			},
		},
	}

	defaults.Default()

	wantStr := decode.ToYAMLString(want, "")
	gotStr := decode.ToYAMLString(defaults, "")

	if gotStr != wantStr {
		t.Errorf("Defaults mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
	}
}

func TestDefaults_SetDefaults(t *testing.T) {
	d := &Defaults{}
	hd := &Defaults{}
	hd.Default()

	d.SetDefaults(hd)

	if d.Require.Docker.Defaults != &hd.Require.Docker {
		t.Errorf(
			"Defaults.Require.Docker.Defaults mismatch\ngot:  %p\nwant: %p",
			d.Require.Docker.Defaults, &hd.Require.Docker,
		)
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	tests := []struct {
		name     string
		input    *Defaults
		errRegex string
	}{
		{
			name: "valid",
			input: &Defaults{
				Require: filter.RequireDefaults{
					Docker: *test.Must(t, func() (*docker.Defaults, error) {
						return docker.DecodeDefaults(
							"yaml", []byte(`type: ghcr`),
							&lvCfg.Soft.Require.Docker,
						)
					}),
				},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid require",
			errRegex: test.TrimYAML(`
				^require:
					docker:
						type: "[^"]+" <invalid>.*$`,
			),
			input: func() *Defaults {
				input := Defaults{
					Require: filter.RequireDefaults{
						Docker: *test.Must(t, func() (*docker.Defaults, error) {
							return docker.DecodeDefaults(
								"yaml", []byte(`type: ghcr`),
								&lvCfg.Soft.Require.Docker,
							)
						}),
					},
				}
				input.Require.Docker.Type = "foo"
				return &input
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				tc.input.CheckValues,
			)
		})
	}
}
