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

package logx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestInit(t *testing.T) {
	// GIVEN: a log level and timestamps setting.
	tests := []struct {
		level      string
		timestamps bool
	}{
		{level: "ERROR", timestamps: true},
		{level: "WARN", timestamps: false},
		{level: "INFO", timestamps: true},
		{level: "VERBOSE", timestamps: false},
		{level: "DEBUG", timestamps: false},
	}

	hadTimestamps := logger.timestamps
	t.Cleanup(func() {
		if logger.timestamps != hadTimestamps {
			logger.SetTimestamps(false)
		}
	})

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, timestamps=%t",
			tc.level, tc.timestamps,
		)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel because of the once variable.

			// Reset the once variable and logger for each test.
			once = sync.Once{}
			logger = nil

			// WHEN: Init is called.
			exitCodeChannel := Init(tc.level, tc.timestamps)

			prefix := fmt.Sprintf(
				"%s\nInit(%s)",
				packageName, name,
			)

			// THEN: the Log should be initialised correctly.
			if logger == nil {
				t.Fatalf("%s logger was not initialised", prefix)
			}
			fieldTests := []test.FieldAssertion{
				{Name: "Level", Got: logger.Level.Load(), Want: levelMap[tc.level], Mode: test.CompareEqual},
				{Name: "timestamps", Got: logger.timestamps, Want: tc.timestamps, Mode: test.CompareEqual},
				{Name: "ExitCodeChannel()", Got: logger.exitCodeChannel, Want: exitCodeChannel, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "logger"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	// GIVEN: a new Logger is wanted.
	tests := []struct {
		level      string
		timestamps bool
	}{
		{level: "ERROR", timestamps: true},
		{level: "ERROR", timestamps: false},
		{level: "WARN", timestamps: true},
		{level: "WARN", timestamps: false},
		{level: "INFO", timestamps: true},
		{level: "INFO", timestamps: false},
		{level: "VERBOSE", timestamps: true},
		{level: "VERBOSE", timestamps: false},
		{level: "DEBUG", timestamps: true},
		{level: "DEBUG", timestamps: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, timestamps=%t",
			tc.level, tc.timestamps,
		)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN: NewLogger is called.
			tLogger := NewLogger(tc.level, tc.timestamps)

			prefix := fmt.Sprintf(
				"%s\nNewLogger(%s)",
				packageName, name,
			)

			// THEN: the correct Logger is returned.
			fieldTests := []test.FieldAssertion{
				{Name: "Level", Got: tLogger.Level.Load(), Want: levelMap[tc.level], Mode: test.CompareEqual},
				{Name: "timestamps", Got: tLogger.timestamps, Want: tc.timestamps, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "logger"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLogger_SetExitCodeChannel(t *testing.T) {
	// GIVEN: a Logger with no exit code channel.
	tLogger := NewLogger("INFO", true)
	tLogger.exitCodeChannel = nil

	// AND: a slice of exit code channels.
	channels := []chan string{
		make(chan string),
		make(chan string),
	}

	for i, channel := range channels {
		// WHEN: SetExitCodeChannel is called.
		tLogger.SetExitCodeChannel(channel)

		// THEN: the shutdown handler is updated.
		if tLogger.exitCodeChannel != channel {
			t.Fatalf(
				"%s\nLogger.SetExitCodeChannel(%p) iteration %d, exitCodeChannel pointer mismatch\ngot:  %p\nwant: %p",
				packageName, channel, i,
				tLogger.exitCodeChannel, channel,
			)
		}
		if got := tLogger.ExitCodeChannel(); got != channel {
			t.Fatalf(
				"%s\nLogger.exitCodeChannel returned a different pointer than Logger,ExitCodeChannel() iteration %d\ngot:  %p\nwant: %p",
				packageName, i,
				got, channel,
			)
		}
	}
}

func TestLogger_SetLevel(t *testing.T) {
	// GIVEN: a Logger and various new log levels.
	tests := []struct {
		level string
		ok    bool
	}{
		{level: "ERROR", ok: true},
		{level: "WARN", ok: true},
		{level: "INFO", ok: true},
		{level: "VERBOSE", ok: true},
		{level: "DEBUG", ok: true},
		{level: "verbose", ok: true},
		{level: "vERbOse", ok: true},
		{level: "PINEAPPLE", ok: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"%s - ok=%t",
			tc.level, tc.ok,
		)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			tLogger := NewLogger("INFO", false)
			releaseStdout := test.CaptureLog(t, tLogger)

			// WHEN: SetLevel is called.
			tLogger.SetLevel(tc.level)

			prefix := fmt.Sprintf(
				"%s\nLogger.SetLevel(%q)",
				packageName, tc.level,
			)

			// THEN: the level is updated.
			level := tLogger.Level.Load()
			if tc.ok {
				if level != levelMap[strings.ToUpper(tc.level)] {
					t.Errorf(
						"%s\nLogger.Level value mismatch (valid level)\ngot:  %d\nwant: %d",
						prefix, level, levelMap[tc.level],
					)
				}
			} else {
				if level != 0 {
					t.Errorf(
						"%s Logger.Level value shouldn't have changed\ngot:  %d\nwant: %d",
						prefix, level, 0,
					)
				}
			}

			// AND: errors are logged for invalid levels.
			stdout := releaseStdout()
			errRegex := `not a valid log\.level`
			stdoutOk := !util.RegexCheck(errRegex, stdout)
			if stdoutOk != tc.ok {
				if tc.ok {
					t.Errorf(
						"%s shouldn't have logged an invalid level error\ngot:     %q\nwant: NO %q",
						prefix, stdout, errRegex,
					)
				} else {
					t.Errorf(
						"%s invalid level error mismatch\ngot:  %q\nwant: %q",
						prefix, stdout, errRegex,
					)
				}
			}
		})
	}
}

func TestLogger_IsLevel(t *testing.T) {
	// GIVEN: you have a valid Logger.
	tests := []struct {
		startLevel, testLevel string
		want                  bool
	}{
		{startLevel: "ERROR", testLevel: "DEBUG", want: false},
		{startLevel: "ERROR", testLevel: "VERBOSE", want: false},
		{startLevel: "ERROR", testLevel: "INFO", want: false},
		{startLevel: "ERROR", testLevel: "WARN", want: false},
		{startLevel: "ERROR", testLevel: "ERROR", want: true},
		{startLevel: "WARN", testLevel: "DEBUG", want: false},
		{startLevel: "WARN", testLevel: "VERBOSE", want: false},
		{startLevel: "WARN", testLevel: "INFO", want: false},
		{startLevel: "WARN", testLevel: "WARN", want: true},
		{startLevel: "WARN", testLevel: "ERROR", want: false},
		{startLevel: "INFO", testLevel: "DEBUG", want: false},
		{startLevel: "INFO", testLevel: "VERBOSE", want: false},
		{startLevel: "INFO", testLevel: "INFO", want: true},
		{startLevel: "INFO", testLevel: "WARN", want: false},
		{startLevel: "INFO", testLevel: "ERROR", want: false},
		{startLevel: "VERBOSE", testLevel: "DEBUG", want: false},
		{startLevel: "VERBOSE", testLevel: "VERBOSE", want: true},
		{startLevel: "VERBOSE", testLevel: "INFO", want: false},
		{startLevel: "VERBOSE", testLevel: "WARN", want: false},
		{startLevel: "VERBOSE", testLevel: "ERROR", want: false},
		{startLevel: "DEBUG", testLevel: "DEBUG", want: true},
		{startLevel: "DEBUG", testLevel: "VERBOSE", want: false},
		{startLevel: "DEBUG", testLevel: "INFO", want: false},
		{startLevel: "DEBUG", testLevel: "WARN", want: false},
		{startLevel: "DEBUG", testLevel: "ERROR", want: false},
		{startLevel: "DEBUG", testLevel: "FOO", want: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, test=%s - want=%t",
			tc.testLevel, tc.testLevel, tc.want,
		)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tLogger := NewLogger(tc.startLevel, false)

			// WHEN: IsLevel is called to check the given level.
			got := tLogger.IsLevel(tc.testLevel)

			// THEN: the correct response is returned.
			if got != tc.want {
				t.Errorf(
					"%s\nLogger(level=%s).IsLevel(%q) value mismatch\ngot:  %t\nwant: %t",
					packageName, tc.startLevel, tc.testLevel,
					got, tc.want,
				)
			}
		})
	}
}

func TestLogger_SetTimestamps(t *testing.T) {
	// GIVEN: a Logger and various tests.
	const msg = "timestamp toggle test"
	tests := []struct {
		start    bool
		changeTo bool
	}{
		{start: true, changeTo: true},
		{start: true, changeTo: false},
		{start: false, changeTo: false},
		{start: false, changeTo: true},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"%t to %t",
			tc.start, tc.changeTo,
		)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			tLogger := NewLogger("INFO", tc.start)
			tLogger.SetOutput(buf)

			// WHEN: SetTimestamps is called.
			tLogger.SetTimestamps(tc.changeTo)

			prefix := fmt.Sprintf(
				"%s\nLogger(timestamps=%t).SetTimestamps(%t)",
				packageName, tc.start, tc.changeTo,
			)

			// THEN: the timestamps flag is set correctly.
			if tLogger.timestamps != tc.changeTo {
				t.Errorf(
					"%s .timestamps value mismatch\ngot:  %t\nwant: %t",
					prefix,
					tLogger.timestamps, tc.changeTo,
				)
			}

			// AND: subsequent log output reflects the updated setting.
			tLogger.logMessage(msg)
			got := buf.String()
			wantRegex := fmt.Sprintf("^%s\n$", msg)
			if tc.changeTo {
				wantRegex = strings.Replace(
					wantRegex,
					`^`,
					"^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} ",
					1,
				)
			}
			if !util.RegexCheck(wantRegex, got) {
				t.Errorf(
					"%s .logMessage(%q) output mismatch\ngot:  %q\nwant: %q",
					prefix, msg, got, wantRegex,
				)
			}
		})
	}
}

func TestLogFrom_String(t *testing.T) {
	// GIVEN: a LogFrom.
	tests := []struct {
		name    string
		logFrom LogFrom
		want    string
	}{
		{
			name:    "primary and secondary",
			logFrom: LogFrom{Primary: "foo", Secondary: "bar"},
			want:    "foo (bar), ",
		},
		{
			name:    "only primary",
			logFrom: LogFrom{Primary: "foo"},
			want:    "foo, ",
		},
		{
			name:    "only secondary",
			logFrom: LogFrom{Secondary: "bar"},
			want:    "bar, ",
		},
		{
			name:    "empty logFrom",
			logFrom: LogFrom{},
			want:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: String() is called on it.
			got := tc.logFrom.String()

			// THEN: it is stringified as expected
			if got != tc.want {
				t.Errorf(
					"%s\nLogFrom.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLogger_Output(t *testing.T) {
	// GIVEN: an initial Output, and an Output to change to.
	tests := []struct {
		name          string
		initialOutput io.Writer
		newOutput     io.Writer
	}{
		{
			name:          "set to stdout",
			initialOutput: new(bytes.Buffer),
			newOutput:     os.Stdout,
		},
		{
			name:          "set to stderr",
			initialOutput: new(bytes.Buffer),
			newOutput:     os.Stderr,
		},
		{
			name:          "set to buffer",
			initialOutput: os.Stdout,
			newOutput:     new(bytes.Buffer),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Logger with this initial output.
			buf := &bytes.Buffer{}
			tLogger := &Logger{
				writer: log.New(buf, "", 0),
				out:    tc.initialOutput,
			}

			// WHEN: we call SetOutput with the new output.
			tLogger.SetOutput(tc.newOutput)

			// THEN: GetOutput returns the new writer.
			got := tLogger.GetOutput()
			if got != tc.newOutput {
				t.Errorf(
					"%s\nLogger.GetOutput() result mismatch\ngot:  %v\nwant: %v",
					packageName, got, tc.newOutput,
				)
			}

			// AND: writing a message uses the new writer.
			testMsg := "test message"
			tLogger.logMessage(testMsg)
			switch w := tc.newOutput.(type) {
			case *bytes.Buffer:
				if !strings.Contains(w.String(), testMsg) {
					t.Errorf(
						"%s\nLogger.SetOutput(x), .logMessage(%q) did not write to new buffer",
						packageName, testMsg,
					)
				}
			}
		})
	}
}

func TestLogger_Fatal(t *testing.T) {
	// GIVEN: a Logger and message.
	msg := "argus"
	tests := []struct {
		level      string
		timestamps bool
	}{
		{level: "ERROR", timestamps: true},
		{level: "ERROR", timestamps: false},
		{level: "WARN", timestamps: true},
		{level: "WARN", timestamps: false},
		{level: "INFO", timestamps: true},
		{level: "INFO", timestamps: false},
		{level: "VERBOSE", timestamps: true},
		{level: "VERBOSE", timestamps: false},
		{level: "DEBUG", timestamps: true},
		{level: "DEBUG", timestamps: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, timestamps=%t",
			tc.level, tc.timestamps,
		)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			tLogger := NewLogger(tc.level, tc.timestamps)
			releaseStdout := test.CaptureLog(t, tLogger)

			// WHEN: Fatal is called.
			tLogger.Fatal(errors.New(msg), LogFrom{})

			// THEN: message was logged.
			stdout := releaseStdout()
			regex := fmt.Sprintf("^FATAL: %s\n$", msg)
			if tc.timestamps {
				regex = strings.Replace(regex, `^`, timestampRegex, 1)
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf(
					"%s\nLogger(%s).Fatal message mismatch\ngot:  %q\nwant: %q",
					packageName, name,
					stdout, regex,
				)
			}

			// AND: the exit code was sent to the channel.
			select {
			case <-tLogger.exitCodeChannel:
				return
			default:
				t.Errorf(
					"%s\nLogger(%q).Fatal didn't send to exitCodeChannel",
					packageName, name,
				)
			}
		})
	}
}

func TestLogger_Error(t *testing.T) {
	// GIVEN: a Logger.
	tests := []struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool
	}{
		{level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "ERROR", timestamps: true, otherCondition: false, shouldPrint: false},
		{level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "WARN", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "WARN", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "WARN", timestamps: true, otherCondition: false, shouldPrint: false},
		{level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "INFO", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "INFO", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "INFO", timestamps: true, otherCondition: false, shouldPrint: false},
		{level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "VERBOSE", timestamps: true, otherCondition: false, shouldPrint: false},
		{level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "DEBUG", timestamps: true, otherCondition: false, shouldPrint: false},
		{level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, timestamps=%t, otherCondition=%t - print=%t",
			tc.level, tc.timestamps, tc.otherCondition, tc.shouldPrint,
		)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			tLogger := NewLogger(tc.level, tc.timestamps)
			releaseStdout := test.CaptureLog(t, tLogger)

			// WHEN: Error is called.
			tLogger.Error(errors.New(name), LogFrom{}, tc.otherCondition)

			// THEN: msg was logged if shouldPrint, with/without timestamps.
			stdout := releaseStdout()
			regex := `^$`
			if tc.shouldPrint {
				regex = fmt.Sprintf("^ERROR: %s\n$", name)
				if tc.timestamps {
					regex = strings.Replace(regex, `^`, timestampRegex, 1)
				}
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf(
					"%s\nLogger.Error(level=%q, timestamps=%t, log=%t) stdout mismatch\ngot:  %q\nwant: %q",
					packageName, tc.level, tc.timestamps, tc.otherCondition,
					stdout, regex,
				)
			}
		})
	}
}

func TestLogger_Warn(t *testing.T) {
	// GIVEN: a Logger.
	tests := []struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool
	}{
		{level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: false},
		{level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: false},
		{level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "WARN", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "WARN", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "INFO", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "INFO", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, timestamps=%t, otherCondition=%t - print=%t",
			tc.level, tc.timestamps, tc.otherCondition, tc.shouldPrint,
		)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			tLogger := NewLogger(tc.level, tc.timestamps)
			releaseStdout := test.CaptureLog(t, tLogger)

			// WHEN: Warn is called.
			tLogger.Warn(errors.New(name), LogFrom{}, tc.otherCondition)

			// THEN: msg was logged if shouldPrint, with/without timestamps.
			stdout := releaseStdout()
			regex := "^$"
			if tc.shouldPrint {
				regex = fmt.Sprintf("^WARNING: %s\n$", name)
				if tc.timestamps {
					regex = strings.Replace(regex, `^`, timestampRegex, 1)
				}
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf(
					"%s\nLogger.Warn(level=%s, timestamps=%t, otherCondition=%t) stdout mismatch\ngot:  %q\nwant: %q",
					packageName, tc.level, tc.timestamps, tc.otherCondition,
					stdout, regex,
				)
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	// GIVEN: a Logger.
	tests := []struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool
	}{
		{level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: false},
		{level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: false},
		{level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "WARN", timestamps: true, otherCondition: true, shouldPrint: false},
		{level: "WARN", timestamps: false, otherCondition: true, shouldPrint: false},
		{level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "INFO", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "INFO", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, timestamps=%t, otherCondition=%t - print=%t",
			tc.level, tc.timestamps, tc.otherCondition, tc.shouldPrint,
		)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			tLogger := NewLogger(tc.level, tc.timestamps)
			releaseStdout := test.CaptureLog(t, tLogger)

			// WHEN: Info is called.
			tLogger.Info(errors.New(name), LogFrom{}, tc.otherCondition)

			// THEN: msg was logged if shouldPrint, with/without timestamps.
			stdout := releaseStdout()
			regex := `^$`
			if tc.shouldPrint {
				regex = fmt.Sprintf("^INFO: %s\n$", name)
				if tc.timestamps {
					regex = strings.Replace(regex, `^`, timestampRegex, 1)
				}
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf(
					"%s\nLogger.Info(level=%s, timestamps=%t, otherCondition=%t) stdout mismatch\ngot:  %q\nwant: %q",
					packageName, tc.level, tc.timestamps, tc.otherCondition,
					stdout, regex,
				)
			}
		})
	}
}

func TestLogger_Verbose(t *testing.T) {
	// GIVEN: a Logger.
	tests := []struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool
	}{
		{level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: false},
		{level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: false},
		{level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "WARN", timestamps: true, otherCondition: true, shouldPrint: false},
		{level: "WARN", timestamps: false, otherCondition: true, shouldPrint: false},
		{level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "INFO", timestamps: true, otherCondition: true, shouldPrint: false},
		{level: "INFO", timestamps: false, otherCondition: true, shouldPrint: false},
		{level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, timestamps=%t, otherCondition=%t - print=%t",
			tc.level, tc.timestamps, tc.otherCondition, tc.timestamps,
		)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			tLogger := NewLogger(tc.level, tc.timestamps)
			releaseStdout := test.CaptureLog(t, tLogger)

			// WHEN: Verbose is called.
			tLogger.Verbose(errors.New(name), LogFrom{}, tc.otherCondition)

			// THEN: msg was logged if shouldPrint, with/without timestamps.
			stdout := releaseStdout()
			regex := `^$`
			if tc.shouldPrint {
				regex = fmt.Sprintf("^VERBOSE: %s\n$", name)
				if tc.timestamps {
					regex = strings.Replace(regex, `^`, timestampRegex, 1)
				}
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf(
					"%s\nLogger.Verbose(level=%s, timestamps=%t, otherCondition=%t) stdout mismatch\ngot:  %q\nwant: %q",
					packageName, tc.level, tc.timestamps, tc.otherCondition,
					stdout, regex,
				)
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	// GIVEN: a Logger and message.
	tests := []struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool
	}{
		{level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: false},
		{level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: false},
		{level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "WARN", timestamps: true, otherCondition: true, shouldPrint: false},
		{level: "WARN", timestamps: false, otherCondition: true, shouldPrint: false},
		{level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "INFO", timestamps: true, otherCondition: true, shouldPrint: false},
		{level: "INFO", timestamps: false, otherCondition: true, shouldPrint: false},
		{level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: false},
		{level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: false},
		{level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		{level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		{level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		{level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, timestamps=%t, otherCondition=%t - print=%t",
			tc.level, tc.timestamps, tc.otherCondition, tc.shouldPrint,
		)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			tLogger := NewLogger(tc.level, tc.timestamps)
			releaseStdout := test.CaptureLog(t, tLogger)

			// WHEN: Debug is called.
			tLogger.Debug(errors.New(name), LogFrom{}, tc.otherCondition)

			// THEN: msg was logged if shouldPrint, with/without timestamps.
			stdout := releaseStdout()
			regex := `^$`
			if tc.shouldPrint {
				regex = fmt.Sprintf("^DEBUG: %s\n$", name)
				if tc.timestamps {
					regex = strings.Replace(regex, `^`, timestampRegex, 1)
				}
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf(
					"%s\nLogger(level=%s, timestamps=%t, otherCondition=%t).Debug message stdout mismatch\ngot:  %q\nwant: %q",
					packageName, tc.level, tc.timestamps, tc.otherCondition,
					stdout, regex,
				)
			}
		})
	}
}

func TestLogger_Deprecated(t *testing.T) {
	// GIVEN: a Logger.
	tests := []struct {
		level      string
		timestamps bool
	}{
		{level: "ERROR", timestamps: true},
		{level: "ERROR", timestamps: false},
		{level: "WARN", timestamps: true},
		{level: "WARN", timestamps: false},
		{level: "INFO", timestamps: true},
		{level: "INFO", timestamps: false},
		{level: "VERBOSE", timestamps: true},
		{level: "VERBOSE", timestamps: false},
		{level: "DEBUG", timestamps: true},
		{level: "DEBUG", timestamps: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, timestamps=%t",
			tc.level, tc.timestamps,
		)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			tLogger := NewLogger(tc.level, tc.timestamps)
			releaseStdout := test.CaptureLog(t, tLogger)

			// WHEN: Deprecated is called.
			tLogger.Deprecated(errors.New(name))

			// THEN: message was logged.
			stdout := releaseStdout()
			regex := fmt.Sprintf("^DEPRECATED: %s\n$", name)
			if tc.timestamps {
				regex = strings.Replace(regex, `^`, timestampRegex, 1)
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf(
					"%s\nLogger(%s).Deprecated message mismatch\ngot:  %q\nwant: %q",
					packageName, name,
					stdout, regex,
				)
			}
		})
	}
}

func TestLogger__lengthCap(t *testing.T) {
	// GIVEN: a Logger.
	tests := []struct {
		level      string
		timestamps bool
	}{
		{level: "DEBUG", timestamps: true},
		{level: "DEBUG", timestamps: false},
		{level: "VERBOSE", timestamps: true},
		{level: "VERBOSE", timestamps: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf(
			"level=%s, timestamps=%t",
			tc.level, tc.timestamps,
		)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			tLogger := NewLogger(tc.level, tc.timestamps)
			releaseStdout := test.CaptureLog(t, tLogger)

			// AND: a message.
			msg := strings.Repeat("a", 9999)
			err := errors.New(msg)

			// WHEN: the message is sent.
			switch tc.level {
			case "DEBUG":
				tLogger.Debug(err, LogFrom{}, true)
			case "VERBOSE":
				tLogger.Verbose(err, LogFrom{}, true)
			case "INFO":
				tLogger.Info(err, LogFrom{}, true)
			case "WARN":
				tLogger.Warn(err, LogFrom{}, true)
			case "ERROR":
				tLogger.Error(err, LogFrom{}, true)
			}

			// THEN: the message was logged with/without timestamps.
			stdout := releaseStdout()
			regex := `^$`
			// Verify timestamps.
			gotTimestamps := util.RegexCheck(timestampRegex, stdout)
			if tc.timestamps != gotTimestamps {
				t.Errorf(
					"%s\nLogger(%s) message timestamp mismatch\ngot:  %q\nwant: %q",
					packageName, name,
					stdout, regex,
				)
			}

			// AND: the message length was limited.
			maxLength := 1001 // 1000 + 1 for the newline.
			if tc.timestamps {
				maxLength += 20
			}
			if gotLength := len(stdout); gotLength != maxLength {
				t.Errorf(
					"%s\nLogger(%s), message length not limited\ngot:    %d\n\nstdout: %q\nwant:   max=%d lines",
					packageName, name,
					gotLength, stdout,
					maxLength,
				)
			}
		})
	}
}

func TestLogger_LogMessage(t *testing.T) {
	// GIVEN: a Logger and a message.
	msg := "test message"
	tests := []struct {
		timestamps bool
	}{
		{timestamps: true},
		{timestamps: false},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("timestamps=%t", tc.timestamps)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			tLogger := NewLogger("INFO", tc.timestamps)
			releaseStdout := test.CaptureLog(t, tLogger)

			// WHEN: logMessage is called.
			tLogger.logMessage(msg)

			// THEN: msg was logged with/without timestamps.
			stdout := releaseStdout()
			var regex string
			if tc.timestamps {
				regex = fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} %s\n$", msg)
			} else {
				regex = fmt.Sprintf("^%s\n$", msg)
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf(
					"%s\nLogger(%s).logMessage message mismatch\ngot:  %q\nwant: %q",
					packageName, name,
					stdout, regex,
				)
			}
		})
	}
}
