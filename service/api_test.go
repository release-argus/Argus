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

//go:build integration

package service

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	dvweb "github.com/release-argus/Argus/service/deployed_version/types/web"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	lvforgejo "github.com/release-argus/Argus/service/latest_version/types/forgejo"
	lvgithub "github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/webhook"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestSecretRefs_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		errRegex string
		want     string
	}{
		{
			name:     "invalid JSON",
			data:     `{`,
			errRegex: `unexpected`,
		},
		{
			name:     "static fields error",
			data:     `{"id": 123}`,
			errRegex: `^json: .*unmarshal.* number.*$`,
		},
		{
			name: "static fields",
			data: test.TrimJSON(`{
				"id": "123",
				"latest_version": {
					"headers": [
						{"old_index": 0},
						{"old_index": 1}
					]
				},
				"deployed_version": {
					"headers": [
						{"old_index": 4},
						{"old_index": 3}
					]
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				id: '123'
				latest_version:
					headers:
						- old_index: 0
						- old_index: 1
				deployed_version:
					headers:
						- old_index: 4
						- old_index: 3
			`),
		},
		{
			name: "notify array -> map",
			data: test.TrimJSON(`{
				"notify": [
					{"name": "foo", "old_index": "foo"},
					{"name": "bar", "old_index": "baz"},
					{"name": "baz", "old_index": "bar"}
				]
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				id: ''
				notify:
					bar:
					  name: bar
					  old_index: baz
					baz:
					  name: baz
					  old_index: bar
					foo:
					  name: foo
					  old_index: foo
			`),
		},
		{
			name: "webhook array -> map",
			data: test.TrimJSON(`{
				"webhook": [
					{"name": "foo", "old_index": "foo"},
					{"name": "bar", "old_index": "baz", "headers": [{"old_index": 1}, {"old_index": 0}]},
					{"name": "baz", "old_index": "bar"}
				]
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				id: ''
				webhook:
					bar:
					  name: bar
					  old_index: baz
						headers:
							- old_index: 1
							- old_index: 0
					baz:
					  name: baz
					  old_index: bar
					foo:
					  name: foo
					  old_index: foo
			`),
		},
		{
			name: "filled",
			data: test.TrimJSON(`{
				"notify": [
					{"name": "foo", "old_index": "foo"},
					{"name": "bar", "old_index": "baz"},
					{"name": "baz", "old_index": "bar"}
				],
				"webhook": [
					{"name": "foo", "old_index": "foo"},
					{"name": "bar", "old_index": "baz", "headers": [{"old_index": 1}, {"old_index": 0}]},
					{"name": "baz", "old_index": "bar"}
				]
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				id: ''
				notify:
					bar:
					  name: bar
					  old_index: baz
					baz:
					  name: baz
					  old_index: bar
					foo:
					  name: foo
					  old_index: foo
				webhook:
					bar:
					  name: bar
					  old_index: baz
						headers:
							- old_index: 1
							- old_index: 0
					baz:
					  name: baz
					  old_index: bar
					foo:
					  name: foo
					  old_index: foo
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*secretRefs, error) {
					var refs secretRefs
					err := decode.Unmarshal(format, data, &refs)
					return &refs, err
				},
				"json", tc.data,
				func(v *secretRefs) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				"secretRefs.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestFromPayload(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	type fromDefaults struct {
		command bool
		notify  bool
		webhook bool
	}

	tests := []struct {
		name       string
		oldService *Service
		payload    string

		svcCfg    DefaultsConfig
		notifyCfg shoutrrr.Config
		whCfg     webhook.Config

		want             *Service
		wantFromDefaults fromDefaults
		errRegex         string
	}{
		{
			name:    "empty payload",
			payload: "",
			errRegex: test.TrimYAML(`
				^unmarshal service payload:
					jsontext:
						unexpected EOF$`,
			),
		},
		{
			name:    "invalid payload",
			payload: strings.Repeat("a", 1048577),
			errRegex: test.TrimYAML(`
				^unmarshal service payload:
					jsontext:
						invalid character 'a' .*$`,
			),
		},
		{
			name:     "invalid Service payload",
			payload:  `{"name": false}`,
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "invalid SecretRefs payload",
			payload: test.TrimJSON(`{
				"webhook": [
					{"name": "foo", "old_index": false}
				]
			}`),
			errRegex: `json: .*unmarshal`,
		},
		{
			name: "active=true becomes nil",
			payload: `{
				"options": {
					"active": true
				},
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						`)),
					"active=true becomes nil",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "active=nil stays nil",
			payload: `{
				"options": {
					"active": null
				},
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						`)),
					"active=nil stays nil",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "active=false stays false",
			payload: `{
				"options": {
					"active": false
				},
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							active: false
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						`)),
					"active=false stays false",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "Require.Docker removed if no image:tag",
			payload: `{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `",
					"require": {
						"docker": {
							"type": "ghcr"
						}
					}
				}
			}`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						`)),
					"Require.Docker removed if no image:tag",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "Require.Docker stays if have Type&Image&Tag",
			payload: `{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `",
					"require": {
						"docker": {
							"type": "ghcr",
							"image": "release-argus-argus",
							"tag": "latest"
						}
					}
				}
			}`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							require:
								docker:
									type: ghcr
									image: release-argus-argus
									tag: latest
						`)),
					"Require.Docker stays if have Type&Image&Tag",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "Give LatestVersion secrets",
			payload: `{
				"latest_version": {
					"type": "github",
					"access_token": "` + util.SecretValue + `",
					"url": "` + test.ArgusGitHubRepo + `",
					"require": {
						"docker": {
							"type": "ghcr",
							"image": "` + test.ArgusDockerGHCRRepo + `",
							"tag": "{{ version }}",
							"auth": {
								"token": "` + util.SecretValue + `"
							}
						}
					}
				}
			}`,
			svcCfg: DefaultsConfig{
				Hard: &Defaults{
					LatestVersion: latestver.Defaults{
						Type: lvgithub.Type,
						Common: lvbase.Defaults{
							Require: filter.RequireDefaults{
								Docker: docker.Defaults{
									Type: "ghcr",
								},
							},
						},
					},
				},
			},
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: `+test.ArgusDockerGHCRRepo+`
									tag: "{{ version }}"
									auth:
										token: anotherToken
						`)),
					"Give LatestVersion secrets",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: `+test.ArgusDockerGHCRRepo+`
									tag: "{{ version }}"
									auth:
										token: anotherToken
						`)),
					"Give LatestVersion secrets",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "Give DeployedVersion secrets",
			payload: `{
				"deployed_version": {
					"type": "url",
					"url": "` + test.LookupPlain.URLValid + `",
					"basic_auth": {
						"password": "` + util.SecretValue + `"
					},
					"headers": [
						{
							"key": "X-Foo",
							"value": "` + util.SecretValue + `",
							"old_index": 0
						}
					]
				}
			}`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						deployed_version:
							type: url
							url: `+test.LookupPlain.URLValid+`
							basic_auth:
								password: aPassword
							headers:
								- key: X-Foo
									value: aFoo
						`)),
					"Give DeployedVersion secrets",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						deployed_version:
							type: url
							url: https://example.com
							basic_auth:
								password: aPassword
							headers:
								- key: X-Foo
									value: aFoo
						`)),
					"Give DeployedVersion secrets",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "Give Notify secrets",
			payload: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				},
				"notify": [
					{
						"name": "join",
						"type": "join",
						"url_fields": {
							"apikey": "` + util.SecretValue + `"
						},
						"params": {
							"devices": "` + util.SecretValue + `",
							"icon": "https://example.com/icon.png"
						},
						"old_index": "join-initial"
					},
					{
						"name": "matrix-",
						"type": "matrix",
						"url_fields": {
							"host":    "matrix.example.com",
							"password": "matrixToken"
						},
						"old_index": "matrix-initial"
					},
					{
						"name": "rocketchat",
						"type": "rocketchat",
						"url_fields": {
							"channel": "argus",
							"host":    "https://example.com",
							"tokena": "` + util.SecretValue + `",
							"tokenb": "` + util.SecretValue + `"
						},
						"old_index": "rocketchat-initial"
					},
					{
						"name": "slack",
						"type": "slack",
						"url_fields": {
							"channel": "ABC",
							"token": "` + util.SecretValue + `"
						},
						"old_index": "slack-initial"
					},
					{
						"name": "gotify",
						"type": "gotify",
						"url_fields": {
							"host":  "gotify.example.com",
							"token": "` + util.SecretValue + `"
						},
						"old_index": "gotify-initial"
					},
					{
						"name": "zulip",
						"type": "zulip",
						"url_fields": {
							"botkey": "` + util.SecretValue + `",
							"botmail": "my-bot@example.com",
							"host":    "zulipchat.example.com"
						},
						"old_index": "zulip-initial"
					}
				]
			}`),
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						notify:
							join:
								url_fields:
									apikey: joinApiKey
								params:
									devices: aDevice
									icon: https://example.com/icon.png
							matrix-:
								type: matrix
								url_fields:
									host: matrix.example.com
									password: matrixToken
							rocketchat:
								url_fields:
									channel: argus
									host: example.com
									tokena: rocketchatTokenA
									tokenb: rocketchatTokenB
							slack:
								url_fields:
									channel: ABC
									token: xoxABCDEFGHI-012345678-abcdefghi01234567abcdefghi
							gotify:
								url_fields:
									host: gotify.example.com
									token: gotifyToken
							zulip:
								url_fields:
									botkey: zulipBotKey
									botmail: my-bot%40example.com
									host: zulipchat.example.com
					`)),
					"Give Notify secrets",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: `+test.ArgusDockerGHCRRepo+`
									tag: "{{ version }}"
									auth:
										token: anotherToken

						notify:
							join-initial:
								type: join
								url_fields:
									apikey: joinApiKey
								params:
									devices: aDevice
							matrix-:
								type: matrix
								url_fields:
									host: matrix.example.com
									password: matrixToken
							rocketchat-initial:
								type: rocketchat
								url_fields:
									tokena: rocketchatTokenA
									tokenb: rocketchatTokenB
							slack-initial:
								type: slack
								url_fields:
									token: xoxABCDEFGHI-012345678-abcdefghi01234567abcdefghi
							gotify-initial:
								type: gotify
								url_fields:
									host: gotify.example.com
									token: gotifyToken
							zulip-initial:
								type: zulip
								url_fields:
									botkey: zulipBotKey
					`)),
					"Give Notify secrets",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "Give Webhook secrets",
			payload: `{
				"latest_version": {
					"access_token": "` + util.SecretValue + `",
					"url": "` + test.ArgusGitHubRepo + `"
				},
				"webhook": [
					{
						"name": "github",
						"type": "github",
						"url": "https://example.com/github",
						"secret": "` + util.SecretValue + `",
						"headers": [
							{
								"key": "X-Foo",
								"value": "` + util.SecretValue + `",
								"old_index": 0
							}
						],
						"old_index": "github-initial"
					},
					{
						"name": "gitlab-",
						"type": "gitlab",
						"url": "https://example.com/gitlab",
						"secret": "` + util.SecretValue + `",
						"headers": [
							{
								"key": "X-Bar",
								"value": "` + util.SecretValue + `",
								"old_index": 0
							}
						],
						"old_index": "gitlab-initial"
					}
				]
			}`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							url: `+test.ArgusGitHubRepo+`
							access_token: aToken

						webhook:
							github:
								headers:
									- key: X-Foo
										value: aFoo
								url: https://example.com/github
								secret: githubSecret
							gitlab-:
								type: gitlab
								headers:
									- key: X-Bar
										value: aBar
								url: https://example.com/gitlab
								secret: gitlabSecret
					`)),
					"Give Webhook secrets",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: `+test.ArgusDockerGHCRRepo+`
									tag: "{{ version }}"
									auth:
										token: anotherToken

						webhook:
							github-initial:
								headers:
									- key: X-Foo
										value: aFoo
								url: https://example.com/github
								secret: githubSecret
							gitlab-initial:
								headers:
									- key: X-Bar
										value: aBar
								url: https://example.com/gitlab
								secret: gitlabSecret
					`)),
					"Give Webhook secrets",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "Give ALL secrets",
			payload: `{
				"latest_version": {
					"access_token": "` + util.SecretValue + `",
					"url": "` + test.ArgusGitHubRepo + `",
					"require": {
						"docker": {
							"type": "ghcr",
							"image": "` + test.ArgusDockerGHCRRepo + `",
							"tag": "{{ version }}",
							"auth": {
								"token": "` + util.SecretValue + `"
							}
						}
					}
				},
				"deployed_version": {
					"url": "` + test.LookupWithHeaderAuth.URLValid + `",
					"basic_auth": {
						"password": "` + util.SecretValue + `"
					},
					"headers": [
						{
							"key": "X-Foo",
							"value": "` + util.SecretValue + `",
							"old_index": 0
						}
					]
				},
				"notify": [
					{
						"name": "join",
						"type": "join",
						"url_fields": {
							"apikey": "` + util.SecretValue + `"
						},
						"params": {
							"devices": "` + util.SecretValue + `",
							"icon": "https://example.com/icon.png"
						},
						"old_index": "join-initial"
					},
					{
						"name": "matrix-",
						"type": "matrix",
						"url_fields": {
							"host":    "matrix.example.com",
							"password": "matrixToken"
						},
						"old_index": "matrix-initial"
					},
					{
						"name": "rocketchat",
						"type": "rocketchat",
						"url_fields": {
							"channel": "argus",
							"host":    "https://example.com",
							"tokena": "` + util.SecretValue + `",
							"tokenb": "` + util.SecretValue + `"
						},
						"old_index": "rocketchat-initial"
					},
					{
						"name": "slack",
						"type": "slack",
						"url_fields": {
							"channel": "ABC",
							"token": "` + util.SecretValue + `"
						},
						"old_index": "slack-initial"
					},
					{
						"name": "gotify",
						"type": "gotify",
						"url_fields": {
							"host":  "gotify.example.com",
							"token": "` + util.SecretValue + `"
						},
						"old_index": "gotify-initial"
					},
					{
						"name": "zulip",
						"type": "zulip",
						"url_fields": {
							"botkey": "` + util.SecretValue + `",
							"botmail": "my-bot@example.com",
							"host":    "zulipchat.example.com"
						},
						"old_index": "zulip-initial"
					}
				],
				"webhook": [
					{
						"name": "github",
						"type": "github",
						"url": "https://example.com/github",
						"secret": "` + util.SecretValue + `",
						"headers": [
							{
								"key": "X-Foo",
								"value": "` + util.SecretValue + `",
								"old_index": 0
							}
						],
						"old_index": "github-initial"
					},
					{
						"name": "gitlab-",
						"type": "gitlab",
						"url": "https://example.com/gitlab",
						"secret": "` + util.SecretValue + `",
						"headers": [
							{
								"key": "X-Bar",
								"value": "` + util.SecretValue + `",
								"old_index": 0
							}
						],
						"old_index": "gitlab-initial"
					}
				]
			}`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							url: `+test.ArgusGitHubRepo+`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: `+test.ArgusDockerGHCRRepo+`
									tag: "{{ version }}"
									auth:
										token: anotherToken

						deployed_version:
							url: `+test.LookupWithHeaderAuth.URLValid+`
							basic_auth:
								password: aPassword
							headers:
								- key: X-Foo
									value: aFoo

						notify:
							join:
								url_fields:
									apikey: joinApiKey
								params:
									devices: aDevice
									icon: https://example.com/icon.png
							matrix-:
								type: matrix
								url_fields:
									host: matrix.example.com
									password: matrixToken
							rocketchat:
								url_fields:
									channel: argus
									host: example.com
									tokena: rocketchatTokenA
									tokenb: rocketchatTokenB
							slack:
								url_fields:
									channel: ABC
									token: xoxABCDEFGHI-012345678-abcdefghi01234567abcdefghi
							gotify:
								url_fields:
									host: gotify.example.com
									token: gotifyToken
							zulip:
								url_fields:
									botkey: zulipBotKey
									botmail: my-bot%40example.com
									host: zulipchat.example.com

						webhook:
							github:
								headers:
									- key: X-Foo
										value: aFoo
								url: https://example.com/github
								secret: githubSecret
							gitlab-:
								type: gitlab
								headers:
									- key: X-Bar
										value: aBar
								url: https://example.com/gitlab
								secret: gitlabSecret
					`)),
					"Give ALL secrets",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: `+test.ArgusDockerGHCRRepo+`
									tag: "{{ version }}"
									auth:
										token: anotherToken

						deployed_version:
							type: url
							url: https://example.com
							basic_auth:
								password: aPassword
							headers:
								- key: X-Foo
									value: aFoo

						notify:
							join-initial:
								type: join
								url_fields:
									apikey: joinApiKey
								params:
									devices: aDevice
							matrix-:
								type: matrix
								url_fields:
									host: matrix.example.com
									password: matrixToken
							rocketchat-initial:
								type: rocketchat
								url_fields:
									tokena: rocketchatTokenA
									tokenb: rocketchatTokenB
							slack-initial:
								type: slack
								url_fields:
									token: xoxABCDEFGHI-012345678-abcdefghi01234567abcdefghi
							gotify-initial:
								type: gotify
								url_fields:
									host: gotify.example.com
									token: gotifyToken
							zulip-initial:
								type: zulip
								url_fields:
									botkey: zulipBotKey

						webhook:
							github-initial:
								headers:
									- key: X-Foo
										value: aFoo
								url: https://example.com/github
								secret: githubSecret
							gitlab-initial:
								headers:
									- key: X-Bar
										value: aBar
								url: https://example.com/gitlab
								secret: gitlabSecret
						`)),
					"Give ALL secrets",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
		{
			name: "CheckValues fail",
			payload: test.TrimJSON(`{
				"latest_version": {
					"type": "url",
					"url": "https://example.com"
				},
				"webhook": [
					{
						"name": "test",
						"old_index": "github-initial",
						"type": "github",
						"url": "https://example.com",
						"secret": "` + util.SecretValue + `"
					}
				]
			}`),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						webhook:
							github-initial:
								type: github
								url: https://example.com/github
								headers:
									- key: X-Foo
										value: aFoo
						`)),
					"CheckValues fail",
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: test.TrimYAML(`
				webhook:
					test:
						secret: <required>.*$`,
			),
		},
		{
			name: "CommandFromDefaults",
			payload: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			svcCfg: DefaultsConfig{
				Soft: &Defaults{
					Command: command.Commands{
						{"ls", "-lah"},
					},
				},
			},
			wantFromDefaults: fromDefaults{
				command: true,
			},
			errRegex: `^$`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					"CommandFromDefaults",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "NotifyFromDefaults",
			payload: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			svcCfg: DefaultsConfig{
				Soft: &Defaults{
					Notify: map[string]struct{}{
						"alpha":   {},
						"bravo":   {},
						"charlie": {},
					},
				},
			},
			notifyCfg: shoutrrr.Config{
				Root: shoutrrr.ShoutrrrsDefaults{
					"alpha":   {Type: "smtp"},
					"bravo":   {Type: "smtp"},
					"charlie": {Type: "smtp"},
				},
				Defaults: shoutrrr.ShoutrrrsDefaults{
					"smtp": {
						URLFields: map[string]string{
							"host": "example.com",
						},
						Params: map[string]string{
							"fromaddress": "foo@exampke.com",
							"toaddresses": "bar@exampke.com",
						},
					},
				},
			},
			wantFromDefaults: fromDefaults{
				notify: true,
			},
			errRegex: `^$`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					"NotifyFromDefaults",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "WebhookFromDefaults",
			payload: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			svcCfg: DefaultsConfig{
				Soft: &Defaults{
					Webhook: map[string]struct{}{
						"alpha":   {},
						"bravo":   {},
						"charlie": {},
					},
				},
			},
			whCfg: webhook.Config{
				Defaults: &webhook.Defaults{
					URL:    "https://example.com/github",
					Secret: "something",
				},
			},
			wantFromDefaults: fromDefaults{
				webhook: true,
			},
			errRegex: `^$`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					"WebhookFromDefaults",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "CommandFromDefaults + NotifyFromDefaults + WebhookFromDefaults",
			payload: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			svcCfg: DefaultsConfig{
				Soft: &Defaults{
					Command: command.Commands{
						{"ls", "-lah"},
					},
					Notify: map[string]struct{}{
						"alpha":   {},
						"bravo":   {},
						"charlie": {},
					},
					Webhook: map[string]struct{}{
						"alpha":   {},
						"bravo":   {},
						"charlie": {},
					},
				},
			},
			notifyCfg: shoutrrr.Config{
				Root: shoutrrr.ShoutrrrsDefaults{
					"alpha":   {Type: "smtp"},
					"bravo":   {Type: "smtp"},
					"charlie": {Type: "smtp"},
				},
				Defaults: shoutrrr.ShoutrrrsDefaults{
					"smtp": {
						URLFields: map[string]string{
							"host": "example.com",
						},
						Params: map[string]string{
							"fromaddress": "foo@exampke.com",
							"toaddresses": "bar@exampke.com",
						},
					},
				},
			},
			whCfg: webhook.Config{
				Defaults: &webhook.Defaults{
					URL:    "https://example.com/github",
					Secret: "something",
				},
			},
			wantFromDefaults: fromDefaults{
				command: true,
				notify:  true,
				webhook: true,
			},
			errRegex: `^$`,
			want: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					"CommandFromDefaults + NotifyFromDefaults + WebhookFromDefaults",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.payload = test.TrimJSON(tc.payload)
			reader := bytes.NewReader([]byte(tc.payload))
			payload := io.NopCloser(reader)
			if tc.svcCfg.Soft == nil {
				tc.svcCfg.Soft = svcCfg.Soft
			}
			if tc.svcCfg.Hard == nil {
				tc.svcCfg.Hard = svcCfg.Hard
			}
			if tc.notifyCfg.Defaults == nil {
				tc.notifyCfg.Defaults = notifyCfg.Defaults
			}
			if tc.notifyCfg.HardDefaults == nil {
				tc.notifyCfg.HardDefaults = notifyCfg.HardDefaults
			}
			if tc.whCfg.Defaults == nil {
				tc.whCfg.Defaults = whCfg.Defaults
			}
			if tc.whCfg.HardDefaults == nil {
				tc.whCfg.HardDefaults = whCfg.HardDefaults
			}
			if tc.oldService != nil {
				tc.oldService.Defaults = svcCfg.Soft
				tc.oldService.HardDefaults = svcCfg.Hard
				tc.oldService.Status.Init(
					len(tc.oldService.Command), len(tc.oldService.Notify), len(tc.oldService.Webhook),
					status.ServiceInfo{
						ID:         tc.oldService.ID,
						Name:       tc.oldService.Name,
						ServiceURL: "https://example.com/service/url",
					},
					&tc.oldService.Dashboard,
				)
				tc.oldService.init(notifyCfg, whCfg, nil, nil, nil)
			}

			svc, err := FromPayload(
				tc.oldService,
				&payload,
				tc.svcCfg, tc.notifyCfg, tc.whCfg,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			gotStr, wantStr := svc.String(""), tc.want.String("")
			if gotStr != wantStr {
				t.Errorf("Service.String() mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}

			if svc.CommandFromDefaults != tc.wantFromDefaults.command {
				t.Errorf(
					"Service.CommandFromDefaults mismatch\ngot:  %t\nwant: %t",
					svc.CommandFromDefaults, tc.wantFromDefaults.command,
				)
			} else if svc.CommandFromDefaults {
				if testErr := test.AssertSlicesEqualFunc(
					t,
					svc.Command,
					tc.svcCfg.Soft.Command,
					func(a, b command.Command) bool {
						return a.JSON() == b.JSON()
					},
					"Service.Command",
				); testErr != nil {
					t.Error(testErr)
				}
			}
			if svc.NotifyFromDefaults != tc.wantFromDefaults.notify {
				t.Errorf(
					"Service.NotifyFromDefaults mismatch\ngot:  %t\nwant: %t",
					svc.NotifyFromDefaults, tc.wantFromDefaults.notify,
				)
			} else if svc.NotifyFromDefaults {
				gotKeys := util.SortedKeys(svc.Notify)
				wantKeys := util.SortedKeys(tc.svcCfg.Soft.Notify)
				if !util.AreSlicesEqual(gotKeys, wantKeys) {
					t.Errorf("Service.Notify keys mismatch\ngot:  %q\nwant: %q", gotKeys, wantKeys)
				}
			}
			if svc.WebhookFromDefaults != tc.wantFromDefaults.webhook {
				t.Errorf(
					"Service.WebhookFromDefaults mismatch\ngot:  %t\nwant: %t",
					svc.WebhookFromDefaults, tc.wantFromDefaults.webhook,
				)
			} else if svc.WebhookFromDefaults {
				gotKeys := util.SortedKeys(svc.Webhook)
				wantKeys := util.SortedKeys(tc.svcCfg.Soft.Webhook)
				if !util.AreSlicesEqual(gotKeys, wantKeys) {
					t.Errorf("Service.Webhook keys mismatch\ngot:  %q\nwant: %q", gotKeys, wantKeys)
				}
			}
		})
	}
}

func TestFromPayload__noServiceCreated(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	original := decodeServiceFromPayload
	decodeServiceFromPayload = func(
		format string,
		data []byte,
		id string,
		defaultsCfg DefaultsConfig,
		notifyCfg shoutrrr.Config,
		whCfg webhook.Config,
	) (*Service, error) {
		return nil, nil
	}
	t.Cleanup(func() { decodeServiceFromPayload = original })
	errRegex := `^no service created from payload$`

	payload := io.NopCloser(bytes.NewReader([]byte(`{
		"latest_version": {
			"type": "github",
			"url": "` + test.ArgusGitHubRepo + `"
		}
	}`)))

	_, err := FromPayload(
		nil,
		&payload,
		svcCfg, notifyCfg, whCfg,
	)
	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestFromPayload_ReadFromFail(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	payload := io.NopCloser(bytes.NewReader([]byte("this is a long payload")))
	payload = http.MaxBytesReader(nil, payload, 5)
	errRegex := `^http: request body too large$`

	_, err := FromPayload(
		&Service{},
		&payload,
		svcCfg, notifyCfg, whCfg,
	)
	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
}

func TestService_CheckFetches(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()
	id := t.Name()

	// Queried for the versions CheckFetches should find.
	testLV := lvtest.Lookup(t, "url", false)
	if _, err := testLV.Query(false, logx.LogFrom{}); err != nil {
		t.Fatalf("latestver.Lookup.Query() failed: %v", err)
	}
	testDVL := dvtest.Lookup(t, "url", false, "")
	if err := testDVL.Query(false, logx.LogFrom{}); err != nil {
		t.Fatalf("deployedver.Lookup.Query() failed: %v", err)
	}
	tests := []struct {
		name                                      string
		svc                                       *Service
		startLatestVersion, wantLatestVersion     string
		startDeployedVersion, wantDeployedVersion string
		errRegex                                  string
	}{
		{
			name: "have LatestVersion/nil DeployedVersionLookup",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
						`+lvtest.Lookup(t, "url", false).String("  ")+`
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			startLatestVersion:   "foo",
			wantLatestVersion:    testLV.GetStatus().LatestVersion(),
			startDeployedVersion: "bar",
			wantDeployedVersion:  "bar",
			errRegex:             `^$`,
		},
		{
			name: "have LatestVersion/have DeployedVersionLookup",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
						`+lvtest.Lookup(t, "url", false).String("  ")+`
						deployed_version:
						`+dvtest.Lookup(t, "url", false, "").String("  ")+`
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			startLatestVersion:   "foo",
			wantLatestVersion:    testLV.GetStatus().LatestVersion(),
			wantDeployedVersion:  testDVL.GetStatus().DeployedVersion(),
			startDeployedVersion: "bar",
			errRegex:             `^$`,
		},
		{
			name: "latest_version query fails",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
						`+lvtest.Lookup(t, "url", true).String("  ")+`
						deployed_version:
						`+dvtest.Lookup(t, "url", false, "").String("  ")+`
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			startLatestVersion: "foo",
			wantLatestVersion:  "foo",
			errRegex: test.TrimYAML(`
				^latest_version fetches failed:
					x509 \(certificate invalid\)$`,
			),
		},
		{
			name: "deployed_version query fails",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
						`+lvtest.Lookup(t, "url", false).String("  ")+`
						deployed_version:
						`+dvtest.Lookup(t, "url", true, "").String("  ")+`
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			startLatestVersion:   "foo",
			wantLatestVersion:    testLV.GetStatus().LatestVersion(),
			startDeployedVersion: "bar",
			wantDeployedVersion:  "bar",
			errRegex: test.TrimYAML(`
				^deployed_version fetches failed:
					x509 \(certificate invalid\)$`,
			),
		},
		{
			name: "both queried",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
						`+lvtest.Lookup(t, "url", false).String("  ")+`
						deployed_version:
						`+dvtest.Lookup(t, "url", false, "").String("  ")+`
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			wantLatestVersion:   testLV.GetStatus().LatestVersion(),
			wantDeployedVersion: testDVL.GetStatus().DeployedVersion(),
			errRegex:            `^$`,
		},
		{
			name: "manual deployed_version is left for AddService to apply",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
						`+lvtest.Lookup(t, "url", false).String("  ")+`
						deployed_version:
							type: manual
							version: 1.2.3
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			startLatestVersion:   "foo",
			wantLatestVersion:    testLV.GetStatus().LatestVersion(),
			startDeployedVersion: "bar",
			wantDeployedVersion:  "bar", // A manual lookup is not queried.
			errRegex:             `^$`,
		},
		{
			name: "active=false queries neither",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						options:
							active: false
						latest_version:
						`+lvtest.Lookup(t, "url", false).String("  ")+`
						deployed_version:
						`+dvtest.Lookup(t, "url", false, "").String("  ")+`
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.svc.Defaults = svcCfg.Soft
			tc.svc.HardDefaults = svcCfg.Hard
			tc.svc.Options.SetDefaults(
				&tc.svc.Defaults.Options,
				&tc.svc.HardDefaults.Options,
			)
			tc.svc.init(notifyCfg, whCfg, nil, nil, nil)
			announceChannel := make(chan []byte, 5)
			tc.svc.Status.AnnounceChannel = announceChannel
			tc.svc.Status.SetLatestVersion(tc.startLatestVersion, "", false)
			tc.svc.Status.SetDeployedVersion(tc.startDeployedVersion, "", false)

			err := tc.svc.CheckFetches()
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if got := tc.svc.Status.LatestVersion(); got != tc.wantLatestVersion {
				t.Errorf(
					"Status.LatestVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.wantLatestVersion,
				)
			}

			if got := tc.svc.Status.DeployedVersion(); got != tc.wantDeployedVersion {
				t.Errorf(
					"Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
					got, tc.wantDeployedVersion,
				)
			}

			if got, want := len(tc.svc.Status.AnnounceChannel), 0; got != want {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					got, want,
				)
			}
		})
	}
}

func TestService_GiveSecrets(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()
	id := t.Name()
	timestamp := "2024-01-02T03:04:05Z"

	type statusTests struct {
		oldLatestVersion, expectedLatestVersion                       string
		oldLatestVersionTimestamp, expectedLatestVersionTimestamp     string
		oldDeployedVersion, expectedDeployedVersion                   string
		oldDeployedVersionTimestamp, expectedDeployedVersionTimestamp string
	}
	type commandTests struct {
		oldFails      []*bool
		expectedFails []*bool
	}
	type webhookTests struct {
		oldFails      map[string]*bool
		expectedFails map[string]*bool
	}

	tests := []struct {
		name            string
		svc, oldService *Service
		statusTests     statusTests
		commandTests    commandTests
		webhookTests    webhookTests
		secretRefs      secretRefs
		want            string
	}{
		{
			name: "no secrets",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: something
						deployed_version:
							type: url
							url: https://example.com
							basic_auth:
								username: user
								password: pass
						notify:
							foo:
								url_fields:
									apikey: saucy
							bar:
								url_fields:
									avatar: https://example.com/logo.png
						webhook:
							foo:
								url: https://example.com/foo
								secret: foo
							bar:
								url: https://example.com/bar
								secret: bar
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: somethingElse

						deployed_version:
							type: url
							url: https://example.com
							basic_auth:
								username: username
								password: password
						notify:
							foo:
								url_fields:
									apikey: sweet
							bar:
								url_fields:
									avatar: https://example.com/logo.jpg
						webhook:
							foo:
								url: https://example.com/foo
								secret: foo
							bar:
								url: https://example.com/bar
								secret: bar
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			secretRefs: secretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
					access_token: something
				deployed_version:
					type: url
					url: https://example.com
					basic_auth:
						username: user
						password: pass
				notify:
					bar:
						url_fields:
							avatar: https://example.com/logo.png
					foo:
						url_fields:
							apikey: saucy
				webhook:
					bar:
						url: https://example.com/bar
						secret: bar
					foo:
						url: https://example.com/foo
						secret: foo
			`),
		},
		{
			name: "minimal CREATE",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: nil,
			secretRefs: secretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
			`),
		},
		{
			name: "no oldService (CREATE)",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: `+util.SecretValue+`

						deployed_version:
							type: url
							url: https://example.com
							basic_auth:
								username: `+util.SecretValue+`
								password: `+util.SecretValue+`

						notify:
							bar:
								url_fields:
									avatar: `+util.SecretValue+`
							foo:
								url_fields:
									apikey: `+util.SecretValue+`

						webhook:
							bar:
								url: https://example.com/bar
								secret: `+util.SecretValue+`
							foo:
								url: https://example.com/foo
								secret: `+util.SecretValue+`
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: nil,
			secretRefs: secretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
					access_token: ` + util.SecretValue + `
				deployed_version:
					type: url
					url: https://example.com
					basic_auth:
						username: ` + util.SecretValue + `
						password: ` + util.SecretValue + `
				notify:
					bar:
						url_fields:
							avatar: ` + util.SecretValue + `
					foo:
						url_fields:
							apikey: ` + util.SecretValue + `
				webhook:
					bar:
						url: https://example.com/bar
						secret: ` + util.SecretValue + `
					foo:
						url: https://example.com/foo
						secret: ` + util.SecretValue + `
			`),
		},
		{
			name: "no secretRefs",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: `+util.SecretValue+`

						deployed_version:
							type: url
							url: https://example.com
							basic_auth:
								username: `+util.SecretValue+`
								password: `+util.SecretValue+`

						notify:
							bar:
								url_fields:
									avatar: `+util.SecretValue+`
							foo:
								url_fields:
									apikey: `+util.SecretValue+`

						webhook:
							bar:
								url: https://example.com/bar
								secret: `+util.SecretValue+`
							foo:
								url: https://example.com/foo
								secret: `+util.SecretValue+`
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
							latest_version:
								type: github
								url: `+test.ArgusGitHubRepo+`
								access_token: somethingElse

							deployed_version:
								type: url
								url: https://example.com
								basic_auth:
									username: username
									password: password

							notify:
								bar:
									url_fields:
										avatar: https://example.com/logo.png
								foo:
									url_fields:
										apikey: sweet

							webhook:
								bar:
									url: https://example.com/bar
									secret: bar
								foo:
									url: https://example.com/foo
									secret: foo
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			webhookTests: webhookTests{
				oldFails: map[string]*bool{
					"foo": new(false),
					"bar": new(true),
				},
				expectedFails: map[string]*bool{
					"foo": nil,
					"bar": nil,
				},
			},
			secretRefs: secretRefs{},
			// Without secretRefs, only access_token and the basic_auth password are inherited.
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
					access_token: somethingElse
				deployed_version:
					type: url
					url: https://example.com
					basic_auth:
						username: ` + util.SecretValue + `
						password: password
				notify:
					bar:
						url_fields:
							avatar: ` + util.SecretValue + `
					foo:
						url_fields:
							apikey: ` + util.SecretValue + `
				webhook:
					bar:
						url: https://example.com/bar
						secret: ` + util.SecretValue + `
					foo:
						url: https://example.com/foo
						secret: ` + util.SecretValue + `
			`),
		},
		{
			name: "matching secretRefs",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: somethingElse

						deployed_version:
							type: url
							url: https://example.com
							basic_auth:
								username: `+util.SecretValue+`
								password: password
							headers:
								- key: X-Foo
									value: `+util.SecretValue+`
								- key: X-Bar
									value: `+util.SecretValue+`

						notify:
							bar:
								url_fields:
									avatar: `+util.SecretValue+`
							foo:
								url_fields:
									apikey: `+util.SecretValue+`

						webhook:
							bar:
								url: https://example.com/bar
								secret: `+util.SecretValue+`
							foo:
								url: https://example.com/foo
								secret: `+util.SecretValue+`
					`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
							access_token: somethingElse

						deployed_version:
							type: url
							url: https://example.com
							basic_auth:
								username: username
								password: password
							headers:
								- key: X-Foo
									value: foo
								- key: X-Bar
									value: bar

						notify:
							bar:
								url_fields:
									avatar: https://example.com/logo.png
							foo:
								url_fields:
									apikey: sweet

						webhook:
							bar:
								url: https://example.com/bar
								secret: bar
							foo:
								url: https://example.com/foo
								secret: foo
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			secretRefs: secretRefs{
				DeployedVersionLookup: shared.VSecretRef{
					Headers: []shared.OldIntIndex{
						{OldIndex: new(0)},
						{OldIndex: new(1)},
					},
				},
				Notify: map[string]shared.OldStringIndex{
					"foo": {OldIndex: "foo"},
					"bar": {OldIndex: "bar"},
				},
				Webhook: map[string]shared.WHSecretRef{
					"foo": {OldIndex: "foo"},
					"bar": {OldIndex: "bar"},
				},
			},
			// avatar is not censorable and the username is not a secret, so both stay masked.
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
					access_token: somethingElse
				deployed_version:
					type: url
					url: https://example.com
					basic_auth:
						username: ` + util.SecretValue + `
						password: password
					headers:
						- key: X-Foo
							value: foo
						- key: X-Bar
							value: bar
				notify:
					bar:
						url_fields:
							avatar: ` + util.SecretValue + `
					foo:
						url_fields:
							apikey: sweet
				webhook:
					bar:
						url: https://example.com/bar
						secret: bar
					foo:
						url: https://example.com/foo
						secret: foo
			`),
		},
		{
			name: "LatestVersion.URL/unchanged retains Status.LatestVersion",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: url
							url: https://example.com
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: url
							url: https://example.com
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			statusTests: statusTests{
				oldLatestVersion:               "1.2.3",
				expectedLatestVersion:          "1.2.3",
				oldLatestVersionTimestamp:      timestamp,
				expectedLatestVersionTimestamp: timestamp,
			},
			secretRefs: secretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name: "LatestVersion.URL/changed loses Status.LatestVersion",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: url
							url: https://example.com
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: url
							url: https://example.com/somewhere-else
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			statusTests: statusTests{
				oldLatestVersion:          "1.2.3",
				oldLatestVersionTimestamp: timestamp,
			},
			secretRefs: secretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name: "DeployedVersion.URL/unchanged retains Status.DeployedVersion",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						deployed_version:
							type: url
							url: https://example.com
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						deployed_version:
							type: url
							url: https://example.com
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			statusTests: statusTests{
				oldDeployedVersion:               "1.2.3",
				expectedDeployedVersion:          "1.2.3",
				oldDeployedVersionTimestamp:      timestamp,
				expectedDeployedVersionTimestamp: timestamp,
			},
			secretRefs: secretRefs{},
			want: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name: "DeployedVersion.URL/changed loses Status.DeployedVersion",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						deployed_version:
							type: url
							url: https://example.com
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						deployed_version:
							type: url
							url: https://example.com/somewhere-else
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			statusTests: statusTests{
				oldDeployedVersion:          "1.2.3",
				oldDeployedVersionTimestamp: timestamp,
			},
			secretRefs: secretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
				deployed_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name: "Webhook/unchanged retains Failed",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						webhook:
							test:
								url: https://example.com
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						webhook:
							test:
								url: https://example.com
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			secretRefs: secretRefs{
				Webhook: map[string]shared.WHSecretRef{
					"test": {OldIndex: "test"},
				},
			},
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
				webhook:
					test:
						url: https://example.com
			`),
			webhookTests: webhookTests{
				oldFails: map[string]*bool{
					"test": new(true),
				},
				expectedFails: map[string]*bool{
					"test": new(true),
				},
			},
		},
		{
			name: "Webhook/changed loses Failed",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						webhook:
							test:
								url: https://example.com/other
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						webhook:
							test:
								url: https://example.com
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			secretRefs: secretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
				webhook:
					test:
						url: https://example.com/other
			`),
			webhookTests: webhookTests{
				oldFails: map[string]*bool{
					"test": new(true),
				},
				expectedFails: map[string]*bool{
					"test": nil,
				},
			},
		},
		{
			name: "Command/unchanged retains Failed",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						command:
							- - ls
								- -la
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						command:
							- - ls
								- -la
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
				command:
					- - ls
						- -la
			`),
			secretRefs: secretRefs{},
			commandTests: commandTests{
				oldFails: []*bool{
					new(true),
				},
				expectedFails: []*bool{
					new(true),
				},
			},
		},
		{
			name: "Command/changed loses Failed",
			svc: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						command:
							- - ls
								- -lah
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			oldService: test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`

						command:
							- - ls
								- -la
						`)),
					id,
					svcCfg, notifyCfg, whCfg,
				)
			}),
			secretRefs: secretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
				command:
					- - ls
						- -lah
			`),
			commandTests: commandTests{
				oldFails: []*bool{
					new(true),
				},
				expectedFails: []*bool{
					nil,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.oldService != nil {
				tc.oldService.Status.SetLatestVersion(
					tc.statusTests.oldLatestVersion, tc.statusTests.oldLatestVersionTimestamp,
					false,
				)
				tc.oldService.Status.SetDeployedVersion(
					tc.statusTests.oldDeployedVersion, tc.statusTests.oldDeployedVersionTimestamp,
					false,
				)
				for k, v := range tc.commandTests.oldFails {
					if v != nil {
						tc.oldService.Status.Fails.Command.Set(k, *v)
					}
				}
				for k, v := range tc.webhookTests.oldFails {
					tc.oldService.Status.Fails.Webhook.Set(k, v)
				}
			}

			tc.svc.giveSecrets(tc.oldService, tc.secretRefs)

			if gotStr := tc.svc.String(""); gotStr != tc.want {
				t.Errorf("Service.String() mismatch\ngot:  %q\nwant: %q", gotStr, tc.want)
			}

			if testErr := test.AssertFields(
				t,
				[]test.FieldAssertion{
					{Name: "LatestVersion()", Got: tc.svc.Status.LatestVersion(), Want: tc.statusTests.expectedLatestVersion, Mode: test.CompareEqual},
					{Name: "LatestVersionTimestamp()", Got: tc.svc.Status.LatestVersionTimestamp(), Want: tc.statusTests.expectedLatestVersionTimestamp, Mode: test.CompareEqual},
					{Name: "DeployedVersion()", Got: tc.svc.Status.DeployedVersion(), Want: tc.statusTests.expectedDeployedVersion, Mode: test.CompareEqual},
					{Name: "DeployedVersionTimestamp()", Got: tc.svc.Status.DeployedVersionTimestamp(), Want: tc.statusTests.expectedDeployedVersionTimestamp, Mode: test.CompareEqual},
				},
				"Status",
			); testErr != nil {
				t.Error(testErr)
			}

			for i, want := range tc.commandTests.expectedFails {
				gotStr := test.StringifyPtr(tc.svc.Status.Fails.Command.Get(i))
				wantStr := test.StringifyPtr(want)
				if gotStr != wantStr {
					t.Errorf(
						"Status.Fails.Command.Get(%d) mismatch\ngot:  %q\nwant: %q",
						i, gotStr, wantStr,
					)
				}
			}

			for key, want := range tc.webhookTests.expectedFails {
				wh, ok := tc.svc.Webhook[key]
				if !ok {
					t.Errorf("Service.Webhook[%q] mismatch\ngot:  absent\nwant: present", key)
					continue
				}
				gotStr := test.StringifyPtr(wh.DidFail())
				wantStr := test.StringifyPtr(want)
				if gotStr != wantStr {
					t.Errorf(
						"Service.Webhook[%q].DidFail() mismatch\ngot:  %q\nwant: %q",
						key, gotStr, wantStr,
					)
				}
			}
		})
	}
}

func TestService_GiveSecretsLatestVersion(t *testing.T) {
	lvCfg := lvtest.PlainDefaultsConfig()
	githubLookupWithETag := func(etag string) latestver.Lookup {
		t.Helper()
		lookup := test.Must(t, func() (latestver.Lookup, error) {
			return latestver.Decode(
				"yaml", []byte(test.TrimYAML(`
					type: github
					url: `+test.ArgusGitHubRepo+`
				`)),
				nil,
				&status.Status{},
				lvCfg,
			)
		})
		githubLookup, ok := lookup.(*lvgithub.Lookup)
		if !ok {
			t.Fatalf("latestver.Decode() type mismatch\ngot:  %T\nwant: *lvgithub.Lookup", lookup)
		}
		githubLookup.GetGitHubData().SetETag(etag)
		return githubLookup
	}

	tests := []struct {
		name                   string
		latestVersion, otherLV latestver.Lookup
		secretRefs             shared.VSecretRef
		expected               latestver.Lookup
	}{
		{
			name: "nil/oldLatestVersion",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: nil,
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "empty AccessToken",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						access_token: foo
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "new AccessToken kept",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						access_token: foo
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						access_token: bar
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						access_token: foo
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "give old AccessToken",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						access_token: "`+util.SecretValue+`"
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						access_token: bar
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						access_token: bar
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "referencing default AccessToken",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						access_token: `+util.SecretValue+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			// The old token is empty, so the mask clears to fall back to the default.
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "nil/Require",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
								auth:
									token: foo
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "empty Require",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
								auth:
									token: foo
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "new Require.Docker.Token kept",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
								auth:
									token: foo
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
								auth:
									token: bar
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
								auth:
									token: foo
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "give old Require.Docker.Token",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
								auth:
									token: `+util.SecretValue+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
								auth:
									token: bar
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
								auth:
									token: bar
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "referencing default Require.Docker.Token",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
								auth:
									token: `+util.SecretValue+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						require:
							docker:
								type: ghcr
								image: `+test.ArgusGitHubRepo+`
								tag: latest
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "githubData/carried over if type still 'github'",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV:  githubLookupWithETag("shazam"),
			expected: githubLookupWithETag("shazam"),
		},
		{
			name: "githubData/not carried over if type wasn't 'github'",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						url: https://example.com
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "githubData/not carried over if type no longer 'github'",
			latestVersion: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						url: https://example.com
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
			otherLV: githubLookupWithETag("shazam"),
			expected: test.Must(t, func() (latestver.Lookup, error) {
				return latestver.Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						url: https://example.com
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "only new/changed Headers with want refs",
			latestVersion: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"},
				},
			},
			otherLV: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			expected: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "only new/changed Headers with no refs",
			latestVersion: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"},
				},
			},
			otherLV: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			expected: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{},
			},
		},
		{
			name: "secretRefs/Header value with no refs",
			latestVersion: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"},
				},
			},
			otherLV: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			expected: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{},
			},
		},
		{
			name: "only new/changed Headers with partial ref (not for all secrets)",
			latestVersion: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bang"},
					{Key: "bosh", Value: util.SecretValue},
				},
			},
			otherLV: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bash"},
				},
			},
			expected: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bang"},
					{Key: "bosh", Value: util.SecretValue},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
					{OldIndex: new(1)},
				},
			},
		},
		{
			name: "secretRefs/Header value",
			latestVersion: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"},
				},
			},
			otherLV: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			expected: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "secretRefs/index not found",
			latestVersion: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"},
				},
			},
			otherLV: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			expected: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(1)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "referencing some old Header values but not others",
			latestVersion: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bang"},
					{Key: "bish", Value: util.SecretValue},
				},
			},
			otherLV: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bong"},
				},
			},
			expected: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bang"},
					{Key: "bish", Value: "bong"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: nil},
					{OldIndex: new(1)},
				},
			},
		},
		{
			name: "swap header values",
			latestVersion: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: util.SecretValue},
					{Key: "foo", Value: util.SecretValue},
				},
			},
			otherLV: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bong"},
				},
			},
			expected: &lvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: "bar"},
					{Key: "foo", Value: "bong"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
					{OldIndex: new(1)},
				},
			},
		},
		{
			name: "forgejo/a masked token on the same host is inherited",
			latestVersion: &lvforgejo.Lookup{
				Host:        "https://codeberg.org",
				AccessToken: util.SecretValue,
			},
			otherLV: &lvforgejo.Lookup{
				Host:        "https://codeberg.org",
				AccessToken: "stored-token",
			},
			expected: &lvforgejo.Lookup{
				Host:        "https://codeberg.org",
				AccessToken: "stored-token",
			},
		},
		{
			name: "forgejo/a masked token on another spelling of the host is cleared",
			latestVersion: &lvforgejo.Lookup{
				Host:        "CODEBERG.org:443/",
				AccessToken: util.SecretValue,
			},
			otherLV: &lvforgejo.Lookup{
				Host:        "https://codeberg.org",
				AccessToken: "stored-token",
			},
			expected: &lvforgejo.Lookup{
				Host: "CODEBERG.org:443/",
			},
		},
		{
			name: "forgejo/a masked token on a changed host is cleared",
			latestVersion: &lvforgejo.Lookup{
				Host:        "https://git.example.com",
				AccessToken: util.SecretValue,
			},
			otherLV: &lvforgejo.Lookup{
				Host:        "https://codeberg.org",
				AccessToken: "stored-token",
			},
			expected: &lvforgejo.Lookup{
				Host: "https://git.example.com",
			},
		},
		{
			name: "forgejo/placeholder to a default is cleared",
			latestVersion: &lvforgejo.Lookup{
				Host:        "https://codeberg.org",
				AccessToken: util.SecretValue,
			},
			otherLV: &lvforgejo.Lookup{
				Host: "https://codeberg.org",
			},
			expected: &lvforgejo.Lookup{
				Host: "https://codeberg.org",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newService := &Service{LatestVersion: tc.latestVersion}
			oldService := &Service{LatestVersion: tc.otherLV}

			newService.giveSecretsLatestVersion(oldService.LatestVersion, &tc.secretRefs)

			// Secrets are inherited in place.
			gotLV := newService.LatestVersion
			if reflect.TypeOf(gotLV) != reflect.TypeOf(tc.expected) {
				t.Fatalf(
					"Service.LatestVersion type mismatch\ngot:  %T\nwant: %T",
					gotLV, tc.expected,
				)
			}

			var gotAccessToken, wantAccessToken string
			switch want := tc.expected.(type) {
			case *lvforgejo.Lookup:
				gotAccessToken = gotLV.(*lvforgejo.Lookup).AccessToken
				wantAccessToken = want.AccessToken
			case *lvgithub.Lookup:
				gotAccessToken = gotLV.(*lvgithub.Lookup).AccessToken
				wantAccessToken = want.AccessToken
			}
			if gotAccessToken != wantAccessToken {
				t.Errorf(
					"Lookup.AccessToken mismatch\ngot:  %q\nwant: %q",
					gotAccessToken, wantAccessToken,
				)
			}

			gotRequire, expectedRequire := gotLV.GetRequire(), tc.expected.GetRequire()
			switch {
			case gotRequire == nil && expectedRequire != nil:
				t.Error("Lookup.GetRequire() mismatch\ngot:  nil\nwant: non-nil")
			case gotRequire != nil && expectedRequire == nil:
				t.Error("Lookup.GetRequire() mismatch\ngot:  non-nil\nwant: nil")
			case gotRequire != nil:
				var gotToken, wantToken string
				if gotRequire.Docker != nil {
					gotToken = gotRequire.Docker.GetAuth().GetTokenSelf()
				}
				if expectedRequire.Docker != nil {
					wantToken = expectedRequire.Docker.GetAuth().GetTokenSelf()
				}
				if gotToken != wantToken {
					t.Errorf(
						"Lookup.GetRequire().Docker.GetAuth().GetTokenSelf() mismatch\n"+
							"got:  %q\nwant: %q",
						gotToken, wantToken,
					)
				}
			}

			if expectedLatestVersion, ok := tc.expected.(*lvgithub.Lookup); ok {
				got := gotLV.(*lvgithub.Lookup).GetGitHubData().String()
				expected := expectedLatestVersion.GetGitHubData().String()
				if got != expected {
					t.Errorf(
						"Lookup.GetGitHubData().String() mismatch\ngot:  %q\nwant: %q",
						got, expected,
					)
				}
			}
			var expectedHeaders shared.Headers
			if expectedLV, ok := tc.expected.(*lvweb.Lookup); ok {
				expectedHeaders = expectedLV.Headers
			}
			var gotHeaders shared.Headers
			if gotWebLV, ok := gotLV.(*lvweb.Lookup); ok {
				gotHeaders = gotWebLV.Headers
			}
			if testErr := test.AssertSlicesEqual(t, gotHeaders, expectedHeaders, "Lookup.Headers"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestService_GiveSecretsDeployedVersion(t *testing.T) {
	tests := []struct {
		name                     string
		deployedVersion, otherDV deployedver.Lookup
		secretRefs               shared.VSecretRef
		want                     deployedver.Lookup
	}{
		{
			name:            "nil/DeployedVersion",
			deployedVersion: nil,
			otherDV:         &dvweb.Lookup{},
			secretRefs:      shared.VSecretRef{},
			want:            nil,
		},
		{
			name: "nil/OldDeployedVersion",
			deployedVersion: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: "foo",
				},
			},
			otherDV: nil,
			want: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: "foo",
				},
			},
		},
		{
			name: "BasicAuth.Password/keep new value",
			deployedVersion: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: "foo",
				},
			},
			otherDV: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: "bar",
				},
			},
			want: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: "foo",
				},
			},
		},
		{
			name: "BasicAuth.Password/inherit",
			deployedVersion: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: util.SecretValue,
				},
			},
			otherDV: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: "bar",
				},
			},
			want: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: "bar",
				},
			},
		},
		{
			name: "BasicAuth.Password/referencing default",
			deployedVersion: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: util.SecretValue,
				},
			},
			otherDV: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{},
			},
			want: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: "",
				},
			},
		},
		{
			name: "BasicAuth.Password/unknown secret",
			deployedVersion: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: util.SecretValue,
				},
			},
			otherDV: &dvweb.Lookup{},
			want: &dvweb.Lookup{
				BasicAuth: &dvweb.BasicAuth{
					Password: util.SecretValue,
				},
			},
		},
		{
			name: "Headers/empty",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{},
			},
		},
		{
			name: "only new Headers",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: "bash"},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: nil},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "Headers/secretRefs index out of range",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: util.SecretValue},
					{Key: "bash", Value: util.SecretValue},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: util.SecretValue},
					{Key: "bash", Value: util.SecretValue},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
					{OldIndex: new(1)},
				},
			},
		},
		{
			name: "Headers/SecretValue but nil index secretRefs",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: util.SecretValue},
					{Key: "bash", Value: util.SecretValue},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: "bash"},
					{Key: "bash", Value: "bop"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: util.SecretValue},
					{Key: "bash", Value: util.SecretValue},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: nil},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "Headers/only changed",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
				},
			},
		},
		{
			name: "Headers/only new + changed",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "Headers/only new + changed with secretRefs unused",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "Headers/only new + changed with no secretRefs",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{},
			},
		},
		{
			name: "secretRefs/Header value with no refs",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{},
			},
		},
		{
			name: "only new/changed Headers with partial ref (not for all secrets)",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bang"},
					{Key: "bosh", Value: util.SecretValue},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bash"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bang"},
					{Key: "bosh", Value: util.SecretValue},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
					{OldIndex: new(1)},
				},
			},
		},
		{
			name: "secretRefs/Header value",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "secretRefs/index out of range",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(1)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "referencing some old Header values but not others",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bang"},
					{Key: "bish", Value: util.SecretValue},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bong"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bang"},
					{Key: "bish", Value: "bong"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: nil},
					{OldIndex: new(1)},
				},
			},
		},
		{
			name: "swap header values",
			deployedVersion: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: util.SecretValue},
					{Key: "foo", Value: util.SecretValue},
				},
			},
			otherDV: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bong"},
				},
			},
			want: &dvweb.Lookup{
				Headers: shared.Headers{
					{Key: "bish", Value: "bar"},
					{Key: "foo", Value: "bong"},
				},
			},
			secretRefs: shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
					{OldIndex: new(1)},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newService := &Service{DeployedVersionLookup: tc.deployedVersion}
			oldService := &Service{DeployedVersionLookup: tc.otherDV}

			newService.giveSecretsDeployedVersion(oldService.DeployedVersionLookup, &tc.secretRefs)

			gotDV := newService.DeployedVersionLookup
			switch {
			case gotDV == nil && tc.want != nil:
				t.Fatal("Service.DeployedVersionLookup mismatch\ngot:  nil\nwant: non-nil")
			case gotDV != nil && tc.want == nil:
				t.Fatal("Service.DeployedVersionLookup mismatch\ngot:  non-nil\nwant: nil")
			case gotDV == nil:
				return
			}
			var gotBasicAuth, expectedBasicAuth *dvweb.BasicAuth
			var gotHeaders, expectedHeaders shared.Headers
			if gotWebDV, ok := gotDV.(*dvweb.Lookup); ok {
				gotBasicAuth, gotHeaders = gotWebDV.BasicAuth, gotWebDV.Headers
			}
			if wantWebDV, ok := tc.want.(*dvweb.Lookup); ok {
				expectedBasicAuth, expectedHeaders = wantWebDV.BasicAuth, wantWebDV.Headers
			}
			switch {
			case gotBasicAuth != nil && expectedBasicAuth == nil:
				t.Error("Lookup.BasicAuth mismatch\ngot:  non-nil\nwant: nil")
			case gotBasicAuth == nil && expectedBasicAuth != nil:
				t.Error("Lookup.BasicAuth mismatch\ngot:  nil\nwant: non-nil")
			case gotBasicAuth != nil && gotBasicAuth.Password != expectedBasicAuth.Password:
				t.Errorf(
					"Lookup.BasicAuth.Password mismatch\ngot:  %q\nwant: %q",
					gotBasicAuth.Password, expectedBasicAuth.Password,
				)
			}
			if testErr := test.AssertSlicesEqual(t, gotHeaders, expectedHeaders, "Lookup.Headers"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestService_GiveSecretsNotify(t *testing.T) {
	notifyCfg := shoutrrrtest.PlainConfig()

	tests := []struct {
		name                string
		notify, otherNotify shoutrrr.Shoutrrrs
		secretRefs          map[string]shared.OldStringIndex
		want                shoutrrr.Shoutrrrs
	}{
		{
			name:   "nil/notify",
			notify: nil,
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{},
			want:       nil,
		},
		{
			name: "nil/oldNotifiers",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: nil,
			secretRefs:  map[string]shared.OldStringIndex{},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/nil",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: nil,
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/none",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/no matches",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"bish": {OldIndex: "bash"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/referencing empty index",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: ""},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/referencing index that doesn't exist",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "baz"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/url_fields.altid",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/url_fields.apikey",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/url_fields.apikey swap vars",
			notify: shoutrrr.Shoutrrrs{
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "shazam",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"bar": {OldIndex: "foo"},
				"foo": {OldIndex: "bar"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "shazam",
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"apikey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/url_fields.botkey",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"botkey": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"botkey": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"botkey": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"botkey": "something",
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"botkey": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/url_fields.password",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"password": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"password": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"password": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"password": "something",
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"password": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/url_fields.token",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"token": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"token": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"token": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"token": "something",
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"token": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/url_fields.tokena",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"tokena": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"tokena": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"tokena": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"tokena": "something",
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"tokena": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/url_fields.tokenb",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"tokenb": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"tokenb": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"tokenb": "something",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"tokenb": "something",
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"tokenb": "yikes",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/url_fields.host ignored as SecretValue",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"host": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"host": "https://example.com",
					},
					nil,
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"host": "https://example.com/foo",
					},
					nil,
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"host": util.SecretValue,
					},
					nil,
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"host": "https://example.com",
					},
					nil,
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/params.devices",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil, nil,
					map[string]string{
						"devices": util.SecretValue,
					},
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil, nil,
					map[string]string{
						"devices": "yikes",
					},
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					nil,
					map[string]string{
						"devices": "something",
					},
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					nil,
					map[string]string{
						"devices": "something",
					},
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					nil,
					map[string]string{
						"devices": "yikes",
					},
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/params.avatar ignored as SecretValue",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					nil,
					map[string]string{
						"avatar": util.SecretValue,
					},
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					nil,
					map[string]string{
						"avatar": "https://example.com",
					},
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					nil,
					map[string]string{
						"avatar": "https://example.com/fooo",
					},
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					nil,
					map[string]string{
						"avatar": util.SecretValue,
					},
					nil, nil, nil,
				),
				"bar": shoutrrr.New(
					nil,
					"", "",
					nil,
					nil,
					map[string]string{
						"avatar": "https://example.com",
					},
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/ALL",
			notify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"altid":    util.SecretValue,
						"apikey":   util.SecretValue,
						"botkey":   util.SecretValue,
						"password": util.SecretValue,
						"token":    util.SecretValue,
						"tokena":   util.SecretValue,
						"tokenb":   util.SecretValue,
					},
					map[string]string{
						"devices": util.SecretValue,
					},
					nil, nil, nil,
				),
			},
			otherNotify: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"altid":    "whoosh",
						"apikey":   "foo",
						"botkey":   "bar",
						"password": "baz",
						"token":    "bish",
						"tokena":   "bosh",
						"tokenb":   "bash",
					},
					map[string]string{
						"devices": "id1,id2",
					},
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.OldStringIndex{
				"foo": {OldIndex: "foo"},
			},
			want: shoutrrr.Shoutrrrs{
				"foo": shoutrrr.New(
					nil,
					"", "",
					nil,
					map[string]string{
						"altid":    "whoosh",
						"apikey":   "foo",
						"botkey":   "bar",
						"password": "baz",
						"token":    "bish",
						"tokena":   "bosh",
						"tokenb":   "bash",
					},
					map[string]string{
						"devices": "id1,id2",
					},
					nil, nil, nil,
				),
			},
		},
	}

	for _, tc := range tests {
		newService := &Service{Notify: tc.notify}
		newService.Status.Init(
			len(newService.Command), len(newService.Notify), len(newService.Webhook),
			status.ServiceInfo{
				ID: tc.name,
			},
			&dashboard.Options{},
		)
		newService.Notify.Init(&newService.Status, notifyCfg)

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newService.giveSecretsNotify(tc.otherNotify, tc.secretRefs)

			gotNotify := newService.Notify
			if gotStr, wantStr := gotNotify.String(""), tc.want.String(""); gotStr != wantStr {
				t.Errorf("Shoutrrrs.String() mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}
		})
	}
}

func TestService_GiveSecretsWebhook(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	tests := []struct {
		name                  string
		webhook, otherWebhook webhook.Webhooks
		secretRefs            map[string]shared.WHSecretRef
		want                  webhook.Webhooks
	}{
		{
			name:    "nil/webhook",
			webhook: nil,
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{},
			want:       nil,
		},
		{
			name: "nil/otherWebhook",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: nil,
			secretRefs:   map[string]shared.WHSecretRef{},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "nil/secretRefs",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: nil,
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "no secretRefs",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "no matching secretRefs",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{
				"bish": {OldIndex: "bash"},
			},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/empty index",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"whoosh",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {OldIndex: ""},
				"bar": {OldIndex: ""},
			},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"whoosh",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/index that doesn't exist",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"whoosh",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {OldIndex: "bash"},
				"bar": {OldIndex: ""},
			},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"whoosh",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/secret",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"whoosh",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {OldIndex: "foo"},
				"bar": {OldIndex: ""},
			},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"whoosh",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "secretRefs/secret swap vars",
			webhook: webhook.Webhooks{
				"bar": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					util.SecretValue,
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"whoosh",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{
				"bar": {OldIndex: "foo"},
				"foo": {OldIndex: "bar"},
			},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"whoosh",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil, nil,
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"shazam",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "headers/no secretRefs",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "baz"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bing"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bang"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "baz"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "headers/no header secretRefs",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "baz"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bing"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bang"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {OldIndex: "foo"},
				"bar": {OldIndex: "bar"},
			},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "baz"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "headers/old secrets unwanted",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bar"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "baz"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bing"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bang"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {
					OldIndex: "foo",
					Headers: []shared.OldIntIndex{
						{OldIndex: new(0)},
					},
				},
				"bar": {
					OldIndex: "bar",
					Headers: []shared.OldIntIndex{
						{OldIndex: new(0)},
					},
				},
			},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bar"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "baz"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "headers/secretRefs some indices out of range",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bish", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bang", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bash"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {
					OldIndex: "foo",
					Headers: []shared.OldIntIndex{
						{OldIndex: new(5)}, {OldIndex: new(1)},
					},
				},
				"bar": {
					OldIndex: "bar",
					Headers: []shared.OldIntIndex{
						{OldIndex: new(0)}, {OldIndex: new(2)},
					},
				},
			},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bish", Value: "bash"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "headers/use all secrets",
			webhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bish", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bang", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bash"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {
					OldIndex: "foo",
					Headers: []shared.OldIntIndex{
						{OldIndex: new(0)},
						{OldIndex: new(1)},
					},
				},
				"bar": {
					OldIndex: "bar",
					Headers: []shared.OldIntIndex{
						{OldIndex: new(0)},
						{OldIndex: new(1)},
					},
				},
			},
			want: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bash"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
		{
			name: "headers/secretRefs swap names of webhook",
			webhook: webhook.Webhooks{
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bish", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bang", Value: util.SecretValue},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			otherWebhook: webhook.Webhooks{
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bash"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
			secretRefs: map[string]shared.WHSecretRef{
				"bar": {
					OldIndex: "foo",
					Headers: []shared.OldIntIndex{
						{OldIndex: new(0)},
						{OldIndex: new(1)},
					},
				},
				"foo": {
					OldIndex: "bar",
					Headers: []shared.OldIntIndex{
						{OldIndex: new(0)},
						{OldIndex: new(1)},
					},
				},
			},
			want: webhook.Webhooks{
				"bar": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bash"},
					},
					"",
					nil, nil,
					"bar",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
				"foo": webhook.New(
					nil,
					webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"},
					},
					"",
					nil, nil,
					"foo",
					nil, webhook.Notifiers{}, nil,
					"",
					nil,
					"", "",
					nil, nil, nil,
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var whStr string
			if len(tc.webhook) != 0 {
				whStr = "  " + strings.ReplaceAll(tc.webhook.String(), "\n", "\n  ")
			}
			newService := test.Must(t, func() (*Service, error) {
				return DecodeService(
					"yaml", []byte(test.TrimYAML(`
						latest_version:
							type: github
							url: `+test.ArgusGitHubRepo+`
						webhook:
						`+whStr+`
					`)),
					tc.name,
					svcCfg,
					notifyCfg,
					whCfg,
				)
			})

			if tc.otherWebhook != nil {
				otherServiceStatus := status.Status{}
				otherServiceStatus.Init(
					0, 0, len(tc.otherWebhook),
					status.ServiceInfo{
						ID: "otherService",
					},
					&dashboard.Options{},
				)
				tc.otherWebhook.Init(
					&otherServiceStatus,
					whCfg,
					nil,
					new("10m"),
				)
			}

			newService.giveSecretsWebhook(tc.otherWebhook, tc.secretRefs)

			gotWebhook := newService.Webhook
			if gotStr, wantStr := gotWebhook.String(), tc.want.String(); gotStr != wantStr {
				t.Errorf("Webhooks.String() mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}
		})
	}
}
