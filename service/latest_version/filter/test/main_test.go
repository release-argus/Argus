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
)

func TestRequire(t *testing.T) {
	tests := []struct {
		dockerType string
		want       string
	}{
		{dockerType: "ecr", want: "*docker.ECRRegistry"},
		{dockerType: "ghcr", want: "*docker.GHCRRegistry"},
		{dockerType: "hub", want: "*docker.HubRegistry"},
		{dockerType: "quay", want: "*docker.QuayRegistry"},
	}

	for _, tc := range tests {
		t.Run(tc.dockerType, func(t *testing.T) {
			t.Parallel()

			got := Require(t, tc.dockerType)

			if gotType := fmt.Sprintf("%T", got.Docker); gotType != tc.want {
				t.Errorf("Require.Docker mismatch\ngot:  %s\nwant: %s", gotType, tc.want)
			}
		})
	}
}
