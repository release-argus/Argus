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

package command

import (
	"testing"
	"time"
)

func TestController_CopyFailsFrom(t *testing.T) {
	// GIVEN a Controller with fails and a Controller to copy them to
	tests := map[string]struct {
		from         *Controller
		to           *Controller
		fails        *[]*bool
		nextRunnable []time.Time
	}{
		"both nil": {
			from:  nil,
			to:    nil,
			fails: nil,
		},
		"from nil": {
			from:  nil,
			to:    &Controller{},
			fails: nil,
		},
		"to nil": {
			from:  &Controller{},
			to:    nil,
			fails: nil,
		},
		"doesn't copy if no commands": {
			from: &Controller{
				Failed: &[]*bool{
					boolPtr(true),
					boolPtr(false),
					nil}},
			to:    &Controller{},
			fails: nil,
		},
		"doesn't copy to new commands": {
			from: &Controller{
				Command: &Slice{
					{"ls", "-la"}},
				Failed: &[]*bool{
					boolPtr(true)},
				NextRunnable: []time.Time{
					time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)}},
			to: &Controller{
				Command: &Slice{
					{"ls", "-lah"}}},
			fails: &[]*bool{
				nil},
		},
		"does copy to retained commands": {
			from: &Controller{
				Command: &Slice{
					{"ls", "-lah"}},
				Failed: &[]*bool{
					boolPtr(true)},
				NextRunnable: []time.Time{
					time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)}},
			to: &Controller{
				Command: &Slice{
					{"ls", "-lah"}}},
			fails: &[]*bool{
				boolPtr(true)},
			nextRunnable: []time.Time{
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		"does copy to reordered retained commands": {
			from: &Controller{
				Command: &Slice{
					{"false"},
					{"ls", "-lah"}},
				Failed: &[]*bool{
					boolPtr(true),
					boolPtr(false)},
				NextRunnable: []time.Time{
					time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)}},
			to: &Controller{
				Command: &Slice{
					{"ls", "-lah"}}},
			fails: &[]*bool{
				boolPtr(false)},
			nextRunnable: []time.Time{
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CopyFailsFrom is called
			tc.to.CopyFailsFrom(tc.from)

			// THEN the fails aren't copied to a nil Controller
			if tc.fails == nil && (tc.to == nil || tc.to.Failed == nil) {
				return
			} else if tc.to == nil {
				t.Fatalf("expected to.fails to be %v, but got %v tc.to",
					tc.fails, tc.to)
			}
			if tc.to.Failed == nil {
				t.Fatalf("expected to.fails to be %v, but got %v tc.to.Failed",
					tc.fails, tc.to.Failed)
			} else if len(*tc.to.Failed) != len(*tc.fails) {
				t.Fatalf("expected fails to be %v, but got %v",
					tc.fails, tc.to.Failed)
			}
			// AND the matching fails are copied to the Controller
			for i := range *tc.fails {
				if stringifyPointer((*tc.fails)[i]) != stringifyPointer((*tc.to.Failed)[i]) {
					t.Errorf("Fail %d: expected %q, got %q",
						i,
						stringifyPointer((*tc.fails)[i]),
						stringifyPointer((*tc.to.Failed)[i]))
				}
			}
			// AND the next_runnables are copied to the Controller
			for i := range tc.nextRunnable {
				if (tc.nextRunnable)[i] != (tc.to.NextRunnable)[i] {
					t.Errorf("Fail %d: expected %q, got %q",
						i,
						tc.nextRunnable[i],
						tc.to.NextRunnable[i])
				}
			}
		})
	}
}
