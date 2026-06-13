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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestHeaders_UnmarshalYAML(t *testing.T) {
	// GIVEN: a string to unmarshal as a Headers.
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
			expected: Headers{},
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
			name: "multiple map Headers, sorted input",
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
			name: "multiple map Headers, unsorted input - sorted output",
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
			name: "expected []Headers format YAML",
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

			// WHEN: the string is unmarshaled.
			var headers Headers
			err := decode.Unmarshal("yaml", []byte(tc.data), &headers)

			prefix := fmt.Sprintf(
				"%s\nHeaders.UnmarshalYAML(%q)",
				packageName, tc.data,
			)

			// THEN: we get an error if expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the Headers are as expected.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				headers,
				tc.expected,
				func(a, b Header) bool { return a.Key == b.Key && a.Value == b.Value },
				prefix,
				"Header",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestWebHook_SetHeaders(t *testing.T) {
	// GIVEN: a WebHook with Headers.
	latestVersion := "1.2.3"
	serviceID := "service"
	tests := []struct {
		name                                                 string
		env                                                  map[string]string
		rootValue, mainValue, defaultValue, hardDefaultValue Headers
		want                                                 map[string]string
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
			want: map[string]string{},
		},
		{
			name: "header with django expression",
			rootValue: Headers{
				{Key: "X-Service", Value: "{% if 'a' == 'a' %}foo{% endif %}"},
			},
			want: map[string]string{
				"X-Service": "foo",
			},
		},
		{
			name: "header with service ID var",
			rootValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
			},
			want: map[string]string{
				"X-Service": serviceID,
			},
		},
		{
			name: "header with version var",
			rootValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion,
			},
		},
		{
			name: "header from .Main",
			mainValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion,
			},
		},
		{
			name: "header from .Defaults",
			defaultValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion,
			},
		},
		{
			name: "header from .HardDefaults",
			hardDefaultValue: Headers{
				{Key: "X-Service", Value: "{{ service_id }}"},
				{Key: "X-Version", Value: "{{ version }}"},
			},
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion,
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
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion,
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
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion,
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
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion,
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
			want: map[string]string{
				"X-Service": serviceID,
				"X-Version": latestVersion,
				"X-Foo":     "_bar-",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)
			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			webhook := WebHook{
				ServiceStatus: &status.Status{},
				Main:          &Defaults{},
				Defaults:      &Defaults{},
				HardDefaults:  &Defaults{},
			}
			webhook.ServiceStatus.ServiceInfo.ID = serviceID
			url := "https://example.com"
			webhook.ServiceStatus.Init(
				0, 0, 0,
				status.ServiceInfo{
					ID: serviceID,
				},
				&dashboard.Options{
					OptionsBase: dashboard.OptionsBase{
						WebURL: url,
					},
				},
			)
			webhook.ServiceStatus.SetLatestVersion(latestVersion, "", false)
			webhook.Headers = tc.rootValue
			webhook.Main.Headers = tc.mainValue
			webhook.Defaults.Headers = tc.defaultValue
			webhook.HardDefaults.Headers = tc.hardDefaultValue

			// WHEN: setHeaders is called on this request.
			webhook.setHeaders(req)

			prefix := fmt.Sprintf(
				"%s\nWebHook.SetHeaders(%+v)",
				packageName, webhook.Headers,
			)

			// THEN: the function returns the correct result.
			if tc.rootValue == nil && tc.mainValue == nil && tc.defaultValue == nil && tc.hardDefaultValue == nil {
				if len(req.Header) != 0 {
					t.Fatalf(
						"%s WebHook.Headers are nil but request.Headers aren't\ngot: %v",
						prefix, req.Header,
					)
				}
				return
			}
			if tc.want == nil {
				for _, header := range tc.rootValue {
					tc.want[header.Key] = header.Value
				}
			}
			for header, val := range tc.want {
				if req.Header[header] == nil {
					t.Fatalf(
						"%s key=%q, value=%q was not given to the request\ngot: headers=%+v",
						prefix,
						header, val,
						req.Header,
					)
				}
				if req.Header[header][0] != val {
					t.Fatalf(
						"%s key=%q, value=%q was not given to the request\ngot: value=%q, headers=%+v",
						prefix,
						header, val,
						req.Header[header][0], req.Header,
					)
				}
			}
		})
	}
}
