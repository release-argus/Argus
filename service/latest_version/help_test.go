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

package latestver

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

var packageName = "latestver"

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

type mockLookup struct {
	base.Lookup
	OverrideErr string `json:"override_err,omitempty" yaml:"override_err,omitempty"`
}

func (f *mockLookup) ApplyOverrides(format string, data []byte) error {
	if f.OverrideErr != "" {
		return errors.New(f.OverrideErr)
	}
	return nil
}
func (f *mockLookup) Copy(*status.Status) base.Interface          { return f }
func (f *mockLookup) DecodeSelf(format string, data []byte) error { return nil }
func (f *mockLookup) GetRequire() *filter.Require                 { return f.Require }
func (f *mockLookup) SetRequire(r *filter.Require)                { f.Require = r }
func (f *mockLookup) String(prefix string) string                 { return decode.ToYAMLString(f, prefix) }

func testLookup(t *testing.T, typ string, fail bool) (lv Lookup) {
	if t != nil {
		t.Helper()
	}

	lvCfg := plainDefaultsConfig(t)

	switch typ {
	case "github":
		lv = testGitHub(t, fail)
	case "url":
		lv = testWeb(t, fail)
	}

	lv.Init(
		lv.GetOptions(),
		lv.GetStatus(),
		lvCfg,
	)
	lv.GetStatus().ServiceInfo.ID = "TEST_LV"

	// Check the values.
	if err := lv.CheckValues(); err != nil {
		t.Fatalf(
			"%s.Lookup(type=%q, fail=%t).CheckValues() unexpected error: %v",
			packageName, typ, fail, err,
		)
	}

	return lv
}

func testGitHub(t *testing.T, fail bool) Lookup {
	lvCfg := plainDefaultsConfig(t)
	accessToken := test.GitHubToken(t)
	if fail {
		accessToken = "invalid"
	}

	svcStatus, _ := statustest.New("yaml", nil)
	lv, _ := Decode(
		"yaml", []byte(test.TrimYAML(`
			type: github
			url: `+test.ArgusGitHubRepo+`
			access_token: `+accessToken+`
		`)),
		opttest.Options(t),
		svcStatus,
		lvCfg,
	)

	return lv
}

func testWeb(t *testing.T, fail bool) Lookup {
	lvCfg := plainDefaultsConfig(t)

	svcStatus, _ := statustest.New("yaml", nil)
	lv, _ := Decode(
		"yaml", []byte(test.TrimYAML(`
			type: url
			url: `+test.LookupBare["url_invalid"]+`/1.2.3
			allow_invalid_certs: `+fmt.Sprint(!fail)+`
		`)),
		opttest.Options(t),
		svcStatus,
		lvCfg,
	)

	return lv
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
	hardDefaults.AccessToken = test.GitHubToken(nil)
	hardDefaults.Options = optHardDefaults

	defaults.Require.SetDefaults(&hardDefaults.Require)

	return base.DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
