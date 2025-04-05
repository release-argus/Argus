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

package web

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestNew(t *testing.T) {
	type args struct {
		format string
		data   string
	}

	// GIVEN a YAML string.
	tests := map[string]struct {
		args     args
		wantStr  string
		errRegex string
	}{
		"valid YAML": {
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
						url: https://example.com
						url_commands:
							- type: regex
								regex: foo
						require:
							regex_version: v.+
						allow_invalid_certs: true
					`)},
			errRegex: `^$`,
		},
		"valid YAML with unmapped vars": {
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
						url: https://example.com
						require:
							regex_version: v.+
						allow_invalid_certs: true
						access_token: token
						foo: bar
						url_commands:
							- type: regex
								regex: foo
					`)},
			wantStr: test.TrimYAML(`
							url: https://example.com
							url_commands:
								- type: regex
									regex: foo
							require:
								regex_version: v.+
							allow_invalid_certs: true
						`),
			errRegex: `^$`,
		},
		"valid YAML with Require.Docker": {
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
						require:
							docker:
								type: ghcr
								image: something
								tag: '{{ version }}'
					`)},
			errRegex: `^$`,
		},
		"invalid YAML": {
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					allow_invalid_certs true
				`)},
			errRegex: test.TrimYAML(`
				^failed to unmarshal web.Lookup:
					line \d: cannot unmarshal.*$`),
		},
		"nil YAML": {
			args: args{
				format: "yaml",
				data:   ""},
			errRegex: `^$`,
		},
		"JSON": {
			args: args{
				format: "json",
				data: test.TrimJSON(`{
					"url": "https://example.com",
					"allow_invalid_certs": true,
					"url_commands": [
						{ "type": "regex", "regex": "foo" },
						{ "type": "replace", "old": "foo", "new": "bar" },
						{ "type": "split", "text": "foo" }
					],
					"access_token": "abc",
					"require": { "regex_version": "v.+" }
				}`),
			},
			errRegex: `^$`,
			wantStr: test.TrimYAML(`
				url: https://example.com
				url_commands:
					- type: regex
						regex: foo
					- type: replace
						new: bar
						old: foo
					- type: split
						text: foo
				require:
					regex_version: v.+
				allow_invalid_certs: true
			`),
		},
		"invalid format": {
			args: args{
				format: "xml",
				data: `
					<url>https://example.com</url>`},
			errRegex: test.TrimYAML(`
				^failed to unmarshal web.Lookup:
					unsupported configFormat: xml$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := opt.Options{}
			status := &status.Status{}
			defaults := &base.Defaults{}
			hardDefaults := &base.Defaults{}
			hardDefaults.Default()

			// WHEN New is called with it.
			lookup, err := New(
				tc.args.format, tc.args.data,
				&options,
				status,
				defaults, hardDefaults)

			// THEN any error is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, err)
			}
			if err != nil {
				return
			}
			// AND the Lookup is created as expected.
			wantStr := "type: url\n" + util.ValueOrValue(tc.wantStr, tc.args.data)
			gotStr := lookup.String(lookup, "")
			if gotStr != wantStr {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, wantStr, gotStr)
			}
			// AND the Defaults are set as expected.
			if lookup.Defaults != defaults {
				t.Errorf("%s\nDefaults not set\nwant: %v\ngot:  %v",
					packageName, lookup.Defaults, defaults)
			}
			// AND the HardDefaults are set as expected.
			if lookup.HardDefaults != hardDefaults {
				t.Errorf("%s\nHardDefaults not set\nwant: %v\ngot:  %v",
					packageName, lookup.HardDefaults, hardDefaults)
			}
			// AND the Status is set as expected.
			if lookup.Status != status {
				t.Errorf("%s\nStatus not set\nwant: %v\ngot:  %v",
					packageName, lookup.Status, status)
			}
			// AND the Options are set as expected.
			if lookup.Options != &options {
				t.Errorf("%s\nOptions not set\nwant: %v\ngot:  %v",
					packageName, lookup.Options, &options)
			}
			// AND the Require is given the correct defaults.
			if lookup.Require != nil && lookup.Require.Docker != nil {
				if lookup.Require.Docker.Defaults != &defaults.Require.Docker {
					t.Errorf("%s\nRequire.Docker.Defaults not set\nwant: %v\ngot:  %v",
						packageName, lookup.Require.Docker.Defaults, defaults.Require.Docker)
				}
			}
		})
	}
}

func TestLookup_UnmarshalJSON(t *testing.T) {
	// GIVEN a JSON string to unmarshal.
	tests := map[string]struct {
		data     string
		want     string
		errRegex string
	}{
		"valid JSON": {
			data: test.TrimJSON(`{
				"url": "https://example.com",
				"allow_invalid_certs": true,
				"use_prerelease": true,
				"url_commands": [
					{ "type": "regex", "regex": "foo" }
				],
				"require": { "regex_version": "v.+" }
			}`),
			want: test.TrimYAML(`
				type: url
				url: https://example.com
				url_commands:
					- type: regex
						regex: foo
				require:
					regex_version: v.+
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		"invalid JSON vars": {
			data:     `{"url": ["https://example.com"]}`,
			errRegex: `json: cannot unmarshal array into Go struct field (\.Lookup)?\.url of type string$`,
		},
		"invalid JSON format": {
			data:     `{"url": "https://example.com`,
			errRegex: `^unexpected end of JSON input$`,
		},
		"empty JSON": {
			data:     `{}`,
			want:     "type: url\n",
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var lookup Lookup

			// WHEN UnmarshalJSON is called with it.
			err := lookup.UnmarshalJSON([]byte(tc.data))

			// THEN any error is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, err)
				return
			}
			gotStr := lookup.String(&lookup, "")
			// AND the Lookup isn't created if it errored.
			if err != nil {
				if gotStr != "{}\n" {
					t.Errorf("%s\nvalue mismatch after non-nil error\nwant: nil\ngot:  value=%q",
						packageName, gotStr)
				}
				return
			}
			// AND the lookup is created if expected.
			if gotStr != tc.want {
				t.Errorf("%s\nvalue mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want, gotStr)
			}
		})
	}
}

func TestLookup_UnmarshalYAML(t *testing.T) {
	// GIVEN a YAML string to unmarshal.
	tests := map[string]struct {
		data     string
		want     string
		errRegex string
	}{
		"valid YAML": {
			data: test.TrimYAML(`
				url: https://example.com
				allow_invalid_certs: true
				url_commands:
					- type: regex
						regex: foo
				require:
					regex_version: v.+
			`),
			want: test.TrimYAML(`
				type: url
				url: https://example.com
				url_commands:
					- type: regex
						regex: foo
				require:
					regex_version: v.+
				allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		"invalid YAML": {
			data: test.TrimYAML(`
				url: [https://example.com]
			`),
			errRegex: `line 1: cannot unmarshal .* into string$`,
		},
		"empty YAML": {
			data:     `{}`,
			want:     "type: url\n",
			errRegex: `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Convert the YAML string to a yaml.Node.
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tc.data), &node); err != nil {
				t.Fatalf("%s\nfailed to unmarshal yaml: %v",
					packageName, err)
			}
			var lookup Lookup

			// WHEN UnmarshalYAML is called with it.
			err := lookup.UnmarshalYAML(&node)

			// THEN any error is expected.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, err)
				return
			}
			gotStr := lookup.String(&lookup, "")
			// AND the Lookup isn't created if it errored.
			if err != nil {
				if gotStr != "{}\n" {
					t.Errorf("%s\nvalue mismatch after non-nil error\nwant: nil\ngot:  value=%q",
						packageName, gotStr)
				}
				return
			}
			// AND the lookup is created if expected.
			if gotStr != tc.want {
				t.Errorf("%s\nvalue mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want, gotStr)
			}
		})
	}
}
