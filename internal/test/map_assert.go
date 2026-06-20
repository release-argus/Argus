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

func AssertMapEqual[T comparable](
	t *testing.T,
	gotMap, wantMap map[string]T,
	prefix, target string,
) error {
	t.Helper()

	if gotLen, wantLen := len(gotMap), len(wantMap); len(gotMap) != len(wantMap) {
		return fmt.Errorf(
			"%s %s length mismatch\ngot:  %d (%+v)\nwant: %d (%+v)",
			prefix, target,
			gotLen, gotMap,
			wantLen, wantMap,
		)
	}

	var errs []error
	for k, want := range wantMap {
		got, ok := gotMap[k]
		if !ok {
			errs = append(
				errs,
				fmt.Errorf(
					"%s %s[%q] missing\ngot:  %+v\nwant: %+v)",
					prefix, target, k,
					gotMap, wantMap,
				),
			)
			continue
		}
		if got != want {
			errs = append(
				errs,
				fmt.Errorf(
					"%s %s[%q] mismatch\ngot:  %v\nwant: %v",
					prefix, target, k,
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
