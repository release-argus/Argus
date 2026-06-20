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

func AssertSlicesEqualFunc[T any, U any](
	t *testing.T,
	got []T,
	want []U,
	eq func(T, U) bool,
	prefix, target string,
) error {
	t.Helper()

	// Length.
	if gotLen, wantLen := len(got), len(want); gotLen != wantLen {
		msg := "%s %s length mismatch\ngot:  %d (%+v)\nwant: %d (%+v)"
		if target == "" {
			msg = "%s%s length mismatch\ngot:  %d (%+v)\nwant: %d (%+v)"
		}

		return fmt.Errorf(
			msg,
			prefix, target,
			gotLen, got,
			wantLen, want,
		)
	}

	var errs []error
	for i := range want {
		if !eq(got[i], want[i]) {
			msg := "%s %s[%d] element mismatch\ngot:  %v (%+v)\nwant: %v (%+v)"
			if target == "" {
				msg = "%s%s[%d] element mismatch\ngot:  %v (%+v)\nwant: %v (%+v)"
			}

			errs = append(
				errs,
				fmt.Errorf(
					msg,
					prefix, target, i,
					got[i], got,
					want[i], want,
				),
			)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
