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

package decode

import (
	"fmt"
	"testing"
)

func TestKeyFieldError_Error(t *testing.T) {
	// GIVEN: a KeyFieldError.
	tests := []struct {
		name     string
		err      KeyFieldError
		expected string
	}{
		{
			name: "key, no error wrapped",
			err: KeyFieldError{
				Key: "testKey",
			},
			expected: `testKey: <nil>`,
		},
		{
			name: "key + error wrapped",
			err: KeyFieldError{
				Key: "testKey",
				Err: fmt.Errorf("foo"),
			},
			expected: `testKey: foo`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the error is stringified.
			got := tc.err.Error()

			// THEN: the error is formatted as expected.
			if got != tc.expected {
				t.Fatalf(
					"%s\nstringified KeyFieldError mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.expected,
				)
			}
		})
	}
}

func TestKeyFieldError_Unwrap(t *testing.T) {
	// GIVEN: a KeyFieldError.
	tests := []struct {
		name     string
		err      KeyFieldError
		expected error
	}{
		{
			name: "no error wrapped",
			err: KeyFieldError{
				Key: "testKey",
			},
			expected: nil,
		},
		{
			name: "error wrapped",
			err: KeyFieldError{
				Key: "testKey",
				Err: fmt.Errorf("foo"),
			},
			expected: fmt.Errorf("foo"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the error is unwrapped.
			got := tc.err.Unwrap()

			// THEN: the error is unwrapped as expected.
			if (got != nil && tc.expected != nil) &&
				got.Error() != tc.expected.Error() {
				t.Fatalf(
					"%s\nKeyFieldError unwrapped error mismatch\ngot:  %q\nwant:  %q",
					packageName, got, tc.expected,
				)
			}
		})
	}
}
