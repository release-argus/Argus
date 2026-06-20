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

// Combinations returns all possible combinations of the input elements.
func Combinations[T comparable](input []T) [][]T {
	var result [][]T

	var generate func(index int, current []T)
	generate = func(index int, current []T) {
		if index == len(input) {
			return
		}

		for i := index; i < len(input); i++ {
			newCombination := append([]T{}, current...)
			newCombination = append(newCombination, input[i])
			result = append(result, newCombination)
			generate(i+1, newCombination)
		}
	}

	generate(0, []T{})
	return result
}
