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
	"errors"
	"testing"
)

func TestErrKeyField_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrKeyField
		expected string
	}{
		{
			name: "key, no error wrapped",
			err: ErrKeyField{
				Key: "testKey",
			},
			expected: `testKey: <nil>`,
		},
		{
			name: "key + error wrapped",
			err: ErrKeyField{
				Key: "testKey",
				Err: errors.New("foo"),
			},
			expected: `testKey: foo`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.err.Error()

			if got != tc.expected {
				t.Fatalf("ErrKeyField.Error() mismatch\ngot:  %q\nwant: %q", got, tc.expected)
			}
		})
	}
}

func TestErrKeyField_Unwrap(t *testing.T) {
	wrapped := errors.New("foo")

	tests := []struct {
		name     string
		err      ErrKeyField
		expected error
	}{
		{
			name: "no error wrapped",
			err: ErrKeyField{
				Key: "testKey",
			},
			expected: nil,
		},
		{
			name: "error wrapped",
			err: ErrKeyField{
				Key: "testKey",
				Err: wrapped,
			},
			expected: wrapped,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.err.Unwrap()

			if got != tc.expected {
				t.Fatalf("unwrapped error mismatch\ngot:  %v\nwant: %v", got, tc.expected)
			}
		})
	}
}
