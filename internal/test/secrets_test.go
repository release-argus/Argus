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
	"fmt"
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	// GIVEN: an env var to fetch.
	tests := []struct {
		name     string
		envKey   string
		envValue *string
		panic    bool
	}{
		{
			name:     "env var set",
			envKey:   "ARGUS_TEST_GET_VALUE",
			envValue: Ptr("hello"),
			panic:    false,
		},
		{
			name:     "empty env var set",
			envKey:   "ARGUS_TEST_GET_EMPTY",
			envValue: Ptr(""),
			panic:    true,
		},
		{
			name:     "env var not set",
			envKey:   "ARGUS_TEST_GET_UNSET",
			envValue: nil,
			panic:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: that env var is set.
			if tc.envValue != nil {
				os.Setenv(tc.envKey, *tc.envValue)
				t.Cleanup(func() { os.Unsetenv(tc.envKey) })
			}

			prefix := fmt.Sprintf(
				"%s\nget(%q)",
				packageName, tc.envKey,
			)

			defer func() {
				if recover() != nil && !tc.panic {
					t.Fatalf(
						"%s unexpected panic when %q env var set to %q",
						prefix, tc.envKey, *tc.envValue,
					)
				}
			}()

			// WHEN: get is called on it.
			got := get(t, tc.envKey)

			// THEN: it doesn't reach this if a panic is expected.
			if tc.panic {
				t.Fatalf(
					"%s expected panic when %q env var not set",
					prefix, tc.envKey,
				)
			}

			// AND: the expected value is returned.
			if got != *tc.envValue {
				t.Fatalf(
					"%s mismatch\ngot:  %q\nwant:  %q",
					prefix, got, *tc.envValue,
				)
			}
		})
	}
}

func TestShoutrrrGotifyToken(t *testing.T) {
	// GIVEN: the environment variable ARGUS_TEST_GOTIFY_TOKEN.
	tests := []struct {
		name string
		env  string
	}{
		{name: "env var empty", env: ""},
		{name: "env var set", env: "test"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're manipulating the environment.

			want := tc.env
			if tc.env != "" {
				env := map[string]string{"ARGUS_TEST_GOTIFY_TOKEN": tc.env}
				SetEnv(t, env)
			}

			// WHEN: ShoutrrrGotifyToken is called.
			token := ShoutrrrGotifyToken()

			// THEN: the token should be as expected.
			if tc.env == "" {
				want = token // default token when env is empty.
			}
			if token != want {
				t.Errorf(
					"%s\nShoutrrrGotifyToken() mismatch\ngot:  %q\nwant: %q",
					packageName, token, want,
				)
			}
		})
	}
}
