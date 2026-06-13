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

package test

import (
	"github.com/release-argus/Argus/webhook"
)

// PlainConfig returns plain defaults and hardDefaults for testing.
func PlainConfig() webhook.Config {
	defaults, _ := webhook.DecodeDefaults("yaml", nil)
	hardDefaults, _ := webhook.DecodeDefaults("yaml", nil)
	hardDefaults.Default()

	return webhook.Config{
		Root:         webhook.WebHooksDefaults{},
		Defaults:     defaults,
		HardDefaults: hardDefaults,
	}
}
