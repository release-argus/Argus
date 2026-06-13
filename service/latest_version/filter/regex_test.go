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
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	ghtypes "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestRequire_RegexCheckVersion(t *testing.T) {
	// GIVEN: a Require.
	tests := []struct {
		name     string
		require  *Require
		errRegex string
	}{
		{
			name:     "nil require",
			require:  nil,
			errRegex: `^$`,
		},
		{
			name:     "empty regex_version",
			require:  &Require{},
			errRegex: `^$`,
		},
		{
			name: "match",
			require: &Require{
				RegexVersion: "^[0-9.]+-beta$",
			},
			errRegex: `^$`,
		},
		{
			name: "no match",
			require: &Require{
				RegexVersion: "^[0-9.]+$",
			},
			errRegex: `regex "[^"]+" not matched on version`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.require != nil {
				tc.require.Status = &status.Status{}
			}
			v := "0.1.1-beta"

			// WHEN: RegexCheckVersion is called on it.
			err := tc.require.RegexCheckVersion(v, logx.LogFrom{})

			// THEN: the decode is what we expect.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s\nRequire.RegexCheckVersion(%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, v,
					e, tc.errRegex,
				)
			}
		})
	}
}

func TestRequire_RegexCheckContent(t *testing.T) {
	// GIVEN: a Require.
	tests := []struct {
		name     string
		require  *Require
		body     string
		errRegex string
	}{
		{
			name:     "nil require",
			require:  nil,
			errRegex: `^$`,
		},
		{
			name:     "empty regex_content",
			require:  &Require{},
			errRegex: `^$`,
		},
		{
			name: "match",
			require: &Require{
				RegexContent: `argus-[0-9.]+.linux-amd64`,
			},
			errRegex: `^$`,
			body:     `darwin amd64 - argus-1.2.3.darwin-amd64, linux amd64 - argus-1.2.3.linux-amd64, windows amd64 - argus-1.2.3.windows-amd64,`,
		},
		{
			name: "no match",
			require: &Require{
				RegexContent: `argus-[0-9.]+.linux-amd64`,
			},
			errRegex: `regex .* not matched on content`,
			body:     `darwin amd64 - argus-1.2.3.darwin-amd64, linux arm64 - argus-1.2.3.linux-arm64, windows amd64 - argus-1.2.3.windows-amd64,`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.require != nil {
				tc.require.Status = &status.Status{}
			}
			v := "0.1.1-beta"

			// WHEN: RegexCheckContent is called on it.
			err := tc.require.RegexCheckContent(v, tc.body, logx.LogFrom{})

			// THEN: the decode is what we expect.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s\nRequire.RegexCheckContent(version=%q, body=%q) error mismatch\ngot:  %q\nwant: %q",
					packageName, v, tc.body,
					e, tc.errRegex,
				)
			}
		})
	}
}

func TestRequire_RegexCheckContentGitHub(t *testing.T) {
	// GIVEN: a Require.
	tests := []struct {
		name                  string
		require               *Require
		body                  []ghtypes.Asset
		wantReleaseDate       string
		stdoutRegex, errRegex string
	}{
		{
			name:        "nil require",
			require:     nil,
			stdoutRegex: `^$`,
			errRegex:    `^$`,
		},
		{
			name:        "empty regex_content",
			require:     &Require{},
			stdoutRegex: `^$`,
			errRegex:    `^$`,
		},
		{
			name: "github api, body match",
			require: &Require{
				RegexContent: `argus-[0-9.]+.linux-amd64`,
			},
			stdoutRegex: `^(DEBUG:.*\s){3}$`, // 3: name+browser_download_url for darwin, name for linux.
			errRegex:    `^$`,
			body: []ghtypes.Asset{
				{Name: "argus-1.2.3.darwin-amd64", CreatedAt: "2020-01-01T00:00:00Z"},
				{Name: "argus-1.2.3.linux-amd64", CreatedAt: "2021-01-01T00:00:00Z"},
				{Name: "argus-1.2.3.windows-amd64", CreatedAt: "2022-01-01T00:00:00Z"},
			},
			wantReleaseDate: "2021-01-01T00:00:00Z",
		},
		{
			name: "github api, body no match",
			require: &Require{
				RegexContent: `argus-[0-9.]+.linux-amd64`,
			},
			stdoutRegex: `^(DEBUG:.*\s){6}INFO: regex.*not matched on content.*\s$`, // 6: name+browser_download_url for darwin/linux/windows.
			errRegex:    `^regex .* not matched on content.*$`,
			body: []ghtypes.Asset{
				{Name: "argus-1.2.3.darwin-amd64"},
				{Name: "argus-1.2.3.linux-arm64"},
				{Name: "argus-1.2.3.windows-amd64"},
			},
		},
		{
			name: "github api, missing created_at",
			require: &Require{
				RegexContent: `argus-[0-9.]+.linux-amd64`,
			},
			stdoutRegex: `^DEBUG:.*\s$`,
			errRegex:    `^$`,
			body: []ghtypes.Asset{
				{Name: "argus-1.2.3.linux-amd64", CreatedAt: ""},
			},
			wantReleaseDate: "",
		},
		{
			name: "github api, invalid created_at",
			require: &Require{
				RegexContent: `argus-[0-9.]+.linux-amd64`,
			},
			stdoutRegex: test.TrimYAML(`
				^DEBUG:.*
				WARNING: ignoring release date of "tomorrow" for version .*:
					parsing time .*
				$`,
			),
			errRegex: `^$`,
			body: []ghtypes.Asset{
				{Name: "argus-1.2.3.linux-amd64", CreatedAt: "tomorrow"},
			},
			wantReleaseDate: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			if tc.require != nil {
				tc.require.Status = &status.Status{}
			}
			v := "0.1.1-beta"

			// WHEN: RegexCheckContent is called on it.
			releaseDate, err := tc.require.RegexCheckContentGitHub(v, tc.body, logx.LogFrom{})

			prefix := fmt.Sprintf(
				"%s\nRequire.RegexCheckContentGitHub(version=%q, body=%q)",
				packageName, v, tc.body,
			)

			// THEN: the decode is what we expect.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the release date is what we expect.
			if releaseDate != tc.wantReleaseDate {
				t.Errorf(
					"%s releaseDate mismatch\ngot:  %q\nwant: %q",
					prefix, releaseDate, tc.wantReleaseDate,
				)
			}

			// AND: any log output is as expected.
			stdout := releaseStdout()
			// stdout finishes.
			if tc.stdoutRegex != "" {
				tc.stdoutRegex = strings.ReplaceAll(tc.stdoutRegex, "__name__", tc.name)
				if !util.RegexCheck(tc.stdoutRegex, stdout) {
					t.Errorf(
						"%s stdout mismatch\ngot:  %q\nwant: %q",
						prefix, stdout, tc.stdoutRegex,
					)
				}
			}
		})
	}
}
