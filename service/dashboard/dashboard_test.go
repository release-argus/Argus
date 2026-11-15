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
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

// ############
// # DECODING #
// ############

func TestOptions_Unmarshal(t *testing.T) {
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
				"Options.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	dashCfg := plainDefaultsConfig()

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
				"Decode",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

// #########
// # STATE #
// #########

func TestOptions_IsZero(t *testing.T) {
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
				iconExpanded:       new("foo"),
				iconNotify:         new("bar"),
				iconLinkToExpanded: new("baz"),
				webURLExpanded:     new("qux"),
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
				AutoApprove: new(true),
			},
			want: false,
		},
		{
			name: "non-empty/Icon",
			opt: &Options{
				Icon: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/IconLinkTo",
			opt: &Options{
				IconLinkTo: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/WebURL",
			opt: &Options{
				WebURL: "foo",
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
				AutoApprove: new(true),
				Icon:        "foo",
				IconLinkTo:  "bar",
				WebURL:      "baz",
				Tags:        []string{"foo"},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.opt.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestOptions_Copy(t *testing.T) {
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
				AutoApprove:        new(true),
				Icon:               "icon-url",
				IconLinkTo:         "icon-link",
				WebURL:             "web-url",
				iconExpanded:       new("expanded-icon-url"),
				iconNotify:         new("notify-icon-url"),
				iconLinkToExpanded: new("expanded-icon-link"),
				webURLExpanded:     new("expanded-web-url"),
			},
			want: &Options{
				AutoApprove:        new(true),
				Icon:               "icon-url",
				IconLinkTo:         "icon-link",
				WebURL:             "web-url",
				iconExpanded:       new("expanded-icon-url"),
				iconNotify:         new("notify-icon-url"),
				iconLinkToExpanded: new("expanded-icon-link"),
				webURLExpanded:     new("expanded-web-url"),
			},
		},
		{
			name: "some fields nil",
			options: &Options{
				AutoApprove:  nil,
				Icon:         "icon-url",
				IconLinkTo:   "",
				WebURL:       "web-url",
				iconExpanded: new("hi"),
			},
			want: &Options{
				AutoApprove:  nil,
				Icon:         "icon-url",
				IconLinkTo:   "",
				WebURL:       "web-url",
				iconExpanded: new("hi"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.options.Copy()

			if result == nil {
				if tc.options != nil {
					t.Error("Options.Copy() mismatch\ngot:  nil\nwant: non-nil")
				}
				return
			}

			if result == tc.options {
				t.Errorf("copy aliases the original\ngot:  %p\nwant: NOT %p", result, tc.options)
			}

			want := decode.ToJSONString(tc.want)
			if got := decode.ToJSONString(result); got != want {
				t.Errorf("Options mismatch\ngot:  %q\nwant: %q", got, want)
			}

			fieldTests := []test.FieldAssertion{
				{Name: "iconExpanded", Got: test.StringifyPtr(result.iconExpanded), Want: test.StringifyPtr(tc.want.iconExpanded), Mode: test.CompareEqual},
				{Name: "iconNotify", Got: test.StringifyPtr(result.iconNotify), Want: test.StringifyPtr(tc.want.iconNotify), Mode: test.CompareEqual},
				{Name: "iconLinkToExpanded", Got: test.StringifyPtr(result.iconLinkToExpanded), Want: test.StringifyPtr(tc.want.iconLinkToExpanded), Mode: test.CompareEqual},
				{Name: "webURLExpanded", Got: test.StringifyPtr(result.webURLExpanded), Want: test.StringifyPtr(tc.want.webURLExpanded), Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Options"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// ##########
// # VALUES #
// ##########

func TestOptions_GetAutoApprove(t *testing.T) {
	tests := []struct {
		name                                      string
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      bool
	}{
		{
			name:             "root overrides all",
			want:             true,
			rootValue:        new(true),
			defaultValue:     new(false),
			hardDefaultValue: new(false),
		},
		{
			name:             "default overrides hardDefault",
			want:             true,
			defaultValue:     new(true),
			hardDefaultValue: new(false),
		},
		{
			name:             "hardDefault is last resort",
			want:             true,
			hardDefaultValue: new(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dashboard := Options{}
			dashboard.AutoApprove = tc.rootValue
			dashboard.Defaults = &Defaults{}
			dashboard.Defaults.AutoApprove = tc.defaultValue
			dashboard.HardDefaults = &Defaults{}
			dashboard.HardDefaults.AutoApprove = tc.hardDefaultValue

			got := dashboard.GetAutoApprove()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestOptions_SetFallbackIcon(t *testing.T) {
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
			want:              new("new-icon-url"),
		},
		{
			name:              "overwrite existing fallback icon",
			initialIconNotify: new("old-icon-url"),
			newIconNotify:     "new-icon-url",
			want:              new("new-icon-url"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := &Options{
				iconNotify: tc.initialIconNotify,
			}

			options.SetFallbackIcon(tc.newIconNotify)

			if options.iconNotify == nil || *options.iconNotify != *tc.want {
				t.Errorf(
					"Options.iconNotify mismatch\ngot:  %q\nwant: %q",
					util.DerefOrZero(options.iconNotify), *tc.want,
				)
			}
		})
	}
}

func TestOptions_GetIcon(t *testing.T) {
	tests := []struct {
		name         string
		iconExpanded *string
		icon         string
		iconNotify   *string
		want         string
	}{
		{
			name:         "iconExpanded overrides all",
			iconExpanded: new("expanded-icon"),
			icon:         "default-icon",
			iconNotify:   new("notify-icon"),
			want:         "expanded-icon",
		},
		{
			name:         "icon overrides iconNotify",
			iconExpanded: nil,
			icon:         "default-icon",
			iconNotify:   new("notify-icon"),
			want:         "default-icon",
		},
		{
			name:         "iconNotify is last resort",
			iconExpanded: nil,
			icon:         "",
			iconNotify:   new("notify-icon"),
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
				Icon:         tc.icon,
				iconExpanded: tc.iconExpanded,
				iconNotify:   tc.iconNotify,
			}

			got := options.GetIcon()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestOptions_GetIconLinkTo(t *testing.T) {
	tests := []struct {
		name               string
		iconLinkToExpanded *string
		iconLinkTo         string
		want               string
	}{
		{
			name:               "iconLinkToExpanded overrides all",
			iconLinkToExpanded: new("expanded-icon-link"),
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
				IconLinkTo:         tc.iconLinkTo,
				iconLinkToExpanded: tc.iconLinkToExpanded,
			}

			got := options.GetIconLinkTo()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestOptions_GetWebURL(t *testing.T) {
	tests := []struct {
		name           string
		webURLExpanded *string
		webURL         string
		want           string
	}{
		{
			name:           "webURLExpanded overrides all",
			webURLExpanded: new("expanded-web-url"),
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
				WebURL:         tc.webURL,
				webURLExpanded: tc.webURLExpanded,
			}

			got := options.GetWebURL()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

// ############
// # DEFAULTS #
// ############

func TestOptions_SetDefaults(t *testing.T) {
	tests := []struct {
		name    string
		options *Options
	}{
		{
			name:    "empty",
			options: &Options{},
		},
		{
			name: "existing defaults, hardDefaults overwritten",
			options: &Options{
				Defaults:     &Defaults{},
				HardDefaults: &Defaults{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defaults, hardDefaults := &Defaults{}, &Defaults{}

			tc.options.SetDefaults(defaults, hardDefaults)

			if tc.options.Defaults != defaults {
				t.Errorf(
					"Options.Defaults mismatch\ngot:  %p\nwant: %p",
					tc.options.Defaults, defaults,
				)
			}

			if tc.options.HardDefaults != hardDefaults {
				t.Errorf(
					"Options.HardDefaults mismatch\ngot:  %p\nwant: %p",
					tc.options.HardDefaults, hardDefaults,
				)
			}
		})
	}
}

// ##############
// # VALIDATION #
// ##############

func TestOptions_CheckValues(t *testing.T) {
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
				WebURL: "https://release-argus.io/{{ version }",
			},
		},
		{
			name:     "valid web_url template",
			errRegex: `^$`,
			input: &Options{
				WebURL: "https://release-argus.io",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_ = test.AssertCheckValuesWithError(
				t,
				tc.errRegex,
				tc.input.CheckValues,
			)
		})
	}
}
