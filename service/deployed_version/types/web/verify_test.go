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

package web

import (
	"net/http"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name              string
		data              string
		emptyHardDefaults bool
		errRegex          string
	}{
		{
			name: "valid service",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				regex: '[0-9.]+'
			`),
			errRegex: `^$`,
		},
		{
			name: "method/empty string",
			data: test.TrimYAML(`
				method: ''
				url: "https://example.com"
			`),
			emptyHardDefaults: true,
			errRegex:          `method: <required>.*` + http.MethodGet,
		},
		{
			name: "method/invalid",
			data: test.TrimYAML(`
				method: 'FOO'
				url: "https://example.com"
			`),
			errRegex: `method: "FOO" <invalid>.*` + http.MethodGet,
		},
		{
			name: "method/valid",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
			`),
			errRegex: `^$`,
		},
		{
			name: "method/case insensitive",
			data: test.TrimYAML(`
				method: 'gEt'
				url: "https://example.com"
			`),
			errRegex: `^$`,
		},
		{
			name: "url/empty string",
			data: test.TrimYAML(`
				method: ''
				url: ''
			`),
			errRegex: `url: <required>`,
		},
		{
			name: "body/removed for GET",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				body: "foo"
			`),
			errRegex: `^$`,
		},
		{
			name: "body/not removed for POST",
			data: test.TrimYAML(`
				method: ` + http.MethodPost + `
				url: "https://example.com"
				body: "foo"
			`),
			errRegex: `^$`,
		},
		{
			name: "JSON/invalid, string in square brackets",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				json: 'foo[bar]'
			`),
			errRegex: `^json: .* <invalid>.*$`,
		},
		{
			name: "regex/invalid",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				regex: '[0-'
			`),
			errRegex: `^regex: .* <invalid>.*$`,
		},
		{
			name: "regex_template, with no regex",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				regex_template: "$1.$2.$3"
			`),
			errRegex: `^$`,
		},
		{
			name: "all decode",
			data: test.TrimYAML(`
				type: url
				method: asd
				url: ""
				json: 'foo[bar]'
				regex: '[0-'
				regex_template: $1.$2.$3
			`),
			errRegex: test.TrimYAML(`
				url: <required>.*
				method: "[^"]+" <invalid>.*
				json: "[^"]+" <invalid>.*
				regex: "[^"]+" <invalid>.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testLookup(t, false)
			if tc.emptyHardDefaults {
				input.HardDefaults = &base.Defaults{}
			}
			// Apply the YAML.
			if err := input.UnmarshalYAML([]byte(tc.data)); err != nil {
				t.Fatalf(
					"%s\nLookup.UnmarshalYAML(%q) failed before Lookup.CheckValues(): %v",
					packageName, tc.data,
					err,
				)
			}
			hadBody := input.Body

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				input.CheckValues,
			)

			// AND: RegexTemplate is empty when Regex is empty.
			if input.RegexTemplate != "" && input.Regex == "" {
				t.Fatalf(
					"%s\nLookup.CheckValues() .RegexTemplate should be empty when Regex is empty",
					packageName,
				)
			}

			// AND: Body is empty when Method is GET.
			if input.Method == http.MethodGet && input.Body != "" {
				t.Fatalf(
					"%s\nLookup.CheckValues() .Body should be empty when Method is GET",
					packageName,
				)
			}

			// AND: Body is kept when Method is POST.
			if input.Method == http.MethodPost && hadBody != "" && input.Body == "" {
				t.Fatalf(
					"%s\nLookup.CheckValues() .Body should be kept when Method is POST",
					packageName,
				)
			}

			// AND: Method is uppercased.
			wantMethod := strings.ToUpper(input.Method)
			if wantMethod == "" {
				wantMethod = http.MethodGet
			}
			if input.Method != "" && input.Method != wantMethod {
				t.Fatalf(
					"%s\nLookup.CheckValues() .Method should be uppercased:\ngot:  %q\nwant: %q",
					packageName, input.Method, wantMethod,
				)
			}
		})
	}
}
