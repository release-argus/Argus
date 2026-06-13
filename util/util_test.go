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

package util

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestStringToBoolPtr(t *testing.T) {
	// GIVEN: a string.
	tests := []struct {
		name  string
		input string
		want  *bool
	}{
		{
			name:  "'true' gives true",
			input: "true", want: test.Ptr(true),
		},
		{
			name:  "'false' gives false",
			input: "false", want: test.Ptr(false),
		},
		{
			name:  "'' gives nil",
			input: "", want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// WHEN: StringToBoolPtr is called.
			got := StringToBoolPtr(tc.input)

			prefix := fmt.Sprintf(
				"%s\nStringToBoolPtr(%q)",
				packageName, tc.input,
			)

			// THEN: the string is converted to a bool pointer.
			if got == tc.want {
				return
			}
			// One of them is nil, but the other is not. Or value mismatch.
			if (tc.want != nil && got == nil) || (tc.want == nil && got != nil) ||
				*got != *tc.want {
				t.Errorf(
					"%s value mismatch\ngot:  %v\nwant: %v",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestValueUnlessDefault(t *testing.T) {
	// GIVEN: a value to check and a value we want when it's not default.
	tests := []struct {
		name  string
		check string
		value string
		want  string
	}{
		{
			name:  "default `check` value",
			check: "", value: "foo",
			want: "",
		},
		{
			name:  "non-default `check` value",
			check: "foo", value: "bar",
			want: "bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: ValueUnlessDefault is run on pointer and a value.
			got := ValueUnlessDefault(tc.check, tc.value)

			prefix := fmt.Sprintf(
				"%s\nValueUnlessDefault(check=%q, value=%q)",
				packageName, tc.check, tc.value,
			)

			// THEN: the correct value is returned.
			if got != tc.want {
				t.Errorf(
					"%s value mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestValueOr(t *testing.T) {
	// GIVEN: two values.
	tests := []struct {
		name          string
		first, second string
		want          string
	}{
		{
			name:   "first value is non-default",
			first:  "foo",
			second: "bar",
			want:   "foo",
		},
		{
			name:   "first value is default",
			first:  "",
			second: "bar",
			want:   "bar",
		},
		{
			name:   "both values are default",
			first:  "",
			second: "",
			want:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: ValueOr is called with two values.
			got := ValueOr(tc.first, tc.second)

			prefix := fmt.Sprintf(
				"%s\nValueOr(a=%q, b=%q)",
				packageName, tc.first, tc.second,
			)

			// THEN: the correct value is returned.
			if got != tc.want {
				t.Errorf(
					"%s value mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestDerefOrZero(t *testing.T) {
	// GIVEN: a value to check and a value we want when it's nil.
	tests := []struct {
		name  string
		check *string
		value string
		want  string
	}{
		{
			name:  "nil `check` pointer",
			check: nil,
			want:  "",
		},
		{
			name:  "non-nil `check` pointer",
			check: test.Ptr("foo"),
			want:  "foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: DerefOrZero is run on pointer and a value.
			got := DerefOrZero(tc.check)

			prefix := fmt.Sprintf(
				"%s\nDerefOrZero(%v)",
				packageName, tc.check,
			)

			// THEN: the correct value is returned.
			if got != tc.want {
				t.Errorf(
					"%s value mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestDerefOr(t *testing.T) {
	// GIVEN: a bunch of comparables pointers and values.
	tests := []struct {
		name       string
		ptr, value any
		want       any
	}{
		{
			name:  "nil string pointer",
			ptr:   (*string)(nil),
			value: "argus",
			want:  "argus",
		},
		{
			name:  "non-nil string pointer",
			ptr:   test.Ptr("foo"),
			value: "bar",
			want:  "foo",
		},
		{
			name:  "nil bool pointer",
			ptr:   (*bool)(nil),
			value: false,
			want:  false,
		},
		{
			name:  "non-nil bool pointer",
			ptr:   test.Ptr(true),
			value: false,
			want:  true,
		},
		{
			name:  "nil int pointer",
			ptr:   (*int)(nil),
			value: 1,
			want:  1,
		},
		{
			name:  "non-nil int pointer",
			ptr:   test.Ptr(3),
			value: 2,
			want:  3,
		},
		{
			name:  "nil string slice",
			ptr:   (*[]string)(nil),
			value: []string{"baz"},
			want:  []string{"baz"},
		},
		{
			name:  "non-nil string slice",
			ptr:   test.Ptr([]string{"foo", "bar"}),
			value: []string{"baz"},
			want:  []string{"foo", "bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: DerefOr is called.
			var got any
			switch v := tc.ptr.(type) {
			case *string:
				got = DerefOr(v, tc.value.(string))
			case *bool:
				got = DerefOr(v, tc.value.(bool))
			case *int:
				got = DerefOr(v, tc.value.(int))
			case *[]string:
				got = DerefOr(v, tc.value.([]string))
			}

			prefix := fmt.Sprintf(
				"%s\nDerefOr(ptr=%v, b=%v)",
				packageName, tc.ptr, tc.value,
			)

			// THEN: the pointer is returned if it's nil, otherwise the value.
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf(
					"%s result mismatch\ngot:  %v\nwant: %v",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestPtrIfNotZero(t *testing.T) {
	// GIVEN: comparable values.
	tests := []struct {
		name  string
		value int
		want  *int
	}{
		{
			name:  "zero value",
			value: 0,
			want:  nil,
		},
		{
			name:  "non-zero value",
			value: 42,
			want:  test.Ptr(42),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: PtrIfNotZero is called.
			got := PtrIfNotZero(tc.value)

			prefix := fmt.Sprintf(
				"%s\nPtrIfNotZero(%v)",
				packageName, tc.value,
			)

			// THEN: nil is returned for zero values.
			if tc.want == nil {
				if got != nil {
					t.Errorf(
						"%s result mismatch\ngot:  %d\nwant: nil",
						prefix, *got,
					)
				}
				return
			}
			// AND: a pointer is returned for non-zero values.
			if got == nil || *got != *tc.want {
				t.Errorf(
					"%s result mismatch\ngot:  %v\nwant: %d",
					prefix, got, *tc.want,
				)
			}
		})
	}
}

func TestCopyPointer(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		doesCopy bool
	}{
		{
			name:     "nil pointer",
			input:    (*string)(nil),
			doesCopy: false,
		},
		{
			name:     "non-nil int pointer",
			input:    test.Ptr(6),
			doesCopy: true,
		},
		{
			name:     "non-nil string pointer",
			input:    test.Ptr("foo"),
			doesCopy: true,
		},
		{
			name:     "non-nil bool pointer",
			input:    test.Ptr(true),
			doesCopy: true,
		},
		{
			name:     "non-nil string slice pointer",
			input:    test.Ptr([]string{"foo", "bar"}),
			doesCopy: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: CopyPointer is called.
			var got any
			switch v := tc.input.(type) {
			case *string:
				got = ClonePtr(v)
			case *bool:
				got = ClonePtr(v)
			case *int:
				got = ClonePtr(v)
			case *[]string:
				got = ClonePtr(v)
			}

			prefix := fmt.Sprintf(
				"%s\nCopyPointer(%v)",
				packageName, tc.input,
			)

			// THEN: the result should be a pointer to a copy of the value.
			if (tc.doesCopy && got == nil) ||
				(tc.doesCopy && !reflect.DeepEqual(got, tc.input)) ||
				(!tc.doesCopy && got != nil && !reflect.ValueOf(tc.input).IsNil()) {
				t.Errorf(
					"%s value mismatch\ngot  %v, want %v",
					prefix, got, tc.input,
				)
			}
		})
	}
}

func TestRestoreMaskedValues(t *testing.T) {
	// GIVEN: maps with secrets to be copied.
	tests := []struct {
		name                  string
		input, copyFrom, want map[string]string
		fields                []string
	}{
		{
			name:  "empty map",
			input: map[string]string{},
			copyFrom: map[string]string{
				"foo": "bar",
			},
			want:   map[string]string{},
			fields: []string{"foo"},
		},
		{
			name: "copy only `SecretValue`s in fields",
			input: map[string]string{
				"test": SecretValue,
				"foo":  SecretValue,
			},
			copyFrom: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			want: map[string]string{
				"test": "123",
				"foo":  SecretValue,
			},
			fields: []string{"test"},
		},
		{
			name: "copy only `SecretValue`s in fields that also exist in from",
			input: map[string]string{
				"test": SecretValue,
				"foo":  SecretValue,
				"bar":  SecretValue,
			},
			copyFrom: map[string]string{
				"test": "123",
				"foo":  "bar",
			},
			want: map[string]string{
				"test": "123",
				"foo":  SecretValue,
				"bar":  SecretValue,
			},
			fields: []string{"test", "bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: RestoreMaskedValues is called.
			RestoreMaskedValues(tc.copyFrom, tc.input, tc.fields)

			prefix := fmt.Sprintf(
				"%s\nRestoreMaskedValues(from=%+v, to=%+v, fields=%+v)",
				packageName, tc.copyFrom, tc.input, tc.fields,
			)

			// THEN: the secrets are copied correctly.
			if testErr := test.AssertMapEqual(
				t,
				tc.input,
				tc.want,
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestIndentation(t *testing.T) {
	// GIVEN: a string with/without indentation.
	tests := []struct {
		name       string
		text       string
		indentSize int
		want       string
	}{
		{
			name:       "indent size 0",
			text:       "foo: bar",
			indentSize: 0,
			want:       "",
		},
		{
			name:       "no indent",
			text:       "foo: bar",
			indentSize: 2,
			want:       "",
		},
		{
			name:       "indent 4, indent size 4",
			text:       "    foo: bar",
			indentSize: 4,
			want:       "    ",
		},
		{
			name:       "indent 4, indent size 2",
			text:       "    foo: bar",
			indentSize: 2,
			want:       "    ",
		},
		{
			name:       "indent 3, indent size 2",
			text:       "   foo: bar",
			indentSize: 2,
			want:       "  ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Indentation is called on it.
			got := Indentation(tc.text, uint8(tc.indentSize))

			prefix := fmt.Sprintf(
				"%s\nIndentation(text=%q, indents=%d)",
				packageName, tc.text, tc.indentSize,
			)

			// THEN: the expected indentation is returned.
			if got != tc.want {
				t.Fatalf(
					"%s value mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestTruncateMessage(t *testing.T) {
	// GIVEN: a message and a maxLength to adhere to.
	tests := []struct {
		name      string
		msg       string
		maxLength int
		want      string
	}{
		{
			name:      "message shorter than maxLength",
			msg:       "short message",
			maxLength: 20,
			want:      "short message",
		},
		{
			name:      "message equal to maxLength",
			msg:       "exact length msg",
			maxLength: 16,
			want:      "exact length msg",
		},
		{
			name:      "message longer than maxLength",
			msg:       "is this message too long",
			maxLength: 10,
			want:      "is this me...",
		},
		{
			name:      "empty message",
			msg:       "",
			maxLength: 10,
			want:      "",
		},
		{
			name:      "maxLength zero",
			msg:       "message",
			maxLength: 0,
			want:      "...",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: TruncateMessage is called.
			got := TruncateMessage(tc.msg, tc.maxLength)

			prefix := fmt.Sprintf(
				"%s\ntruncateMessage(msg=%q, max=%d)",
				packageName, tc.msg, tc.maxLength,
			)

			// THEN: the message is truncated only if it exceeds maxLength.
			if got != tc.want {
				t.Errorf(
					"%s value mismatch\ngot:  %q\nwant: %q",
					prefix, tc.msg, tc.want,
				)
			}
		})
	}
}
