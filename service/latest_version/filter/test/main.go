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

// Package test provides test helpers for the latest_version filter package.
package test

import (
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
)

// Require builds a [filter.Require] configured for dockerType registry tests.
func Require(t *testing.T, dockerType string) *filter.Require {
	t.Helper()

	requireDefaults, _ := filter.DecodeDefaults("yaml", nil)
	requireHardDefaults, _ := filter.DecodeDefaults("yaml", nil)
	requireHardDefaults.Default()

	var image string
	switch dockerType {
	case "ecr":
		image = test.ArgusDockerECRRepo
	case "ghcr":
		image = test.ArgusDockerGHCRRepo
	case "hub":
		image = test.ArgusDockerHubRepo
	case "quay":
		image = test.ArgusDockerQuayRepo
	}

	testRequire := &filter.Require{
		Docker: test.Must(t, func() (docker.Registry, error) {
			return docker.Decode(
				"yaml", []byte(test.TrimYAML(`
					type: `+dockerType+`
					image: `+image+`
					tag: '{{ version }}'
					username: `+dockerType+`-username
					token: `+dockerType+`-token
				`)),
				&requireDefaults.Docker,
			)
		}),
	}
	testRequire.Docker.GetAuth().SetQueryToken("foo", time.Now().Add(1*time.Hour))

	return testRequire
}
