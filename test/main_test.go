// Copyright [2024] [Argus]
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

import "testing"

func TestBoolPtr(t *testing.T) {
	// GIVEN a boolean value
	tests := map[string]struct {
		val bool
	}{
		"true":  {val: true},
		"false": {val: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN BoolPtr is called
			result := BoolPtr(tc.val)

			// THEN the result should be a pointer to the boolean value
			if *result != tc.val {
				t.Errorf("expected %t but got %t",
					tc.val, *result)
			}
		})
	}
}

func TestIntPtr(t *testing.T) {
	// GIVEN an integer value
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

			// WHEN IntPtr is called
			result := IntPtr(tc.val)

			// THEN the result should be a pointer to the integer value
			if *result != tc.val {
				t.Errorf("expected %d but got %d",
					tc.val, *result)
			}
		})
	}
}

func TestStringPtr(t *testing.T) {
	// GIVEN a string value
	tests := map[string]struct {
		val string
	}{
		"empty":     {val: ""},
		"non-empty": {val: "hello"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN StringPtr is called
			result := StringPtr(tc.val)

			// THEN the result should be a pointer to the string value
			if *result != tc.val {
				t.Errorf("expected %q but got %q",
					tc.val, *result)
			}
		})
	}
}

func TestUIntPtr(t *testing.T) {
	// GIVEN an integer value
	tests := map[string]struct {
		val uint
	}{
		"positive": {val: 1},
		"zero":     {val: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN UIntPtr is called
			result := UIntPtr(int(tc.val))

			// THEN the result should be a pointer to the unsigned integer value
			if *result != uint(tc.val) {
				t.Errorf("expected %d but got %d",
					tc.val, *result)
			}
		})
	}
}

func TestStringifyPtr(t *testing.T) {
	// GIVEN a pointer to a value
	tests := map[string]struct {
		ptr  interface{}
		want string
	}{
		"nil":           {ptr: nil, want: "nil"},
		"int, positive": {ptr: IntPtr(1), want: "1"},
		"int, negative": {ptr: IntPtr(-1), want: "-1"},
		"string":        {ptr: StringPtr("hello"), want: "hello"},
		"uint":          {ptr: UIntPtr(1), want: "1"},
		"bool":          {ptr: BoolPtr(true), want: "true"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN StringifyPtr is called
			var result string
			switch v := tc.ptr.(type) {
			case *bool:
				result = StringifyPtr(v)
			case *int:
				result = StringifyPtr(v)
			case *string:
				result = StringifyPtr(v)
			case *uint:
				result = StringifyPtr(v)
			case nil:
				var nilPtr *int
				result = StringifyPtr(nilPtr)
			default:
				t.Fatalf("unexpected type %T",
					tc.ptr)
			}

			// THEN the result should be a string representation of the value
			if result != tc.want {
				t.Errorf("expected %q but got %q",
					tc.want, result)
			}
		})
	}
}

func TestTrimJSON(t *testing.T) {
	// GIVEN a JSON string
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

			// WHEN TrimJSON is called
			result := TrimJSON(tc.str)

			// THEN the result should be the JSON string without newlines and tabs
			if result != tc.want {
				t.Errorf("expected %q but got %q",
					tc.want, result)
			}
		})
	}
}

func TestCombinations(t *testing.T) {
	// GIVEN a slice of values
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

			// WHEN Combinations is called
			result := Combinations(tc.input)

			// THEN the result should be all possible combinations of the values
			if len(result) != len(tc.want) {
				t.Fatalf("length differs:\nwant: %d, %v\ngot:  %d, %v",
					len(tc.want), tc.want, len(result), result)
			}
			for i, want := range tc.want {
				if len(result[i]) != len(want) {
					t.Fatalf("items differ in length:\nwant: %v\ngot:  %v",
						tc.want, result)
				}
				for j, v := range want {
					if result[i][j] != v {
						t.Fatalf("items of items differ:\nwant: %v\ngot:  %v",
							tc.want, result)
					}
				}
			}
		})
	}
}
