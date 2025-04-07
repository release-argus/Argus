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

// Package logutil provides a logger.
package logutil

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/release-argus/Argus/util"
)

var (
	Log  *JLog
	once sync.Once
)

// Init initialises the logging system with the specified log level.
// The log level determines the severity of the messages that will be logged.
// Valid log levels are "debug", "verbose", "info", "warn" and "error".
func Init(level string, timestamps bool) {
	once.Do(func() {
		Log = NewJLog(level, timestamps)
	})
}

var (
	levelMap = map[string]uint8{
		"ERROR":   0,
		"WARN":    1,
		"INFO":    2,
		"VERBOSE": 3,
		"DEBUG":   4,
	}
)

// JLog handles logging at multiple levels.
//
// It supports ERROR, WARNING, INFO, VERBOSE, and DEBUG.
type JLog struct {
	mutex sync.RWMutex
	// Minimum level of logs to print.
	//	0 = ERROR
	//	1 = WARN
	//	2 = INFO
	//	3 = VERBOSE
	//	4 = DEBUG
	Level      uint8
	Timestamps bool // Whether to log timestamps with the msg, or just the msg.

	Testing bool // Indicates if running in tests (avoids panic in Fatal).
}

// LogFrom is the source of the log.
type LogFrom struct {
	Primary   string
	Secondary string
}

// NewJLog creates a new JLog with the given log level and timestamps.
func NewJLog(level string, timestamps bool) *JLog {
	newJLog := JLog{}
	newJLog.SetLevel(level)
	newJLog.SetTimestamps(timestamps)
	return &newJLog
}

// SetLevel modifies the logging level.
func (l *JLog) SetLevel(level string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// New log level.
	level = strings.ToUpper(level)
	value := levelMap[level]

	if value == 0 && level != "ERROR" {
		l.Fatal(fmt.Sprintf("%q is not a valid log.level. It should be one of ERROR, WARN, INFO, VERBOSE or DEBUG.",
			level),
			LogFrom{}, true)
	}
	// Set the log level if it has changed.
	if value != l.Level {
		l.Level = value
	}
}

// SetTimestamps on the logs.
func (l *JLog) SetTimestamps(enable bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Set the timestamps flag if it has changed.
	if enable != l.Timestamps {
		l.Timestamps = enable
	}
}

// String produces the string representation of the LogFrom.
func (lf LogFrom) String() string {
	// Have Primary.
	if lf.Primary != "" {
		// Have Primary and Secondary.
		if lf.Secondary != "" {
			return fmt.Sprintf("%s (%s), ",
				lf.Primary, lf.Secondary)
		}
		// Just Primary.
		return lf.Primary + ", "
	}

	// Just Secondary.
	if lf.Secondary != "" {
		return lf.Secondary + ", "
	}

	// Neither Primary nor Secondary.
	return ""
}

// IsLevel checks if the JLog `level` matches the provided `level`.
func (l *JLog) IsLevel(level string) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	level = strings.ToUpper(level)
	value := levelMap[level]

	if value == 0 && level != "ERROR" {
		return false
	}
	return l.Level == value
}

// Fatal calls Error() followed by an os.Exit(1).
func (l *JLog) Fatal(msg any, from LogFrom, otherCondition bool) {
	if otherCondition {
		l.Error(msg, from, true)
		if !l.Testing {
			os.Exit(1)
		}
		panic(msg)
	}
}

// Error log the msg.
//
// (if `otherCondition` true).
func (l *JLog) Error(msg any, from LogFrom, otherCondition bool) {
	if !otherCondition {
		return
	}

	// ERROR: msg from.Primary (from.Secondary)
	l.logMessage(
		fmt.Sprintf("ERROR: %s%v",
			from, msg))
}

// Warn log msg if l.Level > 0 (WARNING, INFO, VERBOSE or DEBUG).
//
// (if otherCondition true).
func (l *JLog) Warn(msg any, from LogFrom, otherCondition bool) {
	if l.Level == 0 || !otherCondition {
		return
	}

	// WARNING: msg from.Primary (from.Secondary)
	l.logMessage(
		fmt.Sprintf("WARNING: %s%v",
			from, msg))
}

// Info log msg if l.Level > 1 (INFO, VERBOSE or DEBUG).
//
// (if otherCondition true).
func (l *JLog) Info(msg any, from LogFrom, otherCondition bool) {
	if l.Level < 2 || !otherCondition {
		return
	}

	// INFO: msg from.Primary (from.Secondary)
	l.logMessage(
		fmt.Sprintf("INFO: %s%v",
			from, msg))
}

// Verbose log msg if l.Level > 2 (VERBOSE or DEBUG).
//
// (if otherCondition true).
func (l *JLog) Verbose(msg any, from LogFrom, otherCondition bool) {
	if l.Level < 3 || !otherCondition {
		return
	}

	// VERBOSE: msg from.Primary (from.Secondary)
	l.logMessage(
		util.TruncateMessage(
			fmt.Sprintf("VERBOSE: %s%v", from, msg),
			997))
}

// Debug log msg if l.Level 4 (DEBUG).
//
// (if otherCondition true).
func (l *JLog) Debug(msg any, from LogFrom, otherCondition bool) {
	if l.Level != 4 || !otherCondition {
		return
	}

	// DEBUG: msg from.Primary (from.Secondary)
	l.logMessage(
		util.TruncateMessage(
			fmt.Sprintf("DEBUG: %s%v", from, msg),
			997))
}

// logMessage logs a message with/without a timestamp based on the Timestamps flag.
func (l *JLog) logMessage(msg string) {
	if l.Timestamps {
		log.Println(msg)
	} else {
		fmt.Println(msg)
	}
}
