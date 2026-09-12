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
	"regexp"
	"testing"

	"github.com/release-argus/Argus/util/errfmt"
)

func TestAssertSlicesEqualFunc(t *testing.T) {
	prefix := t.Name()
	targetToUse := "target"
	// GIVEN: two slices to compare.
	tests := []struct {
		name                             string
		a, b                             []any
		eq                               func(any, any) bool
		want                             bool
		noTargetErrRegex, targetErrRegex string
	}{
		{
			name: "equal",
			a:    []any{"foo", "bar"},
			b:    []any{"foo", "bar"},
			want: true,
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
			noTargetErrRegex: fmt.Sprintf("%s nil mismatch", prefix),
			targetErrRegex:   fmt.Sprintf("%s %s nil mismatch", prefix, targetToUse),
		},
		{
			name:             "explicitly empty is not unset",
			a:                []any{},
			b:                nil,
			eq:               func(a, b any) bool { return a == b },
			noTargetErrRegex: fmt.Sprintf("%s nil mismatch", prefix),
			targetErrRegex:   fmt.Sprintf("%s %s nil mismatch", prefix, targetToUse),
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
			noTargetErrRegex: fmt.Sprintf("%s length mismatch", prefix),
			targetErrRegex:   fmt.Sprintf("%s %s length mismatch", prefix, targetToUse),
		},
		{
			name: "element mismatch",
			a:    []any{"foo", "bar"},
			b:    []any{"foo", "baz"},
			eq: func(a, b any) bool {
				return a == b
			},
			noTargetErrRegex: fmt.Sprintf(`%s\[1\] element mismatch`, prefix),
			targetErrRegex:   fmt.Sprintf(`%s %s\[1\] element mismatch`, prefix, targetToUse),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := ""

			// WHEN: AssertSlicesEqualFunc is called with no target
			err := AssertSlicesEqualFunc(t, tc.a, tc.b, tc.eq, prefix, target)

			// THEN: The error is as expected.
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.noTargetErrRegex).MatchString(e) {
				t.Errorf(
					"%s\nAssertSlicesEqualFunc(a=%v, b=%v, target=%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.a, tc.b, target,
					e, tc.noTargetErrRegex,
				)
			}

			// GIVEN: a target
			target = targetToUse

			// WHEN: AssertSlicesEqualFunc is called with a target
			err = AssertSlicesEqualFunc(t, tc.a, tc.b, tc.eq, "TestAssertSlicesEqualFunc", target)

			// THEN: The error is as expected.
			e = errfmt.FormatError(err)
			if !regexp.MustCompile(tc.targetErrRegex).MatchString(e) {
				t.Errorf(
					"%s\nAssertSlicesEqualFunc(a=%v, b=%v, target=%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.a, tc.b, target,
					e, tc.targetErrRegex,
				)
			}
		})
	}
}

func TestAssertSlicesEqual(t *testing.T) {
	prefix := t.Name()
	targetToUse := "target"
	// GIVEN: two slices of comparable elements to compare.
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
			noTargetErrRegex: fmt.Sprintf("%s nil mismatch", prefix),
			targetErrRegex:   fmt.Sprintf("%s %s nil mismatch", prefix, targetToUse),
		},
		{
			name:             "empty != nil (unset)",
			a:                []string{},
			noTargetErrRegex: fmt.Sprintf("%s nil mismatch", prefix),
			targetErrRegex:   fmt.Sprintf("%s %s nil mismatch", prefix, targetToUse),
		},
		{
			name:             "length mismatch",
			a:                []string{"foo"},
			b:                []string{"foo", "bar"},
			noTargetErrRegex: fmt.Sprintf("%s length mismatch", prefix),
			targetErrRegex:   fmt.Sprintf("%s %s length mismatch", prefix, targetToUse),
		},
		{
			name:             "element mismatch",
			a:                []string{"foo", "bar"},
			b:                []string{"foo", "baz"},
			noTargetErrRegex: fmt.Sprintf(`%s\[1\] element mismatch`, prefix),
			targetErrRegex:   fmt.Sprintf(`%s %s\[1\] element mismatch`, prefix, targetToUse),
		},
		{
			name:             "order mismatch",
			a:                []string{"foo", "bar"},
			b:                []string{"bar", "foo"},
			noTargetErrRegex: fmt.Sprintf(`%s\[0\] element mismatch`, prefix),
			targetErrRegex:   fmt.Sprintf(`%s %s\[0\] element mismatch`, prefix, targetToUse),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: no target.
			target := ""

			// WHEN: AssertSlicesEqual is called with no target.
			err := AssertSlicesEqual(t, tc.a, tc.b, prefix, target)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.noTargetErrRegex).MatchString(e) {
				t.Errorf(
					"%s\nAssertSlicesEqual(a=%v, b=%v, target=%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.a, tc.b, target,
					e, tc.noTargetErrRegex,
				)
			}

			// GIVEN: a target.
			target = targetToUse

			// WHEN: AssertSlicesEqual is called with a target.
			err = AssertSlicesEqual(t, tc.a, tc.b, prefix, target)

			// THEN: the error names it.
			e = errfmt.FormatError(err)
			if !regexp.MustCompile(tc.targetErrRegex).MatchString(e) {
				t.Errorf(
					"%s\nAssertSlicesEqual(a=%v, b=%v, target=%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.a, tc.b, target,
					e, tc.targetErrRegex,
				)
			}
		})
	}
}

func TestAssertSlicesEqual__otherComparableTypes(t *testing.T) {
	prefix := t.Name()
	// GIVEN: slices of types other than string.
	tests := []struct {
		name     string
		invoke   func(t *testing.T) error
		errRegex string
	}{
		{
			name: "ints/equal",
			invoke: func(t *testing.T) error {
				return AssertSlicesEqual(t, []int{1, 2}, []int{1, 2}, prefix, "")
			},
			errRegex: `^$`,
		},
		{
			name: "ints/mismatch",
			invoke: func(t *testing.T) error {
				return AssertSlicesEqual(t, []int{1, 2}, []int{1, 3}, prefix, "")
			},
			errRegex: `\[1\] element mismatch`,
		},
		{
			name: "bools/mismatch",
			invoke: func(t *testing.T) error {
				return AssertSlicesEqual(t, []bool{true}, []bool{false}, prefix, "")
			},
			errRegex: `\[0\] element mismatch`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: AssertSlicesEqual is called.
			err := tc.invoke(t)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Errorf(
					"%s\nerror mismatch\ngot:  %q\nwant: %q",
					packageName, e, tc.errRegex,
				)
			}
		})
	}
}
