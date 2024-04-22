// Copyright [2023] [Argus]
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
	"regexp"
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestRegexCheck(t *testing.T) {
	// GIVEN a variety of RegEx's to apply to a string
	str := `testing\n"beta-release": "0.1.2-beta"\n"stable-release": "0.1.1"`
	tests := map[string]struct {
		regex string
		match bool
	}{
		"regex match":    {regex: `release": "[0-9.]+"`, match: true},
		"no regex match": {regex: `release": "[0-9.]+-alpha"`, match: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN RegexCheck is called
			got := RegexCheck(tc.regex, str)

			// THEN the regex matches when expected
			if got != tc.match {
				t.Errorf("wanted match=%t, not %t\n%q on %q",
					tc.match, got, tc.regex, str)
			}
		})
	}
}

func TestRegexCheckWithParams(t *testing.T) {
	// GIVEN a variety of RegEx's to apply to a string
	str := `testing\n"beta-release": "0.1.2-beta"\n"stable-release": "0.1.1"`
	tests := map[string]struct {
		regex   string
		version string
		match   bool
	}{
		"regex match": {
			regex:   `release": "{{ version }}"`,
			version: "0.1.1",
			match:   true},
		"no regex match": {
			regex:   `release": "{{ version }}"`,
			version: "0.1.2",
			match:   false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN RegexCheck is called
			got := RegexCheckWithParams(tc.regex, str, tc.version)

			// THEN the regex matches when expected
			if got != tc.match {
				t.Errorf("wanted match=%t, not %t\n%q on %q",
					tc.match, got, tc.regex, str)
			}
		})
	}
}

func TestRegexTemplate(t *testing.T) {
	// GIVEN a RegEx, Index (and possibly a template) and text to run it on
	tests := map[string]struct {
		text     string
		regex    string
		template *string
		want     string
	}{
		"datetime template": {
			text:     "2024-01-01T16-36-33Z",
			regex:    `([\d-]+)T(\d+)-(\d+)-(\d+)Z`,
			template: test.StringPtr("$1T$2:$3:$4Z"),
			want:     "2024-01-01T16:36:33Z",
		},
		"template with 10+ matches": {
			text:     "abcdefghijklmnopqrstuvwxyz",
			regex:    `([a-z])([a-z])([a-z])([a-z])([a-z]{2})([a-z])([a-z])([a-z])([a-z])([a-z])([a-z])`,
			template: test.StringPtr("$1_$2_$3_$4_$5_$6_$7_$8_$9_$10_$11"),
			want:     "a_b_c_d_ef_g_h_i_j_k_l",
		},
		"template with placeholder out of range": {
			text:     "abc123-def456-ghi789",
			regex:    `([a-z]+)(\d+)`,
			template: test.StringPtr("$1$4-$10"),
			want:     "abc$4-abc0",
		},
		"template with all placeholders out of range": {
			text:     "abc123-def456-ghi789",
			regex:    `([a-z]+)(\d+)`,
			template: test.StringPtr("$4$5"),
			want:     "$4$5",
		},
		"no template": {
			text:  "abc123-def456-ghi789",
			regex: `([a-z]+)(\d+)`,
			want:  "123",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			re := regexp.MustCompile(tc.regex)
			texts := re.FindAllStringSubmatch(tc.text, 1)
			regexMatches := texts[0]

			// WHEN RegexTemplate is called on the regex matches
			got := RegexTemplate(regexMatches, tc.template)

			// THEN the expected string is returned
			if got != tc.want {
				t.Fatalf("want: %q\n got: %q",
					tc.want, got)
			}
		})
	}
}
