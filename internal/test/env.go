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
	"sync"
	"testing"
)

var envMu sync.Mutex // Only one test should modify environment variables at a time.

// SetEnv sets the specified environment variables for the duration of the test,
// and restores the original values afterwards.
func SetEnv(t *testing.T, vars map[string]string) {
	t.Helper()

	envMu.Lock()
	originals := make(map[string]*string)

	for k, v := range vars {
		if old, ok := os.LookupEnv(k); ok {
			// Capture existing value.
			orig := old
			originals[k] = &orig
		} else {
			// Mark previously unset.
			originals[k] = nil
		}
		// Set new value.
		os.Setenv(k, v)
	}

	t.Cleanup(func() {
		for k, v := range originals {
			if v == nil {
				// Clear previously unset.
				_ = os.Unsetenv(k)
			} else {
				// Restore previous value.
				_ = os.Setenv(k, *v)
			}
		}
		envMu.Unlock()
	})
}
