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

//go:build unit

package command

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestCommands_CheckValues(t *testing.T) {
	// GIVEN: a Commands.
	tests := []struct {
		name     string
		input    *Commands
		errRegex string
	}{
		{
			name:     "nil slice",
			errRegex: `^$`,
			input:    (*Commands)(nil),
		},
		{
			name:     "valid slice",
			errRegex: `^$`,
			input: &Commands{
				{"ls", "-la"},
			},
		},
		{
			name: "invalid templating",
			errRegex: test.TrimYAML(`
				^- item_1:
				  "ls.+ \(.+\) <invalid>.*
				- item_3:
				  "ls.+ \(.+\) <invalid>.*$`,
			),
			input: &Commands{
				{"ls"},
				{"ls", "-la", "{{ version }"},
				{"ls"},
				{"ls", "-la", "{{ version }"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)
		})
	}
}

func TestCommand_CheckValues(t *testing.T) {
	// GIVEN: a Command.
	tests := []struct {
		name     string
		input    *Command
		errRegex string
	}{
		{
			name:     "nil command",
			errRegex: `^$`,
			input:    (*Command)(nil),
		},
		{
			name:     "valid command",
			errRegex: `^$`,
			input: &Command{
				"ls",
				"-la",
			},
		},
		{
			name:     "invalid command template",
			errRegex: `^.+ (.+) <invalid>`,
			input: &Command{
				"ls",
				"-la",
				"{{ version }",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: CheckValues is called.
			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)
		})
	}
}
