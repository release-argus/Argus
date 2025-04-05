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

package logutil

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

var packageName = "logutil"

func TestInit(t *testing.T) {
	// GIVEN a log level and timestamps setting.
	tests := map[string]struct {
		level      string
		timestamps bool
	}{
		"INFO and timestamps": {
			level: "INFO", timestamps: true},
		"DEBUG, no timestamps": {
			level: "DEBUG", timestamps: false},
		"ERROR and timestamps": {
			level: "ERROR", timestamps: true},
		"WARN, no timestamps": {
			level: "WARN", timestamps: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel because of the once variable.
			// Reset the once variable and Log for each test.
			once = sync.Once{}
			Log = nil

			// WHEN Init is called.
			Init(tc.level, tc.timestamps)

			// THEN the Log should be initialized correctly.
			if Log == nil {
				t.Fatalf("%s\nLog was not initialized",
					packageName)
			}
			if Log.Level != levelMap[tc.level] {
				t.Errorf("%s\nwant: level=%d\ngot:  level=%d",
					packageName, levelMap[tc.level], Log.Level)
			}
			if Log.Timestamps != tc.timestamps {
				t.Errorf("%s\nwant: timestamps=%t\ngot:  timestamps=%t",
					packageName, tc.timestamps, Log.Timestamps)
			}
		})
	}
}

func TestNewJLog(t *testing.T) {
	// GIVEN a new JLog is wanted.
	tests := map[string]struct {
		level      string
		timestamps bool
	}{
		"timestamps JLog": {
			level: "INFO", timestamps: true},
		"no timestamps JLog": {
			level: "INFO", timestamps: false},
		"ERROR JLog": {
			level: "ERROR", timestamps: false},
		"WARN JLog": {
			level: "WARN", timestamps: false},
		"INFO JLog": {
			level: "INFO", timestamps: false},
		"VERBOSE JLog": {
			level: "VERBOSE", timestamps: false},
		"DEBUG JLog": {
			level: "DEBUG", timestamps: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN NewJLog is called.
			jLog := NewJLog(tc.level, tc.timestamps)

			// THEN the correct JLog is returned.
			if jLog.Level != levelMap[tc.level] {
				t.Errorf("%s\nwant: level=%d\ngot:  level=%d",
					packageName, levelMap[tc.level], jLog.Level)
			}
			if jLog.Timestamps != tc.timestamps {
				t.Errorf("%s\nwant: timestamps=%t\ngot:  timestamps=%t",
					packageName, tc.timestamps, jLog.Timestamps)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	// GIVEN a JLog and various new log levels.
	tests := map[string]struct {
		level      string
		panicRegex *string
	}{
		"ERROR":              {level: "ERROR"},
		"WARN":               {level: "WARN"},
		"INFO":               {level: "INFO"},
		"VERBOSE":            {level: "VERBOSE"},
		"DEBUG":              {level: "DEBUG"},
		"lower-case verbose": {level: "verbose"},
		"mixed-case vERbOse": {level: "vERbOse"},
		"invalid level PINEAPPLE": {level: "PINEAPPLE",
			panicRegex: test.StringPtr(`not a valid log\.level`)}}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			jLog := NewJLog("INFO", false)
			if tc.panicRegex != nil {
				jLog.Testing = true
				// Switch Fatal to panic and disable this panic.
				defer func() {
					r := recover()

					rStr := fmt.Sprint(r)
					if !util.RegexCheck(*tc.panicRegex, rStr) {
						t.Errorf("%s\npanic error mismatch\nwant: %q\ngot: %q",
							packageName, *tc.panicRegex, rStr)
					}
				}()
			}

			// WHEN SetLevel is called.
			jLog.SetLevel(tc.level)

			// THEN the correct JLog is returned.
			if jLog.Level != levelMap[strings.ToUpper(tc.level)] {
				t.Errorf("%s\nwant: level=%d\ngot:  level=%d",
					packageName, levelMap[strings.ToUpper(tc.level)], jLog.Level)
			}
		})
	}
}

func TestJLog_SetTimestamps(t *testing.T) {
	// GIVEN a JLog and various tests.
	tests := map[string]struct {
		start    bool
		changeTo bool
	}{
		"true to true":   {start: true, changeTo: true},
		"false to false": {start: false, changeTo: false},
		"true to false":  {start: true, changeTo: false},
		"false to true":  {start: false, changeTo: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			jLog := NewJLog("INFO", tc.start)

			// WHEN SetTimestamps is called.
			jLog.SetTimestamps(tc.changeTo)

			// THEN the timestamps are set correctly.
			if jLog.Timestamps != tc.changeTo {
				t.Errorf("%s\nwant: timestamps=%t\ngot:  timestamps=%t",
					packageName, tc.changeTo, jLog.Timestamps)
			}
		})
	}
}

func TestLogFrom_String(t *testing.T) {
	// GIVEN a LogFrom.
	tests := map[string]struct {
		logFrom LogFrom
		want    string
	}{
		"primary and secondary": {
			logFrom: LogFrom{Primary: "foo", Secondary: "bar"},
			want:    "foo (bar), "},
		"only primary": {
			logFrom: LogFrom{Primary: "foo"},
			want:    "foo, "},
		"only secondary": {
			logFrom: LogFrom{Secondary: "bar"},
			want:    "bar, "},
		"empty logFrom": {
			logFrom: LogFrom{},
			want:    ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN String() is called on it.
			got := tc.logFrom.String()

			// THEN an empty string is returned.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestJLog_IsLevel(t *testing.T) {
	// GIVEN you have a valid JLog.
	tests := map[string]struct {
		startLevel, testLevel string
		want                  bool
	}{
		"@ERROR, test ERROR": {
			startLevel: "ERROR", testLevel: "ERROR", want: true},
		"@ERROR, test WARN": {
			startLevel: "ERROR", testLevel: "WARN", want: false},
		"@ERROR, test INFO": {
			startLevel: "ERROR", testLevel: "INFO", want: false},
		"@ERROR, test VERBOSE": {
			startLevel: "ERROR", testLevel: "VERBOSE", want: false},
		"@ERROR, test DEBUG": {
			startLevel: "ERROR", testLevel: "DEBUG", want: false},
		"@WARN, test ERROR": {
			startLevel: "WARN", testLevel: "ERROR", want: false},
		"@WARN, test WARN": {
			startLevel: "WARN", testLevel: "WARN", want: true},
		"@WARN, test INFO": {
			startLevel: "WARN", testLevel: "INFO", want: false},
		"@WARN, test VERBOSE": {
			startLevel: "WARN", testLevel: "VERBOSE", want: false},
		"@WARN, test DEBUG": {
			startLevel: "WARN", testLevel: "DEBUG", want: false},
		"@INFO, test ERROR": {
			startLevel: "INFO", testLevel: "ERROR", want: false},
		"@INFO, test WARN": {
			startLevel: "INFO", testLevel: "WARN", want: false},
		"@INFO, test INFO": {
			startLevel: "INFO", testLevel: "INFO", want: true},
		"@INFO, test VERBOSE": {
			startLevel: "INFO", testLevel: "VERBOSE", want: false},
		"@INFO, test DEBUG": {
			startLevel: "INFO", testLevel: "DEBUG", want: false},
		"@VERBOSE, test ERROR": {
			startLevel: "VERBOSE", testLevel: "ERROR", want: false},
		"@VERBOSE, test WARN": {
			startLevel: "VERBOSE", testLevel: "WARN", want: false},
		"@VERBOSE, test INFO": {
			startLevel: "VERBOSE", testLevel: "INFO", want: false},
		"@VERBOSE, test VERBOSE": {
			startLevel: "VERBOSE", testLevel: "VERBOSE", want: true},
		"@VERBOSE, test DEBUG": {
			startLevel: "VERBOSE", testLevel: "DEBUG", want: false},
		"@DEBUG, test ERROR": {
			startLevel: "DEBUG", testLevel: "ERROR", want: false},
		"@DEBUG, test WARN": {
			startLevel: "DEBUG", testLevel: "WARN", want: false},
		"@DEBUG, test INFO": {
			startLevel: "DEBUG", testLevel: "INFO", want: false},
		"@DEBUG, test VERBOSE": {
			startLevel: "DEBUG", testLevel: "VERBOSE", want: false},
		"@DEBUG, test DEBUG": {
			startLevel: "DEBUG", testLevel: "DEBUG", want: true},
		"@DEBUG, test level not in level map": {
			startLevel: "DEBUG", testLevel: "FOO", want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			jLog := NewJLog(tc.startLevel, false)

			// WHEN IsLevel is called to check the given level.
			got := jLog.IsLevel(tc.testLevel)

			// THEN the correct response is returned.
			if got != tc.want {
				t.Errorf("%s\nlevel is %s, check of whether it's %s failed\nwant: %t\ngot:  %t",
					packageName,
					tc.startLevel, tc.testLevel,
					tc.want, got)
			}
		})
	}
}

func TestJLog_Error(t *testing.T) {
	// GIVEN a JLog and message.
	msg := "argus"
	tests := map[string]struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool
	}{
		"ERROR log with timestamps": {
			level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: true},
		"ERROR log no timestamps": {
			level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: true},
		"ERROR log with !otherCondition": {
			level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		"WARN log with timestamps": {
			level: "WARN", timestamps: true, otherCondition: true, shouldPrint: true},
		"WARN log no timestamps": {
			level: "WARN", timestamps: false, otherCondition: true, shouldPrint: true},
		"WARN log with !otherCondition": {
			level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		"INFO log with timestamps": {
			level: "INFO", timestamps: true, otherCondition: true, shouldPrint: true},
		"INFO log no timestamps": {
			level: "INFO", timestamps: false, otherCondition: true, shouldPrint: true},
		"INFO log with !otherCondition": {
			level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		"VERBOSE log with timestamps": {
			level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: true},
		"VERBOSE log no timestamps": {
			level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: true},
		"VERBOSE log with !otherCondition": {

			level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		"DEBUG log with timestamps": {
			level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		"DEBUG log no timestamps": {
			level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		"DEBUG log with !otherCondition": {
			level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			jLog := NewJLog(tc.level, tc.timestamps)
			var logOut bytes.Buffer
			log.SetOutput(&logOut)

			// WHEN Error is called with true.
			jLog.Error(errors.New(msg), LogFrom{}, tc.otherCondition)

			// THEN msg was logged if shouldPrint, with/without timestamps.
			stdout := releaseStdout()
			var regex string
			if tc.timestamps {
				stdout = logOut.String()
				regex = fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} ERROR: %s\n$", msg)
			} else if !tc.otherCondition {
				regex = "^$"
			} else {
				regex = fmt.Sprintf("^ERROR: %s\n$", msg)
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf("%s\nerror mismatch on 'ERROR: '\nwant: %q\nGot %q",
					packageName, regex, stdout)
			}
		})
	}
}

func TestJLog_Warn(t *testing.T) {
	// GIVEN a JLog and message.
	msg := "argus"
	tests := map[string]struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool
	}{
		"ERROR log with timestamps": {
			level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: false},
		"ERROR log no timestamps": {
			level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: false},
		"ERROR log with !otherCondition": {
			level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		"WARN log with timestamps": {
			level: "WARN", timestamps: true, otherCondition: true, shouldPrint: true},
		"WARN log no timestamps": {
			level: "WARN", timestamps: false, otherCondition: true, shouldPrint: true},
		"WARN log with !otherCondition": {
			level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		"INFO log with timestamps": {
			level: "INFO", timestamps: true, otherCondition: true, shouldPrint: true},
		"INFO log no timestamps": {
			level: "INFO", timestamps: false, otherCondition: true, shouldPrint: true},
		"INFO log with !otherCondition": {
			level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		"VERBOSE log with timestamps": {
			level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: true},
		"VERBOSE log no timestamps": {
			level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: true},
		"VERBOSE log with !otherCondition": {
			level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		"DEBUG log with timestamps": {
			level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		"DEBUG log no timestamps": {
			level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		"DEBUG log with !otherCondition": {
			level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			jLog := NewJLog(tc.level, tc.timestamps)
			var logOut bytes.Buffer
			log.SetOutput(&logOut)

			// WHEN Warn is called with true.
			jLog.Warn(errors.New(msg), LogFrom{}, tc.otherCondition)

			// THEN msg was logged if shouldPrint, with/without timestamps.
			stdout := releaseStdout()
			var regex string
			if !tc.shouldPrint {
				regex = "^$"
			} else if tc.timestamps {
				stdout = logOut.String()
				regex = fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} WARNING: %s\n$", msg)
			} else {
				regex = fmt.Sprintf("^WARNING: %s\n$", msg)
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf("%s\nerror mismatch on 'WARNING: '\nwant: %q\ngot:  %q",
					packageName, regex, stdout)
			}
		})
	}
}

func TestJLog_Info(t *testing.T) {
	// GIVEN a JLog and message.
	msg := "argus"
	tests := map[string]struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool
	}{
		"ERROR log with timestamps": {
			level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: false},
		"ERROR log no timestamps": {
			level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: false},
		"ERROR log with !otherCondition": {
			level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		"WARN log with timestamps": {
			level: "WARN", timestamps: true, otherCondition: true, shouldPrint: false},
		"WARN log no timestamps": {
			level: "WARN", timestamps: false, otherCondition: true, shouldPrint: false},
		"WARN log with !otherCondition": {
			level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		"INFO log with timestamps": {
			level: "INFO", timestamps: true, otherCondition: true, shouldPrint: true},
		"INFO log no timestamps": {
			level: "INFO", timestamps: false, otherCondition: true, shouldPrint: true},
		"INFO log with !otherCondition": {
			level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		"VERBOSE log with timestamps": {
			level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: true},
		"VERBOSE log no timestamps": {
			level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: true},
		"VERBOSE log with !otherCondition": {
			level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		"DEBUG log with timestamps": {
			level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		"DEBUG log no timestamps": {
			level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		"DEBUG log with !otherCondition": {
			level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			jLog := NewJLog(tc.level, tc.timestamps)
			var logOut bytes.Buffer
			log.SetOutput(&logOut)

			// WHEN Info is called with true.
			jLog.Info(errors.New(msg), LogFrom{}, tc.otherCondition)

			// THEN msg was logged if shouldPrint, with/without timestamps.
			stdout := releaseStdout()
			var regex string
			if !tc.shouldPrint {
				regex = "^$"
			} else if tc.timestamps {
				stdout = logOut.String()
				regex = fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} INFO: %s\n$", msg)
			} else {
				regex = fmt.Sprintf("^INFO: %s\n$", msg)
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf("%s\nError mismatch on 'INFO: '\nwant: %q\ngot:  %q",
					packageName, regex, stdout)
			}
		})
	}
}

func TestJLog_Verbose(t *testing.T) {
	// GIVEN a JLog and message.
	tests := map[string]struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool

		customMsg *string
		trimmed   bool
	}{
		"ERROR log with timestamps": {
			level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: false},
		"ERROR log no timestamps": {
			level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: false},
		"ERROR log with !otherCondition": {
			level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		"WARN log with timestamps": {
			level: "WARN", timestamps: true, otherCondition: true, shouldPrint: false},
		"WARN log no timestamps": {
			level: "WARN", timestamps: false, otherCondition: true, shouldPrint: false},
		"WARN log with !otherCondition": {
			level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		"INFO log with timestamps": {
			level: "INFO", timestamps: true, otherCondition: true, shouldPrint: false},
		"INFO log no timestamps": {
			level: "INFO", timestamps: false, otherCondition: true, shouldPrint: false},
		"INFO log with !otherCondition": {
			level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		"VERBOSE log with timestamps": {
			level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: true},
		"VERBOSE log no timestamps": {
			level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: true},
		"VERBOSE log with !otherCondition": {
			level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		"DEBUG log with timestamps": {
			level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		"DEBUG log no timestamps": {
			level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		"DEBUG log with !otherCondition": {
			level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
		"limits VERBOSE message length": {
			level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: true,
			customMsg: test.StringPtr(strings.Repeat("a", 9999)), trimmed: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			msg := "argus"

			jLog := NewJLog(tc.level, tc.timestamps)
			var logOut bytes.Buffer
			log.SetOutput(&logOut)
			if tc.customMsg != nil {
				msg = *tc.customMsg
			}

			// WHEN Verbose is called with true.
			jLog.Verbose(errors.New(msg), LogFrom{}, tc.otherCondition)

			// THEN msg was logged if shouldPrint, with/without timestamps.
			stdout := releaseStdout()
			var regex string
			if tc.customMsg != nil && tc.trimmed {
				msg = (*tc.customMsg)[:1000-len("VERBOSE: ...")] + "..."
			}
			if !tc.shouldPrint {
				regex = "^$"
			} else if tc.timestamps {
				stdout = logOut.String()
				regex = fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} VERBOSE: %s\n$", msg)
			} else {
				regex = fmt.Sprintf("^VERBOSE: %s\n$", msg)
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf("%s\nVERBOSE print mismatch\nwant: %q\ngot:  %q",
					packageName, regex, stdout)
			}
			if tc.customMsg != nil && tc.trimmed {
				maxLength := 1001 // 1000 + 1 for the newline.
				if len(stdout) != maxLength {
					t.Errorf("%s\nVERBOSE message length not limited\nwant: %d lines\ngot:  %d\n\nstdout: %q",
						packageName, maxLength, len(stdout), stdout)
				}
			}
		})
	}
}

func TestJLog_Debug(t *testing.T) {
	// GIVEN a JLog and message.
	msg := "argus"
	tests := map[string]struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool

		customMsg *string
		trimmed   bool
	}{
		"ERROR log with timestamps": {
			level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: false},
		"ERROR log no timestamps": {
			level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: false},
		"ERROR log with !otherCondition": {
			level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		"WARN log with timestamps": {
			level: "WARN", timestamps: true, otherCondition: true, shouldPrint: false},
		"WARN log no timestamps": {
			level: "WARN", timestamps: false, otherCondition: true, shouldPrint: false},
		"WARN log with !otherCondition": {
			level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		"INFO log with timestamps": {
			level: "INFO", timestamps: true, otherCondition: true, shouldPrint: false},
		"INFO log no timestamps": {
			level: "INFO", timestamps: false, otherCondition: true, shouldPrint: false},
		"INFO log with !otherCondition": {
			level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		"VERBOSE log with timestamps": {
			level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: false},
		"VERBOSE log no timestamps": {
			level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: false},
		"VERBOSE log with !otherCondition": {
			level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		"DEBUG log with timestamps": {
			level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		"DEBUG log no timestamps": {
			level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		"DEBUG log with !otherCondition": {
			level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
		"limits DEBUG message length": {
			level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true,
			customMsg: test.StringPtr(strings.Repeat("a", 9999)), trimmed: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			jLog := NewJLog(tc.level, tc.timestamps)
			var logOut bytes.Buffer
			log.SetOutput(&logOut)
			if tc.customMsg != nil {
				msg = *tc.customMsg
			}

			// WHEN Debug is called with true.
			jLog.Debug(errors.New(msg), LogFrom{}, tc.otherCondition)

			// THEN msg was logged if shouldPrint, with/without timestamps.
			stdout := releaseStdout()
			var regex string
			if tc.customMsg != nil && tc.trimmed {
				msg = (*tc.customMsg)[:1000-len("DEBUG: ...")] + "..."
			}
			if !tc.shouldPrint {
				regex = `^$`
			} else if tc.timestamps {
				stdout = logOut.String()
				regex = fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} DEBUG: %s\n$",
					strings.ReplaceAll(msg, ".", `\.`))
			} else {
				regex = fmt.Sprintf("^DEBUG: %s\n$", msg)
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf("%s\nDEBUG print mismatch\nwant: %q\ngot:  %q",
					packageName, regex, stdout)
			}
			if tc.customMsg != nil && tc.trimmed {
				maxLength := 1001 // 1000 + 1 for the newline.
				if len(stdout) != maxLength {
					t.Errorf("%s\nDEBUG message length not limited\nwant: %d lines\ngot:  %d\n\nstdout: %q",
						packageName, maxLength, len(stdout), stdout)
				}
			}
		})
	}
}

func TestJLog_Fatal(t *testing.T) {
	// GIVEN a JLog and message.
	msg := "argus"
	tests := map[string]struct {
		level          string
		timestamps     bool
		otherCondition bool
		shouldPrint    bool
	}{
		"ERROR log with timestamps": {
			level: "ERROR", timestamps: true, otherCondition: true, shouldPrint: true},
		"ERROR log no timestamps": {
			level: "ERROR", timestamps: false, otherCondition: true, shouldPrint: true},
		"ERROR log with !otherCondition": {
			level: "ERROR", timestamps: false, otherCondition: false, shouldPrint: false},
		"WARN log with timestamps": {
			level: "WARN", timestamps: true, otherCondition: true, shouldPrint: true},
		"WARN log no timestamps": {
			level: "WARN", timestamps: false, otherCondition: true, shouldPrint: true},
		"WARN log with !otherCondition": {
			level: "WARN", timestamps: false, otherCondition: false, shouldPrint: false},
		"INFO log with timestamps": {
			level: "INFO", timestamps: true, otherCondition: true, shouldPrint: true},
		"INFO log no timestamps": {
			level: "INFO", timestamps: false, otherCondition: true, shouldPrint: true},
		"INFO log with !otherCondition": {
			level: "INFO", timestamps: false, otherCondition: false, shouldPrint: false},
		"VERBOSE log with timestamps": {
			level: "VERBOSE", timestamps: true, otherCondition: true, shouldPrint: true},
		"VERBOSE log no timestamps": {
			level: "VERBOSE", timestamps: false, otherCondition: true, shouldPrint: true},
		"VERBOSE log with !otherCondition": {
			level: "VERBOSE", timestamps: false, otherCondition: false, shouldPrint: false},
		"DEBUG log with timestamps": {
			level: "DEBUG", timestamps: true, otherCondition: true, shouldPrint: true},
		"DEBUG log no timestamps": {
			level: "DEBUG", timestamps: false, otherCondition: true, shouldPrint: true},
		"DEBUG log with !otherCondition": {
			level: "DEBUG", timestamps: false, otherCondition: false, shouldPrint: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			jLog := NewJLog(tc.level, tc.timestamps)
			var logOut bytes.Buffer
			log.SetOutput(&logOut)
			if tc.shouldPrint {
				jLog.Testing = true
				defer func() {
					recover()
					stdout := releaseStdout()

					regex := fmt.Sprintf("^ERROR: %s\n$", msg)
					if tc.timestamps {
						stdout = logOut.String()
						regex = fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} ERROR: %s\n$", msg)
					}
					if !util.RegexCheck(regex, stdout) {
						t.Errorf("%s\ndid panic, but error mismatch on 'Error: ' (Fatal)\nwant: %q\ngot:  %q",
							packageName, regex, stdout)
					}
				}()
			}

			// WHEN Fatal is called with true.
			jLog.Fatal(errors.New(msg), LogFrom{}, tc.otherCondition)

			// THEN no message was logged in otherCondition is false.
			stdout := releaseStdout()
			regex := "^$"
			if !util.RegexCheck(regex, stdout) {
				t.Errorf("%s\nerror mismatch (otherCondition=%t)\nwant: %q\ngot:  %q",
					packageName, tc.otherCondition, regex, stdout)
			}
			if tc.otherCondition {
				t.Fatalf("%s\nFatal didn't panic, but should have",
					packageName)
			}
		})
	}
}

func TestJLog_logMessage(t *testing.T) {
	// GIVEN a JLog and a message.
	msg := "test message"
	tests := map[string]struct {
		timestamps bool
	}{
		"with timestamps": {
			timestamps: true},
		"without timestamps": {
			timestamps: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			jLog := NewJLog("INFO", tc.timestamps)
			var logOut bytes.Buffer
			log.SetOutput(&logOut)

			// WHEN logMessage is called.
			jLog.logMessage(msg)

			// THEN msg was logged with/without timestamps.
			stdout := releaseStdout()
			var regex string
			if tc.timestamps {
				stdout = logOut.String()
				regex = fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} %s\n$",
					msg)
			} else {
				regex = fmt.Sprintf("^%s\n$",
					msg)
			}
			if !util.RegexCheck(regex, stdout) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, regex, stdout)
			}
		})
	}
}
