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

//go:build unit || integration

package test

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
)

type Logger interface {
	GetOutput() io.Writer
	SetOutput(io.Writer)
}

var (
	logMu               sync.Mutex // Only one test should write to log at a time.
	testHoldingLogMutex string     // Name of the test holding the log mutex.
)

// CaptureLog temporarily captures all output written to the log
// and returns a function that, when called, restores the original log and
// returns the captured output as a string.
func CaptureLog(t *testing.T, log Logger) func() string {
	var buf bytes.Buffer
	logMu.Lock()
	testHoldingLogMutex = t.Name()
	old := log.GetOutput()
	log.SetOutput(&buf)

	cleanup := func() string {
		if testHoldingLogMutex == t.Name() {
			testHoldingLogMutex = ""
			log.SetOutput(old)
			logMu.Unlock()
			return buf.String()
		}
		return ""
	}

	// Ensure the log is restored when the test ends.
	t.Cleanup(func() {
		_ = cleanup()
	})

	return cleanup
}

var (
	stdoutMu               sync.Mutex // Only one test should write to stdout at a time.
	testHoldingStdoutMutex string     // Name of the test holding the stdout mutex.
)

// CaptureStdout temporarily captures all output written to the standard output
// and returns a function that, when called, restores the original standard output and
// returns the captured output as a string.
func CaptureStdout(t *testing.T) func() string {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	stdoutMu.Lock()
	testHoldingStdoutMutex = t.Name()

	var buf bytes.Buffer
	done := make(chan struct{})

	// Drain the pipe until closed.
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	cleanup := func() string {
		if testHoldingStdoutMutex == t.Name() {
			testHoldingStdoutMutex = ""
			_ = w.Close()
			<-done

			os.Stdout = stdout
			stdoutMu.Unlock()
			return buf.String()
		}
		return ""
	}

	// Ensure stdout is restored when the test ends.
	t.Cleanup(func() {
		_ = cleanup()
	})

	return cleanup
}
