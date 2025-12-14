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

// Package web provides a web-based lookup type.
package web

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/web/metric"
)

func TestHTTPRequest(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		failing  bool
		url      string
		errRegex string
	}{
		"invalid url": {
			url:      "invalid://	test",
			errRegex: `invalid control character in URL`},
		"unknown url": {
			url:      "https://release-argus.invalid-tld",
			errRegex: `no such host`},
		"valid url": {
			url:      "https://release-argus.io",
			errRegex: `^$`},
		"invalid cert": {
			failing:  true,
			errRegex: `x509`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(tc.failing)
			if tc.url != "" {
				lookup.URL = tc.url
			}

			// WHEN httpRequest is called on it.
			_, err := lookup.httpRequest(logutil.LogFrom{})

			// THEN any error is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	// GIVEN a Lookup and a Body to filter.
	body := `
		version 1 is "v0.0.0"
		version 2 is "ver1.2.3-dev"
		version 3 is "ver1.2.4"
		version 4 is "ver1.2.5"
	`
	urlCommand := filter.URLCommand{
		Type: "regex", Regex: `([0-9]\.[0-9.]+)`}

	type wantVars struct {
		version  string
		errRegex string
	}

	tests := map[string]struct {
		bodyOverride    *string
		lookupOverrides string
		noSemVer        bool
		want            wantVars
	}{
		"nil url_commands": {
			lookupOverrides: test.TrimYAML(`
				url_commands: null
			`),
			want: wantVars{
				version:  body,
				errRegex: `^$`},
		},
		"empty body": {
			bodyOverride: test.StringPtr(""),
			lookupOverrides: test.TrimYAML(`
				url_commands: []
			`),
			want: wantVars{
				errRegex: `^no releases were found matching the url_commands$`},
		},
		"nil Require": {
			noSemVer: true,
			want: wantVars{
				version:  "0.0.0",
				errRegex: `^$`},
		},
		"url_commands that error": {
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ver9
						index: 9
			`),
			want: wantVars{
				errRegex: `regex "[^"]+" didn't return any matches`},
		},
		"url_commands that pass": {
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"ver([0-9][^"]+)"'
			`),
			noSemVer: true,
			want: wantVars{
				version:  "1.2.3-dev",
				errRegex: `^$`},
		},
		"regex_version mismatch": {
			lookupOverrides: test.TrimYAML(`
				require:
					regex_version: ver(2\.[0-9.]+)
			`),
			want: wantVars{
				errRegex: test.TrimYAML(`
				^no releases were found matching the require field.*
				regex "[^"]+" not matched on version "[^"]+"$`)},
		},
		"regex_version match": {
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
			`),
			noSemVer: true,
			want: wantVars{
				version:  "1.2.4",
				errRegex: `^$`},
		},
		"regex_content mismatch": {
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '4 is "ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
					regex_content: ver[0-9]+.exe
			`),
			want: wantVars{
				errRegex: test.TrimYAML(`
				^no releases were found matching the require field.*
				regex "[^"]+" not matched on content for version "[^"]+"$`)},
		},
		"regex_content match": {
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
					regex_content: 'version 4 is "ver{{ version }}"'
			`),
			want: wantVars{
				version:  "1.2.5",
				errRegex: `^$`},
		},
		"command fail": {
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '4 is "ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
					regex_content: 'version 4 is "ver{{ version }}"'
					command: [false]
			`),
			want: wantVars{
				errRegex: test.TrimYAML(`
				^no releases were found matching the require field.*
				command failed: .*$`)},
		},
		"command pass": {
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
					regex_content: 'version 4 is "ver{{ version }}"'
					command: [true]
			`),
			want: wantVars{
				version:  "1.2.5",
				errRegex: `^$`},
		},
		"docker fail": {
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
						image: release-argus/argus
						tag: '{{ version }}-unknown'
						token: ` + os.Getenv("GH_TOKEN") + `
			`),
			want: wantVars{
				errRegex: test.TrimYAML(`
				^no releases were found matching the require field.*
				release-argus\/argus:[^ ]+ - .*manifest unknown.*$`)},
		},
		"docker pass": {
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
						image: release-argus/argus
						tag: '0.8.0'
						token: ` + os.Getenv("GH_TOKEN") + `
			`),
			want: wantVars{
				version:  "1.2.5",
				errRegex: `^$`},
		},
		"sorts versions when semantic_versioning enabled": {
			bodyOverride: test.StringPtr(`
				patch for older major "0.4.7"
				patch for latest major "v1.0.1"
				latest major "v1.0.0"
				older major "0.0.0"
			`),
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"(v?[0-9][^"]+)"'
			`),
			want: wantVars{
				version:  "v1.0.1",
				errRegex: `^$`},
		},
		"does not sort versions when semantic_versioning disabled": {
			bodyOverride: test.StringPtr(`
				patch for older major "0.4.7"
				patch for latest major "v1.0.1"
				latest major "v1.0.0"
				older major "0.0.0"
			`),
			lookupOverrides: test.TrimYAML(`
				semantic_versioning: false
				url_commands:
					- type: regex
						regex: '"(v?[0-9][^"]+)"'
			`),
			noSemVer: true,
			want: wantVars{
				version:  "0.4.7",
				errRegex: `^$`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.URLCommands[0] = urlCommand
			err := yaml.Unmarshal([]byte(tc.lookupOverrides), lookup)
			if err != nil {
				t.Fatalf("%s\nfailed to unmarshal lookupOverrides: %v",
					packageName, err)
			}
			lookup.Init(
				lookup.Options,
				lookup.Status,
				lookup.Defaults, lookup.HardDefaults)
			if tc.noSemVer {
				*lookup.Options.SemanticVersioning = false
			}
			testBody := util.DereferenceOrValue(tc.bodyOverride, body)

			// WHEN getVersion is called on it.
			version, err := lookup.getVersion(
				testBody, logutil.LogFrom{})

			// THEN any error is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want.errRegex, e)
			}
			// AND the version is as expected.
			if version != tc.want.version {
				t.Errorf("%s\nversion mismatch:\nwant: %q\ngot:  %q",
					packageName, tc.want.version, version)
			}
		})
	}
}

func TestQuery(t *testing.T) {
	testLookupVersions := testLookup(false)
	_, _ = testLookupVersions.query(logutil.LogFrom{})

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

	// GIVEN a Lookup.
	tests := map[string]struct {
		overrides          string
		semanticVersioning *bool
		hadStatus          statusVars
		want               wantVars
	}{
		"invalid url": {
			overrides: test.TrimYAML(`
				url: "invalid://	test"
			`),
			want: wantVars{
				errRegex: `invalid control character in URL`},
		},
		"query that gets a non-semantic version": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_valid"] + `
				url_commands:
					- type: regex
					  regex: ver[0-9.]+
			`),
			want: wantVars{
				errRegex: `failed to convert "[^"]+" to a semantic version`},
		},
		"query on self-signed https works when allowed": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_invalid"] + `
				url_commands:
					- type: regex
					  regex: ver([0-9.]+)
				allow_invalid_certs: true
			`),
			want: wantVars{
				errRegex: `^$`},
		},
		"query on self-signed https fails when not allowed": {
			overrides: test.TrimYAML(`
				url: ` + test.LookupPlain["url_invalid"] + `
				url_commands:
					- type: regex
					  regex: ver([0-9.]+)
				allow_invalid_certs: false
			`),
			want: wantVars{
				errRegex: `x509`},
		},
		"changed to semantic_versioning but had a non-semantic deployed_version": {
			hadStatus: statusVars{
				deployedVersion: "1.2.3.4"},
			want: wantVars{
				errRegex: `^$`},
		},
		"regex_content mismatch": {
			overrides: test.TrimYAML(`
				require:
					regex_content: argus[0-9]+.exe
			`),
			want: wantVars{
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
					regex "[^"]+" not matched on content for version "[^"]+"$`)},
		},
		"regex_content match": {
			overrides: test.TrimYAML(`
				require:
					regex_content: ver{{ version }}
			`),
			want: wantVars{
				errRegex: `^$`},
		},
		"command fail": {
			overrides: test.TrimYAML(`
				require:
					command: [false]
			`),
			want: wantVars{
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
					command failed: .*$`)},
		},
		"command pass": {
			overrides: test.TrimYAML(`
				require:
					command: [true]
			`),
			want: wantVars{
				errRegex: `^$`},
		},
		"docker tag mismatch": {
			overrides: test.TrimYAML(`
				require:
					docker:
						type: ghcr
						image: release-argus/argus
						tag: 0.9.0-beta
						token: ` + os.Getenv("GH_TOKEN") + `
			`),
			want: wantVars{
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
					release-argus\/argus:[^ ]+ - .*manifest unknown.*$`)},
		},
		"docker tag match": {
			overrides: test.TrimYAML(`
				require:
					docker:
						type: ghcr
						image: release-argus/argus
						tag: 0.9.0
						token: ` + os.Getenv("GH_TOKEN") + `
			`),
			want: wantVars{
				errRegex: `^$`},
		},
		"regex_version mismatch": {
			overrides: test.TrimYAML(`
				require:
					regex_version: ver([0-9.]+)
			`),
			want: wantVars{
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
					regex "[^"]+" not matched on version "[^"]+"$`)},
		},
		"urlCommand regex mismatch": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ^[0-9]+$
			`),
			want: wantVars{
				errRegex: `^no releases were found matching the url_commands\nregex "[^"]+" didn't return any matches on ".*"$`},
		},
		"valid semantic version query": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ver([0-9.]+)
			`),
			want: wantVars{
				errRegex: `^$`},
		},
		"older version found": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ([0-9.]+)
			`),
			hadStatus: statusVars{
				latestVersion:   "0.0.0",
				deployedVersion: "9.9.9"},
			want: wantVars{
				errRegex: `queried version .* is less than the deployed version`},
		},
		"newer version found": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ([0-9.]+)
			`),
			hadStatus: statusVars{
				deployedVersion: "0.0.0"},
			want: wantVars{
				errRegex: `^$`},
		},
		"same version found": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ([0-9.]+)
			`),
			hadStatus: statusVars{
				deployedVersion: "1.2.1"},
			want: wantVars{
				errRegex: `^$`},
		},
		"non-semantic version lookup": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ver[0-9.]+
			`),
			semanticVersioning: test.BoolPtr(false),
			hadStatus: statusVars{
				latestVersionWant: "ver1.2.2"},
			want: wantVars{
				errRegex: `^$`},
		},
		"url_command makes all versions non-semantic": {
			overrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '([0-9.]+\.)'
			`),
			want: wantVars{
				errRegex: `failed to convert "[^"]+" to a semantic version`},
		},
		"no overrides, first version does announce new version (with channel)": {
			hadStatus: statusVars{
				latestVersion: ""},
			want: wantVars{
				announce:    true,
				newVersion:  false,
				stdoutRegex: `Latest Release -`,
				errRegex:    `^$`},
		},
		"no overrides, new version does announce new version (with var)": {
			hadStatus: statusVars{
				latestVersion: "0.0.0"},
			want: wantVars{
				announce:    false,
				newVersion:  true,
				stdoutRegex: `New Release -`,
				errRegex:    `^$`},
		},
		"no overrides, same version does not announce new version": {
			hadStatus: statusVars{
				latestVersion: testLookupVersions.Status.LatestVersion()},
			want: wantVars{
				announce:   true, // LastQueried announce.
				newVersion: false,
				errRegex:   `^$`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.

			serviceID := name
			try := 0
			temporaryFailureInNameResolution := true
			for temporaryFailureInNameResolution != false {
				releaseStdout := test.CaptureStdout()
				try++
				temporaryFailureInNameResolution = false
				lookup := testLookup(false)
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
				err := yaml.Unmarshal([]byte(tc.overrides), lookup)
				if err != nil {
					t.Fatalf("%s\nfailed to unmarshal overrides: %v",
						packageName, err)
				}
				lookup.Init(
					lookup.Options,
					lookup.Status,
					lookup.Defaults, lookup.HardDefaults)

				// WHEN Query is called on it.
				var newVersion bool
				newVersion, err = lookup.Query(true, logutil.LogFrom{})

				// THEN any error is expected.
				stdout := releaseStdout()
				e := util.ErrorToString(err)
				if !util.RegexCheck(tc.want.errRegex, e) {
					if strings.Contains(e, "context deadline exceeded") {
						temporaryFailureInNameResolution = true
						if try != 3 {
							time.Sleep(time.Second)
							continue
						}
					}
					t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.want.errRegex, e)
				}
				// AND the stdout contains the expected strings.
				if !util.RegexCheck(tc.want.stdoutRegex, stdout) {
					t.Fatalf("%s\nstdout mismatch\nwant: %q\ngot:  %q",
						packageName, tc.want.stdoutRegex, stdout)
				}
				// AND the LatestVersion is as expected.
				if tc.hadStatus.latestVersionWant != "" &&
					tc.hadStatus.latestVersionWant != lookup.Status.LatestVersion() {
					t.Fatalf("%s\nLatestVersion mismatch\nwant: %q\ngot:  %q",
						packageName, tc.hadStatus.latestVersionWant, lookup.Status.LatestVersion())
				}
				if want := 1; tc.want.announce && len(lookup.Status.AnnounceChannel) != want {
					t.Fatalf("%s\nannouncement mismatch\nwant: %d\ngot:  %d",
						packageName, want, len(lookup.Status.AnnounceChannel))
				}
				if newVersion != tc.want.newVersion {
					t.Fatalf("%s\nnewVersion mismatch\nwant: %t\ngot:  %t",
						packageName, tc.want.newVersion, newVersion)
				}
				// AND the metrics are as expected.
				// 	FAIL:
				var want float64 = 0
				if tc.want.errRegex != `^$` {
					want = 1
				}
				got := testutil.ToFloat64(metric.LatestVersionQueryResultTotal.WithLabelValues(
					serviceID,
					lookup.Type,
					"FAIL"))
				if got != want {
					t.Fatalf("%s\nLatestVersionQueryResultTotal - FAIL\nwant: %f\ngot:  %f",
						packageName, want, got)
				}
				// 	SUCCESS:
				want = 0
				if tc.want.errRegex == `^$` {
					want = 1
				}
				got = testutil.ToFloat64(metric.LatestVersionQueryResultTotal.WithLabelValues(
					serviceID,
					lookup.Type,
					"SUCCESS"))
				if got != want {
					t.Fatalf("%s\nLatestVersionQueryResultTotal - SUCCESS\nwant: %f\ngot:  %f",
						packageName, want, got)
				}
				lookup.DeleteMetrics(lookup)
			}
		})
	}
}
