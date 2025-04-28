// Copyright [2025] [Argus]
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

package dashboard

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestNewOptions(t *testing.T) {
	// GIVEN a set of input values.
	tests := map[string]struct {
		autoApprove  *bool
		icon         string
		iconLinkTo   string
		webURL       string
		tags         []string
		defaults     *OptionsDefaults
		hardDefaults *OptionsDefaults
		want         *Options
	}{
		"all fields set": {
			autoApprove:  test.BoolPtr(true),
			icon:         "icon-url",
			iconLinkTo:   "icon-link",
			webURL:       "web-url",
			tags:         []string{"tag1", "tag2"},
			defaults:     &OptionsDefaults{OptionsBase: OptionsBase{AutoApprove: test.BoolPtr(false)}},
			hardDefaults: &OptionsDefaults{OptionsBase: OptionsBase{AutoApprove: test.BoolPtr(false)}},
			want: &Options{
				OptionsBase:  OptionsBase{AutoApprove: test.BoolPtr(true)},
				Icon:         "icon-url",
				IconLinkTo:   "icon-link",
				WebURL:       "web-url",
				Tags:         []string{"tag1", "tag2"},
				Defaults:     &OptionsDefaults{OptionsBase: OptionsBase{AutoApprove: test.BoolPtr(false)}},
				HardDefaults: &OptionsDefaults{OptionsBase: OptionsBase{AutoApprove: test.BoolPtr(false)}},
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
			want: &Options{
				OptionsBase:  OptionsBase{AutoApprove: nil},
				Icon:         "",
				IconLinkTo:   "",
				WebURL:       "",
				Tags:         nil,
				Defaults:     nil,
				HardDefaults: nil,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN NewOptions is called with them.
			got := NewOptions(
				tc.autoApprove,
				tc.icon, tc.iconLinkTo,
				tc.webURL,
				tc.tags,
				tc.defaults, tc.hardDefaults)

			// THEN the result is as expected.
			gotStr := util.ToJSONString(got)
			wantStr := util.ToJSONString(tc.want)
			if gotStr != wantStr {
				t.Errorf("%s\nwant: %q\ngot:  %v",
					packageName, wantStr, gotStr)
			}
		})
	}
}

func TestOptions_UnmarshalJSON(t *testing.T) {
	// GIVEN a JSON string that represents a Options.
	tests := map[string]struct {
		jsonData string
		errRegex string
		want     *Options
	}{
		"invalid JSON": {
			jsonData: `{invalid: json}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal service\.Dashboard:
					invalid character.*$`),
			want: &Options{},
		},
		"tags - []string": {
			jsonData: `{
				"tags": [
					"foo",
					"bar"
				]
			}`,
			errRegex: `^$`,
			want: &Options{
				Tags: []string{"foo", "bar"},
			},
		},
		"tags - string": {
			jsonData: `{
				"tags": "foo"
			}`,
			errRegex: `^$`,
			want: &Options{
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
				^failed to unmarshal service\.Dashboard:
					tags: <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Default to an empty Options.
			dashboardOptions := &Options{}

			// WHEN the JSON is unmarshalled into a Options.
			err := dashboardOptions.UnmarshalJSON([]byte(test.TrimJSON(tc.jsonData)))

			// THEN the error is as expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
			// AND the result is as expected.
			gotString := util.ToJSONString(dashboardOptions)
			wantString := util.ToJSONString(tc.want)
			if tc.want != nil && gotString != wantString {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, wantString, gotString)
			}
		})
	}
}

func TestOptions_UnmarshalYAML(t *testing.T) {
	tests := map[string]struct {
		yamlData string
		errRegex string
		want     *Options
	}{
		"invalid YAML": {
			yamlData: `invalid yaml`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal service\.Dashboard:
					line \d: cannot unmarshal.*$`),
			want: &Options{},
		},
		"tags - []string": {
			yamlData: `
				tags:
				- foo
				- bar
			`,
			errRegex: `^$`,
			want: &Options{
				Tags: []string{"foo", "bar"},
			},
		},
		"tags - string": {
			yamlData: `
				tags: foo
			`,
			errRegex: `^$`,
			want: &Options{
				Tags: []string{"foo"},
			},
		},
		"tags - invalid": {
			yamlData: `
				tags:
					foo: bar
			`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal service\.Dashboard:
					tags: <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Default to an empty Options.
			dashboardOptions := &Options{}

			// WHEN the YAML is unmarshalled into a Options.
			err := yaml.Unmarshal([]byte(test.TrimYAML(tc.yamlData)), &dashboardOptions)

			// THEN the error is as expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
			// AND the result is as expected.
			gotStr := util.ToJSONString(dashboardOptions)
			wantStr := util.ToJSONString(tc.want)
			if tc.want != nil && gotStr != wantStr {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, wantStr, gotStr)
			}
		})
	}
}

func TestOptions_Copy(t *testing.T) {
	// GIVEN an Options.
	tests := map[string]struct {
		options *Options
		want    *Options
	}{
		"nil options": {
			options: nil,
			want:    nil,
		},
		"all fields set": {
			options: &Options{
				OptionsBase: OptionsBase{
					AutoApprove: test.BoolPtr(true),
				},
				Icon:               "icon-url",
				IconLinkTo:         "icon-link",
				WebURL:             "web-url",
				iconExpanded:       test.StringPtr("expanded-icon-url"),
				iconNotify:         test.StringPtr("notify-icon-url"),
				iconLinkToExpanded: test.StringPtr("expanded-icon-link"),
				webURLExpanded:     test.StringPtr("expanded-web-url"),
			},
			want: &Options{
				OptionsBase: OptionsBase{
					AutoApprove: test.BoolPtr(true),
				},
				Icon:               "icon-url",
				IconLinkTo:         "icon-link",
				WebURL:             "web-url",
				iconExpanded:       test.StringPtr("expanded-icon-url"),
				iconNotify:         test.StringPtr("notify-icon-url"),
				iconLinkToExpanded: test.StringPtr("expanded-icon-link"),
				webURLExpanded:     test.StringPtr("expanded-web-url"),
			},
		},
		"some fields nil": {
			options: &Options{
				OptionsBase: OptionsBase{
					AutoApprove: nil,
				},
				Icon:         "icon-url",
				iconExpanded: test.StringPtr("hi"),
				IconLinkTo:   "",
				WebURL:       "web-url",
			},
			want: &Options{
				OptionsBase: OptionsBase{
					AutoApprove: nil,
				},
				Icon:         "icon-url",
				iconExpanded: test.StringPtr("hi"),
				IconLinkTo:   "",
				WebURL:       "web-url",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Copy is called.
			got := tc.options.Copy()

			// THEN the copied Options matches the expected result.
			gotStr := util.ToJSONString(got)
			wantStr := util.ToJSONString(tc.want)
			if gotStr != wantStr {
				t.Errorf("%s'nCopy() mismatch\nwant: %q\ngot:  %q",
					packageName, wantStr, gotStr)
			}
		})
	}
}

func TestOptions_GetAutoApprove(t *testing.T) {
	// GIVEN a Options.
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

			dashboard := Options{}
			dashboard.AutoApprove = tc.rootValue
			defaults := NewOptionsDefaults(tc.defaultValue)
			dashboard.Defaults = &defaults
			hardDefaultValues := NewOptionsDefaults(tc.hardDefaultValue)
			dashboard.HardDefaults = &hardDefaultValues

			// WHEN GetAutoApprove is called.
			got := dashboard.GetAutoApprove()

			// THEN the function returns the correct result.
			if got != tc.want {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.want, got)
			}
		})
	}
}

func TestOptions_SetFallbackIcon(t *testing.T) {
	// GIVEN an Options and a fallback icon URL.
	tests := map[string]struct {
		initialIconNotify *string
		newIconNotify     string
		want              *string
	}{
		"set new fallback icon": {
			initialIconNotify: nil,
			newIconNotify:     "new-icon-url",
			want:              test.StringPtr("new-icon-url"),
		},
		"overwrite existing fallback icon": {
			initialIconNotify: test.StringPtr("old-icon-url"),
			newIconNotify:     "new-icon-url",
			want:              test.StringPtr("new-icon-url"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := &Options{
				iconNotify: tc.initialIconNotify,
			}

			// WHEN SetFallbackIcon is called with a new icon URL.
			options.SetFallbackIcon(tc.newIconNotify)

			// THEN the fallback icon is updated as expected.
			if options.iconNotify == nil || *options.iconNotify != *tc.want {
				t.Errorf("%s\nvalue mismatch\nwant: %q\ngot:  %q",
					packageName, *tc.want, util.DereferenceOrDefault(options.iconNotify))
			}
		})
	}
}

func TestOptions_GetIconLinkTo(t *testing.T) {
	// GIVEN an Options with various iconLinkTo-related fields set.
	tests := map[string]struct {
		iconLinkToExpanded *string
		iconLinkTo         string
		want               string
	}{
		"iconLinkToExpanded overrides all": {
			iconLinkToExpanded: test.StringPtr("expanded-icon-link"),
			iconLinkTo:         "default-icon-link",
			want:               "expanded-icon-link",
		},
		"iconLinkTo is last resort": {
			iconLinkToExpanded: nil,
			iconLinkTo:         "default-icon-link",
			want:               "default-icon-link",
		},
		"empty string returned if both nil/empty": {
			iconLinkToExpanded: nil,
			iconLinkTo:         "",
			want:               "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := &Options{
				iconLinkToExpanded: tc.iconLinkToExpanded,
				IconLinkTo:         tc.iconLinkTo,
			}

			// WHEN GetIconLinkTo is called.
			got := options.GetIconLinkTo()

			// THEN the returned icon link matches the expected result.
			if got != tc.want {
				t.Errorf("%s\nresult mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestOptions_GetWebURL(t *testing.T) {
	// GIVEN an Options with various webURL-related fields set.
	tests := map[string]struct {
		webURLExpanded *string
		webURL         string
		want           string
	}{
		"webURLExpanded overrides all": {
			webURLExpanded: test.StringPtr("expanded-web-url"),
			webURL:         "default-web-url",
			want:           "expanded-web-url",
		},
		"webURL is last resort": {
			webURLExpanded: nil,
			webURL:         "default-web-url",
			want:           "default-web-url",
		},
		"empty string returned if both nil/empty": {
			webURLExpanded: nil,
			webURL:         "",
			want:           "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := &Options{
				webURLExpanded: tc.webURLExpanded,
				WebURL:         tc.webURL,
			}

			// WHEN GetWebURL is called.
			got := options.GetWebURL()

			// THEN the returned icon link matches the expected result.
			if got != tc.want {
				t.Errorf("%s\nresult mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestOptions_GetIcon(t *testing.T) {
	// GIVEN an Options with various icon-related fields set.
	tests := map[string]struct {
		iconExpanded *string
		icon         string
		iconNotify   *string
		want         string
	}{
		"iconExpanded overrides all": {
			iconExpanded: test.StringPtr("expanded-icon"),
			icon:         "default-icon",
			iconNotify:   test.StringPtr("notify-icon"),
			want:         "expanded-icon",
		},
		"icon overrides iconNotify": {
			iconExpanded: nil,
			icon:         "default-icon",
			iconNotify:   test.StringPtr("notify-icon"),
			want:         "default-icon",
		},
		"iconNotify is last resort": {
			iconExpanded: nil,
			icon:         "",
			iconNotify:   test.StringPtr("notify-icon"),
			want:         "notify-icon",
		},
		"empty string returned if all nil/empty": {
			iconExpanded: nil,
			icon:         "",
			iconNotify:   nil,
			want:         "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := &Options{
				iconExpanded: tc.iconExpanded,
				Icon:         tc.icon,
				iconNotify:   tc.iconNotify,
			}

			// WHEN GetIcon is called.
			got := options.GetIcon()

			// THEN the returned icon matches the expected result.
			if got != tc.want {
				t.Errorf("%s\nresult mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestOptions_CheckValues(t *testing.T) {
	// GIVEN Options.
	tests := map[string]struct {
		dashboardOptions *Options
		errRegex         string
	}{
		"nil": {
			errRegex:         `^$`,
			dashboardOptions: nil},
		"invalid web_url template": {
			errRegex:         `^web_url: ".*" <invalid>.*$`,
			dashboardOptions: &Options{WebURL: "https://release-argus.io/{{ version }"}},
		"valid web_url template": {
			errRegex:         `^$`,
			dashboardOptions: &Options{WebURL: "https://release-argus.io"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN CheckValues is called on it.
			err := tc.dashboardOptions.CheckValues("")

			// THEN the err is what we expect.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nmismatch\nwant: %d lines of error:\n%q\ngot:  %d\n\nlines:\n%v\n\nstdout: %q",
					packageName, wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}
		})
	}
}
