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

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	svctest "github.com/release-argus/Argus/service/test"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestCommandTest(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig()
	whCfg := whtest.PlainConfig(t)

	tests := []struct {
		name        string
		flag        string
		services    service.Services
		stdoutRegex string
		ok          bool
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
			name:        "unknown service",
			flag:        "something",
			ok:          false,
			stdoutRegex: " could not be found ",
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
			name:        "known service/successful command",
			flag:        "argus",
			ok:          true,
			stdoutRegex: `Executing "echo command did run"\s+.*command did run\s+`,
			services: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						argus:
							options:
								interval: 0s
							command:
								- - 	echo
									- command did run
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name:        "known service/failing command",
			flag:        "argus",
			ok:          true, // A failing command still returns ok.
			stdoutRegex: `.*Executing "false"\s+.*exit status [1-9]\s+`,
			services: test.Must(t, func() (service.Services, error) {
				return service.DecodeServices(
					"yaml", []byte(test.TrimYAML(`
						argus:
							options:
								interval: 0s
							command:
								- - false
					`)),
					svcCfg, notifyCfg, whCfg,
				)
			}),
		},
		{
			name:        "known service/no commands",
			flag:        "argus",
			ok:          false,
			stdoutRegex: " does not have any `command` defined",
			services: service.Services{
				"argus": {
					ID: "argus",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			if tc.services[tc.flag] != nil && tc.services[tc.flag].CommandController != nil {
				tc.services[tc.flag].CommandController = command.NewController(
					&tc.services[tc.flag].Status,
					tc.services[tc.flag].Command,
					nil,
					&tc.services[tc.flag].Options.Interval,
				)
			}

			order := make([]string, 0, len(tc.services))
			for i := range tc.services {
				order = append(order, i)
			}
			cfg := config.Config{
				Service: tc.services,
				Order:   order,
			}

			test.AssertCallLog(
				t,
				tc.ok,
				func() bool { return CommandTest(&tc.flag, &cfg) },
				logx.ExitCodeChannel(),
				releaseStdout,
				tc.stdoutRegex,
			)
		})
	}
}
