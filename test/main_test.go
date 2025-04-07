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

package test

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/util"
)

var packageName = "test"

func TestCaptureStdout(t *testing.T) {
	// GIVEN a function that writes to stdout.
	tests := map[string]struct {
		fn   func()
		want string
	}{
		"single line": {
			fn: func() {
				fmt.Println("hello")
			},
			want: "hello\n",
		},
		"multiple lines": {
			fn: func() {
				fmt.Println("hello")
				fmt.Println("world")
			},
			want: "hello\nworld\n",
		},
		"empty": {
			fn:   func() {},
			want: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			// WHEN CaptureStdout is called.
			capture := CaptureStdout()
			tc.fn()
			result := capture()

			// THEN the result should be the expected stdout output.
			if result != tc.want {
				t.Errorf("%s\nstdout mismatch\n%q\ngot:\n%q",
					packageName, tc.want, result)
			}
		})
	}
}

func TestBoolPtr(t *testing.T) {
	// GIVEN a boolean value.
	tests := map[string]struct {
		val bool
	}{
		"true":  {val: true},
		"false": {val: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN BoolPtr is called.
			result := BoolPtr(tc.val)

			// THEN the result should be a pointer to the boolean value.
			if *result != tc.val {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.val, *result)
			}
		})
	}
}

func TestIntPtr(t *testing.T) {
	// GIVEN an integer value.
	tests := map[string]struct {
		val int
	}{
		"positive": {val: 1},
		"zero":     {val: 0},
		"negative": {val: -1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN IntPtr is called.
			result := IntPtr(tc.val)

			// THEN the result should be a pointer to the integer value.
			if *result != tc.val {
				t.Errorf("%s\nwant: %d\ngot:  %d",
					packageName, tc.val, *result)
			}
		})
	}
}

func TestStringPtr(t *testing.T) {
	// GIVEN a string value.
	tests := map[string]struct {
		val string
	}{
		"empty":     {val: ""},
		"non-empty": {val: "hello"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN StringPtr is called.
			result := StringPtr(tc.val)

			// THEN the result should be a pointer to the string value.
			if *result != tc.val {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.val, *result)
			}
		})
	}
}

func TestUInt8Ptr(t *testing.T) {
	// GIVEN an integer value.
	tests := map[string]struct {
		val uint
	}{
		"positive": {val: 1},
		"zero":     {val: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN UInt8Ptr is called.
			result := UInt8Ptr(int(tc.val))

			// THEN the result should be a pointer to the unsigned integer value.
			if *result != uint8(tc.val) {
				t.Errorf("%s\nwant: %d\ngot:  %d",
					packageName, tc.val, *result)
			}
		})
	}
}

func TestStringifyPtr(t *testing.T) {
	// GIVEN a pointer to a value.
	tests := map[string]struct {
		ptr  any
		want string
	}{
		"nil":           {ptr: nil, want: "<nil>"},
		"int, positive": {ptr: IntPtr(1), want: "1"},
		"int, negative": {ptr: IntPtr(-1), want: "-1"},
		"string":        {ptr: StringPtr("hello"), want: "hello"},
		"uint":          {ptr: UInt8Ptr(1), want: "1"},
		"bool":          {ptr: BoolPtr(true), want: "true"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN StringifyPtr is called.
			var result string
			switch v := tc.ptr.(type) {
			case *bool:
				result = StringifyPtr(v)
			case *int:
				result = StringifyPtr(v)
			case *string:
				result = StringifyPtr(v)
			case *uint8:
				result = StringifyPtr(v)
			case nil:
				var nilPtr *int
				result = StringifyPtr(nilPtr)
			default:
				t.Fatalf("%s\nunexpected type %T",
					packageName, tc.ptr)
			}

			// THEN the result should be a string representation of the value.
			if result != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, result)
			}
		})
	}
}

func TestTrimJSON(t *testing.T) {
	// GIVEN a JSON string.
	tests := map[string]struct {
		str  string
		want string
	}{
		"empty": {
			str:  "",
			want: "",
		},
		"single line": {
			str:  `{"key": "value"}`,
			want: `{"key":"value"}`,
		},
		"multi line": {
			str: `
{
"key": "value"
}`,
			want: `{"key":"value"}`,
		},
		"with tabs": {
			str: `{
				"key": "value"
			}`,
			want: `{"key":"value"}`,
		},
		"with spaces": {
			str: `{
				"key": "value"
			}`,
			want: `{"key":"value"}`,
		},
		"mixed": {
			str: `{
				"key": "value",
				"key2": "value2"
			}`,
			want: `{"key":"value","key2":"value2"}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN TrimJSON is called.
			result := TrimJSON(tc.str)

			// THEN the result should be the JSON string without newlines and tabs.
			if result != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, result)
			}
		})
	}
}

func TestTrimYAML(t *testing.T) {
	// GIVEN a YAML string.
	tests := map[string]struct {
		str  string
		want string
	}{
		"empty": {
			str:  "",
			want: "",
		},
		"single line": {
			str:  "key1: value",
			want: "key1: value",
		},
		"multi line": {
			str: `key1: value
key2: value2`,
			want: `key1: value
key2: value2`,
		},
		"with leading newline": {
			str: `
key1: value
key2: value2`,
			want: `key1: value
key2: value2`,
		},
		"with tabs": {
			str: `	key1: value
	key2: value2`,
			want: `key1: value
key2: value2`,
		},
		"with spaces": {
			str: `  key1: value
  key2: value2`,
			want: `key1: value
key2: value2`,
		},
		"mixed": {
			str: `
	key1:
	  key1.1: value1.1
	  key1.2: value1.2
		key1.3: value1.3
	key2: value2`,
			want: `key1:
  key1.1: value1.1
  key1.2: value1.2
  key1.3: value1.3
key2: value2`,
		},
		"clear whitespace-only lines": {
			str: `

key1: value

key2: value2

`,
			want: `
key1: value

key2: value2

`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN TrimYAML is called.
			result := TrimYAML(tc.str)

			// THEN the result should be the YAML string without unnecessary whitespace.
			if result != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, result)
			}
		})
	}
}

func TestFlattenMultilineString(t *testing.T) {
	// GIVEN a multiline string.
	tests := map[string]struct {
		str  string
		want string
	}{
		"empty": {
			str:  "",
			want: "",
		},
		"single line": {
			str:  "line",
			want: "line",
		},
		"multi line": {
			str:  "line1\nline2",
			want: "line1 line2",
		},
		"multi line with leading spaces": {
			str:  "  line1\n  line2",
			want: "line1 line2",
		},
		"multi line with mixed whitespace": {
			str:  "line1\n\tline2\n  line3",
			want: "line1   line2   line3",
		},
		"multi line with empty lines": {
			str:  "\n  line1\n  \t\n  line2",
			want: "line1  line2",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN FlattenMultilineString is called.
			result := FlattenMultilineString(tc.str)

			// THEN the result should be the string with newlines and tabs replaced with spaces.
			if result != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, result)
			}
		})
	}
}

func TestCombinations(t *testing.T) {
	// GIVEN a slice of values.
	tests := map[string]struct {
		input []int
		want  [][]int
	}{
		"empty": {
			input: []int{},
			want:  [][]int{},
		},
		"single": {
			input: []int{1},
			want:  [][]int{{1}},
		},
		"two": {
			input: []int{1, 2},
			want: [][]int{
				{1},
				{1, 2},
				{2},
			},
		},
		"three": {
			input: []int{1, 2, 3},
			want: [][]int{
				{1},
				{1, 2},
				{1, 2, 3},
				{1, 3},
				{2},
				{2, 3},
				{3},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Combinations is called.
			result := Combinations(tc.input)

			// THEN the result should be all possible combinations of the values.
			if len(result) != len(tc.want) {
				t.Fatalf("%s\nlength mismatch\nwant: %d, %v\ngot:  %d, %v",
					packageName,
					len(tc.want), tc.want,
					len(result), result)
			}
			for i, want := range tc.want {
				if len(result[i]) != len(want) {
					t.Fatalf("%s\nitems at [%d] differ in length:\nwant: %v\ngot:  %v",
						packageName, i,
						tc.want, result)
				}
				for j, v := range want {
					if result[i][j] != v {
						t.Fatalf("%s\nitems of items at [%d][%d] differ:\nwant: %v\ngot:  %v",
							packageName,
							i, j,
							tc.want, result)
					}
				}
			}
		})
	}
}

func TestIndent(t *testing.T) {
	// GIVEN a string and an indent value.
	tests := map[string]struct {
		str    string
		indent int
		want   string
	}{
		"empty string": {
			str:    "",
			indent: 2,
			want:   "",
		},
		"single line": {
			str:    "line",
			indent: 2,
			want:   "line",
		},
		"multi line": {
			str:    "line1\nline2",
			indent: 2,
			want:   "line1\n  line2",
		},
		"multi line with different indent": {
			str:    "line1\nline2",
			indent: 4,
			want:   "line1\n    line2",
		},
		"multi line with zero indent": {
			str:    "line1\nline2",
			indent: 0,
			want:   "line1\nline2",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Indent is called.
			result := Indent(tc.str, tc.indent)

			// THEN the result should be the string with each line indented by the given number of spaces.
			if result != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, result)
			}
		})
	}
}

func TestIgnoreError(t *testing.T) {
	tests := map[string]struct {
		fn    func() (int, error)
		panic bool
		want  int
	}{
		"no error": {
			fn: func() (int, error) {
				return 42, nil
			},
			want: 42,
		},
		"with error": {
			fn: func() (int, error) {
				return 6, fmt.Errorf("some error")
			},
			panic: true,
			want:  0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN IgnoreError is called.
			result := IgnoreError(t, tc.fn)

			// THEN the result should be the expected value.
			if result != tc.want {
				t.Errorf("%s\nwant: %d\ngot:  %d",
					packageName, tc.want, result)
			}
		})
	}
}

func TestYAMLToNode(t *testing.T) {
	tests := map[string]struct {
		yamlStr  string
		errRegex string
	}{
		"valid YAML": {
			yamlStr: `
				key: value
			`,
			errRegex: `^$`,
		},
		"invalid YAML": {
			yamlStr: `
				key: [unclosed
			`,
			errRegex: `^yaml:.*did not find expected.*$`,
		},
		"empty YAML": {
			yamlStr:  ``,
			errRegex: `^$`,
		},
		"complex YAML": {
			yamlStr: `
				key1:
					key2: value2
					key3:
					- item1
					- item2
				`,
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.yamlStr = TrimYAML(tc.yamlStr)

			// WHEN YAMLToNode is called.
			node, err := YAMLToNode(t, tc.yamlStr)

			// THEN the result should be a valid yaml.Node, or an error.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			} else if node == nil && err == nil {
				t.Errorf("%s\nexpected node but got nil",
					packageName)
			}
		})
	}
}

func TestEqualSlices(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []int
		expected bool
	}{
		{
			name:     "Equal slices",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 3},
			expected: true,
		},
		{
			name:     "Different lengths",
			a:        []int{1, 2, 3},
			b:        []int{1, 2},
			expected: false,
		},
		{
			name:     "Same length, different elements",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 4},
			expected: false,
		},
		{
			name:     "Both slices empty",
			a:        []int{},
			b:        []int{},
			expected: true,
		},
		{
			name:     "One slice empty",
			a:        []int{},
			b:        []int{1},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := EqualSlices(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("%s\nEqualSlices(%v, %v) mismatch\nwant: %v\ngot:  %v",
					packageName,
					tc.a, tc.b,
					tc.expected, result)
			}
		})
	}
}
