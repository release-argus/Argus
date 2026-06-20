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

package command

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestCommand_ApplyTemplate(t *testing.T) {
	// GIVEN: various Commands.
	tests := []struct {
		name          string
		input         Command
		want          Command
		serviceStatus *status.Status
		latestVersion string
	}{
		{
			name:          "command with no templating and non-nil service status",
			input:         Command{"ls", "-lah"},
			want:          Command{"ls", "-lah"},
			serviceStatus: &status.Status{},
			latestVersion: "1.2.3",
		},
		{
			name:          "command with templating and non-nil service status",
			input:         Command{"ls", "-lah", "{{ version }}"},
			want:          Command{"ls", "-lah", "1.2.3"},
			serviceStatus: &status.Status{},
			latestVersion: "1.2.3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.serviceStatus != nil {
				tc.serviceStatus.Init(
					1, 0, 0,
					status.ServiceInfo{
						ID: tc.name,
					},
					&dashboard.Options{},
				)
			}
			if tc.latestVersion != "" {
				tc.serviceStatus.SetLatestVersion(tc.latestVersion, "", false)
			}

			sInfo := tc.serviceStatus.GetServiceInfo()
			// WHEN: ApplyTemplate is called on the Command.
			got := tc.input.ApplyTemplate(sInfo)

			// THEN: the result is expected.
			if !util.AreSlicesEqual(got, tc.want) {
				t.Fatalf(
					"%s\nCommand.ApplyTemplate(%+v) on %+v mismatch\ngot:  %v\nwant: %v",
					packageName, sInfo, tc.input,
					got, tc.want,
				)
			}
		})
	}
}

func TestController_Exec(t *testing.T) {
	// GIVEN: a Controller.
	tests := []struct {
		name          string
		nilController bool
		commands      Commands
		err           error
		stdoutRegex   string
		noAnnounce    bool
	}{
		{
			name:          "nil Controller",
			nilController: true,
			stdoutRegex:   `^$`,
			noAnnounce:    true,
		},
		{
			name:        "nil Command",
			stdoutRegex: `^$`,
			noAnnounce:  true,
		},
		{
			name:        "single Command",
			stdoutRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`,
			commands: Commands{
				{"date", "+%m-%d-%Y"},
			},
		},
		{
			name:        "multiple Commands",
			err:         fmt.Errorf("exit status 1"),
			stdoutRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+.*'false'\s.*exit status 1\s+$`,
			commands: Commands{
				{"date", "+%m-%d-%Y"},
				{"false"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			announceChannel := make(chan []byte, 8)
			controller := testController(announceChannel)

			// WHEN: the Command @index is executed.
			controller.Command = tc.commands
			if tc.nilController {
				controller = nil
			}
			err := controller.Exec(logx.LogFrom{})

			prefix := fmt.Sprintf(
				"%s\nController.Exec(%+v)",
				packageName, tc.commands,
			)

			// THEN: the stdout is expected.
			// 	decode:
			gotErr := errfmt.FormatError(err)
			wantErr := errfmt.FormatError(tc.err)
			if gotErr != wantErr {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, gotErr, wantErr,
				)
			}
			// 	stdout:
			if stdout := releaseStdout(); !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Fatalf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}
			// 	announced:
			runNumber := 0
			if !tc.noAnnounce {
				runNumber = len(controller.Command)
			}
			if got := len(announceChannel); got != runNumber {
				t.Fatalf(
					"%s announce message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, runNumber,
				)
			}
		})
	}
}

func TestController_ExecIndex(t *testing.T) {
	// GIVEN: a Status.
	announceChannel := make(chan []byte, 8)
	svcStatus := status.New(
		announceChannel, nil, nil,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{},
	)
	svcStatus.ServiceInfo.ID = "service_id"

	// AND: a Controller with different Commands to execute.
	controller := NewController(
		svcStatus,
		Commands{
			{"date", "+%m-%d-%Y"},
			{"false"},
		},
		nil,
		test.Ptr("13m"),
	)

	tests := []struct {
		name        string
		index       int
		err         error
		stdoutRegex string
		noAnnounce  bool
	}{
		{
			name:        "command index out of range",
			index:       2,
			stdoutRegex: `^$`,
			noAnnounce:  true,
		},
		{
			name:        "command index that will pass",
			index:       0,
			stdoutRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`,
		},
		{
			name:        "command index that will fail",
			index:       1,
			err:         fmt.Errorf("exit status 1"),
			stdoutRegex: `exit status 1\s+$`,
		},
	}

	runNumber := 0
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			// WHEN: the Command @index is executed.
			err := controller.ExecIndex(
				logx.LogFrom{},
				tc.index,
				controller.ServiceStatus.GetServiceInfo(),
			)

			prefix := fmt.Sprintf(
				"%s\nController.ExecIndex(%d)",
				packageName, tc.index,
			)

			// THEN: the stdout is expected.
			// 	decode:
			gotErr := errfmt.FormatError(err)
			wantErr := errfmt.FormatError(tc.err)
			if gotErr != wantErr {
				t.Fatalf(
					"%s error mismatch\ngot:  %s\nwant: %s",
					prefix, gotErr, wantErr,
				)
			}
			// 	stdout:
			if stdout := releaseStdout(); !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Fatalf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix,
					stdout, tc.stdoutRegex,
				)
			}
			// 	announced:
			if !tc.noAnnounce {
				runNumber++
			}
			if got := len(announceChannel); got != runNumber {
				t.Fatalf(
					"%s Command run was not announced\ngot:  %d\nwant: %d",
					prefix, got, runNumber,
				)
			}
		})
	}
}

func TestCommand_Exec(t *testing.T) {
	// GIVEN: different Commands to execute.
	tests := []struct {
		name        string
		cmd         Command
		err         error
		stdoutRegex string
	}{
		{
			name:        "command that will pass",
			cmd:         Command{"date", "+%m-%d-%Y"},
			stdoutRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`,
		},
		{
			name:        "command that will fail",
			cmd:         Command{"false"},
			err:         fmt.Errorf("exit status 1"),
			stdoutRegex: `ERROR: exit status 1\s$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			// WHEN: Exec is called on it.
			err := tc.cmd.Exec(logx.LogFrom{})

			prefix := fmt.Sprintf(
				"%s\nCommand.Exec(%+v)",
				packageName, tc.cmd,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			wantErr := errfmt.FormatError(tc.err)
			if e != wantErr {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, wantErr,
				)
			}

			// AND: the stdout is expected.
			if stdout := releaseStdout(); !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Fatalf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}
		})
	}
}
