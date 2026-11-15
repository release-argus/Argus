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
				defaults, err := DecodeDefaults("yaml", data)
				if err != nil {
					return nil, err
				}
				hardDefaults, err := DecodeDefaults("yaml", data)
				if err != nil {
					return nil, err
				}
				defaults.SetDefaults(hardDefaults)
				svcStatus, _ := statustest.New("yaml", nil)

				req, err := Decode(
					"yaml", []byte(test.TrimYAML(`
						docker: {}
					`)),
					svcStatus,
					defaults,
				)
				if err != nil {
					return nil, err
				}
				req.Docker.(*docker.HubRegistry).Image = ""

				if req.Docker.GetTag() == "" {
					t.Fatal("Require.Docker.GetTag() mismatch\ngot:  \"\"\nwant: non-empty")
				}
				return req, nil
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
				defaults, err := DecodeDefaults("yaml", data)
				if err != nil {
					return nil, err
				}
				hardDefaults, err := DecodeDefaults("yaml", data)
				if err != nil {
					return nil, err
				}
				defaults.SetDefaults(hardDefaults)
				svcStatus, _ := statustest.New("yaml", nil)

				return Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							image: foo
					`)),
					svcStatus,
					defaults,
				)
			}),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.req.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestRequire_String(t *testing.T) {
	defaults, _ := plainDefaults()

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
				"Require.String",
				tc.require.String,
				tc.want,
			)
		})
	}
}

func TestRequire_Unmarshal(t *testing.T) {
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
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
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
				"Require.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestRequire_Copy(t *testing.T) {
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
				name := "non-nil Status"
				if nilStatus {
					name = "nil Status"
				}
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					var newStatus *status.Status
					if !nilStatus {
						newStatus = &status.Status{}
					}

					got := tc.req.Copy(newStatus)

					if got == nil {
						if tc.req != nil {
							t.Error("copy mismatch\ngot:  nil\nwant: non-nil")
						}
						return
					}

					fieldTests := []test.FieldAssertion{
						{Name: "RegexContent", Got: got.RegexContent, Want: tc.req.RegexContent, Mode: test.CompareEqual},
						{Name: "RegexVersion", Got: got.RegexVersion, Want: tc.req.RegexVersion, Mode: test.CompareEqual},
						{Name: "Command", Got: &got.Command, Want: &tc.req.Command, Mode: test.CompareDifferentPointer},
						{Name: "Command.String()", Got: got.Command.String(), Want: tc.req.Command.String(), Mode: test.CompareEqual},
						{Name: "Docker", Got: got.Docker, Want: tc.req.Docker, Mode: test.CopiedPointerMode(tc.req.Docker)},
						{Name: "Status", Got: got.Status, Want: newStatus, Mode: test.CompareSamePointer},
						{Name: "defaults", Got: got.defaults, Want: tc.req.defaults, Mode: test.CompareSamePointer},
					}
					if got.Docker != nil && tc.req.Docker != nil {
						fieldTests = append(fieldTests, test.FieldAssertion{Name: "Docker.String()", Got: got.Docker.String(""), Want: tc.req.Docker.String(""), Mode: test.CompareEqual})
					}
					if testErr := test.AssertFields(t, fieldTests, "Require"); testErr != nil {
						t.Fatal(testErr)
					}
				})
			}
		})
	}
}

func TestRequire_DockerTagCheck__unit(t *testing.T) {
	tests := []struct {
		name     string
		require  *Require
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

			err := tc.require.DockerTagCheck("1.2.3")
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestRequire_Init(t *testing.T) {
	defaults, _ := plainDefaults()

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
							image: foo
							tag: bar
					`)),
					svcStatus,
					defaults,
				)
				if err != nil {
					return nil, err
				}

				req.Docker = docker.RegistryMap["ghcr"]()
				return req, nil
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
							image: foo
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
				WebURL: "https://example.com",
			}
			svcStatus.Init(
				0, 0, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				svcDashboard,
			)
			svcStatus.SetDeployedVersion("1.2.3", "", false)

			tc.req.Init(&svcStatus, defaults)

			if tc.req == nil {
				return
			}

			if tc.req.Status != &svcStatus {
				t.Fatalf("Require.Status mismatch\ngot:  %p\nwant: %p", tc.req.Status, &svcStatus)
			}

			if !tc.wantDockerCheck {
				if tc.req.Docker != nil {
					t.Fatalf("Require.Docker mismatch\ngot:  %T\nwant: nil", tc.req.Docker)
				}
				return
			}

			expectedDefaults, err := dockertest.GetDefaultOfDockerType(tc.req.Docker.GetType(), &defaults.Docker)
			if err != nil {
				t.Fatalf("dockertest.GetDefaultOfDockerType() failed: %v", err)
			}
			if got := tc.req.Docker.Defaults(); got != expectedDefaults {
				t.Fatalf(
					"Require.Docker.Defaults() mismatch\ngot:  %p\nwant: %p",
					got, expectedDefaults,
				)
			}
		})
	}
}

func TestRequire_CheckValues(t *testing.T) {
	defaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults, _ := DecodeDefaults("yaml", nil)
	defaults.SetDefaults(hardDefaults)

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
				tc.errRegex,
				tc.input.CheckValues,
			)

			if gotYAML := tc.input.String(""); gotYAML != tc.wantYAML {
				t.Fatalf("Require.String() mismatch\ngot:  %q\nwant: %q", gotYAML, tc.wantYAML)
			}
		})
	}
}

func TestRequire_Inherit(t *testing.T) {
	defaults, _ := plainDefaults()
	fromQueryToken := "from-query-token"
	fromValidUntil := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)

	type overrides struct {
		overrides string
		isNil     bool
	}

	tests := []struct {
		name               string
		from, to           overrides
		inheritDockerToken bool
	}{
		{
			name: "nil to",
			to: overrides{
				isNil: true,
			},
			inheritDockerToken: false,
		},
		{
			name: "nil from",
			from: overrides{
				isNil: true,
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
			name:               "same token, copy",
			inheritDockerToken: true,
		},
		{
			name:               "secret token, copy",
			inheritDockerToken: true,
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						auth:
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
						type: hub
						auth:
							username: foo
							token: hub-token
				`),
			},
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						type: hub
						auth:
							username: bar
							token: hub-token
				`),
			},
			inheritDockerToken: false,
		},
		{
			name: "change of Token, no copy",
			from: overrides{
				overrides: test.TrimYAML(`
					docker:
						auth:
							token: foo
				`),
			},
			to: overrides{
				overrides: test.TrimYAML(`
					docker:
						auth:
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
			if !tc.from.isNil {
				svcStatus, _ := statustest.New("yaml", nil)
				var err error
				from, err = Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: `+test.ArgusGitHubRepo+`
							tag: '{{ version }}'
							auth:
								token: ghcr-token
					`)),
					svcStatus,
					defaults,
				)
				if err != nil {
					t.Fatalf("from: Decode() failed: %v", err)
				}
				if from, err = from.ApplyOverrides(
					"yaml", []byte(tc.from.overrides),
					from.Status,
					defaults,
				); err != nil {
					t.Fatalf("from: Require.ApplyOverrides() failed: %v", err)
				}
				if from != nil && from.Docker != nil {
					from.Docker.GetAuth().SetQueryToken(fromQueryToken, fromValidUntil)
				}
			}
			var to *Require
			wantToken, wantQueryToken := "", ""
			// to's own ValidUntil moves on each read, so assert only whether from's was copied.
			validUntilMode := test.CompareNotEqual
			if !tc.to.isNil {
				svcStatus, _ := statustest.New("yaml", nil)
				var err error
				to, err = Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: `+test.ArgusGitHubRepo+`
							tag: '{{ version }}'
							auth:
								token: ghcr-token
					`)),
					svcStatus,
					defaults,
				)
				if err != nil {
					t.Fatalf("to: Decode() failed: %v", err)
				}
				if to, err = to.ApplyOverrides(
					"yaml", []byte(tc.to.overrides),
					to.Status,
					defaults,
				); err != nil {
					t.Fatalf("to: Require.ApplyOverrides() failed: %v", err)
				}
				if to != nil && to.Docker != nil {
					wantToken = to.Docker.GetAuth().GetTokenSelf()
					wantQueryToken, _ = to.Docker.GetAuth().GetQueryTokenSelf()
				}
			}
			if tc.inheritDockerToken &&
				to != nil && to.Docker != nil &&
				from != nil && from.Docker != nil {
				wantToken = from.Docker.GetAuth().GetTokenSelf()
				wantQueryToken = fromQueryToken
				validUntilMode = test.CompareEqual
			}

			to.Inherit(from)

			gotToken, gotQueryToken, gotValidUntil := "", "", time.Time{}
			if to != nil && to.Docker != nil {
				gotToken = to.Docker.GetAuth().GetTokenSelf()
				gotQueryToken, gotValidUntil = to.Docker.GetAuth().GetQueryTokenSelf()
			}
			fieldTests := []test.FieldAssertion{
				{Name: "Docker.GetAuth().GetTokenSelf()", Got: gotToken, Want: wantToken, Mode: test.CompareEqual},
				{Name: "Docker.GetAuth().GetQueryTokenSelf() queryToken", Got: gotQueryToken, Want: wantQueryToken, Mode: test.CompareEqual},
				{Name: "Docker.GetAuth().GetQueryTokenSelf() validUntil", Got: gotValidUntil, Want: fromValidUntil, Mode: validUntilMode},
			}
			if testErr := test.AssertFields(t, fieldTests, "Require"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestRequire_RemoveUnusedRequireDocker(t *testing.T) {
	defaults, _ := plainDefaults()

	tests := []struct {
		name    string
		require *Require
		wantNil bool
	}{
		{
			name:    "nil require",
			require: nil,
			wantNil: true,
		},
		{
			name: "nil Docker",
			require: &Require{
				Docker: nil,
			},
			wantNil: true,
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
			wantNil: false,
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
			wantNil: false,
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
			wantNil: false,
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
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.require.removeUnusedRequireDocker()

			got, want := "non-nil", "non-nil"
			if tc.require == nil || tc.require.Docker == nil {
				got = "nil"
			}
			if tc.wantNil {
				want = "nil"
			}
			if got != want {
				t.Errorf("Require.Docker mismatch\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}
