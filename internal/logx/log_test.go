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
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestInit(t *testing.T) {
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

	hadLog := logger
	t.Cleanup(func() {
		logger = hadLog
	})

	for _, tc := range tests {
		t.Run(tc.level, func(t *testing.T) {
			// Serial: resets once and logger.

			once = sync.Once{}
			logger = nil

			exitCodeChannel := Init(tc.level, tc.timestamps)

			if logger == nil {
				t.Fatal("logger mismatch\ngot:  nil\nwant: non-nil")
			}
			fieldTests := []test.FieldAssertion{
				{Name: "Level", Got: logger.Level.Load(), Want: levelMap[tc.level], Mode: test.CompareEqual},
				{Name: "timestamps", Got: logger.timestamps, Want: tc.timestamps, Mode: test.CompareEqual},
				{Name: "exitCodeChannel", Got: logger.exitCodeChannel, Want: exitCodeChannel, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Logger"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			t.Parallel()

			for _, timestamps := range []bool{true, false} {
				t.Run(fmt.Sprintf("timestamps=%t", timestamps), func(t *testing.T) {
					t.Parallel()

					tLogger := NewLogger(level, timestamps)

					fieldTests := []test.FieldAssertion{
						{Name: "Level", Got: tLogger.Level.Load(), Want: levelMap[level], Mode: test.CompareEqual},
						{Name: "timestamps", Got: tLogger.timestamps, Want: timestamps, Mode: test.CompareEqual},
					}
					if testErr := test.AssertFields(t, fieldTests, "Logger"); testErr != nil {
						t.Fatal(testErr)
					}
				})
			}
		})
	}
}

func TestLogger_SetExitCodeChannel(t *testing.T) {
	tLogger := NewLogger("INFO", true)
	tLogger.exitCodeChannel = nil

	channels := []chan string{
		make(chan string),
		make(chan string),
	}

	for i, channel := range channels {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// Serial: shares tLogger.

			tLogger.SetExitCodeChannel(channel)

			if got := tLogger.exitCodeChannel; got != channel {
				t.Fatalf("Logger.exitCodeChannel mismatch\ngot:  %p\nwant: %p", got, channel)
			}

			if got := tLogger.ExitCodeChannel(); got != channel {
				t.Fatalf("Logger.ExitCodeChannel() mismatch\ngot:  %p\nwant: %p", got, channel)
			}
		})
	}
}

func TestLogger_SetLevel(t *testing.T) {
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
		t.Run(tc.level, func(t *testing.T) {
			// Serial: captures stdout.

			tLogger := NewLogger("INFO", false)
			releaseStdout := test.CaptureLog(t, tLogger)

			tLogger.SetLevel(tc.level)

			// An unknown level falls back to ERROR (0).
			level := tLogger.Level.Load()
			if tc.ok {
				if want := levelMap[strings.ToUpper(tc.level)]; level != want {
					t.Errorf("Logger.Level mismatch\ngot:  %d\nwant: %d", level, want)
				}
			} else if level != 0 {
				t.Errorf("Logger.Level mismatch\ngot:  %d\nwant: 0", level)
			}

			stdout := releaseStdout()
			errRegex := `not a valid log\.level`
			stdoutOk := !util.RegexCheck(errRegex, stdout)
			if stdoutOk != tc.ok {
				if tc.ok {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: NOT %q", stdout, errRegex)
				} else {
					t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, errRegex)
				}
			}
		})
	}
}

func TestLogger_IsLevel(t *testing.T) {
	for _, startLevel := range levels {
		t.Run(startLevel, func(t *testing.T) {
			t.Parallel()

			for _, testLevel := range slices.Concat(levels, []string{"FOO"}) {
				t.Run(testLevel, func(t *testing.T) {
					t.Parallel()

					tLogger := NewLogger(startLevel, false)

					got := tLogger.IsLevel(testLevel)

					// Only the logger's own level matches, not the levels it also logs.
					if want := testLevel == startLevel; got != want {
						t.Errorf("mismatch\ngot:  %t\nwant: %t", got, want)
					}
				})
			}
		})
	}
}

func TestLogger_SetTimestamps(t *testing.T) {
	const msg = "timestamp toggle test"
	tests := []struct {
		name     string
		start    bool
		changeTo bool
	}{
		{name: "keep enabled", start: true, changeTo: true},
		{name: "disable", start: true, changeTo: false},
		{name: "keep disabled", start: false, changeTo: false},
		{name: "enable", start: false, changeTo: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			tLogger := NewLogger("INFO", tc.start)
			tLogger.SetOutput(buf)

			tLogger.SetTimestamps(tc.changeTo)

			if tLogger.timestamps != tc.changeTo {
				t.Errorf(
					"Logger.timestamps mismatch\ngot:  %t\nwant: %t",
					tLogger.timestamps, tc.changeTo,
				)
			}

			tLogger.logMessage(msg)
			got := buf.String()
			wantRegex := fmt.Sprintf("^%s\n$", msg)
			if tc.changeTo {
				wantRegex = strings.Replace(wantRegex, `^`, timestampRegex, 1)
			}
			if !util.RegexCheck(wantRegex, got) {
				t.Errorf("Buffer.String() mismatch\ngot:  %q\nwant: %q", got, wantRegex)
			}
		})
	}
}

func TestLogFrom_String(t *testing.T) {
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

			got := tc.logFrom.String()

			if got != tc.want {
				t.Errorf("LogFrom.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestLogger_Output(t *testing.T) {
	tests := []struct {
		name        string
		startOutput io.Writer
		newOutput   *bytes.Buffer
	}{
		{
			name:        "from a buffer",
			startOutput: new(bytes.Buffer),
			newOutput:   new(bytes.Buffer),
		},
		{
			name:        "from stdout",
			startOutput: os.Stdout,
			newOutput:   new(bytes.Buffer),
		},
		{
			name:        "from stderr",
			startOutput: os.Stderr,
			newOutput:   new(bytes.Buffer),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			tLogger := &Logger{
				writer: log.New(buf, "", 0),
				out:    tc.startOutput,
			}

			if got := tLogger.GetOutput(); got != tc.startOutput {
				t.Fatalf(
					"before: Logger.GetOutput() mismatch\ngot:  %p\nwant: %p",
					got, tc.startOutput,
				)
			}

			tLogger.SetOutput(tc.newOutput)

			got := tLogger.GetOutput()
			if got != tc.newOutput {
				t.Errorf(
					"after SetOutput: Logger.GetOutput() mismatch\ngot:  %p\nwant: %p",
					got, tc.newOutput,
				)
			}

			testMsg := "test message"
			tLogger.logMessage(testMsg)
			if buf.Len() != 0 {
				t.Errorf(
					"old: Buffer.String() mismatch\ngot:  %q\nwant: \"\" (replaced)",
					buf.String(),
				)
			}
			if !strings.Contains(tc.newOutput.String(), testMsg) {
				t.Errorf(
					"new: Buffer.String() mismatch\ngot:  %q\nwant: %q",
					tc.newOutput.String(), testMsg,
				)
			}
		})
	}
}

func TestLogger_Fatal(t *testing.T) {
	const msg = "argus"

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			// Serial: its subtests capture stdout.

			for _, timestamps := range []bool{true, false} {
				t.Run(fmt.Sprintf("timestamps=%t", timestamps), func(t *testing.T) {
					// Serial: captures stdout.

					tLogger := NewLogger(level, timestamps)
					releaseStdout := test.CaptureLog(t, tLogger)

					tLogger.Fatal(errors.New(msg), LogFrom{})

					stdout := releaseStdout()
					regex := fmt.Sprintf("^FATAL: %s\n$", msg)
					if timestamps {
						regex = strings.Replace(regex, `^`, timestampRegex, 1)
					}
					if !util.RegexCheck(regex, stdout) {
						t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, regex)
					}

					wantExitMsg := "FATAL: " + msg
					select {
					case got := <-tLogger.exitCodeChannel:
						if got != wantExitMsg {
							t.Errorf(
								"Logger.exitCodeChannel mismatch\ngot:  %q\nwant: %q",
								got, wantExitMsg,
							)
						}
					default:
						t.Error("Logger.exitCodeChannel message count mismatch\ngot:  0\nwant: 1")
					}
				})
			}
		})
	}
}

func TestLogger_Error(t *testing.T) {
	const msg = "argus"
	tests := []struct {
		level       string
		shouldPrint bool
	}{
		{level: "ERROR", shouldPrint: true},
		{level: "WARN", shouldPrint: true},
		{level: "INFO", shouldPrint: true},
		{level: "VERBOSE", shouldPrint: true},
		{level: "DEBUG", shouldPrint: true},
	}

	for _, tc := range tests {
		t.Run(tc.level, func(t *testing.T) {
			// Serial: its subtests capture stdout.

			for _, timestamps := range []bool{true, false} {
				t.Run(fmt.Sprintf("timestamps=%t", timestamps), func(t *testing.T) {
					// Serial: its subtests capture stdout.

					for _, otherCondition := range []bool{true, false} {
						t.Run(fmt.Sprintf("otherCondition=%t", otherCondition), func(t *testing.T) {
							// Serial: captures stdout.

							tLogger := NewLogger(tc.level, timestamps)
							releaseStdout := test.CaptureLog(t, tLogger)

							tLogger.Error(errors.New(msg), LogFrom{}, otherCondition)

							stdout := releaseStdout()
							regex := `^$`
							if tc.shouldPrint && otherCondition {
								regex = fmt.Sprintf("^ERROR: %s\n$", msg)
								if timestamps {
									regex = strings.Replace(regex, `^`, timestampRegex, 1)
								}
							}
							if !util.RegexCheck(regex, stdout) {
								t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, regex)
							}
						})
					}
				})
			}
		})
	}
}

func TestLogger_Warn(t *testing.T) {
	const msg = "argus"
	tests := []struct {
		level       string
		shouldPrint bool
	}{
		{level: "ERROR", shouldPrint: false},
		{level: "WARN", shouldPrint: true},
		{level: "INFO", shouldPrint: true},
		{level: "VERBOSE", shouldPrint: true},
		{level: "DEBUG", shouldPrint: true},
	}

	for _, tc := range tests {
		t.Run(tc.level, func(t *testing.T) {
			// Serial: its subtests capture stdout.

			for _, timestamps := range []bool{true, false} {
				t.Run(fmt.Sprintf("timestamps=%t", timestamps), func(t *testing.T) {
					// Serial: its subtests capture stdout.

					for _, otherCondition := range []bool{true, false} {
						t.Run(fmt.Sprintf("otherCondition=%t", otherCondition), func(t *testing.T) {
							// Serial: captures stdout.

							tLogger := NewLogger(tc.level, timestamps)
							releaseStdout := test.CaptureLog(t, tLogger)

							tLogger.Warn(errors.New(msg), LogFrom{}, otherCondition)

							stdout := releaseStdout()
							regex := `^$`
							if tc.shouldPrint && otherCondition {
								regex = fmt.Sprintf("^WARNING: %s\n$", msg)
								if timestamps {
									regex = strings.Replace(regex, `^`, timestampRegex, 1)
								}
							}
							if !util.RegexCheck(regex, stdout) {
								t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, regex)
							}
						})
					}
				})
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	const msg = "argus"
	tests := []struct {
		level       string
		shouldPrint bool
	}{
		{level: "ERROR", shouldPrint: false},
		{level: "WARN", shouldPrint: false},
		{level: "INFO", shouldPrint: true},
		{level: "VERBOSE", shouldPrint: true},
		{level: "DEBUG", shouldPrint: true},
	}

	for _, tc := range tests {
		t.Run(tc.level, func(t *testing.T) {
			// Serial: its subtests capture stdout.

			for _, timestamps := range []bool{true, false} {
				t.Run(fmt.Sprintf("timestamps=%t", timestamps), func(t *testing.T) {
					// Serial: its subtests capture stdout.

					for _, otherCondition := range []bool{true, false} {
						t.Run(fmt.Sprintf("otherCondition=%t", otherCondition), func(t *testing.T) {
							// Serial: captures stdout.

							tLogger := NewLogger(tc.level, timestamps)
							releaseStdout := test.CaptureLog(t, tLogger)

							tLogger.Info(errors.New(msg), LogFrom{}, otherCondition)

							stdout := releaseStdout()
							regex := `^$`
							if tc.shouldPrint && otherCondition {
								regex = fmt.Sprintf("^INFO: %s\n$", msg)
								if timestamps {
									regex = strings.Replace(regex, `^`, timestampRegex, 1)
								}
							}
							if !util.RegexCheck(regex, stdout) {
								t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, regex)
							}
						})
					}
				})
			}
		})
	}
}

func TestLogger_Verbose(t *testing.T) {
	const msg = "argus"
	tests := []struct {
		level       string
		shouldPrint bool
	}{
		{level: "ERROR", shouldPrint: false},
		{level: "WARN", shouldPrint: false},
		{level: "INFO", shouldPrint: false},
		{level: "VERBOSE", shouldPrint: true},
		{level: "DEBUG", shouldPrint: true},
	}

	for _, tc := range tests {
		t.Run(tc.level, func(t *testing.T) {
			// Serial: its subtests capture stdout.

			for _, timestamps := range []bool{true, false} {
				t.Run(fmt.Sprintf("timestamps=%t", timestamps), func(t *testing.T) {
					// Serial: its subtests capture stdout.

					for _, otherCondition := range []bool{true, false} {
						t.Run(fmt.Sprintf("otherCondition=%t", otherCondition), func(t *testing.T) {
							// Serial: captures stdout.

							tLogger := NewLogger(tc.level, timestamps)
							releaseStdout := test.CaptureLog(t, tLogger)

							tLogger.Verbose(errors.New(msg), LogFrom{}, otherCondition)

							stdout := releaseStdout()
							regex := `^$`
							if tc.shouldPrint && otherCondition {
								regex = fmt.Sprintf("^VERBOSE: %s\n$", msg)
								if timestamps {
									regex = strings.Replace(regex, `^`, timestampRegex, 1)
								}
							}
							if !util.RegexCheck(regex, stdout) {
								t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, regex)
							}
						})
					}
				})
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	const msg = "argus"
	tests := []struct {
		level       string
		shouldPrint bool
	}{
		{level: "ERROR", shouldPrint: false},
		{level: "WARN", shouldPrint: false},
		{level: "INFO", shouldPrint: false},
		{level: "VERBOSE", shouldPrint: false},
		{level: "DEBUG", shouldPrint: true},
	}

	for _, tc := range tests {
		t.Run(tc.level, func(t *testing.T) {
			// Serial: its subtests capture stdout.

			for _, timestamps := range []bool{true, false} {
				t.Run(fmt.Sprintf("timestamps=%t", timestamps), func(t *testing.T) {
					// Serial: its subtests capture stdout.

					for _, otherCondition := range []bool{true, false} {
						t.Run(fmt.Sprintf("otherCondition=%t", otherCondition), func(t *testing.T) {
							// Serial: captures stdout.

							tLogger := NewLogger(tc.level, timestamps)
							releaseStdout := test.CaptureLog(t, tLogger)

							tLogger.Debug(errors.New(msg), LogFrom{}, otherCondition)

							stdout := releaseStdout()
							regex := `^$`
							if tc.shouldPrint && otherCondition {
								regex = fmt.Sprintf("^DEBUG: %s\n$", msg)
								if timestamps {
									regex = strings.Replace(regex, `^`, timestampRegex, 1)
								}
							}
							if !util.RegexCheck(regex, stdout) {
								t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, regex)
							}
						})
					}
				})
			}
		})
	}
}

func TestLogger_Deprecated(t *testing.T) {
	const msg = "argus"

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			// Serial: its subtests capture stdout.

			for _, timestamps := range []bool{true, false} {
				t.Run(fmt.Sprintf("timestamps=%t", timestamps), func(t *testing.T) {
					// Serial: captures stdout.

					tLogger := NewLogger(level, timestamps)
					releaseStdout := test.CaptureLog(t, tLogger)

					tLogger.Deprecated(errors.New(msg))

					stdout := releaseStdout()
					regex := fmt.Sprintf("^DEPRECATED: %s\n$", msg)
					if timestamps {
						regex = strings.Replace(regex, `^`, timestampRegex, 1)
					}
					if !util.RegexCheck(regex, stdout) {
						t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, regex)
					}
				})
			}
		})
	}
}

func TestLogger__lengthCap(t *testing.T) {
	for _, level := range []string{"DEBUG", "VERBOSE"} {
		t.Run(level, func(t *testing.T) {
			// Serial: its subtests capture stdout.

			for _, timestamps := range []bool{true, false} {
				t.Run(fmt.Sprintf("timestamps=%t", timestamps), func(t *testing.T) {
					// Serial: captures stdout.

					tLogger := NewLogger(level, timestamps)
					releaseStdout := test.CaptureLog(t, tLogger)

					err := errors.New(strings.Repeat("a", 9999))

					switch level {
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

					stdout := releaseStdout()
					gotTimestamps := util.RegexCheck(timestampRegex, stdout)
					if gotTimestamps != timestamps {
						t.Errorf(
							"timestamp mismatch\ngot:  %s\nwant: %s",
							test.Presence(gotTimestamps), test.Presence(timestamps),
						)
					}

					maxLength := 1001 // 1000 + 1 for the newline.
					if timestamps {
						maxLength += 20
					}
					if gotLength := len(stdout); gotLength != maxLength {
						t.Errorf("len(stdout) mismatch\ngot:  %d\nwant: %d", gotLength, maxLength)
					}
				})
			}
		})
	}
}

func TestLogger_LogMessage(t *testing.T) {
	msg := "test message"
	tests := []struct {
		timestamps bool
	}{
		{timestamps: true},
		{timestamps: false},
	}

	for _, tc := range tests {
		t.Run(strconv.FormatBool(tc.timestamps), func(t *testing.T) {
			// Serial: captures stdout.

			tLogger := NewLogger("INFO", tc.timestamps)
			releaseStdout := test.CaptureLog(t, tLogger)

			tLogger.logMessage(msg)

			stdout := releaseStdout()
			regex := fmt.Sprintf("^%s\n$", msg)
			if tc.timestamps {
				regex = strings.Replace(regex, `^`, timestampRegex, 1)
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, regex)
			}
		})
	}
}
