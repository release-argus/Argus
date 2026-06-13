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

package polymorphic

import "testing"

func TestInvalidTypeError_Error(t *testing.T) {
	// GIVEN: an InvalidTypeError.
	tests := []struct {
		name     string
		err      *InvalidTypeError
		expected string
	}{
		{
			name: "value provided with multiple allowed types",
			err: &InvalidTypeError{
				Key:     "type",
				Value:   "mysql",
				Allowed: []string{"postgres", "sqlite", "mysql"},
			},
			expected: `type: "mysql" <invalid> (supported values = ['postgres', 'sqlite', 'mysql'])`,
		},
		{
			name: "empty value uses required placeholder",
			err: &InvalidTypeError{
				Key:     "type",
				Value:   "",
				Allowed: []string{"postgres", "sqlite"},
			},
			expected: `type: <required> (supported values = ['postgres', 'sqlite'])`,
		},
		{
			name: "single allowed type",
			err: &InvalidTypeError{
				Key:     "type",
				Value:   "redis",
				Allowed: []string{"redis"},
			},
			expected: `type: "redis" <invalid> (supported values = ['redis'])`,
		},
		{
			name: "multiple allowed types preserve order",
			err: &InvalidTypeError{
				Key:     "type",
				Value:   "mongo",
				Allowed: []string{"mongo", "cassandra", "dynamodb"},
			},
			expected: `type: "mongo" <invalid> (supported values = ['mongo', 'cassandra', 'dynamodb'])`,
		},
		{
			name: "empty allowed list",
			err: &InvalidTypeError{
				Key:     "type",
				Value:   "unknown",
				Allowed: []string{},
			},
			expected: `type: "unknown" <invalid> (supported values = [''])`,
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
					"%s\nInvalidTypeError stringified mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.expected,
				)
			}
		})
	}
}
