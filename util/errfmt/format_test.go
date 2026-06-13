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

package errfmt

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config/decode"
)

func TestFormatError(t *testing.T) {
	// GIVEN: an error.
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		// Base cases.
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "joined nil errors",
			err:      errors.Join(nil, nil),
			expected: "",
		},
		{
			name:     "single flat error",
			err:      fmt.Errorf("root"),
			expected: "root",
		},

		// Single unwrap.
		{
			name: "single unwrap simple",
			err: fmt.Errorf(
				"outer: %w",
				fmt.Errorf("inner"),
			),
			expected: strings.TrimPrefix(`
outer:
  inner`, "\n"),
		},
		{
			name: "single unwrap (no colon)",
			err: fmt.Errorf(
				"outer message %w",
				fmt.Errorf("inner"),
			),
			expected: strings.TrimPrefix(`
outer message
  inner`, "\n"),
		},

		// Deep unwrap chains.
		{
			name: "two level unwrap chain",
			err: fmt.Errorf(
				"a: %w",
				fmt.Errorf(
					"b: %w",
					fmt.Errorf("c"),
				),
			),
			expected: strings.TrimPrefix(`
a:
  b:
    c`, "\n"),
		},
		{
			name: "three level unwrap chain",
			err: fmt.Errorf(
				"a: %w",
				fmt.Errorf(
					"b: %w",
					fmt.Errorf(
						"c: %w",
						fmt.Errorf("d"),
					),
				),
			),
			expected: strings.TrimPrefix(`
a:
  b:
    c:
      d`, "\n"),
		},

		// Multi-line error messages.
		{
			name: "multi-line parent error",
			err: fmt.Errorf(
				"outer\nline2: %w",
				fmt.Errorf("inner"),
			),
			expected: strings.TrimPrefix(`
outer
line2:
  inner`, "\n"),
		},
		{
			name: "multi-line child error",
			err: fmt.Errorf(
				"outer: %w",
				fmt.Errorf("inner\nline2"),
			),
			expected: strings.TrimPrefix(`
outer:
  inner
  line2`, "\n"),
		},
		{
			name: "multi-line both parent and child",
			err: fmt.Errorf(
				"outer\nlineB: %w",
				fmt.Errorf("inner\nlineC"),
			),
			expected: strings.TrimPrefix(`
outer
lineB:
  inner
  lineC`, "\n"),
		},

		// Whitespace normalisation cases.
		{
			name:     "trailing newline trimmed",
			err:      fmt.Errorf("error\n"),
			expected: "error",
		},
		{
			name: "inner trailing newline trimmed",
			err: fmt.Errorf(
				"outer: %w",
				fmt.Errorf("inner\n"),
			),
			expected: strings.TrimPrefix(`
outer:
  inner`, "\n"),
		},

		// Joined errors.
		{
			name: "joined errors",
			err: errors.Join(
				fmt.Errorf("first"),
				fmt.Errorf("second"),
			),
			expected: strings.TrimPrefix(`
first
second`, "\n"),
		},
		{
			name: "joined errors with single wrap",
			err: errors.Join(
				fmt.Errorf(
					"first: %w",
					fmt.Errorf("child"),
				),
				fmt.Errorf(
					"second: %w",
					fmt.Errorf("child"),
				),
			),
			expected: strings.TrimPrefix(`
first:
  child
second:
  child`, "\n"),
		},
		{
			name: "joined errors with multiple wraps and joins",
			err: errors.Join(
				fmt.Errorf("hello"),
				fmt.Errorf(
					"a: %w",
					fmt.Errorf("1"),
				),
				fmt.Errorf(
					"b: %w",
					errors.Join(
						fmt.Errorf("b.1"),
						fmt.Errorf("b.2"),
						fmt.Errorf(
							"what: %w",
							fmt.Errorf("bye"),
						),
					),
				),
			),
			expected: strings.TrimPrefix(`
hello
a:
  1
b:
  b.1
  b.2
  what:
    bye`, "\n"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// WHEN: FormatError is called on it.
			got := FormatError(tc.err)

			// THEN: the error is formatted as expected.
			if got != tc.expected {
				t.Fatalf(
					"%s\nFormatError() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.expected,
				)
			}
		})
	}
}

func TestAppendFormattedErrorLines(t *testing.T) {
	// GIVEN: an error.
	tests := []struct {
		name string
		err  error
		want []string
	}{
		{
			name: "nil error",
			err:  nil,
			want: []string{},
		},
		{
			name: "single error",
			err:  fmt.Errorf("error"),
			want: []string{"error"},
		},
		{
			name: "multiple errors",
			err: errors.Join(
				fmt.Errorf("first"),
				fmt.Errorf("second"),
			),
			want: []string{
				"first",
				"second",
			},
		},
		{
			name: "wrapped error",
			err: fmt.Errorf(
				"key: %w",
				fmt.Errorf("error"),
			),
			want: []string{
				"key:",
				"  error",
			},
		},
		{
			name: "wrapped custom error",
			err: &decode.KeyFieldError{
				Key: "key",
				Err: fmt.Errorf("error"),
			},
			want: []string{
				"key:",
				"  error",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []string

			// WHEN: appendFormattedErrorLines is called.
			lines = appendFormattedErrorLines(tc.err, lines, 0)

			// THEN: the number of lines is as expected.
			if got, want := len(lines), len(tc.want); got != want {
				t.Fatalf(
					"%s\nappendFormattedErrorLines() length mismatch\ngot:  %d\nwant: %d",
					packageName, got, want,
				)
			}

			// AND: the lines are as expected.
			for i, wantLine := range tc.want {
				if lines[i] != wantLine {
					t.Fatalf(
						"%s\nappendFormattedErrorLines() mismatch at index %d\ngot:  %v\nwant: %v",
						packageName, i,
						lines, tc.want,
					)
				}
			}
		})
	}
}
