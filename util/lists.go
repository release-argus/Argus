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

package util

import (
	"reflect"
	"strings"
)

func Swap[T comparable](list *[]T, aStart int, aEnd int, bStart int, bEnd int) {
	// Always have the lower index as a
	if aStart > bStart {
		aStart, bStart = bStart, aStart
		aEnd, bEnd = bEnd, aEnd
	}

	aLen := aEnd - aStart + 1
	bLen := bEnd - bStart + 1
	swapper := reflect.Swapper(*list)

	// Direct swaps
	index := 0
	for aStart+index <= aEnd && bStart+index <= bEnd {
		swapper(aStart+index, bStart+index)
		index++
	}

	shiftNumber := bLen - aLen
	if shiftNumber == 0 {
		return
	}

	// how many elements we need to shift
	// Index to start swapping from (only shift `shitNumber` elements)
	var startAt int
	// Whether we're moving right(+)/left(-)
	var direction int
	// ShiftBy `direction` to get to the appropriate position
	var shiftBy int
	// More on the right, so we need to shift some to the left
	if bLen > aLen {
		direction = -1
		startAt = bStart + index
		// shiftBy the <last-direct-swap-on-b> - <last-direct-swap-on-a>
		shiftBy = (bStart + index - 1) - (aStart + index - 1)
	} else {
		// More on the left, so we need to shift some to the right
		direction = 1
		startAt = aEnd
		// shiftBy the <last-direct-swap-on-b> - <aEnd>
		shiftBy = (bStart + index - 1) - (aEnd)
		// Absolute shiftNumber
		shiftNumber *= -1
	}

	loop := 0
	for loop < shiftNumber {
		at := startAt
		// Moving from left to right, so we'll start on the right side and start 1 left each loop
		at -= loop * direction

		shifted := 0
		// Shift `startAt` forward `shiftBy`
		for shifted < shiftBy {
			swapper(at, at+(direction))
			at += direction
			shifted++
		}
		loop++
	}
}

// RemoveIndex from list
func RemoveIndex[T comparable](list *[]T, index int) {
	if index >= len(*list) {
		return
	}

	*list = append((*list)[:index], (*list)[index+1:]...)[:len(*list)-1]
}

// GetIndentation used in line. Test variations on intentationxindentSize
func GetIndentation(line string, indentSize uint8) (indentation string) {
	for {
		if !strings.HasPrefix(line, indentation+strings.Repeat(" ", int(indentSize))) {
			break
		}
		indentation += strings.Repeat(" ", int(indentSize))
	}
	return
}
