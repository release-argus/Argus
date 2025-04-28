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
func InitMap(m *map[string]string) {
	if *m == nil {
		*m = make(map[string]string)
	}
}

// MergeMaps merges `m2` into `m1`,
// replacing any fields in `fields` with a value of 'SecretValue' with the corresponding value in `m2`.
func MergeMaps(m1, m2 map[string]string, fields []string) (m3 map[string]string) {
	m3 = CopyMap(m1)
	for k, v := range m2 {
		m3[k] = v
	}
	CopySecretValues(m1, m3, fields)
	return
}

// CopyMap will return a copy of the map.
func CopyMap[T, Y comparable](m map[T]Y) map[T]Y {
	m2 := make(map[T]Y, len(m))
	for key := range m {
		m2[key] = m[key]
	}
	return m2
}

// LowercaseStringStringMap converts all keys in the map to lowercase.
func LowercaseStringStringMap(change *map[string]string) {
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
	lowercasedMap := make(map[string]string, len(*change))
	for key, value := range *change {
		lowercasedMap[strings.ToLower(key)] = value
	}

	// Replace the original map with the lowercased map.
	*change = lowercasedMap
}

// SortedKeys will return a sorted list of the keys in a map.
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
