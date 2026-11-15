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
	"slices"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestCopySlice(t *testing.T) {
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

			result := CopySlice(tc.had)

			if !slices.Equal(result, tc.want) {
				t.Fatalf("copy mismatch\ngot:  %v\nwant: %v", result, tc.want)
			}

			if len(result) != 0 && &result[0] == &tc.had[0] {
				t.Errorf(
					"copy aliases the original\ngot:  %p\nwant: NOT %p",
					&result[0], &tc.had[0],
				)
			}
		})
	}
}

func TestSliceReplace(t *testing.T) {
	tests := []struct {
		name      string
		had, want []int
		insert    []int
		index     int
	}{
		{
			name:   "replace at beginning",
			had:    []int{1, 2, 3},
			insert: []int{0},
			index:  0,
			want:   []int{0, 2, 3},
		},
		{
			name:   "replace at end",
			had:    []int{1, 2, 3},
			insert: []int{4},
			index:  2,
			want:   []int{1, 2, 4},
		},
		{
			name:   "replace in middle",
			had:    []int{1, 2, 4, 5},
			insert: []int{3},
			index:  2,
			want:   []int{1, 2, 3, 5},
		},
		{
			name:   "replace with multiple elements",
			had:    []int{1, 2, 5, 6},
			insert: []int{3, 4, 5},
			index:  2,
			want:   []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:   "single into empty list",
			had:    []int{},
			insert: []int{1},
			index:  0,
			want:   []int{1},
		},
		{
			name:   "multiple into empty list",
			had:    []int{},
			insert: []int{1, 2, 3},
			index:  0,
			want:   []int{1, 2, 3},
		},
		{
			name:   "replace at out of range index",
			had:    []int{1, 2, 3},
			insert: []int{4, 5},
			index:  5,
			want:   []int{1, 2, 4, 5},
		},
		{
			name:   "replace at negative index",
			had:    []int{1, 2, 3},
			insert: []int{0},
			index:  -1,
			want:   []int{0, 2, 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := SliceReplace(tc.had, tc.index, tc.insert)

			if testErr := test.AssertSlicesEqual(t, got, tc.want, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestSwapRanges(t *testing.T) {
	tests := []struct {
		name         string
		had, want    []int
		aStart, aEnd int
		bStart, bEnd int
	}{
		{
			name:   "indices wrong way round",
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := SwapRanges(tc.had, tc.aStart, tc.aEnd, tc.bStart, tc.bEnd)

			if testErr := test.AssertSlicesEqual(t, result, tc.want, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestReplaceFirst_String(t *testing.T) {
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

			result := ReplaceFirst(tc.had, tc.replace, tc.with)

			if testErr := test.AssertSlicesEqual(t, result, tc.want, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestRemoveFirst_Int(t *testing.T) {
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

			gotInt := RemoveFirst(tc.hadInt, tc.removeInt)

			if testErr := test.AssertSlicesEqual(t, gotInt, tc.wantInt, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestRemoveFirst_String(t *testing.T) {
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

			got := RemoveFirst(tc.had, tc.remove)

			if testErr := test.AssertSlicesEqual(t, got, tc.want, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestRemoveAt_Int(t *testing.T) {
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

			result := RemoveAt(tc.had, tc.index)

			if testErr := test.AssertSlicesEqual(t, result, tc.want, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestAreSlicesEqual(t *testing.T) {
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

			got := AreSlicesEqual(tc.slice1, tc.slice2)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestFirstNonNilPtr(t *testing.T) {
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
				new("bar"),
			},
			wantIndex: 3,
		},
		{
			name: "2 non-nil pointers",
			pointers: []*string{
				new("foo"),
				nil,
				nil,
				new("bar"),
			},
			wantIndex: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FirstNonNilPtr(tc.pointers...)

			if tc.allNil {
				if got != nil {
					t.Fatalf("mismatch\ngot:  %s\nwant: nil", test.StringifyPtr(got))
				}
				return
			}
			if got != tc.pointers[tc.wantIndex] {
				t.Errorf("mismatch\ngot:  %p\nwant: %p", got, tc.pointers[tc.wantIndex])
			}
		})
	}
}

func TestFirstNonDefault(t *testing.T) {
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

			got := FirstNonDefault(tc.slice...)

			if tc.allDefault {
				want := ""
				if got != want {
					t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, want)
				}
				return
			}
			if got != tc.slice[tc.wantIndex] {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.slice[tc.wantIndex])
			}
		})
	}
}

func TestFirstNonEmptySlice(t *testing.T) {
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

			got := FirstNonEmptySlice(tc.slices...)

			if tc.wantIndex == -1 {
				if got != nil {
					t.Errorf("mismatch\ngot:  %q\nwant: nil", got)
				}
				return
			}
			if want := tc.slices[tc.wantIndex]; got == nil || !AreSlicesEqual(got, want) {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

func TestNormaliseNewlines(t *testing.T) {
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
			input: []byte("hello\rthere"),
			want:  []byte("hello\nthere"),
		},
		{
			name:  "string with multiple mac newlines",
			input: []byte("hello\rthere\r"),
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

			result := NormaliseNewlines(tc.input)

			if got, want := string(result), string(tc.want); got != want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}
