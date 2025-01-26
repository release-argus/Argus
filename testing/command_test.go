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
	// GIVEN a Config with a Service containing a Command
	tests := map[string]struct {
		flag                    string
		slice                   service.Slice
		stdoutRegex, panicRegex *string
	}{
		"flag is empty": {
			flag:        "",
			stdoutRegex: test.StringPtr("^$"),
			slice: service.Slice{
				"argus": {
					ID: "argus",
					Command: command.Slice{
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
			slice: service.Slice{
				"argus": {
					ID: "argus",
					Command: command.Slice{
						{"true", "0"}},
					CommandController: &command.Controller{},
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil)}},
		},
		"known service in flag successful command": {
			flag:        "argus",
			stdoutRegex: test.StringPtr(`Executing 'echo command did run'\s+.*command did run\s+`),
			slice: service.Slice{
				"argus": {
					ID: "argus",
					Command: command.Slice{
						{"echo", "command did run"}},
					CommandController: &command.Controller{},
					Options: *opt.New(
						nil, "0s", nil,
						nil, nil)}},
		},
		"known service in flag failing command": {
			flag:        "argus",
			stdoutRegex: test.StringPtr(`.*Executing 'ls /root'\s+.*exit status [1-9]\s+`),
			slice: service.Slice{
				"argus": {
					ID: "argus",
					Command: command.Slice{
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
			slice: service.Slice{
				"argus": {
					ID: "argus"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
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
			for i := range tc.slice {
				tc.slice[i].Status.ServiceID = &tc.slice[i].ID
			}

			// WHEN CommandTest is called with the test Config
			if tc.slice[tc.flag] != nil && tc.slice[tc.flag].CommandController != nil {
				tc.slice[tc.flag].CommandController.Init(
					&tc.slice[tc.flag].Status,
					&tc.slice[tc.flag].Command,
					nil,
					&tc.slice[tc.flag].Options.Interval)
			}
			order := []string{}
			for i := range tc.slice {
				order = append(order, i)
			}
			cfg := config.Config{
				Service: tc.slice,
				Order:   order,
			}
			CommandTest(&tc.flag, &cfg)

			// THEN we get the expected stdout
			stdout := releaseStdout()
			if tc.stdoutRegex != nil {
				if !util.RegexCheck(*tc.stdoutRegex, stdout) {
					t.Errorf("want match for %q\ngot: %q",
						*tc.stdoutRegex, stdout)
				}
			}
		})
	}
}
