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
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestIsNull(t *testing.T) {
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

			got := IsNull(tc.data)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestErrUnsupportedFormat_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrUnsupportedFormat
		expected string
	}{
		{
			name: "xml",
			err: ErrUnsupportedFormat{
				Format: "xml",
			},
			expected: `unsupported format: "xml"`,
		},
		{
			name: "html",
			err: ErrUnsupportedFormat{
				Format: "html",
			},
			expected: `unsupported format: "html"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.err.Error()

			if got != tc.expected {
				t.Fatalf(
					"ErrUnsupportedFormat.Error() mismatch\ngot:  %q\nwant: %q",
					got, tc.expected,
				)
			}
		})
	}
}

type testStruct struct {
	A string `json:"a,omitzero" yaml:"a,omitzero"`
	B int    `json:"b,omitzero" yaml:"b,omitzero"`
	C bool   `json:"c,omitzero" yaml:"c,omitzero"`
}

func TestUnmarshal(t *testing.T) {
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

			var target testStruct
			err := Unmarshal(tc.format, tc.data, &target)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if got := ToYAMLString(target, ""); got != tc.want {
				t.Errorf("ToYAMLString() mismatch\ngot:  %q\nwant: %q", got, tc.want)
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

			err := Unmarshal(tc.format, tc.data, tc.target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := tc.target.(*customUnmarshaler).got
			if string(got) != string(tc.want) {
				t.Errorf("customUnmarshaler.got mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
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

			gotBytes, err := Marshal(tc.format, tc.data)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if got := string(gotBytes); got != tc.want {
				t.Errorf("output mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
