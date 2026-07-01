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
	// GIVEN: a HubRegistryDefaults and JSON/YAML to unmarshal into it.
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{},
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
			registry: &HubRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{},
				},
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{},
				},
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{},
				},
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
				packageName,
				"HubRegistryDefaults",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: Auth should never be nil.
			if tc.registry.Auth == nil {
				t.Errorf(
					"%s\nHubRegistryDefaults.Unmarshal(format=%q, data=%q) - Auth should not be nil",
					packageName, tc.format, tc.data,
				)
			}
		})
	}
}

func TestHubRegistry_Unmarshal(t *testing.T) {
	// GIVEN: a HubRegistry and JSON/YAML to unmarshal into it.
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
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
				},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
				},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
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
			registry: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
				},
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
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
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
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
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
				packageName,
				"HubRegistry",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: Auth should never be nil.
			if tc.registry.Auth == nil {
				t.Errorf(
					"%s\nHubRegistry.Unmarshal(format=%q, data=%q) - Auth should not be nil",
					packageName, tc.format, tc.data,
				)
			}
		})
	}
}

func TestHubRegistry_ApplyOverrides(t *testing.T) {
	// GIVEN: a HubRegistry and JSON/YAML to decode into it.
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
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
				},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
				},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
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
			registry: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
				},
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
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
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
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{},
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
				CommonRegistry: CommonRegistry{
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
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
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
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
					username: hub-username
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
				func(format string, data []byte, v *HubRegistry) (*HubRegistry, error) {
					err := v.ApplyOverrides(format, data)
					return v, err
				},
				tc.format, tc.data,
				func(v *HubRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				true,
				packageName,
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
	// GIVEN: a HubRegistryDefaults.
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{
						Username: "u",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Token",
			registry: &HubRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{
						Token: "foo",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/queryToken and validUntil",
			registry: &HubRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{
						queryToken: "bar",
						validUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: true,
		},
		{
			name: "non-empty/all",
			registry: &HubRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuthDefaults{
						Username:   "u",
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
					"%s\nHubRegistryDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestHubRegistry_IsZero(t *testing.T) {
	// GIVEN: a HubRegistry.
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
				CommonRegistry: CommonRegistry{
					Type: "abc",
					Auth: RegistryMap["hub"]().GetAuth(),
				},
			},
			want: false,
		},
		{
			name: "non-empty/CommonRegistry",
			data: &HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
					Auth: RegistryMap["hub"]().GetAuth(),
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
					Type: "abc",
					Auth: RegistryMap["hub"]().GetAuth(),
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
					"%s\nHubRegistry IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestHubRegistry_Copy(t *testing.T) {
	// GIVEN: a HubRegistry.
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
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i1",
						Tag:   "t1",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username: "u1",
							Token:    "t1",
							defaults: &HubAuthDefaults{},
						},
					},
				},
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

			// WHEN: Copy() is called on it.
			gotInterface := tc.registry.Copy()

			prefix := fmt.Sprintf("%s\nHubRegistry.Copy()", packageName)

			// THEN: the returned Registry unmarshals the same.
			if g := decode.ToYAMLString(gotInterface, ""); g != tc.want {
				t.Fatalf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, g, tc.want,
				)
			}

			// AND: the returned Registry is a HubRegistry.
			got, ok := gotInterface.(*HubRegistry)
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
			gotAuth, ok := got.GetAuth().(*HubAuth)
			hadAuth, _ := tc.registry.GetAuth().(*HubAuth)
			if !ok ||
				gotAuth.queryToken != hadAuth.queryToken ||
				gotAuth.validUntil != hadAuth.validUntil ||
				got.defaults != tc.registry.defaults {
				t.Fatalf(
					"%s mismatch\ngot:  %+v\nwant: %+v",
					prefix, got, tc.want,
				)
			}

			// AND: the returned HubAuth is at a different address.
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

func TestHubRegistryDefaults_String(t *testing.T) {
	// GIVEN: a HubRegistryDefaults.
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username: "u1",
							Token:    "token1",
							defaults: &HubAuthDefaults{
								Username: "u2",
								Token:    "token2",
							},
						},
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
				packageName,
				tc.data.String,
				tc.want,
			)
		})
	}
}

func TestHubRegistry_String(t *testing.T) {
	// GIVEN: a HubRegistry.
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
				CommonRegistry: CommonRegistry{
					Type: "test-hub",
					ContainerDetail: ContainerDetail{
						Image: "i1",
						Tag:   "t1",
						Defaults: &ContainerDetailDefaults{
							Tag: "t2",
						},
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username: "u1",
							Token:    "token1",
							defaults: &HubAuthDefaults{
								Username: "u2",
								Token:    "token2",
							},
						},
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

func TestHubRegistryDefaults_GetType(t *testing.T) {
	// GIVEN: a HubRegistryDefaults.
	var registry HubRegistryDefaults

	// WHEN: GetType is called on it.
	got := registry.GetType()

	// THEN: the type is returned.
	if want := "hub"; got != want {
		t.Errorf(
			"%s\ngot %q, want %q",
			packageName, got, want,
		)
	}
}

func TestHubRegistry_GetType(t *testing.T) {
	// GIVEN: a HubRegistry.
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
				CommonRegistry: CommonRegistry{
					Type: "hi",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetType is called on it.
			got := tc.registry.GetType()

			// THEN: the type is returned.
			if want := "hub"; got != want {
				t.Errorf(
					"%s\ngot %q, want %q",
					packageName, got, want,
				)
			}
		})
	}
}

// #########################
// # REGISTRY | VALIDATION #
// #########################

func TestHubRegistry_CheckValues(t *testing.T) {
	// GIVEN: a HubRegistry.
	tests := []struct {
		name     string
		input    *HubRegistry
		errRegex string
	}{
		{
			name: "valid",
			input: &HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
						Tag:   "t",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username: "username",
							Token:    "token",
						},
					},
				},
			},
			errRegex: `^$`,
		},
		{
			name: "CommonRegistry: no image",
			input: &HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Tag: "t",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username: "username",
							Token:    "t",
						},
					},
				},
			},
			errRegex: `^image: <required> \([^\)]+\)$`,
		},
		{
			name: "Auth: no token",
			input: &HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
						Tag:   "t",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username: "username",
						},
					},
				},
			},
			errRegex: `^token: <required> \([^\)]+\)$`,
		},
		{
			name: "Auth: no username",
			input: &HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
						Tag:   "t",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Token: "token",
						},
					},
				},
			},
			errRegex: `^username: <required> \([^\)]+\)$`,
		},
		{
			name: "CommonRegistry err and Auth err",
			input: &HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Tag: "t",
					},
					Auth: &HubAuth{
						HubAuthDefaults: HubAuthDefaults{
							Username: "username",
						},
					},
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

func TestHubRegistry_NewRequest(t *testing.T) {
	// GIVEN: a HubRegistry, and a tag.
	tests := []struct {
		name     string
		registry *HubRegistry
		tag      string
		errRegex string
	}{
		{
			name: "no image or tag",
			registry: &HubRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "",
						Tag:   "",
					},
				},
			},
			errRegex: `^$`,
		},
		{
			name: "have image+tag",
			registry: &HubRegistry{
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
			registry: &HubRegistry{
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
					.*invalid control character in URL$`,
			),
		},
		{
			name: "image: invalid",
			registry: &HubRegistry{
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
					.*invalid control character in URL$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: newRequest() is called on it.
			_, err := tc.registry.newRequest(tc.tag)

			prefix := fmt.Sprintf(
				"%s\nHubRegistry.newRequest(%s)",
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
		})
	}
}

// ################
// # AUTH | STATE #
// ################

func TestHubAuthDefaults_IsZero(t *testing.T) {
	// GIVEN: a HubAuthDefaults.
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

			// WHEN: IsZero() is called on it.
			got := tc.data.IsZero()

			// THEN: the expected result is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nHubAuthDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestHubAuth_Copy(t *testing.T) {
	// GIVEN: a HubAuth.
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
				HubAuthDefaults: HubAuthDefaults{
					Username:   "u1",
					Token:      "t1",
					queryToken: "qT",
					validUntil: time.Now(),
					defaults:   &HubAuthDefaults{},
				},
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

			// WHEN: Copy() is called on it.
			authCopy := tc.auth.Copy()

			prefix := fmt.Sprintf("%s\nHubAuth.Copy()", packageName)

			// THEN: the returned RegistryAuth unmarshals the same.
			if got := decode.ToYAMLString(authCopy, ""); got != tc.want {
				t.Fatalf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the returned RegistryAuth is a HubAuth.
			got, ok := authCopy.(*HubAuth)
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

func TestHubAuthDefaults_String(t *testing.T) {
	// GIVEN: a HubAuthDefaults.
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

func TestHubAuthDefaults_Defaults(t *testing.T) {
	// GIVEN: a HubAuthDefaults.
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

			// WHEN: Defaults is called on it.
			got := tc.data.Defaults()

			// THEN: Defaults are returned when expected.
			if gotDefaults := got != nil; gotDefaults != tc.haveDefaults {
				t.Errorf(
					"%s\nHubAuthDefaults.Defaults() mismatch\ngot:  %t\nwant: %t",
					packageName, gotDefaults, tc.haveDefaults,
				)
			}
		})
	}
}

func TestHubAuthDefaults_SetDefaults(t *testing.T) {
	// GIVEN: a RegistryAuthDefaults.
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

			// AND: and HubAuthDefaults to take them.
			data := &HubAuthDefaults{}

			// WHEN: SetDefaults() is called to give these defaults.
			data.SetDefaults(tc.newDefaults)

			// THEN: they are set when expected.
			if got := data.defaults == tc.newDefaults; got != tc.doesSet {
				t.Errorf(
					"%s\nHubAuthDefaults.SetDefaults() .defaults mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.doesSet,
				)
			}
		})
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

func TestHubAuth_CheckValues(t *testing.T) {
	// GIVEN: a HubAuth.
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
				HubAuthDefaults: HubAuthDefaults{
					Username: "username",
					Token:    "token",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "username, no token",
			input: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username: "username",
				},
			},
			errRegex: `^token: <required> \([^\)]+\)$`,
		},
		{
			name: "token, no username",
			input: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token: "token",
				},
			},
			errRegex: `^username: <required> \([^\)]+\)$`,
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

func TestHubAuth_GetUsernameSelf(t *testing.T) {
	// GIVEN: a HubAuth.
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
				HubAuthDefaults: HubAuthDefaults{
					Username: "u",
				},
			},
			want: "u",
		},
		{
			name: "env vars",
			auth: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username: "u-${DOCKER_USERNAME}",
				},
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

			// AND: env vars have been set.
			test.SetEnv(t, tc.env)

			// WHEN: GetUsernameSelf is called.
			got := tc.auth.GetUsernameSelf()

			// THEN: the username returned is correct.
			if got != tc.want {
				t.Errorf(
					"%s\nHubAuthDefaults.GetUsernameSelf() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestHubAuth_GetUsername(t *testing.T) {
	// GIVEN: a HubAuth with/without defaults.
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
				HubAuthDefaults: HubAuthDefaults{
					Username: "root",
				},
			},
			want: "root",
		},
		{
			name: "defaults fallback",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username: "",
					defaults: &HubAuthDefaults{
						Username: "defaults",
					},
				},
			},
			want: "defaults",
		},
		{
			name: "defaults fallback recursive",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username: "",
					defaults: &HubAuthDefaults{
						Username: "hard-defaults",
					},
				},
			},
			want: "hard-defaults",
		},
		{
			name: "root Username prioritised",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username: "root",
					defaults: &HubAuthDefaults{
						Username: "defaults",
					},
				},
			},
			want: "root",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetUsername() is called on it.
			got := tc.data.GetUsername()

			// THEN: the expected result is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nHubAuth.GetUsername() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestHubAuth_GetToken(t *testing.T) {
	// GIVEN: a HubAuth with/without defaults.
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
				HubAuthDefaults: HubAuthDefaults{
					Token: "root",
				},
			},
			want: "root",
		},
		{
			name: "defaults fallback",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token: "",
					defaults: &HubAuthDefaults{
						Token: "defaults",
					},
				},
			},
			want: "defaults",
		},
		{
			name: "defaults fallback recursive",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token: "",
					defaults: &HubAuthDefaults{
						Token: "hard-defaults",
					},
				},
			},
			want: "hard-defaults",
		},
		{
			name: "root Token prioritised",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token: "root",
					defaults: &HubAuthDefaults{
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
					"%s\nHubAuth.GetToken() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}

			// WHEN: GetTokenSelf() is called on it.
			got = tc.data.GetTokenSelf()

			// THEN: the expected result is returned.
			if got != tc.data.Token {
				t.Fatalf(
					"%s\nHubAuth.GetTokenSelf() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.data.Token,
				)
			}
		})
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestHubAuthDefaults_GetQueryTokenSelf(t *testing.T) {
	// GIVEN: a HubAuthDefaults.
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			usable := isUsable(tc.data.queryToken, tc.data.validUntil)

			// WHEN: GetQueryTokenSelf() is called on it.
			queryToken, validUntil := tc.data.GetQueryTokenSelf()

			prefix := fmt.Sprintf("%s\nHubAuthDefaults.GetQueryTokenSelf()", packageName)

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
func TestHubAuth_GetQueryToken(t *testing.T) {
	// GIVEN: a HubAuth.
	tests := []struct {
		name     string
		data     *HubAuth
		want     string
		errRegex string
	}{
		{
			name: "no query token",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token: "token",
				},
			},
			want:     "",
			errRegex: `^$`,
		},
		{
			name: "valid at root",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token:      "token",
					queryToken: "query-token",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
			want:     "query-token",
			errRegex: `^$`,
		},
		{
			name: "valid at defaults",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token:      "token",
					queryToken: "query-token-root",
					validUntil: time.Now().Add(-10 * time.Second),
					defaults: &HubAuthDefaults{
						Token:      "token",
						queryToken: "query-token-defaults",
						validUntil: time.Now().Add(10 * time.Second),
					},
				},
			},
			want:     "query-token-defaults",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: an image:tag to query on.
			detail := ContainerDetail{
				Image: test.ArgusDockerHubRepo,
				Tag:   "v1.2.3",
			}

			// WHEN: GetQueryToken() is called on it.
			queryToken, err := tc.data.GetQueryToken(detail)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nHubAuth.GetQueryToken(%+v) error mismatch:\ngot:  %q\nwant: %q",
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
					"%s\nHubAuth.GetQueryToken(%+v) mismatch\ngot:  %q\nwant: %q",
					packageName, detail,
					queryToken, tc.want,
				)
			}
		})
	}
}

func TestHubAuthDefaults_SetQueryToken(t *testing.T) {
	// GIVEN: a HubAuthDefaults.
	tests := []struct {
		name                                               string
		data                                               *HubAuth
		setRootToken, setDefaultToken, setHardDefaultToken bool
	}{
		{
			name: "root only",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token:      "token",
					queryToken: "query-token",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
			setRootToken: true,
		},
		{
			name: "defaults - hardDefaults ignored",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
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
			setRootToken: true,
		},
		{
			name: "set in defaults",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
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
			},
			setRootToken:    true,
			setDefaultToken: true,
		},
		{
			name: "hard defaults ignored if defaults override",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
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
			setRootToken: true,
		},
		{
			name: "set in defaults and hard defaults",
			data: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
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
			},
			setRootToken:        true,
			setDefaultToken:     true,
			setHardDefaultToken: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			queryToken := "new-query-token"
			validUntil := time.Now().Add(10 * time.Second)

			// WHEN: SetQueryToken() is called on it.
			tc.data.SetQueryToken(queryToken, validUntil)

			check := func(name string, d *HubAuthDefaults, wantMatch bool) {
				if d == nil {
					return
				}

				prefix := fmt.Sprintf(
					"%s\nHubAuthDefaults.SetQueryToken(%queryToken=q, validUntil=%q) (%s - wantMatch=%t)",
					packageName, queryToken, validUntil, name, wantMatch,
				)

				// THEN: the queryToken is set only when expected.
				if (d.queryToken == queryToken) != wantMatch {
					t.Errorf(
						"%s .queryToken\ngot:  %q\nwant: %q",
						prefix, queryToken, d.queryToken,
					)
				}

				// AND: the validUntil is set only when expected.
				if d.validUntil.Equal(validUntil) != wantMatch {
					t.Errorf(
						"%s .validUntil\ngot:  %q\nwant: %q",
						prefix, validUntil, d.validUntil,
					)
				}
			}

			check("root", &tc.data.HubAuthDefaults, tc.setRootToken)
			check("defaults", tc.data.defaults, tc.setDefaultToken)
			if tc.data.defaults != nil {
				check("hardDefaults", tc.data.defaults.defaults, tc.setHardDefaultToken)
			}
		})
	}
}

// ######################
// # AUTH | INHERITANCE #
// ######################

func TestHubAuth_Inherit(t *testing.T) {
	// GIVEN: a HubAuth, and a RegistryAuth to try and inherit from.
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
				HubAuthDefaults: HubAuthDefaults{
					Username: "u",
					Token:    util.SecretValue,
				},
			},
			from: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username:   "u",
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
			name: "inherit from HubAuth when Details do not match (src.Token is SecretValue)",
			auth: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username: "u",
					Token:    util.SecretValue,
				},
			},
			from: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username:   "u",
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
			name: "do not inherit from HubAuth when src.Username is changed",
			auth: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username: "u1",
					Token:    util.SecretValue,
				},
			},
			from: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username:   "u",
					Token:      "abc",
					queryToken: "qt",
					validUntil: time.Now(),
				},
			},
		},
		{
			name: "do not inherit from HubAuth when src.Token is not SecretValue",
			auth: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token: "foo",
				},
			},
			from: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token:      "abc",
					queryToken: "qt",
					validUntil: time.Now(),
				},
			},
		},
		{
			name: "do not inherit from GHCRAuth",
			auth: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token: util.SecretValue,
				},
			},
			from: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "abc",
				},
			},
		},
		{
			name: "do not inherit from QuayAuth",
			auth: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
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
				"%s\nHubAuth.Inherit(from=%T, srcDetail=%+v, dstDetail=%+v)",
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
