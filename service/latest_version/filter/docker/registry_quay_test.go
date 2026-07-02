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

func TestQuayRegistryDefaults_Unmarshal(t *testing.T) {
	// GIVEN: a QuayRegistryDefaults and JSON/YAML to unmarshal into it.
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &QuayAuthDefaults{},
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
			registry: &QuayRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &QuayAuthDefaults{},
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &QuayAuthDefaults{},
				},
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &QuayAuthDefaults{},
				},
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
				packageName,
				"QuayRegistryDefaults",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: Auth should never be nil.
			if tc.registry.Auth == nil {
				t.Errorf(
					"%s\nQuayRegistryDefaults.Unmarshal(format=%q, data=%q) - Auth should not be nil",
					packageName, tc.format, tc.data,
				)
			}
		})
	}
}

func TestQuayRegistry_Unmarshal(t *testing.T) {
	// GIVEN: a QuayRegistry and JSON/YAML to unmarshal into it.
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
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
				},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &QuayRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
				},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &QuayRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
				},
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
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
				},
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
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
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
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
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
				packageName,
				"QuayRegistry",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: Auth should never be nil.
			if tc.registry.Auth == nil {
				t.Errorf(
					"%s\nQuayRegistry.Unmarshal(format=%q, data=%q) - Auth should not be nil",
					packageName, tc.format, tc.data,
				)
			}
		})
	}
}

func TestQuayRegistry_ApplyOverrides(t *testing.T) {
	// GIVEN: a QuayRegistry and JSON/YAML to decode into it.
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
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
				},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:   "YAML/invalid ContainerDetail",
			format: "yaml",
			data:   `image: []`,
			registry: &QuayRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
				},
			},
			errRegex: `^[^\s]+ .*unmarshal .*`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &QuayRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
				},
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
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
				},
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
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
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
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{},
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
				CommonRegistry: CommonRegistry{
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
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
					username: quay-username
			`),
			registry: &QuayRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
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
				func(format string, data []byte, v *QuayRegistry) (*QuayRegistry, error) {
					err := v.ApplyOverrides(format, data)
					return v, err
				},
				tc.format, tc.data,
				func(v *QuayRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				true,
				packageName,
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
	// GIVEN: a QuayRegistryDefaults.
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &QuayAuthDefaults{
						Token: "foo",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			registry: &QuayRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &QuayAuthDefaults{
						Token: "foo",
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
					"%s\nQuayRegistryDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestQuayRegistry_IsZero(t *testing.T) {
	// GIVEN: a QuayRegistry.
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
				CommonRegistry: CommonRegistry{
					Type: "abc",
					Auth: RegistryMap["quay"]().GetAuth(),
				},
			},
			want: false,
		},
		{
			name: "non-empty/CommonRegistry",
			data: &QuayRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
					Auth: RegistryMap["quay"]().GetAuth(),
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &QuayRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
					Type: "abc",
					Auth: RegistryMap["quay"]().GetAuth(),
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
					"%s\nQuayRegistry IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestQuayRegistry_Copy(t *testing.T) {
	// GIVEN: a QuayRegistry.
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
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i1",
						Tag:   "t1",
					},
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
							Token:    "t1",
							defaults: &QuayAuthDefaults{},
						},
					},
				},
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

			// WHEN: Copy() is called on it.
			gotInterface := tc.registry.Copy()

			prefix := fmt.Sprintf("%s\nQuayRegistry.Copy() ", packageName)

			// THEN: the returned Registry unmarshals the same.
			if got := decode.ToYAMLString(gotInterface, ""); got != tc.want {
				t.Fatalf(
					"%s mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}

			// AND: the returned Registry is a QuayRegistry.
			got, ok := gotInterface.(*QuayRegistry)
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
			gotAuth, ok := got.GetAuth().(*QuayAuth)
			hadAuth, _ := tc.registry.GetAuth().(*QuayAuth)
			if !ok || got.defaults != tc.registry.defaults {
				t.Fatalf(
					"%s .Auth mismatch\ngot:  %+v\nwant: %s",
					prefix, got, tc.want,
				)
			}

			// AND: the returned QuayAuth is at a different address.
			if gotAuth == hadAuth {
				t.Fatalf(
					"%s returned pointer to same Auth for instance %q",
					packageName, tc.name,
				)
			}
		})
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

func TestQuayRegistryDefaults_String(t *testing.T) {
	// GIVEN: a QuayRegistryDefaults.
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
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
							Token: "token1",
							defaults: &QuayAuthDefaults{
								Token: "token2",
							},
						},
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
				packageName,
				tc.data.String,
				tc.want,
			)
		})
	}
}

func TestQuayRegistry_String(t *testing.T) {
	// GIVEN: a QuayRegistry.
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
				CommonRegistry: CommonRegistry{
					Type: "test-quay",
					ContainerDetail: ContainerDetail{
						Image: "i1",
						Tag:   "t1",
						Defaults: &ContainerDetailDefaults{
							Tag: "t2",
						},
					},
					Auth: &QuayAuth{
						QuayAuthDefaults: QuayAuthDefaults{
							Token: "token1",
							defaults: &QuayAuthDefaults{
								Token: "token2",
							},
						},
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

func TestQuayRegistryDefaults_GetType(t *testing.T) {
	// GIVEN: a QuayRegistryDefaults.
	var registry QuayRegistryDefaults

	// WHEN: GetType is called on it.
	got := registry.GetType()

	// THEN: the type is returned.
	if want := "quay"; got != want {
		t.Errorf(
			"%s\ngot %q, want %q",
			packageName, got, want,
		)
	}
}
func TestQuayRegistry_GetType(t *testing.T) {
	// GIVEN: a QuayRegistry.
	var registry QuayRegistry

	// WHEN: GetType is called on it.
	got := registry.GetType()

	// THEN: the type is returned.
	if want := "quay"; got != want {
		t.Errorf(
			"%s\ngot %q, want %q",
			packageName, got, want,
		)
	}
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestQuayRegistry_NewRequest(t *testing.T) {
	// GIVEN: a QuayRegistry, and a tag.
	tests := []struct {
		name     string
		registry *QuayRegistry
		tag      string
		errRegex string
	}{
		{
			name: "no image - tag",
			registry: &QuayRegistry{
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
			registry: &QuayRegistry{
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
			registry: &QuayRegistry{
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
			registry: &QuayRegistry{
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
				"%s\nQuayRegistry newRequest(%s)",
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

func TestQuayAuthDefaults_IsZero(t *testing.T) {
	// GIVEN: a QuayAuthDefaults.
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

			// WHEN: IsZero() is called on it.
			got := tc.data.IsZero()

			// THEN: the expected result is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nQuayAuthDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestQuayAuth_Copy(t *testing.T) {
	// GIVEN: a QuayAuth.
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
				QuayAuthDefaults: QuayAuthDefaults{
					Token:    "t1",
					defaults: &QuayAuthDefaults{},
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

			prefix := fmt.Sprintf("%s\nQuayAuth.Copy()", packageName)

			// THEN: the returned RegistryAuth unmarshals the same.
			if got := decode.ToYAMLString(authCopy, ""); got != tc.want {
				t.Fatalf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the returned RegistryAuth is a QuayAuth.
			got, ok := authCopy.(*QuayAuth)
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
			if got.defaults != tc.auth.defaults {
				t.Fatalf(
					"%s mismatch\ngot:  %+v\nwant: %+v",
					prefix, got, tc.auth,
				)
			}

			// AND: the returned QuayAuth is at a different address.
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

func TestQuayAuthDefaults_String(t *testing.T) {
	// GIVEN: a QuayAuthDefaults.
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

func TestQuayAuthDefaults_Defaults(t *testing.T) {
	// GIVEN: a QuayAuthDefaults.
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

			// WHEN: Defaults is called on it.
			got := tc.data.Defaults()

			// THEN: Defaults are returned when expected.
			if gotDefaults := got != nil; gotDefaults != tc.haveDefaults {
				t.Errorf(
					"%s\nQuayAuthDefaults.Defaults() mismatch\ngot:  %t\nwant: %t",
					packageName, gotDefaults, tc.haveDefaults,
				)
			}
		})
	}
}

func TestQuayAuthDefaults_SetDefaults(t *testing.T) {
	// GIVEN: a RegistryAuthDefaults.
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
			name: "doesn't give GHCRAuthDefaults",
			newDefaults: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{},
			},
			doesSet: false,
		},
		{
			name: "doesn't give HubAuthDefaults",
			newDefaults: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{},
			},
			doesSet: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: and QuayAuthDefaults to take them.
			data := &QuayAuthDefaults{}

			// WHEN: SetDefaults() is called to give these defaults.
			data.SetDefaults(tc.newDefaults)

			// THEN: they are set when expected.
			if got := data.defaults == tc.newDefaults; got != tc.doesSet {
				t.Errorf(
					"%s\nQuayAuthDefaults.SetDefaults() .defaults mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.doesSet,
				)
			}
		})
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

func TestQuayAuth_CheckValues(t *testing.T) {
	// GIVEN: a QuayAuth.
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
			name: "valid",
			input: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "token",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "token",
			input: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "token",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "no token",
			input: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
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

func TestQuayAuth_GetToken(t *testing.T) {
	// GIVEN: a QuayAuth with/without defaults.
	tests := []struct {
		name string
		data *QuayAuth
		want string
	}{
		{
			name: "empty",
			data: &QuayAuth{},
			want: "",
		},
		{
			name: "no defaults",
			data: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "root",
				},
			},
			want: "root",
		},
		{
			name: "defaults fallback",
			data: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "",
					defaults: &QuayAuthDefaults{
						Token: "defaults",
					},
				},
			},
			want: "defaults",
		},
		{
			name: "defaults fallback recursive",
			data: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "",
					defaults: &QuayAuthDefaults{
						Token: "hard-defaults",
					},
				},
			},
			want: "hard-defaults",
		},
		{
			name: "root Token prioritised",
			data: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "root",
					defaults: &QuayAuthDefaults{
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
					"%s\nQuayAuth.GetToken() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}

			// WHEN: GetTokenSelf() is called on it.
			got = tc.data.GetTokenSelf()

			// THEN: the expected result is returned.
			if got != tc.data.Token {
				t.Fatalf(
					"%s\nQuayAuthDefaults.GetTokenSelf() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.data.Token,
				)
			}
		})
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestQuayAuthDefaults_GetQueryTokenSelf(t *testing.T) {
	// GIVEN: a QuayAuthDefaults.
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetQueryTokenSelf() is called on it.
			queryToken, _ := tc.data.GetQueryTokenSelf()

			prefix := fmt.Sprintf("%s\nQuayAuthDefaults.GetQueryTokenSelf()", packageName)

			// THEN: the query token is returned as expected.
			if queryToken != tc.want {
				t.Errorf(
					"%s queryToken mismatch\ngot:  %q\nwant: %q",
					prefix, queryToken, tc.want,
				)
			}
		})
	}
}

func TestQuayAuth_SetQueryToken(t *testing.T) {
	// GIVEN: a QuayAuth and QuayAuthDefaults.
	queryToken := "new-query-token"
	validUntil := time.Now().Add(10 * time.Second)

	auth := &QuayAuth{
		QuayAuthDefaults: QuayAuthDefaults{Token: "abc"},
	}
	defaults := &QuayAuthDefaults{Token: "def"}

	// WHEN: SetQueryToken is called via concrete types and RegistryAuth.
	auth.SetQueryToken(queryToken, validUntil)
	defaults.SetQueryToken(queryToken, validUntil)

	var viaInterface RegistryAuth = auth
	viaInterface.SetQueryToken(queryToken, validUntil)

	// THEN: token-only auth is unchanged (no-op).
	if got, want := auth.Token, "abc"; got != want {
		t.Fatalf(
			"%s\nQuayAuth.SetQueryToken() changed Token\ngot:  %q\nwant: %q",
			packageName, got, want,
		)
	}
}

func TestQuayAuth_GetQueryToken(t *testing.T) {
	// GIVEN: a QuayAuth with a token.
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
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "abc",
				},
			},
			want: "abc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetQueryToken() is called on it.
			got, err := tc.auth.GetQueryToken(ContainerDetail{})

			prefix := fmt.Sprintf("%s\nQuayAuth.GetQueryToken()", packageName)

			// THEN: the expected Token is returned.
			if got != tc.want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: no error is returned.
			if err != nil {
				t.Errorf(
					"%s error mismatch\ngot:  %v\nwant: nil",
					prefix, err,
				)
			}
		})
	}
}

// ######################
// # AUTH | INHERITANCE #
// ######################

func TestQuayAuth_Inherit(t *testing.T) {
	// GIVEN: a QuayAuth, and a RegistryAuth to try and inherit from.
	tests := []struct {
		name                 string
		auth                 *QuayAuth
		from                 RegistryAuth
		srcDetail, dstDetail ContainerDetail
		wantToken            string
	}{
		{
			name:      "inherit from nil",
			auth:      &QuayAuth{},
			from:      nil,
			wantToken: "",
		},
		{
			name: "inherit from QuayAuth (src.Token is SecretValue)",
			auth: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: util.SecretValue,
				},
			},
			from: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "abc",
				},
			},
			wantToken: "abc",
		},
		{
			name: "do not inherit from QuayAuth when src.Token is not SecretValue",
			auth: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "foo",
				},
			},
			from: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: "abc",
				},
			},
			wantToken: "foo",
		},
		{
			name: "do not inherit from GHCRAuth",
			auth: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: util.SecretValue,
				},
			},
			from: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "abc",
				},
			},
			wantToken: util.SecretValue,
		},
		{
			name: "do not inherit from HubAuth",
			auth: &QuayAuth{
				QuayAuthDefaults: QuayAuthDefaults{
					Token: util.SecretValue,
				},
			},
			from: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Token: "abc",
				},
			},
			wantToken: util.SecretValue,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Inherit is called on it.
			tc.auth.Inherit(tc.from, tc.srcDetail, tc.dstDetail)

			if got := tc.auth.Token; got != tc.wantToken {
				t.Errorf(
					"%s\nQuayAuth.Inherit() .Token value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.wantToken,
				)
			}
		})
	}
}
