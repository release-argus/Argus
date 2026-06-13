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

package option

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
)

var packageName = "option"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

func testOptions() *Options {
	optCfg := plainDefaultsConfig()

	o, _ := Decode(
		"yaml", []byte(test.TrimYAML(`
			active: true
			interval: 10m
			semantic_versioning: true
		`)),
		optCfg,
	)

	return o
}

func testDefaults() *Defaults {
	defaults, _ := DecodeDefaults(
		"yaml", []byte(test.TrimYAML(`
			interval: 1h
			semantic_versioning: false
		`)),
	)
	return defaults
}

// plainDefaultsConfig returns plain defaults and hardDefaults for testing.
func plainDefaultsConfig() DefaultsConfig {
	defaults := Defaults{}
	hardDefaults := Defaults{}
	hardDefaults.Default()

	return DefaultsConfig{
		Soft: &defaults,
		Hard: &hardDefaults,
	}
}
