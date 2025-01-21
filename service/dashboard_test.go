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

package service

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

func TestNewDashboardOptions(t *testing.T) {
	// GIVEN a set of input values
	tests := map[string]struct {
		autoApprove  *bool
		icon         string
		iconLinkTo   string
		webURL       string
		tags         []string
		defaults     *DashboardOptionsDefaults
		hardDefaults *DashboardOptionsDefaults
		want         *DashboardOptions
	}{
		"all fields set": {
			autoApprove:  test.BoolPtr(true),
			icon:         "icon-url",
			iconLinkTo:   "icon-link",
			webURL:       "web-url",
			tags:         []string{"tag1", "tag2"},
			defaults:     &DashboardOptionsDefaults{DashboardOptionsBase: DashboardOptionsBase{AutoApprove: test.BoolPtr(false)}},
			hardDefaults: &DashboardOptionsDefaults{DashboardOptionsBase: DashboardOptionsBase{AutoApprove: test.BoolPtr(false)}},
			want: &DashboardOptions{
				DashboardOptionsBase: DashboardOptionsBase{AutoApprove: test.BoolPtr(true)},
				Icon:                 "icon-url",
				IconLinkTo:           "icon-link",
				WebURL:               "web-url",
				Tags:                 []string{"tag1", "tag2"},
				Defaults:             &DashboardOptionsDefaults{DashboardOptionsBase: DashboardOptionsBase{AutoApprove: test.BoolPtr(false)}},
				HardDefaults:         &DashboardOptionsDefaults{DashboardOptionsBase: DashboardOptionsBase{AutoApprove: test.BoolPtr(false)}},
			},
		},
		"defaults": {
			autoApprove:  nil,
			icon:         "",
			iconLinkTo:   "",
			webURL:       "",
			tags:         nil,
			defaults:     nil,
			hardDefaults: nil,
			want: &DashboardOptions{
				DashboardOptionsBase: DashboardOptionsBase{AutoApprove: nil},
				Icon:                 "",
				IconLinkTo:           "",
				WebURL:               "",
				Tags:                 nil,
				Defaults:             nil,
				HardDefaults:         nil,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN NewDashboardOptions is called with them.
			got := NewDashboardOptions(tc.autoApprove, tc.icon, tc.iconLinkTo, tc.webURL, tc.tags, tc.defaults, tc.hardDefaults)

			// THEN the result is as expected.
			gotStr := util.ToJSONString(got)
			wantStr := util.ToJSONString(tc.want)
			if gotStr != wantStr {
				t.Errorf("NewDashboardOptions() result mismatch\n%q\ngot:\n%v",
					wantStr, gotStr)
			}
		})
	}
}

func TestDashboardOptions_UnmarshalJSON(t *testing.T) {
	// GIVEN a JSON string that represents a DashboardOptions.
	tests := map[string]struct {
		jsonData string
		errRegex string
		want     *DashboardOptions
	}{
		"invalid json": {
			jsonData: `{invalid: json}`,
			errRegex: test.TrimYAML(`
				failed to unmarshal DashboardOptions:
				invalid character.*$`),
			want: &DashboardOptions{},
		},
		"tags - []string": {
			jsonData: `{
				"tags": [
					"foo",
					"bar"
				]
			}`,
			errRegex: `^$`,
			want: &DashboardOptions{
				Tags: []string{"foo", "bar"},
			},
		},
		"tags - string": {
			jsonData: `{
				"tags": "foo"
			}`,
			errRegex: `^$`,
			want: &DashboardOptions{
				Tags: []string{"foo"},
			},
		},
		"tags - invalid": {
			jsonData: `{
				"tags": {
					"foo": "bar"
				}
			}`,
			errRegex: test.TrimYAML(`
				^error in tags field:
				type: <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Default to an empty DashboardOptions.
			dashboardOptions := &DashboardOptions{}

			// WHEN the JSON is unmarshalled into a DashboardOptions.
			err := dashboardOptions.UnmarshalJSON([]byte(test.TrimJSON(tc.jsonData)))

			// THEN the error is as expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("DashboardOptions.UnmarshalJSON() error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}
			// AND the result is as expected.
			gotString := util.ToJSONString(dashboardOptions)
			wantString := util.ToJSONString(tc.want)
			if tc.want != nil && gotString != wantString {
				t.Errorf("DashboardOptions.UnmarshalJSON() result mismatch\n%q\ngot:\n%q",
					wantString, gotString)
			}
		})
	}
}

func TestDashboardOptions_UnmarshalYAML(t *testing.T) {
	tests := map[string]struct {
		yamlData string
		errRegex string
		want     *DashboardOptions
	}{
		"invalid yaml": {
			yamlData: `invalid yaml`,
			errRegex: test.TrimYAML(`
			failed to unmarshal DashboardOptions:
			yaml: unmarshal errors:
			  .*cannot unmarshal.*$`),
			want: &DashboardOptions{},
		},
		"tags - []string": {
			yamlData: `
				tags:
				- foo
				- bar
			`,
			errRegex: `^$`,
			want: &DashboardOptions{
				Tags: []string{"foo", "bar"},
			},
		},
		"tags - string": {
			yamlData: `
				tags: foo
			`,
			errRegex: `^$`,
			want: &DashboardOptions{
				Tags: []string{"foo"},
			},
		},
		"tags - invalid": {
			yamlData: `
				tags:
					foo: bar
			`,
			errRegex: test.TrimYAML(`
				^error in tags field:
				type: <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Default to an empty DashboardOptions.
			dashboardOptions := &DashboardOptions{}

			// WHEN the YAML is unmarshalled into a DashboardOptions
			err := yaml.Unmarshal([]byte(test.TrimYAML(tc.yamlData)), &dashboardOptions)

			// THEN the error is as expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("DashboardOptions.UnmarshalYAML() error mismatch\nwant: %q\ngot:  %q",
					tc.errRegex, e)
			}
			// AND the result is as expected
			gotStr := util.ToJSONString(dashboardOptions)
			wantStr := util.ToJSONString(tc.want)
			if tc.want != nil && gotStr != wantStr {
				t.Errorf("DashboardOptions.UnmarshalYAML() result mismatch\nwant: %s\ngot:  %s",
					wantStr, gotStr)
			}
		})
	}
}

func TestDashboardOptions_GetAutoApprove(t *testing.T) {
	// GIVEN a DashboardOptions
	tests := map[string]struct {
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      bool
	}{
		"root overrides all": {
			want:             true,
			rootValue:        test.BoolPtr(true),
			defaultValue:     test.BoolPtr(false),
			hardDefaultValue: test.BoolPtr(false)},
		"default overrides hardDefault": {
			want:             true,
			defaultValue:     test.BoolPtr(true),
			hardDefaultValue: test.BoolPtr(false)},
		"hardDefault is last resort": {
			want:             true,
			hardDefaultValue: test.BoolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			dashboard := DashboardOptions{}
			dashboard.AutoApprove = tc.rootValue
			defaults := NewDashboardOptionsDefaults(tc.defaultValue)
			dashboard.Defaults = &defaults
			hardDefaultValues := NewDashboardOptionsDefaults(tc.hardDefaultValue)
			dashboard.HardDefaults = &hardDefaultValues

			// WHEN GetAutoApprove is called
			got := dashboard.GetAutoApprove()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %t\ngot:  %t",
					tc.want, got)
			}
		})
	}
}

func TestDashboardOptions_CheckValues(t *testing.T) {
	// GIVEN DashboardOptions
	tests := map[string]struct {
		dashboardOptions *DashboardOptions
		errRegex         string
	}{
		"nil": {
			errRegex:         `^$`,
			dashboardOptions: nil},
		"invalid web_url template": {
			errRegex:         `^web_url: ".*" <invalid>.*$`,
			dashboardOptions: &DashboardOptions{WebURL: "https://release-argus.io/{{ version }"}},
		"valid web_url template": {
			errRegex:         `^$`,
			dashboardOptions: &DashboardOptions{WebURL: "https://release-argus.io"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN CheckValues is called on it
			err := tc.dashboardOptions.CheckValues("")

			// THEN the err is what we expect
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("DashboardOptions.CheckValues() want %d lines of error:\n%q\ngot %d lines:\n%v\nstdout: %q",
					wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("DashboardOptions.CheckValues() error mismatch\nwant match for:\n%q\ngot:\n%q",
					tc.errRegex, e)
				return
			}
		})
	}
}
