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
	svcStatus := &status.Status{}
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
	l.Init(options, svcStatus, defaults, hardDefaults)

	// THEN the fields are initialised correctly.
	if l.Options != options {
		t.Errorf("%s\nunexpected Options\nwant: %v\ngot:  %v",
			packageName, options, l.Options)
	}
	if l.Status != svcStatus {
		t.Errorf("%s\nunexpected Status\nwant: %v\ngot:  %v",
			packageName, svcStatus, l.Status)
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
	l.Init(options, svcStatus, defaults, hardDefaults)

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

func TestGetServiceID(t *testing.T) {
	// GIVEN a Lookup with a Status containing a ServiceID.
	serviceID := "foo"
	l := &testLookup{
		Lookup: Lookup{
			Status: &status.Status{}}}
	l.Status.ServiceInfo.ID = serviceID

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
	svcStatus := &status.Status{}
	l := &testLookup{
		Lookup: Lookup{
			Status: svcStatus}}

	// WHEN GetStatus is called.
	got := l.GetStatus()

	// THEN the Status is returned.
	if svcStatus != got {
		t.Errorf("%s\nwant: %v\ngot:  %v",
			packageName, svcStatus, got)
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
	// GIVEN a Lookup.
	tests := map[string]struct {
		url string
	}{
		"URL 1": {
			url: "https://example.com",
		},
		"URL 2": {
			url: "https://release-argus.io",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup{}
			lookup.URL = tc.url

			// WHEN ServiceURL is called.
			got := lookup.ServiceURL()

			// THEN the result is as expected.
			if tc.url != got {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.url, got)
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
