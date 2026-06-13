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

package web

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_DecodeSelf(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	// GIVEN: data in a given format to Decode into an existing Lookup.
	tests := []struct {
		name         string
		format, data string
		errRegex     string
		want         string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     `{}`,
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:   "JSON/invalid payload decode error",
			format: "json",
			data:   `{`,
			errRegex: test.TrimYAML(`
				^extract "require":
					[^\s]+ unexpected EOF`,
			),
		},
		{
			name:   "JSON/valid payload, no require",
			format: "json",
			data: test.TrimJSON(`{
				"type": "url",
				"allow_invalid_certs": true,
				"headers": [
					{"key": "x-Foo", "value": "bar"}
				]
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				allow_invalid_certs: true
				headers:
					- key: x-Foo
					  value: bar
			`),
		},
		{
			name:   "'web' -> 'url'",
			format: "json",
			data: test.TrimJSON(`{
				"type": "web"
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
			`),
		},
		{
			name:     "JSON/invalid data types",
			format:   "json",
			data:     `{"allow_invalid_certs": "true"}`,
			errRegex: `^json: .*unmarshal.*$`,
			want:     "type: url\n",
		},
		{
			name:   "JSON/require extraction, invalid type",
			format: "json",
			data: test.TrimJSON(`{
				"type":"url",
				"require": 123
			}`),
			errRegex: test.TrimYAML(`
				^require:
					extract "docker":
						json: .*unmarshal.* number.*$`,
			),
		},
		{
			name:   "JSON/valid require block",
			format: "json",
			data: test.TrimJSON(`{
				"type":"url",
				"require": {
					"regex_content": "v?"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				require:
					regex_content: v?
			`),
		},
		{
			name:   "JSON filled",
			format: "json",
			data: test.TrimJSON(`{
				"type": "url",
				"url": "https://example.com",
				"allow_invalid_certs": true,
				"headers": [
					{"key": "X-Foo", "value": "bar"}
				],
				"require": {
					"regex_version": "v?"
				}
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				url: https://example.com
				require:
					regex_version: v?
				allow_invalid_certs: true
				headers:
					- key: X-Foo
					  value: bar
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Lookup.
			options := opttest.PlainOptions(t, optCfg)
			svcStatus := &status.Status{}
			lookup := &Lookup{}
			lookup.Init(
				options,
				svcStatus,
				lvCfg,
			)

			// WHEN: DecodeSelf is called.
			decoded, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					err := lookup.DecodeSelf(format, data)
					return lookup, err
				},
				tc.format, tc.data,
				func(lv *Lookup) string { return decode.ToYAMLString(lv, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"Lookup.DecodeSelf",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if decoded == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nLookup.DecodeSelf(format=%q, data=%q)",
				tc.format, tc.format, tc.data,
			)

			// THEN: pointers are set as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: lvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: lvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLookup_ApplyOverrides(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	tests := []struct {
		name     string
		format   string
		data     string
		previous *Lookup
		errRegex string
		want     string
	}{
		{
			name:     "empty data returns previous",
			format:   "json",
			data:     "",
			previous: &Lookup{},
			errRegex: `^$`,
		},
		{
			name:     "invalid payload causes decode error",
			format:   "json",
			data:     `{`,
			previous: &Lookup{},
			errRegex: test.TrimYAML(`
				^extract "require":
					[^\s]+ unexpected EOF`,
			),
		},
		{
			name:     "override error - base.Lookup",
			format:   "json",
			data:     `{"url": []}`,
			previous: &Lookup{},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name:     "override error - Lookup",
			format:   "json",
			data:     `{"allow_invalid_certs": "true"}`,
			previous: &Lookup{},
			errRegex: `^json: .*unmarshal.* string.*$`,
		},
		{
			name:   "require removed",
			format: "json",
			data:   `{"require": null}`,
			previous: &Lookup{
				Lookup: base.Lookup{
					URL: "https://example.com",
					Require: &filter.Require{
						RegexContent: "v?",
					},
				},
			},
			want:     "url: https://example.com\n",
			errRegex: `^$`,
		},
		{
			name:   "valid require block",
			format: "json",
			data: test.TrimJSON(`{
				"type":"url",
				"require": {
					"regex_content": "v?"
				}
			}`),
			previous: &Lookup{},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				require:
					regex_content: v?
			`),
		},
		{
			name:   "previous require inherited",
			format: "json",
			data:   `{"type": "-"}`,
			previous: &Lookup{
				Lookup: base.Lookup{
					Require: &filter.Require{
						RegexContent: "v?",
					},
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: '-'
				require:
					regex_content: v?
			`),
		},
		{
			name:   "Headers added",
			format: "json",
			data: test.TrimJSON(`{
				"headers": [
					{"key": "X-Foo", "value": "bar"}
				]
			}`),
			previous: &Lookup{
				Lookup: base.Lookup{
					Type: "url",
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				headers:
					- key: X-Foo
					  value: bar
			`),
		},
		{
			name:   "Headers changed",
			format: "json",
			data: test.TrimJSON(`{
				"headers": [
					{"key": "X-Charlie", "value": "here"}
				]
			}`),
			previous: &Lookup{
				Lookup: base.Lookup{Type: "url"},
				Headers: shared.Headers{
					{Key: "X-Bar", Value: "foo"},
					{Key: "X-Foo", Value: "bar"},
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				headers:
					- key: X-Charlie
					  value: here
			`),
		},
		{
			name:   "Headers removed",
			format: "json",
			data:   `{"headers": null}`,
			previous: &Lookup{
				Lookup: base.Lookup{
					Type: "url",
				},
				Headers: shared.Headers{
					{Key: "X-Bar", Value: "foo"},
					{Key: "X-Foo", Value: "bar"},
				},
			},
			errRegex: `^$`,
			want:     "type: url\n",
		},
		{
			name:   "filled",
			format: "json",
			data: test.TrimJSON(`{
				"type": "url",
				"url": "https://release-argus.io",
				"allow_invalid_certs": false,
				"headers": [
					{"key": "X-Bar", "value": "foo"}
				],
				"require": {
					"regex_version": "v?"
				}
			}`),
			previous: &Lookup{
				Lookup: base.Lookup{
					Type: "url",
					URL:  "https://example.com",
				},
				AllowInvalidCerts: test.Ptr(true),
				Headers: shared.Headers{
					{Key: "X-Foo", Value: "bar"},
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				url: https://release-argus.io
				require:
					regex_version: v?
				allow_invalid_certs: false
				headers:
					- key: X-Bar
					  value: foo
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.previous.Defaults = lvCfg.Soft
			tc.previous.HardDefaults = lvCfg.Hard
			// Default want to the stringified struct.
			if tc.want == "" {
				tc.want = decode.ToYAMLString(tc.previous, "")
			}

			// WHEN: ApplyOverrides is called.
			if _, _, testErr := test.AssertApplyOverrides(
				t,
				tc.previous,
				func(format string, data []byte, v *Lookup) (*Lookup, error) {
					err := v.ApplyOverrides(format, data)
					return v, err
				},
				tc.format, tc.data,
				func(v *Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				true,
				packageName,
				"ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
