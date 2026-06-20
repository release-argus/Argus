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

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/dashboard"
	dvbase "github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func TestDecodeDefaults(t *testing.T) {
	// GIVEN: data in a given format to Decode into Defaults.
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:   "JSON/service, notify, webhook",
			format: "json",
			data: test.TrimJSON(`{
				"service": {
					"options": {
						"interval": "1h"
					}
				},
				"notify": {
					"gotify": {
						"url_fields": {
							"port": "444"
						}
					}
				},
				"webhook": {
					"delay": "2s"
				}
			}`),
			want: test.TrimYAML(`
				service:
					options:
						interval: 1h
				notify:
					gotify:
						url_fields:
							port: '444'
				webhook:
					delay: 2s
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/service, notify, webhook",
			format: "yaml",
			data: test.TrimYAML(`
				service:
					options:
						interval: 1h
				notify:
					gotify:
						url_fields:
							port: '444'
				webhook:
					delay: 2s
			`),
			want: test.TrimYAML(`
				service:
					options:
						interval: 1h
				notify:
					gotify:
						url_fields:
							port: '444'
				webhook:
					delay: 2s
			`),
			errRegex: `^$`,
		},
		{
			name:   "JSON/invalid data type",
			format: "json",
			data: test.TrimJSON(`{
				"notify": {
					"gotify": {
						"url_fields": {
							"port": 444
						}
					}
				}
			}`),
			errRegex: test.TrimYAML(`
				^defaults:
					json: .*unmarshal .*$`,
			),
		},
		{
			name:   "YAML/invalid data type",
			format: "yaml",
			data: test.TrimYAML(`
				notify:
					gotify:
						url_fields:
							port: [444]
			`),
			errRegex: test.TrimYAML(`
				^defaults:
					[^\s]+ .*unmarshal .*`,
			),
		},
		{
			name:   "JSON/invalid Service block",
			format: "json",
			data: test.TrimJSON(`{
				"service": {
					"options": {
						"interval": [ 1, 2 ]
					}
				}
			}`),
			errRegex: test.TrimYAML(`
				^defaults:
					json: .*unmarshal.*$`,
			),
		},
		{
			name:   "YAML/invalid Service block",
			format: "yaml",
			data: test.TrimYAML(`
				service:
					options:
						interval:
							- 1
							- 2
			`),
			errRegex: test.TrimYAML(`
				^defaults:
					[^\s]+ .*unmarshal.*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				DecodeDefaults,
				tc.format, tc.data,
				func(v *Defaults) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestDefaults_Unmarshal(t *testing.T) {
	// GIVEN: a string in a given format to unmarshal into Defaults.
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:   "JSON/empty",
			format: "json",
			data:   "",
			want:   "",
			errRegex: test.TrimYAML(`
				^jsontext:
					unexpected EOF$`,
			),
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
			errRegex: "^$",
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"service": {
					"options": {
						"interval": "1m"
					}
				},
				"notify": {
					"gotify": {
						"url_fields": {
							"foo": "bar"
						}
					}
				},
				"webhook": {
					"allow_invalid_certs": false
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				service:
					options:
						interval: 1m
				notify:
					gotify:
						url_fields:
							foo: bar
				webhook:
					allow_invalid_certs: false
			`),
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				service:
					options:
						interval: 1m
				notify:
					gotify:
						url_fields:
							foo: bar
				webhook:
					allow_invalid_certs: false
			`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				service:
					options:
						interval: 1m
				notify:
					gotify:
						url_fields:
							foo: bar
				webhook:
					allow_invalid_certs: false
			`),
		},
		{
			name:     "JSON/static fields err",
			format:   "json",
			data:     `{"notify": ["abc"]}`,
			errRegex: `^json: .*unmarshal .*array.*$`,
		},
		{
			name:     "YAML/static fields err",
			format:   "yaml",
			data:     `notify: [abc]`,
			errRegex: `^[^\s]+ sequence was used.*`,
		},
		{
			name:     "JSON/dynamic fields (service) err",
			format:   "json",
			data:     `{"service": ["abc"]}`,
			errRegex: `^json: .*unmarshal .*array.*$`,
		},
		{
			name:   "YAML/dynamic fields (service) err",
			format: "yaml",
			data:   `service: [abc]`,
			errRegex: test.TrimYAML(`
				[^\s]+ sequence was used.*
				[^\s]+.*\[abc\]
				\s+\^$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Defaults, error) {
					var zero Defaults
					err := decode.Unmarshal(format, data, &zero)
					return &zero, err
				},
				tc.format, tc.data,
				func(v *Defaults) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"Defaults",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name     string
		defaults *Defaults
		want     bool
	}{
		{
			name:     "nil",
			defaults: nil,
			want:     true,
		},
		{
			name:     "empty",
			defaults: &Defaults{},
			want:     true,
		},
		{
			name: "non-empty/Service",
			defaults: &Defaults{
				Service: service.Defaults{
					Options: opt.Defaults{
						Base: opt.Base{
							Interval: "1s",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Notify",
			defaults: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"foo": &shoutrrr.Defaults{
						Base: shoutrrr.Base{
							Type: "discord",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "non-empty/WebHook",
			defaults: &Defaults{
				WebHook: webhook.Defaults{
					Base: webhook.Base{
						Type: "github",
					},
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on the Defaults.
			got := tc.defaults.IsZero()

			// THEN: the result matches the expected value.
			if got != tc.want {
				t.Errorf(
					"%s\nDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestDefaults_String(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name     string
		defaults *Defaults
		want     string
	}{
		{
			name:     "nil",
			defaults: nil,
			want:     "",
		},
		{
			name:     "empty",
			defaults: &Defaults{},
			want:     "{}\n",
		},
		{
			name: "filled",
			defaults: &Defaults{
				Service: service.Defaults{
					Options: *test.Must(t, func() (*opt.Defaults, error) {
						return opt.DecodeDefaults(
							"yaml", []byte(test.TrimYAML(`
								interval: 1m
								semantic_versioning: false
							`)),
						)
					}),
					LatestVersion: lvbase.Defaults{
						AccessToken:       "foo",
						AllowInvalidCerts: test.Ptr(true),
						UsePreRelease:     test.Ptr(false),
						Options: test.Must(t, func() (*opt.Defaults, error) {
							return opt.DecodeDefaults("yaml", []byte("interval: 1m"))
						}),
						Require: filter.RequireDefaults{
							Docker: *test.Must(t, func() (*docker.Defaults, error) {
								return docker.DecodeDefaults(
									"yaml", []byte(test.TrimYAML(`
										type: ghcr
										image: imageFallback
										tag: tagFallback
										registry:
											ghcr:
												image: imageGHCR
												tag: tagGHCR
												auth:
													username: usernameGHCR
													token: tokenGHCR
											hub:
												image: imageHub
												tag: tagHub
												auth:
													username: usernameHub
													token: tokenHub
											quay:
												image: imageQuay
												tag: tagQuay
												auth:
													username: usernameQuay
													token: tokenQuay
									`)),
									test.Must(t, func() (*docker.Defaults, error) {
										return docker.DecodeDefaults(
											"yaml", []byte(test.TrimYAML(`
												type: ghcr
												ghcr:
													image: imageGHCRother
													auth:
														username: usernameGHCRother
														token: tokenGHCRother
												hub:
													image: imageHubOther
													tag: tagHubOther
													auth:
														username: usernameHubOther
														token: tokenHubOther
												quay:
													image: imageQuayOther
													auth:
														username: usernameQuayOther
														token: tokenQuayOther
											`)),
											nil,
										)
									}),
								)
							}),
						},
					},
					DeployedVersionLookup: dvbase.Defaults{
						AllowInvalidCerts: test.Ptr(false),
					},
					Dashboard: *test.Must(t, func() (*dashboard.Defaults, error) {
						return dashboard.DecodeDefaults("yaml", []byte("auto_approve: true"))
					}),
				},
				Notify: shoutrrr.ShoutrrrsDefaults{
					"discord": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"message": "foo {{ version }}",
						},
						map[string]string{
							"host": "example.com",
						},
						map[string]string{
							"username": "Argus",
						},
					),
				},
				WebHook: *test.Must(t, func() (*webhook.Defaults, error) {
					return webhook.DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							allow_invalid_certs: true
							headers:
								- key: X-Header
								  value: foo
							delay: 0s
							desired_status_code: 203
							max_tries: 2
							secret: secret!!!
							silent_fails: false
							type: github
							url: https://example.com
						`)),
					)
				}),
			},
			want: test.TrimYAML(`
				service:
					options:
						interval: 1m
						semantic_versioning: false
					latest_version:
						access_token: foo
						allow_invalid_certs: true
						use_prerelease: false
						require:
							docker:
								type: ghcr
								image: imageFallback
								tag: tagFallback
								registry:
									ghcr:
										image: imageGHCR
										tag: tagGHCR
										auth:
											token: tokenGHCR
									hub:
										image: imageHub
										tag: tagHub
										auth:
											username: usernameHub
											token: tokenHub
									quay:
										image: imageQuay
										tag: tagQuay
										auth:
											token: tokenQuay
					deployed_version:
						allow_invalid_certs: false
					dashboard:
						auto_approve: true
				notify:
					discord:
						options:
							message: foo {{ version }}
						url_fields:
							host: example.com
						params:
							username: Argus
				webhook:
					type: github
					url: https://example.com
					allow_invalid_certs: true
					headers:
						- key: X-Header
							value: foo
					secret: secret!!!
					desired_status_code: 203
					delay: 0s
					max_tries: 2
					silent_fails: false
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.defaults.String,
				tc.want,
			)
		})
	}
}

func TestDefaults_Default(t *testing.T) {
	// GIVEN: nil defaults.
	var defaults Defaults

	// WHEN: Default is called on it.
	defaults.Default()
	tests := []struct {
		name      string
		got, want string
	}{
		{
			name: "Service.Interval",
			got:  defaults.Service.Options.Interval,
			want: "10m",
		},
		{
			name: "Notify.discord.username",
			got:  defaults.Notify["discord"].GetParam("username"),
			want: "Argus",
		},
		{
			name: "WebHook.Delay",
			got:  defaults.WebHook.Delay,
			want: "0s",
		},
	}

	// THEN: the defaults are set correctly.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Log(tc.name)
				t.Errorf(
					"%s\nDefaults.Default() %q value mismatch\ngot:  %q\nwant: %q",
					packageName, tc.name,
					tc.got, tc.want,
				)
			}
		})
	}
}

func TestDefaults_Default__fail(t *testing.T) {
	releaseStdout := test.CaptureLog(t, logx.Default())
	// GIVEN: Defaults, and an environment variable that will cause MapEnvToStruct to error.
	var defaults Defaults
	env := map[string]string{"ARGUS_SERVICE_OPTIONS_INTERVAL": "99 something"}
	test.SetEnv(t, env)

	resultChannel := make(chan bool, 1)
	// WHEN: Default is called on the Defaults.
	resultChannel <- defaults.Default()

	prefix := fmt.Sprintf("%s\nDefaults.Default()", packageName)

	// THEN: if false is returned, the error is logged.
	if err := test.AssertChannelBool(
		t,
		false,
		resultChannel,
		logx.ExitCodeChannel(),
		releaseStdout,
	); err != nil {
		t.Fatal(prefix + err.Error())
	}

	// AND: the stdout matches the expected result.
	stdout := releaseStdout()
	wantSubstring := `One or more 'ARGUS_' environment variables are invalid:`
	if !strings.Contains(stdout, wantSubstring) {
		t.Errorf(
			"%s stdout mismatch\ngot:  %q\nwant: %q",
			prefix, stdout, wantSubstring,
		)
	}
}

func TestDefaults_MapEnvToStruct(t *testing.T) {
	var unmodifiedDefaults Defaults
	unmodifiedDefaults.Default()
	// GIVEN: Defaults and a bunch of env vars.
	tests := []struct {
		name     string
		env      map[string]string
		want     *Defaults
		errRegex string
	}{
		{
			name: "empty vars ignored",
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99m",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": "",
			},
			want: &Defaults{
				Service: service.Defaults{
					Options: *test.Must(t, func() (*opt.Defaults, error) {
						return opt.DecodeDefaults("yaml", []byte("interval: 99m"))
					}),
				},
			},
		},
		{
			name: "service.options/valid",
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99m",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": "true",
			},
			want: &Defaults{
				Service: service.Defaults{
					Options: *test.Must(t, func() (*opt.Defaults, error) {
						return opt.DecodeDefaults(
							"yaml", []byte(test.TrimYAML(`
								interval: 99m
								semantic_versioning: true`,
							)),
						)
					}),
				},
			},
		},
		{
			name: "service.options/invalid time.duration - interval",
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99 something",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": "true",
			},
			errRegex: `ARGUS_SERVICE_OPTIONS_INTERVAL: "[^"]+" <invalid>`,
		},
		{
			name: "service.options/invalid bool - semantic version",
			env: map[string]string{
				"ARGUS_SERVICE_OPTIONS_INTERVAL":            "99",
				"ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING": "foo",
			},
			errRegex: `ARGUS_SERVICE_OPTIONS_SEMANTIC_VERSIONING: "foo" <invalid>`,
		},
		{
			name: "service.latest_version/valid",
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN":        "ghp_something",
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS": "true",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE":      "true",
			},
			want: &Defaults{
				Service: service.Defaults{
					LatestVersion: lvbase.Defaults{
						AccessToken:       "ghp_something",
						AllowInvalidCerts: test.Ptr(true),
						UsePreRelease:     test.Ptr(true),
					},
				},
			},
		},
		{
			name: "service.latest_version.require/valid",
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE":                        "ghcr",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_IMAGE":                       "imageFallback",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TAG":                         "tagFallback",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_GHCR_IMAGE":         "imageForGHCR",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_GHCR_TAG":           "tagForGHCR",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_GHCR_AUTH_TOKEN":    "tokenForGHCR",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_GHCR_AUTH_USERNAME": "usernameForGHCR",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_HUB_IMAGE":          "imageForHub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_HUB_TAG":            "tagForHub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_HUB_AUTH_TOKEN":     "tokenForHub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_HUB_AUTH_USERNAME":  "usernameForHub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_QUAY_IMAGE":         "imageForQuay",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_QUAY_TAG":           "tagForQuay",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_QUAY_AUTH_TOKEN":    "tokenForQuay",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_QUAY_AUTH_USERNAME": "usernameForQuay",
			},
			want: &Defaults{
				Service: service.Defaults{
					LatestVersion: lvbase.Defaults{
						Require: filter.RequireDefaults{
							Docker: *test.Must(t, func() (*docker.Defaults, error) {
								return docker.DecodeDefaults(
									"yaml", []byte(test.TrimYAML(`
										type: ghcr
										image: imageFallback
										tag: tagFallback
										registry:
											ghcr:
												image: imageForGHCR
												tag: tagForGHCR
												auth:
													token: tokenForGHCR
											hub:
												image: imageForHub
												tag: tagForHub
												auth:
													username: usernameForHub
													token: tokenForHub
											quay:
												image: imageForQuay
												tag: tagForQuay
												auth:
													token: tokenForQuay
									`)),
									nil,
								)
							}),
						},
					},
				},
			},
		},
		{
			name: "service.latest_version.require/invalid type",
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE":                       "foo",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_GHCR_AUTH_TOKEN":   "tokenForGHCR",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_HUB_AUTH_TOKEN":    "tokenForDockerHub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_HUB_AUTH_USERNAME": "usernameForDockerHub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_REGISTRY_QUAY_AUTH_TOKEN":   "tokenForQuay",
			},
			errRegex: test.TrimYAML(`ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE: "foo" <invalid> .+`),
		},
		{
			name: "service.latest_version/invalid bool/allow_invalid_certs",
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN":        "ghp_something",
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS": "bar",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE":      "true",
			},
			errRegex: `ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS: "bar" <invalid>`,
		},
		{
			name: "service.latest_version/invalid bool/use_prerelease",
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN":        "ghp_something",
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS": "true",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE":      "bop",
			},
			errRegex: `ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE: "bop" <invalid>`,
		},
		{
			name: "service.deployed_version/valid",
			env: map[string]string{
				"ARGUS_SERVICE_DEPLOYED_VERSION_ALLOW_INVALID_CERTS": "true",
			},
			want: &Defaults{
				Service: service.Defaults{
					DeployedVersionLookup: dvbase.Defaults{
						AllowInvalidCerts: test.Ptr(true),
					},
				},
			},
		},
		{
			name: "service.deployed_version/invalid bool - allow_invalid_certs",
			env: map[string]string{
				"ARGUS_SERVICE_DEPLOYED_VERSION_ALLOW_INVALID_CERTS": "bang",
			},
			errRegex: `ARGUS_SERVICE_DEPLOYED_VERSION_ALLOW_INVALID_CERTS: "bang" <invalid>`,
		},
		{
			name: "service.dashboard/valid",
			env: map[string]string{
				"ARGUS_SERVICE_DASHBOARD_AUTO_APPROVE": "true",
			},
			want: &Defaults{
				Service: service.Defaults{
					Dashboard: *test.Must(t, func() (*dashboard.Defaults, error) {
						return dashboard.DecodeDefaults("yaml", []byte("auto_approve: true"))
					}),
				},
			},
		},
		{
			name: "service.dashboard/invalid bool - auto_approve",
			env: map[string]string{
				"ARGUS_SERVICE_DASHBOARD_AUTO_APPROVE": "zap",
			},
			errRegex: `ARGUS_SERVICE_DASHBOARD_AUTO_APPROVE: "zap" <invalid>`,
		},
		{
			name: "notify.discord/valid",
			env: map[string]string{
				"ARGUS_NOTIFY_DISCORD_OPTIONS_DELAY":        "1h",
				"ARGUS_NOTIFY_DISCORD_OPTIONS_MAX_TRIES":    "1",
				"ARGUS_NOTIFY_DISCORD_OPTIONS_MESSAGE":      "bish",
				"ARGUS_NOTIFY_DISCORD_URL_FIELDS_TOKEN":     "foo",
				"ARGUS_NOTIFY_DISCORD_URL_FIELDS_WEBHOOKID": "bar",
				"ARGUS_NOTIFY_DISCORD_PARAMS_AVATAR":        ":argus:",
				"ARGUS_NOTIFY_DISCORD_PARAMS_COLOR":         "0x50D9ff",
				"ARGUS_NOTIFY_DISCORD_PARAMS_COLORDEBUG":    "0x7b00ab",
				"ARGUS_NOTIFY_DISCORD_PARAMS_COLORERROR":    "0xd60510",
				"ARGUS_NOTIFY_DISCORD_PARAMS_COLORINFO":     "0x2488ff",
				"ARGUS_NOTIFY_DISCORD_PARAMS_COLORWARN":     "0xffc441",
				"ARGUS_NOTIFY_DISCORD_PARAMS_JSON":          "no",
				"ARGUS_NOTIFY_DISCORD_PARAMS_SPLITLINES":    "yes",
				"ARGUS_NOTIFY_DISCORD_PARAMS_TITLE":         "something",
				"ARGUS_NOTIFY_DISCORD_PARAMS_USERNAME":      "test",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"discord": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "1h",
							"max_tries": "1",
							"message":   "bish",
						},
						map[string]string{
							"token":     "foo",
							"webhookid": "bar",
						},
						map[string]string{
							"avatar":     ":argus:",
							"color":      "0x50D9ff",
							"colordebug": "0x7b00ab",
							"colorerror": "0xd60510",
							"colorinfo":  "0x2488ff",
							"colorwarn":  "0xffc441",
							"json":       "no",
							"splitlines": "yes",
							"title":      "something",
							"username":   "test",
						},
					),
				},
			},
		},
		{
			name: "notify.discord/invalid options.delay",
			env: map[string]string{
				"ARGUS_NOTIFY_DISCORD_OPTIONS_DELAY": "foo",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"discord": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay": "foo",
						},
						nil, nil,
					),
				},
			},
			errRegex: `ARGUS_NOTIFY_DISCORD_OPTIONS_DELAY: "foo" <invalid> .+`,
		},
		{
			name: "notify.gotify",
			env: map[string]string{
				"ARGUS_NOTIFY_GOTIFY_OPTIONS_DELAY":     "3s",
				"ARGUS_NOTIFY_GOTIFY_OPTIONS_MAX_TRIES": "3",
				"ARGUS_NOTIFY_GOTIFY_OPTIONS_MESSAGE":   "shazam",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_HOST":   "gotify.example.com",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_PATH":   "gotify",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_PORT":   "443",
				"ARGUS_NOTIFY_GOTIFY_URL_FIELDS_TOKEN":  "SuperSecretToken",
				"ARGUS_NOTIFY_GOTIFY_PARAMS_DISABLETLS": "no",
				"ARGUS_NOTIFY_GOTIFY_PARAMS_PRIORITY":   "0",
				"ARGUS_NOTIFY_GOTIFY_PARAMS_TITLE":      "Argus Gotify Notification",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"gotify": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "3s",
							"max_tries": "3",
							"message":   "shazam",
						},
						map[string]string{
							"host":  "gotify.example.com",
							"path":  "gotify",
							"port":  "443",
							"token": "SuperSecretToken",
						},
						map[string]string{
							"disabletls": "no",
							"priority":   "0",
							"title":      "Argus Gotify Notification",
						},
					),
				},
			},
		},
		{
			name: "notify.googlechat",
			env: map[string]string{
				"ARGUS_NOTIFY_GOOGLECHAT_OPTIONS_DELAY":     "4h",
				"ARGUS_NOTIFY_GOOGLECHAT_OPTIONS_MAX_TRIES": "4",
				"ARGUS_NOTIFY_GOOGLECHAT_OPTIONS_MESSAGE":   "whoosh",
				"ARGUS_NOTIFY_GOOGLECHAT_URL_FIELDS_RAW":    "chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"googlechat": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "4h",
							"max_tries": "4",
							"message":   "whoosh",
						},
						map[string]string{
							"raw": "chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz",
						},
						nil,
					),
				},
			},
		},
		{
			name: "notify.ifttt",
			env: map[string]string{
				"ARGUS_NOTIFY_IFTTT_OPTIONS_DELAY":            "5m",
				"ARGUS_NOTIFY_IFTTT_OPTIONS_MAX_TRIES":        "5",
				"ARGUS_NOTIFY_IFTTT_OPTIONS_MESSAGE":          "pow",
				"ARGUS_NOTIFY_IFTTT_URL_FIELDS_WEBHOOKID":     "secretWHID",
				"ARGUS_NOTIFY_IFTTT_PARAMS_EVENTS":            "event1,event2",
				"ARGUS_NOTIFY_IFTTT_PARAMS_TITLE":             "Argus IFTTT Notification",
				"ARGUS_NOTIFY_IFTTT_PARAMS_USEMESSAGEASVALUE": "2",
				"ARGUS_NOTIFY_IFTTT_PARAMS_USETITLEASVALUE":   "0",
				"ARGUS_NOTIFY_IFTTT_PARAMS_VALUE1":            "bish",
				"ARGUS_NOTIFY_IFTTT_PARAMS_VALUE2":            "bash",
				"ARGUS_NOTIFY_IFTTT_PARAMS_VALUE3":            "bosh",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"ifttt": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "5m",
							"max_tries": "5",
							"message":   "pow",
						},
						map[string]string{
							"webhookid": "secretWHID",
						},
						map[string]string{
							"events":            "event1,event2",
							"title":             "Argus IFTTT Notification",
							"usemessageasvalue": "2",
							"usetitleasvalue":   "0",
							"value1":            "bish",
							"value2":            "bash",
							"value3":            "bosh",
						},
					),
				},
			},
		},
		{
			name: "notify.join",
			env: map[string]string{
				"ARGUS_NOTIFY_JOIN_OPTIONS_DELAY":     "6s",
				"ARGUS_NOTIFY_JOIN_OPTIONS_MAX_TRIES": "6",
				"ARGUS_NOTIFY_JOIN_OPTIONS_MESSAGE":   "pew",
				"ARGUS_NOTIFY_JOIN_URL_FIELDS_APIKEY": "apiKey",
				"ARGUS_NOTIFY_JOIN_PARAMS_DEVICES":    "device1,device2",
				"ARGUS_NOTIFY_JOIN_PARAMS_ICON":       "example.com/icon.png",
				"ARGUS_NOTIFY_JOIN_PARAMS_TITLE":      "Argus Join Notification",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"join": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "6s",
							"max_tries": "6",
							"message":   "pew",
						},
						map[string]string{
							"apikey": "apiKey",
						},
						map[string]string{
							"devices": "device1,device2",
							"icon":    "example.com/icon.png",
							"title":   "Argus Join Notification",
						},
					),
				},
			},
		},
		{
			name: "notify.mattermost",
			env: map[string]string{
				"ARGUS_NOTIFY_MATTERMOST_OPTIONS_DELAY":       "7h",
				"ARGUS_NOTIFY_MATTERMOST_OPTIONS_MAX_TRIES":   "7",
				"ARGUS_NOTIFY_MATTERMOST_OPTIONS_MESSAGE":     "ping",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_CHANNEL":  "argus",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_HOST":     "mattermost.example.com",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_PATH":     "mattermost",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_PORT":     "443",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_TOKEN":    "mattermostToken",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_USERNAME": "Argus",
				"ARGUS_NOTIFY_MATTERMOST_PARAMS_ICON":         ":argus:",
				"ARGUS_NOTIFY_MATTERMOST_PARAMS_TITLE":        "Argus Mattermost Notification",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"mattermost": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "7h",
							"max_tries": "7",
							"message":   "ping",
						},
						map[string]string{
							"channel":  "argus",
							"host":     "mattermost.example.com",
							"path":     "mattermost",
							"port":     "443",
							"token":    "mattermostToken",
							"username": "Argus",
						},
						map[string]string{
							"icon":  ":argus:",
							"title": "Argus Mattermost Notification",
						},
					),
				},
			},
		},
		{
			name: "notify.matrix",
			env: map[string]string{
				"ARGUS_NOTIFY_MATRIX_OPTIONS_DELAY":       "8m",
				"ARGUS_NOTIFY_MATRIX_OPTIONS_MAX_TRIES":   "8",
				"ARGUS_NOTIFY_MATRIX_OPTIONS_MESSAGE":     "pong",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_HOST":     "matrix.example.com",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_PASSWORD": "matrixPassword",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_PATH":     "matrix",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_PORT":     "443",
				"ARGUS_NOTIFY_MATRIX_URL_FIELDS_USER":     "argus",
				"ARGUS_NOTIFY_MATRIX_PARAMS_DISABLETLS":   "no",
				"ARGUS_NOTIFY_MATRIX_PARAMS_ROOMS":        "room1,room2",
				"ARGUS_NOTIFY_MATRIX_PARAMS_TITLE":        "Argus Matrix Notification",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"matrix": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "8m",
							"max_tries": "8",
							"message":   "pong",
						},
						map[string]string{
							"host":     "matrix.example.com",
							"password": "matrixPassword",
							"path":     "matrix",
							"port":     "443",
							"user":     "argus",
						},
						map[string]string{
							"disabletls": "no",
							"rooms":      "room1,room2",
							"title":      "Argus Matrix Notification",
						},
					),
				},
			},
		},
		{
			name: "notify.opsgenie",
			env: map[string]string{
				"ARGUS_NOTIFY_OPSGENIE_OPTIONS_DELAY":      "9s",
				"ARGUS_NOTIFY_OPSGENIE_OPTIONS_MAX_TRIES":  "9",
				"ARGUS_NOTIFY_OPSGENIE_OPTIONS_MESSAGE":    "pang",
				"ARGUS_NOTIFY_OPSGENIE_URL_FIELDS_APIKEY":  "opsGenieApiKey",
				"ARGUS_NOTIFY_OPSGENIE_URL_FIELDS_HOST":    "opsgenie.example.com",
				"ARGUS_NOTIFY_OPSGENIE_URL_FIELDS_PATH":    "opsgenie",
				"ARGUS_NOTIFY_OPSGENIE_URL_FIELDS_PORT":    "443",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_ACTIONS":     "action1,action2",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_ALIAS":       "argus",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_DESCRIPTION": "Argus OpsGenie DESC",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_DETAILS":     "foo=bar",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_ENTITY":      "argus",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_NOTE":        "testing OpsGenie",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_PRIORITY":    "P1",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_RESPONDERS":  "responder1,responder2",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_SOURCE":      "argus",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_TAGS":        "tag1,tag2",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_TITLE":       "Argus OpsGenie Notification",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_USER":        "argus",
				"ARGUS_NOTIFY_OPSGENIE_PARAMS_VISIBLETO":   "visible1,visible2",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"opsgenie": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "9s",
							"max_tries": "9",
							"message":   "pang",
						},
						map[string]string{
							"apikey": "opsGenieApiKey",
							"host":   "opsgenie.example.com",
							"path":   "opsgenie",
							"port":   "443",
						},
						map[string]string{
							"actions":     "action1,action2",
							"alias":       "argus",
							"description": "Argus OpsGenie DESC",
							"details":     "foo=bar",
							"entity":      "argus",
							"note":        "testing OpsGenie",
							"priority":    "P1",
							"responders":  "responder1,responder2",
							"source":      "argus",
							"tags":        "tag1,tag2",
							"title":       "Argus OpsGenie Notification",
							"user":        "argus",
							"visibleto":   "visible1,visible2",
						},
					),
				},
			},
		},
		{
			name: "notify.pushbullet",
			env: map[string]string{
				"ARGUS_NOTIFY_PUSHBULLET_OPTIONS_DELAY":      "10h",
				"ARGUS_NOTIFY_PUSHBULLET_OPTIONS_MAX_TRIES":  "10",
				"ARGUS_NOTIFY_PUSHBULLET_OPTIONS_MESSAGE":    "pong",
				"ARGUS_NOTIFY_PUSHBULLET_URL_FIELDS_TARGETS": "target1,target2",
				"ARGUS_NOTIFY_PUSHBULLET_URL_FIELDS_TOKEN":   "pushbulletToken",
				"ARGUS_NOTIFY_PUSHBULLET_PARAMS_TITLE":       "Argus Pushbullet Notification",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"pushbullet": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "10h",
							"max_tries": "10",
							"message":   "pong",
						},
						map[string]string{
							"targets": "target1,target2",
							"token":   "pushbulletToken",
						},
						map[string]string{
							"title": "Argus Pushbullet Notification",
						},
					),
				},
			},
		},
		{
			name: "notify.pushover",
			env: map[string]string{
				"ARGUS_NOTIFY_PUSHOVER_OPTIONS_DELAY":     "11m",
				"ARGUS_NOTIFY_PUSHOVER_OPTIONS_MAX_TRIES": "11",
				"ARGUS_NOTIFY_PUSHOVER_OPTIONS_MESSAGE":   "pong",
				"ARGUS_NOTIFY_PUSHOVER_URL_FIELDS_TOKEN":  "pushoverToken",
				"ARGUS_NOTIFY_PUSHOVER_URL_FIELDS_USER":   "pushoverUser",
				"ARGUS_NOTIFY_PUSHOVER_PARAMS_DEVICES":    "device1,device2",
				"ARGUS_NOTIFY_PUSHOVER_PARAMS_PRIORITY":   "0",
				"ARGUS_NOTIFY_PUSHOVER_PARAMS_TITLE":      "Argus Pushbullet Notification",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"pushover": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "11m",
							"max_tries": "11",
							"message":   "pong",
						},
						map[string]string{
							"token": "pushoverToken",
							"user":  "pushoverUser",
						},
						map[string]string{
							"devices":  "device1,device2",
							"priority": "0",
							"title":    "Argus Pushbullet Notification",
						},
					),
				},
			},
		},
		{
			name: "notify.rocketchat",
			env: map[string]string{
				"ARGUS_NOTIFY_ROCKETCHAT_OPTIONS_DELAY":       "12s",
				"ARGUS_NOTIFY_ROCKETCHAT_OPTIONS_MAX_TRIES":   "12",
				"ARGUS_NOTIFY_ROCKETCHAT_OPTIONS_MESSAGE":     "pong",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_CHANNEL":  "rocketchatChannel",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_HOST":     "rocketchat.example.com",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_PORT":     "443",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_PATH":     "rocketchat",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_TOKENA":   "FIRST_token",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_TOKENB":   "SECOND_token",
				"ARGUS_NOTIFY_ROCKETCHAT_URL_FIELDS_USERNAME": "rocketchatUser",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"rocketchat": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "12s",
							"max_tries": "12",
							"message":   "pong",
						},
						map[string]string{
							"channel":  "rocketchatChannel",
							"host":     "rocketchat.example.com",
							"path":     "rocketchat",
							"port":     "443",
							"tokena":   "FIRST_token",
							"tokenb":   "SECOND_token",
							"username": "rocketchatUser",
						},
						nil,
					),
				},
			},
		},
		{
			name: "notify.slack",
			env: map[string]string{
				"ARGUS_NOTIFY_SLACK_OPTIONS_DELAY":      "13h",
				"ARGUS_NOTIFY_SLACK_OPTIONS_MAX_TRIES":  "13",
				"ARGUS_NOTIFY_SLACK_OPTIONS_MESSAGE":    "slung",
				"ARGUS_NOTIFY_SLACK_URL_FIELDS_TOKEN":   "slackToken",
				"ARGUS_NOTIFY_SLACK_URL_FIELDS_CHANNEL": "somewhere",
				"ARGUS_NOTIFY_SLACK_PARAMS_BOTNAME":     "Argus",
				"ARGUS_NOTIFY_SLACK_PARAMS_COLOR":       "#ff8000",
				"ARGUS_NOTIFY_SLACK_PARAMS_ICON":        ":ghost:",
				"ARGUS_NOTIFY_SLACK_PARAMS_THREADTS":    "1234567890.123456",
				"ARGUS_NOTIFY_SLACK_PARAMS_TITLE":       "Argus Slack Notification",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"slack": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "13h",
							"max_tries": "13",
							"message":   "slung",
						},
						map[string]string{
							"channel": "somewhere",
							"token":   "slackToken",
						},
						map[string]string{
							"botname":  "Argus",
							"color":    "%23ff8000",
							"icon":     ":ghost:",
							"threadts": "1234567890.123456",
							"title":    "Argus Slack Notification",
						},
					),
				},
			},
		},
		{
			name: "notify.smtp",
			env: map[string]string{
				"ARGUS_NOTIFY_SMTP_OPTIONS_DELAY":       "2m",
				"ARGUS_NOTIFY_SMTP_OPTIONS_MAX_TRIES":   "2",
				"ARGUS_NOTIFY_SMTP_OPTIONS_MESSAGE":     "bing",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_HOST":     "smtp.example.com",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_PASSWORD": "secret",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_PORT":     "25",
				"ARGUS_NOTIFY_SMTP_URL_FIELDS_USERNAME": "user",
				"ARGUS_NOTIFY_SMTP_PARAMS_AUTH":         "Unknown",
				"ARGUS_NOTIFY_SMTP_PARAMS_CLIENTHOST":   "localhost",
				"ARGUS_NOTIFY_SMTP_PARAMS_ENCRYPTION":   "auto",
				"ARGUS_NOTIFY_SMTP_PARAMS_FROMADDRESS":  "me@example.com",
				"ARGUS_NOTIFY_SMTP_PARAMS_FROMNAME":     "someone",
				"ARGUS_NOTIFY_SMTP_PARAMS_SUBJECT":      "Argus SMTP Notification",
				"ARGUS_NOTIFY_SMTP_PARAMS_TOADDRESSES":  "you@somewhere.com",
				"ARGUS_NOTIFY_SMTP_PARAMS_USEHTML":      "no",
				"ARGUS_NOTIFY_SMTP_PARAMS_USESTARTTLS":  "yes",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"smtp": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "2m",
							"max_tries": "2",
							"message":   "bing",
						},
						map[string]string{
							"host":     "smtp.example.com",
							"password": "secret",
							"port":     "25",
							"username": "user",
						},
						map[string]string{
							"auth":        "Unknown",
							"clienthost":  "localhost",
							"encryption":  "Auto",
							"fromaddress": "me@example.com",
							"fromname":    "someone",
							"subject":     "Argus SMTP Notification",
							"toaddresses": "you@somewhere.com",
							"usehtml":     "no",
							"usestarttls": "yes",
						},
					),
				},
			},
		},
		{
			name: "notify.teams",
			env: map[string]string{
				"ARGUS_NOTIFY_TEAMS_OPTIONS_DELAY":         "14m",
				"ARGUS_NOTIFY_TEAMS_OPTIONS_MAX_TRIES":     "14",
				"ARGUS_NOTIFY_TEAMS_OPTIONS_MESSAGE":       "hi",
				"ARGUS_NOTIFY_TEAMS_URL_FIELDS_GROUP":      "teamsGroup",
				"ARGUS_NOTIFY_TEAMS_URL_FIELDS_TENANT":     "tenant",
				"ARGUS_NOTIFY_TEAMS_URL_FIELDS_ALTID":      "otherID?",
				"ARGUS_NOTIFY_TEAMS_URL_FIELDS_GROUPOWNER": "owner",
				"ARGUS_NOTIFY_TEAMS_PARAMS_COLOR":          "#ff8000",
				"ARGUS_NOTIFY_TEAMS_PARAMS_HOST":           "teams.example.com",
				"ARGUS_NOTIFY_TEAMS_PARAMS_TITLE":          "Argus Teams Notification",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"teams": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "14m",
							"message":   "hi",
							"max_tries": "14",
						},
						map[string]string{
							"altid":      "otherID?",
							"group":      "teamsGroup",
							"groupowner": "owner",
							"tenant":     "tenant",
						},
						map[string]string{
							"color": "#ff8000",
							"host":  "teams.example.com",
							"title": "Argus Teams Notification",
						},
					),
				},
			},
		},
		{
			name: "notify.telegram",
			env: map[string]string{
				"ARGUS_NOTIFY_TELEGRAM_OPTIONS_DELAY":       "15s",
				"ARGUS_NOTIFY_TELEGRAM_OPTIONS_MAX_TRIES":   "15",
				"ARGUS_NOTIFY_TELEGRAM_OPTIONS_MESSAGE":     "tong",
				"ARGUS_NOTIFY_TELEGRAM_URL_FIELDS_TOKEN":    "telegramToken",
				"ARGUS_NOTIFY_TELEGRAM_PARAMS_CHATS":        "chat1,chat2",
				"ARGUS_NOTIFY_TELEGRAM_PARAMS_NOTIFICATION": "yes",
				"ARGUS_NOTIFY_TELEGRAM_PARAMS_PARSEMODE":    "None",
				"ARGUS_NOTIFY_TELEGRAM_PARAMS_PREVIEW":      "yes",
				"ARGUS_NOTIFY_TELEGRAM_PARAMS_TITLE":        "Argus Telegram Notification",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"telegram": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "15s",
							"max_tries": "15",
							"message":   "tong",
						},
						map[string]string{
							"token": "telegramToken",
						},
						map[string]string{
							"chats":        "chat1,chat2",
							"notification": "yes",
							"parsemode":    "None",
							"preview":      "yes",
							"title":        "Argus Telegram Notification",
						},
					),
				},
			},
		},
		{
			name: "notify.zulip",
			env: map[string]string{
				"ARGUS_NOTIFY_ZULIP_OPTIONS_DELAY":      "16h",
				"ARGUS_NOTIFY_ZULIP_OPTIONS_MAX_TRIES":  "16",
				"ARGUS_NOTIFY_ZULIP_OPTIONS_MESSAGE":    "hiya",
				"ARGUS_NOTIFY_ZULIP_URL_FIELDS_BOTMAIL": "botmail",
				"ARGUS_NOTIFY_ZULIP_URL_FIELDS_BOTKEY":  "botkey",
				"ARGUS_NOTIFY_ZULIP_URL_FIELDS_HOST":    "zulip.example.com",
				"ARGUS_NOTIFY_ZULIP_URL_FIELDS_PORT":    "1234",
				"ARGUS_NOTIFY_ZULIP_URL_FIELDS_PATH":    "zulip",
				"ARGUS_NOTIFY_ZULIP_PARAMS_STREAM":      "stream",
				"ARGUS_NOTIFY_ZULIP_PARAMS_TOPIC":       "topic",
			},
			want: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"zulip": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay":     "16h",
							"max_tries": "16",
							"message":   "hiya",
						},
						map[string]string{
							"botkey":  "botkey",
							"botmail": "botmail",
							"host":    "zulip.example.com",
							"path":    "zulip",
							"port":    "1234",
						},
						map[string]string{
							"stream": "stream",
							"topic":  "topic",
						},
					),
				},
			},
		},
		{
			name: "webhook/valid",
			env: map[string]string{
				"ARGUS_WEBHOOK_ALLOW_INVALID_CERTS": "false",
				"ARGUS_WEBHOOK_DELAY":               "99s",
				"ARGUS_WEBHOOK_DESIRED_STATUS_CODE": "201",
				"ARGUS_WEBHOOK_MAX_TRIES":           "88",
				"ARGUS_WEBHOOK_SILENT_FAILS":        "true",
				"ARGUS_WEBHOOK_TYPE":                "github",
				"ARGUS_WEBHOOK_URL":                 "https://webhook.example.com",
			},
			want: &Defaults{
				WebHook: *test.Must(t, func() (*webhook.Defaults, error) {
					return webhook.DecodeDefaults(
						"yaml", []byte(test.TrimYAML(`
							allow_invalid_certs: false
							delay: 99s
							desired_status_code: 201
							max_tries: 88
							silent_fails: true
							type: github
							url: https://webhook.example.com
						`)),
					)
				}),
			},
		},
		{
			name: "webhook/invalid str, type",
			env: map[string]string{
				"ARGUS_WEBHOOK_TYPE": "pizza",
			},
			errRegex: `ARGUS_WEBHOOK_TYPE: "pizza" <invalid>`,
		},
		{
			name: "webhook/invalid time.duration, delay",
			env: map[string]string{
				"ARGUS_WEBHOOK_DELAY": "pasta",
			},
			errRegex: `ARGUS_WEBHOOK_DELAY: "[^"]+" <invalid>`,
		},
		{
			name: "webhook/invalid uint, max_tries",
			env: map[string]string{
				"ARGUS_WEBHOOK_MAX_TRIES": "-1",
			},
			errRegex: `ARGUS_WEBHOOK_MAX_TRIES: "-1" <invalid>`,
		},
		{
			name: "webhook/invalid bool/allow_invalid_certs",
			env: map[string]string{
				"ARGUS_WEBHOOK_ALLOW_INVALID_CERTS": "foo",
			},
			errRegex: `ARGUS_WEBHOOK_ALLOW_INVALID_CERTS: "foo" <invalid>`,
		},
		{
			name: "webhook/invalid int, desired_status_code",
			env: map[string]string{
				"ARGUS_WEBHOOK_DESIRED_STATUS_CODE": "okay",
			},
			errRegex: `ARGUS_WEBHOOK_DESIRED_STATUS_CODE: "okay" <invalid>`,
		},
		{
			name: "webhook/invalid bool/silent_fails",
			env: map[string]string{
				"ARGUS_WEBHOOK_SILENT_FAILS": "bar",
			},
			errRegex: `ARGUS_WEBHOOK_SILENT_FAILS: "bar" <invalid>`,
		},
		{
			name: "multiple fails",
			env: map[string]string{
				"ARGUS_NOTIFY_DISCORD_OPTIONS_DELAY":               "foo",
				"ARGUS_NOTIFY_SLACK_OPTIONS_DELAY":                 "bar",
				"ARGUS_NOTIFY_TEAMS_OPTIONS_DELAY":                 "baz",
				"ARGUS_WEBHOOK_DELAY":                              "pasta",
				"ARGUS_WEBHOOK_TYPE":                               "pizza",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE": "pizza",
			},
			errRegex: test.TrimYAML(`
				ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE: "pizza" <invalid> .+
				ARGUS_NOTIFY_DISCORD_OPTIONS_DELAY: "foo" <invalid> .+
				ARGUS_NOTIFY_SLACK_OPTIONS_DELAY: "bar" <invalid> .+
				ARGUS_NOTIFY_TEAMS_OPTIONS_DELAY: "baz" <invalid> .+
				ARGUS_WEBHOOK_TYPE: "pizza" <invalid> .+
				ARGUS_WEBHOOK_DELAY: "pasta" <invalid> .+`,
			),
		},
		{
			name: "no env vars",
			want: &Defaults{},
		},
		{
			name: "no 'ARGUS_' env vars",
			env: map[string]string{
				"NOT_ARGUS": "foo",
			},
			want: &Defaults{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing some env vars.
			releaseStdout := test.CaptureLog(t, logx.Default())

			defaults := Defaults{
				Service: service.Defaults{
					LatestVersion: lvbase.Defaults{
						Require: filter.RequireDefaults{
							Docker: *test.Must(t, func() (*docker.Defaults, error) {
								return docker.DecodeDefaults("json", []byte("{}"), nil)
							}),
						},
					},
					DeployedVersionLookup: dvbase.Defaults{},
				},
			}
			if tc.want == nil {
				tc.want = &Defaults{
					Notify: shoutrrr.ShoutrrrsDefaults{},
				}
			} else {
				if tc.want.Notify != nil {
					defaults.Notify = shoutrrr.ShoutrrrsDefaults{}
					for notifyType := range unmodifiedDefaults.Notify {
						defaults.Notify[notifyType] = shoutrrr.NewDefaults(
							"",
							nil, nil, nil,
						)

						defaults.Notify[notifyType].InitMaps()
						if tc.want.Notify[notifyType] == nil {
							tc.want.Notify[notifyType] = shoutrrr.NewDefaults(
								"",
								nil, nil, nil,
							)
							tc.want.Notify[notifyType].InitMaps()
						}
					}
				}
			}
			test.SetEnv(t, tc.env)
			wantOk := tc.errRegex == ""

			resultChannel := make(chan bool, 1)
			// WHEN: CheckValues is called on it.
			go func() { resultChannel <- defaults.MapEnvToStruct() }()

			prefix := fmt.Sprintf("%s\nDefaults.MapEnvToStruct()", packageName)

			// THEN: the OK value is as expected.
			if err := test.AssertChannelBool(
				t,
				wantOk,
				resultChannel,
				logx.ExitCodeChannel(),
				releaseStdout,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}

			// AND: any error is as expected.
			stdout := releaseStdout()
			if !wantOk {
				return
			}
			if !util.RegexCheck(tc.errRegex, stdout) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.errRegex,
				)
			}

			// AND: the defaults are set to the appropriate env vars.
			wantStr := tc.want.String("")
			if got := defaults.String(""); got != wantStr {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, wantStr,
				)
			}
		})
	}
}

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN: defaults with a test of invalid vars.
	var defaults Defaults
	defaults.Default()
	tests := []struct {
		name     string
		input    *Defaults
		errRegex string
		changed  bool
	}{
		{
			name: "Service.Interval",
			input: &Defaults{
				Service: service.Defaults{
					Options: *test.Must(t, func() (*opt.Defaults, error) {
						return opt.DecodeDefaults("yaml", []byte("interval: 10x"))
					}),
				},
			},
			errRegex: test.TrimYAML(`
				^service:
					options:
						interval: "10x" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "Service.LatestVersion.Require.Docker.Type",
			input: &Defaults{
				Service: service.Defaults{
					LatestVersion: lvbase.Defaults{
						Require: filter.RequireDefaults{
							Docker: *test.Must(t, func() (*docker.Defaults, error) {
								return docker.DecodeDefaults(
									"yaml", []byte("type: pizza"),
									nil,
								)
							}),
						},
					},
				},
			},
			errRegex: test.TrimYAML(`
				^service:
					latest_version:
						require:
							docker:
								type: "pizza" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "Service.Interval + Service.DeployedVersionLookup.Regex",
			input: &Defaults{
				Service: service.Defaults{
					Options: *test.Must(t, func() (*opt.Defaults, error) {
						return opt.DecodeDefaults("yaml", []byte("interval: 10x"))
					}),
					LatestVersion: lvbase.Defaults{
						Require: filter.RequireDefaults{
							Docker: *test.Must(t, func() (*docker.Defaults, error) {
								return docker.DecodeDefaults(
									"yaml", []byte("type: pizza"),
									nil,
								)
							}),
						},
					},
				},
			},
			errRegex: test.TrimYAML(`
				^service:
					options:
						interval: "10x" <invalid>.*
					latest_version:
						require:
							docker:
								type: "pizza" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "Notify changed",
			input: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"foo": shoutrrr.NewDefaults(
						"generic",
						nil,
						map[string]string{
							"host":           "x",
							"secret":         "y",
							"custom_headers": `{"foo": "bar"}`,
						},
						nil,
					),
				},
			},
			changed: true,
		},
		{
			name: "Notify.x.Delay",
			input: &Defaults{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"slack": shoutrrr.NewDefaults(
						"",
						map[string]string{
							"delay": "10x",
						},
						nil, nil,
					),
				},
			},
			errRegex: test.TrimYAML(`
				^notify:
					slack:
						options:
							delay: "10x" <invalid>.*$`,
			),
			changed: false,
		},
		{
			name: "WebHook changed",
			input: &Defaults{
				WebHook: webhook.Defaults{
					Base: webhook.Base{
						Type:   "github",
						URL:    "example.com",
						Secret: "Argus",
						// CustomHeaders -> Headers.
						CustomHeaders: webhook.Headers{
							{Key: "foo", Value: "bar"},
						},
					},
				},
			},
			changed: true,
		},
		{
			name: "WebHook.Delay",
			input: &Defaults{
				WebHook: *test.Must(t, func() (*webhook.Defaults, error) {
					return webhook.DecodeDefaults("yaml", []byte("delay: 10x"))
				}),
			},
			errRegex: test.TrimYAML(`
				^webhook:
					delay: "10x" <invalid>.*$`,
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

func TestDefaults_Print(t *testing.T) {
	// GIVEN: a set of Defaults.
	var defaults Defaults
	defaults.Default()
	tests := []struct {
		name  string
		input *Defaults
		want  string
	}{
		{
			name:  "unmodified hard defaults",
			input: &defaults,
			want:  hardDefaultsStr,
		},
		{
			name:  "empty defaults",
			input: &Defaults{},
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout(t)

			// WHEN: Print is called.
			tc.input.Print("")

			// THEN: the expected number of lines are printed.
			stdout := releaseStdout()
			if stdout != tc.want {
				t.Errorf(
					"%s\nDefaults.Print() stdout mismatch\ngot:  %q\nwant: %q",
					packageName, stdout, tc.want,
				)
			}
		})
	}
}

var hardDefaultsStr = test.TrimYAML(`
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
					usemessageasvalue: '2'
					usetitleasvalue: '0'
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
`)
