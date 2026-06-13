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

package web

import (
	"fmt"
	"os"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
)

var packageName = "deployedver_web"

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

func testLookup(t *testing.T, failing bool) *Lookup {
	t.Helper()

	// Defaults.
	dvCfg := plainDefaultsConfig(t)
	// Options.
	optCfg := opttest.PlainDefaultsConfig(t)
	options, _ := opt.Decode(
		"yaml", []byte("semantic_versioning: true"),
		optCfg,
	)
	// Status.
	announceChannel := make(chan []byte, 24)
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)
	svcDashboard := &dashboard.Options{
		OptionsBase: dashboard.OptionsBase{
			WebURL: "https://example.com",
		},
	}
	svcStatus := status.New(
		announceChannel, databaseChannel, saveChannel,
		"",
		"", "",
		"", "",
		"",
		svcDashboard,
	)
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: "web-testLookup",
		},
		svcDashboard,
	)

	lookup, _ := Decode(
		"yaml", []byte(test.TrimYAML(`
			method: GET
			url:    `+test.LookupJSON["url_invalid"]+`
			json:   version
		`)),
		options,
		svcStatus,
		dvCfg,
	)
	allowInvalidCerts := !failing
	lookup.AllowInvalidCerts = &allowInvalidCerts

	return lookup
}

// plainDefaultsConfig returns plain defaults and hardDefaults for testing.
func plainDefaultsConfig(t *testing.T) base.DefaultsConfig {
	t.Helper()

	optDefaults, _ := opt.DecodeDefaults("yaml", nil)
	optHardDefaults, _ := opt.DecodeDefaults("yaml", nil)
	optHardDefaults.Default()

	defaults, _ := base.DecodeDefaults("yaml", nil)
	defaults.Options = optDefaults
	hardDefaults, _ := base.DecodeDefaults("yaml", nil)
	hardDefaults.Default()
	hardDefaults.Options = optHardDefaults

	return base.DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
