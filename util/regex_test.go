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
	"regexp"
	"testing"
)

func TestRegexCheck(t *testing.T) {
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

			got := RegexCheck(tc.regex, str)

			if got != tc.match {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.match)
			}
		})
	}
}

func TestRegexCheckWithVersion(t *testing.T) {
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

			got := RegexCheckWithVersion(tc.regex, str, tc.version)

			if got != tc.match {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.match)
			}
		})
	}
}

func TestRegexTemplate(t *testing.T) {
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
			want:     "abc$4-abc0", // $10 is not a group, so its $1 is replaced.
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
			want:  "123", // No template gives the last group.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			regexMatches := regexp.MustCompile(tc.regex).FindStringSubmatch(tc.text)

			got := RegexTemplate(regexMatches, tc.template)

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
