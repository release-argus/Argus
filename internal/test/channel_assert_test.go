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

package test

import (
	"fmt"
	"testing"
)

func TestAssertChannelBool(t *testing.T) {
	// GIVEN: a result channel with a possible bool result,
	// the exit channel state, and the releaseStdout function.
	tests := []struct {
		name                    string
		sendResult              *bool
		want                    bool
		exitCodeMsgs            []string // Messages to pre-fill the exitCodeChannel with.
		provideReleaseStdout    bool     // Whether to provide a releaseStdout function to AssertChannelBool.
		wantReleaseStdoutCalled bool     // Whether releaseStdout is expected to be called.
		expectError             bool     // Whether an error is expected from AssertChannelBool.
	}{
		{
			name:        "ok matches want",
			sendResult:  new(true),
			want:        true,
			expectError: false,
		},
		{
			name:        "ok mismatches want",
			sendResult:  new(false),
			want:        true,
			expectError: true,
		},
		{
			name:        "timeout waiting for result",
			sendResult:  nil,
			want:        true,
			expectError: true,
		},
		{
			name:                    "timeout waiting for result, releaseStdout",
			sendResult:              nil,
			want:                    true,
			provideReleaseStdout:    true,
			wantReleaseStdoutCalled: true,
			expectError:             true,
		},
		{
			name:         "exitCodeChannel drained on success",
			sendResult:   new(true),
			want:         true,
			exitCodeMsgs: []string{"1", "2"},
			expectError:  false,
		},
		{
			name:         "exitCodeChannel drained on fatal",
			sendResult:   new(false),
			want:         true,
			exitCodeMsgs: []string{"1", "2"},
			expectError:  true,
		},
		{
			name:                    "releaseStdout called when result!=want",
			sendResult:              new(false),
			want:                    true,
			provideReleaseStdout:    true,
			wantReleaseStdoutCalled: true,
			expectError:             true,
		},
		{
			name:                    "releaseStdout not called when result==want",
			sendResult:              new(true),
			want:                    true,
			provideReleaseStdout:    true,
			wantReleaseStdoutCalled: false,
			expectError:             false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: result/exit channels ;
			resultCh := make(chan bool, 1)
			exitCh := make(chan string, len(tc.exitCodeMsgs))
			// Fill exit channel with exit codes.
			for _, msg := range tc.exitCodeMsgs {
				exitCh <- msg
			}
			// Send result on result channel.
			if tc.sendResult != nil {
				resultCh <- *tc.sendResult
			}

			// AND: a releaseStdout function local to this subtest.
			var releaseStdoutCalled bool
			var releaseStdout func() string
			if tc.provideReleaseStdout {
				releaseStdout = func() string {
					releaseStdoutCalled = true
					return ""
				}
			}

			// WHEN: AssertChannelBool is called.
			err := AssertChannelBool(
				t,
				tc.want,
				resultCh,
				exitCh,
				releaseStdout,
			)

			prefix := fmt.Sprintf("%s\nAssertChannelBool()", packageName)

			// THEN: any error is expected.
			if got := err != nil; got != tc.expectError {
				t.Fatalf(
					"%s error mismatch\ngot:  %v\nwant: %v",
					packageName, got, tc.expectError,
				)
			}

			// AND: exitCodeChannel is fully drained
			wantLength := max(0, len(tc.exitCodeMsgs)-1)
			if got := len(exitCh); got != wantLength {
				t.Errorf(
					"%s exitCodeChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, wantLength,
				)
			}

			// AND: releaseStdout is called when expected.
			if tc.provideReleaseStdout {
				if releaseStdoutCalled != tc.wantReleaseStdoutCalled {
					t.Errorf(
						"%s releaseStdout called mismatch\ngot:  %v\nwant: %v",
						prefix, releaseStdoutCalled, tc.wantReleaseStdoutCalled,
					)
				}
			}
		})
	}
}

func TestDrainExitCode(t *testing.T) {
	tests := []struct {
		name          string
		initialValues []string
		wantRemaining int
	}{
		{
			name:          "channel with value",
			initialValues: []string{"42"},
			wantRemaining: 0,
		},
		{
			name:          "channel empty",
			initialValues: []string{},
			wantRemaining: 0,
		},
		{
			name:          "channel with multiple values, drains only one",
			initialValues: []string{"first", "second"},
			wantRemaining: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: a channel with predefined initial values.
			ch := make(chan string, len(tc.initialValues))
			for _, v := range tc.initialValues {
				ch <- v
			}

			// WHEN: drainExitCode is called.
			drainExitCode(t, ch)

			// THEN: at most one value has been removed.
			if got := len(ch); got != tc.wantRemaining {
				t.Errorf(
					"%s\ndrainExitCode() remaining ExitCode channel values mismatch\ngot:  %d\nwant: %d",
					packageName, got, tc.wantRemaining,
				)
			}
		})
	}
}
