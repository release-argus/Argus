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

package logx

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestPackageAPI(t *testing.T) {
	hadLog := logger
	t.Cleanup(func() { logger = hadLog })

	once = sync.Once{}
	logger = nil

	exitCodeChannel := Init("DEBUG", false)

	if Default() != logger {
		t.Fatalf("Default() mismatch\ngot:  %p\nwant: %p", Default(), logger)
	}
	if ExitCodeChannel() != exitCodeChannel {
		t.Fatalf(
			"ExitCodeChannel() mismatch\ngot:  %p\nwant: %p",
			ExitCodeChannel(), exitCodeChannel,
		)
	}

	SetLevel("INFO")
	if got, want := logger.Level.Load(), levelMap["INFO"]; got != want {
		t.Fatalf("Logger.Level mismatch\ngot:  %d\nwant: %d", got, want)
	}
	if got, want := Level(), levelMap["INFO"]; got != want {
		t.Fatalf("Level() mismatch\ngot:  %d\nwant: %d", got, want)
	}

	SetTimestamps(true)
	if !logger.timestamps {
		t.Fatal("Logger.timestamps mismatch\ngot:  false\nwant: true")
	}
	if !IsLevel("INFO") {
		t.Fatal("IsLevel() mismatch\ngot:  false\nwant: true")
	}

	buf := &bytes.Buffer{}
	SetOutput(buf)
	if GetOutput() != buf {
		t.Fatalf("GetOutput() mismatch\ngot:  %p\nwant: %p", GetOutput(), buf)
	}

	newChannel := make(chan string, 1)
	SetExitCodeChannel(newChannel)
	if logger.exitCodeChannel != newChannel {
		t.Fatalf(
			"Logger.exitCodeChannel mismatch\ngot:  %p\nwant: %p",
			logger.exitCodeChannel, newChannel,
		)
	}

	releaseStdout := test.CaptureLog(t, Default())
	from := LogFrom{Primary: "pkg"}

	Info("info", from, true)
	Warn("warn", from, true)
	Error(errors.New("error"), from, true)
	Verbose("verbose", from, true)
	Debug("debug", from, true)
	Deprecated("deprecated")
	Fatal("fatal", from)

	if got := releaseStdout(); got == "" {
		t.Fatal("releaseStdout() mismatch\ngot:  \"\"\nwant: NOT \"\"")
	}
}

func TestLoggerInstance_Panic(t *testing.T) {
	hadLog := logger
	t.Cleanup(func() {
		logger = hadLog
	})

	once = sync.Once{}
	logger = nil

	defer func() {
		if recover() == nil {
			t.Fatal("recover() mismatch\ngot:  nil\nwant: non-nil")
		}
	}()

	_ = Level()
}
