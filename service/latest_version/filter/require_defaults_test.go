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

package filter

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
)

func TestRequireDefaults_IsZero(t *testing.T) {
	// GIVEN: a RequireDefaults.
	tests := []struct {
		name string
		req  *RequireDefaults
		want bool
	}{
		{
			name: "empty",
			req:  &RequireDefaults{},
			want: true,
		},
		{
			name: "non-empty",
			req: &RequireDefaults{
				Docker: docker.Defaults{
					Type: "ghcr",
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called.
			got := tc.req.IsZero()

			// THEN: the result is expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRequireDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRequireDefaults_SetDefaults(t *testing.T) {
	// GIVEN: RequireDefaults, and RequireDefaults to give it.
	defaults, _ := plainDefaults(t)
	req, _ := DecodeDefaults("yaml", nil)

	// WHEN: SetDefaults is called.
	req.SetDefaults(defaults)

	prefix := fmt.Sprintf("%s\nRequireDefaults.SetDefaults()", packageName)

	// THEN: defaults are set.
	fieldTests := []test.FieldAssertion{
		{Name: "Docker.Defaults", Got: req.Docker.Defaults, Want: &defaults.Docker, Mode: test.CompareSamePointer},
		{Name: "Docker.Registry.GHCR.Auth.Defaults", Got: req.Docker.Registry.GHCR.Auth.Defaults(), Want: defaults.Docker.Registry.GHCR.Auth, Mode: test.CompareSamePointer},
		{Name: "Docker.Registry.Hub.Auth.Defaults", Got: req.Docker.Registry.Hub.Auth.Defaults(), Want: defaults.Docker.Registry.Hub.Auth, Mode: test.CompareSamePointer},
		{Name: "Docker.Registry.Quay.Auth.Defaults", Got: req.Docker.Registry.Quay.Auth.Defaults(), Want: defaults.Docker.Registry.Quay.Auth, Mode: test.CompareSamePointer},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "RequireDefaults"); err != nil {
		t.Fatal(err)
	}
}

func TestRequireDefaults_CheckValues(t *testing.T) {
	// GIVEN: RequireDefaults.
	tests := []struct {
		name     string
		req      *RequireDefaults
		errRegex string
	}{
		{
			name:     "empty",
			req:      &RequireDefaults{},
			errRegex: `^$`,
		},
		{
			name: "valid Docker",
			req: &RequireDefaults{
				Docker: docker.Defaults{
					Type: "ghcr",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid Docker",
			req: &RequireDefaults{
				Docker: docker.Defaults{
					Type: "abc",
				},
			},
			errRegex: test.TrimYAML(`
				^docker:
					type: .* <invalid>.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.req.CheckValues,
			)
		})
	}
}
