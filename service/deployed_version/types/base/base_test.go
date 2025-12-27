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

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestInit(t *testing.T) {
	// GIVEN a Lookup and its dependencies.
	options := &opt.Options{}
	svcStatus := &status.Status{}
	defaults := &Defaults{}
	hardDefaults := &Defaults{}
	l := &Lookup{}

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
}

func TestString(t *testing.T) {
	// GIVEN a Lookup and a parentLookup.
	parentLookup := &testLookup{
		Lookup: Lookup{
			Type: "foo"}}
	l := &testLookup{
		Lookup: Lookup{
			Type: "test"}}
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
			Status: &status.Status{
				ServiceInfo: serviceinfo.ServiceInfo{
					ID: serviceID}}}}

	// WHEN GetService is called.
	got := l.GetServiceID()

	// THEN the ServiceID is returned.
	if got != serviceID {
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
	if got != lookupType {
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
	if got != options {
		t.Errorf("%s\nwant: %v\ngot:  %v",
			packageName, options, got)
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
	if got != svcStatus {
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
	if got != defaults {
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
	if got != hardDefaults {
		t.Errorf("%s\nwant: %v\ngot:  %v",
			packageName, hardDefaults, got)
	}
}

func TestTrack(t *testing.T) {
	// GIVEN a Lookup.
	l := &testLookup{
		Lookup: Lookup{
			Type: "test"},
	}

	// WHEN Track is called.
	l.Track()

	// THEN no error occurs and nothing is tracked.
	// As the function does nothing, we just ensure it doesn't panic or error.
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
			errRegex: `^$`,
		},
		"have URL": {
			yamlStr: test.TrimYAML(`
				type: url
				url: release-argus/argus
			`),
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			l := &testLookup{}
			// Apply the YAML.
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
				t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
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

func TestQuery(t *testing.T) {
	// GIVEN a Lookup.
	l := &testLookup{
		Lookup: Lookup{
			Type: "test"},
	}

	// WHEN Query is called.
	gotErr := l.Query(true, logutil.LogFrom{})

	// THEN the function returns an error as it is not implemented.
	if gotErr == nil {
		t.Errorf("%s\nunexpected nil error",
			packageName)
	}
}

func TestInheritSecrets(t *testing.T) {
	// GIVEN a Lookup and another Lookup to inherit secrets from.
	otherLookup := &testLookup{
		Lookup: Lookup{
			Type: "other"}}
	secretRefs := &shared.VSecretRef{
		Headers: []shared.OldIntIndex{
			{OldIndex: test.IntPtr(0)}},
	}
	l := &testLookup{
		Lookup: Lookup{
			Type: "test"}}
	strBefore := l.String(l, "")

	// WHEN InheritSecrets is called.
	l.InheritSecrets(otherLookup, secretRefs)

	// THEN no secrets are inherited as the function does nothing.
	// As the function does nothing, we just ensure it doesn't panic or error.
	if strAfter := l.String(l, ""); strBefore != strAfter {
		t.Errorf("%s\nunexpected change in Lookup\nbefore: %q\nafter:  %q",
			packageName, strBefore, l.String(l, ""))
	}
}
