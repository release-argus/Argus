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
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

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
			wantSkipf: key + " is not set",
		},
		{
			name:      "unset/skips when secrets are not required",
			require:   "false",
			wantSkipf: key + " is not set",
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
			fT := &FakeT{Abort: true}

			// WHEN: requireSecret is called on it.
			var got string
			aborted := fT.Aborted(func() { got = requireSecret(fT, key) })

			// THEN: the value is returned, or the test is skipped/failed.
			for _, check := range []struct {
				name, got, want string
			}{
				{name: "value", got: got, want: tc.want},
				{name: "Fatalf", got: strings.Join(fT.Fatals, "\n"), want: tc.wantFatalf},
				{name: "Skipf", got: strings.Join(fT.Skips, "\n"), want: tc.wantSkipf},
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

func TestSkipUnlessRequired(t *testing.T) {
	// GIVEN: a reason to skip, and the flag that makes a missing secret fatal.
	tests := []struct {
		name       string
		require    string
		wantFatalf string
		wantSkipf  string
	}{
		{
			name:      "not required/skips when 'require secrets' is unset",
			wantSkipf: "arg_here is unreachable",
		},
		{
			name:      "not required/skips when secrets are not required",
			require:   "false",
			wantSkipf: "arg_here is unreachable",
		},
		{
			name:       "required/fails",
			require:    "true",
			wantFatalf: "arg_here is unreachable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're manipulating the environment.

			// AND: the env is set.
			SetEnv(t, map[string]string{
				requireSecretsEnv: tc.require,
			})
			fT := &FakeT{Abort: true}

			// WHEN: SkipUnlessRequired is called.
			aborted := fT.Aborted(func() {
				SkipUnlessRequired(fT, "%s is unreachable", "arg_here")
			})

			// THEN: the test is skipped, or failed.
			for _, check := range []struct {
				name, got, want string
			}{
				{name: "Fatalf", got: strings.Join(fT.Fatals, "\n"), want: tc.wantFatalf},
				{name: "Skipf", got: strings.Join(fT.Skips, "\n"), want: tc.wantSkipf},
			} {
				if check.got != check.want {
					t.Errorf(
						"%s\nSkipUnlessRequired() %s mismatch\ngot:  %q\nwant: %q",
						packageName, check.name,
						check.got, check.want,
					)
				}
			}

			// AND: reporting ends the test.
			if !aborted {
				t.Errorf(
					"%s\nSkipUnlessRequired() aborted mismatch\ngot:  %t\nwant: %t",
					packageName,
					aborted, true,
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

func TestSecretGetters(t *testing.T) {
	// GIVEN: the env vars that each getter reads.
	const gitHubToken = "dummy-github-token"
	tests := []struct {
		name  string
		key   string
		value string
		get   func(*testing.T) string
		want  string
	}{
		{
			name:  "DockerHubUsername",
			key:   "DOCKER_HUB_USERNAME",
			value: "dummy-docker-hub-username",
			get:   DockerHubUsername,
			want:  "dummy-docker-hub-username",
		},
		{
			name:  "DockerHubToken",
			key:   "DOCKER_HUB_TOKEN",
			value: "dummy-docker-hub-token",
			get:   DockerHubToken,
			want:  "dummy-docker-hub-token",
		},
		{
			name:  "DockerQuayToken",
			key:   "DOCKER_QUAY_TOKEN",
			value: "dummy-quay-token",
			get:   DockerQuayToken,
			want:  "dummy-quay-token",
		},
		{
			name:  "ForgejoToken",
			key:   "ARGUS_TEST_FORGEJO_TOKEN",
			value: "dummy-forgejo-token",
			get:   ForgejoToken,
			want:  "dummy-forgejo-token",
		},
		{
			name:  "GitHubToken",
			key:   "GITHUB_TOKEN",
			value: gitHubToken,
			get:   GitHubToken,
			want:  gitHubToken,
		},
		{
			name:  "GitHubTokenEncoded",
			key:   "GITHUB_TOKEN",
			value: gitHubToken,
			get:   GitHubTokenEncoded,
			want:  base64.StdEncoding.EncodeToString([]byte(gitHubToken)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're manipulating the environment.

			// AND: the env var holds a value.
			SetEnv(t, map[string]string{tc.key: tc.value})

			// WHEN: the getter is called.
			got := tc.get(t)

			// THEN: the value that env var holds is returned.
			if got != tc.want {
				t.Errorf(
					"%s\n%s() mismatch\ngot:  %q\nwant: %q",
					packageName, tc.name,
					got, tc.want,
				)
			}
		})
	}
}
