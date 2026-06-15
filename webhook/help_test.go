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

package webhook

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
)

var packageName = "defaults"

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

func testWebHook(failing bool, selfSignedCert bool, headers bool) *WebHook {
	desiredStatusCode := uint16(0)
	whMaxTries := uint8(1)
	webhook := New(
		test.Ptr(false),
		nil,
		"0s",
		&desiredStatusCode,
		nil,
		"test",
		&whMaxTries,
		Notifiers{},
		test.Ptr("12m"),
		"argus",
		test.Ptr(false),
		"github",
		test.WebHookGitHub["url_valid"],
		&Defaults{},
		&Defaults{}, &Defaults{},
	)
	webhook.ServiceStatus = &status.Status{}
	webhook.ServiceStatus.Init(
		0, 1, 1,
		status.ServiceInfo{
			ID: "testWebHook",
		},
		&dashboard.Options{
			OptionsBase: dashboard.OptionsBase{
				WebURL: "https://example.com",
			},
		},
	)
	webhook.Failed = &webhook.ServiceStatus.Fails.WebHook
	if selfSignedCert {
		webhook.URL = strings.Replace(webhook.URL, "valid", "invalid", 1)
	}
	if failing {
		webhook.Secret = "invalid"
	}
	if headers {
		webhook.URL = strings.Replace(webhook.URL, "github-style", "single-header", 1)
		testHeaderValue := "secret"
		if failing {
			testHeaderValue = "invalid"
		}
		webhook.Headers = Headers{
			{Key: "X-Test", Value: testHeaderValue},
		}
	}
	webhook.initMetrics()
	return webhook
}

func testDefaults(failing bool, headers bool) *Defaults {
	desiredStatusCode := "0"
	whMaxTries := "1"

	wh, _ := DecodeDefaults(
		"yaml", []byte(test.TrimYAML(`
			allow_invalid_certs: false
			delay: 0s
			desired_status_code: `+desiredStatusCode+`
			max_tries: `+whMaxTries+`
			secret: argus
			silent_fails: false
			type: github
			url: `+test.WebHookGitHub["url_valid"]+`
		`)),
	)

	if failing {
		wh.Secret = "invalid"
	}

	if headers {
		wh.URL = strings.Replace(wh.URL, "github-style", "single-header", 1)
		if failing {
			wh.Headers = Headers{
				{Key: "X-Test", Value: "invalid"},
			}
		} else {
			wh.Headers = Headers{
				{Key: "X-Test", Value: "secret"},
			}
		}
	}

	return wh
}

// plainConfig returns plain defaults and hardDefaults for testing.
func plainConfig(t *testing.T) Config {
	t.Helper()

	defaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults.Default()

	return Config{
		Root:         WebHooksDefaults{},
		Defaults:     defaults,
		HardDefaults: hardDefaults,
	}
}
