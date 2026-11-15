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
	"strconv"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

// dummyGitHubToken stands in for a real token where no request is made.
const dummyGitHubToken = "dummy-github-token"

// testWebVersion is the version the [testWeb] Lookup's URL returns.
const testWebVersion = "1.2.3"

func TestMain(m *testing.M) {
	logtest.InitLog()

	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

type mockLookup struct {
	base.Lookup
	OverrideErr string `json:"override_err,omitzero" yaml:"override_err,omitzero"`
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
	t.Helper()

	lvCfg := plainDefaultsConfig()

	switch typ {
	case "github":
		lv = testGitHub(t, fail)
	case "url":
		lv = testWeb(t, fail)
	}

	lv.Init(
		lv.GetOptions(),
		lv.GetStatus(),
		base.DefaultsConfig{
			Soft: lv.GetDefaults(),
			Hard: lv.GetHardDefaults(),
		},
	)
	applyTypeDefaults(lv, lvCfg)
	lv.GetStatus().ServiceInfo.ID = "TEST_LV"

	if err := lv.CheckValues(); err != nil {
		t.Fatalf("testLookup(type=%q, fail=%t).CheckValues() failed: %v", typ, fail, err)
	}

	return lv
}

func testGitHub(t *testing.T, fail bool) Lookup {
	t.Helper()

	lvCfg := plainDefaultsConfig()
	accessToken := test.GitHubToken(t)
	if fail {
		accessToken = "invalid"
	}

	svcStatus, _ := statustest.New("yaml", nil)
	lv, err := Decode(
		"yaml", []byte(test.TrimYAML(`
			type: github
			url: `+test.ArgusGitHubRepo+`
			access_token: `+accessToken+`
		`)),
		opttest.Options(t),
		svcStatus,
		lvCfg,
	)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	return lv
}

func testWeb(t *testing.T, fail bool) Lookup {
	t.Helper()

	lvCfg := plainDefaultsConfig()

	svcStatus, _ := statustest.New("yaml", nil)
	lv, err := Decode(
		"yaml", []byte(test.TrimYAML(`
			type: url
			url: `+test.LookupBare.URLInvalid+`/`+testWebVersion+`
			allow_invalid_certs: `+strconv.FormatBool(!fail)+`
		`)),
		opttest.Options(t),
		svcStatus,
		lvCfg,
	)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	return lv
}

// plainDefaultsConfig returns plain defaults and hardDefaults with a dummy GitHub token.
func plainDefaultsConfig() DefaultsConfig {
	optCfg := opttest.PlainDefaultsConfig()

	defaults, _ := DecodeDefaults("yaml", nil)
	defaults.Common.Options = optCfg.Soft
	hardDefaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults.Default()
	hardDefaults.GitHub.AccessToken = dummyGitHubToken
	hardDefaults.Common.Options = optCfg.Hard

	defaults.Common.Require.SetDefaults(&hardDefaults.Common.Require)

	return DefaultsConfig{
		Soft: defaults,
		Hard: hardDefaults,
	}
}
