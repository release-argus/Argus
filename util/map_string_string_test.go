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

package util

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/release-argus/Argus/test"
	"gopkg.in/yaml.v3"
)

func TestMapStringStringOmitNull_UnmarshalYAML(t *testing.T) {
	// GIVEN a YAML string to unmarshal into a MapStringStringOmitNull.
	tests := map[string]struct {
		input    string
		expected MapStringStringOmitNull
		errRegex string
	}{
		"empty": {
			input:    "",
			expected: MapStringStringOmitNull{},
			errRegex: `^$`,
		},
		"single kv": {
			input: test.TrimYAML(`
				foo: bar`),
			expected: MapStringStringOmitNull{
				"foo": "bar"},
			errRegex: `^$`,
		},
		"multiple with null and empty string omitted": {
			input: test.TrimYAML(`
				a: 1
				b: null
				c: ''`),
			expected: MapStringStringOmitNull{
				"a": "1"},
			errRegex: `^$`,
		},
		"invalid YAML": {
			input: test.TrimYAML(`
				foo`),
			errRegex: `cannot unmarshal .* into map\[string\]\*string`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the YAML is unmarshalled.
			var got MapStringStringOmitNull
			err := yaml.Unmarshal([]byte(tc.input), &got)

			// THEN we get an error if expected.
			if tc.errRegex != "" || err != nil {
				e := ErrorToString(err)
				if !RegexCheck(tc.errRegex, e) {
					t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.errRegex, e)
				}
				if tc.errRegex != "" {
					return
				}
			}

			// AND the map matches what we expect.
			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("%s\nunmarshal YAML mismatch\nwant: %+v\ngot:  %+v",
					packageName, tc.expected, got)
			}
		})
	}
}

func TestMapStringStringOmitNull_UnmarshalJSON(t *testing.T) {
	// GIVEN a JSON string to unmarshal into a MapStringStringOmitNull.
	tests := map[string]struct {
		input    string
		expected MapStringStringOmitNull
		errRegex string
	}{
		"empty object": {
			input:    `{}`,
			expected: MapStringStringOmitNull{},
			errRegex: `^$`,
		},
		"single kv": {
			input:    test.TrimJSON(`{
				"foo":"bar"
			}`),
			expected: MapStringStringOmitNull{"foo": "bar"},
			errRegex: `^$`,
		},
		"null value omitted": {
			input:    test.TrimJSON(`{
				"a":null
			}`),
			expected: MapStringStringOmitNull{},
			errRegex: `^$`,
		},
		"empty string value kept": {
			input:    test.TrimJSON(`{
				"a":""
			}`),
			expected: MapStringStringOmitNull{"a": ""},
			errRegex: `^$`,
		},
		"invalid JSON (array)": {
			input:    test.TrimJSON(`
				[]`),
			errRegex: `json: cannot unmarshal .* into Go value of type map\[string\]\*string`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the JSON is unmarshalled.
			var got MapStringStringOmitNull
			err := json.Unmarshal([]byte(tc.input), &got)

			// THEN we get an error if expected.
			if tc.errRegex != "" || err != nil {
				e := ErrorToString(err)
				if !RegexCheck(tc.errRegex, e) {
					t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.errRegex, e)
				}
				if tc.errRegex != "" {
					return
				}
			}

			// AND the map matches what we expect.
			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("%s\nunmarshal JSON mismatch\nwant: %+v\ngot:  %+v",
					packageName, tc.expected, got)
			}
		})
	}
}
