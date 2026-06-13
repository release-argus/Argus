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

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

var packageName = "latestver_web"

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

	lvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	// Options.
	options, _ := opt.Decode(
		"yaml", nil,
		optCfg,
	)
	fmt.Println(options.Defaults == nil)
	fmt.Println(options.HardDefaults == nil)
	// Status.
	svcStatus, _ := statustest.New("yaml", nil)
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: "web-testLookup",
		},
		&dashboard.Options{
			OptionsBase: dashboard.OptionsBase{
				WebURL: "https://example.com",
			},
		},
	)

	lookup, _ := Decode(
		"yaml", []byte(test.TrimYAML(`
			url: `+test.LookupPlain["url_invalid"]+`
			url_commands:
				- type: regex
					regex: 'ver([0-9.]+)'
			allow_invalid_certs: `+fmt.Sprint(!failing)+`
		`)),
		options,
		svcStatus,
		lvCfg,
	)

	if failing {
		*lookup.AllowInvalidCerts = false
	}

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
	hardDefaults.AccessToken = test.GitHubToken(t)
	hardDefaults.Options = optHardDefaults

	defaults.Require.SetDefaults(&hardDefaults.Require)

	return base.DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
