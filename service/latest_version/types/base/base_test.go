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
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
	"github.com/release-argus/Argus/util/polymorphic"
)

func TestLookup_IsZero(t *testing.T) {
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

			got := tc.data.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestLookup_Init(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	options := &opt.Options{}
	svcStatus := &status.Status{}
	wantRequireRegexContent := "foo"
	require := &filter.Require{
		RegexContent: wantRequireRegexContent,
	}
	l := &Lookup{
		Require: require,
	}

	l.Init(options, svcStatus, lvCfg)

	fieldTests := []test.FieldAssertion{
		{Name: "Options", Got: l.Options, Want: options, Mode: test.CompareSamePointer},
		{Name: "Status", Got: l.Status, Want: svcStatus, Mode: test.CompareSamePointer},
		{Name: "Defaults", Got: l.Defaults, Want: lvCfg.Soft, Mode: test.CompareSamePointer},
		{Name: "HardDefaults", Got: l.HardDefaults, Want: lvCfg.Hard, Mode: test.CompareSamePointer},
	}
	if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
		t.Fatalf("non-empty Require: %v", testErr)
	}

	if l.Require == nil {
		t.Fatal("non-empty Require: Lookup.Require mismatch\ngot:  nil\nwant: non-nil")
	}
	if l.Require.RegexContent != wantRequireRegexContent {
		t.Errorf(
			"non-empty Require: Lookup.Require.RegexContent mismatch\ngot:  %q\nwant: %q",
			l.Require.RegexContent, wantRequireRegexContent,
		)
	}

	l = &Lookup{
		Require: &filter.Require{},
	}

	l.Init(options, svcStatus, lvCfg)

	if l.Require != nil {
		t.Error("empty Require: Lookup.Require mismatch\ngot:  non-nil\nwant: nil")
	}
}

func TestLookup_GetServiceID(t *testing.T) {
	serviceID := "foo"
	l := &testLookup{
		Status: &status.Status{},
	}
	l.Status.ServiceInfo.ID = serviceID

	got := l.GetServiceID()

	if serviceID != got {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, serviceID)
	}
}

func TestLookup_GetType(t *testing.T) {
	l := &testLookup{
		Type: "test",
	}
	want := "-"

	got := l.GetType()

	if got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestLookup_GetOptions(t *testing.T) {
	options := &opt.Options{}
	l := &testLookup{
		Options: options,
	}

	got := l.GetOptions()

	if options != got {
		t.Errorf("mismatch\ngot:  %p\nwant: %p", got, options)
	}
}

func TestLookup_GetRequire(t *testing.T) {
	require := &filter.Require{}
	l := &testLookup{
		Require: require,
	}

	got := l.GetRequire()

	if require != got {
		t.Errorf("mismatch\ngot:  %p\nwant: %p", got, require)
	}
}

func TestLookup_GetStatus(t *testing.T) {
	svcStatus := &status.Status{}
	l := &testLookup{
		Status: svcStatus,
	}

	got := l.GetStatus()

	if svcStatus != got {
		t.Errorf("mismatch\ngot:  %p\nwant: %p", got, svcStatus)
	}
}

func TestLookup_SetStatus(t *testing.T) {
	l := &Lookup{}
	svcStatus := &status.Status{}

	l.SetStatus(svcStatus)

	if l.Status != svcStatus {
		t.Errorf("set: Lookup.Status mismatch\ngot:  %p\nwant: %p", l.Status, svcStatus)
	}

	// ---

	svcStatus = &status.Status{}

	l.SetStatus(svcStatus)

	if l.Status != svcStatus {
		t.Errorf("replaced: Lookup.Status mismatch\ngot:  %p\nwant: %p", l.Status, svcStatus)
	}
}

func TestLookup_GetDefaults(t *testing.T) {
	defaults := &Defaults{}
	l := &testLookup{
		Defaults: defaults,
	}

	got := l.GetDefaults()

	if defaults != got {
		t.Errorf("mismatch\ngot:  %p\nwant: %p", got, defaults)
	}
}

func TestLookup_GetHardDefaults(t *testing.T) {
	hardDefaults := &Defaults{}
	l := &testLookup{
		HardDefaults: hardDefaults,
	}

	got := l.GetHardDefaults()

	if hardDefaults != got {
		t.Errorf("mismatch\ngot:  %p\nwant: %p", got, hardDefaults)
	}
}

func TestLookup_ServiceURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "example.com", url: "https://example.com"},
		{name: "release-argus.io", url: "https://release-argus.io"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup{}
			lookup.URL = tc.url

			got := lookup.ServiceURL()

			if tc.url != got {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.url)
			}
		})
	}
}

func TestLookup_CheckValues(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	tests := []struct {
		name     string
		data     string
		errRegex string
	}{
		{
			name:     "no URL",
			data:     `type: url`,
			errRegex: `^$`,
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
			errRegex: `^$`,
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
			errRegex: `^$`,
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
				"yaml", []byte(tc.data),
				nil,
				nil,
				lvCfg,
			)
			if err != nil {
				t.Fatalf("decodeTestLookup() failed: %v", err)
			}

			_ = test.AssertCheckValuesWithError(
				t,
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
) {
	t.Helper()

	if gotQueryToken != wantQueryToken {
		t.Errorf(
			"Lookup.Require.Docker queryToken mismatch\ngot:  %q\nwant: %q",
			gotQueryToken, wantQueryToken,
		)
	}
	if !gotValidUntil.Equal(wantValidUntil) {
		t.Errorf(
			"Lookup.Require.Docker validUntil mismatch\ngot:  %s\nwant: %s",
			gotValidUntil, wantValidUntil,
		)
	}
}

func TestLookup_InheritSecrets(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

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
			inheritDockerToken: true, // Hub auth is username+token, so the image doesn't matter.
		},
		{
			name:  "ghcr, differing Require.Docker.Image, does not inherit token",
			dType: "ghcr",
			overrides: test.TrimYAML(`
				require:
					docker:
						image: release-argus/test
			`),
			inheritDockerToken: false, // GHCR query tokens are repo-scoped.
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

			toLookup, err := decodeTestLookup(
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
			if err != nil {
				t.Fatalf("toLookup: decodeTestLookup() failed: %v", err)
			}
			toLookup.Require.Docker.GetAuth().SetQueryToken(
				"", time.Time{},
			)
			fromLookup, err := decodeTestLookup(
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
			if err != nil {
				t.Fatalf("fromLookup: decodeTestLookup() failed: %v", err)
			}
			fromLookup.Require.Docker.GetAuth().SetQueryToken(
				util.SecretValue, time.Now().Add(time.Hour),
			)
			overridesRaw, err := polymorphic.Extract("yaml", []byte(tc.overrides), "require")
			if err != nil {
				t.Fatalf("polymorphic.Extract() failed: %v", err)
			}
			if decode.IsNull(overridesRaw) {
				fromLookup.Require = nil
			} else {
				err = UnmarshalRequire(
					"yaml", []byte(tc.overrides),
					fromLookup,
					fromLookup.Status,
					&lvCfg.Soft.Require,
				)
				if err != nil {
					t.Fatalf("UnmarshalRequire() failed: %v", err)
				}
			}

			toLookup.InheritSecrets(fromLookup, nil)

			if tc.inheritDockerToken {
				if toLookup.Require == nil || toLookup.Require.Docker == nil ||
					fromLookup.Require == nil || fromLookup.Require.Docker == nil {
					t.Fatalf(
						"Lookup.Require.Docker mismatch\n"+
							"got:  to nil=%t, from nil=%t\nwant: to nil=false, from nil=false",
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
				)
			} else if toLookup.Require != nil && toLookup.Require.Docker != nil {
				gotQueryToken, gotValidUntil := toLookup.Require.Docker.GetAuth().GetQueryTokenSelf()
				wantQueryToken, wantValidUntil := "", time.Time{}
				checkDockerToken(
					t,
					wantQueryToken, gotQueryToken,
					wantValidUntil, gotValidUntil,
				)
			}
		})
	}
}

func TestLookup_Query(t *testing.T) {
	l := &testLookup{
		Type: "test",
		URL:  "https://example.com",
	}

	errRegex := `^not implemented$`

	newVersion, err := l.Query(true, logx.LogFrom{})

	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
	if newVersion {
		t.Error("newVersion mismatch\ngot:  true\nwant: false")
	}
}
