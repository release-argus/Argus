// Copyright [2024] [Argus]
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

package util

import (
	"fmt"
	"testing"
)

func TestAppendCheckError(t *testing.T) {
	tests := map[string]struct {
		prefix, label string
		checkErr      error
		errs, want    []error
	}{
		"nil checkErr": {
			errs:     []error{},
			prefix:   "prefix",
			label:    "label",
			checkErr: nil,
			want:     []error{},
		},
		"non-nil checkErr": {
			errs:     []error{},
			prefix:   "prefix_",
			label:    "label",
			checkErr: fmt.Errorf("an error occurred"),
			want: []error{
				fmt.Errorf("%slabel:\nan error occurred",
					"prefix_"),
			},
		},
		"existing errors with non-nil checkErr": {
			errs: []error{
				fmt.Errorf("existing error"),
			},
			prefix:   "prefix_",
			label:    "label",
			checkErr: fmt.Errorf("an error occurred"),
			want: []error{
				fmt.Errorf("existing error"),
				fmt.Errorf("%slabel:\nan error occurred",
					"prefix_"),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			errs := tc.errs
			AppendCheckError(&errs, tc.prefix, tc.label, tc.checkErr)
			if len(errs) != len(tc.want) {
				t.Fatalf("expected %d errors, got %d\n%q",
					len(tc.want), len(errs), errs)
			}
			for i, err := range errs {
				if err.Error() != tc.want[i].Error() {
					t.Errorf("error mismatch\n%q\ngot:\n%q",
						tc.want[i], err)
				}
			}
		})
	}
}

func TestErrorToString(t *testing.T) {
	// GIVEN a bunch of comparables
	tests := map[string]struct {
		err  error
		want string
	}{
		"nil error": {
			err: nil, want: ""},
		"non-nil error": {
			err: fmt.Errorf("test error"), want: "test error"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ErrorToString is called
			got := ErrorToString(tc.err)

			// THEN the var is printed when it should be
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}
