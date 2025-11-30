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

import (
	"sort"
	"strings"
)

// InitMap will initialise the map.
func InitMap[M ~map[K]V, K, V comparable](m *M) {
	if *m == nil {
		*m = make(M)
	}
}

// MergeMaps merges `m2` into `m1`,
// replacing any fields in `fields` with a value of 'SecretValue' with the corresponding value in `m2`.
func MergeMaps[K comparable](m1, m2 map[K]string, fields []K) map[K]string {
	m3 := CopyMap(m1)
	for k, v := range m2 {
		m3[k] = v
	}
	CopySecretValues(m1, m3, fields)
	return m3
}

// CopyMap returns a copy of the map.
func CopyMap[K, V comparable, M ~map[K]V](m M) M {
	m2 := make(M, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

// LowercaseStringStringMap converts all keys in the map to lowercase.
func LowercaseStringStringMap[M ~map[string]string](change *M) {
	// Check for a non-lowercase key.
	allLowercase := true
	for key := range *change {
		if key != strings.ToLower(key) {
			allLowercase = false
			break
		}
	}
	// If all keys lowercase, do nothing.
	if allLowercase && *change != nil {
		return
	}

	// Otherwise, create a new map with lowercase keys.
	lowercasedMap := make(M, len(*change))
	for key, value := range *change {
		lowercasedMap[strings.ToLower(key)] = value
	}

	// Replace the original map with the lowercased map.
	*change = lowercasedMap
}

// SortedKeys returns a sorted list of the keys in a map.
func SortedKeys[V any](m map[string]V) (keys []string) {
	keys = make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return
}
