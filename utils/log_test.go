// Copyright [2022] [Argus]
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

package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"testing"
)

func TestNewJLogGivesJLog(t *testing.T) {
	// GIVEN a new JLog is wanted

	// WHEN NewJLog is called
	var jLog interface{} = NewJLog("DEBUG", false)

	// THEN a JLog pointer is created
	switch v := jLog.(type) {
	default:
		t.Errorf("unexpected type %T, jLog should be of type JLog!",
			v)
	case *JLog:
	}
}

func TestNewJLogWithLogLevelParam(t *testing.T) {
	// GIVEN a new JLog is wanted with certain params
	logLevel := "DEBUG"

	// WHEN NewJLog is called
	jLog := NewJLog(logLevel, false)
	wantedLogLevelUInt := uint(4)

	// THEN a JLog pointer is created with this Log Level
	got := (*jLog).Level
	if got != wantedLogLevelUInt {
		t.Errorf("NewJLog didn't use level param. Wanted %q (%d) but got %d",
			logLevel, wantedLogLevelUInt, got)
	}
}

func TestNewJLogWithTimestampParam(t *testing.T) {
	// GIVEN a new JLog is wanted with certain params
	timestamps := true

	// WHEN NewJLog is called
	jLog := NewJLog("WARN", timestamps)

	// THEN a JLog pointer is created with this Log Level
	got := (*jLog).Timestamps
	if got != timestamps {
		t.Errorf("NewJLog didn't use level param. Wanted %t but got %t",
			timestamps, got)
	}
}

func TestSetLevelWithUppercaseValidLevel(t *testing.T) {
	// GIVEN you have a valid JLog and want to change the Log Level
	jLog := NewJLog("INFO", false)
	level := "WARN"

	// WHEN SetLevel is called with a valid level
	jLog.SetLevel(level)
	want := uint(1)

	// THEN a JLog pointer is created with this Log Level
	got := (*jLog).Level
	if got != want {
		t.Errorf("SetLevel set the level correctly. Wanted %q (%d) but got %d",
			level, want, got)
	}
}

func TestSetLevelWithLowercaseValidLevel(t *testing.T) {
	// GIVEN you have a valid JLog and want to change the Log Level
	jLog := NewJLog("info", false)
	level := "warn"

	// WHEN SetLevel is called with a valid level
	jLog.SetLevel(level)
	want := uint(1)

	// THEN a JLog pointer is created with this Log Level
	got := (*jLog).Level
	if got != want {
		t.Errorf("SetLevel set the level correctly. Wanted %q (%d) but got %d",
			level, want, got)
	}
}

func TestSetLevelWithInvalidLevel(t *testing.T) {
	// GIVEN you have a valid JLog and want to change the Log Level to something undefined
	jLog := NewJLog("WARN", false)
	level := "something123"
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() { _ = recover() }()

	// WHEN SetLevel is called with an unknown level
	jLog.SetLevel(level)

	// THEN this call will crash the program
	t.Errorf("%s is an unknown log level and should have been Fatal",
		level)
}

func TestSetTimestamps(t *testing.T) {
	// GIVEN you have a valid JLog and want to change the Log Level
	jLog := NewJLog("WARN", false)
	timestamps := true

	// WHEN SetTimestamps is called to invert Timestamps
	jLog.SetTimestamps(timestamps)

	// THEN the Timetamps var is flipped
	got := (*jLog).Timestamps
	if got != timestamps {
		t.Errorf("SetTimestamps didn't set Timestamps correctly. Wanted %t but got %t",
			timestamps, got)
	}
}

func TestFormatMessageSourceWithDefaultLogFrom(t *testing.T) {
	// GIVEN a default LogFrom var
	var logFrom LogFrom

	// WHEN FormatMessageSource is called with this LogFrom
	got := FormatMessageSource(logFrom)
	want := ""

	// THEN an empty string is returned
	if got != want {
		t.Errorf("FormatMessageSource should have returned %q with an empty LogFrom, not %q",
			want, got)
	}
}

func TestFormatMessageSourceWithLogFromPrimaryOnly(t *testing.T) {
	// GIVEN a LogFrom var with only Primary non-default
	logFrom := LogFrom{Primary: "primary"}

	// WHEN FormatMessageSource is called with this LogFrom
	got := FormatMessageSource(logFrom)
	want := "primary, "

	// THEN an the string returned is this primary followed by ', '
	if got != want {
		t.Errorf("FormatMessageSource should have returned %q with only a Primary, not %q",
			want, got)
	}
}

func TestFormatMessageSourceWithLogFromSecondaryOnly(t *testing.T) {
	// GIVEN a LogFrom var with only Primary non-default
	logFrom := LogFrom{Secondary: "secondary"}

	// WHEN FormatMessageSource is called with this LogFrom
	got := FormatMessageSource(logFrom)
	want := "secondary, "

	// THEN an the string returned is this primary followed by ', '
	if got != want {
		t.Errorf("FormatMessageSource should have returned %q with only a Secondary, not %q",
			want, got)
	}
}

func TestFormatMessageSourceWithLogFromPrimaryAndSecondary(t *testing.T) {
	// GIVEN a LogFrom var with only Primary non-default
	logFrom := LogFrom{Primary: "primary", Secondary: "secondary"}

	// WHEN FormatMessageSource is called with this LogFrom
	got := FormatMessageSource(logFrom)
	want := "primary (secondary), "

	// THEN an the string returned is this primary followed by ', '
	if got != want {
		t.Errorf("FormatMessageSource should have returned %q with a Primary and Secondary, not %q",
			want, got)
	}
}

func TestIsLevelPass(t *testing.T) {
	// GIVEN you have a valid JLog
	level := "WARN"
	jLog := NewJLog(level, false)

	// WHEN IsLevel is called to check with the matching level
	check := jLog.IsLevel(level)

	// THEN a JLog pointer is created with this Log Level
	if !check {
		t.Errorf("IsLevel should have got a match on %s and %d, but returned %t",
			level, jLog.Level, check)
	}
}

func TestIsLevelFail(t *testing.T) {
	// GIVEN you have a valid JLog
	level := "WARN"
	jLog := NewJLog(level, false)

	// WHEN IsLevel is called to check with a mismatching level
	guess := "something"
	check := jLog.IsLevel("something")

	// THEN a JLog pointer is created with this Log Level
	if check {
		t.Errorf("IsLevel shouldn't have got a match on %s and %d, but returned %t",
			guess, jLog.Level, check)
	}
}

func TestErrorFalse(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("WARN", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Error is called with false
	jLog.Error(fmt.Errorf(msg), LogFrom{}, false)

	// THEN nothing was logged
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != "" {
		t.Errorf("Error printed when otherCondition was false. (%q)",
			string(out))
	}
}

func TestErrorTrueTimestamps(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("WARN", true)
	var out bytes.Buffer
	log.SetOutput(&out)

	// WHEN Error is called with true
	jLog.Error(fmt.Errorf(msg), LogFrom{}, true)

	// THEN nsg was logged with timestamps
	regex := fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} ERROR: %s\n$", msg)
	reg := regexp.MustCompile(regex)
	got := out.String()
	match := reg.MatchString(got)
	if !match {
		t.Errorf("Error printed didn't match %q. Got %q, want %q",
			regex, got, msg)
	}
}

func TestErrorTrueNoTimestamps(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("WARN", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Error is called with true
	jLog.Error(fmt.Errorf(msg), LogFrom{}, true)

	// THEN nsg was
	want := fmt.Sprintf("ERROR: %s\n", msg)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != want {
		t.Errorf("Error printed didn't match desired. Got %q, want %q",
			string(out), msg)
	}
}

func TestWarnFalse(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("ERROR", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Warn is called with the Log Level higher than what we're calling
	jLog.Warn(fmt.Errorf(msg), LogFrom{}, false)

	// THEN nothing was logged
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != "" {
		t.Errorf("Warn printed when otherCondition was false. (%q)",
			string(out))
	}
}

func TestWarnTrueTimestamps(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("WARN", true)
	var out bytes.Buffer
	log.SetOutput(&out)

	// WHEN Warn is called with true
	jLog.Warn(fmt.Errorf(msg), LogFrom{}, true)

	// THEN nsg was logged with timestamps
	regex := fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} WARNING: %s\n$", msg)
	reg := regexp.MustCompile(regex)
	got := out.String()
	match := reg.MatchString(got)
	if !match {
		t.Errorf("Warn printed didn't match %q. Got %q, want %q",
			regex, got, msg)
	}
}

func TestWarnTrueNoTimestamps(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("WARN", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Warn is called with true
	jLog.Warn(fmt.Errorf(msg), LogFrom{}, true)

	// THEN nsg was
	want := fmt.Sprintf("WARNING: %s\n", msg)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != want {
		t.Errorf("Warn printed didn't match desired. Got %q, want %q",
			string(out), msg)
	}
}

func TestInfoFalse(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("INFO", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Info is called with the Log Level higher than what we're calling
	jLog.Info(fmt.Errorf(msg), LogFrom{}, false)

	// THEN nothing was logged
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != "" {
		t.Errorf("Info printed when otherCondition was false. (%q)",
			string(out))
	}
}

func TestInfoTrueTimestamps(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("INFO", true)
	var out bytes.Buffer
	log.SetOutput(&out)

	// WHEN Info is called with true
	jLog.Info(fmt.Errorf(msg), LogFrom{}, true)

	// THEN nsg was logged with timestamps
	regex := fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} INFO: %s\n$", msg)
	reg := regexp.MustCompile(regex)
	got := out.String()
	match := reg.MatchString(got)
	if !match {
		t.Errorf("Info printed didn't match %q. Got %q, want %q",
			regex, got, msg)
	}
}

func TestInfoTrueNoTimestamps(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("INFO", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Info is called with true
	jLog.Info(fmt.Errorf(msg), LogFrom{}, true)

	// THEN nsg was
	want := fmt.Sprintf("INFO: %s\n", msg)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != want {
		t.Errorf("Info printed didn't match desired. Got %q, want %q",
			string(out), msg)
	}
}

func TestVerboseFalse(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("VERBOSE", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Verbose is called with the Log Level higher than what we're calling
	jLog.Verbose(fmt.Errorf(msg), LogFrom{}, false)

	// THEN nothing was logged
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != "" {
		t.Errorf("Verbose printed when otherCondition was false. (%q)",
			string(out))
	}
}

func TestVerboseTrueTimestamps(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("VERBOSE", true)
	var out bytes.Buffer
	log.SetOutput(&out)

	// WHEN Verbose is called with true
	jLog.Verbose(fmt.Errorf(msg), LogFrom{}, true)

	// THEN nsg was logged with timestamps
	regex := fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} VERBOSE: %s\n$", msg)
	reg := regexp.MustCompile(regex)
	got := out.String()
	match := reg.MatchString(got)
	if !match {
		t.Errorf("Verbose printed didn't match %q. Got %q, want %q",
			regex, got, msg)
	}
}

func TestVerboseTrueNoTimestamps(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("VERBOSE", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Verbose is called with true
	jLog.Verbose(fmt.Errorf(msg), LogFrom{}, true)

	// THEN nsg was
	want := fmt.Sprintf("VERBOSE: %s\n", msg)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != want {
		t.Errorf("Verbose printed didn't match desired. Got %q, want %q",
			string(out), msg)
	}
}

func TestDebugFalse(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("DEBUG", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Debug is called with the Log Level higher than what we're calling
	jLog.Debug(fmt.Errorf(msg), LogFrom{}, false)

	// THEN nothing was logged
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != "" {
		t.Errorf("Debug printed when otherCondition was false. (%q)",
			string(out))
	}
}

func TestDebugTrueTimestamps(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("DEBUG", true)
	var out bytes.Buffer
	log.SetOutput(&out)

	// WHEN Debug is called with true
	jLog.Debug(fmt.Errorf(msg), LogFrom{}, true)

	// THEN nsg was logged with timestamps
	regex := fmt.Sprintf("^[0-9]{4}\\/[0-9]{2}\\/[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} DEBUG: %s\n$", msg)
	reg := regexp.MustCompile(regex)
	got := out.String()
	match := reg.MatchString(got)
	if !match {
		t.Errorf("Debug printed didn't match %q. Got %q, want %q",
			regex, got, msg)
	}
}

func TestDebugTrueNoTimestamps(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("DEBUG", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Debug is called with true
	jLog.Debug(fmt.Errorf(msg), LogFrom{}, true)

	// THEN nsg was
	want := fmt.Sprintf("DEBUG: %s\n", msg)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != want {
		t.Errorf("Debug printed didn't match desired. Got %q, want %q",
			string(out), msg)
	}
}

func TestFatalFalse(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("ERROR", false)
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Fatal is called with the Log Level higher than what we're calling
	jLog.Fatal(fmt.Errorf(msg), LogFrom{}, false)

	// THEN nothing was logged
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	if string(out) != "" {
		t.Errorf("Fatal printed when otherCondition was false. (%q)",
			string(out))
	}
}

func TestFatalTrue(t *testing.T) {
	// GIVEN a JLog and message
	msg := "argus"
	jLog := NewJLog("ERROR", true)
	var out bytes.Buffer
	log.SetOutput(&out)

	// WHEN Fatal is called with true
	if os.Getenv("RUN_CRASH") == "1" {
		jLog.Fatal(fmt.Errorf(msg), LogFrom{}, true)
		return
	}

	// THEN nsg will crash the program (os.Exit)
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalTrue")
	cmd.Env = append(os.Environ(), "RUN_CRASH=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	got := out.String()
	t.Errorf("Fatal print didn't os.Exit. Got %q",
		got)
}
