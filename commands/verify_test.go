// Copyright [2023] [Argus]
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
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestCommand_CheckValues(t *testing.T) {
	// GIVEN a Command
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called
			err := tc.command.CheckValues()

			// THEN it err's when expected
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}
func TestCommandSlice_CheckValues(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice    *Slice
		errRegex []string
	}{
		"nil slice": {
			errRegex: []string{`^$`},
			slice:    nil},
		"valid slice": {
			errRegex: []string{`^$`},
			slice: &Slice{
				{"ls", "-la"}}},
		"invalid templating": {
			errRegex: []string{
				`^command:$`,
				`^  item_1: .+ (.+) <invalid>`},
			slice: &Slice{
				{"ls"},
				{"ls", "-la", "{{ version }"}}},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called
			err := tc.slice.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			lines := strings.Split(e, `\`)
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				found := false
				for j := range lines {
					match := re.MatchString(lines[j])
					if match {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("want match for: %q\ngot:  %q",
						tc.errRegex[i], strings.ReplaceAll(e, `\`, "\n"))
				}
			}
		})
	}
}
