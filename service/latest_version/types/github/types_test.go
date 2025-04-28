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

// Package github provides a github-based lookup type.
package github

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestLookup_String(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		lookup *Lookup
		want   string
	}{
		"empty": {
			lookup: &Lookup{},
			want:   "{}\n",
		},
		"filled": {
			lookup: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
					"yaml", test.TrimYAML(`
							url: release-argus/Argus
							require:
								regex_content: foo.tar.gz
							access_token: token
							use_prerelease: true
							url_commands:
								- type: regex
									regex: v([0-9.]+)
						`),
					opt.New(
						nil, "1h2m3s", nil,
						nil, nil),
					nil,
					&base.Defaults{
						AccessToken: "foo"},
					&base.Defaults{
						AccessToken: "foo"})
			}),
			want: test.TrimYAML(`
				type: github
				url: release-argus/Argus
				url_commands:
					- type: regex
						regex: v([0-9.]+)
				require:
					regex_content: foo.tar.gz
				access_token: token
				use_prerelease: true
				`),
		},
		"quotes otherwise invalid YAML strings": {
			lookup: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
					"yaml", test.TrimYAML(`
						access_token: ">123"
						url_commands:
							- type: regex
								regex: '{2}([0-9.]+)'
					`),
					nil,
					nil,
					nil, nil)
			}),
			want: test.TrimYAML(`
				type: github
				url_commands:
					- type: regex
						regex: '{2}([0-9.]+)'
				access_token: '>123'
			`),
		},
		"gives defaults": {
			lookup: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
					"yaml", test.TrimYAML(`
						url: release-argus/Argus
					`),
					nil,
					nil,
					&base.Defaults{
						AccessToken: "foo"},
					&base.Defaults{
						AccessToken: "foo"})
			}),
			want: test.TrimYAML(`
				type: github
				url: release-argus/Argus
			`),
		},
		"gives Require.Docker.Defaults": {
			lookup: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
					"yaml", test.TrimYAML(`
						url: release-argus/Argus
						require:
							docker:
								type: ghcr
								image: release-argus/argus
								tag: "{{ version }}"
				`),
					nil,
					nil,
					&base.Defaults{
						AccessToken: "foo",
						Require: filter.RequireDefaults{
							Docker: *filter.NewDockerCheckDefaults(
								"ghcr", "", "", "", "", nil)}},
					&base.Defaults{
						AccessToken: "foo"})
			}),
			want: test.TrimYAML(`
				type: github
				url: release-argus/Argus
				require:
					docker:
						type: ghcr
						image: release-argus/argus
						tag: '{{ version }}'
			`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Lookup is stringified with String.
			got := tc.lookup.String(tc.lookup, "")

			// THEN the result is as expected.
			tc.want = test.TrimYAML(tc.want)
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		format string
		data   string
	}

	// GIVEN a string to unmarshal into a Lookup.
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
					access_token: token
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
				access_token: token
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
				data: (`
					allow_invalid_certs true
			`)},
			errRegex: test.TrimYAML(`
				^failed to unmarshal github.Lookup:
					line \d: .*$`),
		},
		"empty YAML": {
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
				access_token: abc
			`),
		},
		"Invalid format": {
			args: args{
				format: "xml",
				data: `
					<url>https://example.com</url>`},
			errRegex: test.TrimYAML(`
					^failed to unmarshal github.Lookup:
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
			// AND the lookup is created as expected.
			wantStr := "type: github\n" + test.TrimYAML(util.ValueOrValue(tc.wantStr, tc.args.data))
			gotStr := lookup.String(lookup, "")
			if gotStr != wantStr {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, wantStr, gotStr)
			}
			// AND the defaults are set as expected.
			if lookup.Defaults != defaults {
				t.Errorf("%s\nDefaults not set\nwant: %v\ngot:  %v",
					packageName, lookup.Defaults, defaults)
			}
			// AND the hard defaults are set as expected.
			if lookup.HardDefaults != hardDefaults {
				t.Errorf("%s\nHardDefaults not set\nwant: %v\ngot:  %v",
					packageName, lookup.HardDefaults, hardDefaults)
			}
			// AND the status is set as expected.
			if lookup.Status != status {
				t.Errorf("%s\nStatus not set\nwant: %v\ngot:  %v",
					packageName, lookup.Status, status)
			}
			// AND the options are set as expected.
			if lookup.Options != &options {
				t.Errorf("%s\nOptions not set\nwant: %v\ngot:  %v",
					packageName, lookup.Options, &options)
			}
			// AND the require is given the correct defaults.
			if lookup.Require != nil && lookup.Require.Docker != nil {
				if lookup.Require.Docker.Defaults != &defaults.Require.Docker {
					t.Errorf("%s\nRequire.Docker.Defaults not set\nwant: %v\ngot:  %v",
						packageName, lookup.Require.Docker.Defaults, defaults.Require.Docker)
				}
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	// GIVEN a JSON string to unmarshal.
	tests := map[string]struct {
		data     string
		want     string
		errRegex string
	}{
		"Valid JSON": {
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
				type: github
				url: https://example.com
				url_commands:
					- type: regex
						regex: foo
				require:
					regex_version: v.+
				use_prerelease: true
			`),
			errRegex: `^$`,
		},
		"Invalid JSON vars": {
			data:     `{"url": ["https://example.com"]}`,
			errRegex: `^json: cannot unmarshal array .* field (\.Lookup)?\.url of .*$`,
		},
		"Invalid JSON formatting": {
			data:     `{"url": "https://example.com"`,
			errRegex: `^unexpected end of JSON input$`,
		},
		"Empty JSON": {
			data:     `{}`,
			want:     "type: github\n",
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
			if tc.errRegex != `^$` {
				return
			}
			// AND the lookup is created if expected.
			if gotStr != tc.want {
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
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
				use_prerelease: true
				url_commands:
					- type: regex
						regex: foo
				require:
					regex_version: v.+
			`),
			want: test.TrimYAML(`
				type: github
				url: https://example.com
				url_commands:
					- type: regex
						regex: foo
				require:
					regex_version: v.+
				use_prerelease: true
			`),
			errRegex: `^$`,
		},
		"invalid YAML": {
			data: test.TrimYAML(`
				url: [https://example.com]
			`),
			errRegex: test.TrimYAML(`
				^yaml: unmarshal errors:
					line 1: cannot unmarshal.*$`),
		},
		"empty YAML": {
			data:     `{}`,
			want:     "type: github\n",
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
				t.Errorf("%s\nstringified mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want, gotStr)
			}
		})
	}
}
