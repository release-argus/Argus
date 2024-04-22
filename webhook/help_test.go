// Copyright [2022] [Argus]
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
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestMain(m *testing.M) {
	// initialize jLog
	jLog = util.NewJLog("DEBUG", false)
	jLog.Testing = true
	shoutrrr.LogInit(jLog)

	// run other tests
	exitCode := m.Run()

	// exit
	os.Exit(exitCode)
}

func testWebHook(failing bool, selfSignedCert bool, customHeaders bool) *WebHook {
	desiredStatusCode := 0
	whMaxTries := uint(1)
	webhook := New(
		test.BoolPtr(false),
		nil,
		"0s",
		&desiredStatusCode,
		nil,
		&whMaxTries,
		nil,
		test.StringPtr("12m"),
		"argus",
		test.BoolPtr(false),
		"github",
		"https://valid.release-argus.io/hooks/github-style",
		&WebHookDefaults{},
		&WebHookDefaults{},
		&WebHookDefaults{})
	webhook.ID = "test"
	webhook.ServiceStatus = &svcstatus.Status{}
	webhook.ServiceStatus.Init(
		0, 0, 1,
		test.StringPtr("testServiceID"),
		nil)
	webhook.Failed = &webhook.ServiceStatus.Fails.WebHook
	serviceName := "testServiceID"
	webURL := "https://example.com"
	webhook.ServiceStatus.Init(
		0, 1, 0,
		&serviceName,
		&webURL)
	if selfSignedCert {
		webhook.URL = strings.Replace(webhook.URL, "valid", "invalid", 1)
	}
	if failing {
		webhook.Secret = "invalid"
	}
	if customHeaders {
		webhook.URL = strings.Replace(webhook.URL, "github-style", "single-header", 1)
		if failing {
			webhook.CustomHeaders = &Headers{
				{Key: "X-Test", Value: "invalid"}}
		} else {
			webhook.CustomHeaders = &Headers{
				{Key: "X-Test", Value: "secret"}}
		}
	}
	webhook.initMetrics()
	return webhook
}

func testWebHookDefaults(failing bool, selfSignedCert bool, customHeaders bool) *WebHookDefaults {
	desiredStatusCode := 0
	whMaxTries := uint(1)
	webhook := NewDefaults(
		test.BoolPtr(false),
		nil,
		"0s",
		&desiredStatusCode,
		&whMaxTries,
		"argus",
		test.BoolPtr(false),
		"github",
		"https://valid.release-argus.io/hooks/github-style")
	if failing {
		webhook.Secret = "invalid"
	}
	if customHeaders {
		webhook.URL = strings.Replace(webhook.URL, "github-style", "single-header", 1)
		if failing {
			webhook.CustomHeaders = &Headers{
				{Key: "X-Test", Value: "invalid"}}
		} else {
			webhook.CustomHeaders = &Headers{
				{Key: "X-Test", Value: "secret"}}
		}
	}
	return webhook
}

func testNotifier(failing bool, selfSignedCert bool) *shoutrrr.Shoutrrr {
	url := "valid.release-argus.io"
	if selfSignedCert {
		url = strings.Replace(url, "valid", "invalid", 1)
	}
	notifier := shoutrrr.New(
		nil,
		"test",
		&map[string]string{
			"max_tries": "1"},
		&map[string]string{},
		"gotify",
		&map[string]string{
			"host": url,
			"path": "/gotify",
			// trunk-ignore(gitleaks/generic-api-key)
			"token": "AGE-LlHU89Q56uQ"},
		&shoutrrr.ShoutrrrDefaults{},
		&shoutrrr.ShoutrrrDefaults{},
		&shoutrrr.ShoutrrrDefaults{})
	notifier.ServiceStatus = &svcstatus.Status{}
	notifier.ServiceStatus.Init(
		0, 1, 0,
		test.StringPtr("testServiceID"),
		test.StringPtr("https://example.com"))
	notifier.Failed = &notifier.ServiceStatus.Fails.Shoutrrr
	if failing {
		notifier.URLFields["token"] = "invalid"
	}
	return notifier
}
