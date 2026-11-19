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
	"reflect"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestStringToBoolPtr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *bool
	}{
		{
			name:  "'true' gives true",
			input: "true", want: new(true),
		},
		{
			name:  "'false' gives false",
			input: "false", want: new(false),
		},
		{
			name:  "'' gives nil",
			input: "", want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := StringToBoolPtr(tc.input)

			gotStr, wantStr := test.StringifyPtr(got), test.StringifyPtr(tc.want)
			if gotStr != wantStr {
				t.Errorf("mismatch\ngot:  %s\nwant: %s", gotStr, wantStr)
			}
		})
	}
}

func TestValueUnlessZero(t *testing.T) {
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

			got := ValueUnlessZero(tc.check, tc.value)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestValueOr(t *testing.T) {
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

			got := ValueOr(tc.first, tc.second)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestDerefOrZero(t *testing.T) {
	tests := []struct {
		name  string
		check *string
		want  string
	}{
		{
			name:  "nil `check` pointer",
			check: nil,
			want:  "",
		},
		{
			name:  "non-nil `check` pointer",
			check: new("foo"),
			want:  "foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := DerefOrZero(tc.check)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestDerefOr(t *testing.T) {
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
			ptr:   new("foo"),
			value: "bar",
			want:  "foo",
		},
		{
			name:  "nil bool pointer",
			ptr:   (*bool)(nil),
			value: true,
			want:  true,
		},
		{
			name:  "non-nil bool pointer",
			ptr:   new(true),
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
			ptr:   new(3),
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
			ptr:   new([]string{"foo", "bar"}),
			value: []string{"baz"},
			want:  []string{"foo", "bar"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("mismatch\ngot:  %v\nwant: %v", got, tc.want)
			}
		})
	}
}

func TestPtrIfNotZero(t *testing.T) {
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
			want:  new(42),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := PtrIfNotZero(tc.value)

			gotStr, wantStr := test.StringifyPtr(got), test.StringifyPtr(tc.want)
			if gotStr != wantStr {
				t.Errorf("mismatch\ngot:  %s\nwant: %s", gotStr, wantStr)
			}
		})
	}
}

func TestClonePtr(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{
			name:  "nil",
			input: (*string)(nil),
		},
		{
			name:  "non-nil int",
			input: new(6),
		},
		{
			name:  "non-nil string",
			input: new("foo"),
		},
		{
			name:  "non-nil bool",
			input: new(true),
		},
		{
			name:  "non-nil string slice",
			input: new([]string{"foo", "bar"}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
			default:
				t.Fatalf("unexpected type %T", tc.input)
			}

			gotStr, wantStr := test.StringifyPtr(got), test.StringifyPtr(tc.input)
			if gotStr != wantStr {
				t.Errorf("mismatch\ngot:  %s\nwant: %s", gotStr, wantStr)
			}

			if gotStr != "nil" && got == tc.input {
				t.Errorf("copy aliases the original\ngot:  %p\nwant: NOT %p", got, tc.input)
			}
		})
	}
}

func TestIndentation(t *testing.T) {
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

			got := Indentation(tc.text, uint8(tc.indentSize))

			if got != tc.want {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestTruncateMessage(t *testing.T) {
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

			got := TruncateMessage(tc.msg, tc.maxLength)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
