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

	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/test"
)

func TestTemplate_String(t *testing.T) {
	svcInfo := testServiceInfo()
	// GIVEN a variety of string templates.
	tests := map[string]struct {
		template    string
		serviceInfo serviceinfo.ServiceInfo
		panicRegex  *string
		want        string
	}{
		"no django template": {
			template:    "testing 123",
			want:        "testing 123",
			serviceInfo: svcInfo},
		"valid django template": {
			template: "-{% if 'a' == 'a' %}{{ service_id }}{% endif %}-{{ service_url }}-{{ web_url }}-{{ version }}",
			want: fmt.Sprintf("-%s-%s-%s-%s",
				svcInfo.ID, svcInfo.URL, svcInfo.WebURL, svcInfo.LatestVersion),
			serviceInfo: svcInfo},
		"valid django template with defaulting - had value": {
			template: "{{ service_name | default:service_id }} - {{ version }} released",
			want: fmt.Sprintf("%s - %s released",
				svcInfo.Name, svcInfo.LatestVersion),
			serviceInfo: svcInfo},
		"valid django template with defaulting - had no value (empty string)": {
			template: "{{ service_name | default:service_id }} - {{ version }} released",
			want: fmt.Sprintf("%s - %s released",
				svcInfo.ID, svcInfo.LatestVersion),
			serviceInfo: serviceinfo.ServiceInfo{
				ID:            svcInfo.ID,
				Name:          "",
				URL:           svcInfo.URL,
				WebURL:        svcInfo.WebURL,
				LatestVersion: svcInfo.LatestVersion},
		},
		"valid django template with defaulting - had no value (nil)": {
			template: "{{ service_name | default:service_id }} - {{ web_url }}",
			want: fmt.Sprintf("%s - %s",
				"", svcInfo.WebURL),
			serviceInfo: serviceinfo.ServiceInfo{
				ID:            "",
				Name:          "",
				URL:           svcInfo.URL,
				WebURL:        svcInfo.WebURL,
				LatestVersion: svcInfo.LatestVersion},
		},
		"valid django template with array access": {
			template: "{{ tags | first }}-{{ tags.0 }}_{{ tags|slice:'1:2'|first }}-{{ tags | last }}-{{ tags.1 }}_{{ tags | join:',' }}",
			want: fmt.Sprintf("%s-%s_%s-%s-%s_%s,%s",
				svcInfo.Tags[0], svcInfo.Tags[0],
				svcInfo.Tags[1], svcInfo.Tags[1], svcInfo.Tags[1],
				svcInfo.Tags[0], svcInfo.Tags[1]),
			serviceInfo: svcInfo},
		"valid django template with array access out of bounds": {
			template: "{{ tags.0 }}-{{ tags.1 }}-{{ tags.2 }}-{{ tags.3 }}",
			want: fmt.Sprintf("%s-%s--",
				svcInfo.Tags[0], svcInfo.Tags[1]),
			serviceInfo: svcInfo},
		"invalid django template panic": {
			template:    "-{% 'a' == 'a' %}{{ service_id }}{% endif %}-{{ service_url }}-{{ web_url }}-{{ version }}",
			panicRegex:  test.StringPtr("Tag name must be an identifier"),
			serviceInfo: svcInfo},
		"all django vars": {
			template: "{{ service_id }}-{{ service_name }}-{{ service_url }}--{{ icon }}-{{ icon_link_to }}-{{ web_url }}--{{ version }}-{{ approved_version }}-{{ deployed_version }}-{{ latest_version }}-{{ tags|first }}-{{ tags.1 }}",
			want: fmt.Sprintf("%s-%s-%s--%s-%s-%s--%s-%s-%s-%s-%s-%s",
				svcInfo.ID, svcInfo.Name, svcInfo.URL,
				svcInfo.Icon, svcInfo.IconLinkTo, svcInfo.WebURL,
				svcInfo.LatestVersion, svcInfo.ApprovedVersion, svcInfo.DeployedVersion, svcInfo.LatestVersion,
				svcInfo.Tags[0], svcInfo.Tags[1]),
			serviceInfo: svcInfo},
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
