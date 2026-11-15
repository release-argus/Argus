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
	l := testLookup(t, "github", false)
	githubLookup, ok := l.(*github.Lookup)
	if !ok {
		t.Fatalf("testLookup() type mismatch\ngot:  %T\nwant: *github.Lookup", l)
	}
	if _, err := githubLookup.Query(true, logx.LogFrom{}); err != nil {
		t.Fatalf("Lookup.Query() failed: %v", err)
	}
	testVersionGitHub := githubLookup.Status.LatestVersion()

	type args struct {
		overrides          []byte
		ignoreSecretRefs   bool
		semanticVersioning *string
		latestVersion      string
	}

	headerAuthVersion := "4.5.6"
	headerAuthURL := test.HeaderAuthServer(t, headerAuthVersion)
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
				overrides: []byte(`{"url": "` + test.LookupBare.URLValid + `/5.0.0"}`),
			},
			previous: testLookup(t, "url", true),
			errRegex: `^$`,
			want:     "5.0.0",
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
					"url": "` + test.LookupPlain.URLValid + `",
					"url_commands": [
						{
							"type": "regex",
							"regex": "beta: \"v?([^\\\"]+)\""
						}
					]
				}`)),
				semanticVersioning: new("false"),
			},
			previous: testLookup(t, "url", false),
			errRegex: `^$`,
			want:     "1.2.2-beta",
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
			want:     testWebVersion,
			announce: true,
		},
		{
			name:     "GitHub -> URL",
			previous: testLookup(t, "github", false),
			args: args{
				overrides: []byte(test.TrimJSON(`{
					"type": "url",
					"url": "` + test.LookupBare.URLValid + `/ver1.2.2",
					"url_commands": [
						{
							"type": "regex",
							"regex": "ver([0-9.]+)"
						}
					]
				}`)),
				semanticVersioning: new("false"),
				latestVersion:      "1.2.3",
			},
			errRegex: `^$`,
			want:     "1.2.2",
			announce: false,
		},
		{
			name:     "GitHub -> UNKNOWN",
			previous: testLookup(t, "github", false),
			args: args{
				overrides: []byte(`{
					"type": "unknown"
				}`),
				semanticVersioning: new("false"),
				latestVersion:      "0.0.0",
			},
			errRegex: `type: "unknown" <invalid>.*$`,
			announce: false,
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
				latestVersion: testWebVersion,
			},
			previous: headerAuthLookup(t, test.LookupBare.URLInvalid+"/"+testWebVersion),
			want:     headerAuthVersion,
			announce: false,
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
				latestVersion: "",
			},
			previous: headerAuthLookup(t, headerAuthURL),
			want:     "",
			announce: false,
			errRegex: `Hook rules were not satisfied\.`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var targetStatus *status.Status
			switch l := tc.previous.(type) {
			case *github.Lookup:
				targetStatus = l.Status
			case *web.Lookup:
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
				// Set the latest version, with a timestamp that a query would replace.
				if tc.args.latestVersion != "" {
					targetStatus.SetLatestVersion(tc.args.latestVersion, "2020-01-01T00:00:00Z", false)
				}
				previousStatus = targetStatus.Copy(true)
			}

			var secretRefs shared.VSecretRef
			if !tc.args.ignoreSecretRefs {
				if err := decode.Unmarshal("json", tc.args.overrides, &secretRefs); err != nil {
					t.Fatalf("decode.Unmarshal() failed: %v", err)
				}
			}

			got, gotAnnounce, err := Refresh(
				tc.previous,
				tc.args.overrides,
				tc.args.semanticVersioning,
				&secretRefs,
				plainDefaultsConfig(),
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			// Only a version change with no overrides (incl. semantic_versioning) announces.
			if tc.announce != gotAnnounce {
				t.Errorf("announce mismatch\ngot:  %t\nwant: %t", gotAnnounce, tc.announce)
			}

			if tc.want != got {
				t.Errorf("version mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}

			gotTimestamp := targetStatus.LatestVersionTimestamp()
			hadTimestamp := previousStatus.LatestVersionTimestamp()
			versionChanged := previousStatus.LatestVersion() != targetStatus.LatestVersion()
			// Overrides that may change the query leave the timestamp alone.
			noOverrides := tc.args.overrides == nil && tc.args.semanticVersioning == nil
			if noOverrides && versionChanged {
				if gotTimestamp == hadTimestamp {
					t.Errorf(
						"Status.LatestVersionTimestamp() mismatch\ngot:  %q\nwant: NOT %q",
						gotTimestamp, hadTimestamp,
					)
				}
			} else if gotTimestamp != hadTimestamp {
				t.Errorf(
					"Status.LatestVersionTimestamp() mismatch\ngot:  %q\nwant: %q",
					gotTimestamp, hadTimestamp,
				)
			}
		})
	}
}

// headerAuthLookup returns a url Lookup on hookURL that sends the header-auth hook's passing header.
func headerAuthLookup(t *testing.T, hookURL string) Lookup {
	t.Helper()

	l := testLookup(t, "url", false)
	lTyped, ok := l.(*web.Lookup)
	if !ok {
		t.Fatalf("testLookup() type mismatch\ngot:  %T\nwant: *web.Lookup", l)
	}
	lTyped.URL = hookURL
	lTyped.Headers = shared.Headers{
		{
			Key:   test.LookupWithHeaderAuth.HeaderKey,
			Value: test.LookupWithHeaderAuth.HeaderValuePass,
		},
	}
	return lTyped
}
