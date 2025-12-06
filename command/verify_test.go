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

package command

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestCommand_CheckValues(t *testing.T) {
	// GIVEN a Command.
	tests := map[string]struct {
		command  *Command
		errRegex string
	}{
		"nil command": {
			errRegex: `^$`,
			command:  nil},
		"valid command": {
			errRegex: `^$`,
			command:  &Command{"ls", "-la"}},
		"invalid command template": {
			errRegex: `^.+ (.+) <invalid>`,
			command:  &Command{"ls", "-la", "{{ version }"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called.
			err := tc.command.CheckValues()

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestCommandSlice_CheckValues(t *testing.T) {
	// GIVEN a Commands.
	tests := map[string]struct {
		commands *Commands
		errRegex string
	}{
		"nil slice": {
			errRegex: `^$`,
			commands: nil},
		"valid slice": {
			errRegex: `^$`,
			commands: &Commands{
				{"ls", "-la"}}},
		"invalid templating": {
			errRegex: test.TrimYAML(`
				^item_1: .+ \(.+\) <invalid>.*
				item_3: .+ \(.+\) <invalid>.*$`),
			commands: &Commands{
				{"ls"},
				{"ls", "-la", "{{ version }"},
				{"ls"},
				{"ls", "-la", "{{ version }"}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called.
			err := tc.commands.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nerror line count\nwant: %d\n%q\ngot:  %d:\n%v\n\nstdout: %q",
					packageName, wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
		})
	}
}
