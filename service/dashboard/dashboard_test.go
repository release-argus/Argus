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

package dashboard

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

// ############
// # DECODING #
// ############

func TestDecode(t *testing.T) {
	dashCfg := plainDefaultsConfig(t)

	// GIVEN: data in a given format to Decode into Options.
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"auto_approve": true,
				"icon": "icon-url",
				"icon_link_to": "icon-link",
				"web_url": "web-url",
				"tags": [
					"tag1",
					"tag2"
				],
				"other": "foo"
			}`),
			want: test.TrimYAML(`
				auto_approve: true
				icon: icon-url
				icon_link_to: icon-link
				web_url: web-url
				tags:
					- tag1
					- tag2
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				auto_approve: true
				icon: icon-url
				icon_link_to: icon-link
				web_url: web-url
				tags:
					- tag1
					- tag2
				other: foo
			`),
			want: test.TrimYAML(`
				auto_approve: true
				icon: icon-url
				icon_link_to: icon-link
				web_url: web-url
				tags:
					- tag1
					- tag2
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/bool type mismatch",
			format: "yaml",
			data:   `auto_approve: yah`,
			errRegex: test.TrimYAML(`
				^dashboard:
					[^\s]+ .*unmarshal .*
					[^\s]+ .*auto_approve: yah
					\s+\^$`,
			),
		},
		{
			name:   "JSON/bool type mismatch",
			format: "json",
			data:   `{"auto_approve": "true"}`,
			errRegex: test.TrimYAML(`
				^dashboard:
					json:.* unmarshal.*$`,
			),
		},
		{
			name:   "JSON/invalid (trailing comma)",
			format: "json",
			data:   `{"auto_approve": "true",}`,
			errRegex: test.TrimYAML(`
				^dashboard:
					json:.* unmarshal.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Options, error) {
					return Decode(
						format, data,
						dashCfg,
					)
				},
				tc.format, tc.data,
				func(v *Options) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"Decode",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestOptions_Unmarshal(t *testing.T) {
	// GIVEN: a string in a given format to unmarshal into Options.
	tests := []struct {
		name         string
		format, data string
		errRegex     string
		want         string
	}{
		{
			name:   "JSON/empty",
			format: "json",
			data:   "",
			want:   "",
			errRegex: test.TrimYAML(`
				^jsontext:
					unexpected EOF$`,
			),
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     `{invalid: json}`,
			errRegex: `invalid character`,
			want:     "",
		},
		{
			name:   "YAML/invalid",
			format: "yaml",
			data:   `notMappingHere`,
			errRegex: test.TrimYAML(`
				^[^\s]+.*string was used where mapping is expected
				[^\s]+.*notMappingHere.*
				\s+\^$`,
			),
		},
		{
			name:   "JSON/tags/[]string",
			format: "json",
			data: test.TrimJSON(`{
				"tags": [
					"foo",
					"bar"
				]
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				tags:
					- foo
					- bar
			`),
		},
		{
			name:   "YAML/tags/[]string",
			format: "yaml",
			data: test.TrimYAML(`
				tags:
					- foo
					- bar
			`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				tags:
					- foo
					- bar
			`),
		},
		{
			name:   "JSON/tags/[]dict",
			format: "json",
			data: test.TrimJSON(`{
				"tags": [
					{"a": "foo"},
					{"b": "bar"}
				]
			}`),
			errRegex: `^tags: expected a string inside the list, got .*$`,
		},
		{
			name:   "YAML/tags/[]dict",
			format: "yaml",
			data: test.TrimYAML(`
				tags:
				- {foo}
				- {bar}
			`),
			errRegex: `^tags: expected a string inside the list, got .*$`,
		},
		{
			name:     "JSON/tags/string",
			format:   "json",
			data:     `{"tags": "foo"}`,
			errRegex: `^$`,
			want: test.TrimYAML(`
				tags:
					- foo
			`),
		},
		{
			name:     "YAML/tags/string",
			format:   "yaml",
			data:     `tags: foo`,
			errRegex: `^$`,
			want: test.TrimYAML(`
				tags:
					- foo
			`),
		},
		{
			name:   "JSON/tags/invalid",
			format: "json",
			data: test.TrimJSON(`{
				"tags": {
					"foo": "bar"
				}
			}`),
			errRegex: `^tags: expected a string or a list of strings, got map\[string\]interface.*$`,
		},
		{
			name:   "YAML/tags/invalid",
			format: "yaml",
			data: test.TrimYAML(`
				tags:
					foo: bar
			`),
			errRegex: `^tags: expected a string or a list of strings, got map\[string\]interface.*$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Options, error) {
					var zero Options
					err := decode.Unmarshal(format, data, &zero)
					return &zero, err
				},
				tc.format, tc.data,
				func(v *Options) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"Options",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// #########
// # STATE #
// #########

func TestOptions_IsZero(t *testing.T) {
	// GIVEN: Options.
	tests := []struct {
		name string
		opt  *Options
		want bool
	}{
		{
			name: "empty",
			opt:  &Options{},
			want: true,
		},
		{
			name: "ignored fields",
			opt: &Options{
				iconExpanded:       test.Ptr("foo"),
				iconNotify:         test.Ptr("bar"),
				iconLinkToExpanded: test.Ptr("baz"),
				webURLExpanded:     test.Ptr("qux"),
				Defaults: &Defaults{
					OptionsBase{
						Icon: "foo",
					},
				},
				HardDefaults: &Defaults{
					OptionsBase{
						Icon: "foo",
					},
				},
			},
			want: true,
		},
		{
			name: "non-empty/AutoApprove",
			opt: &Options{
				OptionsBase: OptionsBase{
					AutoApprove: test.Ptr(true),
				},
			},
			want: false,
		},
		{
			name: "non-empty/Icon",
			opt: &Options{
				OptionsBase: OptionsBase{
					Icon: "foo",
				},
			},
			want: false,
		},
		{
			name: "non-empty/IconLinkTo",
			opt: &Options{
				OptionsBase: OptionsBase{
					IconLinkTo: "foo",
				},
			},
			want: false,
		},
		{
			name: "non-empty/WebURL",
			opt: &Options{
				OptionsBase: OptionsBase{
					WebURL: "foo",
				},
			},
			want: false,
		},
		{
			name: "non-empty/Tags",
			opt: &Options{
				Tags: []string{"foo"},
			},
			want: false,
		},
		{
			name: "non-empty/all",
			opt: &Options{
				OptionsBase: OptionsBase{
					AutoApprove: test.Ptr(true),
					Icon:        "foo",
					IconLinkTo:  "bar",
					WebURL:      "baz",
				},
				Tags: []string{"foo"},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero() is called on it.
			got := tc.opt.IsZero()

			// THEN: it should return the expected value.
			if got != tc.want {
				t.Errorf(
					"%s\nOptions.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestOptions_Copy(t *testing.T) {
	// GIVEN: an Options.
	tests := []struct {
		name    string
		options *Options
		want    *Options
	}{
		{
			name:    "nil options",
			options: nil,
			want:    nil,
		},
		{
			name: "filled",
			options: &Options{
				OptionsBase: OptionsBase{
					AutoApprove: test.Ptr(true),
					Icon:        "icon-url",
					IconLinkTo:  "icon-link",
					WebURL:      "web-url",
				},
				iconExpanded:       test.Ptr("expanded-icon-url"),
				iconNotify:         test.Ptr("notify-icon-url"),
				iconLinkToExpanded: test.Ptr("expanded-icon-link"),
				webURLExpanded:     test.Ptr("expanded-web-url"),
			},
			want: &Options{
				OptionsBase: OptionsBase{
					AutoApprove: test.Ptr(true),
					Icon:        "icon-url",
					IconLinkTo:  "icon-link",
					WebURL:      "web-url",
				},
				iconExpanded:       test.Ptr("expanded-icon-url"),
				iconNotify:         test.Ptr("notify-icon-url"),
				iconLinkToExpanded: test.Ptr("expanded-icon-link"),
				webURLExpanded:     test.Ptr("expanded-web-url"),
			},
		},
		{
			name: "some fields nil",
			options: &Options{
				OptionsBase: OptionsBase{
					AutoApprove: nil,
					Icon:        "icon-url",
					IconLinkTo:  "",
					WebURL:      "web-url",
				},
				iconExpanded: test.Ptr("hi"),
			},
			want: &Options{
				OptionsBase: OptionsBase{
					AutoApprove: nil,
					Icon:        "icon-url",
					IconLinkTo:  "",
					WebURL:      "web-url",
				},
				iconExpanded: test.Ptr("hi"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Copy is called.
			result := tc.options.Copy()

			prefix := fmt.Sprintf("%s\nOptions.Copy()", packageName)

			// THEN: nil is returned if the Options are nil.
			if result == nil {
				if tc.options != nil {
					t.Errorf(
						"%s pointer mismatch\ngot: nil\nwant non-nil",
						prefix,
					)
				}
				return
			}
			// THEN: the returned struct is at a different address.
			if result == tc.options {
				t.Errorf(
					"%s gave the same address\ngot:  %p\nwant: NOT %p",
					prefix, result, tc.options,
				)
			}
			// THEN: the copied Options matches the expected result.
			want := decode.ToJSONString(tc.want)
			if got := decode.ToJSONString(result); got != want {
				t.Errorf(
					"%s mismatch\ngot:  %q\nwant: %q",
					prefix, got, want,
				)
			}
		})
	}
}

// ##########
// # VALUES #
// ##########

func TestOptions_GetAutoApprove(t *testing.T) {
	// GIVEN: a Options.
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      bool
	}{
		{
			name:             "root overrides all",
			want:             true,
			rootValue:        test.Ptr(true),
			defaultValue:     test.Ptr(false),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "default overrides hardDefault",
			want:             true,
			defaultValue:     test.Ptr(true),
			hardDefaultValue: test.Ptr(false),
		},
		{
			name:             "hardDefault is last resort",
			want:             true,
			hardDefaultValue: test.Ptr(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dashboard := Options{}
			dashboard.AutoApprove = tc.rootValue
			defaults, _ := DecodeDefaults("yaml", nil)
			defaults.AutoApprove = tc.defaultValue
			dashboard.Defaults = defaults
			hardDefaults, _ := DecodeDefaults("yaml", nil)
			hardDefaults.AutoApprove = tc.hardDefaultValue
			dashboard.HardDefaults = hardDefaults

			// WHEN: GetAutoApprove is called.
			got := dashboard.GetAutoApprove()

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s\nOptions.GetAutoApprove() mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestOptions_SetFallbackIcon(t *testing.T) {
	// GIVEN: an Options and a fallback icon URL.
	tests := []struct {
		name              string
		initialIconNotify *string
		newIconNotify     string
		want              *string
	}{
		{
			name:              "set new fallback icon",
			initialIconNotify: nil,
			newIconNotify:     "new-icon-url",
			want:              test.Ptr("new-icon-url"),
		},
		{
			name:              "overwrite existing fallback icon",
			initialIconNotify: test.Ptr("old-icon-url"),
			newIconNotify:     "new-icon-url",
			want:              test.Ptr("new-icon-url"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := &Options{
				iconNotify: tc.initialIconNotify,
			}

			// WHEN: SetFallbackIcon is called with a new icon URL.
			options.SetFallbackIcon(tc.newIconNotify)

			// THEN: the fallback icon is updated as expected.
			if options.iconNotify == nil || *options.iconNotify != *tc.want {
				t.Errorf(
					"%s\nOptions.SetFallbackIcon(%q) value mismatch\ngot:  %q\nwant: %q",
					packageName, tc.newIconNotify, util.DerefOrZero(options.iconNotify), *tc.want,
				)
			}
		})
	}
}

func TestOptions_GetIcon(t *testing.T) {
	// GIVEN: an Options with various icon-related fields set.
	tests := []struct {
		name         string
		iconExpanded *string
		icon         string
		iconNotify   *string
		want         string
	}{
		{
			name:         "iconExpanded overrides all",
			iconExpanded: test.Ptr("expanded-icon"),
			icon:         "default-icon",
			iconNotify:   test.Ptr("notify-icon"),
			want:         "expanded-icon",
		},
		{
			name:         "icon overrides iconNotify",
			iconExpanded: nil,
			icon:         "default-icon",
			iconNotify:   test.Ptr("notify-icon"),
			want:         "default-icon",
		},
		{
			name:         "iconNotify is last resort",
			iconExpanded: nil,
			icon:         "",
			iconNotify:   test.Ptr("notify-icon"),
			want:         "notify-icon",
		},
		{
			name:         "empty string returned if all nil or empty",
			iconExpanded: nil,
			icon:         "",
			iconNotify:   nil,
			want:         "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := &Options{
				OptionsBase: OptionsBase{
					Icon: tc.icon,
				},
				iconExpanded: tc.iconExpanded,
				iconNotify:   tc.iconNotify,
			}

			// WHEN: GetIcon is called.
			got := options.GetIcon()

			// THEN: the returned icon matches the expected result.
			if got != tc.want {
				t.Errorf(
					"%s\nOptions.GetIcon() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestOptions_GetIconLinkTo(t *testing.T) {
	// GIVEN: an Options with various iconLinkTo-related fields set.
	tests := []struct {
		name               string
		iconLinkToExpanded *string
		iconLinkTo         string
		want               string
	}{
		{
			name:               "iconLinkToExpanded overrides all",
			iconLinkToExpanded: test.Ptr("expanded-icon-link"),
			iconLinkTo:         "default-icon-link",
			want:               "expanded-icon-link",
		},
		{
			name:               "iconLinkTo is last resort",
			iconLinkToExpanded: nil,
			iconLinkTo:         "default-icon-link",
			want:               "default-icon-link",
		},
		{
			name:               "empty string returned if both nil or empty",
			iconLinkToExpanded: nil,
			iconLinkTo:         "",
			want:               "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := &Options{
				OptionsBase: OptionsBase{
					IconLinkTo: tc.iconLinkTo,
				},
				iconLinkToExpanded: tc.iconLinkToExpanded,
			}

			// WHEN: GetIconLinkTo is called.
			got := options.GetIconLinkTo()

			// THEN: the returned icon link matches the expected result.
			if got != tc.want {
				t.Errorf(
					"%s\nOptions.GetIconLinkTo() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestOptions_GetWebURL(t *testing.T) {
	// GIVEN: an Options with various webURL-related fields set.
	tests := []struct {
		name           string
		webURLExpanded *string
		webURL         string
		want           string
	}{
		{
			name:           "webURLExpanded overrides all",
			webURLExpanded: test.Ptr("expanded-web-url"),
			webURL:         "default-web-url",
			want:           "expanded-web-url",
		},
		{
			name:           "webURL is last resort",
			webURLExpanded: nil,
			webURL:         "default-web-url",
			want:           "default-web-url",
		},
		{
			name:           "empty string returned if both nil or empty",
			webURLExpanded: nil,
			webURL:         "",
			want:           "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := &Options{
				OptionsBase: OptionsBase{
					WebURL: tc.webURL,
				},
				webURLExpanded: tc.webURLExpanded,
			}

			// WHEN: GetWebURL is called.
			got := options.GetWebURL()

			// THEN: the returned icon link matches the expected result.
			if got != tc.want {
				t.Errorf(
					"%s\nOptions.GetWebURL() mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

// ############
// # DEFAULTS #
// ############

func TestOptions_SetDefaults(t *testing.T) {
	// GIVEN: Options.
	tests := []struct {
		name    string
		options *Options
	}{
		{
			name:    "empty",
			options: &Options{},
		},
		{
			name: "existing defaults - hardDefaults overwritten",
			options: &Options{
				Defaults:     &Defaults{},
				HardDefaults: &Defaults{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: two sets of Defaults.
			var defaults, hardDefaults *Defaults

			// WHEN: SetDefaults is called.
			tc.options.SetDefaults(
				defaults,
				hardDefaults,
			)

			prefix := fmt.Sprintf(
				"%s\nOptions SetDefaults(defaults=%p, hardDefaults=%p)",
				packageName, defaults, hardDefaults,
			)

			// THEN: the defaults are set as expected.
			if tc.options.Defaults != defaults {
				t.Errorf(
					"%s .Defaults pointer mismatch\ngot:  %v\nwant: %v",
					prefix, tc.options.Defaults, defaults,
				)
			}

			// AND: the hardDefaults are set as expected.
			if tc.options.HardDefaults != hardDefaults {
				t.Errorf(
					"%s .HardDefaults pointer mismatch\ngot:  %v\nwant: %v",
					prefix, tc.options.HardDefaults, hardDefaults,
				)
			}
		})
	}
}

// ##############
// # VALIDATION #
// ##############

func TestOptions_CheckValues(t *testing.T) {
	// GIVEN: Options.
	tests := []struct {
		name     string
		input    *Options
		errRegex string
	}{
		{
			name:     "nil",
			errRegex: `^$`,
			input:    (*Options)(nil),
		},
		{
			name:     "invalid web_url template",
			errRegex: `^web_url: ".*" <invalid>.*$`,
			input: &Options{
				OptionsBase: OptionsBase{
					WebURL: "https://release-argus.io/{{ version }",
				},
			},
		},
		{
			name:     "valid web_url template",
			errRegex: `^$`,
			input: &Options{
				OptionsBase: OptionsBase{
					WebURL: "https://release-argus.io",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)
		})
	}
}
