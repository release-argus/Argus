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

package shoutrrr

import (
	"fmt"
	"regexp"
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
			ShoutrrrBase: ShoutrrrBase{
				Type: typeWithDefaults}},
		"no_main_with_type_and_no_defaults": &Shoutrrr{
			ShoutrrrBase: ShoutrrrBase{
				Type: typeWithNoDefaults}},
		"main_no_type": &Shoutrrr{},
		"main_with_type_and_defaults": &Shoutrrr{
			ShoutrrrBase: ShoutrrrBase{
				Type: typeWithDefaults}},
		"main_with_type_and_no_defaults": &Shoutrrr{
			ShoutrrrBase: ShoutrrrBase{
				Type: typeWithNoDefaults}}}
	mains := SliceDefaults{
		"main_no_type": &ShoutrrrDefaults{
			ShoutrrrBase: ShoutrrrBase{
				URLFields: typeWithNoDefaultsURLFields}},
		"main_with_type_and_defaults": &ShoutrrrDefaults{
			ShoutrrrBase: ShoutrrrBase{
				Type:      typeWithDefaults,
				URLFields: typeWithDefaultsURLFields}},
		"main_with_type_and_no_defaults": &ShoutrrrDefaults{
			ShoutrrrBase: ShoutrrrBase{
				Type:      typeWithNoDefaults,
				URLFields: typeWithNoDefaultsURLFields}},
		"main_not_on_service_with_defaults": &ShoutrrrDefaults{
			ShoutrrrBase: ShoutrrrBase{
				Type:      typeWithDefaults,
				URLFields: typeWithDefaultsURLFields}},
	}
	defaults := SliceDefaults{
		typeWithDefaults: &ShoutrrrDefaults{}}
	hardDefaults := SliceDefaults{
		typeWithDefaults:   &ShoutrrrDefaults{},
		typeWithNoDefaults: &ShoutrrrDefaults{},
		typeOther:          &ShoutrrrDefaults{}}
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		payload           TestPayload
		noServiceNotifies bool
		want              *Shoutrrr
		wantMain          string
		wantDefaults      string
		wantHardDefaults  string
		wantServiceURL    string
		err               string
	}{
		"empty": {
			payload: TestPayload{},
			err:     "name and/or name_previous are required",
		},
		"no name": {
			payload: TestPayload{
				ServiceName: "test"},
			err: "name and/or name_previous are required",
		},
		"no name, have name_previous": {
			payload: TestPayload{
				ServiceName:  "test",
				NamePrevious: "test"},
			err: "invalid type",
		},
		"no Type": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "no_main_no_defaults_no_harddefaults",
				URLFields:   typeWithNoDefaultsURLFields},
			err: "invalid type",
		},
		"edit, no Main, no Defaults - No Type": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "no_main_no_type",
				URLFields:   typeWithNoDefaultsURLFields},
			err: "invalid type",
		},
		"edit, no Main, no Defaults - with Type": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "no_main_no_type",
				Type:        typeWithNoDefaults,
				URLFields:   typeWithNoDefaultsURLFields},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type:      typeWithNoDefaults,
					URLFields: typeWithNoDefaultsURLFields}},
		},
		"edit, no Main, no Defaults - had Type (missing name_previous)": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "no_main_with_type_and_no_defaults",
				URLFields:   typeWithNoDefaultsURLFields},
			err: "invalid type",
		},
		"edit, no Main, no Defaults - had Type (have name_previous)": {
			payload: TestPayload{
				ServiceName:  "test",
				NamePrevious: "no_main_with_type_and_no_defaults",
				URLFields:    typeWithNoDefaultsURLFields},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type:      typeWithNoDefaults,
					URLFields: typeWithNoDefaultsURLFields}},
		},
		"edit, no Main, have Defaults": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "no_main_with_type_and_defaults",
				Type:        typeWithDefaults,
				URLFields:   typeWithDefaultsURLFields},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type:      typeWithDefaults,
					URLFields: typeWithDefaultsURLFields}},
		},
		"edit, have Main, no Defaults - Give Type": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "main_no_type",
				Type:        typeWithNoDefaults},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: typeWithNoDefaults}},
		},
		"edit, have Main, no Defaults - No Type": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "main_no_type"},
			err: "invalid type",
		},
		"edit, have Main, have Defaults": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "main_with_type_and_defaults",
				Type:        typeWithDefaults},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: typeWithDefaults}},
		},
		"edit, have Main, have Defaults - Fail, Different Type to Main": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "main_with_type_and_defaults",
				Type:        typeWithNoDefaults},
			err: `type: "[^"]+" != "[^"]+"`,
		},
		"edit, have Main, have Defaults - Fail, Invalid field": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "main_with_type_and_defaults",
				Options: map[string]string{
					"delay": "number"}},
			err: `delay: "number" <invalid>`,
		},
		"new, no Main, have Defaults, type from name": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        typeWithDefaults,
				URLFields:   typeWithDefaultsURLFields},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type:      typeWithDefaults,
					URLFields: typeWithDefaultsURLFields}},
		},
		"new, no Main, no Defaults, type from name": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        typeWithNoDefaults,
				URLFields:   typeWithNoDefaultsURLFields},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type:      typeWithNoDefaults,
					URLFields: typeWithNoDefaultsURLFields}},
		},
		"new, have Main, have Defaults": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "main_not_on_service_with_defaults"},
			want: &Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: typeWithDefaults}},
		},
		"new, have Main, have Defaults - Fail, Different Type to Main": {
			payload: TestPayload{
				ServiceName: "test",
				Name:        "main_not_on_service_with_defaults",
				Type:        typeWithNoDefaults},
			err: `type: "[^"]+" != "[^"]+"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel()

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

			var testServiceNotifies *Slice
			if !tc.noServiceNotifies {
				testServiceNotifies = serviceNotifies
			}

			// WHEN using the template
			got, serviceURL, err := FromPayload(
				&tc.payload,
				testServiceNotifies,
				mains, defaults, hardDefaults)

			// THEN the Shoutrrr is created as expected
			if tc.err == "" {
				tc.err = "^$"
			}
			gotErr := util.ErrorToString(err)
			if !regexp.MustCompile(tc.err).MatchString(gotErr) {
				t.Errorf("Expecting error to match %q, got %q",
					tc.err, gotErr)
			}
			if tc.err != "^$" {
				return
			}
			// AND the Shoutrrr is as expected
			if got.String("") != tc.want.String("") {
				t.Errorf("Result mismatch:\ngot:  %q\nwant: %q",
					got.String(""), tc.want.String(""))
			}
			// AND the serviceURL is as expected
			if serviceURL != tc.wantServiceURL {
				t.Errorf("ServiceURL mismatch:\ngot:  %q\nwant: %q",
					serviceURL, tc.wantServiceURL)
			}
		})
	}
}

func TestSortDefaults(t *testing.T) {
	// GIVEN a set of values for a Notify
	tests := map[string]struct {
		name         string
		nType        string
		main         *ShoutrrrDefaults
		defaults     SliceDefaults
		hardDefaults SliceDefaults
		wantType     string
		wantMain     string
		wantDefaults string
		err          string
	}{
		"Invalid Type": {
			name:  "test",
			nType: "gotify",
			err:   `invalid type "gotify"`,
		},
		"Type from Name": {
			name: "gotify",
			hardDefaults: SliceDefaults{
				"gotify": &ShoutrrrDefaults{}},
			wantType:     "gotify",
			wantMain:     "hardDefaults",
			wantDefaults: "hardDefaults",
		},
		"Type from Main": {
			name: "test",
			hardDefaults: SliceDefaults{
				"gotify": &ShoutrrrDefaults{}},
			main: &ShoutrrrDefaults{
				ShoutrrrBase: ShoutrrrBase{
					Type: "gotify"}},
			wantType:     "gotify",
			wantMain:     "main",
			wantDefaults: "hardDefaults",
		},
		"No Main, No Defaults": {
			name:  "test",
			nType: "gotify",
			hardDefaults: SliceDefaults{
				"gotify": &ShoutrrrDefaults{}},
			wantType:     "gotify",
			wantMain:     "hardDefaults",
			wantDefaults: "hardDefaults",
		},
		"Main, No Defaults": {
			name:  "test",
			nType: "gotify",
			main:  &ShoutrrrDefaults{},
			hardDefaults: SliceDefaults{
				"gotify": &ShoutrrrDefaults{}},
			wantType:     "gotify",
			wantMain:     "main",
			wantDefaults: "hardDefaults",
		},
		"No Main, Defaults": {
			name:  "test",
			nType: "gotify",
			defaults: SliceDefaults{
				"gotify": &ShoutrrrDefaults{}},
			hardDefaults: SliceDefaults{
				"gotify": &ShoutrrrDefaults{}},
			wantType:     "gotify",
			wantMain:     "defaults",
			wantDefaults: "defaults",
		},
		"Main, Defaults": {
			name:  "test",
			nType: "gotify",
			main:  &ShoutrrrDefaults{},
			defaults: SliceDefaults{
				"gotify": &ShoutrrrDefaults{}},
			hardDefaults: SliceDefaults{
				"gotify": &ShoutrrrDefaults{}},
			wantType:     "gotify",
			wantMain:     "main",
			wantDefaults: "defaults",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel()

			// WHEN sorting the defaults
			gotType, gotMain, gotDefaults, gotHardDefaults, err := sortDefaults(
				tc.name, tc.nType, tc.main, tc.defaults, tc.hardDefaults)

			// THEN the err is as expected
			if tc.err == "" {
				tc.err = "^$"
			}

			gotErr := util.ErrorToString(err)
			if !util.RegexCheck(tc.err, gotErr) {
				t.Errorf("Expecting error to match %q, got %q",
					tc.err, gotErr)
			}
			if tc.err != "^$" {
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
