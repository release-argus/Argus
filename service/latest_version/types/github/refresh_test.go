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
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	filtertest "github.com/release-argus/Argus/service/latest_version/filter/test"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	ghtypes "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	"github.com/release-argus/Argus/util"
)

func checkDockerToken(
	t *testing.T,
	wantQueryToken, gotQueryToken string,
	wantValidUntil, gotValidUntil time.Time,
	prefix, message string,
) {
	if gotQueryToken != wantQueryToken {
		t.Errorf(
			"%s\nRequire.Docker.queryToken %s\ngot:  %q\nwant: %q",
			prefix, message,
			gotQueryToken, wantQueryToken,
		)
	}
	if gotValidUntil != wantValidUntil {
		t.Errorf(
			"%s\nRequire.Docker.validUntil %s\ngot:  %q\nwant: %q",
			prefix, message,
			gotValidUntil, wantValidUntil,
		)
	}
}

func TestLookup_InheritSecrets(t *testing.T) {
	testData := newData(
		"etag",
		&[]ghtypes.Release{
			{URL: "foo"},
			{URL: "bar"},
		},
	)
	testData.SetTagFallback()

	// GIVEN: a Lookup and a Lookup to inherit from.
	tests := []struct {
		name               string
		typeChanged        bool
		overrides          string
		inheritData        bool
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
			inheritAccessToken: true,
			inheritRequire:     false,
		},
		{
			name:               "inherit Require, not Data",
			overrides:          `url: something-else`,
			inheritData:        false,
			inheritAccessToken: true,
			inheritRequire:     true,
		},
		{
			name:               "inherit all",
			inheritData:        true,
			inheritRequire:     true,
			inheritAccessToken: true,
		},
	}

	otherLookupTest := &struct {
		base.Lookup `yaml:",inline"`
		Data        *Data `yaml:"github_data"`
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, false)
			lookup.Require = filtertest.Require(t, "hub")
			lookup.Require.Docker.GetAuth().SetQueryToken(
				"",
				time.Time{},
			)
			var fromLookup base.BaseInterface
			if !tc.typeChanged {
				ghFromLookup := testLookup(t, false)
				ghFromLookup.data.CopyFrom(testData)
				fromLookup = ghFromLookup
			} else {
				otherLookup := *otherLookupTest
				fromLookup = &otherLookup
			}
			fromRequire := filtertest.Require(t, "hub")
			fromLookup.SetRequire(fromRequire)
			// overrides.
			if fromL, ok := fromLookup.(*Lookup); ok {
				if err := fromL.ApplyOverrides("yaml", []byte(tc.overrides)); err != nil {
					t.Fatalf(
						"%s\nfailed to unmarshal Lookup overrides: %v",
						packageName, err,
					)
				}
			}

			inheritableETag := "foo"
			if fl, ok := fromLookup.(*Lookup); ok {
				fl.data.eTag = inheritableETag
			}
			lookup.data.eTag = ""
			hadETag := lookup.data.ETag()
			wantQueryToken, wantValidUntil := fromRequire.Docker.GetAuth().GetQueryTokenSelf()

			inheritableAccessToken := "token-goes-here"
			if fl, ok := fromLookup.(*Lookup); ok {
				fl.AccessToken = inheritableAccessToken
			}
			if tc.inheritAccessToken {
				lookup.AccessToken = util.SecretValue
			}

			// WHEN: we call InheritSecrets.
			lookup.InheritSecrets(fromLookup, nil)

			prefix := fmt.Sprintf("%s\nLookup.InheritSecrets()", packageName)

			// THEN: the Data is copied when expected.
			gotETag := lookup.data.ETag()
			if tc.inheritData {
				if gotETag != inheritableETag {
					t.Errorf(
						"%s ETag not copied over\ngot:  %q\nwant:  %q",
						prefix, gotETag, inheritableETag,
					)
				}
				gotReleases := decode.ToYAMLString(lookup.data.Releases(), "")
				wantReleases := decode.ToYAMLString(testData.releases, "")
				if gotReleases != wantReleases {
					t.Errorf(
						"%s Releases not copied over\ngot:  %q\nwant:  %q",
						prefix, gotReleases, wantReleases,
					)
				}
			} else if gotETag != hadETag {
				t.Errorf(
					"%s Data shouldn't have been inherited\ngot:  %q\nwant: %q",
					packageName, gotETag, hadETag,
				)
			}

			// AND: the access token is copied when expected.
			if tc.inheritAccessToken {
				if got, want := lookup.AccessToken, inheritableAccessToken; got != want {
					t.Errorf(
						"%s AccessToken not copied over\ngot:  %q\nwant: %q",
						prefix, got, want,
					)
				}
			}

			// AND: the Require is copied when expected.
			if tc.inheritRequire {
				if lookup.Require == nil {
					t.Errorf("%s Require not copied over\ngot:  nil\nwant: non-nil", prefix)
				} else if lookup.Require.Docker == nil {
					t.Errorf("%s Require.Docker not copied over\ngot:  nil\nwant: non-nil", prefix)
				} else {
					gotQueryToken, gotValidUntil := lookup.Require.Docker.GetAuth().GetQueryTokenSelf()
					checkDockerToken(
						t,
						wantQueryToken, gotQueryToken,
						wantValidUntil, gotValidUntil,
						prefix,
						"not copied",
					)
				}
			} else if lookup.Require != nil && lookup.Require.Docker != nil {
				gotQueryToken, gotValidUntil := lookup.Require.Docker.GetAuth().GetQueryTokenSelf()
				wantQueryToken, wantValidUntil = "", time.Time{}
				checkDockerToken(
					t,
					wantQueryToken, gotQueryToken,
					wantValidUntil, gotValidUntil,
					prefix,
					"should not be copied",
				)
			}
		})
	}
}
