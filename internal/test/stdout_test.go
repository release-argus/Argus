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

package test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

type Log struct {
	writer *log.Logger
}

func NewLog() *Log {
	return &Log{
		writer: log.New(os.Stdout, "", 0),
	}
}

func (l *Log) SetOutput(w io.Writer) {
	l.writer.SetOutput(w)
}

func (l *Log) GetOutput() io.Writer {
	return l.writer.Writer()
}

func (l *Log) Write(msg string) {
	l.writer.Println(msg)
}

func TestCaptureLog(t *testing.T) {
	// GIVEN: a function that writes to stdout.
	tests := []struct {
		name string
		fn   func(logger *Log)
		want string
	}{
		{
			name: "single line",
			fn: func(logger *Log) {
				logger.Write("hello")
			},
			want: "hello\n",
		},
		{
			name: "multiple lines",
			fn: func(logger *Log) {
				logger.Write("hello")
				logger.Write("world")
			},
			want: "hello\nworld\n",
		},
		{
			name: "empty",
			fn:   func(logger *Log) {},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a logger.
			buf := &bytes.Buffer{}
			logger := NewLog()
			logger.SetOutput(buf)

			// WHEN: CaptureLog is called.
			capture := CaptureLog(t, logger)
			tc.fn(logger)
			result := capture()

			// THEN: the result should be the expected stdout output.
			if result != tc.want {
				t.Errorf(
					"%s\nCaptureLog() stdout mismatch\ngot:  %q\nwant: %q",
					packageName, result, tc.want,
				)
			}
		})
	}
}

func TestCaptureStdout(t *testing.T) {
	// GIVEN: a function that writes to stdout.
	tests := []struct {
		name string
		fn   func()
		want string
	}{
		{
			name: "single line",
			fn: func() {
				fmt.Println("hello")
			},
			want: "hello\n",
		},
		{
			name: "multiple lines",
			fn: func() {
				fmt.Println("hello")
				fmt.Println("world")
			},
			want: "hello\nworld\n",
		},
		{
			name: "empty",
			fn:   func() {},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			// WHEN: CaptureStdout is called.
			capture := CaptureStdout(t)
			tc.fn()
			result := capture()

			// THEN: the result should be the expected stdout output.
			if result != tc.want {
				t.Errorf(
					"%s\nCaptureStdout() stdout mismatch\ngot:  %q\nwant: %q",
					packageName, result, tc.want,
				)
			}
		})
	}
}
