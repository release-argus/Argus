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

package v1

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	lvtest "github.com/release-argus/Argus/service/latest_version/test"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	lvgithub "github.com/release-argus/Argus/service/latest_version/types/github"
	lvweb "github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/webhook"
)

func TestHTTP_Config(t *testing.T) {
	lvCfg := lvtest.PlainDefaultsConfig(t)

	file := t.Name() + ".yml"
	api := testAPI(t, file)

	webSettings := &config.Settings{
		Web: config.WebSettings{
			ListenHost: "127.0.0.1",
		},
	}
	// GHCR and Quay auth render as a token alone, so their username is absent from the JSON.
	requireYAML := []byte(test.TrimYAML(`
		docker:
			type: hub
			tag: t
			registry:
				ghcr:
					auth:
						username: something
						token: ghp_X
				hub:
					auth:
						username: something
						token: hub_X
				quay:
					auth:
						username: something
						token: quay_X
	`))
	serviceDefaults := &config.Defaults{
		Service: service.Defaults{
			Options: opt.Defaults{
				Interval: "1h",
			},
			LatestVersion: latestver.Defaults{
				GitHub: lvgithub.Defaults{
					AccessToken:   "foo",
					UsePreRelease: new(false),
				},
				URL: lvweb.Defaults{
					AllowInvalidCerts: new(true),
				},
				Common: lvbase.Defaults{
					Require: *test.Must(t, func() (*filter.RequireDefaults, error) {
						return filter.DecodeDefaults("yaml", requireYAML)
					}),
				},
			},
			Notify: map[string]struct{}{
				"n1": {},
			},
			Command: command.Commands{
				{"command", "arg1", "arg2"},
			},
			Webhook: map[string]struct{}{
				"wh1": {},
				"wh2": {},
				"wh3": {},
				"wh4": {},
			},
		},
	}
	notifiers := &shoutrrr.ShoutrrrsDefaults{
		"foo": shoutrrr.NewDefaults(
			"gotify",
			map[string]string{
				"message": "hello world",
			},
			map[string]string{
				"token": "secret123",
			},
			map[string]string{
				"title": "UPDATE",
			},
		),
	}
	webhooks := &webhook.WebhooksDefaults{
		"foo": test.Must(t, func() (*webhook.Defaults, error) {
			return webhook.DecodeDefaults(
				"yaml", []byte(test.TrimYAML(`
					allow_invalid_certs: true
					headers:
						- key: X-Header
						  value: value
					delay: 4s
					secret: something
					type: github
					url: https://example.com
				`)),
			)
		}),
	}

	webSettingsJSON := `{
		"web": {
			"listen_host": "127.0.0.1"
		}
	}`
	requireJSON := `{
		"docker": {
			"type": "hub",
			"tag": "t",
			"registry": {
				"ghcr": {
					"auth": {
						"token": ` + secretValueMarshaled + `
					}
				},
				"hub": {
					"auth": {
						"username": "something",
						"token": ` + secretValueMarshaled + `
					}
				},
				"quay": {
					"auth": {
						"token": ` + secretValueMarshaled + `
					}
				}
			}
		}
	}`
	serviceDefaultsJSON := `{
		"service": {
			"options": {
				"interval": "1h"
			},
			"latest_version": {
				"common": {
					"require": ` + requireJSON + `
				},
				"github": {
					"access_token": ` + secretValueMarshaled + `,
					"use_prerelease": false
				},
				"url": {
					"allow_invalid_certs": true
				}
			},
			"notify": ["n1"],
			"command": [
				["command","arg1","arg2"]
			],
			"webhook": [
				"wh1",
				"wh2",
				"wh3",
				"wh4"
			]
		}
	}`
	notifiersJSON := `{
		"foo": {
			"type": "gotify",
			"options": {
				"message": "hello world"
			},
			"url_fields": {
				"token": ` + secretValueMarshaled + `
			},
			"params": {
				"title": "UPDATE"
			}
		}
	}`
	webhooksJSON := `{
		"foo": {
			"type": "github",
			"url": "https://example.com",
			"allow_invalid_certs": true,
			"secret": ` + secretValueMarshaled + `,
			"headers": [
				{
					"key": "X-Header",
					"value": ` + secretValueMarshaled + `
				}
			],
			"delay": "4s"
		}
	}`

	tests := []struct {
		name     string
		settings *config.Settings
		defaults *config.Defaults
		notify   *shoutrrr.ShoutrrrsDefaults
		webhook  *webhook.WebhooksDefaults
		service  *service.Services
		order    *[]string
		wantBody string
	}{
		{
			name: "settings/only",
			settings: &config.Settings{
				Data: config.DataSettings{
					DatabaseFile: "somewhere.db",
					Readonly:     new(true),
				},
				Web: config.WebSettings{
					ListenHost:     "127.0.0.1",
					DisabledRoutes: []string{"version", "websocket"},
				},
			},
			wantBody: `
				{
					"settings": {
						"data": {
							"database_file": "somewhere.db",
							"readonly": true
						},
						"web": {
							"listen_host": "127.0.0.1",
							"disabled_routes": ["version", "websocket"]
						}
					}
				}`,
		},
		{
			name:     "settings/and defaults/only",
			settings: webSettings,
			defaults: &config.Defaults{
				Service: service.Defaults{
					Options: opt.Defaults{
						Interval: "1h",
					},
					LatestVersion: latestver.Defaults{
						Type: "github",
						GitHub: lvgithub.Defaults{
							AccessToken:   "foo",
							UsePreRelease: new(false),
						},
						URL: lvweb.Defaults{
							AllowInvalidCerts: new(true),
						},
						Common: lvbase.Defaults{
							Require: *test.Must(t, func() (*filter.RequireDefaults, error) {
								return filter.DecodeDefaults("yaml", requireYAML)
							}),
						},
					},
				},
			},
			wantBody: `
				{
					"settings": ` + webSettingsJSON + `,
					"defaults": {
						"service": {
							"options": {
								"interval": "1h"
							},
							"latest_version": {
								"type": "github",
								"common": {
									"require": ` + requireJSON + `
								},
								"github": {
									"access_token": ` + secretValueMarshaled + `,
									"use_prerelease": false
								},
								"url": {
									"allow_invalid_certs": true
								}
							}
						}
					}
				}`,
		},
		{
			name:     "settings/and defaults/with notify+command+webhook service defaults/only",
			settings: webSettings,
			defaults: serviceDefaults,
			wantBody: `{
				"settings": ` + webSettingsJSON + `,
				"defaults": ` + serviceDefaultsJSON + `
			}`,
		},
		{
			name:     "settings/and defaults/with notify+command+webhook service defaults/and notify",
			settings: webSettings,
			defaults: serviceDefaults,
			notify:   notifiers,
			wantBody: `{
				"settings": ` + webSettingsJSON + `,
				"defaults": ` + serviceDefaultsJSON + `,
				"notify": ` + notifiersJSON + `
			}`,
		},
		{
			name:     "settings/and defaults/with notify+command+webhook service defaults/and notify and webhook",
			settings: webSettings,
			defaults: serviceDefaults,
			notify:   notifiers,
			webhook:  webhooks,
			wantBody: `{
				"settings": ` + webSettingsJSON + `,
				"defaults": ` + serviceDefaultsJSON + `,
				"notify": ` + notifiersJSON + `,
				"webhook": ` + webhooksJSON + `
			}`,
		},
		{
			name:     "settings/and defaults/with notify+command+webhook service defaults/and notify and webhook and service",
			settings: webSettings,
			defaults: serviceDefaults,
			notify:   notifiers,
			webhook:  webhooks,
			service: &service.Services{
				"alpha": &service.Service{
					LatestVersion: test.Must(t, func() (latestver.Lookup, error) {
						return latestver.Decode(
							"yaml", []byte(test.TrimYAML(`
								type: github
								url: `+test.ArgusGitHubRepo+`
								access_token: aToken
							`)),
							nil,
							nil,
							lvCfg,
						)
					}),
				},
				"bravo": &service.Service{
					LatestVersion: test.Must(t, func() (latestver.Lookup, error) {
						return latestver.Decode(
							"yaml", []byte(test.TrimYAML(`
								type: url
								url: https://example.com/version
								allow_invalid_certs: true
							`)),
							nil,
							nil,
							lvCfg,
						)
					}),
				},
			},
			order: &[]string{"alpha", "bravo"},
			wantBody: `{
				"settings": ` + webSettingsJSON + `,
				"defaults": ` + serviceDefaultsJSON + `,
				"notify": ` + notifiersJSON + `,
				"webhook": ` + webhooksJSON + `,
				"service": {
					"alpha": {
						"latest_version": {
							"type": "github",
							"url": "` + test.ArgusGitHubRepo + `",
							"access_token": ` + secretValueMarshaled + `
						}
					},
					"bravo": {
						"latest_version": {
							"type": "url",
							"url": "https://example.com/version",
							"allow_invalid_certs": true
						}
					}
				},
				"order": [
					"alpha",
					"bravo"
				]
			}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: shares api.Config.

			api.Config.Settings = config.Settings{}
			api.Config.Defaults = config.Defaults{}
			api.Config.Notify = shoutrrr.ShoutrrrsDefaults{}
			api.Config.Webhook = webhook.WebhooksDefaults{}
			api.Config.Service = service.Services{}
			api.Config.Order = []string{}
			if tc.settings != nil {
				api.Config.Settings = *tc.settings
			}
			if tc.defaults != nil {
				api.Config.Defaults = *tc.defaults
			}
			if tc.notify != nil {
				api.Config.Notify = *tc.notify
			}
			if tc.webhook != nil {
				api.Config.Webhook = *tc.webhook
			}
			if tc.service != nil {
				api.Config.Service = *tc.service
			}
			if tc.order != nil {
				api.Config.Order = *tc.order
			}
			tc.wantBody = test.TrimJSON(tc.wantBody) + "\n"

			req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
			w := httptest.NewRecorder()
			api.httpConfig(w, req)
			resp := w.Result()
			t.Cleanup(func() { _ = resp.Body.Close() })

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() failed: %v", err)
			}
			if got := string(data); got != tc.wantBody {
				t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, tc.wantBody)
			}
		})
	}
}
