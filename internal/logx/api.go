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

package logx

import "io"

// Default returns the package logger.
func Default() *Logger {
	return logger
}

// Level returns the current minimum log level of the package logger.
func Level() uint32 {
	return loggerInstance().Level.Load()
}

// SetLevel modifies the logging level of the package logger.
func SetLevel(level string) {
	loggerInstance().SetLevel(level)
}

// SetTimestamps enables or disables timestamps on the package logger.
func SetTimestamps(enable bool) {
	loggerInstance().SetTimestamps(enable)
}

// SetOutput sets the output destination of the package logger.
func SetOutput(w io.Writer) {
	loggerInstance().SetOutput(w)
}

// GetOutput returns the output destination of the package logger.
func GetOutput() io.Writer {
	return loggerInstance().GetOutput()
}

// SetExitCodeChannel sets the exit code channel on the package logger.
func SetExitCodeChannel(exitCodeChannel chan string) {
	loggerInstance().SetExitCodeChannel(exitCodeChannel)
}

// ExitCodeChannel returns the exit code channel of the package logger.
func ExitCodeChannel() chan string {
	return loggerInstance().exitCodeChannel
}

// IsLevel reports whether the package logger is at the given level.
func IsLevel(level string) bool {
	return loggerInstance().IsLevel(level)
}

// Fatal logs a fatal message on the package logger.
func Fatal(msg any, from LogFrom) {
	loggerInstance().Fatal(msg, from)
}

// Error logs an error message on the package logger.
func Error(msg any, from LogFrom, otherCondition bool) {
	loggerInstance().Error(msg, from, otherCondition)
}

// Warn logs a warning message on the package logger.
func Warn(msg any, from LogFrom, otherCondition bool) {
	loggerInstance().Warn(msg, from, otherCondition)
}

// Info logs an info message on the package logger.
func Info(msg any, from LogFrom, otherCondition bool) {
	loggerInstance().Info(msg, from, otherCondition)
}

// Verbose logs a verbose message on the package logger.
func Verbose(msg any, from LogFrom, otherCondition bool) {
	loggerInstance().Verbose(msg, from, otherCondition)
}

// Debug logs a debug message on the package logger.
func Debug(msg any, from LogFrom, otherCondition bool) {
	loggerInstance().Debug(msg, from, otherCondition)
}

// Deprecated logs a deprecation message on the package logger.
func Deprecated(msg any) {
	loggerInstance().Deprecated(msg)
}

func loggerInstance() *Logger {
	if logger == nil {
		panic("logx: Init not called")
	}
	return logger
}
