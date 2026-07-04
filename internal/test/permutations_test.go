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

package test

import (
	"fmt"
	"testing"
)

func TestPermutations(t *testing.T) {
	// GIVEN: a slice of values.
	tests := []struct {
		name  string
		input []int
		want  [][]int
	}{
		{
			name:  "empty",
			input: []int{},
			want:  [][]int{},
		},
		{
			name:  "single",
			input: []int{1},
			want: [][]int{
				{1},
			},
		},
		{
			name:  "two",
			input: []int{1, 2},
			want: [][]int{
				{1, 2},
				{2, 1},
			},
		},
		{
			name:  "three",
			input: []int{1, 2, 3},
			want: [][]int{
				{1, 2, 3},
				{2, 1, 3},
				{3, 1, 2},
				{1, 3, 2},
				{2, 3, 1},
				{3, 2, 1},
			},
		},
		{
			name:  "duplicate values are not deduplicated",
			input: []int{1, 1},
			want: [][]int{
				{1, 1},
				{1, 1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Permutations is called.
			result := Permutations(tc.input)

			prefix := fmt.Sprintf(
				"%s\nPermutations(%v)",
				packageName, tc.input,
			)

			// THEN: the result should be all possible permutations of the values.
			if testErr := AssertSlicesEqualFunc(
				t, result, tc.want, func(a, b []int) bool {
					if len(a) != len(b) {
						return false
					}
					for i := range a {
						if a[i] != b[i] {
							return false
						}
					}
					return true
				},
				prefix,
				"",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestPermutations_ReturnsNilForEmptyInput(t *testing.T) {
	// GIVEN: an empty slice.
	input := []string{}

	// WHEN: Permutations is called.
	result := Permutations(input)

	// THEN: the result should be nil, not an empty, non-nil slice.
	if result != nil {
		t.Errorf(
			"%s\nPermutations(%v) mismatch\ngot:  %#v\nwant: nil",
			packageName, input, result,
		)
	}
}

func TestPermutations_DoesNotMutateInput(t *testing.T) {
	// GIVEN: a slice of values.
	input := []int{1, 2, 3}
	original := append([]int{}, input...)

	// WHEN: Permutations is called.
	Permutations(input)

	// THEN: the input slice is left unmodified.
	for i, v := range input {
		if v != original[i] {
			t.Errorf(
				"%s\nPermutations(%v) mutated its input\ngot:  %v\nwant: %v",
				packageName, original,
				input, original,
			)
			break
		}
	}
}

func TestPermutations_StringType(t *testing.T) {
	// GIVEN: a slice of strings.
	input := []string{"a", "b"}
	want := [][]string{
		{"a", "b"},
		{"b", "a"},
	}

	// WHEN: Permutations is called on a non-int generic type.
	result := Permutations(input)

	prefix := fmt.Sprintf(
		"%s\nPermutations(%v)",
		packageName, input,
	)

	// THEN: the result should be all possible permutations of the values.
	if len(result) != len(want) {
		t.Fatalf(
			"%s length mismatch\ngot:  %d (%v)\nwant: %d (%v)",
			prefix,
			len(result), result,
			len(want), want,
		)
	}
	for i, w := range want {
		if len(result[i]) != len(w) {
			t.Fatalf(
				"%s items at [%d] differ in length:\ngot:  %v\nwant: %v",
				prefix, i,
				result, want,
			)
		}
		for j, v := range w {
			if result[i][j] != v {
				t.Fatalf(
					"%s values at [%d][%d] differ:\ngot:  %v\nwant: %v",
					prefix, i, j,
					result, want,
				)
			}
		}
	}
}
