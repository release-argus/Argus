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

package service

import (
	"testing"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	dvbase "github.com/release-argus/Argus/service/deployed_version/types/base"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
)

func TestDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		opt  *Defaults
		want bool
	}{
		{
			name: "empty",
			opt:  &Defaults{},
			want: true,
		},
		{
			name: "non-empty/Options",
			opt: &Defaults{
				Options: opt.Defaults{
					Interval: "1m",
				},
			},
			want: false,
		},
		{
			name: "non-empty/LatestVersion",
			opt: &Defaults{
				LatestVersion: latestver.Defaults{
					Type: "url",
				},
			},
			want: false,
		},
		{
			name: "non-empty/DeployedVersionLookup",
			opt: &Defaults{
				DeployedVersionLookup: dvbase.Defaults{
					Type: "url",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Notify",
			opt: &Defaults{
				Notify: map[string]struct{}{
					"foo": {},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Command",
			opt: &Defaults{
				Command: command.Commands{
					{"echo", "test"},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Webhook",
			opt: &Defaults{
				Webhook: map[string]struct{}{
					"bar": {},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Dashboard",
			opt: &Defaults{
				Dashboard: *test.Must(t, func() (*dashboard.Defaults, error) {
					return dashboard.DecodeDefaults("yaml", []byte("auto_approve: true"))
				}),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			opt: &Defaults{
				Options: opt.Defaults{
					Interval: "1m",
				},
				LatestVersion: latestver.Defaults{
					Type: "url",
				},
				DeployedVersionLookup: dvbase.Defaults{
					Type: "url",
				},
				Notify: map[string]struct{}{
					"foo": {},
				},
				Command: command.Commands{
					{"echo", "test"},
				},
				Webhook: map[string]struct{}{
					"bar": {},
				},
				Dashboard: *test.Must(t, func() (*dashboard.Defaults, error) {
					return dashboard.DecodeDefaults("yaml", []byte("auto_approve: true"))
				}),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.opt.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestDefaults_Unmarshal(t *testing.T) {
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
				^jsontext:
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
			name:     "YAML/DefaultsDecode invalid data type",
			format:   "yaml",
			data:     "options: false",
			errRegex: `^[^\s]+ boolean was used.*`,
		},
		{
			name:   "YAML/LatestVersion invalid data type",
			format: "yaml",
			data:   "latest_version: false",
			errRegex: test.TrimYAML(`
				^latest_version:
					[^\s]+ boolean was used.*`,
			),
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"options": {
					"interval": "1m"
				},
				"latest_version": {
					"type": "url"
				},
				"deployed_version": {
					"type": "url"
				},
				"notify": {
					"foo": {}
				},
				"command": [
					["echo", "test"]
				],
				"webhook": {
					"bar": {}
				},
				"dashboard": {
					"auto_approve": true
				}
			}`),
			want: test.TrimYAML(`
				options:
					interval: 1m
				latest_version:
					type: url
				deployed_version:
					type: url
				notify:
					foo: {}
				command:
					- - echo
						- test
				webhook:
					bar: {}
				dashboard:
					auto_approve: true
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				options:
					interval: 1m
				latest_version:
					type: url
				deployed_version:
					type: url
				notify:
					foo: {}
				command:
					- ["echo", "test"]
				webhook:
					bar: {}
				dashboard:
					auto_approve: true
			`),
			want: test.TrimYAML(`
				options:
					interval: 1m
				latest_version:
					type: url
				deployed_version:
					type: url
				notify:
					foo: {}
				command:
					- - echo
						- test
				webhook:
					bar: {}
				dashboard:
					auto_approve: true
			`),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Defaults, error) {
					var zero Defaults
					err := decode.Unmarshal(format, data, &zero)
					return &zero, err
				},
				tc.format, tc.data,
				func(d *Defaults) string { return d.String("") },
				tc.want,
				tc.errRegex,
				"Defaults.Unmarshal",
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
		want         string
		errRegex     string
	}{
		{
			name:   "JSON/empty",
			format: "json",
			data:   "",
			errRegex: test.TrimYAML(`
				^service:
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
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"options": {
					"interval": "1m"
				},
				"latest_version": {
					"type": "github"
				},
				"deployed_version": {
					"type": "url"
				},
				"notify": {
					"foo": {}
				},
				"command": [
					["echo", "test"]
				],
				"webhook": {
					"bar": {}
				},
				"dashboard": {
					"auto_approve": true
				}
			}`),
			want: test.TrimYAML(`
				options:
					interval: 1m
				latest_version:
					type: github
				deployed_version:
					type: url
				notify:
					foo: {}
				command:
					- - echo
					  - test
				webhook:
					bar: {}
				dashboard:
					auto_approve: true
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				options:
					interval: 1m
				latest_version:
					type: github
				deployed_version:
					type: url
				notify:
					foo: {}
				command:
					- ["echo", "test"]
				webhook:
					bar: {}
				dashboard:
					auto_approve: true
			`),
			want: test.TrimYAML(`
				options:
					interval: 1m
				latest_version:
					type: github
				deployed_version:
					type: url
				notify:
					foo: {}
				command:
					- - echo
						- test
				webhook:
					bar: {}
				dashboard:
					auto_approve: true
			`),
			errRegex: `^$`,
		},
		{
			name:   "JSON/invalid latest_version format",
			format: "json",
			data: test.TrimJSON(`{
				"options": {
					"interval": "1m"
				},
				"latest_version": "abc",
				"deployed_version": {
					"type": "url"
				},
				"notify": {
					"foo": {}
				},
				"command": [
					["echo", "test"]
				],
				"webhook": {
					"bar": {}
				},
				"dashboard": {
					"auto_approve": true
				}
			}`),
			errRegex: test.TrimYAML(`
				^service:
					latest_version:
						json: .*unmarshal.* string.*$`,
			),
		},
		{
			name:   "YAML/invalid latest_version format",
			format: "yaml",
			data: test.TrimYAML(`
				options:
					interval: 1m
				latest_version: abc
				deployed_version:
					type: url
				notify:
					foo: {}
				command:
					- ["echo", "test"]
				webhook:
					bar: {}
				dashboard:
					auto_approve: true
			`),
			errRegex: test.TrimYAML(`
				^service:
					latest_version:
						[^\s]+.* string .*`,
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
				func(v *Defaults) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestDefaults_String(t *testing.T) {
	tests := []struct {
		name     string
		defaults *Defaults
		want     string
	}{
		{
			name:     "nil",
			defaults: nil,
			want:     "",
		},
		{
			name:     "empty",
			defaults: &Defaults{},
			want:     "{}\n",
		},
		{
			name: "filled",
			defaults: &Defaults{
				Options: *test.Must(t, func() (*opt.Defaults, error) {
					return opt.DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							interval: 1m
							semantic_versioning: false
						`)),
					)
				}),
				LatestVersion: latestver.Defaults{
					GitHub: github.Defaults{
						AccessToken:   "foo",
						UsePreRelease: new(false),
					},
					URL: web.Defaults{
						AllowInvalidCerts: new(true),
					},
					Common: lvbase.Defaults{
						Options: &opt.Defaults{
							Interval: "1m",
						},
						Require: filter.RequireDefaults{
							Docker: *test.Must(t, func() (*docker.Defaults, error) {
								return docker.DecodeDefaults(
									"yaml", []byte(test.TrimYAML(`
										type: ghcr
										registry:
											ghcr:
												auth:
													username: usernameGHCR
													token: tokenGHCR
											hub:
												auth:
													username: usernameHub
													token: tokenHub
											quay:
												auth:
													username: usernameQuay
													token: tokenQuay
									`)),
									test.Must(t, func() (*docker.Defaults, error) {
										return docker.DecodeDefaults(
											"yaml", []byte(test.TrimYAML(`
												type: ghcr
												registry:
													ghcr:
														auth:
															username: usernameGHCR_Other
															token: tokenGHCR_Other
													hub:
														auth:
															username: usernameHub_Other
															token: tokenHub_Other
													quay:
														auth:
															username: usernameQuay_Other
															token: tokenQuay_Other
											`)),
											nil,
										)
									}),
								)
							}),
						},
					},
				},
				DeployedVersionLookup: dvbase.Defaults{
					AllowInvalidCerts: new(false),
				},
				Dashboard: *test.Must(t, func() (*dashboard.Defaults, error) {
					return dashboard.DecodeDefaults("yaml", []byte("auto_approve: true"))
				}),
			},
			want: test.TrimYAML(`
				options:
					interval: 1m
					semantic_versioning: false
				latest_version:
					common:
						require:
							docker:
								type: ghcr
								registry:
									ghcr:
										auth:
											token: tokenGHCR
									hub:
										auth:
											username: usernameHub
											token: tokenHub
									quay:
										auth:
											token: tokenQuay
					github:
						access_token: foo
						use_prerelease: false
					url:
						allow_invalid_certs: true
				deployed_version:
					allow_invalid_certs: false
				dashboard:
					auto_approve: true
				`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				"Defaults.String",
				tc.defaults.String,
				tc.want,
			)
		})
	}
}

func TestDefaults_Default(t *testing.T) {
	d := &Defaults{}

	d.Default()

	if got, want := d.Options.Interval, "10m"; got != want {
		t.Errorf("Defaults.Options.Interval mismatch\ngot:  %q\nwant: %q", got, want)
	}
	if d.Dashboard.AutoApprove == nil {
		t.Error("Defaults.Dashboard.AutoApprove mismatch\ngot:  nil\nwant: non-nil")
	}

	if testErr := test.AssertFields(
		t,
		[]test.FieldAssertion{
			{Name: "LatestVersion.Options -> Options", Got: d.LatestVersion.Common.Options, Want: &d.Options, Mode: test.CompareSamePointer},
			{Name: "DeployedVersionLookup.Options -> Options", Got: d.DeployedVersionLookup.Options, Want: &d.Options, Mode: test.CompareSamePointer},
		},
		"Defaults",
	); testErr != nil {
		t.Fatal(testErr)
	}
}

func TestDefaults_SetDefaults(t *testing.T) {
	d := &Defaults{}

	hd := &Defaults{}
	hd.Default()

	d.SetDefaults(hd)

	fieldTests := []test.FieldAssertion{
		{Name: "LatestVersion.Require.Docker.ContainerDetailDefaults.Defaults -> HardDefaults", Got: d.LatestVersion.Common.Require.Docker.ContainerDetailDefaults.Defaults, Want: &hd.LatestVersion.Common.Require.Docker.ContainerDetailDefaults, Mode: test.CompareSamePointer},
		{Name: "LatestVersion.Require.Docker.Registry.GHCR.Auth.Defaults -> HardDefaults...Registry.GHCR.Auth", Got: d.LatestVersion.Common.Require.Docker.Registry.GHCR.GetAuth().Defaults(), Want: hd.LatestVersion.Common.Require.Docker.Registry.GHCR.GetAuth(), Mode: test.CompareSamePointer},
		{Name: "LatestVersion.Require.Docker.Registry.Hub.Auth.Defaults -> HardDefaults...Registry.Hub.Auth", Got: d.LatestVersion.Common.Require.Docker.Registry.Hub.GetAuth().Defaults(), Want: hd.LatestVersion.Common.Require.Docker.Registry.Hub.GetAuth(), Mode: test.CompareSamePointer},
		{Name: "LatestVersion.Require.Docker.Registry.Quay.Auth.Defaults -> HardDefaults...Registry.Quay.Auth", Got: d.LatestVersion.Common.Require.Docker.Registry.Quay.GetAuth().Defaults(), Want: hd.LatestVersion.Common.Require.Docker.Registry.Quay.GetAuth(), Mode: test.CompareSamePointer},
		{Name: "LatestVersion.Options -> Options", Got: d.LatestVersion.Common.Options, Want: &d.Options, Mode: test.CompareSamePointer},
		{Name: "DeployedVersionLookup.Options -> Options", Got: d.DeployedVersionLookup.Options, Want: &d.Options, Mode: test.CompareSamePointer},
		{Name: "HardDefaults.LatestVersion.Options -> HardDefaults.Options", Got: hd.LatestVersion.Common.Options, Want: &hd.Options, Mode: test.CompareSamePointer},
		{Name: "HardDefaults.DeployedVersionLookup.Options -> HardDefaults.Options", Got: hd.DeployedVersionLookup.Options, Want: &hd.Options, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, "Defaults"); testErr != nil {
		t.Fatal(testErr)
	}
}
