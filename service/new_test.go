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

//go:build integration

package service

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/notify/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	dv_web "github.com/release-argus/Argus/service/deployed_version/types/web"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	lv_github "github.com/release-argus/Argus/service/latest_version/types/github"
	lv_web "github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/webhook"
)

func TestService_GiveSecretsLatestVersion(t *testing.T) {
	type otherData struct {
		githubData            *lv_github.Data
		githubDataTransformed bool
	}

	// GIVEN a LatestVersion that may have secrets in it referencing those in another LatestVersion
	githubData := lv_github.Data{}
	githubData.SetETag("shazam")
	tests := map[string]struct {
		latestVersion latestver.Lookup
		otherLV       latestver.Lookup
		expected      latestver.Lookup
		otherData     otherData
	}{
		"nil oldLatestVersion": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			otherLV: nil,
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
		},
		"empty AccessToken": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						access_token: "foo"
					`),
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
		},
		"new AccessToken kept": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						access_token: "foo"
					`),
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						access_token: "bar"
					`),
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						access_token: "foo"
					`),
					nil,
					nil,
					nil, nil)
			}),
		},
		"give old AccessToken": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						access_token: "`+util.SecretValue+`"
					`),
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						access_token: "bar"
					`),
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						access_token: "bar"
					`),
					nil,
					nil,
					nil, nil)
			}),
		},
		"referencing default AccessToken": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						access_token: util.SecretValue
					`),
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
		},
		"nil Require": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
								token: foo
					`),
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
		},
		"empty Require": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
					`),
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
								token: foo
					`),
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
		},
		"new Require.Docker.Token kept": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
								token: foo
					`),
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
								token: bar
					`),
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
								token: foo
					`),
					nil,
					nil,
					nil, nil)
			}),
		},
		"give old Require.Docker.Token": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
								token: `+util.SecretValue+`
					`),
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
								token: bar
					`),
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
								token: bar
					`),
					nil,
					nil,
					nil, nil)
			}),
		},
		"referencing default Require.Docker.Token": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
								token: `+util.SecretValue+`
					`),
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
					`),
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", test.TrimYAML(`
						require:
							docker:
								type: ghcr
								token: `+util.SecretValue+`
					`),
					nil,
					nil,
					nil, nil)
			}),
		},
		"githubData carried over if type still 'github'": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			otherData: otherData{
				githubData:            &githubData,
				githubDataTransformed: true,
			},
		},
		"githubData not carried over if type wasn't 'github'": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			otherData: otherData{
				githubData:            &githubData,
				githubDataTransformed: false,
			},
		},
		"githubData not carried over if type no longer 'github'": {
			latestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"url",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			otherLV: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New("github",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			expected: test.IgnoreError(t, func() (latestver.Lookup, error) {
				return latestver.New(
					"url",
					"yaml", "",
					nil,
					nil,
					nil, nil)
			}),
			otherData: otherData{
				githubData:            &githubData,
				githubDataTransformed: false,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			newService := &Service{LatestVersion: tc.latestVersion}
			oldService := &Service{LatestVersion: tc.otherLV}

			// WHEN we call GiveSecrets
			newService.giveSecretsLatestVersion(oldService.LatestVersion)

			// THEN we should get a Service with the secrets from the other Service
			gotLV := newService.LatestVersion

			// Only GitHub types have AccessTokens
			if gotLatestVersion, ok := gotLV.(*lv_github.Lookup); ok {
				if hadLatestVersion, ok := tc.latestVersion.(*lv_github.Lookup); ok {
					gotAccessToken := gotLatestVersion.AccessToken
					expectedAccessToken := hadLatestVersion.AccessToken
					if gotAccessToken != expectedAccessToken {
						t.Errorf("Access Token\nwant: %q\ngot:  %q",
							expectedAccessToken, gotAccessToken)
					}
				}
			}

			// Require
			var gotRequire *filter.Require
			var expectedRequire *filter.Require
			// Got
			if gotLatestVersion, ok := gotLV.(*lv_github.Lookup); ok {
				gotRequire = gotLatestVersion.GetRequire()
			} else if gotLatestVersion, ok := gotLV.(*lv_web.Lookup); ok {
				gotRequire = gotLatestVersion.GetRequire()
			}
			// Expected
			if expectedLatestVersion, ok := tc.expected.(*lv_github.Lookup); ok {
				expectedRequire = expectedLatestVersion.GetRequire()
			} else if expectedLatestVersion, ok := tc.expected.(*lv_web.Lookup); ok {
				expectedRequire = expectedLatestVersion.GetRequire()
			}
			// newService has a nil Require, but expected non-nil
			if gotRequire == nil && expectedRequire != nil {
				t.Errorf("Expected Require to be non-nil, got nil")

				// newService Require/Docker isn't nil when expected is or vice versa
			} else if gotRequire != expectedRequire &&
				gotRequire.Docker != expectedRequire.Docker &&
				// newService doesn't have the expected Token
				gotRequire.Docker.Token != expectedRequire.Docker.Token {
				t.Errorf("Expected %q, got %q",
					expectedRequire.Docker.Token, gotRequire.Docker.Token)
			}

			// Data
			if expectedLatestVersion, ok := tc.expected.(*lv_github.Lookup); ok {
				// Ensure gotLV is a *lv_github.Lookup
				if gotLatestVersion, ok := gotLV.(*lv_github.Lookup); ok {
					got := gotLatestVersion.GetGitHubData().String()
					expected := expectedLatestVersion.GetGitHubData().String()
					if got != expected {
						t.Errorf("Expected githubData to be\n%v\ngot\n%v",
							expected, got)
					}
				} else {
					t.Fatalf("Expected *lv_github.Lookup, got %T", gotLV)
				}
			}
		})
	}
}

func TestService_GiveSecretsDeployedVersion(t *testing.T) {
	// GIVEN a DeployedVersion that may have secrets in it referencing those in another DeployedVersion
	tests := map[string]struct {
		deployedVersion, otherDV deployedver.Lookup
		secretRefs               shared.DVSecretRef
		expected                 deployedver.Lookup
	}{
		"nil DeployedVersion": {
			deployedVersion: nil,
			otherDV:         &dv_web.Lookup{},
			secretRefs:      shared.DVSecretRef{},
			expected:        nil,
		},
		"nil OldDeployedVersion": {
			deployedVersion: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: "foo"}},
			otherDV: nil,
			expected: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: "foo"}},
		},
		"keep BasicAuth.Password": {
			deployedVersion: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: "foo"}},
			otherDV: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: "bar"}},
			expected: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: "foo"}},
		},
		"give old BasicAuth.Password": {
			deployedVersion: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: util.SecretValue}},
			otherDV: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: "bar"}},
			expected: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: "bar"}},
		},
		"referencing default BasicAuth.Password": {
			deployedVersion: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: util.SecretValue}},
			otherDV: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{}},
			expected: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: ""}},
		},
		"referencing BasicAuth.Password that doesn't exist": {
			deployedVersion: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: util.SecretValue}},
			otherDV: &dv_web.Lookup{},
			expected: &dv_web.Lookup{
				BasicAuth: &dv_web.BasicAuth{
					Password: util.SecretValue}},
		},
		"empty Headers": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{}},
		},
		"only new Headers": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "bish", Value: "bash"}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "bish", Value: "bash"}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: nil},
					{OldIndex: nil}}},
		},
		"Headers with index out of range": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "bish", Value: util.SecretValue},
					{Key: "bash", Value: util.SecretValue}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "bish", Value: util.SecretValue},
					{Key: "bash", Value: util.SecretValue}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(0)},
					{OldIndex: test.IntPtr(1)}}},
		},
		"Headers with SecretValue but nil index refs": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "bish", Value: util.SecretValue},
					{Key: "bash", Value: util.SecretValue}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "bish", Value: "bash"},
					{Key: "bash", Value: "bop"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "bish", Value: util.SecretValue},
					{Key: "bash", Value: util.SecretValue}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: nil},
					{OldIndex: nil}}},
		},
		"only changed Headers": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "shazam"}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "shazam"}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(0)}}},
		},
		"only new/changed Headers": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(0)}, {OldIndex: nil}}},
		},
		"only new/changed Headers with expected refs": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(0)}, {OldIndex: nil}}},
		},
		"only new/changed Headers with no refs": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{}},
		},
		"referencing old Header value with no refs": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{}},
		},
		"only new/changed Headers with partial ref (not for all secrets)": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bang"},
					{Key: "bosh", Value: util.SecretValue}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bash"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bang"},
					{Key: "bosh", Value: util.SecretValue}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(0)}, {OldIndex: test.IntPtr(1)}}},
		},
		"referencing old Header value": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bash"}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(0)}, {OldIndex: nil}}},
		},
		"referencing old Header value that doesn't exist": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: util.SecretValue},
					{Key: "bish", Value: "bash"}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(1)}, {OldIndex: nil}}},
		},
		"referencing some old Header values but not others": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bang"},
					{Key: "bish", Value: util.SecretValue}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bong"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bang"},
					{Key: "bish", Value: "bong"}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: nil}, {OldIndex: test.IntPtr(1)}}},
		},
		"swap header values": {
			deployedVersion: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "bish", Value: util.SecretValue},
					{Key: "foo", Value: util.SecretValue}}},
			otherDV: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bong"}}},
			expected: &dv_web.Lookup{
				Headers: []dv_web.Header{
					{Key: "bish", Value: "bar"},
					{Key: "foo", Value: "bong"}}},
			secretRefs: shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(0)}, {OldIndex: test.IntPtr(1)}}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			newService := &Service{DeployedVersionLookup: tc.deployedVersion}
			oldService := &Service{DeployedVersionLookup: tc.otherDV}

			// WHEN we call giveSecretsDeployedVersion
			newService.giveSecretsDeployedVersion(oldService.DeployedVersionLookup, &tc.secretRefs)

			// THEN we should get a Service with the secrets from the other Service
			gotDV := newService.DeployedVersionLookup
			if gotDV == tc.expected {
				return
			}
			// Got/Expected nil but not both
			if gotDV == nil && tc.expected != nil ||
				gotDV != nil && tc.expected == nil {
				t.Errorf("Expected %q, got %q",
					tc.expected.String(tc.expected, ""), gotDV.String(gotDV, ""))
			}
			// BasicAuth
			var expectedBasicAuth *dv_web.BasicAuth
			if expectedLV, ok := tc.expected.(*dv_web.Lookup); ok {
				expectedBasicAuth = expectedLV.BasicAuth
			}
			var gotBasicAuth *dv_web.BasicAuth
			if gotLV, ok := gotDV.(*dv_web.Lookup); ok {
				gotBasicAuth = gotLV.BasicAuth
			}
			if expectedBasicAuth != gotBasicAuth {
				if expectedBasicAuth == nil && gotBasicAuth != nil {
					t.Errorf("Expected BasicAuth to be nil, got %q", *gotBasicAuth)
				} else if gotBasicAuth.Password != expectedBasicAuth.Password {
					t.Errorf("Expected %q, got %q",
						util.DereferenceOrDefault(expectedBasicAuth), util.DereferenceOrDefault(gotBasicAuth))
				}
			}
			// Headers
			var expectedHeaders []dv_web.Header
			if expectedLV, ok := tc.expected.(*dv_web.Lookup); ok {
				expectedHeaders = expectedLV.Headers
			}
			var gotHeaders []dv_web.Header
			if gotLV, ok := gotDV.(*dv_web.Lookup); ok {
				gotHeaders = gotLV.Headers
			}
			if len(gotHeaders) != len(expectedHeaders) {
				t.Errorf("Expected %q, got %q",
					expectedHeaders, gotHeaders)
			} else {
				for i, gotHeader := range gotHeaders {
					if gotHeader != expectedHeaders[i] {
						t.Errorf("Expected %q, got %q",
							expectedHeaders[i], gotHeader)
					}
				}
			}
		})
	}
}

func TestService_GiveSecretsNotify(t *testing.T) {
	// GIVEN a NotifySlice that may have secrets in it referencing those in another NotifySliceSlice
	tests := map[string]struct {
		notify, otherNotify shoutrrr.Slice
		secretRefs          map[string]shared.OldStringIndex
		expected            shoutrrr.Slice
	}{
		"nil NotifySlice": {
			notify: nil,
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{},
			expected:   nil,
		},
		"nil oldNotifies": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
			otherNotify: nil,
			secretRefs:  map[string]shared.OldStringIndex{},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
		},
		"nil secretRefs": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: nil,
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil)},
		},
		"no secretRefs": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil)},
		},
		"no matching secretRefs": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"bish": {OldIndex: "bash"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil)},
		},
		"secretRef referencing empty index": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "yikes"},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: ""}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "yikes"},
					nil,
					nil, nil, nil)},
		},
		"secretRef referencing index that doesn't exist": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "yikes"},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "baz"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "yikes"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - url_fields.altid": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "yikes"},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "yikes"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - url_fields.apikey": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "yikes"},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "yikes"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - url_fields.apikey swap vars": {
			notify: shoutrrr.Slice{
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil),
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "shazam"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{
				"bar": {OldIndex: "foo"},
				"foo": {OldIndex: "bar"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "shazam"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - url_fields.apikey swap vars ignores notify order": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": util.SecretValue},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "shazam"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{
				"bar": {OldIndex: "foo"},
				"foo": {OldIndex: "bar"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "shazam"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"apikey": "something"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - url_fields.botkey": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"botkey": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"botkey": "yikes"},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"botkey": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"botkey": "something"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"botkey": "yikes"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - url_fields.password": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"password": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"password": "yikes"},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"password": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"password": "something"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"password": "yikes"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - url_fields.token": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"token": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"token": "yikes"},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"token": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"token": "something"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"token": "yikes"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - url_fields.tokena": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"tokena": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"tokena": "yikes"},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"tokena": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"tokena": "something"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"tokena": "yikes"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - url_fields.tokenb": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"tokenb": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"tokenb": "yikes"},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"tokenb": "something"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"tokenb": "something"},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"tokenb": "yikes"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - url_fields.host ignored as SecretValue": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"host": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"host": "https://example.com"},
					nil,
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"host": "https://example.com/foo"},
					nil,
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"host": util.SecretValue},
					nil,
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "", nil,
					map[string]string{
						"host": "https://example.com"},
					nil,
					nil, nil, nil)},
		},
		"secretRefs - params.devices": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil, nil,
					map[string]string{
						"devices": util.SecretValue},
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "",
					nil, nil,
					map[string]string{
						"devices": "yikes"},
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					nil,
					map[string]string{
						"devices": "something"},
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					nil,
					map[string]string{
						"devices": "something"},
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "",
					nil,
					nil,
					map[string]string{
						"devices": "yikes"},
					nil, nil, nil)},
		},
		"secretRefs - params.avatar ignored as SecretValue": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					nil,
					map[string]string{
						"avatar": util.SecretValue},
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "",
					nil,
					nil,
					map[string]string{
						"avatar": "https://example.com"},
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					nil,
					map[string]string{
						"avatar": "https://example.com/fooo"},
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					nil,
					map[string]string{
						"avatar": util.SecretValue},
					nil, nil, nil),
				"bar": shoutrrr.New(
					nil, "", "",
					nil,
					nil,
					map[string]string{
						"avatar": "https://example.com"},
					nil, nil, nil)},
		},
		"secretRefs - ALL": {
			notify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					map[string]string{
						"altid":    util.SecretValue,
						"apikey":   util.SecretValue,
						"botkey":   util.SecretValue,
						"password": util.SecretValue,
						"token":    util.SecretValue,
						"tokena":   util.SecretValue,
						"tokenb":   util.SecretValue},
					map[string]string{
						"devices": util.SecretValue},
					nil, nil, nil)},
			otherNotify: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					map[string]string{
						"altid":    "whoosh",
						"apikey":   "foo",
						"botkey":   "bar",
						"password": "baz",
						"token":    "bish",
						"tokena":   "bosh",
						"tokenb":   "bash"},
					map[string]string{
						"devices": "id1,id2"},
					nil, nil, nil)},
			secretRefs: map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}},
			expected: shoutrrr.Slice{
				"foo": shoutrrr.New(
					nil, "", "",
					nil,
					map[string]string{
						"altid":    "whoosh",
						"apikey":   "foo",
						"botkey":   "bar",
						"password": "baz",
						"token":    "bish",
						"tokena":   "bosh",
						"tokenb":   "bash"},
					map[string]string{
						"devices": "id1,id2"},
					nil, nil, nil)},
		},
	}

	for name, tc := range tests {
		newService := &Service{Notify: tc.notify}
		newService.Status.Init(
			len(newService.Notify), len(newService.Command), len(newService.WebHook),
			&name, nil,
			nil)
		// Give empty defaults and hardDefaults to the NotifySlice
		newService.Notify.Init(
			&newService.Status,
			&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
		)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN we call giveSecretsNotify
			newService.giveSecretsNotify(tc.otherNotify, tc.secretRefs)

			// THEN we should get a NotifySlice with the secrets from the other Service
			gotNotify := newService.Notify
			gotNotifyStr := gotNotify.String("")
			expectedStr := tc.expected.String("")
			if gotNotifyStr != expectedStr {
				t.Errorf("secrets not passed over\nwant:\n%v\n\ngot:\n%v",
					expectedStr, gotNotifyStr)
			}
		})
	}
}

func TestService_GiveSecretsWebHook(t *testing.T) {
	// GIVEN a WebHookSlice that may have secrets in it referencing those in another WebHookSliceSlice
	tests := map[string]struct {
		webhook, otherWebhook webhook.Slice
		secretRefs            map[string]shared.WHSecretRef
		expected              webhook.Slice
	}{
		"nil WebHookSlice": {
			webhook: nil,
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{},
			expected:   nil,
		},
		"nil otherWebHook": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
			otherWebhook: nil,
			secretRefs:   map[string]shared.WHSecretRef{},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
		},
		"nil secretRefs": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
			secretRefs: nil,
			expected: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil)},
		},
		"no secretRefs": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil)},
		},
		"no matching secretRefs": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"bish": {OldIndex: "bash"}},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil)},
		},
		"secretRef referencing empty index": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"whoosh",
					nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {OldIndex: ""},
				"bar": {OldIndex: ""}},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"whoosh",
					nil, "", "", nil, nil, nil)},
		},
		"secretRef referencing index that doesn't exist": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"whoosh",
					nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {OldIndex: "bash"},
				"bar": {OldIndex: ""}},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"whoosh",
					nil, "", "", nil, nil, nil)},
		},
		"secretRefs - secret": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"whoosh",
					nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {OldIndex: "foo"},
				"bar": {OldIndex: ""}},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"whoosh",
					nil, "", "", nil, nil, nil)},
		},
		"secretRefs - secret swap vars": {
			webhook: webhook.Slice{
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil),
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"whoosh",
					nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"bar": {OldIndex: "foo"},
				"foo": {OldIndex: "bar"}},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"whoosh",
					nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
		},
		"secretRefs - secret swap vars ignores order sent": {
			webhook: webhook.Slice{
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil),
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					util.SecretValue,
					nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"whoosh",
					nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"bar": {OldIndex: "foo"},
				"foo": {OldIndex: "bar"}},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"whoosh",
					nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil, nil, "", nil, nil, nil, nil, nil,
					"shazam",
					nil, "", "", nil, nil, nil)},
		},
		"custom headers - no secretRefs": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "baz"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bing"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bang"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "baz"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
		},
		"custom headers - no header secretRefs": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "baz"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bing"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bang"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {OldIndex: "foo"},
				"bar": {OldIndex: "bar"}},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "baz"},
					},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
		},
		"custom headers - header secretRefs but old secrets unwanted": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bar"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "baz"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bing"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bang"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {
					OldIndex: "foo",
					CustomHeaders: []shared.OldIntIndex{
						{OldIndex: test.IntPtr(0)}}},
				"bar": {
					OldIndex: "bar",
					CustomHeaders: []shared.OldIntIndex{
						{OldIndex: test.IntPtr(0)}}}},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bar"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "baz"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
		},
		"custom headers - header secretRefs, some indices out of range": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bish", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bang", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bash"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {
					OldIndex: "foo",
					CustomHeaders: []shared.OldIntIndex{
						{OldIndex: test.IntPtr(5)}, {OldIndex: test.IntPtr(1)}}},
				"bar": {
					OldIndex: "bar",
					CustomHeaders: []shared.OldIntIndex{
						{OldIndex: test.IntPtr(0)}, {OldIndex: test.IntPtr(2)}}}},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bish", Value: "bash"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
		},
		"custom headers - header secretRefs use all secrets": {
			webhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bish", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bang", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bash"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"foo": {
					OldIndex: "foo",
					CustomHeaders: []shared.OldIntIndex{
						{OldIndex: test.IntPtr(0)},
						{OldIndex: test.IntPtr(1)}}},
				"bar": {
					OldIndex: "bar",
					CustomHeaders: []shared.OldIntIndex{
						{OldIndex: test.IntPtr(0)},
						{OldIndex: test.IntPtr(1)}}}},
			expected: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bash"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
		},
		"custom headers - header secretRefs, swap names of webhook": {
			webhook: webhook.Slice{
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bish", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: util.SecretValue},
						{Key: "bang", Value: util.SecretValue}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			otherWebhook: webhook.Slice{
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bash"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
			secretRefs: map[string]shared.WHSecretRef{
				"bar": {
					OldIndex: "foo",
					CustomHeaders: []shared.OldIntIndex{
						{OldIndex: test.IntPtr(0)},
						{OldIndex: test.IntPtr(1)}}},
				"foo": {
					OldIndex: "bar",
					CustomHeaders: []shared.OldIntIndex{
						{OldIndex: test.IntPtr(0)},
						{OldIndex: test.IntPtr(1)}}}},
			expected: webhook.Slice{
				"bar": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bash"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil),
				"foo": webhook.New(
					nil,
					&webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"}},
					"", nil, nil, nil, nil, nil, "", nil, "", "", nil, nil, nil)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			newService := &Service{
				ID: name,
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				WebHook: tc.webhook}
			// New Service Status.Fails
			newService.Status.Init(
				len(newService.Notify), len(newService.Command), len(newService.WebHook),
				&newService.ID, nil,
				nil)
			newService.Init(
				&Defaults{}, &Defaults{},
				&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
				&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{},
			)
			// Other Service Status.Fails
			if tc.otherWebhook != nil {
				otherServiceStatus := status.Status{}
				otherServiceStatus.Init(
					len(tc.otherWebhook), 0, 0,
					test.StringPtr("otherService"), nil,
					nil)
				tc.otherWebhook.Init(
					&otherServiceStatus,
					&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{},
					nil,
					test.StringPtr("10m"))
			}

			// WHEN we call giveSecretsWebHook
			newService.giveSecretsWebHook(tc.otherWebhook, tc.secretRefs)

			// THEN we should get a WebHookSlice with the secrets from the other Service
			gotWebHook := newService.WebHook
			if gotWebHook.String() != tc.expected.String() {
				t.Errorf("Want:\n%v\n\nGot:\n%v",
					tc.expected, gotWebHook)
			}
		})
	}
}

func TestService_GiveSecrets(t *testing.T) {
	type statusTests struct {
		oldLatestVersion, expectedLatestVersion                       string
		oldLatestVersionTimestamp, expectedLatestVersionTimestamp     string
		oldDeployedVersion, expectedDeployedVersion                   string
		oldDeployedVersionTimestamp, expectedDeployedVersionTimestamp string
	}
	type commandTests struct {
		oldFails      []*bool
		expectedFails []*bool
	}
	type webhookTests struct {
		oldFails      map[string]*bool
		expectedFails map[string]*bool
	}

	// GIVEN a Service that may have secrets in it referencing those in another Service
	tests := map[string]struct {
		svc          *Service
		oldService   *Service
		statusTests  statusTests
		commandTests commandTests
		webhookTests webhookTests
		secretRefs   oldSecretRefs
		expected     *Service
	}{
		"no secrets": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: something
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: user
								password: pass
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "",
						nil,
						map[string]string{
							"apikey": "saucy"},
						nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil,
						nil,
						map[string]string{
							"avatar": "https://example.com"},
						nil, nil, nil),
				},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"bar",
						nil, "",
						"http://foo.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"foo",
						nil, "",
						"http://bar.com",
						nil, nil, nil),
				},
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: somethingElse
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: username
								password: password
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "", nil,
						nil,
						map[string]string{
							"apikey": "sweet"},
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil, nil,
						map[string]string{
							"avatar": "https://example.com/logo.png"},
						nil, nil, nil)},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"bar",
						nil, "",
						"http://bar.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"foo",
						nil, "",
						"http://foo.com",
						nil, nil, nil)},
			},
			secretRefs: oldSecretRefs{},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: something
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: user
								password: pass
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "", nil,
						map[string]string{
							"apikey": "saucy"},
						nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil, nil,
						map[string]string{
							"avatar": "https://example.com"},
						nil, nil, nil),
				},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"bar",
						nil, "",
						"http://foo.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"foo",
						nil, "",
						"http://bar.com",
						nil, nil, nil),
				},
			},
		},
		"minimal CREATE": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
			},
			oldService: nil,
			secretRefs: oldSecretRefs{},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
			},
		},
		"no oldService (CREATE)": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: `+util.SecretValue+`
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: `+util.SecretValue+`
								password: `+util.SecretValue+`
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "",
						nil,
						map[string]string{
							"apikey": util.SecretValue},
						nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil, nil,
						map[string]string{
							"avatar": util.SecretValue},
						nil, nil, nil),
				},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						util.SecretValue,
						nil, "",
						"http://foo.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						util.SecretValue,
						nil, "",
						"http://bar.com",
						nil, nil, nil),
				},
			},
			oldService: nil,
			secretRefs: oldSecretRefs{},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: `+util.SecretValue+`
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: `+util.SecretValue+`
								password: `+util.SecretValue+`
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "", nil,
						map[string]string{
							"apikey": util.SecretValue},
						nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil, nil,
						map[string]string{
							"avatar": util.SecretValue},
						nil, nil, nil)},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						util.SecretValue,
						nil, "",
						"http://foo.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						util.SecretValue,
						nil, "",
						"http://bar.com",
						nil, nil, nil)},
			},
		},
		"no secretRefs": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: `+util.SecretValue+`
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: `+util.SecretValue+`
								password: `+util.SecretValue+`
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "", nil,
						map[string]string{
							"apikey": util.SecretValue},
						nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil,
						nil,
						map[string]string{
							"avatar": util.SecretValue},
						nil, nil, nil),
				},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						util.SecretValue,
						nil, "",
						"http://foo.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						util.SecretValue,
						nil, "",
						"http://bar.com",
						nil, nil, nil),
				},
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: somethingElse
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: username
								password: password
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "", nil,
						map[string]string{
							"apikey": "sweet"},
						nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil, nil,
						map[string]string{
							"avatar": "https://example.com/logo.png"},
						nil, nil, nil)},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"foo",
						nil, "",
						"http://foo.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"bar",
						nil, "",
						"http://bar.com",
						nil, nil, nil)},
			},
			webhookTests: webhookTests{
				oldFails: map[string]*bool{
					"foo": test.BoolPtr(false),
					"bar": test.BoolPtr(true)},
			},
			secretRefs: oldSecretRefs{},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: somethingElse
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: `+util.SecretValue+`
								password: password
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "", nil,
						map[string]string{
							"apikey": util.SecretValue},
						nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil, nil,
						map[string]string{
							"avatar": util.SecretValue},
						nil, nil, nil)},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						util.SecretValue,
						nil, "",
						"http://foo.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						util.SecretValue,
						nil, "",
						"http://bar.com",
						nil, nil, nil)},
			},
		},
		"matching secretRefs": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: somethingElse
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: `+util.SecretValue+`
								password: password
							headers:
								- key: X-Foo
									value: `+util.SecretValue+`
								- key: X-Bar
									value: `+util.SecretValue+`
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "",
						nil,
						map[string]string{
							"apikey": util.SecretValue},
						nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil, nil,
						map[string]string{
							"avatar": util.SecretValue},
						nil, nil, nil),
				},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						util.SecretValue,
						nil, "",
						"http://foo.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						util.SecretValue,
						nil, "",
						"http://bar.com",
						nil, nil, nil),
				},
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: somethingElse
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: username
								password: password
							headers:
								- key: X-Foo
									value: foo
								- key: X-Bar
									value: bar
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "",
						nil,
						map[string]string{
							"apikey": "sweet"},
						nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil, nil,
						map[string]string{
							"avatar": "https://example.com/logo.png"},
						nil, nil, nil),
				},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"foo",
						nil, "",
						"http://foo.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"bar",
						nil, "",
						"http://bar.com",
						nil, nil, nil),
				},
			},
			secretRefs: oldSecretRefs{
				DeployedVersionLookup: shared.DVSecretRef{Headers: []shared.OldIntIndex{{OldIndex: test.IntPtr(0)}, {OldIndex: test.IntPtr(1)}}},
				Notify:                map[string]shared.OldStringIndex{"foo": {OldIndex: "foo"}, "bar": {OldIndex: "bar"}},
				WebHook:               map[string]shared.WHSecretRef{"foo": {OldIndex: "foo"}, "bar": {OldIndex: "bar"}},
			},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: somethingElse
						`),
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								username: `+util.SecretValue+`
								password: password
							headers:
								- key: X-Foo
									value: foo
								- key: X-Bar
									value: bar
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "", "",
						nil,
						map[string]string{
							"apikey": "sweet"},
						nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "", "",
						nil, nil,
						map[string]string{
							"avatar": util.SecretValue},
						nil, nil, nil),
				},
				WebHook: webhook.Slice{
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"foo",
						nil, "",
						"http://foo.com",
						nil, nil, nil),
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil,
						"bar",
						nil, "",
						"http://bar.com",
						nil, nil, nil),
				},
			},
		},
		"unchanged LatestVersion.URL retains Status.LatestVersion": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
			statusTests: statusTests{
				oldLatestVersion:               "1.2.3",
				expectedLatestVersion:          "1.2.3",
				oldLatestVersionTimestamp:      time.Now().Format(time.RFC3339),
				expectedLatestVersionTimestamp: time.Now().Format(time.RFC3339),
			},
			secretRefs: oldSecretRefs{},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
		},
		"changed LatestVersion.URL loses Status.LatestVersion": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"url",
						"yaml", `url: https://example.com`,
						nil,
						nil,
						nil, nil)
				}),
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
			statusTests: statusTests{
				oldLatestVersion:          "1.2.3",
				oldLatestVersionTimestamp: time.Now().Format(time.RFC3339),
			},
			secretRefs: oldSecretRefs{},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
		},
		"unchanged DeployedVersion.URL retains Status.DeployedVersion": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
			statusTests: statusTests{
				oldDeployedVersion:               "1.2.3",
				expectedDeployedVersion:          "1.2.3",
				oldDeployedVersionTimestamp:      time.Now().Format(time.RFC3339),
				expectedDeployedVersionTimestamp: time.Now().Format(time.RFC3339),
			},
			secretRefs: oldSecretRefs{},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
		},
		"changed DeployedVersion.URL loses Status.DeployedVersion": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com/somewhere-else
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
			statusTests: statusTests{
				oldDeployedVersion:          "1.2.3",
				oldDeployedVersionTimestamp: time.Now().Format(time.RFC3339),
			},
			secretRefs: oldSecretRefs{},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							url: https://example.com
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
		},
		"unchanged WebHook retains Failed": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				WebHook: webhook.Slice{
					"test": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil, "",
						"http://example.com",
						nil, nil, nil),
				},
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				WebHook: webhook.Slice{
					"test": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil, "",
						"http://example.com",
						nil, nil, nil),
				},
			},
			secretRefs: oldSecretRefs{
				WebHook: map[string]shared.WHSecretRef{
					"test": {OldIndex: "test"}},
			},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				WebHook: webhook.Slice{
					"test": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil, "",
						"http://example.com",
						nil, nil, nil),
				},
			},
			webhookTests: webhookTests{
				oldFails: map[string]*bool{
					"test": test.BoolPtr(true)},
				expectedFails: map[string]*bool{
					"test": test.BoolPtr(true)},
			},
		},
		"changed WebHook loses Failed": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				WebHook: webhook.Slice{
					"test": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil, "",
						"http://example.com/other",
						nil, nil, nil)},
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				WebHook: webhook.Slice{
					"test": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil, "",
						"http://example.com",
						nil, nil, nil)},
			},
			secretRefs: oldSecretRefs{},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				WebHook: webhook.Slice{
					"test": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil, "",
						"http://example.com/other",
						nil, nil, nil)},
			},
			webhookTests: webhookTests{
				oldFails: map[string]*bool{
					"test": test.BoolPtr(true)},
			},
		},
		"unchanged Command retains Failed": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				Command: command.Slice{
					{"ls", "-la"}},
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				Command: command.Slice{
					{"ls", "-la"}},
				CommandController: &command.Controller{},
			},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				Command: command.Slice{
					{"ls", "-la"}},
				CommandController: &command.Controller{},
			},
			secretRefs: oldSecretRefs{},
			commandTests: commandTests{
				oldFails: []*bool{
					test.BoolPtr(true)},
				expectedFails: []*bool{
					test.BoolPtr(true)},
			},
		},
		"changed Command loses Failed": {
			svc: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				Command: command.Slice{
					{"ls", "-lah"}},
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				Command: command.Slice{
					{"ls", "-la"}},
				CommandController: &command.Controller{},
			},
			secretRefs: oldSecretRefs{},
			expected: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						nil, nil)
				}),
				Command: command.Slice{
					{"ls", "-lah"}},
			},
			commandTests: commandTests{
				oldFails: []*bool{
					test.BoolPtr(true)},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.svc.Init(
				&Defaults{}, &Defaults{},
				&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
				&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{},
			)
			tc.expected.Init(
				&Defaults{}, &Defaults{},
				&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
				&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{},
			)
			if tc.expected != nil {
				for k, v := range tc.commandTests.expectedFails {
					if v != nil {
						tc.expected.Status.Fails.Command.Set(k, *v)
					}
				}
				for k, v := range tc.webhookTests.expectedFails {
					tc.expected.Status.Fails.WebHook.Set(k, v)
				}
			}
			if tc.oldService != nil {
				tc.oldService.Init(
					&Defaults{}, &Defaults{},
					&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
					&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{},
				)
				if tc.oldService.Command != nil {
					tc.oldService.CommandController.Command = &tc.oldService.Command
				}
				for k, v := range tc.commandTests.oldFails {
					if v != nil {
						tc.oldService.Status.Fails.Command.Set(k, *v)
					}
				}
				for k, v := range tc.webhookTests.oldFails {
					tc.oldService.Status.Fails.WebHook.Set(k, v)
				}
			}

			// WHEN we call giveSecrets
			tc.svc.giveSecrets(tc.oldService, tc.secretRefs)

			// THEN we should get a Service with the secrets from the old Service
			gotService := tc.svc
			gotServiceStr := gotService.String("")
			expectedStr := tc.expected.String("")
			if gotServiceStr != expectedStr {
				t.Errorf("secrets weren't passed on with giveSecrets()\n%q\n\nGot:\n%q",
					expectedStr, gotServiceStr)
			}

			if gotService.WebHook != nil {
				var expectedWH string
				for name := range gotService.WebHook {
					expectedWH = name
					break
				}
				// Expecting Failed to be carried over
				// Get failed state being copied
				var wantFailed *bool
				for name, wh := range tc.expected.WebHook {
					wantFailed = wh.Failed.Get(name)
					break
				}
				// Get carried over state
				gotFailed := gotService.WebHook[expectedWH].Failed.Get(expectedWH)
				if gotFailed == wantFailed {
					return
				}
				if gotFailed == nil || wantFailed == nil {
					t.Errorf("Want: %v, got: %v",
						wantFailed, gotFailed)
				} else if *gotFailed != *wantFailed {
					t.Errorf("Want: %t, got: %t",
						*wantFailed, *gotFailed)
				}
			}
		})
	}
}

func TestService_CheckFetches(t *testing.T) {
	// GIVEN a Service
	testLV := testLatestVersion(t, "url", false)
	testLV.Query(false, logutil.LogFrom{})
	testDVL := testDeployedVersionLookup(t, false)
	testDVL.Query(false, logutil.LogFrom{})
	tests := map[string]struct {
		svc                                       *Service
		startLatestVersion, wantLatestVersion     string
		startDeployedVersion, wantDeployedVersion string
		errRegex                                  string
	}{
		"Already have LatestVersion, nil DeployedVersionLookup": {
			svc: &Service{
				LatestVersion:         testLatestVersion(t, "url", false),
				DeployedVersionLookup: nil},
			startLatestVersion:   "foo",
			wantLatestVersion:    testLV.GetStatus().LatestVersion(),
			startDeployedVersion: "bar",
			wantDeployedVersion:  "bar",
			errRegex:             `^$`,
		},
		"Already have LatestVersion and DeployedVersionLookup": {
			svc: &Service{
				LatestVersion:         testLatestVersion(t, "url", false),
				DeployedVersionLookup: testDeployedVersionLookup(t, false)},
			startLatestVersion:   "foo",
			wantLatestVersion:    testLV.GetStatus().LatestVersion(),
			wantDeployedVersion:  testDVL.GetStatus().DeployedVersion(),
			startDeployedVersion: "bar",
			errRegex:             `^$`,
		},
		"latest_version query fails": {
			svc: &Service{
				LatestVersion:         testLatestVersion(t, "url", true),
				DeployedVersionLookup: testDeployedVersionLookup(t, false)},
			errRegex: `latest_version - x509 \(certificate invalid\)`,
		},
		"deployed_version query fails": {
			svc: &Service{
				LatestVersion:         testLatestVersion(t, "url", false),
				DeployedVersionLookup: testDeployedVersionLookup(t, true)},
			wantLatestVersion: testLV.GetStatus().LatestVersion(),
			errRegex:          `deployed_version - x509 \(certificate invalid\)`,
		},
		"both queried": {
			svc: &Service{
				LatestVersion:         testLatestVersion(t, "url", false),
				DeployedVersionLookup: testDeployedVersionLookup(t, false)},
			wantLatestVersion:   testLV.GetStatus().LatestVersion(),
			wantDeployedVersion: testDVL.GetStatus().DeployedVersion(),
			errRegex:            `^$`,
		},
		"inactive queries neither": {
			svc: &Service{
				Options: *opt.New(
					test.BoolPtr(false), // active
					"", nil,
					nil, nil),
				LatestVersion:         testLatestVersion(t, "url", false),
				DeployedVersionLookup: testDeployedVersionLookup(t, false)},
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hardDefaults := Defaults{}
			hardDefaults.Default()
			tc.svc.Init(
				&Defaults{}, &hardDefaults,
				&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
				&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{},
			)
			announceChannel := make(chan []byte, 5)
			tc.svc.Status.AnnounceChannel = &announceChannel
			tc.svc.Status.SetLatestVersion(tc.startLatestVersion, "", false)
			tc.svc.Status.SetDeployedVersion(tc.startDeployedVersion, "", false)

			// WHEN we call CheckFetches
			err := tc.svc.CheckFetches()

			// THEN we get the err we expect
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			// AND we get the expected LatestVersion
			if tc.svc.Status.LatestVersion() != tc.wantLatestVersion {
				t.Errorf("LatestVersion\nWant: %q, got: %q",
					tc.wantLatestVersion, tc.svc.Status.LatestVersion())
			}
			// AND we get the expected DeployedVersion
			if tc.svc.Status.DeployedVersion() != tc.wantDeployedVersion {
				t.Errorf("DeployedVersion\nWant: %q, got: %q",
					tc.wantDeployedVersion, tc.svc.Status.DeployedVersion())
			}
			if len(*tc.svc.Status.AnnounceChannel) != 0 {
				t.Errorf("AnnounceChannel should be empty, got %d",
					len(*tc.svc.Status.AnnounceChannel))
			}
		})
	}
}

func TestRemoveDefaults(t *testing.T) {
	// GIVEN a Service, old Service and defaults
	tests := map[string]struct {
		svc                     *Service
		wasUsingNotifyDefaults  bool
		wasUsingCommandDefaults bool
		wasUsingDefaults        bool
		d                       *Defaults
		want                    *Service
	}{
		"No defaults being used": {
			svc: &Service{
				Comment: "foo",
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "foo", "gotify",
						nil, nil, nil,
						nil, nil, nil)},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: webhook.Slice{
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil)}},
			wasUsingNotifyDefaults:  false,
			wasUsingCommandDefaults: false,
			wasUsingDefaults:        false,
			d: &Defaults{
				Notify: map[string]struct{}{
					"bish": {}},
				Command: command.Slice{{"ls", "-la"}},
				WebHook: map[string]struct{}{
					"bash": {}}},
			want: &Service{
				Comment: "foo",
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "foo", "gotify",
						nil, nil, nil,
						nil, nil, nil)},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: webhook.Slice{
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil)}},
		},
		"All from defaults": {
			svc: &Service{
				Comment: "foo",
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "foo", "gotify",
						nil, nil, nil,
						nil, nil, nil),
					"gotify": shoutrrr.New(
						nil, "bar", "gotify",
						nil, nil, nil,
						nil, nil, nil)},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: webhook.Slice{
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil)}},
			wasUsingNotifyDefaults:  true,
			wasUsingCommandDefaults: true,
			wasUsingDefaults:        true,
			d: &Defaults{
				Notify: map[string]struct{}{
					"foo":    {},
					"gotify": {}},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: map[string]struct{}{
					"bar": {}}},
			want: &Service{
				Comment: "foo"},
		},
		"Notify default changed": {
			svc: &Service{
				Comment: "foo",
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "foo",
						"gotify",
						map[string]string{
							"message": "bar"},
						nil,
						nil, nil, nil, nil)},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: webhook.Slice{
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil)}},
			wasUsingNotifyDefaults:  true,
			wasUsingCommandDefaults: true,
			wasUsingDefaults:        true,
			d: &Defaults{
				Notify: map[string]struct{}{
					"foo": {}},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: map[string]struct{}{
					"bar": {}}},
			want: &Service{
				Comment: "foo",
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "foo",
						"gotify",
						map[string]string{
							"message": "bar"},
						nil,
						nil, nil, nil, nil)}},
		},
		"WebHook default changed": {
			svc: &Service{
				Comment: "foo",
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "foo",
						"gotify",
						nil, nil, nil,
						nil, nil, nil)},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: webhook.Slice{
					"bar": webhook.New(
						nil, nil,
						"1s",
						nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil)}},
			wasUsingNotifyDefaults:  true,
			wasUsingCommandDefaults: true,
			wasUsingDefaults:        true,
			d: &Defaults{
				Notify: map[string]struct{}{
					"foo": {}},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: map[string]struct{}{
					"bar": {}}},
			want: &Service{
				Comment: "foo",
				WebHook: webhook.Slice{
					"bar": webhook.New(
						nil, nil,
						"1s",
						nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil)}},
		},
		"Command default changed": {
			svc: &Service{
				Comment: "foo",
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "foo", "gotify",
						nil, nil, nil,
						nil, nil, nil)},
				Command: command.Slice{{"rm", "-rf", "foo.txt"}},
				WebHook: webhook.Slice{
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil)}},
			wasUsingNotifyDefaults:  true,
			wasUsingCommandDefaults: true,
			wasUsingDefaults:        true,
			d: &Defaults{
				Notify: map[string]struct{}{
					"foo": {}},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: map[string]struct{}{
					"bar": {}}},
			want: &Service{
				Comment: "foo",
				Command: command.Slice{{"rm", "-rf", "foo.txt"}}},
		},
		"defaults overridden by changing size of slice": {
			svc: &Service{
				Comment: "foo",
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "foo", "gotify",
						nil, nil, nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "bar", "gotify",
						nil, nil, nil,
						nil, nil, nil)},
				Command: command.Slice{{"ls", "-lah"}, {"rm", "-rf", "foo.txt"}},
				WebHook: webhook.Slice{
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil),
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil)}},
			wasUsingNotifyDefaults:  true,
			wasUsingCommandDefaults: true,
			wasUsingDefaults:        true,
			d: &Defaults{
				Notify: map[string]struct{}{
					"foo": {}},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: map[string]struct{}{
					"bar": {}}},
			want: &Service{
				Comment: "foo",
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "foo", "gotify",
						nil, nil, nil,
						nil, nil, nil),
					"bar": shoutrrr.New(
						nil, "bar", "gotify",
						nil, nil, nil,
						nil, nil, nil)},
				Command: command.Slice{{"ls", "-lah"}, {"rm", "-rf", "foo.txt"}},
				WebHook: webhook.Slice{
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil),
					"foo": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil)}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			oldService := Service{
				Notify: shoutrrr.Slice{
					"foo": shoutrrr.New(
						nil, "foo", "gotify",
						nil, nil, nil,
						nil, nil, nil)},
				Command: command.Slice{{"ls", "-lah"}},
				WebHook: webhook.Slice{
					"bar": webhook.New(
						nil, nil, "", nil, nil, nil, nil, nil, "", nil,
						"github",
						"", nil, nil, nil)}}
			oldService.notifyFromDefaults = tc.wasUsingNotifyDefaults
			oldService.commandFromDefaults = tc.wasUsingCommandDefaults
			oldService.webhookFromDefaults = tc.wasUsingDefaults

			// WHEN we call RemoveDefaults
			removeDefaults(&oldService, tc.svc, tc.d)

			// THEN we get the expected Service
			if tc.want.String("") != tc.svc.String("") {
				t.Errorf("\nwant: %q\ngot:  %q",
					tc.want.String(""), tc.svc.String(""))
			}
		})
	}
}

func TestFromPayload_ReadFromFail(t *testing.T) {
	// GIVEN an invalid payload
	payloadStr := "this is a long payload"
	payload := io.NopCloser(bytes.NewReader([]byte(payloadStr)))
	payload = http.MaxBytesReader(nil, payload, 5)

	// WHEN we call New
	_, err := FromPayload(
		&Service{},
		&payload,
		&Defaults{}, &Defaults{},
		&shoutrrr.SliceDefaults{},
		&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
		&webhook.SliceDefaults{},
		&webhook.Defaults{}, &webhook.Defaults{},
		logutil.LogFrom{},
	)

	// THEN we should get an error
	if err == nil {
		t.Errorf("Want error, got nil")
	}
}

func TestFromPayload(t *testing.T) {
	// GIVEN a payload and the Service defaults
	tests := map[string]struct {
		oldService *Service
		payload    string

		serviceDefaults, serviceHardDefaults *Defaults

		notifyGlobals, notifyDefaults, notifyHardDefaults *shoutrrr.SliceDefaults

		webhookGlobals                       *webhook.SliceDefaults
		webhookDefaults, webhookHardDefaults *webhook.Defaults

		want     *Service
		errRegex string
	}{
		"empty payload": {
			payload:  "",
			errRegex: `^EOF$`,
		},
		"invalid payload": {
			payload:  strings.Repeat("a", 1048577),
			errRegex: `^invalid character 'a' looking for beginning of value$`,
		},
		"invalid Service payload": {
			payload:  `{"name": false}`,
			errRegex: `json: cannot unmarshal bool into Go struct field [^ ]+ of type string$`,
		},
		"invalid SecretRefs payload": {
			payload: `{
				"webhook": {
					"foo": {
						"oldIndex": false}}}`,
			errRegex: `json: cannot unmarshal bool into Go struct field [^ ]+ of type string`,
		},
		"active True becomes nil": {
			payload: `{
				"options": {
					"active": true}}`,
			// Defaults as otherwise everything will be zero, so won't print
			want: &Service{
				Dashboard: DashboardOptions{Defaults: &DashboardOptionsDefaults{}},
				Options: opt.Options{
					Active:   nil,
					Defaults: &opt.Defaults{}},
			},
			errRegex: `^$`,
		},
		"active nil stays nil": {
			payload: `{
				"options": {
					"active": null}}`,
			// Defaults as otherwise everything will be zero, so won't print
			want: &Service{
				Dashboard: DashboardOptions{Defaults: &DashboardOptionsDefaults{}},
				Options: opt.Options{
					Active:   nil,
					Defaults: &opt.Defaults{}},
			},
			errRegex: `^$`,
		},
		"active False stays false": {
			payload: `{
				"options": {
					"active": false}}`,
			// Defaults as otherwise everything will be zero, so won't print
			want: &Service{
				Dashboard: DashboardOptions{Defaults: &DashboardOptionsDefaults{}},
				Options: opt.Options{
					Active:   test.BoolPtr(false),
					Defaults: &opt.Defaults{}},
			},
			errRegex: `^$`,
		},
		"Require.Docker removed if no Image&Tag": {
			payload: `{
				"latest_version": {
					"type": "github",
					"require": {
						"docker": {
							"type": "ghcr"}}}}`,
			want: &Service{
				Options:   opt.Options{Defaults: &opt.Defaults{}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptionsDefaults{}},
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", "",
						nil,
						nil,
						&latestver_base.Defaults{}, &latestver_base.Defaults{})
				})},
			errRegex: `^$`,
		},
		"Require.Docker stays if have Type&Image&Tag": {
			payload: `{
				"latest_version": {
					"type": "github",
					"require": {
						"docker": {
							"type": "ghcr",
							"image": "release-argus-argus",
							"tag": "latest"}}}}`,
			want: &Service{
				Options: opt.Options{Defaults: &opt.Defaults{}},
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							require:
								docker:
									type: ghcr
									image: release-argus-argus
									tag: latest
						`),
						nil,
						nil,
						&latestver_base.Defaults{}, &latestver_base.Defaults{})
				}),
				Dashboard: DashboardOptions{Defaults: &DashboardOptionsDefaults{}},
			},
			errRegex: `^$`,
		},
		"Give LatestVersion secrets": {
			payload: `{
				"latest_version": {
					"access_token": "` + util.SecretValue + `",
					"require": {
						"docker": {
							"type": "ghcr",
							"image": "release-argus/argus",
							"tag": "{{ version }}",
							"token": "` + util.SecretValue + `"}}}}`,
			serviceHardDefaults: &Defaults{
				LatestVersion: latestver_base.Defaults{
					Require: filter.RequireDefaults{
						Docker: filter.DockerCheckDefaults{
							Type: "ghcr"}}},
			},
			want: &Service{
				Options:   opt.Options{Defaults: &opt.Defaults{}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptionsDefaults{}},
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: release-argus/argus
									tag: "{{ version }}"
									token: anotherToken
						`),
						nil,
						nil,
						&latestver_base.Defaults{}, &latestver_base.Defaults{})
				}),
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: release-argus/argus
									tag: "{{ version }}"
									token: anotherToken
						`),
						nil,
						nil,
						&latestver_base.Defaults{}, &latestver_base.Defaults{})
				}),
			},
			errRegex: `^$`,
		},
		"Give DeployedVersion secrets": {
			payload: `{
				"latest_version": {
					"type": "github",
					"access_token": "` + util.SecretValue + `",
					"require": {
						"docker": {
							"type": "ghcr",
							"image": "release-argus/argus",
							"tag": "{{ version }}",
							"token": "` + util.SecretValue + `"}}},
				"deployed_version": {
					"type": "url",
					"basic_auth": {
						"password": "` + util.SecretValue + `"},
					"headers": [
						{"key": "X-Foo", "value": "` + util.SecretValue + `", "oldIndex": 0}]}}`,
			want: &Service{
				Options:   opt.Options{Defaults: &opt.Defaults{}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptionsDefaults{}},
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: release-argus/argus
									tag: "{{ version }}"
									token: anotherToken
						`),
						nil,
						nil,
						&latestver_base.Defaults{}, &latestver_base.Defaults{})
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								password: aPassword
							headers:
								- key: X-Foo
									value: aFoo
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: release-argus/argus
									tag: "{{ version }}"
									token: anotherToken
						`),
						nil,
						nil,
						&latestver_base.Defaults{}, &latestver_base.Defaults{})
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								password: aPassword
							headers:
								- key: X-Foo
									value: aFoo
						`),
						nil,
						nil,
						nil, nil)
				}),
			},
			errRegex: `^$`,
		},
		"Give Notify secrets": {
			payload: `{
				"latest_version": {
					"type": "github",
					"access_token": "` + util.SecretValue + `",
					"require": {
						"docker": {
							"type": "ghcr",
							"image": "release-argus/argus",
							"tag": "{{ version }}",
							"token": "` + util.SecretValue + `"}}},
				"deployed_version": {
					"type": "url",
					"basic_auth": {
						"password": "` + util.SecretValue + `"},
					"headers": [
						{"key": "X-Foo","value": "` + util.SecretValue + `","oldIndex": 0}]},
				"notify": {
					"slack": {
						"type": "slack",
						"url_fields": {
							"token": "` + util.SecretValue + `"},
						"oldIndex": "slack-initial"},
					"join": {
						"type": "join",
						"url_fields": {
							"apikey": "` + util.SecretValue + `"},
						"params": {
							"devices": "` + util.SecretValue + `",
							"icon": "https://example.com/icon.png"},
						"oldIndex": "join-initial"},
					"zulip": {
						"type": "zulip",
						"url_fields": {
							"botkey": "` + util.SecretValue + `"},
						"oldIndex": "zulip-initial"},
					"matrix-": {
						"type": "matrix",
						"url_fields": {
							"password": "` + util.SecretValue + `"},
						"oldIndex": "matrix-initial"},
					"rocketchat": {
						"type": "rocketchat",
						"url_fields": {
							"tokena": "` + util.SecretValue + `",
							"tokenb": "` + util.SecretValue + `"},
						"oldIndex": "rocketchat-initial"},
					"teams": {
						"type": "teams",
						"url_fields": {
							"altid": "` + util.SecretValue + `"},
						"oldIndex": "teams-initial"}}}`,
			want: &Service{
				Options:   opt.Options{Defaults: &opt.Defaults{}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptionsDefaults{}},
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: release-argus/argus
									tag: "{{ version }}"
									token: anotherToken
						`),
						nil,
						nil,
						&latestver_base.Defaults{}, &latestver_base.Defaults{})
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								password: aPassword
							headers:
								- key: X-Foo
									value: aFoo
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"slack": shoutrrr.New(
						nil, "",
						"", // Type removed as it's in ID
						nil,
						map[string]string{
							"token": "slackToken"},
						nil,
						nil, nil, nil),
					"join": shoutrrr.New(
						nil, "",
						"", // Type removed as it's in ID
						nil,
						map[string]string{
							"devices": "aDevice",
							"icon":    "https://example.com/icon.png"},
						map[string]string{
							"apikey": "joinApiKey"},
						nil, nil, nil),
					"zulip": shoutrrr.New(
						nil, "",
						"", // Type removed as it's in ID
						nil,
						map[string]string{
							"botkey": "zulipBotKey"},
						nil,
						nil, nil, nil),
					"matrix-": shoutrrr.New(
						nil, "",
						"matrix",
						nil,
						map[string]string{
							"password": "matrixToken"},
						nil,
						nil, nil, nil),
					"rocketchat": shoutrrr.New(
						nil, "",
						"", // Type removed as it's in ID
						nil,
						map[string]string{
							"tokena": "rocketchatTokenA",
							"tokenb": "rocketchatTokenB"},
						nil,
						nil, nil, nil),
					"teams": shoutrrr.New(
						nil, "",
						"", // Type removed as it's in ID
						nil,
						map[string]string{
							"altid": "teamsAltID"},
						nil,
						nil, nil, nil),
				},
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: release-argus/argus
									tag: "{{ version }}"
									token: anotherToken
						`),
						nil,
						nil,
						&latestver_base.Defaults{}, &latestver_base.Defaults{})
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								password: aPassword
							headers:
								- key: X-Foo
									value: aFoo
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"slack-initial": shoutrrr.New(
						nil, "",
						"slack",
						nil,
						map[string]string{
							"botname": "testBotName"},
						map[string]string{
							"token":   "slackToken",
							"channel": "slackChannel"},
						nil, nil, nil),
					"join-initial": shoutrrr.New(
						nil, "",
						"join",
						nil,
						map[string]string{
							"apikey": "joinApiKey"},
						map[string]string{
							"devices": "aDevice"},
						nil, nil, nil),
					"zulip-initial": shoutrrr.New(
						nil, "",
						"zulip",
						nil,
						map[string]string{
							"botmail": "zulipBotMail",
							"botkey":  "zulipBotKey",
							"host":    "zulipHost"},
						nil,
						nil, nil, nil),
					"matrix-initial": shoutrrr.New(
						nil, "",
						"matrix",
						map[string]string{
							"title": "matrixTitle"},
						map[string]string{
							"password": "matrixToken",
							"host":     "matrixHost"},
						nil,
						nil, nil, nil),
					"rocketchat-initial": shoutrrr.New(
						nil, "",
						"rocketchat",
						nil,
						map[string]string{
							"host":    "rocketchatHost",
							"tokena":  "rocketchatTokenA",
							"tokenb":  "rocketchatTokenB",
							"channel": "rocketchatChannel"},
						nil,
						nil, nil, nil),
					"teams-initial": shoutrrr.New(
						nil, "",
						"teams",
						nil,
						map[string]string{
							"group":      "teamsGroup",
							"tenant":     "teamsTenant",
							"altid":      "teamsAltID",
							"groupowner": "teamsGroupOwner"},
						map[string]string{
							"host": "teamsHost"},
						nil, nil, nil),
				},
			},
			errRegex: `^$`,
		},
		"Give WebHook secrets": {
			payload: `{
				"latest_version": {
					"type": "github",
					"access_token": "` + util.SecretValue + `",
					"require": {
						"docker": {
							"type": "ghcr",
							"image": "release-argus/args",
							"tag": "{{ version }}",
							"token": "` + util.SecretValue + `"}}},
				"deployed_version": {
					"type": "url",
					"basic_auth": {
						"password": "` + util.SecretValue + `"},
					"headers": [
						{"key": "X-Foo","value": "` + util.SecretValue + `","oldIndex": 0}]},
				"notify": {
					"slack": {
						"type": "slack",
						"url_fields": {
							"token": "` + util.SecretValue + `"},
						"oldIndex": "slack-initial"},
					"join": {
						"type": "join",
						"url_fields": {
							"apikey": "` + util.SecretValue + `"},
						"params": {
							"devices": "` + util.SecretValue + `",
							"icon": "https://example.com/icon.png"},
						"oldIndex": "join-initial"},
					"zulip": {
						"type": "zulip",
						"url_fields": {
							"botkey": "` + util.SecretValue + `"},
						"oldIndex": "zulip-initial"},
					"matrix-": {
						"type": "matrix",
						"url_fields": {
							"password": "` + util.SecretValue + `"},
						"oldIndex": "matrix-initial"},
					"rocketchat": {
						"type": "rocketchat",
						"url_fields": {
							"tokena": "` + util.SecretValue + `",
							"tokenb": "` + util.SecretValue + `"},
						"oldIndex": "rocketchat-initial"},
					"teams": {
						"type": "teams",
						"url_fields": {
							"altid": "` + util.SecretValue + `"},
						"oldIndex": "teams-initial"}},
				"webhook": {
					"github": {
						"type": "github",
						"secret": "` + util.SecretValue + `",
						"custom_headers": [
							{"key": "X-Foo", "Value": "` + util.SecretValue + `", "oldIndex": 0}],
						"oldIndex": "github-initial"},
					"gitlab-": {
						"type": "gitlab",
						"secret": "` + util.SecretValue + `",
						"custom_headers": [
							{"key": "X-Bar", "Value": "` + util.SecretValue + `", "oldIndex": 0}],
						"oldIndex": "gitlab-initial"}}}}`,
			want: &Service{
				Options:   opt.Options{Defaults: &opt.Defaults{}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptionsDefaults{}},
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: release-argus/args
									tag: "{{ version }}"
									token: anotherToken
						`),
						nil,
						nil,
						&latestver_base.Defaults{}, &latestver_base.Defaults{})
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								password: aPassword
							headers:
								- key: X-Foo
									value: aFoo
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"slack": shoutrrr.New(
						nil, "",
						"", // Type removed as it's in ID
						nil,
						map[string]string{
							"token": "slackToken"},
						nil,
						nil, nil, nil),
					"join": shoutrrr.New(
						nil, "",
						"", // Type removed as it's in ID
						nil,
						map[string]string{
							"apikey": "joinApiKey"},
						map[string]string{
							"devices": "aDevice",
							"icon":    "https://example.com/icon.png"},
						nil, nil, nil),
					"zulip": shoutrrr.New(
						nil, "",
						"", // Type removed as it's in ID
						nil,
						map[string]string{
							"botkey": "zulipBotKey"},
						nil,
						nil, nil, nil),
					"matrix-": shoutrrr.New(
						nil, "",
						"matrix",
						nil,
						map[string]string{
							"password": "matrixToken"},
						nil,
						nil, nil, nil),
					"rocketchat": shoutrrr.New(
						nil, "",
						"", // Type removed as it's in ID
						nil,
						map[string]string{
							"tokena": "rocketchatTokenA",
							"tokenb": "rocketchatTokenB"},
						nil,
						nil, nil, nil),
					"teams": shoutrrr.New(
						nil, "",
						"", // Type removed as it's in ID
						nil,
						map[string]string{
							"altid": "teamsAltId"},
						nil,
						nil, nil, nil),
				},
				WebHook: webhook.Slice{
					"github": webhook.New(
						nil,
						&webhook.Headers{
							{Key: "X-Foo", Value: "aFoo"}},
						"", nil, nil, nil, nil, nil,
						"githubSecret",
						nil,
						"", // Type removed as it's in ID
						"", nil, nil, nil),
					"gitlab-": webhook.New(
						nil,
						&webhook.Headers{
							{Key: "X-Bar", Value: "aBar"}},
						"", nil, nil, nil, nil, nil,
						"gitlabSecret",
						nil,
						"gitlab",
						"", nil, nil, nil),
				},
			},
			oldService: &Service{
				LatestVersion: test.IgnoreError(t, func() (latestver.Lookup, error) {
					return latestver.New(
						"github",
						"yaml", test.TrimYAML(`
							access_token: aToken
							require:
								docker:
									type: ghcr
									image: release-argus/args
									tag: "{{ version }}"
									token: anotherToken
						`),
						nil,
						nil,
						&latestver_base.Defaults{}, &latestver_base.Defaults{})
				}),
				DeployedVersionLookup: test.IgnoreError(t, func() (deployedver.Lookup, error) {
					return deployedver.New(
						"url",
						"yaml", test.TrimYAML(`
							basic_auth:
								password: aPassword
							headers:
								- key: X-Foo
									value: aFoo
						`),
						nil,
						nil,
						nil, nil)
				}),
				Notify: shoutrrr.Slice{
					"slack-initial": shoutrrr.New(
						nil, "",
						"slack",
						nil,
						map[string]string{
							"token":   "slackToken",
							"channel": "slackChannel"},
						map[string]string{
							"botname": "testBotName"},
						nil, nil, nil),
					"join-initial": shoutrrr.New(
						nil, "",
						"join",
						nil,
						map[string]string{
							"apikey": "joinApiKey"},
						map[string]string{
							"devices": "aDevice"},
						nil, nil, nil),
					"zulip-initial": shoutrrr.New(
						nil, "",
						"zulip",
						nil,
						map[string]string{
							"botmail": "zulipBotMail",
							"botkey":  "zulipBotKey",
							"host":    "zulipHost"},
						nil,
						nil, nil, nil),
					"matrix-initial": shoutrrr.New(
						nil, "",
						"matrix",
						map[string]string{
							"title": "matrixTitle"},
						map[string]string{
							"password": "matrixToken",
							"host":     "matrixHost"},
						nil,
						nil, nil, nil),
					"rocketchat-initial": shoutrrr.New(
						nil, "",
						"rocketchat",
						nil,
						map[string]string{
							"host":    "rocketchatHost",
							"tokena":  "rocketchatTokenA",
							"tokenb":  "rocketchatTokenB",
							"channel": "rocketchatChannel"},
						nil,
						nil, nil, nil),
					"teams-initial": shoutrrr.New(
						nil, "",
						"teams",
						nil,
						map[string]string{
							"group":      "teamsGroup",
							"tenant":     "teamsTenant",
							"altid":      "teamsAltId",
							"groupowner": "teamsGroupOwner"},
						map[string]string{
							"host": "teamsHost"},
						nil, nil, nil),
				},
				WebHook: webhook.Slice{
					"github-initial": webhook.New(
						nil,
						&webhook.Headers{
							{Key: "X-Foo", Value: "aFoo"}},
						"", nil, nil, nil, nil, nil,
						"githubSecret",
						nil,
						"github-initial",
						"", nil, nil, nil),
					"gitlab-initial": webhook.New(
						nil,
						&webhook.Headers{
							{Key: "X-Bar", Value: "aBar"}},
						"", nil, nil, nil, nil, nil,
						"gitlabSecret",
						nil,
						"gitlab-initial",
						"", nil, nil, nil),
				},
			},
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Convert the string payload to a ReadCloser
			tc.payload = test.TrimJSON(tc.payload)
			reader := bytes.NewReader([]byte(tc.payload))
			payload := io.NopCloser(reader)
			if tc.serviceHardDefaults == nil {
				tc.serviceHardDefaults = &Defaults{}
				tc.serviceHardDefaults.Default()
			}
			if tc.serviceDefaults == nil {
				tc.serviceDefaults = &Defaults{}
			}
			if tc.notifyDefaults == nil {
				tc.notifyDefaults = &shoutrrr.SliceDefaults{}
			}
			if tc.notifyHardDefaults == nil {
				tc.notifyHardDefaults = &shoutrrr.SliceDefaults{}
				tc.notifyHardDefaults.Default()
			}
			if tc.oldService != nil {
				tc.oldService.Init(
					&Defaults{}, &Defaults{},
					&shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{}, &shoutrrr.SliceDefaults{},
					&webhook.SliceDefaults{}, &webhook.Defaults{}, &webhook.Defaults{})
			}

			// WHEN we call FromPayload
			got, err := FromPayload(
				tc.oldService,
				&payload,
				tc.serviceDefaults,
				tc.serviceHardDefaults,
				tc.notifyGlobals,
				tc.notifyDefaults,
				tc.notifyHardDefaults,
				tc.webhookGlobals,
				tc.webhookDefaults,
				tc.webhookHardDefaults,
				logutil.LogFrom{Primary: name})

			// THEN we get an error if the payload is invalid
			if tc.errRegex != "" || err != nil {
				e := util.ErrorToString(err)
				if !util.RegexCheck(tc.errRegex, e) {
					t.Fatalf("error doesn't match regex\nwant match for %q\ngot: %q",
						tc.errRegex, e)
				}
				return
			}
			// AND we should get a new Service otherwise
			if got.String("") != tc.want.String("") {
				t.Errorf("Service mismatch after FromPayload\nwant:\n%s\n\ngot:\n%s",
					tc.want.String(""), got.String(""))
			}
		})
	}
}
