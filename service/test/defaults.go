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

// Package test provides test helpers for the service package.
package test

import (
	"testing"

	"github.com/release-argus/Argus/service"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

// PlainDefaultsConfig returns plain defaults and hardDefaults for testing.
func PlainDefaultsConfig(t *testing.T) service.DefaultsConfig {
	t.Helper()

	optDefaults, _ := opt.DecodeDefaults("yaml", nil)
	optHardDefaults, _ := opt.DecodeDefaults("yaml", nil)
	optHardDefaults.Default()

	defaults, _ := service.DecodeDefaults("yaml", nil)
	defaults.Options = *optDefaults
	hardDefaults, _ := service.DecodeDefaults("yaml", nil)
	hardDefaults.Default()

	hardDefaults.Options = *optHardDefaults

	svcStatus, _ := statustest.New("yaml", nil)
	hardDefaults.Status = status.NewDefaults(
		svcStatus.AnnounceChannel,
		svcStatus.DatabaseChannel,
		svcStatus.SaveChannel,
	)

	defaults.SetDefaults(hardDefaults)

	return service.DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
