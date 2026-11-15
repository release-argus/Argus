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

			buf := &bytes.Buffer{}
			logger := NewLog()
			logger.SetOutput(buf)

			capture := CaptureLog(t, logger)
			tc.fn(logger)
			result := capture()

			if result != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", result, tc.want)
			}
		})
	}
}

func TestCaptureStdout(t *testing.T) {
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
			// Serial: captures stdout.

			capture := CaptureStdout(t)
			tc.fn()
			result := capture()

			if result != tc.want {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", result, tc.want)
			}
		})
	}
}
