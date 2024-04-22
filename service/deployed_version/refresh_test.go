// Copyright [2023] [Argus]
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

package deployedver

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestBasicAuthFromString(t *testing.T) {
	// GIVEN we have a string of basic auth
	exampleBasicAuth := BasicAuth{
		Username: "user",
		Password: "pass"}
	tests := map[string]struct {
		basicAuth *string
		previous  *BasicAuth
		want      *BasicAuth
	}{
		"nil string uses previous": {
			basicAuth: nil,
			previous:  &exampleBasicAuth,
			want:      &exampleBasicAuth,
		},
		"empty string uses previous": {
			basicAuth: test.StringPtr(""),
			previous:  &exampleBasicAuth,
			want:      &exampleBasicAuth,
		},
		"user and pass set": {
			basicAuth: test.StringPtr(`{"username": "foo", "password": "bar"}`),
			previous:  &exampleBasicAuth,
			want: &BasicAuth{
				Username: "foo",
				Password: "bar"},
		},
		"only user set, get pass from previous": {
			basicAuth: test.StringPtr(`{"username": "foo"}`),
			previous:  &exampleBasicAuth,
			want: &BasicAuth{
				Username: "foo",
				Password: exampleBasicAuth.Password},
		},
		"only pass set, get user from previous": {
			basicAuth: test.StringPtr(`{"password": "bar"}`),
			previous:  &exampleBasicAuth,
			want: &BasicAuth{
				Username: exampleBasicAuth.Username,
				Password: "bar"},
		},
		"only user set, no previous": {
			basicAuth: test.StringPtr(`{"username": "foo"}`),
			previous:  nil,
			want: &BasicAuth{
				Username: "foo"},
		},
		"only pass set, no previous": {
			basicAuth: test.StringPtr(`{"password": "bar"}`),
			previous:  nil,
			want: &BasicAuth{
				Password: "bar"},
		},
		"invalid json": {
			basicAuth: test.StringPtr(`{"username": false`),
			previous:  &exampleBasicAuth,
			want:      &exampleBasicAuth,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN we call basicAuthFromString
			got := basicAuthFromString(tc.basicAuth, tc.previous, &util.LogFrom{Primary: name})

			// THEN we get the expected result
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHeadersFromString(t *testing.T) {
	// GIVEN we had previous headers and we're given a string of new headers
	previousHeaders := []Header{
		{Key: "foo", Value: "bar"}}
	tests := map[string]struct {
		headers *string
		want    *[]Header
	}{
		"invalid json": {
			headers: test.StringPtr(`{"key": false, "value": "bash"}`),
			want:    &previousHeaders},
		"nil string": {
			headers: nil,
			want:    &previousHeaders},
		"empty string": {
			headers: test.StringPtr(""),
			want:    &previousHeaders},
		"single header": {
			headers: test.StringPtr(`[{"key": "bish", "value": "bash"}]`),
			want: &[]Header{
				{Key: "bish", Value: "bash"}}},
		"multiple headers": {
			headers: test.StringPtr(`[{"key": "bish", "value": "bash"}, {"key": "bosh", "value": "bosh"}]`),
			want: &[]Header{
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "bosh"}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN we call headersFromString
			got := headersFromString(tc.headers, &previousHeaders, &util.LogFrom{Primary: name})

			// THEN we get the expected headers
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestLookup_ApplyOverrides(t *testing.T) {
	testL := testLookup()
	// GIVEN various json strings to parse as parts of a Lookup
	tests := map[string]struct {
		allowInvalidCerts  *string
		basicAuth          *string
		headers            *string
		json               *string
		regex              *string
		regexTemplate      *string
		semanticVersioning *string
		url                *string
		previous           *Lookup
		previousRegex      string
		errRegex           string
		want               *Lookup
	}{
		"all nil": {
			previous: testLookup(),
			want:     testLookup(),
		},
		"allow invalid certs": {
			allowInvalidCerts: test.StringPtr("false"),
			previous:          testLookup(),
			want: New(
				test.BoolPtr(false), // AllowInvalidCerts
				nil, nil,
				testL.JSON,
				testL.Options,
				"", nil,
				&svcstatus.Status{},
				testL.URL,
				nil, nil),
		},
		"basic auth": {
			basicAuth: test.StringPtr(`{"username": "foo", "password": "bar"}`),
			previous:  testLookup(),
			want: New(
				testL.AllowInvalidCerts,
				&BasicAuth{ // BasicAuth
					Username: "foo",
					Password: "bar"},
				nil,
				testL.JSON,
				testL.Options,
				"", nil,
				&svcstatus.Status{},
				testL.URL,
				nil, nil),
		},
		"headers": {
			headers: test.StringPtr(`[{"key": "bish", "value": "bash"}, {"key": "bosh", "value": "bosh"}]`),

			previous: testLookup(),
			want: New(
				testL.AllowInvalidCerts,
				nil,
				&[]Header{ // Headers
					{Key: "bish", Value: "bash"},
					{Key: "bosh", Value: "bosh"}},
				"version",
				testL.Options,
				"", nil,
				&svcstatus.Status{},
				testL.URL,
				nil, nil),
		},
		"json": {
			json: test.StringPtr("bish"),

			previous: testLookup(),
			want: New(
				testL.AllowInvalidCerts,
				nil, nil,
				"bish", // JSON
				testL.Options,
				"", nil,
				&svcstatus.Status{},
				testL.URL,
				nil, nil),
		},
		"regex": {
			regex: test.StringPtr("bish"),

			previous: testLookup(),
			want: New(
				testL.AllowInvalidCerts,
				nil, nil,
				"version",
				testL.Options,
				"bish", nil, // RegEx
				&svcstatus.Status{},
				testL.URL,
				nil, nil),
		},
		"regex template": {
			regexTemplate: test.StringPtr("$1.$4"),

			previous:      testLookup(),
			previousRegex: "([0-9]+)",
			want: New(
				testL.AllowInvalidCerts,
				nil, nil,
				"version",
				testL.Options,
				"([0-9]+)",
				test.StringPtr("$1.$4"), // RegEx Template
				&svcstatus.Status{},
				testL.URL,
				nil, nil),
		},
		"semantic versioning": {
			semanticVersioning: test.StringPtr("false"),

			previous: testLookup(),
			want: New(
				testL.AllowInvalidCerts,
				nil, nil,
				testL.JSON,
				opt.New(
					test.BoolPtr(false), "", nil,
					nil, nil),
				"", nil,
				&svcstatus.Status{},
				testL.URL,
				nil, nil),
		},
		"url": {
			url: test.StringPtr("https://valid.release-argus.io/json"),

			previous: testLookup(),
			want: New(
				testL.AllowInvalidCerts,
				nil, nil,
				testL.JSON,
				testL.Options,
				"", nil,
				&svcstatus.Status{},
				"https://valid.release-argus.io/json", // URL
				nil, nil),
		},
		"override with invalid (empty) url": {
			url: test.StringPtr(""),

			previous: testLookup(),
			want:     nil,
			errRegex: "url: <required>",
		},
		"override with invalid regex": {
			regex: test.StringPtr("v([0-9))"),

			previous: testLookup(),
			want:     nil,
			errRegex: "regex: .+ <invalid>",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tc.previousRegex != "" {
				tc.previous.Regex = tc.previousRegex
			}

			// WHEN we call applyOverrides
			got, err := tc.previous.applyOverrides(
				tc.allowInvalidCerts,
				tc.basicAuth,
				tc.headers,
				tc.json,
				tc.regex,
				tc.regexTemplate,
				tc.semanticVersioning,
				tc.url,
				&name,
				&util.LogFrom{Primary: name})

			// THEN we get an error if expected
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
			// AND we get the expected result otherwise
			if tc.want.String("") != got.String("") {
				t.Errorf("expected:\n%v\nbut got:\n%v", tc.want, got)
			}
		})
	}
}

func TestLookup_Refresh(t *testing.T) {
	testL := testLookup()
	testVersion, _ := testL.Query(true, &util.LogFrom{Primary: "TestRefresh"})
	if testVersion == "" {
		t.Fatalf("test version is empty")
	}

	// GIVEN a Lookup and various json strings to override parts of it
	tests := map[string]struct {
		allowInvalidCerts        *string
		basicAuth                *string
		headers                  *string
		json                     *string
		regex                    *string
		regexTemplate            *string
		semanticVersioning       *string
		url                      *string
		lookup                   *Lookup
		deployedVersion          string
		deployedVersionTimestamp string
		errRegex                 string
		want                     string
		announce                 bool
	}{
		"Change of URL": {
			url:    test.StringPtr("https://valid.release-argus.io/json"),
			lookup: testLookup(),
			want:   testVersion,
		},
		"Removal of URL": {
			url:      test.StringPtr(""),
			lookup:   testLookup(),
			errRegex: "url: <required>",
			want:     "",
		},
		"Change of a few vars": {
			json:               test.StringPtr("otherVersion"),
			semanticVersioning: test.StringPtr("false"),
			lookup:             testLookup(),
			want:               testVersion + "-beta",
		},
		"Change of vars that fail Query": {
			allowInvalidCerts: test.StringPtr("false"),
			lookup:            testLookup(),
			errRegex:          `x509 \(certificate invalid\)`,
		},
		"Refresh new version": {
			lookup: New(
				testL.AllowInvalidCerts,
				nil, nil,
				testL.JSON,
				testL.Options,
				"", nil,
				&svcstatus.Status{},
				testL.URL,
				testL.Defaults,
				testL.HardDefaults),
			deployedVersion:          "0.0.0",
			deployedVersionTimestamp: time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
			want:                     testVersion,
			announce:                 true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Copy the starting status
			previousStatus := svcstatus.Status{}
			if tc.lookup != nil {
				tc.lookup.Status = &svcstatus.Status{ServiceID: test.StringPtr("serviceID")}
				previousStatus.SetApprovedVersion(tc.lookup.Status.ApprovedVersion(), false)
				previousStatus.SetDeployedVersion(tc.lookup.Status.DeployedVersion(), false)
				previousStatus.SetDeployedVersionTimestamp(tc.lookup.Status.DeployedVersionTimestamp())
				previousStatus.SetLatestVersion(tc.lookup.Status.LatestVersion(), false)
				previousStatus.SetLatestVersionTimestamp(tc.lookup.Status.LatestVersionTimestamp())
				previousStatus.SetLastQueried(tc.lookup.Status.LastQueried())
				if tc.deployedVersion != "" {
					tc.lookup.Status.SetDeployedVersion(tc.deployedVersion, false)
					tc.lookup.Status.SetDeployedVersionTimestamp(tc.deployedVersionTimestamp)
				}
			}

			// WHEN we call Refresh
			got, gotAnnounce, err := tc.lookup.Refresh(
				tc.allowInvalidCerts,
				tc.basicAuth,
				tc.headers,
				tc.json,
				tc.regex,
				tc.regexTemplate,
				tc.semanticVersioning,
				tc.url)

			// THEN we get an error if expected
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
			}
			// AND announce is only true when expected
			if tc.announce != gotAnnounce {
				t.Errorf("expected announce of %t, not %t",
					tc.announce, gotAnnounce)
			}
			// AND we get the expected result otherwise
			if tc.want != got {
				t.Errorf("expected version %q, not %q", tc.want, got)
			}
			// AND the timestamp only changes if the version changed
			// and the possible query-changing overrides are nil
			if tc.headers == nil && tc.json == nil && tc.regex == nil && tc.semanticVersioning == nil && tc.url == nil {
				// If the version changed
				if previousStatus.DeployedVersion() != tc.lookup.Status.DeployedVersion() {
					// then so should the timestamp
					if previousStatus.DeployedVersionTimestamp() == tc.lookup.Status.DeployedVersionTimestamp() {
						t.Errorf("expected timestamp to change from %q, but got %q",
							previousStatus.DeployedVersionTimestamp(), tc.lookup.Status.DeployedVersionTimestamp())
					}
					// otherwise, the timestamp should remain unchanged
				} else if previousStatus.DeployedVersionTimestamp() != tc.lookup.Status.DeployedVersionTimestamp() {
					t.Errorf("expected timestamp %q but got %q",
						previousStatus.DeployedVersionTimestamp(), tc.lookup.Status.DeployedVersionTimestamp())
				}
				// If the overrides are not nil
			} else {
				// The timestamp shouldn't change
				if previousStatus.DeployedVersionTimestamp() != tc.lookup.Status.DeployedVersionTimestamp() {
					t.Errorf("expected timestamp %q but got %q",
						previousStatus.DeployedVersionTimestamp(), tc.lookup.Status.DeployedVersionTimestamp())
				}
			}
		})
	}
}
