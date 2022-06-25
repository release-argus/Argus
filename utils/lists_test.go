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

package utils

import (
	"testing"
)

func TestSwapMoveMatchingSize(t *testing.T) {
	// GIVEN a list of comparable
	lst := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	// WHEN Swap is called on it with matching size swaps
	Swap(&lst, 0, 1, 8, 9)
	wantLst := []int{8, 9, 2, 3, 4, 5, 6, 7, 0, 1}

	// THEN the Swap is successfuly
	for i := range lst {
		if lst[i] != wantLst[i] {
			t.Errorf(`Swap got %v, want %v`, lst, wantLst)
		}
	}
	if len(lst) != len(wantLst) {
		t.Errorf(`Swap added/removed elements! Got %v, want %v`, lst, wantLst)
	}
}

func TestSwapMoveMoreOnLeft(t *testing.T) {
	// GIVEN a list of comparable
	lst := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	// WHEN Swap is called with more elements moving from the left
	Swap(&lst, 5, 9, 14, 15)
	wantLst := []int{0, 1, 2, 3, 4, 14, 15, 10, 11, 12, 13, 5, 6, 7, 8, 9, 16, 17, 18, 19, 20}

	// THEN the Swap is successful
	for i := range lst {
		if lst[i] != wantLst[i] {
			t.Errorf(`Swap got %v, want %v`, lst, wantLst)
		}
	}
	if len(lst) != len(wantLst) {
		t.Errorf(`Swap added/removed elements! Got %v, want %v`, lst, wantLst)
	}
}

func TestSwapMoveMoreOnRight(t *testing.T) {
	// GIVEN a list of comparable
	lst := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	// WHEN Swap is called with more elements moving from the right
	Swap(&lst, 7, 7, 14, 15)
	wantLst := []int{0, 1, 2, 3, 4, 5, 6, 14, 15, 8, 9, 10, 11, 12, 13, 7, 16, 17, 18, 19, 20}

	// THEN the Swap is successful
	for i := range lst {
		if lst[i] != wantLst[i] {
			t.Errorf(`Swap got %v, want %v`, lst, wantLst)
		}
	}
	if len(lst) != len(wantLst) {
		t.Errorf(`Swap added/removed elements! Got %v, want %v`, lst, wantLst)
	}
}

func TestSwapIndicesOppositeWayRound(t *testing.T) {
	// GIVEN a list of comparable
	lst := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	// WHEN Swap is called with more elements moving from the left
	Swap(&lst, 14, 15, 5, 9)
	wantLst := []int{0, 1, 2, 3, 4, 14, 15, 10, 11, 12, 13, 5, 6, 7, 8, 9, 16, 17, 18, 19, 20}

	// THEN the Swap is successful
	for i := range lst {
		if lst[i] != wantLst[i] {
			t.Errorf(`Swap got %v, want %v`, lst, wantLst)
		}
	}
	if len(lst) != len(wantLst) {
		t.Errorf(`Swap added/removed elements! Got %v, want %v`, lst, wantLst)
	}
}

func TestRemoveIndexStart(t *testing.T) {
	// GIVEN a list of comparable
	lstInt := []int{0, 1, 2, 3}

	// WHEN RemoveIndex is called on it with the starting index
	RemoveIndex(&lstInt, 0)
	wantLstInt := []int{1, 2, 3}

	// Then that index is removed
	if len(lstInt) != len(wantLstInt) {
		t.Errorf(`Failed removing first index! Got %v, want %v`, lstInt, wantLstInt)
	}
}

func TestRemoveIndexEnd(t *testing.T) {
	// GIVEN a list of comparable
	lstInt := []int{0, 1, 2, 3}

	// WHEN RemoveIndex is called on it with the ending index
	RemoveIndex(&lstInt, 3)
	wantLstInt := []int{0, 1, 2}

	// Then that index is removed
	if len(lstInt) != len(wantLstInt) {
		t.Errorf(`Failed removing final index! Got %v, want %v`, lstInt, wantLstInt)
	}
}

func TestRemoveIndexMiddle(t *testing.T) {
	// GIVEN a list of comparable
	lstInt := []int{0, 1, 2, 3}

	// WHEN RemoveIndex is called on it with an index not at the start/end
	RemoveIndex(&lstInt, 1)
	wantLstInt := []int{0, 2, 3}

	// Then that index is removed
	if len(lstInt) != len(wantLstInt) {
		t.Errorf(`Failed removing an index not at the edge! Got %v, want %v`, lstInt, wantLstInt)
	}
}

func TestRemoveIndexOutOfRange(t *testing.T) {
	// GIVEN a list of comparable
	lstInt := []int{0, 1, 2, 3}

	// WHEN RemoveIndex is called on it with an index not at the start/end
	RemoveIndex(&lstInt, 10)
	wantLstInt := []int{0, 1, 2, 3}

	// Then that index is removed
	if len(lstInt) != len(wantLstInt) {
		t.Errorf(`Tried to remove an index out of bounds of the array and got %v, want %v`, lstInt, wantLstInt)
	}
}

func TestGetIndentationNone(t *testing.T) {
	// GIVEN a string with no indentation
	str := "foo: bar"

	// WHEN GetIndentation is called on it
	gotIndentation := GetIndentation(str, 2)
	want := ""

	// THEN the returned indentation is correct (none)
	if gotIndentation != want {
		t.Errorf(`GetIndentation gave %v, want match for %q`, gotIndentation, want)
	}
}

func TestGetIndentationIndentSizeOne(t *testing.T) {
	// GIVEN a string with one space indentation
	str := " foo: bar"

	// WHEN GetIndentation is called on it with indentSize 1
	gotIndentation := GetIndentation(str, 1)
	want := " "

	// THEN the returned indentation is correct (one space)
	if gotIndentation != want {
		t.Errorf(`GetIndentation gave %q, want match for %q`, gotIndentation, want)
	}
}
func TestGetIndentationIndentSizeTwo(t *testing.T) {
	// GIVEN a string with less than indentSize an indent
	str := " foo: bar"

	// WHEN GetIndentation is called on it with indentSize 2
	gotIndentation := GetIndentation(str, 2)
	want := ""

	// THEN the returned indentation is none as there's no leading blocks of two spaces
	if gotIndentation != want {
		t.Errorf(`GetIndentation gave %q, want match for %q`, gotIndentation, want)
	}
}
func TestGetIndentationMultipleIndents(t *testing.T) {
	// GIVEN a string with multiple indents
	str := "    foo: bar"

	// WHEN GetIndentation is called on it with indentSize 2
	gotIndentation := GetIndentation(str, 2)
	want := "    "

	// THEN the returned indentation is the correct indentation
	if gotIndentation != want {
		t.Errorf(`GetIndentation gave %v, want match for %q`, gotIndentation, want)
	}
}
