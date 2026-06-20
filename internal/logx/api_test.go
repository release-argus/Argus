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
	once = sync.Once{}
	logger = nil

	exitCodeChannel := Init("DEBUG", false)
	if Default() != logger {
		t.Fatalf("%s Default() pointer mismatch", packageName)
	}
	if ExitCodeChannel() != exitCodeChannel {
		t.Fatalf("%s ExitCodeChannel() pointer mismatch", packageName)
	}

	SetLevel("INFO")
	t.Cleanup(func() { logger.SetLevel("DEBUG") })
	if logger.Level.Load() != levelMap["INFO"] {
		t.Fatalf("%s SetLevel(INFO) returned false", packageName)
	}
	if Level() != levelMap["INFO"] {
		t.Fatalf("%s Level() value mismatch", packageName)
	}

	SetTimestamps(true)
	t.Cleanup(func() { logger.SetTimestamps(false) })
	if !IsLevel("INFO") {
		t.Fatalf("%s IsLevel(INFO) returned false", packageName)
	}

	buf := &bytes.Buffer{}
	SetOutput(buf)
	if GetOutput() != buf {
		t.Fatalf("%s GetOutput() pointer mismatch", packageName)
	}

	newChannel := make(chan string, 1)
	SetExitCodeChannel(newChannel)
	if logger.exitCodeChannel != newChannel {
		t.Fatalf("%s SetExitCodeChannel() pointer mismatch", packageName)
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
		t.Fatalf("%s expected package-level log output", packageName)
	}
}

func TestLoggerInstance_Panic(t *testing.T) {
	hadLog := logger
	t.Cleanup(func() {
		logger = hadLog
	})

	// GIVEN: logger is not initialised.
	once = sync.Once{}
	logger = nil

	// THEN: loggerInstance will panic.
	defer func() {
		if recover() == nil {
			t.Fatalf("%s expected panic when logger is not initialised", packageName)
		}
	}()

	// WHEN: loggerInstance is called.
	_ = Level()
}
