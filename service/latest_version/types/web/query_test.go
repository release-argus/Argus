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
	metric "github.com/release-argus/Argus/web/metric"
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

			// THEN any err is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("web.Lookup.httpRequest() want match for %q\nnot: %q",
					tc.errRegex, e)
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
		Type: "regex", Regex: (`([0-9]\.[0-9.]+)`)}

	type wantVars struct {
		version  string
		errRegex string
	}

	tests := map[string]struct {
		bodyOverride    *string
		lookupOverrides string
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
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.URLCommands[0] = urlCommand
			err := yaml.Unmarshal([]byte(tc.lookupOverrides), lookup)
			if err != nil {
				t.Fatalf("web.Lookup.GetVersion failed to unmarshal lookupOverrides: %v", err)
			}
			lookup.Init(
				lookup.Options,
				lookup.Status,
				lookup.Defaults, lookup.HardDefaults)
			testBody := util.PtrValueOrValue(tc.bodyOverride, body)

			// WHEN getVersion is called on it.
			version, err := lookup.getVersion(
				testBody, logutil.LogFrom{})

			// THEN any err is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf("web.Lookup.getVersion() error mismatch:\nwant: %q\ngot:  %q",
					tc.want.errRegex, e)
			}
			// AND the version is as expected.
			if tc.want.version != version {
				t.Errorf("web.Lookup.getVersion() version mismatch:\nwant: %q\ngot:  %q",
					tc.want.version, version)
			}
		})
	}
}

func TestQuery(t *testing.T) {
	testLookupVersions := testLookup(false)
	testLookupVersions.query(logutil.LogFrom{})

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
				url: https://valid.release-argus.io/plain
				url_commands:
					- type: regex
					  regex: ver[0-9.]+
			`),
			want: wantVars{
				errRegex: `failed converting .* to a semantic version`},
		},
		"query on self-signed https works when allowed": {
			overrides: test.TrimYAML(`
				url: https://invalid.release-argus.io/plain
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
				url: https://invalid.release-argus.io/plain
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
				errRegex: `failed converting "[^"]+" to a semantic version`},
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
				lookup.Status.ServiceID = &name
				lookup.Options.SemanticVersioning = tc.semanticVersioning
				// hadStatus
				lookup.Status.SetLatestVersion(tc.hadStatus.latestVersion, "", false)
				lookup.Status.SetDeployedVersion(tc.hadStatus.deployedVersion, "", false)
				// overrides
				err := yaml.Unmarshal([]byte(tc.overrides), lookup)
				if err != nil {
					t.Fatalf("web.Lookup.Query failed to unmarshal overrides: %v", err)
				}
				lookup.Init(
					lookup.Options,
					lookup.Status,
					lookup.Defaults, lookup.HardDefaults)

				// WHEN Query is called on it.
				var newVersion bool
				newVersion, err = lookup.Query(true, logutil.LogFrom{})

				// THEN any err is expected.
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
					t.Fatalf("web.Lookup.Query() want match for %q\nnot: %q",
						tc.want.errRegex, e)
				}
				// AND the stdout contains the expected strings.
				if !util.RegexCheck(tc.want.stdoutRegex, stdout) {
					t.Fatalf("web.Lookup.Query() match for %q not found in:\n%q",
						tc.want.stdoutRegex, stdout)
				}
				// AND the LatestVersion is as expected.
				if tc.hadStatus.latestVersionWant != "" &&
					tc.hadStatus.latestVersionWant != lookup.Status.LatestVersion() {
					t.Fatalf("web.Lookup.Query() wanted LatestVersion to become %q, not %q",
						tc.hadStatus.latestVersionWant, lookup.Status.LatestVersion())
				}
				if tc.want.announce && len(*lookup.Status.AnnounceChannel) != 1 {
					t.Fatalf("web.Lookup.Query() wanted an announcement")
				}
				if tc.want.newVersion != newVersion {
					t.Fatalf("web.Lookup.Query() wanted newVersion to be %t, not %t",
						tc.want.newVersion, newVersion)
				}
				// AND the metrics are as expected.
				// FAIL
				want := 0
				if tc.want.errRegex != "^$" {
					want = 1
				}
				got := testutil.ToFloat64(metric.LatestVersionQueryResultTotal.WithLabelValues(
					*lookup.Status.ServiceID,
					lookup.Type,
					"FAIL"))
				if got != float64(want) {
					t.Fatalf("web.Lookup.Query() LatestVersionQueryResultTotal - FAIL\nwant: %d\ngot:  %f",
						want, got)
				}
				// SUCCESS
				want = 0
				if tc.want.errRegex == "^$" {
					want = 1
				}
				got = testutil.ToFloat64(metric.LatestVersionQueryResultTotal.WithLabelValues(
					*lookup.Status.ServiceID,
					lookup.Type,
					"SUCCESS"))
				if got != float64(want) {
					t.Fatalf("web.Lookup.Query() LatestVersionQueryResultTotal - FAIL\nwant: %d\ngot:  %f",
						want, got)
				}
				lookup.DeleteMetrics(lookup)
			}
		})
	}
}
