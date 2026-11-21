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

package status

import (
	"strconv"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestFailsBase_Init(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{name: "size=0", size: 0},
		{name: "size=1", size: 1},
		{name: "size=2", size: 2},
		{name: "size=3", size: 3},
		{name: "size=4", size: 4},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			var failsShoutrrr FailsShoutrrr
			var failsWebhook FailsWebhook

			failsCommand.Init(tc.size)
			failsShoutrrr.Init(tc.size)
			failsWebhook.Init(tc.size)

			if got := len(failsCommand.fails); got != tc.size {
				t.Errorf("len(FailsCommand.fails) mismatch\ngot:  %d\nwant: %d", got, tc.size)
			}
			if failsShoutrrr.fails == nil {
				t.Error("FailsShoutrrr.fails mismatch\ngot:  nil\nwant: non-nil")
			}
			if failsWebhook.fails == nil {
				t.Error("FailsWebhook.fails mismatch\ngot:  nil\nwant: non-nil")
			}
		})
	}
}

func TestFailsBase_SetAndGet(t *testing.T) {
	tests := []struct {
		name       string
		size       int
		setAtArray map[int]*bool
		setAtMap   map[string]*bool
	}{
		{
			name:       "can add to empty map",
			size:       0,
			setAtArray: map[int]*bool{},
			setAtMap: map[string]*bool{
				"test": new(true),
			},
		},
		{
			name: "can add to non-empty map or edit array",
			size: 3,
			setAtArray: map[int]*bool{
				0: new(true),
				1: new(false),
				2: new(true),
			},
			setAtMap: map[string]*bool{
				"bish": new(true),
				"bash": new(false),
				"bosh": new(true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			failsCommand.Init(tc.size)
			var failsShoutrrr FailsShoutrrr
			failsShoutrrr.Init(tc.size)
			var failsWebhook FailsWebhook
			failsWebhook.Init(tc.size)
			for i := range tc.setAtArray {
				got := failsCommand.Get(i)
				if got != nil {
					t.Errorf(
						"after Init: FailsCommand.Get(%d) mismatch\ngot:  %s\nwant: nil",
						i, test.StringifyPtr(got),
					)
				}
			}
			for k := range tc.setAtMap {
				got := failsShoutrrr.Get(k)
				if got != nil {
					t.Errorf(
						"after Init: FailsShoutrrr.Get(%q) mismatch\ngot:  %s\nwant: nil",
						k, test.StringifyPtr(got),
					)
				}
				got = failsWebhook.Get(k)
				if got != nil {
					t.Errorf(
						"after Init: FailsWebhook.Get(%q) mismatch\ngot:  %s\nwant: nil",
						k, test.StringifyPtr(got),
					)
				}
			}

			for i, v := range tc.setAtArray {
				failsCommand.Set(i, *v)
			}
			for k, v := range tc.setAtMap {
				failsShoutrrr.Set(k, v)
				failsWebhook.Set(k, v)
			}

			for i, v := range tc.setAtArray {
				got := failsCommand.Get(i)
				gotStr := test.StringifyPtr(got)
				wantStr := test.StringifyPtr(v)
				if got == nil || *got != *v {
					t.Errorf(
						"after Set: FailsCommand.Get(%d) mismatch\ngot:  %s\nwant: %s",
						i, gotStr, wantStr,
					)
				}
			}
			for k, v := range tc.setAtMap {
				got := failsShoutrrr.Get(k)
				gotStr := test.StringifyPtr(got)
				wantStr := test.StringifyPtr(v)
				if got == nil || *got != *v {
					t.Errorf(
						"after Set: FailsShoutrrr.Get(%q) mismatch\ngot:  %s\nwant: %s",
						k, gotStr, wantStr,
					)
				}
				got = failsWebhook.Get(k)
				gotStr = test.StringifyPtr(got)
				if got == nil || *got != *v {
					t.Errorf(
						"after Set: FailsWebhook.Get(%q) mismatch\ngot:  %s\nwant: %s",
						k, gotStr, wantStr,
					)
				}
			}
		})
	}
}

func TestFailsBase_AllPassed(t *testing.T) {
	tests := []struct {
		name  string
		fails map[int]*bool
		want  bool
	}{
		{
			name:  "empty",
			fails: map[int]*bool{},
			want:  true,
		},
		{
			name: "all true (failed)",
			fails: map[int]*bool{
				0: new(true),
				1: new(true),
				2: new(true),
			},
			want: false,
		},
		{
			name: "all false (passed)",
			fails: map[int]*bool{
				0: new(false),
				1: new(false),
				2: new(false),
			},
			want: true,
		},
		{
			name: "all nil (not run)",
			fails: map[int]*bool{
				0: nil,
				1: nil,
				2: nil,
			},
			want: false,
		},
		{
			name: "mixed",
			fails: map[int]*bool{
				0: new(true),
				1: new(false),
				2: nil,
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			failsCommand.Init(len(tc.fails))
			var failsShoutrrr FailsShoutrrr
			failsShoutrrr.Init(len(tc.fails))
			var failsWebhook FailsWebhook
			failsWebhook.Init(len(tc.fails))
			for i, v := range tc.fails {
				if v != nil {
					failsCommand.Set(i, *v)
				}
				iStr := strconv.Itoa(i)
				failsShoutrrr.Set(iStr, v)
				failsWebhook.Set(iStr, v)
			}

			gotC := failsCommand.AllPassed()
			gotS := failsShoutrrr.AllPassed()
			gotWH := failsWebhook.AllPassed()

			if gotC != tc.want {
				t.Errorf("FailsCommand.AllPassed() mismatch\ngot:  %t\nwant: %t", gotC, tc.want)
			}
			if gotS != tc.want {
				t.Errorf("FailsShoutrrr.AllPassed() mismatch\ngot:  %t\nwant: %t", gotS, tc.want)
			}
			if gotWH != tc.want {
				t.Errorf("FailsWebhook.AllPassed() mismatch\ngot:  %t\nwant: %t", gotWH, tc.want)
			}
		})
	}
}

func TestFailsBase_Reset(t *testing.T) {
	tests := []struct {
		name  string
		fails map[int]*bool
	}{
		{
			name:  "empty",
			fails: map[int]*bool{},
		},
		{
			name: "all true (failed)",
			fails: map[int]*bool{
				0: new(true),
				1: new(true),
				2: new(true),
			},
		},
		{
			name: "all false (passed)",
			fails: map[int]*bool{
				0: new(false),
				1: new(false),
				2: new(false),
			},
		},
		{
			name: "all nil (not run)",
			fails: map[int]*bool{
				0: nil,
				1: nil,
				2: nil,
			},
		},
		{
			name: "mixed",
			fails: map[int]*bool{
				0: new(true),
				1: new(false),
				2: nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			failsCommand.Init(len(tc.fails))
			var failsShoutrrr FailsShoutrrr
			failsShoutrrr.Init(len(tc.fails))
			var failsWebhook FailsWebhook
			failsWebhook.Init(len(tc.fails))
			for i, v := range tc.fails {
				if v != nil {
					failsCommand.Set(i, *v)
				}
				iStr := strconv.Itoa(i)
				failsShoutrrr.Set(iStr, v)
				failsWebhook.Set(iStr, v)
			}

			failsCommand.Reset()
			failsShoutrrr.Reset()
			failsWebhook.Reset()

			for i := range tc.fails {
				got := failsCommand.Get(i)
				if got != nil {
					t.Errorf(
						"FailsCommand.Get(%d) mismatch\ngot:  %s\nwant: nil",
						i, test.StringifyPtr(got),
					)
				}
				iStr := strconv.Itoa(i)
				got = failsShoutrrr.Get(iStr)
				if got != nil {
					t.Errorf(
						"FailsShoutrrr.Get(%q) mismatch\ngot:  %s\nwant: nil",
						iStr, test.StringifyPtr(got),
					)
				}
				got = failsWebhook.Get(iStr)
				if got != nil {
					t.Errorf(
						"FailsWebhook.Get(%q) mismatch\ngot:  %s\nwant: nil",
						iStr, test.StringifyPtr(got),
					)
				}
			}
		})
	}
}

func TestFailsBase_Length(t *testing.T) {
	tests := []struct {
		name       string
		size       int
		setAtArray map[int]*bool
		setAtMap   map[string]*bool
	}{
		{
			name:       "can add to empty map",
			size:       0,
			setAtArray: map[int]*bool{},
			setAtMap: map[string]*bool{
				"test": new(true),
			},
		},
		{
			name: "can add to non-empty map or edit array",
			size: 3,
			setAtArray: map[int]*bool{
				0: new(true),
				1: new(false),
				2: new(true),
			},
			setAtMap: map[string]*bool{
				"bish": new(true),
				"bash": new(false),
				"bosh": new(true),
			},
		},
		{
			name: "map length counts its elements, array length is its make size",
			size: 3,
			setAtArray: map[int]*bool{
				0: new(true),
				1: new(false),
			},
			setAtMap: map[string]*bool{
				"bish": new(true),
				"bash": new(false),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			failsCommand.Init(tc.size)
			var failsShoutrrr FailsShoutrrr
			failsShoutrrr.Init(tc.size)
			var failsWebhook FailsWebhook
			failsWebhook.Init(tc.size)
			for i, v := range tc.setAtArray {
				if v != nil {
					failsCommand.Set(i, *v)
				}
			}
			for k, v := range tc.setAtMap {
				failsShoutrrr.Set(k, v)
				failsWebhook.Set(k, v)
			}

			lengthC := failsCommand.Length()
			lengthS := failsShoutrrr.Length()
			lengthWH := failsWebhook.Length()

			if lengthC != tc.size {
				t.Errorf("FailsCommand.Length() mismatch\ngot:  %d\nwant: %d", lengthC, tc.size)
			}
			if lengthS != len(tc.setAtMap) {
				t.Errorf(
					"FailsShoutrrr.Length() mismatch\ngot:  %d\nwant: %d",
					lengthS, len(tc.setAtMap),
				)
			}
			if lengthWH != len(tc.setAtMap) {
				t.Errorf(
					"FailsWebhook.Length() mismatch\ngot:  %d\nwant: %d",
					lengthWH, len(tc.setAtMap),
				)
			}
		})
	}
}

func TestFailsShoutrrr_Copy(t *testing.T) {
	tests := []struct {
		name string
		// If false, the receiver's fails map is left nil.
		initOriginal bool
		original     map[string]*bool
		// Mutations to verify that Copy doesn't alias the underlying map.
		mutateOriginal map[string]*bool
		mutateCopy     map[string]*bool
	}{
		{
			name:           "uninitialised fails map",
			initOriginal:   false,
			original:       nil,
			mutateOriginal: map[string]*bool{"a": new(true)},
			mutateCopy:     map[string]*bool{"a": new(false)},
		},
		{
			name:           "empty receiver",
			initOriginal:   true,
			original:       map[string]*bool{},
			mutateOriginal: map[string]*bool{"a": new(true)},
			mutateCopy:     map[string]*bool{"b": new(false)},
		},
		{
			name:         "filled receiver",
			initOriginal: true,
			original: map[string]*bool{
				"foo": new(false),
				"bar": nil,
				"baz": new(true),
			},
			mutateOriginal: map[string]*bool{
				"foo": new(true),  // Replace a value.
				"tap": new(false), // Add a new entry.
			},
			mutateCopy: map[string]*bool{
				"foo": new(false), // Replace a value.
				"qux": nil,        // Add a nil entry.
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orig := &FailsShoutrrr{}
			if tc.initOriginal {
				orig.Init(len(tc.original))
				for k, v := range tc.original {
					orig.Set(k, v)
				}
			}

			inputCopy := orig.Copy()

			initialCopyWant := tc.original
			if !tc.initOriginal {
				initialCopyWant = map[string]*bool{}
			}
			if testErr := test.AssertMapEqual(
				t,
				failValues(inputCopy.fails),
				failValues(initialCopyWant),
				"FailsShoutrrr.Copy().fails",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}

			if testErr := test.AssertMapEqual(
				t,
				failValues(orig.fails),
				failValues(initialCopyWant),
				"FailsShoutrrr.fails",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}

			if !tc.initOriginal {
				orig.Init(len(tc.mutateOriginal))
			}
			for k, v := range tc.mutateOriginal {
				orig.Set(k, v)
			}

			if testErr := test.AssertMapEqual(
				t,
				failValues(inputCopy.fails),
				failValues(initialCopyWant),
				"FailsShoutrrr.Copy().fails",
			); testErr != nil {
				t.Errorf("after original mutated: %v", testErr)
			}

			for k, v := range tc.mutateCopy {
				inputCopy.Set(k, v)
			}

			wantMutatedInput := mergedMap(tc.original, tc.mutateCopy)
			if testErr := test.AssertMapEqual(
				t,
				failValues(inputCopy.fails),
				failValues(wantMutatedInput),
				"FailsShoutrrr.Copy().fails",
			); testErr != nil {
				t.Errorf("after copy mutated: %v", testErr)
			}
		})
	}
}

func TestFailsWebhook_Copy(t *testing.T) {
	time2021 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	time2022 := time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)
	time2023 := time.Date(2023, 3, 3, 0, 0, 0, 0, time.UTC)
	time2024 := time.Date(2024, 4, 4, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		// If false, the receiver's fails map is left nil.
		initOriginal         bool
		originalFails        map[string]*bool
		originalNextRunnable map[string]time.Time
		// Mutations to verify that Copy doesn't alias the underlying map.
		mutateOriginalFails        map[string]*bool
		mutateOriginalNextRunnable map[string]time.Time
		mutateCopyFails            map[string]*bool
		mutateCopyNextRunnable     map[string]time.Time
	}{
		{
			name:                       "uninitialised fails map",
			initOriginal:               false,
			originalFails:              nil,
			mutateOriginalFails:        map[string]*bool{"a": new(true)},
			mutateOriginalNextRunnable: map[string]time.Time{"a": time2022},
			mutateCopyFails:            map[string]*bool{"a": new(false)},
			mutateCopyNextRunnable:     map[string]time.Time{"b": time2023},
		},
		{
			name:                       "empty receiver",
			initOriginal:               true,
			originalFails:              map[string]*bool{},
			mutateOriginalFails:        map[string]*bool{"a": new(true)},
			mutateOriginalNextRunnable: map[string]time.Time{"a": time2022},
			mutateCopyFails:            map[string]*bool{"b": new(false)},
			mutateCopyNextRunnable:     map[string]time.Time{"b": time2023},
		},
		{
			name:         "filled receiver",
			initOriginal: true,
			originalFails: map[string]*bool{
				"foo": new(false),
				"bar": nil,
				"baz": new(true),
			},
			originalNextRunnable: map[string]time.Time{
				"foo": time2021,
				"bar": time2022,
			},
			mutateOriginalFails: map[string]*bool{
				"foo": new(true),  // Replace a value.
				"tap": new(false), // Add a new entry.
			},
			mutateOriginalNextRunnable: map[string]time.Time{
				"foo": time2023,
				"baz": time2024,
			},
			mutateCopyFails: map[string]*bool{
				"foo": new(false), // Replace a value.
				"qux": nil,        // Add a nil entry.
			},
			mutateCopyNextRunnable: map[string]time.Time{
				"foo": time2024,
				"qux": time2023,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orig := &FailsWebhook{}
			if tc.initOriginal {
				orig.Init(len(tc.originalFails))
				for k, v := range tc.originalFails {
					orig.Set(k, v)
				}
				for k, v := range tc.originalNextRunnable {
					orig.SetNextRunnable(k, v)
				}
			}

			inputCopy := orig.Copy()

			initialCopyWant := tc.originalFails
			if !tc.initOriginal {
				initialCopyWant = map[string]*bool{}
			}
			if testErr := test.AssertMapEqual(
				t,
				failValues(inputCopy.fails),
				failValues(initialCopyWant),
				"FailsWebhook.Copy().fails",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}

			initialNextRunnableWant := tc.originalNextRunnable
			if !tc.initOriginal {
				initialNextRunnableWant = map[string]time.Time{}
			}
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.nextRunnable,
				initialNextRunnableWant,
				"FailsWebhook.Copy().nextRunnable",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}

			if !tc.initOriginal {
				orig.Init(len(tc.mutateOriginalFails))
			}
			for k, v := range tc.mutateOriginalFails {
				orig.Set(k, v)
			}

			if testErr := test.AssertMapEqual(
				t,
				failValues(inputCopy.fails),
				failValues(initialCopyWant),
				"FailsWebhook.Copy().fails",
			); testErr != nil {
				t.Errorf("after original mutated: %v", testErr)
			}

			for k, v := range tc.mutateOriginalNextRunnable {
				orig.SetNextRunnable(k, v)
			}

			if testErr := test.AssertMapEqual(
				t,
				inputCopy.nextRunnable,
				initialNextRunnableWant,
				"FailsWebhook.Copy().nextRunnable",
			); testErr != nil {
				t.Errorf("after original mutated: %v", testErr)
			}

			for k, v := range tc.mutateCopyFails {
				inputCopy.Set(k, v)
			}

			wantMutatedFails := mergedMap(initialCopyWant, tc.mutateCopyFails)
			if testErr := test.AssertMapEqual(
				t,
				failValues(inputCopy.fails),
				failValues(wantMutatedFails),
				"FailsWebhook.Copy().fails",
			); testErr != nil {
				t.Errorf("after copy mutated: %v", testErr)
			}

			for k, v := range tc.mutateCopyNextRunnable {
				inputCopy.SetNextRunnable(k, v)
			}

			wantMutatedNextRunnable := mergedMap(initialNextRunnableWant, tc.mutateCopyNextRunnable)
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.nextRunnable,
				wantMutatedNextRunnable,
				"FailsWebhook.Copy().nextRunnable",
			); testErr != nil {
				t.Errorf("after copy mutated: %v", testErr)
			}

			wantOrigFails := mergedMap(initialCopyWant, tc.mutateOriginalFails)
			if testErr := test.AssertMapEqual(
				t,
				failValues(orig.fails),
				failValues(wantOrigFails),
				"FailsWebhook.fails",
			); testErr != nil {
				t.Errorf("after copy mutated: %v", testErr)
			}

			wantOrigNextRunnable := mergedMap(initialNextRunnableWant, tc.mutateOriginalNextRunnable)
			if testErr := test.AssertMapEqual(
				t,
				orig.nextRunnable,
				wantOrigNextRunnable,
				"FailsWebhook.nextRunnable",
			); testErr != nil {
				t.Errorf("after copy mutated: %v", testErr)
			}
		})
	}
}

func TestFailsWebhook_SetAndGetNextRunnable(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		setAtMap map[string]time.Time
	}{
		{
			name: "can add to empty map",
			size: 0,
			setAtMap: map[string]time.Time{
				"test": time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "can add to non-empty map",
			size: 3,
			setAtMap: map[string]time.Time{
				"bish": time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				"bash": time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				"bosh": time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var failsWebhook FailsWebhook
			failsWebhook.Init(tc.size)
			for k := range tc.setAtMap {
				if got := failsWebhook.NextRunnable(k); !got.IsZero() {
					t.Errorf(
						"after Init: FailsWebhook.NextRunnable(%q) mismatch\ngot:  %s\nwant: %s",
						k, got, time.Time{},
					)
				}
			}

			for k, v := range tc.setAtMap {
				failsWebhook.SetNextRunnable(k, v)
			}

			for k, v := range tc.setAtMap {
				got := failsWebhook.NextRunnable(k)
				if !got.Equal(v) {
					t.Errorf(
						"after SetNextRunnable: FailsWebhook.NextRunnable(%q) mismatch\ngot:  %s\nwant: %s",
						k, got, v,
					)
				}
			}
		})
	}
}

func TestFails_String(t *testing.T) {
	tests := []struct {
		name                        string
		commandFails                []*bool
		shoutrrrFails, webhookFails map[string]*bool
		want                        string
	}{
		{
			name:          "empty fails",
			commandFails:  []*bool{},
			shoutrrrFails: map[string]*bool{},
			webhookFails:  map[string]*bool{},
			want:          "",
		},
		{
			name: "no fails",
			commandFails: []*bool{
				nil, new(false),
			},
			shoutrrrFails: map[string]*bool{
				"bar": new(false),
				"foo": nil,
			},
			webhookFails: map[string]*bool{
				"bar": nil,
				"foo": new(false),
			},
			want: test.TrimYAML(`
				shoutrrr:
					bar: false
					foo: nil
				command:
					- 0: nil
					- 1: false
				webhook:
					bar: nil
					foo: false
			`),
		},
		{
			name: "only shoutrrr",
			shoutrrrFails: map[string]*bool{
				"bash": new(false),
				"bish": nil,
				"bosh": new(true),
			},
			want: test.TrimYAML(`
				shoutrrr:
					bash: false
					bish: nil
					bosh: true
			`),
		},
		{
			name: "only command",
			commandFails: []*bool{
				nil,
				new(false),
				new(true),
			},
			want: test.TrimYAML(`
				command:
					- 0: nil
					- 1: false
					- 2: true
			`),
		},
		{
			name: "only webhook",
			webhookFails: map[string]*bool{
				"bash": new(true),
				"bish": new(false),
				"bosh": nil,
			},
			want: test.TrimYAML(`
				webhook:
					bash: true
					bish: false
					bosh: nil
			`),
		},
		{
			name: "all",
			shoutrrrFails: map[string]*bool{
				"bash": new(false),
				"bish": new(true),
				"bosh": nil,
			},
			commandFails: []*bool{
				nil,
				new(false),
				new(true),
			},
			webhookFails: map[string]*bool{
				"bash": new(false),
				"bish": nil,
				"bosh": new(true),
			},
			want: test.TrimYAML(`
				shoutrrr:
					bash: false
					bish: true
					bosh: nil
				command:
					- 0: nil
					- 1: false
					- 2: true
				webhook:
					bash: false
					bish: nil
					bosh: true
			`),
		},
		{
			name: "maps are alphabetical",
			shoutrrrFails: map[string]*bool{
				"bish": new(true),
				"bash": new(true),
				"bosh": new(true),
			},
			commandFails: []*bool{
				nil,
				new(true),
				new(false),
			},
			webhookFails: map[string]*bool{
				"zip": new(true),
				"zap": new(true),
				"zop": new(true),
			},
			want: test.TrimYAML(`
				shoutrrr:
					bash: true
					bish: true
					bosh: true
				command:
					- 0: nil
					- 1: true
					- 2: false
				webhook:
					zap: true
					zip: true
					zop: true
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fails := Fails{}
			fails.Command.Init(len(tc.commandFails))
			for k, v := range tc.commandFails {
				if v != nil {
					fails.Command.Set(k, *v)
				}
			}
			fails.Shoutrrr.Init(len(tc.shoutrrrFails))
			for k, v := range tc.shoutrrrFails {
				fails.Shoutrrr.Set(k, v)
			}
			fails.Webhook.Init(len(tc.webhookFails))
			for k, v := range tc.webhookFails {
				fails.Webhook.Set(k, v)
			}

			got := fails.String("")

			if got != tc.want {
				t.Errorf("Fails.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestFails_Copy(t *testing.T) {
	time2023 := time.Date(2023, 3, 3, 0, 0, 0, 0, time.UTC)
	time2021 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	time2022 := time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name                               string
		fromCommandFails, toCommandFails   []*bool
		fromShoutrrrFails, toShoutrrrFails map[string]*bool
		fromWebhookFails, toWebhookFails   map[string]*bool
		fromNextRunnable, toNextRunnable   map[string]time.Time
	}{
		{
			name:              "copy empty fails",
			fromCommandFails:  []*bool{},
			toCommandFails:    []*bool{},
			fromShoutrrrFails: map[string]*bool{},
			toShoutrrrFails:   map[string]*bool{},
			fromWebhookFails:  map[string]*bool{},
			toWebhookFails:    map[string]*bool{},
		},
		{
			name: "copy nextRunnable/replaces stale entries",
			fromWebhookFails: map[string]*bool{
				"foo": new(true),
			},
			fromNextRunnable: map[string]time.Time{
				"foo": time2021,
				"bar": time2022,
			},
			toNextRunnable: map[string]time.Time{
				"baz": time2023,
			},
		},
		{
			name: "copy non-empty fails",
			fromCommandFails: []*bool{
				nil,
				new(false),
				new(true),
			},
			toCommandFails: []*bool{
				new(true),
				nil,
				new(false),
			},
			fromShoutrrrFails: map[string]*bool{
				"foo": new(false),
				"bar": new(true),
			},
			toShoutrrrFails: map[string]*bool{
				"baz": new(true),
			},
			fromWebhookFails: map[string]*bool{
				"foo": new(true),
				"bar": new(false),
			},
			toWebhookFails: map[string]*bool{
				"baz": new(false),
			},
		},
		{
			name: "copy to smaller fails",
			fromCommandFails: []*bool{
				nil,
				new(false),
				new(true),
			},
			toCommandFails: []*bool{
				new(true),
				nil,
			},
			fromShoutrrrFails: map[string]*bool{
				"foo": new(false),
				"bar": new(true),
			},
			toShoutrrrFails: map[string]*bool{
				"baz": new(true),
			},
			fromWebhookFails: map[string]*bool{
				"foo": new(true),
				"bar": new(false),
			},
			toWebhookFails: map[string]*bool{
				"baz": new(false),
			},
		},
		{
			name: "copy to larger fails",
			fromCommandFails: []*bool{
				nil,
				new(false),
				new(true),
			},
			toCommandFails: []*bool{
				new(true),
				nil,
				new(false),
				nil,
			},
			fromShoutrrrFails: map[string]*bool{
				"foo": new(false),
				"bar": new(true),
			},
			toShoutrrrFails: map[string]*bool{
				"baz":  new(true),
				"bosh": nil,
			},
			fromWebhookFails: map[string]*bool{
				"foo": new(true),
				"bar": new(false),
			},
			toWebhookFails: map[string]*bool{
				"baz":  new(false),
				"bosh": nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			from := Fails{}
			from.Command.Init(len(tc.fromCommandFails))
			for k, v := range tc.fromCommandFails {
				if v != nil {
					from.Command.Set(k, *v)
				}
			}
			from.Shoutrrr.Init(len(tc.fromShoutrrrFails))
			for k, v := range tc.fromShoutrrrFails {
				from.Shoutrrr.Set(k, v)
			}
			from.Webhook.Init(len(tc.fromWebhookFails))
			for k, v := range tc.fromWebhookFails {
				from.Webhook.Set(k, v)
			}

			to := Fails{}
			to.Command.Init(len(tc.toCommandFails))
			for k, v := range tc.toCommandFails {
				if v != nil {
					to.Command.Set(k, *v)
				}
			}
			to.Shoutrrr.Init(len(tc.toShoutrrrFails))
			for k, v := range tc.toShoutrrrFails {
				to.Shoutrrr.Set(k, v)
			}
			to.Webhook.Init(len(tc.toWebhookFails))
			for k, v := range tc.toWebhookFails {
				to.Webhook.Set(k, v)
			}

			from.Webhook.nextRunnable = util.CopyMap(tc.fromNextRunnable)
			to.Webhook.nextRunnable = util.CopyMap(tc.toNextRunnable)

			to.Copy(&from)

			if testErr := test.AssertSlicesEqual(
				t,
				failSliceValues(to.Command.fails),
				failSliceValues(from.Command.fails),
				"Fails.Command",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				failValues(to.Shoutrrr.fails),
				failValues(from.Shoutrrr.fails),
				"Fails.Shoutrrr",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				failValues(to.Webhook.fails),
				failValues(from.Webhook.fails),
				"Fails.Webhook",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}
			if testErr := test.AssertMapEqual(
				t,
				to.Webhook.nextRunnable,
				tc.fromNextRunnable,
				"Fails.Webhook.nextRunnable",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}

			from.Webhook.nextRunnable["foo"] = time2023
			from.Webhook.nextRunnable["new"] = time2023
			if testErr := test.AssertMapEqual(
				t,
				to.Webhook.nextRunnable,
				tc.fromNextRunnable,
				"Fails.Webhook.nextRunnable",
			); testErr != nil {
				t.Errorf("after original mutated: %v", testErr)
			}
		})
	}
}

func TestFails_ResetFails(t *testing.T) {
	tests := []struct {
		name                        string
		commandFails                *[]*bool
		shoutrrrFails, webhookFails *map[string]*bool
	}{
		{
			name: "all default",
		},
		{
			name:          "all empty",
			commandFails:  &[]*bool{},
			shoutrrrFails: &map[string]*bool{},
			webhookFails:  &map[string]*bool{},
		},
		{
			name: "only notifiers",
			shoutrrrFails: &map[string]*bool{
				"0": nil,
				"1": new(false),
				"3": new(true),
			},
		},
		{
			name: "only commands",
			commandFails: &[]*bool{
				nil,
				new(false),
				new(true),
			},
		},
		{
			name: "only webhooks",
			webhookFails: &map[string]*bool{
				"0": nil,
				"1": new(false),
				"3": new(true),
			},
		},
		{
			name: "all filled",
			shoutrrrFails: &map[string]*bool{
				"0": nil,
				"1": new(false),
				"3": new(true),
			},
			commandFails: &[]*bool{
				nil,
				new(false),
				new(true),
			},
			webhookFails: &map[string]*bool{
				"0": nil,
				"1": new(false),
				"3": new(true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fails := Fails{}
			if tc.commandFails != nil {
				fails.Command.Init(len(*tc.commandFails))
				for i, v := range *tc.commandFails {
					if v != nil {
						fails.Command.Set(i, *v)
					}
				}
			}
			if tc.shoutrrrFails != nil {
				fails.Shoutrrr.Init(len(*tc.shoutrrrFails))
				for k, v := range *tc.shoutrrrFails {
					fails.Shoutrrr.Set(k, v)
				}
			}
			if tc.webhookFails != nil {
				fails.Webhook.Init(len(*tc.webhookFails))
				for k, v := range *tc.webhookFails {
					fails.Webhook.Set(k, v)
				}
			}

			fails.resetFails()

			if tc.shoutrrrFails != nil {
				for i := range *tc.shoutrrrFails {
					if got := fails.Shoutrrr.Get(i); got != nil {
						t.Errorf(
							"Fails.Shoutrrr.Get(%q) mismatch\ngot:  %s\nwant: nil (reset)",
							i, test.StringifyPtr(got),
						)
					}
				}
			}
			if tc.commandFails != nil {
				for i := range *tc.commandFails {
					if got := fails.Command.Get(i); got != nil {
						t.Errorf(
							"Fails.Command.Get(%d) mismatch\ngot:  %s\nwant: nil (reset)",
							i, test.StringifyPtr(got),
						)
					}
				}
			}
			if tc.webhookFails != nil {
				for i := range *tc.webhookFails {
					if got := fails.Webhook.Get(i); got != nil {
						t.Errorf(
							"Fails.Webhook.Get(%q) mismatch\ngot:  %s\nwant: nil (reset)",
							i, test.StringifyPtr(got),
						)
					}
				}
			}
		})
	}
}
