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

func TestAssertFields(t *testing.T) {
	val := "test"
	ptr := &val
	val2 := "test2"
	ptr2 := &val2
	// GIVEN: a slice of FieldAssertion's.
	tests := []struct {
		name           string
		fieldTest      FieldAssertion
		errRegex       string
		targetOverride *string
	}{
		{
			name:      "CompareEqual success",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareEqual},
			errRegex:  `^$`,
		},
		{
			name:      "CompareSamePointer success",
			fieldTest: FieldAssertion{Name: "NAME", Got: ptr, Want: ptr, Mode: CompareSamePointer},
			errRegex:  `^$`,
		},
		{
			name:      "CompareDifferentPointer success",
			fieldTest: FieldAssertion{Name: "NAME", Got: ptr, Want: ptr2, Mode: CompareDifferentPointer},
			errRegex:  `^$`,
		},
		{
			name:      "CompareNotEqual success",
			fieldTest: FieldAssertion{Name: "NAME", Got: "abc", Want: "123", Mode: CompareNotEqual},
			errRegex:  `^$`,
		},
		{
			name:      "CompareEqual mismatch",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val2, Mode: CompareEqual},
			errRegex:  `NAME was not handed to the target correctly\ngot:  .*\nwant: .*`,
		},
		{
			name:      "CompareSamePointer mismatch",
			fieldTest: FieldAssertion{Name: "NAME", Got: ptr, Want: ptr2, Mode: CompareSamePointer},
			errRegex:  `NAME was not handed to the target correctly\ngot:  .*\nwant: .*`,
		},
		{
			name:      "CompareDifferentPointer mismatch",
			fieldTest: FieldAssertion{Name: "NAME", Got: ptr, Want: ptr, Mode: CompareDifferentPointer},
			errRegex:  `NAME was not handed to the target correctly\ngot:  .*\nwant: .*`,
		},
		{
			name:      "CompareNotEqual mismatch",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareNotEqual},
			errRegex:  `NAME was not handed to the target correctly\ngot:  .*\nwant: .*`,
		},
		{
			name: ".Check() success",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareEqual, Check: func() (any, any, bool) {
				return val, val, true
			}},
			errRegex: `^$`,
		},
		{
			name: ".Check() mismatch",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareEqual, Check: func() (any, any, bool) {
				return val, val2, false
			}},
			errRegex: `NAME was not handed to the target correctly\ngot:  .*\nwant: .*`,
		},
		{
			name: "no target param",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareEqual, Check: func() (any, any, bool) {
				return val, val2, false
			}},
			errRegex:       `NAME mismatch\ngot:  .*\nwant: .*`,
			targetOverride: Ptr(""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := "target"
			if tc.targetOverride != nil {
				target = *tc.targetOverride
			}
			fieldTests := []FieldAssertion{tc.fieldTest}

			// WHEN: AssertFields is called.
			err := AssertFields(
				t,
				fieldTests,
				"TestAssertFields",
				target,
			)

			// THEN: The error is as expected.
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Fatalf("%s\nAssertFields() error mismatch\ngot:  %q\nwant: %q",
					packageName, e, tc.errRegex,
				)
			}
		})
	}
}
