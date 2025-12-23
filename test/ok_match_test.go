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

package test

import "testing"

func TestOkMatch(t *testing.T) {
	calledStdout := make(map[string]bool)
	tests := map[string]struct {
		sendResult         *bool
		want               bool
		exitCodeMsgs       []string
		releaseStdout      func() string
		checkReleaseStdout func() bool
		expectError        bool
	}{
		"ok matches want": {
			sendResult:  BoolPtr(true),
			want:        true,
			expectError: false,
		},
		"ok mismatches want": {
			sendResult:  BoolPtr(false),
			want:        true,
			expectError: true,
		},
		"timeout waiting for result": {
			sendResult:  nil,
			want:        true,
			expectError: true,
		},
		"timeout waiting for result, releaseStdout": {
			sendResult: nil,
			want:       true,
			releaseStdout: func() string {
				calledStdout["1"] = true
				return ""
			},
			checkReleaseStdout: func() bool {
				return calledStdout["1"] == true
			},
			expectError: true,
		},
		"exitCodeChannel drained on success": {
			sendResult:   BoolPtr(true),
			want:         true,
			exitCodeMsgs: []string{"1", "2"},
			expectError:  false,
		},
		"exitCodeChannel drained on fatal": {
			sendResult:   BoolPtr(false),
			want:         true,
			exitCodeMsgs: []string{"1", "2"},
			expectError:  true,
		},
		"releaseStdout called when result!=want": {
			sendResult: BoolPtr(false),
			want:       true,
			releaseStdout: func() string {
				calledStdout["1"] = true
				return ""
			},
			checkReleaseStdout: func() bool {
				return calledStdout["1"] == true
			},
			expectError: true,
		},
		"releaseStdout not called when result==want": {
			sendResult: BoolPtr(true),
			want:       true,
			releaseStdout: func() string {
				calledStdout["2"] = true
				return ""
			},
			checkReleaseStdout: func() bool {
				return calledStdout["2"] == false
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			resultCh := make(chan bool, 1)
			exitCh := make(chan string, len(tc.exitCodeMsgs))

			for _, msg := range tc.exitCodeMsgs {
				exitCh <- msg
			}

			if tc.sendResult != nil {
				resultCh <- *tc.sendResult
			}

			// WHEN
			err := OkMatch(t, tc.want, resultCh, exitCh, tc.releaseStdout)

			// THEN
			if tc.expectError != (err != nil) {
				t.Errorf("%s\nerror mismatch\nwant: %v\ngot: %v",
					packageName, tc.expectError, err)
			}

			// AND exitCodeChannel is fully drained
			wantLength := max(0, len(tc.exitCodeMsgs)-1)
			if len(exitCh) != wantLength {
				t.Errorf(
					"%s\nexitCodeChannel length mismatch\nwant: %d\ngot:  %d",
					packageName, wantLength, len(exitCh))
			}
			// AND releaseStdout is called when expected.
			if tc.checkReleaseStdout != nil {
				if val := tc.checkReleaseStdout(); !val {
					t.Errorf("%s\ncheckReleaseStdout mismatch\nwant: true\ngot:  %v",
						packageName, val)
				}
			}
		})
	}
}

func Test_drainExitCode(t *testing.T) {
	tests := map[string]struct {
		initialValues []string
		wantRemaining int
	}{
		"channel with value": {
			initialValues: []string{"42"},
			wantRemaining: 0,
		},
		"channel empty": {
			initialValues: []string{},
			wantRemaining: 0,
		},
		"channel with multiple values, drains only one": {
			initialValues: []string{"first", "second"},
			wantRemaining: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// GIVEN a channel with predefined initial values.
			ch := make(chan string, len(tc.initialValues))
			for _, v := range tc.initialValues {
				ch <- v
			}

			// WHEN drainExitCode is called.
			drainExitCode(t, ch)

			// THEN at most one value has been removed.
			if len(ch) != tc.wantRemaining {
				t.Errorf("%s\nremaining values mismatch\nwant: %d\ngot:  %d",
					packageName, tc.wantRemaining, len(ch))
			}
		})
	}
}
