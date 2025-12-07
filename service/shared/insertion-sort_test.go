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

import (
	"reflect"
	"testing"
)

func TestInsertionSort_int(t *testing.T) {
	tests := map[string]struct {
		name     string
		items    []int
		item     int
		less     func(a, b int) bool
		expected []int
	}{
		"Insert int into empty slice": {
			items:    []int{},
			item:     5,
			less:     func(a, b int) bool { return a < b },
			expected: []int{5},
		},
		"Insert at beginning": {
			items:    []int{3, 2, 1},
			item:     5,
			less:     func(a, b int) bool { return a < b },
			expected: []int{5, 3, 2, 1},
		},
		"Insert at end": {
			items:    []int{8, 6, 4},
			item:     2,
			less:     func(a, b int) bool { return a < b },
			expected: []int{8, 6, 4, 2},
		},
		"Insert in the middle": {
			items:    []int{8, 6, 4},
			item:     5,
			less:     func(a, b int) bool { return a < b },
			expected: []int{8, 6, 5, 4},
		},
		"Insert duplicate value": {
			items:    []int{8, 6, 4},
			item:     6,
			less:     func(a, b int) bool { return a < b },
			expected: []int{8, 6, 6, 4},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN InsertionSort is called.
			result := InsertionSort(tc.items, tc.item, tc.less)

			// THEN the result is as expected.
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("%s\nwant: %v\ngot:  %v",
					packageName, result, tc.expected)
			}
		})
	}
}

func TestInsertionSort_String(t *testing.T) {
	t.Parallel()

	type testCase struct {
		items    []string
		item     string
		less     func(a, b string) bool
		expected []string
	}

	tests := map[string]testCase{
		"Insert string into empty slice": {
			items:    []string{},
			item:     "banana",
			less:     func(a, b string) bool { return a < b },
			expected: []string{"banana"},
		},
		"Insert at beginning (desc order)": {
			items:    []string{"c", "b", "a"},
			item:     "d",
			less:     func(a, b string) bool { return a < b },
			expected: []string{"d", "c", "b", "a"},
		},
		"Insert in the middle": {
			items:    []string{"z", "m", "a"},
			item:     "h",
			less:     func(a, b string) bool { return a < b },
			expected: []string{"z", "m", "h", "a"},
		},
		"Insert at end": {
			items:    []string{"z", "y", "x"},
			item:     "w",
			less:     func(a, b string) bool { return a < b },
			expected: []string{"z", "y", "x", "w"},
		},
		"Insert duplicate": {
			items:    []string{"c", "b"},
			item:     "b",
			less:     func(a, b string) bool { return a < b },
			expected: []string{"c", "b", "b"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN InsertionSort is called.
			result := InsertionSort(tc.items, tc.item, tc.less)

			// THEN the result is as expected.
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("%s\nwant: %v\ngot:  %v",
					packageName, tc.expected, result)
			}
		})
	}
}

func TestInsertionSort_Struct(t *testing.T) {
	type person struct {
		name string
		age  int
	}

	lessByAge := func(a, b person) bool { return a.age < b.age }

	// GIVEN a slice of persons and a person to insert.
	tests := map[string]struct {
		items    []person
		item     person
		expected []person
	}{
		"Insert person into empty slice": {
			items:    []person{},
			item:     person{name: "Bob", age: 30},
			expected: []person{{name: "Bob", age: 30}},
		},
		"Insert at beginning": {
			items: []person{{"Ann", 40}, {"Bill", 35}, {"Cara", 20}},
			item:  person{"Zed", 50},
			expected: []person{
				{"Zed", 50}, {"Ann", 40}, {"Bill", 35}, {"Cara", 20},
			},
		},
		"Insert in the middle": {
			items: []person{{"Ann", 40}, {"Bill", 35}, {"Cara", 20}},
			item:  person{"Dom", 33},
			expected: []person{
				{"Ann", 40}, {"Bill", 35}, {"Dom", 33}, {"Cara", 20},
			},
		},
		"Insert at end": {
			items:    []person{{"Ann", 40}, {"Bill", 35}},
			item:     person{"Eve", 10},
			expected: []person{{"Ann", 40}, {"Bill", 35}, {"Eve", 10}},
		},
		"Insert duplicate age": {
			items:    []person{{"Ann", 40}, {"Bill", 35}},
			item:     person{"Ben", 35},
			expected: []person{{"Ann", 40}, {"Bill", 35}, {"Ben", 35}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN InsertionSort is called.
			result := InsertionSort(tc.items, tc.item, lessByAge)

			// THEN the result is as expected.
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("%s\nwant: %v\ngot:  %v",
					packageName, tc.expected, result)
			}
		})
	}
}
