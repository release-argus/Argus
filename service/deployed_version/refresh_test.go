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

	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestLookup_Refresh(t *testing.T) {
	testL := testLookup()
	testVersion, _ := testL.Query(true, logutil.LogFrom{Primary: "TestLookup_Refresh"})
	if testVersion == "" {
		t.Fatalf("test version is empty")
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

	// GIVEN a Lookup and various json strings to override parts of it
	tests := map[string]struct {
		isNil    bool
		args     args
		errRegex string
		want     string
		announce int
	}{
		"nil Lookup": {
			isNil:    true,
			errRegex: `lookup is nil`,
		},
		"Change of URL": {
			args: args{
				overrides: test.StringPtr(test.TrimJSON(`{
					"url": "https://valid.release-argus.io/json"
				}`))},
			errRegex: `^$`,
			want:     testVersion,
		},
		"Removal of URL": {
			args: args{
				overrides: test.StringPtr(test.TrimJSON(`{
					"url": ""
				}`))},
			errRegex: `url: <required>`,
			want:     "",
		},
		"Change of a few vars": {
			args: args{
				overrides: test.StringPtr(test.TrimJSON(`{
					"json": "otherVersion"
				}`)),
				semanticVersioning: test.StringPtr("false")},
			errRegex: `^$`,
			want:     testVersion + "-beta",
		},
		"Change of vars that fail Query": {
			args: args{
				overrides: test.StringPtr(test.TrimJSON(`{
					"allow_invalid_certs": false
				}`))},
			errRegex: `x509 \(certificate invalid\)`,
		},
		"Refresh new version": {
			args: args{
				version: versions{
					latestVersion:            testVersion,
					deployedVersion:          "0.0.0",
					deployedVersionTimestamp: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339)}},
			errRegex: `^$`,
			want:     testVersion,
			announce: 1,
		},
		"Refresh new version that's newer than latest": {
			args: args{
				version: versions{
					latestVersion:            "0.0.0",
					deployedVersion:          "0.0.0",
					deployedVersionTimestamp: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339)}},
			errRegex: `^$`,
			want:     testVersion,
			announce: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup()

			// Copy the starting status
			announceChannel := make(chan []byte, 4)
			svcStatus := status.New(
				&announceChannel, nil, nil,
				"",
				tc.args.version.deployedVersion, tc.args.version.deployedVersionTimestamp,
				tc.args.version.latestVersion, "",
				"")
			svcStatus.ServiceID = test.StringPtr("TestLookup_Refresh - " + name)
			lookup.Status = svcStatus
			previousStatus := lookup.Status.Copy()
			serviceID := lookup.Status.ServiceID
			if tc.isNil {
				lookup = nil
			}

			// WHEN we call Refresh
			got, err := lookup.Refresh(
				serviceID,
				tc.args.overrides,
				tc.args.semanticVersioning)

			// THEN we get an error if expected
			if tc.errRegex != "" || err != nil {
				e := util.ErrorToString(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Fatalf("want match for %q\nnot: %q",
						tc.errRegex, e)
				}
				if tc.isNil {
					return
				}
			}
			// AND announce is only true when expected
			gotAnnounces := len(*lookup.Status.AnnounceChannel)
			if tc.announce != gotAnnounces {
				t.Errorf("announce count mismatch\n want %d, got %d",
					tc.announce, gotAnnounces)
			}
			// AND we get the expected result otherwise
			if tc.want != got {
				t.Errorf("version mismatch\nwant: %q\ngot:  %q",
					tc.want, got)
			}
			// AND the timestamp only changes if the version changed
			// and the possible query-changing overrides are nil
			if tc.args.overrides == nil && tc.args.semanticVersioning == nil {
				// If the version changed
				if previousStatus.DeployedVersion() != lookup.Status.DeployedVersion() {
					// then so should the timestamp
					if previousStatus.DeployedVersionTimestamp() == lookup.Status.DeployedVersionTimestamp() {
						t.Errorf("expected deployed_version_timestamp to change\nfrom: %q\ngot:  %q",
							previousStatus.DeployedVersionTimestamp(), lookup.Status.DeployedVersionTimestamp())
					}
					// otherwise, the timestamp should remain unchanged
				} else if previousStatus.DeployedVersionTimestamp() != lookup.Status.DeployedVersionTimestamp() {
					t.Errorf("expected deployed_version_timestamp to\nremain: %q\ngot:    %q",
						previousStatus.DeployedVersionTimestamp(), lookup.Status.DeployedVersionTimestamp())
				}
				// If the overrides are not nil
			} else {
				// The timestamp shouldn't change
				if previousStatus.DeployedVersionTimestamp() != lookup.Status.DeployedVersionTimestamp() {
					t.Errorf("expected timestamp %q but got %q",
						previousStatus.DeployedVersionTimestamp(), lookup.Status.DeployedVersionTimestamp())
				}
			}
		})
	}
}

func TestApplyOverridesJSON(t *testing.T) {
	type args struct {
		lookup             *Lookup
		overrides          *string
		semanticVerDiff    bool
		semanticVersioning *string
	}
	tests := map[string]struct {
		args     args
		wantErr  bool
		errRegex string
	}{
		"no overrides, no semantic versioning change": {
			args: args{
				lookup:             testLookup(),
				overrides:          nil,
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantErr: false,
		},
		"invalid semantic versioning JSON": {
			args: args{
				lookup:             testLookup(),
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: test.StringPtr("invalid"),
			},
			wantErr:  true,
			errRegex: `failed to unmarshal semantic_versioning`,
		},
		"valid semantic versioning change": {
			args: args{
				lookup:             testLookup(),
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: test.StringPtr("true"),
			},
			wantErr: false,
		},
		"valid overrides JSON": {
			args: args{
				lookup: testLookup(),
				overrides: test.StringPtr(test.TrimJSON(`{
					"url": "https://valid.release-argus.io/json"
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantErr: false,
		},
		"invalid overrides JSON": {
			args: args{
				lookup: testLookup(),
				overrides: test.StringPtr(test.TrimJSON(`{
					"url": "
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantErr:  true,
			errRegex: `failed to unmarshal deployed_version`,
		},
		"overrides that make CheckValues fail": {
			args: args{
				lookup: testLookup(),
				overrides: test.StringPtr(test.TrimJSON(`{
					"url": ""
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantErr:  true,
			errRegex: `^url: <required>.*$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := applyOverridesJSON(
				tc.args.lookup,
				tc.args.overrides,
				tc.args.semanticVerDiff,
				tc.args.semanticVersioning)

			if (err != nil) != tc.wantErr {
				t.Errorf("applyOverridesJSON() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr && !util.RegexCheck(tc.errRegex, util.ErrorToString(err)) {
				t.Errorf("applyOverridesJSON() error = %v, wantErr %v", err, tc.errRegex)
			}
			if !tc.wantErr && got == nil {
				t.Errorf("applyOverridesJSON() got = nil, want non-nil")
			}
		})
	}
}
