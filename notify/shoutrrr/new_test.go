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
	dashtest "github.com/release-argus/Argus/service/dashboard/test"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestShoutrrr_FromPayload(t *testing.T) {
	dashCfg := dashtest.PlainDefaultsConfig(t)

	testToken := "Aod9Cb0zXCeOrnD"
	typeWithDefaults := "gotify"
	typeWithDefaultsURLFields := map[string]string{
		"host":  "localhost",
		"token": testToken,
	}
	typeWithNoDefaults := "ntfy"
	typeWithNoDefaultsURLFields := map[string]string{
		"topic": "foo",
	}
	typeOther := "slack"

	// GIVEN: a bunch of 'service' notifiers.
	serviceNotifiers := &Shoutrrrs{
		"no_main_no_type": &Shoutrrr{},
		"no_main_with_type_and_defaults": &Shoutrrr{
			Base: Base{
				Type: typeWithDefaults,
			},
		},
		"no_main_with_type_and_no_defaults": &Shoutrrr{
			Base: Base{
				Type: typeWithNoDefaults,
			},
		},
		"main_no_type": &Shoutrrr{},
		"main_with_type_and_defaults": &Shoutrrr{
			Base: Base{
				Type: typeWithDefaults,
			},
		},
		"main_with_type_and_no_defaults": &Shoutrrr{
			Base: Base{
				Type: typeWithNoDefaults,
			},
		},
	}
	mains := ShoutrrrsDefaults{
		"main_no_type": &Defaults{
			Base: Base{
				URLFields: typeWithNoDefaultsURLFields,
			},
		},
		"main_with_type_and_defaults": &Defaults{
			Base: Base{
				Type:      typeWithDefaults,
				URLFields: typeWithDefaultsURLFields,
			},
		},
		"main_with_type_and_no_defaults": &Defaults{
			Base: Base{
				Type:      typeWithNoDefaults,
				URLFields: typeWithNoDefaultsURLFields,
			},
		},
		"main_not_on_service_with_defaults": &Defaults{
			Base: Base{
				Type:      typeWithDefaults,
				URLFields: typeWithDefaultsURLFields,
			},
		},
	}
	defaults := ShoutrrrsDefaults{
		typeWithDefaults: &Defaults{},
	}
	hardDefaults := ShoutrrrsDefaults{
		typeWithDefaults:   &Defaults{},
		typeWithNoDefaults: &Defaults{},
		typeOther:          &Defaults{},
	}

	// AND: a payload for a Shoutrrr.
	tests := []struct {
		name                                     string
		payload                                  TestPayload
		want                                     *Shoutrrr
		wantMain, wantDefaults, wantHardDefaults string
		wantServiceURL                           string
		errRegex                                 string
	}{
		{
			name:     "empty",
			payload:  TestPayload{},
			errRegex: `^name and/or name_previous are required`,
		},
		{
			name: "no name/only",
			payload: TestPayload{
				ServiceIDPrevious: "test",
			},
			errRegex: `^name and/or name_previous are required`,
		},
		{
			name: "no name/have name_previous",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				NamePrevious:      "test",
			},
			errRegex: `^invalid type "[^"]+"$`,
		},
		{
			name: "no type",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "no_main_no_defaults_no_hard_defaults",
				URLFields:         typeWithNoDefaultsURLFields,
			},
			errRegex: `^invalid type "[^"]+"$`,
		},
		{
			name: "edit/no Main/no Defaults/no type",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "no_main_no_type",
				URLFields:         typeWithNoDefaultsURLFields,
			},
			errRegex: `^invalid type "[^"]+"$`,
		},
		{
			name: "edit/no Main/no Defaults/with type",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "no_main_no_type",
				Type:              typeWithNoDefaults,
				URLFields:         typeWithNoDefaultsURLFields,
			},
			want: &Shoutrrr{
				Base: Base{
					Type:      typeWithNoDefaults,
					URLFields: typeWithNoDefaultsURLFields,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "edit/no Main/no Defaults/had type/missing name_previous",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "no_main_with_type_and_no_defaults",
				URLFields:         typeWithNoDefaultsURLFields,
			},
			errRegex: `^invalid type "[^"]+"$`,
		},
		{
			name: "edit/no Main/no Defaults/had type/have name_previous",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				NamePrevious:      "no_main_with_type_and_no_defaults",
				URLFields:         typeWithNoDefaultsURLFields,
			},
			want: &Shoutrrr{
				Base: Base{
					Type:      typeWithNoDefaults,
					URLFields: typeWithNoDefaultsURLFields,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "edit/no Main/have Defaults",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "no_main_with_type_and_defaults",
				Type:              typeWithDefaults,
				URLFields:         typeWithDefaultsURLFields,
			},
			want: &Shoutrrr{
				Base: Base{
					Type:      typeWithDefaults,
					URLFields: typeWithDefaultsURLFields,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "edit/have Main/no Defaults/give type",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "main_no_type",
				Type:              typeWithNoDefaults,
			},
			want: &Shoutrrr{
				Base: Base{
					Type: typeWithNoDefaults,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "edit/have Main/no Defaults/no type",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "main_no_type",
			},
			errRegex: `^invalid type "[^"]+"$`,
		},
		{
			name: "edit/have Main/have Defaults/core",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "main_with_type_and_defaults",
				Type:              typeWithDefaults,
			},
			want: &Shoutrrr{
				Base: Base{
					Type: typeWithDefaults,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "edit/have Main/have Defaults/fail/different type to Main",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "main_with_type_and_defaults",
				Type:              typeWithNoDefaults,
			},
			errRegex: `type: "` + typeWithNoDefaults + `" <invalid> .*\(gotify\)\)`,
		},
		{
			name: "edit/have Main/have Defaults/fail/invalid field",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "main_with_type_and_defaults",
				Options: map[string]string{
					"delay": "number",
				},
			},
			errRegex: test.TrimYAML(`
				^options:
					delay: "number" <invalid>.*$`,
			),
		},
		{
			name: "new/no Main, have Defaults, type from name",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              typeWithDefaults,
				URLFields:         typeWithDefaultsURLFields,
			},
			want: &Shoutrrr{
				Base: Base{
					Type:      typeWithDefaults,
					URLFields: typeWithDefaultsURLFields,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "new/no Main, no Defaults, type from name",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              typeWithNoDefaults,
				URLFields:         typeWithNoDefaultsURLFields,
			},
			want: &Shoutrrr{
				Base: Base{
					Type:      typeWithNoDefaults,
					URLFields: typeWithNoDefaultsURLFields,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "new/have Main, have Defaults/core",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "main_not_on_service_with_defaults",
			},
			want: &Shoutrrr{
				Base: Base{
					Type: typeWithDefaults,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "new/have Main, have Defaults/new service",
			payload: TestPayload{
				ServiceIDPrevious: "",
				ServiceID:         "something",
				Name:              "main_not_on_service_with_defaults",
			},
			want: &Shoutrrr{
				Base: Base{
					Type: typeWithDefaults,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "new/have Main, have Defaults, fail, different type to Main",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "main_not_on_service_with_defaults",
				Type:              typeWithNoDefaults,
			},
			errRegex: `^type: "` + typeWithNoDefaults + `" <invalid> .*\(gotify\)\).*$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.want != nil {
				vars := []struct {
					Target string
					Var    any
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
				testServiceNotify = (*serviceNotifiers)[tc.payload.NamePrevious]
			}
			dash, _ := dashboard.Decode(
				"yaml", nil,
				dashCfg,
			)
			testServiceStatus := status.New(
				nil, nil, nil,
				"",
				"", "",
				"", "",
				"",
				dash,
			)
			testServiceStatus.Init(
				0, 1, 0,
				status.ServiceInfo{
					ID:         tc.payload.ServiceID,
					Name:       tc.payload.ServiceName,
					ServiceURL: "https://example.com/service/url",
				},
				testServiceStatus.Dashboard,
			)

			// WHEN: using the payload.
			result, errRegex := FromPayload(
				tc.payload,
				testServiceNotify, testServiceStatus,
				Config{
					Root:         mains,
					Defaults:     defaults,
					HardDefaults: hardDefaults,
				},
			)

			prefix := fmt.Sprintf(
				"%s\nFromPayload(%+v)",
				packageName, tc.payload,
			)

			// THEN: it errors when expected.
			e := errfmt.FormatError(errRegex)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
				return
			}
			if e != "" {
				return
			}

			// AND: the Shoutrrr is as expected.
			got := result.String("")
			want := tc.want.String("")
			if got != want {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: the serviceName is as expected.
			got = result.ServiceStatus.ServiceInfo.ID
			want = tc.payload.ServiceID
			if got != want {
				t.Errorf(
					"%s .ServiceStatus.ServiceInfo.ID mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}

			// AND: the serviceURL is as expected.
			got = tc.payload.ServiceURL
			want = tc.wantServiceURL
			if got != want {
				t.Errorf(
					"%s .ServiceURL mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}
		})
	}
}

func TestResolveDefaults(t *testing.T) {
	// GIVEN: a set of values for a Notify.
	tests := []struct {
		name                             string
		notifyName, nType                string
		main                             *Defaults
		defaults, hardDefaults           ShoutrrrsDefaults
		wantType, wantMain, wantDefaults string
		errRegex                         string
	}{
		{
			name:       "Invalid Type",
			notifyName: "test",
			nType:      "gotify",
			errRegex:   `invalid type "gotify"`,
		},
		{
			name:       "Type from Name",
			notifyName: "gotify",
			hardDefaults: ShoutrrrsDefaults{
				"gotify": &Defaults{},
			},
			wantType:     "gotify",
			wantMain:     "hardDefaults",
			wantDefaults: "hardDefaults",
			errRegex:     `^$`,
		},
		{
			name:       "Type from Main",
			notifyName: "test",
			hardDefaults: ShoutrrrsDefaults{
				"gotify": &Defaults{},
			},
			main: &Defaults{
				Base: Base{
					Type: "gotify",
				},
			},
			wantType:     "gotify",
			wantMain:     "main",
			wantDefaults: "hardDefaults",
			errRegex:     `^$`,
		},
		{
			name:       "No Main, No Defaults",
			notifyName: "test",
			nType:      "gotify",
			hardDefaults: ShoutrrrsDefaults{
				"gotify": &Defaults{},
			},
			wantType:     "gotify",
			wantMain:     "hardDefaults",
			wantDefaults: "hardDefaults",
			errRegex:     `^$`,
		},
		{
			name:       "Main, No Defaults",
			notifyName: "test",
			nType:      "gotify",
			main:       &Defaults{},
			hardDefaults: ShoutrrrsDefaults{
				"gotify": &Defaults{},
			},
			wantType:     "gotify",
			wantMain:     "main",
			wantDefaults: "hardDefaults",
			errRegex:     `^$`,
		},
		{
			name:       "No Main, Defaults",
			notifyName: "test",
			nType:      "gotify",
			defaults: ShoutrrrsDefaults{
				"gotify": &Defaults{},
			},
			hardDefaults: ShoutrrrsDefaults{
				"gotify": &Defaults{},
			},
			wantType:     "gotify",
			wantMain:     "defaults",
			wantDefaults: "defaults",
			errRegex:     `^$`,
		},
		{
			name:       "Main, Defaults",
			notifyName: "test",
			nType:      "gotify",
			main:       &Defaults{},
			defaults: ShoutrrrsDefaults{
				"gotify": &Defaults{},
			},
			hardDefaults: ShoutrrrsDefaults{
				"gotify": &Defaults{},
			},
			wantType:     "gotify",
			wantMain:     "main",
			wantDefaults: "defaults",
			errRegex:     `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: resolveDefaults is called.
			gotType, gotMain, gotDefaults, gotHardDefaults, err := resolveDefaults(
				tc.notifyName,
				tc.nType,
				tc.main,
				tc.defaults, tc.hardDefaults,
			)

			prefix := fmt.Sprintf("%s\nresolveDefaults()", packageName)

			// THEN: the decode is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch on \ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if tc.errRegex != "^$" {
				return
			}

			// AND: the values are as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Type", Got: gotType, Want: tc.wantType},
			}

			// 	Main ref.
			if tc.wantMain == "defaults" || tc.wantMain == "hardDefaults" {
				wantMain := tc.defaults[gotType]
				if tc.wantMain == "hardDefaults" {
					wantMain = tc.hardDefaults[gotType]
				}
				fieldTests = append(
					fieldTests,
					test.FieldAssertion{Name: "Main", Got: gotMain, Want: wantMain, Mode: test.CompareSamePointer},
				)
			} else {
				fieldTests = append(
					fieldTests,
					test.FieldAssertion{Name: "Main", Got: gotMain, Want: tc.defaults[gotType], Mode: test.CompareDifferentPointer},
					test.FieldAssertion{Name: "Main", Got: gotMain, Want: tc.hardDefaults[gotType], Mode: test.CompareDifferentPointer},
				)
			}
			// 	Defaults/HardDefaults refs.
			wantDefaults := tc.defaults[gotType]
			if tc.wantDefaults == "hardDefaults" {
				wantDefaults = tc.hardDefaults[gotType]
			}
			fieldTests = append(
				fieldTests,
				test.FieldAssertion{Name: "Defaults", Got: gotDefaults, Want: wantDefaults, Mode: test.CompareSamePointer},
				test.FieldAssertion{Name: "HardDefaults", Got: gotHardDefaults, Want: tc.hardDefaults[gotType], Mode: test.CompareSamePointer},
			)

			if testErr := test.AssertFields(t, fieldTests, prefix, ""); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
