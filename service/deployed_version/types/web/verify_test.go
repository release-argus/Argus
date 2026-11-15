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
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
)

func TestLookup_CheckValues(t *testing.T) {
	tests := []struct {
		name              string
		data              string
		emptyHardDefaults bool
		errRegex          string
		wantMethod        string
		wantBody          string
		wantRegexTemplate string
	}{
		{
			name: "valid service",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				regex: '[0-9.]+'
			`),
			errRegex:   `^$`,
			wantMethod: http.MethodGet,
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
			errRegex:   `method: "FOO" <invalid>.*` + http.MethodGet,
			wantMethod: "FOO",
		},
		{
			name: "method/valid",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
			`),
			errRegex:   `^$`,
			wantMethod: http.MethodGet,
		},
		{
			name: "method/case insensitive",
			data: test.TrimYAML(`
				method: 'gEt'
				url: "https://example.com"
			`),
			errRegex:   `^$`,
			wantMethod: http.MethodGet,
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
			errRegex:   `^$`,
			wantMethod: http.MethodGet,
		},
		{
			name: "body/not removed for POST",
			data: test.TrimYAML(`
				method: ` + http.MethodPost + `
				url: "https://example.com"
				body: "foo"
			`),
			errRegex:   `^$`,
			wantMethod: http.MethodPost,
			wantBody:   "foo",
		},
		{
			name: "JSON/invalid, string in square brackets",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				json: 'foo[bar]'
			`),
			errRegex:   `^json: .* <invalid>.*$`,
			wantMethod: http.MethodGet,
		},
		{
			name: "regex/invalid",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				regex: '[0-'
			`),
			errRegex:   `^regex: .* <invalid>.*$`,
			wantMethod: http.MethodGet,
		},
		{
			name: "regex_template, with no regex",
			data: test.TrimYAML(`
				method: ` + http.MethodGet + `
				url: "https://example.com"
				regex_template: "$1.$2.$3"
			`),
			errRegex:          `^$`,
			wantMethod:        http.MethodGet,
			wantRegexTemplate: "",
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
			wantMethod:        "ASD",
			wantRegexTemplate: "$1.$2.$3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := testLookup(t)
			if tc.emptyHardDefaults {
				input.HardDefaults = &base.Defaults{}
			}
			if err := input.UnmarshalYAML([]byte(tc.data)); err != nil {
				t.Fatalf("Lookup.UnmarshalYAML() failed: %v", err)
			}

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				input.CheckValues,
			)

			if input.Method != tc.wantMethod {
				t.Errorf("Lookup.Method mismatch\ngot:  %q\nwant: %q", input.Method, tc.wantMethod)
			}
			if input.Body != tc.wantBody {
				t.Errorf("Lookup.Body mismatch\ngot:  %q\nwant: %q", input.Body, tc.wantBody)
			}
			if input.RegexTemplate != tc.wantRegexTemplate {
				t.Errorf(
					"Lookup.RegexTemplate mismatch\ngot:  %q\nwant: %q",
					input.RegexTemplate, tc.wantRegexTemplate,
				)
			}
		})
	}
}
