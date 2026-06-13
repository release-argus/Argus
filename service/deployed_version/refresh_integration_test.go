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

package deployedver

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/deployed_version/types/manual"
	"github.com/release-argus/Argus/service/deployed_version/types/web"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_Refresh(t *testing.T) {
	type versions struct {
		latestVersion            *string
		deployedVersion          *string
		deployedVersionTimestamp string
	}
	type args struct {
		overrides          []byte
		ignoreSecretRefs   bool
		semanticVersioning *string
		version            versions
	}

	// GIVEN: a Lookup and various JSON strings to override parts of it.
	tests := []struct {
		name     string
		args     args
		previous Lookup
		errRegex string
		want     string
		announce int
	}{
		{
			name:     "nil Lookup",
			errRegex: `lookup is nil`,
		},
		{
			name: "invalid JSON - manual",
			args: args{
				overrides:        []byte(`{`),
				ignoreSecretRefs: true,
			},
			previous: testLookup(t, "manual", false, ""),
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+ could not find flow.*`,
			),
		},
		{
			name: "Change of URL",
			args: args{
				overrides: []byte(`{"url": "` + test.LookupBare["url_valid"] + `/1.2.4"}`),
			},
			previous: testLookup(t, "url", false, "1.2.3"),
			errRegex: `^$`,
			want:     "1.2.4",
		},
		{
			name: "Removal of URL",
			args: args{
				overrides: []byte(`{"url": ""}`),
			},
			previous: testLookup(t, "url", false, ""),
			errRegex: `url: <required>`,
			want:     "",
		},
		{
			name: "Change of a few vars",
			args: args{
				overrides: []byte(test.TrimJSON(`{
					"url": "` + test.LookupBare["url_valid"] + "/" + url.QueryEscape(`{"foo":"1.2.3-beta"}`) + `",
					"json": "foo"
				}`)),
				semanticVersioning: test.Ptr("false"),
			},
			previous: testLookup(t, "url", false, "1.2.2"),
			errRegex: `^$`,
			want:     "1.2.3-beta",
		},
		{
			name: "Change of vars that fail Query",
			args: args{
				overrides: []byte(`{"allow_invalid_certs": false}`),
			},
			previous: testLookup(t, "url", false, ""),
			errRegex: `x509 \(certificate invalid\)`,
		},
		{
			name: "Refresh new version that's newer than latest",
			args: args{
				version: versions{
					latestVersion:            test.Ptr("1.2.2"),
					deployedVersion:          test.Ptr("0.0.0"),
					deployedVersionTimestamp: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339),
				},
			},
			previous: testLookup(t, "url", false, "1.2.3"),
			errRegex: `^$`,
			want:     "1.2.3",
			announce: 1,
		},
		{
			name: "InheritSecrets inherits header secrets",
			args: args{
				overrides: []byte(test.TrimJSON(`{
					"url": "` + test.LookupWithHeaderAuth["url_valid"] + `",
					"headers": [
						{
							"key": "` + test.LookupWithHeaderAuth["header_key"] + `",
							"value": "` + util.SecretValue + `",
							"old_index": 0
						}
					]
				}`)),
				version: versions{
					latestVersion:            test.Ptr("0.0.0"),
					deployedVersion:          test.Ptr("0.0.0"),
					deployedVersionTimestamp: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339),
				},
			},
			previous: test.Must(t, func() (Lookup, error) {
				l := testLookup(t, "url", false, "")
				lTyped, _ := l.(*web.Lookup)
				lTyped.Method = "POST"
				lTyped.URL = test.LookupWithHeaderAuth["url_valid"]
				lTyped.Headers = shared.Headers{
					{
						Key:   test.LookupWithHeaderAuth["header_key"],
						Value: test.LookupWithHeaderAuth["header_value_pass"],
					},
				}
				return lTyped, nil
			}),
			want:     "1.2.3",
			announce: 0,
			errRegex: `^$`,
		},
		{
			name: "InheritSecrets can be overridden with non '<secret>' values",
			args: args{
				overrides: []byte(test.TrimJSON(`{
					"headers": [
						{
							"key": "` + test.LookupWithHeaderAuth["header_key"] + `",
							"value": "` + test.LookupWithHeaderAuth["header_value_fail"] + `",
							"old_index": 0
						}
					]
				}`)),
				version: versions{
					latestVersion:            test.Ptr(""),
					deployedVersion:          test.Ptr(""),
					deployedVersionTimestamp: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339),
				},
			},
			previous: test.Must(t, func() (Lookup, error) {
				l := testLookup(t, "url", false, "")
				lTyped, _ := l.(*web.Lookup)
				lTyped.Method = "POST"
				lTyped.URL = test.LookupWithHeaderAuth["url_valid"]
				lTyped.Headers = shared.Headers{
					{
						Key:   test.LookupWithHeaderAuth["header_key"],
						Value: test.LookupWithHeaderAuth["header_value_pass"],
					},
				}
				return lTyped, nil
			}),
			want:     "",
			announce: 0,
			errRegex: `Hook rules were not satisfied\.`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
					status.ServiceInfo{
						ID: tc.name,
					},
					&dashboard.Options{},
				)
				// Set the latest version.
				if tc.args.version.latestVersion != nil {
					targetStatus.SetLatestVersion(
						*tc.args.version.latestVersion,
						"",
						false,
					)
				}
				if tc.args.version.deployedVersion != nil {
					targetStatus.SetDeployedVersion(
						*tc.args.version.deployedVersion,
						tc.args.version.deployedVersionTimestamp,
						false,
					)
				}
				previousStatus = targetStatus.Copy(true)
			}
			var previousType string
			if tc.previous != nil {
				previousType = tc.previous.GetType()
			}

			// AND: resolve the secretRefs
			var secretRefs shared.VSecretRef
			if !tc.args.ignoreSecretRefs {
				if err := decode.Unmarshal("json", tc.args.overrides, &secretRefs); err != nil {
					t.Fatalf("%s\nfailed to unmarshal secretRefs: %v",
						packageName, err,
					)
				}
			}

			// WHEN: we call Refresh.
			got, err := Refresh(
				tc.previous,
				previousType, tc.args.overrides,
				tc.args.semanticVersioning,
				&secretRefs,
			)

			prefix := fmt.Sprintf("%s\nLookup.Refresh()", packageName)

			// THEN: we get an error if expected.
			if tc.errRegex != "" || err != nil {
				e := errfmt.FormatError(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Fatalf(
						"%s error mismatch\ngot:  %q\nwant: %q",
						prefix, e, tc.errRegex,
					)
				}
				if tc.previous == nil {
					return
				}
			}

			// AND: announce is only true when expected.
			gotAnnounces := len(targetStatus.AnnounceChannel)
			if gotAnnounces != tc.announce {
				t.Errorf(
					"%s Announce channel message count mismatch\ngot:  %d\nwant: %d",
					prefix, gotAnnounces, tc.announce,
				)
			}

			// AND: we get the expected result otherwise.
			if got != tc.want {
				t.Errorf(
					"%s mismatch on version returned\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the timestamp only changes if the version changed,
			// and the possible query-changing overrides are nil.
			gotDVTS := previousStatus.DeployedVersionTimestamp()
			targetDVTS := targetStatus.DeployedVersionTimestamp()
			if tc.args.overrides == nil && tc.args.semanticVersioning == nil {
				// If the version changed.
				if previousStatus.DeployedVersion() != targetStatus.DeployedVersion() {
					// then so should the timestamp.
					if gotDVTS == targetDVTS {
						t.Errorf(
							"%s expected .DeployedVersion() to change\ngot:  %q\nhad: %q",
							prefix, gotDVTS, targetDVTS,
						)
					}
					// otherwise, the timestamp should remain unchanged.
				} else if gotDVTS != targetDVTS {
					t.Errorf(
						"%s expected .DeployedVersionTimestamp() to be unchanged\ngot:  %q\nwant: %q",
						prefix, gotDVTS, targetDVTS,
					)
				}
				// If the overrides are not nil.
			} else {
				// The timestamp shouldn't change.
				if gotDVTS != targetDVTS {
					t.Errorf(
						"%s .DeployedVersionTimestamp() value mismatch\ngot:  %q\nwant: %q",
						prefix, gotDVTS, targetDVTS,
					)
				}
			}
		})
	}
}
