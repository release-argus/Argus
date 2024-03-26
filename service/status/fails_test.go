// Copyright [2023] [Argus]
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

package svcstatus

import (
	"strconv"
	"strings"
	"testing"
)

func TestFailsBase_Init(t *testing.T) {
	// GIVEN a Fails
	tests := map[string]struct {
		size int
	}{
		"0": {size: 0},
		"1": {size: 1},
		"2": {size: 2},
		"3": {size: 3},
		"4": {size: 4},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			var failsShoutrrr FailsShoutrrr
			var failsWebHook FailsWebHook

			// WHEN we Init
			failsCommand.Init(tc.size)
			failsShoutrrr.Init(tc.size)
			failsWebHook.Init(tc.size)

			// THEN the size of the map is as expected
			if len(failsCommand.fails) != tc.size {
				t.Errorf("FailsCommand - want: %d, got: %d",
					tc.size, len(failsCommand.fails))
			}
			if failsShoutrrr.fails == nil {
				t.Errorf("FailsCommand - want: non-nil, got: %v",
					failsShoutrrr.fails)
			}
			if failsWebHook.fails == nil {
				t.Errorf("FailsWebHook - want: non-nil, got: %v",
					failsWebHook.fails)
			}
		})
	}
}

func TestFailsBase_SetAndGet(t *testing.T) {
	// GIVEN a Fails
	tests := map[string]struct {
		size       int
		setAtArray map[int]*bool
		setAtMap   map[string]*bool
	}{
		"can add to empty map": {
			size:       0,
			setAtArray: map[int]*bool{},
			setAtMap: map[string]*bool{
				"test": boolPtr(true)},
		},
		"can add to non-empty map or edit array": {
			size: 3,
			setAtArray: map[int]*bool{
				0: boolPtr(true),
				1: boolPtr(false),
				2: boolPtr(true),
			},
			setAtMap: map[string]*bool{
				"bish": boolPtr(true),
				"bash": boolPtr(false),
				"bosh": boolPtr(true)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			failsCommand.Init(tc.size)
			var failsShoutrrr FailsShoutrrr
			failsShoutrrr.Init(tc.size)
			var failsWebHook FailsWebHook
			failsWebHook.Init(tc.size)
			// ensure they are empty
			for i := range tc.setAtArray {
				got := failsCommand.Get(i)
				if got != nil {
					t.Errorf("FailsCommand - want: nil, got: %v", got)
				}
			}
			for k := range tc.setAtMap {
				got := failsShoutrrr.Get(k)
				if got != nil {
					t.Errorf("FailsShoutrrr - want: nil, got: %v", got)
				}
				got = failsWebHook.Get(k)
				if got != nil {
					t.Errorf("FailsWebHook - want: nil, got: %v", got)
				}
			}

			// WHEN we Set
			for i, v := range tc.setAtArray {
				failsCommand.Set(i, *v)
			}
			for k, v := range tc.setAtMap {
				failsShoutrrr.Set(k, v)
				failsWebHook.Set(k, v)
			}

			// THEN the values can be retrieved with Get
			for i, v := range tc.setAtArray {
				got := failsCommand.Get(i)
				if got == nil {
					t.Errorf("FailsCommand - want: non-nil, got: %v", got)
				}
				if *got != *v {
					t.Errorf("FailsCommand - want: %v, got: %v", v, got)
				}
			}
			for k, v := range tc.setAtMap {
				got := failsShoutrrr.Get(k)
				if got == nil {
					t.Errorf("FailsShoutrrr - want: non-nil, got: %v", got)
				}
				if *got != *v {
					t.Errorf("FailsShoutrrr - want: %v, got: %v", v, got)
				}
				got = failsWebHook.Get(k)
				if got == nil {
					t.Errorf("FailsWebHook - want: non-nil, got: %v", got)
				}
				if *got != *v {
					t.Errorf("FailsWebHook - want: %v, got: %v", v, got)
				}
			}
		})
	}
}

func TestFailsBase_AllPassed(t *testing.T) {
	// GIVEN a Fails
	tests := map[string]struct {
		fails map[int]*bool
		want  bool
	}{
		"empty": {
			fails: map[int]*bool{},
			want:  true,
		},
		"all true (failed)": {
			fails: map[int]*bool{
				0: boolPtr(true),
				1: boolPtr(true),
				2: boolPtr(true)},
			want: false,
		},
		"all false (passed)": {
			fails: map[int]*bool{
				0: boolPtr(false),
				1: boolPtr(false),
				2: boolPtr(false)},
			want: true,
		},
		"all nil (not run)": {
			fails: map[int]*bool{
				0: nil,
				1: nil,
				2: nil},
			want: false,
		},
		"mixed": {
			fails: map[int]*bool{
				0: boolPtr(true),
				1: boolPtr(false),
				2: nil},
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			failsCommand.Init(len(tc.fails))
			var failsShoutrrr FailsShoutrrr
			failsShoutrrr.Init(len(tc.fails))
			var failsWebHook FailsWebHook
			failsWebHook.Init(len(tc.fails))
			for i, v := range tc.fails {
				if v != nil {
					failsCommand.Set(i, *v)
				}
				iStr := strconv.Itoa(i)
				failsShoutrrr.Set(iStr, v)
				failsWebHook.Set(iStr, v)
			}

			// WHEN we call AllPassed
			gotC := failsCommand.AllPassed()
			gotS := failsShoutrrr.AllPassed()
			gotWH := failsWebHook.AllPassed()

			// THEN the result is as expected
			if gotC != tc.want {
				t.Errorf("FailsCommand - want: %v, got: %v", tc.want, gotC)
			}
			if gotS != tc.want {
				t.Errorf("FailsShoutrrr - want: %v, got: %v", tc.want, gotS)
			}
			if gotWH != tc.want {
				t.Errorf("FailsWebHook - want: %v, got: %v", tc.want, gotWH)
			}
		})
	}
}

func TestFailsBase_Reset(t *testing.T) {
	// GIVEN a Fails
	tests := map[string]struct {
		fails map[int]*bool
	}{
		"empty": {
			fails: map[int]*bool{},
		},
		"all true (failed)": {
			fails: map[int]*bool{
				0: boolPtr(true),
				1: boolPtr(true),
				2: boolPtr(true)},
		},
		"all false (passed)": {
			fails: map[int]*bool{
				0: boolPtr(false),
				1: boolPtr(false),
				2: boolPtr(false)},
		},
		"all nil (not run)": {
			fails: map[int]*bool{
				0: nil,
				1: nil,
				2: nil},
		},
		"mixed": {
			fails: map[int]*bool{
				0: boolPtr(true),
				1: boolPtr(false),
				2: nil},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			failsCommand.Init(len(tc.fails))
			var failsShoutrrr FailsShoutrrr
			failsShoutrrr.Init(len(tc.fails))
			var failsWebHook FailsWebHook
			failsWebHook.Init(len(tc.fails))
			for i, v := range tc.fails {
				if v != nil {
					failsCommand.Set(i, *v)
				}
				iStr := strconv.Itoa(i)
				failsShoutrrr.Set(iStr, v)
				failsWebHook.Set(iStr, v)
			}

			// WHEN we call Reset
			failsCommand.Reset()
			failsShoutrrr.Reset()
			failsWebHook.Reset()

			// THEN all the indices are reset to nil
			for i := range tc.fails {
				got := failsCommand.Get(i)
				if got != nil {
					t.Errorf("FailsCommand - want: nil, got: %v", got)
				}
				iStr := strconv.Itoa(i)
				got = failsShoutrrr.Get(iStr)
				if got != nil {
					t.Errorf("FailsShoutrrr - want: nil, got: %v", got)
				}
				got = failsWebHook.Get(iStr)
				if got != nil {
					t.Errorf("FailsWebHook - want: nil, got: %v", got)
				}
			}
		})
	}
}

func TestFailsBase_Length(t *testing.T) {
	// GIVEN a Fails
	tests := map[string]struct {
		size       int
		setAtArray map[int]*bool
		setAtMap   map[string]*bool
	}{
		"can add to empty map": {
			size:       0,
			setAtArray: map[int]*bool{},
			setAtMap: map[string]*bool{
				"test": boolPtr(true)},
		},
		"can add to non-empty map or edit array": {
			size: 3,
			setAtArray: map[int]*bool{
				0: boolPtr(true),
				1: boolPtr(false),
				2: boolPtr(true),
			},
			setAtMap: map[string]*bool{
				"bish": boolPtr(true),
				"bash": boolPtr(false),
				"bosh": boolPtr(true)},
		},
		"length gives number of elements in map, not make size": {
			size: 3,
			setAtArray: map[int]*bool{
				0: boolPtr(true),
				1: boolPtr(false)},
			setAtMap: map[string]*bool{
				"bish": boolPtr(true),
				"bash": boolPtr(false)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			failsCommand.Init(tc.size)
			var failsShoutrrr FailsShoutrrr
			failsShoutrrr.Init(tc.size)
			var failsWebHook FailsWebHook
			failsWebHook.Init(tc.size)
			// Set the values
			for i, v := range tc.setAtArray {
				if v != nil {
					failsCommand.Set(i, *v)
				}
			}
			for k, v := range tc.setAtMap {
				failsShoutrrr.Set(k, v)
				failsWebHook.Set(k, v)
			}

			// WHEN we call Lanegth
			langthC := failsCommand.Length()
			langthS := failsShoutrrr.Length()
			langthWH := failsWebHook.Length()

			// THEN the langth's are returned correctly
			if langthC != tc.size {
				t.Errorf("FailsCommand - want: %v, got: %v", tc.size, langthC)
			}
			if langthS != len(tc.setAtMap) {
				t.Errorf("FailsShoutrrr - want: %v, got: %v", len(tc.setAtMap), langthS)
			}
			if langthWH != len(tc.setAtMap) {
				t.Errorf("FailsWebHook - want: %v, got: %v", len(tc.setAtMap), langthWH)
			}
		})
	}
}

func TestFails_String(t *testing.T) {
	// GIVEN a Fails
	tests := map[string]struct {
		commandFails  []*bool
		shoutrrrFails map[string]*bool
		webhookFails  map[string]*bool
		want          string
	}{
		"empty fails": {
			commandFails:  []*bool{},
			shoutrrrFails: map[string]*bool{},
			webhookFails:  map[string]*bool{},
			want:          "",
		},
		"no fails": {
			commandFails: []*bool{
				nil, boolPtr(false)},
			shoutrrrFails: map[string]*bool{
				"bar": boolPtr(false),
				"foo": nil},
			webhookFails: map[string]*bool{
				"bar": nil,
				"foo": boolPtr(false)},
			want: `
shoutrrr: {bar: false, foo: nil},
 command: [0: nil, 1: false],
 webhook: {bar: nil, foo: false}`,
		},
		"only shoutrrr": {
			shoutrrrFails: map[string]*bool{
				"bash": boolPtr(false),
				"bish": nil,
				"bosh": boolPtr(true)},
			want: `
shoutrrr: {bash: false, bish: nil, bosh: true}`,
		},
		"only command": {
			commandFails: []*bool{
				nil,
				boolPtr(false),
				boolPtr(true)},
			want: `
command: [0: nil, 1: false, 2: true]`,
		},
		"only webhook": {
			webhookFails: map[string]*bool{
				"bash": boolPtr(true),
				"bish": boolPtr(false),
				"bosh": nil},
			want: `
webhook: {bash: true, bish: false, bosh: nil}`,
		},
		"all": {
			shoutrrrFails: map[string]*bool{
				"bash": boolPtr(false),
				"bish": boolPtr(true),
				"bosh": nil},
			commandFails: []*bool{
				nil,
				boolPtr(false),
				boolPtr(true)},
			webhookFails: map[string]*bool{
				"bash": boolPtr(false),
				"bish": nil,
				"bosh": boolPtr(true)},
			want: `
shoutrrr: {bash: false, bish: true, bosh: nil},
 command: [0: nil, 1: false, 2: true],
 webhook: {bash: false, bish: nil, bosh: true}`,
		},
		"maps are alphabetical": {
			shoutrrrFails: map[string]*bool{
				"bish": boolPtr(true),
				"bash": boolPtr(true),
				"bosh": boolPtr(true)},
			commandFails: []*bool{
				nil,
				boolPtr(true),
				boolPtr(false)},
			webhookFails: map[string]*bool{
				"zip":  boolPtr(true),
				"zap":  boolPtr(true),
				"zoop": boolPtr(true)},
			want: `
shoutrrr: {bash: true, bish: true, bosh: true},
 command: [0: nil, 1: true, 2: false],
 webhook: {zap: true, zip: true, zoop: true}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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
			fails.WebHook.Init(len(tc.webhookFails))
			for k, v := range tc.webhookFails {
				fails.WebHook.Set(k, v)
			}

			// WHEN the Fails is stringified with String
			got := fails.String()

			// THEN the result is as expected
			tc.want = strings.ReplaceAll(tc.want, "\n", "")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}
