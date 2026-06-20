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

	"github.com/release-argus/Argus/internal/test"
)

// #########
// # STATE #
// #########

func TestContainerDetail_IsZero(t *testing.T) {
	// GIVEN: a ContainerDetails
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

			// WHEN: IsZero() is called on it.
			got := tc.data.IsZero()

			// THEN: the expected result is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nContainerDetail.IsZero() value mismatch\ngot:  %t\nwant: %t",
					tc.name, got, tc.want,
				)
			}
		})
	}
}

func TestContainerDetail_Copy(t *testing.T) {
	// GIVEN: a ContainerDetails
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy() is called on it.
			got := tc.data.Copy()

			prefix := fmt.Sprintf("%s\nContainerDetail.Copy()", packageName)

			// THEN: the .Image is copied over.
			if got.Image != tc.data.Image {
				t.Fatalf(
					"%s .Image mismatch\ngot:  %q\nwant: %q",
					prefix, got.Image, tc.data.Image,
				)
			}

			// AND: the .Image is at a different address.
			if &got.Image == &tc.data.Image {
				t.Fatalf(
					"%s .Image address mismatch\ngot:      %p\nwant: not-%p",
					prefix, &got.Image, &tc.data.Image,
				)
			}

			// AND: the .Tag is copied over.
			if got.Tag != tc.data.Tag {
				t.Fatalf(
					"%s .Tag mismatch\ngot:  %q\nwant: %q",
					prefix, got.Tag, tc.data.Tag,
				)
			}

			// AND: the .Tag is at a different address.
			if &got.Tag == &tc.data.Tag {
				t.Fatalf(
					"%s .Tag address mismatch\ngot:      %p\nwant: not-%p",
					prefix, &got.Tag, &tc.data.Tag,
				)
			}
		})
	}
}

// ############
// # DEFAULTS #
// ############

func TestContainerDetail_Default(t *testing.T) {
	// GIVEN: a ContainerDetail.
	detail := ContainerDetail{}

	// WHEN: Default() is called on it.
	detail.Default()

	// THEN: .Tag is given a value.
	if detail.Tag == "" {
		t.Fatalf("%s\nContainerDetail.Default() mismatch\nTag is empty", packageName)
	}
}

// ##########
// # VALUES #
// ##########

func TestContainerDetail_GetImage(t *testing.T) {
	envPrefix := "TEST_CONTAINER_DETAIL_GET_IMAGE"
	// GIVEN: a ContainerDetail.
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
			test.SetEnv(t, tc.env)

			// WHEN: GetImage() is called on it.
			got := tc.data.GetImage()

			// THEN: the expected image is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nContainerDetail.GetImage() mismatch\ngot:  %q\nwant: %q",
					tc.name, got, tc.want,
				)
			}
		})
	}
}

func TestContainerDetail_GetTag(t *testing.T) {
	envPrefix := "TEST_CONTAINER_DETAIL_GET_TAG"
	// GIVEN: a ContainerDetail.
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
			name: "Image with env var",
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			test.SetEnv(t, tc.env)

			// WHEN: GetTag() is called on it.
			got := tc.data.GetTag()

			// THEN: the expected tag is returned.
			if got != tc.want {
				t.Fatalf(
					"%s\nContainerDetail.GetTag() mismatch\ngot:  %q\nwant: %q",
					tc.name, got, tc.want,
				)
			}
		})
	}
}
