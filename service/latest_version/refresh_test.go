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

package latestver

import (
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/service/latest_version/filter"
	github "github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestLookup_Refresh(t *testing.T) {
	testURL := testLookup("url", false).(*web.Lookup)
	testURL.Query(true, logutil.LogFrom{})
	testVersionURL := testURL.Status.LatestVersion()
	testGitHub := testLookup("github", false).(*github.Lookup)
	testGitHub.Query(true, logutil.LogFrom{})
	testVersionGitHub := testGitHub.Status.LatestVersion()

	type args struct {
		overrides          *string
		semanticVersioning *string
	}

	// GIVEN a Lookup and a possible YAML string to override parts of it.
	tests := map[string]struct {
		args          args
		latestVersion string
		previous      Lookup
		errRegex      string
		want          string
		announce      bool
	}{
		"nil Lookup": {
			errRegex: `lookup is nil`,
		},
		"Change of URL": {
			args: args{
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"url": "` + test.LookupPlain["url_valid"] + `"
					}`)),
			},
			previous: testLookup("url", true),
			errRegex: `^$`,
			want:     testVersionURL,
		},
		"Fail CheckValues": {
			args: args{
				overrides: test.StringPtr(
					test.TrimYAML(`{
						"url_commands": [
							{"type": "unknown"}
						]
				}`)),
			},
			previous: testLookup("url", false),
			errRegex: test.TrimYAML(`
				^url_commands:
					- item_0:
						type: "[^"]+" <invalid>.*$`),
		},
		"Change of a few vars": {
			args: args{
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"url_commands": [
							{"type": "regex", "regex": "beta: \"v?([^\\\"]+)\""}
						]
				}`)),
				semanticVersioning: test.StringPtr("false"),
			},
			previous: testLookup("url", false),
			errRegex: `^$`,
			want:     testVersionURL + "-beta",
		},
		"Change of vars that fail Query": {
			args: args{
				overrides: test.StringPtr(
					test.TrimYAML(`{
						"allow_invalid_certs": false
				}`)),
			},
			previous: testLookup("url", false),
			errRegex: `x509 \(certificate invalid\)`,
		},
		"GitHub - Refresh new version": {
			previous:      testLookup("github", false),
			latestVersion: "0.0.0",
			errRegex:      `^$`,
			want:          testVersionGitHub,
			announce:      true,
		},
		"URL - Refresh new version": {
			previous:      testLookup("url", false),
			latestVersion: "0.0.0",
			errRegex:      `^$`,
			want:          testVersionURL,
			announce:      true,
		},
		"GitHub -> URL": {
			previous: testLookup("github", false),
			args: args{
				overrides: test.StringPtr(
					test.TrimYAML(`{
						"type": "url",
						"url": "` + test.LookupPlain["url_valid"] + `",
						"url_commands": [
							{"type": "regex", "regex": "ver([0-9.]+)"}
						]
				}`)),
				semanticVersioning: test.StringPtr("false")},
			latestVersion: "0.0.0",
			errRegex:      `^$`,
			want:          testVersionURL,
			announce:      false,
		},
		"GitHub -> UNKNOWN": {
			previous: testLookup("github", false),
			args: args{
				overrides: test.StringPtr(`{
					"type": "unknown"
				}`),
				semanticVersioning: test.StringPtr("false")},
			latestVersion: "0.0.0",
			errRegex:      `type: "unknown" <invalid>.*$`,
			announce:      false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Status we will be working with.
			var targetStatus *status.Status
			switch l := tc.previous.(type) {
			case *github.Lookup:
				targetStatus = l.Status
			case *web.Lookup:
				targetStatus = l.Status
			}

			// Copy the starting Status.
			var previousStatus *status.Status
			if tc.previous != nil {
				targetStatus.Init(
					0, 0, 0,
					&name, nil,
					nil)
				// Set the latest version.
				if tc.latestVersion != "" {
					targetStatus.SetLatestVersion(tc.latestVersion, "", false)
				}
				previousStatus = targetStatus.Copy()
			}

			// WHEN we call Refresh.
			got, gotAnnounce, err := Refresh(
				tc.previous,
				tc.args.overrides,
				tc.args.semanticVersioning)

			// THEN we get an error if expected.
			if tc.errRegex != "" || err != nil {
				e := util.ErrorToString(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Fatalf("want match for %q\nnot: %q",
						tc.errRegex, e)
				}
				return
			}
			// AND announce is only true when expected.
			if tc.announce != gotAnnounce {
				t.Errorf("expected announce of %t, not %t",
					tc.announce, gotAnnounce)
			}
			// AND we get the expected result otherwise.
			if tc.want != got {
				t.Errorf("expected version %q, not %q", tc.want, got)
			}
			// AND the timestamp only changes if the version changed.
			if previousStatus.LatestVersionTimestamp() != "" {
				// If the possible query-changing overrides are nil.
				if tc.args.overrides == nil && tc.args.semanticVersioning == nil {
					// The timestamp should change only if the version changed.
					if previousStatus.LatestVersion() != targetStatus.LatestVersion() &&
						previousStatus.LatestVersionTimestamp() == targetStatus.LatestVersionTimestamp() {
						t.Errorf("expected timestamp to change from %q, but got %q",
							previousStatus.LatestVersionTimestamp(), targetStatus.LatestVersionTimestamp())
						// The timestamp shouldn't change as the version didn't change.
					} else if previousStatus.LatestVersionTimestamp() != targetStatus.LatestVersionTimestamp() {
						t.Errorf("expected timestamp %q but got %q",
							previousStatus.LatestVersionTimestamp(), targetStatus.LatestVersionTimestamp())
					}
					// If the overrides are not nil.
				} else {
					// The timestamp shouldn't change.
					if previousStatus.LatestVersionTimestamp() != targetStatus.LatestVersionTimestamp() {
						t.Errorf("expected timestamp %q but got %q",
							previousStatus.LatestVersionTimestamp(), targetStatus.LatestVersionTimestamp())
					}
				}
			}
		})
	}
}

func TestApplyOverridesJSON(t *testing.T) {
	tLookup := testLookup("url", false)
	// GIVEN a Lookup and a possible JSON string to override parts of it.
	type args struct {
		lookup             Lookup
		overrides          *string
		semanticVerDiff    bool
		semanticVersioning *string
	}
	tests := map[string]struct {
		args               args
		lookupRequire      *filter.Require
		wantStr            string
		wantRequireInherit bool
		errRegex           string
	}{
		"no overrides, no semantic versioning change": {
			args: args{
				lookup:             tLookup,
				overrides:          nil,
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantStr:  tLookup.String(tLookup, ""),
			errRegex: `^$`,
		},
		"invalid semantic versioning JSON": {
			args: args{
				lookup:             tLookup,
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: test.StringPtr("invalid"),
			},
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver\.Lookup\.Options\.SemanticVersioning:
					invalid character.*$`),
		},
		"valid semantic versioning change": {
			args: args{
				lookup:             tLookup,
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: test.StringPtr("true"),
			},
			wantStr:  tLookup.String(tLookup, ""),
			errRegex: `^$`,
		},
		"valid overrides JSON": {
			args: args{
				lookup: tLookup,
				overrides: test.StringPtr(test.TrimJSON(`{
					"url": "` + test.LookupJSON["url_valid"] + `"
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantStr: test.TrimYAML(`
				type: url
				url: ` + test.LookupJSON["url_valid"] + `
				url_commands:
					- type: regex
						regex: ver([0-9.]+)
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		"invalid overrides JSON - Invalid JSON": {
			args: args{
				lookup: tLookup,
				overrides: test.StringPtr(test.TrimJSON(`{
					"url": "
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					unexpected end of JSON input$`),
		},
		"invalid overrides JSON - different var type": {
			args: args{
				lookup: tLookup,
				overrides: test.StringPtr(test.TrimJSON(`{
					"url": ["newType"]
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					cannot unmarshal array into Go struct field \.Lookup\.url of type string$`),
		},
		"overrides that make CheckValues fail": {
			args: args{
				lookup: tLookup,
				overrides: test.StringPtr(test.TrimJSON(`{
					"url": ""
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^url: <required>.*$`,
		},
		"change type with valid overrides": {
			args: args{
				lookup: tLookup,
				overrides: test.StringPtr(test.TrimJSON(`{
					"type": "github",
					"url": "release-argus/Argus",
					"access_token": "token"
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantStr: test.TrimYAML(`
				type: github
				url: release-argus/Argus
				url_commands:
					- type: regex
						regex: ver([0-9.]+)
				access_token: token
			`),
			errRegex: `^$`,
		},
		"change type to unknown type": {
			args: args{
				lookup: tLookup,
				overrides: test.StringPtr(test.TrimJSON(`{
					"type": "newType",
					"url": []
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `\stype: "newType" <invalid> \(expected one of \[github, url\]\)$`,
		},
		"inherit Require.Docker.* - same Lookup.type": {
			args: args{
				lookup:             testLookup("url", false),
				overrides:          nil,
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			lookupRequire: &filter.Require{
				Docker: filter.NewDockerCheck(
					"ghcr",
					"release-argus/argus", "{{ version }}",
					"user", "pass",
					"queryToken!", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					nil)},
			wantRequireInherit: true,
			wantStr: test.TrimYAML(`
				type: url
				url: ` + test.LookupPlain["url_invalid"] + `
				url_commands:
					- type: regex
						regex: ver([0-9.]+)
				require:
					docker:
						type: ghcr
						username: user
						token: pass
						image: release-argus/argus
						tag: '{{ version }}'
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		"inherit Require.Docker.* - different Lookup.type": {
			args: args{
				lookup: testLookup("url", false),
				overrides: test.StringPtr(`{
					"type": "github"
				}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			lookupRequire: &filter.Require{
				Docker: filter.NewDockerCheck(
					"ghcr",
					"release-argus/argus", "{{ version }}",
					"user", "pass",
					"queryToken!", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					nil)},
			wantRequireInherit: true,
			wantStr: test.TrimYAML(`
				type: github
				url: ` +
				strings.Join(
					strings.Split(
						strings.Split(
							test.LookupPlain["url_invalid"], "://")[1],
						"/")[strings.Count(test.LookupPlain["url_invalid"], "/")-3:],
					"/") + `
				url_commands:
					- type: regex
						regex: ver([0-9.]+)
				require:
					docker:
						type: ghcr
						username: user
						token: pass
						image: release-argus/argus
						tag: '{{ version }}'
			`),
			errRegex: `^$`,
		},
		"don't inherit Require.Docker.* - same Lookup.type": {
			args: args{
				lookup: testLookup("url", false),
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"require": {
							"docker": {
								"image": "release-argus/test"
				}}}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			lookupRequire: &filter.Require{
				Docker: filter.NewDockerCheck(
					"ghcr",
					"release-argus/argus", "{{ version }}",
					"user", "pass",
					"queryToken!", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					nil)},
			wantRequireInherit: false,
			wantStr: test.TrimYAML(`
				type: url
				url: ` + test.LookupPlain["url_invalid"] + `
				url_commands:
					- type: regex
						regex: ver([0-9.]+)
				require:
					docker:
						type: ghcr
						username: user
						token: pass
						image: release-argus/test
						tag: '{{ version }}'
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		"don't inherit Require.Docker.* - different Lookup.type": {
			args: args{
				lookup: testLookup("url", false),
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"type": "github",
						"require": {
							"docker": {
								"image": "release-argus/test"
				}}}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			lookupRequire: &filter.Require{
				Docker: filter.NewDockerCheck(
					"ghcr",
					"release-argus/argus", "{{ version }}",
					"user", "pass",
					"queryToken!", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					nil)},
			wantRequireInherit: false,
			wantStr: test.TrimYAML(`
				type: github
				url: ` +
				strings.Join(
					strings.Split(
						strings.Split(
							test.LookupPlain["url_invalid"], "://")[1],
						"/")[strings.Count(test.LookupPlain["url_invalid"], "/")-3:],
					"/") + `
				url_commands:
					- type: regex
						regex: ver([0-9.]+)
				require:
					docker:
						type: ghcr
						username: user
						token: pass
						image: release-argus/test
						tag: '{{ version }}'
			`),
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.lookupRequire != nil {
				if lv, ok := tc.args.lookup.(*web.Lookup); ok {
					lv.Require = tc.lookupRequire
				} else if lv, ok := tc.args.lookup.(*github.Lookup); ok {
					lv.Require = tc.lookupRequire
				}
			}

			// WHEN we call applyOverridesJSON.
			got, err := applyOverridesJSON(
				tc.args.lookup,
				tc.args.overrides,
				tc.args.semanticVerDiff,
				tc.args.semanticVersioning)

			// THEN we get an error matching the format expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("applyOverridesJSON() error mismatch\n%q\ngot:\n%q",
					tc.errRegex, e)
			}
			if tc.errRegex != `^$` {
				return
			}
			// AND the result is as expected.
			if got == nil {
				t.Errorf("applyOverridesJSON() got: nil, want: non-nil\n%s",
					e)
				return
			}
			gotStr := got.String(got, "")
			if gotStr != tc.wantStr {
				t.Errorf("applyOverridesJSON() got:\n%q\nwant:\n%q",
					gotStr, tc.wantStr)
			}
			// AND Require.Docker.* is inherited when expected.
			gotRequire := got.GetRequire()
			if tc.wantRequireInherit {
				if gotRequire == nil || gotRequire.Docker == nil {
					t.Errorf("applyOverridesJSON() Require, got: nil, want: non-nil")
				} else {
					gotQueryToken, gotValidUntil := gotRequire.Docker.CopyQueryToken()
					wantQueryToken, wantValidUntil := tc.lookupRequire.Docker.CopyQueryToken()
					if gotRequire.Docker.Token != tc.lookupRequire.Docker.Token ||
						gotQueryToken != wantQueryToken ||
						gotValidUntil != wantValidUntil {
						t.Errorf("applyOverridesJSON() Require.Docker mismatch, got:\n%+v\nwant:\n%+v",
							gotRequire.Docker, tc.lookupRequire.Docker)
					}
				}
			} else if gotRequire != nil && gotRequire.Docker != nil &&
				tc.lookupRequire != nil && tc.lookupRequire.Docker != nil {
				gotQueryToken, gotValidUntil := gotRequire.Docker.CopyQueryToken()
				avoidQueryToken, avoidValidUntil := tc.lookupRequire.Docker.CopyQueryToken()
				if gotQueryToken == avoidQueryToken &&
					gotValidUntil == avoidValidUntil {
					t.Errorf("applyOverridesJSON() Require.Docker copied over unexpectedly\ngot:\n%+v\nhad:\n%+v",
						gotRequire.Docker, tc.lookupRequire.Docker)
				}
			}
		})
	}
}
