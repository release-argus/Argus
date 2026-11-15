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
	"os"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
)

func TestInitLog(t *testing.T) {
	envKey := "ARGUS_LOG_LEVEL"
	env := map[string]string{envKey: ""}
	test.IsolateEnv(t, env)

	InitLog()

	// Set in the env, so a later Config.Load() keeps it.
	got := os.Getenv(envKey)
	wantLogLevel := "DEBUG"
	if got != wantLogLevel {
		t.Errorf("os.Getenv(envKey) mismatch\ngot:  %q\nwant: %q", got, wantLogLevel)
	}

	if got, want := logx.Level(), uint32(4); got != want {
		t.Errorf("logx.Level() mismatch\ngot:  %d\nwant: %d (DEBUG)", got, want)
	}
}
