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

func TestNotifyTest(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	// GIVEN: a Config with/without Service containing a Shoutrrr and Root Shoutrrrs.
	tests := []struct {
		name          string
		flag          string
		services      service.Services
		mainShoutrrrs shoutrrr.ShoutrrrsDefaults
		stdoutRegex   *string
		ok            bool
	}{
		{
			name:        "empty flag",
			flag:        "",
			ok:          true,
			stdoutRegex: test.Ptr("^$"),
			services: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						argus:
							options:
								interval: 0s
							command:
								- - true
									- 0
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name:        "unknown Notifier",
			flag:        "something",
			ok:          false,
			stdoutRegex: test.Ptr("FATAL: .*Notifier .* could not be found"),
			services: test.Must(t, func() (service.Services, error) {
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
		{
			name:        "known Service Notifier with invalid Gotify token",
			flag:        "bar",
			ok:          false,
			stdoutRegex: test.Ptr(`Message failed to send with "bar" config\s+.*invalid gotify token`),
			services: test.Must(t, func() (service.Services, error) {
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
										token: invalid
								baz: {}
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name:        "invalid Gotify token format",
			flag:        "bar",
			ok:          false,
			stdoutRegex: test.Ptr(`invalid gotify token: "abc"`),
			services: test.Must(t, func() (service.Services, error) {
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
										token: abc
								baz: {}
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name:        "valid Gotify token format",
			flag:        "bar",
			ok:          false,
			stdoutRegex: test.Ptr(`Message failed to send with.*\s.*server responded`),
			services: test.Must(t, func() (service.Services, error) {
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
										token: AGdjFCZugzJGhEG
								baz: {}
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name:        "shoutrrr from Root",
			flag:        "baz",
			ok:          false,
			stdoutRegex: test.Ptr(`Message failed to send with.*\s.*server responded`),
			services:    service.Services{},
			mainShoutrrrs: shoutrrr.ShoutrrrsDefaults{
				"baz": shoutrrr.NewDefaults(
					"gotify",
					map[string]string{
						"max_tries": "1",
					},
					map[string]string{
						"host":  test.ValidCertNoProtocol,
						"token": "AGdjFCZugzJGhEG",
					},
					map[string]string{},
				),
			},
		},
		{
			name:        "successful send",
			flag:        "work",
			ok:          true,
			stdoutRegex: test.Ptr(`Message sent successfully with "work" config`),
			services: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						argus:
							notify:
								foo: {}
								work:
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			serviceHardDefaults := service.Defaults{}
			serviceHardDefaults.Default()
			shoutrrrHardDefaults := shoutrrr.ShoutrrrsDefaults{}
			shoutrrrHardDefaults.Default()

			resultChannel := make(chan bool, 1)
			// WHEN: NotifyTest is called with the test Config.
			cfg := config.Config{
				Service: tc.services,
				Notify:  tc.mainShoutrrrs,
			}
			resultChannel <- NotifyTest(&tc.flag, &cfg)

			prefix := fmt.Sprintf(
				"%s\nNotifyTest(%q)",
				packageName, tc.flag,
			)

			// THEN: it succeeds/fails as expected.
			if err := test.AssertChannelBool(
				t,
				tc.ok,
				resultChannel,
				logx.ExitCodeChannel(),
				nil,
			); err != nil {
				t.Fatal(prefix + err.Error())
			}

			// AND: we get the expected stdout.
			stdout := releaseStdout()
			if tc.stdoutRegex != nil {
				if !util.RegexCheck(*tc.stdoutRegex, stdout) {
					t.Errorf(
						"%s stdout mismatch\ngot:  %q\nwant: %q",
						prefix, stdout, *tc.stdoutRegex,
					)
				}
			}
		})
	}
}
