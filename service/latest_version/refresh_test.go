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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/github"
	"github.com/release-argus/Argus/service/latest_version/types/web"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestApplyOverridesJSON(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	tLookupInterface := testLookup(t, "url", false)
	tLookup := tLookupInterface.(*web.Lookup)
	// GIVEN: a Lookup and a possible JSON string to override parts of it.
	type args struct {
		lookup             Lookup
		overrides          []byte
		semanticVerDiff    bool
		semanticVersioning *string
	}
	tests := []struct {
		name               string
		args               args
		lookupRequire      *filter.Require
		wantStr            string
		wantRequireInherit bool
		errRegex           string
	}{
		{
			name: "no overrides, no semantic versioning change",
			args: args{
				lookup:             tLookup,
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
				lookup:             tLookup,
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: test.Ptr("invalid"),
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
				lookup:             tLookup,
				overrides:          nil,
				semanticVerDiff:    true,
				semanticVersioning: test.Ptr("true"),
			},
			wantStr:  tLookup.String(""),
			errRegex: `^$`,
		},
		{
			name: "valid overrides JSON",
			args: args{
				lookup:             tLookup,
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
				lookup:             tLookup,
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
				lookup:             tLookup,
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
				lookup: tLookup,
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
				lookup: tLookup,
				overrides: []byte(test.TrimJSON(`{
					"type": "newType",
					"url": []
				}`)),
				semanticVerDiff:    false,
				semanticVersioning: nil,
			},
			errRegex: test.TrimYAML(`
				latest_version:
					type: "newType" <invalid> \(supported values = \['github', 'url'\]\)$`,
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
					&lvCfg.Soft.Require,
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
					&lvCfg.Soft.Require,
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
					&lvCfg.Soft.Require,
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

			if tc.lookupRequire != nil {
				switch lv := tc.args.lookup.(type) {
				case *web.Lookup:
					lv.Require = tc.lookupRequire
				case *github.Lookup:
					lv.Require = tc.lookupRequire
				}
			}

			// WHEN: we call applyOverridesJSON.
			got, err := applyOverridesJSON(
				tc.args.lookup,
				tc.args.overrides,
				tc.args.semanticVerDiff,
				tc.args.semanticVersioning,
			)

			prefix := fmt.Sprintf(
				"%s\napplyOverridesJSON(%q)",
				packageName, tc.args.overrides,
			)

			// THEN: we get an error matching the format expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if tc.errRegex != `^$` {
				return
			}

			// AND: the result is as expected.
			if got == nil {
				t.Errorf(
					"%s result mismatch\ngot:  nil\nwant: non-nil (%v)",
					prefix, got,
				)
				return
			}
			if gotStr := got.String(""); gotStr != tc.wantStr {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, tc.wantStr,
				)
			}

			// AND: Require.Docker.* is inherited when expected.
			gotRequire := got.GetRequire()
			if tc.wantRequireInherit {
				if gotRequire == nil || gotRequire.Docker == nil {
					t.Errorf("%s .Require result mismatch\ngot:  nil\nwant: non-nil", prefix)
				} else {
					gotQueryToken, _ := gotRequire.Docker.GetAuth().GetQueryToken(gotRequire.Docker.Detail())
					wantQueryToken, _ := tc.lookupRequire.Docker.GetAuth().GetQueryToken(tc.lookupRequire.Docker.Detail())
					if gotRequire.Docker.GetAuth().GetTokenSelf() != tc.lookupRequire.Docker.GetAuth().GetTokenSelf() ||
						gotQueryToken != wantQueryToken {
						t.Errorf(
							"%s .Require.Docker result mismatch\ngot:  %+v\nwant: %+v",
							prefix, gotRequire.Docker.GetAuth(), tc.lookupRequire.Docker.GetAuth(),
						)
					}
				}
			} else if gotRequire != nil && gotRequire.Docker != nil &&
				tc.lookupRequire != nil && tc.lookupRequire.Docker != nil {
				gotQueryToken, _ := gotRequire.Docker.GetAuth().GetQueryToken(gotRequire.Docker.Detail())
				avoidQueryToken, _ := tc.lookupRequire.Docker.GetAuth().GetQueryToken(tc.lookupRequire.Docker.Detail())
				if gotQueryToken == avoidQueryToken {
					t.Errorf(
						"%s .Require.Docker copied over unexpectedly\ngot:  %+v\nhad: %+v",
						prefix, gotRequire.Docker.GetAuth(), tc.lookupRequire.Docker.GetAuth(),
					)
				}
			}
		})
	}
}
