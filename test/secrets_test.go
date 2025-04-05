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

package test

import (
	"os"
	"testing"
)

func TestShoutrrrGotifyToken(t *testing.T) {
	// GIVEN the environment variable ARGUS_TEST_GOTIFY_TOKEN.
	tests := map[string]struct {
		env string
	}{
		"empty": {env: ""},
		"set":   {env: "test"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're manipulating the environment.

			want := tc.env
			if tc.env != "" {
				os.Setenv("ARGUS_TEST_GOTIFY_TOKEN", tc.env)
				t.Cleanup(func() { os.Unsetenv("ARGUS_TEST_GOTIFY_TOKEN") })
			}

			// WHEN ShoutrrrGotifyToken is called.
			token := ShoutrrrGotifyToken()

			// THEN the token should be as expected.
			if tc.env == "" {
				want = token // default token when env is empty.
			}
			if token != want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, want, token)
			}
		})
	}
}
