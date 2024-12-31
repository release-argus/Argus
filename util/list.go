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

// Package util provides utility functions for the Argus project.
package util

import (
	"reflect"
)

// Contains returns whether `s` contains `e`.
func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// CopyList returns a copy of the list.
func CopyList[T comparable](list []T) []T {
	newList := make([]T, len(list))
	copy(newList, list)
	return newList
}

// ReplaceWithElements replaces the element at `index` in `slice` with the elements in `insert`.
// It expands the slice as needed and shifts existing elements to make space for `insert`.
func ReplaceWithElements[T comparable](slice *[]T, index int, insert []T) {
	// Handle empty slice.
	if len(*slice) == 0 {
		*slice = append(*slice, make([]T, len(insert))...)
		copy(*slice, insert)
		return
	}

	// Ensure the index falls within the slice bounds.
	if index >= len(*slice) || index < 0 {
		index = max(len(*slice)-1, 0)
	}

	// Expand the slice to accommodate the new elements.
	expand := len(insert) - 1
	if expand > 0 {
		*slice = append(*slice, make([]T, expand)...)
		// Shift elements to create space for `insert`.
		copy((*slice)[index+len(insert):], (*slice)[index+1:])
	}

	// Insert the new elements into the created space.
	copy((*slice)[index:], insert)
}

// Swap swaps two sublists within a list, defined by their start and end indices.
//
// Parameters:
//   - list: A pointer to the list containing the sublists to swap.
//   - aStart, aEnd: Indices defining the first sublist.
//   - bStart, bEnd: Indices defining the second sublist.
func Swap[T comparable](
	list *[]T,
	aStart, aEnd int,
	bStart, bEnd int,
) {
	// Always have the lower index as a.
	if aStart > bStart {
		aStart, bStart = bStart, aStart
		aEnd, bEnd = bEnd, aEnd
	}

	aLen := aEnd - aStart + 1
	bLen := bEnd - bStart + 1
	swapper := reflect.Swapper(*list)

	// Direct swaps.
	index := 0
	for aStart+index <= aEnd && bStart+index <= bEnd {
		swapper(aStart+index, bStart+index)
		index++
	}

	// how many elements we need to shift.
	shiftNumber := bLen - aLen
	if shiftNumber == 0 {
		return
	}

	// Index to start swapping.
	var startAt int
	// Whether we're moving right(+)/left(-).
	var direction int
	// Amount of elements to shift by.
	var shiftBy int
	// More on the right, so we need to shift some to the left.
	if bLen > aLen {
		direction = -1
		startAt = bStart + index
		// shiftBy the <last-direct-swap-on-b> - <last-direct-swap-on-a>.
		shiftBy = (bStart + index - 1) - (aStart + index - 1)
	} else {
		// More on the left, so we need to shift some to the right.
		direction = 1
		startAt = aEnd
		// shiftBy the <last-direct-swap-on-b> - <aEnd>.
		shiftBy = (bStart + index - 1) - (aEnd)
		// Absolute shiftNumber.
		shiftNumber *= -1
	}

	loop := 0
	for loop < shiftNumber {
		at := startAt
		// Moving from left to right, so we will start on the right side and start 1 left each loop.
		at -= loop * direction

		shifted := 0
		// Shift `startAt` forward `shiftBy`.
		for shifted < shiftBy {
			swapper(at, at+(direction))
			at += direction
			shifted++
		}
		loop++
	}
}

// ReplaceElement will replace the first occurrence of `oldValue` with `newValue` in `list`.
func ReplaceElement(list []string, oldValue, newValue string) []string {
	for i := range list {
		if list[i] == oldValue {
			list[i] = newValue
			return list
		}
	}
	return list
}

// RemoveIndex from list.
func RemoveIndex[T comparable](list *[]T, index int) {
	if index >= len(*list) {
		return
	}

	*list = append((*list)[:index], (*list)[index+1:]...)[:len(*list)-1]
}

// RemoveElement from list.
func RemoveElement[T comparable](s []T, r T) []T {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
