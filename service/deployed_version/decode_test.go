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

package deployedver

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/deployed_version/types/web"
	opttest "github.com/release-argus/Argus/service/option/test"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util/polymorphic"
)

func TestDecode(t *testing.T) {
	dvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig()

	tests := []struct {
		name         string
		format, data string
		errRegex     string
		want         string
	}{
		{
			name:     "JSON/empty",
			data:     "",
			format:   "json",
			want:     "",
			errRegex: `^$`,
		},
		{
			name:     "JSON/empty object",
			data:     "{}",
			format:   "json",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			data:     "",
			format:   "yaml",
			want:     "",
			errRegex: `^$`,
		},
		{
			name: "JSON/url/minimal",
			data: test.TrimJSON(`{
				"type": "url",
				"url": "https://example.com"
			}`),
			format: "json",
			want: test.TrimYAML(`
				type: url
				url: https://example.com
			`),
			errRegex: `^$`,
		},
		{
			name: "YAML/url/minimal",
			data: test.TrimYAML(`
				type: url
				url: https://example.com
			`),
			format: "yaml",
			want: test.TrimYAML(`
				type: url
				url: https://example.com
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/url/filled",
			format: "yaml",
			data: test.TrimYAML(`
				type: url
				method: GET
				url: https://example.com
				allow_invalid_certs: false
				basic_auth:
					username: user
					password: pass
				headers:
					- key: X-Header
						value: val
					- key: X-Another
						value: val2
				body: removed_on_verify
				json: value.version
				regex: v([0-9.]+)
				regex_template: $1
			`),
			want: test.TrimYAML(`
				type: url
				method: GET
				url: https://example.com
				allow_invalid_certs: false
				basic_auth:
					username: user
					password: pass
				headers:
					- key: X-Header
						value: val
					- key: X-Another
						value: val2
				body: removed_on_verify
				json: value.version
				regex: v([0-9.]+)
				regex_template: $1
			`),
			errRegex: `^$`,
		},
		{
			name:   "invalid format",
			data:   `{"type": "url"}`,
			format: "xml",
			errRegex: test.TrimYAML(`
				^deployed_version:
					unsupported format: "xml"$`,
			),
		},
		{
			name: "JSON/unknown type",
			data: test.TrimJSON(`{
				"type": "unknown",
				"url": "https://example.com"
			}`),
			format: "json",
			errRegex: test.TrimYAML(`
				^deployed_version:
					type: "unknown" <invalid>.*$`,
			),
		},
		{
			name: "YAML/unknown type",
			data: test.TrimYAML(`
				type: unknown
				url: https://example.com
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				^deployed_version:
					type: "unknown" <invalid>.*$`,
			),
		},
		{
			name: "JSON/invalid format",
			data: test.TrimJSON(`{
				"type": "url",
				"url": https://example.com
			}`),
			format: "json",
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+ invalid character`,
			),
		},
		{
			name: "YAML/invalid format",
			data: test.TrimYAML(`
				type: url
				url: https://example.com
				invalid
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+ non-map value is specified.*`,
			),
		},
		{
			name: "JSON/invalid vars",
			data: test.TrimJSON(`{
				"type": "url",
				"url": ["https://example.com"]
			}`),
			format: "json",
			errRegex: test.TrimYAML(`
				^deployed_version:
					json: .*unmarshal .*`,
			),
		},
		{
			name: "YAML/invalid vars",
			data: test.TrimYAML(`
				type: url
				url: [https://example.com]
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+ .*unmarshal.*`,
			),
		},
		{
			name: "YAML/invalid URL format",
			data: test.TrimYAML(`
				type: url
				url:
					invalid: true
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+ .*unmarshal.*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := opttest.PlainOptions(optCfg)
			svcStatus, _ := statustest.New("yaml", nil)

			lookup, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (Lookup, error) {
					return Decode(
						format, data,
						options,
						svcStatus,
						dvCfg,
					)
				},
				tc.format, tc.data,
				func(v Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"Decode",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.GetOptions(), Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.GetStatus(), Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.GetDefaults(), Want: dvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.GetHardDefaults(), Want: dvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

type badInheritable struct{}

func (b *badInheritable) GetType() string                                 { return "bad" }
func (b *badInheritable) DecodeSelf(format string, data []byte) error     { return nil }
func (b *badInheritable) ApplyOverrides(format string, data []byte) error { return nil }

func TestDecode_TypeAssertionFailure(t *testing.T) {
	dvCfg := plainDefaultsConfig()

	orig := ServiceMapInheritable
	t.Cleanup(func() { ServiceMapInheritable = orig })

	bad := map[string]func() polymorphic.Inheritable{
		"bad": func() polymorphic.Inheritable {
			return &badInheritable{}
		},
	}
	ServiceMapInheritable = polymorphic.ToInheritableMap(bad)

	data := `{"type":"bad"}`

	_, err := Decode(
		"json", []byte(data),
		nil,
		nil,
		dvCfg,
	)

	want := "expected deployedver.Lookup"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error mismatch\ngot:  %v\nwant: %q", err, want)
	}
}

func TestApplyOverrides(t *testing.T) {
	dvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig()

	type Args struct {
		format, data string
		target       Lookup
	}

	tests := []struct {
		name        string
		args        Args
		sameAddress bool
		errRegex    string
		want        string
	}{
		{
			name: "empty data returns previous",
			args: Args{
				format: "json",
				data:   "",
				target: &mockLookup{},
			},
			sameAddress: true,
			want:        "lookup: {}\n",
			errRegex:    `^$`,
		},
		{
			name: "null data returns nil",
			args: Args{
				format: "json",
				data:   `null`,
				target: &mockLookup{},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid payload causes decode error",
			args: Args{
				format: "json",
				data:   `{`,
				target: &mockLookup{},
			},
			errRegex: test.TrimYAML(`
				^deployed_version:
					.*unexpected`,
			),
		},
		{
			name: "override error",
			args: Args{
				format: "json",
				data:   `{"OverrideErr": "yes"}`,
				target: &mockLookup{
					OverrideErr: "yes",
				},
			},
			errRegex: test.TrimYAML(`
				^deployed_version:
					yes$`,
			),
		},
		{
			name: "valid payload/new",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "url",
					"url": "http://example.com"
				}`),
				target: nil,
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				url: http://example.com
			`),
		},
		{
			name: "valid payload/edit/keep same type",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "url",
					"url": "http://example.com"
				}`),
				target: &web.Lookup{
					Method: "PUT",
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				method: PUT
				url: http://example.com
			`),
		},
		{
			name: "valid payload/edit/new type",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "manual",
					"version": "1.2.3"
				}`),
				target: &web.Lookup{
					Method: "PUT",
				},
			},
			errRegex:    `^$`,
			sameAddress: false,
			want: test.TrimYAML(`
				type: manual
				version: 1.2.3
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := opttest.PlainOptions(optCfg)
			svcStatus, _ := statustest.New("yaml", nil)

			if _, _, testErr := test.AssertApplyOverrides(
				t,
				tc.args.target,
				func(format string, data []byte, v Lookup) (Lookup, error) {
					return ApplyOverrides(
						format, data,
						v,
						options,
						svcStatus,
						dvCfg,
					)
				},
				tc.args.format, tc.args.data,
				func(v Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				tc.sameAddress,
				"ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
