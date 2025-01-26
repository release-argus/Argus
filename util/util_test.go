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
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestDereferenceOrNilValue(t *testing.T) {
	// GIVEN lists of strings
	tests := map[string]struct {
		ptr    *string
		nilStr string
		want   string
	}{
		"nil *string": {
			ptr: nil, nilStr: "bar",
			want: "bar"},
		"non-nil *string": {
			ptr: test.StringPtr("foo"), nilStr: "bar",
			want: "foo"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN DereferenceOrNilValue is run on a pointer
			got := DereferenceOrNilValue(tc.ptr, tc.nilStr)

			// THEN the correct value is returned
			if got != tc.want {
				t.Errorf("want: %s\ngot:  %s",
					tc.want, got)
			}
		})
	}
}

func TestStringToBoolPtr(t *testing.T) {
	// GIVEN a string
	tests := map[string]struct {
		input string
		want  *bool
	}{
		"'true' gives true": {
			input: "true", want: test.BoolPtr(true)},
		"'false' gives false": {
			input: "false", want: test.BoolPtr(false)},
		"'' gives nil": {
			input: "", want: nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN StringToBoolPtr is called
			got := StringToBoolPtr(tc.input)

			// THEN the string is converted to a bool pointer
			if got == tc.want {
				return
			}
			// One of them is nil, but the other is not
			if (got == nil && tc.want != nil) || (tc.want == nil && got != nil) {
				t.Errorf("want: %v\ngot:  %v",
					tc.want, got)
			}
			// Not the same bool value
			if *got != *tc.want {
				t.Errorf("want: %v\ngot:  %v",
					tc.want, got)
			}
		})
	}
}

func TestValueUnlessDefault(t *testing.T) {
	// GIVEN a value to check and a value we want when it's not default
	tests := map[string]struct {
		check string
		value string
		want  string
	}{
		"default `check` value": {
			check: "", value: "foo",
			want: ""},
		"non-default `check` value": {
			check: "foo", value: "bar",
			want: "bar"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ValueUnlessDefault is run on pointer and a value
			got := ValueUnlessDefault(tc.check, tc.value)

			// THEN the correct value is returned
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestValueOrValue(t *testing.T) {
	// GIVEN two values
	tests := map[string]struct {
		first, second string
		want          string
	}{
		"first value is non-default": {
			first:  "foo",
			second: "bar",
			want:   "foo",
		},
		"first value is default": {
			first:  "",
			second: "bar",
			want:   "bar",
		},
		"both values are default": {
			first:  "",
			second: "",
			want:   "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ValueOrValue is called with two values
			got := ValueOrValue(tc.first, tc.second)

			// THEN the correct value is returned
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestDereferenceOrDefault(t *testing.T) {
	// GIVEN a value to check and a value we want when it's nil
	tests := map[string]struct {
		check *string
		value string
		want  string
	}{
		"nil `check` pointer": {
			check: nil,
			want:  ""},
		"non-nil `check` pointer": {
			check: test.StringPtr("foo"),
			want:  "foo"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN DereferenceOrDefault is run on pointer and a value
			got := DereferenceOrDefault(tc.check)

			// THEN the correct value is returned
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestDereferenceOrValue(t *testing.T) {
	// GIVEN a bunch of comparables
	tests := map[string]struct {
		element *string
		value   string
		want    string
	}{
		"nil pointer": {
			element: nil, want: ""},
		"non-nil pointer": {
			element: test.StringPtr("foo"), value: "bar", want: "bar"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN DereferenceOrValue is called
			got := DereferenceOrValue(tc.element, tc.value)

			// THEN the var is printed when it should be
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestPtrValueOrValue(t *testing.T) {
	// GIVEN a bunch of comparables pointers and values
	tests := map[string]struct {
		ptr, value interface{}
		want       interface{}
	}{
		"nil string pointer": {
			ptr:   (*string)(nil),
			value: "argus", want: "argus"},
		"non-nil string pointer": {
			ptr:   test.StringPtr("foo"),
			value: "bar", want: "foo"},
		"nil bool pointer": {
			ptr:   (*bool)(nil),
			value: false, want: false},
		"non-nil bool pointer": {
			ptr:   test.BoolPtr(true),
			value: false, want: true},
		"nil int pointer": {
			ptr:   (*int)(nil),
			value: 1, want: 1},
		"non-nil int pointer": {
			ptr:   test.IntPtr(3),
			value: 2, want: 3},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN PtrValueOrValue is called
			var got interface{}
			switch v := tc.ptr.(type) {
			case *string:
				got = PtrValueOrValue(v, tc.value.(string))
			case *bool:
				got = PtrValueOrValue(v, tc.value.(bool))
			case *int:
				got = PtrValueOrValue(v, tc.value.(int))
			}

			// THEN the pointer is returned if it's nil, otherwise the value
			if got != tc.want {
				t.Errorf("\nwant: %v\ngot:  %v", tc.want, got)
			}
		})
	}
}

func TestCopyPointer(t *testing.T) {
	tests := map[string]struct {
		input, want *int
	}{
		"nil pointer": {
			input: nil,
			want:  nil,
		},
		"non-nil pointer": {
			input: test.IntPtr(6),
			want:  test.IntPtr(6),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CopyPointer is called
			got := CopyPointer(tc.input)

			// THEN the result should be a pointer to a copy of the value
			if (got == nil && tc.want != nil) ||
				(got != nil && tc.want == nil) ||
				(got != nil && *got != *tc.want) {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestCopySecretValues(t *testing.T) {
	// GIVEN maps with secrets to be copied
	tests := map[string]struct {
		input, copyFrom, want map[string]string
		fields                []string
	}{
		"empty map": {
			input: map[string]string{},
			copyFrom: map[string]string{
				"foo": "bar"},
			want:   map[string]string{},
			fields: []string{"foo"},
		},
		"copy only `SecretValue`s in fields": {
			input: map[string]string{
				"test": SecretValue,
				"foo":  SecretValue},
			copyFrom: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  SecretValue},
			fields: []string{"test"},
		},
		"copy only `SecretValue`s in fields that also exist in from": {
			input: map[string]string{
				"test": SecretValue,
				"foo":  SecretValue,
				"bar":  SecretValue},
			copyFrom: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  SecretValue,
				"bar":  SecretValue},
			fields: []string{"test", "bar"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CopySecretValues is called
			CopySecretValues(tc.copyFrom, tc.input, tc.fields)

			// THEN the secrets are copied correctly
			if len(tc.input) != len(tc.want) {
				t.Fatalf("want: %v\ngot:  %v",
					tc.want, tc.input)
			}
			for i := range tc.input {
				if tc.input[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, tc.input)
				}
			}
		})
	}
}

type CustomErrorMarshal struct{}

func (c CustomErrorMarshal) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("intentional marshal error")
}

func TestTo____String(t *testing.T) {

	// GIVEN a struct to print in YAML format
	tests := map[string]struct {
		input              interface{}
		wantJSON, wantYAML string
	}{
		"invalid input": {
			input:    CustomErrorMarshal{},
			wantJSON: "{}",
			wantYAML: "",
		},
		"empty struct": {
			input:    struct{}{},
			wantJSON: "{}",
			wantYAML: "{}",
		},
		"simple struct": {
			input: struct {
				Test string `yaml:"test" json:"test"`
			}{
				Test: "test"},
			wantJSON: `{"test":"test"}`,
			wantYAML: "test: test",
		},
		"nested struct": {
			input: struct {
				Test struct {
					Foo string `yaml:"foo" json:"foo"`
				} `yaml:"test" json:"test"`
			}{
				Test: struct {
					Foo string `yaml:"foo" json:"foo"`
				}{
					Foo: "bar"}},
			wantJSON: `{"test":{"foo":"bar"}}`,
			wantYAML: "test:\n  foo: bar",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				wantYAML := strings.TrimPrefix(tc.wantYAML, "\n")
				if wantYAML != "" {
					if wantYAML != "{}" {
						wantYAML = prefix + strings.ReplaceAll(wantYAML, "\n", "\n"+prefix)
					}
					wantYAML += "\n"
				}

				// WHEN ToYAMLString is called
				gotYAML := ToYAMLString(tc.input, prefix)

				// THEN the struct is printed in YAML format
				if gotYAML != wantYAML {
					t.Fatalf("YAML (prefix=%q) want:\n%q\ngot:\n%q",
						prefix, wantYAML, gotYAML)
				}
			}

			// WHEN ToJSONString is called
			gotJSON := ToJSONString(tc.input)

			// THEN the struct is printed in JSON format
			if gotJSON != tc.wantJSON {
				t.Fatalf("JSON want:\n%q\ngot:\n%q",
					tc.wantJSON, gotJSON)
			}
		})
	}
}

func TestGetIndentation(t *testing.T) {
	// GIVEN a set of strings with varying indentation
	tests := map[string]struct {
		text       string
		indentSize int
		want       string
	}{
		"no indent": {
			text:       "foo: bar",
			indentSize: 2,
			want:       "",
		},
		"indent 4, indent size 4": {
			text:       "    foo: bar",
			indentSize: 4,
			want:       "    ",
		},
		"indent 4, indent size 2": {
			text:       "    foo: bar",
			indentSize: 2,
			want:       "    ",
		},
		"indent 3, indent size 2": {
			text:       "   foo: bar",
			indentSize: 2,
			want:       "  ",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Indentation is called on a string
			got := Indentation(tc.text, uint8(tc.indentSize))

			// THEN the expected indentation is returned
			if got != tc.want {
				t.Fatalf("want:%q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestTruncateMessage(t *testing.T) {
	// GIVEN a message and a maxLength to adhere to
	tests := map[string]struct {
		msg       string
		maxLength int
		want      string
	}{
		"message shorter than maxLength": {
			msg:       "short message",
			maxLength: 20,
			want:      "short message",
		},
		"message equal to maxLength": {
			msg:       "exact length msg",
			maxLength: 16,
			want:      "exact length msg",
		},
		"message longer than maxLength": {
			msg:       "is this message too long",
			maxLength: 10,
			want:      "is this me...",
		},
		"empty message": {
			msg:       "",
			maxLength: 10,
			want:      "",
		},
		"maxLength zero": {
			msg:       "message",
			maxLength: 0,
			want:      "...",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN TruncateMessage is called
			got := TruncateMessage(tc.msg, tc.maxLength)

			// THEN the message is truncated only if it exceeds maxLength
			if got != tc.want {
				t.Errorf("truncateMessage(%q, %d) = %q; want %q",
					tc.msg, tc.maxLength, got, tc.want)
			}
		})
	}
}
