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

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/util/polymorphic"
)

func TestDecode(t *testing.T) {
	defaults, _ := plainDefaults(t)

	// GIVEN: data in a given format to Decode into a Registry.
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
				packageName,
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

func TestDecode_UnknownStructType(t *testing.T) {
	// GIVEN: RegistryMapInheritable has a non-Registry type.
	prevRegistryMapInheritable := RegistryMapInheritable
	t.Cleanup(func() {
		RegistryMapInheritable = prevRegistryMapInheritable
	})
	rMap := map[string]func() polymorphic.Inheritable{
		"test": func() polymorphic.Inheritable {
			return &unknownStruct{}
		},
	}
	RegistryMapInheritable = polymorphic.ToInheritableMap(rMap)

	defaults, _ := plainDefaults(t)

	// WHEN: We Decode() with data that resolves to this non-Registry type.
	data := "type: test"
	_, err := Decode(
		"yaml", []byte(data),
		defaults,
	)

	errRegex := test.TrimYAML(`
		^docker:
			expected Registry, got .*$`,
	)
	// THEN: We get an error.
	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Fatalf(
			"%s\nDecode(format=\"yaml\", data=%q) error mismatch:\ngot:  %q\nwant: %q",
			packageName, data,
			errRegex, e,
		)
	}
}

func TestApplyOverrides(t *testing.T) {
	defaults, _ := plainDefaults(t)

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
				"image": test/app,
			}`),
			previous: nil,
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
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "test/app",
						Tag:   "{{ version }}",
					},
				},
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
				CommonRegistry: CommonRegistry{
					ContainerDetail: ContainerDetail{
						Image: "test/app",
						Tag:   "{{ version }}",
					},
				},
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
				CommonRegistry: CommonRegistry{
					Type: "ecr",
					ContainerDetail: ContainerDetail{
						Image: "something",
						Tag:   "else",
					},
				},
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
				CommonRegistry: CommonRegistry{
					Type: "ecr",
					ContainerDetail: ContainerDetail{
						Image: "something",
						Tag:   "else",
					},
				},
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
				CommonRegistry: CommonRegistry{
					Type: "ecr",
					ContainerDetail: ContainerDetail{
						Image: "something",
						Tag:   "else",
					},
				},
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
				CommonRegistry: CommonRegistry{
					Type: "ghcr",
					ContainerDetail: ContainerDetail{
						Image: "something",
						Tag:   "else",
					},
				},
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
				CommonRegistry: CommonRegistry{
					Type: "ghcr",
					ContainerDetail: ContainerDetail{
						Image: "something",
						Tag:   "else",
					},
				},
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
				CommonRegistry: CommonRegistry{
					Type: "ghcr",
					ContainerDetail: ContainerDetail{
						Image: "something",
						Tag:   "else",
					},
				},
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
				CommonRegistry: CommonRegistry{
					Type: "hub",
					ContainerDetail: ContainerDetail{
						Image: "something",
						Tag:   "else",
					},
				},
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
				"image": "test/app-ghcr",
			}`),
			previous: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Type: "hub",
					ContainerDetail: ContainerDetail{
						Image: "test/app-hub",
						Tag:   "{{ version }}",
					},
				},
			},
			want: test.TrimYAML(`
				type: ghcr
				image: test/app-ghcr
			`),
		},
		{
			name:   "hub -> quay",
			format: "json",
			data: test.TrimJSON(`{
				"type": "ghcr",
				"image": "test/app-quay",
			}`),
			previous: &HubRegistry{
				CommonRegistry: CommonRegistry{
					Type: "hub",
					ContainerDetail: ContainerDetail{
						Image: "test/app-hub",
						Tag:   "{{ version }}",
					},
				},
			},
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
				CommonRegistry: CommonRegistry{
					Type: "quay",
					ContainerDetail: ContainerDetail{
						Image: "something",
						Tag:   "else",
					},
				},
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
				"image": "test/app-ghcr",
			}`),
			previous: &QuayRegistry{
				CommonRegistry: CommonRegistry{
					Type: "quay",
					ContainerDetail: ContainerDetail{
						Image: "test/app-quay",
						Tag:   "{{ version }}",
					},
				},
			},
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
				"image": "test/app-hub",
			}`),
			previous: &QuayRegistry{
				CommonRegistry: CommonRegistry{
					Type: "quay",
					ContainerDetail: ContainerDetail{
						Image: "test/app-quay",
						Tag:   "{{ version }}",
					},
				},
			},
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
				CommonRegistry: CommonRegistry{
					Type: "hub",
					ContainerDetail: ContainerDetail{
						Image: "test/app",
						Tag:   "{{ version }}",
					},
				},
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
				CommonRegistry: CommonRegistry{
					Type: "hub",
					ContainerDetail: ContainerDetail{
						Image: "test/app",
						Tag:   "{{ version }}",
					},
				},
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
					v.Auth = RegistryMapInheritable["ecr"]().(Registry).GetAuth()
				case *GHCRRegistry:
					v.Auth = RegistryMapInheritable["ghcr"]().(Registry).GetAuth()
				case *HubRegistry:
					v.Auth = RegistryMapInheritable["hub"]().(Registry).GetAuth()
				case *QuayRegistry:
					v.Auth = RegistryMapInheritable["quay"]().(Registry).GetAuth()
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
				packageName,
				"ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestApplyOverrides_UnknownStructType(t *testing.T) {
	// GIVEN: RegistryMapInheritable has a non-Registry type.
	prevRegistryMapInheritable := RegistryMapInheritable
	t.Cleanup(func() {
		RegistryMapInheritable = prevRegistryMapInheritable
	})
	rMap := map[string]func() polymorphic.Inheritable{
		"test": func() polymorphic.Inheritable {
			return &unknownStruct{}
		},
	}
	RegistryMapInheritable = polymorphic.ToInheritableMap(rMap)

	defaults, _ := plainDefaults(t)

	// WHEN: We ApplyOverrides() with data that resolves to this non-Registry type.
	data := "type: test"
	_, err := ApplyOverrides(
		"yaml", []byte(data),
		nil,
		defaults,
	)

	errRegex := test.TrimYAML(`
		^docker:
			expected Registry, got .*$`,
	)
	// THEN: We get an error.
	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Fatalf(
			"%s\nApplyOverrides(format=\"yaml\", data=%q) error mismatch:\ngot:  %q\nwant: %q",
			packageName, data,
			errRegex, e,
		)
	}
}
