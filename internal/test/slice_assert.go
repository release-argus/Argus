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
	"strings"
	"testing"
)

// AssertSlicesEqualFunc compares two slices element-wise with eq, distinguishing nil from empty.
func AssertSlicesEqualFunc[T any, U any](
	t *testing.T,
	got []T,
	want []U,
	eq func(T, U) bool,
	target string,
) error {
	t.Helper()

	if (got == nil) != (want == nil) {
		msg := "%s nil mismatch\ngot:  %s\nwant: %s"
		args := []any{target, formatValue(got), formatValue(want)}
		if target == "" {
			msg = strings.Replace(msg, "%s ", "", 1)
			args = args[1:]
		}
		return fmt.Errorf(msg, args...)
	}

	if gotLen, wantLen := len(got), len(want); gotLen != wantLen {
		msg := "%s length mismatch\ngot:  %d (%s)\nwant: %d (%s)"
		args := []any{target, gotLen, formatValue(got), wantLen, formatValue(want)}
		if target == "" {
			msg = strings.Replace(msg, "%s ", "", 1)
			args = args[1:]
		}
		return fmt.Errorf(msg, args...)
	}

	var errs []error
	for i := range want {
		if !eq(got[i], want[i]) {
			errs = append(errs, fmt.Errorf(
				"%s[%d] element mismatch\ngot:  %s (%s)\nwant: %s (%s)",
				target, i,
				formatValue(got[i]), formatValue(got),
				formatValue(want[i]), formatValue(want),
			))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// AssertSlicesEqual compares two slices of comparable elements, distinguishing
// unset (nil) from an explicitly empty slice.
// target names the compared field in any mismatch, e.g. "hide".
func AssertSlicesEqual[T comparable](
	t *testing.T,
	got, want []T,
	target string,
) error {
	t.Helper()

	return AssertSlicesEqualFunc(
		t,
		got,
		want,
		func(a, b T) bool { return a == b },
		target,
	)
}
