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
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	"github.com/release-argus/Argus/service/deployed_version/types/web/constants"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
)

// ############
// # DECODING #
// ############

func TestLookup_Unmarshal(t *testing.T) {
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"type": "url",
				"method": "GET",
				"url": "https://example.com",
				"allow_invalid_certs": false,
				"basic_auth": {
					"username": "user",
					"password": "pass"
				},
				"headers": [
					{"key": "X-Header",  "value": "val"},
					{"key": "X-Another", "value": "val2"}
				],
				"body": "body_here",
				"json": "value.version",
				"regex": "v([0-9.]+)",
				"regex_template": "$1"
			}`),
			errRegex: `^$`,
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
				regex_template: $1
			`),
		},
		{
			name:   "YAML/filled",
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
				body: body_here
				json: value.version
				regex: v([0-9.]+)
				regex_template: $1
			`),
			errRegex: `^$`,
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
				regex_template: $1
			`),
		},
		{
			name:     "JSON/invalid data type",
			format:   "json",
			data:     `{"url": ["https://example.com"]}`,
			errRegex: `^json: .*unmarshal.* array.*$`,
		},
		{
			name:   "YAML/invalid data type",
			format: "yaml",
			data:   `url: ["https://example.com"]`,
			errRegex: test.TrimYAML(`
				^[^\s]+ .*unmarshal.*
				\>  \d | .+$`,
			),
		},
		{
			name:   "Type: web -> url",
			format: "yaml",
			data: test.TrimYAML(`
				type: web
				method: GET
				url: https://example.com
			`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: url
				method: GET
				url: https://example.com
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					var zero Lookup
					err := decode.Unmarshal(format, data, &zero)
					return &zero, err
				},
				tc.format, tc.data,
				func(v *Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"Lookup.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestString(t *testing.T) {
	dvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig(t)

	tests := []struct {
		name   string
		lookup *Lookup
		want   string
	}{
		{
			name:   "empty",
			lookup: &Lookup{},
			want:   "{}\n",
		},
		{
			name: "filled",
			lookup: test.Must(t, func() (*Lookup, error) {
				options, err := opt.Decode(
					"yaml", []byte("interval: 1h2m3s"),
					optCfg,
				)
				if err != nil {
					return nil, err
				}
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
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
					`)),
					options,
					nil,
					base.DefaultsConfig{
						Soft: &base.Defaults{
							AllowInvalidCerts: new(false),
						},
						Hard: &base.Defaults{
							AllowInvalidCerts: new(false),
						},
					},
				)
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
				regex_template: $1
			`),
		},
		{
			name: "quotes otherwise invalid YAML strings",
			lookup: test.Must(t, func() (*Lookup, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						basic_auth:
							username: '>123'
							password: '{pass}'
					`)),
					nil,
					nil,
					dvCfg,
				)
			}),
			want: test.TrimYAML(`
				type: url
				basic_auth:
					username: '>123'
					password: '{pass}'
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				"Lookup.String",
				func(prefix string) string {
					return tc.lookup.String(prefix)
				},
				tc.want,
			)
		})
	}
}

// #########
// # STATE #
// #########

func TestLookup_Copy(t *testing.T) {
	dvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig(t)

	tests := []struct {
		name   string
		lookup *Lookup
		status *status.Status
	}{
		{
			name:   "nil",
			lookup: nil,
			status: nil,
		},
		{
			name: "basic_auth",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						basic_auth:
							username: user
							password: pass
					`)),
					test.Must(t, func() (*opt.Options, error) {
						return opt.Decode("yaml", nil, optCfg)
					}),
					svcStatus,
					dvCfg,
				)
			}),
			status: nil,
		},
		{
			name: "headers",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						headers:
							- key: X-Something
								value: foo
					`)),
					test.Must(t, func() (*opt.Options, error) {
						return opt.Decode("yaml", nil, optCfg)
					}),
					svcStatus,
					dvCfg,
				)
			}),
			status: nil,
		},
		{
			name: "options",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(`type: url`),
					test.Must(t, func() (*opt.Options, error) {
						return opt.Decode(
							"yaml", []byte(test.TrimYAML(`
								active: false
								interval: 2s
								semantic_versioning: false
							`)),
							optCfg,
						)
					}),
					svcStatus,
					dvCfg,
				)
			}),
			status: nil,
		},
		{
			name: "filled",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
								type: test
								method: `+constants.SupportedMethods[0]+`
								url: test
								allow_invalid_certs: true
								basic_auth:
									username: user
									password: pass
								target_header: X-Foo
								headers:
									- key: X-Something
										value: foo
								body: b
								json: j
								regex: r
								regex_template: t$1
					`)),
					test.Must(t, func() (*opt.Options, error) {
						return opt.Decode(
							"yaml", []byte(test.TrimYAML(`
								active: false
								interval: 2s
								semantic_versioning: false
							`)),
							optCfg,
						)
					}),
					svcStatus,
					dvCfg,
				)
			}),
			status: test.Must(t, func() (*status.Status, error) {
				return statustest.New("yaml", nil)
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wantStr := decode.ToYAMLString(tc.lookup, "")

			gotInterface := tc.lookup.Copy(tc.status)

			if tc.lookup == nil {
				if gotInterface != nil {
					t.Error("Lookup.Copy() mismatch\ngot:  non-nil\nwant: nil")
				}
				return
			}

			if gotInterface == nil {
				t.Fatal("Lookup.Copy() mismatch\ngot:  nil\nwant: non-nil")
			}

			if gotInterface == tc.lookup {
				t.Fatalf(
					"copy aliases the original\ngot:  %p\nwant: NOT %p",
					gotInterface, tc.lookup,
				)
			}

			got, ok := gotInterface.(*Lookup)
			if !ok {
				t.Fatalf("type mismatch\ngot:  %T\nwant: *Lookup", gotInterface)
			}

			if gotStr := got.String(""); gotStr != wantStr {
				t.Fatalf("Lookup.String() mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Type", Got: got.Type, Want: tc.lookup.Type, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Options", Got: got.Options, Want: tc.lookup.Options, Mode: test.CopiedPointerMode(tc.lookup.Options)},
				{Name: "Status", Got: got.Status, Want: tc.status, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Method", Got: got.Method, Want: tc.lookup.Method, Mode: test.CompareEqual},
				{Name: "URL", Got: got.URL, Want: tc.lookup.URL, Mode: test.CompareEqual},
				{Name: "AllowInvalidCerts", Got: got.AllowInvalidCerts, Want: tc.lookup.AllowInvalidCerts, Mode: test.CopiedPointerMode(tc.lookup.AllowInvalidCerts)},
				{Name: "TargetHeader", Got: got.TargetHeader, Want: tc.lookup.TargetHeader, Mode: test.CompareEqual},
				{Name: "BasicAuth", Got: got.BasicAuth, Want: tc.lookup.BasicAuth, Mode: test.CopiedPointerMode(tc.lookup.BasicAuth)},
				{Name: "Body", Got: got.Body, Want: tc.lookup.Body, Mode: test.CompareEqual},
				{Name: "JSON", Got: got.JSON, Want: tc.lookup.JSON, Mode: test.CompareEqual},
				{Name: "Regex", Got: got.Regex, Want: tc.lookup.Regex, Mode: test.CompareEqual},
				{Name: "RegexTemplate", Got: got.RegexTemplate, Want: tc.lookup.RegexTemplate, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			if len(got.Headers) > 0 && len(tc.lookup.Headers) > 0 && &got.Headers[0] == &tc.lookup.Headers[0] {
				t.Errorf(
					"Lookup.Headers: copy aliases the original\ngot:  %p\nwant: NOT %p",
					&got.Headers[0], &tc.lookup.Headers[0],
				)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Defaults", Got: got.Defaults, Want: tc.lookup.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: tc.lookup.HardDefaults, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestBasicAuth_Copy(t *testing.T) {
	tests := []struct {
		name string
		auth *BasicAuth
	}{
		{
			name: "nil",
			auth: nil,
		},
		{
			name: "empty",
			auth: &BasicAuth{},
		},
		{
			name: "filled",
			auth: &BasicAuth{
				Username: "user",
				Password: "password",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.auth.Copy()

			if tc.auth == nil {
				if got != nil {
					t.Error("BasicAuth.Copy() mismatch\ngot:  non-nil\nwant: nil")
				}
				return
			}
			if got == nil {
				t.Fatal("BasicAuth.Copy() mismatch\ngot:  nil\nwant: non-nil")
			}

			if got == tc.auth {
				t.Errorf("copy aliases the original\ngot:  %p\nwant: NOT %p", got, tc.auth)
			}

			if got.Username != tc.auth.Username {
				t.Errorf(
					"BasicAuth.Username mismatch\ngot:  %q\nwant: %q",
					got.Username, tc.auth.Username,
				)
			}
			if got.Password != tc.auth.Password {
				t.Errorf(
					"BasicAuth.Password mismatch\ngot:  %q\nwant: %q",
					got.Password, tc.auth.Password,
				)
			}
		})
	}
}

func TestLookup_InheritSecrets(t *testing.T) {
	tests := []struct {
		name       string
		lookup     *Lookup
		other      *Lookup
		secretRefs *shared.VSecretRef
		want       *Lookup
	}{
		{
			name: "inherit BasicAuth password",
			lookup: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: util.SecretValue,
				},
			},
			other: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "password",
				},
			},
			secretRefs: &shared.VSecretRef{},
			want: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "password",
				},
			},
		},
		{
			name: "inherit headers",
			lookup: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
				},
			},
			other: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: "secret"},
				},
			},
			secretRefs: &shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
				},
			},
			want: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: "secret"},
				},
			},
		},
		{
			name: "no inheritance when no secrets",
			lookup: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "password",
				},
				Headers: shared.Headers{
					{Key: "X-Test", Value: "value"},
				},
			},
			other: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "other_password",
				},
				Headers: shared.Headers{
					{Key: "X-Test", Value: "other_value"},
				},
			},
			secretRefs: &shared.VSecretRef{},
			want: &Lookup{
				BasicAuth: &BasicAuth{
					Username: "user",
					Password: "password",
				},
				Headers: shared.Headers{
					{Key: "X-Test", Value: "value"},
				},
			},
		},
		{
			name: "no inheritance when secretRefs out of bounds",
			lookup: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
					{Key: "X-Other", Value: util.SecretValue},
				},
			},
			other: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: "secret"},
				},
			},
			secretRefs: &shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(1)},
				},
			},
			want: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
					{Key: "X-Other", Value: util.SecretValue},
				},
			},
		},
		{
			name: "no inheritance when secretRef index nil",
			lookup: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
				},
			},
			other: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: "secret"},
				},
			},
			secretRefs: &shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: nil},
				},
			},
			want: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.lookup.InheritSecrets(tc.other, tc.secretRefs)

			if got, want := tc.lookup.String(""), tc.want.String(""); got != want {
				t.Errorf("Lookup.String() mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}
