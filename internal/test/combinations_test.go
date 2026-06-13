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

func TestCombinations(t *testing.T) {
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
				{1},
				{1, 2},
				{2},
			},
		},
		{
			name:  "three",
			input: []int{1, 2, 3},
			want: [][]int{
				{1},
				{1, 2},
				{1, 2, 3},
				{1, 3},
				{2},
				{2, 3},
				{3},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Combinations is called.
			result := Combinations(tc.input)

			prefix := fmt.Sprintf(
				"%s\nCombinations(%v)",
				packageName, tc.input,
			)

			// THEN: the result should be all possible combinations of the values.
			if len(result) != len(tc.want) {
				t.Fatalf(
					"%s length mismatch\ngot:  %d (%v)\nwant: %d (%v)",
					prefix,
					len(result), result,
					len(tc.want), tc.want,
				)
			}
			for i, want := range tc.want {
				if len(result[i]) != len(want) {
					t.Fatalf(
						"%s items at [%d] differ in length:\ngot:  %v\nwant: %v",
						prefix, i,
						result, tc.want,
					)
				}
				for j, v := range want {
					if result[i][j] != v {
						t.Fatalf(
							"%s values at [%d][%d] differ:\ngot:  %v\nwant: %v",
							prefix, i, j,
							result, tc.want,
						)
					}
				}
			}
		})
	}
}
