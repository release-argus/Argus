// Copyright [2025] [Argus]
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

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	logtest "github.com/release-argus/Argus/test/log"
)

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	// Exit.
	os.Exit(exitCode)
}

func testWebHook(failing bool, selfSignedCert bool, customHeaders bool) *WebHook {
	desiredStatusCode := uint16(0)
	whMaxTries := uint8(1)
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
		test.LookupGitHub["url_valid"],
		&Defaults{},
		&Defaults{}, &Defaults{})
	webhook.ID = "test"
	webhook.ServiceStatus = &status.Status{}
	webhook.ServiceStatus.Init(
		0, 1, 1,
		test.StringPtr("testServiceID"), nil,
		test.StringPtr("https://example.com"))
	webhook.Failed = &webhook.ServiceStatus.Fails.WebHook
	if selfSignedCert {
		webhook.URL = strings.Replace(webhook.URL, "valid", "invalid", 1)
	}
	if failing {
		webhook.Secret = "invalid"
	}
	if customHeaders {
		webhook.URL = strings.Replace(webhook.URL, "github-style", "single-header", 1)
		testHeaderValue := "secret"
		if failing {
			testHeaderValue = "invalid"
		}
		webhook.CustomHeaders = &Headers{
			{Key: "X-Test", Value: testHeaderValue}}
	}
	webhook.initMetrics()
	return webhook
}

func testDefaults(failing bool, customHeaders bool) *Defaults {
	desiredStatusCode := uint16(0)
	whMaxTries := uint8(1)
	webhook := NewDefaults(
		test.BoolPtr(false),
		nil,
		"0s",
		&desiredStatusCode,
		&whMaxTries,
		"argus",
		test.BoolPtr(false),
		"github",
		test.LookupGitHub["url_valid"])
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
