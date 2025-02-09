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

package test

import (
	"strings"

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/webhook"
)

func WebHook(failing, selfSignedCert, customHeaders bool) *webhook.WebHook {
	desiredStatusCode := uint16(0)
	whMaxTries := uint8(1)
	wh := webhook.New(
		test.BoolPtr(false),
		nil,
		"0s",
		&desiredStatusCode,
		nil,
		&whMaxTries,
		nil,
		test.StringPtr("12m"),
		test.LookupGitHub["secret_pass"],
		test.BoolPtr(false),
		"github",
		test.LookupGitHub["url_valid"],
		&webhook.Defaults{},
		&webhook.Defaults{}, &webhook.Defaults{})
	wh.ID = "test"
	wh.ServiceStatus = &status.Status{}
	serviceName := "testServiceID"
	wh.Failed = &wh.ServiceStatus.Fails.WebHook
	webURL := "https://example.com"
	wh.ServiceStatus.Init(
		0, 1, 0,
		&serviceName, &serviceName,
		&webURL)
	if selfSignedCert {
		wh.URL = strings.Replace(wh.URL, "valid", "invalid", 1)
	}
	if failing {
		wh.Secret = test.LookupGitHub["secret_fail"]
	}
	if customHeaders {
		wh.URL = strings.Replace(wh.URL, "github-style", "single-header", 1)
		if failing {
			wh.CustomHeaders = &webhook.Headers{
				{Key: test.LookupWithHeaders["header_key"], Value: test.LookupWithHeaders["header_value_fail"]}}
		} else {
			wh.CustomHeaders = &webhook.Headers{
				{Key: test.LookupWithHeaders["header_key"], Value: test.LookupWithHeaders["header_value_pass"]}}
		}
	}

	// Slice to InitMetrics.
	slice := webhook.Slice{"test": wh}
	slice.InitMetrics()

	return wh
}
