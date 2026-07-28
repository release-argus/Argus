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

package service

import (
	"fmt"
	"strings"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/test"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestServices_Print(t *testing.T) {
	// GIVEN: a Services.
	tests := []struct {
		name     string
		services *Services
		ordering []string
		want     string
	}{
		{
			name:     "nil map with no ordering",
			services: nil,
			want:     "",
		},
		{
			name:     "nil map with ordering",
			ordering: []string{"foo", "bar"},
			services: nil,
			want:     "",
		},
		{
			name:     "map with nil Service and empty Service",
			ordering: []string{"foo", "bar"},
			services: &Services{
				"foo": nil,
				"bar": &Service{},
			},
			want: test.TrimYAML(`
				service:
					bar: {}`,
			),
		},
		{
			name:     "respects ordering",
			ordering: []string{"zulu", "alpha"},
			services: &Services{
				"zulu":  &Service{ID: "zulu", Comment: "a"},
				"alpha": &Service{ID: "alpha", Comment: "b"},
			},
			want: test.TrimYAML(`
				service:
					zulu:
						comment: a
					alpha:
						comment: b`,
			),
		},
		{
			name:     "respects reversed ordering",
			ordering: []string{"alpha", "zulu"},
			services: &Services{
				"zulu":  &Service{ID: "zulu", Comment: "a"},
				"alpha": &Service{ID: "alpha", Comment: "b"},
			},
			want: test.TrimYAML(`
				service:
					alpha:
						comment: b
					zulu:
						comment: a`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout(t)

			if tc.want != "" {
				tc.want += "\n"
			}
			tc.want = strings.ReplaceAll(tc.want, "\t", "")

			// WHEN: Print is called.
			tc.services.Print("", tc.ordering)

			prefix := fmt.Sprintf(
				"%s\nServices.Print(prefix=\"\", order=%v)",
				packageName, tc.ordering,
			)

			// THEN: it prints the want stdout.
			stdout := releaseStdout()
			if stdout != tc.want {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.want,
				)
			}
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name          string
		options       opt.Defaults
		latestVersion latestver.Defaults
		errRegex      string
	}{
		{
			name: "valid",
			options: *test.Must(t, func() (*opt.Defaults, error) {
				return opt.DecodeDefaults("yaml", []byte("interval: 10s"))
			}),
			latestVersion: latestver.Defaults{
				Common: lvbase.Defaults{
					Require: filter.RequireDefaults{
						Docker: *test.Must(t, func() (*docker.Defaults, error) {
							return docker.DecodeDefaults(
								"yaml", []byte("type: ghcr"),
								nil,
							)
						}),
					},
				},
			},
		},
		{
			name: "options with decode",
			options: *test.Must(t, func() (*opt.Defaults, error) {
				return opt.DecodeDefaults("yaml", []byte("interval: 10x"))
			}),
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*$`,
			),
		},
		{
			name: "latestVersion with decode",
			options: *test.Must(t, func() (*opt.Defaults, error) {
				return opt.DecodeDefaults("yaml", []byte("interval: 10s"))
			}),
			latestVersion: latestver.Defaults{
				Common: lvbase.Defaults{
					Require: filter.RequireDefaults{
						Docker: *test.Must(t, func() (*docker.Defaults, error) {
							return docker.DecodeDefaults(
								"yaml", []byte("type: randomType"),
								nil,
							)
						}),
					},
				},
			},
			errRegex: test.TrimYAML(`
				^latest_version:
					require:
						docker:
							type: "[^"]+" <invalid>`,
			),
		},
		{
			name: "all decode",
			options: *test.Must(t, func() (*opt.Defaults, error) {
				return opt.DecodeDefaults("yaml", []byte("interval: 10x"))
			}),
			latestVersion: latestver.Defaults{
				Common: lvbase.Defaults{
					Require: filter.RequireDefaults{
						Docker: *test.Must(t, func() (*docker.Defaults, error) {
							return docker.DecodeDefaults(
								"yaml", []byte("type: randomType"),
								nil,
							)
						}),
					},
				},
			},
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					require:
						docker:
							type: "[^"]+" <invalid>.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := &Defaults{
				Options:       tc.options,
				LatestVersion: tc.latestVersion,
			}

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				input.CheckValues,
			)
		})
	}
}

func TestServices_CheckValues(t *testing.T) {
	svcCfg := plainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// GIVEN: a Services.
	tests := []struct {
		name     string
		input    Services
		errRegex string
		changed  bool
	}{
		{
			name:     "nil map",
			input:    (Services)(nil),
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "single valid service",
			input: test.Must(t, func() (Services, error) {
				return DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						first:
							comment: foo_comment
							options:
								interval: 10s
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "multiple valid services/no changes",
			input: test.Must(t, func() (Services, error) {
				return DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						first:
							comment: foo_comment",
							options:
								interval: 10s
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
						second:
							comment: bar_comment
							options:
								interval: 10s
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "multiple valid services/1+ changed",
			input: test.Must(t, func() (Services, error) {
				return DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						first:
							comment: foo_comment
							options:
								interval: 10s
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
							webhook:
								wh:
									type: github
									url: example.com
									secret: Argus
									custom_headers:
										- key: foo
										  value: bar
						second:
							comment: bar_comment
							options:
								interval: 10s
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
			changed:  true,
		},
		{
			name: "multiple valid services/1+ changed but some error",
			input: test.Must(t, func() (Services, error) {
				return DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						first:
							comment: foo_comment
							options:
								interval: 10s
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
							webhook:
								wh:
									type: github
									url: example.com
									secret: Argus
									custom_headers:
										- key: foo
										  value: bar
						second:
							comment: bar_comment
							options:
								interval: 10x
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: test.TrimYAML(`
				^second:
					options:
						interval: "10x" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "single invalid service",
			input: test.Must(t, func() (Services, error) {
				return DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						first:
							comment: foo_comment
							options:
								interval: 10x
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `interval: "[^"]+" <invalid>.*$`,
			changed:  false,
		},
		{
			name: "multiple invalid services",
			input: test.Must(t, func() (Services, error) {
				return DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						foo:
							comment: foo_comment
							options:
								interval: 10x
						bar:
							comment: bar_comment
							options:
								interval: 10y
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
							deployed_version:
								type: url
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: test.TrimYAML(`
				^bar:
					options:
						interval: "10y" <invalid>.*
					deployed_version:
						url: <required>.*
				foo:
					latest_version and\/or deployed_version required`,
			),
			changed: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				packageName,
				tc.errRegex,
				tc.changed,
				tc.input.CheckValues,
			)
		})
	}
}

func TestServices_CheckValues__nil(t *testing.T) {
	// GIVEN: a nil Services pointer.
	var s *Services

	// WHEN/THEN: CheckValues returns no error and changed=false.
	_, _ = test.AssertCheckValuesWithErrorAndChanged(
		t,
		packageName,
		`^$`,
		false,
		s.CheckValues,
	)
}

func TestService_CheckValues(t *testing.T) {
	svcCfg := plainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// GIVEN: a Service.
	tests := []struct {
		name     string
		input    *Service
		noInit   bool
		errRegex string
		changed  bool
	}{
		{
			name:     "nil service",
			input:    (*Service)(nil),
			errRegex: `^$`,
			changed:  false,
		},
		{
			name: "options err",
			input: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							interval: 10x
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "nil latest_version + deployed_version",
			input: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						comment: foo_comment
						options:
							interval: 10x
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^latest_version and/or deployed_version required$`,
			changed:  false,
		},
		{
			name: "options + latest_version err",
			input: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						comment: foo_comment
						options:
							interval: 10x
						latest_version:
							type: github
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					url: <required>.*$`,
			),
			changed: false,
		},
		{
			name: "options + latest_version + deployed_version err",
			input: test.Must(t, func() (*Service, error) {
				cfgOverride := plainDefaultsConfig(t)
				cfgOverride.Hard.DeployedVersionLookup.Method = ""
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						comment: foo_comment
						options:
							interval: 10x
						latest_version:
							type: github
						deployed_version:
							type: url
							regex: '[0-9'
					`)),
					"",
					cfgOverride, notifyCfg, whCfg,
				)
			}),
			errRegex: test.TrimYAML(`
				^options:
					interval: "10x" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
					method: <required>.*
					regex: "\[0-9" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "options + latest_version + deployed_version + notify err",
			input: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						comment: foo_comment
						options:
							interval: 10x
						latest_version:
							type: github
						deployed_version:
							type: url
						notify:
							foo:
								type: discord
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
				notify:
					foo:
						url_fields:
							token: <required>.*
							webhookid: <required>.*$`,
			),
			changed: false,
		},
		{
			name: "options + latest_version + deployed_version + notify + command err",
			input: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						comment: foo_comment
						options:
							interval: 10x
						latest_version:
							type: github
						deployed_version:
							type: url
						notify:
							foo:
								type: discord
						command:
							- ["bash", "update.sh", "{{ version }"]
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: test.TrimYAML(`
				^options:
					interval: "[^"]+" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
				notify:
					foo:
						url_fields:
							token: <required>.*
							webhookid: <required>.*
				command:
					- item_0:
					  "bash .* <invalid>.*templating.*$`,
			),
			changed: false,
		},
		{
			name: "options + latest_version + deployed_version + notify + command + webhook err",
			input: test.Must(t, func() (*Service, error) {
				whCfgOverride := whtest.PlainConfig(t)
				whCfgOverride.HardDefaults.Type = ""
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						comment: foo_comment
						options:
							interval: 10x
						latest_version:
							type: github
						deployed_version:
							type: url
						notify:
							foo:
								type: discord
						command:
							- ["bash", "update.sh", "{{ version }"]
						webhook:
							wh: {}
					`)),
					"",
					svcCfg, notifyCfg, whCfgOverride,
				)
			}),
			errRegex: test.TrimYAML(`
				^options:
					interval: "10x" <invalid>.*
				latest_version:
					url: <required>.*
				deployed_version:
					url: <required>.*
				notify:
					foo:
						url_fields:
							token: <required>.*
							webhookid: <required>.*
				command:
					- item_0:
					  "bash .* <invalid>.*templating.*
				webhook:
					wh:
						type: <required>.*
						url: <required>.*
						secret: <required>.*$`,
			),
			changed: false,
		},
		{
			name: "notify changed",
			input: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						notify:
							foo:
								type: generic
								url_fields:
									host: x
									secret: y
									custom_headers: '{"foo": "bar"}'
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
			changed:  true,
		},
		{
			name: "webhook changed",
			input: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						webhook:
							wh:
								type: github
								url: example.com
								secret: Argus
								custom_headers:
									foo: bar
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
			changed:  true,
		},
		{
			name: "dashboard err",
			input: func() *Service {
				svc := test.Must(t, func() (*Service, error) {
					return DecodeService(
						"yaml", []byte(test.TrimYAML(`
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
						`)),
						"",
						svcCfg, notifyCfg, whCfg,
					)
				})
				svc.Dashboard.WebURL = "{{ invalid"
				return svc
			}(),
			errRegex: test.TrimYAML(`
				^dashboard:
					web_url: ".*" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "not changed if we have errors",
			input: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						comment: foo_comment
						options:
							interval: 10x
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						notify:
							foo:
								type: discord
								url_fields:
									token: x
									webhookid: y
						webhook:
							wh:
								type: github
								url: example.com
								secret: Argus
								custom_headers:
									foo: bar
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: test.TrimYAML(`
				^options:
					interval: "10x" <invalid>.*$`,
			),
			changed: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				packageName,
				tc.errRegex,
				tc.changed,
				tc.input.CheckValues,
			)
		})
	}
}

func TestService_CheckValues__deployed_version__manual(t *testing.T) {
	svcCfg := plainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// GIVEN: a Service with a `manual` deployed_version carrying a version.
	tests := []struct {
		name        string
		version     string
		errRegex    string
		wantVersion string
	}{
		{
			name:        "valid version is seeded into the Status",
			version:     "1.2.3",
			errRegex:    `^$`,
			wantVersion: "1.2.3",
		},
		{
			name:        "non-semantic version is rejected, and not seeded",
			version:     "1_2_3",
			errRegex:    `failed to convert "1_2_3" to a semantic version`,
			wantVersion: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						deployed_version:
							type: manual
							version: "`+tc.version+`"
					`)),
					"",
					svcCfg, notifyCfg, whCfg,
				)
			})

			// AND: a Status wired to the announce/database channels.
			svc.Status.AnnounceChannel = make(chan []byte, 8)
			svc.Status.DatabaseChannel = make(chan dbtype.Message, 8)

			// WHEN: CheckValues is called.
			_, _ = test.AssertCheckValuesWithErrorAndChanged(
				t,
				packageName,
				tc.errRegex,
				false,
				svc.CheckValues,
			)

			prefix := fmt.Sprintf("%s\nService.CheckValues()", packageName)

			// THEN: it only validates - nothing is applied yet.
			if got := svc.Status.DeployedVersion(); got != "" {
				t.Errorf(
					"%s should not apply the version\ngot: %q",
					prefix, got,
				)
			}

			// WHEN: the configured version is applied (as Track and a service edit do).
			svc.DeployedVersionLookup.ApplyConfiguredVersion()
			prefix = fmt.Sprintf("%s\nApplyConfiguredVersion()", packageName)

			// THEN: a valid version reaches the Status, and an invalid one does not.
			if got := svc.Status.DeployedVersion(); got != tc.wantVersion {
				t.Errorf(
					"%s mismatch on Status.DeployedVersion()\ngot:  %q\nwant: %q",
					prefix, got, tc.wantVersion,
				)
			}

			// AND: it is never broadcast.
			if got := len(svc.Status.AnnounceChannel); got != 0 {
				t.Errorf(
					"%s AnnounceChannel message count mismatch\ngot:  %d\nwant: 0",
					prefix, got,
				)
			}

			// AND: a valid version is persisted, so that it survives a restart.
			wantDBMessages := 0
			if tc.wantVersion != "" {
				wantDBMessages = 1
			}
			if got := len(svc.Status.DatabaseChannel); got != wantDBMessages {
				t.Errorf(
					"%s DatabaseChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, wantDBMessages,
				)
			}
		})
	}
}
