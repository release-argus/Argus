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

package web

import (
	"net/http"
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		yamlStr  string
		errRegex string
	}{
		"valid service": {
			yamlStr: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				regex: '[0-9.]+'
			`),
			errRegex: `^$`,
		},
		"method - empty string": {
			yamlStr: test.TrimYAML(`
				method: ''
				url: "https://example.com"
			`),
			errRegex: `^$`,
		},
		"method - invalid": {
			yamlStr: test.TrimYAML(`
				method: 'FOO'
				url: "https://example.com"
			`),
			errRegex: `method: "[^"]+" <invalid>`,
		},
		"method - valid": {
			yamlStr: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
			`),
			errRegex: `^$`,
		},
		"method - case insensitive": {
			yamlStr: test.TrimYAML(`
				method: 'gEt'
				url: "https://example.com"
			`),
			errRegex: `^$`,
		},
		"url - empty string": {
			yamlStr: test.TrimYAML(`
				method: ''
				url: ''
			`),
			errRegex: `url: <required>`,
		},
		"body - removed for GET": {
			yamlStr: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				body: "foo"
			`),
			errRegex: `^$`,
		},
		"body - not removed for POST": {
			yamlStr: test.TrimYAML(`
				method: ` + http.MethodPost + `
				url: "https://example.com"
				body: "foo"
			`),
			errRegex: `^$`,
		},
		"JSON - invalid, string in square brackets": {
			yamlStr: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				json: 'foo[bar]'
			`),
			errRegex: `json: .* <invalid>`,
		},
		"regex - invalid": {
			yamlStr: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				regex: '[0-'
			`),
			errRegex: `regex: .* <invalid>`,
		},
		"regexTemplate - with no regex": {
			yamlStr: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				regex_template: "$1.$2.$3"
			`),
			errRegex: `^$`,
		},
		"all errs": {
			yamlStr: test.TrimYAML(`
				type: url
				method: asd
				url: ""
				json: 'foo[bar]'
				regex: '[0-'
				regex_template: $1.$2.$3
			`),
			errRegex: test.TrimYAML(`
				method: "[^"]+" <invalid>.*
				url: <required>.*
				json: "[^"]+" <invalid>.*
				regex: "[^"]+" <invalid>.*`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			// Apply the YAML.
			if err := yaml.Unmarshal([]byte(tc.yamlStr), lookup); err != nil {
				t.Fatalf("%s\nerror unmarshalling YAML: %v",
					packageName, err)
			}
			hadBody := lookup.Body

			// WHEN CheckValues is called.
			err := lookup.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
					packageName, wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
			// AND RegexTemplate is empty when Regex is empty.
			if lookup.RegexTemplate != "" && lookup.Regex == "" {
				t.Fatalf("%s\nRegexTemplate should be nil when Regex is empty",
					packageName)
			}
			// AND Body is empty when Method is GET.
			if lookup.Method == http.MethodGet && lookup.Body != "" {
				t.Fatalf("%s\nBody should be nil when Method is GET",
					packageName)
			}
			// AND Body is kept when Method is POST.
			if lookup.Method == http.MethodPost && hadBody != "" && lookup.Body == "" {
				t.Fatalf("%s\nBody should be kept when Method is POST",
					packageName)
			}
			// AND Method is uppercased.
			wantMethod := strings.ToUpper(lookup.Method)
			if wantMethod == "" {
				wantMethod = http.MethodGet
			}
			if lookup.Method != wantMethod {
				t.Fatalf("%s\nMethod should be uppercased:\nwant: %q\ngot:  %q",
					packageName, wantMethod, lookup.Method)
			}
		})
	}
}
