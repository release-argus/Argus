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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	svctest "github.com/release-argus/Argus/service/test"
	"github.com/release-argus/Argus/util"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestGetAllShoutrrrNames(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	// GIVEN: various Services and Notifiers.
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

			// WHEN: getAllShoutrrrNames is called on this config.
			got := getAllShoutrrrNames(&cfg)

			// THEN: a list of all Shoutrrrs will be returned.
			if err := test.AssertSlicesEqualFunc(
				t,
				got,
				tc.want,
				func(a, b string) bool { return a == b },
				fmt.Sprintf("%s\ngetAllShoutrrrNames()", packageName),
				"",
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFindShoutrrr(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	// GIVEN: a Config with/without Service containing a Shoutrrr and Root Shoutrrrs.
	tests := []struct {
		name        string
		flag        string
		cfg         *config.Config
		stdoutRegex string
		ok          bool
		foundInRoot *bool
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
											host: `+test.ValidCertNoProtocol+`
											path: /gotify
											token: `+test.ShoutrrrGotifyToken()+`
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
			foundInRoot: test.Ptr(true),
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
			foundInRoot: test.Ptr(false),
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
			foundInRoot: test.Ptr(false),
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
			foundInRoot: test.Ptr(false),
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
			foundInRoot: test.Ptr(false),
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
			foundInRoot: test.Ptr(false),
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
			foundInRoot: test.Ptr(true),
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
			foundInRoot: test.Ptr(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			tc.cfg.HardDefaults.Notify.Default()

			resultChannel := make(chan bool, 1)
			// WHEN: findShoutrrr is called with the test Config.
			got, ok := findShoutrrr(tc.flag, tc.cfg, logx.LogFrom{})
			resultChannel <- ok

			prefix := fmt.Sprintf(
				"%s\nfindShoutrrr(%q)",
				packageName, tc.flag,
			)

			// THEN: we get the expected stdout.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant:  %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}

			// AND: it succeeds/fails as expected.
			if err := test.AssertChannelBool(
				t,
				tc.ok,
				resultChannel,
				logx.ExitCodeChannel(),
				nil,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}
			if !tc.ok {
				return
			}
			// If the notifier should have been found in the root or in a service.
			if tc.foundInRoot != nil {
				if *tc.foundInRoot {
					w := tc.cfg.Notify[tc.flag].String("")
					if got := got.String(""); got != w {
						t.Fatalf(
							"%s mismatch on Shoutrrr that should have been found in Root\ngot:  %q\nwant: %q",
							prefix, got, w,
						)
					}
				} else {
					w := tc.cfg.Service["argus"].Notify[tc.flag].String("")
					if got := got.String(""); got != w {
						t.Fatalf(
							"%s mismatch on Shoutrrr that should have been found on a Service\ngot:  %q\nwant: %q",
							prefix, got, w,
						)
					}
					// Would have been given in the Init.
					got.Defaults = tc.cfg.Defaults.Notify[got.Type]
				}
			}
			// If there were Defaults for that type.
			if tc.cfg.Defaults.Notify[got.Type] != nil {
				want := tc.cfg.Defaults.Notify[got.Type].String("")
				got := got.Defaults.String("")
				if got != want {
					t.Fatalf(
						"%s defaults were not applied correctly to the Shoutrrr returned\ngot:  %v\nwant: %v",
						prefix, got, want,
					)
				}
			}
		})
	}
}
