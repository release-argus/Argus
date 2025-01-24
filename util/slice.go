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

// Package util provides utility functions for the Argus project.
package util

import "bytes"

type customComparable interface {
	bool | int | map[string]string | string | uint8 | uint16
}

// FirstNonNilPtr will return the first non-nil pointer in `pointers`.
func FirstNonNilPtr[T customComparable](pointers ...*T) *T {
	for _, pointer := range pointers {
		if pointer != nil {
			return pointer
		}
	}
	return nil
}

// FirstNonDefault will return the first non-default var in `vars`.
func FirstNonDefault[T comparable](vars ...T) T {
	var fresh T
	for _, v := range vars {
		if v != fresh {
			return v
		}
	}
	return fresh
}

type comparableElement interface {
	comparable
	bool | int | string | uint8 | uint16
}

// AreStringSlicesEqual compares two slices of strings and returns true if they are identical.
// It checks both the length of the slices and the values at each index.
// If the slices have different lengths or any corresponding elements differ, it returns false.
func AreSlicesEqual[T comparableElement](slice1, slice2 []T) bool {
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

// NormaliseNewlines all newlines in `data` to \n.
func NormaliseNewlines(data []byte) []byte {
	// replace CR LF \r\n (Windows) with LF \n (Unix).
	data = bytes.ReplaceAll(data, []byte{13, 10}, []byte{10})
	// replace CF \r (Mac) with LF \n (Unix).
	data = bytes.ReplaceAll(data, []byte{13}, []byte{10})

	return data
}
