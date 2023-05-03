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
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
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
			basicAuth: stringPtr(""),
			previous:  &exampleBasicAuth,
			want:      &exampleBasicAuth,
		},
		"user and pass set": {
			basicAuth: stringPtr(`{"username": "foo", "password": "bar"}`),
			previous:  &exampleBasicAuth,
			want: &BasicAuth{
				Username: "foo",
				Password: "bar"},
		},
		"only user set, get pass from previous": {
			basicAuth: stringPtr(`{"username": "foo"}`),
			previous:  &exampleBasicAuth,
			want: &BasicAuth{
				Username: "foo",
				Password: exampleBasicAuth.Password},
		},
		"only pass set, get user from previous": {
			basicAuth: stringPtr(`{"password": "bar"}`),
			previous:  &exampleBasicAuth,
			want: &BasicAuth{
				Username: exampleBasicAuth.Username,
				Password: "bar"},
		},
		"only user set, no previous": {
			basicAuth: stringPtr(`{"username": "foo"}`),
			previous:  nil,
			want: &BasicAuth{
				Username: "foo"},
		},
		"only pass set, no previous": {
			basicAuth: stringPtr(`{"password": "bar"}`),
			previous:  nil,
			want: &BasicAuth{
				Password: "bar"},
		},
		"invalid json": {
			basicAuth: stringPtr(`{"username": false`),
			previous:  &exampleBasicAuth,
			want:      &exampleBasicAuth,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
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
			headers: stringPtr(`{"key": false, "value": "bash"}`),
			want:    &previousHeaders},
		"nil string": {
			headers: nil,
			want:    &previousHeaders},
		"empty string": {
			headers: stringPtr(""),
			want:    &previousHeaders},
		"single header": {
			headers: stringPtr(`[{"key": "bish", "value": "bash"}]`),
			want: &[]Header{
				{Key: "bish", Value: "bash"}}},
		"multiple headers": {
			headers: stringPtr(`[{"key": "bish", "value": "bash"}, {"key": "bosh", "value": "bosh"}]`),
			want: &[]Header{
				{Key: "bish", Value: "bash"},
				{Key: "bosh", Value: "bosh"}}},
	}

	for name, tc := range tests {
		name, tc := name, tc
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
	test := testLookup()
	// GIVEN various json strings to parse as parts of a Lookup
	tests := map[string]struct {
		allowInvalidCerts  *string
		basicAuth          *string
		headers            *string
		json               *string
		regex              *string
		semanticVersioning *string
		url                *string
		previous           *Lookup
		errRegex           string
		want               *Lookup
	}{
		"all nil": {
			previous: testLookup(),
			want:     testLookup(),
		},
		"allow invalid certs": {
			allowInvalidCerts: stringPtr("false"),
			previous:          testLookup(),
			want: New(
				boolPtr(false), // AllowInvalidCerts
				nil, nil,
				test.JSON,
				test.Options,
				"",
				&svcstatus.Status{},
				test.URL,
				nil, nil),
		},
		"basic auth": {
			basicAuth: stringPtr(`{"username": "foo", "password": "bar"}`),
			previous:  testLookup(),
			want: New(
				test.AllowInvalidCerts,
				&BasicAuth{ // BasicAuth
					Username: "foo",
					Password: "bar"},
				nil,
				test.JSON,
				test.Options,
				"",
				&svcstatus.Status{},
				test.URL,
				nil, nil),
		},
		"headers": {
			headers: stringPtr(`[{"key": "bish", "value": "bash"}, {"key": "bosh", "value": "bosh"}]`),

			previous: testLookup(),
			want: New(
				test.AllowInvalidCerts,
				nil,
				&[]Header{ // Headers
					{Key: "bish", Value: "bash"},
					{Key: "bosh", Value: "bosh"}},
				"version",
				test.Options,
				"",
				&svcstatus.Status{},
				test.URL,
				nil, nil),
		},
		"json": {
			json: stringPtr("bish"),

			previous: testLookup(),
			want: New(
				test.AllowInvalidCerts,
				nil, nil,
				"bish", // JSON
				test.Options,
				"",
				&svcstatus.Status{},
				test.URL,
				nil, nil),
		},
		"regex": {
			regex: stringPtr("bish"),

			previous: testLookup(),
			want: New(
				test.AllowInvalidCerts,
				nil, nil,
				"version",
				test.Options,
				"bish", // RegEx
				&svcstatus.Status{},
				test.URL,
				nil, nil),
		},
		"semantic versioning": {
			semanticVersioning: stringPtr("false"),

			previous: testLookup(),
			want: New(
				test.AllowInvalidCerts,
				nil, nil,
				test.JSON,
				opt.New(
					boolPtr(false), "", nil,
					nil, nil),
				"",
				&svcstatus.Status{},
				test.URL,
				nil, nil),
		},
		"url": {
			url: stringPtr("https://valid.release-argus.io/json"),

			previous: testLookup(),
			want: New(
				test.AllowInvalidCerts,
				nil, nil,
				test.JSON,
				test.Options,
				"",
				&svcstatus.Status{},
				"https://valid.release-argus.io/json", // URL
				nil, nil),
		},
		"override with invalid (empty) url": {
			url: stringPtr(""),

			previous: testLookup(),
			want:     nil,
			errRegex: "url: <required>",
		},
		"override with invalid regex": {
			regex: stringPtr("v([0-9))"),

			previous: testLookup(),
			want:     nil,
			errRegex: "regex: .+ <invalid>",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN we call applyOverrides
			got, err := tc.previous.applyOverrides(
				tc.allowInvalidCerts,
				tc.basicAuth,
				tc.headers,
				tc.json,
				tc.regex,
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
	test := testLookup()
	testVersion, _ := test.Query(true, &util.LogFrom{Primary: "TestRefresh"})
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
			url:    stringPtr("https://valid.release-argus.io/json"),
			lookup: testLookup(),
			want:   testVersion,
		},
		"Removal of URL": {
			url:      stringPtr(""),
			lookup:   testLookup(),
			errRegex: "url: <required>",
			want:     "",
		},
		"Change of a few vars": {
			json:               stringPtr("otherVersion"),
			semanticVersioning: stringPtr("false"),
			lookup:             testLookup(),
			want:               testVersion + "-beta",
		},
		"Change of vars that fail Query": {
			allowInvalidCerts: stringPtr("false"),
			lookup:            testLookup(),
			errRegex:          `x509 \(certificate invalid\)`,
		},
		"Refresh new version": {
			lookup: New(
				test.AllowInvalidCerts,
				nil, nil,
				test.JSON,
				test.Options,
				"",
				&svcstatus.Status{},
				test.URL,
				test.Defaults,
				test.HardDefaults),
			deployedVersion:          "0.0.0",
			deployedVersionTimestamp: time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
			want:                     testVersion,
			announce:                 true,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Copy the starting status
			previousStatus := svcstatus.Status{}
			if tc.lookup != nil {
				tc.lookup.Status = &svcstatus.Status{ServiceID: stringPtr("serviceID")}
				previousStatus.SetApprovedVersion(tc.lookup.Status.GetApprovedVersion(), false)
				previousStatus.SetDeployedVersion(tc.lookup.Status.GetDeployedVersion(), false)
				previousStatus.SetDeployedVersionTimestamp(tc.lookup.Status.GetDeployedVersionTimestamp())
				previousStatus.SetLatestVersion(tc.lookup.Status.GetLatestVersion(), false)
				previousStatus.SetLatestVersionTimestamp(tc.lookup.Status.GetLatestVersionTimestamp())
				previousStatus.SetLastQueried(tc.lookup.Status.GetLastQueried())
				if tc.deployedVersion != "" {
					tc.lookup.Status.SetDeployedVersion(tc.deployedVersion, false)
					tc.lookup.Status.SetDeployedVersionTimestamp(tc.deployedVersionTimestamp)
				}
			}

			if strings.HasPrefix(name, "Change of vars that fail Query") || strings.HasPrefix(name, "Refresh new version") {
				fmt.Println()
			}

			// WHEN we call Refresh
			got, gotAnnounce, err := tc.lookup.Refresh(
				tc.allowInvalidCerts,
				tc.basicAuth,
				tc.headers,
				tc.json,
				tc.regex,
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
				if previousStatus.GetDeployedVersion() != tc.lookup.Status.GetDeployedVersion() {
					// then so should the timestamp
					if previousStatus.GetDeployedVersionTimestamp() == tc.lookup.Status.GetDeployedVersionTimestamp() {
						t.Errorf("expected timestamp to change from %q, but got %q",
							previousStatus.GetDeployedVersionTimestamp(), tc.lookup.Status.GetDeployedVersionTimestamp())
					}
					// otherwise, the timestamp should remain unchanged
				} else if previousStatus.GetDeployedVersionTimestamp() != tc.lookup.Status.GetDeployedVersionTimestamp() {
					t.Errorf("expected timestamp %q but got %q",
						previousStatus.GetDeployedVersionTimestamp(), tc.lookup.Status.GetDeployedVersionTimestamp())
				}
				// If the overrides are not nil
			} else {
				// The timestamp shouldn't change
				if previousStatus.GetDeployedVersionTimestamp() != tc.lookup.Status.GetDeployedVersionTimestamp() {
					t.Errorf("expected timestamp %q but got %q",
						previousStatus.GetDeployedVersionTimestamp(), tc.lookup.Status.GetDeployedVersionTimestamp())
				}
			}
		})
	}
}
