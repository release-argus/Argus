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

func TestSetCustomHeaders(t *testing.T) {
	// GIVEN a WebHook with CustomHeaders
	latestVersion := "1.2.3"
	serviceID := "service"
	tests := map[string]struct {
		customHeaders     map[string]string
		mainCustomHeaders map[string]string
		wantHeaders       map[string]string
	}{
		"no customHeaders": {customHeaders: map[string]string{}},
		"standard headers": {customHeaders: map[string]string{"X-Something": "foo", "X-Service": "test"}},
		"header with jinja expression": {customHeaders: map[string]string{"X-Service": "{% if 'a' == 'a' %}foo{% endif %}"},
			wantHeaders: map[string]string{"X-Service": "foo"}},
		"header with service id var": {customHeaders: map[string]string{"X-Service": "{{ service_id }}"},
			wantHeaders: map[string]string{"X-Service": serviceID}},
		"header with version var": {customHeaders: map[string]string{"X-Service": "{{ service_id }}", "X-Version": "{{ version }}"},
			wantHeaders: map[string]string{"X-Service": serviceID, "X-Version": latestVersion}},
		"header from .Main": {mainCustomHeaders: map[string]string{"X-Service": "{{ service_id }}", "X-Version": "{{ version }}"},
			wantHeaders: map[string]string{"X-Service": serviceID, "X-Version": latestVersion}},
		"header from .CustomHeaders overrides those in .Main": {customHeaders: map[string]string{"X-Service": "{{ service_id }}", "X-Version": "{{ version }}"},
			mainCustomHeaders: map[string]string{"X-Service": "{{ service_id }}", "X-Version": "----"},
			wantHeaders:       map[string]string{"X-Service": serviceID, "X-Version": latestVersion}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			webhook := WebHook{ServiceStatus: &svcstatus.Status{ServiceID: &serviceID, LatestVersion: latestVersion},
				Main:     &WebHook{},
				Defaults: &WebHook{}}
			if tc.customHeaders != nil {
				webhook.CustomHeaders = &tc.customHeaders
			}
			if tc.mainCustomHeaders != nil {
				webhook.Main.CustomHeaders = &tc.mainCustomHeaders
			}

			// WHEN SetCustomHeaders is called on this request
			webhook.SetCustomHeaders(req)

			// THEN the function returns the correct result
			if tc.customHeaders == nil && tc.mainCustomHeaders == nil {
				if len(req.Header) != 0 {
					t.Fatalf("custom headers was nil but Headers are %v",
						req.Header)
				}
				return
			}
			if tc.wantHeaders == nil {
				tc.wantHeaders = tc.customHeaders
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
