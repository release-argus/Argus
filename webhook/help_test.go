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

func TestMain(m *testing.M) {
	logtest.InitLog()

	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

func testWebhook(failing bool, headers bool) *Webhook {
	desiredStatusCode := uint16(0)
	whMaxTries := uint8(1)
	webhook := New(
		new(false),
		nil,
		"0s",
		&desiredStatusCode,
		nil,
		"test",
		&whMaxTries,
		Notifiers{},
		new("12m"),
		"argus",
		new(false),
		"github",
		test.WebhookGitHub.URLValid,
		&Defaults{},
		&Defaults{}, &Defaults{},
	)
	webhook.ServiceStatus = &status.Status{}
	webhook.ServiceStatus.Init(
		0, 1, 1,
		status.ServiceInfo{
			ID: "testWebhook",
		},
		&dashboard.Options{
			WebURL: "https://example.com",
		},
	)
	webhook.Failed = &webhook.ServiceStatus.Fails.Webhook
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

func testDefaults(t *testing.T, failing bool) *Defaults {
	t.Helper()

	wh, err := DecodeDefaults(
		"yaml", []byte(test.TrimYAML(`
			allow_invalid_certs: false
			delay: 0s
			desired_status_code: 0
			max_tries: 1
			secret: argus
			silent_fails: false
			type: github
			url: `+test.WebhookGitHub.URLValid+`
		`)),
	)
	if err != nil {
		t.Fatalf("DecodeDefaults() failed: %v", err)
	}

	if failing {
		wh.Secret = "invalid"
	}

	return wh
}

// plainConfig returns plain defaults and hardDefaults for testing.
func plainConfig() Config {
	defaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults.Default()

	return Config{
		Root:         WebhooksDefaults{},
		Defaults:     defaults,
		HardDefaults: hardDefaults,
	}
}
