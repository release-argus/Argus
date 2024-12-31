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

//go:build unit

package util

import (
	"testing"
)

func TestContains(t *testing.T) {
	// GIVEN lists of strings
	tests := map[string]struct {
		list        []string
		contain     string
		doesContain bool
	}{
		"[]string does contain": {
			list:    []string{"hello", "hi", "hiya"},
			contain: "hi", doesContain: true},
		"[]string does not contain": {
			list:    []string{"hello", "hi", "hiya"},
			contain: "howdy", doesContain: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Contains is run on this list with a element inside it
			var found bool
			found = Contains(tc.list, tc.contain)

			// THEN true is returned if it does contain the item
			if found != tc.doesContain {
				t.Errorf("want Contains=%t, got Contains=%t",
					found, tc.doesContain)
			}
		})
	}
}

func TestCopyList(t *testing.T) {
	// GIVEN a set of lists
	tests := map[string]struct {
		had, want []int
	}{
		"empty list": {
			had:  []int{},
			want: []int{},
		},
		"one element": {
			had:  []int{1},
			want: []int{1},
		},
		"multiple elements": {
			had:  []int{1, 2, 3, 4, 5},
			want: []int{1, 2, 3, 4, 5},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CopyList is called on a list
			got := CopyList(tc.had)

			// THEN the copy is successful
			if len(got) != len(tc.want) {
				t.Fatalf("CopyList failed:\nwant:%v\ngot:  %v",
					tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("\nwant: %v\ngot:  %v",
						tc.want, got)
				}
			}
		})
	}
}

func TestReplaceWithElements(t *testing.T) {
	// GIVEN a set of lists and an element to replace
	tests := map[string]struct {
		had, want []int
		insert    []int
		index     int
	}{
		"insert at beginning": {
			had:    []int{1, 2, 3},
			insert: []int{0},
			index:  0,
			want:   []int{0, 2, 3},
		},
		"insert at end": {
			had:    []int{1, 2, 3},
			insert: []int{4},
			index:  2,
			want:   []int{1, 2, 4},
		},
		"insert in middle": {
			had:    []int{1, 2, 4, 5},
			insert: []int{3},
			index:  2,
			want:   []int{1, 2, 3, 5},
		},
		"insert multiple elements": {
			had:    []int{1, 2, 5, 6},
			insert: []int{3, 4, 5},
			index:  2,
			want:   []int{1, 2, 3, 4, 5, 6},
		},
		"insert single into empty list": {
			had:    []int{},
			insert: []int{1},
			index:  0,
			want:   []int{1},
		},
		"insert multiple into empty list": {
			had:    []int{},
			insert: []int{1, 2, 3},
			index:  0,
			want:   []int{1, 2, 3},
		},
		"insert at out of range index": {
			had:    []int{1, 2, 3},
			insert: []int{4, 5},
			index:  5,
			want:   []int{1, 2, 4, 5},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ReplaceWithElements is called on a list
			ReplaceWithElements(&tc.had, tc.index, tc.insert)

			// THEN the replacement is successful
			if len(tc.had) != len(tc.want) {
				t.Fatalf("ReplaceWithElements failed:\nwant:%v\ngot:  %v",
					tc.want, tc.had)
			}
			for i := range tc.had {
				if tc.had[i] != tc.want[i] {
					t.Fatalf("\nwant: %v\ngot:  %v",
						tc.want, tc.had)
				}
			}
		})
	}
}

func TestSwap(t *testing.T) {
	// GIVEN a set of lists
	tests := map[string]struct {
		had, want    []int
		aStart, aEnd int
		bStart, bEnd int
	}{
		"handles []int": {
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{0, 7, 8, 9, 2, 3, 4, 5, 6, 1},
			aStart: 7, aEnd: 9,
			bStart: 1, bEnd: 1,
		},
		"swap single element": {
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{9, 1, 2, 3, 4, 5, 6, 7, 8, 0},
			aStart: 0, aEnd: 0,
			bStart: 9, bEnd: 9,
		},
		"matching swap sizes": {
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{8, 9, 2, 3, 4, 5, 6, 7, 0, 1},
			aStart: 0, aEnd: 1,
			bStart: 8, bEnd: 9,
		},
		"more on left": {
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{0, 8, 9, 5, 6, 7, 1, 2, 3, 4},
			aStart: 1, aEnd: 4,
			bStart: 8, bEnd: 9,
		},
		"more on right": {
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{0, 7, 8, 9, 2, 3, 4, 5, 6, 1},
			aStart: 1, aEnd: 1,
			bStart: 7, bEnd: 9,
		},
		"indices wrong way round": {
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{0, 7, 8, 9, 2, 3, 4, 5, 6, 1},
			aStart: 7, aEnd: 9,
			bStart: 1, bEnd: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Swap is called on a list
			Swap(&tc.had, tc.aStart, tc.aEnd, tc.bStart, tc.bEnd)

			// THEN the Swap is successful
			// int
			if len(tc.had) != len(tc.want) {
				t.Fatalf("Swap added/removed elements!\nwant:%v\ngot:  %v",
					tc.want, tc.had)
			}
			for i := range tc.had {
				if tc.had[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, tc.had)
				}
			}
		})
	}
}

func TestReplaceElement(t *testing.T) {
	// GIVEN a set of lists
	tests := map[string]struct {
		had, want     []string
		replace, with string
	}{
		"Replace element at end": {
			had:     []string{"foo", "bar", "baz"},
			want:    []string{"foo", "bar", "bash"},
			replace: "baz",
			with:    "bash",
		},
		"Replace element at start": {
			had:     []string{"foo", "bar", "baz"},
			want:    []string{"bash", "bar", "baz"},
			replace: "foo",
			with:    "bash",
		},
		"Replace element in middle": {
			had:     []string{"foo", "bar", "baz"},
			want:    []string{"foo", "bash", "baz"},
			replace: "bar",
			with:    "bash",
		},
		"Replace element not in list": {
			had:     []string{"foo", "bar", "baz"},
			want:    []string{"foo", "bar", "baz"},
			replace: "bash",
			with:    "bosh",
		}}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ReplaceElement is called on a list
			got := ReplaceElement(tc.had, tc.replace, tc.with)

			// THEN the Replacement is successful
			if len(got) != len(tc.want) {
				t.Fatalf("ReplaceElement failed:\nwant:%v\ngot:  %v",
					tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("\nwant: %v\ngot:  %v",
						tc.want, got)
				}
			}
		})
	}
}

func TestRemoveIndex(t *testing.T) {
	// GIVEN a set of lists
	tests := map[string]struct {
		had, want []int
		wantStr   []string
		index     int
	}{
		"handles []int": {
			had:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:  []int{0, 1, 2, 3, 5, 6, 7, 8, 9},
			index: 4,
		},
		"out of range": {
			had:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:  []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			index: 10,
		},
		"first index": {
			had:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			index: 0,
		},
		"last index": {
			had:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:  []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			index: 9,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN RemoveIndex is called on a list
			RemoveIndex(&tc.had, tc.index)

			// THEN the Removal is successful
			// int
			if len(tc.had) != len(tc.want) {
				t.Fatalf("Remove index failed\nwant:%v\ngot:  %v",
					tc.want, tc.had)
			}
			for i := range tc.had {
				if tc.had[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, tc.had)
				}
			}
		})
	}
}

func TestRemoveElement(t *testing.T) {
	// GIVEN a set of lists
	tests := map[string]struct {
		removeInt       int
		hadInt, wantInt []int
		removeStr       string
		hadStr, wantStr []string
	}{
		"handles []int": {
			removeInt: 7,
			hadInt:    []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
			wantInt:   []int{9, 8, 6, 5, 4, 3, 2, 1, 0},
		},
		"handles []string": {
			removeStr: "bash",
			hadStr:    []string{"bish", "bash", "bosh", "bush"},
			wantStr:   []string{"bish", "bosh", "bush"},
		},
		"handles []string with duplicates": {
			removeStr: "bash",
			hadStr:    []string{"bish", "bash", "bosh", "bush", "bash"},
			wantStr:   []string{"bish", "bosh", "bush", "bash"},
		},
		"handles []string with no match": {
			removeStr: "bush",
			hadStr:    []string{"bish", "bosh", "bosh"},
			wantStr:   []string{"bish", "bosh", "bosh"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN RemoveElement is called on a list
			gotInt := RemoveElement(tc.hadInt, tc.removeInt)
			gotStr := RemoveElement(tc.hadStr, tc.removeStr)

			// THEN the Removal is successful
			// int
			if len(gotInt) != len(tc.wantInt) {
				t.Fatalf("Remove element failed\nwant:%v\ngot:  %v",
					tc.wantInt, gotInt)
			}
			for i := range gotInt {
				if gotInt[i] != tc.wantInt[i] {
					t.Fatalf("Mismatch at index %d: got %d, want %d\ngot:  %v\nwant: %v",
						i, gotInt[i], tc.wantInt[i], gotInt, tc.wantInt)
				}
			}
			// string
			if len(gotStr) != len(tc.wantStr) {
				t.Fatalf("Remove element failed\nwant:%v\ngot:  %v",
					tc.wantStr, gotStr)
			}
			for i := range gotStr {
				if gotStr[i] != tc.wantStr[i] {
					t.Fatalf("Mismatch at index %d: got %s, want %s\ngot:  %v\nwant: %v",
						i, gotStr[i], tc.wantStr[i], gotStr, tc.wantStr)
				}
			}
		})
	}
}
