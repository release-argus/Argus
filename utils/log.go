// Copyright [2022] [Hymenaios]
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

package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// JLog is a log for various levels of logging.
//
// It supports ERROR, WARNING, INFO, VERBOSE and DEBUG.
type JLog struct {
	// Level is the level of the Log
	// 0 = ERROR  1 = WARN,
	// 2 = INFO,  3 = VERBOSE,
	// 4 = DEBUG
	Level      uint
	LevelStr   string // The string value of Level (ERROR/WARN/INFO/VERBOSE/DEBUG)
	Timestamps bool   // whether to log timestamps with the msg, or just the msg.
}

type LogFrom struct {
	Primary   string
	Secondary string
}

func NewJLog(level string, timestamps bool) *JLog {
	new := JLog{}
	new.SetLevel(level)
	new.SetTimestamps(timestamps)
	return &new
}

// SetLevel of the JLog.
//
// If value is out of the range (<0 or >4), then exit.
func (l *JLog) SetLevel(level string) {
	level = strings.ToUpper(level)
	levelMap := map[string]int{
		"ERROR":   0,
		"WARN":    1,
		"INFO":    2,
		"VERBOSE": 3,
		"DEBUG":   4,
	}
	value := levelMap[level]
	l.LevelStr = level

	msg := fmt.Sprintf("%q is not a valid log.level. It should be one of ERROR, WARN, INFO, VERBOSE or DEBUG.", level)
	l.Fatal(msg, LogFrom{}, value == 0 && level != "ERROR")

	l.Level = uint(value)
}

// SetTimestamps on the logs.
func (l *JLog) SetTimestamps(enable bool) {
	l.Timestamps = enable
}

// FormatMessageSource for logging.
//
// from.Primary and from.Secondary defined = `from.Primary (from.Secondary) `
//
// from.Primary defined = `from.Primary `
//
// from.Secondary defined = `from.Secondary `
func FormatMessageSource(from LogFrom) string {
	// from.Primary defined
	if from.Primary != "" {
		// from.Primary and from.Secondary are defined
		if from.Secondary != "" {
			return fmt.Sprintf("%s (%s), ", from.Primary, from.Secondary)
		}
		// Just from.Primary defined
		return fmt.Sprintf("%s, ", from.Primary)
	}

	// Just from.Secondary defined
	if from.Secondary != "" {
		return fmt.Sprintf("%s, ", from.Secondary)
	}

	// Neither from.Primary nor from.Secondary defined
	return ""
}

// Error log the msg.
//
// (if otherCondition is true)
func (l *JLog) Error(msg interface{}, from LogFrom, otherCondition bool) {
	msgString := fmt.Sprintf("%s%v", FormatMessageSource(from), msg)

	if otherCondition {
		if l.Timestamps {
			log.Printf("ERROR: %s\n", msgString)
		} else {
			fmt.Printf("ERROR: %s\n", msgString)
		}
	}
}

// Warn log msg if l.Level is > 0 (WARNING, INFO, VERBOSE or DEBUG).
//
// (if otherCondition is true)
func (l *JLog) Warn(msg interface{}, from LogFrom, otherCondition bool) {
	msgString := fmt.Sprintf("%s%v", FormatMessageSource(from), msg)

	if l.Level > 0 && otherCondition {
		if l.Timestamps {
			log.Printf("WARNING: %s\n", msgString)
		} else {
			fmt.Printf("WARNING: %s\n", msgString)
		}
	}
}

// Info log msg if l.Level is > 1 (INFO, VERBOSE or DEBUG).
//
// (if otherCondition is true)
func (l *JLog) Info(msg interface{}, from LogFrom, otherCondition bool) {
	msgString := fmt.Sprintf("%s%v", FormatMessageSource(from), msg)

	if l.Level > 1 && otherCondition {
		// INFO: msg from.Primary (from.Secondary)
		if l.Timestamps {
			log.Printf("INFO: %s\n", msgString)
		} else {
			fmt.Printf("INFO: %s\n", msgString)
		}
	}
}

// Verbose log msg if l.Level is > 2 (VERBOSE or DEBUG).
//
// (if otherCondition is true)
func (l *JLog) Verbose(msg interface{}, from LogFrom, otherCondition bool) {
	msgString := fmt.Sprintf("%s%v", FormatMessageSource(from), msg)

	if l.Level > 2 && otherCondition {
		if l.Timestamps {
			log.Printf("VERBOSE: %s\n", msgString)
		} else {
			fmt.Printf("VERBOSE: %s\n", msgString)
		}
	}
}

// Debug log msg if l.Level is 4 (DEBUG).
//
// (if otherCondition is true)
func (l *JLog) Debug(msg interface{}, from LogFrom, otherCondition bool) {
	msgString := fmt.Sprintf("%s%v", FormatMessageSource(from), msg)

	if l.Level == 4 && otherCondition {
		if l.Timestamps {
			log.Printf("DEBUG: %s\n", msgString)
		} else {
			fmt.Printf("DEBUG: %s\n", msgString)
		}
	}
}

// Fatal is equivalent to Error() followed by a call to os.Exit(1).
func (l *JLog) Fatal(msg interface{}, from LogFrom, otherCondition bool) {
	if otherCondition {
		l.Error(msg, from, true)
		os.Exit(1)
	}
}
