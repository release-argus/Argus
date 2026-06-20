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
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestIsNull(t *testing.T) {
	// GIVEN: []byte.
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "empty",
			data: []byte{},
			want: false,
		},
		{
			name: "non-empty",
			data: []byte{1, 2, 3},
			want: false,
		},
		{
			name: "null/bare",
			data: []byte("null"),
			want: true,
		},
		{
			name: "null/trailing spaces trimmed",
			data: []byte("null  "),
			want: true,
		},
		{
			name: "null/leading spaces trimmed",
			data: []byte("   null"),
			want: true,
		},
		{
			name: "null/leading and trailing spaces trimmed",
			data: []byte("   null "),
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsNull is called on it.
			got := IsNull(tc.data)

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nIsNull(%v) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.data, got, tc.want,
				)
			}
		})
	}
}

func TestUnsupportedFormatError_Error(t *testing.T) {
	// GIVEN: a UnsupportedFormatError.
	tests := []struct {
		name     string
		err      UnsupportedFormatError
		expected string
	}{
		{
			name: "xml",
			err: UnsupportedFormatError{
				Format: "xml",
			},
			expected: `unsupported format: "xml"`,
		},
		{
			name: "html",
			err: UnsupportedFormatError{
				Format: "html",
			},
			expected: `unsupported format: "html"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the error is stringified.
			got := tc.err.Error()

			// THEN: the error is formatted as expected.
			if got != tc.expected {
				t.Fatalf(
					"%s\nUnsupportedFormatError stringified mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.expected,
				)
			}
		})
	}
}

type testStruct struct {
	A string `json:"a,omitempty" yaml:"a,omitempty"`
	B int    `json:"b,omitempty" yaml:"b,omitempty"`
	C bool   `json:"c,omitempty" yaml:"c,omitempty"`
}

func TestUnmarshal(t *testing.T) {
	// GIVEN: []byte and a format.
	tests := []struct {
		name     string
		data     []byte
		format   string
		want     string
		errRegex string
	}{
		{
			name:   "JSON/no data",
			data:   []byte{},
			format: "json",
			errRegex: test.TrimYAML(`
				jsontext:
					unexpected EOF`,
			),
			want: "{}\n",
		},
		{
			name:     "JSON/empty object",
			data:     []byte("{}"),
			format:   "json",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/no data",
			data:     []byte{},
			format:   "yaml",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "unsupported format/no data",
			data:     []byte{},
			format:   "x",
			errRegex: `^unsupported format: "x"$`,
			want:     "{}\n",
		},
		{
			name: "JSON/valid data",
			data: []byte(test.TrimJSON(`{
				"a": "hi",
				"b": 6,
				"c": true
			}`)),
			format:   "json",
			errRegex: `^$`,
			want: test.TrimYAML(`
				a: hi
				b: 6
				c: true
			`),
		},
		{
			name: "YAML/valid data",
			data: []byte(test.TrimYAML(`
				a: foo
				b: 42
				c: true
			`)),
			format:   "yaml",
			errRegex: `^$`,
			want: test.TrimYAML(`
				a: foo
				b: 42
				c: true
			`),
		},
		{
			name:     "unsupported format/core",
			data:     []byte("{}"),
			format:   "txt",
			errRegex: `^unsupported format: "txt"$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Unmarshal is called on it.
			var target testStruct
			err := Unmarshal(tc.format, tc.data, &target)

			prefix := fmt.Sprintf(
				"%s\nUnmarshal(format=%s, data=%v)",
				packageName, tc.format, tc.data,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the result stringifies as expected.
			if got := ToYAMLString(target, ""); got != tc.want {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

type customUnmarshaler struct {
	got []byte
}

func (c *customUnmarshaler) UnmarshalJSON(data []byte) error {
	c.got = append([]byte(nil), data...)
	return nil
}

func (c *customUnmarshaler) UnmarshalYAML(data []byte) error {
	c.got = append([]byte(nil), data...)
	return nil
}

func TestUnmarshal__customUnmarshaler(t *testing.T) {
	// GIVEN: a target that implements a custom unmarshaler.
	tests := []struct {
		name   string
		format string
		data   []byte
		target any
		want   []byte
	}{
		{
			name:   "json.Unmarshaler",
			format: "json",
			data:   []byte(`{"a":"hi"}`),
			target: &customUnmarshaler{},
			want:   []byte(`{"a":"hi"}`),
		},
		{
			name:   "yaml.Unmarshaler",
			format: "yaml",
			data:   []byte("a: hi\n"),
			target: &customUnmarshaler{},
			want:   []byte("a: hi\n"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Unmarshal is called on it.
			err := Unmarshal(tc.format, tc.data, tc.target)

			prefix := fmt.Sprintf(
				"%s\nUnmarshal(format=%q, data=%q)",
				packageName, tc.format, tc.data,
			)

			// THEN: no error is returned.
			if err != nil {
				t.Fatalf(
					"%s unexpected error: %q",
					prefix, err,
				)
			}

			// AND: the custom unmarshaler received the raw data.
			got := tc.target.(*customUnmarshaler).got
			if string(got) != string(tc.want) {
				t.Errorf(
					"%s custom unmarshaler data mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	// GIVEN: struct and a format.
	tests := []struct {
		name     string
		data     testStruct
		format   string
		want     string
		errRegex string
	}{
		{
			name:     "JSON",
			data:     testStruct{A: "foo", B: 1, C: true},
			format:   "json",
			errRegex: `^$`,
			want: test.TrimJSON(`{
				"a": "foo",
				"b": 1,
				"c": true
			}`),
		},
		{
			name:     "YAML",
			data:     testStruct{A: "bar", B: 2, C: true},
			format:   "yaml",
			errRegex: `^$`,
			want: test.TrimYAML(`
				a: bar
				b: 2
				c: true
			`),
		},
		{
			name:     "unsupported format",
			data:     testStruct{A: "foo", B: 42, C: true},
			format:   "txt",
			errRegex: `^unsupported format: "txt"$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Marshal is called on it.
			gotBytes, err := Marshal(tc.format, tc.data)

			prefix := fmt.Sprintf(
				"%s\nMarshal(format=%q, data=%v)",
				packageName, tc.format, tc.data,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the result stringifies as expected.
			if got := string(gotBytes); got != tc.want {
				t.Errorf(
					"%s value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}
