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

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// ############
// # DECODING #
// ############

func TestDecodeDefaults(t *testing.T) {
	// GIVEN: data in a given format to Decode into Defaults.
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
			want: "",
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
							token: ghcr-secret
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
							token: ghcr-secret
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
					token: ghcr-secret
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
							token: ghcr-secret
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
					ghcr": ["a", "b", "c"]
				}
			}`),
			errRegex: test.TrimYAML(`
				^docker:
					jsontext: invalid character.*`,
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

			// AND: we have defaults for each registry type.
			var defaults Defaults
			defaults.Default()
			defaults.Type = ""
			defaults.Tag = ""

			// WHEN: DecodeDefaults is called.
			got, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Defaults, error) {
					return DecodeDefaults(format, data, &defaults)
				},
				tc.format, tc.data,
				func(v *Defaults) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"DecodeDefaults",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nDecodeDefaults(format=%q, data=%q)",
				packageName, tc.format, tc.data,
			)

			// THEN: The Defaults were passed over correctly.
			fieldTests := []test.FieldAssertion{
				{Name: "Defaults", Got: got.Defaults, Want: &defaults, Mode: test.CompareSamePointer},
				{Name: "ECR.Auth.Defaults", Got: got.Registry.ECR.GetAuth().Defaults(), Want: defaults.Registry.ECR.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "GHCR.Auth.Defaults", Got: got.Registry.GHCR.GetAuth().Defaults(), Want: defaults.Registry.GHCR.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "Hub.Auth.Defaults", Got: got.Registry.Hub.GetAuth().Defaults(), Want: defaults.Registry.Hub.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "Quay.Auth.Defaults", Got: got.Registry.Quay.GetAuth().Defaults(), Want: defaults.Registry.Quay.GetAuth(), Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Defaults"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDefaults_MarshalYAML(t *testing.T) {
	// GIVEN: a Defaults.
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
				ContainerDetailDefaults: ContainerDetailDefaults{
					Tag: "1.2.3",
				},
			},
			want: test.TrimYAML(`
				type: ` + PossibleTypes[0] + `
				tag: 1.2.3
			`),
		},
		{
			name: "dynamic fields",
			defaults: &Defaults{
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &GHCRAuth{
								GHCRAuthDefaults: GHCRAuthDefaults{
									Token: "ghcr-token",
								},
							},
						},
					},
					Hub: &HubRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &HubAuthDefaults{
								Username: "hub-username",
								Token:    "hub-token",
							},
						},
					},
					Quay: &QuayRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &QuayAuthDefaults{
								Token: "quay-token",
							},
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
		},
		{
			name: "filled",
			defaults: &Defaults{
				Type: PossibleTypes[0],
				ContainerDetailDefaults: ContainerDetailDefaults{
					Tag: "1.2.3",
				},
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &GHCRAuth{
								GHCRAuthDefaults: GHCRAuthDefaults{
									Token: "ghcr-token",
								},
							},
						},
					},
					Hub: &HubRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &HubAuthDefaults{
								Username: "hub-username",
								Token:    "hub-token",
							},
						},
					},
					Quay: &QuayRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &QuayAuthDefaults{
								Token: "quay-token",
							},
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
			want: "type: " + PossibleTypes[0] + "\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Marshaled to YAML.
			gotYAML, err := decode.Marshal("yaml", tc.defaults)

			prefix := fmt.Sprintf("%s\nDefaults.MarshalYAML()", packageName)

			// AND: the error is as expected.
			e := errfmt.FormatError(err)
			if util.RegexCheck(tc.errRegex, e) == false {
				t.Errorf(
					"%s error mismatch:\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: The YAML matches.
			if got := string(gotYAML); got != tc.want {
				t.Errorf(
					"%s value mismatch:\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

// #########
// # STATE #
// #########

func TestDefaults_IsZero(t *testing.T) {
	// GIVEN: a Defaults.
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
				ContainerDetailDefaults: ContainerDetailDefaults{
					Tag: "a",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Registry",
			data: &Defaults{
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &GHCRAuthDefaults{
								Token: "a",
							},
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

			// WHEN: IsZero() is called on it.
			got := tc.data.IsZero()

			// THEN: the expected result is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRegistryDefaults_IsZero(t *testing.T) {
	// GIVEN: a registryDefaults.
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
					CommonRegistryDefaults: CommonRegistryDefaults{
						Auth: &GHCRAuthDefaults{
							Token: "ghcr-token",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Hub",
			data: &RegistryDefaultsSet{
				Hub: &HubRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						Auth: &HubAuthDefaults{
							Token: "hub-token",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Quay",
			data: &RegistryDefaultsSet{
				Quay: &QuayRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						Auth: &QuayAuthDefaults{
							Token: "quay-token",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &RegistryDefaultsSet{
				GHCR: &GHCRRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						Auth: &GHCRAuthDefaults{
							Token: "ghcr-token",
						},
					},
				},
				Hub: &HubRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						Auth: &HubAuthDefaults{
							Token: "hub-token",
						},
					},
				},
				Quay: &QuayRegistryDefaults{
					CommonRegistryDefaults: CommonRegistryDefaults{
						Auth: &QuayAuthDefaults{
							Token: "quay-token",
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

			// WHEN: IsZero() is called on it.
			got := tc.data.IsZero()

			// THEN: the expected result is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nregistryDefaults.IsZero() values mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestDefaults_String(t *testing.T) {
	// GIVEN: a registryDefaults.
	tests := []struct {
		name      string
		rDefaults *Defaults
		want      string
	}{
		{
			name:      "nil",
			rDefaults: nil,
			want:      "null\n",
		},
		{
			name:      "empty",
			rDefaults: &Defaults{},
			want:      "{}\n",
		},
		{
			name: "filled",
			rDefaults: &Defaults{
				Type: "ghcr",
				ContainerDetailDefaults: ContainerDetailDefaults{
					Tag: "1.2.3",
				},
				Registry: RegistryDefaultsSet{
					GHCR: &GHCRRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &GHCRAuth{
								GHCRAuthDefaults: GHCRAuthDefaults{
									Token: "t1",
								},
							},
						},
					},
					Hub: &HubRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &HubAuthDefaults{
								Username: "u1",
								Token:    "t2",
							},
						},
					},
					Quay: &QuayRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &QuayAuthDefaults{
								Token: "t1",
							},
						},
					},
				},
				Defaults: &Defaults{
					Type: "ghcr",
					ContainerDetailDefaults: ContainerDetailDefaults{
						Tag: "Bar",
					},
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
				packageName,
				tc.rDefaults.String,
				tc.want,
			)
		})
	}
}

// ############
// # DEFAULTS #
// ############

func TestDefaults_Default(t *testing.T) {
	// GIVEN: a Defaults.
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
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &GHCRAuthDefaults{},
						},
					},
					Hub: &HubRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &HubAuthDefaults{},
						},
					},
					Quay: &QuayRegistryDefaults{
						CommonRegistryDefaults: CommonRegistryDefaults{
							Auth: &QuayAuthDefaults{},
						},
					},
				},
				ContainerDetailDefaults: ContainerDetailDefaults{
					Tag: "t",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defaultType := "hub"
			defaultContainerDetailTag := "{{ version }}"

			// WHEN: Default() is called on it.
			tc.data.Default()

			prefix := fmt.Sprintf("%s\nDefaults.Default()", packageName)

			// THEN: the Type is set to the default value.
			if tc.data.Type != defaultType {
				t.Fatalf(
					"%s .Type value mismatch\ngot:  %q\nwant: %q",
					prefix, tc.data.Type, defaultType,
				)
			}

			// AND: the registries are non-nil.
			if tc.data.Registry.GHCR == nil ||
				tc.data.Registry.Hub == nil ||
				tc.data.Registry.Quay == nil {
				t.Fatalf(
					"%s got 1+ nil registries\ngot:  %+v\nwant: non-nil registries",
					prefix, tc.data.Registry,
				)
			}

			// AND: the registries have Auth's set.
			if tc.data.Registry.GHCR.GetAuth() == nil ||
				tc.data.Registry.Hub.GetAuth() == nil ||
				tc.data.Registry.Quay.GetAuth() == nil {
				t.Fatalf(
					"%s got 1+ nil .Registry.X.Auth\ngot:  %+v\nwant: non-nil auth on registries",
					prefix, tc.data.Registry,
				)
			}

			// AND: the ContainerDetailDefaults Tag is defaulted.
			if got := tc.data.ContainerDetailDefaults.Tag; got != defaultContainerDetailTag {
				t.Fatalf(
					"%s .ContainerDetailDefaults.Tag value mismatch\ngot:  %q\nwant: %q",
					prefix, got, defaultContainerDetailTag,
				)
			}
		})
	}
}

func TestDefaults_Defaults(t *testing.T) {
	_, hardDefaults := plainDefaults(t)
	// GIVEN: a Defaults.
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

			// AND: Those defaults have no Defaults.
			defaults, _ := plainDefaults(t)
			defaults.Defaults = nil

			// WHEN: Defaults is accessed on it before it is set.
			got := defaults.Defaults

			// THEN: nil is returned from Defaults.
			if got != nil {
				t.Errorf("%s\nfresh Defaults, .Defaults is non-nil", packageName)
			}

			// WHEN: SetDefaults is called on it.
			defaults.SetDefaults(tc.defaults)

			prefix := fmt.Sprintf(
				"%s\nDefaults.SetDefaults(type=%+v)",
				packageName, tc.defaults,
			)

			// THEN: Those Defaults are been set.
			if got := defaults.Defaults; got != tc.defaults {
				t.Fatalf(
					"%s mismatch on .Defaults\ngot:  %v\nwant: %v",
					prefix, got, tc.defaults,
				)
			}
			if tc.defaults == nil {
				return
			}
			fieldTests := []test.FieldAssertion{
				{Name: "ECR.Auth.Defaults", Got: defaults.Registry.ECR.GetAuth().Defaults(), Want: tc.defaults.Registry.ECR.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "GHCR.Auth.Defaults", Got: defaults.Registry.GHCR.GetAuth().Defaults(), Want: tc.defaults.Registry.GHCR.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "Hub.Auth.Defaults", Got: defaults.Registry.Hub.GetAuth().Defaults(), Want: tc.defaults.Registry.Hub.GetAuth(), Mode: test.CompareSamePointer},
				{Name: "Quay.Auth.Defaults", Got: defaults.Registry.Quay.GetAuth().Defaults(), Want: tc.defaults.Registry.Quay.GetAuth(), Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Defaults"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// ##########
// # VALUES #
// ##########

func TestDefaults_GetType(t *testing.T) {
	// GIVEN: a Defaults.
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

			// WHEN: GetType is called on it.
			got := tc.data.GetType()

			// THEN: the type is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nDefaults.GetType() mismatch\ngot %q, want %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

// ##############
// # VALIDATION #
// ##############

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN: a Defaults.
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
				input.Type = "foo"
				return input, err
			}),
			errRegex: `^type: .* <invalid>.*$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
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
	// GIVEN: a fresh Defaults.
	var d Defaults

	// THEN: all registries are nil
	if d.Registry.ECR != nil ||
		d.Registry.GHCR != nil ||
		d.Registry.Hub != nil ||
		d.Registry.Quay != nil {
		t.Fatalf(
			"%s\nfresh Defaults\ngot: non-nil registry %+v\nwant: all nil",
			packageName, d.Registry,
		)
	}

	// WHEN: initRegistries is called on it.
	d.initRegistries()

	// THEN: all registries are initialised
	if d.Registry.ECR == nil ||
		d.Registry.GHCR == nil ||
		d.Registry.Hub == nil ||
		d.Registry.Quay == nil {
		t.Fatalf(
			"%s\nDefaults.initRegistries() didn't initialise all registries\ngot: %+v",
			packageName, d.Registry,
		)
	}
}

func TestGetRegistryDefaults(t *testing.T) {
	// GIVEN: a registry type.
	tests := []struct {
		name  string
		dType string
	}{
		{
			name:  "known: ecr",
			dType: "ecr",
		},
		{
			name:  "known: ghcr",
			dType: "ghcr",
		},
		{
			name:  "known: hub",
			dType: "hub",
		},
		{
			name:  "known: quay",
			dType: "quay",
		},
		{
			name:  "unknown",
			dType: "foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Defaults struct.
			_, defaults := plainDefaults(t)

			var want RegistryDefaults
			switch tc.dType {
			case "ecr":
				want = defaults.Registry.ECR
			case "ghcr":
				want = defaults.Registry.GHCR
			case "hub":
				want = defaults.Registry.Hub
			case "quay":
				want = defaults.Registry.Quay
			}

			// WHEN: getRegistryDefaults is called with these.
			got := getRegistryDefaults(tc.dType, defaults)

			prefix := fmt.Sprintf(
				"%s\ngetRegistryDefaults(type=%q, defaults=%+v)",
				packageName, tc.dType, defaults,
			)

			// THEN: the correct registry defaults are returned.
			if got != want {
				t.Fatalf(
					"%s pointer mismatch\ngot:  %p\nwant: %p",
					prefix, got, want,
				)
			}
		})
	}
}

func TestGetRegistryDefaults_NilDefaults(t *testing.T) {
	// GIVEN: a registry type.
	tests := []struct {
		name  string
		dType string
	}{
		{
			name:  "known: ecr",
			dType: "ecr",
		},
		{
			name:  "known: ghcr",
			dType: "ghcr",
		},
		{
			name:  "known: hub",
			dType: "hub",
		},
		{
			name:  "known: quay",
			dType: "quay",
		},
		{
			name:  "unknown",
			dType: "foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Defaults struct.
			_, defaults := plainDefaults(t)

			switch tc.dType {
			case "ecr":
				defaults.Registry.ECR = nil
			case "ghcr":
				defaults.Registry.GHCR = nil
			case "hub":
				defaults.Registry.Hub = nil
			case "quay":
				defaults.Registry.Quay = nil
			}

			// WHEN: getRegistryDefaults is called with these.
			got := getRegistryDefaults(tc.dType, defaults)

			prefix := fmt.Sprintf(
				"%s\ngetRegistryDefaults(type=%q, defaults=%+v)",
				packageName, tc.dType, defaults,
			)

			// THEN: the nil registry defaults are returned.
			if got != nil {
				t.Fatalf(
					"%s pointer mismatch\ngot:  %p\nwant: nil",
					prefix, got,
				)
			}
		})
	}
}

func TestSetRegistryDefaults(t *testing.T) {
	// GIVEN: a registry type.
	tests := []struct {
		name            string
		dType           string
		registry        RegistryDefaults
		defaultRegistry RegistryDefaults
	}{
		{
			name:            "nil registry",
			registry:        nil,
			defaultRegistry: RegistryDefaultsMap["ghcr"](),
		},
		{
			name:            "nil defaultRegistry",
			registry:        RegistryDefaultsMap["ghcr"](),
			defaultRegistry: nil,
		},
		{
			name:            "ghcr",
			dType:           "ghcr",
			registry:        RegistryDefaultsMap["ghcr"](),
			defaultRegistry: RegistryDefaultsMap["ghcr"](),
		},
		{
			name:            "hub",
			dType:           "hub",
			registry:        RegistryDefaultsMap["hub"](),
			defaultRegistry: RegistryDefaultsMap["hub"](),
		},
		{
			name:            "quay",
			dType:           "quay",
			registry:        RegistryDefaultsMap["quay"](),
			defaultRegistry: RegistryDefaultsMap["quay"](),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: setRegistryDefaults is called with these.
			setRegistryDefaults(tc.registry, tc.defaultRegistry)

			if tc.registry == nil || tc.defaultRegistry == nil {
				return
			}
			prefix := fmt.Sprintf(
				"%s\nsetRegistryDefaults(registry=%p, defaultRegistry=%p)",
				packageName, tc.registry, tc.defaultRegistry,
			)

			// THEN: the auth defaults have been linked as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Auth.Defaults", Got: tc.registry.GetAuth().Defaults(), Want: tc.defaultRegistry.GetAuth(), Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, prefix, ""); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
