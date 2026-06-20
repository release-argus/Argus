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

package filter

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestRequire_ExecCommand(t *testing.T) {
	// GIVEN: a Require with a Command.
	tests := []struct {
		name        string
		cmd         command.Command
		sInfo       *serviceinfo.ServiceInfo
		version     string
		stdoutRegex string
		errRegex    string
	}{
		{
			name:     "no command",
			errRegex: `^$`,
		},
		{
			name:     "valid command",
			cmd:      []string{"true"},
			errRegex: `^$`,
		},
		{
			name:     "valid multi-arg command",
			cmd:      []string{"ls", "-lah"},
			errRegex: `^$`,
		},
		{
			name:     "invalid command",
			cmd:      []string{"false"},
			errRegex: `exit status 1`,
		},
		{
			name: "new version overrides previous latest_version",
			cmd: []string{
				"echo",
				"approved_version='{{ approved_version}}', deployed_version='{{ deployed_version }}', version='{{ version }}', latest_version='{{ latest_version }}',",
			},
			sInfo: &serviceinfo.ServiceInfo{
				ApprovedVersion: "v1.0.0-approved",
				DeployedVersion: "v2.0.0-deployed",
				LatestVersion:   "v3.0.0-latest",
			},
			version:     "v4.0.0",
			stdoutRegex: `approved_version='v1.0.0-approved', deployed_version='v2.0.0-deployed', version='v4.0.0', latest_version='v4.0.0'`,
			errRegex:    `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout and sharing log resultChannel.
			releaseStdout := test.CaptureLog(t, logx.Default())

			svcDashboard := &dashboard.Options{
				OptionsBase: dashboard.OptionsBase{
					WebURL: "https://example.com",
				},
			}
			require := Require{Command: tc.cmd}
			require.Status = &status.Status{}
			require.Status.Init(
				0, 1, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				svcDashboard,
			)
			if tc.sInfo != nil {
				require.Status.ServiceInfo = *tc.sInfo
			}

			// WHEN: ApplyTemplate is called on the Command.
			err := require.ExecCommand(tc.version, logx.LogFrom{})

			prefix := fmt.Sprintf(
				"%s\nRequire ExecCommand(%q)",
				packageName, require.Command,
			)

			// THEN: the decode errors as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the stdout is expected.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Fatalf(
					"%s stdout mismatch\nwant: %q\ngot:  %q",
					prefix, tc.stdoutRegex, stdout,
				)
			}
		})
	}
}
