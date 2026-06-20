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

// Package test provides test helpers for the docker filter package.
package test

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/service/latest_version/filter/docker"
)

const packageName = "docker.test"

// GetDefaultOfDockerType returns the registry defaults for dType from defaults.
func GetDefaultOfDockerType(t *testing.T, dType string, defaults *docker.Defaults) (docker.RegistryDefaults, error) {
	t.Helper()

	switch dType {
	case "ghcr":
		return defaults.Registry.GHCR, nil
	case "hub":
		return defaults.Registry.Hub, nil
	case "quay":
		return defaults.Registry.Quay, nil
	}

	return nil, fmt.Errorf(
		"%s\nunknown docker registry type: %s",
		packageName, dType,
	)
}

// MockRegistryDefaults is a minimal RegistryDefaults used in tests.
type MockRegistryDefaults struct {
	docker.CommonRegistryDefaults `json:",inline" yaml:",inline"`
}

// IsZero implements the yaml.IsZeroer interface.
func (f *MockRegistryDefaults) IsZero() bool {
	return false
}

// GetType returns the mock registry type identifier.
func (f *MockRegistryDefaults) GetType() string {
	return "fake"
}

// String implements RegistryDefaults.
func (f *MockRegistryDefaults) String(string) string {
	return ""
}
