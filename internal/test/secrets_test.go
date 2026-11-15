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
			wantSkipf: fmt.Sprintf("%q is not set", key),
		},
		{
			name:      "unset/skips when secrets are not required",
			require:   "false",
			wantSkipf: fmt.Sprintf("%q is not set", key),
		},
		{
			name:       "unset/fails when secrets are required",
			require:    "true",
			wantFatalf: fmt.Sprintf("%q is not set, but %q is true", key, requireSecretsEnv),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars.

			IsolateEnv(t, map[string]string{
				key:               tc.value,
				requireSecretsEnv: tc.require,
			})
			fakeT := &FakeT{Abort: true}

			var got string
			aborted := fakeT.Aborted(func() {
				got = requireSecret(fakeT, key)
			})

			for _, check := range []struct {
				name, got, want string
			}{
				{name: "value", got: got, want: tc.want},
				{name: "FakeT.Fatals", got: strings.Join(fakeT.Fatals, "\n"), want: tc.wantFatalf},
				{name: "FakeT.Skips", got: strings.Join(fakeT.Skips, "\n"), want: tc.wantSkipf},
			} {
				if check.got != check.want {
					t.Errorf("%s mismatch\ngot:  %q\nwant: %q", check.name, check.got, check.want)
				}
			}

			wantAborted := tc.wantFatalf != "" || tc.wantSkipf != ""
			if aborted != wantAborted {
				t.Errorf("FakeT.Aborted() mismatch\ngot:  %t\nwant: %t", aborted, wantAborted)
			}
		})
	}
}

func TestSkipUnlessRequired(t *testing.T) {
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
			// Serial: sets env vars.

			IsolateEnv(t, map[string]string{
				requireSecretsEnv: tc.require,
			})
			fakeT := &FakeT{Abort: true}

			aborted := fakeT.Aborted(func() {
				SkipUnlessRequired(fakeT, "%s is unreachable", "arg_here")
			})

			for _, check := range []struct {
				name, got, want string
			}{
				{name: "FakeT.Fatals", got: strings.Join(fakeT.Fatals, "\n"), want: tc.wantFatalf},
				{name: "FakeT.Skips", got: strings.Join(fakeT.Skips, "\n"), want: tc.wantSkipf},
			} {
				if check.got != check.want {
					t.Errorf("%s mismatch\ngot:  %q\nwant: %q", check.name, check.got, check.want)
				}
			}

			if !aborted {
				t.Error("FakeT.Aborted() mismatch\ngot:  false\nwant: true")
			}
		})
	}
}

func TestShoutrrrGotifyToken(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{name: "env var empty", env: ""},
		{name: "env var set", env: "test"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars.

			IsolateEnv(t, map[string]string{"ARGUS_TEST_GOTIFY_TOKEN": tc.env})

			token := ShoutrrrGotifyToken()

			if tc.env == "" {
				if token == "" {
					t.Error("token mismatch\ngot:  \"\"\nwant: NOT \"\" (default)")
				}
				return
			}
			if token != tc.env {
				t.Errorf("token mismatch\ngot:  %q\nwant: %q", token, tc.env)
			}
		})
	}
}

func TestSecretGetters(t *testing.T) {
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
			// Serial: sets env vars.

			IsolateEnv(t, map[string]string{tc.key: tc.value})

			got := tc.get(t)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
