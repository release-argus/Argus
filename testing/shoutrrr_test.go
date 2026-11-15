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

// Package testing provides utilities for CLI-based testing.
package testing

import (
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	svctest "github.com/release-argus/Argus/service/test"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestFindShoutrrr(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig(t)

	tests := []struct {
		name        string
		flag        string
		cfg         *config.Config
		stdoutRegex string
		ok          bool
		foundInRoot bool
	}{
		{
			name: "empty search with only Service notifiers",
			flag: "",
			ok:   false,
			stdoutRegex: test.TrimYAML(`
				^FATAL: Notifier .* could not be found.*
				[^\s]+.*one of these?.*
					- bar
					- baz
					- foo\s$`,
			),
			cfg: &config.Config{
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bar:
										type: gotify
										options:
											max_tries: 1
										url_fields:
											host: `+test.NotifyGotify.HostValid+`
											path: `+test.NotifyGotify.Path+`
											token: `+test.NotifyGotify.TokenPass+`
									baz: {}
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
			},
		},
		{
			name: "empty search with only Root notifiers",
			flag: "",
			ok:   false,
			stdoutRegex: test.TrimYAML(`
				^FATAL: Notifier .* could not be found.*
				[^\s]+.*one of these?.*
					- bar
					- baz
					- foo\s$`,
			),
			cfg: &config.Config{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"foo": {},
					"bar": {},
					"baz": {},
				},
			},
		},
		{
			name: "empty search with Root notifiers and Service notifiers and no duplicates",
			flag: "",
			ok:   false,
			stdoutRegex: test.TrimYAML(`
				^FATAL: Notifier .* could not be found.*
				[^\s]+.*one of these?.*
					- bar
					- baz
					- foo\s$`,
			),
			cfg: &config.Config{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"foo": {},
					"bar": {},
					"baz": {},
				},
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bar: {}
									baz: {}
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
			},
		},
		{
			name: "empty search with Root notifiers and Service notifiers and duplicates",
			flag: "",
			ok:   false,
			stdoutRegex: test.TrimYAML(`
				^FATAL: Notifier .* could not be found.*
				[^\s]+.*one of these?.*
					- bar
					- baz
					- foo
					- shazam\s$`,
			),
			cfg: &config.Config{
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bar: {}
									baz: {}
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
				Notify: shoutrrr.ShoutrrrsDefaults{
					"foo":    {},
					"shazam": {},
					"baz":    {},
				},
			},
		},
		{
			name:        "matching search of notifier in Root",
			flag:        "bosh",
			ok:          true,
			stdoutRegex: `^$`,
			cfg: &config.Config{
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bar: {}
									baz: {}
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
				Notify: shoutrrr.ShoutrrrsDefaults{
					"bosh": shoutrrr.NewDefaults(
						"gotify",
						nil,
						map[string]string{
							"host":  "example.com",
							"token": "example",
						},
						nil,
					),
				},
			},
			foundInRoot: true,
		},
		{
			name:        "matching search of notifier in Service",
			flag:        "baz",
			ok:          true,
			stdoutRegex: `^$`,
			cfg: &config.Config{
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bar: {}
									baz:
										type: gotify
										url_fields:
											host: example.com
											token: example
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
			},
		},
		{
			name:        "matching search of notifier in Root and a Service",
			flag:        "bar",
			ok:          true,
			stdoutRegex: `^$`,
			cfg: &config.Config{
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bar:
										type: gotify
										url_fields:
											host: example.com
											token: example
									baz: {}
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
				Notify: shoutrrr.ShoutrrrsDefaults{
					"bar": shoutrrr.NewDefaults(
						"gotify",
						nil,
						map[string]string{
							"host":  "example.com",
							"token": "example",
						},
						nil,
					),
				},
			},
		},
		{
			name:        "matching search of Service notifier with incomplete config filled by Defaults",
			flag:        "bar",
			ok:          true,
			stdoutRegex: `^$`,
			cfg: &config.Config{
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bar:
										type: smtp
										url_fields:
											host: example.com
										params:
											fromaddress: test@release-argus.io
									baz: {}
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
				Defaults: config.Defaults{
					Notify: shoutrrr.ShoutrrrsDefaults{
						"something": shoutrrr.NewDefaults(
							"something",
							map[string]string{
								"title": "bar",
							},
							nil,
							nil,
						),
						"smtp": shoutrrr.NewDefaults(
							"",
							nil, nil,
							map[string]string{
								"toaddresses": "me@you.com",
							}),
					},
				},
			},
		},
		{
			name:        "matching search of Service notifier with incomplete config filled by Root",
			flag:        "bar",
			ok:          true,
			stdoutRegex: `^$`,
			cfg: &config.Config{
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bar:
										type: smtp
										url_fields:
											host: example.com
										params:
											fromaddress: test@release-argus.io
									baz: {}
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
				Notify: shoutrrr.ShoutrrrsDefaults{
					"something": shoutrrr.NewDefaults(
						"something",
						map[string]string{
							"title": "bar",
						},
						nil,
						nil,
					),
					"smtp": shoutrrr.NewDefaults(
						"",
						nil, nil,
						map[string]string{
							"toaddresses": "me@you.com",
						}),
				},
			},
		},
		{
			name:        "matching search of Service notifier with incomplete config filled by Root and Defaults",
			flag:        "bosh",
			ok:          true,
			stdoutRegex: `^$`,
			cfg: &config.Config{
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bosh:
										params:
											fromaddress: test@release-argus.io
									baz: {}
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
				Notify: shoutrrr.ShoutrrrsDefaults{
					"bosh": shoutrrr.NewDefaults(
						"smtp",
						nil,
						map[string]string{
							"host": "example.com",
						},
						nil,
					),
				},
				Defaults: config.Defaults{
					Notify: shoutrrr.ShoutrrrsDefaults{
						"something": shoutrrr.NewDefaults(
							"something",
							map[string]string{
								"title": "bar",
							},
							nil, nil,
						),
						"smtp": shoutrrr.NewDefaults(
							"",
							nil, nil,
							map[string]string{
								"toaddresses": "me@you.com",
							}),
					},
				},
			},
		},
		{
			name: "matching search of Root notifier with invalid config",
			flag: "bosh",
			ok:   false,
			stdoutRegex: test.TrimYAML(`
				^FATAL: notify:
					bosh:
						params:
							toaddresses: <required>.*`,
			),
			cfg: &config.Config{
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bar: {}
									baz: {}
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
				Notify: shoutrrr.ShoutrrrsDefaults{
					"bosh": shoutrrr.NewDefaults(
						"smtp",
						nil,
						map[string]string{
							"host": "example.com",
						},
						map[string]string{
							"fromaddress": "test@release-argus.io",
						}),
				},
			},
		},
		{
			name:        "matching search of Root notifier with incomplete config filled by Defaults",
			flag:        "bosh",
			ok:          true,
			stdoutRegex: `^$`,
			cfg: &config.Config{
				Service: test.Must(t, func() (service.Services, error) {
					return service.DecodeServices(
						"yaml", []byte(test.TrimYAML(`
							argus:
								notify:
									foo: {}
									bar: {}
									baz: {}
						`)),
						svcCfg, notifyCfg, whCfg,
					)
				}),
				Notify: shoutrrr.ShoutrrrsDefaults{
					"bosh": shoutrrr.NewDefaults(
						"smtp",
						nil,
						map[string]string{
							"host": "example.com",
						},
						map[string]string{
							"fromaddress": "test@release-argus.io",
						}),
				},
				Defaults: config.Defaults{
					Notify: shoutrrr.ShoutrrrsDefaults{
						"something": shoutrrr.NewDefaults(
							"something",
							nil, nil,
							map[string]string{
								"title": "bar",
							}),
						"smtp": shoutrrr.NewDefaults(
							"",
							nil, nil,
							map[string]string{
								"toaddresses": "me@you.com",
							},
						),
					},
				},
			},
			foundInRoot: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			tc.cfg.HardDefaults.Notify.Default()

			var got *shoutrrr.Shoutrrr
			test.AssertCallLog(
				t,
				tc.ok,
				func() bool {
					var ok bool
					got, ok = findShoutrrr(tc.flag, tc.cfg, logx.LogFrom{})
					return ok
				},
				logx.ExitCodeChannel(),
				releaseStdout,
				tc.stdoutRegex,
			)
			if !tc.ok {
				return
			}
			if !tc.foundInRoot {
				if want := tc.cfg.Service["argus"].Notify[tc.flag]; got != want {
					t.Fatalf("service: Shoutrrr mismatch\ngot:  %p\nwant: %p", got, want)
				}
				// A Service Notifier gets its Defaults in Init, which this doesn't call.
				return
			}
			wantStr := tc.cfg.Notify[tc.flag].String("")
			if gotStr := got.String(""); gotStr != wantStr {
				t.Fatalf("root: Shoutrrr.String() mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}
			if want := tc.cfg.Defaults.Notify[got.Type]; want != nil && got.Defaults != want {
				t.Fatalf("root: Shoutrrr.Defaults mismatch\ngot:  %p\nwant: %p", got.Defaults, want)
			}
		})
	}
}

func TestGetAllShoutrrrNames(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig(t)

	tests := []struct {
		name          string
		svc           service.Services
		rootNotifiers shoutrrr.ShoutrrrsDefaults
		want          []string
	}{
		{
			name: "nothing",
		},
		{
			name: "only service notifiers",
			svc: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						0:
							notify:
								foo: {}
						1:
							notify:
								bar: {}
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: []string{"bar", "foo"},
		},
		{
			name: "only service notifiers with duplicates",
			svc: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						0:
							notify:
								foo: {}
						1:
							notify:
								foo: {}
								bar: {}
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
			want: []string{"bar", "foo"},
		},
		{
			name: "only root notifiers",
			rootNotifiers: shoutrrr.ShoutrrrsDefaults{
				"foo": {},
				"bar": {},
			},
			want: []string{"bar", "foo"},
		},
		{
			name: "root + service notifiers",
			svc: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						0:
							notify:
								foo: {}
						1:
							notify:
								bar: {}
								foo: {}
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
			rootNotifiers: shoutrrr.ShoutrrrsDefaults{
				"baz": {},
			},
			want: []string{"bar", "baz", "foo"},
		},
		{
			name: "root + service notifiers with duplicates",
			svc: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						0:
							notify:
								foo: {}
						1:
							notify:
								bar: {}
								foo: {}
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
			rootNotifiers: shoutrrr.ShoutrrrsDefaults{
				"foo": {},
				"bar": {},
				"baz": {},
			},
			want: []string{"bar", "baz", "foo"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Config{
				Service: tc.svc,
				Notify:  tc.rootNotifiers,
			}

			got := getAllShoutrrrNames(&cfg)

			if testErr := test.AssertSlicesEqual(t, got, tc.want, ""); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
