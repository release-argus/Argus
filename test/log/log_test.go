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

//go:build unit

package logtest

import (
	"os"
	"testing"

	logutil "github.com/release-argus/Argus/util/log"
)

var packageName = "logtest"

func TestInitLog(t *testing.T) {
	// GIVEN the environment variable is not set.
	envKey := "ARGUS_LOG_LEVEL"
	os.Unsetenv(envKey)

	// WHEN InitLog is called.
	InitLog()

	// THEN the environment variable should be set to "DEBUG".
	got := os.Getenv(envKey)
	wantLogLevel := "DEBUG"
	if got != wantLogLevel {
		t.Errorf("%s\nunexpected %s\nwant: %q\ngot:  %q",
			packageName, envKey, wantLogLevel, got)
	}
	// AND the log level should be set to DEBUG.
	hadLogLevel := logutil.Log.Level
	logutil.Log.SetLevel("DEBUG")
	debugLogLevel := logutil.Log.Level
	if debugLogLevel != hadLogLevel {
		t.Errorf("%s\nlog level mismatch\nwant: %d DEBUG\ngot %d",
			packageName, debugLogLevel, hadLogLevel)
	}
	// AND the log should be in testing mode.
	if !logutil.Log.Testing {
		t.Errorf("%s\nexpected log to be in testing mode",
			packageName)
	}
}
