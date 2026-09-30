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

func TestFakeT(t *testing.T) {
	// GIVEN: a functionalto call on a FakeT.
	tests := []struct {
		name       string
		fn         func(*FakeT, string, ...any)
		wantErrors []string
		wantFatals []string
		wantSkips  []string
	}{
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

			// WHEN: the function is called.
			tc.fn(&fakeT, "hello %s", "world")

			// THEN: it lands in that function's slice, and no other.
			for _, check := range []struct {
				name string
				got  []string
				want []string
			}{
				{name: "Errors", got: fakeT.Errors, want: tc.wantErrors},
				{name: "Fatals", got: fakeT.Fatals, want: tc.wantFatals},
				{name: "Skips", got: fakeT.Skips, want: tc.wantSkips},
			} {
				if !slices.Equal(check.got, check.want) {
					t.Errorf(
						"%s\nFakeT.%s(\"hello %%s\", \"world\") %s mismatch\ngot:  %q\nwant: %q",
						packageName, tc.name, check.name,
						check.got, check.want,
					)
				}
			}
		})
	}
}

func TestFakeT_Aborted(t *testing.T) {
	// GIVEN: a FakeT that may abort, and a function to make inside Aborted.
	tests := []struct {
		name        string
		abort       bool
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

			// WHEN: Aborted calls a function that may abort.
			aborted := fakeT.Aborted(func() {
				if tc.fn != nil {
					tc.fn(fakeT, "stop %s", "here")
				}
				reachedEnd = true
			})

			// THEN: an aborting function is reported only if tc.abort was set.
			if aborted != tc.wantAborted {
				t.Errorf(
					"%s\nFakeT.Aborted() mismatch\ngot:  %t\nwant: %t",
					packageName, aborted, tc.wantAborted,
				)
			}

			// AND: the call stopped exactly when it aborted.
			if wantEnd := !tc.wantAborted; reachedEnd != wantEnd {
				t.Errorf(
					"%s\nFakeT.Aborted() reached the end of the call\ngot:  %t\nwant: %t",
					packageName, reachedEnd, wantEnd,
				)
			}
		})
	}
}

func TestFakeT_Aborted__foreignPanic(t *testing.T) {
	// GIVEN: a FakeT, and a panic that is not one of its allowed aborts.
	fakeT := &FakeT{Abort: true}
	want := "not an abort"

	// THEN: the panic is re-raised rather than reported as an abort.
	defer func() {
		switch got := recover(); got {
		case nil:
			t.Errorf(
				"%s\nFakeT.Aborted() swallowed a foreign panic\nwant: %q",
				packageName, want,
			)
		case want:
			// Re-raised, as it should be.
		default:
			t.Errorf(
				"%s\nFakeT.Aborted() re-raised the wrong value\ngot:  %v\nwant: %q",
				packageName, got, want,
			)
		}
	}()

	// WHEN: Aborted calls something that panics with it.
	fakeT.Aborted(func() { panic(want) })
}
