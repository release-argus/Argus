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
	"strings"
	"testing"
)

func TestRandAlphaNumericLower(t *testing.T) {
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

			got := RandAlphaNumericLower(tc.wanted)

			if gotLen := len(got); gotLen != tc.wanted {
				t.Errorf("length mismatch\ngot:  %d\nwant: %d", gotLen, tc.wanted)
			}

			for _, v := range got {
				if !strings.ContainsRune(alphanumericLower, v) {
					t.Errorf(
						"character %q mismatch\ngot:  %q\nwant: only characters in %q",
						v, got, alphanumericLower,
					)
				}
			}
		})
	}
}

func TestRandNumeric(t *testing.T) {
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

			got := RandNumeric(tc.wanted)

			if gotLen := len(got); gotLen != tc.wanted {
				t.Errorf("length mismatch\ngot:  %d\nwant: %d", gotLen, tc.wanted)
			}

			for _, v := range got {
				if !strings.ContainsRune(numeric, v) {
					t.Errorf(
						"character %q mismatch\ngot:  %q\nwant: only characters in %q",
						v, got, numeric,
					)
				}
			}
		})
	}
}

func TestRandString(t *testing.T) {
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

			result := RandString(tc.wanted, tc.alphabet)

			if got := len(result); got != tc.wanted {
				t.Errorf("length mismatch\ngot:  %d\nwant: %d", got, tc.wanted)
			}

			for _, v := range result {
				if !strings.ContainsRune(tc.alphabet, v) {
					t.Errorf(
						"character %q mismatch\ngot:  %q\nwant: only characters in %q",
						v, result, tc.alphabet,
					)
				}
			}
		})
	}
}
