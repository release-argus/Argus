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

package util

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestTemplate_String(t *testing.T) {
	// GIVEN a variety of string templates.
	tests := map[string]struct {
		template    string
		serviceInfo ServiceInfo
		panicRegex  *string
		want        string
	}{
		"no django template": {
			template:    "testing 123",
			want:        "testing 123",
			serviceInfo: testServiceInfo()},
		"valid django template": {
			template:    "-{% if 'a' == 'a' %}{{ service_id }}{% endif %}-{{ service_url }}-{{ web_url }}-{{ version }}",
			want:        "-something-example.com-other.com-NEW",
			serviceInfo: testServiceInfo()},
		"valid django template with defaulting - had value": {
			template:    "{{ service_name | default:service_id }} - {{ version }} released",
			want:        "another - NEW released",
			serviceInfo: testServiceInfo()},
		"valid django template with defaulting - had no value (empty string)": {
			template: "{{ service_name | default:service_id }} - {{ version }} released",
			want:     "something - NEW released",
			serviceInfo: ServiceInfo{
				ID:            "something",
				Name:          "",
				URL:           "example.com",
				WebURL:        test.StringPtr("other.com"),
				LatestVersion: "NEW",
			}},
		"valid django template with defaulting - had no value (nil)": {
			template: "{{ service_name | default:service_id }} - {{ web_url }}",
			want:     "else - ",
			serviceInfo: ServiceInfo{
				ID:            "something",
				Name:          "else",
				URL:           "example.com",
				WebURL:        nil,
				LatestVersion: "NEW",
			}},
		"invalid django template panic": {
			template:    "-{% 'a' == 'a' %}{{ service_id }}{% endif %}-{{ service_url }}-{{ web_url }}-{{ version }}",
			panicRegex:  test.StringPtr("Tag name must be an identifier"),
			serviceInfo: testServiceInfo()},
		"all django vars": {
			template:    "{{ service_id }}-{{ service_name }}-{{ service_url }}-{{ web_url }}-{{ version }}",
			want:        "something-another-example.com-other.com-NEW",
			serviceInfo: testServiceInfo()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.panicRegex != nil {
				// Switch Fatal to panic and disable this panic.
				defer func() {
					r := recover()

					rStr := fmt.Sprint(r)
					if !RegexCheck(*tc.panicRegex, rStr) {
						t.Errorf("%s\npanic mismatch\nwant: %q\ngot:  %q",
							packageName, *tc.panicRegex, rStr)
					}
				}()
			}

			// WHEN TemplateString is called.
			got := TemplateString(tc.template, tc.serviceInfo)

			// THEN the string stays the same.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestCheckTemplate(t *testing.T) {
	// GIVEN a variety of string templates.
	tests := map[string]struct {
		template string
		pass     bool
	}{
		"no django template":            {template: "testing 123", pass: true},
		"valid django template":         {template: "{{ version }}-foo", pass: true},
		"invalid django template panic": {template: "{{ version }", pass: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckTemplate is called.
			got := CheckTemplate(tc.template)

			// THEN the string stays the same.
			if got != tc.pass {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.pass, got)
			}
		})
	}
}
