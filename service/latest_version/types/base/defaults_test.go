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
			data: &Defaults{
				Type: "a",
			},
			want: false,
		},
		{
			name: "non-empty/AccessToken",
			data: &Defaults{
				AccessToken: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/AllowInvalidCerts",
			data: &Defaults{
				AllowInvalidCerts: test.Ptr(true),
			},
			want: false,
		},
		{
			name: "non-empty/UsePreRelease",
			data: &Defaults{
				UsePreRelease: test.Ptr(true),
			},
			want: false,
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
		{
			name: "non-empty/all",
			data: &Defaults{
				Type:              "a",
				AccessToken:       "foo",
				AllowInvalidCerts: test.Ptr(true),
				UsePreRelease:     test.Ptr(true),
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
						[^\s]+ .*unmarshal.*
						[^\s]+.*
						\s+\^$`,
			),
		},
		{
			name:   "YAML/full",
			format: "yaml",
			data: test.TrimYAML(`
				type: github
				access_token: foo
				allow_invalid_certs: false
				use_prerelease: true
				foo: bar
				require:
					docker:
						tag: t
			`),
			want: test.TrimYAML(`
				type: github
				access_token: foo
				allow_invalid_certs: false
				use_prerelease: true
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
				packageName,
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestDefaults_Default(t *testing.T) {
	// GIVEN: a LookupDefault.
	defaults := Defaults{}
	want := Defaults{
		Type:              "github",
		AllowInvalidCerts: test.Ptr(false),
		UsePreRelease:     test.Ptr(false),
		Require: filter.RequireDefaults{
			Docker: docker.Defaults{
				Type: "hub",
				ContainerDetailDefaults: docker.ContainerDetailDefaults{
					Tag: "{{ version }}",
				},
			},
		},
	}

	// WHEN: Default is called.
	defaults.Default()

	wantStr := decode.ToYAMLString(want, "")
	gotStr := decode.ToYAMLString(defaults, "")
	// THEN: it should set the defaults as expected.
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

	// THEN: the struct is populated with default values.
	if d.Require.Docker.Defaults != &hd.Require.Docker {
		t.Errorf(
			"%s\nDefaults.SetDefaults() pointer mismatch\ngot:  %p\nwant: %p",
			packageName, d.Require.Docker.Defaults, hd.Require.Docker.Defaults,
		)
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	// GIVEN: a Defaults.
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
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)
		})
	}
}
