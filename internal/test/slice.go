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

// Attributes returns selector applied to each item, in input order.
func Attributes[T any, V ~string](items []T, selector func(T) V) []V {
	result := make([]V, len(items))
	for i, item := range items {
		result[i] = selector(item)
	}
	return result
}

// AttributesSorted returns selector applied to each item, in ascending order.
func AttributesSorted[T any, V ~string](items []T, selector func(T) V) []V {
	result := Attributes(items, selector)
	slices.Sort(result)
	return result
}
