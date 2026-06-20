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

package latestver

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_Refresh(t *testing.T) {
	testURL := testLookup(t, "url", false).(*web.Lookup)
	_, _ = testURL.Query(true, logx.LogFrom{})
	testVersionURL := testURL.Status.LatestVersion()
	testGitHub := testLookup(t, "github", false).(*github.Lookup)
	_, _ = testGitHub.Query(true, logx.LogFrom{})
	testVersionGitHub := testGitHub.Status.LatestVersion()

	type args struct {
		overrides          []byte
		ignoreSecretRefs   bool
		semanticVersioning *string
		latestVersion      string
	}

	// GIVEN: a Lookup and a possible YAML string to override parts of it.
	tests := []struct {
		name     string
		args     args
		previous Lookup
		errRegex string
		want     string
		announce bool
	}{
		{
			name: "nil Lookup",
			args: args{
				ignoreSecretRefs: true,
			},
			errRegex: `lookup is nil`,
		},
		{
			name: "Change of URL",
			args: args{
				overrides: []byte(`{"url": "` + test.LookupBare["url_valid"] + `/5.0.0"}`),
			},
			previous: testLookup(t, "url", true),
			errRegex: `^$`,
			want:     testVersionURL,
		},
		{
			name: "Fail applyOverridesJSON",
			args: args{
				overrides: []byte(test.TrimJSON(`{
					"url_commands": [
						{"type": "unknown"}
					]
				}`)),
			},
			previous: testLookup(t, "url", false),
			errRegex: test.TrimYAML(`
				^url_commands:
					- item_0:
						type: "[^"]+" <invalid>.*$`,
			),
		},
		{
			name: "Removal of URL",
			args: args{
				overrides: []byte(`{"url": ""}`),
			},
			previous: testLookup(t, "url", false),
			errRegex: `url: <required>`,
			want:     "",
		},
		{
			name: "Change multiple vars",
			args: args{
				overrides: []byte(test.TrimJSON(`{
					"url": "` + test.LookupPlain["url_valid"] + `",
					"url_commands": [
						{
							"type": "regex",
							"regex": "beta: \"v?([^\\\"]+)\""
						}
					]
				}`)),
				semanticVersioning: test.Ptr("false"),
			},
			previous: testLookup(t, "url", false),
			errRegex: `^$`,
			want:     testVersionURL + "-beta",
		},
		{
			name: "Change of vars that fail Query",
			args: args{
				overrides: []byte(`{"allow_invalid_certs": false}`),
			},
			previous: testLookup(t, "url", false),
			errRegex: `x509 \(certificate invalid\)`,
		},
		{
			name: "github, refresh new version",
			args: args{
				ignoreSecretRefs: true,
				latestVersion:    "0.0.0",
			},
			previous: testLookup(t, "github", false),
			errRegex: `^$`,
			want:     testVersionGitHub,
			announce: true,
		},
		{
			name: "url, refresh new version",
			args: args{
				ignoreSecretRefs: true,
				latestVersion:    "0.0.0",
			},
			previous: testLookup(t, "url", false),
			errRegex: `^$`,
			want:     testVersionURL,
			announce: true,
		},
		{
			name:     "GitHub -> URL",
			previous: testLookup(t, "github", false),
			args: args{
				overrides: []byte(test.TrimJSON(`{
					"type": "url",
					"url": "` + test.LookupPlain["url_valid"] + `",
					"url_commands": [
						{
							"type": "regex",
							"regex": "ver([0-9.]+)"
						}
					]
				}`)),
				semanticVersioning: test.Ptr("false"),
				latestVersion:      "0.0.0",
			},
			errRegex: `^$`,
			want:     testVersionURL,
			announce: false,
		},
		{
			name:     "GitHub -> UNKNOWN",
			previous: testLookup(t, "github", false),
			args: args{
				overrides: []byte(`{
					"type": "unknown"
				}`),
				semanticVersioning: test.Ptr("false"),
				latestVersion:      "0.0.0",
			},
			errRegex: `type: "unknown" <invalid>.*$`,
			announce: false,
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
				latestVersion: testVersionURL,
			},
			previous: test.Must(t, func() (Lookup, error) {
				l := testLookup(t, "url", false)
				lTyped, _ := l.(*web.Lookup)
				lTyped.Headers = shared.Headers{
					{
						Key:   test.LookupWithHeaderAuth["header_key"],
						Value: test.LookupWithHeaderAuth["header_value_pass"],
					},
				}
				return lTyped, nil
			}),
			want:     testVersionURL,
			announce: false,
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
				latestVersion: "",
			},
			previous: test.Must(t, func() (Lookup, error) {
				l := testLookup(t, "url", false)
				lTyped, _ := l.(*web.Lookup)
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
			announce: false,
			errRegex: `Hook rules were not satisfied\.`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
					status.ServiceInfo{
						ID: tc.name,
					},
					&dashboard.Options{},
				)
				// Set the latest version.
				if tc.args.latestVersion != "" {
					targetStatus.SetLatestVersion(tc.args.latestVersion, "", false)
				}
				previousStatus = targetStatus.Copy(true)
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
			got, gotAnnounce, err := Refresh(
				tc.previous,
				tc.args.overrides,
				tc.args.semanticVersioning,
				&secretRefs,
			)

			prefix := fmt.Sprintf("%s\nRefresh()", packageName)

			// THEN: we get an error if expected.
			if tc.errRegex != "" || err != nil {
				e := errfmt.FormatError(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Fatalf(
						"%s error mismatch\ngot:  %q\nwant: %q",
						prefix, e, tc.errRegex,
					)
				}
				return
			}

			// AND: announce is only true when expected.
			if tc.announce != gotAnnounce {
				t.Errorf(
					"%s got announce %t, want%t",
					prefix, gotAnnounce, tc.announce,
				)
			}

			// AND: we get the expected result otherwise.
			if tc.want != got {
				t.Errorf(
					"%s expected version %q, not %q",
					prefix, got, tc.want,
				)
			}

			// AND: the timestamp only changes if the version changed.
			gotTimestamp := targetStatus.LatestVersionTimestamp()
			hadTimestamp := previousStatus.LatestVersionTimestamp()
			if previousStatus.LatestVersionTimestamp() != "" {
				// If the possible query-changing overrides are nil.
				if tc.args.overrides == nil && tc.args.semanticVersioning == nil {
					// The timestamp should change only if the version changed.
					if previousStatus.LatestVersion() != targetStatus.LatestVersion() &&
						gotTimestamp == hadTimestamp {
						t.Errorf(
							"%s .LatestVersionTimestamp() should have changed from %q when version changed\nwant: %q",
							prefix, gotTimestamp, hadTimestamp,
						)
						// The timestamp shouldn't change as the version didn't change.
					} else if gotTimestamp != hadTimestamp {
						t.Errorf(
							"%s\n .LatestVersionTimestamp() value mismatch\ngot:  %q\nwant: %q",
							prefix, gotTimestamp, hadTimestamp,
						)
					}
					// If the overrides are not nil.
				} else {
					// The timestamp shouldn't change.
					if gotTimestamp != hadTimestamp {
						t.Errorf(
							"%s .LatestVersionTimestamp() = %q (changed) after non-nil overrides\nwant: %q",
							prefix, gotTimestamp, hadTimestamp,
						)
					}
				}
			}
		})
	}
}
