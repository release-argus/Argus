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

func TestQuayRegistryDefaults_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     string
		registry *QuayRegistryDefaults
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "{}",
			registry: &QuayRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &QuayRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &QuayRegistryDefaults{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &QuayRegistryDefaults{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &QuayRegistryDefaults{
				Auth: &QuayAuthDefaults{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					json: .*unmarshal.*$`,
			),
		},
		{
			name:   "YAML/invalid Auth",
			format: "yaml",
			data:   `auth: []`,
			registry: &QuayRegistryDefaults{
				Auth: &QuayAuthDefaults{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					[^\s]+ sequence was used where mapping is expected`,
			),
		},
		{
			name:     "JSON/auth-null",
			format:   "json",
			data:     `{"auth": null}`,
			registry: &QuayRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
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
			registry: &QuayRegistryDefaults{
				Auth: &QuayAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				auth:
					token: tOKEn
			`),
		},
		{
			name:   "YAML/auth-quay",
			format: "yaml",
			data: test.TrimYAML(`
				auth:
					username: quay-username
					token: tOKEn
			`),
			registry: &QuayRegistryDefaults{
				Auth: &QuayAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				auth:
					token: tOKEn
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*QuayRegistryDefaults, error) {
					err := decode.Unmarshal(format, data, tc.registry)
					return tc.registry, err
				},
				tc.format, tc.data,
				func(v *QuayRegistryDefaults) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"QuayRegistryDefaults.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}

			if tc.registry.Auth == nil {
				t.Error("QuayRegistryDefaults.Auth mismatch\ngot:  nil\nwant: non-nil")
			}
		})
	}
}

func TestQuayRegistry_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     string
		registry *QuayRegistry
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "{}",
			registry: &QuayRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &QuayRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &QuayRegistry{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &QuayRegistry{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid ContainerDetail",
			format: "json",
			data:   `{"image": []}`,
			registry: &QuayRegistry{
				Auth: &QuayAuth{},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &QuayRegistry{
				Auth: &QuayAuth{},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &QuayRegistry{
				Auth: &QuayAuth{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					[^\s]+ .*unmarshal .*`,
			),
		},
		{
			name:   "YAML/invalid Auth",
			format: "yaml",
			data:   `auth: []`,
			registry: &QuayRegistry{
				Auth: &QuayAuth{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					[^\s]+ sequence was used where mapping is expected.*`,
			),
		},
		{
			name:     "JSON/auth-null",
			format:   "json",
			data:     `{"auth": null}`,
			registry: &QuayRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
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
			registry: &QuayRegistry{
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
			registry: &QuayRegistry{
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

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*QuayRegistry, error) {
					err := decode.Unmarshal(format, data, tc.registry)
					return tc.registry, err
				},
				tc.format, tc.data,
				func(v *QuayRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"QuayRegistry.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}

			if tc.registry.Auth == nil {
				t.Error("QuayRegistry.Auth mismatch\ngot:  nil\nwant: non-nil")
			}
		})
	}
}

func TestQuayRegistry_ApplyOverrides(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     string
		registry *QuayRegistry
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "{}",
			registry: &QuayRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &QuayRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &QuayRegistry{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &QuayRegistry{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid ContainerDetail",
			format: "json",
			data:   `{"image": []}`,
			registry: &QuayRegistry{
				Auth: &QuayAuth{},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &QuayRegistry{
				Auth: &QuayAuth{},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &QuayRegistry{
				Auth: &QuayAuth{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					json: .*unmarshal .*`,
			),
		},
		{
			name:   "YAML/invalid Auth",
			format: "yaml",
			data:   `auth: []`,
			registry: &QuayRegistry{
				Auth: &QuayAuth{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					[^\s]+ sequence was used where mapping is expected`,
			),
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
			registry: &QuayRegistry{
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
			registry: &QuayRegistry{
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
			name:   "auth-quay replaced",
			format: "yaml",
			data: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: quay-username
					token: tOKEn
			`),
			registry: &QuayRegistry{
				Auth: &QuayAuth{
					Token: "abc",
				},
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
			name:   "mutate",
			format: "yaml",
			data: test.TrimYAML(`
				tag: t
				auth:
					username: quay-username
			`),
			registry: &QuayRegistry{
				Image: "i",
				Auth: &QuayAuth{
					Token: "abc",
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					token: abc
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertApplyOverrides(
				t,
				tc.registry,
				func(format string, data []byte, v *QuayRegistry) (*QuayRegistry, error) {
					err := v.ApplyOverrides(format, data)
					return v, err
				},
				tc.format, tc.data,
				func(v *QuayRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				true,
				"QuayRegistry.ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// ####################
// # REGISTRY | STATE #
// ####################

func TestQuayRegistryDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		registry *QuayRegistryDefaults
		want     bool
	}{
		{
			name:     "nil",
			registry: nil,
			want:     true,
		},
		{
			name:     "empty",
			registry: &QuayRegistryDefaults{},
			want:     true,
		},
		{
			name: "non-empty/Token",
			registry: &QuayRegistryDefaults{
				Auth: &QuayAuthDefaults{
					Token: "foo",
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.registry.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestQuayRegistry_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data *QuayRegistry
		want bool
	}{
		{
			name: "nil",
			data: nil,
			want: true,
		},
		{
			name: "empty",
			data: RegistryMap["quay"]().(*QuayRegistry),
			want: true,
		},
		{
			name: "non-empty/Type",
			data: &QuayRegistry{
				Type: "abc",
				Auth: RegistryMap["quay"]().GetAuth(),
			},
			want: false,
		},
		{
			name: "non-empty/CommonRegistry",
			data: &QuayRegistry{
				Image: "i",
				Auth:  RegistryMap["quay"]().GetAuth(),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &QuayRegistry{
				Image: "i",
				Type:  "abc",
				Auth:  RegistryMap["quay"]().GetAuth(),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.IsZero()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestQuayRegistry_Copy(t *testing.T) {
	tests := []struct {
		name     string
		registry *QuayRegistry
		want     string
	}{
		{
			name:     "nil",
			registry: nil,
			want:     "null\n",
		},
		{
			name:     "empty",
			registry: RegistryMap["quay"]().(*QuayRegistry),
			want:     "{}\n",
		},
		{
			name: "filled",
			registry: &QuayRegistry{
				Image: "i1",
				Tag:   "t1",
				Auth: &QuayAuth{
					Token:    "t1",
					defaults: &QuayAuthDefaults{},
				},
				defaults: &QuayRegistryDefaults{},
			},
			want: test.TrimYAML(`
				image: i1
				tag: t1
				auth:
					token: t1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotInterface := tc.registry.Copy()

			if got := decode.ToYAMLString(gotInterface, ""); got != tc.want {
				t.Fatalf("QuayRegistry mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			got, ok := gotInterface.(*QuayRegistry)
			if !ok {
				if gotInterface == nil {
					return
				}
				t.Fatalf(
					"QuayRegistry.Copy() type mismatch\ngot:  %T\nwant: *QuayRegistry",
					gotInterface,
				)
			}

			gotAuth, ok := got.GetAuth().(*QuayAuth)
			hadAuth, _ := tc.registry.GetAuth().(*QuayAuth)
			if !ok {
				t.Fatalf(
					"QuayRegistry.GetAuth() type mismatch\ngot:  %T\nwant: *QuayAuth",
					got.GetAuth(),
				)
			}
			if got.defaults != tc.registry.defaults {
				t.Fatalf(
					"QuayRegistry.defaults mismatch\ngot:  %p\nwant: %p",
					got.defaults, tc.registry.defaults,
				)
			}

			if gotAuth == hadAuth {
				t.Fatalf(
					"QuayRegistry.GetAuth(): copy aliases the original\ngot:  %p\nwant: NOT %p",
					gotAuth, hadAuth,
				)
			}
		})
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

func TestQuayRegistryDefaults_String(t *testing.T) {
	tests := []struct {
		name string
		data *QuayRegistryDefaults
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "",
		},
		{
			name: "empty",
			data: &QuayRegistryDefaults{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &QuayRegistryDefaults{
				Auth: &QuayAuth{
					Token: "token1",
					defaults: &QuayAuthDefaults{
						Token: "token2",
					},
				},
			},
			want: test.TrimYAML(`
				auth:
					token: token1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				"QuayRegistryDefaults.String",
				tc.data.String,
				tc.want,
			)
		})
	}
}

func TestQuayRegistry_String(t *testing.T) {
	tests := []struct {
		name string
		data *QuayRegistry
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "",
		},
		{
			name: "empty",
			data: &QuayRegistry{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &QuayRegistry{
				Type: "test-quay",
				ContainerDetail: ContainerDetail{
					Image: "i1",
					Tag:   "t1",
					Defaults: &ContainerDetailDefaults{
						Tag: "t2",
					},
				},
				Auth: &QuayAuth{
					Token: "token1",
					defaults: &QuayAuthDefaults{
						Token: "token2",
					},
				},
			},
			want: test.TrimYAML(`
				type: test-quay
				image: i1
				tag: t1
				auth:
					token: token1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				"QuayRegistry.String",
				tc.data.String,
				tc.want,
			)
		})
	}
}

// #######################
// # REGISTRY | METADATA #
// #######################

func TestQuayRegistryDefaults_GetType(t *testing.T) {
	var registry QuayRegistryDefaults

	got := registry.GetType()

	if want := "quay"; got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestQuayRegistry_GetType(t *testing.T) {
	var registry QuayRegistry

	got := registry.GetType()

	if want := "quay"; got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestQuayRegistry_NewRequest(t *testing.T) {
	tests := []struct {
		name     string
		registry *QuayRegistry
		tag      string
		errRegex string
	}{
		{
			name: "no image - tag",
			registry: &QuayRegistry{
				Image: "",
				Tag:   "",
			},
			errRegex: `^$`,
		},
		{
			name: "have image+tag",
			registry: &QuayRegistry{
				Image: "123",
				Tag:   "not-used",
			},
			tag:      "foo",
			errRegex: `^$`,
		},
		{
			name: "tag: invalid",
			registry: &QuayRegistry{
				Image: "123",
				Tag:   "not-used",
			},
			tag: "	foo",
			errRegex: test.TrimYAML(`
				^parse "https://.*
					.*invalid control character in URL$`,
			),
		},
		{
			name: "image: invalid",
			registry: &QuayRegistry{
				Image: "	123",
				Tag:   "not-used",
			},
			tag: "foo",
			errRegex: test.TrimYAML(`
				^parse "https://.*
					.*invalid control character in URL$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.registry.newRequest(tc.tag)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

// ################
// # AUTH | STATE #
// ################

func TestQuayAuthDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data *QuayAuthDefaults
		want bool
	}{
		{
			name: "nil",
			data: nil,
			want: true,
		},
		{
			name: "empty",
			data: &QuayAuthDefaults{},
			want: true,
		},
		{
			name: "non-empty",
			data: &QuayAuthDefaults{Token: "t1"},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.IsZero()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestQuayAuth_Copy(t *testing.T) {
	tests := []struct {
		name string
		auth *QuayAuth
		want string
	}{
		{
			name: "nil",
			auth: nil,
			want: "null\n",
		},
		{
			name: "empty",
			auth: &QuayAuth{},
			want: "{}\n",
		},
		{
			name: "filled",
			auth: &QuayAuth{
				Token:    "t1",
				defaults: &QuayAuthDefaults{},
			},
			want: "token: t1\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authCopy := tc.auth.Copy()

			if got := decode.ToYAMLString(authCopy, ""); got != tc.want {
				t.Fatalf("QuayAuth mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			got, ok := authCopy.(*QuayAuth)
			if !ok {
				if authCopy == nil {
					return
				}
				t.Fatalf("QuayAuth.Copy() type mismatch\ngot:  %T\nwant: *QuayAuth", authCopy)
			}
			if tc.auth == nil {
				return
			}

			if got.defaults != tc.auth.defaults {
				t.Fatalf(
					"QuayAuth.defaults mismatch\ngot:  %p\nwant: %p",
					got.defaults, tc.auth.defaults,
				)
			}

			if got == tc.auth {
				t.Fatalf("copy aliases the original\ngot:  %p\nwant: NOT %p", got, tc.auth)
			}
		})
	}
}

// ####################
// # AUTH | STRINGIFY #
// ####################

func TestQuayAuthDefaults_String(t *testing.T) {
	tests := []struct {
		name string
		data *QuayAuthDefaults
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "null\n",
		},
		{
			name: "empty",
			data: &QuayAuthDefaults{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &QuayAuthDefaults{
				Token: "t1",
			},
			want: "token: t1\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				"QuayAuthDefaults.String",
				tc.data.String,
				tc.want,
			)
		})
	}
}

// ###################
// # AUTH | DEFAULTS #
// ###################

func TestQuayAuthDefaults_Defaults(t *testing.T) {
	tests := []struct {
		name         string
		data         *QuayAuthDefaults
		haveDefaults bool
	}{
		{
			name:         "no defaults",
			data:         &QuayAuthDefaults{},
			haveDefaults: false,
		},
		{
			name: "defaults",
			data: &QuayAuthDefaults{
				defaults: &QuayAuthDefaults{
					Token: "defaults",
				},
			},
			haveDefaults: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.Defaults()

			if gotDefaults := got != nil; gotDefaults != tc.haveDefaults {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", gotDefaults, tc.haveDefaults)
			}
		})
	}
}

func TestQuayAuthDefaults_SetDefaults(t *testing.T) {
	tests := []struct {
		name        string
		newDefaults RegistryAuthDefaults
		doesSet     bool
	}{
		{
			name:        "give QuayAuthDefaults",
			newDefaults: &QuayAuthDefaults{},
			doesSet:     true,
		},
		{
			name:        "doesn't give ECRAuthDefaults",
			newDefaults: &ECRAuthDefaults{},
			doesSet:     false,
		},
		{
			name:        "doesn't give GHCRAuthDefaults",
			newDefaults: &GHCRAuthDefaults{},
			doesSet:     false,
		},
		{
			name:        "doesn't give HubAuthDefaults",
			newDefaults: &HubAuthDefaults{},
			doesSet:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			data := &QuayAuthDefaults{}

			data.SetDefaults(tc.newDefaults)

			var want *QuayAuthDefaults
			if tc.doesSet {
				want = tc.newDefaults.(*QuayAuthDefaults)
			}
			if data.defaults != want {
				t.Errorf(
					"QuayAuthDefaults.defaults mismatch\ngot:  %p\nwant: %p",
					data.defaults, want,
				)
			}
		})
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

func TestQuayAuth_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *QuayAuth
		errRegex string
	}{
		{
			name:     "nil",
			input:    (*QuayAuth)(nil),
			errRegex: `^$`,
		},
		{
			name: "token",
			input: &QuayAuth{
				Token: "token",
			},
			errRegex: `^$`,
		},
		{
			name:     "no token",
			input:    &QuayAuth{},
			errRegex: `^$`,
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

// ######################
// # AUTH | CREDENTIALS #
// ######################

func TestQuayAuth_GetToken(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		data     *QuayAuth
		want     string
		wantSelf string
	}{
		{
			name:     "empty",
			data:     &QuayAuth{},
			want:     "",
			wantSelf: "",
		},
		{
			name: "no defaults",
			data: &QuayAuth{
				Token: "root",
			},
			want:     "root",
			wantSelf: "root",
		},
		{
			name: "defaults fallback",
			data: &QuayAuth{
				Token: "",
				defaults: &QuayAuthDefaults{
					Token: "defaults",
				},
			},
			want:     "defaults",
			wantSelf: "",
		},
		{
			name: "defaults fallback recursive",
			data: &QuayAuth{
				Token: "",
				defaults: &QuayAuthDefaults{
					Token: "",
					defaults: &QuayAuthDefaults{
						Token: "hard-defaults",
					},
				},
			},
			want:     "hard-defaults",
			wantSelf: "",
		},
		{
			name: "root Token prioritised",
			data: &QuayAuth{
				Token: "root",
				defaults: &QuayAuthDefaults{
					Token: "defaults",
				},
			},
			want:     "root",
			wantSelf: "root",
		},
		{
			name: "env var expanded",
			env:  map[string]string{"TEST_QUAY_AUTH_GET_TOKEN": "from-env"},
			data: &QuayAuth{
				Token: "${TEST_QUAY_AUTH_GET_TOKEN}",
			},
			want:     "from-env",
			wantSelf: "from-env",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.env != nil {
				test.IsolateEnv(t, tc.env)
			}

			got := tc.data.GetToken()

			if got != tc.want {
				t.Fatalf("QuayAuth.GetToken() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			got = tc.data.GetTokenSelf()

			if got != tc.wantSelf {
				t.Fatalf("QuayAuth.GetTokenSelf() mismatch\ngot:  %q\nwant: %q", got, tc.wantSelf)
			}
		})
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestQuayAuthDefaults_GetQueryTokenSelf(t *testing.T) {
	tests := []struct {
		name string
		data *QuayAuthDefaults
		want string
	}{
		{
			name: "empty",
			data: &QuayAuthDefaults{},
			want: "",
		},
		{
			name: "non-empty",
			data: &QuayAuthDefaults{
				Token: "token",
			},
			want: "token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			queryToken, _ := tc.data.GetQueryTokenSelf()

			if queryToken != tc.want {
				t.Errorf("queryToken mismatch\ngot:  %q\nwant: %q", queryToken, tc.want)
			}
		})
	}
}

func TestQuayAuth_SetQueryToken(t *testing.T) {
	queryToken := "new-query-token"
	validUntil := time.Now().Add(10 * time.Second)

	auth := &QuayAuth{
		Token: "abc",
	}
	defaults := &QuayAuthDefaults{Token: "def"}

	auth.SetQueryToken(queryToken, validUntil)
	defaults.SetQueryToken(queryToken, validUntil)

	var viaInterface RegistryAuth = auth
	viaInterface.SetQueryToken(queryToken, validUntil)

	// Token-only auth makes SetQueryToken a no-op.
	if got, want := auth.Token, "abc"; got != want {
		t.Errorf("QuayAuth.Token mismatch\ngot:  %q\nwant: %q", got, want)
	}
	if got, want := defaults.Token, "def"; got != want {
		t.Errorf("QuayAuthDefaults.Token mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestQuayAuth_GetQueryToken(t *testing.T) {
	tests := []struct {
		name string
		auth QuayAuth
		want string
	}{
		{
			name: "empty token",
			auth: QuayAuth{},
		},
		{
			name: "existing token",
			auth: QuayAuth{
				Token: "abc",
			},
			want: "abc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.auth.GetQueryToken(ContainerDetail{})

			if got != tc.want {
				t.Errorf("queryToken mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			if err != nil {
				t.Errorf("error mismatch\ngot:  %v\nwant: nil", err)
			}
		})
	}
}

// ######################
// # AUTH | INHERITANCE #
// ######################

func TestQuayAuth_Inherit(t *testing.T) {
	tests := []struct {
		name      string
		auth      *QuayAuth
		from      RegistryAuth
		wantToken string
	}{
		{
			name:      "inherit from nil",
			auth:      &QuayAuth{},
			from:      nil,
			wantToken: "",
		},
		{
			name: "inherit from QuayAuth (Token is SecretValue)",
			auth: &QuayAuth{
				Token: util.SecretValue,
			},
			from: &QuayAuth{
				Token: "abc",
			},
			wantToken: "abc",
		},
		{
			name: "do not inherit from QuayAuth when Token is not SecretValue",
			auth: &QuayAuth{
				Token: "foo",
			},
			from: &QuayAuth{
				Token: "abc",
			},
			wantToken: "foo",
		},
		{
			name: "do not inherit from GHCRAuth",
			auth: &QuayAuth{
				Token: util.SecretValue,
			},
			from: &GHCRAuth{
				Token: "abc",
			},
			wantToken: util.SecretValue,
		},
		{
			name: "do not inherit from HubAuth",
			auth: &QuayAuth{
				Token: util.SecretValue,
			},
			from: &HubAuth{
				Token: "abc",
			},
			wantToken: util.SecretValue,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.auth.Inherit(tc.from, ContainerDetail{}, ContainerDetail{})

			if got := tc.auth.Token; got != tc.wantToken {
				t.Errorf("QuayAuth.Token mismatch\ngot:  %q\nwant: %q", got, tc.wantToken)
			}
		})
	}
}
