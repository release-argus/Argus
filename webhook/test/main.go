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

package test

import (
	"strings"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/webhook"
)

// WebHook returns a configured WebHook for tests.
func WebHook(failing, selfSignedCert, headers bool) *webhook.WebHook {
	defaults, _ := webhook.DecodeDefaults("yaml", nil)
	hardDefaults, _ := webhook.DecodeDefaults("yaml", nil)

	desiredStatusCode := uint16(0)
	whMaxTries := uint8(1)
	wh := webhook.New(
		test.Ptr(false),
		nil,
		"0s",
		&desiredStatusCode,
		nil,
		"test",
		&whMaxTries,
		webhook.Notifiers{},
		test.Ptr("12m"),
		test.WebHookGitHub["secret_pass"],
		test.Ptr(false),
		"github",
		test.WebHookGitHub["url_valid"],
		&webhook.Defaults{},
		defaults, hardDefaults,
	)
	wh.ServiceStatus = &status.Status{}
	serviceName := "testServiceID"
	wh.Failed = &wh.ServiceStatus.Fails.WebHook
	wh.ServiceStatus.Init(
		0, 1, 1,
		status.ServiceInfo{
			ID:         serviceName,
			Name:       serviceName,
			ServiceURL: "https://example.com/service/url",
		},
		&dashboard.Options{
			OptionsBase: dashboard.OptionsBase{
				WebURL: "https://example.com/web_url",
			},
		},
	)
	if selfSignedCert {
		wh.URL = strings.Replace(
			wh.URL,
			"valid",
			"invalid",
			1,
		)
	}
	if failing {
		wh.Secret = test.WebHookGitHub["secret_fail"]
	}
	if headers {
		wh.URL = strings.Replace(
			wh.URL,
			"github-style",
			"single-header",
			1,
		)
		if failing {
			wh.Headers = webhook.Headers{
				{
					Key:   test.LookupWithHeaderAuth["header_key"],
					Value: test.LookupWithHeaderAuth["header_value_fail"],
				},
			}
		} else {
			wh.Headers = webhook.Headers{
				{
					Key:   test.LookupWithHeaderAuth["header_key"],
					Value: test.LookupWithHeaderAuth["header_value_pass"],
				},
			}
		}
	}

	// WebHooks to InitMetrics.
	webhooks := webhook.WebHooks{"test": wh}
	webhooks.InitMetrics()

	return wh
}
