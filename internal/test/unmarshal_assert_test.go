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

package test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/release-argus/Argus/util/errfmt"
)

func TestAssertUnmarshal(t *testing.T) {
	// GIVEN: a struct type.
	type testStruct struct {
		String string `json:"string,omitempty" yaml:"string,omitempty"`
		Int    int    `json:"int,omitempty" yaml:"int,omitempty"`
		Bool   bool   `json:"bool,omitempty" yaml:"bool,omitempty"`
	}

	// AND: data to unmarshal into it.
	tests := []struct {
		name                            string
		format, data                    string
		want                            string
		stringify                       func(*testStruct) string
		unmarshalErrRegex, testErrRegex string
	}{
		{
			name:              "unknown format",
			format:            "foo",
			data:              "",
			stringify:         nil,
			want:              "",
			unmarshalErrRegex: `^unsupported format: "foo"$`,
			testErrRegex:      `^$`,
		},
		{
			name:              "JSON/empty",
			format:            "json",
			data:              "",
			stringify:         genericStringify[*testStruct],
			want:              "{}\n",
			unmarshalErrRegex: `^$`,
			testErrRegex:      `^$`,
		},
		{
			name:              "JSON/empty object",
			format:            "json",
			data:              "{}",
			stringify:         genericStringify[*testStruct],
			want:              "{}\n",
			unmarshalErrRegex: `^$`,
			testErrRegex:      `^$`,
		},
		{
			name:              "YAML/empty",
			format:            "yaml",
			data:              "",
			stringify:         genericStringify[*testStruct],
			want:              "{}\n",
			unmarshalErrRegex: `^$`,
			testErrRegex:      `^$`,
		},
		{
			name:              "JSON/error caught",
			format:            "json",
			data:              `{invalid: json}`,
			stringify:         genericStringify[*testStruct],
			want:              "",
			unmarshalErrRegex: `invalid character`,
			testErrRegex:      `^$`,
		},
		{
			name:              "YAML/error caught",
			format:            "yaml",
			data:              `invalid: "yaml`,
			stringify:         genericStringify[*testStruct],
			want:              "",
			unmarshalErrRegex: `could not find end character`,
			testErrRegex:      `^$`,
		},
		{
			name:              "YAML/error mismatch",
			format:            "yaml",
			data:              `invalid: "yaml`,
			stringify:         genericStringify[*testStruct],
			want:              "",
			unmarshalErrRegex: `^$`,
			testErrRegex:      `could not find end character`,
		},
		{
			name:   "JSON/stringified match",
			format: "json",
			data: TrimJSON(`{
				"string": "foo",
				"int": 42,
				"bool": true
			}`),
			stringify: genericStringify[*testStruct],
			want: TrimYAML(`
				string: foo
				int: 42
				bool: true
			`),
			unmarshalErrRegex: `^$`,
			testErrRegex:      `^$`,
		},
		{
			name:   "YAML/stringified match",
			format: "yaml",
			data: TrimYAML(`
				string: foo
				int: 42
				bool: true
			`),
			stringify: genericStringify[*testStruct],
			want: TrimYAML(`
				string: foo
				int: 42
				bool: true
			`),
			unmarshalErrRegex: `^$`,
			testErrRegex:      `^$`,
		},
		{
			name:   "JSON/stringified mismatch caught",
			format: "json",
			data: TrimJSON(`{
				"string": "foo",
				"int": 42,
				"bool": true
			}`),
			stringify: func(ts *testStruct) string {
				return ""
			},
			want: TrimYAML(`
				string: foo
				int: 42
				bool: true
			`),
			unmarshalErrRegex: `stringified mismatch`,
			testErrRegex:      `stringified mismatch`,
		},
		{
			name:   "YAML/stringified mismatch caught",
			format: "yaml",
			data: TrimYAML(`
				string: foo
				int: 42
				bool: true
			`),
			stringify: func(ts *testStruct) string {
				return ""
			},
			want: TrimYAML(`
				string: foo
				int: 42
				bool: true
			`),
			unmarshalErrRegex: `stringified mismatch`,
			testErrRegex:      `stringified mismatch`,
		},
		{
			name:   "JSON/stringified mismatch uncaught",
			format: "json",
			data: TrimJSON(`{
				"string": "foo",
				"int": 42,
				"bool": true
			}`),
			stringify: func(ts *testStruct) string {
				return ""
			},
			want: TrimYAML(`
				string: foo
				int: 42
				bool: true
			`),
			unmarshalErrRegex: `^$`,
			testErrRegex:      `stringified mismatch`,
		},
		{
			name:   "YAML/stringified mismatch uncaught",
			format: "yaml",
			data: TrimYAML(`
				string: foo
				int: 42
				bool: true
			`),
			stringify: func(ts *testStruct) string {
				return ""
			},
			want: TrimYAML(`
				string: foo
				int: 42
				bool: true
			`),
			unmarshalErrRegex: `^$`,
			testErrRegex:      `stringified mismatch`,
		},
		{
			name:              "stringify function missing",
			format:            "yaml",
			data:              "",
			stringify:         nil,
			want:              "",
			unmarshalErrRegex: `^$`,
			testErrRegex:      `stringify function is required`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: an instance of that struct.
			var v testStruct

			// WHEN: it is unmarshaled.
			_, testErr := AssertUnmarshal(
				t,
				tc.format,
				tc.data,
				&v,
				tc.unmarshalErrRegex,
				tc.stringify,
				tc.want,
				packageName,
				"TestAssertUnmarshal",
			)

			prefix := fmt.Sprintf(
				"%s\nAssertUnmarshal(format=%q, data=%q)",
				packageName, tc.format, tc.data,
			)

			// THEN: the test errors when expected.
			e := errfmt.FormatError(testErr)
			if !regexp.MustCompile(tc.testErrRegex).MatchString(e) {
				t.Errorf(
					"%s test error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.testErrRegex,
				)
			}
		})
	}
}
