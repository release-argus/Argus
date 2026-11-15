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

package test

import (
	"testing"

	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestGetDefaultOfDockerType(t *testing.T) {
	defaults := &docker.Defaults{}
	defaults.Default()

	tests := []struct {
		name     string
		dType    string
		want     docker.RegistryDefaults
		errRegex string
	}{
		{
			name:     "ecr",
			dType:    "ecr",
			want:     defaults.Registry.ECR,
			errRegex: `^$`,
		},
		{
			name:     "ghcr",
			dType:    "ghcr",
			want:     defaults.Registry.GHCR,
			errRegex: `^$`,
		},
		{
			name:     "hub",
			dType:    "hub",
			want:     defaults.Registry.Hub,
			errRegex: `^$`,
		},
		{
			name:     "quay",
			dType:    "quay",
			want:     defaults.Registry.Quay,
			errRegex: `^$`,
		},
		{
			name:     "unknown",
			dType:    "foo",
			errRegex: `^unknown docker registry type: "foo"$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := GetDefaultOfDockerType(tc.dType, defaults)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if got != tc.want {
				t.Fatalf("RegistryDefaults mismatch\ngot:  %p\nwant: %p", got, tc.want)
			}
		})
	}
}

func TestMockRegistryDefaults_GetType(t *testing.T) {
	fake := &MockRegistryDefaults{}

	got := fake.GetType()

	if want := "fake"; got != want {
		t.Errorf("MockRegistryDefaults.GetType() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestMockRegistryDefaults_String(t *testing.T) {
	fake := &MockRegistryDefaults{}

	got := fake.String("  ")

	if got != "" {
		t.Errorf("MockRegistryDefaults.String() mismatch\ngot:  %q\nwant: \"\"", got)
	}
}
