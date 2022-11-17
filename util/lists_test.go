// Copyright [2022] [Argus]
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

func TestSwap(t *testing.T) {
	// GIVEN a set of lists
	tests := map[string]struct {
		had    []int
		want   []int
		aStart int
		aEnd   int
		bStart int
		bEnd   int
	}{
		"handles []int": {
			had:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			want:   []int{0, 7, 8, 9, 2, 3, 4, 5, 6, 1},
			aStart: 7, aEnd: 9,
			bStart: 1, bEnd: 1,
		},
		"swap singl element": {
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
		name, tc := name, tc
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

func TestRemoveIndex(t *testing.T) {
	// GIVEN a set of lists
	tests := map[string]struct {
		had     []int
		want    []int
		wantStr []string
		index   int
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
		name, tc := name, tc
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

func TestGetIndentation(t *testing.T) {
	// GIVEN a set of strings with varying indentation
	tests := map[string]struct {
		text       string
		indentSize int
		want       string
	}{
		"no indent": {
			text:       "foo: bar",
			indentSize: 2,
			want:       "",
		},
		"indent 4, indent size 4": {
			text:       "    foo: bar",
			indentSize: 4,
			want:       "    ",
		},
		"indent 4, indent size 2": {
			text:       "    foo: bar",
			indentSize: 2,
			want:       "    ",
		},
		"indent 3, indent size 2": {
			text:       "   foo: bar",
			indentSize: 2,
			want:       "  ",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// WHEN GetIndentation is called on a string
			got := GetIndentation(tc.text, uint8(tc.indentSize))

			// THEN the expected indentation is returned
			if got != tc.want {
				t.Fatalf("want:%q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}
