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
		service       service.Slice
		rootNotifiers shoutrrr.SliceDefaults
		want          []string
	}{
		"nothing": {},
		"only service notifiers": {
			service: service.Slice{
				"0": {Notify: shoutrrr.Slice{"foo": {}}},
				"1": {Notify: shoutrrr.Slice{"bar": {}}},
			},
			want: []string{"bar", "foo"},
		},
		"only service notifiers with duplicates": {
			service: service.Slice{
				"0": {Notify: shoutrrr.Slice{"foo": {}}},
				"1": {Notify: shoutrrr.Slice{"foo": {}, "bar": {}}},
			},
			want: []string{"bar", "foo"},
		},
		"only root notifiers": {rootNotifiers: shoutrrr.SliceDefaults{
			"foo": {}, "bar": {}},
			want: []string{"bar", "foo"},
		},
		"root + service notifiers": {
			service: service.Slice{
				"0": {Notify: shoutrrr.Slice{"foo": {}}},
				"1": {Notify: shoutrrr.Slice{"foo": {}, "bar": {}}},
			},
			rootNotifiers: shoutrrr.SliceDefaults{
				"baz": {}},
			want: []string{"bar", "baz", "foo"},
		},
		"root + service notifiers with duplicates": {
			service: service.Slice{
				"0": {Notify: shoutrrr.Slice{"foo": {}}},
				"1": {Notify: shoutrrr.Slice{"foo": {}, "bar": {}}},
			},
			rootNotifiers: shoutrrr.SliceDefaults{
				"foo": {}, "bar": {}, "baz": {}},
			want: []string{"bar", "baz", "foo"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			cfg := config.Config{
				Service: tc.service,
				Notify:  tc.rootNotifiers,
			}

			// WHEN getAllShoutrrrNames is called on this config.
			got := getAllShoutrrrNames(&cfg)

			// THEN a list of all Shoutrrrs will be returned.
			if len(got) != len(tc.want) {
				t.Fatalf("lists differ in length\nwant: %s\ngot:  %s",
					tc.want, got)
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
					t.Fatalf("want: %v\ngot: %v",
						tc.want, got)
				}
				gotIndex--
			}
		})
	}
}

func TestFindShoutrrr(t *testing.T) {
	// GIVEN a Config with/without Service containing a Shoutrrr and Root Shoutrrr(s).
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
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{
							"foo": {},
							"bar": {},
							"baz": {}}}}},
		},
		"empty search with only Root notifiers": {
			flag:       "",
			panicRegex: test.StringPtr(`could not be found.*\s+.*one of these?.*\s  - bar\s  - baz\s  - foo\s$`),
			cfg: &config.Config{
				Notify: shoutrrr.SliceDefaults{
					"foo": {},
					"bar": {},
					"baz": {}}},
		},
		"empty search with Root notifiers and Service notifiers and no duplicates": {
			flag:       "",
			panicRegex: test.StringPtr(`could not be found.*\s+.*one of these?.*\s  - bar\s  - baz\s  - foo\s$`),
			cfg: &config.Config{
				Notify: shoutrrr.SliceDefaults{
					"foo": {},
					"bar": {},
					"baz": {}},
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{
							"foo": {},
							"bar": {},
							"baz": {}}}}},
		},
		"empty search with Root notifiers and Service notifiers and duplicates": {
			flag:       "",
			panicRegex: test.StringPtr(`could not be found.*\s+.*one of these?.*\s  - bar\s  - baz\s  - foo\s  - shazam\s$`),
			cfg: &config.Config{
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{"foo": {}, "bar": {}, "baz": {}}}},
				Notify: shoutrrr.SliceDefaults{
					"foo":    {},
					"shazam": {},
					"baz":    {}}},
		},
		"matching search of notifier in Root": {
			flag:        "bosh",
			stdoutRegex: test.StringPtr("^$"),
			cfg: &config.Config{
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{
							"foo": {}, "bar": {}, "baz": {}}}},
				Notify: shoutrrr.SliceDefaults{
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
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{
							"foo": {}, "bar": {},
							"baz": shoutrrr.New(
								nil, "",
								"gotify",
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
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{
							"foo": {},
							"bar": shoutrrr.New(
								nil, "",
								"gotify",
								nil,
								map[string]string{
									"host":  "example.com",
									"token": "foo"},
								nil,
								nil, nil, nil),
							"baz": {}}}},
				Notify: shoutrrr.SliceDefaults{
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
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{
							"foo": {},
							"bar": shoutrrr.New(
								nil, "",
								"smtp",
								nil,
								map[string]string{
									"fromaddress": "test@release-argus.io"},
								map[string]string{
									"host": "example.com"},
								nil, nil, nil),
							"baz": {}}}},
				Defaults: config.Defaults{
					Notify: shoutrrr.SliceDefaults{
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
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{
							"foo": {},
							"bar": shoutrrr.New(
								nil, "",
								"smtp",
								nil,
								map[string]string{
									"host": "example.com"},
								map[string]string{
									"fromaddress": "test@release-argus.io"},
								nil, nil, nil),
							"baz": {}}}},
				Notify: shoutrrr.SliceDefaults{
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
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{
							"foo": {},
							"bosh": shoutrrr.New(
								nil, "",
								"smtp",
								nil,
								map[string]string{
									"host": "example.com"},
								map[string]string{
									"fromaddress": "test@release-argus.io"},
								nil, nil, nil),
							"baz": {}}}},
				Notify: shoutrrr.SliceDefaults{
					"bosh": shoutrrr.NewDefaults(
						"smtp",
						nil,
						map[string]string{
							"host": "example.com"},
						nil)},
				Defaults: config.Defaults{
					Notify: shoutrrr.SliceDefaults{
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
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{
							"foo": {},
							"bar": {},
							"baz": {}}}},
				Notify: shoutrrr.SliceDefaults{
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
				Service: service.Slice{
					"argus": {
						Notify: shoutrrr.Slice{
							"foo": {},
							"bar": {},
							"baz": {}}}},
				Notify: shoutrrr.SliceDefaults{
					"bosh": shoutrrr.NewDefaults(
						"smtp",
						nil,
						map[string]string{
							"host": "example.com"},
						map[string]string{
							"fromaddress": "test@release-argus.io"})},
				Defaults: config.Defaults{
					Notify: shoutrrr.SliceDefaults{
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
						t.Errorf("expected a panic that matched %q\ngot: %q",
							*tc.panicRegex, rStr)
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
					t.Fatalf("want match for %q\nnot: %q",
						*tc.stdoutRegex, stdout)
				}
			}
			// If the notifier should have been found in the root or in a service.
			if tc.foundInRoot != nil {
				if *tc.foundInRoot {
					if tc.cfg.Notify[tc.flag].String("") != got.String("") {
						t.Fatalf("want:\n%v\n\ngot:\n%v",
							tc.cfg.Notify[tc.flag].String(""), got.String(""))
					}
				} else {
					if tc.cfg.Service["argus"].Notify[tc.flag].String("") != got.String("") {
						t.Fatalf("want: %v\ngot: %v",
							tc.cfg.Service["argus"].Notify[tc.flag].String(""), got.String(""))
					}
					// Would have been given in the Init.
					got.Defaults = tc.cfg.Defaults.Notify[got.Type]
				}
			}
			// If there were Defaults for that type.
			if tc.cfg.Defaults.Notify[got.Type] != nil {
				if tc.cfg.Defaults.Notify[got.Type].String("") != got.Defaults.String("") {
					t.Fatalf("defaults were not applied\nwant: %v\ngot: %v",
						tc.cfg.Defaults.Notify[got.Type].String(""), got.Defaults.String(""))
				}
			}
		})
	}
}

func TestNotifyTest(t *testing.T) {
	// GIVEN a Config with/without Service containing a Shoutrrr and Root Shoutrrr(s).
	emptyShoutrrr := shoutrrr.NewDefaults(
		"",
		map[string]string{},
		map[string]string{},
		map[string]string{})
	tests := map[string]struct {
		flag                    string
		slice                   service.Slice
		rootSlice               shoutrrr.SliceDefaults
		stdoutRegex, panicRegex *string
	}{
		"empty flag": {flag: "",
			stdoutRegex: test.StringPtr("^$"),
			slice: service.Slice{
				"argus": {
					Notify: shoutrrr.Slice{
						"foo": {},
						"bar": {},
						"baz": {},
					},
				}}},
		"unknown Notifier": {flag: "something",
			panicRegex: test.StringPtr("Notifier.* could not be found"),
			slice: service.Slice{
				"argus": {
					Notify: shoutrrr.Slice{
						"foo": {},
						"bar": {},
						"baz": {},
					},
				}}},
		"known Service Notifier with invalid Gotify token": {
			flag:       "bar",
			panicRegex: test.StringPtr(`Message failed to send with "bar" config\s+invalid gotify token`),
			slice: service.Slice{
				"argus": {
					Notify: shoutrrr.Slice{
						"foo": {},
						"bar": shoutrrr.New(
							nil,
							"bar",
							"gotify",
							map[string]string{},
							map[string]string{
								"host":  "example.com",
								"token": "invalid"},
							map[string]string{
								"max_tries": "1"},
							emptyShoutrrr,
							emptyShoutrrr,
							emptyShoutrrr),
						"baz": {}}}},
		},
		"valid Gotify token": {
			flag:       "bar",
			panicRegex: test.StringPtr(`HTTP 40\d`),
			slice: service.Slice{
				"argus": {
					Notify: shoutrrr.Slice{
						"foo": {},
						"bar": shoutrrr.New(
							nil,
							"bar",
							"gotify",
							map[string]string{
								"max_tries": "1"},
							map[string]string{
								"host":  "example.com",
								"token": "AGdjFCZugzJGhEG"},
							map[string]string{},
							emptyShoutrrr,
							emptyShoutrrr,
							emptyShoutrrr),
						"baz": {}}}},
		},
		"shoutrrr from Root": {
			flag:       "baz",
			panicRegex: test.StringPtr(`HTTP 40\d`),
			slice:      service.Slice{},
			rootSlice: shoutrrr.SliceDefaults{
				"baz": shoutrrr.NewDefaults(
					"gotify",
					map[string]string{
						"max_tries": "1"},
					map[string]string{
						"host":  "example.com",
						"token": "AGdjFCZugzJGhEG"},
					map[string]string{})},
		},
		"successful send": {
			flag:        "work",
			stdoutRegex: test.StringPtr(`Message sent successfully with "work" config`),
			slice: service.Slice{
				"argus": {
					Notify: shoutrrr.Slice{
						"work": shoutrrr.New(
							nil,
							"bar",
							"gotify",
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
			shoutrrrHardDefaults := shoutrrr.SliceDefaults{}
			shoutrrrHardDefaults.Default()
			for i := range tc.slice {
				(*tc.slice[i]).Init(
					&service.Defaults{}, &serviceHardDefaults,
					&tc.rootSlice, &shoutrrr.SliceDefaults{}, &shoutrrrHardDefaults,
					&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})
			}
			if tc.panicRegex != nil {
				// Switch Fatal to panic and disable this panic.
				defer func() {
					r := recover()
					releaseStdout()

					rStr := fmt.Sprint(r)
					if !util.RegexCheck(*tc.panicRegex, rStr) {
						t.Errorf("expected a panic that matched %q\ngot: %q",
							*tc.panicRegex, rStr)
					}
				}()
			}

			// WHEN NotifyTest is called with the test Config.
			cfg := config.Config{
				Service: tc.slice,
				Notify:  tc.rootSlice}
			NotifyTest(&tc.flag, &cfg)

			// THEN we get the expected stdout.
			stdout := releaseStdout()
			if tc.stdoutRegex != nil {
				if !util.RegexCheck(*tc.stdoutRegex, stdout) {
					t.Errorf("want match for %q\nnot: %q",
						*tc.stdoutRegex, stdout)
				}
			}
		})
	}
}
