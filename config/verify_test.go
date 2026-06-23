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
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	latestver "github.com/release-argus/Argus/service/latest_version"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
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

	defaults, hardDefaults := plainDefaults(t)
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

	cfg.WebHook = webhook.WebHooksDefaults{
		"test": &cfg.Defaults.WebHook,
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
					lvbase.DefaultsConfig{
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
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// GIVEN: variations of Config to test.
	tests := []struct {
		name     string
		cfg      *Config
		errRegex string
		logRegex *string
		ok       bool
		wantSave bool
	}{
		{
			name: "valid Config",
			cfg:  testVerify(t),
			ok:   true,
		},
		{
			name: "invalid Settings",
			cfg: &Config{
				Settings: Settings{
					SettingsBase: SettingsBase{
						Web: WebSettings{
							CertFile: "does_not_exist.pem",
						},
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
			name: "invalid WebHook",
			cfg: &Config{
				WebHook: webhook.WebHooksDefaults{
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
			name: "valid Config that gets changed is saved",
			cfg: test.Must(t, func() (*Config, error) {
				cfg := testVerify(t)

				newService, _ := service.DecodeService(
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
				cfg.Service[t.Name()] = newService

				return cfg, nil
			}),
			logRegex: test.Ptr(`^DEPRECATED: .*\s$`),
			ok:       true,
			wantSave: true,
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
				cfg.Service[t.Name()] = newService

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
			logRegex: test.Ptr(test.TrimYAML(`
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
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseLog := test.CaptureLog(t, logx.Default())
			releaseStdout := test.CaptureStdout(t)

			resultChannel := make(chan bool, 1)
			// WHEN: CheckValues is called on them.
			resultChannel <- tc.cfg.CheckValues()

			prefix := fmt.Sprintf("%s\nConfig.CheckValues()", packageName)

			// THEN: this call will/won't crash the program.
			if err := test.AssertChannelBool(
				t,
				tc.ok,
				resultChannel,
				logx.ExitCodeChannel(),
				nil,
			); err != nil {
				t.Fatalf("%s %v", prefix, err)
			}

			// AND: the error line count matches.
			stdout := releaseStdout()
			lines := strings.Split(stdout, "\n")
			gotLines := len(lines)
			wantLines := strings.Count(tc.errRegex, "\n")
			if gotLines < wantLines {
				t.Fatalf(
					"%s error line count mismatch\ngot:  %d\nwant: %d\nerrRegex: %q\nstdout:   %q",
					prefix,
					gotLines, wantLines,
					tc.errRegex, stdout,
				)
				return
			}

			// AND: the error regex matches.
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.errRegex,
				)
				return
			}

			// AND: the log matches.
			logWant := `^FATAL: Config could not be parsed.*\s$`
			if tc.ok {
				logWant = `^$`
			}
			logWant = util.DerefOr(tc.logRegex, logWant)
			logOut := releaseLog()
			if !util.RegexCheck(logWant, logOut) {
				t.Errorf(
					"%s log mismatch\ngot:  %q\nwant: %q",
					prefix, logOut, logWant,
				)
			}

			// AND: saves are queued as expected.
			saveQueued := len(tc.cfg.SaveChannel) > 0
			if saveQueued != tc.wantSave {
				t.Errorf(
					"%s save queue mismatch\ngot:  %t\nwant: %t",
					prefix, saveQueued, tc.wantSave,
				)
			}
		})
	}
}

func TestConfig_Print(t *testing.T) {
	// GIVEN: a Config and print flags of true and false.
	config := testVerify(t)
	tests := []struct {
		flag bool
		want string
	}{
		{
			flag: true,
			want: configStr,
		},
		{
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
		name := fmt.Sprintf("flag: %t", tc.flag)
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout(t)

			// WHEN: Print is called with these flags.
			config.Print(&tc.flag)

			// THEN: config is printed only when the flag is true.
			stdout := releaseStdout()
			if stdout != tc.want {
				t.Errorf(
					"%s\nConfig.Print() mismatch with %s\ngot:  %q\n\nwant: %q",
					packageName, name,
					stdout, tc.want,
				)
			}
		})
	}
}

func TestConfig_Print__exitsWhenNotTesting(t *testing.T) {
	// GIVEN: print flag enabled.
	config := testVerify(t)
	flag := true

	originalExit := exitAfterPrint
	t.Cleanup(func() {
		exitAfterPrint = originalExit
	})

	var exitCode int
	exitAfterPrint = func(code int) { exitCode = code }

	_ = test.CaptureStdout(t)

	// WHEN: Print is called.
	config.Print(&flag)

	// THEN: the process exit path is invoked with code 0.
	if exitCode != 0 {
		t.Errorf(
			"%s\nConfig.Print() exit code mismatch\ngot:  %d\nwant: 0",
			packageName, exitCode,
		)
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
				allow_invalid_certs: false
				use_prerelease: false
				require:
					docker:
						type: hub
						tag: '{{ version }}'
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
					requestmethod: POST
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
