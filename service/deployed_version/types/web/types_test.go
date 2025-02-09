// Copyright [2025] [Argus]
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

package web

import (
	"strings"
	"testing"

	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestNew(t *testing.T) {
	type args struct {
		format string
		data   string
	}
	type wants struct {
		yaml     string
		errRegex string
	}
	// GIVEN a string to unmarshal, and a set of options/status/defaults.
	tests := map[string]struct {
		args  args
		wants wants
	}{
		"valid YAML": {
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
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
			wants: wants{
				errRegex: `^$`,
				yaml: test.TrimYAML(`
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
				`)},
		},
		"valid JSON": {
			args: args{
				format: "json",
				data: test.TrimJSON(`{
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
			wants: wants{
				errRegex: `^$`,
				yaml: test.TrimYAML(`
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
				`)},
		},
		"invalid format": {
			args: args{
				format: "invalid",
				data: `
					<latest_version>Argus</latest_version>
					<url>release-argus/argus</url>`,
			},
			wants: wants{
				errRegex: `^failed to unmarshal web.Lookup`},
		},
		"invalid YAML": {
			args: args{
				format: "yaml",
				data:   "invalid_yaml",
			},
			wants: wants{
				errRegex: `^failed to unmarshal web.Lookup`},
		},
		"invalid JSON": {
			args: args{
				format: "json",
				data:   "invalid_json",
			},
			wants: wants{
				errRegex: `^failed to unmarshal web.Lookup`},
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
			if !util.RegexCheck(tc.wants.errRegex, e) {
				if err == nil {
					t.Error("web.Lookup.New() expected error, got nil")
				} else {
					t.Errorf("web.Lookup.New() unexpected error: %v", err)
				}
			}
			if err != nil {
				return
			}
			// AND the lookup is created as expected.ValueOrValue(tc.wants.yaml, tc.args.data)).
			gotStr := lookup.String(lookup, "")
			if gotStr != tc.wants.yaml {
				t.Errorf("web.Lookup.String() mismatch\nwant: %q\ngot:  %q",
					tc.wants.yaml, gotStr)
			}
			// AND the defaults are set as expected.
			if lookup.Defaults != defaults {
				t.Errorf("web.Lookup.Defaults not set\nwant: %v\ngot:  %v",
					lookup.Defaults, defaults)
			}
			// AND the hard defaults are set as expected.
			if lookup.HardDefaults != hardDefaults {
				t.Errorf("web.Lookup.HardDefaults not set\nwant: %v\ngot:  %v",
					lookup.HardDefaults, hardDefaults)
			}
			// AND the status is set as expected.
			if lookup.Status != status {
				t.Errorf("web.Lookup.Status not set\nwant: %v\ngot:  %v",
					lookup.Status, status)
			}
			// AND the options are set as expected.
			if lookup.Options != &options {
				t.Errorf("web.Lookup.Options not set\nwant: %v\ngot:  %v",
					lookup.Options, &options)
			}
		})
	}
}

func TestString(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		lookup *Lookup
		want   string
	}{
		"empty": {
			lookup: &Lookup{},
			want:   "{}",
		},
		"filled": {
			lookup: test.IgnoreError(t, func() (*Lookup, error) {
				return New("yaml", test.TrimYAML(`
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
					regex_template: $1
				`),
					opt.New(
						nil, "1h2m3s", nil,
						nil, nil),
					nil,
					&base.Defaults{
						AllowInvalidCerts: test.BoolPtr(false)},
					&base.Defaults{
						AllowInvalidCerts: test.BoolPtr(false)})
			}),
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
				body: body_here
				json: value.version
				regex: v([0-9.]+)
				regex_template: $1`),
		},
		"quotes otherwise invalid YAML strings": {
			lookup: test.IgnoreError(t, func() (*Lookup, error) {
				return New(
					"yaml", test.TrimYAML(`
						basic_auth:
							username: '>123'
							password: '{pass}'
							`),
					nil,
					nil,
					nil, nil)
			}),
			want: test.TrimYAML(`
				type: url
				basic_auth:
					username: '>123'
					password: '{pass}'`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Lookup is stringified with String.
				got := tc.lookup.String(tc.lookup, prefix)

				// THEN the result is as expected.
				if got != want {
					t.Errorf("(prefix=%q) got:\n%q\nwant:\n%q",
						prefix, got, want)
				}
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	// GIVEN a Lookup to unmarshal from JSON.
	tests := map[string]struct {
		json    string
		wantStr string
		wantErr bool
	}{
		"empty": {
			json: "{}",
			wantStr: test.TrimJSON(`{
				"type": "url"
			}`),
			wantErr: false,
		},
		"filled": {
			json: test.TrimJSON(`{
				"type": "url",
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
			wantErr: false,
		},
		"invalid type - url": {
			json: test.TrimJSON(`{
				"url": ["https://example.com"]
			}`),
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{}

			// WHEN the JSON is unmarshalled.
			err := lookup.UnmarshalJSON([]byte(tc.json))

			// THEN it errors when expected.
			if (err != nil) != tc.wantErr {
				t.Errorf("Lookup.UnmarshalJSON() error = %v, wantErr %v",
					err, tc.wantErr)
			}
			if err == nil {
				gotStr := util.ToJSONString(lookup)
				wantStr := util.ValueOrValue(tc.wantStr, tc.json)
				if gotStr != wantStr {
					t.Errorf("Lookup.UnmarshalJSON()\ngot: \n%v\nwant:\n%v",
						gotStr, wantStr)
				}
			}
		})
	}
}

func TestUnmarshalYAML(t *testing.T) {
	// GIVEN a Lookup to unmarshal from YAML.
	tests := map[string]struct {
		yaml    string
		wantStr string
		wantErr bool
	}{
		"empty": {
			yaml: test.TrimYAML(``),
			wantStr: test.TrimYAML(`
				type: url
			`),
			wantErr: false,
		},
		"filled": {
			yaml: test.TrimYAML(`
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
			wantErr: false,
		},
		"invalid type - url": {
			yaml: test.TrimYAML(`
				url: ["https://example.com"]
			`),
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{}
			yamlNode, err := test.YAMLToNode(t, tc.yaml)
			if err != nil {
				t.Errorf("failed to convert YAML to yaml.Node: %v", err)
			}

			// WHEN the YAML is unmarshalled.
			err = lookup.UnmarshalYAML(yamlNode)

			// THEN it errors when expected.
			if (err != nil) != tc.wantErr {
				t.Errorf("Lookup.UnmarshalYAML() error = %v, wantErr %v",
					err, tc.wantErr)
			}
			if err == nil {
				gotStr := lookup.String(lookup, "")
				wantStr := util.ValueOrValue(tc.wantStr, tc.yaml)
				if gotStr != wantStr {
					t.Errorf("Lookup.UnmarshalYAML()\ngot: \n%v\nwant:\n%v",
						gotStr, wantStr)
				}
			}
		})
	}
}

func TestLookup_InheritSecrets(t *testing.T) {
	// GIVEN a Lookup with secrets, and another Lookup to inherit from.
	tests := map[string]struct {
		lookup     *Lookup
		other      *Lookup
		secretRefs *shared.DVSecretRef
		want       *Lookup
	}{
		"inherit BasicAuth password": {
			lookup: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: util.SecretValue,
				}},
			other: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "password",
				}},
			secretRefs: &shared.DVSecretRef{},
			want: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "password",
				}},
		},
		"inherit headers": {
			lookup: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: util.SecretValue},
				}},
			other: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: "secret"},
				}},
			secretRefs: &shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(0)},
				}},
			want: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: "secret"}}},
		},
		"no inheritance when no secrets": {
			lookup: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "password"},
				Headers: []Header{
					{Key: "X-Test", Value: "value"}},
			},
			other: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "other_password"},
				Headers: []Header{
					{Key: "X-Test", Value: "other_value"}},
			},
			secretRefs: &shared.DVSecretRef{},
			want: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "password"},
				Headers: []Header{
					{Key: "X-Test", Value: "value"}},
			},
		},
		"no inheritance when no matching headers": {
			lookup: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: util.SecretValue}},
			},
			other: &Lookup{
				Headers: []Header{
					{Key: "X-Other", Value: "secret"}},
			},
			secretRefs: &shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(1)}},
			},
			want: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: util.SecretValue}},
			},
		},
		"no inheritance when secretRefs out of bounds": {
			lookup: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: util.SecretValue},
					{Key: "X-Other", Value: util.SecretValue}},
			},
			other: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: "secret"}},
			},
			secretRefs: &shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.IntPtr(1)}},
			},
			want: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: util.SecretValue},
					{Key: "X-Other", Value: util.SecretValue}},
			},
		},
		"no inheritance when secretRef index nil": {
			lookup: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: util.SecretValue}},
			},
			other: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: "secret"}},
			},
			secretRefs: &shared.DVSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: nil}},
			},
			want: &Lookup{
				Headers: []Header{
					{Key: "X-Test", Value: util.SecretValue}},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN InheritSecrets is called.
			tc.lookup.InheritSecrets(tc.other, tc.secretRefs)

			// THEN the Lookup inherits the secrets as expected.
			wantStr := tc.want.String(tc.want, "")
			gotStr := tc.lookup.String(tc.lookup, "")
			if wantStr != gotStr {
				t.Errorf("web.Lookup.InheritSecrets() = %v, want %v",
					gotStr, wantStr)
			}
		})
	}
}
