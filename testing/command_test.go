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

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	shoutrrrtest "github.com/release-argus/Argus/notify/shoutrrr/test"
	"github.com/release-argus/Argus/service"
	svctest "github.com/release-argus/Argus/service/test"
	"github.com/release-argus/Argus/util"
	whtest "github.com/release-argus/Argus/webhook/test"
)

func TestCommandTest(t *testing.T) {
	svcCfg := svctest.PlainDefaultsConfig(t)
	notifyCfg := shoutrrrtest.PlainConfig(t)
	whCfg := whtest.PlainConfig(t)

	// GIVEN: a Config with a Service containing a Command.
	tests := []struct {
		name        string
		flag        string
		services    service.Services
		stdoutRegex *string
		ok          bool
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
			name:        "unknown service",
			flag:        "something",
			ok:          false,
			stdoutRegex: test.Ptr(" could not be found "),
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
			name:        "known service - successful command",
			flag:        "argus",
			ok:          true,
			stdoutRegex: test.Ptr(`Executing 'echo command did run'\s+.*command did run\s+`),
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
			name:        "known service - failing command",
			flag:        "argus",
			ok:          true,
			stdoutRegex: test.Ptr(`.*Executing 'false'\s+.*exit status [1-9]\s+`),
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
			name:        "service - no commands",
			flag:        "argus",
			ok:          false,
			stdoutRegex: test.Ptr(" does not have any `command` defined"),
			services: service.Services{
				"argus": {
					ID: "argus",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			if tc.services[tc.flag] != nil && tc.services[tc.flag].CommandController != nil {
				tc.services[tc.flag].CommandController = command.NewController(
					&tc.services[tc.flag].Status,
					tc.services[tc.flag].Command,
					nil,
					&tc.services[tc.flag].Options.Interval,
				)
			}

			order := make([]string, len(tc.services))
			for i := range tc.services {
				order = append(order, i)
			}
			cfg := config.Config{
				Service: tc.services,
				Order:   order,
			}

			resultChannel := make(chan bool, 1)
			// WHEN: CommandTest is called with the test Config.
			resultChannel <- CommandTest(&tc.flag, &cfg)

			prefix := fmt.Sprintf(
				"%s\nCommandTest(%q)",
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
						"%s error mismatch\ngot:  %q\nwant: %q",
						prefix, stdout, *tc.stdoutRegex,
					)
				}
			}
		})
	}
}
