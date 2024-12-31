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

package latestver

import (
	"testing"

	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestNew(t *testing.T) {
	// GIVEN a set of args
	type args struct {
		lType                  string
		overrides              interface{}
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
			errRegex: `^type: "foo" <invalid> .*$`,
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
				unsupported configData type: int`),
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
				unsupported configData type: int`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if overrides, ok := tc.args.overrides.(string); ok {
				tc.args.overrides = test.TrimYAML(overrides)
			}

			// WHEN New is called with the args
			got, err := New(
				tc.args.lType,
				"yaml", tc.args.overrides,
				opt.New(
					nil, "",
					tc.args.semanticVersioning,
					nil, nil),
				nil,
				tc.args.defaults, tc.args.hardDefaults)

			// THEN any error is as expected
			if err != nil {
				if !util.RegexCheck(tc.errRegex, err.Error()) {
					t.Errorf("New() error mismatch\n%q\ngot:  %q",
						tc.errRegex, err)
				}
				return
			}
			// THEN the correct type is returned
			if getType(got) != tc.args.lType {
				t.Fatalf("New() Type mismatch\nwant: %q\ngot:  %T",
					tc.args.lType, got)
			}

			// AND the variables are set correctly
			gotYAML := util.ToYAMLString(got, "")
			if gotYAML != tc.wantYAML {
				t.Errorf("YAML mismatch\nwant: %q\ngot:  %q",
					tc.wantYAML, gotYAML)
			}
			if got.GetDefaults() != tc.args.defaults {
				t.Errorf("defaults mismatch\nwant: %p\ngot:  %p",
					tc.args.defaults, got.GetDefaults())
			}
			if got.GetHardDefaults() != tc.args.hardDefaults {
				t.Errorf("hardDefaults mismatch\nwant: %p\ngot:  %p",
					tc.args.hardDefaults, got.GetHardDefaults())
			}
		})
	}
}

func TestCopy(t *testing.T) {
	// GIVEN a Lookup
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
					&status.Status{},
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
					&status.Status{},
					&base.Defaults{}, &base.Defaults{})
			}),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Copy is called
			got := Copy(tc.lookup)

			// THEN the variables are copied over
			gotYAML := util.ToYAMLString(got, "")
			wantYAML := util.ToYAMLString(tc.lookup, "")
			if gotYAML != wantYAML {
				t.Errorf("YAML mismatch\nwant: %q\ngot:  %q",
					wantYAML, gotYAML)
			}
			if tc.lookup == nil {
				return // No further checks
			}

			if got.GetOptions() == tc.lookup.GetOptions() {
				t.Error("options shouldn't point to the same memory address")
			} else if got.GetOptions().String() != tc.lookup.GetOptions().String() {
				t.Errorf("options mismatch\nwant: %q\ngot:  %q",
					tc.lookup.GetOptions(), got.GetOptions())
			}

			if got.GetDefaults() != tc.lookup.GetDefaults() {
				t.Errorf("defaults mismatch\nwant: %v\ngot:  %v",
					tc.lookup.GetDefaults(), got.GetDefaults())
			}

			if got.GetHardDefaults() != tc.lookup.GetHardDefaults() {
				t.Errorf("hardDefaults mismatch\nwant: %v\ngot:  %v",
					tc.lookup.GetHardDefaults(), got.GetHardDefaults())
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

	// GIVEN a Lookup
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
			errRegex: `^type: \"foo\" <invalid>.*$`,
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

			// WHEN ChangeType is called
			gotLookup, err := ChangeType(
				tc.args.newType,
				tc.args.lookup,
				tc.args.overrides)

			// THEN the correct type is returned
			if gotLookup == nil {
				// Unknown type
				if tc.errRegex != "" {
					if !util.RegexCheck(tc.errRegex, err.Error()) {
						t.Errorf("error mismatch\nwant: %q\ngot:  %q",
							tc.errRegex, err)
					}
					// return // Got expected error
				} else {
					t.Errorf("ChangeType() gave nil latestver.Lookup, expected type %q\nerr=%c",
						tc.args.newType, err)
				}
				return
			}
			gotType := gotLookup.GetType()
			if tc.args.newType != gotType {
				t.Fatalf("Type mismatch\nwant: %q\ngot:  %T",
					tc.args.newType, gotType)
			}

			// AND the variables are copied over
			gotYAML := util.ToYAMLString(gotLookup, "")
			tc.wantYAML = test.TrimYAML(tc.wantYAML)
			if gotYAML != tc.wantYAML {
				t.Errorf("YAML mismatch\nwant: %q\ngot:  %q",
					tc.wantYAML, gotYAML)
			}

			if gotLookup.GetOptions() != tc.args.lookup.GetOptions() &&
				tc.args.lookup.GetOptions() != nil || gotLookup.GetOptions() == nil {
				t.Errorf("options mismatch\nwant: %v\ngot:  %v",
					tc.args.lookup.GetOptions(), gotLookup.GetOptions())
			}

			if gotLookup.GetDefaults() != tc.args.lookup.GetDefaults() {
				t.Errorf("defaults mismatch\nwant: %v\ngot:  %v",
					tc.args.lookup.GetDefaults(), gotLookup.GetDefaults())
			}

			if gotLookup.GetHardDefaults() != tc.args.lookup.GetHardDefaults() {
				t.Errorf("hardDefaults mismatch\nwant: %v\ngot:  %v",
					tc.args.lookup.GetHardDefaults(), gotLookup.GetHardDefaults())
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
		"Empty": {
			jsonStr:  "",
			errRegex: `unexpected end of JSON input`,
			wantJSON: test.StringPtr(""),
		},
		"Invalid formatting": {
			jsonStr:  "invalid",
			errRegex: `invalid character`,
		},
		"Valid - GitHub": {
			jsonStr: test.TrimJSON(`{
				"type":"github",
				"url":"release-argus/Argus",
				"access_token":"token"
			}`),
			errRegex: `^$`,
		},
		"Valid - URL": {
			jsonStr: test.TrimJSON(`{
				"type":"url",
				"url":"https://example.com",
				"allow_invalid_certs":true
			}`),
			errRegex: `^$`,
		},
		"Invalid - GitHub": {
			jsonStr: test.TrimJSON(`{"
				type":"github",
				"url":"release-argus/Argus",
				"access_token":1
			}`),
			errRegex: `failed to unmarshal github.Lookup`,
		},
		"Invalid - URL": {
			jsonStr: test.TrimJSON(`{
				"type":"url",
				"url":"https://example.com",
				"allow_invalid_certs":"true"
			}`),
			errRegex: `failed to unmarshal web.Lookup`,
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

			// WHEN unmarshal is called
			lookupJSON, errJSON := UnmarshalJSON([]byte(tc.jsonStr))

			// THEN any error is as expected
			eJSON := util.ErrorToString(errJSON)
			if !util.RegexCheck(tc.errRegex, eJSON) {
				t.Errorf("error mismatch on JSON unmarshal of latestver.Lookup:\n%q\ngot:\n%q",
					tc.errRegex, eJSON)
			}
			if tc.errRegex != "^$" {
				return
			}
			// AND the Lookup is unmarshalled as expected
			gotFromJSON := util.ToJSONString(lookupJSON)
			if *tc.wantJSON != gotFromJSON {
				t.Errorf("latestver.Lookup.String() mismatch on JSON unmarshal\n%q\ngot:\n%q",
					*tc.wantJSON, gotFromJSON)
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
			yamlStr:  "{ invalid",
			errRegex: `did not find expected`,
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
			errRegex: `failed to unmarshal github.Lookup`,
		},
		"Invalid - URL": {
			yamlStr: test.TrimYAML(`
				type: url
				url: https://example.com
				allow_invalid_certs: "true"
			`),
			errRegex: `failed to unmarshal web.Lookup`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.wantYAML == nil {
				tc.wantYAML = &tc.yamlStr
			}

			// WHEN unmarshal is called
			lookupYAML, errYAML := UnmarshalYAML([]byte(tc.yamlStr))

			// THEN any error is as expected
			eYAML := util.ErrorToString(errYAML)
			if !util.RegexCheck(tc.errRegex, eYAML) {
				t.Errorf("error mismatch on YAML unmarshal of latestver.Lookup:\n%q\ngot:  %q",
					tc.errRegex, eYAML)
			}
			if tc.errRegex != "^$" {
				return
			}
			// AND the Lookup is unmarshalled as expected
			gotFromYAML := lookupYAML.String(lookupYAML, "")
			if *tc.wantYAML != gotFromYAML {
				t.Errorf("latestver.Lookup.String() mismatch on YAML unmarshal\n%q\ngot:  %q",
					*tc.wantYAML, gotFromYAML)
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
		"Valid JSON - GitHub": {
			data: test.TrimJSON(`{
				"type": "github",
				"url": "release-argus/Argus"
			}`),
			format:   "json",
			wantType: "github",
		},
		"Valid JSON - URL": {
			data: test.TrimJSON(`{
				"type": "url",
				"url": "https://example.com"
			}`),
			format:   "json",
			wantType: "url",
		},
		"Valid YAML - GitHub": {
			data: test.TrimYAML(`
				type: github
				url: release-argus/Argus
			`),
			format:   "yaml",
			wantType: "github",
		},
		"Valid YAML - URL": {
			data: test.TrimYAML(`
				type: url
				url: https://example.com
			`),
			format:   "yaml",
			wantType: "url",
		},
		"Invalid format": {
			data:     `{"type": "github"}`,
			format:   "xml",
			errRegex: `unknown format: "xml"`,
		},
		"Unknown type": {
			data: test.TrimJSON(`{
				"type": "unknown",
				"url": "https://example.com"
			}`),
			format: "json",
			errRegex: test.TrimYAML(`
			^failed to unmarshal latestver.Lookup:
			type: "unknown" <invalid>.*$`),
		},
		"Invalid JSON": {
			data: test.TrimYAML(`{
				"type": "github",
				"url": release-argus/Argus
			}`),
			format: "json",
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
				invalid character.*$`),
		},
		"Invalid YAML": {
			data: test.TrimYAML(`
				type: github
				url: release-argus/Argus
				invalid
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				^failed to unmarshal latestver.Lookup:
				yaml: .*$`),
		},
		"Invalid GitHub": {
			data: test.TrimYAML(`
				type: github
				url:
					repo: release-argus
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				failed to unmarshal github.Lookup:
				yaml: .+
				.* cannot unmarshal .*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN unmarshal is called
			got, err := unmarshal([]byte(tc.data), tc.format)

			// THEN any error is as expected
			if err != nil {
				if !util.RegexCheck(tc.errRegex, err.Error()) {
					t.Errorf("unmarshal() error mismatch\nwant: %q\ngot:  %q",
						tc.errRegex, err)
				}
				return
			}

			// AND the correct type is returned
			if got.GetType() != tc.wantType {
				t.Errorf("unmarshal() type mismatch\nwant: %q\ngot:  %q",
					tc.wantType, got.GetType())
			}
		})
	}
}
