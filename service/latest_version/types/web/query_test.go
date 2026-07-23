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

// Package web provides a web-based lookup type.
package web

import (
	"fmt"
	"sort"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_HTTPRequest(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name      string
		failing   bool
		overrides string
		bodyRegex string
		errRegex  string
	}{
		{
			name:      "invalid url",
			overrides: `url: "https://	test"`,
			bodyRegex: `^$`,
			errRegex:  `invalid control character in URL`,
		},
		{
			name:      "unknown url",
			overrides: `url: https://release-argus.invalid-tld`,
			bodyRegex: `^$`,
			errRegex:  `no such host`,
		},
		{
			name:      "valid url",
			overrides: `url: https://release-argus.io`,
			bodyRegex: `.*`,
			errRegex:  `^$`,
		},
		{
			name:      "invalid cert",
			failing:   true,
			bodyRegex: `^$`,
			errRegex:  `x509`,
		},
		{
			name: "headers/pass",
			overrides: test.TrimYAML(`
				method: POST
				url: ` + test.LookupWithHeaderAuth["url_valid"] + `
				headers:
					- key: ` + test.LookupWithHeaderAuth["header_key"] + `
						value: ` + test.LookupWithHeaderAuth["header_value_pass"] + `
			`),
			bodyRegex: `^[\d.]+$`,
			errRegex:  `^$`,
		},
		{
			name: "headers/fail",
			overrides: test.TrimYAML(`
				method: POST
				url: ` + test.LookupWithHeaderAuth["url_valid"] + `
				headers:
					- key: ` + test.LookupWithHeaderAuth["header_key"] + `
						value: ` + test.LookupWithHeaderAuth["header_value_fail"] + `
			`),
			bodyRegex: `Hook rules were not satisfied\.`,
			errRegex:  `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, tc.failing)
			// Apply overrides.
			if err := lookup.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal overrides: %s",
					packageName, err,
				)
			}

			// WHEN: httpRequest is called on it.
			body, err := lookup.httpRequest(logx.LogFrom{})

			prefix := fmt.Sprintf("%s\nLookup.httpRequest", packageName)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the body matches the expected regex.
			if tc.bodyRegex != "" {
				if !util.RegexCheck(tc.bodyRegex, string(body)) {
					t.Errorf(
						"%s body mismatch\ngot:  %q\nwant: %q",
						prefix, string(body), tc.bodyRegex,
					)
				}
			}
		})
	}
}

func TestLookup_GetVersion(t *testing.T) {
	// GIVEN: a Lookup and a Body to filter.
	body := `
		version 1 is "v0.0.0"
		version 2 is "ver1.2.3-dev"
		version 3 is "ver1.2.4"
		version 4 is "ver1.2.5"
	`
	urlCommand := filter.URLCommand{Type: "regex", Regex: `([0-9]\.[0-9.]+)`}

	type wantVars struct {
		version  string
		errRegex string
	}

	tests := []struct {
		name            string
		bodyOverride    *string
		lookupOverrides string
		semVer          bool
		want            wantVars
	}{
		{
			name:            "nil url_commands",
			lookupOverrides: `url_commands: []`,
			semVer:          false,
			want: wantVars{
				version:  body,
				errRegex: `^$`,
			},
		},
		{
			name:            "empty body",
			bodyOverride:    new(""),
			lookupOverrides: `url_commands: []`,
			semVer:          true,
			want: wantVars{
				errRegex: `^no releases were found matching the url_commands`,
			},
		},
		{
			name:   "nil Require",
			semVer: false,
			want: wantVars{
				version:  "0.0.0",
				errRegex: `^$`,
			},
		},
		{
			name: "url_commands that error",
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: ver9
						index: 9
			`),
			semVer: true,
			want: wantVars{
				errRegex: `regex "[^"]+" didn't return any matches`,
			},
		},
		{
			name: "url_commands that pass",
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"ver([0-9][^"]+)"'
			`),
			semVer: false,
			want: wantVars{
				version:  "1.2.3-dev",
				errRegex: `^$`,
			},
		},
		{
			name: "regex_version mismatch",
			lookupOverrides: test.TrimYAML(`
				require:
					regex_version: ver(2\.[0-9.]+)
			`),
			semVer: true,
			want: wantVars{
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
						regex "[^"]+" not matched on version "[^"]+"$`,
				),
			},
		},
		{
			name: "regex_version match",
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
			`),
			semVer: false,
			want: wantVars{
				version:  "1.2.4",
				errRegex: `^$`,
			},
		},
		{
			name: "regex_content mismatch",
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '4 is "ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
					regex_content: ver[0-9]+.exe
			`),
			semVer: true,
			want: wantVars{
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
						regex "[^"]+" not matched on content for version "[^"]+"$`,
				),
			},
		},
		{
			name: "regex_content match",
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
					regex_content: 'version 4 is "ver{{ version }}"'
			`),
			semVer: true,
			want: wantVars{
				version:  "1.2.5",
				errRegex: `^$`,
			},
		},
		{
			name: "command fail",
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '4 is "ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
					regex_content: 'version 4 is "ver{{ version }}"'
					command: [false]
			`),
			semVer: true,
			want: wantVars{
				errRegex: test.TrimYAML(`
					^no releases were found matching the require field.*
						command failed:
							exit status 1$`,
				),
			},
		},
		{
			name: "command pass",
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"ver([0-9][^"]+)"'
				require:
					regex_version: ^1\.[0-9.]+$
					regex_content: 'version 4 is "ver{{ version }}"'
					command: [true]
			`),
			semVer: true,
			want: wantVars{
				version:  "1.2.5",
				errRegex: `^$`,
			},
		},
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
						token: ` + test.DockerHubToken(t) + `
			`),
			semVer: true,
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
						tag: '0.8.0'
						token: ` + test.GitHubToken(t) + `
			`),
			semVer: true,
			want: wantVars{
				version:  "1.2.5",
				errRegex: `^$`,
			},
		},
		{
			name:         "version fails semver",
			bodyOverride: new("1_0_0"),
			lookupOverrides: test.TrimYAML(`
				url_commands: []
			`),
			semVer: true,
			want: wantVars{
				errRegex: `failed to convert "1_0_0" to a semantic version`,
			},
		},
		{
			name: "semver skips non-semver values",
			bodyOverride: new(`
				bad "1_0_0"
				good "v1.0.0"
				good "v2.0.0"
				bad "v3_0_0"
			`),
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"(v?[0-9][^"]+)"'
				require:
					regex_version: ^v2\.[0-9.]+$
			`),
			semVer: true,
			want: wantVars{
				version:  "v2.0.0",
				errRegex: `^$`,
			},
		},
		{
			name: "mixed semver/non-semver picks same winner regardless of candidate order (order A)",
			bodyOverride: new(`
				bad "1_0_0"
				good "v1.0.0"
				good "v2.0.0"
				bad "v3_0_0"
			`),
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"(v?[0-9][^"]+)"'
			`),
			semVer: true,
			want: wantVars{
				version:  "v2.0.0",
				errRegex: `^$`,
			},
		},
		{
			name: "mixed semver/non-semver picks same winner regardless of candidate order (order B)",
			bodyOverride: new(`
				bad "v3_0_0"
				good "v2.0.0"
				bad "1_0_0"
				good "v1.0.0"
			`),
			lookupOverrides: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '"(v?[0-9][^"]+)"'
			`),
			semVer: true,
			want: wantVars{
				version:  "v2.0.0",
				errRegex: `^$`,
			},
		},
		{
			name: "sorts versions when semantic_versioning enabled",
			bodyOverride: new(`
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
			semVer: true,
			want: wantVars{
				version:  "v1.0.1",
				errRegex: `^$`,
			},
		},
		{
			name: "does not sort versions when semantic_versioning disabled",
			bodyOverride: new(`
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
			semVer: false,
			want: wantVars{
				version:  "0.4.7",
				errRegex: `^$`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			// overrides.
			lookup.URLCommands[0] = urlCommand
			if err := lookup.ApplyOverrides("yaml", []byte(tc.lookupOverrides)); err != nil {
				t.Fatalf(
					"%s\nfailed to unmarshal Lookup overrides: %v",
					packageName, err,
				)
			}
			lookup.Options.SemanticVersioning = &tc.semVer
			if err := lookup.CheckValues(); err != nil {
				t.Fatalf(
					"%s\nfailed to check Lookup values after applying test overrides: %v",
					packageName, err,
				)
			}
			testBody := util.DerefOr(tc.bodyOverride, body)

			// WHEN: getVersion is called on it.
			version, err := lookup.getVersion(testBody, logx.LogFrom{})

			prefix := fmt.Sprintf(
				"%s\nLookup.getVersion(%q)",
				packageName, testBody,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, tc.want.errRegex, e,
				)
			}

			// AND: the version is as expected.
			if version != tc.want.version {
				t.Errorf(
					"%s version mismatch:\ngot:  %q\nwant: %q",
					prefix, version, tc.want.version,
				)
			}
		})
	}
}

func TestVersionSortsBefore(t *testing.T) {
	// GIVEN: two version strings.
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "higher semver sorts before lower semver",
			a:    "1.10.0",
			b:    "1.2.3",
			want: true,
		},
		{
			name: "lower semver does not sort before higher semver",
			a:    "1.2.3",
			b:    "1.10.0",
			want: false,
		},
		{
			name: "release sorts before its own pre-release",
			a:    "1.2.3",
			b:    "1.2.3-rc.1",
			want: true,
		},
		{
			name: "pre-release does not sort before its own release",
			a:    "1.2.3-rc.1",
			b:    "1.2.3",
			want: false,
		},
		{
			name: "pre-release with more identifiers sorts before one with fewer",
			a:    "1.2.3-alpha.1",
			b:    "1.2.3-alpha",
			want: true,
		},
		{
			name: "alphanumeric pre-release identifier sorts before numeric identifier",
			a:    "1.2.3-alpha",
			b:    "1.2.3-1",
			want: true,
		},
		{
			name: "numeric pre-release identifiers compare numerically, not lexically",
			a:    "1.2.3-rc.10",
			b:    "1.2.3-rc.2",
			want: true,
		},
		{
			name: "v-prefixed and bare versions of equal precedence are not ordered",
			a:    "v1.2.3",
			b:    "1.2.3",
			want: false,
		},
		{
			name: "bare and v-prefixed versions of equal precedence are not ordered (reverse)",
			a:    "1.2.3",
			b:    "v1.2.3",
			want: false,
		},
		{
			name: "build metadata is ignored for precedence",
			a:    "1.2.3+build.1",
			b:    "1.2.3+build.2",
			want: false,
		},
		{
			name: "coerced partial versions of equal precedence are not ordered",
			a:    "1.2",
			b:    "1.2.0",
			want: false,
		},
		{
			name: "a parseable version sorts before an unparseable one",
			a:    "1.2.3",
			b:    "latest",
			want: true,
		},
		{
			name: "an unparseable version does not sort before a parseable one",
			a:    "latest",
			b:    "1.2.3",
			want: false,
		},
		{
			name: "unparseable versions fall back to a deterministic lexical tiebreak",
			a:    "latest",
			b:    "nightly",
			want: true,
		},
		{
			name: "unparseable versions fall back to a deterministic lexical tiebreak (reverse)",
			a:    "nightly",
			b:    "latest",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: versionSortsBefore is called on the pair.
			got := versionSortsBefore(tc.a, tc.b)

			// THEN: it should return the expected value.
			if got != tc.want {
				t.Errorf(
					"%s\nversionSortsBefore(%q, %q) mismatch\ngot:  %t\nwant: %t",
					packageName, tc.a, tc.b, got, tc.want,
				)
			}
		})
	}
}

func TestVersionSortsBefore__OrderIndependence(t *testing.T) {
	// GIVEN: the same set of semver and non-semver candidates, supplied in all input orders.
	want := []string{"1.10.0", "1.9.9", "1.2.3", "latest", "nightly"}
	permutations := test.Permutations(want)

	for i, perm := range permutations {
		name := fmt.Sprintf("permutation %d", i)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := append([]string{}, perm...)

			// WHEN: the set is sorted using versionSortsBefore.
			sort.Slice(got, func(i, j int) bool {
				return versionSortsBefore(got[i], got[j])
			})

			// THEN: the output order is identical regardless of the input order.
			if !util.AreSlicesEqual(got, want) {
				t.Errorf(
					"%s\nsort.Slice(%v, versionSortsBefore) mismatch\ngot:  %v\nwant: %v",
					packageName, perm, got, want,
				)
			}
		})
	}
}
