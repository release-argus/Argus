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

package docker

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// #######################
// # REGISTRY | DECODING #
// #######################

func TestCommonRegistryDefaults_Unmarshal(t *testing.T) {
	// GIVEN: a CommonRegistryDefaults and JSON/YAML to unmarshal into it.
	tests := []struct {
		name     string
		format   string
		data     string
		registry *CommonRegistryDefaults
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "{}",
			registry: &CommonRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &CommonRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &CommonRegistryDefaults{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &CommonRegistryDefaults{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:     "JSON/invalid ContainerDetail",
			format:   "json",
			data:     `{"image": []}`,
			registry: &CommonRegistryDefaults{},
			errRegex: `^json: .*unmarshal.*`,
		},
		{
			name:     "YAML/invalid ContainerDetail",
			format:   "yaml",
			data:     `image: []`,
			registry: &CommonRegistryDefaults{},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &CommonRegistryDefaults{
				Auth: &GHCRAuthDefaults{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					json: .*unmarshal.*`,
			),
		},
		{
			name:   "YAML/invalid Auth",
			format: "yaml",
			data:   `auth: []`,
			registry: &CommonRegistryDefaults{
				Auth: &GHCRAuthDefaults{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					[^\s]+ sequence was used where mapping is expected`,
			),
		},
		{
			name:   "JSON/auth-ghcr",
			format: "json",
			data: test.TrimJSON(`{
				"image": "i",
				"tag": "t",
				"auth": {
					"username": "ghcr-username",
					"token": "tOKEn"
				}
			}`),
			registry: &CommonRegistryDefaults{
				Auth: &GHCRAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					token: tOKEn
			`),
		},
		{
			name:   "YAML/auth-ghcr",
			format: "yaml",
			data: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: ghcr-username
					token: tOKEn
			`),
			registry: &CommonRegistryDefaults{
				Auth: &GHCRAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					token: tOKEn
			`),
		},
		{
			name:   "JSON/auth-hub",
			format: "json",
			data: test.TrimJSON(`{
				"image": "i",
				"tag": "t",
				"auth": {
					"username": "hub-username",
					"token": "tOKEn"
				}
			}`),
			registry: &CommonRegistryDefaults{
				Auth: &HubAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: hub-username
					token: tOKEn
			`),
		},
		{
			name:   "YAML/auth-hub",
			format: "yaml",
			data: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: hub-username
					token: tOKEn
			`),
			registry: &CommonRegistryDefaults{
				Auth: &HubAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: hub-username
					token: tOKEn
			`),
		},
		{
			name:   "JSON/auth-quay",
			format: "json",
			data: test.TrimJSON(`{
				"image": "i",
				"tag": "t",
				"auth": {
					"username": "quay-username",
					"token": "tOKEn"
				}
			}`),
			registry: &CommonRegistryDefaults{
				Auth: &QuayAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					token: tOKEn
			`),
		},
		{
			name:   "YAML/auth-quay",
			format: "yaml",
			data: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: quay-username
					token: tOKEn
			`),
			registry: &CommonRegistryDefaults{
				Auth: &QuayAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					token: tOKEn
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, testErr := test.AssertUnmarshal(
				t,
				tc.format, tc.data,
				tc.registry,
				tc.errRegex,
				func(v *CommonRegistryDefaults) string { return decode.ToYAMLString(v, "") },
				tc.want,
				packageName,
				"CommonRegistryDefaults",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestCommonRegistry_Unmarshal(t *testing.T) {
	// GIVEN: a CommonRegistry and JSON/YAML to unmarshal into it.
	tests := []struct {
		name     string
		format   string
		data     string
		registry *CommonRegistry
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "{}",
			registry: &CommonRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &CommonRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &CommonRegistry{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &CommonRegistry{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:     "JSON/invalid ContainerDetail",
			format:   "json",
			data:     `{"image": []}`,
			registry: &CommonRegistry{},
			errRegex: `^json: .*unmarshal.*`,
		},
		{
			name:     "YAML/invalid ContainerDetail",
			format:   "yaml",
			data:     `image: []`,
			registry: &CommonRegistry{},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &CommonRegistry{
				Auth: &GHCRAuth{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					json: .*unmarshal.*`,
			),
		},
		{
			name:   "YAML/invalid Auth",
			format: "yaml",
			data:   `auth: []`,
			registry: &CommonRegistry{
				Auth: &GHCRAuth{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					[^\s]+ sequence was used where mapping is expected`,
			),
		},
		{
			name:   "JSON/auth-ghcr",
			format: "json",
			data: test.TrimJSON(`{
				"image": "i",
				"tag": "t",
				"auth": {
					"username": "ghcr-username",
					"token": "tOKEn"
				}
			}`),
			registry: &CommonRegistry{
				Auth: &GHCRAuth{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					token: tOKEn
			`),
		},
		{
			name:   "YAML/auth-ghcr",
			format: "yaml",
			data: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: ghcr-username
					token: tOKEn
			`),
			registry: &CommonRegistry{
				Auth: &GHCRAuth{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					token: tOKEn
			`),
		},
		{
			name:   "JSON/auth-hub",
			format: "json",
			data: test.TrimJSON(`{
				"image": "i",
				"tag": "t",
				"auth": {
					"username": "hub-username",
					"token": "tOKEn"
				}
			}`),
			registry: &CommonRegistry{
				Auth: &HubAuth{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: hub-username
					token: tOKEn
			`),
		},
		{
			name:   "YAML/auth-hub",
			format: "yaml",
			data: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: hub-username
					token: tOKEn
			`),
			registry: &CommonRegistry{
				Auth: &HubAuth{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: hub-username
					token: tOKEn
			`),
		},
		{
			name:   "JSON/auth-quay",
			format: "json",
			data: test.TrimJSON(`{
				"image": "i",
				"tag": "t",
				"auth": {
					"username": "quay-username",
					"token": "tOKEn"
				}
			}`),
			registry: &CommonRegistry{
				Auth: &QuayAuth{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					token: tOKEn
			`),
		},
		{
			name:   "YAML/auth-quay",
			format: "yaml",
			data: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: quay-username
					token: tOKEn
			`),
			registry: &CommonRegistry{
				Auth: &QuayAuth{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					token: tOKEn
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, testErr := test.AssertUnmarshal(
				t,
				tc.format, tc.data,
				tc.registry,
				tc.errRegex,
				func(v *CommonRegistry) string { return v.String("") },
				tc.want,
				packageName,
				"CommonRegistry",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// ####################
// # REGISTRY | STATE #
// ####################

func TestCommonRegistry_IsZero(t *testing.T) {
	// GIVEN: a CommonRegistry.
	tests := []struct {
		name     string
		registry *CommonRegistry
		want     bool
	}{
		{
			name:     "nil",
			registry: nil,
			want:     true,
		},
		{
			name:     "empty",
			registry: &CommonRegistry{},
			want:     true,
		},
		{
			name: "have type",
			registry: &CommonRegistry{
				Type: "hub",
			},
			want: false,
		},
		{
			name: "have image",
			registry: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app",
				},
			},
			want: false,
		},
		{
			name: "have tag",
			registry: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "1.2.3",
				},
			},
			want: false,
		},
		{
			name: "have image+tag",
			registry: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app",
					Tag:   "1.2.3",
				},
			},
			want: false,
		},
		{
			name: "have auth",
			registry: &CommonRegistry{
				Auth: &HubAuth{
					HubAuthDefaults: HubAuthDefaults{
						Token:      "foo",
						queryToken: "bar",
						validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: false,
		},
		{
			name: "have type, have image/tag, have auth",
			registry: &CommonRegistry{
				Type: "hub",
				ContainerDetail: ContainerDetail{
					Image: "test/app",
					Tag:   "1.2.3",
				},
				Auth: &HubAuth{
					HubAuthDefaults: HubAuthDefaults{
						Token:      "foo",
						queryToken: "bar",
						validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called with it.
			got := tc.registry.IsZero()

			// THEN: true is returned if all fields are empty.
			if got != tc.want {
				t.Errorf(
					"%s\nCommonRegistry.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestCommonRegistry_Clone(t *testing.T) {
	// GIVEN: a CommonRegistry.
	tests := []struct {
		name     string
		registry *CommonRegistry
		want     string
	}{
		{
			name:     "nil",
			registry: nil,
			want:     "null\n",
		},
		{
			name:     "empty",
			registry: &CommonRegistry{},
			want:     "{}\n",
		},
		{
			name: "filled",
			registry: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: test.ArgusDockerGHCRRepo,
					Tag:   "1.2.3",
				},
				Auth: &HubAuth{
					HubAuthDefaults: HubAuthDefaults{
						Username:   "u1",
						Token:      "t1",
						queryToken: "qT",
						validUntil: time.Now(),
						defaults:   &HubAuthDefaults{},
					},
				},
			},
			want: test.TrimYAML(`
				image: ` + test.ArgusDockerGHCRRepo + `
				tag: 1.2.3
				auth:
					username: u1
					token: t1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy() is called on it.
			result := tc.registry.Clone()

			prefix := fmt.Sprintf("%s\nCommonRegistry.Clone()", packageName)

			// THEN: the returned RegistryAuth unmarshals the same.
			if got := decode.ToYAMLString(result, ""); got != tc.want {
				t.Fatalf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
			if result == nil {
				return
			}

			// AND: the defaults are at the same address.
			if got := result.defaults; got != tc.registry.defaults {
				t.Fatalf(
					"%s .defaults pointer lost for %q",
					prefix, tc.name,
				)
			}

			// AND: the returned GHCRAuth is at a different address.
			if result.Auth == tc.registry.Auth && (result.Auth != nil || tc.registry.Auth != nil) {
				t.Fatalf(
					"%s pointer to same Auth instance for %q",
					prefix, tc.name,
				)
			}
		})
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

func TestCommonRegistry_String(t *testing.T) {
	// GIVEN: a CommonRegistry.
	tests := []struct {
		name     string
		registry *CommonRegistry
		want     string
	}{
		{
			name:     "nil",
			registry: nil,
			want:     "",
		},
		{
			name:     "empty",
			registry: &CommonRegistry{},
			want:     "{}\n",
		},
		{
			name: "filled",
			registry: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app",
					Tag:   "{{ version }}",
				},
				Auth: &HubAuth{
					HubAuthDefaults: HubAuthDefaults{
						Token:      "_token_",
						queryToken: "_queryToken_",
						validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
						defaults: &HubAuthDefaults{
							Token:      "_other_token_",
							queryToken: "_other_queryToken_",
							validUntil: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					},
				},
				defaults: &HubRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						Auth: &HubAuthDefaults{
							Token:      "_other_token_",
							queryToken: "_other_queryToken_",
							validUntil: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
			want: test.TrimYAML(`
				image: test/app
				tag: '{{ version }}'
				auth:
					token: _token_
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.registry.String,
				tc.want,
			)
		})
	}
}

// #######################
// # REGISTRY | DEFAULTS #
// #######################

func TestCommonRegistryDefaults_Defaults(t *testing.T) {
	// GIVEN: a fresh CommonRegistryDefaults.
	var registry CommonRegistryDefaults

	// WHEN: Defaults is called on it.
	got := registry.Defaults()

	// THEN: nil is received.
	if got != nil {
		t.Errorf(
			"%s\nfresh CommonRegistryDefaults\ngot:  %v\nwant: nil",
			packageName, got,
		)
	}

	for _, dType := range PossibleTypes {
		t.Run(dType, func(t *testing.T) {
			// GIVEN: Defaults.
			defaults := Defaults{
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &HubAuthDefaults{
								Token: "ghcr-token",
							},
						},
					},
					Hub: &HubRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &HubAuthDefaults{
								Token:    "hub-token",
								Username: "hub-username",
							},
						},
					},
					Quay: &QuayRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &HubAuthDefaults{
								Token: "hub-token",
							},
						},
					},
				},
			}
			registry.Auth = RegistryMap[dType]().GetAuth()
			rDefaults := getRegistryDefaults(dType, &defaults)
			rootContainerDetail := ContainerDetail{}

			// WHEN: SetDefaults is called with them.
			registry.SetDefaults(rDefaults, &rootContainerDetail, &defaults.ContainerDetail)

			prefix := fmt.Sprintf(
				"%s\nSetDefaults(type=%q, defaults=%+v)\n",
				packageName, dType, defaults,
			)

			// THEN: Defaults is set to the corresponding type defaults.
			got := registry.Defaults()
			var want RegistryDefaults
			if want = getRegistryDefaults(dType, &defaults); want == nil {
				t.Fatalf(
					"%s expected defaults for registry type %q, but none found in %+v",
					prefix, dType, defaults,
				)
			}
			if got != want {
				t.Fatalf(
					"%s .Defaults() pointer mismatch\ngot:  %p\nwant: %p",
					prefix, got, want,
				)
			}

			// AND: Defaults are handed out as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "GetContainerDetail()", Got: got.GetContainerDetail(), Want: want.GetContainerDetail(), Mode: test.CompareSamePointer},
				{Name: "GetContainerDetail().Defaults", Got: got.GetContainerDetail().Defaults, Want: want.GetContainerDetail().Defaults, Mode: test.CompareSamePointer},
				{Name: "GetAuth()", Got: got.GetAuth(), Want: want.GetAuth(), Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "CommonRegistryDefaults"); err != nil {
				t.Fatal(err)
			}

			// AND: ContainerDetail is chained as expected.
			fieldTests = []test.FieldAssertion{
				{ // Layer 1: Registry HardDefaults
					Name: "L1: Registry.ContainerDetail.Defaults -> HardDefaults.Registry.ContainerDetail",
					Got:  registry.ContainerDetail.Defaults,
					Want: rDefaults.GetContainerDetail(),
					Mode: test.CompareSamePointer,
				},
				{ // Layer 2: Root Defaults
					Name: "L2: LatestVersion.Require.Docker.Registry.GHCR.ContainerDetail.Defaults.Defaults -> Defaults.ContainerDetail",
					Got:  registry.ContainerDetail.Defaults.Defaults,
					Want: &rootContainerDetail,
					Mode: test.CompareSamePointer,
				},
				{ // Layer 3: Root HardDefaults
					Name: "L3: LatestVersion.Require.Docker.Registry.GHCR.ContainerDetail.Defaults.Defaults.Defaults -> HardDefaults.ContainerDetail",
					Got:  registry.ContainerDetail.Defaults.Defaults.Defaults,
					Want: &defaults.ContainerDetail,
					Mode: test.CompareSamePointer,
				},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "CommonRegistryDefaults.ContainerDetail"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCommonRegistryDefaults_Defaults__nil(t *testing.T) {
	// GIVEN: a fresh CommonRegistryDefaults.
	var registry CommonRegistryDefaults

	// WHEN: Defaults is called on it.
	got := registry.Defaults()

	// THEN: nil is received.
	if got != nil {
		t.Errorf(
			"%s\nfresh CommonRegistryDefaults\ngot:  %v\nwant: nil",
			packageName, got,
		)
	}

	// ---

	// GIVEN: nil Defaults.
	var defaults *Defaults

	for _, dType := range PossibleTypes {
		var rDefaults RegistryDefaults
		registry.Auth = RegistryMap[dType]().GetAuth()

		// WHEN: SetDefaults is called with them.
		registry.SetDefaults(
			rDefaults,
			&ContainerDetail{},
			&ContainerDetail{},
		)

		prefix := fmt.Sprintf(
			"%s\nSetDefaults(%q)",
			packageName, dType,
		)

		// THEN: Defaults remain nil.
		if registry.Defaults(); got != nil {
			t.Fatalf(
				"%s Registry defaults mismatch\ngot:  %v\nwant: nil",
				prefix, got,
			)
		}
	}

	// ---

	// GIVEN: Defaults with nil Registry.* values.
	defaults = &Defaults{
		Registry: RegistryDefaultsSet{
			GHCR: nil,
			Hub:  nil,
			Quay: nil,
		},
	}
	for _, dType := range PossibleTypes {
		registry.Auth = RegistryMap[dType]().GetAuth()

		rDefaults := getRegistryDefaults(dType, defaults)

		// WHEN: SetDefaults is called with them.
		registry.SetDefaults(
			rDefaults,
			&ContainerDetail{},
			&ContainerDetail{},
		)

		// THEN: Defaults is set to the corresponding type defaults.
		if got := registry.Defaults(); got != nil {
			t.Fatalf(
				"%s\nDefaults mismatch after SetDefaults(%q)\ngot:  %v\nwant: nil",
				packageName, dType, got,
			)
		}

		// AND: Defaults of Auth are set to the corresponding type defaults.
		if got := registry.Auth.Defaults(); got != nil {
			t.Fatalf(
				"%s\nAuth defaults mismatch after SetDefaults(%q)\ngot:  %v\nwant: nil",
				packageName, dType, got,
			)
		}
	}
}

func TestCommonRegistry_Defaults(t *testing.T) {
	// GIVEN: a fresh CommonRegistry.
	var registry CommonRegistry

	// WHEN: Defaults is called on it.
	got := registry.Defaults()

	// THEN: nil is received.
	if got != nil {
		t.Errorf(
			"%s\nfresh CommonRegistry\ngot:  %v\nwant: nil",
			packageName, got,
		)
	}

	for _, dType := range PossibleTypes {
		t.Run(dType, func(t *testing.T) {
			// GIVEN: registryDefaults.
			defaults := Defaults{
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &HubAuthDefaults{
								Token: "ghcr-token",
							},
						},
					},
					Hub: &HubRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &HubAuthDefaults{
								Token:    "hub-token",
								Username: "hub-username",
							},
						},
					},
					Quay: &QuayRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &HubAuthDefaults{
								Token: "hub-token",
							},
						},
					},
				},
			}
			registry.Auth = RegistryMap[dType]().GetAuth()

			// WHEN: SetDefaults is called with them.
			registry.SetDefaults(dType, &defaults)

			prefix := fmt.Sprintf(
				"%s\nSetDefaults(type=%q, defaults=%+v)",
				packageName, dType, defaults,
			)

			// THEN: Defaults is set to the corresponding type defaults.
			got := registry.Defaults()
			var want RegistryDefaults
			if want = getRegistryDefaults(dType, &defaults); want == nil {
				t.Fatalf(
					"%s\nUnknown registry type: %q",
					prefix, dType,
				)
			}
			if got != want {
				t.Fatalf(
					"%s .Defaults mismatch\ngot:  %v\nwant: %v",
					prefix, got, want,
				)
			}

			// AND: Defaults of Auth are set to the corresponding type defaults.
			gotAuth := got.GetAuth()
			wantAuth := want.GetAuth()
			if gotAuth != wantAuth {
				t.Fatalf(
					"%s .Defaults.Auth mismatch\ngot:  %v\nwant: %v",
					prefix, gotAuth, wantAuth,
				)
			}
		})
	}
}

func TestCommonRegistry_Defaults__nil(t *testing.T) {
	// GIVEN: a fresh CommonRegistry.
	var registry CommonRegistry

	// WHEN: SetDefaults is called on it with nil Defaults.
	dType := PossibleTypes[0]
	registry.SetDefaults(dType, nil)

	// THEN: the Defaults are unchanged.
	if got := registry.Defaults(); got != nil {
		t.Fatalf("%s\nDefaults remain nil after SetDefaults with nil Defaults", packageName)
	}
}

func TestCommonRegistry_SetDefaults__UnknownType(t *testing.T) {
	dockerType := "ghcr"
	// GIVEN: a CommonRegistry with Defaults.
	registry := RegistryMap[dockerType]()
	defaults := Defaults{
		Registry: RegistryDefaultsSet{
			GHCR: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{
						Token: "ghcr-token",
					},
				},
			},
			Hub: &HubRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{
						Token:    "hub-token",
						Username: "hub-username",
					},
				},
			},
			Quay: &QuayRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{
						Token: "hub-token",
					},
				},
			},
		},
	}
	prefix := fmt.Sprintf(
		"%s\nSetDefaults(type=%q, defaults=%+v)",
		packageName, dockerType, defaults,
	)

	registry.SetDefaults(dockerType, &defaults)
	originalDefaults := registry.Defaults()
	if originalDefaults == nil {
		t.Fatalf("%s .Defaults() should not be nil for known type", prefix)
	}

	// WHEN: SetDefaults is called with an unknown type.
	registry.SetDefaults("unknown", &defaults)

	prefix = fmt.Sprintf(
		"%s\nSetDefaults(type=\"unknown\", defaults=%+v)",
		packageName, defaults,
	)

	// THEN: The defaults are unchanged.
	if registry.Defaults() != originalDefaults {
		t.Fatalf("%s .Defaults() should remain unchanged because of the unknown type", prefix)
	}

	// WHEN: SetDefaults is called with a known type, but no defaults exist for this type.
	switch dockerType {
	case "ghcr":
		defaults.Registry.GHCR = nil
	case "hub":
		defaults.Registry.Hub = nil
	case "quay":
		defaults.Registry.Quay = nil
	}
	registry.SetDefaults(dockerType, &defaults)

	prefix = fmt.Sprintf(
		"%s\nSetDefaults(type=%q, defaults=%+v)",
		packageName, dockerType, defaults,
	)

	// THEN: The defaults are unchanged.
	if registry.Defaults() != originalDefaults {
		t.Fatalf(
			"%s .Defaults() should remain unchanged when known type has nil defaults",
			packageName,
		)
	}
}

// #######################
// # REGISTRY | METADATA #
// #######################

func TestCommonRegistry_GetTypeSelf(t *testing.T) {
	// GIVEN: a CommonRegistry.
	tests := []struct {
		name     string
		registry *CommonRegistry
		want     string
	}{
		{
			name:     "no type",
			registry: &CommonRegistry{},
		},
		{
			name: "have type",
			registry: &CommonRegistry{
				Type: "hi",
			},
			want: "hi",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetTypeSelf is called on it.
			got := tc.registry.GetTypeSelf()

			// THEN: the type is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nGetTypeSelf() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestCommonRegistry_GetImageSelf(t *testing.T) {
	// GIVEN: a CommonRegistry.
	tests := []struct {
		name     string
		registry CommonRegistry
		want     string
	}{
		{
			name: "GHCR",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "image-here",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "other-image-here",
						},
					},
				},
			},
			want: "image-here",
		},
		{
			name: "Hub",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "image-here",
				},
				defaults: &HubRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "other-image-here",
						},
					},
				},
			},
			want: "image-here",
		},
		{
			name: "Quay",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "image-here",
				},
				defaults: &QuayRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "other-image-here",
						},
					},
				},
			},
			want: "image-here",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetImageSelf is called on it.
			got := tc.registry.GetImageSelf()

			// THEN: the image is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nCommonRegistry.GetImageSelf() mismatch\ngot:  %q\nwant: %q",
					tc.name, got, tc.want,
				)
			}
		})
	}
}

func TestCommonRegistry_GetImage(t *testing.T) {
	// GIVEN: a DockerCheck with an image.
	tests := []struct {
		name     string
		registry CommonRegistry
		want     string
	}{
		{
			name: "empty image",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "",
						},
					},
				},
			},
			want: "",
		},
		{
			name: "Image from root",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app-root",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "",
						},
					},
				},
			},
			want: "test/app-root",
		},
		{
			name: "Image from defaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "test/app-defaults",
						},
						defaults: &GHCRRegistryDefaults{
							CommonRegistryDefaults: CommonRegistryDefaults{
								ContainerDetail: ContainerDetail{
									Image: "",
								},
							},
						},
					},
				},
			},
			want: "test/app-defaults",
		},
		{
			name: "Image from hardDefaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "",
						},
						defaults: &GHCRRegistryDefaults{
							CommonRegistryDefaults: CommonRegistryDefaults{
								ContainerDetail: ContainerDetail{
									Image: "test/app-hardDefaults",
								},
							},
						},
					},
				},
			},
			want: "test/app-hardDefaults",
		},
		{
			name: "Image from root, ignore defaults/hardDefaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app-root",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "test/app-defaults",
						},
						defaults: &GHCRRegistryDefaults{
							CommonRegistryDefaults: CommonRegistryDefaults{
								ContainerDetail: ContainerDetail{
									Image: "test/app-hardDefaults",
								},
							},
						},
					},
				},
			},
			want: "test/app-root",
		},
		{
			name: "Image from defaults, ignore hardDefaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "test/app-defaults",
						},
						defaults: &GHCRRegistryDefaults{
							CommonRegistryDefaults: CommonRegistryDefaults{
								ContainerDetail: ContainerDetail{
									Image: "test/app-hardDefaults",
								},
							},
						},
					},
				},
			},
			want: "test/app-defaults",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetImage is called on it.
			got := tc.registry.GetImage()

			// THEN: the expected Tag is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nCommonRegistry.GetImage() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestCommonRegistry_GetTagSelf(t *testing.T) {
	// GIVEN: a CommonRegistry.
	tests := []struct {
		name     string
		registry CommonRegistry
		want     string
	}{
		{
			name: "GHCR",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "tag-here",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "other-tag-here",
						},
					},
				},
			},
			want: "tag-here",
		},
		{
			name: "Hub",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "tag-here",
				},
				defaults: &HubRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "other-tag-here",
						},
					},
				},
			},
			want: "tag-here",
		},
		{
			name: "Quay",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "tag-here",
				},
				defaults: &QuayRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "other-tag-here",
						},
					},
				},
			},
			want: "tag-here",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetTagSelf is called on it.
			got := tc.registry.GetTagSelf()

			// THEN: the tag is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nCommonRegistry.GetTagSelf() mismatch\ngot:  %q\nwant: %q",
					tc.name, got, tc.want,
				)
			}
		})
	}
}

func TestCommonRegistry_GetTag(t *testing.T) {
	// GIVEN: a DockerCheck with a tag.
	tests := []struct {
		name     string
		registry CommonRegistry
		want     string
	}{
		{
			name: "empty tag",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "",
						},
					},
				},
			},
			want: "",
		},
		{
			name: "Tag from root",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "test/app-root",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "",
						},
					},
				},
			},
			want: "test/app-root",
		},
		{
			name: "Tag from defaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "test/app-defaults",
						},
						defaults: &GHCRRegistryDefaults{
							CommonRegistryDefaults: CommonRegistryDefaults{
								ContainerDetail: ContainerDetail{
									Tag: "",
								},
							},
						},
					},
				},
			},
			want: "test/app-defaults",
		},
		{
			name: "Tag from hardDefaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "",
						},
						defaults: &GHCRRegistryDefaults{
							CommonRegistryDefaults: CommonRegistryDefaults{
								ContainerDetail: ContainerDetail{
									Tag: "test/app-hardDefaults",
								},
							},
						},
					},
				},
			},
			want: "test/app-hardDefaults",
		},
		{
			name: "Tag from root, ignore defaults/hardDefaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "test/app-root",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "test/app-defaults",
						},
						defaults: &GHCRRegistryDefaults{
							CommonRegistryDefaults: CommonRegistryDefaults{
								ContainerDetail: ContainerDetail{
									Tag: "test/app-hardDefaults",
								},
							},
						},
					},
				},
			},
			want: "test/app-root",
		},
		{
			name: "Tag from defaults, ignore hardDefaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "test/app-defaults",
						},
						defaults: &GHCRRegistryDefaults{
							CommonRegistryDefaults: CommonRegistryDefaults{
								ContainerDetail: ContainerDetail{
									Tag: "test/app-hardDefaults",
								},
							},
						},
					},
				},
			},
			want: "test/app-defaults",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetTag is called on it.
			got := tc.registry.GetTag()

			// THEN: the expected Tag is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nCommonRegistry.GetTag() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

// #########################
// # REGISTRY | VALIDATION #
// #########################

func TestCommonRegistry_CheckValues(t *testing.T) {
	// GIVEN: a CommonRegistry.
	tests := []struct {
		name      string
		input     *CommonRegistry
		wantImage string
		errRegex  string
	}{
		{
			name:     "image:tag at root",
			errRegex: `^$`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app",
					Tag:   "1.2.3",
				},
				defaults: nil,
			},
		},
		{
			name:     "image in defaults",
			errRegex: `^$`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "",
					Tag:   "1.2.3",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "test/app",
						},
					},
				},
			},
		},
		{
			name:     "image: missing",
			errRegex: `^image: <required>.*$`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "",
					Tag:   "1.2.3",
				},
				defaults: nil,
			},
		},
		{
			name:     "image: with period in name",
			errRegex: `^$`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/image.io",
					Tag:   "1.2.3",
				},
				defaults: nil,
			},
		},
		{
			name:     "image: invalid",
			errRegex: `image: .* <invalid>`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "	test/app",
					Tag:   "1.2.3",
				},
				defaults: nil,
			},
		},
		{
			name:     "tag: missing",
			errRegex: `^tag: <required>.*$`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app",
					Tag:   "",
				},
				defaults: nil,
			},
		},
		{
			name:     "tag: invalid templating",
			errRegex: `^tag: .* <invalid>.*$`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app",
					Tag:   "{{ version }",
				},
				defaults: nil,
			},
		},
		{
			name:     "tag: invalid url encoding",
			errRegex: `^tag: .* <invalid>.*$`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app",
					Tag:   "1.2	.3+",
				},
				defaults: nil,
			},
		},
		{
			name:     "tag in defaults",
			errRegex: `^$`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app",
					Tag:   "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "1.2.3",
						},
					},
				},
			},
		},
		{
			name:     "auth err",
			errRegex: `^token: <required>.*$`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app",
					Tag:   "1.2.3",
				},
				Auth: &HubAuth{
					HubAuthDefaults: HubAuthDefaults{
						Username: "someone",
					},
				},
			},
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

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestCommonRegistry_GetTagForVersion(t *testing.T) {
	// GIVEN: a CommonRegistry with a tag.
	tests := []struct {
		name     string
		registry CommonRegistry
		version  string
		want     string
	}{
		{
			name: "empty tag",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "",
						},
					},
				},
			},
			version: "3.2.1",
			want:    "",
		},
		{
			name: "plain version",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "1.2.3",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "1.2.3",
						},
					},
				},
			},
			version: "3.2.1",
			want:    "1.2.3",
		},
		{
			name: "version template",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "{{ version }}.1",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Tag: "{{ version }}.1",
						},
					},
				},
			},
			version: "3.2",
			want:    "3.2.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetTagForVersion is called on it.
			got := tc.registry.GetTagForVersion(tc.version)

			// THEN: the expected Tag is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nCommonRegistry.GetTagForVersion(%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.version,
					got, tc.want,
				)
			}
		})
	}
}

func TestCommonRegistry_ParseBody(t *testing.T) {
	// GIVEN: a HTTP response from a query.
	tests := []struct {
		name     string
		errRegex string
		resp     *http.Response
	}{
		{
			name:     "200",
			errRegex: "^$",
			resp: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		{
			name:     "404",
			errRegex: "tag not found",
			resp: &http.Response{
				StatusCode: http.StatusNotFound,
			},
		},
		{
			name:     "500",
			errRegex: "body error message here",
			resp: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("body error message here")),
			},
		},
	}

	// AND: Container detail.
	registry := CommonRegistry{
		ContainerDetail: ContainerDetail{
			Image: "test/app",
			Tag:   "ver",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: parseBody is called on this response.
			err := registry.parseBody(registry.Tag, tc.resp)

			// THEN: the expected error is returned.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nCommonRegistry parseBody(tag=%q, resp=%+v) mismatch\ngot:  %q\nwant: %q",
					packageName, registry.Tag, tc.resp,
					e, tc.errRegex,
				)
			}
		})
	}
}

func TestCommonRegistryDefaults_GetContainerDetail(t *testing.T) {
	// GIVEN: a ContainerDetail.
	tests := []struct {
		name   string
		detail ContainerDetail
	}{
		{
			name:   "empty",
			detail: ContainerDetail{},
		},
		{
			name: "image and tag",
			detail: ContainerDetail{
				Image: "foo",
				Tag:   "bar",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a CommonRegistryDefaults with this ContainerDetail.
			defaults := CommonRegistryDefaults{
				ContainerDetail: tc.detail,
			}
			want := &defaults.ContainerDetail

			// WHEN: GetContainerDetail is called on it.
			got := defaults.GetContainerDetail()

			// THEN: a pointer to this ContainerDetail is returned.
			if got != want {
				t.Errorf(
					"%s\nCommonRegistryDefaults.GetContainerDetail() pointer mismatch\ngot:  %p\nwant: %p",
					packageName, got, want,
				)
			}
		})
	}
}

func TestCommonRegistry_Detail(t *testing.T) {
	// GIVEN: a CommonRegistry.
	tests := []struct {
		name     string
		registry CommonRegistry
		want     ContainerDetail
	}{
		{
			name: "empty",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{},
			},
			want: ContainerDetail{},
		},
		{
			name: "image+tag from root",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "foo",
					Tag:   "bar",
				},
			},
			want: ContainerDetail{
				Image: "foo",
				Tag:   "bar",
			},
		},
		{
			name: "image from root, tag from defaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "foo",
					Tag:   "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "default-foo",
							Tag:   "default-bar",
						},
					},
				},
			},
			want: ContainerDetail{
				Image: "foo",
				Tag:   "default-bar",
			},
		},
		{
			name: "image from defaults, tag from root",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "",
					Tag:   "bar",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "default-foo",
							Tag:   "default-bar",
						},
					},
				},
			},
			want: ContainerDetail{
				Image: "default-foo",
				Tag:   "bar",
			},
		},
		{
			name: "image from defaults, tag from defaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "",
					Tag:   "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "default-foo",
							Tag:   "default-bar",
						},
					},
				},
			},
			want: ContainerDetail{
				Image: "default-foo",
				Tag:   "default-bar",
			},
		},
		{
			name: "image from defaults of defaults, tag from defaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "",
					Tag:   "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "default-foo",
							Tag:   "",
						},
						defaults: &GHCRRegistryDefaults{
							CommonRegistryDefaults: CommonRegistryDefaults{
								ContainerDetail: ContainerDetail{
									Image: "",
									Tag:   "default-default-bar",
								},
							},
						},
					},
				},
			},
			want: ContainerDetail{
				Image: "default-foo",
				Tag:   "default-default-bar",
			},
		},
		{
			name: "ignore defaults when found value",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "foo",
					Tag:   "",
				},
				defaults: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						ContainerDetail: ContainerDetail{
							Image: "default-foo",
							Tag:   "default-bar",
						},
						defaults: &GHCRRegistryDefaults{
							CommonRegistryDefaults: CommonRegistryDefaults{
								ContainerDetail: ContainerDetail{
									Image: "default-default-foo",
									Tag:   "default-default-bar",
								},
							},
						},
					},
				},
			},
			want: ContainerDetail{
				Image: "foo",
				Tag:   "default-bar",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: we call Detail.
			result := tc.registry.Detail()

			// THEN: The expected ContainerDetail is returned.
			want := decode.ToYAMLString(tc.want, "")
			if got := decode.ToYAMLString(result, ""); got != want {
				t.Errorf(
					"%s\nCommonRegistry.Detail() mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}
		})
	}
}

// ##########################
// # REGISTRY | INHERITANCE #
// ##########################

func TestCommonRegistry_Inherit(t *testing.T) {
	// GIVEN: A CommonRegistry and a Registry.
	tests := []struct {
		name     string
		registry CommonRegistry
		from     Registry
		inherit  bool
	}{
		{
			name: "nil registry to inherit from",
			registry: CommonRegistry{
				Auth: &HubAuth{
					HubAuthDefaults: HubAuthDefaults{
						Username:   "u",
						Token:      "t",
						queryToken: "qt",
						validUntil: time.Now().UTC().Add(10 * time.Minute),
					},
				},
				ContainerDetail: ContainerDetail{
					Image: "releaseargus/argus",
					Tag:   "{{ latest_version }}",
				},
			},
			from:    nil,
			inherit: false,
		},
		{
			name: "nil registry Auth to inherit from",
			registry: CommonRegistry{
				Auth: &HubAuth{
					HubAuthDefaults: HubAuthDefaults{
						Username:   "u",
						Token:      "t",
						queryToken: "qt",
						validUntil: time.Now().UTC().Add(10 * time.Minute),
					},
				},
				ContainerDetail: ContainerDetail{
					Image: "releaseargus/argus",
					Tag:   "{{ latest_version }}",
				},
			},
			from:    &GHCRRegistry{},
			inherit: false,
		},
		{
			name: "inherit from GHCR - same auth",
			registry: CommonRegistry{
				Auth: &GHCRAuth{
					GHCRAuthDefaults: GHCRAuthDefaults{
						Token:      "t",
						queryToken: "qt",
						validUntil: time.Now().UTC().Add(10 * time.Minute),
					},
				},
				ContainerDetail: ContainerDetail{
					Image: "releaseargus/argus",
					Tag:   "{{ latest_version }}",
				},
			},
			from: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token:      "t",
							queryToken: "qt-new",
							validUntil: time.Now().UTC().Add(10 * time.Minute),
						},
					},
					ContainerDetail: ContainerDetail{
						Image: "releaseargus/argus",
						Tag:   "{{ latest_version }}",
					},
				},
			},
			inherit: true,
		},
		{
			name: "inherit from GHCR - different auth",
			registry: CommonRegistry{
				Auth: &GHCRAuth{
					GHCRAuthDefaults: GHCRAuthDefaults{
						Token:      "t",
						queryToken: "qt",
						validUntil: time.Now().UTC().Add(10 * time.Minute),
					},
				},
				ContainerDetail: ContainerDetail{
					Image: "releaseargus/argus",
					Tag:   "{{ latest_version }}",
				},
			},
			from: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token:      "t1",
							queryToken: "qt-new",
							validUntil: time.Now().UTC().Add(10 * time.Minute),
						},
					},
					ContainerDetail: ContainerDetail{
						Image: "releaseargus/argus",
						Tag:   "{{ latest_version }}",
					},
				},
			},
			inherit: false,
		},
		{
			name: "inherit from Hub - same auth",
			registry: CommonRegistry{
				Auth: &HubAuth{
					HubAuthDefaults: HubAuthDefaults{
						Username:   "u",
						Token:      "t",
						queryToken: "qt",
						validUntil: time.Now().UTC().Add(10 * time.Minute),
					},
				},
				ContainerDetail: ContainerDetail{
					Image: "releaseargus/argus",
					Tag:   "{{ latest_version }}",
				},
			},
			from: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username:   "u",
							Token:      "t",
							queryToken: "qt-new",
							validUntil: time.Now().UTC().Add(10 * time.Minute),
						},
					},
					ContainerDetail: ContainerDetail{
						Image: "releaseargus/argus",
						Tag:   "{{ latest_version }}",
					},
				},
			},
			inherit: true,
		},
		{
			name: "inherit from Hub - different auth (different username, same token)",
			registry: CommonRegistry{
				Auth: &HubAuth{
					HubAuthDefaults: HubAuthDefaults{
						Username:   "u",
						Token:      "t",
						queryToken: "qt",
						validUntil: time.Now().UTC().Add(10 * time.Minute),
					},
				},
				ContainerDetail: ContainerDetail{
					Image: "releaseargus/argus",
					Tag:   "{{ latest_version }}",
				},
			},
			from: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username:   "u1",
							Token:      "t",
							queryToken: "qt-new",
							validUntil: time.Now().UTC().Add(10 * time.Minute),
						},
					},
					ContainerDetail: ContainerDetail{
						Image: "releaseargus/argus",
						Tag:   "{{ latest_version }}",
					},
				},
			},
			inherit: false,
		},
		{
			name: "inherit from Hub - different auth (same username, different token)",
			registry: CommonRegistry{
				Auth: &HubAuth{
					HubAuthDefaults: HubAuthDefaults{
						Username:   "u",
						Token:      "t",
						queryToken: "qt",
						validUntil: time.Now().UTC().Add(10 * time.Minute),
					},
				},
				ContainerDetail: ContainerDetail{
					Image: "releaseargus/argus",
					Tag:   "{{ latest_version }}",
				},
			},
			from: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username:   "u",
							Token:      "t1",
							queryToken: "qt-new",
							validUntil: time.Now().UTC().Add(10 * time.Minute),
						},
					},
					ContainerDetail: ContainerDetail{
						Image: "releaseargus/argus",
						Tag:   "{{ latest_version }}",
					},
				},
			},
			inherit: false,
		},
		{
			name: "inherit from Quay - same auth",
			registry: CommonRegistry{
				Auth: &QuayAuth{
					QuayAuthDefaults: QuayAuthDefaults{
						Token: "t",
					},
				},
				ContainerDetail: ContainerDetail{
					Image: "releaseargus/argus",
					Tag:   "{{ latest_version }}",
				},
			},
			from: &QuayRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
							Token: "t",
						},
					},
					ContainerDetail: ContainerDetail{
						Image: "releaseargus/argus",
						Tag:   "{{ latest_version }}",
					},
				},
			},
			inherit: false, // We don't use query tokens for Quay.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hadToken, hadQueryToken, hadValidUntil := getTokenData(t, tc.registry.Auth)

			// WHEN: Inherit is called on them.
			tc.registry.Inherit(tc.from)

			prefix := fmt.Sprintf(
				"%s\nCommonRegistry.Inherit(from=%T)",
				packageName, tc.from,
			)

			// THEN: the expected inheritance has occurred.
			if tc.from == nil {
				return
			}
			gotAuth := tc.registry.GetAuth()
			gotToken, gotQueryToken, gotValidUntil := getTokenData(t, gotAuth)
			wantToken, wantQueryToken, wantValidUntil := getTokenData(t, tc.from.GetAuth())
			tokenInherited := hadToken != gotToken
			queryTokenInherited := hadQueryToken != gotQueryToken
			validUntilInherited := !hadValidUntil.Equal(gotValidUntil)
			if (tokenInherited != tc.inherit && hadToken != gotToken) ||
				(queryTokenInherited != tc.inherit && hadQueryToken != gotQueryToken) ||
				(validUntilInherited != tc.inherit && !hadValidUntil.Equal(gotValidUntil)) {
				msg := "%s expected to inherit but did not\ntoken inherited: %t (had %q, got %q, want %q)\nqueryToken inherited: %t (had %q, got %q, want %q)\nvalidUntil inherited: %t (had %q, got %q, want %q)"
				if !tc.inherit {
					msg = "%s expected not to inherit but did\ntoken inherited: %t (had: %q, got: %q, want: NOT %q)\nqueryToken inherited: %t (had: %q, got: %q, want: NOT %q)\nvalidUntil inherited: %t (had: %q, got: %q, want: NOT %q)"
				}
				t.Errorf(
					msg,
					prefix,
					tokenInherited, hadToken, gotToken, wantToken,
					queryTokenInherited, hadQueryToken, gotQueryToken, wantQueryToken,
					validUntilInherited, hadValidUntil, gotValidUntil, wantValidUntil,
				)
			}
		})
	}
}

// ###################
// # REGISTRY | AUTH #
// ###################

func TestCommonRegistryDefaults_GetAuth(t *testing.T) {
	// GIVEN: a RegistryAuthDefaults.
	tests := []struct {
		name string
		auth RegistryAuthDefaults
	}{
		{
			name: "ghcr auth",
			auth: &GHCRAuthDefaults{
				Token: "abc",
			},
		},
		{
			name: "hub auth",
			auth: &HubAuthDefaults{
				Username: "123",
				Token:    "abc",
			},
		},
		{
			name: "ghcr auth",
			auth: &QuayAuthDefaults{
				Token: "abc",
			},
		},
		{
			name: "nil auth",
			auth: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a CommonRegistryDefaults with this Auth.
			registry := CommonRegistryDefaults{}
			registry.Auth = tc.auth

			// WHEN: GetAuth() is called on it.
			got := registry.GetAuth()

			// THEN: the expected Auth is returned.
			if got != tc.auth {
				t.Errorf(
					"%s\nCommonRegistry.GetAuth() result mismatch\ngot:  %v\nwant: %v",
					packageName, got, tc.auth,
				)
			}
		})
	}
}

func TestCommonRegistry_GetAuth(t *testing.T) {
	// GIVEN: a RegistryAuth.
	tests := []struct {
		name string
		auth RegistryAuth
	}{
		{
			name: "ghcr auth",
			auth: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "abc",
				},
			},
		},
		{
			name: "hub auth",
			auth: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username: "123",
					Token:    "abc",
				},
			},
		},
		{
			name: "ghcr auth",
			auth: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "abc",
				},
			},
		},
		{
			name: "nil auth",
			auth: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a CommonRegistry with this Auth.
			registry := CommonRegistry{}
			registry.Auth = tc.auth

			// WHEN: GetAuth() is called on it.
			got := registry.GetAuth()

			// THEN: the expected Auth is returned.
			if got != tc.auth {
				t.Errorf(
					"%s\nCommonRegistry.GetAuth() result mismatch\ngot:  %v\nwant: %v",
					packageName, got, tc.auth,
				)
			}
		})
	}
}
