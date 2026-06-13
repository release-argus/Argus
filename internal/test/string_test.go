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
	"testing"
)

func TestIndent(t *testing.T) {
	// GIVEN: a string and an indent value.
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

			// WHEN: Indent is called.
			result := Indent(tc.str, tc.indent)

			// THEN: the result should be the string with each line indented by the given number of spaces.
			if result != tc.want {
				t.Errorf(
					"%s\nIndent(str=%q, indents=%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.str, tc.indent,
					result, tc.want,
				)
			}
		})
	}
}

func TestTrimJSON(t *testing.T) {
	// GIVEN: a JSON string.
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
			name: "with spaces",
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

			// WHEN: TrimJSON is called.
			result := TrimJSON(tc.str)

			// THEN: the result should be the JSON string without newlines and tabs.
			if result != tc.want {
				t.Errorf(
					"%s\nTrimJSON() mismatch\ngot:  %q\nwant: %q",
					packageName, result, tc.want,
				)
			}
		})
	}
}

func TestTrimYAML(t *testing.T) {
	// GIVEN: a YAML string.
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

			// WHEN: TrimYAML is called.
			result := TrimYAML(tc.str)

			// THEN: the result should be the YAML string without unnecessary whitespace.
			if result != tc.want {
				t.Errorf(
					"%s\nTrimYAML(%q) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.str,
					result, tc.want,
				)
			}
		})
	}
}

func TestFakeT(t *testing.T) {
	// GIVEN: a FakeT.
	var fakeT FakeT
	fakeT.Helper()
	if len(fakeT.Errors) != 0 {
		t.Fatalf("%s\nfakeT should be empty", packageName)
	}

	// WHEN: Errorf is called.
	format, arg1 := "hello %s", "world"
	fakeT.Errorf(format, arg1)

	prefix := fmt.Sprintf(
		"%s\nFakeT.Errorf(format=%q, arg=%q)",
		packageName, format, arg1,
	)

	// THEN: the error is recorded.
	if len(fakeT.Errors) != 1 {
		t.Fatalf("%s should give fakeT an error", prefix)
	}
	want := fmt.Sprintf(format, arg1)
	if got := fakeT.Errors[0]; got != want {
		t.Fatalf(
			"%s format mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
	}

	// WHEN: Error is called.
	arg2 := "again"
	fakeT.Fatalf(format, arg2)

	prefix = fmt.Sprintf(
		"%s\nFakeT.Fatalf(format=%q, arg=%q)",
		packageName, format, arg2,
	)

	// THEN: the error is recorded.
	if len(fakeT.Errors) != 2 {
		t.Fatalf("%s should have given fakeT another error", prefix)
	}
	want = fmt.Sprintf(format, arg2)
	if got := fakeT.Errors[1]; got != want {
		t.Fatalf(
			"%s Fatalf mismatch\ngot:  %q\nwant: %q",
			packageName, fakeT.Errors[1], "hello again",
		)
	}

	// No-op.
	fakeT.Helper()
}

func TestAddPrefix(t *testing.T) {
	// GIVEN: a string and a prefix.
	tests := []struct {
		name   string
		want   string
		prefix string
		expect string
	}{
		{
			name:   "empty string",
			want:   "",
			prefix: "  ",
			expect: "",
		},
		{
			name:   "single line",
			want:   "hello",
			prefix: "- ",
			expect: "- hello",
		},
		{
			name:   "leading newline is stripped",
			want:   "\nhello",
			prefix: "  ",
			expect: "  hello",
		},
		{
			name:   "multi-line string",
			want:   "a\nb\nc",
			prefix: ">> ",
			expect: ">> a\n>> b\n>> c",
		},
		{
			name:   "multi-line with leading newline",
			want:   "\na\nb",
			prefix: "-- ",
			expect: "-- a\n-- b",
		},
		{
			name:   "json empty object",
			want:   "{}",
			prefix: "  ",
			expect: "  {}",
		},
		{
			name:   "json empty object with leading newline",
			want:   "\n{}",
			prefix: "  ",
			expect: "  {}",
		},
		{
			name:   "prefix is empty",
			want:   "hello\nworld",
			prefix: "",
			expect: "hello\nworld",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: addPrefix is called.
			got := addPrefix(tc.want, tc.prefix)

			// THEN: the prefix is added, as expected.
			if got != tc.expect {
				t.Fatalf(
					"%s\naddPrefix(str=%q, prefix=%q) mismatch\ngot:  %q\nwant: %q",
					packageName,
					tc.want, tc.prefix,
					got, tc.expect,
				)
			}
		})
	}
}
