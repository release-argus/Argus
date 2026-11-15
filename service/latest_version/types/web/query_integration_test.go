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

package web

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/util/polymorphic"
	"github.com/release-argus/Argus/web/metric"
)

func TestLookup_Query(t *testing.T) {
	testLookupVersions := testLookup(t, false)
	if _, err := testLookupVersions.query(logx.LogFrom{}); err != nil {
		t.Fatalf("Lookup.query() failed: %v", err)
	}

	type statusVars struct {
		latestVersion, latestVersionWant string
		deployedVersion                  string
	}
	type wantVars struct {
		announce    bool
		newVersion  bool
		stdoutRegex string
		errRegex    string
	}

	tests := []struct {
		name               string
		overrides          string
		semanticVersioning *bool
		hadStatus          statusVars
		want               wantVars
	}{
		{
			name:      "invalid url",
			overrides: `url: "invalid://	test"`,
			want: wantVars{
				announce: false,
				errRegex: `invalid control character in URL`,
			},
		},
		{
			name: "query that gets a non-semantic version",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLValid + `
				url_commands:
					- type: regex
					  regex: ver[0-9.]+
			`),
			want: wantVars{
				announce: false,
				errRegex: `failed to convert "[^"]+" to a semantic version`,
			},
		},
		{
			name: "query on self-signed https works when allowed",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLInvalid + `
				url_commands:
					- type: regex
					  regex: ver([0-9.]+)
				allow_invalid_certs: true
			`),
			want: wantVars{
				announce: true,
				errRegex: `^$`,
			},
		},
		{
			name: "query on self-signed https fails when not allowed",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain.URLInvalid + `
				url_commands:
					- type: regex
					  regex: ver([0-9.]+)
				allow_invalid_certs: false
			`),
			want: wantVars{
				announce: false,
				errRegex: `x509`,
			},
		},
		{
			name: "changed to semantic_versioning but had a non-semantic deployed_version",
			hadStatus: statusVars{
				deployedVersion: "1.2.3.4",
			},
			want: wantVars{
				announce: true,
				errRegex: `^$`,
			},
		},
		{
			name: "regex_content mismatch",
			overrides: test.TrimYAML(`
				require:
					regex_content: argus[0-9]+.exe
			`),
			want: wantVars{
				announce: false,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require fields
						regex "[^"]+" not matched on content for version "[^"]+"$`,
				),
			},
		},
		{
			name: "regex_content match",
			overrides: test.TrimYAML(`
				require:
					regex_content: ver{{ version }}
			`),
			want: wantVars{
				announce: true,
				errRegex: `^$`,
			},
		},
		{
			name: "command fail",
			overrides: test.TrimYAML(`
				require:
					command: [false]
			`),
			want: wantVars{
				announce: false,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require fields
						command failed:
							exit status 1$`,
				),
			},
		},
		{
			name: "command pass",
			overrides: test.TrimYAML(`
				require:
					command: [true]
			`),
			want: wantVars{
				announce: true,
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
						token: ` + test.GitHubToken(t) + `
			`),
			want: wantVars{
				announce: false,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require fields
						release-argus\/argus:[^ ]+ - .*tag not found.*$`,
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
			want: wantVars{
				announce: true,
				errRegex: `^$`,
			},
		},
		{
			name: "regex_version mismatch",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: .*
				require:
					regex_version: ver([0-9.]+)-unknown
			`),
			semanticVersioning: new(false),
			want: wantVars{
				announce: false,
				errRegex: test.TrimYAML(`
					^no releases were found matching the require fields
						regex "[^"]+" not matched on version ("[^"]+".*)+$`,
				),
			},
		},
		{
			name: "urlCommand regex mismatch",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ^[0-9]+$
			`),
			want: wantVars{
				announce: false,
				errRegex: test.TrimYAML(`
					^no releases were found matching the url_commands
						regex "[^"]+" didn't return any matches on ".*"$`,
				),
			},
		},
		{
			name: "valid semantic version query",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ver([0-9.]+)
			`),
			want: wantVars{
				announce: true,
				errRegex: `^$`,
			},
		},
		{
			name: "older version found",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ([0-9.]+)
			`),
			hadStatus: statusVars{
				latestVersion:   "0.0.0",
				deployedVersion: "9.9.9",
			},
			want: wantVars{
				announce:   false,
				newVersion: true,
				errRegex:   `^$`,
			},
		},
		{
			name: "newer version found",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ([0-9.]+)
			`),
			hadStatus: statusVars{
				deployedVersion: "0.0.0",
			},
			want: wantVars{
				announce: true,
				errRegex: `^$`,
			},
		},
		{
			name: "same version found",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ([0-9.]+)
			`),
			hadStatus: statusVars{
				deployedVersion: "1.2.1",
			},
			want: wantVars{
				announce: true,
				errRegex: `^$`,
			},
		},
		{
			name: "non-semantic version lookup",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ver[0-9.]+
			`),
			semanticVersioning: new(false),
			hadStatus: statusVars{
				latestVersionWant: "ver1.2.2",
			},
			want: wantVars{
				announce: true,
				errRegex: `^$`,
			},
		},
		{
			name: "url_command makes all versions non-semantic",
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '([0-9.]+-)'
			`),
			want: wantVars{
				announce: false,
				errRegex: `failed to convert "[^"]+" to a semantic version`,
			},
		},
		{
			name: "no overrides/first version does announce new version, with channel",
			hadStatus: statusVars{
				latestVersion: "",
			},
			want: wantVars{
				announce:    true,
				newVersion:  false,
				stdoutRegex: `Latest Release -`,
				errRegex:    `^$`,
			},
		},
		{
			name: "no overrides/new version does announce new version, with var",
			hadStatus: statusVars{
				latestVersion: "0.0.0",
			},
			want: wantVars{
				announce:    false,
				newVersion:  true,
				stdoutRegex: `New Release -`,
				errRegex:    `^$`,
			},
		},
		{
			name: "no overrides/same version does not announce new version",
			hadStatus: statusVars{
				latestVersion: testLookupVersions.Status.LatestVersion(),
			},
			want: wantVars{
				announce:   true, // LastQueried announce.
				newVersion: false,
				errRegex:   `^$`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			serviceID := tc.name
			const maxAttempts = 3
			retryOnTimeout := true
			for attempt := 0; retryOnTimeout; attempt++ {
				t.Logf("attempt %d", attempt)
				releaseStdout := test.CaptureLog(t, logx.Default())
				retryOnTimeout = false
				lookup := testLookup(t, false)
				lookup.InitMetrics(lookup)
				lookup.URL = strings.Replace(lookup.URL, "://invalid", "://valid", 1)
				lookup.AllowInvalidCerts = nil
				lookup.Status.ServiceInfo.ID = serviceID
				lookup.Options.SemanticVersioning = tc.semanticVersioning
				lookup.Status.SetLatestVersion(tc.hadStatus.latestVersion, "", false)
				lookup.Status.SetDeployedVersion(tc.hadStatus.deployedVersion, "", false)
				if err := lookup.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
					t.Fatalf("Lookup.ApplyOverrides() failed: %v", err)
				}
				reqOverride, err := polymorphic.Extract("yaml", []byte(tc.overrides), "require")
				if err != nil {
					t.Fatalf("polymorphic.Extract() failed: %v", err)
				}
				if reqOverride != nil {
					if lookup.Require, err = filter.Decode(
						"yaml", reqOverride,
						lookup.Status,
						&lookup.Defaults.Require,
					); err != nil {
						t.Fatalf("filter.Decode() failed: %v", err)
					}
				}
				lookup.Init(
					lookup.Options,
					lookup.Status,
					base.DefaultsConfig{
						Soft: lookup.Defaults,
						Hard: lookup.HardDefaults,
					},
				)

				newVersion, err := lookup.Query(true, logx.LogFrom{})

				stdout := releaseStdout()
				e := errfmt.FormatError(err)
				if !util.RegexCheck(tc.want.errRegex, e) {
					if strings.Contains(e, "context deadline exceeded") && attempt < maxAttempts-1 {
						retryOnTimeout = true
						lookup.DeleteMetrics(lookup)
						time.Sleep(time.Second)
						continue
					}
					t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.want.errRegex)
				}

				if !util.RegexCheck(tc.want.stdoutRegex, stdout) {
					t.Fatalf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.want.stdoutRegex)
				}

				if got := lookup.Status.LatestVersion(); tc.hadStatus.latestVersionWant != "" &&
					tc.hadStatus.latestVersionWant != got {
					t.Fatalf(
						"Status.LatestVersion() mismatch\ngot:  %q\nwant: %q",
						got, tc.hadStatus.latestVersionWant,
					)
				}
				wantAnnounces := 0
				if tc.want.announce {
					wantAnnounces = 1
				}
				if got := len(lookup.Status.AnnounceChannel); got != wantAnnounces {
					t.Fatalf(
						"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
						got, wantAnnounces,
					)
				}
				if newVersion != tc.want.newVersion {
					t.Fatalf(
						"newVersion mismatch\ngot:  %t\nwant: %t",
						newVersion, tc.want.newVersion,
					)
				}

				want := 0.0
				if tc.want.errRegex != `^$` {
					want = 1
				}
				got := testutil.ToFloat64(
					metric.LatestVersionQueryResultTotal.WithLabelValues(
						serviceID,
						lookup.GetType(),
						metric.ActionResultFail,
					),
				)
				if got != want {
					t.Fatalf(
						"fail: metric.LatestVersionQueryResultTotal mismatch\ngot:  %f\nwant: %f",
						got, want,
					)
				}
				want = 0
				if tc.want.errRegex == `^$` {
					want = 1
				}
				got = testutil.ToFloat64(
					metric.LatestVersionQueryResultTotal.WithLabelValues(
						serviceID,
						lookup.GetType(),
						metric.ActionResultSuccess,
					),
				)
				if got != want {
					t.Fatalf(
						"success: metric.LatestVersionQueryResultTotal mismatch\n"+
							"got:  %f\nwant: %f",
						got, want,
					)
				}
				lookup.DeleteMetrics(lookup)
			}
		})
	}
}

func TestLookup_GetVersion__integration__docker(t *testing.T) {
	body := `
		version 1 is "v0.0.0"
		version 2 is "ver1.2.3-dev"
		version 3 is "ver1.2.4"
		version 4 is "ver1.2.5"
	`

	type wantVars struct {
		version  string
		errRegex string
	}

	tests := []struct {
		name            string
		lookupOverrides string
		want            wantVars
	}{
		{
			name: "docker fail",
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '4 is "ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
					regex_content: 'version 4 is "ver{{ version }}"'
					command: [true]
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: '{{ version }}-unknown'
						token: ` + test.GitHubToken(t) + `
			`),
			want: wantVars{
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
						` + test.ArgusDockerGHCRRepo + `:[^ ]+ - .*tag not found.*$`,
				),
			},
		},
		{
			name: "docker pass",
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
					regex_content: 'version 4 is "ver{{ version }}"'
					command: [true]
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: '0.39.0'
						token: ` + test.GitHubToken(t) + `
			`),
			want: wantVars{
				version:  "1.2.5",
				errRegex: `^$`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			if err := lookup.ApplyOverrides("yaml", []byte(tc.lookupOverrides)); err != nil {
				t.Fatalf("Lookup.ApplyOverrides() failed: %v", err)
			}
			lookup.Options.SemanticVersioning = new(true)
			if err := lookup.CheckValues(); err != nil {
				t.Fatalf("Lookup.CheckValues() failed: %v", err)
			}

			version, err := lookup.getVersion(body, logx.LogFrom{})
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.want.errRegex)
			}

			if version != tc.want.version {
				t.Errorf("version mismatch\ngot:  %q\nwant: %q", version, tc.want.version)
			}
		})
	}
}
