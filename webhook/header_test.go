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

//go:build unit

package webhook

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestHeaders_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		errRegex string
		expected Headers
	}{
		{
			name:     "empty",
			data:     "",
			errRegex: `^$`,
			expected: nil,
		},
		{
			name:     "single map Header",
			data:     "foo: bar",
			errRegex: `^$`,
			expected: Headers{
				{Key: "foo", Value: "bar"},
			},
		},
		{
			name: "multiple map Headers/sorted input",
			data: test.TrimYAML(`
				bish: bash
				bosh: boom
				foo: bar
			`),
			errRegex: `^$`,
			expected: Headers{
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "boom"},
				{Key: "foo", Value: "bar"},
			},
		},
		{
			name: "multiple map Headers/unsorted input gives sorted output",
			data: test.TrimYAML(`
				foo: bar
				bish: bash
				bosh: boom`,
			),
			errRegex: `^$`,
			expected: Headers{
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "boom"},
				{Key: "foo", Value: "bar"},
			},
		},
		{
			name: "list format",
			data: test.TrimYAML(`
				- key: foo
					value: bar
				- key: bish
					value: bash
				- key: bosh
					value: boom`,
			),
			errRegex: `^$`,
			expected: Headers{
				{Key: "foo", Value: "bar"},
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "boom"},
			},
		},
		{
			name: "invalid YAML",
			data: "notMappingHere",
			errRegex: test.TrimYAML(`
				^[^\s]+.*string was used where mapping is expected
				[^\s]+ .* notMappingHere.*
				\s+\^$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var headers Headers
			err := decode.Unmarshal("yaml", []byte(tc.data), &headers)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if testErr := test.AssertSlicesEqual(
				t,
				headers,
				tc.expected,
				"Headers",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestWebhook_SetHeaders(t *testing.T) {
	latestVersion := "1.2.3"
	serviceID := "service"
	tests := []struct {
		name                                                 string
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue Headers
		want                                                 http.Header
	}{
		{
			name:      "nil root",
			rootValue: nil,
		},
		{
			name:      "no root",
			rootValue: Headers{},
		},
		{
			name: "standard headers",
			rootValue: Headers{
				{Key: "X-Something", Value: "foo"},
				{Key: "X-Service", Value: "test"},
			},
			want: http.Header{
				"X-Something": {"foo"},
				"X-Service":   {"test"},
			},
		},
		{
			name: "header with django expression",
			rootValue: Headers{
				{Key: "X-Service", Value: "{% if 'a' == 'a' %}foo{% endif %}"},
			},
			want: http.Header{
				"X-Service": {"foo"},
			},
		},
		{
			name: "header with service ID var",
			rootValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
			},
			want: http.Header{
				"X-Service": {serviceID},
			},
		},
		{
			name: "header with version var",
			rootValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			want: http.Header{
				"X-Service": {serviceID},
				"X-Version": {latestVersion},
			},
		},
		{
			name: "header from .Main",
			mainValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			want: http.Header{
				"X-Service": {serviceID},
				"X-Version": {latestVersion},
			},
		},
		{
			name: "header from .Defaults",
			defaultValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			want: http.Header{
				"X-Service": {serviceID},
				"X-Version": {latestVersion},
			},
		},
		{
			name: "header from .HardDefaults",
			hardDefaultValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			want: http.Header{
				"X-Service": {serviceID},
				"X-Version": {latestVersion},
			},
		},
		{
			name: "header from .Headers overrides those in .Main",
			rootValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			mainValue: Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"},
			},
			want: http.Header{
				"X-Service": {serviceID},
				"X-Version": {latestVersion},
			},
		},
		{
			name: "header from .Headers overrides those in .Defaults",
			rootValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			defaultValue: Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"},
			},
			want: http.Header{
				"X-Service": {serviceID},
				"X-Version": {latestVersion},
			},
		},
		{
			name: "header from .Main overrides those in .Defaults",
			mainValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			defaultValue: Headers{
				{Key: "X-Service", Value: "===="},
				{Key: "X-Version", Value: "----"},
			},
			want: http.Header{
				"X-Service": {serviceID},
				"X-Version": {latestVersion},
			},
		},
		{
			name: "header with env var",
			env: map[string]string{
				"FOO": "bar",
			},
			rootValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
				{Key: "X-Foo", Value: "_${FOO}-"},
			},
			want: http.Header{
				"X-Service": {serviceID},
				"X-Version": {latestVersion},
				"X-Foo":     {"_bar-"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.IsolateEnv(t, tc.env)
			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			webhook := Webhook{
				ServiceStatus: &status.Status{},
				Main:          &Defaults{},
				Defaults:      &Defaults{},
				HardDefaults:  &Defaults{},
			}
			webhook.ServiceStatus.Init(
				0, 0, 0,
				status.ServiceInfo{
					ID: serviceID,
				},
				&dashboard.Options{
					WebURL: "https://example.com",
				},
			)
			webhook.ServiceStatus.SetLatestVersion(latestVersion, "", false)
			webhook.Headers = tc.rootValue
			webhook.Main.Headers = tc.mainValue
			webhook.Defaults.Headers = tc.defaultValue
			webhook.HardDefaults.Headers = tc.hardDefaultValue

			webhook.setHeaders(req)

			if testErr := test.AssertMapEqualFunc(
				t,
				req.Header,
				tc.want,
				slices.Equal,
				"http.Request.Header",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
