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

package file

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestLookup_CheckValues(t *testing.T) {
	tests := []struct {
		name              string
		data              string
		errRegex          string
		wantRegexTemplate string // Cleared if no regex.
	}{
		{
			name: "valid lookup",
			data: test.TrimYAML(`
				path: /VERSION
				regex: 'v([0-9.]+)'
			`),
			errRegex: `^$`,
		},
		{
			name: "path/empty string",
			data: test.TrimYAML(`
				path: ''
			`),
			errRegex: `^path: <required>`,
		},
		{
			name: "path/relative",
			data: test.TrimYAML(`
				path: relative/VERSION
			`),
			errRegex: `^path: "relative/VERSION" <invalid>.*must be an absolute path`,
		},
		{
			name: "regex/invalid",
			data: test.TrimYAML(`
				path: /VERSION
				regex: '[0-'
			`),
			errRegex: `^regex: .* <invalid>.*$`,
		},
		{
			name: "regex_template/with regex",
			data: test.TrimYAML(`
				path: /VERSION
				regex: 'v([0-9]+)-([0-9]+)'
				regex_template: "$1.$2"
			`),
			errRegex:          `^$`,
			wantRegexTemplate: "$1.$2",
		},
		{
			name: "regex_template/without regex",
			data: test.TrimYAML(`
				path: /VERSION
				regex_template: "$1.$2.$3"
			`),
			errRegex:          `^$`,
			wantRegexTemplate: "",
		},
		{
			name: "all decode",
			data: test.TrimYAML(`
				type: file
				path: relative
				regex: '[0-'
				regex_template: $1.$2.$3
			`),
			errRegex: test.TrimYAML(`
				path: "[^"]+" <invalid>.*
				regex: "[^"]+" <invalid>.*$`,
			),
			wantRegexTemplate: "$1.$2.$3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, "")
			input, err := Decode(
				"yaml", []byte(tc.data),
				lookup.Options,
				lookup.Status,
				plainDefaultsConfig(),
			)
			if err != nil {
				t.Fatalf("Decode() failed: %v", err)
			}

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				input.CheckValues,
			)

			if input.RegexTemplate != tc.wantRegexTemplate {
				t.Errorf(
					"Lookup.RegexTemplate mismatch\ngot:  %q\nwant: %q",
					input.RegexTemplate, tc.wantRegexTemplate,
				)
			}
		})
	}
}
