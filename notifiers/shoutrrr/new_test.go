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
	"encoding/json"
	"regexp"
	"testing"

	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestShoutrrr_UseTemplate(t *testing.T) {
	testToken := "Aod9Cb0zXCeOrnD"
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		template  Shoutrrr
		options   string
		urlFields string
		params    string
		want      Shoutrrr
		err       string
	}{
		"empty": {
			template: Shoutrrr{},
			want:     Shoutrrr{},
			err:      "host: <required>.* token: <required>",
		},
		"valid": {
			template: Shoutrrr{},
			urlFields: `
				{
					"host": "example.com",
					"token": "` + testToken + `"
				}`,
			want: Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					URLFields: map[string]string{
						"host":  "example.com",
						"token": testToken}}},
		},
		"inherit secrets": {
			template: Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: "gotify",
					URLFields: map[string]string{
						"host":  "release-argus.com",
						"token": testToken,
					}}},
			urlFields: `
				{
					"host": "example.com",
					"token": "<secret>"
				}`,
			want: Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					URLFields: map[string]string{
						"host":  "example.com",
						"token": testToken}}},
		},
		"invalid CheckValues": {
			urlFields: `
				{"host": "release-argus.com"}`,
			template: Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					Type: "gotify",
					URLFields: map[string]string{
						"host": "example.com",
					}}},
			want: Shoutrrr{
				ShoutrrrBase: ShoutrrrBase{
					URLFields: map[string]string{
						"host": "release-argus.com",
					}}},
			err: "token: <required>",
		},
		"invalid JSON - options": {
			template: Shoutrrr{},
			options:  "invalid",
			err:      "options:.* invalid character",
		},
		"invalid JSON - urlFields": {
			template:  Shoutrrr{},
			urlFields: "invalid",
			err:       "url_fields:.* invalid character",
		},
		"invalid JSON - params": {
			template: Shoutrrr{},
			params:   "invalid",
			err:      "params:.* invalid character",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel()

			tc.template.Type = "gotify"
			tc.want.Type = tc.template.Type
			tc.template.Main = &ShoutrrrDefaults{}
			tc.template.Defaults = &ShoutrrrDefaults{}
			tc.template.HardDefaults = &ShoutrrrDefaults{}
			tc.template.ServiceStatus = &svcstatus.Status{}
			vars := []*string{&tc.options, &tc.urlFields, &tc.params}
			for _, v := range vars {
				*v = trimJSON(*v)
				if *v == "" {
					*v = "{}"
				}
			}

			// WHEN using the template
			got, err := tc.template.UseTemplate(
				&tc.options,
				&tc.urlFields,
				&tc.params,
				&util.LogFrom{
					Primary: "TestCopyAndModify", Secondary: name})

			// THEN the Shoutrrr is created as expected
			if tc.err == "" {
				tc.err = "^$"
			}
			gotErr := util.ErrorToString(err)
			if !regexp.MustCompile(tc.err).MatchString(gotErr) {
				t.Errorf("Expecting error to match %q, got %q",
					tc.err, gotErr)
			}
			if tc.err != "^$" && got.String("") != tc.want.String("") {
				t.Errorf("Result mismatch:\ngot:  %q\nwant: %q",
					got.String(""), tc.want.String(""))
			}
		})
	}

}

func TestShoutrrr_StringToMap(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		str     *string
		baseMap map[string]string
		want    map[string]string
		err     string
	}{
		"nil": {
			str: nil,
			baseMap: map[string]string{
				"test": "foo"},
			want: map[string]string{
				"test": "foo"},
		},
		"empty": {
			str: stringPtr(""),
			baseMap: map[string]string{
				"test": "foo"},
			want: map[string]string{
				"test": "foo"},
			err: "unexpected end of JSON input",
		},
		"empty JSON": {
			str: stringPtr("{}"),
			baseMap: map[string]string{
				"test": "foo"},
			want: map[string]string{
				"test": "foo"},
		},
		"invalid JSON": {
			str: stringPtr(`
				{
					"devices": "foo",
					"other": "<secret>",
				}`),
			baseMap: map[string]string{
				"test": "foo"},
			want: map[string]string{
				"test": "foo"},
			err: `invalid character.*\}`,
		},
		"<secret> ref": {
			str: stringPtr(`
				{
					"devices": "foo",
					"other": "<secret>"
				}`),
			baseMap: map[string]string{
				"devices": "<secret>",
				"other":   "bar",
			},
			want: map[string]string{
				"devices": "foo",
				"other":   "bar",
			},
		},
		"<secret> ref <secret> stays <secret>": {
			str: stringPtr(`
				{
					"devices": "foo",
					"other": "<secret>"
				}`),
			baseMap: map[string]string{
				"devices": "<secret>",
				"other":   "<secret>",
			},
			want: map[string]string{
				"devices": "foo",
				"other":   "<secret>",
			},
		},
		"<secret> ref empty is kept as <secret>": {
			str: stringPtr(`
				{
					"devices": "foo",
					"other": "<secret>"
				}`),
			baseMap: map[string]string{
				"devices": "<secret>",
			},
			want: map[string]string{
				"devices": "foo",
				"other":   "<secret>",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			targetMap := make(map[string]string)
			if tc.str != nil {
				*tc.str = trimJSON(*tc.str)
			}

			// WHEN inheriting secrets
			err := stringToMap(tc.str, &targetMap, &tc.baseMap)

			// THEN the secrets are inherited
			got, _ := json.Marshal(targetMap)
			want, _ := json.Marshal(tc.want)
			if string(got) != string(want) {
				t.Errorf("InheritSecrets() mismatch:\ngot:  %q\nwant: %q",
					got, want)
			}
			// AND any errs are expected
			errStr := util.ErrorToString(err)
			wantErr := "^$"
			if tc.err != "" {
				wantErr = tc.err
			}
			if !regexp.MustCompile(wantErr).MatchString(errStr) {
				t.Errorf("Expecting error to match %q, got %q",
					wantErr, errStr)
			}
		})
	}
}
