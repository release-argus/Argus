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

//go:build unit || integration

// Package test provides test helpers for the latest_version package.
package test

import (
	"testing"

	latestver "github.com/release-argus/Argus/service/latest_version"
	opttest "github.com/release-argus/Argus/service/option/test"
)

// dummyGitHubToken stands in for a real token where no request is made.
const dummyGitHubToken = "dummy-github-token"

// PlainDefaultsConfig returns plain defaults and hardDefaults with a dummy GitHub token.
func PlainDefaultsConfig(t *testing.T) latestver.DefaultsConfig {
	t.Helper()

	optCfg := opttest.PlainDefaultsConfig(t)

	defaults, _ := latestver.DecodeDefaults("yaml", nil)
	defaults.Common.Options = optCfg.Soft
	hardDefaults, _ := latestver.DecodeDefaults("yaml", nil)
	hardDefaults.Default()
	hardDefaults.GitHub.AccessToken = dummyGitHubToken
	hardDefaults.Common.Options = optCfg.Hard

	defaults.Common.Require.SetDefaults(&hardDefaults.Common.Require)

	return latestver.DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
