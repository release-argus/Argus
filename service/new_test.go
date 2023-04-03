// Copyright [2022] [Argus]
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
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

func TestService_GiveSecretsLatestVersion(t *testing.T) {
	// GIVEN a LatestVersion that may have secrets in it referencing those in another LatestVersion
	githubData := latestver.GitHubData{ETag: "shazam"}
	tests := map[string]struct {
		latestVersion latestver.Lookup
		otherLV       latestver.Lookup
		expected      latestver.Lookup
	}{
		"empty AccessToken": {
			latestVersion: latestver.Lookup{},
			otherLV: latestver.Lookup{
				AccessToken: stringPtr("foo")},
			expected: latestver.Lookup{},
		},
		"new AccessToken kept": {
			latestVersion: latestver.Lookup{
				AccessToken: stringPtr("foo")},
			otherLV: latestver.Lookup{
				AccessToken: stringPtr("bar")},
			expected: latestver.Lookup{
				AccessToken: stringPtr("foo")},
		},
		"give old AccessToken": {
			latestVersion: latestver.Lookup{
				AccessToken: stringPtr("<secret>")},
			otherLV: latestver.Lookup{
				AccessToken: stringPtr("bar")},
			expected: latestver.Lookup{
				AccessToken: stringPtr("bar")},
		},
		"referncing default AccessToken": {
			latestVersion: latestver.Lookup{
				AccessToken: stringPtr("<secret>")},
			otherLV: latestver.Lookup{
				AccessToken: nil},
			expected: latestver.Lookup{
				AccessToken: nil},
		},
		"nil Require": {
			latestVersion: latestver.Lookup{
				Require: nil},
			otherLV: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Token: "foo"}}},
			expected: latestver.Lookup{
				Require: nil,
			},
		},
		"empty Require": {
			latestVersion: latestver.Lookup{
				Require: &filter.Require{}},
			otherLV: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Token: "foo"}}},
			expected: latestver.Lookup{
				Require: &filter.Require{}},
		},
		"new Require.Docker.Token kept": {
			latestVersion: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Token: "foo"}}},
			otherLV: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Token: "bar"}}},
			expected: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Token: "foo"}}},
		},
		"give old Require.Docker.Token": {
			latestVersion: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Token: "<secret>"}}},
			otherLV: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Token: "bar"}}},
			expected: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Token: "bar"}}},
		},
		"referencing default Require.Docker.Token": {
			latestVersion: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Token: "<secret>"}}},
			otherLV: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{}}},
			expected: latestver.Lookup{
				Require: &filter.Require{
					Docker: &filter.DockerCheck{
						Token: ""}}},
		},
		"GitHubData carried over if type still 'github'": {
			latestVersion: latestver.Lookup{
				Type: "github"},
			otherLV: latestver.Lookup{
				Type:       "github",
				GitHubData: &githubData},
			expected: latestver.Lookup{
				Type:       "github",
				GitHubData: &githubData},
		},
		"GitHubData not carried over if type wasn't 'github'": {
			latestVersion: latestver.Lookup{
				Type: "github"},
			otherLV: latestver.Lookup{
				Type:       "url",
				GitHubData: &githubData}, // would be nil for type url
			expected: latestver.Lookup{
				Type: "github"},
		},
		"GitHubData not carried over if type no longer 'github'": {
			latestVersion: latestver.Lookup{
				Type: "url"},
			otherLV: latestver.Lookup{
				Type:       "github",
				GitHubData: &githubData},
			expected: latestver.Lookup{
				Type: "github"},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			newService := &Service{LatestVersion: tc.latestVersion}
			oldService := &Service{LatestVersion: tc.otherLV}

			// WHEN we call GiveSecrets
			newService.giveSecretsLatestVersion(&oldService.LatestVersion)

			// THEN we should get a Service with the secrets from the other Service
			gotLV := newService.LatestVersion
			if tc.expected.AccessToken == nil && gotLV.AccessToken != nil {
				t.Errorf("Expected AccessToken to be nil, got %q", *gotLV.AccessToken)
			} else if util.DefaultIfNil(gotLV.AccessToken) != util.DefaultIfNil(tc.expected.AccessToken) {
				t.Errorf("Expected %q, got %q",
					util.DefaultIfNil(tc.expected.AccessToken), util.DefaultIfNil(gotLV.AccessToken))
			}

			// newService has a nil Require, but expected non-nil
			if gotLV.Require == nil && tc.expected.Require != nil {
				t.Errorf("Expected Require to be non-nil, got nil")

				// newService Require/Docker isn't nil when expected is or vice versa
			} else if gotLV.Require != tc.expected.Require &&
				gotLV.Require.Docker != tc.expected.Require.Docker &&
				// newService doesn't have the expected Token
				gotLV.Require.Docker.Token != tc.expected.Require.Docker.Token {
				t.Errorf("Expected %q, got %q",
					tc.expected.Require.Docker.Token, gotLV.Require.Docker.Token)
			}

			// GitHubData
			if gotLV.GitHubData != tc.expected.GitHubData {
				t.Errorf("Expected GitHubData to be %v, got %q",
					tc.expected.GitHubData, gotLV.GitHubData)
			}
		})
	}
}

func TestService_GiveSecretsDeployedVersion(t *testing.T) {
	// GIVEN a DeployedVersion that may have secrets in it referencing those in another DeployedVersion
	tests := map[string]struct {
		deployedVersion *deployedver.Lookup
		otherDV         *deployedver.Lookup
		secretRefs      dvlSecretRef
		expected        *deployedver.Lookup
	}{
		"nil DeployedVersion": {
			deployedVersion: nil,
			otherDV:         &deployedver.Lookup{},
			secretRefs:      dvlSecretRef{},
			expected:        nil,
		},
		"nil OldDeployedVersion": {
			deployedVersion: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "foo"}},
			otherDV: nil,
			expected: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "foo"}},
		},
		"keep BasicAuth.Password": {
			deployedVersion: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "foo"}},
			otherDV: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "bar"}},
			expected: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "foo"}},
		},
		"give old BasicAuth.Password": {
			deployedVersion: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "<secret>"}},
			otherDV: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "bar"}},
			expected: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "bar"}},
		},
		"referencing default BasicAuth.Password": {
			deployedVersion: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "<secret>"}},
			otherDV: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{}},
			expected: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: ""}},
		},
		"referencing BasicAuth.Password that doesn't exist": {
			deployedVersion: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "<secret>"}},
			otherDV: &deployedver.Lookup{},
			expected: &deployedver.Lookup{
				BasicAuth: &deployedver.BasicAuth{
					Password: "<secret>"}},
		},
		"empty Headers": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{}},
		},
		"only new Headers": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "bish", Value: "bash"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "bish", Value: "bash"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: nil},
					{OldIndex: nil}}},
		},
		"Headers with index out of range": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "bish", Value: "<secret>"},
					{Key: "bash", Value: "<secret>"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "bish", Value: "<secret>"},
					{Key: "bash", Value: "<secret>"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: intPtr(0)},
					{OldIndex: intPtr(1)}}},
		},
		"Headers with <secret> but nil index refs": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "bish", Value: "<secret>"},
					{Key: "bash", Value: "<secret>"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "bish", Value: "bash"},
					{Key: "bash", Value: "boop"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "bish", Value: "<secret>"},
					{Key: "bash", Value: "<secret>"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: nil},
					{OldIndex: nil}}},
		},
		"only changed Headers": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "shazam"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "shazam"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: intPtr(0)}}},
		},
		"only new/changed Headers": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: intPtr(0)}, {OldIndex: nil}}},
		},
		"only new/changed Headers with expected refs": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: intPtr(0)}, {OldIndex: nil}}},
		},
		"only new/changed Headers with no refs": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "shazam"},
					{Key: "bish", Value: "bash"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{}},
		},
		"referencing old Header value with no refs": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "<secret>"},
					{Key: "bish", Value: "bash"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "<secret>"},
					{Key: "bish", Value: "bash"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{}},
		},
		"only new/changed Headers with partial ref (not for all secrets)": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "<secret>"},
					{Key: "bish", Value: "bang"},
					{Key: "bosh", Value: "<secret>"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bash"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bang"},
					{Key: "bosh", Value: "<secret>"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: intPtr(0)}, {OldIndex: intPtr(1)}}},
		},
		"referencing old Header value": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "<secret>"},
					{Key: "bish", Value: "bash"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bash"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: intPtr(0)}, {OldIndex: nil}}},
		},
		"referencing old Header value that doesn't exist": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "<secret>"},
					{Key: "bish", Value: "bash"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "<secret>"},
					{Key: "bish", Value: "bash"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: intPtr(1)}, {OldIndex: nil}}},
		},
		"referencing some old Header values but not others": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bang"},
					{Key: "bish", Value: "<secret>"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bong"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bang"},
					{Key: "bish", Value: "bong"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: nil}, {OldIndex: intPtr(1)}}},
		},
		"swap header values": {
			deployedVersion: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "bish", Value: "<secret>"},
					{Key: "foo", Value: "<secret>"}}},
			otherDV: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "foo", Value: "bar"},
					{Key: "bish", Value: "bong"}}},
			expected: &deployedver.Lookup{
				Headers: []deployedver.Header{
					{Key: "bish", Value: "bar"},
					{Key: "foo", Value: "bong"}}},
			secretRefs: dvlSecretRef{
				Headers: []oldIntIndex{
					{OldIndex: intPtr(0)}, {OldIndex: intPtr(1)}}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
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
			// Got/Epxected nil but not both
			if gotDV == nil && tc.expected != nil ||
				gotDV != nil && tc.expected == nil {
				t.Errorf("Expected %q, got %q",
					tc.expected, gotDV)
			}
			// BasicAuth
			if tc.expected.BasicAuth != gotDV.BasicAuth {
				if tc.expected.BasicAuth == nil && gotDV.BasicAuth != nil {
					t.Errorf("Expected BasicAuth to be nil, got %q", *gotDV.BasicAuth)
				} else if gotDV.BasicAuth.Password != tc.expected.BasicAuth.Password {
					t.Errorf("Expected %q, got %q",
						util.DefaultIfNil(tc.expected.BasicAuth), util.DefaultIfNil(gotDV.BasicAuth))
				}
			}
			// Headers
			if len(gotDV.Headers) != len(tc.expected.Headers) {
				t.Errorf("Expected %q, got %q",
					tc.expected.Headers, gotDV.Headers)
			} else {
				for i, gotHeader := range gotDV.Headers {
					if gotHeader != tc.expected.Headers[i] {
						t.Errorf("Expected %q, got %q",
							tc.expected.Headers[i], gotHeader)
					}
				}
			}
		})
	}
}

func TestService_GiveSecretsNotify(t *testing.T) {
	// GIVEN a NotifySlice that may have secrets in it referencing those in another NotifySliceSlice
	tests := map[string]struct {
		notify      shoutrrr.Slice
		otherNotify *shoutrrr.Slice
		expected    shoutrrr.Slice
		secretRefs  *map[string]oldStringIndex
	}{
		"nil NotifySlice": {
			notify: nil,
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			expected:   nil,
			secretRefs: &map[string]oldStringIndex{},
		},
		"nil oldNotifies": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			otherNotify: nil,
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			secretRefs: &map[string]oldStringIndex{},
		},
		"nil secretRefs": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}}},
			secretRefs: nil,
		},
		"no secretRefs": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}}},
			secretRefs: &map[string]oldStringIndex{},
		},
		"no matching secretRefs": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}}},
			secretRefs: &map[string]oldStringIndex{"bish": {OldIndex: stringPtr("bash")}},
		},
		"secretRef referencing nil index": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "yikes"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "yikes"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: nil}},
		},
		"secretRef referencing index that doesn't exist": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "yikes"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "yikes"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("baz")}},
		},
		"secretRefs - url_fields.altid": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"altid": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"altid": "yikes"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"altid": "something"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"altid": "something"}},
				"bar": {
					URLFields: map[string]string{
						"altid": "yikes"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
		"secretRefs - url_fields.apikey": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "yikes"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "yikes"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
		"secretRefs - url_fields.apikey swap vars": {
			notify: shoutrrr.Slice{
				"bar": {
					URLFields: map[string]string{
						"apikey": "<secret>"}},
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "shazam"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "shazam"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			secretRefs: &map[string]oldStringIndex{
				"bar": {OldIndex: stringPtr("foo")},
				"foo": {OldIndex: stringPtr("bar")}},
		},
		"secretRefs - url_fields.apikey swap vars ignores notify order": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "<secret>"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "something"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "shazam"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"apikey": "shazam"}},
				"bar": {
					URLFields: map[string]string{
						"apikey": "something"}}},
			secretRefs: &map[string]oldStringIndex{
				"bar": {OldIndex: stringPtr("foo")},
				"foo": {OldIndex: stringPtr("bar")}},
		},
		"secretRefs - url_fields.botkey": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"botkey": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"botkey": "yikes"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"botkey": "something"}},
			},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"botkey": "something"}},
				"bar": {
					URLFields: map[string]string{
						"botkey": "yikes"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
		"secretRefs - url_fields.password": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"password": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"password": "yikes"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"password": "something"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"password": "something"}},
				"bar": {
					URLFields: map[string]string{
						"password": "yikes"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
		"secretRefs - url_fields.token": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"token": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"token": "yikes"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"token": "something"}},
			},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"token": "something"}},
				"bar": {
					URLFields: map[string]string{
						"token": "yikes"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
		"secretRefs - url_fields.tokena": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"tokena": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"tokena": "yikes"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"tokena": "something"}},
			},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"tokena": "something"}},
				"bar": {
					URLFields: map[string]string{
						"tokena": "yikes"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
		"secretRefs - url_fields.tokenb": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"tokenb": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"tokenb": "yikes"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"tokenb": "something"}},
			},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"tokenb": "something"}},
				"bar": {
					URLFields: map[string]string{
						"tokenb": "yikes"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
		"secretRefs - url_fields.host ignored as <secret>": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"host": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"host": "https://example.com"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"host": "https://example.com/foo"}},
			},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"host": "<secret>"}},
				"bar": {
					URLFields: map[string]string{
						"host": "https://example.com"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
		"secretRefs - params.devices": {
			notify: shoutrrr.Slice{
				"foo": {
					Params: map[string]string{
						"devices": "<secret>"}},
				"bar": {
					Params: map[string]string{
						"devices": "yikes"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					Params: map[string]string{
						"devices": "something"}},
			},
			expected: shoutrrr.Slice{
				"foo": {
					Params: map[string]string{
						"devices": "something"}},
				"bar": {
					Params: map[string]string{
						"devices": "yikes"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
		"secretRefs - params.avatar ignored as <secret>": {
			notify: shoutrrr.Slice{
				"foo": {
					Params: map[string]string{
						"avatar": "<secret>"}},
				"bar": {
					Params: map[string]string{
						"avatar": "https://example.com"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					Params: map[string]string{
						"avatar": "https://example.com/foo"}},
			},
			expected: shoutrrr.Slice{
				"foo": {
					Params: map[string]string{
						"avatar": "<secret>"}},
				"bar": {
					Params: map[string]string{
						"avatar": "https://example.com"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
		"secretRefs - ALL": {
			notify: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"altid":    "<secret>",
						"apikey":   "<secret>",
						"botkey":   "<secret>",
						"password": "<secret>",
						"token":    "<secret>",
						"tokena":   "<secret>",
						"tokenb":   "<secret>"},
					Params: map[string]string{
						"devices": "<secret>"}}},
			otherNotify: &shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"altid":    "whoosh",
						"apikey":   "foo",
						"botkey":   "bar",
						"password": "baz",
						"token":    "bish",
						"tokena":   "bosh",
						"tokenb":   "bash"},
					Params: map[string]string{
						"devices": "id1,id2"}}},
			expected: shoutrrr.Slice{
				"foo": {
					URLFields: map[string]string{
						"altid":    "whoosh",
						"apikey":   "foo",
						"botkey":   "bar",
						"password": "baz",
						"token":    "bish",
						"tokena":   "bosh",
						"tokenb":   "bash"},
					Params: map[string]string{
						"devices": "id1,id2"}}},
			secretRefs: &map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		newService := &Service{Notify: tc.notify}
		newService.Status.Init(
			len(newService.Notify), len(newService.Command), len(newService.WebHook),
			&name,
			nil)
		// Give empty defaults and hardDefaults to the NotifySlice
		newService.Notify.Init(
			&newService.Status,
			&shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{},
		)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN we call giveSecretsNotify
			newService.giveSecretsNotify(tc.otherNotify, tc.secretRefs)

			// THEN we should get a NotifySlice with the secrets from the other Service
			gotNotify := newService.Notify
			if gotNotify.String() != tc.expected.String() {
				t.Errorf("Want:\n%v\n\nGot:\n%v",
					tc.expected, gotNotify)
			}
		})
	}
}

func TestService_GiveSecretsWebHook(t *testing.T) {
	// GIVEN a WebHookSlice that may have secrets in it referencing those in another WebHookSliceSlice
	test := map[string]struct {
		webhook      webhook.Slice
		otherWebhook *webhook.Slice
		expected     webhook.Slice
		secretRefs   *map[string]whSecretRef
	}{
		"nil WebHookSlice": {
			webhook: nil,
			otherWebhook: &webhook.Slice{
				"foo": {Secret: "shazam"}},
			expected:   nil,
			secretRefs: &map[string]whSecretRef{},
		},
		"nil otherWebHook": {
			webhook: webhook.Slice{
				"foo": {Secret: "shazam"}},
			otherWebhook: nil,
			expected: webhook.Slice{
				"foo": {Secret: "shazam"}},
			secretRefs: &map[string]whSecretRef{},
		},
		"nil secretRefs": {
			webhook: webhook.Slice{
				"foo": {Secret: "<secret>"}},
			otherWebhook: &webhook.Slice{
				"foo": {Secret: "shazam"}},
			expected: webhook.Slice{
				"foo": {Secret: "<secret>"}},
			secretRefs: nil,
		},
		"no secretRefs": {
			webhook: webhook.Slice{
				"foo": {Secret: "<secret>"}},
			otherWebhook: &webhook.Slice{
				"foo": {Secret: "shazam"}},
			expected: webhook.Slice{
				"foo": {Secret: "<secret>"}},
			secretRefs: &map[string]whSecretRef{},
		},
		"no matching secretRefs": {
			webhook: webhook.Slice{
				"foo": {Secret: "<secret>"}},
			otherWebhook: &webhook.Slice{
				"foo": {Secret: "shazam"}},
			expected: webhook.Slice{
				"foo": {Secret: "<secret>"}},
			secretRefs: &map[string]whSecretRef{
				"bish": {OldIndex: stringPtr("bash")}},
		},
		"secretRef referencing nil index": {
			webhook: webhook.Slice{
				"foo": {Secret: "<secret>"},
				"bar": {Secret: "whoosh"}},
			otherWebhook: &webhook.Slice{
				"foo": {Secret: "shazam"}},
			expected: webhook.Slice{
				"foo": {Secret: "<secret>"},
				"bar": {Secret: "whoosh"}},
			secretRefs: &map[string]whSecretRef{
				"foo": {OldIndex: nil},
				"bar": {OldIndex: nil}},
		},
		"secretRef referencing index that doesn't exist": {
			webhook: webhook.Slice{
				"foo": {Secret: "<secret>"},
				"bar": {Secret: "whoosh"}},
			otherWebhook: &webhook.Slice{
				"foo": {Secret: "shazam"}},
			expected: webhook.Slice{
				"foo": {Secret: "<secret>"},
				"bar": {Secret: "whoosh"}},
			secretRefs: &map[string]whSecretRef{
				"foo": {OldIndex: stringPtr("bash")},
				"bar": {OldIndex: nil}},
		},
		"secretRefs - secret": {
			webhook: webhook.Slice{
				"foo": {Secret: "<secret>"},
				"bar": {Secret: "whoosh"}},
			otherWebhook: &webhook.Slice{
				"foo": {Secret: "shazam"}},
			expected: webhook.Slice{
				"foo": {Secret: "shazam"},
				"bar": {Secret: "whoosh"}},
			secretRefs: &map[string]whSecretRef{
				"foo": {OldIndex: stringPtr("foo")},
				"bar": {OldIndex: nil}},
		},
		"secretRefs - secret swap vars": {
			webhook: webhook.Slice{
				"bar": {Secret: "<secret>"},
				"foo": {Secret: "<secret>"}},
			otherWebhook: &webhook.Slice{
				"foo": {Secret: "shazam"},
				"bar": {Secret: "whoosh"}},
			expected: webhook.Slice{
				"foo": {Secret: "whoosh"},
				"bar": {Secret: "shazam"}},
			secretRefs: &map[string]whSecretRef{
				"bar": {OldIndex: stringPtr("foo")},
				"foo": {OldIndex: stringPtr("bar")}},
		},
		"secretRefs - secret swap vars ignores order sent": {
			webhook: webhook.Slice{
				"bar": {Secret: "<secret>"},
				"foo": {Secret: "<secret>"}},
			otherWebhook: &webhook.Slice{
				"foo": {Secret: "shazam"},
				"bar": {Secret: "whoosh"}},
			expected: webhook.Slice{
				"foo": {Secret: "whoosh"},
				"bar": {Secret: "shazam"}},
			secretRefs: &map[string]whSecretRef{
				"bar": {OldIndex: stringPtr("foo")},
				"foo": {OldIndex: stringPtr("bar")}},
		},
		"custom headers - no secretRefs": {
			webhook: webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "baz"}}},
			},
			otherWebhook: &webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bing"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bang"}}},
			},
			expected: webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "baz"}}},
			},
			secretRefs: &map[string]whSecretRef{},
		},
		"custom headers - no header secretRefs": {
			webhook: webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "baz"}}},
			},
			otherWebhook: &webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bing"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bang"}}},
			},
			expected: webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "baz"}}},
			},
			secretRefs: &map[string]whSecretRef{
				"foo": {OldIndex: stringPtr("foo")},
				"bar": {OldIndex: stringPtr("bar")},
			},
		},
		"custom headers - header secretRefs but old secrets unwanted": {
			webhook: webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bar"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "baz"}}},
			},
			otherWebhook: &webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bing"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bang"}}},
			},
			expected: webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bar"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "baz"}}},
			},
			secretRefs: &map[string]whSecretRef{
				"foo": {
					OldIndex: stringPtr("foo"),
					CustomHeaders: []oldIntIndex{
						{OldIndex: intPtr(0)}}},
				"bar": {
					OldIndex: stringPtr("bar"),
					CustomHeaders: []oldIntIndex{
						{OldIndex: intPtr(0)}}},
			},
		},
		"custom headers - header secretRefs, some indices out of range": {
			webhook: webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"},
						{Key: "bish", Value: "<secret>"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"},
						{Key: "bang", Value: "<secret>"}}},
			},
			otherWebhook: &webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bosh"}}},
				"bar": {CustomHeaders: &webhook.Headers{
					{Key: "foo", Value: "bang"},
					{Key: "bang", Value: "boom"}}},
			},
			expected: webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"},
						{Key: "bish", Value: "bosh"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "<secret>"}}},
			},
			secretRefs: &map[string]whSecretRef{
				"foo": {
					OldIndex: stringPtr("foo"),
					CustomHeaders: []oldIntIndex{
						{OldIndex: intPtr(5)}, {OldIndex: intPtr(1)}}},
				"bar": {
					OldIndex: stringPtr("bar"),
					CustomHeaders: []oldIntIndex{
						{OldIndex: intPtr(0)}, {OldIndex: intPtr(2)}}},
			},
		},
		"custom headers - header secretRefs use all secrets": {
			webhook: webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"},
						{Key: "bish", Value: "<secret>"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"},
						{Key: "bang", Value: "<secret>"}}},
			},
			otherWebhook: &webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bosh"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"}}},
			},
			expected: webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bosh"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"}}},
			},
			secretRefs: &map[string]whSecretRef{
				"foo": {
					OldIndex: stringPtr("foo"),
					CustomHeaders: []oldIntIndex{
						{OldIndex: intPtr(0)},
						{OldIndex: intPtr(1)}}},
				"bar": {
					OldIndex: stringPtr("bar"),
					CustomHeaders: []oldIntIndex{
						{OldIndex: intPtr(0)},
						{OldIndex: intPtr(1)}}},
			},
		},
		"custom headers - header secretRefs, swap names of webhook": {
			webhook: webhook.Slice{
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"},
						{Key: "bish", Value: "<secret>"}}},
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "<secret>"},
						{Key: "bang", Value: "<secret>"}}},
			},
			otherWebhook: &webhook.Slice{
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bosh"}}},
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"}}},
			},
			expected: webhook.Slice{
				"bar": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bing"},
						{Key: "bish", Value: "bosh"}}},
				"foo": {
					CustomHeaders: &webhook.Headers{
						{Key: "foo", Value: "bang"},
						{Key: "bang", Value: "boom"}}},
			},
			secretRefs: &map[string]whSecretRef{
				"bar": {
					OldIndex: stringPtr("foo"),
					CustomHeaders: []oldIntIndex{
						{OldIndex: intPtr(0)},
						{OldIndex: intPtr(1)}}},
				"foo": {
					OldIndex: stringPtr("bar"),
					CustomHeaders: []oldIntIndex{
						{OldIndex: intPtr(0)},
						{OldIndex: intPtr(1)}}},
			},
		},
	}

	for name, tc := range test {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			newService := &Service{
				ID:      name,
				WebHook: tc.webhook}
			// New Service Status.Fails
			newService.Status.Init(
				len(newService.Notify), len(newService.Command), len(newService.WebHook),
				&newService.ID,
				nil)
			newService.Init(
				&Service{}, &Service{},
				&shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{},
				&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
			)
			// Other Service Status.Fails
			if tc.otherWebhook != nil {
				otherServiceStatus := svcstatus.Status{}
				otherServiceStatus.Init(
					len(*tc.otherWebhook), 0, 0,
					stringPtr("otherService"),
					nil)
				tc.otherWebhook.Init(
					&otherServiceStatus,
					&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
					nil,
					stringPtr("10m"))
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
	// GIVEN a Service that may have secrets in it referencing those in another Service
	tests := map[string]struct {
		svc                              *Service
		oldService                       *Service
		oldLatestVersion                 string
		expectedLatestVersion            string
		oldLatestVersionTimestamp        string
		expectedLatestVersionTimestamp   string
		oldDeployedVersion               string
		expectedDeployedVersion          string
		oldDeployedVersionTimestamp      string
		expectedDeployedVersionTimestamp string
		oldCommandFails                  []*bool
		expectedCommandFails             []*bool
		oldWebHookFails                  map[string]*bool
		expectedWebHookFails             map[string]*bool
		secretRefs                       oldSecretRefs
		expected                         *Service
	}{
		"no secrets": {
			svc: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("something")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "user",
						Password: "pass"},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "salsy"}},
					"bar": {
						Params: map[string]string{
							"avatar": "https://example.com"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "bar"},
					"bar": {URL: "http://bar.com", Secret: "foo"},
				},
			},
			oldService: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("somethingelse")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "username",
						Password: "password"},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "sweet"}},
					"bar": {
						Params: map[string]string{
							"avatar": "https://example.com/logo.png"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "foo"},
					"bar": {URL: "http://bar.com", Secret: "bar"},
				},
			},
			expected: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("something")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "user",
						Password: "pass"},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "salsy"}},
					"bar": {
						Params: map[string]string{
							"avatar": "https://example.com"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "bar"},
					"bar": {URL: "http://bar.com", Secret: "foo"},
				},
			},
			secretRefs: oldSecretRefs{},
		},
		"minimal CREATE": {
			svc:        &Service{},
			oldService: nil,
			expected:   &Service{},
			secretRefs: oldSecretRefs{},
		},
		"no oldService (CREATE)": {
			svc: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("<secret>")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "<secret>",
						Password: "<secret>"},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "<secret>"}},
					"bar": {
						Params: map[string]string{
							"avatar": "<secret>"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "<secret>"},
					"bar": {URL: "http://bar.com", Secret: "<secret>"},
				},
			},
			oldService: nil,
			expected: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("<secret>")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "<secret>",
						Password: "<secret>"},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "<secret>"}},
					"bar": {
						Params: map[string]string{
							"avatar": "<secret>"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "<secret>"},
					"bar": {URL: "http://bar.com", Secret: "<secret>"},
				},
			},
			secretRefs: oldSecretRefs{},
		},
		"no secretRefs": {
			svc: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("<secret>")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "<secret>",
						Password: "<secret>"},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "<secret>"}},
					"bar": {
						Params: map[string]string{
							"avatar": "<secret>"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "<secret>"},
					"bar": {URL: "http://bar.com", Secret: "<secret>"},
				},
			},
			oldService: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("somethingelse")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "username",
						Password: "password"},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "sweet"}},
					"bar": {
						Params: map[string]string{
							"avatar": "https://example.com/logo.png"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "foo"},
					"bar": {URL: "http://bar.com", Secret: "bar"},
				},
			},
			oldWebHookFails: map[string]*bool{
				"foo": boolPtr(false),
				"bar": boolPtr(true)},
			expected: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("somethingelse")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "<secret>",
						Password: "password"},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "<secret>"}},
					"bar": {
						Params: map[string]string{
							"avatar": "<secret>"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "<secret>"},
					"bar": {URL: "http://bar.com", Secret: "<secret>"},
				},
			},
			secretRefs: oldSecretRefs{},
		},
		"matching secretRefs": {
			svc: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("somethingelse")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "<secret>",
						Password: "password"},
					Headers: []deployedver.Header{
						{Key: "X-Foo", Value: "<secret>"},
						{Key: "X-Bar", Value: "<secret>"},
					},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "<secret>"}},
					"bar": {
						Params: map[string]string{
							"avatar": "<secret>"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "<secret>"},
					"bar": {URL: "http://bar.com", Secret: "<secret>"},
				},
			},
			oldService: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("somethingelse")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "username",
						Password: "password"},
					Headers: []deployedver.Header{
						{Key: "X-Foo", Value: "foo"},
						{Key: "X-Bar", Value: "bar"},
					},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "sweet"}},
					"bar": {
						Params: map[string]string{
							"avatar": "https://example.com/logo.png"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "foo"},
					"bar": {URL: "http://bar.com", Secret: "bar"},
				},
			},
			expected: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("somethingelse")},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Username: "<secret>",
						Password: "password"},
					Headers: []deployedver.Header{
						{Key: "X-Foo", Value: "foo"},
						{Key: "X-Bar", Value: "bar"},
					},
				},
				Notify: shoutrrr.Slice{
					"foo": {
						URLFields: map[string]string{
							"apikey": "sweet"}},
					"bar": {
						Params: map[string]string{
							"avatar": "<secret>"}},
				},
				WebHook: webhook.Slice{
					"foo": {URL: "http://foo.com", Secret: "foo"},
					"bar": {URL: "http://bar.com", Secret: "bar"},
				},
			},
			secretRefs: oldSecretRefs{
				DeployedVersionLookup: dvlSecretRef{Headers: []oldIntIndex{{OldIndex: intPtr(0)}, {OldIndex: intPtr(1)}}},
				Notify:                map[string]oldStringIndex{"foo": {OldIndex: stringPtr("foo")}, "bar": {OldIndex: stringPtr("bar")}},
				WebHook:               map[string]whSecretRef{"foo": {OldIndex: stringPtr("foo")}, "bar": {OldIndex: stringPtr("bar")}},
			},
		},
		"unchanged LatestVersion.URL retains Status.LatestVersion": {
			svc: &Service{
				LatestVersion: latestver.Lookup{
					Type: "URL",
					URL:  "https://example.com",
				},
			},
			oldService: &Service{
				LatestVersion: latestver.Lookup{
					Type: "URL",
					URL:  "https://example.com",
				},
			},
			oldLatestVersion:               "1.2.3",
			expectedLatestVersion:          "1.2.3",
			oldLatestVersionTimestamp:      time.Now().Format(time.RFC3339),
			expectedLatestVersionTimestamp: time.Now().Format(time.RFC3339),
			expected: &Service{
				LatestVersion: latestver.Lookup{
					Type: "URL",
					URL:  "https://example.com"},
			},
			secretRefs: oldSecretRefs{},
		},
		"changed LatestVersion.URL loses Status.LatestVersion": {
			svc: &Service{
				LatestVersion: latestver.Lookup{
					Type: "URL",
					URL:  "https://example.com"},
			},
			oldService: &Service{
				LatestVersion: latestver.Lookup{
					Type: "URL",
					URL:  "https://example.com"},
			},
			oldLatestVersion:          "1.2.3",
			oldLatestVersionTimestamp: time.Now().Format(time.RFC3339),
			expected: &Service{
				LatestVersion: latestver.Lookup{
					Type: "URL",
					URL:  "https://example.com"},
			},
			secretRefs: oldSecretRefs{},
		},
		"unchanged DeployedVersion.URL retains Status.DeployedVersion": {
			svc: &Service{
				DeployedVersionLookup: &deployedver.Lookup{
					URL: "https://example.com",
				},
			},
			oldService: &Service{
				DeployedVersionLookup: &deployedver.Lookup{
					URL: "https://example.com",
				},
			},
			oldDeployedVersion:               "1.2.3",
			expectedDeployedVersion:          "1.2.3",
			oldDeployedVersionTimestamp:      time.Now().Format(time.RFC3339),
			expectedDeployedVersionTimestamp: time.Now().Format(time.RFC3339),
			expected: &Service{
				DeployedVersionLookup: &deployedver.Lookup{
					URL: "https://example.com",
				},
			},
			secretRefs: oldSecretRefs{},
		},
		"changed DeployedVersion.URL loses Status.DeployedVersion": {
			svc: &Service{
				DeployedVersionLookup: &deployedver.Lookup{
					URL: "https://example.com",
				},
			},
			oldService: &Service{
				DeployedVersionLookup: &deployedver.Lookup{
					URL: "https://example.com/somewhere-else"},
			},
			oldDeployedVersion:          "1.2.3",
			oldDeployedVersionTimestamp: time.Now().Format(time.RFC3339),
			expected: &Service{
				DeployedVersionLookup: &deployedver.Lookup{
					URL: "https://example.com"},
			},
			secretRefs: oldSecretRefs{},
		},
		"unchanged WebHook retains Failed": {
			svc: &Service{
				WebHook: webhook.Slice{
					"test": &webhook.WebHook{
						URL: "http://example.com"}}},
			oldService: &Service{
				WebHook: webhook.Slice{
					"test": &webhook.WebHook{
						ID:  "test",
						URL: "http://example.com"}}},
			expected: &Service{
				WebHook: webhook.Slice{
					"test": &webhook.WebHook{
						ID:  "test",
						URL: "http://example.com"}}},
			oldWebHookFails: map[string]*bool{
				"test": boolPtr(true)},
			expectedWebHookFails: map[string]*bool{
				"test": boolPtr(true)},
			secretRefs: oldSecretRefs{
				WebHook: map[string]whSecretRef{"test": {OldIndex: stringPtr("test")}}},
		},
		"changed WebHook loses Failed": {
			svc: &Service{
				WebHook: webhook.Slice{
					"test": &webhook.WebHook{
						URL: "http://example.com/other"}}},
			oldService: &Service{
				WebHook: webhook.Slice{
					"test": &webhook.WebHook{
						ID:  "test",
						URL: "http://example.com"}}},
			expected: &Service{
				WebHook: webhook.Slice{
					"test": &webhook.WebHook{
						ID:  "test",
						URL: "http://example.com/other"}}},
			oldWebHookFails: map[string]*bool{
				"test": boolPtr(true)},
			secretRefs: oldSecretRefs{
				WebHook: map[string]whSecretRef{"test": {OldIndex: stringPtr("test")}}},
		},
		"unchanged Command retains Failed": {
			svc: &Service{
				Command: command.Slice{
					{"ls", "-la"}}},
			oldService: &Service{
				Command: command.Slice{
					{"ls", "-la"}},
				CommandController: &command.Controller{}},
			expected: &Service{
				Command: command.Slice{
					{"ls", "-la"}},
				CommandController: &command.Controller{}},
			oldCommandFails: []*bool{
				boolPtr(true)},
			expectedCommandFails: []*bool{
				boolPtr(true)},
			secretRefs: oldSecretRefs{},
		},
		"changed Command loses Failed": {
			svc: &Service{
				Command: command.Slice{
					{"ls", "-lah"}}},
			oldService: &Service{
				Command: command.Slice{
					{"ls", "-la"}},
				CommandController: &command.Controller{}},
			expected: &Service{
				Command: command.Slice{
					{"ls", "-lah"}}},
			oldCommandFails: []*bool{
				boolPtr(true)},
			secretRefs: oldSecretRefs{},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.svc.Init(
				&Service{}, &Service{},
				&shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{},
				&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
			)
			tc.expected.Init(
				&Service{}, &Service{},
				&shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{},
				&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
			)
			if tc.expected != nil {
				for k, v := range tc.expectedCommandFails {
					if v != nil {
						tc.expected.Status.Fails.Command.Set(k, *v)
					}
				}
				for k, v := range tc.expectedWebHookFails {
					tc.expected.Status.Fails.WebHook.Set(k, v)
				}
			}
			if tc.oldService != nil {
				tc.oldService.Init(
					&Service{}, &Service{},
					&shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{},
					&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
				)
				if tc.oldService.Command != nil {
					tc.oldService.CommandController.Command = &tc.oldService.Command
				}
				for k, v := range tc.oldCommandFails {
					if v != nil {
						tc.oldService.Status.Fails.Command.Set(k, *v)
					}
				}
				for k, v := range tc.oldWebHookFails {
					tc.oldService.Status.Fails.WebHook.Set(k, v)
				}
			}

			// WHEN we call giveSecrets
			tc.svc.giveSecrets(tc.oldService, tc.secretRefs)

			// THEN we should get a Service with the secrets from the old Service
			gotService := tc.svc
			if gotService.String() != tc.expected.String() {
				t.Errorf("Want:\n%v\n\nGot:\n%v",
					tc.expected, gotService)
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

func TestNew_ReadFromFail(t *testing.T) {
	testLogging()
	// GIVEN an invalid payload
	payloadStr := "this is a long payload"
	payload := ioutil.NopCloser(bytes.NewReader([]byte(payloadStr)))
	payload = http.MaxBytesReader(nil, payload, 5)

	// WHEN we call New
	_, err := New(
		&Service{},
		&payload,
		&Service{},
		&Service{},
		&shoutrrr.Slice{},
		&shoutrrr.Slice{},
		&shoutrrr.Slice{},
		&webhook.Slice{},
		&webhook.WebHook{},
		&webhook.WebHook{},
		&util.LogFrom{},
	)

	// THEN we should get an error
	if err == nil {
		t.Errorf("Want error, got nil")
	}
}

func TestNew(t *testing.T) {
	testLogging()
	// GIVEN a payload and the Service defaults
	tests := map[string]struct {
		oldService *Service
		payload    string

		serviceDefaults     *Service
		serviceHardDefaults *Service

		notifyGlobals      *shoutrrr.Slice
		notifyDefaults     *shoutrrr.Slice
		notifyHardDefaults *shoutrrr.Slice

		webhookGlobals      *webhook.Slice
		webhookDefaults     *webhook.WebHook
		webhookHardDefaults *webhook.WebHook

		want     *Service
		errRegex string
	}{
		"empty payload": {
			payload:  "",
			errRegex: "^EOF$",
		},
		"invalid payload": {
			payload:  strings.Repeat("a", 1048577),
			errRegex: "^invalid character 'a' looking for beginning of value$",
		},
		"invalid Service payload": {
			payload:  `{"name": false}`,
			errRegex: "^json: cannot unmarshal bool into Go struct field Service.name of type string$",
		},
		"invalid SecretRefs payload": {
			payload:  `{"webhook": {"foo": {"oldIndex": false}}}`,
			errRegex: "^json: cannot unmarshal bool into Go struct field whSecretRef.webhook.oldIndex of type string",
		},
		"active True becomes nil": {
			payload: `{"options":{"active":true}}`,
			// Defaults as otherwise everything will be zero, so won't print
			want: &Service{
				LatestVersion: latestver.Lookup{Defaults: &latestver.Lookup{}},
				Dashboard:     DashboardOptions{Defaults: &DashboardOptions{}},
				Options: opt.Options{
					Active:   nil,
					Defaults: &opt.Options{}}},
		},
		"active nil stays nil": {
			payload: `{"options":{"active":null}}`,
			// Defaults as otherwise everything will be zero, so won't print
			want: &Service{
				LatestVersion: latestver.Lookup{Defaults: &latestver.Lookup{}},
				Dashboard:     DashboardOptions{Defaults: &DashboardOptions{}},
				Options: opt.Options{
					Active:   nil,
					Defaults: &opt.Options{}}},
		},
		"active False stays false": {
			payload: `{"options":{"active":false}}`,
			// Defaults as otherwise everything will be zero, so won't print
			want: &Service{
				LatestVersion: latestver.Lookup{Defaults: &latestver.Lookup{}},
				Dashboard:     DashboardOptions{Defaults: &DashboardOptions{}},
				Options: opt.Options{
					Active:   boolPtr(false),
					Defaults: &opt.Options{}}},
		},
		"Require.Docker removed if no Image&Tag": {
			payload: `{"latest_version":{"require":{"docker":{"type":"ghcr"}}}}`,
			want: &Service{
				Options:   opt.Options{Defaults: &opt.Options{}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptions{}},
				LatestVersion: latestver.Lookup{
					Defaults: &latestver.Lookup{},
					Require:  &filter.Require{}}},
		},
		"Require.Docker stays if have Type&Image&Tag": {
			payload: `{"latest_version":{"require":{"docker":{"type":"ghcr","image":"release-argus-argus","tag":"latest"}}}}`,
			want: &Service{
				Options: opt.Options{Defaults: &opt.Options{}},
				LatestVersion: latestver.Lookup{
					Require: &filter.Require{
						Docker: &filter.DockerCheck{
							Type:  "ghcr",
							Image: "release-argus-argus",
							Tag:   "latest"}}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptions{}}},
		},
		"Give LatestVersion secrets": {
			payload: `{
				"latest_version": {
					"access_token": "<secret>",
					"require": {
						"docker": {
							"token": "<secret>"}}}}`,
			want: &Service{
				Options:   opt.Options{Defaults: &opt.Options{}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptions{}},
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("aToken"),
					Require: &filter.Require{
						Docker: &filter.DockerCheck{
							Token: "anotherToken"}}},
			},
			oldService: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("aToken"),
					Require: &filter.Require{
						Docker: &filter.DockerCheck{
							Token: "anotherToken"}}},
			},
		},
		"Give DeployedVersion secrets": {
			payload: `{
				"latest_version": {
					"access_token": "<secret>",
					"require": {
						"docker": {
							"token": "<secret>"}}},
				"deployed_version": {
					"basic_auth": {
						"password": "<secret>"},
					"headers": [
						{"key": "X-Foo", "value": "<secret>", "oldIndex": 0}
					]}}`,
			want: &Service{
				Options:   opt.Options{Defaults: &opt.Options{}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptions{}},
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("aToken"),
					Require: &filter.Require{
						Docker: &filter.DockerCheck{
							Token: "anotherToken"}}},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Password: "aPassword"},
					Headers: []deployedver.Header{
						{Key: "X-Foo", Value: "aFoo"}}},
			},
			oldService: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("aToken"),
					Require: &filter.Require{
						Docker: &filter.DockerCheck{
							Token: "anotherToken"}}},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Password: "aPassword"},
					Headers: []deployedver.Header{
						{Key: "X-Foo", Value: "aFoo"}}},
			},
		},
		"Give Notify secrets": {
			payload: `{
				"latest_version": {
					"access_token": "<secret>",
					"require": {
						"docker": {
							"token": "<secret>"}}},
				"deployed_version": {
					"basic_auth": {
						"password": "<secret>"},
					"headers": [
						{"key": "X-Foo","value": "<secret>","oldIndex": 0}]},
				"notify": {
					"slack": {
						"type": "slack",
						"url_fields": {
							"token": "<secret>"},
						"oldIndex": "slack-initial"},
					"join": {
						"type": "join",
						"url_fields": {
							"apikey": "<secret>"},
						"params": {
							"devices": "<secret>",
							"icon": "https://example.com/icon.png"},
						"oldIndex": "join-initial"},
					"zulip": {
						"type": "zulip",
						"url_fields": {
							"botkey": "<secret>"},
						"oldIndex": "zulip-initial"},
					"matrix": {
						"type": "matrix",
						"url_fields": {
							"password": "<secret>"},
						"oldIndex": "matrix-initial"},
					"rocketchat": {
						"type": "rocketchat",
						"url_fields": {
							"tokena": "<secret>",
							"tokenb": "<secret>"},
						"oldIndex": "rocketchat-initial"},
					"teams": {
						"type": "teams",
						"url_fields": {
							"altid": "<secret>"},
						"oldIndex": "teams-initial"}
				}}`,
			want: &Service{
				Options:   opt.Options{Defaults: &opt.Options{}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptions{}},
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("aToken"),
					Require: &filter.Require{
						Docker: &filter.DockerCheck{
							Token: "anotherToken"}}},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Password: "aPassword"},
					Headers: []deployedver.Header{
						{Key: "X-Foo", Value: "aFoo"}}},
				Notify: shoutrrr.Slice{
					"slack": {
						Type: "slack",
						URLFields: map[string]string{
							"token": "slackToken"}},
					"join": {
						Type: "join",
						URLFields: map[string]string{
							"apikey": "joinApiKey"},
						Params: map[string]string{
							"devices": "aDevice",
							"icon":    "https://example.com/icon.png"}},
					"zulip": {
						Type: "zulip",
						URLFields: map[string]string{
							"botkey": "zulipBotKey"}},
					"matrix": {
						Type: "matrix",
						URLFields: map[string]string{
							"password": "matrixToken"}},
					"rocketchat": {
						Type: "rocketchat",
						URLFields: map[string]string{
							"tokena": "rocketchatTokenA",
							"tokenb": "rocketchatTokenB"}},
					"teams": {
						Type: "teams",
						URLFields: map[string]string{
							"altid": "teamsAltId"}},
				},
			},
			oldService: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("aToken"),
					Require: &filter.Require{
						Docker: &filter.DockerCheck{
							Token: "anotherToken"}}},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Password: "aPassword"},
					Headers: []deployedver.Header{
						{Key: "X-Foo", Value: "aFoo"}}},
				Notify: shoutrrr.Slice{
					"slack-initial": {
						Type: "slack",
						URLFields: map[string]string{
							"token":   "slackToken",
							"channel": "slackChannel"},
						Params: map[string]string{
							"botname": "testBotName"}},
					"join-initial": {
						Type: "join",
						URLFields: map[string]string{
							"apikey": "joinApiKey"},
						Params: map[string]string{
							"devices": "aDevice"}},
					"zulip-initial": {
						Type: "zulip",
						URLFields: map[string]string{
							"botmail": "zulipBotMail",
							"botkey":  "zulipBotKey",
							"host":    "zulipHost"}},
					"matrix-initial": {
						Type: "matrix",
						URLFields: map[string]string{
							"password": "matrixToken",
							"host":     "matrixHost"},
						Params: map[string]string{
							"title": "matrixTitle"}},
					"rocketchat-initial": {
						Type: "rocketchat",
						URLFields: map[string]string{
							"host":    "rocketchatHost",
							"tokena":  "rocketchatTokenA",
							"tokenb":  "rocketchatTokenB",
							"channel": "rocketchatChannel"}},
					"teams-initial": {
						Type: "teams",
						URLFields: map[string]string{
							"group":      "teamsGroup",
							"tenant":     "teamsTenant",
							"altid":      "teamsAltId",
							"groupowner": "teamsGroupOwner"},
						Params: map[string]string{
							"host": "teamsHost"}},
				},
			},
		},
		"Give WebHook secrets": {
			payload: `{
				"latest_version": {
					"access_token": "<secret>",
					"require": {
						"docker": {
							"token": "<secret>"}}},
				"deployed_version": {
					"basic_auth": {
						"password": "<secret>"},
					"headers": [
						{"key": "X-Foo","value": "<secret>","oldIndex": 0}]},
				"notify": {
					"slack": {
						"type": "slack",
						"url_fields": {
							"token": "<secret>"},
						"oldIndex": "slack-initial"},
					"join": {
						"type": "join",
						"url_fields": {
							"apikey": "<secret>"},
						"params": {
							"devices": "<secret>",
							"icon": "https://example.com/icon.png"},
						"oldIndex": "join-initial"},
					"zulip": {
						"type": "zulip",
						"url_fields": {
							"botkey": "<secret>"},
						"oldIndex": "zulip-initial"},
					"matrix": {
						"type": "matrix",
						"url_fields": {
							"password": "<secret>"},
						"oldIndex": "matrix-initial"},
					"rocketchat": {
						"type": "rocketchat",
						"url_fields": {
							"tokena": "<secret>",
							"tokenb": "<secret>"},
						"oldIndex": "rocketchat-initial"},
					"teams": {
						"type": "teams",
						"url_fields": {
							"altid": "<secret>"},
						"oldIndex": "teams-initial"}},
				"webhook": {
					"github": {"type": "github",
						"secret": "<secret>",
						"custom_headers": [
							{"key":"X-Foo", "Value": "<secret>", "oldIndex": 0}],
						"oldIndex": "github-initial"},
					"gitlab": {"type": "gitlab",
						"secret": "<secret>",
						"custom_headers": [
							{"key":"X-Bar", "Value": "<secret>", "oldIndex": 0}],
						"oldIndex": "gitlab-initial"}}}}`,
			want: &Service{
				Options:   opt.Options{Defaults: &opt.Options{}},
				Dashboard: DashboardOptions{Defaults: &DashboardOptions{}},
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("aToken"),
					Require: &filter.Require{
						Docker: &filter.DockerCheck{
							Token: "anotherToken"}}},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Password: "aPassword"},
					Headers: []deployedver.Header{
						{Key: "X-Foo", Value: "aFoo"}}},
				Notify: shoutrrr.Slice{
					"slack": {
						Type: "slack",
						URLFields: map[string]string{
							"token": "slackToken"}},
					"join": {
						Type: "join",
						URLFields: map[string]string{
							"apikey": "joinApiKey"},
						Params: map[string]string{
							"devices": "aDevice",
							"icon":    "https://example.com/icon.png"}},
					"zulip": {
						Type: "zulip",
						URLFields: map[string]string{
							"botkey": "zulipBotKey"}},
					"matrix": {
						Type: "matrix",
						URLFields: map[string]string{
							"password": "matrixToken"}},
					"rocketchat": {
						Type: "rocketchat",
						URLFields: map[string]string{
							"tokena": "rocketchatTokenA",
							"tokenb": "rocketchatTokenB"}},
					"teams": {
						Type: "teams",
						URLFields: map[string]string{
							"altid": "teamsAltId"}},
				},
				WebHook: webhook.Slice{
					"github": {
						Type:   "github",
						Secret: "githubSecret",
						CustomHeaders: &webhook.Headers{
							{Key: "X-Foo", Value: "aFoo"}}},
					"gitlab": {
						Type:   "gitlab",
						Secret: "gitlabSecret",
						CustomHeaders: &webhook.Headers{
							{Key: "X-Bar", Value: "aBar"}}},
				},
			},
			oldService: &Service{
				LatestVersion: latestver.Lookup{
					AccessToken: stringPtr("aToken"),
					Require: &filter.Require{
						Docker: &filter.DockerCheck{
							Token: "anotherToken"}}},
				DeployedVersionLookup: &deployedver.Lookup{
					BasicAuth: &deployedver.BasicAuth{
						Password: "aPassword"},
					Headers: []deployedver.Header{
						{Key: "X-Foo", Value: "aFoo"}}},
				Notify: shoutrrr.Slice{
					"slack-initial": {
						Type: "slack",
						URLFields: map[string]string{
							"token":   "slackToken",
							"channel": "slackChannel"},
						Params: map[string]string{
							"botname": "testBotName"}},
					"join-initial": {
						Type: "join",
						URLFields: map[string]string{
							"apikey": "joinApiKey"},
						Params: map[string]string{
							"devices": "aDevice"}},
					"zulip-initial": {
						Type: "zulip",
						URLFields: map[string]string{
							"botmail": "zulipBotMail",
							"botkey":  "zulipBotKey",
							"host":    "zulipHost"}},
					"matrix-initial": {
						Type: "matrix",
						URLFields: map[string]string{
							"password": "matrixToken",
							"host":     "matrixHost"},
						Params: map[string]string{
							"title": "matrixTitle"}},
					"rocketchat-initial": {
						Type: "rocketchat",
						URLFields: map[string]string{
							"host":    "rocketchatHost",
							"tokena":  "rocketchatTokenA",
							"tokenb":  "rocketchatTokenB",
							"channel": "rocketchatChannel"}},
					"teams-initial": {
						Type: "teams",
						URLFields: map[string]string{
							"group":      "teamsGroup",
							"tenant":     "teamsTenant",
							"altid":      "teamsAltId",
							"groupowner": "teamsGroupOwner"},
						Params: map[string]string{
							"host": "teamsHost"}},
				},
				WebHook: webhook.Slice{
					"github-initial": {
						Type:   "github",
						Secret: "githubSecret",
						CustomHeaders: &webhook.Headers{
							{Key: "X-Foo", Value: "aFoo"}}},
					"gitlab-initial": {
						Type:   "gitlab",
						Secret: "gitlabSecret",
						CustomHeaders: &webhook.Headers{
							{Key: "X-Bar", Value: "aBar"}}},
				},
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Convert the string payload to a ReadCloser
			reader := bytes.NewReader([]byte(tc.payload))
			payload := ioutil.NopCloser(reader)
			if tc.serviceHardDefaults == nil {
				tc.serviceHardDefaults = &Service{}
				tc.serviceHardDefaults.Status.Init(
					0, 0, 0,
					stringPtr("serviceID"),
					stringPtr("https://example.com"),
				)
			}
			if tc.serviceDefaults == nil {
				tc.serviceDefaults = &Service{}
			}
			if tc.notifyDefaults == nil {
				tc.notifyDefaults = &shoutrrr.Slice{}
			}
			if tc.notifyHardDefaults == nil {
				tc.notifyHardDefaults = &shoutrrr.Slice{}
			}
			if tc.oldService != nil {
				tc.oldService.Init(
					&Service{}, &Service{},
					&shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{},
					&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{})
			}

			// WHEN we call New
			got, err := New(
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
				&util.LogFrom{Primary: name})

			// THEN we get an error if the payload is invalid
			if tc.errRegex != "" || err != nil {
				// No error expected
				if tc.errRegex == "" {
					tc.errRegex = "^$"
				}
				e := util.ErrorToString(err)
				re := regexp.MustCompile(tc.errRegex)
				match := re.MatchString(e)
				if !match {
					t.Fatalf("want match for %q\nnot: %q",
						tc.errRegex, e)
				}
				return
			}
			// AND we should get a new Service otherwise
			if got.String() != tc.want.String() {
				t.Errorf("Want:\n%v\n\nGot:\n%v",
					tc.want, got)
			}
		})
	}
}

func TestService_CheckFetches(t *testing.T) {
	// GIVEN a Service
	testLogging()
	testLV := testLatestVersionLookupURL(false)
	testLV.Query()
	testDVL := testDeployedVersionLookup(false)
	v, _ := testDVL.Query(&util.LogFrom{})
	testDVL.Status.SetDeployedVersion(v, false)
	tests := map[string]struct {
		svc                  *Service
		startLatestVersion   string
		wantLatestVersion    string
		startDeployedVersion string
		wantDeployedVersion  string
		errRegex             string
	}{
		"Already have LatestVersion, nil DeployedVersionLookup": {
			svc: &Service{
				LatestVersion:         testLatestVersionLookupURL(false),
				DeployedVersionLookup: nil},
			startLatestVersion:   "foo",
			wantLatestVersion:    testLV.Status.GetLatestVersion(),
			startDeployedVersion: "bar",
			wantDeployedVersion:  "bar",
			errRegex:             "^$",
		},
		"Already have LatestVersion and DeployedVersionLookup": {
			svc: &Service{
				LatestVersion:         testLatestVersionLookupURL(false),
				DeployedVersionLookup: testDeployedVersionLookup(false)},
			startLatestVersion:   "foo",
			wantLatestVersion:    testLV.Status.GetLatestVersion(),
			wantDeployedVersion:  testDVL.Status.GetDeployedVersion(),
			startDeployedVersion: "bar",
			errRegex:             "^$",
		},
		"latest_version query fails": {
			svc: &Service{
				LatestVersion:         testLatestVersionLookupURL(true),
				DeployedVersionLookup: testDeployedVersionLookup(false)},
			errRegex: `latest_version - x509 \(certificate invalid\)`,
		},
		"deployed_version query fails": {
			svc: &Service{
				LatestVersion:         testLatestVersionLookupURL(false),
				DeployedVersionLookup: testDeployedVersionLookup(true)},
			wantLatestVersion: "1.2.2",
			errRegex:          `deployed_version - x509 \(certificate invalid\)`,
		},
		"both queried": {
			svc: &Service{
				LatestVersion:         testLatestVersionLookupURL(false),
				DeployedVersionLookup: testDeployedVersionLookup(false)},
			wantLatestVersion:   "1.2.2",
			wantDeployedVersion: "1.2.3",
			errRegex:            "^$",
		},
		"inactive queries neither": {
			svc: &Service{
				Options: opt.Options{
					Active: boolPtr(false)},
				LatestVersion:         testLatestVersionLookupURL(false),
				DeployedVersionLookup: testDeployedVersionLookup(false)},
			errRegex: "^$",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		tc.svc.Init(
			&Service{
				LatestVersion:         latestver.Lookup{},
				DeployedVersionLookup: &deployedver.Lookup{},
				Options:               opt.Options{}},
			&Service{
				LatestVersion:         latestver.Lookup{},
				DeployedVersionLookup: &deployedver.Lookup{},
				Options: opt.Options{
					SemanticVersioning: boolPtr(true)}},
			&shoutrrr.Slice{}, &shoutrrr.Slice{}, &shoutrrr.Slice{},
			&webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{},
		)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			announceChannel := make(chan []byte, 5)
			tc.svc.Status.AnnounceChannel = &announceChannel
			tc.svc.Status.SetLatestVersion(tc.startLatestVersion, false)
			tc.svc.Status.SetDeployedVersion(tc.startDeployedVersion, false)

			// WHEN we call CheckFetches
			err := tc.svc.CheckFetches()

			// THEN we get the err we expect
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			// AND we get the expected LatestVersion
			if tc.svc.Status.GetLatestVersion() != tc.wantLatestVersion {
				t.Errorf("LatestVersion\nWant: %q, got: %q",
					tc.wantLatestVersion, tc.svc.Status.GetLatestVersion())
			}
			// AND we get the expected DeployedVersion
			if tc.svc.Status.GetDeployedVersion() != tc.wantDeployedVersion {
				t.Errorf("DeployedVersion\nWant: %q, got: %q",
					tc.wantDeployedVersion, tc.svc.Status.GetDeployedVersion())
			}
			if len(*tc.svc.Status.AnnounceChannel) != 0 {
				t.Errorf("AnnounceChannel should be empty, got %d",
					len(*tc.svc.Status.AnnounceChannel))
			}
		})
	}
}
