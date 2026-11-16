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
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "base/nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "base/joined nil errors",
			err:      errors.Join(nil, nil),
			expected: "",
		},
		{
			name:     "base/single flat error",
			err:      errors.New("root"),
			expected: "root",
		},
		{
			name: "single unwrap/simple",
			err: fmt.Errorf(
				"outer: %w",
				errors.New("inner"),
			),
			expected: `
outer:
  inner`,
		},
		{
			name: "single unwrap/no colon",
			err: fmt.Errorf(
				"outer message %w",
				errors.New("inner"),
			),
			expected: `
outer message
  inner`,
		},
		{
			name: "single unwrap/trailing wrap whose text the parent also starts with",
			err: fmt.Errorf(
				"timeout waiting: %w",
				errors.New("timeout"),
			),
			expected: `
timeout waiting:
  timeout`,
		},
		{
			name: "single unwrap/wrapped error at beginning",
			err: fmt.Errorf(
				"%w: test:read @ global",
				errors.New("invalid grant"),
			),
			expected: `
invalid grant: test:read @ global`,
		},
		{
			name: "single unwrap/wrapped error at beginning without colon",
			err: fmt.Errorf(
				"%w test:read @ global",
				errors.New("invalid grant"),
			),
			expected: `
invalid grant test:read @ global`,
		},
		{
			name: "single unwrap/wrapped error at beginning with trailing newline",
			err: fmt.Errorf(
				"%w: context\n",
				errors.New("invalid grant"),
			),
			expected: `
invalid grant: context`,
		},
		{
			name: "single unwrap/bare wrap",
			err:  fmt.Errorf("%w", fmt.Errorf("outer: %w", errors.New("inner"))),
			expected: `
outer:
  inner`,
		},
		{
			name: "single unwrap/wrapped error in the middle",
			err:  fmt.Errorf("before %w after", errors.New("inner")),
			expected: `
before inner after`,
		},
		{
			name: "single unwrap/wrapped text glued to parent text",
			err:  fmt.Errorf("timeout%w", errors.New("timeout")),
			expected: `
timeouttimeout`,
		},
		{
			name: "two wraps/category then cause",
			err: fmt.Errorf(
				"%w: salt: %w",
				errors.New("malformed hash"), errors.New("illegal base64"),
			),
			expected: `
malformed hash:
  salt: illegal base64`,
		},
		{
			name: "two wraps/category then joined causes",
			err: fmt.Errorf(
				"%w: %w",
				errors.New("invalid preference"),
				errors.Join(
					errors.New("first"),
					errors.New("second"),
				),
			),
			expected: `
invalid preference:
  first
  second`,
		},
		{
			name: "two level unwrap chain/wrapped errors at end",
			err: fmt.Errorf(
				"a: %w",
				fmt.Errorf(
					"b: %w",
					errors.New("c"),
				),
			),
			expected: `
a:
  b:
    c`,
		},
		{
			name: "two level unwrap chain/wrapped errors at beginning",
			err: fmt.Errorf(
				"%w: context A",
				fmt.Errorf(
					"%w: context B",
					errors.New("root"),
				),
			),
			expected: `
root: context B: context A`,
		},
		{
			name: "three level unwrap chain",
			err: fmt.Errorf(
				"a: %w",
				fmt.Errorf(
					"b: %w",
					fmt.Errorf(
						"c: %w",
						errors.New("d"),
					),
				),
			),
			expected: `
a:
  b:
    c:
      d`,
		},
		{
			name: "multi-line/parent error",
			err: fmt.Errorf(
				"outer\nline2: %w",
				errors.New("inner"),
			),
			expected: `
outer
line2:
  inner`,
		},
		{
			name: "multi-line/child error",
			err: fmt.Errorf(
				"outer: %w",
				errors.New("inner\nline2"),
			),
			expected: `
outer:
  inner
  line2`,
		},
		{
			name: "multi-line/parent and child errors",
			err: fmt.Errorf(
				"outer\nlineB: %w",
				errors.New("inner\nlineC"),
			),
			expected: `
outer
lineB:
  inner
  lineC`,
		},
		{
			name: "multi-line/wrapped error",
			err: fmt.Errorf(
				"%w: context line1\ncontext line2",
				errors.New("invalid grant"),
			),
			expected: `
invalid grant: context line1
context line2`,
		},
		{
			name:     "whitespace/trailing newline trimmed",
			err:      errors.New("error\n"),
			expected: "error",
		},
		{
			name: "whitespace/inner trailing newline trimmed",
			err: fmt.Errorf(
				"outer: %w",
				errors.New("inner\n"),
			),
			expected: `
outer:
  inner`,
		},
		{
			name: "joined errors/simple",
			err: errors.Join(
				errors.New("first"),
				errors.New("second"),
			),
			expected: `
first
second`,
		},
		{
			name: "joined errors/single wrap",
			err: errors.Join(
				fmt.Errorf(
					"first: %w",
					errors.New("child"),
				),
				fmt.Errorf(
					"second: %w",
					errors.New("child"),
				),
			),
			expected: `
first:
  child
second:
  child`,
		},
		{
			name: "joined errors/multiple wraps and joins",
			err: errors.Join(
				errors.New("hello"),
				fmt.Errorf(
					"a: %w",
					errors.New("1"),
				),
				fmt.Errorf(
					"b: %w",
					errors.Join(
						errors.New("b.1"),
						errors.New("b.2"),
						fmt.Errorf(
							"what: %w",
							errors.New("bye"),
						),
					),
				),
			),
			expected: `
hello
a:
  1
b:
  b.1
  b.2
  what:
    bye`,
		},
		{
			name: "joined errors/wrapped errors at beginning",
			err: errors.Join(
				fmt.Errorf(
					"%w: first context",
					errors.New("first"),
				),
				fmt.Errorf(
					"%w trailing context",
					errors.New("second"),
				),
			),
			expected: `
first: first context
second trailing context`,
		},
		{
			name: "mixed prefix and suffix wrapping",
			err: fmt.Errorf(
				"outer: %w",
				fmt.Errorf(
					"%w: context",
					errors.New("inner"),
				),
			),
			expected: `
outer:
  inner: context`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FormatError(tc.err)

			tc.expected = strings.TrimPrefix(tc.expected, "\n")
			if got != tc.expected {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.expected)
			}
		})
	}
}

func TestAppendFormattedErrorLines(t *testing.T) {
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
			err:  errors.New("error"),
			want: []string{"error"},
		},
		{
			name: "multiple errors",
			err: errors.Join(
				errors.New("first"),
				errors.New("second"),
			),
			want: []string{
				"first",
				"second",
			},
		},
		{
			name: "wrapped error/simple",
			err: fmt.Errorf(
				"key: %w",
				errors.New("error"),
			),
			want: []string{
				"key:",
				"  error",
			},
		},
		{
			name: "wrapped error/custom",
			err: &decode.ErrKeyField{
				Key: "key",
				Err: errors.New("error"),
			},
			want: []string{
				"key:",
				"  error",
			},
		}, {
			name: "wrapped error/at beginning",
			err: fmt.Errorf(
				"%w: context",
				errors.New("error"),
			),
			want: []string{
				"error: context",
			},
		},
		{
			name: "wrapped error/at beginning with multiline context",
			err: fmt.Errorf(
				"%w: context line1\ncontext line2",
				errors.New("error"),
			),
			want: []string{
				"error: context line1\ncontext line2",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lines := appendFormattedErrorLines(tc.err, nil, 0)

			if got, want := len(lines), len(tc.want); got != want {
				t.Fatalf(
					"len(lines) mismatch\ngot:  %d (%q)\nwant: %d (%q)",
					got, lines,
					want, tc.want,
				)
			}

			for i, wantLine := range tc.want {
				if lines[i] != wantLine {
					t.Fatalf(
						"lines[%d] mismatch\ngot:  %q (%q)\nwant: %q (%q)",
						i,
						lines[i], lines,
						wantLine, tc.want,
					)
				}
			}
		})
	}
}
