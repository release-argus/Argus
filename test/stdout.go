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

//go:build unit || integration

package test

import (
	"bytes"
	"io"
	"os"
	"sync"
)

type Logger interface {
	GetOutput() io.Writer
	SetOutput(io.Writer)
}

var LogMutex sync.Mutex // Only one test should write to log at a time.

// CaptureLog temporarily captures all output written to the log
// and returns a function that, when called, restores the original log and
// returns the captured output as a string.
func CaptureLog(log Logger) func() string {
	var buf bytes.Buffer
	LogMutex.Lock()
	old := log.GetOutput()
	log.SetOutput(&buf)

	return func() string {
		log.SetOutput(old)
		LogMutex.Unlock()
		return buf.String()
	}
}

var StdoutMutex sync.Mutex // Only one test should write to stdout at a time.

// CaptureStdout temporarily captures all output written to the standard output
// and returns a function that, when called, restores the original standard output and
// returns the captured output as a string.
func CaptureStdout() func() string {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	StdoutMutex.Lock()

	var buf bytes.Buffer
	done := make(chan struct{})

	// Drain the pipe until closed.
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	return func() string {
		_ = w.Close()
		<-done

		os.Stdout = stdout
		StdoutMutex.Unlock()
		return buf.String()
	}
}
