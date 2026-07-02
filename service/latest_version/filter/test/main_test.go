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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/service/latest_version/filter/docker"
)

func TestRequire(t *testing.T) {
	// GIVEN: a Docker type.
	tests := []struct {
		dockerType string
	}{
		{dockerType: "ecr"},
		{dockerType: "ghcr"},
		{dockerType: "hub"},
		{dockerType: "quay"},
	}

	for _, tc := range tests {
		t.Run(tc.dockerType, func(t *testing.T) {
			t.Parallel()

			// WHEN: Require is called with it.
			got := Require(t, tc.dockerType)

			prefix := fmt.Sprintf(
				"%s\nRequire(type=%q)",
				packageName, tc.dockerType,
			)

			// THEN: the expected Docker type is returned.
			switch tc.dockerType {
			case "ecr":
				if _, ok := got.Docker.(*docker.ECRRegistry); !ok {
					t.Errorf(
						"%s type mismatch\ngot:  %t\nwant: %q",
						prefix, got.Docker, "ECRRegistry",
					)
				}
			case "ghcr":
				if _, ok := got.Docker.(*docker.GHCRRegistry); !ok {
					t.Errorf(
						"%s type mismatch\ngot:  %t\nwant: %q",
						prefix, got.Docker, "GHCRRegistry",
					)
				}
			case "hub":
				if _, ok := got.Docker.(*docker.HubRegistry); !ok {
					t.Errorf(
						"%s type mismatch\ngot:  %t\nwant: %q",
						prefix, got.Docker, "HubRegistry",
					)
				}
			case "quay":
				if _, ok := got.Docker.(*docker.QuayRegistry); !ok {
					t.Errorf(
						"%s type mismatch\ngot:  %t\nwant: %q",
						prefix, got.Docker, "QuayRegistry",
					)
				}
			}
		})
	}
}
