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
		root   *Headers
		main   *Headers
		dfault *Headers
		want   map[string]string
	}{
		"nil root": {
			root: nil,
		},
		"no root": {
			root: &Headers{},
		},
		"standard headers": {
			root: &Headers{
				{Key: "X-Something", Value: "foo"},
				{Key: "X-Service", Value: "test"}},
			want: map[string]string{},
		},
		"header with jinja expression": {
			root: &Headers{
				{Key: "X-Service", Value: "{% if 'a' == 'a' %}foo{% endif %}"}},
			want: map[string]string{
				"X-Service": "foo"},
		},
		"header with service id var": {
			root: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"}},
			want: map[string]string{
				"X-Service": serviceID},
		},
		"header with version var": {
			root: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .Main": {
			main: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .Defaults": {
			dfault: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .CustomHeaders overrides those in .Main": {
			root: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			main: &Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .CustomHeaders overrides those in .Defaults": {
			root: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			dfault: &Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .Main overrides those in .Defaults": {
			main: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			dfault: &Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			webhook := WebHook{
				ServiceStatus: &svcstatus.Status{
					ServiceID: &serviceID},
				Main:     &WebHookDefaults{},
				Defaults: &WebHookDefaults{}}
			url := "https://example.com"
			webhook.ServiceStatus.Init(
				0, 0, 0,
				&serviceID,
				&url)
			webhook.ServiceStatus.SetLatestVersion(latestVersion, false)
			webhook.CustomHeaders = tc.root
			webhook.Main.CustomHeaders = tc.main
			webhook.Defaults.CustomHeaders = tc.dfault

			// WHEN setCustomHeaders is called on this request
			webhook.setCustomHeaders(req)

			// THEN the function returns the correct result
			if tc.root == nil && tc.main == nil && tc.dfault == nil {
				if len(req.Header) != 0 {
					t.Fatalf("custom headers was nil but Headers are %v",
						req.Header)
				}
				return
			}
			if tc.want == nil {
				for _, header := range *tc.root {
					tc.want[header.Key] = header.Value
				}
			}
			for header, val := range tc.want {
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
