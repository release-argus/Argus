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

package util

import (
	"fmt"
	"testing"
)

func TestRandAlphaNumericLower(t *testing.T) {
	// GIVEN: different size strings are wanted.
	tests := []struct {
		name   string
		wanted int
	}{
		{
			name:   "length 1",
			wanted: 1,
		},
		{
			name:   "length 2",
			wanted: 2,
		},
		{
			name:   "length 3",
			wanted: 3,
		},
		{
			name:   "length 10",
			wanted: 10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: RandAlphaNumericLower is called.
			got := RandAlphaNumericLower(tc.wanted)

			prefix := fmt.Sprintf(
				"%s\nRandAlphaNumericLower(%d)",
				packageName, tc.wanted,
			)

			// THEN: we get a random string of the specified length.
			if gotLen := len(got); gotLen != tc.wanted {
				t.Errorf(
					"%s length mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, tc.wanted,
				)
			}

			// AND: it contains only lowercase alphanumeric characters.
			for _, v := range got {
				valid := false
				for _, char := range alphanumericLower {
					if v == char {
						valid = true
						break
					}
				}
				if !valid {
					t.Errorf(
						"%s contains non-alphanumeric character %q\ngot:  %q",
						prefix, string(v), got,
					)
				}
			}
		})
	}
}

func TestRandNumeric(t *testing.T) {
	// GIVEN: different size strings are wanted.
	tests := []struct {
		name   string
		wanted int
	}{
		{
			name:   "length 1",
			wanted: 1,
		},
		{
			name:   "length 2",
			wanted: 2,
		},
		{
			name:   "length 3",
			wanted: 3,
		},
		{
			name:   "length 10",
			wanted: 10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: RandNumeric is called.
			got := RandNumeric(tc.wanted)

			prefix := fmt.Sprintf(
				"%s\nRandNumeric(%d)",
				packageName, tc.wanted,
			)

			// THEN: we get a random numeric string of the specified length.
			if gotLen := len(got); gotLen != tc.wanted {
				t.Errorf(
					"%s length mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, tc.wanted,
				)
			}

			// AND: it contains only numeric characters.
			for _, v := range got {
				valid := false
				for _, char := range numeric {
					if v == char {
						valid = true
						break
					}
				}
				if !valid {
					t.Errorf(
						"%s contains non-numeric character %q\ngot:  %q",
						prefix, string(v), got,
					)
				}
			}
		})
	}
}

func TestRandString(t *testing.T) {
	// GIVEN: different size strings are wanted with different alphabets.
	tests := []struct {
		name     string
		wanted   int
		alphabet string
	}{
		{
			name:     "length 1 string, length 1 alphabet",
			wanted:   1,
			alphabet: "a",
		},
		{
			name:     "length 2, length 1 alphabet",
			wanted:   2,
			alphabet: "b",
		},
		{
			name:     "length 3, length 1 alphabet",
			wanted:   3,
			alphabet: "c",
		},
		{
			name:     "length 10, length 1 alphabet",
			wanted:   10,
			alphabet: "d",
		},
		{
			name:     "length 10, length 5 alphabet",
			wanted:   10,
			alphabet: "abcde",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: RandString is called.
			result := RandString(tc.wanted, tc.alphabet)

			prefix := fmt.Sprintf(
				"%s\nRandString(length=%d, alphabet=%q)",
				packageName, tc.wanted, tc.alphabet,
			)

			// THEN: we get a random alphabet string of the specified length.
			if got := len(result); got != tc.wanted {
				t.Errorf(
					"%s length mismatch\ngot:  %d\nwant: %d",
					prefix, got, tc.wanted,
				)
			}

			// AND: it contains only alphabet characters.
			for _, v := range result {
				valid := false
				for _, char := range tc.alphabet {
					if v == char {
						valid = true
						break
					}
				}
				if !valid {
					t.Errorf(
						"%s contains non-alphabet character %q\ngot:  %q",
						prefix, string(v), result,
					)
				}
			}
		})
	}
}
