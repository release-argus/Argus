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

package shoutrrr

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestShoutrrrs_Init(t *testing.T) {
	// GIVEN: Shoutrrrs.
	tests := []struct {
		name                   string
		nilMap                 bool
		shoutrrrs              *Shoutrrrs
		had, want              map[string]string
		mains                  *ShoutrrrsDefaults
		defaults, hardDefaults ShoutrrrsDefaults
	}{
		{
			name:      "nil slice",
			shoutrrrs: nil,
			nilMap:    true,
		},
		{
			name:      "empty slice",
			shoutrrrs: &Shoutrrrs{},
		},
		{
			name: "nil mains",
			shoutrrrs: &Shoutrrrs{
				"fail": testShoutrrr(true, false),
				"pass": testShoutrrr(false, false),
			},
		},
		{
			name: "slice with nil element and matching main",
			shoutrrrs: &Shoutrrrs{
				"fail": nil,
			},
			mains: &ShoutrrrsDefaults{
				"fail": testDefaults(false, false),
			},
		},
		{
			name: "have matching mains",
			shoutrrrs: &Shoutrrrs{
				"fail": testShoutrrr(true, false),
				"pass": testShoutrrr(false, false),
			},
			mains: &ShoutrrrsDefaults{
				"fail": testDefaults(false, false),
				"pass": testDefaults(true, false),
			},
		},
		{
			name: "some matching mains",
			shoutrrrs: &Shoutrrrs{
				"fail": testShoutrrr(true, false),
				"pass": testShoutrrr(false, false),
			},
			mains: &ShoutrrrsDefaults{
				"other": testDefaults(false, false),
				"pass":  testDefaults(true, false),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			serviceStatus := status.Status{}
			mainCount := 0
			if tc.mains == nil {
				tc.mains = &ShoutrrrsDefaults{}
			} else {
				mainCount = len(*tc.mains)
			}
			serviceStatus.Init(
				0, mainCount, 0,
				status.ServiceInfo{
					ID: tc.name,
				},
				&dashboard.Options{},
			)
			for i := range tc.defaults {
				tc.defaults[i].URLFields = tc.had
			}
			if tc.defaults == nil {
				tc.defaults = make(ShoutrrrsDefaults)
			}
			for i := range tc.hardDefaults {
				tc.hardDefaults[i].Params = tc.had
			}
			if tc.hardDefaults == nil {
				tc.hardDefaults = make(ShoutrrrsDefaults)
			}
			if tc.nilMap {
				tc.shoutrrrs = nil
			}

			// WHEN: Init is called on it.
			tc.shoutrrrs.Init(
				&serviceStatus,
				Config{
					Root:         *tc.mains,
					Defaults:     tc.defaults,
					HardDefaults: tc.hardDefaults,
				},
			)

			// THEN: the Shoutrrr is initialised correctly:
			if tc.nilMap {
				if tc.shoutrrrs != nil {
					t.Fatalf(
						"%s\nShoutrrrs.Init() nil shoutrrr should still be nil\ngot:  %v\nwant: nil",
						packageName, tc.shoutrrrs,
					)
				}
				return
			}

			for id, s := range *tc.shoutrrrs {
				prefix := fmt.Sprintf(
					"%s\nShoutrrrs.Init() id=%s",
					packageName, id,
				)

				fieldTests := []test.FieldAssertion{
					{
						Name: "Main",
						Check: func() (any, any, bool) {
							got := s.Main
							want := (*tc.mains)[id]

							if want != nil && got != want {
								return got, want, false
							}
							return got, want, true
						},
						Mode: test.CompareSamePointer,
					},
					{
						Name: "Defaults",
						Check: func() (any, any, bool) {
							got := s.Defaults
							want := tc.defaults[id]

							if want != nil && got != want {
								return got, want, false
							}
							return got, want, true
						},
						Mode: test.CompareSamePointer,
					},
					{
						Name: "HardDefaults",
						Check: func() (any, any, bool) {
							got := s.HardDefaults
							want := tc.hardDefaults[id]

							if want != nil && got != want {
								return got, want, false
							}
							return got, want, true
						},
						Mode: test.CompareSamePointer,
					},
					{
						Name: "Status",
						Got:  s.ServiceStatus,
						Want: &serviceStatus,
						Mode: test.CompareSamePointer,
					},
					{
						Name: "Status.Fails",
						Got:  s.Failed,
						Want: &s.ServiceStatus.Fails.Shoutrrr,
						Mode: test.CompareSamePointer,
					},
				}
				if err := test.AssertFields(t, fieldTests, prefix, "Shoutrrr"); err != nil {
					t.Fatal(err)
				}

				// 	Options, URLFields, Params:
				for key, val := range tc.want {
					if gotOpt := s.Options[key]; gotOpt != val {
						t.Errorf(
							"%s Options mismatch on %q\ngot:  %q (%+v)\nwant: %q (%+v)",
							prefix, key,
							gotOpt, s.Options,
							val, tc.want,
						)
					}
					if gotURL := s.Defaults.URLFields[key]; gotURL != val {
						t.Errorf(
							"%s Defaults.URLFields mismatch on %q\ngot:  %q (%+v)\nwant: %q (%+v)",
							prefix, key,
							gotURL, s.Defaults.URLFields,
							val, tc.want,
						)
					}
					if gotParam := s.HardDefaults.Params[key]; gotParam != val {
						t.Errorf(
							"%s HardDefaults.Params mismatch on %q\ngot:  %q (%+v)\nwant: %q (%+v)",
							prefix, key,
							gotParam, s.HardDefaults.Params,
							val, tc.want,
						)
					}
				}
			}
		})
	}
}

func TestShoutrrr_Init(t *testing.T) {
	tShoutrrr := testShoutrrr(false, false)
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name                         string
		id                           string
		clearType                    bool
		had, want                    map[string]string
		giveMain                     bool
		main                         *Defaults
		serviceShoutrrr, nilShoutrrr bool
	}{
		{
			name:        "nil shoutrrr",
			nilShoutrrr: true,
		},
		{
			name:            "all lowercase keys",
			id:              "lowercase",
			serviceShoutrrr: true,
			had: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash",
			},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash",
			},
		},
		{
			name:            "mixed-case keys",
			id:              "mixed-case",
			serviceShoutrrr: true,
			had: map[string]string{
				"hello": "TEST123", "FOO": "bAr", "bIsh": "bash",
			},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash",
			},
		},
		{
			name:            "gives matching main",
			id:              "matching-main",
			serviceShoutrrr: true,
			main:            &Defaults{},
			giveMain:        true,
		},
		{
			name:            "creates new main if none match",
			id:              "no-matching-main",
			serviceShoutrrr: true,
			main:            nil,
		},
		{
			name:      ".Type cleared when it matches .ID",
			id:        tShoutrrr.Type,
			clearType: true,
		},
		{
			name: ".Type cleared when it matches Main.Type",
			id:   "something",
			main: &Defaults{
				Base: Base{
					Type: tShoutrrr.Type,
				},
			},
			giveMain:  true,
			clearType: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ID = tc.id
			serviceStatus := shoutrrr.ServiceStatus
			shoutrrr.ServiceStatus.ServiceInfo.ID = tc.name
			if tc.giveMain {
				tc.main.Options = tc.had
			}
			shoutrrr.Options = map[string]string{}
			shoutrrr.Params = map[string]string{}
			shoutrrr.URLFields = map[string]string{}
			defaults := NewDefaults(
				"",
				make(map[string]string),
				make(map[string]string),
				make(map[string]string),
			)
			hardDefaults := NewDefaults(
				"",
				make(map[string]string),
				make(map[string]string),
				make(map[string]string),
			)
			for key := range tc.had {
				// Options.
				shoutrrr.Options[key] = tc.had[key]
				defaults.Options[key] = tc.had[key]
				hardDefaults.Options[key] = tc.had[key]
				// Params.
				shoutrrr.Params[key] = tc.had[key]
				defaults.Params[key] = tc.had[key]
				hardDefaults.Params[key] = tc.had[key]
				// URLFields.
				shoutrrr.URLFields[key] = tc.had[key]
				defaults.URLFields[key] = tc.had[key]
				hardDefaults.URLFields[key] = tc.had[key]
			}
			if tc.nilShoutrrr {
				shoutrrr = nil
			}

			// WHEN: Init is called on it.
			shoutrrr.Init(
				serviceStatus,
				tc.main,
				defaults, hardDefaults,
			)

			prefix := fmt.Sprintf("%s\nShoutrrr.Init()", packageName)

			// THEN: Shoutrrr remains nil if nil previously.
			if tc.nilShoutrrr {
				if shoutrrr != nil {
					t.Fatalf(
						"%s nil shoutrrr should remain nil\ngot:  %v\nwant: nil",
						prefix, shoutrrr,
					)
				}
				return
			}

			// AND: Pointers are set to the given values.
			fieldTests := []test.FieldAssertion{
				{Name: "Status", Got: shoutrrr.ServiceStatus, Want: serviceStatus, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: shoutrrr.HardDefaults, Want: hardDefaults, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: shoutrrr.Defaults, Want: defaults, Mode: test.CompareSamePointer},
			}
			mainFieldAssertionMode := test.CompareSamePointer
			if !tc.giveMain {
				mainFieldAssertionMode = test.CompareDifferentPointer
			}
			fieldTests = append(fieldTests, test.FieldAssertion{Name: "Main", Got: shoutrrr.Main, Want: tc.main, Mode: mainFieldAssertionMode})
			if err := test.AssertFields(t, fieldTests, prefix, "Shoutrrr"); err != nil {
				t.Fatal(err)
			}

			// AND: keys are lower-cased as expected.
			for key := range tc.want {
				fieldTests := []test.FieldAssertion{
					{Name: fmt.Sprintf("Options[%q]", key), Got: shoutrrr.Options[key], Want: tc.want[key], Mode: test.CompareEqual},
					{Name: fmt.Sprintf("URLFields[%q]", key), Got: shoutrrr.URLFields[key], Want: tc.want[key], Mode: test.CompareEqual},
					{Name: fmt.Sprintf("Params[%q]", key), Got: shoutrrr.Params[key], Want: tc.want[key], Mode: test.CompareEqual},
				}
				if err := test.AssertFields(t, fieldTests, prefix, "Shoutrrr"); err != nil {
					t.Fatal(err)
				}
			}

			// AND: the ID is cleared when expected.
			want := shoutrrr.Type
			if tc.clearType {
				want = ""
			}
			if shoutrrr.Type != want {
				t.Errorf(
					"%s\nShoutrrr.Init() cleared .Type unexpectedly\ngot:  %q\nwant: %q",
					packageName, shoutrrr.Type, want,
				)
			}
		})
	}
}

func TestShoutrrr_InitMaps(t *testing.T) {
	// GIVEN: a Shoutrrr.
	tests := []struct {
		name        string
		had, want   map[string]string
		nilShoutrrr bool
	}{
		{
			name:        "nil shoutrrr",
			nilShoutrrr: true,
		},
		{
			name: "all lowercase keys",
			had: map[string]string{
				"hello": "TEST123",
				"foo":   "bAr",
				"bish":  "bash",
			},
			want: map[string]string{
				"hello": "TEST123",
				"foo":   "bAr",
				"bish":  "bash",
			},
		},
		{
			name: "mixed-case keys",
			had: map[string]string{
				"hello": "TEST123",
				"FOO":   "bAr",
				"bIsh":  "bash",
			},
			want: map[string]string{
				"hello": "TEST123",
				"foo":   "bAr",
				"bish":  "bash",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ServiceStatus.ServiceInfo.ID = tc.name
			shoutrrr.Options = tc.had
			shoutrrr.URLFields = tc.had
			shoutrrr.Params = tc.had
			if tc.nilShoutrrr {
				shoutrrr = nil
			}

			// WHEN: InitMaps is called on it.
			shoutrrr.InitMaps()

			// THEN: the keys in the options/urlFields/params maps will have been converted to lowercase.
			if tc.nilShoutrrr {
				if shoutrrr != nil {
					t.Fatalf(
						"%s\nShoutrrr.InitMaps() nil shoutrrr should remain nil\ngot: %v",
						packageName, shoutrrr,
					)
				}
				return
			}
			errStr := "%s\nShoutrrr.InitMaps() mismatch on %q\ngot:  %v\nwant: %v"
			checkKeys := func(key string, gotMap MapStringStringOmitNull) {
				if !util.AreSlicesEqual(util.SortedKeys(tc.want), util.SortedKeys(gotMap)) {
					t.Fatalf(
						errStr,
						packageName, key, tc.want, gotMap,
					)
				}
			}
			checkKeys("Options", shoutrrr.Options)
			checkKeys("URLFields", shoutrrr.URLFields)
			checkKeys("Params", shoutrrr.Params)
		})
	}
}
