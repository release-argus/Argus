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
	"testing"
)

// abortSecretT stands in for the [runtime.Goexit] that t.Fatalf/t.Skipf perform.
type abortSecretT struct{}

// fakeSecretT records what requireSecret reports, aborting as [testing.T] does.
type fakeSecretT struct {
	fatalf string
	skipf  string
}

func (f *fakeSecretT) Helper() {}

func (f *fakeSecretT) Fatalf(format string, args ...any) {
	f.fatalf = fmt.Sprintf(format, args...)
	panic(abortSecretT{})
}

func (f *fakeSecretT) Skipf(format string, args ...any) {
	f.skipf = fmt.Sprintf(format, args...)
	panic(abortSecretT{})
}

// callRequireSecret calls [requireSecret], recovering the abort that a report triggers.
func callRequireSecret(t *fakeSecretT, key string) (value string, aborted bool) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if _, ok := r.(abortSecretT); !ok {
			panic(r)
		}
		aborted = true
	}()

	return requireSecret(t, key), false
}

func TestRequireSecret(t *testing.T) {
	// GIVEN: a secret env var, and the flag that makes a missing one fatal.
	key := "ARGUS_TEST_REQUIRE_SECRET_VALUE"
	tests := []struct {
		name       string
		value      string
		require    string
		want       string
		wantFatalf string
		wantSkipf  string
	}{
		{
			name:  "set/returns the value",
			value: "hello",
			want:  "hello",
		},
		{
			name:    "set/returns the value, even when secrets are required",
			value:   "hello",
			require: "true",
			want:    "hello",
		},
		{
			name:      "unset/skips",
			wantSkipf: fmt.Sprintf("%s is not set", key),
		},
		{
			name:      "unset/skips when secrets are not required",
			require:   "false",
			wantSkipf: fmt.Sprintf("%s is not set", key),
		},
		{
			name:    "unset/fails when secrets are required",
			require: "true",
			wantFatalf: fmt.Sprintf(
				"%s is not set, but %s is true",
				key, requireSecretsEnv,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're manipulating the environment.

			// AND: the env is set.
			SetEnv(t, map[string]string{
				key:               tc.value,
				requireSecretsEnv: tc.require,
			})
			fake := &fakeSecretT{}

			// WHEN: requireSecret is called on it.
			got, aborted := callRequireSecret(fake, key)

			// THEN: the value is returned, or the test is skipped/failed.
			for _, check := range []struct {
				name, got, want string
			}{
				{name: "value", got: got, want: tc.want},
				{name: "Fatalf", got: fake.fatalf, want: tc.wantFatalf},
				{name: "Skipf", got: fake.skipf, want: tc.wantSkipf},
			} {
				if check.got != check.want {
					t.Errorf(
						"%s\nrequireSecret(%q) %s mismatch\ngot:  %q\nwant: %q",
						packageName, key, check.name,
						check.got, check.want,
					)
				}
			}

			// AND: reporting ends the test.
			wantAborted := tc.wantFatalf != "" || tc.wantSkipf != ""
			if aborted != wantAborted {
				t.Errorf(
					"%s\nrequireSecret(%q) aborted mismatch\ngot:  %t\nwant: %t",
					packageName, key,
					aborted, wantAborted,
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
