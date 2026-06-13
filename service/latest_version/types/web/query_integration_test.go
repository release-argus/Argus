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
	"fmt"
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
	_, _ = testLookupVersions.query(logx.LogFrom{})

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

	// GIVEN: a Lookup.
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
				errRegex: `invalid control character in URL`,
			},
		},
		{
			name: "query that gets a non-semantic version",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				url_commands:
					- type: regex
					  regex: ver[0-9.]+
			`),
			want: wantVars{
				errRegex: `failed to convert "[^"]+" to a semantic version`,
			},
		},
		{
			name: "query on self-signed https works when allowed",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_invalid"] + `
				url_commands:
					- type: regex
					  regex: ver([0-9.]+)
				allow_invalid_certs: true
			`),
			want: wantVars{
				errRegex: `^$`,
			},
		},
		{
			name: "query on self-signed https fails when not allowed",
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_invalid"] + `
				url_commands:
					- type: regex
					  regex: ver([0-9.]+)
				allow_invalid_certs: false
			`),
			want: wantVars{
				errRegex: `x509`,
			},
		},
		{
			name: "changed to semantic_versioning but had a non-semantic deployed_version",
			hadStatus: statusVars{
				deployedVersion: "1.2.3.4",
			},
			want: wantVars{
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
			want: wantVars{
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
						token: ` + test.DockerHubToken(t) + `
			`),
			want: wantVars{
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
			semanticVersioning: test.Ptr(false),
			want: wantVars{
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
			semanticVersioning: test.Ptr(false),
			hadStatus: statusVars{
				latestVersionWant: "ver1.2.2",
			},
			want: wantVars{
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
				errRegex: `failed to convert "[^"]+" to a semantic version`,
			},
		},
		{
			name: "no overrides, first version does announce new version (with channel)",
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
			name: "no overrides, new version does announce new version (with var)",
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
			name: "no overrides, same version does not announce new version",
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
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			serviceID := tc.name
			try := 0
			temporaryFailureInNameResolution := true
			for temporaryFailureInNameResolution != false {
				releaseStdout := test.CaptureLog(t, logx.Default())
				try++
				temporaryFailureInNameResolution = false
				lookup := testLookup(t, false)
				lookup.InitMetrics(lookup)
				// Valid cert switch.
				lookup.URL = strings.Replace(lookup.URL, "://invalid", "://valid", 1)
				lookup.AllowInvalidCerts = nil
				lookup.Status.ServiceInfo.ID = serviceID
				lookup.Options.SemanticVersioning = tc.semanticVersioning
				// hadStatus.
				lookup.Status.SetLatestVersion(tc.hadStatus.latestVersion, "", false)
				lookup.Status.SetDeployedVersion(tc.hadStatus.deployedVersion, "", false)
				// overrides.
				if err := lookup.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
					t.Fatalf(
						"%s\nfailed to unmarshal overrides: %v",
						packageName, err,
					)
				}
				if regOverride, _ := polymorphic.Extract("yaml", []byte(tc.overrides), "require"); regOverride != nil {
					lookup.Require, _ = filter.Decode(
						"yaml", regOverride,
						lookup.Status,
						&lookup.Defaults.Require,
					)
				}
				lookup.Init(
					lookup.Options,
					lookup.Status,
					base.DefaultsConfig{
						Soft: lookup.Defaults,
						Hard: lookup.HardDefaults,
					},
				)

				// WHEN: Query is called on it.
				var newVersion bool
				newVersion, err := lookup.Query(true, logx.LogFrom{})

				prefix := fmt.Sprintf("%s\nLookup.Query()", packageName)

				// THEN: the error is as expected.
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
						prefix, e, tc.want.errRegex,
					)
				}

				// AND: the stdout contains the expected strings.
				if !util.RegexCheck(tc.want.stdoutRegex, stdout) {
					t.Fatalf(
						"%s stdout mismatch\ngot:  %q\nwant: %q",
						prefix, stdout, tc.want.stdoutRegex,
					)
				}

				// AND: the LatestVersion is as expected.
				if got := lookup.Status.LatestVersion(); tc.hadStatus.latestVersionWant != "" &&
					tc.hadStatus.latestVersionWant != got {
					t.Fatalf(
						"%s .LatestVersion() mismatch\ngot:  %q\nwant:   %q",
						prefix, got, tc.hadStatus.latestVersionWant,
					)
				}
				if got := len(lookup.Status.AnnounceChannel); tc.want.announce && got != 1 {
					t.Fatalf(
						"%s announcement mismatch\ngot:  %d\nwant: %d",
						prefix, got, 1,
					)
				}
				if newVersion != tc.want.newVersion {
					t.Fatalf(
						"%s newVersion mismatch\ngot:  %t\nwant: %t",
						prefix, newVersion, tc.want.newVersion,
					)
				}

				// AND: the metrics are as expected.
				// 	FAIL:
				var want float64 = 0
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
						"%s LatestVersionQueryResultTotal metric - FAIL\ngot:  %f\nwant: %f",
						prefix, got, want,
					)
				}
				// 	SUCCESS:
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
						"%s LatestVersionQueryResultTotal metric - SUCCESS\ngot:  %f\nwant: %f",
						prefix, got, want,
					)
				}
				lookup.DeleteMetrics(lookup)
			}
		})
	}
}
