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

//go:build unit

package webhook

import (
	"net/http"
	"net/http/httptest"
	"testing"

	svcstatus "github.com/release-argus/Argus/service/status"
)

func TestWebHook_SetCustomHeaders(t *testing.T) {
	// GIVEN a WebHook with CustomHeaders
	latestVersion := "1.2.3"
	serviceID := "service"
	tests := map[string]struct {
		customHeaders        *Headers
		mainCustomHeaders    *Headers
		defaultCustomHeaders *Headers
		wantHeaders          map[string]string
	}{
		"nil customHeaders": {
			customHeaders: nil,
		},
		"no customHeaders": {
			customHeaders: &Headers{},
		},
		"standard headers": {
			customHeaders: &Headers{
				{Key: "X-Something", Value: "foo"},
				{Key: "X-Service", Value: "test"}},
			wantHeaders: map[string]string{},
		},
		"header with jinja expression": {
			customHeaders: &Headers{
				{Key: "X-Service", Value: "{% if 'a' == 'a' %}foo{% endif %}"}},
			wantHeaders: map[string]string{
				"X-Service": "foo"},
		},
		"header with service id var": {
			customHeaders: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"}},
			wantHeaders: map[string]string{
				"X-Service": serviceID},
		},
		"header with version var": {
			customHeaders: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			wantHeaders: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .Main": {
			mainCustomHeaders: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			wantHeaders: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .Defaults": {
			defaultCustomHeaders: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			wantHeaders: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .CustomHeaders overrides those in .Main": {
			customHeaders: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			mainCustomHeaders: &Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"}},
			wantHeaders: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .CustomHeaders overrides those in .Defaults": {
			customHeaders: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			defaultCustomHeaders: &Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"}},
			wantHeaders: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .Main overrides those in .Defaults": {
			mainCustomHeaders: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			defaultCustomHeaders: &Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"}},
			wantHeaders: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			webhook := WebHook{
				ServiceStatus: &svcstatus.Status{
					ServiceID: &serviceID},
				Main:     &WebHook{},
				Defaults: &WebHook{}}
			url := "https://example.com"
			webhook.ServiceStatus.Init(
				0, 0, 0,
				&serviceID,
				&url)
			webhook.ServiceStatus.SetLatestVersion(latestVersion, false)
			webhook.CustomHeaders = tc.customHeaders
			webhook.Main.CustomHeaders = tc.mainCustomHeaders
			webhook.Defaults.CustomHeaders = tc.defaultCustomHeaders

			// WHEN setCustomHeaders is called on this request
			webhook.setCustomHeaders(req)

			// THEN the function returns the correct result
			if tc.customHeaders == nil && tc.mainCustomHeaders == nil && tc.defaultCustomHeaders == nil {
				if len(req.Header) != 0 {
					t.Fatalf("custom headers was nil but Headers are %v",
						req.Header)
				}
				return
			}
			if tc.wantHeaders == nil {
				for _, header := range *tc.customHeaders {
					tc.wantHeaders[header.Key] = header.Value
				}
			}
			for header, val := range tc.wantHeaders {
				if req.Header[header] == nil {
					t.Fatalf("%s: %s was not given to the request, got\n%v",
						header, val, req.Header)
				}
				if req.Header[header][0] != val {
					t.Fatalf("%s: %s was not given to the request, got\n%v\n%v",
						header, val, req.Header[header][0], req.Header)
				}
			}
		})
	}
}
