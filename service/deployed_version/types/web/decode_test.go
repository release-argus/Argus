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
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_DecodeSelf(t *testing.T) {
	dvCfg := plainDefaultsConfig(t)
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
			data:     ``,
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid payload decode error",
			format:   "json",
			data:     `{`,
			errRegex: `unexpected`,
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
			name:     "JSON/invalid data types",
			format:   "json",
			data:     `{"allow_invalid_certs": "true"}`,
			errRegex: `^json: .*unmarshal.*$`,
			want:     "type: url\n",
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"type": "url",
				"method": "GET",
				"url": "https://example.com",
				"allow_invalid_certs": true,
				"target_header": "X-Version",
				"basic_auth": {
					"username": "user",
					"password": "pass"
				},
				"headers": [
					{"key": "X-Foo", "value": "bar"}
				],
				"body": "foo",
				"json": "x.y.z",
				"regex": "v?[\\d.]+",
				"regex_template": "$0-beta"
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				method: GET
				url: https://example.com
				allow_invalid_certs: true
				target_header: X-Version
				basic_auth:
					username: user
					password: pass
				headers:
					- key: X-Foo
					  value: bar
				body: foo
				json: x.y.z
				regex: 'v?[\d.]+'
				regex_template: $0-beta
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
				dvCfg,
			)

			// WHEN: DecodeSelf is called.
			lookup, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					err := lookup.DecodeSelf(format, data)
					return lookup, err
				},
				tc.format, tc.data,
				func(v *Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"Lookup.DecodeSelf",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nLookup.DecodeSelf(format=%q, data=%q)",
				packageName, tc.format, tc.data,
			)

			// AND: Pointers are handed out to it correctly.
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: dvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: dvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLookup_ApplyOverrides(t *testing.T) {
	dvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	type Args struct {
		format, data string
		target       *Lookup
	}
	tests := []struct {
		name     string
		args     Args
		errRegex string
		want     string
	}{
		{
			name: "empty data returns previous",
			args: Args{
				format: "json",
				data:   "",
				target: &Lookup{},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid payload causes decode error",
			args: Args{
				format: "json",
				data:   `{`,
				target: &Lookup{},
			},
			errRegex: `unexpected`,
		},
		{
			name: "override error - base.Lookup",
			args: Args{
				format: "json",
				data:   `{"type": []}`,
				target: &Lookup{},
			},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "override error - Lookup",
			args: Args{
				format: "json",
				data:   `{"allow_invalid_certs": "true"}`,
				target: &Lookup{},
			},
			errRegex: `^json: .*unmarshal.* string.*$`,
		},
		{
			name: "Headers added",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"headers": [
						{"key": "X-Foo", "value": "bar"}
					]
				}`),
				target: &Lookup{
					Lookup: base.Lookup{
						Type: "web",
					},
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
			name: "Headers changed",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"headers": [
						{"key": "X-Charlie", "value": "here"}
					]
				}`),
				target: &Lookup{
					Lookup: base.Lookup{
						Type: "web",
					},
					Headers: shared.Headers{
						{Key: "X-Bar", Value: "foo"},
						{Key: "X-Foo", Value: "bar"},
					},
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
			name: "Headers removed",
			args: Args{
				format: "json",
				data:   `{"headers": null}`,
				target: &Lookup{
					Lookup: base.Lookup{
						Type: "web",
					},
					Headers: shared.Headers{
						{Key: "X-Bar", Value: "foo"},
						{Key: "X-Foo", Value: "bar"},
					},
				},
			},
			errRegex: `^$`,
			want:     "type: url\n",
		},
		{
			name: "filled",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "url",
					"method": "GET",
					"url": "https://release-argus.io",
					"allow_invalid_certs": false,
					"target_header": "X-Version-Other",
					"basic_auth": {
						"username": "foo",
						"password": "bar"
					},
					"headers": [
						{"key": "X-Bar", "value": "foo"}
					],
					"body": "foo",
					"json": "x.y.z",
					"regex": "v?[\\d.]+",
					"regex_template": "$0-beta"
				}`),
				target: &Lookup{
					Lookup: base.Lookup{
						Type: "url",
					},
					Method:            "POST",
					URL:               "https://example.com",
					AllowInvalidCerts: test.Ptr(true),
					BasicAuth: &BasicAuth{
						Username: "user",
						Password: "pass",
					},
					Headers: shared.Headers{
						{Key: "X-Foo", Value: "bar"},
					},
					Body:          "b",
					JSON:          "j",
					Regex:         "re",
					RegexTemplate: "template",
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				method: GET
				url: https://release-argus.io
				allow_invalid_certs: false
				target_header: X-Version-Other
				basic_auth:
					username: foo
					password: bar
				headers:
					- key: X-Bar
					  value: foo
				body: foo
				json: x.y.z
				regex: 'v?[\d.]+'
				regex_template: $0-beta
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: Options + Status.
			options := opttest.PlainOptions(t, optCfg)
			svcStatus := &status.Status{}
			tc.args.target.Init(options, svcStatus, dvCfg)
			// Default want to the unchanged stringified struct.
			if tc.want == "" {
				tc.want = decode.ToYAMLString(tc.args.target, "")
			}

			tc.args.target.Defaults = dvCfg.Soft
			tc.args.target.HardDefaults = dvCfg.Hard
			// Default want to the stringified struct.
			if tc.want == "" {
				tc.want = decode.ToYAMLString(tc.args.target, "")
			}

			// WHEN: ApplyOverrides is called.
			lookup, err, testErr := test.AssertApplyOverrides(
				t,
				tc.args.target,
				func(format string, data []byte, v *Lookup) (*Lookup, error) {
					err := v.ApplyOverrides(format, data)
					return v, err
				},
				tc.args.format, tc.args.data,
				func(v *Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				true,
				packageName,
				"Lookup.ApplyOverrides",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nLookup.ApplyOverrides(format=%q, data=%q)",
				packageName, tc.args.format, tc.args.data,
			)

			// AND: pointers are handed out as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: dvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: dvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
