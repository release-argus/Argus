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
	"io"
	"net/http"
	"net/http/httptest"
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

func TestHubRegistryDefaults_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     string
		registry *HubRegistryDefaults
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			registry: &HubRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &HubRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &HubRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &HubRegistryDefaults{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &HubRegistryDefaults{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &HubRegistryDefaults{
				Auth: &HubAuthDefaults{},
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
			registry: &HubRegistryDefaults{
				Auth: &HubAuthDefaults{},
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
			registry: &HubRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
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
			registry: &HubRegistryDefaults{
				Auth: &HubAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				auth:
					username: hub-username
					token: tOKEn
			`),
		},
		{
			name:   "YAML/auth-hub",
			format: "yaml",
			data: test.TrimYAML(`
				auth:
					username: hub-username
					token: tOKEn
			`),
			registry: &HubRegistryDefaults{
				Auth: &HubAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				auth:
					username: hub-username
					token: tOKEn
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*HubRegistryDefaults, error) {
					err := decode.Unmarshal(format, data, tc.registry)
					return tc.registry, err
				},
				tc.format, tc.data,
				func(v *HubRegistryDefaults) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"HubRegistryDefaults.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}

			if tc.registry.Auth == nil {
				t.Error("HubRegistryDefaults.Auth mismatch\ngot:  nil\nwant: non-nil")
			}
		})
	}
}

func TestHubRegistry_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     string
		registry *HubRegistry
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			registry: &HubRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &HubRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &HubRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &HubRegistry{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &HubRegistry{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid ContainerDetail",
			format: "json",
			data:   `{"image": []}`,
			registry: &HubRegistry{
				Auth: &HubAuth{},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &HubRegistry{
				Auth: &HubAuth{},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &HubRegistry{
				Auth: &HubAuth{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					json: .*unmarshal .*$`,
			),
		},
		{
			name:   "YAML/invalid Auth",
			format: "yaml",
			data:   `auth: []`,
			registry: &HubRegistry{
				Auth: &HubAuth{},
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
			registry: &HubRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
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
			registry: &HubRegistry{
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
			registry: &HubRegistry{
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*HubRegistry, error) {
					err := decode.Unmarshal(format, data, tc.registry)
					return tc.registry, err
				},
				tc.format, tc.data,
				func(v *HubRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"HubRegistry.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}

			if tc.registry.Auth == nil {
				t.Error("HubRegistry.Auth mismatch\ngot:  nil\nwant: non-nil")
			}
		})
	}
}

func TestHubRegistry_ApplyOverrides(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     string
		registry *HubRegistry
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			registry: &HubRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &HubRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &HubRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &HubRegistry{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &HubRegistry{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid ContainerDetail",
			format: "json",
			data:   `{"image": []}`,
			registry: &HubRegistry{
				Auth: &HubAuth{},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &HubRegistry{
				Auth: &HubAuth{},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &HubRegistry{
				Auth: &HubAuth{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					json: .*unmarshal .*$`,
			),
		},
		{
			name:   "YAML/invalid Auth",
			format: "yaml",
			data:   `auth: []`,
			registry: &HubRegistry{
				Auth: &HubAuth{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					[^\s]+ sequence was used where mapping is expected`,
			),
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
			registry: &HubRegistry{
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
			registry: &HubRegistry{
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
			name:   "auth-hub replaced",
			format: "yaml",
			data: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: hub-username
					token: tOKEn
			`),
			registry: &HubRegistry{
				Auth: &HubAuth{
					Token: "abc",
				},
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
			name:   "mutate",
			format: "yaml",
			data: test.TrimYAML(`
				tag: t
				auth:
					username: hub-username
			`),
			registry: &HubRegistry{
				Image: "i",
				Auth: &HubAuth{
					Token: "abc",
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: hub-username
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
				func(format string, data []byte, v *HubRegistry) (*HubRegistry, error) {
					err := v.ApplyOverrides(format, data)
					return v, err
				},
				tc.format, tc.data,
				func(v *HubRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				true,
				"HubRegistry.ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// ####################
// # REGISTRY | STATE #
// ####################

func TestHubRegistryDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		registry *HubRegistryDefaults
		want     bool
	}{
		{
			name:     "nil",
			registry: nil,
			want:     true,
		},
		{
			name:     "empty",
			registry: &HubRegistryDefaults{},
			want:     true,
		},
		{
			name: "non-empty/Username",
			registry: &HubRegistryDefaults{
				Auth: &HubAuthDefaults{
					Username: "u",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Token",
			registry: &HubRegistryDefaults{
				Auth: &HubAuthDefaults{
					Token: "foo",
				},
			},
			want: false,
		},
		{
			name: "non-empty/queryToken and validUntil",
			registry: &HubRegistryDefaults{
				Auth: &HubAuthDefaults{
					queryToken: "bar",
					validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want: true, // IsZero ignores the cached query token.
		},
		{
			name: "non-empty/all",
			registry: &HubRegistryDefaults{
				Auth: &HubAuthDefaults{
					Username:   "u",
					Token:      "foo",
					queryToken: "bar",
					validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
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

func TestHubRegistry_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data *HubRegistry
		want bool
	}{
		{
			name: "nil",
			data: nil,
			want: true,
		},
		{
			name: "empty",
			data: RegistryMap["hub"]().(*HubRegistry),
			want: true,
		},
		{
			name: "non-empty/Type",
			data: &HubRegistry{
				Type: "abc",
				Auth: RegistryMap["hub"]().GetAuth(),
			},
			want: false,
		},
		{
			name: "non-empty/CommonRegistry",
			data: &HubRegistry{
				Image: "i",
				Auth:  RegistryMap["hub"]().GetAuth(),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &HubRegistry{
				Image: "i",
				Type:  "abc",
				Auth:  RegistryMap["hub"]().GetAuth(),
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

func TestHubRegistry_Copy(t *testing.T) {
	tests := []struct {
		name     string
		registry *HubRegistry
		want     string
	}{
		{
			name:     "nil",
			registry: nil,
			want:     "null\n",
		},
		{
			name:     "empty",
			registry: RegistryMap["hub"]().(*HubRegistry),
			want:     "{}\n",
		},
		{
			name: "filled",
			registry: &HubRegistry{
				Image: "i1",
				Tag:   "t1",
				Auth: &HubAuth{
					Username:   "u1",
					Token:      "t1",
					queryToken: "qT",
					validUntil: time.Now().Add(time.Hour),
					defaults:   &HubAuthDefaults{},
				},
				defaults: &HubRegistryDefaults{},
			},
			want: test.TrimYAML(`
				image: i1
				tag: t1
				auth:
					username: u1
					token: t1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotInterface := tc.registry.Copy()

			if g := decode.ToYAMLString(gotInterface, ""); g != tc.want {
				t.Fatalf("HubRegistry mismatch\ngot:  %q\nwant: %q", g, tc.want)
			}

			got, ok := gotInterface.(*HubRegistry)
			if !ok {
				if gotInterface == nil {
					return
				}
				t.Fatalf(
					"HubRegistry.Copy() type mismatch\ngot:  %T\nwant: *HubRegistry",
					gotInterface,
				)
			}

			gotAuth, ok := got.GetAuth().(*HubAuth)
			hadAuth, _ := tc.registry.GetAuth().(*HubAuth)
			if !ok {
				t.Fatalf(
					"HubRegistry.GetAuth() type mismatch\ngot:  %T\nwant: *HubAuth",
					got.GetAuth(),
				)
			}
			if gotAuth.queryToken != hadAuth.queryToken ||
				gotAuth.validUntil != hadAuth.validUntil ||
				got.defaults != tc.registry.defaults {
				t.Fatalf(
					"HubRegistry mismatch\n"+
						"got:  Auth.queryToken=%q, Auth.validUntil=%s, defaults=%p\n"+
						"want: Auth.queryToken=%q, Auth.validUntil=%s, defaults=%p",
					gotAuth.queryToken, gotAuth.validUntil, got.defaults,
					hadAuth.queryToken, hadAuth.validUntil, tc.registry.defaults,
				)
			}

			if gotAuth == hadAuth {
				t.Fatalf(
					"HubRegistry.GetAuth(): copy aliases the original\ngot:  %p\nwant: NOT %p",
					gotAuth, hadAuth,
				)
			}
		})
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

func TestHubRegistryDefaults_String(t *testing.T) {
	tests := []struct {
		name string
		data *HubRegistryDefaults
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "",
		},
		{
			name: "empty",
			data: &HubRegistryDefaults{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &HubRegistryDefaults{
				Auth: &HubAuth{
					Username: "u1",
					Token:    "token1",
					defaults: &HubAuthDefaults{
						Username: "u2",
						Token:    "token2",
					},
				},
			},
			want: test.TrimYAML(`
				auth:
					username: u1
					token: token1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				"HubRegistryDefaults.String",
				tc.data.String,
				tc.want,
			)
		})
	}
}

func TestHubRegistry_String(t *testing.T) {
	tests := []struct {
		name string
		data *HubRegistry
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "",
		},
		{
			name: "empty",
			data: &HubRegistry{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &HubRegistry{
				Type: "test-hub",
				ContainerDetail: ContainerDetail{
					Image: "i1",
					Tag:   "t1",
					Defaults: &ContainerDetailDefaults{
						Tag: "t2",
					},
				},
				Auth: &HubAuth{
					Username: "u1",
					Token:    "token1",
					defaults: &HubAuthDefaults{
						Username: "u2",
						Token:    "token2",
					},
				},
			},
			want: test.TrimYAML(`
				type: test-hub
				image: i1
				tag: t1
				auth:
					username: u1
					token: token1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				"HubRegistry.String",
				tc.data.String,
				tc.want,
			)
		})
	}
}

// #######################
// # REGISTRY | METADATA #
// #######################

func TestHubRegistryDefaults_GetType(t *testing.T) {
	var registry HubRegistryDefaults

	got := registry.GetType()

	if want := "hub"; got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestHubRegistry_GetType(t *testing.T) {
	tests := []struct {
		name     string
		registry *HubRegistry
	}{
		{
			name:     "no type",
			registry: &HubRegistry{},
		},
		{
			name: "ignore type",
			registry: &HubRegistry{
				Type: "hi",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.registry.GetType()

			if want := "hub"; got != want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

// #########################
// # REGISTRY | VALIDATION #
// #########################

func TestHubRegistry_CheckValues(t *testing.T) {
	tests := []struct {
		name      string
		input     *HubRegistry
		wantImage string
		errRegex  string
	}{
		{
			name: "valid/official image gets library namespace",
			input: &HubRegistry{
				Image: "i",
				Tag:   "t",
				Auth: &HubAuth{
					Username: "username",
					Token:    "token",
				},
			},
			wantImage: "library/i",
			errRegex:  `^$`,
		},
		{
			name: "valid/namespaced image unchanged",
			input: &HubRegistry{
				Image: "org/i",
				Tag:   "t",
				Auth: &HubAuth{
					Username: "username",
					Token:    "token",
				},
			},
			wantImage: "org/i",
			errRegex:  `^$`,
		},
		{
			name: "CommonRegistry: no image",
			input: &HubRegistry{
				Tag: "t",
				Auth: &HubAuth{
					Username: "username",
					Token:    "t",
				},
			},
			errRegex: `^image: <required> \([^\)]+\)$`,
		},
		{
			name: "Auth: no token",
			input: &HubRegistry{
				Image: "i",
				Tag:   "t",
				Auth: &HubAuth{
					Username: "username",
				},
			},
			wantImage: "i", // library/ is added only once validation passes.
			errRegex:  `^token: <required> \([^\)]+\)$`,
		},
		{
			name: "Auth: no username",
			input: &HubRegistry{
				Image: "i",
				Tag:   "t",
				Auth: &HubAuth{
					Token: "token",
				},
			},
			wantImage: "i", // library/ is added only once validation passes.
			errRegex:  `^username: <required> \([^\)]+\)$`,
		},
		{
			name: "CommonRegistry err and Auth err",
			input: &HubRegistry{
				Tag: "t",
				Auth: &HubAuth{
					Username: "username",
				},
			},
			errRegex: test.TrimYAML(`
				^image: <required> \([^\)]+\)
				token: <required> \([^\)]+\)$`,
			),
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

			if tc.input.Image != tc.wantImage {
				t.Errorf(
					"HubRegistry.Image mismatch\ngot:  %q\nwant: %q",
					tc.input.Image, tc.wantImage,
				)
			}
		})
	}
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestHubRegistry_NewRequest(t *testing.T) {
	tests := []struct {
		name     string
		registry *HubRegistry
		tag      string
		errRegex string
	}{
		{
			name: "no image or tag",
			registry: &HubRegistry{
				Image: "",
				Tag:   "",
			},
			errRegex: `^$`,
		},
		{
			name: "have image+tag",
			registry: &HubRegistry{
				Image: "123",
				Tag:   "not-used",
			},
			tag:      "foo",
			errRegex: `^$`,
		},
		{
			name: "tag: invalid",
			registry: &HubRegistry{
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
			registry: &HubRegistry{
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

// ###################
// # AUTH | DECODING #
// ###################

func TestHubAuthDefaults_Unmarshal(t *testing.T) {
	tests := []struct {
		name         string
		format, data string
		want         *HubAuthDefaults
		errRegex     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "{}",
			errRegex: `^$`,
			want:     &HubAuthDefaults{},
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}\n",
			errRegex: `^$`,
			want:     &HubAuthDefaults{},
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"username": "u1",
				"token": "t1"
			}`),
			errRegex: `^$`,
			want: &HubAuthDefaults{
				Username: "u1",
				Token:    "t1",
			},
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				username: u1
				token: t1
			`),
			errRegex: `^$`,
			want: &HubAuthDefaults{
				Username: "u1",
				Token:    "t1",
			},
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "invalid",
			errRegex: `invalid character`,
		},
		{
			name:   "YAML/invalid",
			format: "yaml",
			data:   "invalid",
			errRegex: test.TrimYAML(`
				^[^\s]+ string was used where mapping is expected
				[^\s]+ .+
				\s*\^$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*HubAuthDefaults, error) {
					var decoded HubAuthDefaults
					err := decode.Unmarshal(format, data, &decoded)
					return &decoded, err
				},
				tc.format, tc.data,
				func(v *HubAuthDefaults) string { return v.String("") },
				decode.ToYAMLString(tc.want, ""),
				tc.errRegex,
				"HubAuthDefaults.Unmarshal",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

// ################
// # AUTH | STATE #
// ################

func TestHubAuthDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data *HubAuthDefaults
		want bool
	}{
		{
			name: "nil",
			data: nil,
			want: true,
		},
		{
			name: "empty",
			data: &HubAuthDefaults{},
			want: true,
		},
		{
			name: "non-empty/Username",
			data: &HubAuthDefaults{
				Username: "u1",
			},
			want: false,
		},
		{
			name: "non-empty/Token",
			data: &HubAuthDefaults{
				Token: "t1",
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &HubAuthDefaults{
				Username: "u1",
				Token:    "t1",
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

func TestHubAuth_Copy(t *testing.T) {
	tests := []struct {
		name string
		auth *HubAuth
		want string
	}{
		{
			name: "nil",
			auth: nil,
			want: "null\n",
		},
		{
			name: "empty",
			auth: &HubAuth{},
			want: "{}\n",
		},
		{
			name: "filled",
			auth: &HubAuth{
				Username:   "u1",
				Token:      "t1",
				queryToken: "qT",
				validUntil: time.Now(),
				defaults:   &HubAuthDefaults{},
			},
			want: test.TrimYAML(`
				username: u1
				token: t1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authCopy := tc.auth.Copy()

			if got := decode.ToYAMLString(authCopy, ""); got != tc.want {
				t.Fatalf("HubAuth mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			got, ok := authCopy.(*HubAuth)
			if !ok {
				if authCopy == nil {
					return
				}
				t.Fatalf("HubAuth.Copy() type mismatch\ngot:  %T\nwant: *HubAuth", authCopy)
			}
			if tc.auth == nil {
				return
			}

			if got.queryToken != tc.auth.queryToken ||
				got.validUntil != tc.auth.validUntil ||
				got.defaults != tc.auth.defaults {
				t.Fatalf(
					"HubAuth mismatch\n"+
						"got:  queryToken=%q, validUntil=%s, defaults=%p\n"+
						"want: queryToken=%q, validUntil=%s, defaults=%p",
					got.queryToken, got.validUntil, got.defaults,
					tc.auth.queryToken, tc.auth.validUntil, tc.auth.defaults,
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

func TestHubAuthDefaults_String(t *testing.T) {
	tests := []struct {
		name string
		data *HubAuthDefaults
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "null\n",
		},
		{
			name: "empty",
			data: &HubAuthDefaults{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &HubAuthDefaults{
				Username: "u1",
				Token:    "t1",
			},
			want: test.TrimYAML(`
				username: u1
				token: t1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				"HubAuthDefaults.String",
				tc.data.String,
				tc.want,
			)
		})
	}
}

// ###################
// # AUTH | DEFAULTS #
// ###################

func TestHubAuthDefaults_Defaults(t *testing.T) {
	tests := []struct {
		name         string
		data         *HubAuthDefaults
		haveDefaults bool
	}{
		{
			name:         "no defaults",
			data:         &HubAuthDefaults{},
			haveDefaults: false,
		},
		{
			name: "defaults",
			data: &HubAuthDefaults{
				defaults: &HubAuthDefaults{
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

func TestHubAuthDefaults_SetDefaults(t *testing.T) {
	tests := []struct {
		name        string
		newDefaults RegistryAuthDefaults
		doesSet     bool
	}{
		{
			name:        "give HubAuthDefaults",
			newDefaults: &HubAuthDefaults{},
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
			name:        "doesn't give QuayAuthDefaults",
			newDefaults: &QuayAuthDefaults{},
			doesSet:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			data := &HubAuthDefaults{}

			data.SetDefaults(tc.newDefaults)

			var want *HubAuthDefaults
			if tc.doesSet {
				want = tc.newDefaults.(*HubAuthDefaults)
			}
			if data.defaults != want {
				t.Errorf(
					"HubAuthDefaults.defaults mismatch\ngot:  %p\nwant: %p",
					data.defaults, want,
				)
			}
		})
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

func TestHubAuth_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *HubAuth
		errRegex string
	}{
		{
			name:     "nil",
			input:    (*HubAuth)(nil),
			errRegex: `^$`,
		},
		{
			name: "valid",
			input: &HubAuth{
				Username: "username",
				Token:    "token",
			},
			errRegex: `^$`,
		},
		{
			name: "username, no token",
			input: &HubAuth{
				Username: "username",
			},
			errRegex: `^token: <required> \([^\)]+\)$`,
		},
		{
			name: "token, no username",
			input: &HubAuth{
				Token: "token",
			},
			errRegex: `^username: <required> \([^\)]+\)$`,
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

func TestHubAuth_GetUsernameSelf(t *testing.T) {
	tests := []struct {
		name string
		auth *HubAuth
		env  map[string]string
		want string
	}{
		{
			name: "empty auth",
			auth: &HubAuth{},
			want: "",
		},
		{
			name: "username",
			auth: &HubAuth{
				Username: "u",
			},
			want: "u",
		},
		{
			name: "env vars",
			auth: &HubAuth{
				Username: "u-${DOCKER_USERNAME}",
			},
			env: map[string]string{
				"DOCKER_USERNAME": "1",
			},
			want: "u-1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.IsolateEnv(t, tc.env)

			got := tc.auth.GetUsernameSelf()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestHubAuth_GetUsername(t *testing.T) {
	tests := []struct {
		name string
		data *HubAuth
		want string
	}{
		{
			name: "empty",
			data: &HubAuth{},
			want: "",
		},
		{
			name: "no defaults",
			data: &HubAuth{
				Username: "root",
			},
			want: "root",
		},
		{
			name: "defaults fallback",
			data: &HubAuth{
				Username: "",
				defaults: &HubAuthDefaults{
					Username: "defaults",
				},
			},
			want: "defaults",
		},
		{
			name: "defaults fallback recursive",
			data: &HubAuth{
				Username: "",
				defaults: &HubAuthDefaults{
					Username: "hard-defaults",
				},
			},
			want: "hard-defaults",
		},
		{
			name: "root Username prioritised",
			data: &HubAuth{
				Username: "root",
				defaults: &HubAuthDefaults{
					Username: "defaults",
				},
			},
			want: "root",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.GetUsername()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestHubAuth_GetToken(t *testing.T) {
	tests := []struct {
		name     string
		data     *HubAuth
		want     string
		wantSelf string
	}{
		{
			name:     "empty",
			data:     &HubAuth{},
			want:     "",
			wantSelf: "",
		},
		{
			name: "no defaults",
			data: &HubAuth{
				Token: "root",
			},
			want:     "root",
			wantSelf: "root",
		},
		{
			name: "defaults fallback",
			data: &HubAuth{
				Token: "",
				defaults: &HubAuthDefaults{
					Token: "defaults",
				},
			},
			want:     "defaults",
			wantSelf: "",
		},
		{
			name: "defaults fallback recursive",
			data: &HubAuth{
				Token: "",
				defaults: &HubAuthDefaults{
					Token: "",
					defaults: &HubAuthDefaults{
						Token: "hard-defaults",
					},
				},
			},
			want:     "hard-defaults",
			wantSelf: "",
		},
		{
			name: "root Token prioritised",
			data: &HubAuth{
				Token: "root",
				defaults: &HubAuthDefaults{
					Token: "defaults",
				},
			},
			want:     "root",
			wantSelf: "root",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.GetToken()

			if got != tc.want {
				t.Fatalf("HubAuth.GetToken() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			got = tc.data.GetTokenSelf()

			if got != tc.wantSelf {
				t.Fatalf("HubAuth.GetTokenSelf() mismatch\ngot:  %q\nwant: %q", got, tc.wantSelf)
			}
		})
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestHubAuthDefaults_GetQueryTokenSelf(t *testing.T) {
	tests := []struct {
		name string
		data *HubAuthDefaults
		want string
	}{
		{
			name: "empty",
			data: &HubAuthDefaults{},
			want: "",
		},
		{
			name: "expired",
			data: &HubAuthDefaults{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(-10 * time.Second),
			},
			want: "",
		},
		{
			name: "valid",
			data: &HubAuthDefaults{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(10 * time.Second),
			},
			want: "query-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			usable := isUsable(tc.data.queryToken, tc.data.validUntil)

			queryToken, validUntil := tc.data.GetQueryTokenSelf()

			if queryToken != tc.want {
				t.Errorf("queryToken mismatch\ngot:  %q\nwant: %q", queryToken, tc.want)
			}

			if !validUntil.Equal(tc.data.validUntil) && usable {
				t.Errorf("validUntil mismatch\ngot:  %s\nwant: %s", validUntil, tc.data.validUntil)
			}
		})
	}
}

func TestHubAuth_GetQueryToken(t *testing.T) {
	tests := []struct {
		name     string
		data     *HubAuth
		want     string
		errRegex string
	}{
		{
			name: "no query token",
			data: &HubAuth{
				Token: "token",
			},
			want:     "", // No Username, so refreshQueryToken returns before any request.
			errRegex: `^$`,
		},
		{
			name: "valid at root",
			data: &HubAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(10 * time.Second),
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "valid at defaults",
			data: &HubAuth{
				Token:      "token",
				queryToken: "query-token-root",
				validUntil: time.Now().Add(-10 * time.Second),
				defaults: &HubAuthDefaults{
					Token:      "token",
					queryToken: "query-token-defaults",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
			want:     "query-token-defaults",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			detail := ContainerDetail{
				Image: test.ArgusDockerHubRepo,
				Tag:   "v1.2.3",
			}

			queryToken, err := tc.data.GetQueryToken(detail)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if queryToken != tc.want {
				t.Errorf("queryToken mismatch\ngot:  %q\nwant: %q", queryToken, tc.want)
			}
		})
	}
}

func TestHubAuth_GetQueryToken__refresh(t *testing.T) {
	tests := []struct {
		name            string
		username, token string
		status          int
		wantRequests    int
		want            string
		errRegex        string
	}{
		{
			name:         "refreshed",
			username:     "username",
			token:        "token",
			status:       http.StatusOK,
			wantRequests: 1,
			want:         "fresh-query-token",
			errRegex:     `^$`,
		},
		{
			name:         "no token - not refreshed",
			username:     "username",
			status:       http.StatusOK,
			wantRequests: 0,
			want:         "",
			errRegex:     `^$`,
		},
		{
			name:         "refresh fails",
			username:     "username",
			token:        "token",
			status:       http.StatusInternalServerError,
			wantRequests: 1,
			errRegex:     `^docker-hub token request failed \(status=500\)`,
		},
	}
	_hubTokenAddress := hubTokenAddress

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: swaps hubTokenAddress.

			type tokenRequest struct {
				method, contentType string
				body                []byte
			}
			requests := make(chan tokenRequest, 1)
			handler := func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("io.ReadAll() failed: %v", err)
				}
				requests <- tokenRequest{
					method:      r.Method,
					contentType: r.Header.Get("Content-Type"),
					body:        body,
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"access_token":"fresh-query-token"}`))
			}
			srv := httptest.NewServer(http.HandlerFunc(handler))
			t.Cleanup(srv.Close)
			hubTokenAddress = srv.URL
			t.Cleanup(func() {
				hubTokenAddress = _hubTokenAddress
			})
			data := &HubAuth{
				Username: tc.username,
				Token:    tc.token,
			}

			queryToken, err := data.GetQueryToken(ContainerDetail{Image: test.ArgusDockerHubRepo})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if got := len(requests); got != tc.wantRequests {
				t.Fatalf("len(requests) mismatch\ngot:  %d\nwant: %d", got, tc.wantRequests)
			}
			if tc.wantRequests > 0 {
				req := <-requests
				if req.method != http.MethodPost {
					t.Errorf(
						"Request.Method mismatch\ngot:  %q\nwant: %q",
						req.method, http.MethodPost,
					)
				}
				if req.contentType != "application/json" {
					t.Errorf(
						"Request.Header[\"Content-Type\"] mismatch\ngot:  %q\nwant: %q",
						req.contentType, "application/json",
					)
				}
				var body hubTokenRequest
				if decodeErr := decode.Unmarshal("json", req.body, &body); decodeErr != nil {
					t.Fatalf("decode.Unmarshal() failed: %v", decodeErr)
				}
				if body.Identifier != tc.username || body.Secret != tc.token {
					t.Errorf(
						"hubTokenRequest mismatch\n"+
							"got:  identifier=%q, secret=%q\nwant: identifier=%q, secret=%q",
						body.Identifier, body.Secret,
						tc.username, tc.token,
					)
				}
			}
			if err != nil {
				return
			}

			if queryToken != tc.want {
				t.Errorf("queryToken mismatch\ngot:  %q\nwant: %q", queryToken, tc.want)
			}
			if cached, _ := data.GetQueryTokenSelf(); cached != tc.want {
				t.Errorf(
					"HubAuth.GetQueryTokenSelf() mismatch\ngot:  %q\nwant: %q",
					cached, tc.want,
				)
			}
		})
	}
}

func TestHubAuthDefaults_SetQueryToken(t *testing.T) {
	tests := []struct {
		name                                 string
		data                                 *HubAuth
		setDefaultToken, setHardDefaultToken bool
	}{
		{
			name: "root only",
			data: &HubAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(10 * time.Second),
			},
		},
		{
			name: "defaults - hardDefaults ignored",
			data: &HubAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(11 * time.Second),
				defaults: &HubAuthDefaults{
					Token:      "token1",
					queryToken: "query-token",
					validUntil: time.Now().Add(12 * time.Second),
					defaults: &HubAuthDefaults{
						Token:      "token2",
						queryToken: "query-token",
						validUntil: time.Now().Add(13 * time.Second),
					},
				},
			},
		},
		{
			name: "set in defaults",
			data: &HubAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(11 * time.Second),
				defaults: &HubAuthDefaults{
					Token:      "token",
					queryToken: "query-token",
					validUntil: time.Now().Add(12 * time.Second),
					defaults: &HubAuthDefaults{
						Token:      "token1",
						queryToken: "query-token",
						validUntil: time.Now().Add(13 * time.Second),
					},
				},
			},
			setDefaultToken: true,
		},
		{
			name: "no root credentials - set in defaults",
			data: &HubAuth{
				queryToken: "query-token",
				validUntil: time.Now().Add(11 * time.Second),
				defaults: &HubAuthDefaults{
					Token:      "token",
					queryToken: "query-token",
					validUntil: time.Now().Add(12 * time.Second),
				},
			},
			setDefaultToken: true,
		},
		{
			name: "hard defaults ignored if defaults override",
			data: &HubAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(11 * time.Second),
				defaults: &HubAuthDefaults{
					Token:      "token1",
					queryToken: "query-token",
					validUntil: time.Now().Add(12 * time.Second),
					defaults: &HubAuthDefaults{
						Token:      "token",
						queryToken: "query-token",
						validUntil: time.Now().Add(13 * time.Second),
					},
				},
			},
		},
		{
			name: "set in defaults and hard defaults",
			data: &HubAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(11 * time.Second),
				defaults: &HubAuthDefaults{
					Token:      "token",
					queryToken: "query-token",
					validUntil: time.Now().Add(12 * time.Second),
					defaults: &HubAuthDefaults{
						Token:      "token",
						queryToken: "query-token",
						validUntil: time.Now().Add(13 * time.Second),
					},
				},
			},
			setDefaultToken:     true,
			setHardDefaultToken: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			queryToken := "new-query-token"
			validUntil := time.Now().Add(10 * time.Second)

			tc.data.SetQueryToken(queryToken, validUntil)

			check := func(d *HubAuthDefaults, wantMatch bool) {
				t.Helper()
				if d == nil {
					if wantMatch {
						t.Error("HubAuthDefaults mismatch\ngot:  nil\nwant: non-nil")
					}
					return
				}
				assertQueryTokenSet(
					t, "HubAuthDefaults",
					d.queryToken, d.validUntil,
					queryToken, validUntil,
					wantMatch,
				)
			}

			check(&tc.data.HubAuthDefaults, true)
			check(tc.data.defaults, tc.setDefaultToken)
			if tc.data.defaults != nil {
				check(tc.data.defaults.defaults, tc.setHardDefaultToken)
			}
		})
	}
}

// ######################
// # AUTH | INHERITANCE #
// ######################

func TestHubAuth_Inherit(t *testing.T) {
	tests := []struct {
		name                 string
		auth                 *HubAuth
		from                 RegistryAuth
		srcDetail, dstDetail ContainerDetail
		inherit              bool
	}{
		{
			name: "inherit from nil",
			auth: &HubAuth{},
			from: nil,
		},
		{
			name: "inherit from HubAuth (src.Token is SecretValue)",
			auth: &HubAuth{
				Username: "u",
				Token:    util.SecretValue,
			},
			from: &HubAuth{
				Username:   "u",
				Token:      "abc",
				queryToken: "qt",
				validUntil: time.Now(),
			},
			srcDetail: ContainerDetail{
				Image: "a",
				Tag:   "b",
			},
			dstDetail: ContainerDetail{
				Image: "a",
				Tag:   "b",
			},
			inherit: true,
		},
		{
			name: "inherit from HubAuth when Details do not match (src.Token is SecretValue)",
			auth: &HubAuth{
				Username: "u",
				Token:    util.SecretValue,
			},
			from: &HubAuth{
				Username:   "u",
				Token:      "abc",
				queryToken: "qt",
				validUntil: time.Now(),
			},
			srcDetail: ContainerDetail{
				Image: "a",
				Tag:   "b",
			},
			dstDetail: ContainerDetail{
				Image: "c",
				Tag:   "d",
			},
			inherit: true,
		},
		{
			name: "do not inherit from HubAuth when src.Username is changed",
			auth: &HubAuth{
				Username: "u1",
				Token:    util.SecretValue,
			},
			from: &HubAuth{
				Username:   "u",
				Token:      "abc",
				queryToken: "qt",
				validUntil: time.Now(),
			},
		},
		{
			name: "do not inherit from HubAuth when src.Token is not SecretValue",
			auth: &HubAuth{
				Token: "foo",
			},
			from: &HubAuth{
				Token:      "abc",
				queryToken: "qt",
				validUntil: time.Now(),
			},
		},
		{
			name: "do not inherit from GHCRAuth",
			auth: &HubAuth{
				Token: util.SecretValue,
			},
			from: &GHCRAuth{
				Token: "abc",
			},
		},
		{
			name: "do not inherit from QuayAuth",
			auth: &HubAuth{
				Token: util.SecretValue,
			},
			from: &QuayAuth{
				Token: "abc",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := getTokenData(tc.auth)

			tc.auth.Inherit(tc.from, tc.srcDetail, tc.dstDetail)

			assertInherited(t, "HubAuth", had, tc.auth, tc.from, tc.inherit)
		})
	}
}
