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

func TestErrInvalidType_Error(t *testing.T) {
	// GIVEN: an ErrInvalidType.
	tests := []struct {
		name     string
		err      *ErrInvalidType
		expected string
	}{
		{
			name: "value provided with multiple allowed types",
			err: &ErrInvalidType{
				Key:     "type",
				Value:   "mysql",
				Allowed: []string{"postgres", "sqlite", "mysql"},
			},
			expected: `type: "mysql" <invalid> (supported values = ['postgres', 'sqlite', 'mysql'])`,
		},
		{
			name: "empty value uses required placeholder",
			err: &ErrInvalidType{
				Key:     "type",
				Value:   "",
				Allowed: []string{"postgres", "sqlite"},
			},
			expected: `type: <required> (supported values = ['postgres', 'sqlite'])`,
		},
		{
			name: "single allowed type",
			err: &ErrInvalidType{
				Key:     "type",
				Value:   "redis",
				Allowed: []string{"redis"},
			},
			expected: `type: "redis" <invalid> (supported values = ['redis'])`,
		},
		{
			name: "multiple allowed types preserve order",
			err: &ErrInvalidType{
				Key:     "type",
				Value:   "mongo",
				Allowed: []string{"mongo", "cassandra", "dynamodb"},
			},
			expected: `type: "mongo" <invalid> (supported values = ['mongo', 'cassandra', 'dynamodb'])`,
		},
		{
			name: "empty allowed list",
			err: &ErrInvalidType{
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
					"%s\nErrInvalidType stringified mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.expected,
				)
			}
		})
	}
}

func TestErrInvalidValues_Error(t *testing.T) {
	// GIVEN: an ErrInvalidValues.
	tests := []struct {
		name     string
		err      *ErrInvalidValues
		expected string
	}{
		{
			name: "single value",
			err: &ErrInvalidValues{
				Key:     "hide",
				Values:  []string{"mysql"},
				Allowed: []string{"postgres", "sqlite", "mysql"},
			},
			expected: `hide: "mysql" <invalid> (supported values = ['postgres', 'sqlite', 'mysql'])`,
		},
		{
			name: "values are named in the order given",
			err: &ErrInvalidValues{
				Key:     "hide",
				Values:  []string{"mongo", "cassandra"},
				Allowed: []string{"postgres", "sqlite"},
			},
			expected: `hide: "mongo", "cassandra" <invalid> (supported values = ['postgres', 'sqlite'])`,
		},
		{
			name: "at the cap, every value is named",
			err: &ErrInvalidValues{
				Key:     "hide",
				Values:  []string{"a", "b", "c", "d", "e"},
				Allowed: []string{"postgres"},
			},
			expected: `hide: "a", "b", "c", "d", "e" <invalid> (supported values = ['postgres'])`,
		},
		{
			name: "over the cap, the rest are counted",
			err: &ErrInvalidValues{
				Key:     "hide",
				Values:  []string{"a", "b", "c", "d", "e", "f", "g"},
				Allowed: []string{"postgres"},
			},
			expected: `hide: "a", "b", "c", "d", "e", and 2 more <invalid> (supported values = ['postgres'])`,
		},
		{
			name: "one over the cap",
			err: &ErrInvalidValues{
				Key:     "hide",
				Values:  []string{"a", "b", "c", "d", "e", "f"},
				Allowed: []string{"postgres"},
			},
			expected: `hide: "a", "b", "c", "d", "e", and 1 more <invalid> (supported values = ['postgres'])`,
		},
		{
			name: "empty value stays visible",
			err: &ErrInvalidValues{
				Key:     "hide",
				Values:  []string{""},
				Allowed: []string{"postgres"},
			},
			expected: `hide: "" <invalid> (supported values = ['postgres'])`,
		},
		{
			name: "no values uses 'required' placeholder",
			err: &ErrInvalidValues{
				Key:     "hide",
				Values:  nil,
				Allowed: []string{"postgres", "sqlite"},
			},
			expected: `hide: <required> (supported values = ['postgres', 'sqlite'])`,
		},
		{
			name: "empty allowed list",
			err: &ErrInvalidValues{
				Key:     "hide",
				Values:  []string{"unknown"},
				Allowed: []string{},
			},
			expected: `hide: "unknown" <invalid> (supported values = [''])`,
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
					"%s\nErrInvalidValues stringified mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.expected,
				)
			}
		})
	}
}
