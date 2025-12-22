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

//go:build unit || integration

package test

import (
	"fmt"
	"testing"
	"time"
)

// OkMatch waits for a value on the channel and checks it against the expected value.
func OkMatch(
	t *testing.T,
	want bool,
	channel chan bool,
	exitCodeChannel chan string,
	releaseStdout func() string,
) error {
	t.Helper()
	timeout := 2500 * time.Millisecond

	select {
	case got := <-channel:
		drainExitCode(t, exitCodeChannel)
		// Fatal if ok value not expected.
		if got != want {
			if releaseStdout != nil {
				_ = releaseStdout()
			}
			return fmt.Errorf(
				"%s\nok mismatch:\nwant: %t\ngot:  %t",
				t.Name(), want, got,
			)
		}
	case <-time.After(timeout):
		if releaseStdout != nil {
			_ = releaseStdout()
		}
		drainExitCode(t, exitCodeChannel)
		return fmt.Errorf("%s\ntimeout (%s) waiting for result channel",
			t.Name(), timeout)
	}

	return nil
}

func drainExitCode(t *testing.T, ch <-chan string) {
	t.Helper()

	select {
	case msg := <-ch:
		t.Logf("%s\ndrained exitCodeChannel - %q",
			t.Name(), msg)
	default:
	}
}
