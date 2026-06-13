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
	"fmt"
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
	// GIVEN: a GHCRRegistryDefaults and JSON/YAML to unmarshal into it.
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
			name:   "JSON/invalid ContainerDetail",
			format: "json",
			data:   `{"image": []}`,
			registry: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &GHCRAuth{},
				},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &GHCRAuth{},
				},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &GHCRAuth{},
				},
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &GHCRAuth{},
				},
			},
			errRegex: test.TrimYAML(`
				^auth:
					[^\s]+ sequence was used where mapping is expected`,
			),
		},
		{
			name:     "auth: null",
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &GHCRAuth{},
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
			name:   "YAML/auth-ghcr",
			format: "yaml",
			data: test.TrimYAML(`
				image: i
				tag: t
				auth:
					username: ghcr-username
					token: tOKEn
			`),
			registry: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &GHCRAuth{},
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, testErr := test.AssertUnmarshal(
				t,
				tc.format, tc.data,
				tc.registry,
				tc.errRegex,
				func(v *GHCRRegistryDefaults) string { return v.String("") },
				tc.want,
				packageName,
				"GHCRRegistryDefaults",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: Auth should never be nil.
			if tc.registry.Auth == nil {
				t.Errorf(
					"%s\nGHCRRegistryDefaults.Unmarshal(format=%q, data=%q) - Auth should not be nil",
					packageName, tc.format, tc.data,
				)
			}
		})
	}
}

func TestGHCRRegistry_Unmarshal(t *testing.T) {
	// GIVEN: a GHCRRegistry and JSON/YAML to unmarshal into it.
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
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
				},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
				},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
				},
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
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
				},
			},
			errRegex: test.TrimYAML(`
				^auth:
					[^\s]+ sequence was used where mapping is expected`,
			),
		},
		{
			name:     "auth: null",
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
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
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
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, testErr := test.AssertUnmarshal(
				t,
				tc.format, tc.data,
				tc.registry,
				tc.errRegex,
				func(v *GHCRRegistry) string { return v.String("") },
				tc.want,
				packageName,
				"GHCRRegistry",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: Auth should never be nil.
			if tc.registry.Auth == nil {
				t.Errorf(
					"%s\nGHCRRegistry.Unmarshal(format=%q, data=%q) - Auth should not be nil",
					packageName, tc.format, tc.data,
				)
			}
		})
	}
}

func TestGHCRRegistry_ApplyOverrides(t *testing.T) {
	// GIVEN: a GHCRRegistry and JSON/YAML to decode into it.
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
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
				},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
				},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
				},
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
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
				},
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
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
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
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{},
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
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token: "abc",
						},
					},
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
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token: "abc",
						},
					},
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

			// WHEN: it is decoded into itself.
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
				packageName,
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
	// GIVEN: a GHCRRegistryDefaults.
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
			name: "only token",
			registry: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &GHCRAuthDefaults{
						Token: "foo",
					},
				},
			},
			want: false,
		},
		{
			name: "only query-token/valid-until",
			registry: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &GHCRAuthDefaults{
						queryToken: "bar",
						validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: true,
		},
		{
			name: "only image",
			registry: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
				},
			},
			want: false,
		},
		{
			name: "only tag",
			registry: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					ContainerDetail: ContainerDetail{
						Tag: "t",
					},
				},
			},
			want: false,
		},
		{
			name: "only image/tag",
			registry: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					ContainerDetail: ContainerDetail{
						Image: "i",
						Tag:   "t",
					},
				},
			},
			want: false,
		},
		{
			name: "full",
			registry: &GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					ContainerDetail: ContainerDetail{
						Image: "i",
						Tag:   "t",
					},
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token:      "foo",
							queryToken: "bar",
							validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
						},
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
					"%s\nGHCRRegistryDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestGHCRRegistry_IsZero(t *testing.T) {
	// GIVEN: a GHCRRegistry.
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
			name: "non-empty - type",
			data: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					Type: "abc",
					Auth: RegistryMap["ghcr"]().GetAuth(),
				},
			},
			want: false,
		},
		{
			name: "non-empty - CommonRegistry",
			data: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
					Auth: RegistryMap["ghcr"]().GetAuth(),
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.data.IsZero()

			// THEN: the expected result is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nGHCRRegistry.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestGHCRRegistry_Copy(t *testing.T) {
	// GIVEN: a GHCRRegistry.
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
				CommonRegistry: CommonRegistry{
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token:    "t1",
							defaults: &GHCRAuthDefaults{},
						},
					},
				},
			},
			want: test.TrimYAML(`
				auth:
					token: t1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy() is called on it.
			gotInterface := tc.registry.Copy()

			prefix := fmt.Sprintf("%s\nGHCRRegistry.Copy()", packageName)

			// THEN: the returned Registry unmarshals the same.
			if got := decode.ToYAMLString(gotInterface, ""); got != tc.want {
				t.Fatalf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the returned Registry is a GHCRRegistry.
			got, ok := gotInterface.(*GHCRRegistry)
			if !ok {
				if gotInterface == nil {
					return
				}
				t.Fatalf(
					"%s returned wrong type: %T",
					prefix, gotInterface,
				)
			}

			// AND: the unmarshaled values are the same.
			gotAuth, ok := got.GetAuth().(*GHCRAuth)
			hadAuth, _ := tc.registry.GetAuth().(*GHCRAuth)
			if !ok ||
				gotAuth.queryToken != hadAuth.queryToken ||
				gotAuth.validUntil != hadAuth.validUntil ||
				got.defaults != tc.registry.defaults {
				t.Fatalf(
					"%s mismatch\ngot:  %+v\nwant: %s",
					prefix, got, tc.want,
				)
			}

			// AND: the returned GHCRAuth is at a different address.
			if gotAuth == hadAuth {
				t.Fatalf(
					"%s returned pointer to same Auth for instance %q",
					prefix, tc.name,
				)
			}
		})
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

func TestGHCRRegistryDefaults_String(t *testing.T) {
	// GIVEN: a GHCRRegistryDefaults.
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					ContainerDetail: ContainerDetail{
						Image: "i1",
						Tag:   "t1",
						Defaults: &ContainerDetail{
							Image: "i2",
							Tag:   "t2",
						},
					},
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token: "token1",
							defaults: &GHCRAuthDefaults{
								Token: "token2",
							},
						},
					},
					defaults: &GHCRRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							ContainerDetail: ContainerDetail{
								Image: "i2",
								Tag:   "t2",
								Defaults: &ContainerDetail{
									Image: "i3",
									Tag:   "t3",
								},
							},
							Auth: &GHCRAuth{
								GHCRAuthDefaults: GHCRAuthDefaults{
									Token: "token2",
									defaults: &GHCRAuthDefaults{
										Token: "token3",
									},
								},
							},
						},
					},
				},
			},
			want: test.TrimYAML(`
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
				packageName,
				tc.data.String,
				tc.want,
			)
		})
	}
}

func TestGHCRRegistry_String(t *testing.T) {
	// GIVEN: a GHCRRegistry.
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
				CommonRegistry: CommonRegistry{
					Type: "test-ghcr",
					ContainerDetail: ContainerDetail{
						Image: "i1",
						Tag:   "t1",
						Defaults: &ContainerDetail{
							Image: "i2",
							Tag:   "t2",
						},
					},
					Auth: &GHCRAuth{
						GHCRAuthDefaults: GHCRAuthDefaults{
							Token: "token1",
							defaults: &GHCRAuthDefaults{
								Token: "token2",
							},
						},
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
				packageName,
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
	// GIVEN: a GHCRRegistryDefaults.
	var registry GHCRRegistryDefaults

	// WHEN: GetType is called on it.
	got := registry.GetType()

	// THEN: the type is returned.
	if want := "ghcr"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGHCRRegistry_GetType(t *testing.T) {
	// GIVEN: a GHCRRegistry.
	var registry GHCRRegistry

	// WHEN: GetType is called on it.
	got := registry.GetType()

	// THEN: the type is returned.
	if want := "ghcr"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestGHCRRegistry_NewRequest(t *testing.T) {
	// GIVEN: a GHCRRegistry, and a tag.
	tests := []struct {
		name     string
		registry *GHCRRegistry
		tag      string
		errRegex string
	}{
		{
			name: "no image/tag",
			registry: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "",
					},
				},
			},
			errRegex: `^$`,
		},
		{
			name: "have image+tag",
			registry: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "123",
						Tag:   "not-used",
					},
				},
			},
			tag:      "foo",
			errRegex: `^$`,
		},
		{
			name: "tag: invalid",
			registry: &GHCRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "123",
						Tag:   "not-used",
					},
				},
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
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "	123",
						Tag:   "not-used",
					},
				},
			},
			tag: "foo",
			errRegex: test.TrimYAML(`
				^parse "https://.*
					[^\s]+ invalid control character in URL$`,
			),
		},
	}

	// AND: Headers to verify being set.
	headers := map[string]string{
		"Accept": "application/vnd.oci.image.index.v1+json",
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: getQueryURL() is called on it.
			req, err := tc.registry.newRequest(tc.tag)

			prefix := fmt.Sprintf(
				"%s\nGHCRRegistry.newRequest(%q)",
				packageName, tc.tag,
			)

			// THEN: The error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if err != nil {
				return
			}

			// AND: The headers are set.
			for key, value := range headers {
				if got := req.Header.Get(key); got != value {
					t.Errorf(
						"%s .Header[%q] mismatch\ngot:  %q\nwant: %q",
						prefix, key,
						got, value,
					)
				}
			}
		})
	}
}

// #######################
// # REGISTRY | DEFAULTS #
// #######################

func TestGHCRRegistryDefaults_Defaults(t *testing.T) {
	// GIVEN: a GHCRRegistryDefaults with defaults.
	tests := []struct {
		name     string
		registry GHCRRegistryDefaults
		wantNil  bool
	}{
		{
			name:     "nil defaults",
			registry: GHCRRegistryDefaults{},
			wantNil:  true,
		},
		{
			name: "GHCRRegistryDefaults",
			registry: GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					defaults: &GHCRRegistryDefaults{},
				},
			},
			wantNil: false,
		},
		{
			name: "HubRegistryDefaults",
			registry: GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					defaults: &HubRegistryDefaults{},
				},
			},
			wantNil: false,
		},
		{
			name: "QuayRegistryDefaults",
			registry: GHCRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					defaults: &QuayRegistryDefaults{},
				},
			},
			wantNil: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Defaults is called on it.
			got := tc.registry.Defaults()

			// THEN: a pointer to the defaults is returned.
			if got == nil != tc.wantNil {
				want := "nil"
				got := "nil"
				if tc.wantNil {
					want = "non-nil"
				} else {
					got = "non-nil"
				}
				t.Errorf(
					"%s\nGHCRRegistryDefaults.Defaults() mismatch\ngot:  %s\nwant: %s",
					packageName, got, want,
				)
			}
		})
	}
}

// ###################
// # AUTH | DECODING #
// ###################

func TestGHCRAuthDefaults_Unmarshal(t *testing.T) {
	// GIVEN: a GHCRAuthDefaults.
	tests := []struct {
		name         string
		format, data string
		want         *GHCRAuthDefaults
		errRegex     string
	}{
		{
			name:     "JSONL empty",
			format:   "json",
			data:     "{}",
			errRegex: `^$`,
			want:     &GHCRAuthDefaults{},
		},
		{
			name:     "YAML empty",
			format:   "yaml",
			data:     "{}\n",
			errRegex: `^$`,
			want:     &GHCRAuthDefaults{},
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "invalid",
			want:     nil,
			errRegex: `invalid character`,
		},
		{
			name:   "YAML/invalid",
			format: "yaml",
			data:   "invalid",
			want:   nil,
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v GHCRAuthDefaults
			if err, testErr := test.AssertUnmarshal(
				t,
				tc.format, tc.data,
				&v,
				tc.errRegex,
				func(v *GHCRAuthDefaults) string { return v.String("") },
				func() string { return decode.ToYAMLString(tc.want, "") }(),
				packageName,
				"GHCRAuthDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			} else if err != nil {
				return
			}

			// AND: The queryToken is encoded only if not already.
			if v.queryToken != tc.want.queryToken {
				t.Errorf(
					"%s\nGHCRAuthDefaults.Unmarshal(format=%q, data=%q) .queryToken mismatch:\ngot:  %q\nwant: %q",
					packageName, tc.format, tc.data,
					v.queryToken, tc.want.queryToken,
				)
			}
		})
	}
}

// ################
// # AUTH | STATE #
// ################

func TestGHCRAuthDefaults_IsZero(t *testing.T) {
	// GIVEN: a GHCRAuthDefaults.
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
			name: "non-empty",
			data: &GHCRAuthDefaults{Token: "t1"},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.data.IsZero()

			// THEN: the expected result is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nGHCRAuthDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestGHCRAuth_Copy(t *testing.T) {
	// GIVEN: a GHCRAuth.
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
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token:      "t1",
					queryToken: "qT",
					validUntil: time.Now(),
					defaults:   &GHCRAuthDefaults{},
				},
			},
			want: "token: t1\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy() is called on it.
			authCopy := tc.auth.Copy()

			prefix := fmt.Sprintf("%s\nGHCRAuth.Copy()", packageName)

			// THEN: the returned RegistryAuth unmarshals the same.
			if got := decode.ToYAMLString(authCopy, ""); got != tc.want {
				t.Fatalf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the returned RegistryAuth is a GHCRAuth.
			got, ok := authCopy.(*GHCRAuth)
			if !ok {
				if authCopy == nil {
					return
				}
				t.Fatalf(
					"%s returned wrong type: %T",
					prefix, authCopy,
				)
			}
			if tc.auth == nil {
				return
			}

			// AND: the unmarshaled values are the same.
			if got.queryToken != tc.auth.queryToken ||
				got.validUntil != tc.auth.validUntil ||
				got.defaults != tc.auth.defaults {
				t.Fatalf(
					"%s mismatch\ngot:  %+v\nwant: %+v",
					prefix, got, tc.auth,
				)
			}

			// AND: the returned GHCRAuth is at a different address.
			if got == tc.auth {
				t.Fatalf(
					"%s returned pointer to the same Auth for instance %q",
					prefix, tc.name,
				)
			}
		})
	}
}

// ####################
// # AUTH | STRINGIFY #
// ####################

func TestGHCRAuthDefaults_String(t *testing.T) {
	// GIVEN: a GHCRAuthDefaults.
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
				packageName,
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
	// GIVEN: a GHCRAuthDefaults.
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

			// WHEN: Defaults is called on it.
			got := tc.data.Defaults()

			// THEN: Defaults are returned when expected.
			if gotDefaults := got != nil; gotDefaults != tc.haveDefaults {
				t.Errorf(
					"%s\nGHCRAuthDefaults.Defaults() mismatch\ngot:  %t\nwant: %t",
					packageName, gotDefaults, tc.haveDefaults,
				)
			}
		})
	}
}

func TestGHCRAuthDefaults_SetDefaults(t *testing.T) {
	// GIVEN: a RegistryAuth.
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

			// AND: and GHCRAuthDefaults to take them.
			data := &GHCRAuthDefaults{}

			// WHEN: SetDefaults() is called to give these defaults.
			data.SetDefaults(tc.newDefaults)

			// THEN: they are set when expected.
			if got := data.defaults == tc.newDefaults; got != tc.doesSet {
				t.Errorf(
					"%s\nGHCRAuthDefaults.SetDefaults() .defaults mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.doesSet,
				)
			}
		})
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

func TestGHCRAuth_CheckValues(t *testing.T) {
	// GIVEN: a GHCRAuth.
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
			name: "valid",
			input: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "token",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "token",
			input: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "token",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "no token",
			input: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "token",
				},
			},
			errRegex: `^$`,
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

// ######################
// # AUTH | CREDENTIALS #
// ######################

func TestGHCRAuth_GetToken(t *testing.T) {
	// GIVEN: a GHCRAuth with/without defaults.
	tests := []struct {
		name string
		data *GHCRAuth
		want string
	}{
		{
			name: "empty",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{},
			},
			want: "",
		},
		{
			name: "no defaults",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "root",
				},
			},
			want: "root",
		},
		{
			name: "defaults fallback",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "",
					defaults: &GHCRAuthDefaults{
						Token: "defaults",
					},
				},
			},
			want: "defaults",
		},
		{
			name: "defaults fallback recursive",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "",
					defaults: &GHCRAuthDefaults{
						Token: "",
						defaults: &GHCRAuthDefaults{
							Token: "hard-defaults",
						},
					},
				},
			},
			want: "hard-defaults",
		},
		{
			name: "root Token prioritised",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "root",
					defaults: &GHCRAuthDefaults{
						Token: "defaults",
					},
				},
			},
			want: "root",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetToken() is called on it.
			got := tc.data.GetToken()

			// THEN: the expected result is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nGHCRAuth.GetToken() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}

			// WHEN: GetTokenSelf() is called on it.
			got = tc.data.GetTokenSelf()

			// THEN: the expected result is returned.
			if got != tc.data.Token {
				t.Fatalf(
					"%s\nGHCRAuth.GetTokenSelf() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.data.Token,
				)
			}
		})
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestGHCRAuthDefaults_GetQueryTokenSelf(t *testing.T) {
	// GIVEN: a GHCRAuthDefaults.
	tests := []struct {
		name string
		data *GHCRAuthDefaults
		want string
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
			want: "query-token", // auto-renews.
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			usable := isUsable(tc.data.queryToken, tc.data.validUntil)

			// WHEN: GetQueryTokenSelf() is called on it.
			queryToken, validUntil := tc.data.GetQueryTokenSelf()

			prefix := fmt.Sprintf("%s\nGHCRAuthDefaults.GetQueryTokenSelf()", packageName)

			// THEN: the query token is returned as expected.
			if queryToken != tc.want {
				t.Errorf(
					"%s queryToken mismatch\ngot:  %q\nwant: %q",
					prefix, queryToken, tc.want,
				)
			}

			// AND: the validUntil is returned as before if it was usable.
			if !validUntil.Equal(tc.data.validUntil) && usable {
				t.Errorf(
					"%s validUntil mismatch\ngot:  %q\nwant: %q",
					prefix, validUntil, tc.data.validUntil,
				)
			}
		})
	}
}

func TestGHCRAuthDefaults_GetQueryTokenSelf__Parallel(t *testing.T) {
	// GIVEN: a GHCRAuthDefaults with a queryToken that has expired.
	data := &GHCRAuthDefaults{
		Token:      "token",
		queryToken: "query-token",
		validUntil: time.Now().Add(-10 * time.Second),
	}

	// AND: the mutex is held.
	data.mu.Lock()
	go func() {
		time.Sleep(time.Second)
		data.mu.Unlock()
	}()

	// WHEN: GetQueryTokenSelf() is queued on it by many goroutines.
	for range 10 {
		go func() { _, _ = data.GetQueryTokenSelf() }()
	}
	// Call again to verify we still get a queryToken.
	queryToken, validUntil := data.GetQueryTokenSelf()

	prefix := fmt.Sprintf("%s\nGHCRAuthDefaults.GetQueryTokenSelf()", packageName)

	// THEN: the query token is returned as expected.
	if queryToken != data.queryToken {
		t.Errorf(
			"%s queryToken mismatch\ngot:  %q\nwant: %q",
			prefix, queryToken, data.queryToken,
		)
	}

	// AND: the validUntil is returned as before if it was usable.
	if !validUntil.Equal(data.validUntil) {
		t.Errorf(
			"%s validUntil mismatch\ngot:  %q\nwant: %q",
			prefix, validUntil, data.validUntil,
		)
	}
}

func TestGHCRAuth_GetQueryToken(t *testing.T) {
	// GIVEN: a GHCRAuth.
	tests := []struct {
		name     string
		data     *GHCRAuth
		want     string
		errRegex string
	}{
		{
			name: "valid",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token:      "token",
					queryToken: "query-token",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "expired root refreshed before defaults fallback",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token:      "token",
					queryToken: "query-token",
					validUntil: time.Now().Add(-10 * time.Second),
					defaults: &GHCRAuthDefaults{
						Token:      "token",
						queryToken: "default-query-token",
						validUntil: time.Now().Add(10 * time.Second),
					},
				},
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "expired root refreshed",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token:      "token",
					queryToken: "query-token",
					validUntil: time.Now().Add(-10 * time.Second),
				},
			},
			want:     "query-token",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: an image:tag to query on.
			detail := ContainerDetail{
				Image: test.ArgusDockerGHCRRepo,
				Tag:   "v1.2.3",
			}

			// WHEN: GetQueryToken() is called on it.
			queryToken, err := tc.data.GetQueryToken(detail)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nGHCRAuth.GetQueryToken(%+v) error mismatch:\ngot:  %q\nwant: %q",
					packageName, detail,
					e, tc.errRegex,
				)
			}
			if err != nil {
				return
			}

			// AND: the query token is returned as expected.
			if queryToken != tc.want {
				t.Errorf(
					"%s\nGHCRAuth.GetQueryTokenSelf() mismatch on queryToken\ngot:  %q\nwant: %q",
					packageName, queryToken, tc.want,
				)
			}
		})
	}
}

func TestGHCRAuthDefaults_SetQueryToken(t *testing.T) {
	// GIVEN: a GHCRAuthDefaults.
	tests := []struct {
		name string
		data *GHCRAuth
	}{
		{
			name: "root only",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token:      "token",
					queryToken: "query-token",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
		},
		{
			name: "defaults/hardDefaults ignored",
			data: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
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
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			queryToken := "new-query-token"
			validUntil := time.Now().Add(10 * time.Second)

			// WHEN: SetQueryToken() is called on it.
			tc.data.SetQueryToken(queryToken, validUntil)

			check := func(name string, d *GHCRAuthDefaults, wantMatch bool) {
				if d == nil {
					return
				}

				prefix := fmt.Sprintf(
					"%s\nGHCRAuthDefaults.SetQueryToken(queryToken=%q, validUntil=%q) (%s - wantMatch=%t)",
					packageName, queryToken, validUntil, name, wantMatch,
				)

				// THEN: the queryToken is set only when expected.
				if (d.queryToken == queryToken) != wantMatch {
					t.Errorf(
						"%s .queryToken mismatch\ngot:  %q\nwant: %q",
						prefix, queryToken, d.queryToken,
					)
				}

				// AND: the validUntil is set only when expected.
				if d.validUntil.Equal(validUntil) != wantMatch {
					t.Errorf(
						"%s .validUntil mismatch\ngot:  %q\nwant: %q",
						prefix, validUntil, d.validUntil,
					)
				}
			}

			check("root", &tc.data.GHCRAuthDefaults, true)
			check("defaults", tc.data.defaults, false)
			if tc.data.defaults != nil {
				check("hardDefaults", tc.data.defaults.defaults, false)
			}
		})
	}
}

func TestGHCRAuth_RefreshQueryToken__cached(t *testing.T) {
	// GIVEN: a GHCRAuth with a cached query token that is valid for a while.
	auth := &GHCRAuth{
		GHCRAuthDefaults: GHCRAuthDefaults{
			queryToken: "cached-token",
			validUntil: time.Now().Add(time.Hour),
		},
	}
	detail := ContainerDetail{Image: test.ArgusDockerGHCRRepo}

	// WHEN: refreshQueryToken is called on it.
	queryToken, err := auth.refreshQueryToken(detail)

	prefix := fmt.Sprintf(
		"%s\nGHCRAuth.refreshQueryToken(%+v)",
		packageName, detail,
	)

	// THEN: the cached token is returned without error.
	e := errfmt.FormatError(err)
	if !util.RegexCheck(`^$`, e) {
		t.Fatalf("%s error mismatch\ngot:  %q\nwant: %q", prefix, e, `^$`)
	}
	if got, want := queryToken, "cached-token"; got != want {
		t.Fatalf(
			"%s queryToken mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}
}

// ######################
// # AUTH | INHERITANCE #
// ######################

func TestGHCRAuth_Inherit(t *testing.T) {
	// GIVEN: a GHCRAuth, and a RegistryAuth to try and inherit from.
	tests := []struct {
		name                 string
		auth                 *GHCRAuth
		from                 RegistryAuth
		srcDetail, dstDetail ContainerDetail
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
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: util.SecretValue,
				},
			},
			from: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token:      "abc",
					queryToken: "qt",
					validUntil: time.Now(),
				},
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
			name: "inherit from GHCRAuth when Details do not match (src.Token is SecretValue)",
			auth: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: util.SecretValue,
				},
			},
			from: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token:      "abc",
					queryToken: "qt",
					validUntil: time.Now(),
				},
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
			name: "do not inherit from GHCRAuth when src.Token is not SecretValue",
			auth: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "foo",
				},
			},
			from: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token:      "abc",
					queryToken: "qt",
					validUntil: time.Now(),
				},
			},
		},
		{
			name: "do not inherit from HubAuth",
			auth: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: util.SecretValue,
				},
			},
			from: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token: "abc",
				},
			},
		},
		{
			name: "do not inherit from QuayAuth",
			auth: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: util.SecretValue,
				},
			},
			from: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "abc",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hadToken, hadQueryToken, hadValidUntil := getTokenData(t, tc.auth)

			// WHEN: Inherit is called on it.
			tc.auth.Inherit(tc.from, tc.srcDetail, tc.dstDetail)

			prefix := fmt.Sprintf(
				"%s\nGHCRAuth.Inherit(from=%T, srcDetail=%+v, dstDetail=%+v)",
				packageName, tc.from, tc.srcDetail, tc.dstDetail,
			)

			// THEN: the expected inheritance has occurred.
			gotToken, gotQueryToken, gotValidUntil := getTokenData(t, tc.auth)
			wantToken, wantQueryToken, wantValidUntil := getTokenData(t, tc.from)
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
