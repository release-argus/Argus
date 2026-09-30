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

import "fmt"

// TB is the subset of [testing.TB] used by these helpers, so that their
// reporting paths can be exercised with a [FakeT].
type TB interface {
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Helper()
	Skipf(format string, args ...any)
}

// abortFakeT is what an aborting [FakeT] panics with.
type abortFakeT struct{}

// FakeT records what a helper reports rather than failing the test, so that a
// helper's own failure paths can be asserted on.
type FakeT struct {
	Errors []string
	Fatals []string
	Skips  []string

	// Abort makes Fatalf and Skipf stop the caller, standing in for the
	// [runtime.Goexit] a real [testing.T] performs. Read the outcome with
	// [FakeT.Aborted].
	Abort bool
}

func (f *FakeT) Errorf(format string, args ...any) {
	f.Errors = append(f.Errors, fmt.Sprintf(format, args...))
}

func (f *FakeT) Fatalf(format string, args ...any) {
	f.Fatals = append(f.Fatals, fmt.Sprintf(format, args...))
	f.abort()
}

func (f *FakeT) Skipf(format string, args ...any) {
	f.Skips = append(f.Skips, fmt.Sprintf(format, args...))
	f.abort()
}

func (f *FakeT) Helper() {
	// No-op.
}

func (f *FakeT) abort() {
	if f.Abort {
		panic(abortFakeT{})
	}
}

// Aborted calls fn, reporting whether a Fatalf or Skipf stopped it.
//
// It only ever reports true when [FakeT.Abort] is set. Any other panic is
// re-raised, so a genuine failure is not mistaken for an abort.
func (f *FakeT) Aborted(fn func()) (aborted bool) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if _, isAbort := r.(abortFakeT); !isAbort {
			panic(r)
		}
		aborted = true
	}()

	fn()
	return false
}
