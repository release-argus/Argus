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

import "testing"

func TestFieldError_Error(t *testing.T) {
	// GIVEN: a FieldError.
	tests := []struct {
		name     string
		err      FieldError
		expected string
	}{
		{
			name: "key",
			err: FieldError{
				Key: "testKey",
			},
			expected: `testKey: <required>`,
		},
		{
			name: "key + description",
			err: FieldError{
				Key:         "testKey",
				Description: "must be set",
			},
			expected: `testKey: <required> (must be set)`,
		},
		{
			name: "key-value",
			err: FieldError{
				Key:   "testKey",
				Value: "foo",
			},
			expected: `testKey: "foo" <invalid>`,
		},
		{
			name: "key-value + description",
			err: FieldError{
				Key:         "type",
				Value:       "foo",
				Description: "should be X",
			},
			expected: `type: "foo" <invalid> (should be X)`,
		},
		{
			name: "key-value + description + allowed values",
			err: FieldError{
				Key:         "type",
				Value:       "foo",
				Description: "description goes here",
			},
			expected: `type: "foo" <invalid> (description goes here)`,
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
					"%s\nstringified FieldError mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.expected,
				)
			}
		})
	}
}
