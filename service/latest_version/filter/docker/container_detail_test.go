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
)

// #########
// # STATE #
// #########

func TestContainerDetail_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data ContainerDetail
		want bool
	}{
		{
			name: "empty",
			data: ContainerDetail{},
			want: true,
		},
		{
			name: "non-empty/Image",
			data: ContainerDetail{
				Image: "i",
			},
			want: false,
		},
		{
			name: "non-empty/Tag",
			data: ContainerDetail{
				Tag: "t",
			},
			want: false,
		},
		{
			name: "non-empty/all",
			want: false,
			data: ContainerDetail{
				Image: "i",
				Tag:   "t",
			},
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

func TestContainerDetailDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data *ContainerDetailDefaults
		want bool
	}{
		{
			name: "nil",
			data: nil,
			want: true,
		},
		{
			name: "empty",
			data: &ContainerDetailDefaults{},
			want: true,
		},
		{
			name: "non-empty/Tag",
			data: &ContainerDetailDefaults{Tag: "t"},
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

func TestContainerDetail_Copy(t *testing.T) {
	tests := []struct {
		name string
		data ContainerDetail
	}{
		{
			name: "empty",
			data: ContainerDetail{},
		},
		{
			name: "Image and Tag",
			data: ContainerDetail{
				Image: "i",
				Tag:   "t",
			},
		},
		{
			name: "Image, Tag and Defaults",
			data: ContainerDetail{
				Image:    "i",
				Tag:      "t",
				Defaults: &ContainerDetailDefaults{Tag: "dt"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.Copy()

			if got.Image != tc.data.Image {
				t.Fatalf(
					"ContainerDetail.Image mismatch\ngot:  %q\nwant: %q",
					got.Image, tc.data.Image,
				)
			}

			if got.Tag != tc.data.Tag {
				t.Fatalf("ContainerDetail.Tag mismatch\ngot:  %q\nwant: %q", got.Tag, tc.data.Tag)
			}

			if got.Defaults != tc.data.Defaults {
				t.Fatalf(
					"ContainerDetail.Defaults mismatch\ngot:  %p\nwant: %p",
					got.Defaults, tc.data.Defaults,
				)
			}
		})
	}
}

// ############
// # DEFAULTS #
// ############

func TestContainerDetailDefaults_Default(t *testing.T) {
	detail := ContainerDetailDefaults{}

	detail.Default()

	if detail.Tag == "" {
		t.Fatal("ContainerDetailDefaults.Tag mismatch\ngot:  \"\"\nwant: non-empty")
	}
}

// ##########
// # VALUES #
// ##########

func TestContainerDetail_GetImage(t *testing.T) {
	envPrefix := "TEST_CONTAINER_DETAIL_GET_IMAGE"

	tests := []struct {
		name string
		data ContainerDetail
		env  map[string]string
		want string
	}{
		{
			name: "empty",
			data: ContainerDetail{},
			want: "",
		},
		{
			name: "Image",
			data: ContainerDetail{
				Image: "i",
				Tag:   "t",
			},
			want: "i",
		},
		{
			name: "Image with env var",
			data: ContainerDetail{
				Image: `i-${` + envPrefix + `_VAR}`,
				Tag:   "t",
			},
			env: map[string]string{
				envPrefix + "_VAR": "value",
			},
			want: "i-value",
		},
		{
			name: "no Image",
			data: ContainerDetail{
				Tag: "t",
			},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			test.IsolateEnv(t, tc.env)

			got := tc.data.GetImage()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestContainerDetail_GetTag(t *testing.T) {
	envPrefix := "TEST_CONTAINER_DETAIL_GET_TAG"

	tests := []struct {
		name string
		data ContainerDetail
		env  map[string]string
		want string
	}{
		{
			name: "empty",
			data: ContainerDetail{},
			want: "",
		},
		{
			name: "Tag",
			data: ContainerDetail{
				Image: "i",
				Tag:   "t",
			},
			want: "t",
		},
		{
			name: "Tag with env var",
			data: ContainerDetail{
				Image: "i",
				Tag:   `t-${` + envPrefix + `_VAR}`,
			},
			env: map[string]string{
				envPrefix + "_VAR": "value",
			},
			want: "t-value",
		},
		{
			name: "no Tag",
			data: ContainerDetail{
				Image: "i",
			},
			want: "",
		},
		{
			name: "Tag from defaults",
			data: ContainerDetail{
				Image:    "i",
				Defaults: &ContainerDetailDefaults{Tag: "t-default"},
			},
			want: "t-default",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			test.IsolateEnv(t, tc.env)

			got := tc.data.GetTag()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestContainerDetailDefaults_GetTag(t *testing.T) {
	tests := []struct {
		name string
		data *ContainerDetailDefaults
		want string
	}{
		{
			name: "nil",
			data: nil,
			want: "",
		},
		{
			name: "own Tag",
			data: &ContainerDetailDefaults{Tag: "t"},
			want: "t",
		},
		{
			name: "Tag from chain",
			data: &ContainerDetailDefaults{
				Defaults: &ContainerDetailDefaults{Tag: "t-deep"},
			},
			want: "t-deep",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.GetTag()

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
