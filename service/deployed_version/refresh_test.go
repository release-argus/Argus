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

//go:build unit

package deployedver

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestBasicAuthFromString(t *testing.T) {
	testLogging()
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
	testLogging()
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
	testLogging()
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
			want: &Lookup{
				AllowInvalidCerts: boolPtr(false),

				URL:     test.URL,
				JSON:    test.JSON,
				Options: test.Options,
				Status:  test.Status},
		},
		"basic auth": {
			basicAuth: stringPtr(`{"username": "foo", "password": "bar"}`),
			previous:  testLookup(),
			want: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "foo",
					Password: "bar"},

				URL:               test.URL,
				AllowInvalidCerts: test.AllowInvalidCerts,
				JSON:              test.JSON,
				Options:           test.Options,
				Status:            test.Status},
		},
		"headers": {
			headers: stringPtr(`[{"key": "bish", "value": "bash"}, {"key": "bosh", "value": "bosh"}]`),

			previous: testLookup(),
			want: &Lookup{
				Headers: []Header{
					{Key: "bish", Value: "bash"},
					{Key: "bosh", Value: "bosh"}},

				URL:               test.URL,
				AllowInvalidCerts: test.AllowInvalidCerts,
				JSON:              test.JSON,
				Options:           test.Options,
				Status:            test.Status},
		},
		"json": {
			json: stringPtr("bish"),

			previous: testLookup(),
			want: &Lookup{
				JSON: "bish",

				URL:               test.URL,
				AllowInvalidCerts: test.AllowInvalidCerts,
				Options:           test.Options,
				Status:            test.Status},
		},
		"regex": {
			regex: stringPtr("bish"),

			previous: testLookup(),
			want: &Lookup{
				Regex: "bish",

				URL:               test.URL,
				AllowInvalidCerts: test.AllowInvalidCerts,
				JSON:              test.JSON,
				Options:           test.Options,
				Status:            test.Status},
		},
		"semantic versioning": {
			semanticVersioning: stringPtr("false"),

			previous: testLookup(),
			want: &Lookup{
				Options: &opt.Options{
					SemanticVersioning: boolPtr(false),
					Defaults:           test.Options.Defaults,
					HardDefaults:       test.Options.HardDefaults,
				},

				URL:               test.URL,
				AllowInvalidCerts: test.AllowInvalidCerts,
				JSON:              test.JSON,
				Status:            test.Status},
		},
		"url": {
			url: stringPtr("https://valid.release-argus.io/json"),

			previous: testLookup(),
			want: &Lookup{
				URL: "https://valid.release-argus.io/json",

				AllowInvalidCerts: test.AllowInvalidCerts,
				JSON:              test.JSON,
				Options:           test.Options,
				Status:            test.Status},
		},
		"override with invalid (empty) url": {
			url:      stringPtr(""),
			previous: testLookup(),
			want:     nil,
			errRegex: "url: <missing>",
		},
		"override with invalid regex": {
			regex:    stringPtr("v([0-9))"),
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
			if tc.want.String() != got.String() {
				t.Errorf("expected:\n%v\nbut got:\n%v", tc.want, got)
			}
		})
	}
}

func TestLookup_Refresh(t *testing.T) {
	testLogging()
	test := testLookup()
	testVersion, _ := test.Query(&util.LogFrom{Primary: "TestRefresh"})
	if testVersion == "" {
		t.Fatalf("test version is empty")
	}

	// GIVEN a Lookup and various json strings to override parts of it
	tests := map[string]struct {
		allowInvalidCerts  *string
		basicAuth          *string
		headers            *string
		json               *string
		regex              *string
		semanticVersioning *string
		url                *string
		previous           *Lookup
		previousStatus     svcstatus.Status
		errRegex           string
		want               string
	}{
		"Change of URL": {
			url:      stringPtr("https://valid.release-argus.io/json"),
			previous: testLookup(),
			want:     testVersion,
		},
		"Removal of URL": {
			url:      stringPtr(""),
			previous: testLookup(),
			errRegex: "url: <missing>",
			want:     "",
		},
		"Change of a few vars": {
			json:               stringPtr("otherVersion"),
			semanticVersioning: stringPtr("false"),
			previous:           testLookup(),
			want:               testVersion + "-beta",
		},
		"Change of vars that fail Query": {
			allowInvalidCerts: stringPtr("false"),
			previous:          testLookup(),
			errRegex:          `x509 \(certificate invalid\)`,
		},
		"Refresh new version": {
			previous: &Lookup{
				URL:               test.URL,
				AllowInvalidCerts: test.AllowInvalidCerts,
				JSON:              test.JSON,
				Options:           test.Options,
				Status: &svcstatus.Status{
					ServiceID:                stringPtr("Refresh new version"),
					DeployedVersion:          "0.0.0",
					DeployedVersionTimestamp: time.Now().UTC().Format(time.RFC3339),
				},
				Defaults:     test.Defaults,
				HardDefaults: test.HardDefaults},
			want: testVersion,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Copy tc.PreviousStatus
			previousStatus := tc.previousStatus

			// WHEN we call Refresh
			got, err := tc.previous.Refresh(
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
			// AND we get the expected result otherwise
			if tc.want != got {
				t.Errorf("expected %q but got %q", tc.want, got)
			}
			// AND the timestamp only changes if the version changed
			if previousStatus.DeployedVersionTimestamp != "" {
				// If the possible query-changing overrides are nil
				if tc.headers == nil && tc.json == nil && tc.regex == nil && tc.semanticVersioning == nil && tc.url == nil {
					// The timestamp should change only if the version changed
					if previousStatus.DeployedVersion != tc.previous.Status.DeployedVersion &&
						previousStatus.DeployedVersionTimestamp == tc.previous.Status.DeployedVersionTimestamp {
						t.Errorf("expected timestamp to change from %q, but got %q",
							previousStatus.DeployedVersionTimestamp, tc.previous.Status.DeployedVersionTimestamp)
						// The timestamp shouldn't change as the version didn't change
					} else if previousStatus.DeployedVersionTimestamp != tc.previous.Status.DeployedVersionTimestamp {
						t.Errorf("expected timestamp %q but got %q",
							previousStatus.DeployedVersionTimestamp, tc.previous.Status.DeployedVersionTimestamp)
					}
					// If the overrides are not nil
				} else {
					// The timestamp shouldn't change
					if previousStatus.DeployedVersionTimestamp != tc.previous.Status.DeployedVersionTimestamp {
						t.Errorf("expected timestamp %q but got %q",
							previousStatus.DeployedVersionTimestamp, tc.previous.Status.DeployedVersionTimestamp)
					}
				}
			}
		})
	}
}
