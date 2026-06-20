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

package log

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
)

var packageName = "logtest"

func TestInitLog(t *testing.T) {
	// GIVEN: the environment variable is not set.
	envKey := "ARGUS_LOG_LEVEL"
	env := map[string]string{envKey: ""}
	test.SetEnv(t, env)

	// WHEN: InitLog is called.
	InitLog()

	prefix := fmt.Sprintf("%s\nInitLog()", packageName)

	// THEN: the environment variable should be set to "DEBUG".
	got := os.Getenv(envKey)
	wantLogLevel := "DEBUG"
	if got != wantLogLevel {
		t.Errorf(
			"%s unexpected %s value\ngot:  %q\nwant: %q",
			prefix, envKey,
			got, wantLogLevel,
		)
	}

	// AND: the log level should be set to DEBUG.
	logx.SetLevel("DEBUG")
	if logx.Level() != 4 {
		t.Errorf(
			"%s log level mismatch\ngot %d\nwant: %d (DEBUG)",
			prefix, logx.Level(), 4,
		)
	}
}
