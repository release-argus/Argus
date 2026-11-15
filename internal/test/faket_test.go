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

//go:build unit || integration

package test

import (
	"slices"
	"testing"
)

func TestFakeT_Context(t *testing.T) {
	var fakeT FakeT

	ctx := fakeT.Context()

	if ctx == nil {
		t.Fatal("FakeT.Context() mismatch\ngot:  nil\nwant: non-nil")
	}
	if err := ctx.Err(); err != nil {
		t.Errorf("FakeT.Context().Err() mismatch\ngot:  %v\nwant: nil", err)
	}
}

func TestFakeT(t *testing.T) {
	tests := []struct {
		name       string
		logFn      func(*FakeT, ...any)
		fn         func(*FakeT, string, ...any)
		wantLogs   []string
		wantErrors []string
		wantFatals []string
		wantSkips  []string
	}{
		{
			name:     "Log",
			logFn:    (*FakeT).Log,
			wantLogs: []string{"hello %s world"},
		},
		{
			name:     "Logf",
			fn:       (*FakeT).Logf,
			wantLogs: []string{"hello world"},
		},
		{
			name:       "Errorf",
			fn:         (*FakeT).Errorf,
			wantErrors: []string{"hello world"},
		},
		{
			name:       "Fatalf",
			fn:         (*FakeT).Fatalf,
			wantFatals: []string{"hello world"},
		},
		{
			name:      "Skipf",
			fn:        (*FakeT).Skipf,
			wantSkips: []string{"hello world"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var fakeT FakeT
			fakeT.Helper() // No-op.

			arg1, arg2 := "hello %s", "world"
			if tc.logFn != nil {
				tc.logFn(&fakeT, arg1, arg2)
			} else {
				tc.fn(&fakeT, arg1, arg2)
			}

			for _, check := range []struct {
				name string
				got  []string
				want []string
			}{
				{name: "Logs", got: fakeT.Logs, want: tc.wantLogs},
				{name: "Errors", got: fakeT.Errors, want: tc.wantErrors},
				{name: "Fatals", got: fakeT.Fatals, want: tc.wantFatals},
				{name: "Skips", got: fakeT.Skips, want: tc.wantSkips},
			} {
				if !slices.Equal(check.got, check.want) {
					t.Errorf(
						"FakeT.%s mismatch\ngot:  %q\nwant: %q",
						check.name, check.got, check.want,
					)
				}
			}
		})
	}
}

func TestFakeT_Aborted(t *testing.T) {
	tests := []struct {
		name        string
		abort       bool
		logFn       func(*FakeT, ...any)
		fn          func(*FakeT, string, ...any)
		wantAborted bool
	}{
		{
			name: "nothing reported/runs to the end",
		},
		{
			name: "Abort unset/Fatalf returns",
			fn:   (*FakeT).Fatalf,
		},
		{
			name: "Abort unset/Skipf returns",
			fn:   (*FakeT).Skipf,
		},
		{
			name:        "Abort set/Fatalf stops the call",
			abort:       true,
			fn:          (*FakeT).Fatalf,
			wantAborted: true,
		},
		{
			name:        "Abort set/Skipf stops the call",
			abort:       true,
			fn:          (*FakeT).Skipf,
			wantAborted: true,
		},
		{
			name:  "Abort set/Log does not stop the call",
			abort: true,
			logFn: (*FakeT).Log,
		},
		{
			name:  "Abort set/Logf does not stop the call",
			abort: true,
			fn:    (*FakeT).Logf,
		},
		{
			name:  "Abort set/Errorf does not stop the call",
			abort: true,
			fn:    (*FakeT).Errorf,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fakeT := &FakeT{Abort: tc.abort}
			reachedEnd := false

			aborted := fakeT.Aborted(func() {
				switch {
				case tc.logFn != nil:
					tc.logFn(fakeT, "stop %s", "here")
				case tc.fn != nil:
					tc.fn(fakeT, "stop %s", "here")
				}
				reachedEnd = true
			})

			if aborted != tc.wantAborted {
				t.Errorf("aborted mismatch\ngot:  %t\nwant: %t", aborted, tc.wantAborted)
			}

			if wantEnd := !tc.wantAborted; reachedEnd != wantEnd {
				t.Errorf("reachedEnd mismatch\ngot:  %t\nwant: %t", reachedEnd, wantEnd)
			}
		})
	}
}

func TestFakeT_Aborted__foreignPanic(t *testing.T) {
	fakeT := &FakeT{Abort: true}
	want := "not an abort"

	defer func() {
		if got := recover(); got != want {
			t.Errorf("recover() mismatch\ngot:  %v\nwant: %q", got, want)
		}
	}()

	fakeT.Aborted(func() {
		panic(want)
	})
}
