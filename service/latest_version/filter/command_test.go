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

package filter

import (
	"regexp"
	"testing"

	command "github.com/release-argus/Argus/commands"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestRequire_ExecCommand(t *testing.T) {
	// GIVEN a Require with a Command
	testLogging("WARN")
	tests := map[string]struct {
		cmd      command.Command
		errRegex string
	}{
		"no command": {
			errRegex: "^$"},
		"valid command": {
			cmd:      []string{"true"},
			errRegex: "^$"},
		"valid multi-arg command": {
			cmd:      []string{"ls", "-lah"},
			errRegex: "^$"},
		"invalid command": {
			cmd:      []string{"false"},
			errRegex: "exit status 1"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require := Require{Command: tc.cmd}
			require.Status = &svcstatus.Status{}

			// WHEN ApplyTemplate is called on the Command
			err := require.ExecCommand(&util.LogFrom{})

			// THEN the err is expected
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Errorf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}
