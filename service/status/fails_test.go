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

	"github.com/release-argus/Argus/test"
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
				"test": test.BoolPtr(true)},
		},
		"can add to non-empty map or edit array": {
			size: 3,
			setAtArray: map[int]*bool{
				0: test.BoolPtr(true),
				1: test.BoolPtr(false),
				2: test.BoolPtr(true),
			},
			setAtMap: map[string]*bool{
				"bish": test.BoolPtr(true),
				"bash": test.BoolPtr(false),
				"bosh": test.BoolPtr(true)},
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
				0: test.BoolPtr(true),
				1: test.BoolPtr(true),
				2: test.BoolPtr(true)},
			want: false,
		},
		"all false (passed)": {
			fails: map[int]*bool{
				0: test.BoolPtr(false),
				1: test.BoolPtr(false),
				2: test.BoolPtr(false)},
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
				0: test.BoolPtr(true),
				1: test.BoolPtr(false),
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
				0: test.BoolPtr(true),
				1: test.BoolPtr(true),
				2: test.BoolPtr(true)},
		},
		"all false (passed)": {
			fails: map[int]*bool{
				0: test.BoolPtr(false),
				1: test.BoolPtr(false),
				2: test.BoolPtr(false)},
		},
		"all nil (not run)": {
			fails: map[int]*bool{
				0: nil,
				1: nil,
				2: nil},
		},
		"mixed": {
			fails: map[int]*bool{
				0: test.BoolPtr(true),
				1: test.BoolPtr(false),
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
				"test": test.BoolPtr(true)},
		},
		"can add to non-empty map or edit array": {
			size: 3,
			setAtArray: map[int]*bool{
				0: test.BoolPtr(true),
				1: test.BoolPtr(false),
				2: test.BoolPtr(true),
			},
			setAtMap: map[string]*bool{
				"bish": test.BoolPtr(true),
				"bash": test.BoolPtr(false),
				"bosh": test.BoolPtr(true)},
		},
		"length gives number of elements in map, not make size": {
			size: 3,
			setAtArray: map[int]*bool{
				0: test.BoolPtr(true),
				1: test.BoolPtr(false)},
			setAtMap: map[string]*bool{
				"bish": test.BoolPtr(true),
				"bash": test.BoolPtr(false)},
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
				nil, test.BoolPtr(false)},
			shoutrrrFails: map[string]*bool{
				"bar": test.BoolPtr(false),
				"foo": nil},
			webhookFails: map[string]*bool{
				"bar": nil,
				"foo": test.BoolPtr(false)},
			want: `
shoutrrr: {bar: false, foo: nil},
 command: [0: nil, 1: false],
 webhook: {bar: nil, foo: false}`,
		},
		"only shoutrrr": {
			shoutrrrFails: map[string]*bool{
				"bash": test.BoolPtr(false),
				"bish": nil,
				"bosh": test.BoolPtr(true)},
			want: `
shoutrrr: {bash: false, bish: nil, bosh: true}`,
		},
		"only command": {
			commandFails: []*bool{
				nil,
				test.BoolPtr(false),
				test.BoolPtr(true)},
			want: `
command: [0: nil, 1: false, 2: true]`,
		},
		"only webhook": {
			webhookFails: map[string]*bool{
				"bash": test.BoolPtr(true),
				"bish": test.BoolPtr(false),
				"bosh": nil},
			want: `
webhook: {bash: true, bish: false, bosh: nil}`,
		},
		"all": {
			shoutrrrFails: map[string]*bool{
				"bash": test.BoolPtr(false),
				"bish": test.BoolPtr(true),
				"bosh": nil},
			commandFails: []*bool{
				nil,
				test.BoolPtr(false),
				test.BoolPtr(true)},
			webhookFails: map[string]*bool{
				"bash": test.BoolPtr(false),
				"bish": nil,
				"bosh": test.BoolPtr(true)},
			want: `
shoutrrr: {bash: false, bish: true, bosh: nil},
 command: [0: nil, 1: false, 2: true],
 webhook: {bash: false, bish: nil, bosh: true}`,
		},
		"maps are alphabetical": {
			shoutrrrFails: map[string]*bool{
				"bish": test.BoolPtr(true),
				"bash": test.BoolPtr(true),
				"bosh": test.BoolPtr(true)},
			commandFails: []*bool{
				nil,
				test.BoolPtr(true),
				test.BoolPtr(false)},
			webhookFails: map[string]*bool{
				"zip":  test.BoolPtr(true),
				"zap":  test.BoolPtr(true),
				"zoop": test.BoolPtr(true)},
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
