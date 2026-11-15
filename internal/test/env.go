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
	"os"
	"strings"
	"sync"
	"testing"
)

var envMu sync.Mutex // Only one test should modify environment variables at a time.

// argusEnvPrefix is the prefix of the env vars the config MapEnvToStruct methods read.
const argusEnvPrefix = "ARGUS_"

// argusTestEnvPrefix marks test-infra vars (e.g. ARGUS_TEST_REQUIRE_SECRETS) that test.IsolateEnv keeps.
const argusTestEnvPrefix = "ARGUS_TEST_"

// IsolateEnv unsets ARGUS_* env vars (except ARGUS_TEST_*) and sets vars. Cleanup reverts only
// the keys that changed, so parallel tests reading other vars never see them cleared.
func IsolateEnv(t *testing.T, vars map[string]string) {
	t.Helper()

	envMu.Lock()
	originals := SplitEnvVars(os.Environ())
	t.Cleanup(func() {
		for k := range SplitEnvVars(os.Environ()) {
			if _, ok := originals[k]; !ok {
				_ = os.Unsetenv(k)
			}
		}
		for k, v := range originals {
			if got, ok := os.LookupEnv(k); !ok || got != v {
				_ = os.Setenv(k, v)
			}
		}
		envMu.Unlock()
	})

	for k := range originals {
		if strings.HasPrefix(k, argusEnvPrefix) && !strings.HasPrefix(k, argusTestEnvPrefix) {
			_ = os.Unsetenv(k)
		}
	}
	for k, v := range vars {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("os.Setenv(%q) failed: %v", k, err)
		}
	}
}

// SplitEnvVars maps each KEY=value entry to its value, skipping any without an "=".
func SplitEnvVars(envVars []string) map[string]string {
	result := make(map[string]string, len(envVars))

	for _, kv := range envVars {
		key, value, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		result[key] = value
	}

	return result
}
