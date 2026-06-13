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

// Package logx provides a logger for Argus.
package logx

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

var (
	logger *Logger
	once   sync.Once
)

func formatMsg(msg any) string {
	if err, ok := msg.(error); ok {
		return errfmt.FormatError(err)
	}
	return fmt.Sprint(msg)
}

// Init initialises the logging system with the specified log level.
// The log level determines the severity of the messages that will be logged.
// Valid log levels are "debug", "verbose", "info", "warn" and "error".
func Init(level string, timestamps bool) chan string {
	once.Do(func() {
		logger = NewLogger(level, timestamps)
	})
	return logger.exitCodeChannel
}

var (
	levelMap = map[string]uint32{
		"ERROR":   0,
		"WARN":    1,
		"INFO":    2,
		"VERBOSE": 3,
		"DEBUG":   4,
	}
)

// Logger handles logging at multiple levels.
//
// It supports ERROR, WARNING, INFO, VERBOSE, and DEBUG.
type Logger struct {
	mu sync.RWMutex
	// Minimum level of logs to print.
	//	0 = ERROR
	//	1 = WARN
	//	2 = INFO
	//	3 = VERBOSE
	//	4 = DEBUG
	// Level           uint8
	Level           atomic.Uint32
	timestamps      bool        // Whether to log a timestamp with each msg or just the msg.
	exitCodeChannel chan string // Shutdown handler.

	writer *log.Logger // Internal logger used for printing messages with the configured format and output.
	out    io.Writer   // The current destination for log output.
}

// LogFrom is the source of the log.
type LogFrom struct {
	Primary   string
	Secondary string
}

// writerFor returns a logger that writes to w.
// The logger includes standard timestamps when timestamps is true.
func writerFor(w io.Writer, timestamps bool) *log.Logger {
	flags := 0
	if timestamps {
		flags = log.LstdFlags
	}
	return log.New(w, "", flags)
}

// NewLogger creates a new Logger with the given log level and timestamps.
func NewLogger(level string, timestamps bool) *Logger {
	newLogger := Logger{
		out:        os.Stdout,
		timestamps: timestamps,
		writer:     writerFor(os.Stdout, timestamps),
	}
	newLogger.SetExitCodeChannel(make(chan string, 1))
	newLogger.SetLevel(level)

	return &newLogger
}

// SetExitCodeChannel sets the exit code to send to on 'Fatal' errors.
func (l *Logger) SetExitCodeChannel(exitCodeChannel chan string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.exitCodeChannel = exitCodeChannel
}

// ExitCodeChannel returns the [Logger.exitCodeChannel].
func (l *Logger) ExitCodeChannel() chan string {
	return l.exitCodeChannel
}

// SetLevel modifies the logging level.
func (l *Logger) SetLevel(level string) {
	// New log level.
	level = strings.ToUpper(level)
	value := levelMap[level]

	if value == 0 && level != "ERROR" {
		l.Fatal(
			fmt.Sprintf("%q is not a valid log.level. It should be one of %s",
				level, strings.Join(util.SortedKeys(levelMap), ", "),
			),
			LogFrom{},
		)
	}
	// Set the log level if it has changed.
	if currentLevel := l.Level.Load(); currentLevel != value {
		l.Level.Store(value)
	}
}

// IsLevel checks if the [Logger.level] matches the provided `level`.
func (l *Logger) IsLevel(level string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	level = strings.ToUpper(level)
	value := levelMap[level]

	if value == 0 && level != "ERROR" {
		return false
	}
	return l.Level.Load() == value
}

// SetTimestamps on the logs.
func (l *Logger) SetTimestamps(enable bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if enable != l.timestamps {
		l.timestamps = enable
		l.writer = writerFor(l.out, enable)
	}
}

// String implements [fmt.Stringer].
func (lf LogFrom) String() string {
	// Have Primary.
	if lf.Primary != "" {
		// Have Primary and Secondary.
		if lf.Secondary != "" {
			return fmt.Sprintf(
				"%s (%s), ",
				lf.Primary, lf.Secondary,
			)
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

// SetOutput sets the destination writer for log output.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.writer.SetOutput(w)
	l.out = w
}

// GetOutput returns the current log output writer.
func (l *Logger) GetOutput() io.Writer {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.out
}

// Fatal calls Error() followed by a message to the exit code channel
// (which signals a shutdown of the process).
func (l *Logger) Fatal(msg any, from LogFrom) {
	fullMsg := fmt.Sprintf(
		"FATAL: %s%s",
		from, formatMsg(msg),
	)
	l.logMessage(fullMsg)

	if l.exitCodeChannel != nil {
		l.exitCodeChannel <- fullMsg
	}
}

// Error will log the message if the log level is ERROR or higher.
//
// (if `otherCondition` true).
func (l *Logger) Error(msg any, from LogFrom, otherCondition bool) {
	if !otherCondition {
		return
	}

	// ERROR: msg from.Primary (from.Secondary)
	l.logMessage(
		fmt.Sprintf(
			"ERROR: %s%v",
			from, formatMsg(msg),
		),
	)
}

// Warn will log the message if the log level is WARN or higher.
//
// (if otherCondition true).
func (l *Logger) Warn(msg any, from LogFrom, otherCondition bool) {
	if l.Level.Load() == 0 || !otherCondition {
		return
	}

	// WARNING: msg from.Primary (from.Secondary)
	l.logMessage(
		fmt.Sprintf(
			"WARNING: %s%v",
			from, formatMsg(msg),
		),
	)
}

// Info will log the message if the log level is INFO or higher.
//
// (if otherCondition true).
func (l *Logger) Info(msg any, from LogFrom, otherCondition bool) {
	if l.Level.Load() < 2 || !otherCondition {
		return
	}

	// INFO: msg from.Primary (from.Secondary)
	l.logMessage(
		fmt.Sprintf(
			"INFO: %s%v",
			from, formatMsg(msg),
		),
	)
}

// Verbose will log the message if the log level is VERBOSE or higher.
//
// (if otherCondition true).
func (l *Logger) Verbose(msg any, from LogFrom, otherCondition bool) {
	if l.Level.Load() < 3 || !otherCondition {
		return
	}

	// VERBOSE: msg from.Primary (from.Secondary)
	l.logMessage(
		util.TruncateMessage(
			fmt.Sprintf(
				"VERBOSE: %s%v",
				from, formatMsg(msg),
			),
			997,
		),
	)
}

// Debug will log the message if the log level is DEBUG.
//
// (if otherCondition true).
func (l *Logger) Debug(msg any, from LogFrom, otherCondition bool) {
	if l.Level.Load() != 4 || !otherCondition {
		return
	}

	// DEBUG: msg from.Primary (from.Secondary)
	l.logMessage(
		util.TruncateMessage(
			fmt.Sprintf(
				"DEBUG: %s%v",
				from, formatMsg(msg),
			),
			997,
		),
	)
}

// Deprecated will log the deprecation message.
func (l *Logger) Deprecated(msg any) {
	// DEPRECATED: msg
	l.logMessage(fmt.Sprintf("DEPRECATED: %v", msg))
}

// logMessage logs a message with/without a timestamp based on the Timestamps flag.
func (l *Logger) logMessage(msg string) {
	msg = strings.TrimRight(msg, "\n")
	l.writer.Println(msg)
}
