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

package deployedver

import (
	"testing"
	"time"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/service/deployed_version/types/web"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestRefresh(t *testing.T) {
	testL := testLookup("url", false)
	_ = testL.Query(true, logutil.LogFrom{Primary: "TestLookup_Refresh"})
	testVersion := testL.GetStatus().DeployedVersion()
	if testVersion == "" {
		t.Fatalf("%s\ntest version is empty",
			packageName)
	}

	type versions struct {
		latestVersion            string
		deployedVersion          string
		deployedVersionTimestamp string
	}
	type args struct {
		overrides          *string
		semanticVersioning *string
		version            versions
	}

	// GIVEN a Lookup and various JSON strings to override parts of it.
	tests := map[string]struct {
		args     args
		previous Lookup
		errRegex string
		want     string
		announce int
	}{
		"nil Lookup": {
			errRegex: `lookup is nil`,
		},
		"invalid JSON - manual": {
			args: args{
				overrides: test.StringPtr(`{`),
			},
			previous: testLookup("manual", false),
			errRegex: `failed to unmarshal deployedver.Lookup`,
		},
		"Change of URL": {
			args: args{
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"url": "` + test.LookupJSON["url_valid"] + `"
					}`)),
			},
			previous: testLookup("url", false),
			errRegex: `^$`,
			want:     testVersion,
		},
		"Removal of URL": {
			args: args{
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"url": ""
					}`)),
			},
			previous: testLookup("url", false),
			errRegex: `url: <required>`,
			want:     "",
		},
		"Change of a few vars": {
			args: args{
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"json": "otherVersion"
					}`)),
				semanticVersioning: test.StringPtr("false"),
			},
			previous: testLookup("url", false),
			errRegex: `^$`,
			want:     testVersion + "-beta",
		},
		"Change of vars that fail Query": {
			args: args{
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"allow_invalid_certs": false
					}`)),
			},
			previous: testLookup("url", false),
			errRegex: `x509 \(certificate invalid\)`,
		},
		"Refresh new version": {
			args: args{
				version: versions{
					latestVersion:            testVersion,
					deployedVersion:          "0.0.0",
					deployedVersionTimestamp: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339)},
			},
			previous: testLookup("url", false),
			errRegex: `^$`,
			want:     testVersion,
			announce: 1,
		},
		"Refresh new version that's newer than latest": {
			args: args{
				version: versions{
					latestVersion:            "0.0.0",
					deployedVersion:          "0.0.0",
					deployedVersionTimestamp: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339)},
			},
			previous: testLookup("url", false),
			errRegex: `^$`,
			want:     testVersion,
			announce: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Status we will be working with.
			var targetStatus *status.Status
			switch l := tc.previous.(type) {
			case *web.Lookup:
				targetStatus = l.Status
			case *manual.Lookup:
				targetStatus = l.Status
			}

			// Copy the starting Status.
			var previousStatus *status.Status
			if tc.previous != nil {
				targetStatus.Init(
					0, 0, 0,
					name, "", "",
					&dashboard.Options{})
				// Set the latest version.
				if tc.args.version.latestVersion != "" {
					targetStatus.SetLatestVersion(
						tc.args.version.latestVersion, "",
						false)
				}
				if tc.args.version.deployedVersion != "" {
					targetStatus.SetDeployedVersion(
						tc.args.version.deployedVersion, tc.args.version.deployedVersionTimestamp,
						false)
				}
				previousStatus = targetStatus.Copy(true)
			}
			var previousType string
			if tc.previous != nil {
				previousType = tc.previous.GetType()
			}

			// WHEN we call Refresh.
			got, err := Refresh(
				tc.previous,
				previousType, tc.args.overrides,
				tc.args.semanticVersioning)

			// THEN we get an error if expected.
			if tc.errRegex != "" || err != nil {
				e := util.ErrorToString(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.errRegex, e)
				}
				if tc.previous == nil {
					return
				}
			}
			// AND announce is only true when expected.
			gotAnnounces := len(*targetStatus.AnnounceChannel)
			if gotAnnounces != tc.announce {
				t.Errorf("%s\nannounce channel count mismatch\nwant: %d\ngot:  %d",
					packageName, tc.announce, gotAnnounces)
			}
			// AND we get the expected result otherwise.
			if got != tc.want {
				t.Errorf("%s\nversion mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
			// AND the timestamp only changes if the version changed,
			// and the possible query-changing overrides are nil.
			if tc.args.overrides == nil && tc.args.semanticVersioning == nil {
				// If the version changed.
				if previousStatus.DeployedVersion() != targetStatus.DeployedVersion() {
					// then so should the timestamp.
					if previousStatus.DeployedVersionTimestamp() == targetStatus.DeployedVersionTimestamp() {
						t.Errorf("%s\nexpected deployed_version_timestamp to change\nhad: %q\ngot: %q",
							packageName, previousStatus.DeployedVersionTimestamp(), targetStatus.DeployedVersionTimestamp())
					}
					// otherwise, the timestamp should remain unchanged.
				} else if previousStatus.DeployedVersionTimestamp() != targetStatus.DeployedVersionTimestamp() {
					t.Errorf("%s\nexpected deployed_version_timestamp to be unchanged\nwant: %q\ngot:  %q",
						packageName, previousStatus.DeployedVersionTimestamp(), targetStatus.DeployedVersionTimestamp())
				}
				// If the overrides are not nil.
			} else {
				// The timestamp shouldn't change.
				if previousStatus.DeployedVersionTimestamp() != targetStatus.DeployedVersionTimestamp() {
					t.Errorf("%s\ntimestamp mismatch\nwant: %q\ngot:  %q",
						packageName, previousStatus.DeployedVersionTimestamp(), targetStatus.DeployedVersionTimestamp())
				}
			}
		})
	}
}

func TestApplyOverridesJSON(t *testing.T) {
	type args struct {
		lookup             Lookup
		overrides          *string
		semanticVerDiff    bool
		semanticVersioning *string
	}
	tests := map[string]struct {
		args     args
		errRegex string
	}{
		"no overrides, no semantic versioning change": {
			args: args{
				lookup:             testLookup("url", false),
				overrides:          nil,
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^$`,
		},
		"invalid semantic versioning JSON": {
			args: args{
				lookup:             testLookup("url", false),
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: test.StringPtr("invalid"),
			},
			errRegex: `failed to unmarshal deployedver\.Lookup\.Options\.SemanticVersioning`,
		},
		"valid semantic versioning change": {
			args: args{
				lookup:             testLookup("url", false),
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: test.StringPtr("true"),
			},
			errRegex: `^$`,
		},
		"valid overrides JSON": {
			args: args{
				lookup: testLookup("url", false),
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"url": "` + test.LookupJSON["url_valid"] + `"
					}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^$`,
		},
		"invalid overrides JSON - Invalid JSON": {
			args: args{
				lookup: testLookup("url", false),
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"url": "
					}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `failed to unmarshal deployedver.Lookup`,
		},
		"invalid overrides JSON - different var type": {
			args: args{
				lookup: testLookup("url", false),
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"url": [""]
					}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^failed to unmarshal deployedver.Lookup`,
		},
		"overrides that fail CheckValues": {
			args: args{
				lookup: testLookup("url", false),
				overrides: test.StringPtr(
					test.TrimJSON(`{
						"url": ""
					}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^url: <required>.*$`,
		},
		"change type with valid overrides - url to manual": {
			args: args{
				lookup: testLookup("url", false),
				overrides: test.StringPtr(test.TrimJSON(`{
					"type": "manual",
					"version": "1.2.3"
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^$`,
		},
		"change type with valid overrides - manual to url": {
			args: args{
				lookup: testLookup("manual", false),
				overrides: test.StringPtr(test.TrimJSON(`{
					"type": "url",
					"url": "` + test.LookupJSON["url_valid"] + `"
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `^$`,
		},
		"change type to unknown type": {
			args: args{
				lookup: testLookup("url", false),
				overrides: test.StringPtr(test.TrimJSON(`{
					"type": "newType",
					"url": []
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: `\stype: "newType" <invalid> \(supported types = \['url', 'manual'\]\)$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := applyOverridesJSON(
				tc.args.lookup,
				tc.args.overrides,
				tc.args.semanticVerDiff,
				tc.args.semanticVersioning)

			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, util.ErrorToString(err)) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}
