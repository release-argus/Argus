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

package latestver

import (
	"testing"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestNew(t *testing.T) {
	// GIVEN a set of args to create a Lookup.
	type args struct {
		lType                  string
		overrides              any
		semanticVersioning     *bool
		defaults, hardDefaults *base.Defaults
	}
	tests := map[string]struct {
		args     args
		wantYAML string
		errRegex string
	}{
		"github - bare": {
			args: args{
				lType: "github",
				overrides: `
					url: release-argus/Argus
				`,
				defaults:     &base.Defaults{},
				hardDefaults: &base.Defaults{},
			},
			wantYAML: test.TrimYAML(`
				type: github
				url: release-argus/Argus
			`),
		},
		"github - full": {
			args: args{
				lType: "github",
				overrides: `
					url: release-argus/Argus
					access_token: token
					url_commands:
						- type: split
							text: v
					require:
						regex_version: v[\d.]+
					allow_invalid_certs: true
					use_prerelease: true
				`,
				semanticVersioning: test.BoolPtr(false),
				defaults:           &base.Defaults{},
				hardDefaults:       &base.Defaults{},
			},
			wantYAML: test.TrimYAML(`
				type: github
				url: release-argus/Argus
				url_commands:
					- type: split
						text: v
				require:
					regex_version: v[\d.]+
				access_token: token
				use_prerelease: true
			`),
		},
		"url - bare": {
			args: args{
				lType: "url",
				overrides: `
					url: https://example.com
					`,
				defaults:     &base.Defaults{},
				hardDefaults: &base.Defaults{},
			},
			wantYAML: test.TrimYAML(`
				type: url
				url: https://example.com
				`),
		},
		"url - full": {
			args: args{
				lType: "url",
				overrides: `
					url: release-argus/Argus
					access_token: token
					url_commands:
					- type: split
						text: v
					require:
						regex_version: v[\d.]+
					allow_invalid_certs: false
					use_prerelease: true
				`,
				semanticVersioning: test.BoolPtr(true),
				defaults:           &base.Defaults{},
				hardDefaults:       &base.Defaults{},
			},
			wantYAML: test.TrimYAML(`
				type: url
				url: release-argus/Argus
				url_commands:
					- type: split
						text: v
				require:
					regex_version: v[\d.]+
				allow_invalid_certs: false
				`),
		},
		"invalid type": {
			args: args{
				lType: "foo",
				overrides: `
					url: release-argus/Argus
				`,
				defaults:     &base.Defaults{},
				hardDefaults: &base.Defaults{},
			},
			errRegex: `type: "foo" <invalid> .*$`,
		},
		"GitHub - invalid configData type": {
			args: args{
				lType:        "github",
				overrides:    1,
				defaults:     &base.Defaults{},
				hardDefaults: &base.Defaults{},
			},
			errRegex: test.TrimYAML(`
				^failed to unmarshal github.Lookup:
					unsupported configData type: int$`),
		},
		"URL - invalid configData type": {
			args: args{
				lType:        "url",
				overrides:    1,
				defaults:     &base.Defaults{},
				hardDefaults: &base.Defaults{},
			},
			errRegex: test.TrimYAML(`
				^failed to unmarshal web.Lookup:
					unsupported configData type: int$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if overrides, ok := tc.args.overrides.(string); ok {
				tc.args.overrides = test.TrimYAML(overrides)
			}

			// WHEN New is called with the args.
			got, err := New(
				tc.args.lType,
				"yaml", tc.args.overrides,
				opt.New(
					nil, "",
					tc.args.semanticVersioning,
					nil, nil),
				nil,
				tc.args.defaults, tc.args.hardDefaults)

			// THEN any error is as expected.
			if err != nil {
				if !util.RegexCheck(tc.errRegex, err.Error()) {
					t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.errRegex, err)
				}
				return
			}
			// THEN the correct type is returned.
			if gotType := getType(got); gotType != tc.args.lType {
				t.Fatalf("%s\nType mismatch\nwant: %q\ngot:  %q (%T)",
					packageName, tc.args.lType, gotType, got)
			}

			// AND the variables are set correctly.
			gotYAML := util.ToYAMLString(got, "")
			if gotYAML != tc.wantYAML {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantYAML, gotYAML)
			}
			if got.GetDefaults() != tc.args.defaults {
				t.Errorf("%s\nDefaults mismatch\nwant: %p\ngot:  %p",
					packageName, tc.args.defaults, got.GetDefaults())
			}
			if got.GetHardDefaults() != tc.args.hardDefaults {
				t.Errorf("%s\nHardDefaults mismatch\nwant: %p\ngot:  %p",
					packageName, tc.args.hardDefaults, got.GetHardDefaults())
			}
		})
	}
}

func TestCopy(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		lookup   Lookup
		wantYAML string
	}{
		"nil": {
			lookup: nil,
		},
		"github": {
			lookup: test.IgnoreError(t, func() (base.Interface, error) {
				return New(
					"github",
					"yaml", test.TrimYAML(
						test.TrimYAML(`
					url: release-argus/Argus
					url_commands:
						- type: split
							text: v
					require:
						regex_version: v[\d.]+
					use_prerelease: true
			`)),
					&opt.Options{},
					status.New(
						nil, nil, nil,
						"",
						"", "",
						"", "",
						"",
						&dashboard.Options{}),
					&base.Defaults{}, &base.Defaults{})
			}),
		},
		"url": {
			lookup: test.IgnoreError(t, func() (base.Interface, error) {
				return New(
					"url",
					"yaml", test.TrimYAML(`
						url: release-argus/Argus
						url_commands:
							- type: split
								text: v
						require:
							regex_version: v[\d.]+
						allow_invalid_certs: true
					`),
					&opt.Options{},
					status.New(
						nil, nil, nil,
						"",
						"", "",
						"", "",
						"",
						&dashboard.Options{}),
					&base.Defaults{}, &base.Defaults{})
			}),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Copy is called.
			got := Copy(tc.lookup)

			// THEN the variables are copied over.
			gotYAML := util.ToYAMLString(got, "")
			wantYAML := util.ToYAMLString(tc.lookup, "")
			if gotYAML != wantYAML {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, wantYAML, gotYAML)
			}
			if tc.lookup == nil {
				return // No further checks.
			}

			if got.GetOptions() == tc.lookup.GetOptions() {
				t.Errorf("%s\nOptions shouldn't point to the same memory address",
					packageName)
			} else if got.GetOptions().String() != tc.lookup.GetOptions().String() {
				t.Errorf("%s\nOptions mismatch\nwant: %q\ngot:  %q",
					packageName, tc.lookup.GetOptions(), got.GetOptions())
			}

			if got.GetDefaults() != tc.lookup.GetDefaults() {
				t.Errorf("%s\nDefaults mismatch\nwant: %v\ngot:  %v",
					packageName, tc.lookup.GetDefaults(), got.GetDefaults())
			}

			if got.GetHardDefaults() != tc.lookup.GetHardDefaults() {
				t.Errorf("%s\nHardDefaults mismatch\nwant: %v\ngot:  %v",
					packageName, tc.lookup.GetHardDefaults(), got.GetHardDefaults())
			}
		})
	}
}

func TestChangeType(t *testing.T) {
	type args struct {
		newType   string
		lookup    Lookup
		overrides string
	}

	// GIVEN a Lookup.
	tests := map[string]struct {
		args     args
		wantYAML string
		errRegex string
	}{
		"github -> github": {
			args: args{
				lookup: test.IgnoreError(t, func() (base.Interface, error) {
					return New(
						"github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
							url_commands:
								- type: split
									text: v
							require:
								regex_version: v[\d.]+
							use_prerelease: true
						`),
						&opt.Options{},
						&status.Status{},
						&base.Defaults{}, &base.Defaults{})
				}),
				newType: "github",
			},
			wantYAML: `
				type: github
				url: release-argus/Argus
				url_commands:
					- type: split
						text: v
				require:
					regex_version: v[\d.]+
				use_prerelease: true
				`,
		},
		"github -> url": {
			args: args{
				lookup: test.IgnoreError(t, func() (base.Interface, error) {
					return New(
						"github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
							url_commands:
								- type: split
									text: v
							require:
								regex_version: v[\d.]+
							use_prerelease: true
						`),
						&opt.Options{},
						&status.Status{},
						&base.Defaults{}, &base.Defaults{})
				}),
				newType: "url",
				overrides: test.TrimYAML(`
					access_token: token
					allow_invalid_certs: true
				`),
			},
			wantYAML: `
				type: url
				url: release-argus/Argus
				url_commands:
					- type: split
						text: v
				require:
					regex_version: v[\d.]+
				allow_invalid_certs: true
				`,
		},
		"url -> url": {
			args: args{
				lookup: test.IgnoreError(t, func() (base.Interface, error) {
					return New(
						"url",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
							url_commands:
								- type: split
									text: v
							require:
								regex_version: v[\d.]+
							allow_invalid_certs: true
						`),
						&opt.Options{},
						&status.Status{},
						&base.Defaults{}, &base.Defaults{})
				}),
				newType: "url",
			},
			wantYAML: `
				type: url
				url: release-argus/Argus
				url_commands:
					- type: split
						text: v
				require:
					regex_version: v[\d.]+
				allow_invalid_certs: true
				`,
		},
		"url -> github": {
			args: args{
				lookup: test.IgnoreError(t, func() (base.Interface, error) {
					return New(
						"github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
							require:
								regex_version: v[\d.]+
							allow_invalid_certs: true
							url_commands:
								- type: split
									text: v
						`),
						&opt.Options{},
						&status.Status{},
						&base.Defaults{}, &base.Defaults{})
				}),
				newType: "github",
			},
			wantYAML: `
				type: github
				url: release-argus/Argus
				url_commands:
					- type: split
						text: v
				require:
					regex_version: v[\d.]+
				`,
		},
		"url -> unknown": {
			args: args{
				lookup: test.IgnoreError(t, func() (base.Interface, error) {
					return New(
						"github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
							url_commands:
								- type: split
									text: v
							require:
								regex_version: v[\d.]+
							allow_invalid_certs: true
						`),
						&opt.Options{},
						&status.Status{},
						&base.Defaults{}, &base.Defaults{})
				}),
				newType: "foo",
			},
			errRegex: `type: "foo" <invalid>.*$`,
			wantYAML: "",
		},
		"invalid YAML": {
			args: args{
				lookup: test.IgnoreError(t, func() (base.Interface, error) {
					return New(
						"github",
						"yaml", test.TrimYAML(`
							url: release-argus/Argus
							url_commands:
								- type: split
									text: v
							require:
								regex_version: v[\d.]+
							allow_invalid_certs: true
						`),
						&opt.Options{},
						&status.Status{},
						&base.Defaults{}, &base.Defaults{})
				}),
				newType:   "github",
				overrides: "invalid",
			},
			errRegex: `cannot unmarshal`,
			wantYAML: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ChangeType is called.
			gotLookup, err := ChangeType(
				tc.args.newType,
				tc.args.lookup,
				tc.args.overrides)

			// THEN the correct type is returned.
			if gotLookup == nil {
				// Unknown type.
				if tc.errRegex != "" {
					if !util.RegexCheck(tc.errRegex, err.Error()) {
						t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
							packageName, tc.errRegex, err)
					}
				} else {
					t.Errorf("%s\nType mismatch\nwant%q\ngot:  nil\nerr=%c",
						packageName, tc.args.newType, err)
				}
				return
			}
			gotType := gotLookup.GetType()
			if gotType != tc.args.newType {
				t.Fatalf("%s\nType mismatch\nwant: %q\ngot:  %T",
					packageName, tc.args.newType, gotType)
			}

			// AND the variables are copied over.
			gotYAML := util.ToYAMLString(gotLookup, "")
			tc.wantYAML = test.TrimYAML(tc.wantYAML)
			if gotYAML != tc.wantYAML {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantYAML, gotYAML)
			}

			if gotLookup.GetOptions() != tc.args.lookup.GetOptions() &&
				tc.args.lookup.GetOptions() != nil || gotLookup.GetOptions() == nil {
				t.Errorf("%s\nOptions mismatch\nwant: %v\ngot:  %v",
					packageName, tc.args.lookup.GetOptions(), gotLookup.GetOptions())
			}

			if gotLookup.GetDefaults() != tc.args.lookup.GetDefaults() {
				t.Errorf("%s\nDefaults mismatch\nwant: %v\ngot:  %v",
					packageName, tc.args.lookup.GetDefaults(), gotLookup.GetDefaults())
			}

			if gotLookup.GetHardDefaults() != tc.args.lookup.GetHardDefaults() {
				t.Errorf("%s\nHardDefaults mismatch\nwant: %v\ngot:  %v",
					packageName, tc.args.lookup.GetHardDefaults(), gotLookup.GetHardDefaults())
			}
		})
	}
}

func TestIsEqual(t *testing.T) {
	// GIVEN two Lookups.
	tests := map[string]struct {
		a, b Lookup
		want bool
	}{
		"empty": {
			a:    &web.Lookup{},
			b:    &web.Lookup{},
			want: true,
		},
		"defaults ignored": {
			a: &web.Lookup{
				Lookup: base.Lookup{
					Defaults: &base.Defaults{
						AllowInvalidCerts: test.BoolPtr(false)}}},
			b:    &web.Lookup{},
			want: true,
		},
		"hard_defaults ignored": {
			a: &web.Lookup{
				Lookup: base.Lookup{
					Defaults: &base.Defaults{
						AllowInvalidCerts: test.BoolPtr(false)}}},
			b:    &web.Lookup{},
			want: true,
		},
		"equal - url": {
			a: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"url",
					"yaml", test.TrimYAML(`
						url: https://example.com
						allow_invalid_certs: false
						url_commands:
							- type: split
								text: v
						require:
							regex_version: v([0-9.]+)`),
					nil,
					nil,
					nil, nil)
			}),
			b: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"url",
					"yaml", test.TrimYAML(`
						url: https://example.com
						allow_invalid_certs: false
						url_commands:
							- type: split
								text: v
						require:
							regex_version: v([0-9.]+)`),
					nil,
					nil,
					nil, nil)
			}),
			want: true,
		},
		"equal - github": {
			a: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"github",
					"yaml", test.TrimYAML(`
						url: release-argus/Argus
						access_token: token
						url_commands:
							- type: split
								text: v
						require:
							regex_version: v([0-9.]+)`),
					nil,
					nil,
					nil, nil)
			}),
			b: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"github",
					"yaml", test.TrimYAML(`
						url: release-argus/Argus
						access_token: token
						url_commands:
							- type: split
								text: v
						require:
							regex_version: v([0-9.]+)`),
					nil,
					nil,
					nil, nil)
			}),
			want: true,
		},
		"not equal": {
			a: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"github",
					"yaml", test.TrimYAML(`
						url: release-argus/Argus`),
					nil,
					nil,
					nil, nil)
			}),
			b: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"url",
					"yaml", test.TrimYAML(`
						url: release-argus/ARGUS`),
					nil,
					nil,
					nil, nil)
			}),
			want: false,
		},
		"not equal with nil": {
			a: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"github",
					"yaml", test.TrimYAML(`
						url: release-argus/Argus`),
					nil,
					nil,
					nil, nil)
			}),
			b:    nil,
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the two Lookups are compared.
			got := IsEqual(tc.a, tc.b)

			// THEN the result is as expected.
			if got != tc.want {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.want, got)
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		jsonStr  string
		errRegex string
		wantJSON *string
	}{
		"empty": {
			jsonStr:  "",
			errRegex: `unexpected end of JSON input`,
			wantJSON: test.StringPtr(""),
		},
		"invalid formatting": {
			jsonStr: "invalid",
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					invalid character .*$`),
		},
		"valid - GitHub": {
			jsonStr: test.TrimJSON(`{
				"type":"github",
				"url":"release-argus/Argus",
				"access_token":"token"
			}`),
			errRegex: `^$`,
		},
		"valid - URL": {
			jsonStr: test.TrimJSON(`{
				"type":"url",
				"url":"https://example.com",
				"allow_invalid_certs":true
			}`),
			errRegex: `^$`,
		},
		"invalid - GitHub": {
			jsonStr: test.TrimJSON(`{"
				type":"github",
				"url":"release-argus/Argus",
				"access_token":1,
				"use_prerelease":"sometimes"
			}`),
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					failed to unmarshal github.Lookup:
						cannot unmarshal number.* \.access_token of type string$`),
		},
		"invalid - URL": {
			jsonStr: test.TrimJSON(`{
				"type":"url",
				"url":"https://example.com",
				"allow_invalid_certs":"true",
				"target_header":true
			}`),
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					failed to unmarshal web.Lookup:
						cannot unmarshal string.* \.allow_invalid_certs of type bool$`),
		},
		"unknown type": {
			jsonStr: test.TrimJSON(`{
				"type":"foo",
				"url":"https://example.com",
				"allow_invalid_certs":true
			}`),
			errRegex: `failed to unmarshal latestver.Lookup`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.wantJSON == nil {
				tc.wantJSON = &tc.jsonStr
			}

			// WHEN UnmarshalJSON is called.
			lookupJSON, errJSON := UnmarshalJSON([]byte(tc.jsonStr))

			// THEN any error is as expected.
			eJSON := util.ErrorToString(errJSON)
			if !util.RegexCheck(tc.errRegex, eJSON) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, eJSON)
			}
			if tc.errRegex != "^$" {
				return
			}
			// AND the Lookup is unmarshalled as expected.
			gotFromJSON := util.ToJSONString(lookupJSON)
			if gotFromJSON != *tc.wantJSON {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, *tc.wantJSON, gotFromJSON)
			}
		})
	}
}

func TestUnmarshalYAML(t *testing.T) {
	tests := map[string]struct {
		yamlStr  string
		errRegex string
		wantYAML *string
	}{
		"Empty": {
			yamlStr:  "",
			errRegex: `failed to unmarshal`,
			wantYAML: test.StringPtr(""),
		},
		"Invalid formatting": {
			yamlStr: "{ invalid",
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					line \d: did not find expected.*$`),
		},
		"Valid - GitHub": {
			yamlStr: test.TrimYAML(`
				type: github
				url: release-argus/Argus
				access_token: token
			`),
			errRegex: `^$`,
		},
		"Valid - URL": {
			yamlStr: test.TrimYAML(`
				type: url
				url: https://example.com
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		"Invalid - GitHub": {
			yamlStr: test.TrimYAML(`
				type: github
				url: release-argus/Argus
				access_token:
					sub: token
			`),
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					failed to unmarshal github.Lookup:
						line \d: cannot unmarshal.*$`),
		},
		"Invalid - URL": {
			yamlStr: test.TrimYAML(`
				type: url
				url: https://example.com
				allow_invalid_certs: "true"
			`),
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					failed to unmarshal web.Lookup:
						line \d: cannot unmarshal.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.wantYAML == nil {
				tc.wantYAML = &tc.yamlStr
			}

			// WHEN UnmarshalYAML is called.
			lookupYAML, errYAML := UnmarshalYAML([]byte(tc.yamlStr))

			// THEN any error is as expected.
			eYAML := util.ErrorToString(errYAML)
			if !util.RegexCheck(tc.errRegex, eYAML) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, eYAML)
			}
			if tc.errRegex != "^$" {
				return
			}
			// AND the Lookup is unmarshalled as expected.
			gotFromYAML := lookupYAML.String(lookupYAML, "")
			if gotFromYAML != *tc.wantYAML {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, *tc.wantYAML, gotFromYAML)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := map[string]struct {
		format, data string
		wantType     string
		errRegex     string
	}{
		"valid JSON - GitHub": {
			data: test.TrimJSON(`{
				"type": "github",
				"url": "release-argus/Argus"
			}`),
			format:   "json",
			wantType: "github",
		},
		"valid JSON - URL": {
			data: test.TrimJSON(`{
				"type": "url",
				"url": "https://example.com"
			}`),
			format:   "json",
			wantType: "url",
		},
		"valid YAML - GitHub": {
			data: test.TrimYAML(`
				type: github
				url: release-argus/Argus
			`),
			format:   "yaml",
			wantType: "github",
		},
		"valid YAML - URL": {
			data: test.TrimYAML(`
				type: url
				url: https://example.com
			`),
			format:   "yaml",
			wantType: "url",
		},
		"invalid format": {
			data:     `{"type": "github"}`,
			format:   "xml",
			errRegex: `unknown format: "xml"`,
		},
		"unknown type": {
			data: test.TrimJSON(`{
				"type": "unknown",
				"url": "https://example.com"
			}`),
			format: "json",
			errRegex: test.TrimYAML(`
			^failed to unmarshal latestver.Lookup:
				type: "unknown" <invalid>.*$`),
		},
		"invalid JSON": {
			data: test.TrimYAML(`{
				"type": "github",
				"url": release-argus/Argus
			}`),
			format: "json",
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					invalid character.*$`),
		},
		"invalid YAML": {
			data: test.TrimYAML(`
				type: github
				url: release-argus/Argus
				invalid
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					line \d: could not find .*$`),
		},
		"invalid GitHub": {
			data: test.TrimYAML(`
				type: github
				url:
					repo: release-argus
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
					failed to unmarshal github.Lookup:
						line \d: cannot unmarshal .*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN unmarshal is called.
			got, err := unmarshal([]byte(tc.data), tc.format)

			// THEN any error is as expected.
			if err != nil {
				if !util.RegexCheck(tc.errRegex, err.Error()) {
					t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.errRegex, err)
				}
				return
			}

			// AND the correct type is returned.
			if got.GetType() != tc.wantType {
				t.Errorf("%s\nType mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantType, got.GetType())
			}
		})
	}
}
