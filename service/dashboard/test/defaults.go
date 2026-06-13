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

// Package test provides test helpers for the dashboard package.
package test

import (
	"testing"

	"github.com/release-argus/Argus/service/dashboard"
)

// PlainDefaultsConfig returns plain defaults and hardDefaults for testing.
func PlainDefaultsConfig(t *testing.T) dashboard.DefaultsConfig {
	t.Helper()

	defaults, _ := dashboard.DecodeDefaults("yaml", nil)
	hardDefaults, _ := dashboard.DecodeDefaults("yaml", nil)
	hardDefaults.Default()

	return dashboard.DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
