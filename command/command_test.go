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

package command

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestCommand_ApplyTemplate(t *testing.T) {
	// GIVEN various Commands
	tests := map[string]struct {
		input         Command
		want          Command
		serviceStatus *status.Status
		latestVersion string
	}{
		"command with no templating and non-nil service status": {
			input:         Command{"ls", "-lah"},
			want:          Command{"ls", "-lah"},
			serviceStatus: &status.Status{},
			latestVersion: "1.2.3"},
		"command with no templating and nil service status": {
			input: Command{"ls", "-lah"},
			want:  Command{"ls", "-lah"}},
		"command with templating and nil service status": {
			input: Command{"ls", "-lah", "{{ version }}"},
			want:  Command{"ls", "-lah", "{{ version }}"}},
		"command with templating and non-nil service status": {
			input:         Command{"ls", "-lah", "{{ version }}"},
			want:          Command{"ls", "-lah", "1.2.3"},
			serviceStatus: &status.Status{},
			latestVersion: "1.2.3"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.latestVersion != "" {
				tc.serviceStatus.SetLatestVersion(tc.latestVersion, "", false)
			}

			// WHEN ApplyTemplate is called on the Command
			got := tc.input.ApplyTemplate(tc.serviceStatus)

			// THEN the result is expected
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want: %v\ngot:  %v",
					tc.want, got)
			}
		})
	}
}

func TestCommand_Exec(t *testing.T) {
	// GIVEN different Commands to execute
	tests := map[string]struct {
		cmd         Command
		err         error
		stdoutRegex string
	}{
		"command that will pass": {
			cmd:         Command{"date", "+%m-%d-%Y"},
			stdoutRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`},
		"command that will fail": {
			cmd:         Command{"false"},
			err:         fmt.Errorf("exit status 1"),
			stdoutRegex: `exit status 1\s+$`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			// WHEN Exec is called on it
			err := tc.cmd.Exec(logutil.LogFrom{})

			// THEN the stdout is expected
			if util.ErrorToString(err) != util.ErrorToString(tc.err) {
				t.Fatalf("error mismatch%q\ngot:%q",
					tc.err, err)
			}
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("want match for %q\nnot: %q",
					tc.stdoutRegex, stdout)
			}
		})
	}
}

func TestController_ExecIndex(t *testing.T) {
	// GIVEN a Controller with different Commands to execute
	announce := make(chan []byte, 8)
	controller := Controller{}
	svcStatus := status.New(
		&announce, nil, nil,
		"", "", "", "", "", "")
	svcStatus.ServiceID = test.StringPtr("service_id")
	controller.Init(
		svcStatus,
		&Slice{
			{"date", "+%m-%d-%Y"},
			{"false"}},
		nil,
		test.StringPtr("13m"),
	)
	tests := map[string]struct {
		index       int
		err         error
		stdoutRegex string
		noAnnounce  bool
	}{
		"command index out of range": {
			index:       2,
			stdoutRegex: `^$`,
			noAnnounce:  true},
		"command index that will pass": {
			index:       0,
			stdoutRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`},
		"command index that will fail": {
			index:       1,
			err:         fmt.Errorf("exit status 1"),
			stdoutRegex: `exit status 1\s+$`},
	}

	runNumber := 0
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			// WHEN the Command @index is executed
			err := controller.ExecIndex(logutil.LogFrom{}, tc.index)

			// THEN the stdout is expected
			// err
			if util.ErrorToString(err) != util.ErrorToString(tc.err) {
				t.Fatalf("errors differ\nwant: %s\ngot:  %s",
					tc.err, err)
			}
			// stdout
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Fatalf("want match for %q\nnot: %q",
					tc.stdoutRegex, stdout)
			}
			// announced
			if !tc.noAnnounce {
				runNumber++
			}
			if len(announce) != runNumber {
				t.Fatalf("Command run not announced\nat %d, want %d",
					len(announce), runNumber)
			}
		})
	}
}

func TestController_Exec(t *testing.T) {
	// GIVEN a Controller
	tests := map[string]struct {
		nilController bool
		commands      *Slice
		err           error
		stdoutRegex   string
		noAnnounce    bool
	}{
		"nil Controller": {
			nilController: true,
			stdoutRegex:   `^$`,
			noAnnounce:    true},
		"nil Command": {
			stdoutRegex: `^$`,
			noAnnounce:  true},
		"single Command": {
			stdoutRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`,
			commands: &Slice{
				{"date", "+%m-%d-%Y"}}},
		"multiple Commands": {
			err:         fmt.Errorf("exit status 1"),
			stdoutRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+.*'false'\s.*exit status 1\s+$`,
			commands: &Slice{
				{"date", "+%m-%d-%Y"},
				{"false"}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			announce := make(chan []byte, 8)
			controller := testController(&announce)

			// WHEN the Command @index is executed
			controller.Command = tc.commands
			if tc.nilController {
				controller = nil
			}
			err := controller.Exec(logutil.LogFrom{})

			// THEN the stdout is expected
			// err
			if util.ErrorToString(err) != util.ErrorToString(tc.err) {
				t.Fatalf("errors differ\nwant: %q\ngot:  %q",
					util.ErrorToString(tc.err), util.ErrorToString(err))
			}
			// stdout
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Fatalf("want match for %q\nnot: %q",
					tc.stdoutRegex, stdout)
			}
			// announced
			runNumber := 0
			if !tc.noAnnounce {
				runNumber = len(*controller.Command)
			}
			if len(announce) != runNumber {
				t.Fatalf("Command run not announced\nat %d, want %d",
					len(announce), runNumber)
			}
		})
	}
}
