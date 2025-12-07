// Copyright [2025] [Argus]
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

package shared

import "sort"

// InsertionSort performs an 'insertion sort' of item into items.
// It inserts an item into a sorted slice, maintaining the order as determined by the less function.
func InsertionSort[T any](items []T, item T, less func(a, b T) bool) []T {
	// Find insertion index.
	i := sort.Search(len(items), func(idx int) bool {
		return less(items[idx], item)
	})

	// Grow slice.
	items = append(items, *new(T))
	// Shift right.
	copy(items[i+1:], items[i:])
	// Insert.
	items[i] = item

	return items
}
