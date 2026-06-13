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
	"regexp"
	"testing"
)

func TestRegexCheck(t *testing.T) {
	// GIVEN: a variety of Regexes to apply to a string.
	str := `testing\n"beta-release": "0.1.2-beta"\n"stable-release": "0.1.1"`
	tests := []struct {
		name  string
		regex string
		match bool
	}{
		{
			name:  "regex match",
			regex: `release": "[0-9.]+"`,
			match: true,
		},
		{
			name:  "no regex match",
			regex: `release": "[0-9.]+-alpha"`,
			match: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: RegexCheck is called.
			got := RegexCheck(tc.regex, str)

			prefix := fmt.Sprintf(
				"%s\nRegexCheck(re=%q, str=%q)",
				packageName, tc.regex, str,
			)

			// THEN: the regex matches when expected.
			if got != tc.match {
				t.Errorf("%s mismatch\ngot:  %t\nwant: %t", prefix, got, tc.match)
			}
		})
	}
}

func TestRegexCheckWithVersion(t *testing.T) {
	// GIVEN: a variety of Regexes to apply to a string.
	str := `testing\n"beta-release": "0.1.2-beta"\n"stable-release": "0.1.1"`
	tests := []struct {
		name    string
		regex   string
		version string
		match   bool
	}{
		{
			name:    "regex match",
			regex:   `release": "{{ version }}"`,
			version: "0.1.1",
			match:   true,
		},
		{
			name:    "no regex match",
			regex:   `release": "{{ version }}"`,
			version: "0.1.2",
			match:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: RegexCheckWithVersion is called.
			got := RegexCheckWithVersion(tc.regex, str, tc.version)

			prefix := fmt.Sprintf(
				"%s\nRegexCheckWithVersion(re=%q, str=%q, version=%q)",
				packageName, tc.regex, str, tc.version,
			)

			// THEN: the regex matches when expected.
			if got != tc.match {
				t.Errorf(
					"%s mismatch\ngot:  %t\nwant: %t",
					prefix, got, tc.match,
				)
			}
		})
	}
}

func TestRegexTemplate(t *testing.T) {
	// GIVEN: a RegEx, Index (and possibly a template) and text to run it on.
	tests := []struct {
		name            string
		text            string
		regex, template string
		want            string
	}{
		{
			name:     "datetime template",
			text:     "2024-01-01T16-36-33Z",
			regex:    `([\d-]+)T(\d+)-(\d+)-(\d+)Z`,
			template: "$1T$2:$3:$4Z",
			want:     "2024-01-01T16:36:33Z",
		},
		{
			name:     "template with 10+ matches",
			text:     "abcdefghijklmnopqrstuvwxyz",
			regex:    `([a-z])([a-z])([a-z])([a-z])([a-z]{2})([a-z])([a-z])([a-z])([a-z])([a-z])([a-z])`,
			template: "$1_$2_$3_$4_$5_$6_$7_$8_$9_$10_$11",
			want:     "a_b_c_d_ef_g_h_i_j_k_l",
		},
		{
			name:     "template with placeholder out of range",
			text:     "abc123-def456-ghi789",
			regex:    `([a-z]+)(\d+)`,
			template: "$1$4-$10",
			want:     "abc$4-abc0",
		},
		{
			name:     "template with all placeholders out of range",
			text:     "abc123-def456-ghi789",
			regex:    `([a-z]+)(\d+)`,
			template: "$4$5",
			want:     "$4$5",
		},
		{
			name:  "no template",
			text:  "abc123-def456-ghi789",
			regex: `([a-z]+)(\d+)`,
			want:  "123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			re := regexp.MustCompile(tc.regex)
			texts := re.FindAllStringSubmatch(tc.text, 1)
			regexMatches := texts[0]

			// WHEN: RegexTemplate is called on the regex matches.
			got := RegexTemplate(regexMatches, tc.template)

			prefix := fmt.Sprintf(
				"%s\nRegexTemplate(matches=%v, template=%q)",
				packageName, regexMatches, tc.template,
			)

			// THEN: the expected string is returned.
			if got != tc.want {
				t.Fatalf(
					"%s result mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}
