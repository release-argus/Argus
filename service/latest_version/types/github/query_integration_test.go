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

//go:build integration

package github

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_Query(t *testing.T) {
	tLookup := testLookup(t, false)
	tLookup.URL = "release-argus/.github"
	_, _ = tLookup.Query(false, logx.LogFrom{})
	emptyReleasesETag := tLookup.data.eTag

	type statusVars struct {
		deployedVersion   string
		latestVersion     string
		wantLatestVersion *string
	}
	type want struct {
		stdoutRegex string
		errRegex    string
	}
	// GIVEN: a Lookup.
	tests := []struct {
		name         string
		overrides    string
		overrideETag *string
		semVer       bool
		status       statusVars
		want         want
	}{
		{
			name:      "invalid url",
			overrides: `url: "release-argus	Argus"`,
			semVer:    true,
			want: want{
				errRegex: `invalid control character in URL`,
			},
		},
		{
			name: "query that gets a non-semantic version - not allowed",
			overrides: test.TrimYAML(`
				url: release-argus/test
				url_commands:
					- type: regex
						regex: 'ver[0-9.]+'
			`),
			semVer: true,
			want: want{
				errRegex: `no releases were found matching the url_commands`,
			},
		},
		{
			name: "query that gets a non-semantic version - allowed",
			overrides: test.TrimYAML(`
				url: release-argus/test
				url_commands:
					- type: regex
						regex: 'ver[0-9.]+'
			`),
			semVer: false,
			status: statusVars{
				wantLatestVersion: test.Ptr("ver1.1.1"),
			},
			want: want{
				errRegex: `^$`,
			},
		},
		{
			name: "changed to semantic_versioning but had a non-semantic deployed_version",
			status: statusVars{
				deployedVersion: "1.2.3.4",
			},
			semVer: true,
			want: want{
				errRegex: `^$`,
			},
		},
		{
			name: "regex_content mismatch",
			overrides: test.TrimYAML(`
				require:
					regex_content: "argus[0-9]+.exe"
			`),
			semVer: true,
			want: want{
				stdoutRegex: `regex "[^"]+" not matched on content for version`,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
						regex "[^"]+" not matched on content for version "[^"]+"$`,
				),
			},
		},
		{
			name: "regex_content match",
			overrides: test.TrimYAML(`
				require:
					regex_content: "{{ version }}.linux-amd64"
			`),
			semVer: false,
			want: want{
				errRegex: `^$`,
			},
		},
		{
			name: "command fail",
			overrides: test.TrimYAML(`
				require:
					command: ["false"]
			`),
			semVer: true,
			want: want{
				stdoutRegex: `exit status 1`,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
						command failed:
							exit status 1$`,
				),
			},
		},
		{
			name: "command pass",
			overrides: test.TrimYAML(`
				require:
					command: ["true"]
			`),
			semVer: true,
			want: want{
				errRegex: `^$`,
			},
		},
		{
			name: "docker tag mismatch",
			overrides: test.TrimYAML(`
				require:
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: 0.9.0-beta
						token: ` + test.DockerHubToken(t) + `
			`),
			semVer: true,
			want: want{
				stdoutRegex: `tag not found`,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
						` + test.ArgusDockerGHCRRepo + `:0.9.0-beta - tag not found`,
				),
			},
		},
		{
			name: "docker tag match",
			overrides: test.TrimYAML(`
				require:
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: 0.9.0
						token: ` + test.GitHubToken(t) + `
			`),
			semVer: true,
			want: want{
				errRegex: `^$`,
			},
		},
		{
			name: "regex_version mismatch",
			overrides: test.TrimYAML(`
				require:
					regex_version: "ver([0-9.]+)"
			`),
			semVer: true,
			want: want{
				stdoutRegex: `regex "[^"]+" not matched on version`,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
						regex "[^"]+" not matched on version "[^"]+"$`,
				),
			},
		},
		{
			name: "urlCommand regex mismatch",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: 'x([0-9.]+)'
			`),
			semVer: true,
			want: want{
				errRegex:    `no releases were found matching the url_commands`,
				stdoutRegex: `regex "[^"]+" didn't return any matches`,
			},
		},
		{
			name: "valid semantic version query",
			overrides: test.TrimYAML(`
				url: release-argus/test
				url_commands:
					- type: regex
						regex: 'ver([0-9.]+)'
			`),
			semVer: true,
			want: want{
				errRegex: `^$`,
			},
		},
		{
			name: "older version found",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '([0-9.]+)'
			`),
			status: statusVars{
				latestVersion:   "0.0.0",
				deployedVersion: "9.9.9",
			},
			semVer: true,
			want: want{
				errRegex: `^$`,
			},
		},
		{
			name: "newer version found",
			overrides: test.TrimYAML(`
				url: release-argus/test
				url_commands:
					- type: regex
						regex: '([0-9.]+)'
			`),
			status: statusVars{
				latestVersion:     "1.1.0",
				wantLatestVersion: test.Ptr("1.1.1"),
			},
			semVer: true,
			want: want{
				stdoutRegex: `New Release - "[^"]+"`,
				errRegex:    `^$`,
			},
		},
		{
			name: "same version found",
			overrides: test.TrimYAML(`
				url: release-argus/test
				url_commands:
					- type: regex
						regex: '([0-9.]+)'
			`),
			status: statusVars{
				latestVersion:     "1.1.1",
				wantLatestVersion: test.Ptr("1.1.1"),
			},
			semVer: true,
			want: want{
				errRegex: `^$`,
			},
		},
		{
			name:      "repo that uses tags, not releases - has tags",
			overrides: `url: "release-argus/test"`,
			status: statusVars{
				wantLatestVersion: test.Ptr("1.1.1"),
			},
			semVer: true,
			want: want{
				errRegex:    `^$`,
				stdoutRegex: `no tags found on /releases, trying /tags`,
			},
		},
		{
			name:         "repo that uses tags, not releases - no tags",
			overrides:    `url: "release-argus/.github"`,
			overrideETag: &emptyReleasesETag,
			semVer:       true,
			want: want{
				errRegex:    `no releases were found matching the url_commands`,
				stdoutRegex: `no tags found on /releases, trying /tags`,
			},
		},
		{
			name:         "repo that uses tags, not releases - no tags - emptyListETag changed",
			overrides:    `url: "release-argus/.github"`,
			overrideETag: test.Ptr(""),
			semVer:       true,
			want: want{
				errRegex:    `no releases were found`,
				stdoutRegex: `/releases gave .*, trying /tags`,
			},
		},
		{
			name:         "version from 2nd page",
			overrides:    `url: "release-argus/test-pagination"`,
			overrideETag: test.Ptr(""),
			semVer:       true,
			want: want{
				stdoutRegex: `on page 1`,
			},
		},
		{
			name:      "no access token",
			overrides: `access_token: null`,
			semVer:    true,
			want: want{
				errRegex: `^$`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			try := 0
			temporaryFailureInNameResolution := true
			for temporaryFailureInNameResolution != false {
				releaseStdout := test.CaptureLog(t, logx.Default())
				try++
				temporaryFailureInNameResolution = false
				lookup := testLookup(t, false)
				if strings.Contains(tc.overrides, "access_token: null") {
					lookup.HardDefaults.AccessToken = ""
				}
				lookup.Status.ServiceInfo.ID = tc.name
				if err := lookup.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
					t.Fatalf(
						"%s\nfailed to unmarshal Lookup overrides: %v",
						packageName, err,
					)
				}
				lookup.Options.SemanticVersioning = &tc.semVer
				lookup.Status.SetLatestVersion(tc.status.latestVersion, "", false)
				lookup.Status.SetDeployedVersion(tc.status.deployedVersion, "", false)
				// In case require is non-nil, Init to give it Status.
				lookup.Init(
					lookup.Options,
					lookup.Status,
					base.DefaultsConfig{
						Soft: lookup.Defaults,
						Hard: lookup.HardDefaults,
					},
				)
				// Clear ETag/Releases if URL changed.
				if tc.overrideETag != nil {
					lookup.data.eTag = *tc.overrideETag
				}

				// WHEN: Query is called on it.
				_, err := lookup.Query(true, logx.LogFrom{})

				prefix := fmt.Sprintf("%s\nLookup.Query()", packageName)

				// THEN: any decode is expected.
				stdout := releaseStdout()
				e := errfmt.FormatError(err)
				if !util.RegexCheck(tc.want.errRegex, e) {
					if strings.Contains(e, "context deadline exceeded") {
						temporaryFailureInNameResolution = true
						if try != 3 {
							time.Sleep(time.Second)
							continue
						}
					}
					t.Fatalf(
						"%s error mismatch\ngot:  %q\nwant: %q",
						prefix, tc.want.errRegex, e,
					)
				}

				// AND: the stdout contains the expected strings.
				if !util.RegexCheck(tc.want.stdoutRegex, stdout) {
					t.Fatalf(
						"%s stdout mismatch\n%q\nnot found in:\n%q",
						prefix, tc.want.stdoutRegex, stdout,
					)
				}

				// AND: the LatestVersion is as expected.
				if tc.status.wantLatestVersion != nil &&
					*tc.status.wantLatestVersion != lookup.Status.LatestVersion() {
					t.Fatalf(
						"%s .LatestVersion() mismatch\ngot:  %q\nwant: %q",
						prefix, *tc.status.wantLatestVersion, lookup.Status.LatestVersion(),
					)
				}
			}
		})
	}
}

func TestLookup_Query__githubETag(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name                               string
		attempts                           int
		eTagChanged, eTagUnchangedUseCache int
		initialRequireRegexVersion         string
		urlCommands                        filter.URLCommands
		errRegex                           string
	}{
		// Keeps `.Releases` in case filters change.
		{
			name:                  "three requests only uses 1 api limit",
			attempts:              3,
			eTagChanged:           1,
			eTagUnchangedUseCache: 3, // 2 attempts + 1 recheck.
			errRegex:              `^$`,
		},
		{
			name:                       "if initial request fails filter, cached results will be used",
			attempts:                   3,
			eTagChanged:                3, // page1+2, page1.
			eTagUnchangedUseCache:      2, // 1 last attempt + 1 recheck.
			initialRequireRegexVersion: `^FOO$`,
			errRegex: test.TrimYAML(`
				^no releases were found matching the require field.*
					regex "[^"]+" not matched on version "[^"]+"$`,
			),
		},
		{
			name:                  "invalid url_commands will catch no versions",
			attempts:              2,
			eTagChanged:           4, // page1+2, page1+2.
			eTagUnchangedUseCache: 0, // 0 recheck.
			urlCommands: filter.URLCommands{
				{Type: "regex", Regex: `^FOO$`},
			},
			errRegex: test.TrimYAML(`
				^no releases were found matching the url_commands on page 2 of the API response
				no releases were found matching the url_commands on page 2 of the API response$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			lookup := testLookup(t, false)
			lookup.URL = "release-argus/test-pagination"
			lookup.UsePreRelease = test.Ptr(true)
			lookup.GetGitHubData().SetETag("foo")
			lookup.Status.ServiceInfo.ID = tc.name
			lookup.Require = &filter.Require{
				RegexVersion: tc.initialRequireRegexVersion,
				Status:       lookup.Status,
			}
			lookup.URLCommands = tc.urlCommands

			attempt := 0
			prefix := fmt.Sprintf("%s\nLookup.Query()", packageName)

			// WHEN: Query is called on it 'attempts' number of times.
			var errs []error
			for tc.attempts != attempt {
				attempt++
				if attempt == 2 {
					lookup.Require = &filter.Require{}
				}

				_, err := lookup.Query(true, logx.LogFrom{})
				if err != nil {
					errs = append(errs, err)
				}
				t.Logf(
					"%s - attempt %d, ETag: %s",
					prefix, attempt, lookup.GetGitHubData().ETag(),
				)
			}

			// THEN: any decode is expected.
			stdout := releaseStdout()
			e := errfmt.FormatError(errors.Join(errs...))
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			gotETagChanged := strings.Count(stdout, "new ETag")
			if gotETagChanged != tc.eTagChanged {
				t.Errorf(
					"%s unexpected ETag produced\ngot:  %d\nwant: %d\nstdout: %q",
					prefix, gotETagChanged, tc.eTagChanged, stdout,
				)
			}
			gotETagUnchangedUseCache := strings.Count(stdout, "Using cached releases")
			if gotETagUnchangedUseCache != tc.eTagUnchangedUseCache {
				t.Errorf(
					"%s ETag unchanged use cache count mismatch\ngot:  %d\nwant: %d\nstdout: %q",
					prefix, gotETagUnchangedUseCache, tc.eTagUnchangedUseCache, stdout,
				)
			}
		})
	}
}
