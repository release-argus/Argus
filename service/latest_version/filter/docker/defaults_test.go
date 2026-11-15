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

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// ############
// # DECODING #
// ############

func TestDecodeDefaults(t *testing.T) {
	tests := []struct {
		name         string
		format, data string
		errRegex     string
		want         string
	}{
		{
			name:   "JSON/empty",
			format: "json",
			data:   "",
			errRegex: test.TrimYAML(`
				docker:
					jsontext:
						unexpected EOF`,
			),
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:   "JSON/invalid",
			format: "json",
			data:   "invalid",
			errRegex: test.TrimYAML(`
				^docker:
					jsontext:
						invalid character.*$`,
			),
		},
		{
			name:   "YAML/invalid",
			format: "yaml",
			data:   "invalid",
			errRegex: test.TrimYAML(`
				^docker:
					[^\s]+ string was used where mapping is expected`,
			),
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ghcr",
				"tag": "{{ version }}",
				"registry": {
					"ghcr": {
						"auth": {
							"token": "ghcr-secret"
						}
					},
					"hub": {
						"auth": {
							"username": "hub-user",
							"token": "hub-secret"
						}
					},
					"quay": {
						"auth": {
							"token": "quay-secret"
						}
					}
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: ghcr
				tag: '{{ version }}'
				registry:
					ghcr:
						auth:
							token: ghcr-secret
					hub:
						auth:
							username: hub-user
							token: hub-secret
					quay:
						auth:
							token: quay-secret
			`),
		},
		{
			name:   "YAML/filled/bare",
			format: "yaml",
			data: test.TrimYAML(`
				type: ghcr
				tag: '{{ version }}'
				registry:
					ghcr:
						auth:
							token: ghcr-secret
					hub:
						auth:
							username: hub-user
							token: hub-secret
					quay:
						auth:
							token: quay-secret
			`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: ghcr
				tag: '{{ version }}'
				registry:
					ghcr:
						auth:
							token: ghcr-secret
					hub:
						auth:
							username: hub-user
							token: hub-secret
					quay:
						auth:
							token: quay-secret
			`),
		},
		{
			name:   "YAML/filled/oldStyle",
			format: "yaml",
			data: test.TrimYAML(`
				type: ghcr
				ghcr:
					token: ghcr-secret
				hub:
					username: hub-user
					token: hub-secret
				quay:
					token: quay-secret
			`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: ghcr
				registry:
					ghcr:
						auth:
							token: ghcr-secret
					hub:
						auth:
							username: hub-user
							token: hub-secret
					quay:
						auth:
							token: quay-secret
			`),
		},
		{
			name:   "JSON/invalid registry values",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ghcr",
				"image": "{{ service_id }}",
				"tag": "{{ version }}",
				"registry": {
					"ghcr": ["a", "b", "c"]
				}
			}`),
			errRegex: test.TrimYAML(`
				^docker:
					json: .*unmarshal JSON array into Go docker\.GHCRRegistryDefaults within "/registry/ghcr":`,
			),
		},
		{
			name:   "JSON/invalid registry auth values",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ghcr",
				"image": "{{ service_id }}",
				"tag": "{{ version }}",
				"registry": {
					"ghcr": {
						"auth": {
							"token": ["ghcr-secret"]
						}
					}
				}
			}`),
			errRegex: test.TrimYAML(`
				^docker:
					[^\s]+ .*unmarshal.*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var defaults Defaults
			defaults.Default()
			defaults.Type = ""
			defaults.Tag = ""

			got, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Defaults, error) {
					return DecodeDefaults(format, data, &defaults)
				},
				tc.format, tc.data,
				func(v *Defaults) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"DecodeDefaults",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil {
				return
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Defaults", Got: got.Defaults, Want: &defaults, Mode: test.CompareSamePointer},
				{Name: "Registry.ECR.GetAuth().Defaults()", Got: got.Registry.ECR.GetAuth().Defaults(), Want: defaults.Registry.ECR.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "Registry.GHCR.GetAuth().Defaults()", Got: got.Registry.GHCR.GetAuth().Defaults(), Want: defaults.Registry.GHCR.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "Registry.Hub.GetAuth().Defaults()", Got: got.Registry.Hub.GetAuth().Defaults(), Want: defaults.Registry.Hub.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "Registry.Quay.GetAuth().Defaults()", Got: got.Registry.Quay.GetAuth().Defaults(), Want: defaults.Registry.Quay.GetAuth(), Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Defaults"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestDefaults_MarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		defaults *Defaults
		want     string
		errRegex string
	}{
		{
			name:     "empty",
			defaults: &Defaults{},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "static fields",
			defaults: &Defaults{
				Type: PossibleTypes[0],
				Tag:  "1.2.3",
			},
			want: test.TrimYAML(`
				type: ` + PossibleTypes[0] + `
				tag: 1.2.3
			`),
			errRegex: `^$`,
		},
		{
			name: "dynamic fields",
			defaults: &Defaults{
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						Auth: &GHCRAuth{
							Token: "ghcr-token",
						},
					},
					Hub: &HubRegistryDefaults{
						Auth: &HubAuthDefaults{
							Username: "hub-username",
							Token:    "hub-token",
						},
					},
					Quay: &QuayRegistryDefaults{
						Auth: &QuayAuthDefaults{
							Token: "quay-token",
						},
					},
				},
			},
			want: test.TrimYAML(`
				registry:
					ghcr:
						auth:
							token: ghcr-token
					hub:
						auth:
							username: hub-username
							token: hub-token
					quay:
						auth:
							token: quay-token
			`),
			errRegex: `^$`,
		},
		{
			name: "filled",
			defaults: &Defaults{
				Type: PossibleTypes[0],
				Tag:  "1.2.3",
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						Auth: &GHCRAuth{
							Token: "ghcr-token",
						},
					},
					Hub: &HubRegistryDefaults{
						Auth: &HubAuthDefaults{
							Username: "hub-username",
							Token:    "hub-token",
						},
					},
					Quay: &QuayRegistryDefaults{
						Auth: &QuayAuthDefaults{
							Token: "quay-token",
						},
					},
				},
			},
			want: test.TrimYAML(`
				type: ` + PossibleTypes[0] + `
				tag: 1.2.3
				registry:
					ghcr:
						auth:
							token: ghcr-token
					hub:
						auth:
							username: hub-username
							token: hub-token
					quay:
						auth:
							token: quay-token
			`),
			errRegex: `^$`,
		},
		{
			name: "empty dynamic fields",
			defaults: &Defaults{
				Type: PossibleTypes[0],
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{},
					Hub:  &HubRegistryDefaults{},
					Quay: &QuayRegistryDefaults{},
				},
			},
			want:     "type: " + PossibleTypes[0] + "\n",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotYAML, err := decode.Marshal("yaml", tc.defaults)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if got := string(gotYAML); got != tc.want {
				t.Errorf("decode.Marshal() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

// #########
// # STATE #
// #########

func TestDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data *Defaults
		want bool
	}{
		{
			name: "empty",
			data: &Defaults{},
			want: true,
		},
		{
			name: "non-empty/Type",
			data: &Defaults{
				Type: "a",
			},
			want: false,
		},
		{
			name: "non-empty/ContainerDetailDefaults",
			data: &Defaults{
				Tag: "a",
			},
			want: false,
		},
		{
			name: "non-empty/Registry",
			data: &Defaults{
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						Auth: &GHCRAuthDefaults{
							Token: "a",
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

			got := tc.data.IsZero()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestRegistryDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data *RegistryDefaultsSet
		want bool
	}{
		{
			name: "nil",
			data: nil,
			want: true,
		},
		{
			name: "empty",
			data: &RegistryDefaultsSet{},
			want: true,
		},
		{
			name: "non-empty/GHCR",
			data: &RegistryDefaultsSet{
				GHCR: &GHCRRegistryDefaults{
					Auth: &GHCRAuthDefaults{
						Token: "ghcr-token",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Hub",
			data: &RegistryDefaultsSet{
				Hub: &HubRegistryDefaults{
					Auth: &HubAuthDefaults{
						Token: "hub-token",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Quay",
			data: &RegistryDefaultsSet{
				Quay: &QuayRegistryDefaults{
					Auth: &QuayAuthDefaults{
						Token: "quay-token",
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &RegistryDefaultsSet{
				GHCR: &GHCRRegistryDefaults{
					Auth: &GHCRAuthDefaults{
						Token: "ghcr-token",
					},
				},
				Hub: &HubRegistryDefaults{
					Auth: &HubAuthDefaults{
						Token: "hub-token",
					},
				},
				Quay: &QuayRegistryDefaults{
					Auth: &QuayAuthDefaults{
						Token: "quay-token",
					},
				},
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

// #############
// # STRINGIFY #
// #############

func TestDefaults_String(t *testing.T) {
	tests := []struct {
		name     string
		defaults *Defaults
		want     string
	}{
		{
			name:     "nil",
			defaults: nil,
			want:     "null\n",
		},
		{
			name:     "empty",
			defaults: &Defaults{},
			want:     "{}\n",
		},
		{
			name: "filled",
			defaults: &Defaults{
				Type: "ghcr",
				Tag:  "1.2.3",
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						Auth: &GHCRAuth{
							Token: "t1",
						},
					},
					Hub: &HubRegistryDefaults{
						Auth: &HubAuthDefaults{
							Username: "u1",
							Token:    "t2",
						},
					},
					Quay: &QuayRegistryDefaults{
						Auth: &QuayAuthDefaults{
							Token: "t1",
						},
					},
				},
				Defaults: &Defaults{
					Type: "ghcr",
					Tag:  "Bar",
				},
			},
			want: test.TrimYAML(`
				type: ghcr
				tag: 1.2.3
				registry:
					ghcr:
						auth:
							token: t1
					hub:
						auth:
							username: u1
							token: t2
					quay:
						auth:
							token: t1
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				"Defaults.String",
				tc.defaults.String,
				tc.want,
			)
		})
	}
}

// ############
// # DEFAULTS #
// ############

func TestDefaults_Default(t *testing.T) {
	tests := []struct {
		name string
		data *Defaults
	}{
		{
			name: "empty",
			data: &Defaults{},
		},
		{
			name: "filled",
			data: &Defaults{
				Type: "abc",
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						Auth: &GHCRAuthDefaults{},
					},
					Hub: &HubRegistryDefaults{
						Auth: &HubAuthDefaults{},
					},
					Quay: &QuayRegistryDefaults{
						Auth: &QuayAuthDefaults{},
					},
				},
				Tag: "t",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defaultType := "hub"
			defaultContainerDetailTag := "{{ version }}"

			tc.data.Default()

			if tc.data.Type != defaultType {
				t.Fatalf("Defaults.Type mismatch\ngot:  %q\nwant: %q", tc.data.Type, defaultType)
			}

			assertRegistriesNonNil(t, tc.data.Registry)

			if testErr := test.AssertFields(
				t,
				[]test.FieldAssertion{
					{Name: "Registry.ECR.GetAuth()", Got: tc.data.Registry.ECR.GetAuth(), Mode: test.CompareNotNil},
					{Name: "Registry.GHCR.GetAuth()", Got: tc.data.Registry.GHCR.GetAuth(), Mode: test.CompareNotNil},
					{Name: "Registry.Hub.GetAuth()", Got: tc.data.Registry.Hub.GetAuth(), Mode: test.CompareNotNil},
					{Name: "Registry.Quay.GetAuth()", Got: tc.data.Registry.Quay.GetAuth(), Mode: test.CompareNotNil},
				},
				"Defaults",
			); testErr != nil {
				t.Fatal(testErr)
			}

			if got := tc.data.Tag; got != defaultContainerDetailTag {
				t.Fatalf(
					"Defaults.Tag mismatch\ngot:  %q\nwant: %q",
					got, defaultContainerDetailTag,
				)
			}
		})
	}
}

func TestDefaults_SetDefaults(t *testing.T) {
	_, hardDefaults := plainDefaults(t)

	tests := []struct {
		name     string
		defaults *Defaults
	}{
		{
			name:     "nil",
			defaults: nil,
		},
		{
			name:     "empty",
			defaults: &Defaults{},
		},
		{
			name:     "non-empty",
			defaults: hardDefaults,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defaults, _ := plainDefaults(t)
			defaults.Defaults = nil

			defaults.SetDefaults(tc.defaults)

			if got := defaults.Defaults; got != tc.defaults {
				t.Fatalf("Defaults.Defaults mismatch\ngot:  %p\nwant: %p", got, tc.defaults)
			}
			if tc.defaults == nil {
				return
			}
			fieldTests := []test.FieldAssertion{
				{Name: "Registry.ECR.GetAuth().Defaults()", Got: defaults.Registry.ECR.GetAuth().Defaults(), Want: tc.defaults.Registry.ECR.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "Registry.GHCR.GetAuth().Defaults()", Got: defaults.Registry.GHCR.GetAuth().Defaults(), Want: tc.defaults.Registry.GHCR.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "Registry.Hub.GetAuth().Defaults()", Got: defaults.Registry.Hub.GetAuth().Defaults(), Want: tc.defaults.Registry.Hub.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "Registry.Quay.GetAuth().Defaults()", Got: defaults.Registry.Quay.GetAuth().Defaults(), Want: tc.defaults.Registry.Quay.GetAuth(), Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Defaults"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

// ##########
// # VALUES #
// ##########

func TestDefaults_GetType(t *testing.T) {
	tests := []struct {
		name string
		data *Defaults
		want string
	}{
		{
			name: "no type",
			data: &Defaults{},
			want: "",
		},
		{
			name: "root",
			data: &Defaults{
				Type: "abc",
				Defaults: &Defaults{
					Type: "def",
					Defaults: &Defaults{
						Type: "ghi",
					},
				},
			},
			want: "abc",
		},
		{
			name: "defaults",
			data: &Defaults{
				Type: "",
				Defaults: &Defaults{
					Type: "def",
					Defaults: &Defaults{
						Type: "ghi",
					},
				},
			},
			want: "def",
		},
		{
			name: "hard defaults",
			data: &Defaults{
				Type: "",
				Defaults: &Defaults{
					Type: "",
					Defaults: &Defaults{
						Type: "ghi",
					},
				},
			},
			want: "ghi",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.GetType()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

// ##############
// # VALIDATION #
// ##############

func TestDefaults_CheckValues(t *testing.T) {
	tests := []struct {
		name     string
		input    *Defaults
		errRegex string
	}{
		{
			name:     "nil",
			input:    (*Defaults)(nil),
			errRegex: `^$`,
		},
		{
			name: "valid",
			input: test.Must(t, func() (*Defaults, error) {
				return DecodeDefaults(
					"yaml", []byte("type: ghcr"),
					nil,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "invalid docker",
			input: test.Must(t, func() (*Defaults, error) {
				input, err := DecodeDefaults(
					"yaml", []byte("type: ghcr"),
					nil,
				)
				if err != nil {
					return nil, err
				}
				input.Type = "foo"
				return input, nil
			}),
			errRegex: `^type: .* <invalid>.*$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				tc.input.CheckValues,
			)
		})
	}
}

// #############
// # UTILITIES #
// #############

func TestDefaults_InitRegistries(t *testing.T) {
	var d Defaults

	d.initRegistries()

	assertRegistriesNonNil(t, d.Registry)
}

func TestGetRegistryDefaults(t *testing.T) {
	tests := []struct {
		name    string
		dType   string
		arrange func(d *Defaults) (want RegistryDefaults)
	}{
		{name: "ecr/populated", dType: "ecr", arrange: func(d *Defaults) RegistryDefaults { return d.Registry.ECR }},
		{name: "ecr/nil", dType: "ecr", arrange: func(d *Defaults) RegistryDefaults {
			d.Registry.ECR = nil
			return nil
		}},
		{name: "ghcr/populated", dType: "ghcr", arrange: func(d *Defaults) RegistryDefaults { return d.Registry.GHCR }},
		{name: "ghcr/nil", dType: "ghcr", arrange: func(d *Defaults) RegistryDefaults {
			d.Registry.GHCR = nil
			return nil
		}},
		{name: "hub/populated", dType: "hub", arrange: func(d *Defaults) RegistryDefaults { return d.Registry.Hub }},
		{name: "hub/nil", dType: "hub", arrange: func(d *Defaults) RegistryDefaults {
			d.Registry.Hub = nil
			return nil
		}},
		{name: "quay/populated", dType: "quay", arrange: func(d *Defaults) RegistryDefaults { return d.Registry.Quay }},
		{name: "quay/nil", dType: "quay", arrange: func(d *Defaults) RegistryDefaults {
			d.Registry.Quay = nil
			return nil
		}},
		{name: "unknown", dType: "foo", arrange: func(*Defaults) RegistryDefaults { return nil }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, defaults := plainDefaults(t)
			want := tc.arrange(defaults)

			// Siblings stay populated, so a wrong switch arm returns a non-nil sibling and fails.
			got := getRegistryDefaults(tc.dType, defaults)

			if got != want {
				t.Fatalf(
					"mismatch\ngot:  %s\nwant: %s",
					registryDefaultsString(got), registryDefaultsString(want),
				)
			}
		})
	}
}

func TestSetRegistryDefaults(t *testing.T) {
	tests := []struct {
		name            string
		registry        RegistryDefaults
		defaultRegistry RegistryDefaults
	}{
		{
			name:            "nil registry",
			registry:        nil,
			defaultRegistry: RegistryDefaultsMap["ghcr"](),
		},
		{
			name: "nil defaultRegistry keeps the existing link",
			registry: func() RegistryDefaults {
				registry := RegistryDefaultsMap["ghcr"]()
				registry.GetAuth().SetDefaults(RegistryDefaultsMap["ghcr"]().GetAuth())
				return registry
			}(),
			defaultRegistry: nil,
		},
		{
			name:            "ghcr",
			registry:        RegistryDefaultsMap["ghcr"](),
			defaultRegistry: RegistryDefaultsMap["ghcr"](),
		},
		{
			name:            "hub",
			registry:        RegistryDefaultsMap["hub"](),
			defaultRegistry: RegistryDefaultsMap["hub"](),
		},
		{
			name:            "quay",
			registry:        RegistryDefaultsMap["quay"](),
			defaultRegistry: RegistryDefaultsMap["quay"](),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var want RegistryAuthDefaults
			if tc.registry != nil {
				want = tc.registry.GetAuth().Defaults()
			}

			setRegistryDefaults(tc.registry, tc.defaultRegistry)

			if tc.registry == nil {
				return
			}

			if tc.defaultRegistry != nil {
				want = tc.defaultRegistry.GetAuth()
			}
			fieldTests := []test.FieldAssertion{
				{Name: "GetAuth().Defaults()", Got: tc.registry.GetAuth().Defaults(), Want: want, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "RegistryDefaults"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
