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

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/util/polymorphic"
)

func TestDecode(t *testing.T) {
	defaults, _ := plainDefaults()

	tests := []struct {
		name         string
		format, data string
		errRegex     string
		want         string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     `{}`,
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
			name:     "JSON/null",
			format:   "json",
			data:     "null",
			errRegex: `^$`,
			want:     "",
		},
		{
			name:     "YAML/null",
			format:   "yaml",
			data:     "null",
			errRegex: `^$`,
			want:     "",
		},
		{
			name:   "JSON/invalid payload decode error",
			format: "json",
			data:   `{`,
			errRegex: test.TrimYAML(`
				^docker:
					[^\s]+ unexpected EOF`,
			),
		},
		{
			name:   "JSON/docker extraction, invalid data type",
			format: "json",
			data:   `{"type": 123}`,
			errRegex: test.TrimYAML(`
				^docker:
					json: .*unmarshal.* number.*$`,
			),
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
					"type": "hub",
					"image": "test/app",
					"tag": "{{ version }}"
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: hub
				image: test/app
				tag: '{{ version }}'
			`),
		},
		{
			name:   "JSON/static fields unmarshal fail",
			format: "json",
			data:   `{"image": ["-"]}`,
			errRegex: test.TrimYAML(`
				^docker:
					json: .*unmarshal.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (Registry, error) {
					return Decode(
						format, data,
						defaults,
					)
				},
				tc.format, tc.data,
				func(v Registry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"Decode",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

type unknownStruct struct {
	polymorphic.Inheritable
	Type string `json:"type" yaml:"type"`
}

func (u *unknownStruct) GetType() string {
	return u.Type
}

// withUnknownRegistryType swaps [registryMapInheritable] for one whose "test" type is not a [Registry].
func withUnknownRegistryType(t *testing.T) {
	prevRegistryMapInheritable := registryMapInheritable
	t.Cleanup(func() {
		registryMapInheritable = prevRegistryMapInheritable
	})
	registryMapInheritable = polymorphic.ToInheritableMap(map[string]func() polymorphic.Inheritable{
		"test": func() polymorphic.Inheritable {
			return &unknownStruct{}
		},
	})
}

func TestDecode_UnknownStructType(t *testing.T) {
	withUnknownRegistryType(t)

	defaults, _ := plainDefaults()

	errRegex := test.TrimYAML(`
		^docker:
			expected Registry, got .*$`,
	)

	_, err := Decode("yaml", []byte("type: test"), defaults)
	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestApplyOverrides(t *testing.T) {
	defaults, _ := plainDefaults()

	tests := []struct {
		name        string
		format      string
		data        string
		previous    Registry
		errRegex    string
		sameAddress bool
		want        string
	}{
		{
			name:   "New Registry",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ghcr",
				"image": "test/app"
			}`),
			previous: nil,
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: ghcr
				image: test/app
			`),
		},
		{
			name:   "empty data returns previous",
			format: "json",
			data:   "",
			previous: &GHCRRegistry{
				Image: "test/app",
				Tag:   "{{ version }}",
			},
			sameAddress: true,
			errRegex:    `^$`,
			want: test.TrimYAML(`
				image: test/app
				tag: '{{ version }}'
			`),
		},
		{
			name:   "null data omits previous",
			format: "json",
			data:   `null`,
			previous: &GHCRRegistry{
				Image: "test/app",
				Tag:   "{{ version }}",
			},
			errRegex: `^$`,
			want:     "",
		},
		{
			name:     "invalid payload causes decode error",
			format:   "json",
			data:     `{"`,
			previous: &GHCRRegistry{},
			errRegex: test.TrimYAML(`
				^docker:
					[^\s]+ unexpected EOF`,
			),
		},
		{
			name:     "docker extraction, invalid data type",
			format:   "json",
			data:     `{"type": 123}`,
			previous: &GHCRRegistry{},
			errRegex: test.TrimYAML(`
				^docker:
					json: .*unmarshal.* number .*`,
			),
		},
		{
			name:   "ecr -> ghcr",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ghcr",
				"image": "test/app-ghcr",
				"tag": "{{ version }}"
			}`),
			previous: &ECRRegistry{
				Type:  "ecr",
				Image: "something",
				Tag:   "else",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: ghcr
				image: test/app-ghcr
				tag: '{{ version }}'
			`),
		},
		{
			name:   "ecr -> hub",
			format: "json",
			data: test.TrimJSON(`{
				"type": "hub",
				"image": "test/app-hub",
				"tag": "{{ version }}"
			}`),
			previous: &ECRRegistry{
				Type:  "ecr",
				Image: "something",
				Tag:   "else",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: hub
				image: test/app-hub
				tag: '{{ version }}'
			`),
		},
		{
			name:   "ecr -> quay",
			format: "json",
			data: test.TrimJSON(`{
				"type": "quay",
				"image": "test/app-quay",
				"tag": "{{ version }}"
			}`),
			previous: &ECRRegistry{
				Type:  "ecr",
				Image: "something",
				Tag:   "else",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: quay
				image: test/app-quay
				tag: '{{ version }}'
			`),
		},
		{
			name:   "ghcr -> ecr",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ecr",
				"image": "test/app-ecr",
				"tag": "{{ version }}"
			}`),
			previous: &GHCRRegistry{
				Type:  "ghcr",
				Image: "something",
				Tag:   "else",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: ecr
				image: test/app-ecr
				tag: '{{ version }}'
			`),
		},
		{
			name:   "ghcr -> hub",
			format: "json",
			data: test.TrimJSON(`{
				"type": "hub",
				"image": "test/app-hub",
				"tag": "{{ version }}"
			}`),
			previous: &GHCRRegistry{
				Type:  "ghcr",
				Image: "something",
				Tag:   "else",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: hub
				image: test/app-hub
				tag: '{{ version }}'
			`),
		},
		{
			name:   "ghcr -> quay",
			format: "json",
			data: test.TrimJSON(`{
				"type": "quay",
				"image": "test/app-quay",
				"tag": "{{ version }}"
			}`),
			previous: &GHCRRegistry{
				Type:  "ghcr",
				Image: "something",
				Tag:   "else",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: quay
				image: test/app-quay
				tag: '{{ version }}'
			`),
		},
		{
			name:   "hub -> ecr",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ecr",
				"image": "test/app-ecr",
				"tag": "{{ version }}"
			}`),
			previous: &HubRegistry{
				Type:  "hub",
				Image: "something",
				Tag:   "else",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: ecr
				image: test/app-ecr
				tag: '{{ version }}'
			`),
		},
		{
			name:   "hub -> ghcr",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ghcr",
				"image": "test/app-ghcr"
			}`),
			previous: &HubRegistry{
				Type:  "hub",
				Image: "test/app-hub",
				Tag:   "{{ version }}",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: ghcr
				image: test/app-ghcr
			`),
		},
		{
			name:   "hub -> quay",
			format: "json",
			data: test.TrimJSON(`{
				"type": "quay",
				"image": "test/app-quay"
			}`),
			previous: &HubRegistry{
				Type:  "hub",
				Image: "test/app-hub",
				Tag:   "{{ version }}",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: quay
				image: test/app-quay
			`),
		},
		{
			name:   "quay -> ecr",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ecr",
				"image": "test/app-ecr",
				"tag": "{{ version }}"
			}`),
			previous: &QuayRegistry{
				Type:  "quay",
				Image: "something",
				Tag:   "else",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: ecr
				image: test/app-ecr
				tag: '{{ version }}'
			`),
		},
		{
			name:   "quay -> ghcr",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ghcr",
				"image": "test/app-ghcr"
			}`),
			previous: &QuayRegistry{
				Type:  "quay",
				Image: "test/app-quay",
				Tag:   "{{ version }}",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: ghcr
				image: test/app-ghcr
			`),
		},
		{
			name:   "quay -> hub",
			format: "json",
			data: test.TrimJSON(`{
				"type": "hub",
				"image": "test/app-hub"
			}`),
			previous: &QuayRegistry{
				Type:  "quay",
				Image: "test/app-quay",
				Tag:   "{{ version }}",
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: hub
				image: test/app-hub
			`),
		},
		{
			name:   "image-only override keeps existing tag and type",
			format: "json",
			data: test.TrimJSON(`{
				"image": "test/other"
			}`),
			previous: &HubRegistry{
				Type:  "hub",
				Image: "test/app",
				Tag:   "{{ version }}",
			},
			sameAddress: true,
			errRegex:    `^$`,
			want: test.TrimYAML(`
				type: hub
				image: test/other
				tag: '{{ version }}'
			`),
		},
		{
			name:   "tag-only override keeps existing image and type",
			format: "json",
			data: test.TrimJSON(`{
				"tag": "1.2.3"
			}`),
			previous: &HubRegistry{
				Type:  "hub",
				Image: "test/app",
				Tag:   "{{ version }}",
			},
			sameAddress: true,
			errRegex:    `^$`,
			want: test.TrimYAML(`
				type: hub
				image: test/app
				tag: 1.2.3
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.previous != nil {
				switch v := tc.previous.(type) {
				case *ECRRegistry:
					v.Auth = registryMapInheritable["ecr"]().(Registry).GetAuth()
				case *GHCRRegistry:
					v.Auth = registryMapInheritable["ghcr"]().(Registry).GetAuth()
				case *HubRegistry:
					v.Auth = registryMapInheritable["hub"]().(Registry).GetAuth()
				case *QuayRegistry:
					v.Auth = registryMapInheritable["quay"]().(Registry).GetAuth()
				}
				tc.previous.SetDefaults(tc.previous.GetType(), defaults)
			}

			if _, _, testErr := test.AssertApplyOverrides(
				t,
				tc.previous,
				func(format string, data []byte, v Registry) (Registry, error) {
					return ApplyOverrides(format, data, v, defaults)
				},
				tc.format, tc.data,
				func(v Registry) string { return v.String("") },
				tc.want,
				tc.errRegex,
				tc.sameAddress,
				"ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestApplyOverrides_QueryToken(t *testing.T) {
	const (
		queryToken = "query-token"
		ecrInit    = `{"type":"ecr", "image":"a/b"}`
		hubInit    = `{"type":"hub", "image":"a/b","auth":{"username":"u","token":"t"}}`
		ghcrInit   = `{"type":"ghcr","image":"a/b"}`
	)

	tests := []struct {
		name     string
		initial  string
		override string
		want     string
	}{
		{
			name:     "hub/unchanged credentials keep the token",
			initial:  hubInit,
			override: `{"image":"c/d"}`,
			want:     queryToken,
		},
		{
			name:     "hub/secret placeholder keeps the token",
			initial:  hubInit,
			override: `{"auth":{"token":"` + util.SecretValue + `"}}`,
			want:     queryToken,
		},
		{
			name:     "hub/changed token clears the token",
			initial:  hubInit,
			override: `{"auth":{"token":"t2"}}`,
			want:     "",
		},
		{
			name:     "hub/changed username clears the token",
			initial:  hubInit,
			override: `{"auth":{"username":"u2"}}`,
			want:     "",
		},
		{
			name:     "ghcr/unchanged credentials and image keep the token",
			initial:  ghcrInit,
			override: `{"tag":"{{ version }}"}`,
			want:     queryToken,
		},
		{
			name:     "ghcr/changed image clears the token",
			initial:  ghcrInit,
			override: `{"image":"c/d"}`,
			want:     "",
		},
		{
			name:     "ghcr/changed token is not inherited",
			initial:  ghcrInit,
			override: `{"auth":{"token":"t2"}}`,
			want:     "t2", // Decoding the new token derives the query token from it.
		},
		{
			name:     "ecr/anonymous token is always kept",
			initial:  ecrInit,
			override: `{"image":"c/d"}`,
			want:     queryToken,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defaults, _ := plainDefaults()
			target, err := Decode("json", []byte(tc.initial), defaults)
			if err != nil {
				t.Fatalf("Decode() failed: %v", err)
			}
			target.GetAuth().SetQueryToken(queryToken, time.Now().Add(time.Hour))

			got, err := ApplyOverrides("json", []byte(tc.override), target, defaults)
			if err != nil {
				t.Fatalf("ApplyOverrides() failed: %v", err)
			}

			gotToken, _ := got.GetAuth().GetQueryTokenSelf()
			if gotToken != tc.want {
				t.Errorf(
					"Registry.GetAuth().GetQueryTokenSelf() mismatch\ngot:  %q\nwant: %q",
					gotToken, tc.want,
				)
			}
		})
	}
}

func TestApplyOverrides_SharedDefaultsQueryToken(t *testing.T) {
	const sharedToken = "shared-token"

	tests := []struct {
		name     string
		initial  string
		override string
	}{
		{
			name:     "hub/unchanged credentials",
			initial:  `{"type":"hub","image":"a/b"}`,
			override: `{"image":"c/d"}`,
		},
		{
			name:     "hub/changed token",
			initial:  `{"type":"hub","image":"a/b"}`,
			override: `{"auth":{"token":"t2"}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defaults, _ := plainDefaults()
			sharedAuth := defaults.Registry.Hub.GetAuth()
			hub, ok := sharedAuth.(*HubAuthDefaults)
			if !ok {
				t.Fatalf(
					"Registry.Hub.GetAuth() type mismatch\ngot:  %T\nwant: *HubAuthDefaults",
					sharedAuth,
				)
			}
			hub.Username, hub.Token = "u", "t"
			target, err := Decode("json", []byte(tc.initial), defaults)
			if err != nil {
				t.Fatalf("Decode() failed: %v", err)
			}
			sharedAuth.SetQueryToken(sharedToken, time.Now().Add(time.Hour))

			if _, err := ApplyOverrides("json", []byte(tc.override), target, defaults); err != nil {
				t.Fatalf("ApplyOverrides() failed: %v", err)
			}

			gotToken, _ := sharedAuth.GetQueryTokenSelf()
			if gotToken != sharedToken {
				t.Errorf(
					"Defaults.Registry.Hub.GetAuth().GetQueryTokenSelf() mismatch\n"+
						"got:  %q\nwant: %q",
					gotToken, sharedToken,
				)
			}
		})
	}
}

func TestApplyOverrides_UnknownStructType(t *testing.T) {
	withUnknownRegistryType(t)

	defaults, _ := plainDefaults()

	errRegex := test.TrimYAML(`
		^docker:
			expected Registry, got .*$`,
	)

	_, err := ApplyOverrides("yaml", []byte("type: test"), nil, defaults)
	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}
