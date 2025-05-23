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

package deployedver

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/deployed_version/types/web"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestNew(t *testing.T) {
	// GIVEN a set of args to create a Lookup.
	type args struct {
		lType, wantType string
		configFormat    string
		configData      any
	}
	tests := map[string]struct {
		args     args
		wantYAML string
		errRegex string
	}{
		"string, YAML - url": {
			args: args{
				lType:        "url",
				configFormat: "yaml",
				configData: test.TrimYAML(`
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
					body: body_here
					json: value.version
					regex: v([0-9.]+)
					regex_template: $1
				`),
			},
			wantYAML: test.TrimYAML(`
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
				body: body_here
				json: value.version
				regex: v([0-9.]+)
				regex_template: $1
			`),
			errRegex: `^$`,
		},
		"yaml.Node - url": {
			args: args{
				lType:        "url",
				configFormat: "something?",
				configData: test.IgnoreError(t, func() (*yaml.Node, error) {
					return test.YAMLToNode(t,
						test.TrimYAML(`
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
							body: body_here
							json: value.version
							regex: v([0-9.]+)
							regex_template: $1
					`))
				}),
			},
			wantYAML: test.TrimYAML(`
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
				body: body_here
				json: value.version
				regex: v([0-9.]+)
				regex_template: $1
			`),
			errRegex: `^$`,
		},
		"string. JSON - web": {
			args: args{
				lType:        "web",
				wantType:     "url",
				configFormat: "json",
				configData: test.TrimJSON(`{
					"method": "GET",
					"url": "https://example.com",
					"allow_invalid_certs": false,
					"basic_auth": {
						"username": "user",
						"password": "pass"
					},
					"headers": [
						{"key": "X-Header", "value": "val"},
						{"key": "X-Another", "value": "val2"}
					],
					"body": "body_here",
					"json": "value.version",
					"regex": "v([0-9.]+)",
					"regex_template": "$1"
				}`),
			},
			wantYAML: test.TrimYAML(`
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
				body: body_here
				json: value.version
				regex: v([0-9.]+)
				regex_template: $1
			`),
			errRegex: `^$`,
		},
		"json.RawMessage - web": {
			args: args{
				lType:        "web",
				wantType:     "url",
				configFormat: "json",
				configData: json.RawMessage(
					test.TrimJSON(`{
						"method": "GET",
						"url": "https://example.com",
						"allow_invalid_certs": false,
						"basic_auth": {
							"username": "user",
							"password": "pass"
						},
						"headers": [
							{"key": "X-Header", "value": "val"},
							{"key": "X-Another", "value": "val2"}
						],
						"body": "body_here",
						"json": "value.version",
						"regex": "v([0-9.]+)",
						"regex_template": "$1"
					}`)),
			},
			wantYAML: test.TrimYAML(`
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
				body: body_here
				json: value.version
				regex: v([0-9.]+)
				regex_template: $1
			`),
			errRegex: `^$`,
		},
		"invalid type": {
			args: args{
				lType:        "invalid",
				configFormat: "yaml",
				configData:   "invalid_yaml",
			},
			errRegex: test.TrimYAML(`
				^failed to unmarshal deployedver\.Lookup:
					type: "[^"]+" <invalid>.*$`),
		},
		"empty type": {
			args: args{
				lType:        "",
				configFormat: "yaml",
				configData:   "invalid_yaml",
			},
			errRegex: test.TrimYAML(`
				^failed to unmarshal deployedver\.Lookup:
					type: <required>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := &opt.Options{}
			status := &status.Status{}
			defaults := &base.Defaults{}
			hardDefaults := &base.Defaults{}

			// WHEN New is called with the args.
			got, err := New(
				tc.args.lType,
				tc.args.configFormat,
				tc.args.configData,
				options,
				status,
				defaults,
				hardDefaults,
			)

			// THEN any error is as expected.
			if err != nil {
				if !util.RegexCheck(tc.errRegex, err.Error()) {
					t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
						packageName, tc.errRegex, err)
				}
				return
			}
			// THEN the correct type is returned.
			tc.args.lType = util.FirstNonDefault(tc.args.wantType, tc.args.lType)
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
			if got.GetDefaults() != defaults {
				t.Errorf("%s\nDefaults mismatch\nwant: %p\ngot:  %p",
					packageName, defaults, got.GetDefaults())
			}
			if got.GetHardDefaults() != hardDefaults {
				t.Errorf("%s\nHardDefaults mismatch\nwant: %p\ngot:  %p",
					packageName, hardDefaults, got.GetHardDefaults())
			}
		})
	}
}

func TestCopy(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		lookup Lookup
	}{
		"nil lookup": {
			lookup: nil,
		},
		"url": {
			lookup: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"url",
					"yaml", test.TrimYAML(`
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
						body: body_here
						json: value.version
						regex: v([0-9.]+)
						regex_template: $1`),
					opt.New(
						test.BoolPtr(true),
						"9m",
						test.BoolPtr(true),
						nil, nil),
					&status.Status{
						Dashboard: &dashboard.Options{}},
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
		"equal": {
			a: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"url",
					"yaml", test.TrimYAML(`
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
						body: body_here
						json: value.version
						regex: v([0-9.]+)
						regex_template: $1`),
					nil,
					nil,
					nil, nil)
			}),
			b: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"url",
					"yaml", test.TrimYAML(`
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
						body: body_here
						json: value.version
						regex: v([0-9.]+)
						regex_template: $1`),
					nil,
					nil,
					nil, nil)
			}),
			want: true,
		},
		"not equal": {
			a: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"url",
					"yaml", test.TrimYAML(`
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
						body: body_here
						json: value.version
						regex: v([0-9.]+)
						regex_template: $1`),
					nil,
					nil,
					nil, nil)
			}),
			b: test.IgnoreError(t, func() (Lookup, error) {
				return New(
					"url",
					"yaml", test.TrimYAML(`
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
								value: val3
						body: body_here
						json: value.version
						regex: v([0-9.]+)
						regex_template: $1`),
					nil,
					nil,
					nil, nil)
			}),
			want: false,
		},
		"not equal with nil": {
			a: nil,
			b: &web.Lookup{
				URL: "https://example.com"},
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
	// GIVEN a JSON string.
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
		"Minimal JSON": {
			jsonStr:  test.TrimJSON(`{}`),
			errRegex: `^$`,
			wantJSON: test.StringPtr(test.TrimJSON(`{
				"type":"url"
			}`)),
		},
		"Valid - URL": {
			jsonStr: test.TrimJSON(`{
				"type":"url",
				"url":"https://example.com",
				"allow_invalid_certs":true
			}`),
			errRegex: `^$`,
		},
		"Invalid - URL": {
			jsonStr: test.TrimJSON(`{
				"type":"url",
				"url":"https://example.com",
				"allow_invalid_certs":"true"
			}`),
			errRegex: `failed to unmarshal web.Lookup`,
		},
		"Unknown type": {
			jsonStr: test.TrimJSON(`{
				"type":"foo",
				"url":"https://example.com",
				"allow_invalid_certs":true
			}`),
			errRegex: `failed to unmarshal deployedver.Lookup`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.wantJSON == nil {
				tc.wantJSON = &tc.jsonStr
			}

			// WHEN UnmarshalJSON is called on it.
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
	// GIVEN a YAML string.
	tests := map[string]struct {
		yamlStr  string
		errRegex string
		wantYAML *string
	}{
		"Empty": {
			yamlStr: "",
			wantYAML: test.StringPtr(`
				type: url
			`),
			// TODO: When url default removed, this will error.
			// errRegex: `failed to unmarshal`,
			// wantYAML: test.StringPtr(""),
		},
		"Invalid formatting": {
			yamlStr:  "{ invalid",
			errRegex: `did not find expected`,
		},
		"Valid - URL": {
			yamlStr: test.TrimYAML(`
				type: url
				url: https://example.com
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		"Invalid - URL": {
			yamlStr: test.TrimYAML(`
				type: url
				url: https://example.com
				allow_invalid_certs: "true"
			`),
			errRegex: `failed to unmarshal web.Lookup`,
		},
		"Unknown type": {
			yamlStr: test.TrimJSON(`{
				type: foo,
				url: https://example.com",
				allow_invalid_certs: true
			}`),
			errRegex: `failed to unmarshal deployedver.Lookup`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.wantYAML == nil {
				tc.wantYAML = &tc.yamlStr
			}

			// WHEN UnmarshalYAML is called on it.
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
	// GIVEN an input string in a specified format.
	tests := map[string]struct {
		format, data string
		wantType     string
		errRegex     string
	}{
		"valid JSON - URL": {
			data: test.TrimJSON(`{
				"type": "url",
				"url": "https://example.com"
			}`),
			format:   "json",
			wantType: "url",
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
			data:     `{"type": "url"}`,
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
			^failed to unmarshal deployedver.Lookup:
				type: "unknown" <invalid>.*$`),
		},
		"invalid JSON": {
			data: test.TrimYAML(`{
				"type": "url",
				"url": https://example.com
			}`),
			format: "json",
			errRegex: test.TrimYAML(`
				^failed to unmarshal deployedver.Lookup:
					invalid character.*$`),
		},
		"invalid YAML vars": {
			data: test.TrimYAML(`
				type: url
				url: [https://example.com]
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				^failed to unmarshal web.Lookup:
					line \d+: .*$`),
		},
		"invalid YAML format": {
			data: test.TrimYAML(`
				type: url
				url: https://example.com
				invalid
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				^failed to unmarshal deployedver.Lookup:
					line \d+: .*$`),
		},
		"invalid URL": {
			data: test.TrimYAML(`
				type: url
				url:
					invalid: true
			`),
			format: "yaml",
			errRegex: test.TrimYAML(`
				failed to unmarshal web.Lookup:
					line \d: cannot unmarshal .*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN unmarshal is called on the input.
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
				t.Errorf("%s\ntype mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantType, got.GetType())
			}
		})
	}
}
