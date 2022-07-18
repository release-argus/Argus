// Copyright [2022] [Argus]
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
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/utils"
)

func TestCommandTest(t *testing.T) {
	// GIVEN a Config with a Service containing a Command
	jLog = utils.NewJLog("INFO", false)
	InitJLog(jLog)
	tests := map[string]struct {
		flag        string
		slice       service.Slice
		outputRegex *string
		panicRegex  *string
	}{
		"flag is empty": {flag: "",
			outputRegex: stringPtr("^$"),
			slice: service.Slice{
				"argus": {
					ID: stringPtr("argus"),
					Command: &command.Slice{
						command.Command{"true", "0"},
					},
					CommandController: &command.Controller{},
					Interval:          stringPtr("0s"),
				},
			}},
		"unknown service in flag": {flag: "something",
			panicRegex:  stringPtr(" could not be found "),
			outputRegex: stringPtr("should have panic'd before reaching this"),
			slice: service.Slice{
				"argus": {
					ID: stringPtr("argus"),
					Command: &command.Slice{
						command.Command{"true", "0"},
					},
					CommandController: &command.Controller{},
					Interval:          stringPtr("0s"),
				},
			}},
		"known service in flag successful command": {flag: "argus",
			outputRegex: stringPtr(`Executing 'echo command did run'\s+.*command did run\s+`),
			slice: service.Slice{
				"argus": {
					ID: stringPtr("argus"),
					Command: &command.Slice{
						command.Command{"echo", "command did run"},
					},
					CommandController: &command.Controller{},
					Interval:          stringPtr("0s"),
				},
			}},
		"known service in flag failing command": {flag: "argus",
			outputRegex: stringPtr(`.*Executing 'ls /root'\s+.*exit status 2\s+`),
			slice: service.Slice{
				"argus": {
					ID: stringPtr("argus"),
					Command: &command.Slice{
						command.Command{"ls", "/root"},
					},
					CommandController: &command.Controller{},
					Interval:          stringPtr("0s"),
				},
			}},
		"service with no commands": {flag: "argus",
			panicRegex:  stringPtr(" does not have any `command` defined"),
			outputRegex: stringPtr("should have panic'd before reaching this"),
			slice: service.Slice{
				"argus": {
					ID: stringPtr("argus"),
				},
			}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			jLog.Testing = true
			if tc.panicRegex != nil {
				// Switch Fatal to panic and disable this panic.
				defer func() {
					r := recover()
					rStr := fmt.Sprint(r)
					re := regexp.MustCompile(*tc.panicRegex)
					match := re.MatchString(rStr)
					if !match {
						t.Errorf("%s:\nexpected a panic that matched %q\ngot: %q",
							name, *tc.panicRegex, rStr)
					}
				}()
			}

			// WHEN CommandTest is called with the test Config
			if tc.slice[tc.flag] != nil && tc.slice[tc.flag].CommandController != nil {
				tc.slice[tc.flag].CommandController.Init(jLog, &tc.flag, nil, tc.slice[tc.flag].Command, nil, tc.slice[tc.flag].Interval)
			}
			cfg := config.Config{
				Service: tc.slice,
			}
			CommandTest(&tc.flag, &cfg, jLog)

			// THEN we get the expected output
			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = stdout
			output := string(out)
			if tc.outputRegex != nil {
				re := regexp.MustCompile(*tc.outputRegex)
				match := re.MatchString(output)
				if !match {
					t.Errorf("%s:\nwant match for %q\non: %q",
						name, *tc.outputRegex, output)
				}
			}
		})
	}
}
