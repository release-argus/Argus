// Copyright [2025] [Argus]
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

	"github.com/release-argus/Argus/test"
)

func TestFailsBase_Init(t *testing.T) {
	// GIVEN a Fails.
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

			// WHEN we Init.
			failsCommand.Init(tc.size)
			failsShoutrrr.Init(tc.size)
			failsWebHook.Init(tc.size)

			// THEN the size of the map is as expected.
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
	// GIVEN a Fails.
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
			// Ensure they are empty.
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

			// WHEN we Set.
			for i, v := range tc.setAtArray {
				failsCommand.Set(i, *v)
			}
			for k, v := range tc.setAtMap {
				failsShoutrrr.Set(k, v)
				failsWebHook.Set(k, v)
			}

			// THEN the values can be retrieved with Get.
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
	// GIVEN a Fails.
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

			// WHEN we call AllPassed.
			gotC := failsCommand.AllPassed()
			gotS := failsShoutrrr.AllPassed()
			gotWH := failsWebHook.AllPassed()

			// THEN the result is as expected.
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
	// GIVEN a Fails.
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

			// WHEN we call Reset.
			failsCommand.Reset()
			failsShoutrrr.Reset()
			failsWebHook.Reset()

			// THEN all the indices are reset to nil.
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
	// GIVEN a Fails.
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
			// Set the values.
			for i, v := range tc.setAtArray {
				if v != nil {
					failsCommand.Set(i, *v)
				}
			}
			for k, v := range tc.setAtMap {
				failsShoutrrr.Set(k, v)
				failsWebHook.Set(k, v)
			}

			// WHEN we call Length.
			lengthC := failsCommand.Length()
			lengthS := failsShoutrrr.Length()
			lengthWH := failsWebHook.Length()

			// THEN the lengths of the maps are returned.
			if lengthC != tc.size {
				t.Errorf("FailsCommand - want: %v, got: %v",
					tc.size, lengthC)
			}
			if lengthS != len(tc.setAtMap) {
				t.Errorf("FailsShoutrrr - want: %v, got: %v",
					len(tc.setAtMap), lengthS)
			}
			if lengthWH != len(tc.setAtMap) {
				t.Errorf("FailsWebHook - want: %v, got: %v",
					len(tc.setAtMap), lengthWH)
			}
		})
	}
}

func TestFails_String(t *testing.T) {
	// GIVEN a Fails.
	tests := map[string]struct {
		commandFails                []*bool
		shoutrrrFails, webhookFails map[string]*bool
		want                        string
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
		"only shoutrrr": {
			shoutrrrFails: map[string]*bool{
				"bash": test.BoolPtr(false),
				"bish": nil,
				"bosh": test.BoolPtr(true)},
			want: test.TrimYAML(`
				shoutrrr:
					bash: false
					bish: nil
					bosh: true
				`),
		},
		"only command": {
			commandFails: []*bool{
				nil,
				test.BoolPtr(false),
				test.BoolPtr(true)},
			want: test.TrimYAML(`
				command:
					- 0: nil
					- 1: false
					- 2: true
			`),
		},
		"only webhook": {
			webhookFails: map[string]*bool{
				"bash": test.BoolPtr(true),
				"bish": test.BoolPtr(false),
				"bosh": nil},
			want: test.TrimYAML(`
				webhook:
					bash: true
					bish: false
					bosh: nil
			`),
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
				"zip": test.BoolPtr(true),
				"zap": test.BoolPtr(true),
				"zop": test.BoolPtr(true)},
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

			// WHEN the Fails are stringified with String.
			got := fails.String("")

			// THEN the result is as expected.
			if got != tc.want {
				t.Errorf("Fails.String() mismatch\n%q\ngot:\n%q",
					tc.want, got)
			}
		})
	}
}

func TestFails_Copy(t *testing.T) {
	// GIVEN a fails to copy from and a fails to copy to.
	tests := map[string]struct {
		fromCommandFails, toCommandFails   []*bool
		fromShoutrrrFails, toShoutrrrFails map[string]*bool
		fromWebHookFails, toWebHookFails   map[string]*bool
	}{
		"copy empty fails": {
			fromCommandFails:  []*bool{},
			toCommandFails:    []*bool{},
			fromShoutrrrFails: map[string]*bool{},
			toShoutrrrFails:   map[string]*bool{},
			fromWebHookFails:  map[string]*bool{},
			toWebHookFails:    map[string]*bool{},
		},
		"copy non-empty fails": {
			fromCommandFails: []*bool{
				nil, test.BoolPtr(false), test.BoolPtr(true)},
			toCommandFails: []*bool{
				test.BoolPtr(true), nil, test.BoolPtr(false)},
			fromShoutrrrFails: map[string]*bool{
				"foo": test.BoolPtr(false), "bar": test.BoolPtr(true)},
			toShoutrrrFails: map[string]*bool{
				"baz": test.BoolPtr(true)},
			fromWebHookFails: map[string]*bool{
				"foo": test.BoolPtr(true), "bar": test.BoolPtr(false)},
			toWebHookFails: map[string]*bool{
				"baz": test.BoolPtr(false)},
		},
		"copy to smaller fails": {
			fromCommandFails: []*bool{
				nil, test.BoolPtr(false), test.BoolPtr(true)},
			toCommandFails: []*bool{
				test.BoolPtr(true), nil},
			fromShoutrrrFails: map[string]*bool{
				"foo": test.BoolPtr(false), "bar": test.BoolPtr(true)},
			toShoutrrrFails: map[string]*bool{
				"baz": test.BoolPtr(true)},
			fromWebHookFails: map[string]*bool{
				"foo": test.BoolPtr(true), "bar": test.BoolPtr(false)},
			toWebHookFails: map[string]*bool{
				"baz": test.BoolPtr(false)},
		},
		"copy to larger fails": {
			fromCommandFails: []*bool{
				nil, test.BoolPtr(false), test.BoolPtr(true)},
			toCommandFails: []*bool{
				test.BoolPtr(true), nil, test.BoolPtr(false), nil},
			fromShoutrrrFails: map[string]*bool{
				"foo": test.BoolPtr(false), "bar": test.BoolPtr(true)},
			toShoutrrrFails: map[string]*bool{
				"baz": test.BoolPtr(true), "bosh": nil},
			fromWebHookFails: map[string]*bool{
				"foo": test.BoolPtr(true), "bar": test.BoolPtr(false)},
			toWebHookFails: map[string]*bool{
				"baz": test.BoolPtr(false), "bosh": nil},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// fails to copy from.
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
			from.WebHook.Init(len(tc.fromWebHookFails))
			for k, v := range tc.fromWebHookFails {
				from.WebHook.Set(k, v)
			}

			// fails to copy to.
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
			to.WebHook.Init(len(tc.toWebHookFails))
			for k, v := range tc.toWebHookFails {
				to.WebHook.Set(k, v)
			}

			// WHEN we call Copy.
			to.Copy(&from)

			// THEN the values are copied correctly.
			// Command.
			if len(to.Command.fails) > len(from.Command.fails) {
				t.Errorf("Command - want: %d, got: %d",
					len(tc.fromCommandFails), len(tc.toCommandFails))
			}
			for k, v := range tc.fromCommandFails {
				fromStr := test.StringifyPtr(v)
				gotStr := test.StringifyPtr(to.Command.Get(k))
				if fromStr != gotStr {
					t.Errorf("Command - want: %s, got: %s",
						fromStr, gotStr)
				}
			}
			// Shoutrrr.
			if len(to.Shoutrrr.fails) > len(from.Shoutrrr.fails) {
				t.Errorf("Shoutrrr - want: %d, got: %d",
					len(tc.fromShoutrrrFails), len(tc.toShoutrrrFails))
			}
			for k, v := range tc.fromShoutrrrFails {
				fromStr := test.StringifyPtr(v)
				gotStr := test.StringifyPtr(to.Shoutrrr.Get(k))
				if fromStr != gotStr {
					t.Errorf("Shoutrrr - want: %s, got: %s",
						fromStr, gotStr)
				}
			}
			// WebHook.
			if len(to.WebHook.fails) > len(from.WebHook.fails) {
				t.Errorf("WebHook - want: %d, got: %d",
					len(tc.fromWebHookFails), len(tc.toWebHookFails))
			}
			for k, v := range tc.fromWebHookFails {
				fromStr := test.StringifyPtr(v)
				gotStr := test.StringifyPtr(to.WebHook.Get(k))
				if fromStr != gotStr {
					t.Errorf("WebHook - want: %s, got: %s",
						fromStr, gotStr)
				}
			}
		})
	}
}
