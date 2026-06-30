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

package filter

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	dockertest "github.com/release-argus/Argus/service/latest_version/filter/docker/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestRequire_IsZero(t *testing.T) {
	// GIVEN: a Require.
	tests := []struct {
		name string
		req  *Require
		want bool
	}{
		{
			name: "nil",
			req:  nil,
			want: true,
		},
		{
			name: "empty",
			req:  &Require{},
			want: true,
		},
		{
			name: "non-empty/RegexContent",
			req: &Require{
				RegexContent: "abc",
			},
			want: false,
		},
		{
			name: "non-empty/RegexVersion",
			req: &Require{
				RegexVersion: "abc",
			},
			want: false,
		},
		{
			name: "non-empty/Command",
			req: &Require{
				Command: command.Command{"ls", "-lah"},
			},
			want: false,
		},
		{
			name: "non-empty/Docker/.Tag from defaults, no .Image",
			req: test.Must(t, func() (*Require, error) {
				data := []byte(test.TrimYAML(`
					docker:
						type: hub
						tag: foo
				`))
				defaults, _ := DecodeDefaults("yaml", data)
				hardDefaults, _ := DecodeDefaults("yaml", data)
				defaults.SetDefaults(hardDefaults)
				svcStatus, _ := statustest.New("yaml", nil)

				req, err := Decode(
					"yaml", []byte(test.TrimYAML(`
						docker: {}
					`)),
					svcStatus,
					defaults,
				)
				req.Docker.(*docker.HubRegistry).Image = ""

				if req.Docker.GetTag() == "" {
					t.Fatal("Docker.Tag should not be empty")
				}
				return req, err
			}),
			want: true,
		},
		{
			name: "non-empty/Docker/.Tag from defaults, .Image set",
			req: test.Must(t, func() (*Require, error) {
				data := []byte(test.TrimYAML(`
					docker:
						type: hub
						tag: foo
				`))
				defaults, _ := DecodeDefaults("yaml", data)
				hardDefaults, _ := DecodeDefaults("yaml", data)
				defaults.SetDefaults(hardDefaults)
				svcStatus, _ := statustest.New("yaml", nil)

				req, err := Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							image: foo
					`)),
					svcStatus,
					defaults,
				)

				return req, err
			}),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called.
			got := tc.req.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nRequire.IsZero() value mismatch\ngot:  %v\nwant: %v",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestRequire_String(t *testing.T) {
	defaults, _ := plainDefaults(t)

	tests := []struct {
		name    string
		require *Require
		want    string
	}{
		{
			name:    "nil",
			require: nil,
			want:    "",
		},
		{
			name:    "empty",
			require: &Require{},
			want:    "{}\n",
		},
		{
			name: "filled",
			require: test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						regex_content: abc{{ version }}.tar.gz
						regex_version: v([0-9.]+)
						command: ["ls", "-la"]
						docker:
							type: hub
					`)),
					svcStatus,
					defaults,
				)
			}),
			want: test.TrimYAML(`
				regex_content: abc{{ version }}.tar.gz
				regex_version: v([0-9.]+)
				command:
					- ls
					- -la
				docker:
					type: hub
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.require.String,
				tc.want,
			)
		})
	}
}

func TestRequire_Unmarshal(t *testing.T) {
	// GIVEN: JSON and/or YAML string to unmarshal into Require.
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			want:     "{}\n",
			errRegex: "^$",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: "^$",
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"regex_content": "a",
				"regex_version": "b",
				"command": ["ls", "-lah"],
				"docker": {
					"type": "ghcr",
					"image": "i",
					"tag":   "t",
					"auth": {
						"token": "abc"
					}
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				regex_content: a
				regex_version: b
				command:
					- ls
					- -lah
			`),
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				regex_content: a
				regex_version: b
				command: ["ls", "-lah"]
				docker:
					type: ghcr
					image: i
					tag: t
					auth:
						token: abc
			`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				regex_content: a
				regex_version: b
				command:
					- ls
					- -lah
			`),
		},
		{
			name:     "YAML/invalid data types",
			format:   "yaml",
			data:     "regex_content: [a]\n",
			errRegex: `^[^\s]+ .*unmarshal.*`,
		},
		{
			name:   "YAML/invalid docker subtree, ignored",
			format: "yaml",
			data: test.TrimYAML(`
				regex_content: a
				regex_version: b
				command: ["ls", "-lah"]
				docker:
					- type: ghcr
						image: i
						tag: t
						auth:
						token: abc
			`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				regex_content: a
				regex_version: b
				command:
					- ls
					- -lah
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Require, error) {
					var zero Require
					err := decode.Unmarshal(format, data, &zero)
					return &zero, err
				},
				tc.format, tc.data,
				func(val *Require) string { return val.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"Require",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestRequire_Copy(t *testing.T) {
	// GIVEN: a Require.
	tests := []struct {
		name string
		req  *Require
	}{
		{
			name: "nil",
			req:  nil,
		},
		{
			name: "command",
			req: &Require{
				Command: command.Command{"ls", "-lah"},
			},
		},
		{
			name: "docker",
			req: &Require{
				Docker: test.Must(t, func() (docker.Registry, error) {
					defaults := docker.Defaults{}
					defaults.Default()
					return docker.Decode(
						"yaml", []byte(test.TrimYAML(`
							image: foo
							tag: bar
						`)),
						&defaults,
					)
				}),
			},
		},
		{
			name: "filled",
			req: &Require{
				RegexContent: "rc",
				RegexVersion: "rv",
				Command:      command.Command{"ls", "-lah"},
				Docker: test.Must(t, func() (docker.Registry, error) {
					defaults := docker.Defaults{}
					defaults.Default()
					return docker.Decode(
						"yaml", []byte(test.TrimYAML(`
							image: foo
							tag: bar
						`)),
						&defaults,
					)
				}),
				defaults: &RequireDefaults{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, nilStatus := range []bool{true, false} {

				// AND: a Status to Copy.
				var newStatus *status.Status
				if !nilStatus {
					newStatus = &status.Status{}
				}

				// WHEN: Copy() is called on it with this Status.
				got := tc.req.Copy(newStatus)

				prefix := fmt.Sprintf(
					"%s\nRequire.Copy(%p)",
					packageName, newStatus,
				)

				// THEN: if nil was copied, we got nil
				if got == nil {
					if !nilStatus {
						t.Errorf(
							"%s of nil got %v, want nil",
							prefix, got,
						)
					}
					return
				}

				// AND: the fields are copied/shared as expected.
				fieldTests := []test.FieldAssertion{
					{Name: "Command", Got: &got.Command, Want: &tc.req.Command, Mode: test.CompareDifferentPointer},
					{Name: "Status", Got: got.Status, Want: newStatus, Mode: test.CompareEqual},
					{Name: "Defaults", Got: got.defaults, Want: tc.req.defaults, Mode: test.CompareSamePointer},
				}
				if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestRequire_Init(t *testing.T) {
	defaults, _ := plainDefaults(t)

	// GIVEN: a Require, JLog and a Status.
	tests := []struct {
		name            string
		req             *Require
		wantDockerCheck bool
	}{
		{
			name: "nil require",
			req:  nil,
		},
		{
			name: "non-nil require",
			req:  &Require{},
		},
		{
			name: "non-nil require with empty DockerCheck",
			req: test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				req, err := Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: hub
							image: foo,
							tag: bar
					`)),
					svcStatus,
					defaults,
				)

				req.Docker = docker.RegistryMap["ghcr"]()
				return req, err
			}),
		},
		{
			name: "non-nil require with non-empty DockerCheck",
			req: test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: hub
							image: foo,
							tag: bar
					`)),
					svcStatus,
					defaults,
				)
			}),
			wantDockerCheck: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			svcStatus := status.Status{}
			svcDashboard := &dashboard.Options{
				OptionsBase: dashboard.OptionsBase{
					WebURL: "https://example.com",
				},
			}
			svcStatus.Init(
				0, 0, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				svcDashboard,
			)
			svcStatus.SetDeployedVersion("1.2.3", "", false)

			// WHEN: Init is called with it.
			tc.req.Init(&svcStatus, defaults)

			prefix := fmt.Sprintf("%s\nRequire.Init()", packageName)

			// THEN: the global JLog is set to its address.
			if tc.req == nil {
				// THEN: the Require is still nil.
				if tc.req != nil {
					t.Fatalf("%s with a nil require shouldn't initialise it", prefix)
				}
			} else {
				// THEN: the status is given to the Require.
				if tc.req.Status != &svcStatus {
					t.Fatalf(
						"%s .Status should be the address of the var given to it\ngot:  %v\nwant: %v",
						prefix, &svcStatus, tc.req.Status,
					)
				}

				// AND: the DockerCheck remains nil if it was initially.
				if !tc.wantDockerCheck {
					if tc.req.Docker != nil {
						t.Fatalf("%s with a nil DockerCheck shouldn't initialise it", prefix)
					}
					return
				}

				// AND: the defaults are handed to it otherwise.
				expectedDefaults, _ := dockertest.GetDefaultOfDockerType(
					t,
					tc.req.Docker.GetType(),
					&defaults.Docker,
				)
				if got := tc.req.Docker.Defaults(); got != expectedDefaults {
					t.Fatalf(
						"%s .Docker.Defaults() should be the address of the var given to it\ngot:  %v\nwant: %v",
						prefix, got, expectedDefaults,
					)
				}
			}
		})
	}
}

func TestRequire_CheckValues(t *testing.T) {
	defaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults, _ := DecodeDefaults("yaml", nil)
	defaults.SetDefaults(hardDefaults)

	// GIVEN: a Require.
	tests := []struct {
		name     string
		input    *Require
		wantYAML string
		errRegex string
	}{
		{
			name:     "nil",
			input:    (*Require)(nil),
			errRegex: `^$`,
		},
		{
			name: "valid regex_content regex",
			input: &Require{
				RegexContent: "[0-9]",
			},
			wantYAML: "regex_content: '[0-9]'\n",
			errRegex: `^$`,
		},
		{
			name: "invalid regex_content regex",
			input: &Require{
				RegexContent: "[0-",
			},
			wantYAML: "regex_content: '[0-'\n",
			errRegex: `^regex_content: .* <invalid>.*RegEx.*$`,
		},
		{
			name: "valid regex_content template",
			input: &Require{
				RegexContent: `{% if version %}.linux-amd64{% endif %}`,
			},
			wantYAML: "regex_content: '{% if version %}.linux-amd64{% endif %}'\n",
			errRegex: `^$`,
		},
		{
			name: "invalid regex_content template",
			input: &Require{
				RegexContent: "{% if version }.linux-amd64",
			},
			wantYAML: "regex_content: '{% if version }.linux-amd64'\n",
			errRegex: `^regex_content: .* <invalid>.*templating`,
		},
		{
			name: "valid regex_version",
			input: &Require{
				RegexVersion: "[0-9]",
			},
			wantYAML: "regex_version: '[0-9]'\n",
			errRegex: `^$`,
		},
		{
			name: "invalid regex_version",
			input: &Require{
				RegexVersion: "[0-",
			},
			wantYAML: "regex_version: '[0-'\n",
			errRegex: `^regex_version: .* <invalid>.*$`,
		},
		{
			name: "valid command",
			input: &Require{
				Command: []string{
					"bash", "update.sh", "{{ version }}"},
			},
			wantYAML: test.TrimYAML(`
				command:
					- bash
					- update.sh
					- '{{ version }}'
			`),
			errRegex: `^$`,
		},
		{
			name: "invalid command",
			input: &Require{
				Command: []string{"{{ version }", "cmd", "1"},
			},
			wantYAML: test.TrimYAML(`
				command:
					- '{{ version }'
					- cmd
					- '1'
			`),
			errRegex: `^command: .* <invalid>.*templating.*$`,
		},
		{
			name: "valid docker",
			input: test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: `+test.ArgusGitHubRepo+`
							tag: '{{ version }}'
					`)),
					svcStatus,
					defaults,
				)
			}),
			wantYAML: test.TrimYAML(`
				docker:
					type: ghcr
					image: ` + test.ArgusGitHubRepo + `
					tag: '{{ version }}'
			`),
			errRegex: `^$`,
		},
		{
			name: "docker, no image:tag removes Docker",
			input: test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							auth:
								token: ghp_TEST
					`)),
					svcStatus,
					defaults,
				)
			}),
			wantYAML: "{}\n",
			errRegex: `^$`,
		},
		{
			name: "all possible errors",
			input: test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						regex_content: '[0-'
						regex_version: '[0-'
						docker:
							type: ghcr
							image: `+test.ArgusGitHubRepo+`
					`)),
					svcStatus,
					defaults,
				)
			}),
			wantYAML: test.TrimYAML(`
				regex_content: '[0-'
				regex_version: '[0-'
				docker:
					type: ghcr
					image: ` + test.ArgusGitHubRepo + `
			`),
			errRegex: test.TrimYAML(`
				^regex_content: .* <invalid>.*
				regex_version: .* <invalid>.*
				docker:
					tag: <required>.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)

			// THEN: it stringifies correctly.
			if gotYAML := tc.input.String(""); gotYAML != tc.wantYAML {
				t.Fatalf(
					"%s\npost Require.CheckValues() stringified mismatch\ngot:  %q\nwant: %q",
					packageName, gotYAML, tc.wantYAML,
				)
			}
		})
	}
}

func TestRequire_Inherit(t *testing.T) {
	defaults, _ := plainDefaults(t)

	type overrides struct {
		overrides string
		nil       bool
	}

	// GIVEN: two Require objects.
	tests := []struct {
		name               string
		from, to           overrides
		inheritDockerToken bool
	}{
		{
			name: "nil to",
			to: overrides{
				nil: true,
			},
			inheritDockerToken: false,
		},
		{
			name: "nil from",
			from: overrides{
				nil: true,
			},
			inheritDockerToken: false,
		},
		{
			name: "no Docker to",
			to: overrides{
				overrides: `docker: null`,
			},
			inheritDockerToken: false,
		},
		{
			name: "no Docker from",
			from: overrides{
				overrides: `docker: null`,
			},
			inheritDockerToken: false,
		},
		{
			name:               "no change Docker",
			inheritDockerToken: true,
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						token: ` + util.SecretValue,
				),
			},
		},
		{
			name: "change of Type, no copy",
			from: overrides{
				overrides: test.TrimYAML(`
					docker:
						type: ghcr
				`),
			},
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						type: hub
				`),
			},
			inheritDockerToken: false,
		},
		{
			name: "change of Image, no copy",
			from: overrides{
				overrides: test.TrimYAML(`
					docker:
						image: test/app
				`),
			},
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						image: test/app2
				`),
			},
			inheritDockerToken: false,
		},
		{
			name: "change of Username, no copy",
			from: overrides{
				overrides: test.TrimYAML(`
					docker:
						username: foo
				`),
			},
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						username: bar
				`),
			},
			inheritDockerToken: false,
		},
		{
			name: "change of Token, no copy",
			from: overrides{
				overrides: test.TrimYAML(`
					docker:
						token: foo
				`),
			},
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						token: bar
				`),
			},
			inheritDockerToken: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var from *Require
			if !tc.from.nil {
				svcStatus, _ := statustest.New("yaml", nil)
				from, _ = Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: `+test.ArgusGitHubRepo+`
							tag: '{{ version }}'
							username: ghcr-username
							password: ghcr-password
					`)),
					svcStatus,
					defaults,
				)
				var err error
				if from, err = from.ApplyOverrides(
					"yaml", []byte(tc.from.overrides),
					from.Status,
					defaults,
				); err != nil {
					t.Fatalf(
						"%s\nerror unmarshaling RequireDefaults overrides: %v",
						packageName, err,
					)
				}
			}
			var to *Require
			wantToken, wantQueryToken, wantValidUntil := "", "", time.Time{}
			if !tc.to.nil {
				svcStatus, _ := statustest.New("yaml", nil)
				to, _ = Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: `+test.ArgusGitHubRepo+`
							tag: '{{ version }}'
							username: ghcr-username
							password: ghcr-password
					`)),
					svcStatus,
					defaults,
				)
				var err error
				if to, err = from.ApplyOverrides(
					"yaml", []byte(tc.to.overrides),
					to.Status,
					defaults,
				); err != nil {
					t.Fatalf(
						"%s\nerror unmarshaling RequireDefaults overrides: %v",
						packageName, err,
					)
				}
				if to != nil && to.Docker != nil {
					wantToken = to.Docker.GetAuth().GetTokenSelf()
					_, wantValidUntil = to.Docker.GetAuth().GetQueryTokenSelf()
				}
			}
			if tc.inheritDockerToken &&
				to != nil && to.Docker != nil &&
				from != nil && from.Docker != nil {
				wantToken = from.Docker.GetAuth().GetTokenSelf()
				wantQueryToken, wantValidUntil = from.Docker.GetAuth().GetQueryTokenSelf()
			}

			// WHEN: Inherit is called on them.
			to.Inherit(from)

			prefix := fmt.Sprintf("%s\nRequire.Inherit()", packageName)

			// THEN: the Require tokens are inherited.
			gotToken, gotQueryToken, gotValidUntil := "", "", time.Time{}
			if to != nil && to.Docker != nil {
				gotToken = to.Docker.GetAuth().GetTokenSelf()
				gotQueryToken, gotValidUntil = to.Docker.GetAuth().GetQueryTokenSelf()
			}
			fieldTests := []test.FieldAssertion{
				{Name: "Token", Got: gotToken, Want: wantToken, Mode: test.CompareEqual},
				{Name: "QueryToken", Got: gotQueryToken, Want: wantQueryToken, Mode: test.CompareEqual},
				{Name: "ValidUntil", Got: gotValidUntil, Want: wantValidUntil, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Require"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRequire_DockerTagCheck__unit(t *testing.T) {
	// GIVEN: a Require and version.
	tests := []struct {
		name     string
		require  *Require
		version  string
		errRegex string
	}{
		{
			name:     "nil require",
			require:  nil,
			errRegex: `^$`,
		},
		{
			name:     "nil docker",
			require:  &Require{},
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: DockerTagCheck is called.
			err := tc.require.DockerTagCheck(tc.version)

			prefix := fmt.Sprintf(
				"%s\nRequire.DockerTagCheck(%q)",
				packageName, tc.version,
			)

			// THEN: the error matches expectation.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
		})
	}
}

func TestRequire_RemoveUnusedRequireDocker(t *testing.T) {
	defaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults, _ := DecodeDefaults("yaml", nil)
	defaults.SetDefaults(hardDefaults)

	tests := []struct {
		name    string
		require *Require
		nil     bool
	}{
		{
			name:    "nil require",
			require: nil,
			nil:     true,
		},
		{
			name: "nil Docker",
			require: &Require{
				Docker: nil,
			},
			nil: true,
		},
		{
			name: "Docker/Image and Tag, kept",
			require: test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: `+test.ArgusGitHubRepo+`
							tag: latest
					`)),
					svcStatus,
					defaults,
				)
			}),
			nil: false,
		},
		{
			name: "Docker/Image only",
			require: test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: `+test.ArgusGitHubRepo+`
					`)),
					svcStatus,
					defaults,
				)
			}),
			nil: false,
		},
		{
			name: "Docker/Tag only",
			require: test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							tag: latest
					`)),
					svcStatus,
					defaults,
				)
			}),
			nil: false,
		},
		{
			name: "Docker/empty Image and empty Tag",
			require: test.Must(t, func() (*Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: ''
							tag: ''
					`)),
					svcStatus,
					defaults,
				)
			}),
			nil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: removeUnusedRequireDocker is called.
			tc.require.removeUnusedRequireDocker()

			// THEN: the Docker is removed if it has no Image or Tag.
			if got := (tc.require == nil || tc.require.Docker == nil); got != tc.nil {
				t.Errorf(
					"%s\nRequire.removeUnusedRequireDocker() mismatch:\ngot:  nil=%t\nwant: nil=%t",
					packageName, got, tc.nil,
				)
			}
		})
	}
}
