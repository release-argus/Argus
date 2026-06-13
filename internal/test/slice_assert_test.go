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
	prefix := "TestAssertSlicesEqualFunc"
	targetToUse := "target"
	// GIVEN: two slices to compare.
	tests := []struct {
		name                             string
		a, b                             []interface{}
		eq                               func(interface{}, interface{}) bool
		want                             bool
		noTargetErrRegex, targetErrRegex string
	}{
		{
			name: "equal",
			a:    []interface{}{"foo", "bar"},
			b:    []interface{}{"foo", "bar"},
			want: true,
			eq: func(a, b interface{}) bool {
				return a == b
			},
			noTargetErrRegex: `^$`,
			targetErrRegex:   `^$`,
		},
		{
			name: "length mismatch",
			a:    []interface{}{"foo", "bar"},
			b:    []interface{}{"foo", "bar", "baz", "qux"},
			eq: func(a, b interface{}) bool {
				return a == b
			},
			noTargetErrRegex: fmt.Sprintf("%s length mismatch", prefix),
			targetErrRegex:   fmt.Sprintf("%s %s length mismatch", prefix, targetToUse),
		},
		{
			name: "element mismatch",
			a:    []interface{}{"foo", "bar"},
			b:    []interface{}{"foo", "baz"},
			eq: func(a, b interface{}) bool {
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
				t.Errorf("%s\nAssertSlicesEqualFunc(a=%v, b=%v, target=%q) error mismatch\ngot:  %q\nwant: %q",
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
				t.Errorf("%s\nAssertSlicesEqualFunc(a=%v, b=%v, target=%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.a, tc.b, target,
					e, tc.targetErrRegex,
				)
			}
		})
	}
}
