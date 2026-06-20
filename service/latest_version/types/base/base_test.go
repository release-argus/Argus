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

// Package base provides the base struct for latest_version lookups.
package base

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
)

func TestLookup_IsZero(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		name string
		data Lookup
		want bool
	}{
		{
			name: "empty",
			data: Lookup{},
			want: true,
		},
		{
			name: "non-empty/Type",
			data: Lookup{
				Type: "abc",
			},
			want: false,
		},
		{
			name: "non-empty/URL",
			data: Lookup{
				URL: "https://example.com",
			},
			want: false,
		},
		{
			name: "non-empty/URLCommands",
			data: Lookup{
				URLCommands: filter.URLCommands{
					{Type: "regex", Regex: "[0-9.]+"},
				},
			},
			want: false,
		},
		{
			name: "non-empty/Require",
			data: Lookup{
				Require: &filter.Require{
					RegexVersion: "[0-9.]+",
				},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: Lookup{
				Type: "abc",
				URL:  "https://example.com",
				URLCommands: filter.URLCommands{
					{Type: "regex", Regex: "[0-9.]+"},
				},
				Require: &filter.Require{
					RegexVersion: "[0-9.]+",
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on it.
			got := tc.data.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nLookup.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLookup_Init(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	// GIVEN: a Lookup and its dependencies.
	options := &opt.Options{}
	svcStatus := &status.Status{}
	wantRequireRegexContent := "foo"
	require := &filter.Require{
		RegexContent: wantRequireRegexContent,
	}
	l := &Lookup{
		Require: require,
	}

	// WHEN: Init is called.
	l.Init(options, svcStatus, lvCfg)

	prefix := fmt.Sprintf(
		"%s\nLookup.Init(options=%p, status=%p, defaults=%v)",
		packageName, options, svcStatus, lvCfg,
	)

	// THEN: pointers to those vars are handed out to the Lookup.
	fieldTests := []test.FieldAssertion{
		{Name: "Options", Got: l.Options, Want: options, Mode: test.CompareSamePointer},
		{Name: "Status", Got: l.Status, Want: svcStatus, Mode: test.CompareSamePointer},
		{Name: "Defaults", Got: l.Defaults, Want: lvCfg.Soft, Mode: test.CompareSamePointer},
		{Name: "HardDefaults", Got: l.HardDefaults, Want: lvCfg.Hard, Mode: test.CompareSamePointer},
	}
	if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
		t.Fatal(err)
	}

	// AND: the Require field is initialised correctly.
	if l.Require == nil {
		t.Fatalf("%s .Require should not be nil on Lookup", prefix)
	}
	if l.Require.RegexContent != wantRequireRegexContent {
		t.Errorf(
			"%s .Require.RegexContent was not handed to the Lookup correctly\ngot:  %q\nwant: %q",
			prefix, wantRequireRegexContent, l.Require.RegexContent,
		)
	}

	// GIVEN: a Lookup with an empty Require.
	l = &Lookup{
		Require: &filter.Require{},
	}

	// WHEN: Init is called.
	l.Init(options, svcStatus, lvCfg)

	// THEN: the Require field is set to nil.
	if l.Require != nil {
		t.Errorf(
			"%s (Require is empty), gave Lookup a non-nil Require\ngot:  %+v\nwant: nil",
			prefix, l.Require,
		)
	}
}

func TestLookup_GetServiceID(t *testing.T) {
	// GIVEN: a Lookup with a Status containing a ServiceID.
	serviceID := "foo"
	l := &testLookup{
		Lookup: Lookup{
			Status: &status.Status{},
		},
	}
	l.Status.ServiceInfo.ID = serviceID

	// WHEN: GetService is called.
	got := l.GetServiceID()

	// THEN: the ServiceID is returned.
	if serviceID != got {
		t.Errorf(
			"%s\nLookup.GetServiceID() value mismatch\ngot:  %q\nwant: %q",
			packageName, got, serviceID,
		)
	}
}

func TestLookup_GetType(t *testing.T) {
	// GIVEN: a Lookup with a Type.
	lookupType := "test"
	l := &testLookup{
		Lookup: Lookup{Type: lookupType},
	}
	want := "-"

	// WHEN: GetType is called.
	got := l.GetType()

	// THEN: an empty string is returned.
	if got != want {
		t.Errorf(
			"%s\nLookup.GetType() value mismatch\ngot:  %q\nwant: %q",
			packageName, got, want,
		)
	}
}

func TestLookup_GetOptions(t *testing.T) {
	// GIVEN: a Lookup with Options.
	options := &opt.Options{}
	l := &testLookup{
		Lookup: Lookup{
			Options: options,
		},
	}

	// WHEN: GetOptions is called.
	got := l.GetOptions()

	// THEN: the Options are returned.
	if options != got {
		t.Errorf(
			"%s\nLookup.GetOptions() pointer mismatch\ngot:  %p\nwant: %p",
			packageName, got, options,
		)
	}
}

func TestLookup_GetRequite(t *testing.T) {
	// GIVEN: a Lookup with Require.
	require := &filter.Require{}
	l := &testLookup{
		Lookup: Lookup{Require: require},
	}

	// WHEN: GetRequire is called.
	got := l.GetRequire()

	// THEN: the Require is returned.
	if require != got {
		t.Errorf(
			"%s\nLookup.GetRequire() pointer mismatch\ngot:  %p\nwant: %p",
			packageName, got, require,
		)
	}
}

func TestLookup_GetStatus(t *testing.T) {
	// GIVEN: a Lookup with Status.
	svcStatus := &status.Status{}
	l := &testLookup{
		Lookup: Lookup{
			Status: svcStatus,
		},
	}

	// WHEN: GetStatus is called.
	got := l.GetStatus()

	// THEN: the Status is returned.
	if svcStatus != got {
		t.Errorf(
			"%s\nLookup.GetStatus() pointer mismatch\ngot:  %p\nwant: %p",
			packageName, got, svcStatus,
		)
	}
}

func TestLookup_SetStatus(t *testing.T) {
	// GIVEN: a Lookup and a Status.
	l := &Lookup{}
	svcStatus := &status.Status{}

	// WHEN: SetStatus is called.
	l.SetStatus(svcStatus)

	// THEN: the Status is set.
	if l.Status != svcStatus {
		t.Errorf(
			"%s\nLookup.SetStatus(%p) pointer mismatch\ngot:  %v\nwant: %v",
			packageName, svcStatus,
			l.Status, svcStatus,
		)
	}

	// ---

	// GIVEN: a new Status.
	svcStatus = &status.Status{}

	// WHEN: SetStatus is called.
	l.SetStatus(svcStatus)

	// THEN: the Status is set.
	if l.Status != svcStatus {
		t.Errorf(
			"%s\nLookup.SetStatus(%p) pointer mismatch\ngot:  %v\nwant: %v",
			packageName, svcStatus,
			l.Status, svcStatus,
		)
	}
}

func TestLookup_GetDefaults(t *testing.T) {
	// GIVEN: a Lookup with Defaults.
	defaults := &Defaults{}
	l := &testLookup{
		Lookup: Lookup{Defaults: defaults},
	}

	// WHEN: GetDefaults is called.
	got := l.GetDefaults()

	// THEN: the Defaults are returned.
	if defaults != got {
		t.Errorf(
			"%s\nLookup.GetDefaults() pointer mismatch\ngot:  %p\nwant: %p",
			packageName, got, defaults,
		)
	}
}

func TestLookup_GetHardDefaults(t *testing.T) {
	// GIVEN: a Lookup with HardDefaults.
	hardDefaults := &Defaults{}
	l := &testLookup{
		Lookup: Lookup{HardDefaults: hardDefaults},
	}

	// WHEN: GetHardDefaults is called.
	got := l.GetHardDefaults()

	// THEN: the HardDefaults are returned.
	if hardDefaults != got {
		t.Errorf(
			"%s\nLookup.GetHardDefaults() pointer mismatch\ngot:  %p\nwant: %p",
			packageName, got, hardDefaults,
		)
	}
}

func TestLookup_ServiceURL(t *testing.T) {
	// GIVEN: a Lookup.
	tests := []struct {
		url string
	}{
		{url: "https://example.com"},
		{url: "https://release-argus.io"},
	}

	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup{}
			lookup.URL = tc.url

			// WHEN: ServiceURL is called.
			got := lookup.ServiceURL()

			// THEN: the result is as expected.
			if tc.url != got {
				t.Errorf(
					"%s\nLookup.ServiceURL() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.url,
				)
			}
		})
	}
}

func TestLookup_CheckValues(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	// GIVEN: a Lookup.
	tests := []struct {
		name     string
		data     string
		errRegex string
	}{
		{
			name: "no URL",
			data: `type: url`,
		},
		{
			name: "valid URLCommands",
			data: test.TrimYAML(`
				type: url
				url: https://example.com
				url_commands:
					- type: regex
						regex: foo
			`),
		},
		{
			name: "invalid URLCommands",
			data: test.TrimYAML(`
				type: url
				url: https://example.com
				url_commands:
					- type: foo
			`),
			errRegex: test.TrimYAML(`
				^url_commands:
					- item_0:
						type: "foo" <invalid>.*$`,
			),
		},
		{
			name: "valid Require",
			data: test.TrimYAML(`
				type: url
				url: https://example.com
				require:
					regex_version: v.+
			`),
		},
		{
			name: "invalid Require",
			data: test.TrimYAML(`
				type: url
				url: https://example.com
				require:
					regex_version: "[0-"
			`),
			errRegex: test.TrimYAML(`
				^require:
					regex_version: "[^"]+" <invalid>.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input, err := decodeTestLookup(
				t,
				"yaml", []byte(tc.data),
				nil,
				nil,
				lvCfg,
			)
			// Apply the YAML.
			if err != nil {
				t.Fatalf(
					"%s\ndecodeTestLookup(%q) failed: %v",
					packageName, tc.data,
					err,
				)
			}

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				input.CheckValues,
			)
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
		t.Errorf(
			"%s\nRequire.Docker.queryToken %s\ngot:  %q\nwant: %q",
			packageName, message,
			gotQueryToken, wantQueryToken,
		)
	}
	if gotValidUntil != wantValidUntil {
		t.Errorf(
			"%s\nRequire.Docker.validUntil %s\ngot:  %q\nwant: %q",
			packageName, message,
			gotValidUntil, wantValidUntil,
		)
	}
}

func TestLookup_InheritSecrets(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	// GIVEN: two Lookups with/without Require.
	tests := []struct {
		name               string
		dType              string
		overrides          string
		inheritDockerToken bool
	}{
		{
			name:               "no overrides, inherits token",
			dType:              "hub",
			inheritDockerToken: true,
		},
		{
			name:  "hub, differing Require.Docker.Image, does inherit token",
			dType: "hub",
			overrides: test.TrimYAML(`
				require:
					docker:
						image: release-argus/test
			`),
			inheritDockerToken: true,
		},
		{
			name:  "ghcr, differing Require.Docker.Image, does not inherit token",
			dType: "ghcr",
			overrides: test.TrimYAML(`
				require:
					docker:
						image: release-argus/test
			`),
			inheritDockerToken: false,
		},
		{
			name:  "removed Require.Docker",
			dType: "hub",
			overrides: test.TrimYAML(`
				require:
					docker: null
			`),
			inheritDockerToken: false,
		},
		{
			name:               "Removed Require",
			dType:              "hub",
			overrides:          `require: null`,
			inheritDockerToken: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			toLookup, _ := decodeTestLookup(
				t,
				"yaml", []byte(test.TrimYAML(`
					require:
						docker:
							type: `+tc.dType+`
							image: test/app
							tag: '{{ version }}'
							username: hub-username
							token: hub-token
				`)),
				nil, nil,
				lvCfg,
			)
			toLookup.Require.Docker.GetAuth().SetQueryToken(
				"", time.Time{},
			)
			fromLookup, _ := decodeTestLookup(
				t,
				"yaml", []byte(test.TrimYAML(`
					require:
						docker:
							type: `+tc.dType+`
							image: test/app
							tag: '{{ version }}'
							username: hub-username
							token: hub-token
				`)),
				nil, nil,
				lvCfg,
			)
			fromLookup.Require.Docker.GetAuth().SetQueryToken(
				util.SecretValue, time.Now().Add(time.Hour),
			)
			// Require.
			overridesRaw, _ := polymorphic.Extract("yaml", []byte(tc.overrides), "require")
			if decode.IsNull(overridesRaw) {
				fromLookup.Require = nil
			} else {
				err := UnmarshalRequire(
					"yaml", []byte(tc.overrides),
					fromLookup,
					fromLookup.Status,
					&lvCfg.Soft.Require,
				)
				if err != nil {
					t.Fatalf(
						"%s\nUnmarshalRequire(%q) failed: %v",
						packageName, tc.overrides,
						err,
					)
				}
			}

			// WHEN: InheritSecrets is called.
			toLookup.InheritSecrets(fromLookup, nil)

			// THEN: the Docker.(QueryToken|ValidUntil) are copied over when expected.
			if tc.inheritDockerToken {
				if toLookup.Require == nil || toLookup.Require.Docker == nil ||
					fromLookup.Require == nil || fromLookup.Require.Docker == nil {
					t.Fatalf(
						"%s\nLookup.InheritSecrets() unexpected nil values\n"+
							"toLookup.Require.Docker, nil=%t\n"+
							"fromLookup.Require.Docker, nil=%t",
						packageName,
						toLookup.Require == nil || toLookup.Require.Docker == nil,
						fromLookup.Require == nil || fromLookup.Require.Docker == nil,
					)
				}
				gotQueryToken, gotValidUntil := toLookup.Require.Docker.GetAuth().GetQueryTokenSelf()
				wantQueryToken, wantValidUntil := fromLookup.Require.Docker.GetAuth().GetQueryTokenSelf()
				checkDockerToken(
					t,
					wantQueryToken, gotQueryToken,
					wantValidUntil, gotValidUntil,
					"not copied",
				)
			} else if toLookup.Require != nil && toLookup.Require.Docker != nil {
				gotQueryToken, gotValidUntil := toLookup.Require.Docker.GetAuth().GetQueryTokenSelf()
				wantQueryToken, wantValidUntil := "", time.Time{}
				checkDockerToken(
					t,
					wantQueryToken, gotQueryToken,
					wantValidUntil, gotValidUntil,
					"should not be copied",
				)
			}
		})
	}
}

func TestLookup_Query(t *testing.T) {
	// GIVEN: a Lookup.
	l := &testLookup{
		Lookup: Lookup{
			Type: "test",
			URL:  "https://example.com",
		},
	}

	// WHEN: Query is called.
	gotBool, gotErr := l.Query(true, logx.LogFrom{})

	prefix := fmt.Sprintf("%s\nLookup.Query()", packageName)

	// THEN: the function returns false and an error as it is not implemented.
	if gotBool != false {
		t.Errorf(
			"%s unexpected value\ngot:  %t\nwant: %t",
			prefix, false, gotBool,
		)
	}
	if gotErr == nil {
		t.Errorf("%s unexpected nil error", prefix)
	}
}
