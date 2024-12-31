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

import "testing"

func TestRandAlphaNumericLower(t *testing.T) {
	// GIVEN different size strings are wanted
	tests := map[string]struct {
		wanted int
	}{
		"length 1": {
			wanted: 1},
		"length 2": {
			wanted: 2},
		"length 3": {
			wanted: 3},
		"length 10": {
			wanted: 10},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN RandAlphaNumericLower is called
			got := RandAlphaNumericLower(tc.wanted)

			// THEN we get a random alphanumeric string of the specified length
			if len(got) != tc.wanted {
				t.Errorf("got length %d. wanted %d",
					tc.wanted, len(got))
			}
			charactersVerified := 0
			for charactersVerified != tc.wanted {
				var characters []string
				for i := range alphanumericLower {
					characters = append(characters, string(alphanumericLower[i]))
				}

				for i := range characters {
					if got == characters[i] {
						RemoveIndex(&characters, i)
						break
					}
				}
				charactersVerified++
			}
		})
	}
}

func TestRandNumeric(t *testing.T) {
	// GIVEN different size strings are wanted
	tests := map[string]struct {
		wanted int
	}{
		"length 1": {
			wanted: 1},
		"length 2": {
			wanted: 2},
		"length 3": {
			wanted: 3},
		"length 10": {
			wanted: 10},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN RandNumeric is called
			got := RandNumeric(tc.wanted)

			// THEN we get a random numeric string of the specified length
			if len(got) != tc.wanted {
				t.Errorf("got length %d. wanted %d",
					tc.wanted, len(got))
			}
			charactersVerified := 0
			for charactersVerified != tc.wanted {
				var characters []string
				for i := range numeric {
					characters = append(characters, string(numeric[i]))
				}

				for i := range characters {
					if got == characters[i] {
						RemoveIndex(&characters, i)
						break
					}
				}
				charactersVerified++
			}
		})
	}
}

func TestRandString(t *testing.T) {
	// GIVEN different size strings are wanted with different alphabets
	tests := map[string]struct {
		wanted   int
		alphabet string
	}{
		"length 1 string, length 1 alphabet": {
			wanted: 1, alphabet: "a"},
		"length 2, length 1 alphabet": {
			wanted: 2, alphabet: "b"},
		"length 3, length 1 alphabet": {
			wanted: 3, alphabet: "c"},
		"length 10, length 1 alphabet": {
			wanted: 10, alphabet: "d"},
		"length 10, length 5 alphabet": {
			wanted: 10, alphabet: "abcde"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN RandString is called
			got := RandString(tc.wanted, tc.alphabet)

			// THEN we get a random alphabet string of the specified length
			if len(got) != tc.wanted {
				t.Errorf("got length %d. wanted %d",
					tc.wanted, len(got))
			}
			charactersVerified := 0
			for charactersVerified != tc.wanted {
				var characters []string
				for i := range tc.alphabet {
					characters = append(characters, string(tc.alphabet[i]))
				}

				for i := range characters {
					if got == characters[i] {
						RemoveIndex(&characters, i)
						break
					}
				}
				charactersVerified++
			}
		})
	}
}
