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

func TestECRRegistryDefaults_Unmarshal(t *testing.T) {
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
			data:     "",
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
				Auth: &ECRAuth{},
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
				Auth: &ECRAuth{},
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
				"ECRRegistryDefaults.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}

			if tc.registry.Auth == nil {
				t.Error("ECRRegistryDefaults.Auth mismatch\ngot:  nil\nwant: non-nil")
			}
		})
	}
}

func TestECRRegistry_Unmarshal(t *testing.T) {
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
				Auth: &ECRAuth{},
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
			name:   "JSON/image+tag/empty auth omitted",
			format: "json",
			data: test.TrimJSON(`{
				"image": "i",
				"tag": "t",
				"auth": {}
			}`),
			registry: &ECRRegistry{
				Auth: &ECRAuth{},
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
				"ECRRegistry.Unmarshal",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}

			if registry.Auth == nil {
				t.Error("ECRRegistry.Auth mismatch\ngot:  nil\nwant: non-nil")
			}
		})
	}
}

func TestECRRegistry_ApplyOverrides(t *testing.T) {
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
				Image: "i",
				Auth:  &ECRAuth{},
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
			name: "cache only/queryToken and validUntil",
			registry: &ECRRegistryDefaults{
				Auth: &ECRAuthDefaults{
					queryToken: "abc",
					validUntil: time.Now().UTC().Add(time.Hour),
				},
			},
			want: true,
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

func TestECRRegistry_IsZero(t *testing.T) {
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
				Type: "abc",
				Auth: RegistryMap["ecr"]().GetAuth(),
			},
			want: false,
		},
		{
			name: "non-empty/ContainerDetail",
			data: &ECRRegistry{
				Image: "i",
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &ECRRegistry{
				Image: "i",
				Type:  "abc",
				Auth:  RegistryMap["ecr"]().GetAuth(),
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

func TestECRRegistry_Copy(t *testing.T) {
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
				Image: "i1",
				Tag:   "t1",
				Auth: &ECRAuth{
					queryToken: "qT",
					validUntil: time.Now().Add(time.Hour),
					defaults:   &ECRAuthDefaults{},
				},
				defaults: &ECRRegistryDefaults{},
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

			gotInterface := tc.registry.Copy()

			if got := decode.ToYAMLString(gotInterface, ""); got != tc.want {
				t.Fatalf("ECRRegistry mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			got, ok := gotInterface.(*ECRRegistry)
			if !ok {
				if gotInterface == nil {
					return
				}
				t.Fatalf(
					"ECRRegistry.Copy() type mismatch\ngot:  %T\nwant: *ECRRegistry",
					gotInterface,
				)
			}
			hadAuth, hasAuth := tc.registry.GetAuth().(*ECRAuth)
			if !hasAuth {
				return
			}

			gotAuth, ok := got.GetAuth().(*ECRAuth)
			if !ok {
				t.Fatalf(
					"ECRRegistry.GetAuth() type mismatch\ngot:  %T\nwant: *ECRAuth",
					got.GetAuth(),
				)
			}
			if gotAuth.queryToken != hadAuth.queryToken ||
				gotAuth.validUntil != hadAuth.validUntil ||
				got.defaults != tc.registry.defaults {
				t.Fatalf(
					"ECRRegistry mismatch\n"+
						"got:  Auth.queryToken=%q, Auth.validUntil=%s, defaults=%p\n"+
						"want: Auth.queryToken=%q, Auth.validUntil=%s, defaults=%p",
					gotAuth.queryToken, gotAuth.validUntil, got.defaults,
					hadAuth.queryToken, hadAuth.validUntil, tc.registry.defaults,
				)
			}

			if gotAuth == hadAuth {
				t.Fatalf(
					"ECRRegistry.GetAuth(): copy aliases the original\ngot:  %p\nwant: NOT %p",
					gotAuth, hadAuth,
				)
			}
		})
	}
}

// ########################
// # REGISTRY | STRINGIFY #
// ########################

func TestECRRegistryDefaults_String(t *testing.T) {
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
				Auth: &ECRAuthDefaults{
					queryToken: "qT",
					validUntil: time.Now().Add(time.Hour),
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
				"ECRRegistryDefaults.String",
				tc.data.String,
				tc.want,
			)
		})
	}
}

func TestECRRegistry_String(t *testing.T) {
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
				Type:  "test-ecr",
				Image: "i1",
				Tag:   "t1",
				Auth: &ECRAuth{
					queryToken: "qT",
					validUntil: time.Now().Add(time.Hour),
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
				"ECRRegistry.String",
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
	var registry ECRRegistryDefaults

	got := registry.GetType()

	if want := "ecr"; got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestECRRegistry_GetType(t *testing.T) {
	var registry ECRRegistry

	got := registry.GetType()

	if want := "ecr"; got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

// #########################
// # REGISTRY | OPERATIONS #
// #########################

func TestECRRegistry_NewRequest(t *testing.T) {
	tests := []struct {
		name     string
		registry *ECRRegistry
		tag      string
		errRegex string
	}{
		{
			name: "no image or tag",
			registry: &ECRRegistry{
				Image: "",
				Tag:   "",
			},
			errRegex: `^$`,
		},
		{
			name: "have image+tag",
			registry: &ECRRegistry{
				Image: "123",
				Tag:   "not-used",
			},
			tag:      "foo",
			errRegex: `^$`,
		},
		{
			name: "tag: invalid",
			registry: &ECRRegistry{
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
			registry: &ECRRegistry{
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

	headers := map[string]string{
		"Accept": "application/vnd.oci.image.index.v1+json",
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

			for key, value := range headers {
				if got := req.Header.Get(key); got != value {
					t.Errorf(
						"http.Request.Header[%q] mismatch\ngot:  %q\nwant: %q",
						key, got, value,
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
			data:     "",
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

func TestECRAuthDefaults_IsZero(t *testing.T) {
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
			name: "cache only/queryToken and validUntil",
			data: &ECRAuthDefaults{
				queryToken: "qT",
				validUntil: time.Now().Add(time.Hour),
			},
			want: true, // ECR auth carries no configuration.
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

func TestECRAuth_Copy(t *testing.T) {
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
				queryToken: "qT",
				validUntil: time.Now(),
				defaults:   &ECRAuthDefaults{},
			},
			want: "{}\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authCopy := tc.auth.Copy()

			if got := decode.ToYAMLString(authCopy, ""); got != tc.want {
				t.Fatalf("ECRAuth mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			got, ok := authCopy.(*ECRAuth)
			if !ok {
				if authCopy == nil {
					return
				}
				t.Fatalf("ECRAuth.Copy() type mismatch\ngot:  %T\nwant: *ECRAuth", authCopy)
			}
			if tc.auth == nil {
				return
			}

			if got.queryToken != tc.auth.queryToken ||
				got.validUntil != tc.auth.validUntil ||
				got.defaults != tc.auth.defaults {
				t.Fatalf(
					"ECRAuth mismatch\n"+
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

func TestECRAuthDefaults_String(t *testing.T) {
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
				"ECRAuthDefaults.String",
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

			got := tc.data.Defaults()

			if gotDefaults := got != nil; gotDefaults != tc.haveDefaults {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", gotDefaults, tc.haveDefaults)
			}
		})
	}
}

func TestECRAuthDefaults_SetDefaults(t *testing.T) {
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

			data := &ECRAuthDefaults{}

			data.SetDefaults(tc.newDefaults)

			var want *ECRAuthDefaults
			if tc.doesSet {
				want = tc.newDefaults.(*ECRAuthDefaults)
			}
			if data.defaults != want {
				t.Errorf(
					"ECRAuthDefaults.defaults mismatch\ngot:  %p\nwant: %p",
					data.defaults, want,
				)
			}
		})
	}
}

// #####################
// # AUTH | VALIDATION #
// #####################

func TestECRAuth_CheckValues(t *testing.T) {
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
	data := &ECRAuthDefaults{}

	got := data.GetTokenSelf()

	// Amazon ECR Public Gallery is anonymous.
	if want := ""; got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

// #######################
// # AUTH | QUERY TOKENS #
// #######################

func TestECRAuthDefaults_GetQueryTokenSelf(t *testing.T) {
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

			queryToken, _ := tc.data.GetQueryTokenSelf()

			if queryToken != tc.want {
				t.Errorf("queryToken mismatch\ngot:  %q\nwant: %q", queryToken, tc.want)
			}
		})
	}
}

func TestECRAuthDefaults_GetQueryTokenSelf__parallel(t *testing.T) {
	data := &ECRAuthDefaults{
		queryToken: "query-token",
		validUntil: time.Now().Add(time.Hour),
	}

	data.mu.Lock()

	queryTokens := make([]string, 10)
	done := make(chan int, len(queryTokens))
	for i := range queryTokens {
		go func() {
			queryTokens[i], _ = data.GetQueryTokenSelf()
			done <- i
		}()
	}

	select {
	case i := <-done:
		data.mu.Unlock()
		t.Fatalf(
			"goroutine %d: ECRAuthDefaults.GetQueryTokenSelf() returned while the lock was held",
			i,
		)
	case <-time.After(50 * time.Millisecond):
	}

	data.mu.Unlock()
	for range queryTokens {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("ECRAuthDefaults.GetQueryTokenSelf() did not return once the lock was released")
		}
	}
	for i, queryToken := range queryTokens {
		if queryToken != data.queryToken {
			t.Errorf(
				"goroutine %d: ECRAuthDefaults.GetQueryTokenSelf() mismatch\ngot:  %q\nwant: %q",
				i, queryToken, data.queryToken,
			)
		}
	}
}

func TestECRAuth_GetQueryToken__cached(t *testing.T) {
	tests := []struct {
		name string
		data *ECRAuth
		want string
	}{
		{
			name: "valid on self",
			data: &ECRAuth{
				queryToken: "query-token",
				validUntil: time.Now().Add(10 * time.Second),
			},
			want: "query-token",
		},
		{
			name: "valid via defaults chain",
			data: &ECRAuth{
				defaults: &ECRAuthDefaults{
					queryToken: "default-query-token",
					validUntil: time.Now().Add(10 * time.Second),
				},
			},
			want: "default-query-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			queryToken, err := tc.data.GetQueryToken(ContainerDetail{Image: test.ArgusDockerECRRepo})
			if err != nil {
				t.Fatalf("unexpected error: %s", errfmt.FormatError(err))
			}

			if queryToken != tc.want {
				t.Errorf("queryToken mismatch\ngot:  %q\nwant: %q", queryToken, tc.want)
			}
		})
	}
}

func TestECRAuthDefaults_SetQueryToken(t *testing.T) {
	data := &ECRAuth{
		defaults: &ECRAuthDefaults{},
	}

	queryToken := "new-query-token"
	validUntil := time.Now().Add(10 * time.Second)

	data.SetQueryToken(queryToken, validUntil)

	if data.queryToken != queryToken || !data.validUntil.Equal(validUntil) {
		t.Errorf(
			"ECRAuth mismatch\n"+
				"got:  queryToken=%q, validUntil=%s\n"+
				"want: queryToken=%q, validUntil=%s",
			data.queryToken, data.validUntil,
			queryToken, validUntil,
		)
	}

	// SetQueryToken writes only the receiver.
	if data.defaults.queryToken != "" {
		t.Errorf(
			"ECRAuth.defaults.queryToken mismatch\ngot:  %q\nwant: \"\"",
			data.defaults.queryToken,
		)
	}
}

func TestECRAuth_RefreshQueryToken__cached(t *testing.T) {
	auth := &ECRAuth{
		queryToken: "cached-token",
		validUntil: time.Now().Add(time.Hour),
	}

	queryToken, err := auth.refreshQueryToken(ContainerDetail{Image: test.ArgusDockerECRRepo})
	e := errfmt.FormatError(err)
	if !util.RegexCheck(`^$`, e) {
		t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, `^$`)
	}
	if got, want := queryToken, "cached-token"; got != want {
		t.Fatalf("queryToken mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

// ######################
// # AUTH | INHERITANCE #
// ######################

func TestECRAuth_Inherit(t *testing.T) {
	tests := []struct {
		name                 string
		from                 RegistryAuth
		srcDetail, dstDetail ContainerDetail
		inherit              bool
	}{
		{
			name: "inherit from nil",
			from: nil,
		},
		{
			name: "inherit from ECRAuth (tokens are global, images differ)",
			from: &ECRAuth{
				queryToken: "qt",
				validUntil: time.Now(),
			},
			srcDetail: ContainerDetail{Image: "a", Tag: "b"},
			dstDetail: ContainerDetail{Image: "c", Tag: "d"},
			inherit:   true,
		},
		{
			name: "do not inherit from GHCRAuth",
			from: &GHCRAuth{
				Token: "abc",
			},
		},
		{
			name: "do not inherit from HubAuth",
			from: &HubAuth{
				Username: "user",
				Token:    "abc",
			},
		},
		{
			name: "do not inherit from QuayAuth",
			from: &QuayAuth{
				Token: "abc",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			auth := &ECRAuth{
				queryToken: "keep",
				validUntil: time.Now().UTC().Add(10 * time.Minute),
			}
			had := getTokenData(auth)

			auth.Inherit(tc.from, tc.srcDetail, tc.dstDetail)

			assertInherited(t, "ECRAuth", had, auth, tc.from, tc.inherit)
		})
	}
}
