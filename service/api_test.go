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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/nicholas-fedor/shoutrrr/pkg/services/chat/teams"
	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config/decode"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	dvtest "github.com/release-argus/Argus/service/deployed_version/test"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	whtest "github.com/release-argus/Argus/webhook/test"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	dvweb "github.com/release-argus/Argus/service/deployed_version/types/web"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	lvgithub "github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/webhook"
)

func TestOldSecretRefs_UnmarshalJSON(t *testing.T) {
	// GIVEN: oldSecretRefs to unmarshal from JSON.
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

			// WHEN: UnmarshalJSON is called.
			var v oldSecretRefs
			if _, testErr := test.AssertUnmarshal(
				t,
				"json", tc.data,
				&v,
				tc.errRegex,
				func(v *oldSecretRefs) string { return decode.ToYAMLString(v, "") },
				tc.want,
				packageName,
				"oldSecretRefs",
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

	// GIVEN: a payload and the Service defaults.
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
			name:     "empty payload",
			payload:  "",
			errRegex: `^latest_version and\/or deployed_version required$`,
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
					"active=nil stays nil",
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
					"Require.Docker stays if have Type&Image&Tag",
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
					LatestVersion: lvbase.Defaults{
						Type: lvgithub.Type,
						Require: filter.RequireDefaults{
							Docker: docker.Defaults{
								Type: "ghcr",
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
					"url": "` + test.LookupPlain["url_valid"] + `",
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
							url: `+test.LookupPlain["url_valid"]+`
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
					"Give Notify secrets",
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
						"name": "teams",
						"type": "teams",
						"url_fields": {
							"altid": "` + util.SecretValue + `",
							"group":      "` + strings.Repeat("g", teams.UUID4Length) + `",
							"groupowner": "` + strings.Repeat("o", teams.UUID4Length) + `",
							"tenant":     "` + strings.Repeat("t", teams.UUID4Length) + `",
							"extraid":     "eID"
						},
						"params": {
							"host": "example.webhook.office.com"
						},
						"old_index": "teams-initial"
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
							teams:
								type: teams
								url_fields:
									altid: `+strings.Repeat("a", teams.HashLength)+`
									group: `+strings.Repeat("g", teams.UUID4Length)+`
									groupowner: `+strings.Repeat("o", teams.UUID4Length)+`
									tenant: `+strings.Repeat("t", teams.UUID4Length)+`
									extraid: eID
								params:
									host: example.webhook.office.com
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
							teams-initial:
								type: teams
								url_fields:
									altid: `+strings.Repeat("a", teams.HashLength)+`
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
			name: "Give WebHook secrets",
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
					"Give WebHook secrets",
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
					"Give WebHook secrets",
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
					"url": "` + test.LookupWithHeaderAuth["url_valid"] + `",
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
						"name": "teams",
						"type": "teams",
						"url_fields": {
							"altid": "` + util.SecretValue + `",
							"group":      "` + strings.Repeat("g", teams.UUID4Length) + `",
							"groupowner": "` + strings.Repeat("o", teams.UUID4Length) + `",
							"tenant":     "` + strings.Repeat("t", teams.UUID4Length) + `",
							"extraid":     "eID"
						},
						"params": {
							"host": "example.webhook.office.com"
						},
						"old_index": "teams-initial"
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
							url: `+test.LookupWithHeaderAuth["url_valid"]+`
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
							teams:
								type: teams
								url_fields:
									altid: `+strings.Repeat("a", teams.HashLength)+`
									group: `+strings.Repeat("g", teams.UUID4Length)+`
									groupowner: `+strings.Repeat("o", teams.UUID4Length)+`
									tenant: `+strings.Repeat("t", teams.UUID4Length)+`
									extraid: eID
								params:
									host: example.webhook.office.com
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
					"active=nil stays nil",
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
							teams-initial:
								type: teams
								url_fields:
									altid: `+strings.Repeat("a", teams.HashLength)+`
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
					"active=nil stays nil",
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
					"alpha":   {Base: shoutrrr.Base{Type: "smtp"}},
					"bravo":   {Base: shoutrrr.Base{Type: "smtp"}},
					"charlie": {Base: shoutrrr.Base{Type: "smtp"}},
				},
				Defaults: shoutrrr.ShoutrrrsDefaults{
					"smtp": {
						Base: shoutrrr.Base{
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
			name: "WebHookFromDefaults",
			payload: test.TrimJSON(`{
				"latest_version": {
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `"
				}
			}`),
			svcCfg: DefaultsConfig{
				Soft: &Defaults{
					WebHook: map[string]struct{}{
						"alpha":   {},
						"bravo":   {},
						"charlie": {},
					},
				},
			},
			whCfg: webhook.Config{
				Defaults: &webhook.Defaults{
					Base: webhook.Base{
						URL:    "https://example.com/github",
						Secret: "something",
					},
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
					"WebHookFromDefaults",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name: "CommandFromDefaults + NotifyFromDefaults + WebHookFromDefaults",
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
					WebHook: map[string]struct{}{
						"alpha":   {},
						"bravo":   {},
						"charlie": {},
					},
				},
			},
			notifyCfg: shoutrrr.Config{
				Root: shoutrrr.ShoutrrrsDefaults{
					"alpha":   {Base: shoutrrr.Base{Type: "smtp"}},
					"bravo":   {Base: shoutrrr.Base{Type: "smtp"}},
					"charlie": {Base: shoutrrr.Base{Type: "smtp"}},
				},
				Defaults: shoutrrr.ShoutrrrsDefaults{
					"smtp": {
						Base: shoutrrr.Base{
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
			},
			whCfg: webhook.Config{
				Defaults: &webhook.Defaults{
					Base: webhook.Base{
						URL:    "https://example.com/github",
						Secret: "something",
					},
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
					"All FromDefaults",
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Convert the string payload to a ReadCloser.
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
					len(tc.oldService.Command), len(tc.oldService.Notify), len(tc.oldService.WebHook),
					status.ServiceInfo{
						ID:         tc.oldService.ID,
						Name:       tc.oldService.Name,
						ServiceURL: "https://example.com/service/url",
					},
					&tc.oldService.Dashboard,
				)
				tc.oldService.init(notifyCfg, whCfg, nil, nil, nil)
			}

			// WHEN: we call FromPayload.
			svc, err := FromPayload(
				tc.oldService,
				&payload,
				tc.svcCfg, tc.notifyCfg, tc.whCfg,
				logx.LogFrom{Primary: tc.name},
			)

			prefix := fmt.Sprintf(
				"%s\nFromPayload(%q)",
				packageName, tc.payload,
			)

			// THEN: we get an error if the payload is invalid.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s\nerror mismatch from\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: we should get a new Service otherwise.
			gotStr, wantStr := svc.String(""), tc.want.String("")
			if gotStr != wantStr {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}

			// AND: the Service fromDefaults is set when expected.
			if svc.CommandFromDefaults != tc.wantFromDefaults.command {
				t.Errorf(
					"%s CommandFromDefaults mismatch\ngot:  %v\nwant: %v",
					prefix, svc.CommandFromDefaults, tc.wantFromDefaults.command,
				)
			} else if svc.CommandFromDefaults {
				if testErr := test.AssertSlicesEqualFunc(
					t,
					svc.Command,
					tc.svcCfg.Soft.Command,
					func(a, b command.Command) bool {
						return a.JSON() == b.JSON()
					},
					prefix,
					"Command",
				); testErr != nil {
					t.Errorf(
						"%s CommandFromDefaults=true, Command should match defaults:\n%v",
						prefix, testErr,
					)
				}
			}
			if svc.NotifyFromDefaults != tc.wantFromDefaults.notify {
				t.Errorf(
					"%s NotifyFromDefaults mismatch\ngot:  %v\nwant: %v",
					prefix, svc.NotifyFromDefaults, tc.wantFromDefaults.notify,
				)
			} else if svc.NotifyFromDefaults {
				gotKeys := util.SortedKeys(svc.Notify)
				wantKeys := util.SortedKeys(tc.svcCfg.Soft.Notify)
				if !util.AreSlicesEqual(gotKeys, wantKeys) {
					t.Errorf(
						"%s NotifyFromDefaults=true, Notify should match defaults:\ngot:  %v\nwant: %v",
						prefix, gotKeys, wantKeys,
					)
				}
			}
			if svc.WebHookFromDefaults != tc.wantFromDefaults.webhook {
				t.Errorf(
					"%s WebHookFromDefaults mismatch\ngot:  %v\nwant: %v",
					prefix, svc.WebHookFromDefaults, tc.wantFromDefaults.webhook,
				)
			} else if svc.WebHookFromDefaults {
				gotKeys := util.SortedKeys(svc.WebHook)
				wantKeys := util.SortedKeys(tc.svcCfg.Soft.WebHook)
				if !util.AreSlicesEqual(gotKeys, wantKeys) {
					t.Errorf(
						"%s WebHookFromDefaults=true, WebHook should match defaults:\ngot:  %v\nwant: %v",
						prefix, gotKeys, wantKeys,
					)
				}
			}
		})
	}
}

func TestFromPayload__NoServiceCreated(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	// GIVEN: the function that decodes a Service payload returns nil.
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

	// AND: a valid payload to create a service.
	payloadStr := `{
		"latest_version": {
			"type": "github",
			"url": "` + test.ArgusGitHubRepo + `"
		}
	}`
	reader := bytes.NewReader([]byte(payloadStr))
	payload := io.NopCloser(reader)

	// WHEN: we call FromPayload.
	_, err := FromPayload(
		nil,
		&payload,
		svcCfg, notifyCfg, whCfg,
		logx.LogFrom{Primary: "TestFromPayload_noServiceCreated"},
	)

	// THEN: we should get an error.
	prefix := fmt.Sprintf("%s\nFromPayload()", packageName)
	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Fatalf(
			"%s error mismatch\ngot:  %q\nwant: %q",
			prefix, e, errRegex,
		)
	}
}

func TestFromPayload_ReadFromFail(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()
	// GIVEN: an invalid payload.
	payloadStr := "this is a long payload"
	payload := io.NopCloser(bytes.NewReader([]byte(payloadStr)))
	payload = http.MaxBytesReader(nil, payload, 5)

	// WHEN: we call Decode.
	_, err := FromPayload(
		&Service{},
		&payload,
		svcCfg,
		notifyCfg,
		whCfg,
		logx.LogFrom{},
	)

	// THEN: we should get an error.
	if err == nil {
		t.Errorf(
			"%s\nFromPayload() gave no error with an invalid payload",
			packageName,
		)
	}
}

func TestService_CheckFetches(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()
	id := "TestService_CheckFetches"

	// GIVEN: a Service.
	testLV := lvtest.Lookup(t, "url", false)
	_, _ = testLV.Query(false, logx.LogFrom{})
	testDVL := dvtest.Lookup(t, "url", false, "")
	_ = testDVL.Query(false, logx.LogFrom{})
	tests := []struct {
		name                                      string
		svc                                       *Service
		startLatestVersion, wantLatestVersion     string
		startDeployedVersion, wantDeployedVersion string
		errRegex                                  string
	}{
		{
			name: "have LatestVersion, nil DeployedVersionLookup",
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
			name: "have LatestVersion, have DeployedVersionLookup",
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
			wantLatestVersion: testLV.GetStatus().LatestVersion(),
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

			// WHEN: we call CheckFetches.
			err := tc.svc.CheckFetches()

			prefix := fmt.Sprintf("%s\nService.CheckFetches()", packageName)

			// THEN: the error is as want.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: we get the want LatestVersion.
			if got := tc.svc.Status.LatestVersion(); got != tc.wantLatestVersion {
				t.Errorf(
					"%sStatus.LatestVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, tc.wantLatestVersion, got,
				)
			}

			// AND: we get the want DeployedVersion.
			if got := tc.svc.Status.DeployedVersion(); got != tc.wantDeployedVersion {
				t.Errorf(
					"%s Status.DeployedVersion()\ngot:  %q\nwant: %q",
					prefix, tc.wantDeployedVersion, got,
				)
			}
			want := 0
			if got := len(tc.svc.Status.AnnounceChannel); got != want {
				t.Errorf(
					"%s Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					prefix, got, want,
				)
			}
		})
	}
}

func TestService_GiveSecrets(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()
	id := "TestService_GiveSecrets"

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

	// GIVEN: a Service that may have secrets in it referencing those in another Service.
	tests := []struct {
		name            string
		svc, oldService *Service
		statusTests     statusTests
		commandTests    commandTests
		webhookTests    webhookTests
		secretRefs      oldSecretRefs
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
			secretRefs: oldSecretRefs{},
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
			secretRefs: oldSecretRefs{},
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
			secretRefs: oldSecretRefs{},
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
					"foo": test.Ptr(false),
					"bar": test.Ptr(true),
				},
			},
			secretRefs: oldSecretRefs{},
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
			secretRefs: oldSecretRefs{
				DeployedVersionLookup: shared.VSecretRef{
					Headers: []shared.OldIntIndex{
						{OldIndex: test.Ptr(0)},
						{OldIndex: test.Ptr(1)},
					},
				},
				Notify: map[string]shared.OldStringIndex{
					"foo": {OldIndex: "foo"},
					"bar": {OldIndex: "bar"},
				},
				WebHook: map[string]shared.WHSecretRef{
					"foo": {OldIndex: "foo"},
					"bar": {OldIndex: "bar"},
				},
			},
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
			name: "unchanged LatestVersion.URL retains Status.LatestVersion",
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
			statusTests: statusTests{
				oldLatestVersion:               "1.2.3",
				expectedLatestVersion:          "1.2.3",
				oldLatestVersionTimestamp:      time.Now().Format(time.RFC3339),
				expectedLatestVersionTimestamp: time.Now().Format(time.RFC3339),
			},
			secretRefs: oldSecretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name: "changed LatestVersion.URL loses Status.LatestVersion",
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
				oldLatestVersion:          "1.2.3",
				oldLatestVersionTimestamp: time.Now().Format(time.RFC3339),
			},
			secretRefs: oldSecretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name: "unchanged DeployedVersion.URL retains Status.DeployedVersion",
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
				oldDeployedVersionTimestamp:      time.Now().Format(time.RFC3339),
				expectedDeployedVersionTimestamp: time.Now().Format(time.RFC3339),
			},
			secretRefs: oldSecretRefs{},
			want: test.TrimYAML(`
				deployed_version:
					type: url
					url: https://example.com
			`),
		},
		{
			name: "changed DeployedVersion.URL loses Status.DeployedVersion",
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
				oldDeployedVersionTimestamp: time.Now().Format(time.RFC3339),
			},
			secretRefs: oldSecretRefs{},
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
			name: "unchanged WebHook retains Failed",
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
			secretRefs: oldSecretRefs{
				WebHook: map[string]shared.WHSecretRef{
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
					"test": test.Ptr(true),
				},
				expectedFails: map[string]*bool{
					"test": test.Ptr(true),
				},
			},
		},
		{
			name: "changed WebHook loses Failed",
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
			secretRefs: oldSecretRefs{},
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
					"test": test.Ptr(true),
				},
			},
		},
		{
			name: "unchanged Command retains Failed",
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
			secretRefs: oldSecretRefs{},
			commandTests: commandTests{
				oldFails: []*bool{
					test.Ptr(true),
				},
				expectedFails: []*bool{
					test.Ptr(true),
				},
			},
		},
		{
			name: "changed Command loses Failed",
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
			secretRefs: oldSecretRefs{},
			want: test.TrimYAML(`
				latest_version:
					type: github
					url: ` + test.ArgusGitHubRepo + `
				command:
					- - ls
						- -lah
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.oldService != nil {
				for k, v := range tc.commandTests.oldFails {
					if v != nil {
						tc.oldService.Status.Fails.Command.Set(k, *v)
					}
				}
				for k, v := range tc.webhookTests.oldFails {
					tc.oldService.Status.Fails.WebHook.Set(k, v)
				}
			}

			// WHEN: we call giveSecrets.
			tc.svc.giveSecrets(tc.oldService, tc.secretRefs)

			prefix := fmt.Sprintf("%s\nService.giveSecrets()", packageName)

			// THEN: we should get a Service with the secrets from the old Service.
			gotService := tc.svc
			gotServiceStr := gotService.String("")
			if gotServiceStr != tc.want {
				t.Errorf(
					"%s secrets weren't passed on\ngot:  %q\nwant: %q",
					prefix, gotServiceStr, tc.want,
				)
			}

			if gotService.WebHook != nil {
				var expectedWH string
				for name := range gotService.WebHook {
					expectedWH = name
					break
				}
				// Expecting `Failed` to be carried over.
				for key := range tc.webhookTests.expectedFails {
					want := tc.webhookTests.expectedFails[key]
					got := gotService.WebHook[expectedWH].DidFail()
					wantStr := test.StringifyPtr(want)
					gotStr := test.StringifyPtr(got)
					if gotStr != wantStr {
						t.Errorf(
							"%s stringified mismatch on .Failed\ngot:  %q\nwant: %q",
							packageName, gotStr, wantStr,
						)
					}
				}
			}
		})
	}
}

func TestService_GiveSecretsLatestVersion(t *testing.T) {
	lvCfg := lvtest.PlainDefaultsConfig(t)
	type otherData struct {
		githubData            *lvgithub.Data
		githubDataTransformed bool
	}

	// GIVEN: a LatestVersion that may have secrets in it referencing those in another LatestVersion.
	githubData := lvgithub.Data{}
	githubData.SetETag("shazam")
	tests := []struct {
		name                   string
		latestVersion, otherLV latestver.Lookup
		secretRefs             shared.VSecretRef
		expected               latestver.Lookup
		otherData              otherData
	}{
		{
			name: "nil oldLatestVersion",
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
			name: "nil Require",
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
								token: `+util.SecretValue+`
					`)),
					nil,
					&status.Status{},
					lvCfg,
				)
			}),
		},
		{
			name: "githubData carried over if type still 'github'",
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
			otherData: otherData{
				githubData:            &githubData,
				githubDataTransformed: true,
			},
		},
		{
			name: "githubData not carried over if type wasn't 'github'",
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
			otherData: otherData{
				githubData:            &githubData,
				githubDataTransformed: false,
			},
		},
		{
			name: "githubData not carried over if type no longer 'github'",
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
			otherData: otherData{
				githubData:            &githubData,
				githubDataTransformed: false,
			},
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
					{OldIndex: test.Ptr(0)},
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
			name: "referencing old Header value with no refs",
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
					{OldIndex: test.Ptr(0)},
					{OldIndex: test.Ptr(1)},
				},
			},
		},
		{
			name: "referencing old Header value",
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
					{OldIndex: test.Ptr(0)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "referencing old Header value that doesn't exist",
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
					{OldIndex: test.Ptr(1)},
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
					{OldIndex: test.Ptr(1)},
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
					{OldIndex: test.Ptr(0)},
					{OldIndex: test.Ptr(1)},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newService := &Service{LatestVersion: tc.latestVersion}
			oldService := &Service{LatestVersion: tc.otherLV}

			// WHEN: we call giveSecretsLatestVersion.
			newService.giveSecretsLatestVersion(oldService.LatestVersion, &tc.secretRefs)

			prefix := fmt.Sprintf("%s\nService.giveSecretsLatestVersion()", packageName)

			// THEN: we should get a Service with the secrets from the other Service.
			gotLV := newService.LatestVersion

			// Only GitHub types have AccessTokens.
			if gotLatestVersion, ok := gotLV.(*lvgithub.Lookup); ok {
				if hadLatestVersion, ok := tc.latestVersion.(*lvgithub.Lookup); ok {
					gotAccessToken := gotLatestVersion.AccessToken
					expectedAccessToken := hadLatestVersion.AccessToken
					if gotAccessToken != expectedAccessToken {
						t.Errorf(
							"%s .AccessToken mismatch\ngot:  %q\nwant: %q",
							prefix, gotAccessToken, expectedAccessToken,
						)
					}
				}
			}

			// Require:
			var gotRequire *filter.Require
			var expectedRequire *filter.Require
			// 	Got:
			if gotLatestVersion, ok := gotLV.(*lvgithub.Lookup); ok {
				gotRequire = gotLatestVersion.GetRequire()
			} else if gotLatestVersion, ok := gotLV.(*lvweb.Lookup); ok {
				gotRequire = gotLatestVersion.GetRequire()
			}
			// 	Expected:
			if expectedLatestVersion, ok := tc.expected.(*lvgithub.Lookup); ok {
				expectedRequire = expectedLatestVersion.GetRequire()
			} else if expectedLatestVersion, ok := tc.expected.(*lvweb.Lookup); ok {
				expectedRequire = expectedLatestVersion.GetRequire()
			}
			// newService has a nil Require, but want non-nil.
			if gotRequire == nil && expectedRequire != nil {
				t.Errorf("%s .Require mismatch\ngot:  nil\nwant: non-nil", prefix)

				// newService Require/Docker isn't nil when want is or vice-versa.
			} else if gotRequire != expectedRequire &&
				gotRequire.Docker != expectedRequire.Docker &&
				// newService doesn't have the want Token.
				gotRequire.Docker.GetAuth().GetTokenSelf() != expectedRequire.Docker.GetAuth().GetTokenSelf() {
				t.Errorf(
					"%s .Require.Docker.Token mismatch\ngot:  %q\nwant: %q",
					prefix,
					gotRequire.Docker.GetAuth().GetTokenSelf(),
					expectedRequire.Docker.GetAuth().GetTokenSelf(),
				)
			}

			// githubData:
			if expectedLatestVersion, ok := tc.expected.(*lvgithub.Lookup); ok {
				// Ensure gotLV is a *lvgithub.Lookup.
				if gotLatestVersion, ok := gotLV.(*lvgithub.Lookup); ok {
					got := gotLatestVersion.GetGitHubData().String()
					expected := expectedLatestVersion.GetGitHubData().String()
					if got != expected {
						t.Errorf(
							"%s .githubData mismatch\ngot:  %v\nwant: %v",
							prefix, got, expected,
						)
					}
				} else {
					t.Fatalf(
						"%s *lvgithub.Lookup type mismatch\ngot:  %T\nwant: github.Lookup",
						prefix, gotLV,
					)
				}
			}
			// Headers:
			var expectedHeaders shared.Headers
			if expectedLV, ok := tc.expected.(*lvweb.Lookup); ok {
				expectedHeaders = expectedLV.Headers
			}
			var gotHeaders shared.Headers
			if gotLV, ok := gotLV.(*lvweb.Lookup); ok {
				gotHeaders = gotLV.Headers
			}
			if !util.AreSlicesEqual(expectedHeaders, gotHeaders) {
				t.Errorf(
					"%s .Headers mismatch\ngot:  %+v\nwant: %+v",
					prefix, gotHeaders, expectedHeaders,
				)
			}
		})
	}
}

func TestService_GiveSecretsDeployedVersion(t *testing.T) {
	// GIVEN: a DeployedVersion that may have secrets in it referencing those in another DeployedVersion.
	tests := []struct {
		name                     string
		deployedVersion, otherDV deployedver.Lookup
		secretRefs               shared.VSecretRef
		want                     deployedver.Lookup
	}{
		{
			name:            "nil DeployedVersion",
			deployedVersion: nil,
			otherDV:         &dvweb.Lookup{},
			secretRefs:      shared.VSecretRef{},
			want:            nil,
		},
		{
			name: "nil OldDeployedVersion",
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
			name: "keep BasicAuth.Password",
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
			name: "give old BasicAuth.Password",
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
			name: "referencing default BasicAuth.Password",
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
			name: "referencing BasicAuth.Password that doesn't exist",
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
			name: "empty Headers",
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
			name: "Headers with index out of range",
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
					{OldIndex: test.Ptr(0)},
					{OldIndex: test.Ptr(1)},
				},
			},
		},
		{
			name: "Headers with SecretValue but nil index refs",
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
			name: "only changed Headers",
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
					{OldIndex: test.Ptr(0)},
				},
			},
		},
		{
			name: "only new/changed Headers",
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
					{OldIndex: test.Ptr(0)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "only new/changed Headers with want refs",
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
					{OldIndex: test.Ptr(0)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "only new/changed Headers with no refs",
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
			name: "referencing old Header value with no refs",
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
					{OldIndex: test.Ptr(0)},
					{OldIndex: test.Ptr(1)},
				},
			},
		},
		{
			name: "referencing old Header value",
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
					{OldIndex: test.Ptr(0)},
					{OldIndex: nil},
				},
			},
		},
		{
			name: "referencing old Header value that doesn't exist",
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
					{OldIndex: test.Ptr(1)},
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
					{OldIndex: test.Ptr(1)},
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
					{OldIndex: test.Ptr(0)},
					{OldIndex: test.Ptr(1)},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			newService := &Service{DeployedVersionLookup: tc.deployedVersion}
			oldService := &Service{DeployedVersionLookup: tc.otherDV}

			// WHEN: we call giveSecretsDeployedVersion.
			newService.giveSecretsDeployedVersion(oldService.DeployedVersionLookup, &tc.secretRefs)

			prefix := fmt.Sprintf("%s\nService.giveSecretsDeployedVersion()", packageName)

			// THEN: we should get a Service with the secrets from the other Service.
			gotDV := newService.DeployedVersionLookup
			if gotDV == tc.want {
				return
			}
			// Got/Expected nil but not both.
			if gotDV == nil && tc.want != nil ||
				gotDV != nil && tc.want == nil {
				t.Errorf(
					"%s nil/not-nil state unexpected\ngot:  %q\nwant: %q",
					prefix, gotDV.String(""), tc.want.String(""),
				)
			}
			// BasicAuth:
			var expectedBasicAuth *dvweb.BasicAuth
			if expectedLV, ok := tc.want.(*dvweb.Lookup); ok {
				expectedBasicAuth = expectedLV.BasicAuth
			}
			var gotBasicAuth *dvweb.BasicAuth
			if gotLV, ok := gotDV.(*dvweb.Lookup); ok {
				gotBasicAuth = gotLV.BasicAuth
			}
			if gotBasicAuth != expectedBasicAuth {
				if gotBasicAuth != nil && expectedBasicAuth == nil {
					t.Errorf(
						"%s .BasicAuth mismatch\ngot:  %q\nwant: nil",
						prefix, *gotBasicAuth,
					)
				} else if gotBasicAuth.Password != expectedBasicAuth.Password {
					t.Errorf(
						"%s .BasicAuth.Password mismatch\ngot:  %q\nwant: %q",
						prefix, util.DerefOrZero(gotBasicAuth), util.DerefOrZero(expectedBasicAuth),
					)
				}
			}
			// Headers:
			var expectedHeaders shared.Headers
			if expectedDV, ok := tc.want.(*dvweb.Lookup); ok {
				expectedHeaders = expectedDV.Headers
			}
			var gotHeaders shared.Headers
			if gotLV, ok := gotDV.(*dvweb.Lookup); ok {
				gotHeaders = gotLV.Headers
			}
			if !util.AreSlicesEqual(expectedHeaders, gotHeaders) {
				t.Errorf(
					"%s .Headers mismatch\ngot:  %+v\nwant: %+v",
					prefix, gotHeaders, expectedHeaders,
				)
			}
		})
	}
}

func TestService_GiveSecretsNotify(t *testing.T) {
	notifyCfg := shoutrrrtest.PlainConfig()

	// GIVEN: a NotifySlice that may have secrets in it referencing those in another NotifySliceSlice.
	tests := []struct {
		name                string
		notify, otherNotify shoutrrr.Shoutrrrs
		secretRefs          map[string]shared.OldStringIndex
		expected            shoutrrr.Shoutrrrs
	}{
		{
			name:   "nil NotifySlice",
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
			expected:   nil,
		},
		{
			name: "nil oldNotifies",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "nil secretRefs",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "no secretRefs",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "no matching secretRefs",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRef referencing empty index",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRef referencing index that doesn't exist",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - url_fields.altid",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - url_fields.apikey",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - url_fields.apikey swap vars",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - url_fields.apikey swap vars ignores notify order",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - url_fields.botkey",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - url_fields.password",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - url_fields.token",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - url_fields.tokena",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - url_fields.tokenb",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - url_fields.host ignored as SecretValue",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - params.devices",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - params.avatar ignored as SecretValue",
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
			expected: shoutrrr.Shoutrrrs{
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
			name: "secretRefs - ALL",
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
			expected: shoutrrr.Shoutrrrs{
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
			len(newService.Command), len(newService.Notify), len(newService.WebHook),
			status.ServiceInfo{
				ID: tc.name,
			},
			&dashboard.Options{},
		)
		// Give empty defaults and hardDefaults to the NotifySlice.
		newService.Notify.Init(&newService.Status, notifyCfg)

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: we call giveSecretsNotify.
			newService.giveSecretsNotify(tc.otherNotify, tc.secretRefs)

			prefix := fmt.Sprintf("%s\nService.giveSecretsNotify()", packageName)

			// THEN: we should get a NotifySlice with the secrets from the other Service.
			gotNotify := newService.Notify
			gotNotifyStr := gotNotify.String("")
			expectedStr := tc.expected.String("")
			if gotNotifyStr != expectedStr {
				t.Errorf(
					"%s secrets weren't passed on\ngot:  %+v\nwant: %+v",
					prefix, gotNotifyStr, expectedStr,
				)
			}
		})
	}
}

func TestService_GiveSecretsWebHook(t *testing.T) {
	svcCfg := plainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()
	// GIVEN: a WebHookSlice that may have secrets in it referencing those in another WebHookSliceSlice.
	tests := []struct {
		name                  string
		webhook, otherWebhook webhook.WebHooks
		secretRefs            map[string]shared.WHSecretRef
		want                  webhook.WebHooks
	}{
		{
			name:    "nil WebHookSlice",
			webhook: nil,
			otherWebhook: webhook.WebHooks{
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
			name: "nil otherWebHook",
			webhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			name: "nil secretRefs",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			name: "secretRef referencing empty index",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			name: "secretRef referencing index that doesn't exist",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			name: "secretRefs - secret",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			name: "secretRefs - secret swap vars",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			name: "secretRefs - secret swap vars ignores order sent",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			name: "headers - no secretRefs",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			name: "headers - no header secretRefs",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
			want: webhook.WebHooks{
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
			name: "headers - header secretRefs but old secrets unwanted",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
						{OldIndex: test.Ptr(0)},
					},
				},
				"bar": {
					OldIndex: "bar",
					Headers: []shared.OldIntIndex{
						{OldIndex: test.Ptr(0)},
					},
				},
			},
			want: webhook.WebHooks{
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
			name: "headers - header secretRefs, some indices out of range",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
						{OldIndex: test.Ptr(5)}, {OldIndex: test.Ptr(1)},
					},
				},
				"bar": {
					OldIndex: "bar",
					Headers: []shared.OldIntIndex{
						{OldIndex: test.Ptr(0)}, {OldIndex: test.Ptr(2)},
					},
				},
			},
			want: webhook.WebHooks{
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
			name: "headers - header secretRefs use all secrets",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
						{OldIndex: test.Ptr(0)},
						{OldIndex: test.Ptr(1)},
					},
				},
				"bar": {
					OldIndex: "bar",
					Headers: []shared.OldIntIndex{
						{OldIndex: test.Ptr(0)},
						{OldIndex: test.Ptr(1)},
					},
				},
			},
			want: webhook.WebHooks{
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
			name: "headers - header secretRefs, swap names of webhook",
			webhook: webhook.WebHooks{
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
			otherWebhook: webhook.WebHooks{
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
						{OldIndex: test.Ptr(0)},
						{OldIndex: test.Ptr(1)},
					},
				},
				"foo": {
					OldIndex: "bar",
					Headers: []shared.OldIntIndex{
						{OldIndex: test.Ptr(0)},
						{OldIndex: test.Ptr(1)},
					},
				},
			},
			want: webhook.WebHooks{
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
			// Other Service Status.Fails.
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
					test.Ptr("10m"),
				)
			}

			// WHEN: we call giveSecretsWebHook.
			newService.giveSecretsWebHook(tc.otherWebhook, tc.secretRefs)

			prefix := fmt.Sprintf("%s\nService.giveSecretsWebHook()", packageName)

			// THEN: we should get a WebHookSlice with the secrets from the other Service.
			gotWebHook := newService.WebHook
			if gotStr, wantStr := gotWebHook.String(), tc.want.String(); gotStr != wantStr {
				t.Errorf(
					"%s secrets weren't passed on\ngot:  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}
		})
	}
}
