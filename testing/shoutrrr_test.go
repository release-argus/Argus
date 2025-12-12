// Copyright [2025] [Argus]
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

package testing

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/webhook"
)

func TestGetAllShoutrrrNames(t *testing.T) {
	// GIVEN various Services and Notifiers.
	tests := map[string]struct {
		service       service.Services
		rootNotifiers shoutrrr.ShoutrrrsDefaults
		want          []string
	}{
		"nothing": {},
		"only service notifiers": {
			service: service.Services{
				"0": {Notify: shoutrrr.Shoutrrrs{"foo": {}}},
				"1": {Notify: shoutrrr.Shoutrrrs{"bar": {}}},
			},
			want: []string{"bar", "foo"},
		},
		"only service notifiers with duplicates": {
			service: service.Services{
				"0": {Notify: shoutrrr.Shoutrrrs{"foo": {}}},
				"1": {Notify: shoutrrr.Shoutrrrs{"foo": {}, "bar": {}}},
			},
			want: []string{"bar", "foo"},
		},
		"only root notifiers": {rootNotifiers: shoutrrr.ShoutrrrsDefaults{
			"foo": {}, "bar": {}},
			want: []string{"bar", "foo"},
		},
		"root + service notifiers": {
			service: service.Services{
				"0": {Notify: shoutrrr.Shoutrrrs{"foo": {}}},
				"1": {Notify: shoutrrr.Shoutrrrs{"foo": {}, "bar": {}}},
			},
			rootNotifiers: shoutrrr.ShoutrrrsDefaults{
				"baz": {}},
			want: []string{"bar", "baz", "foo"},
		},
		"root + service notifiers with duplicates": {
			service: service.Services{
				"0": {Notify: shoutrrr.Shoutrrrs{"foo": {}}},
				"1": {Notify: shoutrrr.Shoutrrrs{"foo": {}, "bar": {}}},
			},
			rootNotifiers: shoutrrr.ShoutrrrsDefaults{
				"foo": {}, "bar": {}, "baz": {}},
			want: []string{"bar", "baz", "foo"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Config{
				Service: tc.service,
				Notify:  tc.rootNotifiers,
			}

			// WHEN getAllShoutrrrNames is called on this config.
			got := getAllShoutrrrNames(&cfg)

			// THEN a list of all Shoutrrrs will be returned.
			if len(got) != len(tc.want) {
				t.Fatalf("%s\nlists length mismatch\nwant: %s\ngot:  %s",
					packageName, tc.want, got)
			}
			gotIndex := 0
			for gotIndex != 0 {
				found := false
				for wantIndex := range tc.want {
					if got[gotIndex] == tc.want[wantIndex] {
						found = true
						util.RemoveIndex(&got, gotIndex)
						util.RemoveIndex(&tc.want, wantIndex)
						break
					}
				}
				if !found {
					t.Fatalf("%s\nwant: %v\ngot:  %v",
						packageName, tc.want, got)
				}
				gotIndex--
			}
		})
	}
}

func TestFindShoutrrr(t *testing.T) {
	// GIVEN a Config with/without Service containing a Shoutrrr and Root Shoutrrrs.
	tests := map[string]struct {
		flag                    string
		cfg                     *config.Config
		stdoutRegex, panicRegex *string
		foundInRoot             *bool
	}{
		"empty search with only Service notifiers": {
			flag:       "",
			panicRegex: test.StringPtr(`could not be found.*\s+.*one of these?.*\s  - bar\s  - baz\s  - foo\s$`),
			cfg: &config.Config{
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{
							"foo": {},
							"bar": {},
							"baz": {}}}}},
		},
		"empty search with only Root notifiers": {
			flag:       "",
			panicRegex: test.StringPtr(`could not be found.*\s+.*one of these?.*\s  - bar\s  - baz\s  - foo\s$`),
			cfg: &config.Config{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"foo": {},
					"bar": {},
					"baz": {}}},
		},
		"empty search with Root notifiers and Service notifiers and no duplicates": {
			flag:       "",
			panicRegex: test.StringPtr(`could not be found.*\s+.*one of these?.*\s  - bar\s  - baz\s  - foo\s$`),
			cfg: &config.Config{
				Notify: shoutrrr.ShoutrrrsDefaults{
					"foo": {},
					"bar": {},
					"baz": {}},
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{
							"foo": {},
							"bar": {},
							"baz": {}}}}},
		},
		"empty search with Root notifiers and Service notifiers and duplicates": {
			flag:       "",
			panicRegex: test.StringPtr(`could not be found.*\s+.*one of these?.*\s  - bar\s  - baz\s  - foo\s  - shazam\s$`),
			cfg: &config.Config{
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{"foo": {}, "bar": {}, "baz": {}}}},
				Notify: shoutrrr.ShoutrrrsDefaults{
					"foo":    {},
					"shazam": {},
					"baz":    {}}},
		},
		"matching search of notifier in Root": {
			flag:        "bosh",
			stdoutRegex: test.StringPtr("^$"),
			cfg: &config.Config{
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{
							"foo": {}, "bar": {}, "baz": {}}}},
				Notify: shoutrrr.ShoutrrrsDefaults{
					"bosh": shoutrrr.NewDefaults(
						"gotify",
						nil,
						map[string]string{
							"host":  "example.com",
							"token": "example"},
						nil)}},
			foundInRoot: test.BoolPtr(true),
		},
		"matching search of notifier in Service": {
			flag:        "baz",
			stdoutRegex: test.StringPtr("^$"),
			cfg: &config.Config{
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{
							"foo": {}, "bar": {},
							"baz": shoutrrr.New(
								nil,
								"", "gotify",
								nil,
								map[string]string{
									"host":  "example.com",
									"token": "example"},
								nil,
								nil, nil, nil)}}}},
			foundInRoot: test.BoolPtr(false),
		},
		"matching search of notifier in Root and a Service": {
			flag:        "bar",
			stdoutRegex: test.StringPtr("^$"),
			cfg: &config.Config{
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{
							"foo": {},
							"bar": shoutrrr.New(
								nil,
								"", "gotify",
								nil,
								map[string]string{
									"host":  "example.com",
									"token": "foo"},
								nil,
								nil, nil, nil),
							"baz": {}}}},
				Notify: shoutrrr.ShoutrrrsDefaults{
					"bar": shoutrrr.NewDefaults(
						"gotify",
						nil,
						map[string]string{
							"host":  "example.com",
							"token": "example"},
						nil)}},
			foundInRoot: test.BoolPtr(false),
		},
		"matching search of Service notifier with incomplete config filled by Defaults": {
			flag: "bar",
			cfg: &config.Config{
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{
							"foo": {},
							"bar": shoutrrr.New(
								nil,
								"", "smtp",
								nil,
								map[string]string{
									"fromaddress": "test@release-argus.io"},
								map[string]string{
									"host": "example.com"},
								nil, nil, nil),
							"baz": {}}}},
				Defaults: config.Defaults{
					Notify: shoutrrr.ShoutrrrsDefaults{
						"something": shoutrrr.NewDefaults(
							"something",
							map[string]string{
								"title": "bar"},
							nil,
							nil),
						"smtp": shoutrrr.NewDefaults(
							"",
							nil, nil,
							map[string]string{
								"toaddresses": "me@you.com"})}}},
			foundInRoot: test.BoolPtr(false),
		},
		"matching search of Service notifier with incomplete config filled by Root": {
			flag: "bar",
			cfg: &config.Config{
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{
							"foo": {},
							"bar": shoutrrr.New(
								nil,
								"", "smtp",
								nil,
								map[string]string{
									"host": "example.com"},
								map[string]string{
									"fromaddress": "test@release-argus.io"},
								nil, nil, nil),
							"baz": {}}}},
				Notify: shoutrrr.ShoutrrrsDefaults{
					"something": shoutrrr.NewDefaults(
						"something",
						map[string]string{
							"title": "bar"},
						nil,
						nil),
					"smtp": shoutrrr.NewDefaults(
						"",
						nil, nil,
						map[string]string{
							"toaddresses": "me@you.com"})}},
			foundInRoot: test.BoolPtr(false),
		},
		"matching search of Service notifier with incomplete config filled by Root and Defaults": {
			flag: "bosh",
			cfg: &config.Config{
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{
							"foo": {},
							"bosh": shoutrrr.New(
								nil,
								"", "smtp",
								nil,
								map[string]string{
									"host": "example.com"},
								map[string]string{
									"fromaddress": "test@release-argus.io"},
								nil, nil, nil),
							"baz": {}}}},
				Notify: shoutrrr.ShoutrrrsDefaults{
					"bosh": shoutrrr.NewDefaults(
						"smtp",
						nil,
						map[string]string{
							"host": "example.com"},
						nil)},
				Defaults: config.Defaults{
					Notify: shoutrrr.ShoutrrrsDefaults{
						"something": shoutrrr.NewDefaults(
							"something",
							map[string]string{
								"title": "bar"},
							nil, nil),
						"smtp": shoutrrr.NewDefaults(
							"",
							nil, nil,
							map[string]string{
								"toaddresses": "me@you.com"})}}},
			foundInRoot: test.BoolPtr(false),
		},
		"matching search of Root notifier with invalid config": {
			flag:       "bosh",
			panicRegex: test.StringPtr(`^notify:\s  bosh:\s    params:\s      toaddresses: <required>`),
			cfg: &config.Config{
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{
							"foo": {},
							"bar": {},
							"baz": {}}}},
				Notify: shoutrrr.ShoutrrrsDefaults{
					"bosh": shoutrrr.NewDefaults(
						"smtp",
						nil,
						map[string]string{
							"host": "example.com"},
						map[string]string{
							"fromaddress": "test@release-argus.io"})}},
			foundInRoot: test.BoolPtr(true),
		},
		"matching search of Root notifier with incomplete config filled by Defaults": {
			flag: "bosh",
			cfg: &config.Config{
				Service: service.Services{
					"argus": {
						Notify: shoutrrr.Shoutrrrs{
							"foo": {},
							"bar": {},
							"baz": {}}}},
				Notify: shoutrrr.ShoutrrrsDefaults{
					"bosh": shoutrrr.NewDefaults(
						"smtp",
						nil,
						map[string]string{
							"host": "example.com"},
						map[string]string{
							"fromaddress": "test@release-argus.io"})},
				Defaults: config.Defaults{
					Notify: shoutrrr.ShoutrrrsDefaults{
						"something": shoutrrr.NewDefaults(
							"something",
							nil, nil,
							map[string]string{
								"title": "bar"}),
						"smtp": shoutrrr.NewDefaults(
							"",
							nil, nil,
							map[string]string{
								"toaddresses": "me@you.com"})}}},
			foundInRoot: test.BoolPtr(true),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			if tc.panicRegex != nil {
				// Switch Fatal to panic and disable this panic.
				defer func() {
					r := recover()
					releaseStdout()

					rStr := fmt.Sprint(r)
					if !util.RegexCheck(*tc.panicRegex, rStr) {
						t.Errorf("%s\nexpected a panic that matched %q\ngot: %q",
							packageName, *tc.panicRegex, rStr)
					}
				}()
			}
			tc.cfg.HardDefaults.Notify.Default()
			for _, svc := range tc.cfg.Service {
				svc.Init(
					&tc.cfg.Defaults.Service, &tc.cfg.HardDefaults.Service,
					&tc.cfg.Notify, &tc.cfg.Defaults.Notify, &tc.cfg.HardDefaults.Notify,
					&tc.cfg.WebHook, &tc.cfg.Defaults.WebHook, &tc.cfg.HardDefaults.WebHook)
			}

			// WHEN findShoutrrr is called with the test Config.
			got := findShoutrrr(tc.flag, tc.cfg, logutil.LogFrom{})

			// THEN we get the expected stdout.
			stdout := releaseStdout()
			if tc.stdoutRegex != nil {
				if !util.RegexCheck(*tc.stdoutRegex, stdout) {
					t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, *tc.stdoutRegex, stdout)
				}
			}
			// If the notifier should have been found in the root or in a service.
			if tc.foundInRoot != nil {
				if *tc.foundInRoot {
					if got.String("") != tc.cfg.Notify[tc.flag].String("") {
						t.Fatalf("%s\nwasn't found in .Notify\nwant: %q\n\ngot:  %q",
							packageName, tc.cfg.Notify[tc.flag].String(""), got.String(""))
					}
				} else {
					if got.String("") != tc.cfg.Service["argus"].Notify[tc.flag].String("") {
						t.Fatalf("%s\nwasn't found in Service\nwant: %q\ngot:  %q",
							packageName, tc.cfg.Service["argus"].Notify[tc.flag].String(""), got.String(""))
					}
					// Would have been given in the Init.
					got.Defaults = tc.cfg.Defaults.Notify[got.Type]
				}
			}
			// If there were Defaults for that type.
			if tc.cfg.Defaults.Notify[got.Type] != nil {
				if got.Defaults.String("") != tc.cfg.Defaults.Notify[got.Type].String("") {
					t.Fatalf("%s\ndefaults were not applied\nwant: %v\ngot:  %v",
						packageName, tc.cfg.Defaults.Notify[got.Type].String(""), got.Defaults.String(""))
				}
			}
		})
	}
}

func TestNotifyTest(t *testing.T) {
	// GIVEN a Config with/without Service containing a Shoutrrr and Root Shoutrrrs.
	emptyShoutrrr := shoutrrr.NewDefaults(
		"",
		map[string]string{},
		map[string]string{},
		map[string]string{})
	tests := map[string]struct {
		flag                    string
		services                service.Services
		mainShoutrrrs           shoutrrr.ShoutrrrsDefaults
		stdoutRegex, panicRegex *string
	}{
		"empty flag": {flag: "",
			stdoutRegex: test.StringPtr("^$"),
			services: service.Services{
				"argus": {
					Notify: shoutrrr.Shoutrrrs{
						"foo": {},
						"bar": {},
						"baz": {},
					},
				}}},
		"unknown Notifier": {flag: "something",
			panicRegex: test.StringPtr("Notifier.* could not be found"),
			services: service.Services{
				"argus": {
					Notify: shoutrrr.Shoutrrrs{
						"foo": {},
						"bar": {},
						"baz": {},
					},
				}}},
		"known Service Notifier with invalid Gotify token": {
			flag:       "bar",
			panicRegex: test.StringPtr(`Message failed to send with "bar" config\s+.*invalid gotify token`),
			services: service.Services{
				"argus": {
					Notify: shoutrrr.Shoutrrrs{
						"foo": {},
						"bar": shoutrrr.New(
							nil,
							"bar", "gotify",
							map[string]string{
								"max_tries": "1"},
							map[string]string{
								"host":  test.ValidCertNoProtocol,
								"token": "invalid"},
							map[string]string{},
							emptyShoutrrr,
							emptyShoutrrr,
							emptyShoutrrr),
						"baz": {}}}},
		},
		"invalid Gotify token format": {
			flag:       "bar",
			panicRegex: test.StringPtr(`invalid gotify token: "abc"`),
			services: service.Services{
				"argus": {
					Notify: shoutrrr.Shoutrrrs{
						"foo": {},
						"bar": shoutrrr.New(
							nil,
							"bar", "gotify",
							map[string]string{
								"max_tries": "1"},
							map[string]string{
								"host":  test.ValidCertNoProtocol,
								"token": "abc"},
							map[string]string{},
							emptyShoutrrr,
							emptyShoutrrr,
							emptyShoutrrr),
						"baz": {}}}},
		},
		"valid Gotify token format": {
			flag:       "bar",
			panicRegex: test.StringPtr(`Message failed to send with.*\s.*server responded`),
			services: service.Services{
				"argus": {
					Notify: shoutrrr.Shoutrrrs{
						"foo": {},
						"bar": shoutrrr.New(
							nil,
							"bar", "gotify",
							map[string]string{
								"max_tries": "1"},
							map[string]string{
								"host":  test.ValidCertNoProtocol,
								"token": "AGdjFCZugzJGhEG"},
							map[string]string{},
							emptyShoutrrr,
							emptyShoutrrr,
							emptyShoutrrr),
						"baz": {}}}},
		},
		"shoutrrr from Root": {
			flag:       "baz",
			panicRegex: test.StringPtr(`Message failed to send with.*\s.*server responded`),
			services:   service.Services{},
			mainShoutrrrs: shoutrrr.ShoutrrrsDefaults{
				"baz": shoutrrr.NewDefaults(
					"gotify",
					map[string]string{
						"max_tries": "1"},
					map[string]string{
						"host":  test.ValidCertNoProtocol,
						"token": "AGdjFCZugzJGhEG"},
					map[string]string{})},
		},
		"successful send": {
			flag:        "work",
			stdoutRegex: test.StringPtr(`Message sent successfully with "work" config`),
			services: service.Services{
				"argus": {
					Notify: shoutrrr.Shoutrrrs{
						"work": shoutrrr.New(
							nil,
							"bar", "gotify",
							map[string]string{
								"max_tries": "1"},
							map[string]string{
								"host":  test.ValidCertNoProtocol,
								"path":  "/gotify",
								"token": test.ShoutrrrGotifyToken()},
							map[string]string{},
							emptyShoutrrr,
							emptyShoutrrr,
							emptyShoutrrr),
					}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureStdout()

			serviceHardDefaults := service.Defaults{}
			serviceHardDefaults.Default()
			shoutrrrHardDefaults := shoutrrr.ShoutrrrsDefaults{}
			shoutrrrHardDefaults.Default()
			for i := range tc.services {
				(*tc.services[i]).Init(
					&service.Defaults{}, &serviceHardDefaults,
					&tc.mainShoutrrrs, &shoutrrr.ShoutrrrsDefaults{}, &shoutrrrHardDefaults,
					&webhook.WebHooksDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})
			}
			if tc.panicRegex != nil {
				// Switch Fatal to panic and disable this panic.
				defer func() {
					r := recover()
					releaseStdout()

					rStr := fmt.Sprint(r)
					if !util.RegexCheck(*tc.panicRegex, rStr) {
						t.Errorf("%s\nexpected a panic that matched %q\ngot: %q",
							packageName, *tc.panicRegex, rStr)
					}
				}()
			}

			// WHEN NotifyTest is called with the test Config.
			cfg := config.Config{
				Service: tc.services,
				Notify:  tc.mainShoutrrrs}
			NotifyTest(&tc.flag, &cfg)

			// THEN we get the expected stdout.
			stdout := releaseStdout()
			if tc.stdoutRegex != nil {
				if !util.RegexCheck(*tc.stdoutRegex, stdout) {
					t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, *tc.stdoutRegex, stdout)
				}
			}
		})
	}
}
