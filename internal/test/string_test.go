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
	"testing"
)

func TestIndent(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		indent int
		want   string
	}{
		{
			name:   "empty string",
			str:    "",
			indent: 2,
			want:   "",
		},
		{
			name:   "single line",
			str:    "line",
			indent: 2,
			want:   "line",
		},
		{
			name:   "multi line",
			str:    "line1\nline2",
			indent: 2,
			want:   "line1\n  line2",
		},
		{
			name:   "multi line with different indent",
			str:    "line1\nline2",
			indent: 4,
			want:   "line1\n    line2",
		},
		{
			name:   "multi line with zero indent",
			str:    "line1\nline2",
			indent: 0,
			want:   "line1\nline2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := Indent(tc.str, tc.indent)

			if result != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", result, tc.want)
			}
		})
	}
}

func TestTrimJSON(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "empty",
			str:  "",
			want: "",
		},
		{
			name: "single line",
			str:  `{"key": "value"}`,
			want: `{"key":"value"}`,
		},
		{
			name: "multi line",
			str: `
{
"key": "value"
}`,
			want: `{"key":"value"}`,
		},
		{
			name: "with tabs",
			str: `{
				"key": "value"
			}`,
			want: `{"key":"value"}`,
		},
		{
			name: "mixed",
			str: `{
				"key": "value",
				"key2": "value2"
			}`,
			want: `{"key":"value","key2":"value2"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := TrimJSON(tc.str)

			if result != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", result, tc.want)
			}
		})
	}
}

func TestTrimYAML(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "empty",
			str:  "",
			want: "",
		},
		{
			name: "single line",
			str:  "key1: value",
			want: "key1: value",
		},
		{
			name: "multi line",
			str: `key1: value
key2: value2`,
			want: `key1: value
key2: value2`,
		},
		{
			name: "with leading newline",
			str: `
key1: value
key2: value2`,
			want: `key1: value
key2: value2`,
		},
		{
			name: "with tabs",
			str: `	key1: value
	key2: value2`,
			want: `key1: value
key2: value2`,
		},
		{
			name: "with spaces",
			str: `  key1: value
  key2: value2`,
			want: `key1: value
key2: value2`,
		},
		{
			name: "mixed",
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
		{
			name: "indented list",
			str: `
				- key1.1: value1.1
				- key1.2: value1.2
				- key1.3: value1.3
			`,
			want: `- key1.1: value1.1
- key1.2: value1.2
- key1.3: value1.3
`,
		},
		{
			name: "clear whitespace-only lines",
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := TrimYAML(tc.str)

			if result != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", result, tc.want)
			}
		})
	}
}

func TestAddPrefix(t *testing.T) {
	tests := []struct {
		name   string
		target string
		prefix string
		expect string
	}{
		{
			name:   "empty string",
			target: "",
			prefix: "  ",
			expect: "",
		},
		{
			name:   "single line",
			target: "hello",
			prefix: "- ",
			expect: "- hello",
		},
		{
			name:   "leading newline is stripped",
			target: "\nhello",
			prefix: "  ",
			expect: "  hello",
		},
		{
			name:   "multi-line string",
			target: "a\nb\nc",
			prefix: ">> ",
			expect: ">> a\n>> b\n>> c",
		},
		{
			name:   "multi-line with leading newline",
			target: "\na\nb",
			prefix: "-- ",
			expect: "-- a\n-- b",
		},
		{
			name:   "json empty object",
			target: "{}",
			prefix: "  ",
			expect: "  {}",
		},
		{
			name:   "json empty object with trailing newline",
			target: "{}\n",
			prefix: "  ",
			expect: "{}\n",
		},
		{
			name:   "json empty object with leading newline",
			target: "\n{}",
			prefix: "  ",
			expect: "  {}",
		},
		{
			name:   "prefix is empty",
			target: "hello\nworld",
			prefix: "",
			expect: "hello\nworld",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := addPrefix(tc.target, tc.prefix)

			if got != tc.expect {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.expect)
			}
		})
	}
}
