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

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
)

// PlainDefaultsConfig returns plain defaults and hardDefaults for testing.
func PlainDefaultsConfig(t *testing.T) base.DefaultsConfig {
	t.Helper()

	optDefaults, _ := opt.DecodeDefaults("yaml", nil)
	optHardDefaults, _ := opt.DecodeDefaults("yaml", nil)
	optHardDefaults.Default()

	defaults, _ := base.DecodeDefaults("yaml", nil)
	defaults.Options = optDefaults
	hardDefaults, _ := base.DecodeDefaults("yaml", nil)
	hardDefaults.Default()
	hardDefaults.AccessToken = test.GitHubToken(t)
	hardDefaults.Options = optHardDefaults

	defaults.Require.SetDefaults(&hardDefaults.Require)

	return base.DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
