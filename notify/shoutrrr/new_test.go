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
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	dashtest "github.com/release-argus/Argus/service/dashboard/test"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestShoutrrr_FromPayload(t *testing.T) {
	dashCfg := dashtest.PlainDefaultsConfig()

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
	typeWithMapSecrets := "generic"

	serviceNotifiers := &Shoutrrrs{
		"no_main_no_type": &Shoutrrr{},
		"no_main_with_type_and_defaults": &Shoutrrr{
			Type: typeWithDefaults,
		},
		"no_main_with_type_and_no_defaults": &Shoutrrr{
			Type: typeWithNoDefaults,
		},
		"main_no_type": &Shoutrrr{},
		"main_with_type_and_defaults": &Shoutrrr{
			Type: typeWithDefaults,
		},
		"main_with_type_and_no_defaults": &Shoutrrr{
			Type: typeWithNoDefaults,
		},
		"no_main_with_map_secrets": &Shoutrrr{
			Type: typeWithMapSecrets,
			URLFields: map[string]string{
				"host":       "example.com",
				"headers":    `{"Authorization":"Bearer x","X-Old":"old"}`,
				"query_vars": `{"token":"abc"}`,
			},
			Params: map[string]string{
				"actions":  `[{"label":"A","url":"https://a"},{"label":"B","url":"https://b"}]`,
				"unlisted": "VALUE",
			},
		},
	}
	mains := ShoutrrrsDefaults{
		"main_no_type": &Defaults{
			URLFields: typeWithNoDefaultsURLFields,
		},
		"main_with_type_and_defaults": &Defaults{
			Type:      typeWithDefaults,
			URLFields: typeWithDefaultsURLFields,
		},
		"main_with_type_and_no_defaults": &Defaults{
			Type:      typeWithNoDefaults,
			URLFields: typeWithNoDefaultsURLFields,
		},
		"main_not_on_service_with_defaults": &Defaults{
			Type:      typeWithDefaults,
			URLFields: typeWithDefaultsURLFields,
		},
	}
	defaults := ShoutrrrsDefaults{
		typeWithDefaults: &Defaults{},
	}
	hardDefaults := ShoutrrrsDefaults{
		typeWithDefaults:   &Defaults{},
		typeWithNoDefaults: &Defaults{},
		typeOther:          &Defaults{},
		typeWithMapSecrets: &Defaults{},
	}

	tests := []struct {
		name     string
		payload  TestPayload
		want     *Shoutrrr
		errRegex string
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
			// No notifier is named "test", so the name is taken as the type.
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
				Type:      typeWithNoDefaults,
				URLFields: typeWithNoDefaultsURLFields,
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
				Type:      typeWithNoDefaults,
				URLFields: typeWithNoDefaultsURLFields,
			},
			errRegex: `^$`,
		},
		{
			name: "edit/no Main/no Defaults/had type/masked values restored",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				NamePrevious:      "no_main_with_map_secrets",
				URLFields: map[string]string{
					"headers":    `{"Authorization":"<secret>","X-New":"new"}`,
					"query_vars": `{"token":"<secret>"}`,
				},
				Params: map[string]string{
					"actions":  `[{"label":"B","old_index":1,"url":"<secret>"}]`,
					"unlisted": util.SecretValue,
				},
			},
			want: &Shoutrrr{
				Type: typeWithMapSecrets,
				URLFields: map[string]string{
					"host":       "example.com",
					"headers":    `{"Authorization":"Bearer x","X-New":"new"}`,
					"query_vars": `{"token":"abc"}`,
				},
				Params: map[string]string{
					"actions":  `[{"label":"B","url":"https://b"}]`,
					"unlisted": "VALUE",
				},
			},
			errRegex: `^$`,
		},
		{
			name: "edit/no Main/no Defaults/had type/renamed key's masked value restored",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				NamePrevious:      "no_main_with_map_secrets",
				URLFields: map[string]string{
					"headers": `{"X-Auth":{"old_key":"Authorization","value":"<secret>"}}`,
				},
			},
			want: &Shoutrrr{
				Type: typeWithMapSecrets,
				URLFields: map[string]string{
					"host":       "example.com",
					"headers":    `{"X-Auth":"Bearer x"}`,
					"query_vars": `{"token":"abc"}`,
				},
				Params: map[string]string{
					"actions":  `[{"label":"A","url":"https://a"},{"label":"B","url":"https://b"}]`,
					"unlisted": "VALUE",
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
				Type:      typeWithDefaults,
				URLFields: typeWithDefaultsURLFields,
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
				Type: typeWithNoDefaults,
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
				Type: typeWithDefaults,
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
				Type:      typeWithDefaults,
				URLFields: typeWithDefaultsURLFields,
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
				Type:      typeWithNoDefaults,
				URLFields: typeWithNoDefaultsURLFields,
			},
			errRegex: `^$`,
		},
		{
			name: "new/JSON old indices dropped",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              typeWithNoDefaults,
				URLFields:         typeWithNoDefaultsURLFields,
				Params: map[string]string{
					"actions": `[{"action":"view","label":"A","old_index":null,"url":"https://a"}]`,
				},
			},
			want: &Shoutrrr{
				Type:      typeWithNoDefaults,
				URLFields: typeWithNoDefaultsURLFields,
				Params: map[string]string{
					"actions": `[{"action":"view","label":"A","url":"https://a"}]`,
				},
			},
			errRegex: `^$`,
		},
		{
			name: "new/fail/URL not located",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              typeWithDefaults,
				URLFields: map[string]string{
					"host":  "https://\texample.com",
					"token": testToken,
				},
			},
			errRegex: `failed to parse URL`,
		},
		{
			name: "new/have Main, have Defaults/core",
			payload: TestPayload{
				ServiceIDPrevious: "test",
				Name:              "main_not_on_service_with_defaults",
			},
			want: &Shoutrrr{
				Type: typeWithDefaults,
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
				Type: typeWithDefaults,
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

			var testServiceNotify *Shoutrrr
			if tc.payload.NamePrevious != "" {
				testServiceNotify = (*serviceNotifiers)[tc.payload.NamePrevious]
			}
			dash, _ := dashboard.Decode("yaml", nil, dashCfg)
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

			result, err := FromPayload(
				tc.payload,
				testServiceNotify, testServiceStatus,
				Config{
					Root:         mains,
					Defaults:     defaults,
					HardDefaults: hardDefaults,
				},
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
				return
			}
			if err != nil {
				return
			}

			got := result.String("")
			want := tc.want.String("")
			if got != want {
				t.Errorf("Shoutrrr.String() mismatch\ngot:  %q\nwant: %q", got, want)
			}

			if result.ServiceStatus != testServiceStatus {
				t.Errorf(
					"Shoutrrr.ServiceStatus mismatch\ngot:  %p\nwant: %p",
					result.ServiceStatus, testServiceStatus,
				)
			}
		})
	}
}

func TestResolveDefaults(t *testing.T) {
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
				Type: "gotify",
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

			gotType, gotMain, gotDefaults, gotHardDefaults, err := resolveDefaults(
				tc.notifyName,
				tc.nType,
				tc.main,
				tc.defaults, tc.hardDefaults,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
				return
			}
			if err != nil {
				return
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Type", Got: gotType, Want: tc.wantType, Mode: test.CompareEqual},
			}

			wantMain := tc.main
			switch tc.wantMain {
			case "defaults":
				wantMain = tc.defaults[tc.wantType]
			case "hardDefaults":
				wantMain = tc.hardDefaults[tc.wantType]
			}
			fieldTests = append(
				fieldTests,
				test.FieldAssertion{Name: "Main", Got: gotMain, Want: wantMain, Mode: test.CompareSamePointer},
			)

			wantDefaults := tc.defaults[tc.wantType]
			if tc.wantDefaults == "hardDefaults" {
				wantDefaults = tc.hardDefaults[tc.wantType]
			}
			fieldTests = append(
				fieldTests,
				test.FieldAssertion{Name: "Defaults", Got: gotDefaults, Want: wantDefaults, Mode: test.CompareSamePointer},
				test.FieldAssertion{Name: "HardDefaults", Got: gotHardDefaults, Want: tc.hardDefaults[tc.wantType], Mode: test.CompareSamePointer},
			)

			if testErr := test.AssertFields(t, fieldTests, ""); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
