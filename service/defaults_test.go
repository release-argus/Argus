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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	dvbase "github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
)

func TestDefaults_IsZero(t *testing.T) {
	// GIVEN: Defaults.
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
					Base: opt.Base{
						Interval: "1m",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/LatestVersion",
			opt: &Defaults{
				LatestVersion: lvbase.Defaults{
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
			name: "non-empty/WebHook",
			opt: &Defaults{
				WebHook: map[string]struct{}{
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
					Base: opt.Base{
						Interval: "1m",
					},
				},
				LatestVersion: lvbase.Defaults{
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
				WebHook: map[string]struct{}{
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

			// WHEN: IsZero() is called on it.
			got := tc.opt.IsZero()

			// THEN: it should return the want value.
			if got != tc.want {
				t.Errorf(
					"%s\nDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDefaults_Unmarshal(t *testing.T) {
	// GIVEN: JSON and/or YAML string to unmarshal into Defaults.
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
				latest_version:
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
				packageName,
				"Defaults",
			); testErr != nil {
				t.Error(testErr)
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
				service:
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
			errRegex: "^$",
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
				packageName,
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestDefaults_String(t *testing.T) {
	// GIVEN: Defaults.
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
				LatestVersion: lvbase.Defaults{
					AccessToken:       "foo",
					AllowInvalidCerts: test.Ptr(true),
					UsePreRelease:     test.Ptr(false),
					Options: &opt.Defaults{
						Base: opt.Base{
							Interval: "1m",
						},
					},
					Require: filter.RequireDefaults{
						Docker: *test.Must(t, func() (*docker.Defaults, error) {
							return docker.DecodeDefaults(
								"yaml", []byte(test.TrimYAML(`
									type: ghcr
									registry:
										ghcr:
											image: imageGHCR
											tag: tagGHCR
											auth:
												username: usernameGHCR
												token: tokenGHCR
										hub:
											image: imageHub
											tag: tagHub
											auth:
												username: usernameHub
												token: tokenHub
										quay:
											image: imageQuay
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
													image: imageGHCRother
													tag: imageGHCRother
													auth:
														username: usernameGHCR_Other
														token: tokenGHCR_Other
												hub:
													image: imageHub_Other
													tag: tagHub_Other
													auth:
														username: usernameHub_Other
														token: tokenHub_Other
												quay:
													image: imageQuay_Other
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
				DeployedVersionLookup: dvbase.Defaults{
					AllowInvalidCerts: test.Ptr(false),
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
					access_token: foo
					allow_invalid_certs: true
					use_prerelease: false
					require:
						docker:
							type: ghcr
							registry:
								ghcr:
									image: imageGHCR
									tag: tagGHCR
									auth:
										token: tokenGHCR
								hub:
									image: imageHub
									tag: tagHub
									auth:
										username: usernameHub
										token: tokenHub
								quay:
									image: imageQuay
									auth:
										token: tokenQuay
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
				packageName,
				tc.defaults.String,
				tc.want,
			)
		})
	}
}

func TestDefaults_Default(t *testing.T) {
	// GIVEN: Defaults.
	d := &Defaults{}

	// WHEN: Default is called.
	d.Default()

	prefix := fmt.Sprintf("%s\nDefaults.Default()", packageName)

	// THEN: the struct is populated with default values.
	if d.Options.Interval != "10m" {
		t.Errorf(
			"%s .Options.Interval mismatch\ngot:  %s\nwant: 10m",
			prefix, d.Options.Interval,
		)
	}
	if d.Dashboard.AutoApprove == nil {
		t.Errorf("%s .Dashboard.AutoApprove should be non-nil", packageName)
	}

	// AND: the X.Options vars are pointing to the Options struct.
	if d.LatestVersion.Options != &d.Options {
		t.Errorf(
			"%s invalid .LatestVersion.Options pointer:\ngot:  %p\nwant: %p",
			packageName, d.LatestVersion.Options, &d.Options,
		)
	}
	if d.DeployedVersionLookup.Options != &d.Options {
		t.Errorf(
			"%s\ninvalid .DeployedVersionLookup.Options pointer:\ngot:  %p\nwant: %p",
			packageName, d.DeployedVersionLookup.Options, &d.Options,
		)
	}
}

func TestDefaults_SetDefaults(t *testing.T) {
	// GIVEN: Defaults.
	d := &Defaults{}

	// AND: HardDefaults.
	hd := &Defaults{}
	hd.Default()

	// WHEN: SetDefaults is called.
	d.SetDefaults(hd)

	prefix := fmt.Sprintf("%s\nDefaults.SetDefaults(&Defaults{})", packageName)

	// THEN: the struct is populated with default values.
	fieldTests := []test.FieldAssertion{
		{
			Name: "LatestVersion.Require.Docker.ContainerDetail.Defaults -> HardDefaults",
			Got:  d.LatestVersion.Require.Docker.ContainerDetail.Defaults,
			Want: &hd.LatestVersion.Require.Docker.ContainerDetail,
			Mode: test.CompareSamePointer,
		},
		{ // Layer 1: Registry HardDefaults
			Name: "L1: LatestVersion.Require.Docker.Registry.GHCR.ContainerDetail.Defaults -> HardDefaults...Registry.GHCR.ContainerDetail",
			Got:  d.LatestVersion.Require.Docker.Registry.GHCR.ContainerDetail.Defaults,
			Want: &hd.LatestVersion.Require.Docker.Registry.GHCR.ContainerDetail,
			Mode: test.CompareSamePointer,
		},
		{ // Layer 2: Root Defaults
			Name: "L2: LatestVersion.Require.Docker.Registry.GHCR.ContainerDetail.Defaults.Defaults -> Defaults.ContainerDetail",
			Got:  d.LatestVersion.Require.Docker.Registry.GHCR.ContainerDetail.Defaults.Defaults,
			Want: &d.LatestVersion.Require.Docker.ContainerDetail,
			Mode: test.CompareSamePointer,
		},
		{ // Layer 3: Root HardDefaults
			Name: "L3: LatestVersion.Require.Docker.Registry.GHCR.ContainerDetail.Defaults.Defaults.Defaults -> HardDefaults.ContainerDetail",
			Got:  d.LatestVersion.Require.Docker.Registry.GHCR.ContainerDetail.Defaults.Defaults.Defaults,
			Want: &hd.LatestVersion.Require.Docker.ContainerDetail,
			Mode: test.CompareSamePointer,
		},
		{ // Layer 1: Registry HardDefaults
			Name: "L1: LatestVersion.Require.Docker.Registry.Hub.ContainerDetail.Defaults -> HardDefaults...Registry.Hub.ContainerDetail",
			Got:  d.LatestVersion.Require.Docker.Registry.Hub.ContainerDetail.Defaults,
			Want: &hd.LatestVersion.Require.Docker.Registry.Hub.ContainerDetail,
			Mode: test.CompareSamePointer,
		},
		{ // Layer 2: Root Defaults
			Name: "L2: LatestVersion.Require.Docker.Registry.Hub.ContainerDetail.Defaults.Defaults -> Defaults.ContainerDetail",
			Got:  d.LatestVersion.Require.Docker.Registry.Hub.ContainerDetail.Defaults.Defaults,
			Want: &d.LatestVersion.Require.Docker.ContainerDetail,
			Mode: test.CompareSamePointer,
		},
		{ // Layer 3: Root HardDefaults
			Name: "L3: LatestVersion.Require.Docker.Registry.Hub.ContainerDetail.Defaults.Defaults.Defaults -> HardDefaults.ContainerDetail",
			Got:  d.LatestVersion.Require.Docker.Registry.Hub.ContainerDetail.Defaults.Defaults.Defaults,
			Want: &hd.LatestVersion.Require.Docker.ContainerDetail,
			Mode: test.CompareSamePointer,
		},
		{ // Layer 1: Registry HardDefaults
			Name: "L1: LatestVersion.Require.Docker.Registry.Quay.ContainerDetail.Defaults -> HardDefaults...Registry.Quay.ContainerDetail",
			Got:  d.LatestVersion.Require.Docker.Registry.Quay.ContainerDetail.Defaults,
			Want: &hd.LatestVersion.Require.Docker.Registry.Quay.ContainerDetail,
			Mode: test.CompareSamePointer,
		},
		{ // Layer 2: Root Defaults
			Name: "L2: LatestVersion.Require.Docker.Registry.Quay.ContainerDetail.Defaults.Defaults -> Defaults.ContainerDetail",
			Got:  d.LatestVersion.Require.Docker.Registry.Quay.ContainerDetail.Defaults.Defaults,
			Want: &d.LatestVersion.Require.Docker.ContainerDetail,
			Mode: test.CompareSamePointer,
		},
		{ // Layer 3: Root HardDefaults
			Name: "L3: LatestVersion.Require.Docker.Registry.Quay.ContainerDetail.Defaults.Defaults.Defaults -> HardDefaults.ContainerDetail",
			Got:  d.LatestVersion.Require.Docker.Registry.Quay.ContainerDetail.Defaults.Defaults.Defaults,
			Want: &hd.LatestVersion.Require.Docker.ContainerDetail,
			Mode: test.CompareSamePointer,
		},
		{
			Name: "LatestVersion.Options -> Options",
			Got:  d.LatestVersion.Options,
			Want: &d.Options,
			Mode: test.CompareSamePointer,
		},
		{
			Name: "DeployedVersionLookup.Options -> Options",
			Got:  d.DeployedVersionLookup.Options,
			Want: &d.Options,
			Mode: test.CompareSamePointer,
		},
		{
			Name: "LatestVersion.Options.Defaults -> HardDefaults.Options",
			Got:  hd.LatestVersion.Options,
			Want: &hd.Options,
			Mode: test.CompareSamePointer,
		},
		{
			Name: "DeployedVersionLookup.Options.Defaults -> HardDefaults.Options",
			Got:  hd.DeployedVersionLookup.Options,
			Want: &hd.Options,
			Mode: test.CompareSamePointer,
		},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "Defaults"); err != nil {
		t.Fatal(err)
	}
}
