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

func TestECRRegistryDefaults_Unmarshal(t *testing.T) {
	// GIVEN: an ECRRegistryDefaults and JSON/YAML to unmarshal into it.
	tests := []struct {
		name     string
		format   string
		data     string
		registry *ECRRegistryDefaults
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			registry: &ECRRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &ECRRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}",
			registry: &ECRRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "foo",
			registry: &ECRRegistryDefaults{},
			errRegex: `invalid character`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &ECRRegistryDefaults{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid Auth",
			format: "json",
			data:   `{"auth": []}`,
			registry: &ECRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &ECRAuth{},
				},
			},
			errRegex: test.TrimYAML(`
				^auth:
					json: .*unmarshal.*$`,
			),
		},
		{
			name:     "JSON/auth-null",
			format:   "json",
			data:     `{"auth": null}`,
			registry: &ECRRegistryDefaults{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:   "YAML/auth-empty",
			format: "yaml",
			data: test.TrimYAML(`
				auth: {}
			`),
			registry: &ECRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &ECRAuth{},
				},
			},
			errRegex: `^$`,
			want:     "{}\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*ECRRegistryDefaults, error) {
					err := decode.Unmarshal(format, data, tc.registry)
					return tc.registry, err
				},
				tc.format, tc.data,
				func(v *ECRRegistryDefaults) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"ECRRegistryDefaults",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: Auth should never be nil.
			if tc.registry.Auth == nil {
				t.Errorf(
					"%s\nECRRegistryDefaults.Unmarshal(format=%q, data=%q) - Auth should not be nil",
					packageName, tc.format, tc.data,
				)
			}
		})
	}
}

func TestECRRegistry_Unmarshal(t *testing.T) {
	// GIVEN: an ECRRegistry and JSON/YAML to unmarshal into it.
	tests := []struct {
		name     string
		format   string
		data     string
		registry *ECRRegistry
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			registry: &ECRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &ECRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &ECRRegistry{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "JSON/invalid ContainerDetail",
			format: "json",
			data:   `{"image": []}`,
			registry: &ECRRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &ECRAuth{},
				},
			},
			errRegex: `^json: .*unmarshal .*$`,
		},
		{
			name:     "JSON/auth-null",
			format:   "json",
			data:     `{"auth": null}`,
			registry: &ECRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:   "JSON/image+tag (auth omitted, ECR is anonymous)",
			format: "json",
			data: test.TrimJSON(`{
				"image": "i",
				"tag": "t",
				"auth": {}
			}`),
			registry: &ECRRegistry{
				CommonRegistry: CommonRegistry{
					Auth: &ECRAuth{},
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			registry, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*ECRRegistry, error) {
					err := decode.Unmarshal(format, data, tc.registry)
					return tc.registry, err
				},
				tc.format, tc.data,
				func(v *ECRRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"ECRRegistry",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}

			// AND: Auth should never be nil.
			if registry.Auth == nil {
				t.Errorf(
					"%s\nECRRegistry.Unmarshal(format=%q, data=%q) - Auth should not be nil",
					packageName, tc.format, tc.data,
				)
			}
		})
	}
}

func TestECRRegistry_ApplyOverrides(t *testing.T) {
	// GIVEN: an ECRRegistry and JSON/YAML to decode into it.
	tests := []struct {
		name     string
		format   string
		data     string
		registry *ECRRegistry
		errRegex string
		want     string
	}{
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			registry: &ECRRegistry{},
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     "foo",
			registry: &ECRRegistry{},
			errRegex: `string was used where mapping is expected`,
		},
		{
			name:   "mutate image+tag",
			format: "yaml",
			data: test.TrimYAML(`
				tag: t
			`),
			registry: &ECRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
					Auth: &ECRAuth{},
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				image: i
				tag: t
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
				func(format string, data []byte, v *ECRRegistry) (*ECRRegistry, error) {
					err := v.ApplyOverrides(format, data)
					return v, err
				},
				tc.format, tc.data,
				func(v *ECRRegistry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				true,
				packageName,
				"ECRRegistry.ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// ####################
// # REGISTRY | STATE #
// ####################

func TestECRRegistryDefaults_IsZero(t *testing.T) {
	// GIVEN: an ECRRegistryDefaults.
	tests := []struct {
		name     string
		registry *ECRRegistryDefaults
		want     bool
	}{
		{
			name:     "nil",
			registry: nil,
			want:     true,
		},
		{
			name:     "empty",
			registry: &ECRRegistryDefaults{},
			want:     true,
		},
		{
			name: "non-empty/queryToken and validUntil",
			registry: &ECRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &ECRAuthDefaults{
						queryToken: "abc",
						validUntil: time.Now().UTC().Add(time.Hour),
					},
				},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called with it.
			got := tc.registry.IsZero()

			// THEN: the expected result is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nECRRegistryDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestECRRegistry_IsZero(t *testing.T) {
	// GIVEN: an ECRRegistry.
	tests := []struct {
		name string
		data *ECRRegistry
		want bool
	}{
		{
			name: "nil",
			data: nil,
			want: true,
		},
		{
			name: "empty",
			data: RegistryMap["ecr"]().(*ECRRegistry),
			want: true,
		},
		{
			name: "non-empty/Type",
			data: &ECRRegistry{
				CommonRegistry: CommonRegistry{
					Type: "abc",
					Auth: RegistryMap["ecr"]().GetAuth(),
				},
			},
			want: false,
		},
		{
			name: "non-empty/ContainerDetail",
			data: &ECRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &ECRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i",
					},
					Type: "abc",
					Auth: RegistryMap["ecr"]().GetAuth(),
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
					"%s\nECRRegistry.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestECRRegistry_Copy(t *testing.T) {
	// GIVEN: an ECRRegistry.
	tests := []struct {
		name     string
		registry *ECRRegistry
		want     string
	}{
		{
			name:     "nil",
			registry: nil,
			want:     "null\n",
		},
		{
			name:     "empty",
			registry: RegistryMap["ecr"]().(*ECRRegistry),
			want:     "{}\n",
		},
		{
			name: "filled",
			registry: &ECRRegistry{
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "i1",
						Tag:   "t1",
					},
					Auth: &ECRAuth{
						ECRAuthDefaults: ECRAuthDefaults{
							queryToken: "qT",
							validUntil: time.Now().Add(time.Hour),
							defaults:   &ECRAuthDefaults{},
						},
					},
				},
			},
			want: test.TrimYAML(`
				image: i1
				tag: t1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy() is called on it.
			gotInterface := tc.registry.Copy()

			prefix := fmt.Sprintf("%s\nECRRegistry.Copy()", packageName)

			// THEN: the returned Registry unmarshals the same.
			if got := decode.ToYAMLString(gotInterface, ""); got != tc.want {
				t.Fatalf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the returned Registry is an ECRRegistry.
			got, ok := gotInterface.(*ECRRegistry)
			if !ok {
				if gotInterface == nil {
					return
				}
				t.Fatalf(
					"%s returned wrong type: %T",
					prefix, gotInterface,
				)
			}
			hadAuth, hasAuth := tc.registry.GetAuth().(*ECRAuth)
			if !hasAuth {
				return
			}

			// AND: the cache values are copied.
			gotAuth, ok := got.GetAuth().(*ECRAuth)
			if !ok ||
				gotAuth.queryToken != hadAuth.queryToken ||
				gotAuth.validUntil != hadAuth.validUntil ||
				got.defaults != tc.registry.defaults {
				t.Fatalf(
					"%s mismatch\ngot:  %+v\nwant: %s",
					prefix, got, tc.want,
				)
			}

			// AND: the returned ECRAuth is at a different address.
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

func TestECRRegistryDefaults_String(t *testing.T) {
	// GIVEN: an ECRRegistryDefaults.
	tests := []struct {
		name string
		data *ECRRegistryDefaults
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "",
		},
		{
			name: "empty",
			data: &ECRRegistryDefaults{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &ECRRegistryDefaults{
				CommonRegistryDefaults: CommonRegistryDefaults{
					Auth: &ECRAuthDefaults{
						queryToken: "qT",
						validUntil: time.Now().Add(time.Hour),
					},
				},
			},
			want: "{}\n",
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

func TestECRRegistry_String(t *testing.T) {
	// GIVEN: an ECRRegistry.
	tests := []struct {
		name string
		data *ECRRegistry
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "",
		},
		{
			name: "empty",
			data: &ECRRegistry{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &ECRRegistry{
				CommonRegistry: CommonRegistry{
					Type: "test-ecr",
					ContainerDetail: ContainerDetail{
						Image: "i1",
						Tag:   "t1",
					},
					Auth: &ECRAuth{
						ECRAuthDefaults: ECRAuthDefaults{
							queryToken: "qT",
							validUntil: time.Now().Add(time.Hour),
						},
					},
				},
			},
			want: test.TrimYAML(`
				type: test-ecr
				image: i1
				tag: t1
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

func TestECRRegistryDefaults_GetType(t *testing.T) {
	// GIVEN: an ECRRegistryDefaults.
	var registry ECRRegistryDefaults

	// WHEN: GetType is called on it.
	got := registry.GetType()

	// THEN: the type is returned.
	if want := "ecr"; got != want {
		t.Errorf(
			"%s\ngot %q, want %q",
			packageName, got, want,
		)
	}
}

func TestECRRegistry_GetType(t *testing.T) {
	// GIVEN: an ECRRegistry.
	var registry ECRRegistry

	// WHEN: GetType is called on it.
	got := registry.GetType()

	// THEN: the type is returned.
	if want := "ecr"; got != want {
		t.Errorf(
			"%s\ngot %q, want %q",
			packageName, got, want,
		)
	}
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestECRRegistry_NewRequest(t *testing.T) {
	// GIVEN: an ECRRegistry, and a tag.
	tests := []struct {
		name     string
		registry *ECRRegistry
		tag      string
		errRegex string
	}{
		{
			name: "no image or tag",
			registry: &ECRRegistry{
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
			registry: &ECRRegistry{
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
			registry: &ECRRegistry{
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
			registry: &ECRRegistry{
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

			// WHEN: newRequest() is called on it.
			req, err := tc.registry.newRequest(tc.tag)

			prefix := fmt.Sprintf(
				"%s\nECRRegistry.newRequest(%q)",
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

// ###################
// # AUTH | DECODING #
// ###################

func TestECRAuthDefaults_Unmarshal(t *testing.T) {
	// GIVEN: an ECRAuthDefaults.
	tests := []struct {
		name         string
		format, data string
		errRegex     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "{}",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "invalid",
			errRegex: `invalid character`,
		},
		{
			name:     "JSON/sequence rejected",
			format:   "json",
			data:     `[]`,
			errRegex: `unmarshal JSON array`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var auth ECRAuthDefaults
			err := decode.Unmarshal(tc.format, []byte(tc.data), &auth)

			// THEN: any error case is expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s\nECRAuthDefaults.Unmarshal(format=%q, data=%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.format, tc.data, e, tc.errRegex,
				)
			}
		})
	}
}

// ################
// # AUTH | STATE #
// ################

func TestECRAuthDefaults_IsZero(t *testing.T) {
	// GIVEN: an ECRAuthDefaults.
	tests := []struct {
		name string
		data *ECRAuthDefaults
		want bool
	}{
		{
			name: "nil",
			data: nil,
			want: true,
		},
		{
			name: "empty",
			data: &ECRAuthDefaults{},
			want: true,
		},
		{
			name: "filled",
			data: &ECRAuthDefaults{
				queryToken: "qT",
				validUntil: time.Now().Add(time.Hour),
			},
			want: true,
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
					"%s\nECRAuthDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestECRAuth_Copy(t *testing.T) {
	// GIVEN: an ECRAuth.
	tests := []struct {
		name string
		auth *ECRAuth
		want string
	}{
		{
			name: "nil",
			auth: nil,
			want: "null\n",
		},
		{
			name: "empty",
			auth: &ECRAuth{},
			want: "{}\n",
		},
		{
			name: "filled",
			auth: &ECRAuth{
				ECRAuthDefaults: ECRAuthDefaults{
					queryToken: "qT",
					validUntil: time.Now(),
					defaults:   &ECRAuthDefaults{},
				},
			},
			want: "{}\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy() is called on it.
			authCopy := tc.auth.Copy()

			prefix := fmt.Sprintf("%s\nECRAuth.Copy()", packageName)

			// THEN: the returned RegistryAuth unmarshals the same.
			if got := decode.ToYAMLString(authCopy, ""); got != tc.want {
				t.Fatalf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the returned RegistryAuth is an ECRAuth.
			got, ok := authCopy.(*ECRAuth)
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

			// AND: the cache values are copied.
			if got.queryToken != tc.auth.queryToken ||
				got.validUntil != tc.auth.validUntil ||
				got.defaults != tc.auth.defaults {
				t.Fatalf(
					"%s mismatch\ngot:  %+v\nwant: %+v",
					prefix, got, tc.auth,
				)
			}

			// AND: the returned ECRAuth is at a different address.
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

func TestECRAuthDefaults_String(t *testing.T) {
	// GIVEN: an ECRAuthDefaults.
	tests := []struct {
		name string
		data *ECRAuthDefaults
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "null\n",
		},
		{
			name: "empty",
			data: &ECRAuthDefaults{},
			want: "{}\n",
		},
		{
			name: "filled",
			data: &ECRAuthDefaults{
				queryToken: "t1",
				validUntil: time.Now().Add(time.Hour),
			},
			want: "{}\n",
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

func TestECRAuthDefaults_Defaults(t *testing.T) {
	// GIVEN: an ECRAuthDefaults.
	tests := []struct {
		name         string
		data         *ECRAuthDefaults
		haveDefaults bool
	}{
		{
			name:         "no defaults",
			data:         &ECRAuthDefaults{},
			haveDefaults: false,
		},
		{
			name: "defaults",
			data: &ECRAuthDefaults{
				defaults: &ECRAuthDefaults{},
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
					"%s\nECRAuthDefaults.Defaults() mismatch\ngot:  %t\nwant: %t",
					packageName, gotDefaults, tc.haveDefaults,
				)
			}
		})
	}
}

func TestECRAuthDefaults_SetDefaults(t *testing.T) {
	// GIVEN: a RegistryAuthDefaults.
	tests := []struct {
		name        string
		newDefaults RegistryAuthDefaults
		doesSet     bool
	}{
		{
			name:        "give ECRAuthDefaults",
			newDefaults: &ECRAuthDefaults{},
			doesSet:     true,
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
		{
			name:        "doesn't give QuayAuthDefaults",
			newDefaults: &QuayAuthDefaults{},
			doesSet:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: an ECRAuthDefaults to take them.
			data := &ECRAuthDefaults{}

			// WHEN: SetDefaults() is called to give these defaults.
			data.SetDefaults(tc.newDefaults)

			// THEN: they are set when expected.
			if got := data.defaults == tc.newDefaults; got != tc.doesSet {
				t.Errorf(
					"%s\nECRAuthDefaults.SetDefaults() .defaults mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.doesSet,
				)
			}
		})
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

func TestECRAuth_CheckValues(t *testing.T) {
	// GIVEN: an ECRAuth.
	tests := []struct {
		name     string
		input    *ECRAuth
		errRegex string
	}{
		{
			name:     "nil",
			input:    (*ECRAuth)(nil),
			errRegex: `^$`,
		},
		{
			name:     "empty",
			input:    &ECRAuth{},
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

func TestECRAuthDefaults_GetTokenSelf(t *testing.T) {
	// GIVEN: an ECRAuthDefaults.
	data := &ECRAuthDefaults{}

	// WHEN: GetTokenSelf() is called on it.
	got := data.GetTokenSelf()

	// THEN: the empty token is returned (Amazon ECR Public Gallery is anonymous).
	if got != "" {
		t.Errorf(
			"%s\nECRAuthDefaults.GetTokenSelf() mismatch\ngot:  %q\nwant: %q",
			packageName, got, "",
		)
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestECRAuthDefaults_GetQueryTokenSelf(t *testing.T) {
	// GIVEN: an ECRAuthDefaults.
	tests := []struct {
		name string
		data *ECRAuthDefaults
		want string
	}{
		{
			name: "empty",
			data: &ECRAuthDefaults{},
			want: "",
		},
		{
			name: "expired",
			data: &ECRAuthDefaults{
				queryToken: "query-token",
				validUntil: time.Now().Add(-1 * time.Second),
			},
			want: "",
		},
		{
			name: "valid",
			data: &ECRAuthDefaults{
				queryToken: "query-token",
				validUntil: time.Now().Add(10 * time.Second),
			},
			want: "query-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetQueryTokenSelf() is called on it.
			queryToken, _ := tc.data.GetQueryTokenSelf()

			// THEN: the query token is returned as expected.
			if queryToken != tc.want {
				t.Errorf(
					"%s\nECRAuthDefaults.GetQueryTokenSelf() queryToken mismatch\ngot:  %q\nwant: %q",
					packageName, queryToken, tc.want,
				)
			}
		})
	}
}

func TestECRAuthDefaults_GetQueryTokenSelf__parallel(t *testing.T) {
	// GIVEN: an ECRAuthDefaults with a valid queryToken.
	data := &ECRAuthDefaults{
		queryToken: "query-token",
		validUntil: time.Now().Add(time.Hour),
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
	// Call again to verify we still get the queryToken once the lock is released.
	queryToken, _ := data.GetQueryTokenSelf()

	// THEN: the query token is returned as expected.
	if queryToken != data.queryToken {
		t.Errorf(
			"%s\nECRAuthDefaults.GetQueryTokenSelf() queryToken mismatch\ngot:  %q\nwant: %q",
			packageName, queryToken, data.queryToken,
		)
	}
}

func TestECRAuth_GetQueryToken__cached(t *testing.T) {
	// GIVEN: an ECRAuth with a valid cached query token.
	tests := []struct {
		name string
		data *ECRAuth
		want string
	}{
		{
			name: "valid on self",
			data: &ECRAuth{
				ECRAuthDefaults: ECRAuthDefaults{
					queryToken: "query-token",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
			want: "query-token",
		},
		{
			name: "valid via defaults chain",
			data: &ECRAuth{
				ECRAuthDefaults: ECRAuthDefaults{
					defaults: &ECRAuthDefaults{
						queryToken: "default-query-token",
						validUntil: time.Now().Add(10 * time.Second),
					},
				},
			},
			want: "default-query-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetQueryToken() is called on it.
			queryToken, err := tc.data.GetQueryToken(ContainerDetail{Image: test.ArgusDockerECRRepo})

			// THEN: no refresh is needed, so no error.
			if err != nil {
				t.Fatalf(
					"%s\nECRAuth.GetQueryToken() unexpected error: %s",
					packageName, errfmt.FormatError(err),
				)
			}

			// AND: the cached query token is returned.
			if queryToken != tc.want {
				t.Errorf(
					"%s\nECRAuth.GetQueryToken() mismatch\ngot:  %q\nwant: %q",
					packageName, queryToken, tc.want,
				)
			}
		})
	}
}

func TestECRAuthDefaults_SetQueryToken(t *testing.T) {
	// GIVEN: an ECRAuth with a defaults chain.
	data := &ECRAuth{
		ECRAuthDefaults: ECRAuthDefaults{
			defaults: &ECRAuthDefaults{},
		},
	}

	queryToken := "new-query-token"
	validUntil := time.Now().Add(10 * time.Second)

	// WHEN: SetQueryToken() is called on it.
	data.SetQueryToken(queryToken, validUntil)

	// THEN: the queryToken is set on self.
	if data.queryToken != queryToken || !data.validUntil.Equal(validUntil) {
		t.Errorf(
			"%s\nECRAuthDefaults.SetQueryToken() self not set\ngot:  %q/%v\nwant: %q/%v",
			packageName, data.queryToken, data.validUntil, queryToken, validUntil,
		)
	}

	// AND: the defaults are not touched (repo-global token cached per-instance).
	if data.defaults.queryToken != "" {
		t.Errorf(
			"%s\nECRAuthDefaults.SetQueryToken() should not propagate to defaults\ngot:  %q",
			packageName, data.defaults.queryToken,
		)
	}
}

func TestECRAuth_RefreshQueryToken__cached(t *testing.T) {
	// GIVEN: an ECRAuth with a cached query token that is valid for a while.
	auth := &ECRAuth{
		ECRAuthDefaults: ECRAuthDefaults{
			queryToken: "cached-token",
			validUntil: time.Now().Add(time.Hour),
		},
	}

	// WHEN: refreshQueryToken is called on it.
	queryToken, err := auth.refreshQueryToken(ContainerDetail{Image: test.ArgusDockerECRRepo})

	prefix := fmt.Sprintf("%s\nECRAuth.refreshQueryToken()", packageName)

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

func TestECRAuth_Inherit(t *testing.T) {
	// GIVEN: an ECRAuth, and a RegistryAuth to try and inherit from.
	tests := []struct {
		name                 string
		auth                 *ECRAuth
		from                 RegistryAuth
		srcDetail, dstDetail ContainerDetail
		inherit              bool
	}{
		{
			name: "inherit from nil",
			auth: &ECRAuth{},
			from: nil,
		},
		{
			name: "inherit from ECRAuth (tokens are global, images differ)",
			auth: &ECRAuth{},
			from: &ECRAuth{
				ECRAuthDefaults: ECRAuthDefaults{
					queryToken: "qt",
					validUntil: time.Now(),
				},
			},
			srcDetail: ContainerDetail{Image: "a", Tag: "b"},
			dstDetail: ContainerDetail{Image: "c", Tag: "d"},
			inherit:   true,
		},
		{
			name: "do not inherit from GHCRAuth",
			auth: &ECRAuth{},
			from: &GHCRAuth{
				GHCRAuthDefaults: GHCRAuthDefaults{
					Token: "abc",
				},
			},
		},
		{
			name: "do not inherit from HubAuth",
			auth: &ECRAuth{},
			from: &HubAuth{
				HubAuthDefaults: HubAuthDefaults{
					Username: "user",
					Token:    "abc",
				},
			},
		},
		{
			name: "do not inherit from QuayAuth",
			auth: &ECRAuth{},
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

			_, hadQueryToken, hadValidUntil := getTokenData(t, tc.auth)

			// WHEN: Inherit is called on it.
			tc.auth.Inherit(tc.from, tc.srcDetail, tc.dstDetail)

			prefix := fmt.Sprintf(
				"%s\nECRAuth.Inherit(from=%T)",
				packageName, tc.from,
			)

			// THEN: the expected inheritance has occurred.
			_, gotQueryToken, gotValidUntil := getTokenData(t, tc.auth)
			queryTokenInherited := hadQueryToken != gotQueryToken
			validUntilInherited := !hadValidUntil.Equal(gotValidUntil)
			if queryTokenInherited != tc.inherit || validUntilInherited != tc.inherit {
				t.Errorf(
					"%s inheritance mismatch (want inherit=%t)\nqueryToken: had %q, got %q\nvalidUntil: had %q, got %q",
					prefix, tc.inherit,
					hadQueryToken, gotQueryToken,
					hadValidUntil, gotValidUntil,
				)
			}
		})
	}
}
