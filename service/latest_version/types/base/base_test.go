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

//go:build unit

// Package base provides the base struct for latest_version lookups.
package base

import (
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestInit(t *testing.T) {
	// GIVEN a Lookup and its dependencies.
	options := &opt.Options{}
	status := &status.Status{}
	defaults := &Defaults{
		Require: filter.RequireDefaults{
			Docker: filter.DockerCheckDefaults{
				Type: "ghcr"}}}
	hardDefaults := &Defaults{}
	wantRequireRegexContent := "foo"
	require := &filter.Require{
		RegexContent: wantRequireRegexContent}
	l := &Lookup{
		Require: require,
	}

	// WHEN Init is called.
	l.Init(options, status, defaults, hardDefaults)

	// THEN the fields are initialised correctly.
	if l.Options != options {
		t.Errorf("%s\nunexpected Options\nwant: %v\ngot:  %v",
			packageName, options, l.Options)
	}
	if l.Status != status {
		t.Errorf("%s\nunexpected Status\nwant: %v\ngot:  %v",
			packageName, status, l.Status)
	}
	if l.Defaults != defaults {
		t.Errorf("%s\nunexpected Defaults\nwant: %v\ngot:  %v",
			packageName, defaults, l.Defaults)
	}
	if l.HardDefaults != hardDefaults {
		t.Errorf("%s\nunexpected HardDefaults\nwant: %v\ngot:  %v",
			packageName, hardDefaults, l.HardDefaults)
	}

	// AND the Require field is initialised correctly.
	if l.Require == nil {
		t.Fatalf("%s\nRequire should not be nil",
			packageName)
	}
	if l.Require.RegexContent != wantRequireRegexContent {
		t.Errorf("%s\nunexpected Require.RegexContent\nwant: %q\ngot:  %q",
			packageName, wantRequireRegexContent, l.Require.RegexContent)
	}

	// GIVEN a Lookup with an empty Require.
	l = &Lookup{
		Require: &filter.Require{},
	}

	// WHEN Init is called.
	l.Init(options, status, defaults, hardDefaults)

	// THEN the Require field is set to nil.
	if l.Require != nil {
		t.Errorf("%s\nRequire should be nil when empty",
			packageName)
	}
}

func TestString(t *testing.T) {
	// GIVEN a Lookup and a parentLookup.
	parentLookup := &testLookup{
		Lookup: Lookup{
			Type: "foo",
			URL:  "https://example.com/other"}}
	l := &testLookup{
		Lookup: Lookup{
			Type: "test",
			URL:  "https://example.com"}}
	prefix := "  "

	// WHEN String is called.
	got := l.String(parentLookup, prefix)

	// THEN the result is as expected.
	want := util.ToYAMLString(parentLookup, prefix)
	if got != want {
		t.Errorf("%s\nwant: %q\ngot:  %q",
			packageName, want, got)
	}
}

func TestIsEqual(t *testing.T) {
	// GIVEN two Lookups.
	tests := map[string]struct {
		a, b Interface
		want bool
	}{
		"empty": {
			a:    &testLookup{},
			b:    &testLookup{},
			want: true,
		},
		"defaults ignored": {
			a: &testLookup{
				Lookup: Lookup{
					Defaults: &Defaults{
						AccessToken: "foo"}}},
			b:    &testLookup{},
			want: true,
		},
		"hard_defaults ignored": {
			a: &testLookup{
				Lookup: Lookup{
					HardDefaults: &Defaults{
						AccessToken: "foo"}}},
			b:    &testLookup{},
			want: true,
		},
		"equal": {
			a: &testLookup{
				Lookup: Lookup{
					Type: "github",
					URL:  "release-argus/Argus",
					Require: &filter.Require{
						RegexContent: "foo.tar.gz"}}},
			b: &testLookup{
				Lookup: Lookup{
					Type: "github",
					URL:  "release-argus/Argus",
					Require: &filter.Require{
						RegexContent: "foo.tar.gz"}}},
			want: true,
		},
		"not equal": {
			a: &testLookup{
				Lookup: Lookup{
					Type: "github",
					URL:  "release-argus/Argus",
					Require: &filter.Require{
						RegexContent: "foo.tar.gz"}}},
			b: &testLookup{
				Lookup: Lookup{
					Type: "github",
					URL:  "release-argus/Argus",
					Require: &filter.Require{
						RegexContent: "foo.tar.gz-"}}},
			want: false,
		},
		"not equal with nil": {
			a: &testLookup{
				Lookup: Lookup{URL: "release-argus/Argus"}},
			b:    nil,
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Set Status vars just to ensure they are not printed.
			status := &status.Status{}
			status.Init(
				0, 0, 0,
				&name, nil,
				test.StringPtr("http://example.com"))
			status.SetLatestVersion("foo", "", false)
			tc.a.(*testLookup).Status = status

			// WHEN the two Lookups are compared.
			got := tc.a.IsEqual(tc.a, tc.b)

			// THEN the result is as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.want, got)
			}
		})
	}
}

func TestGetServiceID(t *testing.T) {
	// GIVEN a Lookup with a Status containing a ServiceID.
	serviceID := "foo"
	l := &testLookup{
		Lookup: Lookup{
			Status: &status.Status{
				ServiceID: test.StringPtr(serviceID)}}}

	// WHEN GetService is called.
	got := l.GetServiceID()

	// THEN the ServiceID is returned.
	if serviceID != got {
		t.Errorf("%s\nwant: %q\ngot:  %q",
			packageName, serviceID, got)
	}
}

func TestGetType(t *testing.T) {
	// GIVEN a Lookup with a Type.
	lookupType := "test"
	l := &testLookup{
		Lookup: Lookup{Type: lookupType}}

	// WHEN GetType is called.
	got := l.GetType()

	// THEN the Type is returned.
	if lookupType != got {
		t.Errorf("%s\nwant: %q\ngot:  %q",
			packageName, lookupType, got)
	}
}

func TestGetOptions(t *testing.T) {
	// GIVEN a Lookup with Options.
	options := &opt.Options{}
	l := &testLookup{
		Lookup: Lookup{
			Options: options}}

	// WHEN GetOptions is called.
	got := l.GetOptions()

	// THEN the Options are returned.
	if options != got {
		t.Errorf("%s\nwant: %v\ngot:  %v",
			packageName, options, got)
	}
}

func TestGetRequite(t *testing.T) {
	// GIVEN a Lookup with Require.
	require := &filter.Require{}
	l := &testLookup{
		Lookup: Lookup{Require: require}}

	// WHEN GetRequire is called.
	got := l.GetRequire()

	// THEN the Require is returned.
	if require != got {
		t.Errorf("%s\nwant: %v\ngot:  %v",
			packageName, require, got)
	}
}

func TestGetStatus(t *testing.T) {
	// GIVEN a Lookup with Status.
	status := &status.Status{}
	l := &testLookup{
		Lookup: Lookup{
			Status: status}}

	// WHEN GetStatus is called.
	got := l.GetStatus()

	// THEN the Status is returned.
	if status != got {
		t.Errorf("%s\nwant: %v\ngot:  %v",
			packageName, status, got)
	}
}

func TestGetDefaults(t *testing.T) {
	// GIVEN a Lookup with Defaults.
	defaults := &Defaults{}
	l := &testLookup{
		Lookup: Lookup{Defaults: defaults}}

	// WHEN GetDefaults is called.
	got := l.GetDefaults()

	// THEN the Defaults are returned.
	if defaults != got {
		t.Errorf("%s\nwant: %v\ngot:  %v",
			packageName, defaults, got)
	}
}

func TestGetHardDefaults(t *testing.T) {
	// GIVEN a Lookup with HardDefaults.
	hardDefaults := &Defaults{}
	l := &testLookup{
		Lookup: Lookup{HardDefaults: hardDefaults}}

	// WHEN GetHardDefaults is called.
	got := l.GetHardDefaults()

	// THEN the HardDefaults are returned.
	if hardDefaults != got {
		t.Errorf("%s\nwant: %v\ngot:  %v",
			packageName, hardDefaults, got)
	}
}

func TestServiceURL(t *testing.T) {
	type lookupArgs struct {
		webURL        string
		latestVersion string
	}
	testURL := "https://example.com"
	testWebURL := "https://example.com/release"
	testWebURLTemplate := "https://example.com/release/{{ version }}"

	// GIVEN a Lookup.
	tests := map[string]struct {
		lookupArgs   lookupArgs
		ignoreWebURL bool
		expectedURL  string
	}{
		"URL without WebURL": {
			ignoreWebURL: false,
			expectedURL:  testURL,
		},
		"URL with WebURL and ignoreWebURL false": {
			lookupArgs: lookupArgs{
				webURL: testWebURL},
			ignoreWebURL: false,
			expectedURL:  testWebURL,
		},
		"URL with WebURL and ignoreWebURL true": {
			lookupArgs: lookupArgs{
				webURL: testWebURL},
			ignoreWebURL: true,
			expectedURL:  testURL,
		},
		"URL with WebURL containing version template and no latest version": {
			lookupArgs: lookupArgs{
				webURL: testWebURLTemplate},
			ignoreWebURL: false,
			expectedURL:  testURL,
		},
		"URL with WebURL containing version template and latest version": {
			lookupArgs: lookupArgs{
				webURL:        testWebURLTemplate,
				latestVersion: "1.0.0"},
			ignoreWebURL: false,
			expectedURL:  testWebURL + "/1.0.0",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup{}
			lookup.URL = testURL
			lookup.Status = &status.Status{}
			lookup.Status.WebURL = &tc.lookupArgs.webURL
			lookup.Status.SetLatestVersion(tc.lookupArgs.latestVersion, "", false)

			// WHEN ServiceURL is called.
			got := lookup.ServiceURL(tc.ignoreWebURL)

			// THEN the result is as expected.
			if tc.expectedURL != got {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.expectedURL, got)
			}
		})
	}
}

func TestCheckValues(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		yamlStr  string
		errRegex string
	}{
		"no URL": {
			yamlStr: test.TrimYAML(`
				type: url
			`),
		},
		"valid URLCommands": {
			yamlStr: test.TrimYAML(`
				type: url
				url: https://example.com
				url_commands:
					- type: regex
						regex: foo
			`),
		},
		"invalid URLCommands": {
			yamlStr: test.TrimYAML(`
				type: url
				url: https://example.com
				url_commands:
					- type: foo
			`),
			errRegex: test.TrimYAML(`
				^url_commands:
					- item_0:
						type: "foo" <invalid>.*$`),
		},
		"valid Require": {
			yamlStr: test.TrimYAML(`
				type: url
				url: https://example.com
				require:
					regex_version: v.+
			`),
		},
		"invalid Require": {
			yamlStr: test.TrimYAML(`
				type: url
				url: https://example.com
				require:
					regex_version: "[0-"
			`),
			errRegex: test.TrimYAML(`
				^require:
					regex_version: "[^"]+" <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			l := &testLookup{}
			// apply the YAML.
			if err := yaml.Unmarshal([]byte(tc.yamlStr), l); err != nil {
				t.Fatalf("%s\nerror unmarshalling YAML: %v",
					packageName, err)
			}

			// WHEN CheckValues is called.
			err := l.CheckValues("")

			// THEN it errors when expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Errorf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
					packageName, wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
		})
	}
}

func checkDockerToken(
	t *testing.T,
	wantQueryToken, gotQueryToken string,
	wantValidUntil, gotValidUntil time.Time,
	message string,
) {
	if gotQueryToken != wantQueryToken {
		t.Errorf("%s\nRequire.Docker.queryToken %s\nwant: %q\ngot:  %q",
			packageName, message, wantQueryToken, gotQueryToken)
	}
	if gotValidUntil != wantValidUntil {
		t.Errorf("%s\nRequire.Docker.validUntil %s\nwant: %q\ngot:  %q",
			packageName, message, wantValidUntil, gotValidUntil)
	}
}

func TestInherit(t *testing.T) {
	// GIVEN two Lookups with/without Require.
	tests := map[string]struct {
		overrides          string
		inheritDockerToken bool
	}{
		"no overrides - inherits token": {
			inheritDockerToken: true,
		},
		"differing Require.Docker - does not inherit token": {
			overrides: test.TrimYAML(`
				require:
					docker:
						image: release-argus/test
			`),
			inheritDockerToken: false,
		},
		"removed Require.Docker": {
			overrides: test.TrimYAML(`
				require:
					docker: null
			`),
			inheritDockerToken: false,
		},
		"Removed Require": {
			overrides: test.TrimYAML(`
				require: null
			`),
			inheritDockerToken: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			toLookup := &testLookup{}
			toLookup.Require = &filter.Require{
				Docker: filter.NewDockerCheck(
					"ghcr",
					"release-argus/argus", "{{ version }}",
					"ghcr-username", "ghcr-token",
					"ghcr-query-token", time.Now(),
					nil)}
			toLookup.Require.Docker.SetQueryToken(
				"",
				"", time.Time{})
			fromLookup := &testLookup{}
			fromLookup.Require = &filter.Require{
				Docker: filter.NewDockerCheck(
					"ghcr",
					"release-argus/argus", "{{ version }}",
					"ghcr-username", "ghcr-token",
					util.SecretValue, time.Time(time.Now().Add(time.Hour)),
					nil)}
			err := yaml.Unmarshal([]byte(tc.overrides), fromLookup)
			if err != nil {
				t.Fatalf("%s\nerror unmarshalling overrides: %v",
					packageName, err)
			}

			// WHEN Inherit is called.
			toLookup.Inherit(fromLookup)

			// THEN the Docker.(QueryToken|ValidUntil) are copied over when expected.
			if tc.inheritDockerToken {
				if toLookup.Require == nil || toLookup.Require.Docker == nil ||
					fromLookup.Require == nil || fromLookup.Require.Docker == nil {
					t.Fatalf("%s\nunexpected nil values\ntoLookup.Require.Docker, nil=%t\nfromLookup.Require(.Docker, nil=%t",
						packageName,
						toLookup.Require == nil || toLookup.Require.Docker == nil,
						fromLookup.Require == nil || fromLookup.Require.Docker == nil)
				}
				gotQueryToken, gotValidUntil := toLookup.Require.Docker.CopyQueryToken()
				wantQueryToken, wantValidUntil := fromLookup.Require.Docker.CopyQueryToken()
				checkDockerToken(t,
					wantQueryToken, gotQueryToken,
					wantValidUntil, gotValidUntil,
					"not copied")
			} else if toLookup.Require != nil && toLookup.Require.Docker != nil {
				gotQueryToken, gotValidUntil := toLookup.Require.Docker.CopyQueryToken()
				wantQueryToken, wantValidUntil := "", time.Time{}
				checkDockerToken(t,
					wantQueryToken, gotQueryToken,
					wantValidUntil, gotValidUntil,
					"should not be copied")
			}
		})
	}
}

func TestQuery(t *testing.T) {
	// GIVEN a Lookup.
	l := &testLookup{
		Lookup: Lookup{
			Type: "test",
			URL:  "https://example.com"}}

	// WHEN Query is called.
	gotBool, gotErr := l.Query(true, logutil.LogFrom{})

	// THEN the function returns false and an error as it is not implemented.
	want := false
	if gotBool != want {
		t.Errorf("%s\nunexpected return value\nwant: %t\ngot:  %t",
			packageName, want, gotBool)
	}
	if gotErr == nil {
		t.Errorf("%s\nunexpected nil error",
			packageName)
	}
}
