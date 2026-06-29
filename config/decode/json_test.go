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

package decode

import (
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestParseKeys(t *testing.T) {
	// GIVEN: a JSON key string.
	tests := []struct {
		name     string
		input    string
		want     []any
		errRegex string
	}{
		{
			name:     "empty string",
			input:    "",
			want:     []any{},
			errRegex: `^$`,
		},
		{
			name:     "single key",
			input:    "foo",
			want:     []any{"foo"},
			errRegex: `^$`,
		},
		{
			name:     "multiple keys",
			input:    "foo.bar",
			want:     []any{"foo", "bar"},
			errRegex: `^$`,
		},
		{
			name:     "multiple keys with array",
			input:    "foo.bar[1]",
			want:     []any{"foo", "bar", 1},
			errRegex: `^$`,
		},
		{
			name:     "multiple keys with array of objects",
			input:    "foo.bar[1].baz",
			want:     []any{"foo", "bar", 1, "baz"},
			errRegex: `^$`,
		},
		{
			name:     "multiple keys with array of arrays",
			input:    "foo.bar[1][2]",
			want:     []any{"foo", "bar", 1, 2},
			errRegex: `^$`,
		},
		{
			name:     "multiple keys with array of arrays of objects",
			input:    "foo.bar[1][2].baz",
			want:     []any{"foo", "bar", 1, 2, "baz"},
			errRegex: `^$`,
		},
		{
			name:     "multiple keys with array of arrays of objects with array",
			input:    "foo.bar[1][2].baz[3]",
			want:     []any{"foo", "bar", 1, 2, "baz", 3},
			errRegex: `^$`,
		},
		{
			name:     "non-int index",
			input:    "foo.bar[1.1][2].baz[3]",
			want:     nil,
			errRegex: `failed to parse index "1.1" in `,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: ParseKeys is called.
			got, err := ParseKeys(tc.input)

			prefix := fmt.Sprintf(
				"%s\nParseKeys(%q)",
				packageName, tc.input,
			)

			// THEN: the keys are returned correctly.
			if len(got) != len(tc.want) {
				t.Fatalf(
					"%s unexpected number of keys returned\ngot:  %v\nwant: %v",
					prefix, got, tc.want,
				)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf(
						"%s wrong key [%d]\ngot:  %v\nwant: %v",
						prefix, i,
						got, tc.want,
					)
				}
			}

			// AND: the error is returned correctly.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
		})
	}
}

func TestNavigateJSON(t *testing.T) {
	// GIVEN: a JSON string.
	tests := []struct {
		name     string
		input    string
		jsonData any
		key      string
		want     string
		errRegex string
	}{
		{
			name:     "empty key",
			input:    `{"foo": "bar"}`,
			key:      "",
			errRegex: `no key was given`,
		},
		{
			name:     "object not found",
			input:    "{}",
			key:      "foo",
			errRegex: `failed to find value for "[^"]+" in `,
		},
		{
			name:     "simple JSON",
			input:    `{"foo": "bar"}`,
			key:      "foo",
			want:     "bar",
			errRegex: `^$`,
		},
		{
			name: "multi-level JSON",
			input: test.TrimJSON(`{
				"foo": {
					"bar": "baz"
				}
			}`),
			key:      "foo.bar",
			want:     "baz",
			errRegex: `^$`,
		},
		{
			name: "multi-level JSON with array",
			input: test.TrimJSON(`{
				"foo": {
					"bar": [
						"baz",
						"bish"
					]
				}
			}`),
			key:      "foo.bar[1]",
			want:     "bish",
			errRegex: `^$`,
		},
		{
			name: "multi-level JSON with array of objects",
			input: test.TrimJSON(`{
				"foo": {
					"bar": [
						{
							"baz": "bish"
						},
						{
							"bash": "uniform"
						}
					]
				}
			}`),
			key:      "foo.bar[1].bash",
			want:     "uniform",
			errRegex: `^$`,
		},
		{
			name: "multi-level JSON with array of arrays",
			input: test.TrimJSON(`{
				"foo": {
					"bar": [
						["baz", "bish"],
						["bash", "uniform"]
					]
				}
			}`),
			key:      "foo.bar[1][1]",
			want:     "uniform",
			errRegex: `^$`,
		},
		{
			name: "negative index",
			input: test.TrimJSON(`{
				"foo": {
					"bar": [
						["baz", "bish"],
						["bash", "uniform"]
					]
				}
			}`),
			key:      "foo.bar[-1][1]",
			want:     "uniform",
			errRegex: `^$`,
		},
		{
			name: "fail: index of map",
			input: test.TrimJSON(`{
				"foo": {
					"bar": {
						"baz": "bish"
					}
				}
			}`),
			key:      "foo.bar[1]",
			errRegex: `got a map, but the wanted key is not a string`,
		},
		{
			name: "fail: non-int index",
			input: test.TrimJSON(`{
				"foo": {
					"bar": [
						["baz", "bish"],
						["bash", "uniform"]
					]
				}
			}`),
			key:      "foo.bar.bar",
			errRegex: `got an array, but the key is not an integer index`,
		},
		{
			name: "fail: index out of range",
			input: test.TrimJSON(`{
				"foo": {
					"bar": [
						["baz", "bish"],
						["bash", "uniform"]
					]
				}
			}`),
			key:      "foo.bar[1][2]",
			errRegex: `index \d \([^)]+\) out of range`,
		},
		{
			name: "fail: index out of range (negative)",
			input: test.TrimJSON(`{
				"foo": {
					"bar": [
						["baz", "bish"],
						["bash", "uniform"]
					]
				}
			}`),
			key:      "foo.bar[-4][3]",
			errRegex: `index -\d \([^)]+\) out of range`,
		},
		{
			name: "fail: null array element mid-path",
			input: test.TrimJSON(`{
				"foo": [
					null,
					{"bar": "baz"}
				]
			}`),
			key:      "foo[0].bar",
			errRegex: `got null at index 0 while navigating "foo\[0\]\.bar"`,
		},
		{
			name: "fail: got value instead of object",
			input: test.TrimJSON(`{
				"foo": {
					"bar": "baz"
				}
			}`),
			key:      "foo.bar.baz",
			errRegex: `failed to find key "[^"]+" while navigating "[^"]+": [^ ]+ is not an object or array`,
		},
		{
			name:     "boolean leaf, true",
			input:    `{"enabled": true}`,
			key:      "enabled",
			want:     "true",
			errRegex: `^$`,
		},
		{
			name:     "boolean leaf, false",
			input:    `{"enabled": false}`,
			key:      "enabled",
			want:     "false",
			errRegex: `^$`,
		},
		{
			name:     "fail: boolean mid-path",
			input:    `{"foo": true}`,
			key:      "foo.bar",
			errRegex: `failed to find key "bar" while navigating "foo.bar": true is not an object or array`,
		},
		{
			name:     "null leaf",
			input:    `{"version": null}`,
			key:      "version",
			want:     "",
			errRegex: `^$`,
		},
		{
			name: "fail: leaf is object",
			input: test.TrimJSON(`{
				"foo": {
					"bar": "baz"
				}
			}`),
			key:      "foo",
			errRegex: `^failed to find value for "foo" in `,
		},
		{
			name:     "fail: leaf is array",
			input:    `{"foo": ["a", "b"]}`,
			key:      "foo",
			errRegex: `^failed to find value for "foo" in `,
		},
		{
			name:     "fail: case nil at root",
			jsonData: nil,
			key:      "foo",
			errRegex: `^got null at "foo" while navigating "foo"$`,
		},
		{
			name:     "fail: null mid-path",
			input:    `{"foo": null}`,
			key:      "foo.bar",
			errRegex: `got null at "foo" while navigating "foo.bar"`,
		},
		{
			name:     "fail: case nil at root with remaining path",
			jsonData: nil,
			key:      "foo.bar",
			errRegex: `^got null at "foo" while navigating "foo.bar"$`,
		},
		{
			name: "fail: case nil from JSON null document",
			jsonData: func() any {
				var data any
				_ = Unmarshal("json", []byte(`null`), &data)
				return data
			}(),
			key:      "version",
			errRegex: `^got null at "version" while navigating "version"$`,
		},
		{
			name: "fail: default unsupported type mid-path",
			jsonData: map[string]any{
				"foo": struct{}{},
			},
			key:      "foo.bar",
			errRegex: `got unsupported type struct \{\} at "bar" while navigating "foo.bar"`,
		},
		{
			name:     "fail: default unsupported type at root",
			jsonData: struct{}{},
			key:      "foo",
			errRegex: `got unsupported type struct \{\} at "foo" while navigating "foo"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.input != "" {
				_ = Unmarshal("json", []byte(tc.input), &tc.jsonData)
			}

			// WHEN: navigateJSON is called.
			got, err := navigateJSON(&tc.jsonData, tc.key)

			prefix := fmt.Sprintf(
				"%s\nnavigateJSON(data=%v, key=%q)",
				packageName, tc.jsonData, tc.key,
			)

			// THEN: the value is returned correctly.
			if got != tc.want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the error is returned correctly.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
		})
	}
}

func TestGetValueByKey(t *testing.T) {
	// GIVEN: a JSON string.
	tests := []struct {
		name     string
		input    string
		key      string
		want     string
		errRegex string
	}{
		{
			name:  "fail unmarshal",
			input: "{",
			key:   "foo",
			errRegex: test.TrimYAML(`
				^failed to unmarshal response from "[^"]+" into JSON:
					jsontext: unexpected EOF`,
			),
		},
		{
			name:     "empty key",
			input:    `{"foo": "bar"}`,
			key:      "",
			want:     "__root",
			errRegex: `^$`,
		},
		{
			name:     "object not found",
			input:    "{}",
			key:      "foo",
			errRegex: `failed to find value for "[^"]+" in `,
		},
		{
			name:     "simple JSON",
			input:    `{"foo": "bar"}`,
			key:      "foo",
			want:     "bar",
			errRegex: `^$`,
		},
		{
			name: "multi-level JSON",
			input: test.TrimJSON(`{
				"foo": {
					"bar": "baz"
				}
			}`),
			key:      "foo.bar",
			want:     "baz",
			errRegex: `^$`,
		},
		{
			name: "negative index",
			input: test.TrimJSON(`{
				"foo": {
					"bar": [
						["baz", "bish"],
						["bash", "uniform"]
					]
				}
			}`),
			key:      "foo.bar[-1][1]",
			want:     "uniform",
			errRegex: `^$`,
		},
		{
			name: "fail: index out of range",
			input: test.TrimJSON(`{
				"foo": {
					"bar": [
						["baz", "bish"],
						["bash", "uniform"]
					]
				}
			}`),
			key:      "foo.bar[1][2]",
			errRegex: `index \d \([^)]+\) out of range`,
		},
		{
			name: "fail: index out of range (negative)",
			input: test.TrimJSON(`{
				"foo": {
					"bar": [
						["baz", "bish"],
						["bash", "uniform"]
					]
				}
			}`),
			key:      "foo.bar[-4][3]",
			errRegex: `index -\d \([^)]+\) out of range`,
		},
		{
			name: "fail: got value instead of object",
			input: test.TrimJSON(`{
				"foo": {
					"bar": "baz"
				}
			}`),
			key:      "foo.bar.baz",
			errRegex: `failed to find key "baz" while navigating "[^"]+": [^ ]+ is not an object or array`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetValueByKey is called.
			got, err := GetValueByKey([]byte(tc.input), tc.key, "https://release-argus.com")

			prefix := fmt.Sprintf(
				"%s\nGetValueByKey(data=%q, key=%q)",
				packageName, tc.input, tc.key,
			)

			// THEN: the value is returned correctly.
			tc.want = strings.ReplaceAll(tc.want, "__root", tc.input)
			if got != tc.want {
				t.Errorf(
					"%s value mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the error is returned correctly.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
		})
	}
}

func TestToJSONString(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "invalid data",
			input: make(chan int),
			want:  "",
		},
		{
			name:  "nil data",
			input: nil,
			want:  "null",
		},
		{
			name:  "empty string",
			input: "",
			want:  `""`,
		},
		{
			name:  "simple string",
			input: "test",
			want:  `"test"`,
		},
		{
			name:  "integer",
			input: 123,
			want:  "123",
		},
		{
			name:  "float",
			input: 123.45,
			want:  "123.45",
		},
		{
			name:  "boolean true",
			input: true,
			want:  "true",
		},
		{
			name:  "boolean false",
			input: false,
			want:  "false",
		},
		{
			name: "simple map",
			input: map[string]any{
				"foo": "bar",
			},
			want: `{"foo":"bar"}`,
		},
		{
			name: "nested map",
			input: map[string]any{
				"foo": map[string]any{
					"bar": "baz",
				},
			},
			want: `{"foo":{"bar":"baz"}}`,
		},
		{
			name:  "simple slice",
			input: []any{"foo", "bar"},
			want:  `["foo","bar"]`,
		},
		{
			name: "nested slice",
			input: []any{
				[]any{"foo", "bar"},
				"baz",
			},
			want: `[["foo","bar"],"baz"]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: ToJSONString is called.
			got := ToJSONString(tc.input)

			// THEN: the JSON string is returned correctly.
			if got != tc.want {
				t.Errorf(
					"%s\nToJSONString(%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.input,
					got, tc.want,
				)
			}
		})
	}
}
