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

package deployedver

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/service/deployed_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

var packageName = "deployedver"

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
func (f *mockLookup) GetType() string                             { return "fake" }
func (f *mockLookup) String(prefix string) string                 { return decode.ToYAMLString(f, prefix) }
func (f *mockLookup) Track()                                      {}

func testLookup(t *testing.T, typ string, fail bool, version string) (dv Lookup) {
	dvCfg := plainDefaultsConfig(t)

	switch typ {
	case "manual":
		dv = testManual(t, version)
	case "url":
		dv = testWeb(t, fail, version)
	}

	dv.Init(
		dv.GetOptions(),
		dv.GetStatus(),
		dvCfg,
	)
	dv.GetStatus().ServiceInfo.ID = "TEST_DV"

	// Check the values.
	if err := dv.CheckValues(); err != nil {
		t.Fatalf(
			"%s.Lookup(type=%q, fail=%t).CheckValues() unexpected error: %v",
			packageName, dv, fail, err,
		)
	}

	return dv
}

func testManual(t *testing.T, version string) Lookup {
	dvCfg := plainDefaultsConfig(t)

	svcStatus, _ := statustest.New("yaml", nil)
	dv, _ := Decode(
		"yaml", []byte(test.TrimYAML(`
			type: manual
			version: `+version+`
		`)),
		opttest.Options(t),
		svcStatus,
		dvCfg,
	)

	return dv
}

func testWeb(t *testing.T, fail bool, version string) Lookup {
	dvCfg := plainDefaultsConfig(t)

	svcStatus, _ := statustest.New("yaml", nil)
	dv, _ := Decode(
		"yaml", []byte(test.TrimYAML(`
			type: url
			method: GET
			url: `+test.LookupBare["url_invalid"]+`/`+version+`
			allow_invalid_certs: `+fmt.Sprint(!fail)+`
		`)),
		opttest.Options(t),
		svcStatus,
		dvCfg,
	)

	return dv
}

// plainDefaults returns plain defaults and hardDefaults for testing.
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

func getType(lookup Lookup) string {
	switch lookup.(type) {
	case *web.Lookup:
		return "url"
	case *manual.Lookup:
		return "manual"
	}
	return "unknown"
}
