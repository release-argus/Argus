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
)

func TestDecodeDefaults(t *testing.T) {
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
				^dashboard:
					jsontext:
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
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"auto_approve": false,
				"icon": "https://example.com/icon_here",
				"icon_link_to": "https://example.com/not_an_icon",
				"web_url": "https://example.com/somewhere"
			}`),
			want: test.TrimYAML(`
				auto_approve: false
				icon: https://example.com/icon_here
				icon_link_to: https://example.com/not_an_icon
				web_url: https://example.com/somewhere
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				auto_approve: false
				icon: https://example.com/icon_here
				icon_link_to: https://example.com/not_an_icon
				web_url: https://example.com/somewhere
			`),
			want: test.TrimYAML(`
				auto_approve: false
				icon: https://example.com/icon_here
				icon_link_to: https://example.com/not_an_icon
				web_url: https://example.com/somewhere
			`),
			errRegex: `^$`,
		},
		{
			name:   "invalid format",
			data:   `{"type": "url"}`,
			format: "xml",
			errRegex: test.TrimYAML(`
				^dashboard:
					unsupported format: "xml"$`,
			),
		},
		{
			name:   "JSON/invalid format",
			data:   `{"icon": https://example.com}`,
			format: "json",
			errRegex: test.TrimYAML(`
				^dashboard:
					[^\s]+ invalid character .*`,
			),
		},
		{
			name: "YAML/invalid format",
			data: test.TrimYAML(`
				icon: https://example.com
				invalid
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				^dashboard:
					[^\s]+ non-map value is specified.*`,
			),
		},
		{
			name:   "JSON/invalid vars",
			data:   `{"icon": ["https://example.com"]}`,
			format: "json",
			errRegex: test.TrimYAML(`
				^dashboard:
					json: .*unmarshal .*`,
			),
		},
		{
			name:   "YAML/invalid data type",
			data:   `icon: [https://example.com]`,
			format: "yaml",
			errRegex: test.TrimYAML(`
				^dashboard:
					[^\s]+ .*unmarshal.*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				DecodeDefaults,
				tc.format, tc.data,
				func(v *Defaults) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		opt  *Defaults
		want bool
	}{
		{
			name: "empty",
			opt:  &Defaults{},
			want: true,
		},
		{
			name: "non-empty/AutoApprove",
			opt: &Defaults{
				AutoApprove: new(true),
			},
			want: false,
		},
		{
			name: "non-empty/Icon",
			opt: &Defaults{
				Icon: "icon-url",
			},
			want: false,
		},
		{
			name: "non-empty/IconLinkTo",
			opt: &Defaults{
				IconLinkTo: "icon-link",
			},
			want: false,
		},
		{
			name: "non-empty/WebURL",
			opt: &Defaults{
				WebURL: "web-url",
			},
			want: false,
		},
		{
			name: "non-empty/all",
			opt: &Defaults{
				AutoApprove: new(true),
				Icon:        "icon-url",
				IconLinkTo:  "icon-link",
				WebURL:      "web-url",
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

func TestDefaults_Default(t *testing.T) {
	defaults := Defaults{}

	defaults.Default()

	if defaults.AutoApprove == nil || *defaults.AutoApprove {
		t.Errorf(
			"Defaults.AutoApprove mismatch\ngot:  %s\nwant: false",
			test.StringifyPtr(defaults.AutoApprove),
		)
	}
}
