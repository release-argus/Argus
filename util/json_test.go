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
	"regexp"
	"strings"
	"testing"
)

func TestParseKeys(t *testing.T) {
	// GIVEN a JSON key string.
	tests := map[string]struct {
		input    string
		want     []any
		errRegex string
	}{
		"empty string": {
			input:    "",
			want:     []any{},
			errRegex: `^$`,
		},
		"single key": {
			input:    "foo",
			want:     []any{"foo"},
			errRegex: `^$`,
		},
		"multiple keys": {
			input:    "foo.bar",
			want:     []any{"foo", "bar"},
			errRegex: `^$`,
		},
		"multiple keys with array": {
			input:    "foo.bar[1]",
			want:     []any{"foo", "bar", 1},
			errRegex: `^$`,
		},
		"multiple keys with array of objects": {
			input:    "foo.bar[1].baz",
			want:     []any{"foo", "bar", 1, "baz"},
			errRegex: `^$`,
		},
		"multiple keys with array of arrays": {
			input:    "foo.bar[1][2]",
			want:     []any{"foo", "bar", 1, 2},
			errRegex: `^$`,
		},
		"multiple keys with array of arrays of objects": {
			input:    "foo.bar[1][2].baz",
			want:     []any{"foo", "bar", 1, 2, "baz"},
			errRegex: `^$`,
		},
		"multiple keys with array of arrays of objects with array": {
			input:    "foo.bar[1][2].baz[3]",
			want:     []any{"foo", "bar", 1, 2, "baz", 3},
			errRegex: `^$`,
		},
		"non-int index": {
			input:    "foo.bar[1.1][2].baz[3]",
			want:     nil,
			errRegex: `failed to parse index "1.1" in `,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ParseKeys is called.
			got, err := ParseKeys(tc.input)

			// THEN the keys are returned correctly.
			if len(got) != len(tc.want) {
				t.Fatalf("%s\ndifferent amount of keys returned\nwant: %v\ngot:  %v",
					packageName, tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("%s - wrong key [%d]\nwant: %v\ngot:  %v",
						packageName, i, tc.want, got)
				}
			}
			// AND the error is returned correctly.
			e := ErrorToString(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q\n",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestNavigateJSON(t *testing.T) {
	// GIVEN a JSON string.
	tests := map[string]struct {
		input    string
		key      string
		want     string
		errRegex string
	}{
		"empty key": {
			input:    `{ "foo": "bar" }`,
			key:      "",
			errRegex: `no key was given`,
		},
		"object not found": {
			input:    "{}",
			key:      "foo",
			errRegex: `failed to find value for "[^"]+" in `,
		},
		"simple JSON": {
			input:    `{"foo": "bar"}`,
			key:      "foo",
			want:     "bar",
			errRegex: `^$`,
		},
		"multi-level JSON": {
			input:    `{"foo": {"bar": "baz"}}`,
			key:      "foo.bar",
			want:     "baz",
			errRegex: `^$`,
		},
		"multi-level JSON with array": {
			input:    `{"foo": {"bar": ["baz", "bish"]}}`,
			key:      "foo.bar[1]",
			want:     "bish",
			errRegex: `^$`,
		},
		"multi-level JSON with array of objects": {
			input:    `{"foo": {"bar": [{"baz": "bish"}, {"bash": "uniform"}]}}`,
			key:      "foo.bar[1].bash",
			want:     "uniform",
			errRegex: `^$`,
		},
		"multi-level JSON with array of arrays": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "uniform"]]}}`,
			key:      "foo.bar[1][1]",
			want:     "uniform",
			errRegex: `^$`,
		},
		"negative index": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "uniform"]]}}`,
			key:      "foo.bar[-1][1]",
			want:     "uniform",
			errRegex: `^$`,
		},
		"fail: index of map": {
			input:    `{"foo": {"bar": {"baz": "bish"}}}`,
			key:      "foo.bar[1]",
			errRegex: `got a map, but the key is not a string`,
		},
		"fail: non-int index": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "uniform"]]}}`,
			key:      "foo.bar.bar",
			errRegex: `got an array, but the key is not an integer index`,
		},
		"fail: index out of range": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "uniform"]]}}`,
			key:      "foo.bar[1][2]",
			errRegex: `index \d \([^)]+\) out of range`,
		},
		"fail: index out of range (negative)": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "uniform"]]}}`,
			key:      "foo.bar[-4][3]",
			errRegex: `index -\d \([^)]+\) out of range`,
		},
		"fail: got value instead of object": {
			input:    `{"foo": {"bar": "baz"}}`,
			key:      "foo.bar.baz",
			errRegex: `got a value of "[^"]+" at "[^"]+", but there are more keys to navigate`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var jsonData any
			err := json.Unmarshal([]byte(tc.input), &jsonData)

			// WHEN navigateJSON is called.
			got, err := navigateJSON(&jsonData, tc.key)

			// THEN the value is returned correctly.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
			// AND the error is returned correctly.
			e := ErrorToString(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q\n",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestGetValueByKey(t *testing.T) {
	// GIVEN a JSON string.
	tests := map[string]struct {
		input    string
		key      string
		want     string
		errRegex string
	}{
		"fail unmarshal": {
			input:    "{",
			key:      "foo",
			errRegex: `failed to unmarshal the following from`,
		},
		"empty key": {
			input:    `{ "foo": "bar" }`,
			key:      "",
			want:     "__root",
			errRegex: `^$`,
		},
		"object not found": {
			input:    "{}",
			key:      "foo",
			errRegex: `failed to find value for "[^"]+" in `,
		},
		"simple JSON": {
			input:    `{"foo": "bar"}`,
			key:      "foo",
			want:     "bar",
			errRegex: `^$`,
		},
		"multi-level JSON": {
			input:    `{"foo": {"bar": "baz"}}`,
			key:      "foo.bar",
			want:     "baz",
			errRegex: `^$`,
		},
		"negative index": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "uniform"]]}}`,
			key:      "foo.bar[-1][1]",
			want:     "uniform",
			errRegex: `^$`,
		},
		"fail: index out of range": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "uniform"]]}}`,
			key:      "foo.bar[1][2]",
			errRegex: `index \d \([^)]+\) out of range`,
		},
		"fail: index out of range (negative)": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "uniform"]]}}`,
			key:      "foo.bar[-4][3]",
			errRegex: `index -\d \([^)]+\) out of range`,
		},
		"fail: got value instead of object": {
			input:    `{"foo": {"bar": "baz"}}`,
			key:      "foo.bar.baz",
			errRegex: `got a value of "[^"]+" at "[^"]+", but there are more keys to navigate`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN GetValueByKey is called.
			got, err := GetValueByKey([]byte(tc.input), tc.key, "https://release-argus.com")

			// THEN the value is returned correctly.
			tc.want = strings.ReplaceAll(tc.want, "__root", tc.input)
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
			// AND the error is returned correctly.
			e := ErrorToString(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q\n",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestToJSONString(t *testing.T) {
	tests := map[string]struct {
		input any
		want  string
	}{
		"invalid input": {
			input: make(chan int),
			want:  "",
		},
		"nil input": {
			input: nil,
			want:  "null",
		},
		"empty string": {
			input: "",
			want:  `""`,
		},
		"simple string": {
			input: "test",
			want:  `"test"`,
		},
		"integer": {
			input: 123,
			want:  "123",
		},
		"float": {
			input: 123.45,
			want:  "123.45",
		},
		"boolean true": {
			input: true,
			want:  "true",
		},
		"boolean false": {
			input: false,
			want:  "false",
		},
		"simple map": {
			input: map[string]any{"foo": "bar"},
			want:  `{"foo":"bar"}`,
		},
		"nested map": {
			input: map[string]any{"foo": map[string]any{"bar": "baz"}},
			want:  `{"foo":{"bar":"baz"}}`,
		},
		"simple slice": {
			input: []any{"foo", "bar"},
			want:  `["foo","bar"]`,
		},
		"nested slice": {
			input: []any{[]any{"foo", "bar"}, "baz"},
			want:  `[["foo","bar"],"baz"]`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ToJSONString is called.
			got := ToJSONString(tc.input)

			// THEN the JSON string is returned correctly.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}
