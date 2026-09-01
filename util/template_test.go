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

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	serviceinfo "github.com/release-argus/Argus/service/status/info"
)

func TestTemplate_String(t *testing.T) {
	svcInfo := testServiceInfo()
	// GIVEN: a variety of string templates.
	tests := []struct {
		name        string
		template    string
		serviceInfo serviceinfo.ServiceInfo
		want        string
	}{
		{
			name:        "no django template",
			template:    "testing 123",
			want:        "testing 123",
			serviceInfo: svcInfo,
		},
		{
			name:     "valid django template",
			template: "-{% if 'a' == 'a' %}{{ service_id }}{% endif %}-{{ service_url }}-{{ web_url }}-{{ version }}",
			want: fmt.Sprintf(
				"-%s-%s-%s-%s",
				svcInfo.ID, svcInfo.URL, svcInfo.WebURL, svcInfo.LatestVersion,
			),
			serviceInfo: svcInfo,
		},
		{
			name:     "valid django template with defaulting/had value",
			template: "{{ service_name | default:service_id }} - {{ version }} released",
			want: fmt.Sprintf(
				"%s - %s released",
				svcInfo.Name, svcInfo.LatestVersion,
			),
			serviceInfo: svcInfo,
		},
		{
			name:     "valid django template with defaulting/had no value/empty string",
			template: "{{ service_name | default:service_id }} - {{ version }} released",
			want: fmt.Sprintf(
				"%s - %s released",
				svcInfo.ID, svcInfo.LatestVersion,
			),
			serviceInfo: serviceinfo.ServiceInfo{
				ID:            svcInfo.ID,
				Name:          "",
				URL:           svcInfo.URL,
				WebURL:        svcInfo.WebURL,
				LatestVersion: svcInfo.LatestVersion,
			},
		},
		{
			name:     "valid django template with defaulting/had no value/nil",
			template: "{{ service_name | default:service_id }} - {{ web_url }}",
			want:     fmt.Sprintf(" - %s", svcInfo.WebURL),
			serviceInfo: serviceinfo.ServiceInfo{
				ID:            "",
				Name:          "",
				URL:           svcInfo.URL,
				WebURL:        svcInfo.WebURL,
				LatestVersion: svcInfo.LatestVersion,
			},
		},
		{
			name:     "valid django template with array access",
			template: "{{ tags | first }}-{{ tags.0 }}_{{ tags|slice:'1:2'|first }}-{{ tags | last }}-{{ tags.1 }}_{{ tags | join:',' }}",
			want: fmt.Sprintf(
				"%s-%s_%s-%s-%s_%s,%s",
				svcInfo.Tags[0], svcInfo.Tags[0],
				svcInfo.Tags[1], svcInfo.Tags[1], svcInfo.Tags[1],
				svcInfo.Tags[0], svcInfo.Tags[1],
			),
			serviceInfo: svcInfo,
		},
		{
			name:     "valid django template with array access out of bounds",
			template: "{{ tags.0 }}-{{ tags.1 }}-{{ tags.2 }}-{{ tags.3 }}",
			want: fmt.Sprintf(
				"%s-%s--",
				svcInfo.Tags[0], svcInfo.Tags[1],
			),
			serviceInfo: svcInfo,
		},
		{
			name:        "uncompilable/template is returned unchanged",
			template:    "-{% 'a' == 'a' %}{{ service_id }}{% endif %}-{{ service_url }}-{{ web_url }}-{{ version }}",
			want:        "-{% 'a' == 'a' %}{{ service_id }}{% endif %}-{{ service_url }}-{{ web_url }}-{{ version }}",
			serviceInfo: svcInfo,
		},
		{
			name:        "unrenderable/template is returned unchanged",
			template:    "{{ tags.0.bar }}",
			want:        "{{ tags.0.bar }}",
			serviceInfo: svcInfo,
		},
		{
			name:        "unrenderable/a render panic does not escape",
			template:    "{{ 1/0 }}",
			want:        "{{ 1/0 }}",
			serviceInfo: svcInfo,
		},
		{
			name:     "all django vars",
			template: "{{ service_id }}-{{ service_name }}-{{ service_url }}--{{ icon }}-{{ icon_link_to }}-{{ web_url }}--{{ version }}-{{ approved_version }}-{{ deployed_version }}-{{ latest_version }}-{{ tags|first }}-{{ tags.1 }}",
			want: fmt.Sprintf(
				"%s-%s-%s--%s-%s-%s--%s-%s-%s-%s-%s-%s",
				svcInfo.ID, svcInfo.Name, svcInfo.URL,
				svcInfo.Icon, svcInfo.IconLinkTo, svcInfo.WebURL,
				svcInfo.LatestVersion, svcInfo.ApprovedVersion, svcInfo.DeployedVersion, svcInfo.LatestVersion,
				svcInfo.Tags[0], svcInfo.Tags[1],
			),
			serviceInfo: svcInfo,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf(
				"%s\nTemplateString(tmpl=%q, info=%+v)",
				packageName, tc.template, tc.serviceInfo,
			)

			// WHEN: TemplateString is called.
			got := TemplateString(tc.template, tc.serviceInfo)

			// THEN: the string stays the same.
			if got != tc.want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestCheckTemplate(t *testing.T) {
	// GIVEN: a variety of string templates.
	tests := []struct {
		name     string
		template string
		pass     bool
	}{
		{
			name:     "no django template",
			template: "testing 123",
			pass:     true,
		},
		{
			name:     "valid django template",
			template: "{{ version }}-foo",
			pass:     true,
		},
		{
			name:     "invalid django template panic",
			template: "{{ version }",
			pass:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: CheckTemplate is called.
			got := CheckTemplate(tc.template)

			// THEN: the string stays the same.
			if got != tc.pass {
				t.Errorf(
					"%s\nCheckTemplate(%q) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.template,
					got, tc.pass,
				)
			}
		})
	}
}

func TestTemplateString__cannotReadLocalFiles(t *testing.T) {
	// GIVEN: a readable file holding a secret, and templates that try to pull
	// it in.
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret.txt")
	const contents = "TOP-SECRET-TOKEN"
	if err := os.WriteFile(secret, []byte(contents), 0o600); err != nil {
		t.Fatalf(
			"%s\nwrite fixture: %v",
			packageName, err,
		)
	}

	tests := []struct {
		name     string
		template string
	}{
		{name: "denied/include", template: `{% include "` + secret + `" %}`},
		{name: "denied/extends", template: `{% extends "` + secret + `" %}`},
		{name: "denied/ssi", template: `{% ssi "` + secret + `" %}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prefix := fmt.Sprintf(
				"%s\nTemplateString(tmpl=%q)",
				packageName, tc.template,
			)

			// WHEN: the template is rendered.
			got := TemplateString(tc.template, serviceinfo.ServiceInfo{})

			// THEN: the file's contents never reach the output.
			if strings.Contains(got, contents) {
				t.Errorf(
					"%s read a local file\ngot: %q",
					prefix, got,
				)
			}
			// AND: it degrades to the template rather than erroring out.
			if got != tc.template {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.template,
				)
			}
		})
	}
}
