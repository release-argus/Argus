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

//go:build unit

package webhook

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
)

func TestWebHook_SetCustomHeaders(t *testing.T) {
	// GIVEN a WebHook with CustomHeaders.
	latestVersion := "1.2.3"
	serviceID := "service"
	tests := map[string]struct {
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue *Headers
		want                                                 map[string]string
	}{
		"nil root": {
			rootValue: nil,
		},
		"no root": {
			rootValue: &Headers{},
		},
		"standard headers": {
			rootValue: &Headers{
				{Key: "X-Something", Value: "foo"},
				{Key: "X-Service", Value: "test"}},
			want: map[string]string{},
		},
		"header with django expression": {
			rootValue: &Headers{
				{Key: "X-Service", Value: "{% if 'a' == 'a' %}foo{% endif %}"}},
			want: map[string]string{
				"X-Service": "foo"},
		},
		"header with service ID var": {
			rootValue: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"}},
			want: map[string]string{
				"X-Service": serviceID},
		},
		"header with version var": {
			rootValue: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .Main": {
			mainValue: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .Defaults": {
			defaultValue: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .HardDefaults": {
			hardDefaultValue: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .CustomHeaders overrides those in .Main": {
			rootValue: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			mainValue: &Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .CustomHeaders overrides those in .Defaults": {
			rootValue: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			defaultValue: &Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header from .Main overrides those in .Defaults": {
			mainValue: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"}},
			defaultValue: &Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion},
		},
		"header with env var": {
			env: map[string]string{"FOO": "bar"},
			rootValue: &Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
				{Key: "X-Foo", Value: "_${FOO}-"}},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion,
				"X-Foo":     "_bar-"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}
			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			webhook := WebHook{
				ServiceStatus: &status.Status{},
				Main:          &Defaults{},
				Defaults:      &Defaults{},
				HardDefaults:  &Defaults{}}
			webhook.ServiceStatus.ServiceInfo.ID = serviceID
			url := "https://example.com"
			webhook.ServiceStatus.Init(
				0, 0, 0,
				serviceID, "", "",
				&dashboard.Options{
					WebURL: url})
			webhook.ServiceStatus.SetLatestVersion(latestVersion, "", false)
			webhook.CustomHeaders = tc.rootValue
			webhook.Main.CustomHeaders = tc.mainValue
			webhook.Defaults.CustomHeaders = tc.defaultValue
			webhook.HardDefaults.CustomHeaders = tc.hardDefaultValue

			// WHEN setCustomHeaders is called on this request.
			webhook.setCustomHeaders(req)

			// THEN the function returns the correct result.
			if tc.rootValue == nil && tc.mainValue == nil && tc.defaultValue == nil && tc.hardDefaultValue == nil {
				if len(req.Header) != 0 {
					t.Fatalf("%s\ncustom headers is nil but Headers are %v",
						packageName, req.Header)
				}
				return
			}
			if tc.want == nil {
				for _, header := range *tc.rootValue {
					tc.want[header.Key] = header.Value
				}
			}
			for header, val := range tc.want {
				if req.Header[header] == nil {
					t.Fatalf("%s\n%s: %s was not given to the request, got\n%v",
						packageName,
						header, val,
						req.Header)
				}
				if req.Header[header][0] != val {
					t.Fatalf("%s\n%s: %s was not given to the request, got\n%v\n%v",
						packageName,
						header, val,
						req.Header[header][0], req.Header)
				}
			}
		})
	}
}
