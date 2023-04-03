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

	svcstatus "github.com/release-argus/Argus/service/status"
)

func TestController_CopyFailsFrom(t *testing.T) {
	// GIVEN a Controller with fails and a Controller to copy them to
	tests := map[string]struct {
		from             *Controller
		to               *Controller
		fromFails        []*bool
		toFails          []*bool
		fromNextRunnable []time.Time
		toNextRunnable   []time.Time
	}{
		"both nil": {
			from:    nil,
			to:      nil,
			toFails: nil,
		},
		"from nil": {
			from:    nil,
			to:      &Controller{},
			toFails: nil,
		},
		"to nil": {
			from:    &Controller{},
			to:      nil,
			toFails: nil,
		},
		"doesn't copy if no commands": {
			from: &Controller{},
			to:   &Controller{},
			fromFails: []*bool{
				boolPtr(true),
				boolPtr(false),
				nil},
			toFails: nil,
		},
		"doesn't copy to new commands": {
			from: &Controller{
				Command: &Slice{
					{"ls", "-la"}}},
			to: &Controller{
				Command: &Slice{
					{"ls", "-lah"}}},
			fromFails: []*bool{
				boolPtr(true)},
			toFails: []*bool{
				nil},
			fromNextRunnable: []time.Time{
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		"does copy to retained commands": {
			from: &Controller{
				Command: &Slice{
					{"ls", "-lah"}}},
			to: &Controller{
				Command: &Slice{
					{"ls", "-lah"}}},
			fromFails: []*bool{
				boolPtr(true)},
			toFails: []*bool{
				boolPtr(true)},
			fromNextRunnable: []time.Time{
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
			toNextRunnable: []time.Time{
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		"does copy to reordered retained commands": {
			from: &Controller{
				Command: &Slice{
					{"false"},
					{"ls", "-lah"}}},
			to: &Controller{
				Command: &Slice{
					{"ls", "-lah"}}},
			fromFails: []*bool{
				boolPtr(true),
				boolPtr(false)},
			toFails: []*bool{
				boolPtr(false)},
			fromNextRunnable: []time.Time{
				time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC),
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
			toNextRunnable: []time.Time{
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.from != nil && tc.from.Command != nil {
				tc.from.Init(
					&svcstatus.Status{},
					tc.from.Command,
					nil,
					nil)
				for k, v := range tc.fromFails {
					if v != nil {
						tc.from.ServiceStatus.Fails.Command.Set(k, *v)
					}
				}
				for i, v := range tc.fromNextRunnable {
					tc.from.NextRunnable[i] = v
				}
			}
			if tc.to != nil && tc.to.Command != nil {
				tc.to.Init(
					&svcstatus.Status{},
					tc.to.Command,
					nil,
					nil)
			}

			// WHEN CopyFailsFrom is called
			tc.to.CopyFailsFrom(tc.from)

			// THEN the fails aren't copied to a nil Controller
			if tc.toFails == nil && (tc.to == nil || tc.to.Failed == nil) {
				return
			} else if tc.to == nil {
				t.Fatalf("expected to.fails to be %v, but got %v tc.to",
					tc.toFails, tc.to)
			}
			if tc.to.Failed.Length() != len(tc.toFails) {
				t.Fatalf("expected fails to be %v, but got %v",
					tc.toFails, tc.to.Failed)
			}
			// AND the matching fails are copied to the Controller
			for i := range tc.toFails {
				if stringifyPointer(tc.toFails[i]) != stringifyPointer(tc.to.Failed.Get(i)) {
					t.Errorf("Fail %d: expected %q, got %q",
						i,
						stringifyPointer(tc.toFails[i]),
						stringifyPointer(tc.to.Failed.Get(i)))
				}
			}
			// AND the next_runnables are copied to the Controller
			for i := range tc.toNextRunnable {
				if (tc.toNextRunnable)[i] != (tc.to.NextRunnable)[i] {
					t.Errorf("Fail %d: expected %q, got %q",
						i,
						tc.toNextRunnable[i],
						tc.to.NextRunnable[i])
				}
			}
		})
	}
}
