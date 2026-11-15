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
	"strconv"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	lvbasetest "github.com/release-argus/Argus/service/latest_version/types/base/test"
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

// testLookup returns a Lookup on the invalid-cert server, refusing its cert when failing.
func testLookup(t *testing.T, failing bool) *Lookup {
	t.Helper()

	lvCfg := lvbasetest.PlainDefaultsConfig()

	options := opttest.PlainOptions(opttest.PlainDefaultsConfig())
	svcStatus, _ := statustest.New("yaml", nil)
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: "web-testLookup",
		},
		&dashboard.Options{
			WebURL: "https://example.com",
		},
	)

	lookup, err := Decode(
		"yaml", []byte(test.TrimYAML(`
			url: `+test.LookupPlain.URLInvalid+`
			url_commands:
				- type: regex
					regex: 'ver([0-9.]+)'
			allow_invalid_certs: `+strconv.FormatBool(!failing)+`
		`)),
		options,
		svcStatus,
		lvCfg,
	)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	var typeHardDefaults Defaults
	typeHardDefaults.Default()
	lookup.SetTypeDefaults(&Defaults{}, &typeHardDefaults)

	return lookup
}
