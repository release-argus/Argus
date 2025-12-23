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

// Package testing provides utilities for CLI-based testing.
package testing

import (
	"testing"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/service"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestCommandTest(t *testing.T) {
	// GIVEN a Config with a Service containing a Command.
	tests := map[string]struct {
		flag        string
		services    service.Services
		stdoutRegex *string
		ok          bool
	}{
		"flag is empty": {
			flag:        "",
			ok:          true,
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
			ok:          false,
			stdoutRegex: test.StringPtr(" could not be found "),
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
			ok:          true,
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
			ok:          true,
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
			ok:          false,
			stdoutRegex: test.StringPtr(" does not have any `command` defined"),
			services: service.Services{
				"argus": {
					ID: "argus"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(logutil.Log)

			if tc.services[tc.flag] != nil && tc.services[tc.flag].CommandController != nil {
				tc.services[tc.flag].CommandController.Init(
					&tc.services[tc.flag].Status,
					&tc.services[tc.flag].Command,
					nil,
					&tc.services[tc.flag].Options.Interval)
			}
			defaults := service.Defaults{}
			defaults.Default()
			for _, svc := range tc.services {
				svc.Init(
					&defaults, &defaults,
					nil, nil, nil,
					nil, nil, nil)
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
			// WHEN CommandTest is called with the test Config.
			resultChannel <- CommandTest(&tc.flag, &cfg)

			// THEN we get the expected stdout.
			stdout := releaseStdout()
			if tc.stdoutRegex != nil {
				if !util.RegexCheck(*tc.stdoutRegex, stdout) {
					t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, *tc.stdoutRegex, stdout)
				}
			}
			// AND it succeeds/fails as expected.
			if err := test.OkMatch(t, tc.ok, resultChannel, logutil.ExitCodeChannel(), nil); err != nil {
				t.Fatalf("%s\n%s",
					packageName, err.Error())
			}
		})
	}
}
