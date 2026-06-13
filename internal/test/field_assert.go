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

//go:build unit || integration

package test

import (
	"errors"
	"fmt"
	"testing"
)

type CompareMode int

const (
	CompareEqual CompareMode = iota
	CompareNotEqual
	CompareSamePointer
	CompareDifferentPointer
)

type FieldAssertion struct {
	Name, Target string
	Got          any
	Want         any
	Mode         CompareMode
	Check        func() (got, want any, ok bool)
	NotSame      bool
}

func AssertFields(
	t *testing.T,
	fieldTests []FieldAssertion,
	prefix, target string,
) error {
	t.Helper()

	var errs []error
	for _, tc := range fieldTests {
		var got, want any
		var ok bool

		if tc.Check != nil {
			got, want, ok = tc.Check()
		} else {
			got, want = tc.Got, tc.Want
			switch tc.Mode {
			case CompareNotEqual:
				ok = got != want
			case CompareSamePointer:
				ok = got == want || (isNil(want) && isNil(got))
			case CompareDifferentPointer:
				// Ignore check if want is nil.
				ok = got != want || (isNil(want) && isNil(got))
			default:
				ok = got == want
			}
		}
		if ok {
			continue
		}

		// tc.Target>target.
		if tc.Target != "" {
			target = tc.Target
		}
		if target != "" {
			var msg string
			switch tc.Mode {
			case CompareEqual:
				msg = "%s %s was not handed to the %s correctly\ngot:  %v\nwant: %v"
			case CompareNotEqual:
				msg = "%s %s was not handed to the %s correctly\ngot:  %v\nwant: NOT %v"
			case CompareSamePointer:
				msg = "%s %s was not handed to the %s correctly\ngot:  %p\nwant: %p"
			case CompareDifferentPointer:
				msg = "%s %s was not handed to the %s correctly\ngot:  %p\nwant: NOT %v"
			}

			errs = append(
				errs, fmt.Errorf(
					msg,
					prefix, tc.Name,
					target, got, want,
				),
			)
		} else {
			var msg string
			switch tc.Mode {
			case CompareEqual:
				msg = "%s %s mismatch\ngot:  %v\nwant: %v"
			case CompareNotEqual:
				msg = "%s %s mismatch\ngot:  %v\nwant: NOT %v"
			case CompareSamePointer:
				msg = "%s %s mismatch\ngot:  %p\nwant: %p"
			case CompareDifferentPointer:
				msg = "%s %s mismatch\ngot:  %p\nwant: NOT %p"
			}

			errs = append(
				errs, fmt.Errorf(
					msg,
					prefix, tc.Name,
					got, want,
				),
			)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
