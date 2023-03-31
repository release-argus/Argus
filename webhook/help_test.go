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
	"fmt"
	"strings"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func boolPtr(val bool) *bool {
	return &val
}
func intPtr(val int) *int {
	return &val
}
func stringPtr(val string) *string {
	return &val
}
func uintPtr(val int) *uint {
	converted := uint(val)
	return &converted
}
func stringifyPointer[T comparable](ptr *T) string {
	str := "nil"
	if ptr != nil {
		str = fmt.Sprint(*ptr)
	}
	return str
}
func testLogging(level string) {
	jLog = util.NewJLog(level, false)
	jLog.Testing = true
	shoutrrr.LogInit(jLog)
}

func testWebHook(failing bool, forService bool, selfSignedCert bool, customHeaders bool) *WebHook {
	desiredStatusCode := 0
	whMaxTries := uint(1)
	webhook := &WebHook{
		Type:              "github",
		URL:               "https://valid.release-argus.io/hooks/github-style",
		Secret:            "argus",
		AllowInvalidCerts: boolPtr(false),
		DesiredStatusCode: &desiredStatusCode,
		Delay:             "0s",
		SilentFails:       boolPtr(false),
		MaxTries:          &whMaxTries,
	}
	if forService {
		webhook.ID = "test"
		webhook.ParentInterval = stringPtr("12m")
		webhook.ServiceStatus = &svcstatus.Status{}
		webhook.ServiceStatus.Fails.WebHook = make(map[string]*bool, 1)
		webhook.Failed = &webhook.ServiceStatus.Fails.WebHook
		webhook.Main = &WebHook{}
		webhook.Defaults = &WebHook{}
		webhook.HardDefaults = &WebHook{}
	}
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

func testNotifier(failing bool, selfSignedCert bool) *shoutrrr.Shoutrrr {
	url := "valid.release-argus.io"
	if selfSignedCert {
		url = strings.Replace(url, "valid", "invalid", 1)
	}
	notifier := &shoutrrr.Shoutrrr{
		Type:   "gotify",
		ID:     "test",
		Failed: nil,
		ServiceStatus: &svcstatus.Status{
			ServiceID: stringPtr("service"),
			Fails:     svcstatus.Fails{Shoutrrr: make(map[string]*bool, 1)}},
		Main:         &shoutrrr.Shoutrrr{},
		Defaults:     &shoutrrr.Shoutrrr{},
		HardDefaults: &shoutrrr.Shoutrrr{},
		Options:      map[string]string{"max_tries": "1"},
		// trunk-ignore(gitleaks/generic-api-key)
		URLFields: map[string]string{"host": url, "path": "/gotify", "token": "AGE-LlHU89Q56uQ"},
		Params:    map[string]string{},
	}
	notifier.Failed = &notifier.ServiceStatus.Fails.Shoutrrr
	if failing {
		notifier.URLFields["token"] = "invalid"
	}
	return notifier
}
