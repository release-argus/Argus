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

func TestNotifyTest(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig()
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig()

	tests := []struct {
		name          string
		flag          string
		services      service.Services
		mainShoutrrrs shoutrrr.ShoutrrrsDefaults
		stdoutRegex   string
		ok            bool
	}{
		{
			name:        "empty flag",
			flag:        "",
			ok:          true,
			stdoutRegex: "^$",
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
			stdoutRegex: "FATAL: .*Notifier .* could not be found",
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
			stdoutRegex: `Message failed to send with "bar" config\s+.*invalid gotify token`,
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
										host: `+test.NotifyGotify.HostValid+`
										token: `+test.NotifyGotify.TokenMalformed+`
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
			stdoutRegex: `invalid gotify token: "abc"`,
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
										host: `+test.NotifyGotify.HostValid+`
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
			stdoutRegex: `Message failed to send with.*\s.*server responded`,
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
										host: `+test.NotifyGotify.HostValid+`
										path: `+test.NotifyGotify.Path+`
										token: `+test.NotifyGotify.TokenRejected+`
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
			stdoutRegex: `Message failed to send with.*\s.*server responded`,
			services:    service.Services{},
			mainShoutrrrs: shoutrrr.ShoutrrrsDefaults{
				"baz": shoutrrr.NewDefaults(
					"gotify",
					map[string]string{
						"max_tries": "1",
					},
					map[string]string{
						"host":  test.NotifyGotify.HostValid,
						"path":  test.NotifyGotify.Path,
						"token": test.NotifyGotify.TokenRejected,
					},
					map[string]string{},
				),
			},
		},
		{
			name:        "successful send",
			flag:        "work",
			ok:          true,
			stdoutRegex: `Message sent successfully with "work" config`,
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
										host: `+test.NotifyGotify.HostValid+`
										path: `+test.NotifyGotify.Path+`
										token: `+test.NotifyGotify.TokenPass+`
								baz: {}
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			cfg := config.Config{
				Service: tc.services,
				Notify:  tc.mainShoutrrrs,
			}

			test.AssertCallLog(
				t,
				tc.ok,
				func() bool { return NotifyTest(&tc.flag, &cfg) },
				logx.ExitCodeChannel(),
				releaseStdout,
				tc.stdoutRegex,
			)
		})
	}
}
