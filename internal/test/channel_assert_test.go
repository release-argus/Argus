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

import "testing"

func TestAssertCallLog(t *testing.T) {
	tests := []struct {
		name        string
		result      bool
		noCapture   bool
		stdout      string
		want        bool
		stdoutRegex string
		wantLogs    []string
		wantErrors  []string
		wantFatals  []string
	}{
		{
			name:        "result and log match",
			result:      true,
			stdout:      "INFO: done\n",
			want:        true,
			stdoutRegex: `^INFO: done\n$`,
		},
		{
			name:        "log mismatch",
			result:      false,
			stdout:      "ERROR: broke\n",
			want:        false,
			stdoutRegex: `^FATAL: `,
			wantErrors:  []string{"stdout mismatch\ngot:  \"ERROR: broke\\n\"\nwant: \"^FATAL: \""},
		},
		{
			name:        "result mismatch logs the captured log",
			result:      false,
			stdout:      "FATAL: broke\n",
			want:        true,
			stdoutRegex: `^$`,
			wantLogs:    []string{`stdout: "FATAL: broke\n"`},
			wantFatals:  []string{"result mismatch\ngot:  false\nwant: true"},
		},
		{
			name:        "no capture/log unchecked",
			result:      false,
			noCapture:   true,
			want:        false,
			stdoutRegex: `^FATAL: `,
		},
		{
			name:       "no capture/result mismatch",
			result:     false,
			noCapture:  true,
			want:       true,
			wantFatals: []string{"result mismatch\ngot:  false\nwant: true"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			releaseCalls := 0
			releaseStdout := func() string {
				releaseCalls++
				return tc.stdout
			}
			wantReleaseCalls := 1
			if tc.noCapture {
				releaseStdout = nil
				wantReleaseCalls = 0
			}

			var fakeT FakeT
			AssertCallLog(
				&fakeT,
				tc.want,
				func() bool { return tc.result },
				make(chan string, 1),
				releaseStdout,
				tc.stdoutRegex,
			)

			if testErr := AssertSlicesEqual(t, fakeT.Logs, tc.wantLogs, "FakeT.Logs"); testErr != nil {
				t.Error(testErr)
			}
			if testErr := AssertSlicesEqual(t, fakeT.Errors, tc.wantErrors, "FakeT.Errors"); testErr != nil {
				t.Error(testErr)
			}
			if testErr := AssertSlicesEqual(t, fakeT.Fatals, tc.wantFatals, "FakeT.Fatals"); testErr != nil {
				t.Error(testErr)
			}

			if releaseCalls != wantReleaseCalls {
				t.Errorf(
					"releaseStdout() call count mismatch\ngot:  %d\nwant: %d",
					releaseCalls, wantReleaseCalls,
				)
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

			ch := make(chan string, len(tc.initialValues))
			for _, v := range tc.initialValues {
				ch <- v
			}

			drainExitCode(t, ch)

			if got := len(ch); got != tc.wantRemaining {
				t.Errorf("len(ch) mismatch\ngot:  %d\nwant: %d", got, tc.wantRemaining)
			}
		})
	}
}

func TestLogChannel(t *testing.T) {
	type message struct {
		ID   string
		Done bool
	}
	tests := []struct {
		name     string
		log      func(t TB)
		wantLogs []string
	}{
		{
			name: "empty channel",
			log: func(t TB) {
				LogChannel(t, "ch", make(chan string, 1))
			},
		},
		{
			name: "[]byte messages as text",
			log: func(t TB) {
				ch := make(chan []byte, 2)
				ch <- []byte(`{"id":"a"}`)
				ch <- []byte(`{"id":"b"}`)
				LogChannel(t, "ch", ch)
			},
			wantLogs: []string{
				`ch: {"id":"a"}`,
				`ch: {"id":"b"}`,
			},
		},
		{
			name: "struct messages with field names",
			log: func(t TB) {
				ch := make(chan message, 1)
				ch <- message{ID: "a", Done: true}
				LogChannel(t, "ch", ch)
			},
			wantLogs: []string{"ch: {ID:a Done:true}"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var fakeT FakeT
			tc.log(&fakeT)

			if testErr := AssertSlicesEqual(t, fakeT.Logs, tc.wantLogs, "FakeT.Logs"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestAssertChannelLen(t *testing.T) {
	tests := []struct {
		name       string
		queued     []string
		want       int
		wantOK     bool
		wantErrors []string
		wantLogs   []string
	}{
		{
			name:   "matching count",
			queued: []string{"a"},
			want:   1,
			wantOK: true,
		},
		{
			name:       "mismatch reports the count and drains the channel",
			queued:     []string{"a", "b"},
			want:       1,
			wantErrors: []string{"ch message count mismatch\ngot:  2\nwant: 1"},
			wantLogs:   []string{"ch: a", "ch: b"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ch := make(chan string, len(tc.queued))
			for _, msg := range tc.queued {
				ch <- msg
			}
			var fakeT FakeT

			ok := AssertChannelLen(&fakeT, "ch", ch, tc.want)

			if ok != tc.wantOK {
				t.Errorf("ok mismatch\ngot:  %t\nwant: %t", ok, tc.wantOK)
			}
			if testErr := AssertSlicesEqual(t, fakeT.Errors, tc.wantErrors, "FakeT.Errors"); testErr != nil {
				t.Error(testErr)
			}
			if testErr := AssertSlicesEqual(t, fakeT.Logs, tc.wantLogs, "FakeT.Logs"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
