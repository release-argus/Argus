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
	tests := []struct {
		name         string
		format, data string
		registry     *CommonRegistryDefaults
		errRegex     string
		want         string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			registry: &CommonRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &CommonRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
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
			name:   "JSON/auth invalid data type",
			format: "json",
			data:   `{"auth": []}`,
			registry: &CommonRegistryDefaults{
				Auth: &GHCRAuthDefaults{},
			},
			errRegex: test.TrimYAML(`
				^auth:
					json: .*unmarshal.*$`,
			),
		},
		{
			name:   "YAML/auth invalid data type",
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
			registry: &CommonRegistryDefaults{
				Auth: &GHCRAuthDefaults{},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
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
			registry: &CommonRegistryDefaults{
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
			registry: &CommonRegistryDefaults{
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
				func(format string, data []byte) (*CommonRegistryDefaults, error) {
					err := decode.Unmarshal(format, data, tc.registry)
					return tc.registry, err
				},
				tc.format, tc.data,
				func(v *CommonRegistryDefaults) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				"CommonRegistryDefaults.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestCommonRegistry_Unmarshal(t *testing.T) {
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
			data:     "",
			registry: &CommonRegistry{},
			errRegex: test.TrimYAML(`
				^jsontext:
					unexpected EOF$`,
			),
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &CommonRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*CommonRegistry, error) {
					err := decode.Unmarshal(format, data, tc.registry)
					return tc.registry, err
				},
				tc.format, tc.data,
				func(v *CommonRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"CommonRegistry.Unmarshal",
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
			name: "non-empty/Type",
			registry: &CommonRegistry{
				Type: "hub",
			},
			want: false,
		},
		{
			name: "non-empty/Image",
			registry: &CommonRegistry{
				Image: "test/app",
			},
			want: false,
		},
		{
			name: "non-empty/Tag",
			registry: &CommonRegistry{
				Tag: "1.2.3",
			},
			want: false,
		},
		{
			name: "non-empty/ContainerDetail",
			registry: &CommonRegistry{
				Image: "test/app",
				Tag:   "1.2.3",
			},
			want: false,
		},
		{
			name: "non-empty/Auth",
			registry: &CommonRegistry{
				Auth: &HubAuth{
					Token:      "foo",
					queryToken: "bar",
					validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			registry: &CommonRegistry{
				Type:  "hub",
				Image: "test/app",
				Tag:   "1.2.3",
				Auth: &HubAuth{
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

func TestCommonRegistry_Clone(t *testing.T) {
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
				Image: test.ArgusDockerGHCRRepo,
				Tag:   "1.2.3",
				Auth: &HubAuth{
					Username:   "u1",
					Token:      "t1",
					queryToken: "qT",
					validUntil: time.Now(),
					defaults:   &HubAuthDefaults{},
				},
				defaults: &HubRegistryDefaults{},
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

			result := tc.registry.Clone()

			if got := decode.ToYAMLString(result, ""); got != tc.want {
				t.Fatalf("CommonRegistry mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
			if result == nil {
				return
			}

			if got := result.defaults; got != tc.registry.defaults {
				t.Fatalf(
					"CommonRegistry.defaults mismatch\ngot:  %p\nwant: %p",
					got, tc.registry.defaults,
				)
			}

			if result.Auth == tc.registry.Auth && (result.Auth != nil || tc.registry.Auth != nil) {
				t.Fatalf(
					"CommonRegistry.Auth: copy aliases the original\ngot:  %p\nwant: NOT %p",
					result.Auth, tc.registry.Auth,
				)
			}
		})
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

func TestCommonRegistry_String(t *testing.T) {
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
				Image: "test/app",
				Tag:   "{{ version }}",
				Auth: &HubAuth{
					Token:      "_token_",
					queryToken: "_queryToken_",
					validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					defaults: &HubAuthDefaults{
						Token:      "_other_token_",
						queryToken: "_other_queryToken_",
						validUntil: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				},
				defaults: &HubRegistryDefaults{
					Auth: &HubAuthDefaults{
						Token:      "_other_token_",
						queryToken: "_other_queryToken_",
						validUntil: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
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
				"CommonRegistry.String",
				tc.registry.String,
				tc.want,
			)
		})
	}
}

// #######################
// # REGISTRY | DEFAULTS #
// #######################

func TestCommonRegistry_Defaults(t *testing.T) {
	var fresh CommonRegistry

	got := fresh.Defaults()

	if got != nil {
		t.Errorf("CommonRegistry.Defaults() mismatch\ngot:  %p\nwant: nil", got)
	}

	for _, dType := range PossibleTypes {
		t.Run(dType, func(t *testing.T) {
			t.Parallel()

			defaults := Defaults{
				Registry: RegistryDefaultsSet{
					ECR: &ECRRegistryDefaults{
						Auth: &ECRAuthDefaults{},
					},
					GHCR: &GHCRRegistryDefaults{
						Auth: &GHCRAuthDefaults{
							Token: "ghcr-token",
						},
					},
					Hub: &HubRegistryDefaults{
						Auth: &HubAuthDefaults{
							Token:    "hub-token",
							Username: "hub-username",
						},
					},
					Quay: &QuayRegistryDefaults{
						Auth: &QuayAuthDefaults{
							Token: "quay-token",
						},
					},
				},
			}
			registry := CommonRegistry{Auth: RegistryMap[dType]().GetAuth()}

			registry.SetDefaults(dType, &defaults)

			gotDefaults := registry.Defaults()
			if gotDefaults == nil {
				t.Fatal("CommonRegistry.Defaults() mismatch\ngot:  nil\nwant: non-nil")
			}
			var wantDefaults RegistryDefaults
			if wantDefaults = getRegistryDefaults(dType, &defaults); wantDefaults == nil {
				t.Fatal("getRegistryDefaults() mismatch\ngot:  nil\nwant: non-nil")
			}
			if testErr := test.AssertFields(
				t,
				[]test.FieldAssertion{
					{Name: "Defaults()", Got: gotDefaults, Want: wantDefaults, Mode: test.CompareSamePointer},
					{Name: "Defaults().GetAuth()", Got: gotDefaults.GetAuth(), Want: wantDefaults.GetAuth(), Mode: test.CompareSamePointer},
				},
				"CommonRegistry",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestCommonRegistry_Defaults__nil(t *testing.T) {
	var registry CommonRegistry

	registry.SetDefaults(PossibleTypes[0], nil)

	if got := registry.Defaults(); got != nil {
		t.Fatalf("CommonRegistry.Defaults() mismatch\ngot:  %p\nwant: nil", got)
	}
}

func TestCommonRegistry_SetDefaults__unknownType(t *testing.T) {
	dockerType := "ghcr"

	registry := RegistryMap[dockerType]()
	defaults := Defaults{
		Registry: RegistryDefaultsSet{
			GHCR: &GHCRRegistryDefaults{
				Auth: &GHCRAuthDefaults{
					Token: "ghcr-token",
				},
			},
			Hub: &HubRegistryDefaults{
				Auth: &HubAuthDefaults{
					Token:    "hub-token",
					Username: "hub-username",
				},
			},
			Quay: &QuayRegistryDefaults{
				Auth: &QuayAuthDefaults{
					Token: "quay-token",
				},
			},
		},
	}
	registry.SetDefaults(dockerType, &defaults)
	originalDefaults := registry.Defaults()
	if originalDefaults == nil {
		t.Fatal("known type: CommonRegistry.Defaults() mismatch\ngot:  nil\nwant: non-nil")
	}

	registry.SetDefaults("unknown", &defaults)

	if got := registry.Defaults(); got != originalDefaults {
		t.Fatalf(
			"unknown type: CommonRegistry.Defaults() mismatch\ngot:  %p\nwant: %p",
			got, originalDefaults,
		)
	}

	defaults.Registry.GHCR = nil
	registry.SetDefaults(dockerType, &defaults)

	if got := registry.Defaults(); got != originalDefaults {
		t.Fatalf(
			"nil type defaults: CommonRegistry.Defaults() mismatch\ngot:  %p\nwant: %p",
			got, originalDefaults,
		)
	}
}

// #######################
// # REGISTRY | METADATA #
// #######################

func TestCommonRegistry_GetTypeSelf(t *testing.T) {
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

			got := tc.registry.GetTypeSelf()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestCommonRegistry_GetImageSelf(t *testing.T) {
	tests := []struct {
		name     string
		registry CommonRegistry
		want     string
	}{
		{
			name:     "empty",
			registry: CommonRegistry{},
			want:     "",
		},
		{
			name: "image set",
			registry: CommonRegistry{
				Image: "image-here",
			},
			want: "image-here",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.registry.GetImageSelf()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestCommonRegistry_GetImage(t *testing.T) {
	tests := []struct {
		name     string
		registry CommonRegistry
		want     string
	}{
		{
			name:     "empty image",
			registry: CommonRegistry{},
			want:     "",
		},
		{
			name: "image set",
			registry: CommonRegistry{
				Image: "test/app",
			},
			want: "test/app",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.registry.GetImage()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestCommonRegistry_GetTagSelf(t *testing.T) {
	tests := []struct {
		name     string
		registry CommonRegistry
		want     string
	}{
		{
			name:     "empty",
			registry: CommonRegistry{},
			want:     "",
		},
		{
			name: "tag set",
			registry: CommonRegistry{
				Tag: "tag-here",
			},
			want: "tag-here",
		},
		{
			name: "tag ignored in defaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Defaults: &ContainerDetailDefaults{
						Tag: "tag-defaults",
					},
				},
			},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.registry.GetTagSelf()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestCommonRegistry_GetTag(t *testing.T) {
	tests := []struct {
		name     string
		registry CommonRegistry
		want     string
	}{
		{
			name:     "empty tag",
			registry: CommonRegistry{},
			want:     "",
		},
		{
			name: "Tag from instance",
			registry: CommonRegistry{
				Tag: "t-instance",
			},
			want: "t-instance",
		},
		{
			name: "Tag from defaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Defaults: &ContainerDetailDefaults{
						Tag: "t-defaults",
					},
				},
			},
			want: "t-defaults",
		},
		{
			name: "Tag from hardDefaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Defaults: &ContainerDetailDefaults{
						Defaults: &ContainerDetailDefaults{
							Tag: "t-hardDefaults",
						},
					},
				},
			},
			want: "t-hardDefaults",
		},
		{
			name: "Tag from instance, ignore defaults and hardDefaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "t-instance",
					Defaults: &ContainerDetailDefaults{
						Tag: "t-defaults",
						Defaults: &ContainerDetailDefaults{
							Tag: "t-hardDefaults",
						},
					},
				},
			},
			want: "t-instance",
		},
		{
			name: "Tag from defaults, ignore hardDefaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Defaults: &ContainerDetailDefaults{
						Tag: "t-defaults",
						Defaults: &ContainerDetailDefaults{
							Tag: "t-hardDefaults",
						},
					},
				},
			},
			want: "t-defaults",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.registry.GetTag()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

// #########################
// # REGISTRY | VALIDATION #
// #########################

func TestCommonRegistry_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *CommonRegistry
		errRegex string
	}{
		{
			name:     "image:tag at root",
			errRegex: `^$`,
			input: &CommonRegistry{
				Image: "test/app",
				Tag:   "1.2.3",
			},
		},
		{
			name:     "image: missing",
			errRegex: `^image: <required>.*$`,
			input: &CommonRegistry{
				Image: "",
				Tag:   "1.2.3",
			},
		},
		{
			name:     "image: with period in name",
			errRegex: `^$`,
			input: &CommonRegistry{
				Image: "test/image.io",
				Tag:   "1.2.3",
			},
		},
		{
			name:     "image: invalid",
			errRegex: `image: .* <invalid>`,
			input: &CommonRegistry{
				Image: "	test/app",
				Tag:   "1.2.3",
			},
		},
		{
			name:     "tag: missing",
			errRegex: `^tag: <required>.*$`,
			input: &CommonRegistry{
				Image: "test/app",
				Tag:   "",
			},
		},
		{
			name:     "tag: invalid templating",
			errRegex: `^tag: .* <invalid>.*$`,
			input: &CommonRegistry{
				Image: "test/app",
				Tag:   "{{ version }",
			},
		},
		{
			name:     "tag: invalid url encoding",
			errRegex: `^tag: .* <invalid>.*$`,
			input: &CommonRegistry{
				Image: "test/app",
				Tag:   "1.2	.3+",
			},
		},
		{
			name:     "tag in defaults",
			errRegex: `^$`,
			input: &CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "test/app",
					Tag:   "",
					Defaults: &ContainerDetailDefaults{
						Tag: "1.2.3",
					},
				},
			},
		},
		{
			name:     "auth err",
			errRegex: `^token: <required>.*$`,
			input: &CommonRegistry{
				Image: "test/app",
				Tag:   "1.2.3",
				Auth: &HubAuth{
					Username: "someone",
				},
			},
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

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestCommonRegistry_GetTagForVersion(t *testing.T) {
	tests := []struct {
		name     string
		registry CommonRegistry
		version  string
		want     string
	}{
		{
			name: "empty tag",
			registry: CommonRegistry{
				Tag: "",
			},
			version: "3.2.1",
			want:    "",
		},
		{
			name: "plain version",
			registry: CommonRegistry{
				Tag: "1.2.3",
			},
			version: "3.2.1",
			want:    "1.2.3",
		},
		{
			name: "version template",
			registry: CommonRegistry{
				Tag: "{{ version }}.1",
			},
			version: "3.2",
			want:    "3.2.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.registry.GetTagForVersion(tc.version)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestCommonRegistry_ParseBody(t *testing.T) {
	tests := []struct {
		name     string
		errRegex string
		resp     *http.Response
	}{
		{
			name:     "200",
			errRegex: `^$`,
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

	registry := CommonRegistry{
		Image: "test/app",
		Tag:   "ver",
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := registry.parseBody(registry.Tag, tc.resp)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestCommonRegistry_Detail(t *testing.T) {
	tests := []struct {
		name     string
		registry CommonRegistry
		want     ContainerDetail
	}{
		{
			name:     "empty",
			registry: CommonRegistry{},
			want:     ContainerDetail{},
		},
		{
			name: "image+tag from root",
			registry: CommonRegistry{
				Image: "foo",
				Tag:   "bar",
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
					Defaults: &ContainerDetailDefaults{
						Tag: "default-bar",
					},
				},
			},
			want: ContainerDetail{
				Image: "foo",
				Tag:   "default-bar",
			},
		},
		{
			name: "tag from root, no image",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Tag: "bar",
					Defaults: &ContainerDetailDefaults{
						Tag: "default-bar",
					},
				},
			},
			want: ContainerDetail{
				Image: "",
				Tag:   "bar",
			},
		},
		{
			name: "tag from defaults of defaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "foo",
					Defaults: &ContainerDetailDefaults{
						Defaults: &ContainerDetailDefaults{
							Tag: "default-default-bar",
						},
					},
				},
			},
			want: ContainerDetail{
				Image: "foo",
				Tag:   "default-default-bar",
			},
		},
		{
			name: "root tag wins over defaults",
			registry: CommonRegistry{
				ContainerDetail: ContainerDetail{
					Image: "foo",
					Tag:   "bar",
					Defaults: &ContainerDetailDefaults{
						Tag: "default-bar",
					},
				},
			},
			want: ContainerDetail{
				Image: "foo",
				Tag:   "bar",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.registry.Detail()

			want := decode.ToYAMLString(tc.want, "")
			if got := decode.ToYAMLString(result, ""); got != want {
				t.Errorf("ContainerDetail mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

// ##########################
// # REGISTRY | INHERITANCE #
// ##########################

func TestCommonRegistry_Inherit(t *testing.T) {
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
					Username:   "u",
					Token:      "t",
					queryToken: "qt",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			from:    nil,
			inherit: false,
		},
		{
			name: "nil registry Auth to inherit from",
			registry: CommonRegistry{
				Auth: &HubAuth{
					Username:   "u",
					Token:      "t",
					queryToken: "qt",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			from:    &GHCRRegistry{},
			inherit: false,
		},
		{
			name: "inherit from GHCR/same auth",
			registry: CommonRegistry{
				Auth: &GHCRAuth{
					Token:      "t",
					queryToken: "qt",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			from: &GHCRRegistry{
				Auth: &GHCRAuth{
					Token:      "t",
					queryToken: "qt-new",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			inherit: true,
		},
		{
			name: "inherit from GHCR/different auth",
			registry: CommonRegistry{
				Auth: &GHCRAuth{
					Token:      "t",
					queryToken: "qt",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			from: &GHCRRegistry{
				Auth: &GHCRAuth{
					Token:      "t1",
					queryToken: "qt-new",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			inherit: false,
		},
		{
			name: "inherit from Hub/same auth",
			registry: CommonRegistry{
				Auth: &HubAuth{
					Username:   "u",
					Token:      "t",
					queryToken: "qt",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			from: &HubRegistry{
				Auth: &HubAuth{
					Username:   "u",
					Token:      "t",
					queryToken: "qt-new",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			inherit: true,
		},
		{
			name: "inherit from Hub/different auth/different username, same token",
			registry: CommonRegistry{
				Auth: &HubAuth{
					Username:   "u",
					Token:      "t",
					queryToken: "qt",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			from: &HubRegistry{
				Auth: &HubAuth{
					Username:   "u1",
					Token:      "t",
					queryToken: "qt-new",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			inherit: false,
		},
		{
			name: "inherit from Hub/different auth/same username, different token",
			registry: CommonRegistry{
				Auth: &HubAuth{
					Username:   "u",
					Token:      "t",
					queryToken: "qt",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			from: &HubRegistry{
				Auth: &HubAuth{
					Username:   "u",
					Token:      "t1",
					queryToken: "qt-new",
					validUntil: time.Now().UTC().Add(10 * time.Minute),
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			inherit: false,
		},
		{
			name: "inherit from Quay/different token",
			registry: CommonRegistry{
				Auth: &QuayAuth{
					Token: "t",
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			from: &QuayRegistry{
				Auth: &QuayAuth{
					Token: "t-new",
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			inherit: false,
		},
		{
			name: "inherit from Quay/secret token restored",
			registry: CommonRegistry{
				Auth: &QuayAuth{
					Token: util.SecretValue,
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			from: &QuayRegistry{
				Auth: &QuayAuth{
					Token: "t",
				},
				Image: "test/app",
				Tag:   "{{ latest_version }}",
			},
			inherit: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			had := getTokenData(tc.registry.Auth)
			var fromAuth RegistryAuth
			if tc.from != nil {
				fromAuth = tc.from.GetAuth()
			}

			tc.registry.Inherit(tc.from)

			assertInherited(
				t, "CommonRegistry.GetAuth()",
				had, tc.registry.GetAuth(), fromAuth, tc.inherit,
			)
		})
	}
}

// ###################
// # REGISTRY | AUTH #
// ###################

func TestCommonRegistryDefaults_GetAuth(t *testing.T) {
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
			name: "quay auth",
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

			registry := CommonRegistryDefaults{}
			registry.Auth = tc.auth

			got := registry.GetAuth()

			if testErr := test.AssertFields(
				t,
				[]test.FieldAssertion{
					{Name: "GetAuth()", Got: got, Want: tc.auth, Mode: test.CompareSamePointer},
				},
				"CommonRegistryDefaults",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestCommonRegistry_GetAuth(t *testing.T) {
	tests := []struct {
		name string
		auth RegistryAuth
	}{
		{
			name: "ghcr auth",
			auth: &GHCRAuth{
				Token: "abc",
			},
		},
		{
			name: "hub auth",
			auth: &HubAuth{
				Username: "123",
				Token:    "abc",
			},
		},
		{
			name: "quay auth",
			auth: &QuayAuth{
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

			registry := CommonRegistry{}
			registry.Auth = tc.auth

			got := registry.GetAuth()

			if testErr := test.AssertFields(
				t,
				[]test.FieldAssertion{
					{Name: "GetAuth()", Got: got, Want: tc.auth, Mode: test.CompareSamePointer},
				},
				"CommonRegistry",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
