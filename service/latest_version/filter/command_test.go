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

package filter

import (
	"testing"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestRequire_ExecCommand(t *testing.T) {
	// GIVEN a Require with a Command.
	tests := map[string]struct {
		cmd      command.Command
		errRegex string
	}{
		"no command": {
			errRegex: `^$`},
		"valid command": {
			cmd:      []string{"true"},
			errRegex: `^$`},
		"valid multi-arg command": {
			cmd:      []string{"ls", "-lah"},
			errRegex: `^$`},
		"invalid command": {
			cmd:      []string{"false"},
			errRegex: `exit status 1`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			require := Require{Command: tc.cmd}
			require.Status = &status.Status{}
			require.Status.Init(
				0, 1, 0,
				&name, nil,
				test.StringPtr("http://example.com"))

			// WHEN ApplyTemplate is called on the Command.
			err := require.ExecCommand(logutil.LogFrom{})

			// THEN the err is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}
