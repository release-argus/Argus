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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_Query(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name                        string
		env                         map[string]string
		overrides, optionsOverrides string
		errRegex                    string
		wantVersion                 string
	}{
		{
			name: "version from stdout",
			overrides: test.TrimYAML(`
				command:
				  - printf
				  - 1.2.3
			`),
			wantVersion: `^[0-9.]+\.[0-9.]+\.[0-9.]+$`,
			errRegex:    `^$`,
		},
		{
			name: "version with regex capture group",
			overrides: test.TrimYAML(`
				command:
				  - sh
				  - -c
				  - "printf 'version 1.2.3 rest'"
				regex: version ([0-9.]+)
			`),
			wantVersion: `^1\.2\.3$`,
			errRegex:    `^$`,
		},
		{
			name: "regex without capture group",
			overrides: test.TrimYAML(`
				command:
				  - printf
				  - 1.2.3
				regex: '[0-9.]+'
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^[0-9.]+$`,
			errRegex:         `^$`,
		},
		{
			name: "failing regex",
			overrides: test.TrimYAML(`
				command:
				  - printf
				  - 1.2.3
				regex: '^bishBashBosh$'
			`),
			errRegex: `didn't return any matches`,
		},
		{
			name: "empty stdout",
			overrides: test.TrimYAML(`
				command:
				  - sh
				  - -c
				  - "printf ''"
			`),
			errRegex: `no version found`,
		},
		{
			name: "command failure",
			overrides: test.TrimYAML(`
				command:
				  - false
			`),
			errRegex: `failed running command`,
		},
		{
			name: "no command",
			overrides: test.TrimYAML(`
				regex: 'v([0-9.]+)'
			`),
			errRegex: `no command specified`,
		},
		{
			name: "command arg from env",
			env: map[string]string{
				"TEST_LOOKUP__DV_QUERY_CMD": "1.2.3",
			},
			overrides: test.TrimYAML(`
				command:
				  - printf
				  - ${TEST_LOOKUP__DV_QUERY_CMD}
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^[0-9.]+$`,
			errRegex:         `^$`,
		},
		{
			name: "want semantic versioning but get non-semantic version",
			overrides: test.TrimYAML(`
				command:
				  - printf
				  - dev
			`),
			optionsOverrides: `semantic_versioning: true`,
			errRegex:         `failed to convert`,
		},
		{
			name: "allow non-semantic version",
			overrides: test.TrimYAML(`
				command:
				  - printf
				  - dev
			`),
			optionsOverrides: `semantic_versioning: false`,
			wantVersion:      `^dev$`,
			errRegex:         `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.SetEnv(t, tc.env)
			dvl := testLookup(t, nil, "")
			if err := dvl.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Lookup overrides: %s",
					packageName, err,
				)
			}
			if tc.optionsOverrides != "" {
				if err := decode.Unmarshal("yaml", []byte(tc.optionsOverrides), dvl.Options); err != nil {
					t.Fatalf(
						"%s\nfailed to unmarshal Lookup.Options overrides: %s",
						packageName, err,
					)
				}
			}

			// WHEN: Query is called on it.
			err := dvl.Query(true, logx.LogFrom{})

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s\nLookup.Query() error mismatch\ngot:  %q\nwant: %q",
					packageName, e, tc.errRegex,
				)
			}

			// AND: the version matches the expected regex.
			if tc.wantVersion != "" {
				if version := dvl.Status.DeployedVersion(); !util.RegexCheck(tc.wantVersion, version) {
					t.Errorf(
						"%s\nLookup.Query() .DeployedVersion() mismatch\ngot:  %q\nwant %q",
						packageName, version, tc.wantVersion,
					)
				}
			}
		})
	}
}

func TestLookup_getVersion(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name        string
		cmd         []string
		regex       string
		semVer      bool
		wantVersion string
		errRegex    string
	}{
		{
			name:     "no command",
			cmd:      nil,
			errRegex: `no command specified`,
		},
		{
			name:        "plain stdout",
			cmd:         []string{"printf", "1.2.3"},
			semVer:      true,
			wantVersion: "1.2.3",
			errRegex:    `^$`,
		},
		{
			name:        "regex with capture group",
			cmd:         []string{"sh", "-c", `printf 'version 1.2.3 rest'`},
			regex:       `version ([0-9.]+)`,
			semVer:      true,
			wantVersion: "1.2.3",
			errRegex:    `^$`,
		},
		{
			name:        "regex without capture group",
			cmd:         []string{"sh", "-c", `printf 'version 1.2.3 rest'`},
			regex:       `[0-9.]+`,
			semVer:      true,
			wantVersion: "1.2.3",
			errRegex:    `^$`,
		},
		{
			name:     "regex no match",
			cmd:      []string{"printf", "1.2.3"},
			regex:    `^bishBashBosh$`,
			semVer:   true,
			errRegex: `^regex .* didn't return any matches on "1\.2\.3"$`,
		},
		{
			name:     "empty output",
			cmd:      []string{"sh", "-c", `printf ''`},
			semVer:   true,
			errRegex: `^no version found in command .* output$`,
		},
		{
			name:     "command failure",
			cmd:      []string{"false"},
			semVer:   true,
			errRegex: `^failed running command "false": `,
		},
		{
			name:     "semantic versioning rejects non-semantic version",
			cmd:      []string{"printf", "dev"},
			semVer:   true,
			errRegex: `^failed to convert .* to a semantic version`,
		},
		{
			name:        "semantic versioning disabled allows non-semantic version",
			cmd:         []string{"printf", "dev"},
			semVer:      false,
			wantVersion: "dev",
			errRegex:    `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, tc.cmd, tc.regex)
			lookup.Options.SemanticVersioning = &tc.semVer

			version, err := lookup.getVersion(logx.LogFrom{})

			prefix := fmt.Sprintf("%s\nLookup.getVersion()", packageName)

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if got, want := version, tc.wantVersion; got != want {
				t.Errorf(
					"%s version mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}
		})
	}
}