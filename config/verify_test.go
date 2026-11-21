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

package config

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/option"
	svctest "github.com/release-argus/Argus/service/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func testVerify(t *testing.T) *Config {
	t.Helper()

	cfg := &Config{}
	cfg.SaveChannel = make(chan bool, 10)

	cfg.Order = []string{"test"}

	defaults, hardDefaults := plainDefaults()
	cfg.Defaults = *defaults
	cfg.Defaults.Default()
	cfg.HardDefaults = *hardDefaults

	cfg.Notify = shoutrrr.ShoutrrrsDefaults{
		"test": shoutrrr.NewDefaults(
			"discord",
			cfg.Defaults.Notify["discord"].Options,
			cfg.Defaults.Notify["discord"].URLFields,
			cfg.Defaults.Notify["discord"].Params,
		),
	}

	cfg.Webhook = webhook.WebhooksDefaults{
		"test": &cfg.Defaults.Webhook,
	}

	serviceID := "test"
	cfg.Service = service.Services{
		serviceID: &service.Service{
			ID: serviceID,
			LatestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: owner/repo
					`)),
					nil,
					nil,
					latestver.DefaultsConfig{
						Soft: &cfg.Defaults.Service.LatestVersion,
						Hard: &cfg.HardDefaults.Service.LatestVersion,
					},
				)
			}),
		},
	}

	return cfg
}

func TestConfig_CheckValues(t *testing.T) {
	loadMu.Lock()
	nilUndefinedFlags(t, &Settings{})
	loadMu.Unlock()

	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	tests := []struct {
		name     string
		cfg      *Config
		errRegex string
		logRegex *string
		ok       bool
		wantSave bool
	}{
		{
			name:     "valid Config",
			cfg:      testVerify(t),
			ok:       true,
			errRegex: `^$`,
		},
		{
			name: "invalid Settings",
			cfg: &Config{
				Settings: Settings{
					Web: WebSettings{
						CertFile: "does_not_exist.pem",
					},
				},
			},
			errRegex: test.TrimYAML(`
				^settings:
					web:
						cert_file: .*does_not_exist.pem.* no such file.*`,
			),
		},
		{
			name: "invalid file read paths",
			cfg: &Config{
				Settings: Settings{
					Data: DataSettings{
						FileReadRoots: []string{""},
					},
				},
			},
			errRegex: test.TrimYAML(`
				^settings:
					data:
						file_read_roots: "\[\]" <invalid> \(must not contain an empty path\)\s$`,
			),
		},
		{
			name: "invalid Defaults",
			cfg: &Config{
				Defaults: Defaults{
					Service: service.Defaults{
						Options: *test.Must(t, func() (*opt.Defaults, error) {
							return opt.DecodeDefaults("yaml", []byte("interval: 1x"))
						}),
					},
				},
			},
			errRegex: test.TrimYAML(`
				^defaults:
					service:
						options:
							interval: "[^"]+" <invalid>`,
			),
		},
		{
			name: "invalid Notify",
			cfg: &Config{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"test": shoutrrr.NewDefaults(
						"discord",
						map[string]string{
							"delay": "2x",
						},
						nil, nil,
					),
				},
			},
			errRegex: test.TrimYAML(`
				^notify:
					test:
						options:
							delay: "[^"]+" <invalid>`,
			),
		},
		{
			name: "invalid Webhook",
			cfg: &Config{
				Webhook: webhook.WebhooksDefaults{
					"test": test.Must(t, func() (*webhook.Defaults, error) {
						return webhook.DecodeDefaults(
							"yaml", []byte(test.TrimYAML(`
									delay: 10x
									type: github
								`)),
						)
					}),
				},
			},
			errRegex: test.TrimYAML(`
				^webhook:
					test:
						delay: "10x" <invalid>`,
			),
		},
		{
			name: "invalid Service",
			cfg: test.Must(t, func() (*Config, error) {
				var cfg Config
				err := cfg.Decode([]byte(test.TrimYAML(`
					service:
						test:
							options:
								interval: 4x
							latest_version:
								type: github
								url: ` + test.ArgusGitHubRepo + `
				`)))
				return &cfg, err
			}),
			errRegex: test.TrimYAML(`
				^service:
					test:
						options:
							interval: "4x" <invalid>.*`,
			),
		},
		{
			name: "changed/only Defaults is saved",
			cfg: &Config{
				Defaults: Defaults{
					Notify: shoutrrr.ShoutrrrsDefaults{
						"gotify": shoutrrr.NewDefaults(
							"gotify",
							nil,
							map[string]string{"port": ":8080"}, // CheckValues strips the colon.
							nil,
						),
					},
				},
				SaveChannel: make(chan bool, 1),
			},
			ok:       true,
			wantSave: true,
			errRegex: `^$`,
		},
		{
			name: "changed/only Notify is saved",
			cfg: &Config{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"test": shoutrrr.NewDefaults(
						"gotify",
						nil,
						map[string]string{"port": ":8080"}, // CheckValues strips the colon.
						nil,
					),
				},
				SaveChannel: make(chan bool, 1),
			},
			ok:       true,
			wantSave: true,
			errRegex: `^$`,
		},
		{
			name: "changed/only Webhook is saved",
			cfg: &Config{
				Webhook: webhook.WebhooksDefaults{
					"test": test.Must(t, func() (*webhook.Defaults, error) {
						return webhook.DecodeDefaults(
							"yaml", []byte(test.TrimYAML(`
								type: github
								custom_headers:
									foo: bar
							`)),
						)
					}),
				},
				SaveChannel: make(chan bool, 1),
			},
			logRegex: new(`^DEPRECATED: .*webhook.custom_headers.*\s$`),
			ok:       true,
			wantSave: true,
			errRegex: `^$`,
		},
		{
			name: "valid Config that gets changed is saved",
			cfg: test.Must(t, func() (*Config, error) {
				cfg := testVerify(t)

				newService, err := service.DecodeService(
					"yaml", []byte(test.TrimYAML(`
						comment: foo_comment
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						webhook:
							"wh":
								type: "github"
								url: "example.com"
								secret: "Argus"
								custom_headers:
									foo: bar
					`)),
					"Test",
					svcCfg, notifyCfg, whCfg,
				)
				if err != nil {
					return nil, err
				}
				cfg.Service["added"] = newService

				return cfg, nil
			}),
			logRegex: new(`^DEPRECATED: .*\s$`),
			ok:       true,
			wantSave: true,
			errRegex: `^$`,
		},
		{
			name: "invalid Config that gets changed is not saved",
			cfg: test.Must(t, func() (*Config, error) {
				cfg := testVerify(t)

				newService := test.Must(t, func() (*service.Service, error) {
					return service.DecodeService(
						"yaml", []byte(test.TrimYAML(`
							comment: foo_comment
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
							webhook:
								"wh":
									type: "github"
									url: "example.com"
									secret: "Argus"
									custom_headers:
										foo: bar
						`)),
						"Test",
						svcCfg, notifyCfg, whCfg,
					)
				})
				cfg.Service["added"] = newService

				badService := test.Must(t, func() (*service.Service, error) {
					return service.DecodeService(
						"yaml", []byte(test.TrimYAML(`
							comment: foo_comment
							options:
								interval: 10x
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
							notify:
								"foo":
									type: "generic"
									url_fields:
										host: x
										secret: y
										custom_headers: '{"foo": "bar"}'
									headers:
										foo: bar
							webhook:
								"wh":
									type: "github"
									url: "example.com"
									secret: "Argus"
									custom_headers:
										foo: bar
						`)),
						"bad",
						svcCfg, notifyCfg, whCfg,
					)
				})
				cfg.Service["badService"] = badService

				return cfg, nil
			}),
			errRegex: test.TrimYAML(`
				^service:
					badService:
						options:
							interval: "10x" <invalid>.*\s$`,
			),
			logRegex: new(test.TrimYAML(`
				^DEPRECATED: .*webhook.custom_headers.*
				DEPRECATED: .*notify.generic.url_fields.custom_headers.*
				DEPRECATED: .*webhook.custom_headers.*
				FATAL: Config could not be parsed.*\s$`,
			)),
			ok:       false,
			wantSave: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseLog := test.CaptureLog(t, logx.Default())
			releaseStdout := test.CaptureStdout(t)

			test.AssertCallLog(t, tc.ok, tc.cfg.CheckValues, logx.ExitCodeChannel(), nil, "")

			stdout := releaseStdout()
			lines := strings.Split(stdout, "\n")
			gotLines := len(lines)
			wantLines := strings.Count(tc.errRegex, "\n") + 1
			if gotLines < wantLines {
				t.Fatalf(
					"stdout line count mismatch\ngot:  %d (%q)\nwant: >= %d (%q)",
					gotLines, stdout,
					wantLines, tc.errRegex,
				)
			}

			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.errRegex)
				return
			}

			logWant := `^FATAL: Config could not be parsed.*\s$`
			if tc.ok {
				logWant = `^$`
			}
			logWant = util.DerefOr(tc.logRegex, logWant)
			logOut := releaseLog()
			if !util.RegexCheck(logWant, logOut) {
				t.Errorf("log mismatch\ngot:  %q\nwant: %q", logOut, logWant)
			}

			saveQueued := len(tc.cfg.SaveChannel) > 0
			if saveQueued != tc.wantSave {
				t.Errorf("save queue mismatch\ngot:  %t\nwant: %t", saveQueued, tc.wantSave)
			}
		})
	}
}

func TestConfig_Print(t *testing.T) {
	config := testVerify(t)
	tests := []struct {
		name string
		flag bool
		want string
	}{
		{
			name: "flag: true",
			flag: true,
			want: configStr,
		},
		{
			name: "flag: false",
			flag: false,
			want: "",
		},
	}

	originalExit := exitAfterPrint
	t.Cleanup(func() {
		exitAfterPrint = originalExit
	})
	exitAfterPrint = func(_ int) {}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureStdout(t)

			config.Print(&tc.flag)

			stdout := releaseStdout()
			if stdout != tc.want {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.want)
			}
		})
	}
}

func TestConfig_Print__exitsWhenNotTesting(t *testing.T) {
	config := testVerify(t)
	flag := true

	originalExit := exitAfterPrint
	t.Cleanup(func() {
		exitAfterPrint = originalExit
	})

	exitCode := -1
	exitAfterPrint = func(code int) { exitCode = code }

	_ = test.CaptureStdout(t)

	config.Print(&flag)

	if got, want := exitCode, 0; got != want {
		t.Errorf("exit code mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

var configStr = test.TrimYAML(`
	defaults:
		service:
			options:
				interval: 10m
				semantic_versioning: true
			latest_version:
				type: github
				common:
					require:
						docker:
							type: hub
							tag: '{{ version }}'
				forgejo:
					common:
						use_prerelease: false
				github:
					use_prerelease: false
				url:
					allow_invalid_certs: false
			deployed_version:
				type: url
				allow_invalid_certs: false
				method: GET
			dashboard:
				auto_approve: false
		notify:
			bark:
				type: bark
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				url_fields:
					port: '443'
				params:
					title: Argus
			discord:
				type: discord
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				params:
					splitlines: 'yes'
					username: Argus
			generic:
				type: generic
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				params:
					contenttype: application/json
					disabletls: 'no'
					messagekey: message
					method: POST
					titlekey: title
			googlechat:
				type: googlechat
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
			gotify:
				type: gotify
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				url_fields:
					port: '443'
				params:
					disabletls: 'no'
					insecureskipverify: 'no'
					priority: '0'
					title: Argus
					useheader: 'no'
			homeassistant:
				type: homeassistant
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				params:
					disabletls: 'no'
					skiptlsverify: 'no'
					title: Argus
			ifttt:
				type: ifttt
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				params:
					messagevalue: '2'
					titlevalue: '0'
			join:
				type: join
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
			matrix:
				type: matrix
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				url_fields:
					port: '443'
				params:
					disabletls: 'no'
			mattermost:
				type: mattermost
				options:
					delay: 0s
					max_tries: '3'
					message: '<{{ service_url }}|{{ service_name | default:service_id }}> - {{ version }} released{% if web_url %} (<{{ web_url }}|changelog>){% endif %}'
				url_fields:
					port: '443'
					username: Argus
				params:
					disabletls: 'no'
			notifiarr:
				type: notifiarr
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
			ntfy:
				type: ntfy
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				url_fields:
					host: ntfy.sh
				params:
					disabletlsverification: 'no'
					title: Argus
			opsgenie:
				type: opsgenie
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
			pushbullet:
				type: pushbullet
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				params:
					title: Argus
			pushover:
				type: pushover
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
			rocketchat:
				type: rocketchat
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				url_fields:
					port: '443'
			shoutrrr:
				type: shoutrrr
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
			slack:
				type: slack
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				params:
					botname: Argus
			smtp:
				type: smtp
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				params:
					requirestarttls: 'no'
					skiptlsverify: 'no'
					usehtml: 'no'
					usestarttls: 'yes'
			teams:
				type: teams
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
			telegram:
				type: telegram
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
				params:
					notification: 'yes'
					preview: 'yes'
			zulip:
				type: zulip
				options:
					delay: 0s
					max_tries: '3'
					message: '{{ service_name | default:service_id }} - {{ version }} released'
		webhook:
			type: github
			allow_invalid_certs: false
			desired_status_code: 0
			delay: 0s
			max_tries: 3
			silent_fails: false
	notify:
		test:
			type: discord
			options:
				delay: 0s
				max_tries: '3'
				message: '{{ service_name | default:service_id }} - {{ version }} released'
			params:
				splitlines: 'yes'
				username: Argus

	webhook:
		test:
			type: github
			allow_invalid_certs: false
			desired_status_code: 0
			delay: 0s
			max_tries: 3
			silent_fails: false

	service:
		test:
			latest_version:
				type: github
				url: owner/repo
`) + "\n"
