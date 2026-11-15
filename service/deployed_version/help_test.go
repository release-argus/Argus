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
	"strconv"
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
func (f *mockLookup) GetType() string                             { return "fake" }
func (f *mockLookup) String(prefix string) string                 { return decode.ToYAMLString(f, prefix) }
func (f *mockLookup) Track()                                      {}

func testLookup(t *testing.T, typ string, fail bool, version string) (dv Lookup) {
	t.Helper()

	switch typ {
	case manual.Type:
		dv = testManual(t, version)
	case web.Type:
		dv = testWeb(t, fail, version)
	default:
		t.Fatalf("testLookup(type=%q): unsupported type", typ)
	}

	dv.Init(
		dv.GetOptions(),
		dv.GetStatus(),
		base.DefaultsConfig{
			Soft: dv.GetDefaults(),
			Hard: dv.GetHardDefaults(),
		},
	)
	dv.GetStatus().ServiceInfo.ID = "TEST_DV"

	if err := dv.CheckValues(); err != nil {
		t.Fatalf("testLookup(type=%q, fail=%t).CheckValues() failed: %v", typ, fail, err)
	}

	return dv
}

func testManual(t *testing.T, version string) Lookup {
	t.Helper()

	dvCfg := plainDefaultsConfig()

	svcStatus, _ := statustest.New("yaml", nil)
	dv, err := Decode(
		"yaml", []byte(test.TrimYAML(`
			type: manual
			version: `+version+`
		`)),
		opttest.Options(t),
		svcStatus,
		dvCfg,
	)
	if err != nil {
		t.Fatalf("Decode(version=%q) failed: %v", version, err)
	}

	return dv
}

func testWeb(t *testing.T, fail bool, version string) Lookup {
	t.Helper()

	dvCfg := plainDefaultsConfig()

	svcStatus, _ := statustest.New("yaml", nil)
	dv, err := Decode(
		"yaml", []byte(test.TrimYAML(`
			type: url
			method: GET
			url: `+test.LookupBare.URLInvalid+`/`+version+`
			allow_invalid_certs: `+strconv.FormatBool(!fail)+`
		`)),
		opttest.Options(t),
		svcStatus,
		dvCfg,
	)
	if err != nil {
		t.Fatalf("Decode(version=%q) failed: %v", version, err)
	}

	return dv
}

// plainDefaultsConfig returns plain defaults and hardDefaults for testing.
func plainDefaultsConfig() base.DefaultsConfig {
	hardDefaults := &base.Defaults{Options: &opt.Defaults{}}
	hardDefaults.Default()
	hardDefaults.Options.Default()

	return base.DefaultsConfig{
		Soft: &base.Defaults{Options: &opt.Defaults{}},
		Hard: hardDefaults,
	}
}

func getType(lookup Lookup) string {
	switch lookup.(type) {
	case *web.Lookup:
		return web.Type
	case *manual.Lookup:
		return manual.Type
	}
	return "unknown"
}
