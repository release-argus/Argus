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
	"regexp"
	"testing"

	"github.com/release-argus/Argus/util/errfmt"
)

func TestAssertSlicesEqualFunc(t *testing.T) {
	targetToUse := "target"

	tests := []struct {
		name                             string
		a, b                             []any
		eq                               func(any, any) bool
		noTargetErrRegex, targetErrRegex string
	}{
		{
			name: "equal",
			a:    []any{"foo", "bar"},
			b:    []any{"foo", "bar"},
			eq: func(a, b any) bool {
				return a == b
			},
			noTargetErrRegex: `^$`,
			targetErrRegex:   `^$`,
		},
		{
			name:             "unset is not explicitly empty",
			a:                nil,
			b:                []any{},
			eq:               func(a, b any) bool { return a == b },
			noTargetErrRegex: `^nil mismatch\n`,
			targetErrRegex:   `^` + targetToUse + ` nil mismatch\n`,
		},
		{
			name:             "explicitly empty is not unset",
			a:                []any{},
			b:                nil,
			eq:               func(a, b any) bool { return a == b },
			noTargetErrRegex: `^nil mismatch\n`,
			targetErrRegex:   `^` + targetToUse + ` nil mismatch\n`,
		},
		{
			name:             "both unset",
			a:                nil,
			b:                nil,
			eq:               func(a, b any) bool { return a == b },
			noTargetErrRegex: `^$`,
			targetErrRegex:   `^$`,
		},
		{
			name:             "both explicitly empty",
			a:                []any{},
			b:                []any{},
			eq:               func(a, b any) bool { return a == b },
			noTargetErrRegex: `^$`,
			targetErrRegex:   `^$`,
		},
		{
			name: "length mismatch",
			a:    []any{"foo", "bar"},
			b:    []any{"foo", "bar", "baz", "qux"},
			eq: func(a, b any) bool {
				return a == b
			},
			noTargetErrRegex: `^length mismatch\n`,
			targetErrRegex:   `^` + targetToUse + ` length mismatch\n`,
		},
		{
			name: "element mismatch",
			a:    []any{"foo", "bar"},
			b:    []any{"foo", "baz"},
			eq: func(a, b any) bool {
				return a == b
			},
			noTargetErrRegex: `\[1\] element mismatch`,
			targetErrRegex:   targetToUse + `\[1\] element mismatch`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := ""

			err := AssertSlicesEqualFunc(t, tc.a, tc.b, tc.eq, target)
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.noTargetErrRegex).MatchString(e) {
				t.Errorf("no target: error mismatch\ngot:  %q\nwant: %q", e, tc.noTargetErrRegex)
			}

			target = targetToUse

			err = AssertSlicesEqualFunc(t, tc.a, tc.b, tc.eq, target)
			e = errfmt.FormatError(err)
			if !regexp.MustCompile(tc.targetErrRegex).MatchString(e) {
				t.Errorf("target: error mismatch\ngot:  %q\nwant: %q", e, tc.targetErrRegex)
			}
		})
	}
}

func TestAssertSlicesEqual(t *testing.T) {
	targetToUse := "target"

	tests := []struct {
		name                             string
		a, b                             []string
		noTargetErrRegex, targetErrRegex string
	}{
		{
			name:             "equal/both unset",
			noTargetErrRegex: `^$`,
			targetErrRegex:   `^$`,
		},
		{
			name:             "equal/both explicitly empty",
			a:                []string{},
			b:                []string{},
			noTargetErrRegex: `^$`,
			targetErrRegex:   `^$`,
		},
		{
			name:             "equal/same values",
			a:                []string{"foo", "bar"},
			b:                []string{"foo", "bar"},
			noTargetErrRegex: `^$`,
			targetErrRegex:   `^$`,
		},
		{
			name:             "nil (unset) != empty",
			b:                []string{},
			noTargetErrRegex: `^nil mismatch\ngot:  nil\nwant: \[\]$`,
			targetErrRegex:   `^` + targetToUse + ` nil mismatch\n`,
		},
		{
			name:             "empty != nil (unset)",
			a:                []string{},
			noTargetErrRegex: `^nil mismatch\n`,
			targetErrRegex:   `^` + targetToUse + ` nil mismatch\n`,
		},
		{
			name:             "length mismatch",
			a:                []string{"foo"},
			b:                []string{"foo", "bar"},
			noTargetErrRegex: `^length mismatch\n`,
			targetErrRegex:   `^` + targetToUse + ` length mismatch\n`,
		},
		{
			name:             "element mismatch",
			a:                []string{"foo", "bar"},
			b:                []string{"foo", "baz"},
			noTargetErrRegex: `^\[1\] element mismatch\ngot:  "bar" \(\["foo" "bar"\]\)\nwant: "baz" \(\["foo" "baz"\]\)$`,
			targetErrRegex:   targetToUse + `\[1\] element mismatch`,
		},
		{
			name:             "order mismatch",
			a:                []string{"foo", "bar"},
			b:                []string{"bar", "foo"},
			noTargetErrRegex: `\[0\] element mismatch`,
			targetErrRegex:   targetToUse + `\[0\] element mismatch`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := ""

			err := AssertSlicesEqual(t, tc.a, tc.b, target)
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.noTargetErrRegex).MatchString(e) {
				t.Errorf("no target: error mismatch\ngot:  %q\nwant: %q", e, tc.noTargetErrRegex)
			}

			target = targetToUse

			err = AssertSlicesEqual(t, tc.a, tc.b, target)
			e = errfmt.FormatError(err)
			if !regexp.MustCompile(tc.targetErrRegex).MatchString(e) {
				t.Errorf("target: error mismatch\ngot:  %q\nwant: %q", e, tc.targetErrRegex)
			}
		})
	}
}

func TestAssertSlicesEqual__otherComparableTypes(t *testing.T) {
	tests := []struct {
		name     string
		invoke   func(t *testing.T) error
		errRegex string
	}{
		{
			name: "ints/equal",
			invoke: func(t *testing.T) error {
				return AssertSlicesEqual(t, []int{1, 2}, []int{1, 2}, "")
			},
			errRegex: `^$`,
		},
		{
			name: "ints/mismatch",
			invoke: func(t *testing.T) error {
				return AssertSlicesEqual(t, []int{1, 2}, []int{1, 3}, "")
			},
			errRegex: `\[1\] element mismatch`,
		},
		{
			name: "bools/mismatch",
			invoke: func(t *testing.T) error {
				return AssertSlicesEqual(t, []bool{true}, []bool{false}, "")
			},
			errRegex: `\[0\] element mismatch`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.invoke(t)
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}
