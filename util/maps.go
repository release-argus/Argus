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

import (
	"maps"
	"slices"
	"strings"
)

// EnsureMap will ensure the map is initialised.
func EnsureMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return make(map[K]V)
	}
	return m
}

// MergeMaps merges m2 into a copy of m1 and returns the result.
func MergeMaps[K comparable](m1, m2 map[K]string) map[K]string {
	m3 := CopyMap(m1)
	maps.Copy(m3, m2)
	return m3
}

// CopyMap returns a copy of the map.
func CopyMap[K comparable, V any](m map[K]V) map[K]V {
	out := make(map[K]V, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// LowercaseKeys converts all keys in the map to lowercase.
func LowercaseKeys(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}

	out := make(map[string]string, len(m))
	for k, v := range m {
		out[strings.ToLower(k)] = v
	}

	return out
}

// SortedKeys returns a sorted list of the keys in a map.
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)
	return keys
}

// FirstNonEmptyMap returns the first map in maps with at least one entry.
// It returns nil when every map is empty or nil.
func FirstNonEmptyMap[K comparable, V any](maps ...map[K]V) map[K]V {
	for _, m := range maps {
		if len(m) > 0 {
			return m
		}
	}
	return nil
}
