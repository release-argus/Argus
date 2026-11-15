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
	"encoding/base64"
	"sync"
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

func TestGHCRRegistryDefaults_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     string
		registry *GHCRRegistryDefaults
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			registry: &GHCRRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &GHCRRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &GHCRRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &GHCRRegistryDefaults{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &GHCRRegistryDefaults{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &GHCRRegistryDefaults{
				Auth: &GHCRAuth{},
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
			registry: &GHCRRegistryDefaults{
				Auth: &GHCRAuth{},
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
			registry: &GHCRRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
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
			registry: &GHCRRegistryDefaults{
				Auth: &GHCRAuth{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				auth:
					token: tOKEn
			`),
		},
		{
			name:   "YAML/auth-ghcr",
			format: "yaml",
			data: test.TrimYAML(`
				auth:
					username: ghcr-username
					token: tOKEn
			`),
			registry: &GHCRRegistryDefaults{
				Auth: &GHCRAuth{},
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
				func(format string, data []byte) (*GHCRRegistryDefaults, error) {
					err := decode.Unmarshal(format, data, tc.registry)
					return tc.registry, err
				},
				tc.format, tc.data,
				func(v *GHCRRegistryDefaults) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"GHCRRegistryDefaults.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}

			if tc.registry.Auth == nil {
				t.Error("GHCRRegistryDefaults.Auth mismatch\ngot:  nil\nwant: non-nil")
			}
		})
	}
}

func TestGHCRRegistry_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     string
		registry *GHCRRegistry
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			registry: &GHCRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &GHCRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &GHCRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &GHCRRegistry{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &GHCRRegistry{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid ContainerDetail",
			format: "json",
			data:   `{"image": []}`,
			registry: &GHCRRegistry{
				Auth: &GHCRAuth{},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &GHCRRegistry{
				Auth: &GHCRAuth{},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &GHCRRegistry{
				Auth: &GHCRAuth{},
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
			registry: &GHCRRegistry{
				Auth: &GHCRAuth{},
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
			registry: &GHCRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
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
			registry: &GHCRRegistry{
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
			registry: &GHCRRegistry{
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			registry, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*GHCRRegistry, error) {
					err := decode.Unmarshal(format, data, tc.registry)
					return tc.registry, err
				},
				tc.format, tc.data,
				func(v *GHCRRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"GHCRRegistry.Unmarshal",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}

			if registry.Auth == nil {
				t.Error("GHCRRegistry.Auth mismatch\ngot:  nil\nwant: non-nil")
			}
		})
	}
}

func TestGHCRRegistry_ApplyOverrides(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     string
		registry *GHCRRegistry
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			registry: &GHCRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &GHCRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &GHCRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &GHCRRegistry{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &GHCRRegistry{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid ContainerDetail",
			format: "json",
			data:   `{"image": []}`,
			registry: &GHCRRegistry{
				Auth: &GHCRAuth{},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &GHCRRegistry{
				Auth: &GHCRAuth{},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &GHCRRegistry{
				Auth: &GHCRAuth{},
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
			registry: &GHCRRegistry{
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
			registry: &GHCRRegistry{
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
			registry: &GHCRRegistry{
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
			name:   "auth-ghcr replaced",
			format: "yaml",
			data: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: ghcr-username
					token: tOKEn
			`),
			registry: &GHCRRegistry{
				Auth: &GHCRAuth{
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
					username: ghcr-username
			`),
			registry: &GHCRRegistry{
				Image: "i",
				Auth: &GHCRAuth{
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
				func(format string, data []byte, v *GHCRRegistry) (*GHCRRegistry, error) {
					err := v.ApplyOverrides(format, data)
					return v, err
				},
				tc.format, tc.data,
				func(v *GHCRRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				true,
				"GHCRRegistry.ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// ####################
// # REGISTRY | STATE #
// ####################

func TestGHCRRegistryDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		registry *GHCRRegistryDefaults
		want     bool
	}{
		{
			name:     "nil",
			registry: nil,
			want:     true,
		},
		{
			name:     "empty",
			registry: &GHCRRegistryDefaults{},
			want:     true,
		},
		{
			name: "non-empty/Token",
			registry: &GHCRRegistryDefaults{
				Auth: &GHCRAuthDefaults{
					Token: "foo",
				},
			},
			want: false,
		},
		{
			name: "non-empty/queryToken and validUntil",
			registry: &GHCRRegistryDefaults{
				Auth: &GHCRAuthDefaults{
					queryToken: "bar",
					validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want: true, // IsZero ignores the cached query token.
		},
		{
			name: "non-empty/Token, queryToken and validUntil",
			registry: &GHCRRegistryDefaults{
				Auth: &GHCRAuthDefaults{
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

func TestGHCRRegistry_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data *GHCRRegistry
		want bool
	}{
		{
			name: "nil",
			data: nil,
			want: true,
		},
		{
			name: "empty",
			data: RegistryMap["ghcr"]().(*GHCRRegistry),
			want: true,
		},
		{
			name: "non-empty/Type",
			data: &GHCRRegistry{
				Type: "abc",
				Auth: RegistryMap["ghcr"]().GetAuth(),
			},
			want: false,
		},
		{
			name: "non-empty/ContainerDetail",
			data: &GHCRRegistry{
				Image: "i",
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &GHCRRegistry{
				Image: "i",
				Type:  "abc",
				Auth:  RegistryMap["ghcr"]().GetAuth(),
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

func TestGHCRRegistry_Copy(t *testing.T) {
	tests := []struct {
		name     string
		registry *GHCRRegistry
		want     string
	}{
		{
			name:     "nil",
			registry: nil,
			want:     "null\n",
		},
		{
			name:     "empty",
			registry: RegistryMap["ghcr"]().(*GHCRRegistry),
			want:     "{}\n",
		},
		{
			name: "filled",
			registry: &GHCRRegistry{
				Image: "i1",
				Tag:   "t1",
				Auth: &GHCRAuth{
					Token:      "t1",
					queryToken: "qT",
					validUntil: time.Now().Add(time.Hour),
					defaults:   &GHCRAuthDefaults{},
				},
				defaults: &GHCRRegistryDefaults{},
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
				t.Fatalf("GHCRRegistry mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			got, ok := gotInterface.(*GHCRRegistry)
			if !ok {
				if gotInterface == nil {
					return
				}
				t.Fatalf(
					"GHCRRegistry.Copy() type mismatch\ngot:  %T\nwant: *GHCRRegistry",
					gotInterface,
				)
			}

			gotAuth, ok := got.GetAuth().(*GHCRAuth)
			hadAuth, _ := tc.registry.GetAuth().(*GHCRAuth)
			if !ok {
				t.Fatalf(
					"GHCRRegistry.GetAuth() type mismatch\ngot:  %T\nwant: *GHCRAuth",
					got.GetAuth(),
				)
			}
			if gotAuth.queryToken != hadAuth.queryToken ||
				gotAuth.validUntil != hadAuth.validUntil ||
				got.defaults != tc.registry.defaults {
				t.Fatalf(
					"GHCRRegistry mismatch\n"+
						"got:  Auth.queryToken=%q, Auth.validUntil=%s, defaults=%p\n"+
						"want: Auth.queryToken=%q, Auth.validUntil=%s, defaults=%p",
					gotAuth.queryToken, gotAuth.validUntil, got.defaults,
					hadAuth.queryToken, hadAuth.validUntil, tc.registry.defaults,
				)
			}

			if gotAuth == hadAuth {
				t.Fatalf(
					"GHCRRegistry.GetAuth(): copy aliases the original\ngot:  %p\nwant: NOT %p",
					gotAuth, hadAuth,
				)
			}
		})
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

func TestGHCRRegistryDefaults_String(t *testing.T) {
	tests := []struct {
		name string
		data *GHCRRegistryDefaults
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "",
		},
		{
			name: "empty",
			data: &GHCRRegistryDefaults{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &GHCRRegistryDefaults{
				Auth: &GHCRAuth{
					Token: "token1",
					defaults: &GHCRAuthDefaults{
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
				"GHCRRegistryDefaults.String",
				tc.data.String,
				tc.want,
			)
		})
	}
}

func TestGHCRRegistry_String(t *testing.T) {
	tests := []struct {
		name string
		data *GHCRRegistry
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "",
		},
		{
			name: "empty",
			data: &GHCRRegistry{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &GHCRRegistry{
				Type: "test-ghcr",
				ContainerDetail: ContainerDetail{
					Image: "i1",
					Tag:   "t1",
					Defaults: &ContainerDetailDefaults{
						Tag: "t2",
					},
				},
				Auth: &GHCRAuth{
					Token: "token1",
					defaults: &GHCRAuthDefaults{
						Token: "token2",
					},
				},
			},
			want: test.TrimYAML(`
				type: test-ghcr
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
				"GHCRRegistry.String",
				tc.data.String,
				tc.want,
			)
		})
	}
}

// #######################
// # REGISTRY | METADATA #
// #######################

func TestGHCRRegistryDefaults_GetType(t *testing.T) {
	var registry GHCRRegistryDefaults

	got := registry.GetType()

	if want := "ghcr"; got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestGHCRRegistry_GetType(t *testing.T) {
	var registry GHCRRegistry

	got := registry.GetType()

	if want := "ghcr"; got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestGHCRRegistry_NewRequest(t *testing.T) {
	tests := []struct {
		name     string
		registry *GHCRRegistry
		tag      string
		errRegex string
	}{
		{
			name: "no image or tag",
			registry: &GHCRRegistry{
				Image: "",
				Tag:   "",
			},
			errRegex: `^$`,
		},
		{
			name: "have image+tag",
			registry: &GHCRRegistry{
				Image: "123",
				Tag:   "not-used",
			},
			tag:      "foo",
			errRegex: `^$`,
		},
		{
			name: "tag: invalid",
			registry: &GHCRRegistry{
				Image: "123",
				Tag:   "not-used",
			},
			tag: "	foo",
			errRegex: test.TrimYAML(`
				^parse "https://.*
					[^\s]+ invalid control character in URL$`,
			),
		},
		{
			name: "image: invalid",
			registry: &GHCRRegistry{
				Image: "	123",
				Tag:   "not-used",
			},
			tag: "foo",
			errRegex: test.TrimYAML(`
				^parse "https://.*
					[^\s]+ invalid control character in URL$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := tc.registry.newRequest(tc.tag)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			want := "application/vnd.oci.image.index.v1+json"
			if got := req.Header.Get("Accept"); got != want {
				t.Errorf("http.Request.Header[\"Accept\"] mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

// ###################
// # AUTH | DECODING #
// ###################

func TestGHCRAuthDefaults_Unmarshal(t *testing.T) {
	tests := []struct {
		name         string
		format, data string
		want         *GHCRAuthDefaults
		errRegex     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "{}",
			errRegex: `^$`,
			want:     &GHCRAuthDefaults{},
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}\n",
			errRegex: `^$`,
			want:     &GHCRAuthDefaults{},
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
		{
			name:     "JSON/unencoded Token",
			format:   "json",
			data:     `{"token": "ghp_token"}`,
			errRegex: `^$`,
			want: &GHCRAuthDefaults{
				Token:      "ghp_token",
				queryToken: base64.StdEncoding.EncodeToString([]byte("ghp_token")),
			},
		},
		{
			name:     "YAML/unencoded Token",
			format:   "yaml",
			data:     `token: ghp_token`,
			errRegex: `^$`,
			want: &GHCRAuthDefaults{
				Token:      "ghp_token",
				queryToken: base64.StdEncoding.EncodeToString([]byte("ghp_token")),
			},
		},
		{
			name:     "JSON/already encoded Token",
			format:   "json",
			data:     `{"token": "Z2hwX3Rva2Vu"}`,
			errRegex: `^$`,
			want: &GHCRAuthDefaults{
				Token:      "Z2hwX3Rva2Vu",
				queryToken: "Z2hwX3Rva2Vu",
			},
		},
		{
			name:     "YAML/already encoded Token",
			format:   "yaml",
			data:     `token: Z2hwX3Rva2Vu`,
			errRegex: `^$`,
			want: &GHCRAuthDefaults{
				Token:      "Z2hwX3Rva2Vu",
				queryToken: "Z2hwX3Rva2Vu",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			auth, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*GHCRAuthDefaults, error) {
					var decoded GHCRAuthDefaults
					err := decode.Unmarshal(format, data, &decoded)
					return &decoded, err
				},
				tc.format, tc.data,
				func(v *GHCRAuthDefaults) string { return v.String("") },
				decode.ToYAMLString(tc.want, ""),
				tc.errRegex,
				"GHCRAuthDefaults.Unmarshal",
			)
			if testErr != nil {
				t.Fatal(testErr)
			} else if err != nil {
				return
			}

			if auth.queryToken != tc.want.queryToken {
				t.Errorf(
					"GHCRAuthDefaults.queryToken mismatch\ngot:  %q\nwant: %q",
					auth.queryToken, tc.want.queryToken,
				)
			}
		})
	}
}

// ################
// # AUTH | STATE #
// ################

func TestGHCRAuthDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data *GHCRAuthDefaults
		want bool
	}{
		{
			name: "nil",
			data: nil,
			want: true,
		},
		{
			name: "empty",
			data: &GHCRAuthDefaults{},
			want: true,
		},
		{
			name: "filled",
			data: &GHCRAuthDefaults{Token: "t1"},
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

func TestGHCRAuth_Copy(t *testing.T) {
	tests := []struct {
		name string
		auth *GHCRAuth
		want string
	}{
		{
			name: "nil",
			auth: nil,
			want: "null\n",
		},
		{
			name: "empty",
			auth: &GHCRAuth{},
			want: "{}\n",
		},
		{
			name: "filled",
			auth: &GHCRAuth{
				Token:      "t1",
				queryToken: "qT",
				validUntil: time.Now(),
				defaults:   &GHCRAuthDefaults{},
			},
			want: "token: t1\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authCopy := tc.auth.Copy()

			if got := decode.ToYAMLString(authCopy, ""); got != tc.want {
				t.Fatalf("GHCRAuth mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			got, ok := authCopy.(*GHCRAuth)
			if !ok {
				if authCopy == nil {
					return
				}
				t.Fatalf("GHCRAuth.Copy() type mismatch\ngot:  %T\nwant: *GHCRAuth", authCopy)
			}
			if tc.auth == nil {
				return
			}

			if got.queryToken != tc.auth.queryToken ||
				got.validUntil != tc.auth.validUntil ||
				got.defaults != tc.auth.defaults {
				t.Fatalf(
					"GHCRAuth mismatch\n"+
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

func TestGHCRAuthDefaults_String(t *testing.T) {
	tests := []struct {
		name string
		data *GHCRAuthDefaults
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "null\n",
		},
		{
			name: "empty",
			data: &GHCRAuthDefaults{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &GHCRAuthDefaults{
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
				"GHCRAuthDefaults.String",
				tc.data.String,
				tc.want,
			)
		})
	}
}

// ###################
// # AUTH | DEFAULTS #
// ###################

func TestGHCRAuthDefaults_Defaults(t *testing.T) {
	tests := []struct {
		name         string
		data         *GHCRAuthDefaults
		haveDefaults bool
	}{
		{
			name:         "no defaults",
			data:         &GHCRAuthDefaults{},
			haveDefaults: false,
		},
		{
			name: "defaults",
			data: &GHCRAuthDefaults{
				defaults: &GHCRAuthDefaults{
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

func TestGHCRAuthDefaults_SetDefaults(t *testing.T) {
	tests := []struct {
		name        string
		newDefaults RegistryAuthDefaults
		doesSet     bool
	}{
		{
			name:        "give GHCRAuthDefaults",
			newDefaults: &GHCRAuthDefaults{},
			doesSet:     true,
		},
		{
			name:        "doesn't give ECRAuthDefaults",
			newDefaults: &ECRAuthDefaults{},
			doesSet:     false,
		},
		{
			name:        "doesn't give HubAuthDefaults",
			newDefaults: &HubAuthDefaults{},
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

			data := &GHCRAuthDefaults{}

			data.SetDefaults(tc.newDefaults)

			var want *GHCRAuthDefaults
			if tc.doesSet {
				want = tc.newDefaults.(*GHCRAuthDefaults)
			}
			if data.defaults != want {
				t.Errorf(
					"GHCRAuthDefaults.defaults mismatch\ngot:  %p\nwant: %p",
					data.defaults, want,
				)
			}
		})
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

func TestGHCRAuth_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *GHCRAuth
		errRegex string
	}{
		{
			name:     "nil",
			input:    (*GHCRAuth)(nil),
			errRegex: `^$`,
		},
		{
			name: "token",
			input: &GHCRAuth{
				Token: "token",
			},
			errRegex: `^$`,
		},
		{
			name:     "no token",
			input:    &GHCRAuth{},
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

func TestGHCRAuth_GetToken(t *testing.T) {
	tests := []struct {
		name     string
		data     *GHCRAuth
		want     string
		wantSelf string
	}{
		{
			name:     "empty",
			data:     &GHCRAuth{},
			want:     "",
			wantSelf: "",
		},
		{
			name: "no defaults",
			data: &GHCRAuth{
				Token: "root",
			},
			want:     "root",
			wantSelf: "root",
		},
		{
			name: "defaults fallback",
			data: &GHCRAuth{
				Token: "",
				defaults: &GHCRAuthDefaults{
					Token: "defaults",
				},
			},
			want:     "defaults",
			wantSelf: "",
		},
		{
			name: "defaults fallback recursive",
			data: &GHCRAuth{
				Token: "",
				defaults: &GHCRAuthDefaults{
					Token: "",
					defaults: &GHCRAuthDefaults{
						Token: "hard-defaults",
					},
				},
			},
			want:     "hard-defaults",
			wantSelf: "",
		},
		{
			name: "root Token prioritised",
			data: &GHCRAuth{
				Token: "root",
				defaults: &GHCRAuthDefaults{
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
				t.Fatalf("GHCRAuth.GetToken() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			got = tc.data.GetTokenSelf()

			if got != tc.wantSelf {
				t.Fatalf("GHCRAuth.GetTokenSelf() mismatch\ngot:  %q\nwant: %q", got, tc.wantSelf)
			}
		})
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestGHCRAuthDefaults_GetQueryTokenSelf(t *testing.T) {
	tests := []struct {
		name        string
		data        *GHCRAuthDefaults
		want        string
		wantRenewed bool
	}{
		{
			name: "empty",
			data: &GHCRAuthDefaults{},
			want: "",
		},
		{
			name: "expired",
			data: &GHCRAuthDefaults{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(-1 * time.Second),
			},
			want:        "query-token",
			wantRenewed: true,
		},
		{
			name: "valid",
			data: &GHCRAuthDefaults{
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
			originalValidUntil := tc.data.validUntil

			before := time.Now()
			queryToken, validUntil := tc.data.GetQueryTokenSelf()
			after := time.Now()

			if queryToken != tc.want {
				t.Errorf("queryToken mismatch\ngot:  %q\nwant: %q", queryToken, tc.want)
			}

			switch {
			case tc.wantRenewed:
				earliest, latest := before.Add(24*time.Hour), after.Add(24*time.Hour)
				if validUntil.Before(earliest) || validUntil.After(latest) {
					t.Errorf(
						"validUntil mismatch\ngot:  %s\nwant: between %s and %s",
						validUntil, earliest, latest,
					)
				}
			case usable:
				if !validUntil.Equal(originalValidUntil) {
					t.Errorf(
						"validUntil mismatch\ngot:  %s\nwant: %s",
						validUntil, originalValidUntil,
					)
				}
			}
		})
	}
}

func TestGHCRAuthDefaults_GetQueryTokenSelf__parallel(t *testing.T) {
	data := &GHCRAuthDefaults{
		Token:      "token",
		queryToken: "query-token",
		validUntil: time.Now().Add(-10 * time.Second),
	}
	originalQueryToken := data.queryToken

	data.mu.Lock()
	go func() {
		time.Sleep(time.Second)
		data.mu.Unlock()
	}()

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			_, _ = data.GetQueryTokenSelf()
		})
	}
	queryToken, validUntil := data.GetQueryTokenSelf()
	wg.Wait()

	if queryToken != originalQueryToken {
		t.Errorf("queryToken mismatch\ngot:  %q\nwant: %q", queryToken, originalQueryToken)
	}

	// The Token renews the expired validUntil.
	if !isUsable(queryToken, validUntil) {
		t.Errorf(
			"isUsable() mismatch\ngot:  false (validUntil=%s, now=%s)\nwant: true",
			validUntil, time.Now(),
		)
	}
}

func TestGHCRAuth_GetQueryToken(t *testing.T) {
	tests := []struct {
		name     string
		data     *GHCRAuth
		want     string
		errRegex string
	}{
		{
			name: "valid",
			data: &GHCRAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(10 * time.Second),
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "expired root refreshed before defaults fallback",
			data: &GHCRAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(-10 * time.Second),
				defaults: &GHCRAuthDefaults{
					Token:      "token",
					queryToken: "default-query-token",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "expired root refreshed",
			data: &GHCRAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(-10 * time.Second),
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "no root token - defaults query token",
			data: &GHCRAuth{
				defaults: &GHCRAuthDefaults{
					Token:      "token",
					queryToken: "default-query-token",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
			want:     "default-query-token",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			detail := ContainerDetail{
				Image: test.ArgusDockerGHCRRepo,
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

func TestGHCRAuthDefaults_SetQueryToken(t *testing.T) {
	tests := []struct {
		name string
		data *GHCRAuth
	}{
		{
			name: "root only",
			data: &GHCRAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(10 * time.Second),
			},
		},
		{
			name: "defaults - hardDefaults ignored",
			data: &GHCRAuth{
				Token:      "token",
				queryToken: "query-token",
				validUntil: time.Now().Add(11 * time.Second),
				defaults: &GHCRAuthDefaults{
					Token:      "token",
					queryToken: "query-token",
					validUntil: time.Now().Add(12 * time.Second),
					defaults: &GHCRAuthDefaults{
						Token:      "token",
						queryToken: "query-token",
						validUntil: time.Now().Add(13 * time.Second),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			queryToken := "new-query-token"
			validUntil := time.Now().Add(10 * time.Second)

			tc.data.SetQueryToken(queryToken, validUntil)

			check := func(d *GHCRAuthDefaults, wantMatch bool) {
				t.Helper()
				if d == nil {
					if wantMatch {
						t.Error("GHCRAuthDefaults mismatch\ngot:  nil\nwant: non-nil")
					}
					return
				}
				assertQueryTokenSet(
					t, "GHCRAuthDefaults",
					d.queryToken, d.validUntil,
					queryToken, validUntil,
					wantMatch,
				)
			}

			check(&tc.data.GHCRAuthDefaults, true)
			// GHCR query tokens are repo-scoped.
			check(tc.data.defaults, false)
			if tc.data.defaults != nil {
				check(tc.data.defaults.defaults, false)
			}
		})
	}
}

func TestGHCRAuth_RefreshQueryToken__cached(t *testing.T) {
	auth := &GHCRAuth{
		queryToken: "cached-token",
		validUntil: time.Now().Add(time.Hour),
	}
	detail := ContainerDetail{Image: test.ArgusDockerGHCRRepo}

	queryToken, err := auth.refreshQueryToken(detail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := queryToken, "cached-token"; got != want {
		t.Fatalf("queryToken mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

// ######################
// # AUTH | INHERITANCE #
// ######################

func TestGHCRAuth_Inherit(t *testing.T) {
	tests := []struct {
		name                 string
		auth                 *GHCRAuth
		from                 RegistryAuth
		srcDetail, dstDetail ContainerDetail
		wantToken            string
		inherit              bool
	}{
		{
			name: "inherit from nil",
			auth: &GHCRAuth{},
			from: nil,
		},
		{
			name: "inherit from GHCRAuth (src.Token is SecretValue)",
			auth: &GHCRAuth{
				Token: util.SecretValue,
			},
			from: &GHCRAuth{
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
			wantToken: "abc",
			inherit:   true,
		},
		{
			name: "do not inherit when Details do not match (Token is SecretValue)",
			auth: &GHCRAuth{
				Token: util.SecretValue,
			},
			from: &GHCRAuth{
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
			wantToken: util.SecretValue,
		},
		{
			name: "do not inherit from GHCRAuth when src.Token is not SecretValue",
			auth: &GHCRAuth{
				Token: "foo",
			},
			from: &GHCRAuth{
				Token:      "abc",
				queryToken: "qt",
				validUntil: time.Now(),
			},
			wantToken: "foo",
		},
		{
			name: "do not inherit from HubAuth",
			auth: &GHCRAuth{
				Token: util.SecretValue,
			},
			from: &HubAuth{
				Token: "abc",
			},
			wantToken: util.SecretValue,
		},
		{
			name: "do not inherit from QuayAuth",
			auth: &GHCRAuth{
				Token: util.SecretValue,
			},
			from: &QuayAuth{
				Token: "abc",
			},
			wantToken: util.SecretValue,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hadQueryToken, hadValidUntil := tc.auth.queryToken, tc.auth.validUntil

			tc.auth.Inherit(tc.from, tc.srcDetail, tc.dstDetail)

			if tc.auth.Token != tc.wantToken {
				t.Errorf("GHCRAuth.Token mismatch\ngot:  %q\nwant: %q", tc.auth.Token, tc.wantToken)
			}

			wantQueryToken, wantValidUntil := hadQueryToken, hadValidUntil
			if tc.inherit {
				from := getTokenData(tc.from)
				wantQueryToken, wantValidUntil = from.queryToken, from.validUntil
			}
			if tc.auth.queryToken != wantQueryToken {
				t.Errorf(
					"GHCRAuth.queryToken mismatch\ngot:  %q\nwant: %q",
					tc.auth.queryToken, wantQueryToken,
				)
			}
			if !tc.auth.validUntil.Equal(wantValidUntil) {
				t.Errorf(
					"GHCRAuth.validUntil mismatch\ngot:  %s\nwant: %s",
					tc.auth.validUntil, wantValidUntil,
				)
			}
		})
	}
}
