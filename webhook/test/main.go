// Copyright [2024] [Argus]
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

package test

import (
	"strings"

	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/webhook"
)

func WebHook(failing bool, selfSignedCert bool, customHeaders bool) *webhook.WebHook {
	desiredStatusCode := 0
	whMaxTries := uint(1)
	wh := webhook.New(
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
		&webhook.WebHookDefaults{},
		&webhook.WebHookDefaults{},
		&webhook.WebHookDefaults{})
	wh.ID = "test"
	wh.ServiceStatus = &svcstatus.Status{}
	wh.ServiceStatus.Init(
		0, 0, 1,
		test.StringPtr("testServiceID"),
		nil)
	wh.Failed = &wh.ServiceStatus.Fails.WebHook
	serviceName := "testServiceID"
	webURL := "https://example.com"
	wh.ServiceStatus.Init(
		0, 1, 0,
		&serviceName,
		&webURL)
	if selfSignedCert {
		wh.URL = strings.Replace(wh.URL, "valid", "invalid", 1)
	}
	if failing {
		wh.Secret = "invalid"
	}
	if customHeaders {
		wh.URL = strings.Replace(wh.URL, "github-style", "single-header", 1)
		if failing {
			wh.CustomHeaders = &webhook.Headers{
				{Key: "X-Test", Value: "invalid"}}
		} else {
			wh.CustomHeaders = &webhook.Headers{
				{Key: "X-Test", Value: "secret"}}
		}
	}

	// Slice to InitMetrics
	slice := webhook.Slice{"test": wh}
	slice.InitMetrics()

	return wh
}
