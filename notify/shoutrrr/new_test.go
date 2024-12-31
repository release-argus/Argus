// Copyright [2024] [Argus]
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

	"github.com/release-argus/Argus/util"
)

func TestShoutrrr_FromPayload(t *testing.T) {
	testToken := "Aod9Cb0zXCeOrnD"
	typeWithDefaults := "gotify"
	typeWithDefaultsURLFields := map[string]string{
		"host":  "localhost",
		"token": testToken}
	typeWithNoDefaults := "ntfy"
	typeWithNoDefaultsURLFields := map[string]string{
		"topic": "foo"}
	typeOther := "slack"
	serviceNotifies := &Slice{
		"no_main_no_type": &Shoutrrr{},
		"no_main_with_type_and_defaults": &Shoutrrr{
			Base: Base{
				Type: typeWithDefaults}},
		"no_main_with_type_and_no_defaults": &Shoutrrr{
			Base: Base{
				Type: typeWithNoDefaults}},
		"main_no_type": &Shoutrrr{},
		"main_with_type_and_defaults": &Shoutrrr{
			Base: Base{
				Type: typeWithDefaults}},
		"main_with_type_and_no_defaults": &Shoutrrr{
			Base: Base{
				Type: typeWithNoDefaults}}}
	mains := SliceDefaults{
		"main_no_type": &Defaults{
			Base: Base{
				URLFields: typeWithNoDefaultsURLFields}},
		"main_with_type_and_defaults": &Defaults{
			Base: Base{
				Type:      typeWithDefaults,
				URLFields: typeWithDefaultsURLFields}},
		"main_with_type_and_no_defaults": &Defaults{
			Base: Base{
				Type:      typeWithNoDefaults,
				URLFields: typeWithNoDefaultsURLFields}},
		"main_not_on_service_with_defaults": &Defaults{
			Base: Base{
				Type:      typeWithDefaults,
				URLFields: typeWithDefaultsURLFields}},
	}
	defaults := SliceDefaults{
		typeWithDefaults: &Defaults{}}
	hardDefaults := SliceDefaults{
		typeWithDefaults:   &Defaults{},
		typeWithNoDefaults: &Defaults{},
		typeOther:          &Defaults{}}
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		payload          TestPayload
		want             *Shoutrrr
		wantMain         string
		wantDefaults     string
		wantHardDefaults string
		wantServiceURL   string
		errRegex         string
	}{
		"empty": {
			payload:  TestPayload{},
			errRegex: "name and/or name_previous are required",
		},
		"no name": {
			payload: TestPayload{
				ServiceNamePrevious: "test"},
			errRegex: "name and/or name_previous are required",
		},
		"no name, have name_previous": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				NamePrevious:        "test"},
			errRegex: "invalid type",
		},
		"no Type": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "no_main_no_defaults_no_hard_defaults",
				URLFields:           typeWithNoDefaultsURLFields},
			errRegex: "invalid type",
		},
		"edit, no Main, no Defaults - No Type": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "no_main_no_type",
				URLFields:           typeWithNoDefaultsURLFields},
			errRegex: "invalid type",
		},
		"edit, no Main, no Defaults - with Type": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "no_main_no_type",
				Type:                typeWithNoDefaults,
				URLFields:           typeWithNoDefaultsURLFields},
			want: &Shoutrrr{
				Base: Base{
					Type:      typeWithNoDefaults,
					URLFields: typeWithNoDefaultsURLFields}},
			errRegex: `^$`,
		},
		"edit, no Main, no Defaults - had Type (missing name_previous)": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "no_main_with_type_and_no_defaults",
				URLFields:           typeWithNoDefaultsURLFields},
			errRegex: "invalid type",
		},
		"edit, no Main, no Defaults - had Type (have name_previous)": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				NamePrevious:        "no_main_with_type_and_no_defaults",
				URLFields:           typeWithNoDefaultsURLFields},
			want: &Shoutrrr{
				Base: Base{
					Type:      typeWithNoDefaults,
					URLFields: typeWithNoDefaultsURLFields}},
			errRegex: `^$`,
		},
		"edit, no Main, have Defaults": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "no_main_with_type_and_defaults",
				Type:                typeWithDefaults,
				URLFields:           typeWithDefaultsURLFields},
			want: &Shoutrrr{
				Base: Base{
					Type:      typeWithDefaults,
					URLFields: typeWithDefaultsURLFields}},
			errRegex: `^$`,
		},
		"edit, have Main, no Defaults - Give Type": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "main_no_type",
				Type:                typeWithNoDefaults},
			want: &Shoutrrr{
				Base: Base{
					Type: typeWithNoDefaults}},
			errRegex: `^$`,
		},
		"edit, have Main, no Defaults - No Type": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "main_no_type"},
			errRegex: "invalid type",
		},
		"edit, have Main, have Defaults": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "main_with_type_and_defaults",
				Type:                typeWithDefaults},
			want: &Shoutrrr{
				Base: Base{
					Type: typeWithDefaults}},
			errRegex: `^$`,
		},
		"edit, have Main, have Defaults - Fail, Different Type to Main": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "main_with_type_and_defaults",
				Type:                typeWithNoDefaults},
			errRegex: `type: "[^"]+" != "[^"]+"`,
		},
		"edit, have Main, have Defaults - Fail, Invalid field": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "main_with_type_and_defaults",
				Options: map[string]string{
					"delay": "number"}},
			errRegex: `delay: "number" <invalid>`,
		},
		"new, no Main, have Defaults, type from name": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                typeWithDefaults,
				URLFields:           typeWithDefaultsURLFields},
			want: &Shoutrrr{
				Base: Base{
					Type:      typeWithDefaults,
					URLFields: typeWithDefaultsURLFields}},
			errRegex: `^$`,
		},
		"new, no Main, no Defaults, type from name": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                typeWithNoDefaults,
				URLFields:           typeWithNoDefaultsURLFields},
			want: &Shoutrrr{
				Base: Base{
					Type:      typeWithNoDefaults,
					URLFields: typeWithNoDefaultsURLFields}},
			errRegex: `^$`,
		},
		"new, have Main, have Defaults": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "main_not_on_service_with_defaults"},
			want: &Shoutrrr{
				Base: Base{
					Type: typeWithDefaults}},
			errRegex: `^$`,
		},
		"new, have Main, have Defaults - new service": {
			payload: TestPayload{
				ServiceNamePrevious: "",
				ServiceName:         "something",
				Name:                "main_not_on_service_with_defaults"},
			want: &Shoutrrr{
				Base: Base{
					Type: typeWithDefaults}},
			errRegex: `^$`,
		},
		"new, have Main, have Defaults - Fail, Different Type to Main": {
			payload: TestPayload{
				ServiceNamePrevious: "test",
				Name:                "main_not_on_service_with_defaults",
				Type:                typeWithNoDefaults},
			errRegex: `type: "[^"]+" != "[^"]+"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.want != nil {
				vars := []struct {
					Target string
					Var    interface{}
				}{
					{Target: tc.wantMain, Var: tc.want.Main},
					{Target: tc.wantDefaults, Var: tc.want.Defaults},
					{Target: tc.wantHardDefaults, Var: tc.want.HardDefaults},
				}
				for _, v := range vars {
					switch v.Target {
					case "main":
						v.Var = tc.want.Main
					case "defaults":
						v.Var = tc.want.Defaults
					case "hardDefaults":
						v.Var = tc.want.HardDefaults
					}
				}
			}

			var testServiceNotify *Shoutrrr
			if tc.payload.NamePrevious != "" {
				testServiceNotify = (*serviceNotifies)[tc.payload.NamePrevious]
			}

			// WHEN using the template
			got, serviceURL, errRegex := FromPayload(
				tc.payload,
				testServiceNotify,
				mains,
				defaults, hardDefaults)

			// THEN the Shoutrrr is created as expected
			e := util.ErrorToString(errRegex)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Expecting error to match %q, got %q",
					tc.errRegex, e)
				return
			}
			if tc.errRegex != "^$" {
				return
			}
			// AND the Shoutrrr is as expected
			if got.String("") != tc.want.String("") {
				t.Errorf("Result mismatch:\ngot:  %q\nwant: %q",
					got.String(""), tc.want.String(""))
			}
			// AND the serviceName is as expected
			if *got.ServiceStatus.ServiceID != tc.payload.ServiceName {
				t.Errorf("ServiceName mismatch:\ngot:  %q\nwant: %q",
					*got.ServiceStatus.ServiceID, tc.payload.ServiceName)
			}
			// AND the serviceURL is as expected
			if serviceURL != tc.wantServiceURL {
				t.Errorf("ServiceURL mismatch:\ngot:  %q\nwant: %q",
					serviceURL, tc.wantServiceURL)
			}
		})
	}
}

func TestResolveDefaults(t *testing.T) {
	// GIVEN a set of values for a Notify
	tests := map[string]struct {
		name, nType                      string
		main                             *Defaults
		defaults, hardDefaults           SliceDefaults
		wantType, wantMain, wantDefaults string
		errRegex                         string
	}{
		"Invalid Type": {
			name:     "test",
			nType:    "gotify",
			errRegex: `invalid type "gotify"`,
		},
		"Type from Name": {
			name: "gotify",
			hardDefaults: SliceDefaults{
				"gotify": &Defaults{}},
			wantType:     "gotify",
			wantMain:     "hardDefaults",
			wantDefaults: "hardDefaults",
			errRegex:     `^$`,
		},
		"Type from Main": {
			name: "test",
			hardDefaults: SliceDefaults{
				"gotify": &Defaults{}},
			main: &Defaults{
				Base: Base{
					Type: "gotify"}},
			wantType:     "gotify",
			wantMain:     "main",
			wantDefaults: "hardDefaults",
			errRegex:     `^$`,
		},
		"No Main, No Defaults": {
			name:  "test",
			nType: "gotify",
			hardDefaults: SliceDefaults{
				"gotify": &Defaults{}},
			wantType:     "gotify",
			wantMain:     "hardDefaults",
			wantDefaults: "hardDefaults",
			errRegex:     `^$`,
		},
		"Main, No Defaults": {
			name:  "test",
			nType: "gotify",
			main:  &Defaults{},
			hardDefaults: SliceDefaults{
				"gotify": &Defaults{}},
			wantType:     "gotify",
			wantMain:     "main",
			wantDefaults: "hardDefaults",
			errRegex:     `^$`,
		},
		"No Main, Defaults": {
			name:  "test",
			nType: "gotify",
			defaults: SliceDefaults{
				"gotify": &Defaults{}},
			hardDefaults: SliceDefaults{
				"gotify": &Defaults{}},
			wantType:     "gotify",
			wantMain:     "defaults",
			wantDefaults: "defaults",
			errRegex:     `^$`,
		},
		"Main, Defaults": {
			name:  "test",
			nType: "gotify",
			main:  &Defaults{},
			defaults: SliceDefaults{
				"gotify": &Defaults{}},
			hardDefaults: SliceDefaults{
				"gotify": &Defaults{}},
			wantType:     "gotify",
			wantMain:     "main",
			wantDefaults: "defaults",
			errRegex:     `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN resolveDefaults is called
			gotType, gotMain, gotDefaults, gotHardDefaults, err := resolveDefaults(
				tc.name, tc.nType, tc.main, tc.defaults, tc.hardDefaults)

			// THEN the err is as expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("Expecting error to match %q, got %q",
					tc.errRegex, e)
			}
			if tc.errRegex != "^$" {
				return
			}
			// AND the values are as expected
			if gotType != tc.wantType {
				t.Fatalf("Type mismatch:\ngot:  %q\nwant: %q",
					gotType, tc.wantType)
			}
			allAddresses := fmt.Sprintf("main=%p, defaults=%p, hardDefaults=%p", gotMain, gotDefaults, gotHardDefaults)
			// Main ref
			if tc.wantMain == "defaults" || tc.wantMain == "hardDefaults" {
				if (tc.wantMain == "defaults" && gotMain != tc.defaults[gotType]) ||
					(tc.wantMain == "hardDefaults" && gotMain != tc.hardDefaults[gotType]) {
					t.Errorf("Main mismatch:\ngot:  %p\nwant: %q\n%s",
						gotMain, tc.wantMain, allAddresses)
				}
			} else if tc.wantMain != "defaults" && tc.wantMain != "hardDefaults" {
				if gotMain == gotDefaults ||
					gotMain == gotHardDefaults ||
					gotMain != tc.main {
					t.Errorf("Main mismatch:\ngot:  %p\nwant: %p\n%s",
						gotMain, tc.main, allAddresses)
				}
			}
			// Defaults ref
			if tc.wantDefaults == "hardDefaults" {
				if gotDefaults != tc.hardDefaults[gotType] {
					t.Errorf("Defaults mismatch:\ngot:  %p\nwant: %q\n%s",
						gotDefaults, tc.wantDefaults, allAddresses)
				}
			} else if gotDefaults != tc.defaults[gotType] {
				t.Errorf("Defaults mismatch:\ngot:  %p\nwant: %p\n%s",
					gotDefaults, tc.defaults[gotType], allAddresses)
			}
			// HardDefaults ref
			if gotHardDefaults != tc.hardDefaults[gotType] {
				t.Errorf("HardDefaults mismatch:\ngot:  %p\nwant: %p\n%s",
					gotHardDefaults, tc.hardDefaults[gotType], allAddresses)
			}
		})
	}
}
