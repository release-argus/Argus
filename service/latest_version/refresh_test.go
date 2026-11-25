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

package latestver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/web"
	opttest "github.com/release-argus/Argus/service/option/test"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestRefresh__overridesRemovingTheLookup(t *testing.T) {
	tests := []struct {
		name               string
		semanticVersioning *string
	}{
		{
			name:               "semantic_versioning unchanged",
			semanticVersioning: nil,
		},
		{
			name:               "semantic_versioning toggled off",
			semanticVersioning: new("false"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, "url", false)
			lookup.GetOptions().SemanticVersioning = new(true)

			version, announce, err := Refresh(
				lookup,
				[]byte("null"),
				tc.semanticVersioning,
				nil,
				plainDefaultsConfig(),
			)

			// An error, not a panic on the removed Lookup.
			e := errfmt.FormatError(err)
			if wantErrRegex := `removed by overrides`; !util.RegexCheck(wantErrRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, wantErrRegex)
			}

			if version != "" || announce {
				t.Errorf(
					"version, announce mismatch\ngot:  %q, %t\nwant: %q, %t",
					version, announce,
					"", false,
				)
			}
		})
	}
}

func TestApplyOverridesJSON(t *testing.T) {
	lvCfg := plainDefaultsConfig()

	tLookupInterface := testLookup(t, "url", false)
	tLookup := tLookupInterface.(*web.Lookup)

	type args struct {
		lookup             Lookup
		overrides          []byte
		semanticVerDiff    bool
		semanticVersioning *string
	}
	tests := []struct {
		name                   string
		args                   args
		lookupRequire          *filter.Require
		wantStr                string
		wantSemanticVersioning *bool
		wantRequireInherit     bool
		errRegex               string
	}{
		{
			name: "no overrides, no semantic versioning change",
			args: args{
				lookup:             testLookup(t, "url", false),
				overrides:          nil,
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantStr:  tLookup.String(""),
			errRegex: `^$`,
		},
		{
			name: "invalid semantic versioning JSON",
			args: args{
				lookup:             testLookup(t, "url", false),
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: new("invalid"),
			},
			errRegex: test.TrimYAML(`
				^semantic_versioning:
					jsontext:
						invalid character .*$`,
			),
		},
		{
			name: "valid semantic versioning change",
			args: args{
				lookup:             testLookup(t, "url", false),
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: new("true"),
			},
			wantStr:                tLookup.String(""),
			wantSemanticVersioning: new(true),
			errRegex:               `^$`,
		},
		{
			name: "valid overrides JSON",
			args: args{
				lookup:             testLookup(t, "url", false),
				overrides:          []byte(`{"url": "https://example.com"}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantStr: test.TrimYAML(`
				type: url
				url: https://example.com
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		{
			name: "invalid overrides JSON/Invalid JSON",
			args: args{
				lookup:             testLookup(t, "url", false),
				overrides:          []byte(`{"url": "}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: test.TrimYAML(`
				^latest_version:
					[^\s]+ could not find.*`,
			),
		},
		{
			name: "invalid overrides JSON/different var type",
			args: args{
				lookup:             testLookup(t, "url", false),
				overrides:          []byte(`{"url": ["newType"]}`),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: test.TrimYAML(`
				^latest_version:
					[^\s]+ .*unmarshal.*`,
			),
		},
		{
			name: "change type with valid overrides",
			args: args{
				lookup: testLookup(t, "url", false),
				overrides: []byte(test.TrimJSON(`{
					"type": "github",
					"url": "` + test.ArgusGitHubRepo + `",
					"access_token": "token"
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			wantStr: test.TrimYAML(`
				type: github
				url: ` + test.ArgusGitHubRepo + `
				access_token: token
			`),
			errRegex: `^$`,
		},
		{
			name: "change type to unknown type",
			args: args{
				lookup: testLookup(t, "url", false),
				overrides: []byte(test.TrimJSON(`{
					"type": "newType",
					"url": []
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: test.TrimYAML(`
				latest_version:
					type: "newType" <invalid> \(supported values = \['forgejo', 'github', 'url'\]\)$`,
			),
		},
		{
			name: "inherit Require.Docker.*, same Lookup.type, same Docker",
			args: args{
				lookup:             testLookup(t, "url", false),
				overrides:          nil,
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			lookupRequire: test.Must(t, func() (*filter.Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return filter.Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: `+test.ArgusDockerGHCRRepo+`
							tag: '{{ version }}'
							auth:
								token: pass
					`)),
					svcStatus,
					&lvCfg.Soft.Common.Require,
				)
			}),
			wantRequireInherit: true,
			wantStr: test.TrimYAML(`
				type: url
				url: ` + tLookup.URL + `
				require:
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: '{{ version }}'
						auth:
							token: pass
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		{
			name: "don't inherit Require.Docker.*, same Lookup.type, different Docker.Type",
			args: args{
				lookup: testLookup(t, "url", false),
				overrides: []byte(
					test.TrimJSON(`{
						"require": {
							"docker": {
								"type": "hub",
								"image": "release-argus/test"
							}
						}
					}`),
				),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			lookupRequire: test.Must(t, func() (*filter.Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return filter.Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: `+test.ArgusDockerGHCRRepo+`
							tag: '{{ version }}'
							auth:
								token: 'pass'
					`)),
					svcStatus,
					&lvCfg.Soft.Common.Require,
				)
			}),
			wantRequireInherit: false,
			wantStr: test.TrimYAML(`
				type: url
				url: ` + tLookup.URL + `
				require:
					docker:
						type: hub
						image: release-argus/test
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		{
			name: "changing type only uses overrides, does inherit Docker if same type or image or auth",
			args: args{
				lookup: testLookup(t, "url", false),
				overrides: []byte(
					test.TrimJSON(`{
						"type": "github",
						"url": "` + test.ArgusGitHubRepo + `",
						"require": {
							"docker": {
								"type": "ghcr",
								"image": "` + test.ArgusDockerGHCRRepo + `",
								"tag": "{{ version }}-beta",
								"auth": {
									"token": "pass"
								}
							}
						}
					}`),
				),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			lookupRequire: test.Must(t, func() (*filter.Require, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return filter.Decode(
					"yaml", []byte(test.TrimYAML(`
						docker:
							type: ghcr
							image: `+test.ArgusDockerGHCRRepo+`
							tag: '{{ version }}'
							auth:
								token: 'pass'
					`)),
					svcStatus,
					&lvCfg.Soft.Common.Require,
				)
			}),
			wantRequireInherit: true,
			wantStr: test.TrimYAML(`
				type: github
				url: ` + test.ArgusGitHubRepo + `
				require:
					docker:
						type: ghcr
						image: ` + test.ArgusDockerGHCRRepo + `
						tag: '{{ version }}-beta'
						auth:
							token: pass
			`),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var hadToken, hadQueryToken string
			if tc.lookupRequire != nil {
				tc.args.lookup.(*web.Lookup).Require = tc.lookupRequire
				if tc.lookupRequire.Docker != nil {
					auth := tc.lookupRequire.Docker.GetAuth()
					hadToken = auth.GetTokenSelf()
					// Seeded so inheritance has a query token to copy without a live token request.
					hadQueryToken = "seed-token"
					auth.SetQueryToken(hadQueryToken, time.Now().Add(time.Hour))
				}
			}

			got, err := applyOverridesJSON(
				tc.args.lookup,
				tc.args.overrides,
				tc.args.semanticVerDiff,
				tc.args.semanticVersioning,
				lvCfg,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if got == nil {
				t.Error("lookup mismatch\ngot:  nil\nwant: non-nil")
				return
			}
			if gotStr := got.String(""); gotStr != tc.wantStr {
				t.Errorf("Lookup.String() mismatch\ngot:  %q\nwant: %q", gotStr, tc.wantStr)
			}

			gotSemVer := test.StringifyPtr(got.GetOptions().SemanticVersioning)
			if wantSemVer := test.StringifyPtr(tc.wantSemanticVersioning); gotSemVer != wantSemVer {
				t.Errorf(
					"Lookup.GetOptions().SemanticVersioning mismatch\ngot:  %s\nwant: %s",
					gotSemVer, wantSemVer,
				)
			}

			gotRequire := got.GetRequire()
			if tc.wantRequireInherit {
				if gotRequire == nil || gotRequire.Docker == nil {
					t.Error("Lookup.GetRequire().Docker mismatch\ngot:  nil\nwant: non-nil")
				} else {
					gotQueryToken, _ := gotRequire.Docker.GetAuth().GetQueryTokenSelf()
					gotToken := gotRequire.Docker.GetAuth().GetTokenSelf()
					if gotToken != hadToken || gotQueryToken != hadQueryToken {
						t.Errorf(
							"Lookup.GetRequire().Docker token, queryToken mismatch\n"+
								"got:  %q, %q\nwant: %q, %q",
							gotToken, gotQueryToken,
							hadToken, hadQueryToken,
						)
					}
				}
			} else if gotRequire != nil && gotRequire.Docker != nil &&
				tc.lookupRequire != nil && tc.lookupRequire.Docker != nil {
				gotQueryToken, _ := gotRequire.Docker.GetAuth().GetQueryTokenSelf()
				if gotQueryToken == hadQueryToken {
					t.Errorf(
						"Lookup.GetRequire().Docker queryToken mismatch\ngot:  %q\nwant: NOT %q",
						gotQueryToken, hadQueryToken,
					)
				}
			}
		})
	}
}

func TestRefresh__semanticVersioningOnlyIsPreview(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("2.0.0"))
	}))
	t.Cleanup(srv.Close)

	tests := []struct {
		name               string
		semanticVersioning string
	}{
		{
			name:               "semantic_versioning toggled off",
			semanticVersioning: "false",
		},
		{
			name:               "semantic_versioning reset to the default",
			semanticVersioning: "null",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svcStatus, _ := statustest.New("yaml", nil)
			svcStatus.SetLatestVersion("1.0.0", "", false)
			options := opttest.Options(t)
			options.SemanticVersioning = new(true)
			lookup, err := Decode(
				"yaml", []byte("type: url\nurl: "+srv.URL+"\nurl_commands: []\n"),
				options,
				svcStatus,
				plainDefaultsConfig(),
			)
			if err != nil {
				t.Fatalf("Decode() failed: %v", err)
			}
			if tc.semanticVersioning == "null" {
				options.SemanticVersioning = new(false)
			}
			hadSemanticVersioning := options.GetSemanticVersioning()

			version, announce, err := Refresh(
				lookup,
				nil,
				&tc.semanticVersioning,
				nil,
				plainDefaultsConfig(),
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if testErr := test.AssertFields(t, []test.FieldAssertion{
				{Name: "version", Got: version, Want: "2.0.0", Mode: test.CompareEqual},
				{Name: "announce", Got: announce, Want: false, Mode: test.CompareEqual},
				{Name: "Status.LatestVersion()", Got: svcStatus.LatestVersion(), Want: "1.0.0", Mode: test.CompareEqual},
				{Name: "len(Status.AnnounceChannel)", Got: len(svcStatus.AnnounceChannel), Want: 0, Mode: test.CompareEqual},
				{Name: "len(Status.DatabaseChannel)", Got: len(svcStatus.DatabaseChannel), Want: 0, Mode: test.CompareEqual},
				{Name: "Options.GetSemanticVersioning()", Got: options.GetSemanticVersioning(), Want: hadSemanticVersioning, Mode: test.CompareEqual},
			}, ""); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestApplyOverridesJSON__typeChangeCopiesOptions(t *testing.T) {
	lookup := testLookup(t, "url", false)
	options := lookup.GetOptions()
	options.SemanticVersioning = new(true)

	newLookup, err := applyOverridesJSON(
		lookup,
		[]byte(`{"type": "github", "url": "release-argus/Argus"}`),
		true,
		new("false"),
		plainDefaultsConfig(),
	)
	if err != nil {
		t.Fatalf("applyOverridesJSON() failed: %v", err)
	}

	if testErr := test.AssertFields(t, []test.FieldAssertion{
		{Name: "newLookup.GetOptions()", Got: newLookup.GetDefaults(), Want: options, Mode: test.CompareDifferentPointer},
		{Name: "Options.GetSemanticVersioning()", Got: options.GetSemanticVersioning(), Want: true, Mode: test.CompareEqual},
	}, ""); testErr != nil {
		t.Error(testErr)
	}
}
