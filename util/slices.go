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

// Package util provides utility functions for the Argus project.
package util

import "bytes"

// Contains returns whether `s` contains `e`.
func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// CopySlice returns a copy of the slice.
func CopySlice[T any](s []T) []T {
	newList := make([]T, len(s))
	copy(newList, s)
	return newList
}

// SliceReplace replaces the element at `index` in `s` with the elements in `insert`.
func SliceReplace[T any](s []T, index int, insert []T) []T {
	if len(s) == 0 {
		return append([]T(nil), insert...)
	}

	if index < 0 {
		index = 0
	}
	if index >= len(s) {
		index = len(s) - 1
	}

	out := make([]T, 0, len(s)-1+len(insert))
	out = append(out, s[:index]...)
	out = append(out, insert...)
	out = append(out, s[index+1:]...)

	return out
}

// SwapRanges swaps two sub-slices within s, defined by their start and end indices.
func SwapRanges[T any](
	s []T,
	aStart, aEnd int,
	bStart, bEnd int,
) []T {
	// Always have the lower index as `a`.
	if aStart > bStart {
		aStart, bStart = bStart, aStart
		aEnd, bEnd = bEnd, aEnd
	}

	// Extract parts.
	a := append([]T(nil), s[aStart:aEnd+1]...)
	b := append([]T(nil), s[bStart:bEnd+1]...)

	// Build result.
	result := make([]T, 0, len(s))

	result = append(result, s[:aStart]...)       // before 'A'.
	result = append(result, b...)                // 'B' in place of 'A'.
	result = append(result, s[aEnd+1:bStart]...) // middle.
	result = append(result, a...)                // 'A' in place of 'B'.
	result = append(result, s[bEnd+1:]...)       // after 'B'.

	return result
}

// ReplaceFirst replaces the first occurrence of `old` with `new`.
func ReplaceFirst[T comparable](s []T, old, new T) []T {
	for i := range s {
		if s[i] == old {
			s[i] = new
			return s
		}
	}
	return s
}

// RemoveFirst removes the first instance of this element from the slice.
func RemoveFirst[T comparable](s []T, r T) []T {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// RemoveAt removes `index` from the slice.
func RemoveAt[T any](s []T, i int) []T {
	if i >= len(s) {
		return s
	}

	return append(s[:i], s[i+1:]...)[:len(s)-1]
}

// AreSlicesEqual reports whether two slices are identical in length and element values.
func AreSlicesEqual[T comparable](slice1, slice2 []T) bool {
	// Check if the lengths of the slices differ.
	if len(slice1) != len(slice2) {
		return false
	}

	// Compare each element in the slices.
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}

	// All elements are identical.
	return true
}

// FirstNonNilPtr returns the first non-nil pointer in `pointers`.
func FirstNonNilPtr[T any](pointers ...*T) *T {
	for _, pointer := range pointers {
		if pointer != nil {
			return pointer
		}
	}
	return nil
}

// FirstNonDefault returns the first non-default var in `vars`.
func FirstNonDefault[T comparable](vars ...T) T {
	var zero T
	for _, v := range vars {
		if v != zero {
			return v
		}
	}
	return zero
}

// FirstNonEmptySlice returns the first non-empty slice in `vars`.
func FirstNonEmptySlice[T ~[]E, E any](vars ...T) T {
	var zero T
	for _, v := range vars {
		if len(v) > 0 {
			return v
		}
	}
	return zero
}

// NormaliseNewlines replaces all newlines (Mac/Windows) in data with \n (Unix).
func NormaliseNewlines(data []byte) []byte {
	// replace CR LF \r\n (Windows) with LF \n (Unix).
	data = bytes.ReplaceAll(data, []byte{13, 10}, []byte{10})
	// replace CF \r (Mac) with LF \n (Unix).
	data = bytes.ReplaceAll(data, []byte{13}, []byte{10})

	return data
}
