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

// Package github provides a github-based lookup type.
package github

import (
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	filtertest "github.com/release-argus/Argus/service/latest_version/filter/test"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	"github.com/release-argus/Argus/util"
)

func checkDockerToken(
	t *testing.T,
	wantQueryToken, gotQueryToken string,
	wantValidUntil, gotValidUntil time.Time,
) {
	t.Helper()

	if gotQueryToken != wantQueryToken {
		t.Errorf(
			"Lookup.Require.Docker queryToken mismatch\ngot:  %q\nwant: %q",
			gotQueryToken, wantQueryToken,
		)
	}
	if !gotValidUntil.Equal(wantValidUntil) {
		t.Errorf(
			"Lookup.Require.Docker validUntil mismatch\ngot:  %s\nwant: %s",
			gotValidUntil, wantValidUntil,
		)
	}
}

func TestLookup_InheritSecrets(t *testing.T) {
	testData := newData(
		"etag",
		&[]forgetypes.Release{
			{URL: "foo"},
			{URL: "bar"},
		},
	)
	testData.SetTagFallback()

	tests := []struct {
		name               string
		typeChanged        bool
		overrides          string
		inheritData        bool
		accessToken        string // Token the edit carries, when not masked.
		maskAccessToken    bool
		inheritAccessToken bool
		inheritRequire     bool
	}{
		{
			name:        "don't inherit Data as Type changed",
			typeChanged: true,
			overrides: test.TrimYAML(`
				type: something-else
				url: something-else
			`),
			inheritData:    false,
			inheritRequire: true,
		},
		{
			name:               "don't inherit Data as URL changed",
			overrides:          "url: something-else",
			inheritData:        false,
			maskAccessToken:    true,
			inheritAccessToken: true,
			inheritRequire:     true,
		},
		{
			name: "inherit Data, not Require when Docker.Type changed",
			overrides: test.TrimYAML(`
				require:
					docker:
						type: ghcr
			`),
			inheritData:    true,
			inheritRequire: false,
		},
		{
			name: "inherit AccessToken",
			overrides: test.TrimYAML(`
				url: something-else
				require:
					docker:
						type: ` + docker.PossibleTypes[len(docker.PossibleTypes)-1] + `
			`),
			inheritData:        false,
			maskAccessToken:    true,
			inheritAccessToken: true,
			inheritRequire:     false,
		},
		{
			name:               "inherit Require, not Data",
			overrides:          `url: something-else`,
			inheritData:        false,
			maskAccessToken:    true,
			inheritAccessToken: true,
			inheritRequire:     true,
		},
		{
			name:               "inherit all",
			inheritData:        true,
			inheritRequire:     true,
			maskAccessToken:    true,
			inheritAccessToken: true,
		},
		{
			name:            "a masked token cannot be inherited from another type",
			typeChanged:     true,
			overrides:       "type: something-else",
			maskAccessToken: true,
			inheritRequire:  true,
		},
		{
			name:           "a new token is kept",
			overrides:      "url: something-else",
			accessToken:    "new-token",
			inheritRequire: true,
		},
		{
			name:           "an emptied token is kept empty",
			accessToken:    "",
			inheritData:    true,
			inheritRequire: true,
		},
	}

	otherLookupTest := &struct {
		base.Lookup `yaml:",inline"`
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t)
			lookup.Require = filtertest.Require(t, "hub")
			lookup.Require.Docker.GetAuth().SetQueryToken(
				"",
				time.Time{},
			)
			var fromLookup base.BaseInterface
			if !tc.typeChanged {
				ghFromLookup := testLookup(t)
				ghFromLookup.data.CopyFrom(testData)
				fromLookup = ghFromLookup
			} else {
				otherLookup := *otherLookupTest
				fromLookup = &otherLookup
			}
			fromRequire := filtertest.Require(t, "hub")
			fromLookup.SetRequire(fromRequire)
			inheritableETag := "foo"
			inheritableAccessToken := "token-goes-here"
			if fl, ok := fromLookup.(*Lookup); ok {
				if err := fl.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
					t.Fatalf("Lookup.ApplyOverrides() failed: %v", err)
				}
				fl.data.etag = inheritableETag
				fl.AccessToken = inheritableAccessToken
			}
			lookup.data.etag = ""
			hadETag := lookup.data.ETag()
			hadReleases := decode.ToYAMLString(lookup.data.Releases(), "")
			wantQueryToken, wantValidUntil := fromRequire.Docker.GetAuth().GetQueryTokenSelf()

			lookup.AccessToken = tc.accessToken
			if tc.maskAccessToken {
				lookup.AccessToken = util.SecretValue
			}

			lookup.InheritSecrets(fromLookup, nil)

			wantETag, wantReleases := hadETag, hadReleases
			if tc.inheritData {
				wantETag = inheritableETag
				wantReleases = decode.ToYAMLString(testData.releases, "")
			}
			if got := lookup.data.ETag(); got != wantETag {
				t.Errorf("Data.ETag() mismatch\ngot:  %q\nwant: %q", got, wantETag)
			}
			if got := decode.ToYAMLString(lookup.data.Releases(), ""); got != wantReleases {
				t.Errorf("Data.Releases() mismatch\ngot:  %q\nwant: %q", got, wantReleases)
			}

			wantAccessToken := tc.accessToken
			if tc.maskAccessToken {
				wantAccessToken = ""
				if tc.inheritAccessToken {
					wantAccessToken = inheritableAccessToken
				}
			}
			if got := lookup.AccessToken; got != wantAccessToken {
				t.Errorf("Lookup.AccessToken mismatch\ngot:  %q\nwant: %q", got, wantAccessToken)
			}

			if tc.inheritRequire {
				switch {
				case lookup.Require == nil:
					t.Error("Lookup.Require mismatch\ngot:  nil\nwant: non-nil")
				case lookup.Require.Docker == nil:
					t.Error("Lookup.Require.Docker mismatch\ngot:  nil\nwant: non-nil")
				default:
					gotQueryToken, gotValidUntil := lookup.Require.Docker.GetAuth().GetQueryTokenSelf()
					checkDockerToken(
						t,
						wantQueryToken, gotQueryToken,
						wantValidUntil, gotValidUntil,
					)
				}
			} else if lookup.Require != nil && lookup.Require.Docker != nil {
				gotQueryToken, gotValidUntil := lookup.Require.Docker.GetAuth().GetQueryTokenSelf()
				wantQueryToken, wantValidUntil = "", time.Time{}
				checkDockerToken(
					t,
					wantQueryToken, gotQueryToken,
					wantValidUntil, gotValidUntil,
				)
			}
		})
	}
}
