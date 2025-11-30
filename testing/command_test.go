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

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/service"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestCommandTest(t *testing.T) {
	// GIVEN a Config with a Service containing a Command.
	tests := map[string]struct {
		flag                    string
		services                service.Services
		stdoutRegex, panicRegex *string
	}{
		"flag is empty": {
			flag:        "",
			stdoutRegex: test.StringPtr("^$"),
			services: service.Services{
				"argus": {
					ID: "argus",
					Command: command.Commands{
						{"true", "0"}},
					CommandController: &command.Controller{},
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil)}},
		},
		"unknown service in flag": {
			flag:        "something",
			panicRegex:  test.StringPtr(" could not be found "),
			stdoutRegex: test.StringPtr("should have panic'd before reaching this"),
			services: service.Services{
				"argus": {
					ID: "argus",
					Command: command.Commands{
						{"true", "0"}},
					CommandController: &command.Controller{},
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil)}},
		},
		"known service in flag successful command": {
			flag:        "argus",
			stdoutRegex: test.StringPtr(`Executing 'echo command did run'\s+.*command did run\s+`),
			services: service.Services{
				"argus": {
					ID: "argus",
					Command: command.Commands{
						{"echo", "command did run"}},
					CommandController: &command.Controller{},
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil)}},
		},
		"known service in flag failing command": {
			flag:        "argus",
			stdoutRegex: test.StringPtr(`.*Executing 'ls /root'\s+.*exit status [1-9]\s+`),
			services: service.Services{
				"argus": {
					ID: "argus",
					Command: command.Commands{
						{"ls", "/root"}},
					CommandController: &command.Controller{},
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil)}},
		},
		"service with no commands": {
			flag:        "argus",
			panicRegex:  test.StringPtr(" does not have any `command` defined"),
			stdoutRegex: test.StringPtr("should have panic'd before reaching this"),
			services: service.Services{
				"argus": {
					ID: "argus"}},
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
						t.Errorf("%s\npanic mismatch\nwant: %q\ngot:  %q",
							packageName, *tc.panicRegex, rStr)
					}
				}()
			}
			for i := range tc.services {
				tc.services[i].Status.ServiceInfo.ID = tc.services[i].ID
			}

			// WHEN CommandTest is called with the test Config.
			if tc.services[tc.flag] != nil && tc.services[tc.flag].CommandController != nil {
				tc.services[tc.flag].CommandController.Init(
					&tc.services[tc.flag].Status,
					&tc.services[tc.flag].Command,
					nil,
					&tc.services[tc.flag].Options.Interval)
			}
			order := make([]string, len(tc.services))
			for i := range tc.services {
				order = append(order, i)
			}
			cfg := config.Config{
				Service: tc.services,
				Order:   order,
			}
			CommandTest(&tc.flag, &cfg)

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
