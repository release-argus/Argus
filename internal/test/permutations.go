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

//go:build unit || integration

package test

import "slices"

// Permutations generates all permutations of a slice of any type T.
func Permutations[T any](input []T) [][]T {
	if len(input) == 0 {
		return nil
	}

	a := slices.Clone(input)
	result := make([][]T, 0)

	var generate func(int)
	generate = func(n int) {
		if n == 1 {
			result = append(result, slices.Clone(a))
			return
		}

		generate(n - 1)

		for i := 0; i < n-1; i++ {
			if n%2 == 0 {
				a[i], a[n-1] = a[n-1], a[i]
			} else {
				a[0], a[n-1] = a[n-1], a[0]
			}

			generate(n - 1)
		}
	}

	generate(len(a))
	return result
}
