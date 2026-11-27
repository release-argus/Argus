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

	headerAuthURL := test.HeaderAuthServer(t, "1.2.3")
	tests := []struct {
		name     string
		args     args
		previous Lookup
		errRegex string
		want     string
		announce int
	}{
		{
			name: "nil Lookup",
			args: args{
				ignoreSecretRefs: true,
			},
			errRegex: `lookup is nil`,
		},
		{
			name: "JSON/manual, invalid",
			args: args{
				overrides:        []byte(`{`),
				ignoreSecretRefs: true,
			},
			previous: testLookup(t, manual.Type, false, ""),
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+ could not find flow.*`,
			),
		},
		{
			name: "Change of URL",
			args: args{
				overrides: []byte(`{"url": "` + test.LookupBare.URLValid + `/1.2.4"}`),
			},
			previous: testLookup(t, web.Type, false, "1.2.3"),
			errRegex: `^$`,
			want:     "1.2.4",
		},
		{
			name: "Change of Type",
			args: args{
				overrides: []byte(`{"type": "url", "url": "` + test.LookupBare.URLValid + `/1.2.4"}`),
				version: versions{
					deployedVersion: new("1.2.3"),
				},
			},
			previous: testLookup(t, manual.Type, false, ""),
			errRegex: `^$`,
			want:     "1.2.4",
			announce: 0,
		},
		{
			name: "Preview finding an empty version errors",
			args: args{
				overrides:          []byte(`{"url": "` + test.LookupBare.URLValid + `/v", "regex": "v([0-9.]*)"}`),
				semanticVersioning: new("false"),
				version: versions{
					deployedVersion: new("1.2.3"),
				},
			},
			previous: testLookup(t, web.Type, false, "1.2.3"),
			errRegex: `^regex "[^"]+" returned an empty version on "v"$`,
			want:     "",
		},
		{
			name: "Removal of URL",
			args: args{
				overrides: []byte(`{"url": ""}`),
			},
			previous: testLookup(t, web.Type, false, ""),
			errRegex: `url: <required>`,
			want:     "",
		},
		{
			name: "Change of a few vars",
			args: args{
				overrides: []byte(test.TrimJSON(`{
					"url": "` + test.LookupBare.URLValid + "/" + url.QueryEscape(`{"foo":"1.2.3-beta"}`) + `",
					"json": "foo"
				}`)),
				semanticVersioning: new("false"),
			},
			previous: testLookup(t, web.Type, false, "1.2.2"),
			errRegex: `^$`,
			want:     "1.2.3-beta",
		},
		{
			name: "Change of vars that fail Query",
			args: args{
				overrides: []byte(`{"allow_invalid_certs": false}`),
			},
			previous: testLookup(t, web.Type, false, ""),
			errRegex: `x509 \(certificate invalid\)`,
		},
		{
			name: "Refresh new version that's newer than latest",
			args: args{
				ignoreSecretRefs: true,
				version: versions{
					latestVersion:            new("1.2.2"),
					deployedVersion:          new("0.0.0"),
					deployedVersionTimestamp: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339),
				},
			},
			previous: testLookup(t, web.Type, false, "1.2.3"),
			errRegex: `^$`,
			want:     "1.2.3",
			announce: 1,
		},
		{
			name: "InheritSecrets inherits header secrets",
			args: args{
				overrides: []byte(test.TrimJSON(`{
					"url": "` + headerAuthURL + `",
					"headers": [
						{
							"key": "` + test.LookupWithHeaderAuth.HeaderKey + `",
							"value": "` + util.SecretValue + `",
							"old_index": 0
						}
					]
				}`)),
				version: versions{
					latestVersion:            new("0.0.0"),
					deployedVersion:          new("0.0.0"),
					deployedVersionTimestamp: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339),
				},
			},
			previous: headerAuthLookup(t, headerAuthURL),
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
							"key": "` + test.LookupWithHeaderAuth.HeaderKey + `",
							"value": "` + test.LookupWithHeaderAuth.HeaderValueFail + `",
							"old_index": 0
						}
					]
				}`)),
				version: versions{
					latestVersion:            new(""),
					deployedVersion:          new(""),
					deployedVersionTimestamp: time.Now().UTC().Add(-4 * time.Hour).Format(time.RFC3339),
				},
			},
			previous: headerAuthLookup(t, headerAuthURL),
			want:     "",
			announce: 0,
			errRegex: `Hook rules were not satisfied\.`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var targetStatus *status.Status
			switch l := tc.previous.(type) {
			case *web.Lookup:
				targetStatus = l.Status
			case *manual.Lookup:
				targetStatus = l.Status
			}

			var previousStatus *status.Status
			if tc.previous != nil {
				targetStatus.Init(
					0, 0, 0,
					status.ServiceInfo{
						ID: tc.name,
					},
					&dashboard.Options{},
				)
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

			var secretRefs shared.VSecretRef
			if !tc.args.ignoreSecretRefs {
				if err := decode.Unmarshal("json", tc.args.overrides, &secretRefs); err != nil {
					t.Fatalf("decode.Unmarshal() failed: %v", err)
				}
			}

			got, err := Refresh(
				tc.previous,
				previousType, tc.args.overrides,
				tc.args.semanticVersioning,
				&secretRefs,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if tc.previous == nil {
				return
			}

			// A preview never announces; overrides on an already-manual lookup commit instead.
			if gotAnnounces := len(targetStatus.AnnounceChannel); gotAnnounces != tc.announce {
				t.Errorf(
					"Status.AnnounceChannel message count mismatch\ngot:  %d\nwant: %d",
					gotAnnounces, tc.announce,
				)
			}

			if got != tc.want {
				t.Errorf("version mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			hadTS := previousStatus.DeployedVersionTimestamp()
			gotTS := targetStatus.DeployedVersionTimestamp()
			if tc.args.overrides == nil && tc.args.semanticVersioning == nil {
				if previousStatus.DeployedVersion() != targetStatus.DeployedVersion() {
					if gotTS == hadTS {
						t.Errorf(
							"version changed: Status.DeployedVersionTimestamp() mismatch\n"+
								"got:  %q\nwant: NOT %q",
							gotTS, hadTS,
						)
					}
				} else if gotTS != hadTS {
					t.Errorf(
						"version unchanged: Status.DeployedVersionTimestamp() mismatch\n"+
							"got:  %q\nwant: %q",
						gotTS, hadTS,
					)
				}
			} else {
				if gotTS != hadTS {
					t.Errorf(
						"preview: Status.DeployedVersionTimestamp() mismatch\ngot:  %q\nwant: %q",
						gotTS, hadTS,
					)
				}

				if gotDV, hadDV := targetStatus.DeployedVersion(), previousStatus.DeployedVersion(); gotDV != hadDV {
					t.Errorf(
						"preview: Status.DeployedVersion() mismatch\ngot:  %q\nwant: %q",
						gotDV, hadDV,
					)
				}
			}
		})
	}
}

// headerAuthLookup returns a web Lookup that POSTs to hookURL with the passing header.
func headerAuthLookup(t *testing.T, hookURL string) Lookup {
	t.Helper()

	l := testLookup(t, web.Type, false, "")
	lTyped, ok := l.(*web.Lookup)
	if !ok {
		t.Fatalf("testLookup() type mismatch\ngot:  %T\nwant: *web.Lookup", l)
	}
	lTyped.Method = "POST"
	lTyped.URL = hookURL
	lTyped.Headers = shared.Headers{
		{
			Key:   test.LookupWithHeaderAuth.HeaderKey,
			Value: test.LookupWithHeaderAuth.HeaderValuePass,
		},
	}
	return lTyped
}
