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

package service

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	statustest "github.com/release-argus/Argus/service/status/test"
	whtest "github.com/release-argus/Argus/webhook/test"
)

var packageName = "service"

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

func testService(t *testing.T, id string, lvType, dvType string) *Service {
	svcCfg := plainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	svc := test.Must(t, func() (*Service, error) {
		return DecodeService(
			"yaml", []byte(test.TrimYAML(`
				latest_version:
				`+lvtest.Lookup(t, lvType, false).String(" ")+`
				deployed_version:
				`+dvtest.Lookup(t, dvType, false, "").String("  ")+`
				dashboard:
					auto_approve: false
					icon: "test"
					icon_link_to: "https://release-argus.io"
					web_url: "https://release-argus.io"
			`)),
			id,
			svcCfg, notifyCfg, whCfg,
		)
	})

	// Status.
	svcStatus, _ := statustest.New("yaml", nil)
	svc.Status = *svcStatus.Copy(true)
	svc.HardDefaults.Status.AnnounceChannel = svc.Status.AnnounceChannel
	svc.HardDefaults.Status.DatabaseChannel = svc.Status.DatabaseChannel
	svc.HardDefaults.Status.SaveChannel = svc.Status.SaveChannel
	svc.Status.ServiceInfo.ID = svc.ID

	// Check the values.
	if err, _ := svc.CheckValues(); err != nil {
		t.Fatalf(
			"%s\ntestService().CheckValues() error: %v",
			packageName, err,
		)
	}

	return svc
}

// plainDefaultsConfig returns plain defaults and hardDefaults for testing.
func plainDefaultsConfig(t *testing.T) DefaultsConfig {
	t.Helper()

	defaults := Defaults{}
	hardDefaults := Defaults{}
	hardDefaults.Default()
	hardDefaults.LatestVersion.AccessToken = test.GitHubToken(nil)

	defaults.SetDefaults(&hardDefaults)
	return DefaultsConfig{
		Soft: &defaults,
		Hard: &hardDefaults,
	}
}
