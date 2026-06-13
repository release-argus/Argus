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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/webhook"
)

func TestHTTP_Config(t *testing.T) {
	// GIVEN: defaults/hardDefaults.
	lvCfg := lvtest.PlainDefaultsConfig(t)

	// AND: an API and a request for the config.
	file := filepath.Join(t.TempDir(), "config.yml")
	api := testAPI(t, file)

	tests := []struct {
		name     string
		settings *config.Settings
		defaults *config.Defaults
		notify   *shoutrrr.ShoutrrrsDefaults
		webhook  *webhook.WebHooksDefaults
		service  *service.Services
		order    *[]string
		wantBody string
	}{
		{
			name: "settings",
			settings: &config.Settings{
				SettingsBase: config.SettingsBase{
					Web: config.WebSettings{
						ListenHost: "127.0.0.1",
					},
				},
			},
			wantBody: `
				{
					"settings": {
						"web": {
							"listen_host": "127.0.0.1"
						}
					}
				}`,
		},
		{
			name: "settings + defaults",
			settings: &config.Settings{
				SettingsBase: config.SettingsBase{
					Web: config.WebSettings{
						ListenHost: "127.0.0.1",
					},
				},
			},
			defaults: &config.Defaults{
				Service: service.Defaults{
					Options: opt.Defaults{
						Base: opt.Base{
							Interval: "1h",
						},
					},
					LatestVersion: lvbase.Defaults{
						AccessToken:       "foo",
						AllowInvalidCerts: test.Ptr(true),
						UsePreRelease:     test.Ptr(false),
						Require: *test.Must(t, func() (*filter.RequireDefaults, error) {
							return filter.DecodeDefaults(
								"yaml", []byte(test.TrimYAML(`
									docker:
										type: hub
										image: i
										tag: t
										registry:
											ghcr:
												image: iGHCR
												tag: tGHCR
												auth:
													username: something
													token: ghp_X
											hub:
												image: iHub
												tag: tHub
												auth:
													username: something
													token: hub_X
											quay:
												image: iQuay
												tag: tQuay
												auth:
													username: something
													token: quay_X
								`)),
							)
						}),
					},
				},
			},
			wantBody: `
				{
					"settings": {
						"web": {
							"listen_host": "127.0.0.1"
						}
					},
					"defaults": {
						"service": {
							"options": {
								"interval": "1h"
							},
							"latest_version": {
								"access_token": ` + secretValueMarshalled + `,
								"allow_invalid_certs": true,
								"use_prerelease": false,
								"require": {
									"docker": {
										"type": "hub",
										"image": "i",
										"tag": "t",
										"registry": {
											"ghcr": {
												"image": "iGHCR",
												"tag": "tGHCR",
												"auth": {
													"token": ` + secretValueMarshalled + `
												}
											},
											"hub": {
												"image": "iHub",
												"tag": "tHub",
												"auth": {
													"username": "something",
													"token": ` + secretValueMarshalled + `
												}
											},
											"quay": {
												"image": "iQuay",
												"tag": "tQuay",
												"auth": {
													"token": ` + secretValueMarshalled + `
												}
											}
										}
									}
								}
							}
						}
					}
				}`,
		},
		{
			name: "settings + defaults (with notify+command+webhook service defaults)",
			settings: &config.Settings{
				SettingsBase: config.SettingsBase{
					Web: config.WebSettings{
						ListenHost: "127.0.0.1",
					},
				},
			},
			defaults: &config.Defaults{
				Service: service.Defaults{
					Options: opt.Defaults{
						Base: opt.Base{
							Interval: "1h",
						},
					},
					LatestVersion: lvbase.Defaults{
						AccessToken:       "foo",
						AllowInvalidCerts: test.Ptr(true),
						UsePreRelease:     test.Ptr(false),
						Require: *test.Must(t, func() (*filter.RequireDefaults, error) {
							return filter.DecodeDefaults(
								"yaml", []byte(test.TrimYAML(`
									docker:
										type: hub
										image: i
										tag: t
										registry:
											ghcr:
												image: iGHCR
												tag: tGHCR
												auth:
													username: something
													token: ghp_X
											hub:
												image: iHub
												tag: tHub
												auth:
													username: something
													token: hub_X
											quay:
												image: iQuay
												tag: tQuay
												auth:
													username: something
													token: quay_X
								`)),
							)
						}),
					},
					Notify: map[string]struct{}{
						"n1": {},
					},
					Command: command.Commands{
						{"command", "arg1", "arg2"},
					},
					WebHook: map[string]struct{}{
						"wh1": {},
						"wh2": {},
						"wh3": {},
						"wh4": {},
					},
				},
			},
			wantBody: `
				{
					"settings": {
						"web": {
							"listen_host": "127.0.0.1"
						}
					},
					"defaults": {
						"service": {
							"options": {
								"interval": "1h"
							},
							"latest_version": {
								"access_token": ` + secretValueMarshalled + `,
								"allow_invalid_certs": true,
								"use_prerelease": false,
								"require": {
									"docker": {
										"type": "hub",
										"image": "i",
										"tag": "t",
										"registry": {
											"ghcr": {
												"image": "iGHCR",
												"tag": "tGHCR",
												"auth": {
													"token": ` + secretValueMarshalled + `
												}
											},
											"hub": {
												"image": "iHub",
												"tag": "tHub",
												"auth": {
													"username": "something",
													"token": ` + secretValueMarshalled + `
												}
											},
											"quay": {
												"image": "iQuay",
												"tag": "tQuay",
												"auth": {
													"token": ` + secretValueMarshalled + `
												}
											}
										}
									}
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
					}
				}`,
		},
		{
			name: "settings + defaults (with notify+command+webhook service defaults) + notify",
			settings: &config.Settings{
				SettingsBase: config.SettingsBase{
					Web: config.WebSettings{
						ListenHost: "127.0.0.1",
					},
				},
			},
			defaults: &config.Defaults{
				Service: service.Defaults{
					Options: opt.Defaults{
						Base: opt.Base{
							Interval: "1h",
						},
					},
					LatestVersion: lvbase.Defaults{
						AccessToken:       "foo",
						AllowInvalidCerts: test.Ptr(true),
						UsePreRelease:     test.Ptr(false),
						Require: *test.Must(t, func() (*filter.RequireDefaults, error) {
							return filter.DecodeDefaults(
								"yaml", []byte(test.TrimYAML(`
									docker:
										type: hub
										image: i
										tag: t
										registry:
											ghcr:
												image: iGHCR
												tag: tGHCR
												auth:
													username: something
													token: ghp_X
											hub:
												image: iHub
												tag: tHub
												auth:
													username: something
													token: hub_X
											quay:
												image: iQuay
												tag: tQuay
												auth:
													username: something
													token: quay_X
								`)),
							)
						}),
					},
					Notify: map[string]struct{}{
						"n1": {},
					},
					Command: command.Commands{
						{"command", "arg1", "arg2"},
					},
					WebHook: map[string]struct{}{
						"wh1": {},
						"wh2": {},
						"wh3": {},
						"wh4": {},
					},
				},
			},
			notify: &shoutrrr.ShoutrrrsDefaults{
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
			},
			wantBody: `
				{
					"settings": {
						"web": {
							"listen_host": "127.0.0.1"
						}
					},
					"defaults": {
						"service": {
							"options": {
								"interval": "1h"
							},
							"latest_version": {
								"access_token": ` + secretValueMarshalled + `,
								"allow_invalid_certs": true,
								"use_prerelease": false,
								"require": {
									"docker": {
										"type": "hub",
										"image": "i",
										"tag": "t",
										"registry": {
											"ghcr": {
												"image": "iGHCR",
												"tag": "tGHCR",
												"auth": {
													"token": ` + secretValueMarshalled + `
												}
											},
											"hub": {
												"image": "iHub",
												"tag": "tHub",
												"auth": {
													"username": "something",
													"token": ` + secretValueMarshalled + `
												}
											},
											"quay": {
												"image": "iQuay",
												"tag": "tQuay",
												"auth": {
													"token": ` + secretValueMarshalled + `
												}
											}
										}
									}
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
					},
					"notify": {
						"foo": {
							"type": "gotify",
							"options": {
								"message": "hello world"
							},
							"url_fields": {
								"token": ` + secretValueMarshalled + `
							},
							"params": {
								"title": "UPDATE"
							}
						}
					}
				}`,
		},
		{
			name: "settings + defaults (with notify+command+webhook service defaults) + notify + webhook",
			webhook: &webhook.WebHooksDefaults{
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
			},
			wantBody: `
				{
					"settings": {
						"web": {
							"listen_host": "127.0.0.1"
						}
					},
					"defaults": {
						"service": {
							"options": {
								"interval": "1h"
							},
							"latest_version": {
								"access_token": ` + secretValueMarshalled + `,
								"allow_invalid_certs": true,
								"use_prerelease": false,
								"require": {
									"docker": {
										"type": "hub",
										"image": "i",
										"tag": "t",
										"registry": {
											"ghcr": {
												"image": "iGHCR",
												"tag": "tGHCR",
												"auth": {
													"token": ` + secretValueMarshalled + `
												}
											},
											"hub": {
												"image": "iHub",
												"tag": "tHub",
												"auth": {
													"username": "something",
													"token": ` + secretValueMarshalled + `
												}
											},
											"quay": {
												"image": "iQuay",
												"tag": "tQuay",
												"auth": {
													"token": ` + secretValueMarshalled + `
												}
											}
										}
									}
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
					},
					"notify": {
						"foo": {
							"type": "gotify",
							"options": {
								"message": "hello world"
							},
							"url_fields": {
								"token": ` + secretValueMarshalled + `
							},
							"params": {
								"title": "UPDATE"
							}
						}
					},
					"webhook": {
						"foo": {
							"type": "github",
							"url": "https://example.com",
							"allow_invalid_certs": true,
							"secret": ` + secretValueMarshalled + `,
							"headers": [
								{
									"key": "X-Header",
									"value": ` + secretValueMarshalled + `
								}
							],
							"delay": "4s"
						}
					}
				}`,
		},
		{
			name: "settings + defaults (with notify+command+webhook service defaults) + notify + webhook + service",
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
			wantBody: `
				{
					"settings": {
						"web": {
							"listen_host": "127.0.0.1"
						}
					},
					"defaults": {
						"service": {
							"options": {
								"interval": "1h"
							},
							"latest_version": {
								"access_token": ` + secretValueMarshalled + `,
								"allow_invalid_certs": true,
								"use_prerelease": false,
								"require": {
									"docker": {
										"type": "hub",
										"image": "i",
										"tag": "t",
										"registry": {
											"ghcr": {
												"image": "iGHCR",
												"tag": "tGHCR",
												"auth": {
													"token": ` + secretValueMarshalled + `
												}
											},
											"hub": {
												"image": "iHub",
												"tag": "tHub",
												"auth": {
													"username": "something",
													"token": ` + secretValueMarshalled + `
												}
											},
											"quay": {
												"image": "iQuay",
												"tag": "tQuay",
												"auth": {
													"token": ` + secretValueMarshalled + `
												}
											}
										}
									}
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
					},
					"notify": {
						"foo": {
							"type": "gotify",
							"options": {
								"message": "hello world"
							},
							"url_fields": {
								"token": ` + secretValueMarshalled + `
							},
							"params": {
								"title": "UPDATE"
							}
						}
					},
					"webhook": {
						"foo": {
							"type": "github",
							"url": "https://example.com",
							"allow_invalid_certs": true,
							"secret": ` + secretValueMarshalled + `,
							"headers": [
								{
									"key": "X-Header",
									"value": ` + secretValueMarshalled + `
								}
							],
							"delay": "4s"
						}
					},
					"service": {
						"alpha": {
							"latest_version": {
								"type": "github",
								"url": "` + test.ArgusGitHubRepo + `",
								"access_token": ` + secretValueMarshalled + `
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

	api.Config.Settings = config.Settings{}
	api.Config.Defaults = config.Defaults{}
	api.Config.Notify = shoutrrr.ShoutrrrsDefaults{}
	api.Config.WebHook = webhook.WebHooksDefaults{}
	api.Config.Service = service.Services{}
	api.Config.Order = []string{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
				api.Config.WebHook = *tc.webhook
			}
			if tc.service != nil {
				api.Config.Service = *tc.service
			}
			if tc.order != nil {
				api.Config.Order = *tc.order
			}
			tc.wantBody = test.TrimJSON(tc.wantBody) + "\n"

			// WHEN: that HTTP request is sent.
			req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
			w := httptest.NewRecorder()
			api.httpConfig(w, req)
			res := w.Result()
			t.Cleanup(func() { _ = res.Body.Close() })

			prefix := fmt.Sprintf("%s\nAPI.httpConfig()", packageName)

			// THEN: the expected body is returned.
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("%s unexpected error:\n%v", prefix, err)
			}
			if got := string(data); got != tc.wantBody {
				t.Errorf(
					"%s non-matching response\ngot:  %q\nwant: %q",
					prefix, got, tc.wantBody,
				)
			}
		})
	}
}
