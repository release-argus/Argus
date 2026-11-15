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
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestCommand_ApplyTemplate(t *testing.T) {
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

			tc.serviceStatus.Init(
				1, 0, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			tc.serviceStatus.SetLatestVersion(tc.latestVersion, "", false)

			sInfo := tc.serviceStatus.GetServiceInfo()

			got := tc.input.ApplyTemplate(sInfo)

			if !util.AreSlicesEqual(got, tc.want) {
				t.Fatalf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestController_Exec(t *testing.T) {
	tests := []struct {
		name          string
		nilController bool
		commands      Commands
		errRegex      string
		stdoutRegexes []string
		noAnnounce    bool
	}{
		{
			name:          "nil Controller",
			nilController: true,
			errRegex:      `^$`,
			stdoutRegexes: []string{`^$`},
			noAnnounce:    true,
		},
		{
			name:          "nil Command",
			errRegex:      `^$`,
			stdoutRegexes: []string{`^$`},
			noAnnounce:    true,
		},
		{
			name:          "single Command",
			errRegex:      `^$`,
			stdoutRegexes: []string{`[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`},
			commands: Commands{
				{"date", "+%m-%d-%Y"},
			},
		},
		{
			name:     "multiple Commands",
			errRegex: `^exit status 1$`,
			stdoutRegexes: []string{
				`Executing "date \+%m-%d-%Y"`,
				`Executing "false"`,
				`[0-9]{2}-[0-9]{2}-[0-9]{4}`,
				`exit status 1`,
			},
			commands: Commands{
				{"date", "+%m-%d-%Y"},
				{"false"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			announceChannel := make(chan []byte, 8)
			controller := testController(announceChannel)

			controller.Command = tc.commands
			if tc.nilController {
				controller = nil
			}
			err := controller.Exec(logx.LogFrom{})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			stdout := releaseStdout()
			for _, stdoutRegex := range tc.stdoutRegexes {
				if !util.RegexCheck(stdoutRegex, stdout) {
					t.Fatalf("stdout mismatch\ngot:  %q\nwant: %q", stdout, stdoutRegex)
				}
			}

			runNumber := 0
			if !tc.noAnnounce {
				runNumber = len(controller.Command)
			}
			if got := len(announceChannel); got != runNumber {
				test.LogChannel(t, "Status.AnnounceChannel", announceChannel)
				t.Fatalf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, runNumber,
				)
			}
		})
	}
}

func TestController_ExecIndex(t *testing.T) {
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

	controller := NewController(
		svcStatus,
		Commands{
			{"date", "+%m-%d-%Y"},
			{"false"},
		},
		nil,
		new("13m"),
	)

	tests := []struct {
		name        string
		index       int
		errRegex    string
		stdoutRegex string
		noAnnounce  bool
	}{
		{
			name:        "command index out of range",
			index:       2,
			errRegex:    `^$`,
			stdoutRegex: `^$`,
			noAnnounce:  true,
		},
		{
			name:        "command index that will pass",
			index:       0,
			errRegex:    `^$`,
			stdoutRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`,
		},
		{
			name:        "command index that will fail",
			index:       1,
			errRegex:    `^exit status 1$`,
			stdoutRegex: `exit status 1\s+$`,
		},
	}

	runNumber := 0
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			err := controller.ExecIndex(
				logx.LogFrom{},
				tc.index,
				controller.ServiceStatus.GetServiceInfo(),
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if stdout := releaseStdout(); !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Fatalf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
			}

			if !tc.noAnnounce {
				runNumber++
			}
			if got := len(announceChannel); got != runNumber {
				test.LogChannel(t, "Status.AnnounceChannel", announceChannel)
				t.Fatalf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, runNumber,
				)
			}
		})
	}
}

func TestCommand_Exec(t *testing.T) {
	tests := []struct {
		name        string
		cmd         Command
		errRegex    string
		stdoutRegex string
	}{
		{
			name:        "command that will pass",
			cmd:         Command{"date", "+%m-%d-%Y"},
			errRegex:    `^$`,
			stdoutRegex: `[0-9]{2}-[0-9]{2}-[0-9]{4}\s+$`,
		},
		{
			name:        "command that will fail",
			cmd:         Command{"false"},
			errRegex:    `^exit status 1$`,
			stdoutRegex: `ERROR: exit status 1\s$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			err := tc.cmd.Exec(logx.LogFrom{})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if stdout := releaseStdout(); !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Fatalf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
			}
		})
	}
}
