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
	"regexp"
)

// AssertCallLog runs call, then checks it returns want and that the log releaseStdout captured
// matches stdoutRegex. When call returns anything else, the captured log is logged.
// A nil releaseStdout skips the log check.
func AssertCallLog(
	t TB,
	want bool,
	call func() bool,
	exitCodeChannel chan string,
	releaseStdout func() string,
	stdoutRegex string,
) {
	t.Helper()

	got := call()
	drainExitCode(t, exitCodeChannel)
	if releaseStdout == nil {
		if got != want {
			t.Fatalf("result mismatch\ngot:  %t\nwant: %t", got, want)
		}
		return
	}

	stdout := releaseStdout()
	if got != want {
		t.Logf("stdout: %q", stdout)
		t.Fatalf("result mismatch\ngot:  %t\nwant: %t", got, want)
		return
	}

	if !regexp.MustCompile(stdoutRegex).MatchString(stdout) {
		t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, stdoutRegex)
	}
}

func drainExitCode(t TB, ch <-chan string) {
	t.Helper()

	select {
	case msg := <-ch:
		t.Logf("drained %q from the exit code channel", msg)
	default:
	}
}

// LogChannel drains ch without blocking, logging each message under label: a []byte as text,
// anything else with %+v.
func LogChannel[T any](t TB, label string, ch chan T) {
	t.Helper()

	for len(ch) > 0 {
		switch m := any(<-ch).(type) {
		case []byte:
			t.Logf("%s: %s", label, m)
		default:
			t.Logf("%s: %+v", label, m)
		}
	}
}

// AssertChannelLen reports whether ch holds want messages; on a mismatch it fails and drains ch into the log.
func AssertChannelLen[T any](t TB, label string, ch chan T, want int) bool {
	t.Helper()

	if got := len(ch); got != want {
		t.Errorf("%s message count mismatch\ngot:  %d\nwant: %d", label, got, want)
		LogChannel(t, label, ch)
		return false
	}
	return true
}
