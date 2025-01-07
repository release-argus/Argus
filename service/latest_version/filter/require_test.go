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

package filter

import (
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

func TestRequireDefaults_Default(t *testing.T) {
	// GIVEN a RequireDefaults
	tests := map[string]struct {
		require RequireDefaults
	}{
		"empty DockerCheckDefaults": {
			require: RequireDefaults{
				Docker: DockerCheckDefaults{}}},
		"non-empty DockerCheckDefaults": {
			require: RequireDefaults{
				Docker: DockerCheckDefaults{
					Type: "ghcr"}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Default is called on it
			tc.require.Default()

			// THEN the DockerCheckDefaults is set to its default values
			defaultType := "hub"
			if tc.require.Docker.Type != defaultType {
				t.Errorf("filter.Require.Default() mismatch on Docker.Type:\nwant: %q\ngot:  %q",
					defaultType, tc.require.Docker.Type)
			}
		})
	}
}

func TestRequire_Init(t *testing.T) {
	// GIVEN a Require, JLog and a Status
	tests := map[string]struct {
		req             *Require
		wantDockerCheck bool
	}{
		"nil require": {
			req: nil},
		"non-nil require": {
			req: &Require{}},
		"non-nil require with empty DockerCheck": {
			req: &Require{
				Docker: &DockerCheck{}}},
		"non-nil require with non-empty DockerCheck": {
			req: &Require{
				Docker: &DockerCheck{
					Image: "foo",
					Tag:   "bar"}},
			wantDockerCheck: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			status := status.Status{}
			status.Init(
				0, 0, 0,
				test.StringPtr("test"), nil,
				test.StringPtr("http://example.com"))
			status.SetDeployedVersion("1.2.3", "", false)
			defaults := RequireDefaults{
				Docker: *NewDockerCheckDefaults(
					"ghcr",
					"",
					"foo", "",
					"",
					nil)}

			// WHEN Init is called with it
			tc.req.Init(&status, &defaults)

			// THEN the global JLog is set to its address
			if tc.req == nil {
				// THEN the Require is still nil
				if tc.req != nil {
					t.Fatal("Init with a nil require shouldn't initialise it")
				}
			} else {
				// THEN the status is given to the Require
				if tc.req.Status != &status {
					t.Fatalf("Status should be the address of the var given to it %v, not %v",
						&status, tc.req.Status)
				}
				// AND the DockerCheck remains nil if it was initially
				if !tc.wantDockerCheck {
					if tc.req.Docker != nil {
						t.Fatal("Init with a nil DockerCheck shouldn't initialise it")
					}
					return
				}
				// AND the defaults are handed to it otherwise
				if tc.req.Docker.Defaults != &defaults.Docker {
					t.Fatalf("Docker defaults should be the address of the var given to it %v, not %v",
						&defaults.Docker, tc.req.Docker.Defaults)
				}
			}
		})
	}
}

func TestRequireDefaults_CheckValues(t *testing.T) {
	// GIVEN a RequireDefaults
	tests := map[string]struct {
		docker   DockerCheckDefaults
		errRegex string
	}{
		"valid": {
			docker: *NewDockerCheckDefaults(
				"ghcr", "", "", "", "", nil),
			errRegex: `^$`,
		},
		"invalid docker": {
			docker: *NewDockerCheckDefaults(
				"foo", "", "", "", "", nil),
			errRegex: test.TrimYAML(`
				^docker:
					type: .* <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			require := RequireDefaults{
				Docker: tc.docker}

			// WHEN CheckValues is called on it
			err := require.CheckValues("")

			// THEN err is expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("RequireDefaults.CheckValues() want %d lines of error:\n%q\ngot %d lines:\n%v\nstdout: %q",
					wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("RequireDefaults.CheckValues() error mismatch\nwant match for:\n%q\ngot:\n%q",
					tc.errRegex, e)
				return
			}
		})
	}
}

func TestRequire_CheckValues(t *testing.T) {
	// GIVEN a Require
	tests := map[string]struct {
		require  *Require
		errRegex string
	}{
		"nil": {
			require:  nil,
			errRegex: `^$`,
		},
		"valid regex_content regex": {
			require: &Require{
				RegexContent: "[0-9]"},
			errRegex: `^$`,
		},
		"invalid regex_content regex": {
			require: &Require{
				RegexContent: "[0-"},
			errRegex: `^regex_content: .* <invalid>.*RegEx.*$`,
		},
		"valid regex_content template": {
			require: &Require{
				RegexContent: `{% if version %}.linux-amd64{% endif %}`},
			errRegex: `^$`,
		},
		"invalid regex_content template": {
			require: &Require{
				RegexContent: "{% if version }.linux-amd64"},
			errRegex: `^regex_content: .* <invalid>.*templating`,
		},
		"valid regex_version": {
			require: &Require{
				RegexVersion: "[0-9]"},
			errRegex: `^$`,
		},
		"invalid regex_version": {
			require: &Require{
				RegexVersion: "[0-"},
			errRegex: `^regex_version: .* <invalid>.*$`,
		},
		"valid command": {
			require: &Require{
				Command: []string{
					"bash", "update.sh", "{{ version }}"}},
			errRegex: `^$`,
		},
		"invalid command": {
			require: &Require{
				Command: []string{"{{ version }"}},
			errRegex: `^command: .* <invalid>.*templating.*$`,
		},
		"valid docker": {
			require: &Require{
				Docker: NewDockerCheck(
					"ghcr",
					"release-argus/Argus",
					"{{ version }}",
					"", "", "", time.Now(), nil)},
			errRegex: `^$`,
		},
		"invalid docker": {
			require: &Require{
				Docker: NewDockerCheck(
					"foo",
					"", "", "", "", "", time.Now(), nil)},
			errRegex: test.TrimYAML(`
				^docker:
					type: "foo" <invalid>.*
					image: <required>.*
					tag: <required>.*$`),
		},
		"all possible errors": {
			require: &Require{
				RegexContent: "[0-",
				RegexVersion: "[0-",
				Docker: NewDockerCheck(
					"foo",
					"", "", "", "", "", time.Now(), nil)},
			errRegex: test.TrimYAML(`
				^regex_content: .* <invalid>.*
				regex_version: .* <invalid>.*
				docker:
					type: .* <invalid>.*
					image: <required>.*
					tag: <required>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN CheckValues is called on it
			err := tc.require.CheckValues("")

			// THEN err is expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("RequireDefaults.CheckValues() want %d lines of error:\n%q\ngot %d lines:\n%v\nstdout: %q",
					wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("RequireDefaults.CheckValues() error mismatch\nwant match for:\n%q\ngot:\n%q",
					tc.errRegex, e)
				return
			}
		})
	}
}

func TestRequire__String(t *testing.T) {
	tests := map[string]struct {
		require *Require
		want    string
	}{
		"nil": {
			require: nil,
			want:    ""},
		"empty": {
			require: &Require{},
			want:    "{}\n"},
		"all fields defined": {
			require: &Require{
				Status:       &status.Status{},
				RegexContent: "abc{{ version }}.tar.gz",
				RegexVersion: "v([0-9.]+)",
				Command:      command.Command{"ls", "-la"},
				Docker: NewDockerCheck(
					"hub",
					"", "", "", "", "", time.Now(), nil)},
			want: test.TrimYAML(`
				regex_content: abc{{ version }}.tar.gz
				regex_version: v([0-9.]+)
				command:
					- ls
					- -la
				docker:
					type: hub
			`)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Require is stringified with String
			got := tc.require.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestRequire_Inherit(t *testing.T) {
	type overrides struct {
		overrides string
		nil       bool
	}
	// GIVEN two Require objects
	tests := map[string]struct {
		from               overrides
		to                 overrides
		inheritDockerToken bool
	}{
		"nil to": {
			to: overrides{
				nil: true},
			inheritDockerToken: false,
		},
		"nil from": {
			from: overrides{
				nil: true},
			inheritDockerToken: false,
		},
		"no Docker to": {
			to: overrides{
				overrides: test.TrimYAML(`
					docker: null
				`)},
			inheritDockerToken: false,
		},
		"no Docker from": {
			from: overrides{
				overrides: test.TrimYAML(`
					docker: null
				`)},
			inheritDockerToken: false,
		},
		"no change Docker": {
			inheritDockerToken: true,
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						token: ` + util.SecretValue)},
		},
		"change of Type - no copy": {
			from: overrides{
				overrides: test.TrimYAML(`
					docker:
						type: ghcr
				`)},
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						type: hub
				`)},
			inheritDockerToken: false,
		},
		"change of Image - no copy": {
			from: overrides{
				overrides: test.TrimYAML(`
					docker:
						image: release-argus/argus
				`)},
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						image: release-argus/test
				`)},
			inheritDockerToken: false,
		},
		"change of Username - no copy": {
			from: overrides{
				overrides: test.TrimYAML(`
					docker:
						username: foo
				`)},
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						username: bar
				`)},
			inheritDockerToken: false,
		},
		"change of Token - no copy": {
			from: overrides{
				overrides: test.TrimYAML(`
					docker:
						token: foo
				`)},
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						token: bar
				`)},
			inheritDockerToken: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var from *Require
			if !tc.from.nil {
				from = &Require{
					Docker: NewDockerCheck(
						"ghcr",
						"release-argus/argus", "{{ version }}",
						"ghcr-username", "ghcr-token",
						"ghcr-query-token", time.Now(),
						nil)}
				err := yaml.Unmarshal([]byte(tc.from.overrides), from)
				if err != nil {
					t.Fatalf("error unmarshalling overrides: %v",
						err)
				}
			}
			var to *Require
			wantToken, wantQueryToken, wantValidUntil := "", "", ""
			if !tc.to.nil {
				to = &Require{
					Docker: NewDockerCheck(
						"ghcr",
						"release-argus/argus", "{{ version }}",
						"ghcr-username", "",
						"", time.Now(),
						nil)}
				err := yaml.Unmarshal([]byte(tc.to.overrides), to)
				if err != nil {
					t.Fatalf("error unmarshalling overrides: %v",
						err)
				}
				if to.Docker != nil {
					wantToken = to.Docker.Token
					wantValidUntil = to.Docker.validUntil.String()
				}
			}
			if tc.inheritDockerToken &&
				to != nil && to.Docker != nil &&
				from != nil && from.Docker != nil {
				wantToken = from.Docker.Token
				wantQueryToken = from.Docker.queryToken
				wantValidUntil = from.Docker.validUntil.String()
			}

			// WHEN Inherit is called on them
			to.Inherit(from)

			// THEN the Require tokens are inherited
			gotToken, gotQueryToken, gotValidUntil := "", "", ""
			if to != nil && to.Docker != nil {
				gotToken = to.Docker.Token
				gotQueryToken = to.Docker.queryToken
				gotValidUntil = to.Docker.validUntil.String()
			}
			if gotToken != wantToken {
				t.Errorf("filter.Require.Inherit() Token mismatch:\nwant: %q\ngot:  %q",
					wantToken, gotToken)
			}
			if gotQueryToken != wantQueryToken {
				t.Errorf("filter.Require.Inherit() QueryToken mismatch:\nwant: %q\ngot:  %q",
					wantQueryToken, gotQueryToken)
			}
			if gotValidUntil != wantValidUntil {
				t.Errorf("filter.Require.Inherit()  ValidUntil mismatch:\nwant: %q\ngot:  %q",
					wantValidUntil, gotValidUntil)
			}
		})
	}
}

func TestRequire_removeUnusedRequireDocker(t *testing.T) {
	tests := map[string]struct {
		require *Require
		nil     bool
	}{
		"nil require": {
			require: nil,
			nil:     true,
		},
		"nil Docker": {
			require: &Require{
				Docker: nil},
			nil: true,
		},
		"Docker with Image and Tag - kept": {
			require: &Require{
				Docker: &DockerCheck{
					Image: "release-argus/argus",
					Tag:   "latest"}},
			nil: false,
		},
		"Docker with Image only": {
			require: &Require{
				Docker: &DockerCheck{
					Image: "release-argus/argus"}},
			nil: true,
		},
		"Docker with Tag only": {
			require: &Require{
				Docker: &DockerCheck{
					Tag: "latest"}},
			nil: true,
		},
		"Docker with empty Image and Tag": {
			require: &Require{
				Docker: &DockerCheck{
					Image: "",
					Tag:   ""}},
			nil: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN removeUnusedRequireDocker is called
			tc.require.removeUnusedRequireDocker()

			// THEN the Docker is removed if it has no Image or Tag
			if tc.nil != (tc.require == nil || tc.require.Docker == nil) {
				t.Errorf("Docker:\nwant nil=%t\ngot  nil=%t",
					tc.nil, tc.require == nil || tc.require.Docker == nil)
			}
		})
	}
}
