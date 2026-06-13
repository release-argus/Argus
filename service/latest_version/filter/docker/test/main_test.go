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
)

func TestGetDefaultOfDockerType(t *testing.T) {
	// GIVEN: a Defaults.
	defaults := &docker.Defaults{}

	tests := []struct {
		name   string
		dType  string
		expect docker.RegistryDefaults
		err    bool
	}{
		{
			name:   "ghcr",
			dType:  "ghcr",
			expect: defaults.Registry.GHCR,
		},
		{
			name:   "hub",
			dType:  "hub",
			expect: defaults.Registry.Hub,
		},
		{
			name:   "quay",
			dType:  "quay",
			expect: defaults.Registry.Quay,
		},
		{
			name:  "unknown",
			dType: "foo",
			err:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetDefaultOfDockerType is called with the Defaults and a registry type.
			got, err := GetDefaultOfDockerType(t, tc.dType, defaults)

			// THEN: an error is returned when expected *unknown type).
			if err != nil {
				if tc.err {
					return
				}
				t.Fatalf(
					"%s\nunexpected error for type %q: %v",
					packageName, tc.dType,
					err,
				)
			}

			if got != tc.expect {
				t.Fatalf(
					"%s\npointer mismatch for type %q:\ngot:  %p\nwant: %p",
					packageName, tc.dType,
					got, tc.expect,
				)
			}
		})
	}
}

func TestFakeRegistryDefaults_IsZero(t *testing.T) {
	// GIVEN: a FakeRegistryDefaults.
	fake := &MockRegistryDefaults{}
	// WHEN: IsZero is called.
	got := fake.IsZero()
	// THEN: false is returned.
	if got != false {
		t.Fatalf(
			"%s\nunexpected IsZero result:\ngot:  %v\nwant: %v",
			packageName, got, false,
		)
	}
}

func TestFakeRegistryDefaults_GetType(t *testing.T) {
	// GIVEN: a FakeRegistryDefaults.
	fake := &MockRegistryDefaults{}
	// WHEN: GetType is called.
	got := fake.GetType()
	// THEN: "fake" is returned.
	if got != "fake" {
		t.Fatalf(
			"%s\nunexpected GetType result:\ngot:  %q\nwant: %q",
			packageName, got, "fake",
		)
	}
}

func TestFakeRegistryDefaults_String(t *testing.T) {
	// GIVEN: a FakeRegistryDefaults.
	fake := &MockRegistryDefaults{}
	// WHEN: String is called.
	got := fake.String("")
	// THEN: an empty string is returned.
	if got != "" {
		t.Fatalf(
			"%s\nunexpected String result:\ngot:  %q\nwant: %q",
			packageName, got, "",
		)
	}
}
