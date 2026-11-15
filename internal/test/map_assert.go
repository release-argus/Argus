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
	"maps"
	"slices"
	"strings"
	"testing"
)

// AssertMapEqualFunc compares two maps key by key with eq.
// target names the compared field from its root receiver type, e.g. "Lookup.Headers".
func AssertMapEqualFunc[T any, U any](
	t *testing.T,
	gotMap map[string]T,
	wantMap map[string]U,
	eq func(T, U) bool,
	target string,
) error {
	t.Helper()

	if gotLen, wantLen := len(gotMap), len(wantMap); gotLen != wantLen {
		msg := "%s length mismatch\ngot:  %d (%s)\nwant: %d (%s)"
		args := []any{target, gotLen, formatValue(gotMap), wantLen, formatValue(wantMap)}
		if target == "" {
			msg = strings.Replace(msg, "%s ", "", 1)
			args = args[1:]
		}
		return fmt.Errorf(msg, args...)
	}

	var errs []error
	for _, k := range slices.Sorted(maps.Keys(wantMap)) {
		want := wantMap[k]
		got, ok := gotMap[k]
		if !ok {
			errs = append(errs, fmt.Errorf(
				"%s[%q] mismatch\ngot:  absent (%s)\nwant: %s",
				target, k, formatValue(gotMap), formatValue(want),
			))
			continue
		}
		if !eq(got, want) {
			errs = append(errs, fmt.Errorf(
				"%s[%q] mismatch\ngot:  %s\nwant: %s",
				target, k, formatValue(got), formatValue(want),
			))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// AssertMapEqual compares two maps of comparable values.
// target names the compared field from its root receiver type, e.g. "Lookup.Headers".
func AssertMapEqual[T comparable](
	t *testing.T,
	gotMap, wantMap map[string]T,
	target string,
) error {
	t.Helper()

	return AssertMapEqualFunc(
		t,
		gotMap,
		wantMap,
		func(a, b T) bool { return a == b },
		target,
	)
}
