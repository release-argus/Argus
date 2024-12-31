// Copyright [2024] [Argus]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use 10s file except in compliance with the License.
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

	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestNew(t *testing.T) {
	type args struct {
		configFormat, configData string
		options                  *opt.Options
		status                   *status.Status
		defaults, hardDefaults   *Defaults
	}
	type wants struct {
		yaml     string
		errRegex string
	}
	// GIVEN a set of args
	tests := map[string]struct {
		args  args
		wants wants
	}{
		"valid yaml": {
			args: args{
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
					regex_template: $1`),
				options:      nil,
				status:       nil,
				defaults:     nil,
				hardDefaults: nil,
			},
			wants: wants{
				errRegex: `^$`,
				yaml: test.TrimYAML(`
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
				`)},
		},
		"valid json": {
			args: args{
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
				options:      nil,
				status:       nil,
				defaults:     nil,
				hardDefaults: nil,
			},
			wants: wants{
				errRegex: `^$`,
				yaml: test.TrimYAML(`
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
				`)},
		},
		"invalid format": {
			args: args{
				configFormat: "invalid",
				configData: `
<latest_version>Argus</latest_version>
<url>release-argus/argus</url>`,
				options:      nil,
				status:       nil,
				defaults:     nil,
				hardDefaults: nil,
			},
			wants: wants{
				errRegex: `^failed to unmarshal deployedver.Lookup`},
		},
		"invalid yaml": {
			args: args{
				configFormat: "yaml",
				configData:   "invalid_yaml",
				options:      nil,
				status:       nil,
				defaults:     nil,
				hardDefaults: nil,
			},
			wants: wants{
				errRegex: `^failed to unmarshal deployedver.Lookup`},
		},
		"invalid json": {
			args: args{
				configFormat: "json",
				configData:   "invalid_json",
				options:      nil,
				status:       nil,
				defaults:     nil,
				hardDefaults: nil,
			},
			wants: wants{
				errRegex: `^failed to unmarshal deployedver.Lookup`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN New is called with the test case parameters
			got, err := New(
				tc.args.configFormat, tc.args.configData,
				tc.args.options,
				tc.args.status,
				tc.args.defaults, tc.args.hardDefaults)

			// THEN the error is as expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.wants.errRegex, e) {
				t.Errorf("deployedver.New() error mismatch\n%q\ngot:\n%q\n",
					tc.wants.errRegex, e)
			}
			// AND the Lookup is returned as expected
			gotYAML := got.String("")
			if gotYAML != tc.wants.yaml {
				t.Errorf("deployedver.New() mismatch\nwant:\n%q\ngot:\n%q\n",
					tc.wants.yaml, gotYAML)
			}
		})
	}
}

func TestCopy(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		lookup *Lookup
	}{
		"nil lookup": {
			lookup: nil,
		},
		"empty lookup": {
			lookup: &Lookup{},
		},
		"filled lookup": {
			lookup: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
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
					nil,
					&Defaults{}, &Defaults{})
			}),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hadStr := tc.lookup.String("")

			// WHEN Copy is called with the test case parameters
			got := Copy(tc.lookup)

			// THEN the result is as expected
			gotStr := got.String("")
			if gotStr != hadStr {
				t.Errorf("deployedver.Copy() mismatch\n%q\ngot:\n%q\n",
					hadStr, gotStr)
			}
			if tc.lookup != nil {
				// AND the Lookup is not the same instance
				if got == tc.lookup {
					t.Errorf("deployedver.Copy() got same instance")
				}
				// AND the options are not the same instance
				if got.Options != nil && got.Options == tc.lookup.Options {
					t.Errorf("deployedver.Copy() got same Options instance")
				}
				// AND the defaults are the same instance
				if got.Defaults != tc.lookup.Defaults {
					t.Errorf("deployedver.Copy() got different Defaults instance")
				}
				// AND the hardDefaults are the same instance
				if got.HardDefaults != tc.lookup.HardDefaults {
					t.Errorf("deployedver.Copy() got different HardDefaults instance")
				}
			}
		})
	}
}

func TestLookup_String(t *testing.T) {
	tests := map[string]struct {
		lookup *string
		want   string
	}{
		"nil": {
			lookup: nil,
			want:   "",
		},
		"empty": {
			lookup: test.StringPtr(""),
			want:   "{}",
		},
		"filled": {
			lookup: test.StringPtr(test.TrimYAML(`
				method: GET
				url: https://example.com
				allow_invalid_certs: false
				options:
					interval: 9m
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
				regex_template: $1`)),
			want: test.TrimYAML(`
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
		},
		"quotes otherwise invalid yaml strings": {
			lookup: test.StringPtr(test.TrimYAML(`
				basic_auth:
					username: '>123'
					password: '{pass}'`)),
			want: test.TrimYAML(`
				basic_auth:
					username: '>123'
					password: '{pass}'`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var lookup *Lookup
			if tc.lookup != nil {
				lookup, _ = New("yaml", *tc.lookup, nil, nil, nil, nil)
			}

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Lookup is stringified with String
				got := lookup.String(prefix)

				// THEN the result is as expected
				if got != want {
					t.Errorf("(prefix=%q) got:\n%q\nwant:\n%q",
						prefix, got, want)
				}
			}
		})
	}
}

func TestLookup_IsEqual(t *testing.T) {
	// GIVEN two Lookups
	tests := map[string]struct {
		a, b *Lookup
		want bool
	}{
		"empty": {
			a:    &Lookup{},
			b:    &Lookup{},
			want: true,
		},
		"defaults ignored": {
			a: &Lookup{
				Defaults: NewDefaults(
					test.BoolPtr(false))},
			b:    &Lookup{},
			want: true,
		},
		"hard_defaults ignored": {
			a: &Lookup{
				HardDefaults: NewDefaults(
					test.BoolPtr(false))},
			b:    &Lookup{},
			want: true,
		},
		"equal": {
			a: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
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
			b: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
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
			a: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
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
			b: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
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
			b: &Lookup{
				URL: "https://example.com"},
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the two Lookups are compared
			got := tc.a.IsEqual(tc.b)

			// THEN the result is as expected
			if got != tc.want {
				t.Errorf("got %t, want %t", got, tc.want)
			}
		})
	}
}
