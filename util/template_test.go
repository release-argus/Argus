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

package util

import (
	"fmt"
	"regexp"
	"testing"
)

func TestTemplate_String(t *testing.T) {
	// GIVEN a variety of string templates
	serviceInfo := testServiceInfo()
	tests := map[string]struct {
		tmpl       string
		panicRegex *string
		want       string
	}{
		"no jinja template": {
			tmpl: "testing 123",
			want: "testing 123"},
		"valid jinja template": {
			tmpl: "-{% if 'a' == 'a' %}{{ service_id }}{% endif %}-{{ service_url }}-{{ web_url }}-{{ version }}",
			want: "-something-example.com-other.com-NEW"},
		"invalid jinja template panic": {
			tmpl:       "-{% 'a' == 'a' %}{{ service_id }}{% endif %}-{{ service_url }}-{{ web_url }}-{{ version }}",
			panicRegex: stringPtr("Tag name must be an identifier")},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.panicRegex != nil {
				// Switch Fatal to panic and disable this panic.
				defer func() {
					r := recover()
					rStr := fmt.Sprint(r)
					re := regexp.MustCompile(*tc.panicRegex)
					match := re.MatchString(rStr)
					if !match {
						t.Errorf("expected a panic that matched %q\ngot: %q",
							*tc.panicRegex, rStr)
					}
				}()
			}

			// WHEN TemplateString is called
			got := TemplateString(tc.tmpl, serviceInfo)

			// THEN the string stays the same
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestCheckTemplate(t *testing.T) {
	// GIVEN a variety of string templates
	tests := map[string]struct {
		tmpl string
		pass bool
	}{
		"no jinja template":            {tmpl: "testing 123", pass: true},
		"valid jinja template":         {tmpl: "{{ version }}-foo", pass: true},
		"invalid jinja template panic": {tmpl: "{{ version }", pass: false},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckTemplate is called
			got := CheckTemplate(tc.tmpl)

			// THEN the string stays the same
			if got != tc.pass {
				t.Errorf("want: %t\ngot:  %t",
					tc.pass, got)
			}
		})
	}
}
