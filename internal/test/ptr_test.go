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
	"strings"
	"testing"
)

func TestPtr_Bool(t *testing.T) {
	// GIVEN: a boolean value.
	tests := []struct {
		val bool
	}{
		{val: true},
		{val: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("val=%t", tc.val)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Ptr is called.
			result := Ptr(tc.val)

			// THEN: the result should be a pointer to the boolean value.
			if *result != tc.val {
				t.Errorf(
					"%s\nPtr[bool](%t) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.val,
					*result, tc.val,
				)
			}
		})
	}
}

func TestPtr_Int(t *testing.T) {
	// GIVEN: an integer value.
	tests := []struct {
		name string
		val  int
	}{
		{name: "positive", val: 1},
		{name: "zero", val: 0},
		{name: "negative", val: -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Ptr is called.
			result := Ptr(tc.val)

			// THEN: the result should be a pointer to the integer value.
			if *result != tc.val {
				t.Errorf(
					"%s\nPtr[int](%d) mismatch\ngot:  %d\nwant: %d",
					packageName, tc.val,
					*result, tc.val,
				)
			}
		})
	}
}

func TestPtr_String(t *testing.T) {
	// GIVEN: a string value.
	tests := []struct {
		name string
		val  string
	}{
		{name: "empty", val: ""},
		{name: "non-empty", val: "hello"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Ptr is called.
			result := Ptr(tc.val)

			// THEN: the result should be a pointer to the string value.
			if *result != tc.val {
				t.Errorf(
					"%s\nPtr[string](%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.val,
					*result, tc.val,
				)
			}
		})
	}
}

func TestStringSlicePtr(t *testing.T) {
	// GIVEN: a string slice value.
	tests := []struct {
		name string
		val  []string
	}{
		{name: "empty", val: []string{}},
		{name: "single element", val: []string{"hello"}},
		{name: "multiple elements", val: []string{"hello", "argus"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Ptr is called on this value.
			result := Ptr(tc.val)

			prefix := fmt.Sprintf(
				"%s\nPtr[[]string](%v)",
				packageName, tc.val,
			)

			// THEN: the result should be a pointer to the string slice value.
			want := strings.Join(tc.val, ", ")
			if result == nil {
				t.Fatalf(
					"%s mismatch\ngot:  <nil>\nwant: %q",
					prefix, want,
				)
			}
			got := strings.Join(*result, ", ")
			if len(*result) != len(tc.val) {
				t.Errorf(
					"%s result mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}
			for i, val := range *result {
				if val != tc.val[i] {
					t.Errorf(
						"%s[%d] value mismatch\ngot:  %q (%q)\nwant: %q (%q)",
						prefix, i,
						val, got,
						tc.val[i], want,
					)
				}
			}
		})
	}
}

func TestPtr_UInt8(t *testing.T) {
	// GIVEN: a UInt7 value.
	tests := []struct {
		name string
		val  uint8
	}{
		{name: "positive", val: 1},
		{name: "zero", val: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Ptr is called.
			result := Ptr(tc.val)

			// THEN: the result should be a pointer to the unsigned integer value.
			if *result != tc.val {
				t.Errorf(
					"%s\nPtr[uint8](%d) mismatch\ngot:  %d\nwant: %d",
					packageName, tc.val,
					*result, tc.val,
				)
			}
		})
	}
}

func TestPtr_UInt16(t *testing.T) {
	// GIVEN: a UInt16 value.
	tests := []struct {
		name string
		val  uint16
	}{
		{name: "positive", val: 1},
		{name: "zero", val: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Ptr is called.
			result := Ptr(tc.val)

			// THEN: the result should be a pointer to the unsigned integer value.
			if *result != tc.val {
				t.Errorf(
					"%s\nPtr[uint16](%d) mismatch\ngot:  %d\nwant: %d",
					packageName, tc.val,
					*result, tc.val,
				)
			}
		})
	}
}

func TestStringifyPtr(t *testing.T) {
	// GIVEN: a pointer to a value.
	tests := []struct {
		name string
		ptr  any
		want string
	}{
		{name: "nil", ptr: nil, want: "<nil>"},
		{name: "int, positive", ptr: Ptr(1), want: "1"},
		{name: "int, negative", ptr: Ptr(-1), want: "-1"},
		{name: "string", ptr: Ptr("hello"), want: "hello"},
		{name: "uint", ptr: Ptr(1), want: "1"},
		{name: "bool", ptr: Ptr(true), want: "true"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: StringifyPtr is called.
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
				t.Fatalf("%s\nunexpected type %T", packageName, tc.ptr)
			}

			// THEN: the result should be a string representation of the value.
			if result != tc.want {
				t.Errorf(
					"%s\nStringifyPtr() mismatch\ngot:  %q\nwant: %q",
					packageName, result, tc.want,
				)
			}
		})
	}
}
