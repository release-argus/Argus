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
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
)

var packageName = "status"

func TestFailsBase_Init(t *testing.T) {
	// GIVEN: a Fails.
	tests := []struct {
		size int
	}{
		{size: 0},
		{size: 1},
		{size: 2},
		{size: 3},
		{size: 4},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("size=%d", tc.size)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var failsCommand FailsCommand
			var failsShoutrrr FailsShoutrrr
			var failsWebHook FailsWebHook

			// WHEN: we Init.
			failsCommand.Init(tc.size)
			failsShoutrrr.Init(tc.size)
			failsWebHook.Init(tc.size)

			// THEN: the size of the map is as expected.
			if got := len(failsCommand.fails); got != tc.size {
				t.Errorf(
					"%s\nFailsCommand.Init(%d) length mismatch\ngot:  %d\nwant: %d",
					packageName, tc.size,
					got, tc.size,
				)
			}
			if failsShoutrrr.fails == nil {
				t.Errorf(
					"%s\nFailsShoutrrr.Init(%d) length mismatch\ngot:  nil\nwant: non-nil",
					packageName, tc.size,
				)
			}
			if failsWebHook.fails == nil {
				t.Errorf(
					"%s\nFailsWebHook.Init(%d) length mismatch\ngot:  nil\n\nwant: non-nil",
					packageName, tc.size,
				)
			}
		})
	}
}

func TestFailsBase_SetAndGet(t *testing.T) {
	// GIVEN: a Fails.
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
				"test": test.Ptr(true),
			},
		},
		{
			name: "can add to non-empty map or edit array",
			size: 3,
			setAtArray: map[int]*bool{
				0: test.Ptr(true),
				1: test.Ptr(false),
				2: test.Ptr(true),
			},
			setAtMap: map[string]*bool{
				"bish": test.Ptr(true),
				"bash": test.Ptr(false),
				"bosh": test.Ptr(true),
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
			var failsWebHook FailsWebHook
			failsWebHook.Init(tc.size)
			// Ensure they are empty.
			for i := range tc.setAtArray {
				got := failsCommand.Get(i)
				if got != nil {
					t.Errorf(
						"%s\nFailsCommand, Get(%d) after Init(%d)\ngot:  %v\nwant: nil",
						packageName, i, tc.size,
						got,
					)
				}
			}
			for k := range tc.setAtMap {
				got := failsShoutrrr.Get(k)
				if got != nil {
					t.Errorf(
						"%s\nFailsShoutrrr, Get(%q) after Init(%d)\ngot:  %v\nwant: nil",
						packageName, k, tc.size,
						got,
					)
				}
				got = failsWebHook.Get(k)
				if got != nil {
					t.Errorf(
						"%s\nFailsWebHook, Get(%q) after Init(%d)\ngot:  %v\nwant: nil",
						packageName, k, tc.size,
						got,
					)
				}
			}

			// WHEN: we Set.
			for i, v := range tc.setAtArray {
				failsCommand.Set(i, *v)
			}
			for k, v := range tc.setAtMap {
				failsShoutrrr.Set(k, v)
				failsWebHook.Set(k, v)
			}

			// THEN: the values can be retrieved with Get.
			for i, v := range tc.setAtArray {
				got := failsCommand.Get(i)
				gotStr := test.StringifyPtr(got)
				wantStr := test.StringifyPtr(v)
				if got == nil || *got != *v {
					t.Errorf(
						"%s\nFailsCommand, Get(%d) after Set(&%t)\ngot:  %s\nwant: %v",
						packageName, i, *v,
						gotStr, wantStr,
					)
				}
			}
			for k, v := range tc.setAtMap {
				got := failsShoutrrr.Get(k)
				gotStr := test.StringifyPtr(got)
				wantStr := test.StringifyPtr(v)
				if got == nil || *got != *v {
					t.Errorf(
						"%s\nFailsShoutrrr, Get(%q) after Set(&%v)\ngot:  %v\nwant: %v",
						packageName, k, v,
						gotStr, wantStr,
					)
				}
				got = failsWebHook.Get(k)
				gotStr = test.StringifyPtr(got)
				if got == nil || *got != *v {
					t.Errorf(
						"%s\nFailsWebHook, Get(%q) after Set(&%v)\ngot:  %v\nwant: %v",
						packageName, k, v,
						gotStr, wantStr,
					)
				}
			}
		})
	}
}

func TestFailsBase_AllPassed(t *testing.T) {
	// GIVEN: a Fails.
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
				0: test.Ptr(true),
				1: test.Ptr(true),
				2: test.Ptr(true),
			},
			want: false,
		},
		{
			name: "all false (passed)",
			fails: map[int]*bool{
				0: test.Ptr(false),
				1: test.Ptr(false),
				2: test.Ptr(false),
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
				0: test.Ptr(true),
				1: test.Ptr(false),
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

			// WHEN: we call AllPassed.
			gotC := failsCommand.AllPassed()
			gotS := failsShoutrrr.AllPassed()
			gotWH := failsWebHook.AllPassed()

			// THEN: the result is as expected.
			if gotC != tc.want {
				t.Errorf(
					"%s\nFailsCommand.AllPassed() mismatch\ngot:  %t\nwant: %t",
					packageName, gotC, tc.want,
				)
			}
			if gotS != tc.want {
				t.Errorf(
					"%s\nFailsShoutrrr.AllPassed() mismatch\ngot:  %t\nwant: %t",
					packageName, gotS, tc.want,
				)
			}
			if gotWH != tc.want {
				t.Errorf(
					"%s\nFailsWebHook.AllPassed() mismatch\ngot:  %t\nwant: %t",
					packageName, gotWH, tc.want,
				)
			}
		})
	}
}

func TestFailsBase_Reset(t *testing.T) {
	// GIVEN: a Fails.
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
				0: test.Ptr(true),
				1: test.Ptr(true),
				2: test.Ptr(true),
			},
		},
		{
			name: "all false (passed)",
			fails: map[int]*bool{
				0: test.Ptr(false),
				1: test.Ptr(false),
				2: test.Ptr(false),
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
				0: test.Ptr(true),
				1: test.Ptr(false),
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

			// WHEN: we call Reset.
			failsCommand.Reset()
			failsShoutrrr.Reset()
			failsWebHook.Reset()

			// THEN: all the indices are reset to nil.
			for i := range tc.fails {
				got := failsCommand.Get(i)
				if got != nil {
					t.Errorf(
						"%s\nFailsCommand\n\ngot:  %vwant: nil",
						packageName, got,
					)
				}
				iStr := strconv.Itoa(i)
				got = failsShoutrrr.Get(iStr)
				if got != nil {
					t.Errorf(
						"%s\nFailsShoutrrr\ngot:  %v\nwant: nil",
						packageName, got,
					)
				}
				got = failsWebHook.Get(iStr)
				if got != nil {
					t.Errorf(
						"%s\nFailsWebHook\ngot:  %v\nwant: nil",
						packageName, got,
					)
				}
			}
		})
	}
}

func TestFailsBase_Length(t *testing.T) {
	// GIVEN: a Fails.
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
				"test": test.Ptr(true),
			},
		},
		{
			name: "can add to non-empty map or edit array",
			size: 3,
			setAtArray: map[int]*bool{
				0: test.Ptr(true),
				1: test.Ptr(false),
				2: test.Ptr(true),
			},
			setAtMap: map[string]*bool{
				"bish": test.Ptr(true),
				"bash": test.Ptr(false),
				"bosh": test.Ptr(true),
			},
		},
		{
			name: "length gives number of elements in map, not make size",
			size: 3,
			setAtArray: map[int]*bool{
				0: test.Ptr(true),
				1: test.Ptr(false),
			},
			setAtMap: map[string]*bool{
				"bish": test.Ptr(true),
				"bash": test.Ptr(false),
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

			// WHEN: we call Length.
			lengthC := failsCommand.Length()
			lengthS := failsShoutrrr.Length()
			lengthWH := failsWebHook.Length()

			// THEN: the lengths of the maps are returned.
			if lengthC != tc.size {
				t.Errorf(
					"%s\nFailsCommand\ngot:  %v\nwant: %v",
					packageName, lengthC, tc.size,
				)
			}
			if lengthS != len(tc.setAtMap) {
				t.Errorf(
					"%s\nFailsShoutrrr\ngot:  %v\nwant: %v",
					packageName, lengthS, len(tc.setAtMap),
				)
			}
			if lengthWH != len(tc.setAtMap) {
				t.Errorf(
					"%s\nFailsWebHook\ngot:  %v\nwant: %v",
					packageName, lengthWH, len(tc.setAtMap),
				)
			}
		})
	}
}

func TestFails_String(t *testing.T) {
	// GIVEN: a Fails.
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
				nil, test.Ptr(false),
			},
			shoutrrrFails: map[string]*bool{
				"bar": test.Ptr(false),
				"foo": nil,
			},
			webhookFails: map[string]*bool{
				"bar": nil,
				"foo": test.Ptr(false),
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
				"bash": test.Ptr(false),
				"bish": nil,
				"bosh": test.Ptr(true),
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
				test.Ptr(false),
				test.Ptr(true),
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
				"bash": test.Ptr(true),
				"bish": test.Ptr(false),
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
				"bash": test.Ptr(false),
				"bish": test.Ptr(true),
				"bosh": nil,
			},
			commandFails: []*bool{
				nil,
				test.Ptr(false),
				test.Ptr(true),
			},
			webhookFails: map[string]*bool{
				"bash": test.Ptr(false),
				"bish": nil,
				"bosh": test.Ptr(true),
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
				"bish": test.Ptr(true),
				"bash": test.Ptr(true),
				"bosh": test.Ptr(true),
			},
			commandFails: []*bool{
				nil,
				test.Ptr(true),
				test.Ptr(false),
			},
			webhookFails: map[string]*bool{
				"zip": test.Ptr(true),
				"zap": test.Ptr(true),
				"zop": test.Ptr(true),
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
			fails.WebHook.Init(len(tc.webhookFails))
			for k, v := range tc.webhookFails {
				fails.WebHook.Set(k, v)
			}

			// WHEN: the Fails are stringified with String.
			got := fails.String("")

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nFails.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestFailsWebHook_SetAndGetNextRunnable(t *testing.T) {
	// GIVEN: a FailsWebHook.
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
			name: "can add to non-empty map or edit array",
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

			var failsWebHook FailsWebHook
			failsWebHook.Init(tc.size)
			// Ensure they are empty.
			for k := range tc.setAtMap {
				got := failsWebHook.Get(k)
				if got != nil {
					t.Errorf(
						"%s\nFailsWebHook, NextRunnable after Init\ngot:  %v\nwant: nil",
						packageName, got,
					)
				}
			}

			// WHEN: we Set.
			for k, v := range tc.setAtMap {
				failsWebHook.SetNextRunnable(k, v)
			}

			// THEN: the values can be retrieved with Get.
			for k, v := range tc.setAtMap {
				got := failsWebHook.NextRunnable(k)
				if got != v {
					t.Errorf(
						"%s\nFailsWebHook, NextRunnable(%q) after SetNextRunnable(%q)\ngot:  %s\nwant: %s",
						packageName, k, v,
						got, v,
					)
				}
			}
		})
	}
}

func TestFails_Copy(t *testing.T) {
	// GIVEN: a fails to copy from and a fails to copy to.
	tests := []struct {
		name                               string
		fromCommandFails, toCommandFails   []*bool
		fromShoutrrrFails, toShoutrrrFails map[string]*bool
		fromWebHookFails, toWebHookFails   map[string]*bool
	}{
		{
			name:              "copy empty fails",
			fromCommandFails:  []*bool{},
			toCommandFails:    []*bool{},
			fromShoutrrrFails: map[string]*bool{},
			toShoutrrrFails:   map[string]*bool{},
			fromWebHookFails:  map[string]*bool{},
			toWebHookFails:    map[string]*bool{},
		},
		{
			name: "copy non-empty fails",
			fromCommandFails: []*bool{
				nil,
				test.Ptr(false),
				test.Ptr(true),
			},
			toCommandFails: []*bool{
				test.Ptr(true),
				nil,
				test.Ptr(false),
			},
			fromShoutrrrFails: map[string]*bool{
				"foo": test.Ptr(false),
				"bar": test.Ptr(true),
			},
			toShoutrrrFails: map[string]*bool{
				"baz": test.Ptr(true),
			},
			fromWebHookFails: map[string]*bool{
				"foo": test.Ptr(true),
				"bar": test.Ptr(false),
			},
			toWebHookFails: map[string]*bool{
				"baz": test.Ptr(false),
			},
		},
		{
			name: "copy to smaller fails",
			fromCommandFails: []*bool{
				nil,
				test.Ptr(false),
				test.Ptr(true),
			},
			toCommandFails: []*bool{
				test.Ptr(true),
				nil,
			},
			fromShoutrrrFails: map[string]*bool{
				"foo": test.Ptr(false),
				"bar": test.Ptr(true),
			},
			toShoutrrrFails: map[string]*bool{
				"baz": test.Ptr(true),
			},
			fromWebHookFails: map[string]*bool{
				"foo": test.Ptr(true),
				"bar": test.Ptr(false),
			},
			toWebHookFails: map[string]*bool{
				"baz": test.Ptr(false),
			},
		},
		{
			name: "copy to larger fails",
			fromCommandFails: []*bool{
				nil,
				test.Ptr(false),
				test.Ptr(true),
			},
			toCommandFails: []*bool{
				test.Ptr(true),
				nil,
				test.Ptr(false),
				nil,
			},
			fromShoutrrrFails: map[string]*bool{
				"foo": test.Ptr(false),
				"bar": test.Ptr(true),
			},
			toShoutrrrFails: map[string]*bool{
				"baz":  test.Ptr(true),
				"bosh": nil,
			},
			fromWebHookFails: map[string]*bool{
				"foo": test.Ptr(true),
				"bar": test.Ptr(false),
			},
			toWebHookFails: map[string]*bool{
				"baz":  test.Ptr(false),
				"bosh": nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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

			// WHEN: we call Copy.
			to.Copy(&from)

			prefix := fmt.Sprintf("%q\nFails.Copy()", packageName)

			// THEN: the values are copied correctly.
			// Command.
			if gotLen, wantLen := len(to.Command.fails), len(from.Command.fails); gotLen > wantLen {
				t.Errorf(
					"%s Command, length mismatch\ngot:  %d\nwant: %d",
					prefix, gotLen, wantLen,
				)
			}
			for k, v := range tc.fromCommandFails {
				got := test.StringifyPtr(to.Command.Get(k))
				want := test.StringifyPtr(v)
				if got != want {
					t.Errorf(
						"%s Command[%d] mismatch, got:  %s\nwant: %s",
						prefix, k,
						got, want,
					)
				}
			}
			// Shoutrrr.
			wantLen := len(from.Shoutrrr.fails)
			gotLen := len(to.Shoutrrr.fails)
			if gotLen > wantLen {
				t.Errorf(
					"%s\nShoutrrr, length mismatch\ngot:  %d\nwant: %d",
					packageName,
					gotLen, wantLen,
				)
			}
			for k, v := range tc.fromShoutrrrFails {
				got := test.StringifyPtr(to.Shoutrrr.Get(k))
				want := test.StringifyPtr(v)
				if got != want {
					t.Errorf(
						"%s\nShoutrrr[%q]\ngot:  %s\nwant: %s",
						packageName, k,
						got, want,
					)
				}
			}
			// WebHook.
			wantLen = len(from.WebHook.fails)
			gotLen = len(to.WebHook.fails)
			if gotLen > wantLen {
				t.Errorf(
					"%s\nWebHook, length mismatch\ngot:  %d\nwant: %d",
					packageName, gotLen, wantLen,
				)
			}
			for k, v := range tc.fromWebHookFails {
				got := test.StringifyPtr(to.WebHook.Get(k))
				want := test.StringifyPtr(v)
				if got != want {
					t.Errorf(
						"%s\nWebHook[%q]\ngot:  %s\nwant: %s",
						packageName, k,
						got, want,
					)
				}
			}
		})
	}
}

func TestFails_ResetFails(t *testing.T) {
	// GIVEN: a Fails struct.
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
			name: "only notifies",
			shoutrrrFails: &map[string]*bool{
				"0": nil,
				"1": test.Ptr(false),
				"3": test.Ptr(true),
			},
		},
		{
			name: "only commands",
			commandFails: &[]*bool{
				nil,
				test.Ptr(false),
				test.Ptr(true),
			},
		},
		{
			name: "only webhooks",
			webhookFails: &map[string]*bool{
				"0": nil,
				"1": test.Ptr(false),
				"3": test.Ptr(true),
			},
		},
		{
			name: "all filled",
			shoutrrrFails: &map[string]*bool{
				"0": nil,
				"1": test.Ptr(false),
				"3": test.Ptr(true),
			},
			commandFails: &[]*bool{
				nil,
				test.Ptr(false),
				test.Ptr(true),
			},
			webhookFails: &map[string]*bool{
				"0": nil,
				"1": test.Ptr(false),
				"3": test.Ptr(true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fails := Fails{}
			if tc.commandFails != nil {
				fails.Command.Init(len(*tc.commandFails))
			}
			if tc.shoutrrrFails != nil {
				fails.Shoutrrr.Init(len(*tc.shoutrrrFails))
			}
			if tc.webhookFails != nil {
				fails.WebHook.Init(len(*tc.webhookFails))
			}

			// WHEN: resetFails is called on it.
			fails.resetFails()

			// THEN: all the fails become nil.
			if tc.shoutrrrFails != nil {
				for i := range *tc.shoutrrrFails {
					if got := fails.Shoutrrr.Get(i); got != nil {
						t.Errorf(
							"%s\nStatus.Fails.Shoutrrr.Get(%s) not reset with .resetFails()\ngot:  %v\nwant: nil",
							packageName, i, got,
						)
					}
				}
			}
			if tc.commandFails != nil {
				for i := range *tc.commandFails {
					if got := fails.Command.Get(i); got != nil {
						t.Errorf(
							"%s\nStatus.Fails.Command.Get(%d) not reset with .resetFails()\ngot:  %t\nwant: nil",
							packageName, i, *got,
						)
					}
				}
			}
			if tc.webhookFails != nil {
				for i := range *tc.webhookFails {
					if got := fails.WebHook.Get(i); got != nil {
						t.Errorf(
							"%s\nStatus.Fails.WebHook.Get(%s) not reset with .resetFails()\ngot:  %t\nwant: nil",
							packageName, i, *got,
						)
					}
				}
			}
		})
	}
}

func TestFailsShoutrrr_Copy(t *testing.T) {
	tests := []struct {
		name string
		// If initOriginal is false, the receiver's fails map is left nil.
		initOriginal bool
		original     map[string]*bool
		// Mutations to verify that Copy doesn't alias the underlying map.
		mutateOriginal map[string]*bool
		mutateCopy     map[string]*bool
	}{
		{
			name:           "nil receiver",
			initOriginal:   false,
			original:       nil,
			mutateOriginal: map[string]*bool{"a": test.Ptr(true)},
			mutateCopy:     map[string]*bool{"a": test.Ptr(false)},
		},
		{
			name:           "empty receiver",
			initOriginal:   true,
			original:       map[string]*bool{},
			mutateOriginal: map[string]*bool{"a": test.Ptr(true)},
			mutateCopy:     map[string]*bool{"b": test.Ptr(false)},
		},
		{
			name:         "filled receiver",
			initOriginal: true,
			original: map[string]*bool{
				"foo": test.Ptr(false),
				"bar": nil,
				"baz": test.Ptr(true),
			},
			mutateOriginal: map[string]*bool{
				"foo": test.Ptr(true),  // replace value
				"baz": test.Ptr(false), // add new key
			},
			mutateCopy: map[string]*bool{
				"foo": test.Ptr(false), // mutate copy
				"qux": nil,             // add nil entry
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: a FailsShoutrrr.
			orig := &FailsShoutrrr{}
			if tc.initOriginal {
				orig.Init(len(tc.original))
				for k, v := range tc.original {
					orig.Set(k, v)
				}
			}

			// WHEN: Copy is called.
			inputCopy := orig.Copy()

			prefix := fmt.Sprintf("%s\nFailsShoutrrr.Copy()", packageName)

			// THEN: initial state matches.
			initialCopyWant := tc.original
			if !tc.initOriginal {
				initialCopyWant = map[string]*bool{}
			}
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.fails,
				orig.fails,
				prefix,
				"result",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: the original should not be mutated.
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.fails,
				orig.fails,
				prefix,
				"original",
			); testErr != nil {
				t.Error(testErr)
			}

			// WHEN: the original is mutated.
			if tc.initOriginal {
				for k, v := range tc.mutateOriginal {
					orig.Set(k, v)
				}
			} else {
				orig.Init(len(tc.mutateOriginal))
				for k, v := range tc.mutateOriginal {
					orig.Set(k, v)
				}
			}

			// THEN: the copy result should be unchanged.
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.fails,
				initialCopyWant,
				fmt.Sprintf("%s after mutating original", prefix),
				"result",
			); testErr != nil {
				t.Error(testErr)
			}

			// WHEN: the copy is mutated.
			for k, v := range tc.mutateCopy {
				inputCopy.Set(k, v)
			}

			// THEN: the copy result should be mutated.
			wantMutatedInput := tc.original
			if wantMutatedInput == nil {
				wantMutatedInput = tc.mutateCopy
			} else {
				for k, v := range tc.mutateCopy {
					wantMutatedInput[k] = v
				}
			}
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.fails,
				wantMutatedInput,
				fmt.Sprintf("%s after mutating copy", prefix),
				"result",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestFailsWebHook_Copy(t *testing.T) {
	time2021 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	time2022 := time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)
	time2023 := time.Date(2023, 3, 3, 0, 0, 0, 0, time.UTC)
	time2024 := time.Date(2024, 4, 4, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		// If initOriginal is false, the receiver's fails map is left nil.
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
			name:                       "nil receiver",
			initOriginal:               false,
			originalFails:              nil,
			mutateOriginalFails:        map[string]*bool{"a": test.Ptr(true)},
			mutateOriginalNextRunnable: map[string]time.Time{"a": time2022},
			mutateCopyFails:            map[string]*bool{"a": test.Ptr(false)},
			mutateCopyNextRunnable:     map[string]time.Time{"b": time2023},
		},
		{
			name:                       "empty receiver",
			initOriginal:               true,
			originalFails:              map[string]*bool{},
			mutateOriginalFails:        map[string]*bool{"a": test.Ptr(true)},
			mutateOriginalNextRunnable: map[string]time.Time{"a": time2022},
			mutateCopyFails:            map[string]*bool{"b": test.Ptr(false)},
			mutateCopyNextRunnable:     map[string]time.Time{"b": time2023},
		},
		{
			name:         "filled receiver",
			initOriginal: true,
			originalFails: map[string]*bool{
				"foo": test.Ptr(false),
				"bar": nil,
				"baz": test.Ptr(true),
			},
			originalNextRunnable: map[string]time.Time{
				"foo": time2021,
				"bar": time2022,
			},
			mutateOriginalFails: map[string]*bool{
				"foo": test.Ptr(true),  // replace value
				"baz": test.Ptr(false), // add new key
			},
			mutateOriginalNextRunnable: map[string]time.Time{
				"foo": time2023,
				"baz": time2024,
			},
			mutateCopyFails: map[string]*bool{
				"foo": test.Ptr(false), // mutate copy
				"qux": nil,             // add nil entry
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

			// GIVEN: a FailsWebHook.
			orig := &FailsWebHook{}
			if tc.initOriginal {
				orig.Init(len(tc.originalFails))
				for k, v := range tc.originalFails {
					orig.Set(k, v)
				}
				for k, v := range tc.originalNextRunnable {
					orig.SetNextRunnable(k, v)
				}
			}

			// WHEN: Copy is called.
			inputCopy := orig.Copy()

			prefix := fmt.Sprintf("%s\nFailsWebHook.Copy()", packageName)

			// THEN: initial state matches.
			//   Fails
			initialCopyWant := tc.originalFails
			if !tc.initOriginal {
				initialCopyWant = map[string]*bool{}
			}
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.fails,
				initialCopyWant,
				prefix,
				"fails",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: initial state matches.
			//   NextRunnable.
			initialNextRunnableWant := tc.originalNextRunnable
			if !tc.initOriginal {
				initialNextRunnableWant = map[string]time.Time{}
			}
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.nextRunnable,
				initialNextRunnableWant,
				prefix,
				"nextRunnable",
			); testErr != nil {
				t.Error(testErr)
			}

			// WHEN: the original is mutated.
			//   Fails.
			if !tc.initOriginal {
				orig.Init(len(tc.mutateOriginalFails))
			}
			for k, v := range tc.mutateOriginalFails {
				orig.Set(k, v)
			}
			// THEN: the copy Fails result should be unchanged.
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.fails,
				initialCopyWant,
				fmt.Sprintf("%s after mutating original", prefix),
				"fails",
			); testErr != nil {
				t.Error(testErr)
			}
			// WHEN: the original is mutated.
			//   NextRunnable.
			for k, v := range tc.mutateOriginalNextRunnable {
				orig.SetNextRunnable(k, v)
			}
			// THEN: the copy NextRunnable result should be unchanged.
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.nextRunnable,
				initialNextRunnableWant,
				fmt.Sprintf("%s after mutating original", prefix),
				"nextRunnable",
			); testErr != nil {
				t.Error(testErr)
			}

			// WHEN: the copy is mutated.
			//   Fails.
			for k, v := range tc.mutateCopyFails {
				inputCopy.Set(k, v)
			}
			// THEN: the copy Fails result should be mutated.
			wantMutatedFails := initialCopyWant
			if !tc.initOriginal {
				wantMutatedFails = tc.mutateCopyFails
			} else {
				wantMutatedFails = make(map[string]*bool, len(initialCopyWant)+len(tc.mutateCopyFails))
				for k, v := range initialCopyWant {
					wantMutatedFails[k] = v
				}
				for k, v := range tc.mutateCopyFails {
					wantMutatedFails[k] = v
				}
			}
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.fails,
				wantMutatedFails,
				fmt.Sprintf("%s after mutating copy", prefix),
				"fails",
			); testErr != nil {
				t.Error(testErr)
			}

			// WHEN: the copy is mutated.
			//   NextRunnable.
			for k, v := range tc.mutateCopyNextRunnable {
				inputCopy.SetNextRunnable(k, v)
			}
			// THEN: the copy NextRunnable result should be mutated.
			wantMutatedNextRunnable := initialNextRunnableWant
			if !tc.initOriginal {
				wantMutatedNextRunnable = tc.mutateCopyNextRunnable
			} else {
				wantMutatedNextRunnable = make(
					map[string]time.Time,
					len(initialNextRunnableWant)+len(tc.mutateCopyNextRunnable),
				)
				for k, v := range initialNextRunnableWant {
					wantMutatedNextRunnable[k] = v
				}
				for k, v := range tc.mutateCopyNextRunnable {
					wantMutatedNextRunnable[k] = v
				}
			}
			if testErr := test.AssertMapEqual(
				t,
				inputCopy.nextRunnable,
				wantMutatedNextRunnable,
				fmt.Sprintf("%s after mutating copy", prefix),
				"nextRunnable",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: the original should not reflect copy mutations.
			//   Fails.
			wantOrigFails := initialCopyWant
			if !tc.initOriginal {
				wantOrigFails = tc.mutateOriginalFails
			} else {
				wantOrigFails = make(
					map[string]*bool,
					len(initialCopyWant)+len(tc.mutateOriginalFails),
				)
				for k, v := range initialCopyWant {
					wantOrigFails[k] = v
				}
				for k, v := range tc.mutateOriginalFails {
					wantOrigFails[k] = v
				}
			}
			if testErr := test.AssertMapEqual(
				t,
				orig.fails,
				wantOrigFails,
				fmt.Sprintf("%s after mutating copy", prefix),
				"original fails",
			); testErr != nil {
				t.Error(testErr)
			}

			// AND: the original should not reflect copy mutations.
			//   NextRunnable.
			wantOrigNextRunnable := initialNextRunnableWant
			if !tc.initOriginal {
				wantOrigNextRunnable = tc.mutateOriginalNextRunnable
			} else {
				wantOrigNextRunnable = make(
					map[string]time.Time,
					len(initialNextRunnableWant)+len(tc.mutateOriginalNextRunnable),
				)
				for k, v := range initialNextRunnableWant {
					wantOrigNextRunnable[k] = v
				}
				for k, v := range tc.mutateOriginalNextRunnable {
					wantOrigNextRunnable[k] = v
				}
			}
			if testErr := test.AssertMapEqual(
				t,
				orig.nextRunnable,
				wantOrigNextRunnable,
				fmt.Sprintf("%s after mutating copy", prefix),
				"original nextRunnable",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
