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

//go:build unit

package util

import (
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestContains(t *testing.T) {
	// GIVEN: lists of strings.
	tests := []struct {
		name        string
		list        []string
		contain     string
		doesContain bool
	}{
		{
			name:        "[]string does contain",
			list:        []string{"hello", "hi", "hiya"},
			contain:     "hi",
			doesContain: true,
		},
		{
			name:        "[]string does not contain",
			list:        []string{"hello", "hi", "hiya"},
			contain:     "howdy",
			doesContain: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Contains is run on this slice.
			var found bool
			found = Contains(tc.list, tc.contain)

			prefix := fmt.Sprintf(
				"%s\nContains(list=%v, contain=%q)",
				packageName, tc.list, tc.contain,
			)

			// THEN: true is returned if it does contain the item.
			if found != tc.doesContain {
				t.Errorf(
					"%s mismatch\ngot:  %t\nwant: %t",
					prefix, found, tc.doesContain,
				)
			}
		})
	}
}

func TestCopySlice(t *testing.T) {
	// GIVEN: a set of int slices.
	tests := []struct {
		name      string
		had, want []int
	}{
		{
			name: "empty list",
			had:  []int{},
			want: []int{},
		},
		{
			name: "one element",
			had:  []int{1},
			want: []int{1},
		},
		{
			name: "multiple elements",
			had:  []int{1, 2, 3, 4, 5},
			want: []int{1, 2, 3, 4, 5},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: CopySlice is called on a list.
			result := CopySlice(tc.had)

			prefix := fmt.Sprintf("%s\nCopySlice()", packageName)

			// THEN: the copy is successful.
			if gotLen, wantLen := len(result), len(tc.want); gotLen != wantLen {
				t.Fatalf(
					"%s length mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, wantLen,
				)
			}
			for i := range result {
				if got, want := result[i], tc.want[i]; got != want {
					t.Fatalf(
						"%s mismatch at index %d\ngot:  %v (%+v)\nwant: %v (%+v)",
						prefix, i,
						got, result,
						want, tc.want,
					)
				}
			}
		})
	}
}

func TestSliceReplace(t *testing.T) {
	// GIVEN: a slice and an element to replace at a given index.
	tests := []struct {
		name      string
		had, want []int
		insert    []int
		index     int
	}{
		{
			name:   "insert at beginning",
			had:    []int{1, 2, 3},
			insert: []int{0},
			index:  0,
			want:   []int{0, 2, 3},
		},
		{
			name:   "insert at end",
			had:    []int{1, 2, 3},
			insert: []int{4},
			index:  2,
			want:   []int{1, 2, 4},
		},
		{
			name:   "insert in middle",
			had:    []int{1, 2, 4, 5},
			insert: []int{3},
			index:  2,
			want:   []int{1, 2, 3, 5},
		},
		{
			name:   "insert multiple elements",
			had:    []int{1, 2, 5, 6},
			insert: []int{3, 4, 5},
			index:  2,
			want:   []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:   "insert single into empty list",
			had:    []int{},
			insert: []int{1},
			index:  0,
			want:   []int{1},
		},
		{
			name:   "insert multiple into empty list",
			had:    []int{},
			insert: []int{1, 2, 3},
			index:  0,
			want:   []int{1, 2, 3},
		},
		{
			name:   "insert at out of range index",
			had:    []int{1, 2, 3},
			insert: []int{4, 5},
			index:  5,
			want:   []int{1, 2, 4, 5},
		},
		{
			name:   "insert at negative index",
			had:    []int{1, 2, 3},
			insert: []int{0},
			index:  -1,
			want:   []int{0, 2, 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: SliceReplace is called on a slice.
			got := SliceReplace(tc.had, tc.index, tc.insert)

			prefix := fmt.Sprintf(
				"%s\nSliceReplace(slice=%v, index=%d, insert=%v)",
				packageName, tc.had, tc.index, tc.insert,
			)

			// THEN: the replacement is successful.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a, b int) bool { return a == b },
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestSwapRanges(t *testing.T) {
	// GIVEN: a set of slices.
	tests := []struct {
		name         string
		had, want    []int
		aStart, aEnd int
		bStart, bEnd int
	}{
		{
			name:   "handles []int",
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{0, 7, 8, 9, 2, 3, 4, 5, 6, 1},
			aStart: 7, aEnd: 9,
			bStart: 1, bEnd: 1,
		},
		{
			name:   "swap single element",
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{9, 1, 2, 3, 4, 5, 6, 7, 8, 0},
			aStart: 0, aEnd: 0,
			bStart: 9, bEnd: 9,
		},
		{
			name:   "matching swap sizes",
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{8, 9, 2, 3, 4, 5, 6, 7, 0, 1},
			aStart: 0, aEnd: 1,
			bStart: 8, bEnd: 9,
		},
		{
			name:   "more on left",
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{0, 8, 9, 5, 6, 7, 1, 2, 3, 4},
			aStart: 1, aEnd: 4,
			bStart: 8, bEnd: 9,
		},
		{
			name:   "more on right",
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{0, 7, 8, 9, 2, 3, 4, 5, 6, 1},
			aStart: 1, aEnd: 1,
			bStart: 7, bEnd: 9,
		},
		{
			name:   "indices wrong way round",
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{0, 7, 8, 9, 2, 3, 4, 5, 6, 1},
			aStart: 7, aEnd: 9,
			bStart: 1, bEnd: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: SwapRanges is called on them.
			result := SwapRanges(tc.had, tc.aStart, tc.aEnd, tc.bStart, tc.bEnd)

			prefix := fmt.Sprintf(
				"%s\nSwapRanges(%v, indexes %d-%d with %d-%d)",
				packageName,
				tc.had,
				tc.aEnd, tc.bStart,
				tc.bEnd, tc.aEnd,
			)

			// THEN: the swap is successful.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				result,
				tc.want,
				func(a, b int) bool { return a == b },
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestReplaceFirst_String(t *testing.T) {
	// GIVEN: a slice and a new value for a given element in it.
	tests := []struct {
		name          string
		had, want     []string
		replace, with string
	}{
		{
			name:    "Replace element at end",
			had:     []string{"foo", "bar", "baz"},
			want:    []string{"foo", "bar", "bash"},
			replace: "baz",
			with:    "bash",
		},
		{
			name:    "Replace element at start",
			had:     []string{"foo", "bar", "baz"},
			want:    []string{"bash", "bar", "baz"},
			replace: "foo",
			with:    "bash",
		},
		{
			name:    "Replace element in middle",
			had:     []string{"foo", "bar", "baz"},
			want:    []string{"foo", "bash", "baz"},
			replace: "bar",
			with:    "bash",
		},
		{
			name:    "Replace element not in list",
			had:     []string{"foo", "bar", "baz"},
			want:    []string{"foo", "bar", "baz"},
			replace: "bash",
			with:    "bosh",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: ReplaceFirst is called on a list.
			result := ReplaceFirst(tc.had, tc.replace, tc.with)

			prefix := fmt.Sprintf(
				"%s\nReplaceFirst(slice=%v, old=%v, new=%v)",
				packageName, tc.had, tc.replace, tc.with,
			)

			// THEN: the Replacement is successful.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				result,
				tc.want,
				func(a, b string) bool { return a == b },
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestRemoveFirst_Int(t *testing.T) {
	// GIVEN: a set of lists.
	tests := []struct {
		name            string
		removeInt       int
		hadInt, wantInt []int
	}{
		{
			name:      "one occurrence/first",
			removeInt: 3,
			hadInt:    []int{3, 2, 1},
			wantInt:   []int{2, 1},
		},
		{
			name:      "one occurrence/middle",
			removeInt: 2,
			hadInt:    []int{3, 2, 1},
			wantInt:   []int{3, 1},
		},
		{
			name:      "one occurrence/end",
			removeInt: 1,
			hadInt:    []int{3, 2, 1},
			wantInt:   []int{3, 2},
		},
		{
			name:      "only removes occurrence",
			removeInt: 2,
			hadInt:    []int{3, 2, 1, 2},
			wantInt:   []int{3, 1, 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: RemoveFirst is called on a list.
			gotInt := RemoveFirst(tc.hadInt, tc.removeInt)

			prefix := fmt.Sprintf(
				"%s\nRemoveElement(from=%v, target=%v)",
				packageName, tc.hadInt, tc.removeInt,
			)

			// THEN: the Removal is successful.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				gotInt,
				tc.wantInt,
				func(a, b int) bool { return a == b },
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestRemoveFirst_String(t *testing.T) {
	// GIVEN: a string slice.
	tests := []struct {
		name      string
		remove    string
		had, want []string
	}{
		{
			name:   "handles []string",
			remove: "bash",
			had:    []string{"bish", "bash", "bosh", "bush"},
			want:   []string{"bish", "bosh", "bush"},
		},
		{
			name:   "handles []string with duplicates",
			remove: "bash",
			had:    []string{"bish", "bash", "bosh", "bush", "bash"},
			want:   []string{"bish", "bosh", "bush", "bash"},
		},
		{
			name:   "handles []string with no match",
			remove: "bush",
			had:    []string{"bish", "bosh", "bosh"},
			want:   []string{"bish", "bosh", "bosh"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: RemoveFirst is called on a list.
			got := RemoveFirst(tc.had, tc.remove)

			prefix := fmt.Sprintf(
				"%s\nRemoveElement(from=%v, target=%v)",
				packageName, tc.had, tc.remove,
			)

			// THEN: the Removal is successful.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a, b string) bool { return a == b },
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestRemoveAt_Int(t *testing.T) {
	// GIVEN: a slice and an index to remove in it.
	tests := []struct {
		name      string
		had, want []int
		index     int
	}{
		{
			name:  "simple []int",
			had:   []int{3, 2, 1},
			want:  []int{3, 1},
			index: 1,
		},
		{
			name:  "out of range",
			had:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:  []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			index: 10,
		},
		{
			name:  "first index",
			had:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			index: 0,
		},
		{
			name:  "last index",
			had:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:  []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			index: 9,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: RemoveAt is called on it.
			result := RemoveAt(tc.had, tc.index)

			prefix := fmt.Sprintf(
				"%s\nRemoveIndex(from=%v, imdex=%d)",
				packageName, tc.index, tc.index,
			)

			// THEN: the Removal is successful.
			if testErr := test.AssertSlicesEqualFunc(
				t,
				result,
				tc.want,
				func(a, b int) bool { return a == b },
				prefix,
				"",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestAreSlicesEqual(t *testing.T) {
	// GIVEN: different slices.
	tests := []struct {
		name           string
		slice1, slice2 []string
		want           bool
	}{
		{
			name:   "both empty",
			slice1: []string{},
			slice2: []string{},
			want:   true,
		},
		{
			name:   "one empty",
			slice1: []string{"foo"},
			slice2: []string{},
			want:   false,
		},
		{
			name:   "same length, same elements",
			slice1: []string{"foo", "bar"},
			slice2: []string{"foo", "bar"},
			want:   true,
		},
		{
			name:   "different elements",
			slice1: []string{"foo", "bar"},
			slice2: []string{"bar", "foo"},
			want:   false,
		},
		{
			name:   "different lengths",
			slice1: []string{"foo", "bar"},
			slice2: []string{"foo"},
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: AreSlicesEqual is called.
			got := AreSlicesEqual(tc.slice1, tc.slice2)

			prefix := fmt.Sprintf(
				"%s\nAreSlicesEqual(a=%v, b=%v)",
				packageName, tc.slice1, tc.slice2,
			)

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s mismatch\ngot:  %v\nwant: %v",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestFirstNonNilPtr(t *testing.T) {
	// GIVEN: a bunch of pointers.
	tests := []struct {
		name      string
		pointers  []*string
		allNil    bool
		wantIndex int
	}{
		{
			name:     "no pointers",
			pointers: []*string{},
			allNil:   true,
		},
		{
			name: "all nil pointers",
			pointers: []*string{
				nil,
				nil,
				nil,
				nil,
			},
			allNil: true,
		},
		{
			name: "1 non-nil pointer",
			pointers: []*string{
				nil,
				nil,
				nil,
				test.Ptr("bar"),
			},
			wantIndex: 3,
		},
		{
			name: "2 non-nil pointers",
			pointers: []*string{
				test.Ptr("foo"),
				nil,
				nil,
				test.Ptr("bar"),
			},
			wantIndex: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: FirstNonNilPtr is run on a slice of pointers.
			got := FirstNonNilPtr(tc.pointers...)

			prefix := fmt.Sprintf(
				"%s\nFirstNonNilPtr(%v)",
				packageName, tc.pointers,
			)

			// THEN: the correct pointer (or nil) is returned.
			if tc.allNil {
				if got != nil {
					t.Fatalf(
						"%s pointer mismatch\ngot:  %v\nwant: nil",
						prefix, got,
					)
				}
				return
			}
			if got != tc.pointers[tc.wantIndex] {
				t.Errorf(
					"%s pointer mismatch\ngot:  %p\nwant: %p",
					prefix, got, tc.pointers[tc.wantIndex],
				)
			}
		})
	}
}

func TestFirstNonDefault(t *testing.T) {
	// GIVEN: a bunch of comparables.
	tests := []struct {
		name       string
		slice      []string
		allDefault bool
		wantIndex  int
	}{
		{
			name:       "no vars",
			slice:      []string{},
			allDefault: true,
		},
		{
			name: "all default vars",
			slice: []string{
				"",
				"",
				"",
				"",
			},
			allDefault: true,
		},
		{
			name: "1 non-default var",
			slice: []string{
				"",
				"",
				"",
				"bar",
			},
			wantIndex: 3,
		},
		{
			name: "2 non-default vars",
			slice: []string{
				"foo",
				"",
				"",
				"bar",
			},
			wantIndex: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: FirstNonDefault is run on a slice of slice.
			got := FirstNonDefault(tc.slice...)

			prefix := fmt.Sprintf(
				"%s\nFirstNonDefault(%v)",
				packageName, tc.slice,
			)

			// THEN: the correct var (or "") is returned.
			if tc.allDefault {
				want := ""
				if got != want {
					t.Fatalf(
						"%s mismatch\ngot:  %v\nwant: %q",
						prefix, got, want,
					)
				}
				return
			}
			if got != tc.slice[tc.wantIndex] {
				t.Errorf(
					"%s mismatch\ngot:  %v\nwant: %v",
					prefix, got, tc.slice[tc.wantIndex],
				)
			}
		})
	}
}

func TestFirstNonEmptySlice(t *testing.T) {
	// GIVEN: a bunch of slices.
	tests := []struct {
		name      string
		slices    [][]string
		wantIndex int
	}{
		{
			name:      "no slices",
			slices:    [][]string{},
			wantIndex: -1,
		},
		{
			name: "all empty slices",
			slices: [][]string{
				{},
				{},
				{},
				{},
			},
			wantIndex: -1,
		},
		{
			name: "all filled slices",
			slices: [][]string{
				{"1"},
				{"1", "2"},
				{"1", "2", "3"},
				{"1", "2", "4"},
			},
			wantIndex: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: FirstNonEmptySlice is run on a slice of slices.
			got := FirstNonEmptySlice(tc.slices...)

			prefix := fmt.Sprintf(
				"%s\nFirstNonEmptySlice(%v)",
				packageName, tc.slices,
			)

			// THEN: the first non-empty slice is returned.
			if got == nil {
				if tc.wantIndex != -1 {
					t.Errorf(
						"%s mismatch\ngot:  nil\nwant: %v",
						prefix, tc.slices[tc.wantIndex],
					)
				}
			} else {
				gotStr := strings.Join(got, ",")
				wantStr := strings.Join(tc.slices[tc.wantIndex], ",")
				if gotStr != wantStr {
					t.Errorf(
						"%s mismatch\ngot:  %v\nwant: %v",
						prefix, gotStr, wantStr,
					)
				}
			}
		})
	}
}

func TestNormaliseNewlines(t *testing.T) {
	// GIVEN: different byte strings.
	tests := []struct {
		name        string
		input, want []byte
	}{
		{
			name:  "string with no newlines",
			input: []byte("hello there"),
			want:  []byte("hello there"),
		},
		{
			name:  "string with linux newlines",
			input: []byte("hello\nthere"),
			want:  []byte("hello\nthere"),
		},
		{
			name:  "string with multiple linux newlines",
			input: []byte("hello\nthere\n"),
			want:  []byte("hello\nthere\n"),
		},
		{
			name:  "string with windows newlines",
			input: []byte("hello\r\nthere"),
			want:  []byte("hello\nthere"),
		},
		{
			name:  "string with multiple windows newlines",
			input: []byte("hello\r\nthere\r\n"),
			want:  []byte("hello\nthere\n"),
		},
		{
			name:  "string with mac newlines",
			input: []byte("hello\r\nthere"),
			want:  []byte("hello\nthere"),
		},
		{
			name:  "string with multiple mac newlines",
			input: []byte("hello\r\nthere\r\n"),
			want:  []byte("hello\nthere\n"),
		},
		{
			name:  "string with multiple mac and windows newlines",
			input: []byte("\rhello\r\nthere\r\n. hi\r"),
			want:  []byte("\nhello\nthere\n. hi\n"),
		},
		{
			name:  "string with multiple mac, windows and linux newlines",
			input: []byte("\rhello\r\nthere\r\n. hi\r. foo\nbar\n"),
			want:  []byte("\nhello\nthere\n. hi\n. foo\nbar\n"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: NormaliseNewlines is called.
			result := NormaliseNewlines(tc.input)

			prefix := fmt.Sprintf(
				"%s\nNormaliseNewlines(%q)",
				packageName, tc.input,
			)

			// THEN: the newlines are normalised correctly.
			if got, want := string(result), string(tc.want); got != want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}
		})
	}
}
