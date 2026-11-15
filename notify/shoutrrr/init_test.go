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
	"maps"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestShoutrrr_Init(t *testing.T) {
	tShoutrrr := testShoutrrr(false, false)

	tests := []struct {
		name        string
		id          string
		clearType   bool
		had, want   map[string]string
		giveMain    bool
		main        *Defaults
		nilShoutrrr bool
	}{
		{
			name:        "nil shoutrrr",
			nilShoutrrr: true,
		},
		{
			name: "all lowercase keys",
			id:   "lowercase",
			had: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash",
			},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash",
			},
		},
		{
			name: "mixed-case keys",
			id:   "mixed-case",
			had: map[string]string{
				"hello": "TEST123", "FOO": "bAr", "bIsh": "bash",
			},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash",
			},
		},
		{
			name:     "gives matching main",
			id:       "matching-main",
			main:     &Defaults{},
			giveMain: true,
		},
		{
			name: "creates new main if none match",
			id:   "no-matching-main",
			main: nil,
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
				Type: tShoutrrr.Type,
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
				shoutrrr.Options[key] = tc.had[key]
				defaults.Options[key] = tc.had[key]
				hardDefaults.Options[key] = tc.had[key]
				shoutrrr.Params[key] = tc.had[key]
				defaults.Params[key] = tc.had[key]
				hardDefaults.Params[key] = tc.had[key]
				shoutrrr.URLFields[key] = tc.had[key]
				defaults.URLFields[key] = tc.had[key]
				hardDefaults.URLFields[key] = tc.had[key]
			}
			if tc.nilShoutrrr {
				shoutrrr = nil
			}
			var hadType string
			if shoutrrr != nil {
				hadType = shoutrrr.Type
			}

			shoutrrr.Init(
				serviceStatus,
				tc.main,
				defaults, hardDefaults,
			)

			// A nil Shoutrrr need only not panic.
			if tc.nilShoutrrr {
				return
			}

			fieldTests := []test.FieldAssertion{
				{Name: "ServiceStatus", Got: shoutrrr.ServiceStatus, Want: serviceStatus, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: shoutrrr.HardDefaults, Want: hardDefaults, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: shoutrrr.Defaults, Want: defaults, Mode: test.CompareSamePointer},
			}
			mainFieldAssertionMode := test.CompareSamePointer
			if !tc.giveMain {
				mainFieldAssertionMode = test.CompareDifferentPointer
			}
			fieldTests = append(fieldTests, test.FieldAssertion{Name: "Main", Got: shoutrrr.Main, Want: tc.main, Mode: mainFieldAssertionMode})
			if testErr := test.AssertFields(t, fieldTests, "Shoutrrr"); testErr != nil {
				t.Fatal(testErr)
			}

			if testErr := test.AssertMapEqual(t, shoutrrr.Options, tc.want, "Shoutrrr.Options"); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqual(t, shoutrrr.URLFields, tc.want, "Shoutrrr.URLFields"); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqual(t, shoutrrr.Params, tc.want, "Shoutrrr.Params"); testErr != nil {
				t.Error(testErr)
			}

			want := hadType
			if tc.clearType {
				want = ""
			}
			if shoutrrr.Type != want {
				t.Errorf("Shoutrrr.Type mismatch\ngot:  %q\nwant: %q", shoutrrr.Type, want)
			}
		})
	}
}

func TestShoutrrrs_Init(t *testing.T) {
	tests := []struct {
		name      string
		shoutrrrs *Shoutrrrs
		mains     *ShoutrrrsDefaults
	}{
		{
			name:      "nil map",
			shoutrrrs: nil,
		},
		{
			name:      "empty map",
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
			name: "map with nil element and matching main",
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

			mainCount := 0
			if tc.mains == nil {
				tc.mains = &ShoutrrrsDefaults{}
			} else {
				mainCount = len(*tc.mains)
			}
			serviceStatus := newTestStatus(tc.name, mainCount)
			defaults := ShoutrrrsDefaults{"gotify": &Defaults{}}
			hardDefaults := ShoutrrrsDefaults{"gotify": &Defaults{}}

			tc.shoutrrrs.Init(
				serviceStatus,
				Config{
					Root:         *tc.mains,
					Defaults:     defaults,
					HardDefaults: hardDefaults,
				},
			)

			// A nil Shoutrrrs need only not panic.
			if tc.shoutrrrs == nil {
				return
			}

			for id, s := range *tc.shoutrrrs {
				t.Run(id, func(t *testing.T) {
					t.Parallel()

					notifyType := s.GetType()
					mainMode := test.CompareNotNil
					wantMain := (*tc.mains)[id]
					if wantMain != nil {
						mainMode = test.CompareSamePointer
					}
					fieldTests := []test.FieldAssertion{
						{Name: "Main", Got: s.Main, Want: wantMain, Mode: mainMode},
						{Name: "Defaults", Got: s.Defaults, Want: defaults[notifyType], Mode: test.CompareSamePointer},
						{Name: "HardDefaults", Got: s.HardDefaults, Want: hardDefaults[notifyType], Mode: test.CompareSamePointer},
						{Name: "ServiceStatus", Got: s.ServiceStatus, Want: serviceStatus, Mode: test.CompareSamePointer},
						{Name: "Failed", Got: s.Failed, Want: &serviceStatus.Fails.Shoutrrr, Mode: test.CompareSamePointer},
					}
					if testErr := test.AssertFields(t, fieldTests, "Shoutrrr"); testErr != nil {
						t.Fatal(testErr)
					}
				})
			}
		})
	}
}

func TestShoutrrr_InitMaps(t *testing.T) {
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
			shoutrrr.Options = maps.Clone(tc.had)
			shoutrrr.URLFields = maps.Clone(tc.had)
			shoutrrr.Params = maps.Clone(tc.had)
			if tc.nilShoutrrr {
				shoutrrr = nil
			}

			shoutrrr.InitMaps()

			// A nil Shoutrrr need only not panic.
			if tc.nilShoutrrr {
				return
			}

			if testErr := test.AssertMapEqual(t, shoutrrr.Options, tc.want, "Shoutrrr.Options"); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqual(t, shoutrrr.URLFields, tc.want, "Shoutrrr.URLFields"); testErr != nil {
				t.Error(testErr)
			}
			if testErr := test.AssertMapEqual(t, shoutrrr.Params, tc.want, "Shoutrrr.Params"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
