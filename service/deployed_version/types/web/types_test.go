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
	// GIVEN: data to unmarshal into a Lookup.
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
				packageName,
				"Lookup",
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
	dvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	// GIVEN: a Lookup.
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
				options, _ := opt.Decode(
					"yaml", []byte("interval: 1h2m3s"),
					optCfg,
				)
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
							AllowInvalidCerts: test.Ptr(false),
						},
						Hard: &base.Defaults{
							AllowInvalidCerts: test.Ptr(false),
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
				packageName,
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
	dvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	// GIVEN: a Lookup.
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
					"yaml", nil,
					test.Must(t, func() (*opt.Options, error) {
						return opt.Decode(
							"yaml", []byte(test.TrimYAML(`
								basic_auth:
									username: user
									password: pass
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
			name: "headers",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", nil,
					test.Must(t, func() (*opt.Options, error) {
						return opt.Decode(
							"yaml", []byte(test.TrimYAML(`
								headers:
									- key: X-Something
										value: foo
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
			name: "options",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", nil,
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
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wantStr := decode.ToYAMLString(tc.lookup, "")

			// WHEN: Copy() is called on it.
			gotInterface := tc.lookup.Copy(tc.status)

			prefix := fmt.Sprintf(
				"%s\nLookup.Copy(status=%p)",
				packageName, tc.status,
			)

			// THEN: if nil was copied, we get nil.
			if tc.lookup == nil {
				if gotInterface != nil {
					t.Errorf(
						"%s of nil mismatch\ngot:  %v\nwant: nil",
						prefix, gotInterface,
					)
				}
				return
			}

			// AND: the copy is non-nil.
			if gotInterface == nil {
				t.Fatalf("%s got nil want non-nil", prefix)
			}

			// AND: the copy is distinct.
			if gotInterface == tc.lookup {
				t.Fatalf(
					"%s should return a distinct copy\ngot:  %p\nwant: %p",
					prefix, gotInterface, tc.lookup,
				)
			}

			// AND: the type is unchanged.
			got, ok := gotInterface.(*Lookup)
			if !ok {
				t.Fatalf(
					"%s type shouldn't have changed\ngot:  %T\nwant: Lookup",
					prefix, gotInterface,
				)
			}

			// AND: the copy unmarshals the same.
			if gotStr := got.String(""); gotStr != wantStr {
				t.Fatalf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}

			// AND: the fields are copied as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Type", Got: got.Type, Want: tc.lookup.Type, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}

			// AND: copied pointers should be value-equal and non-aliased.
			fieldTests = []test.FieldAssertion{
				{Name: "Options", Got: got.Options, Want: tc.lookup.Options, Mode: test.CompareDifferentPointer},
				{Name: "Status", Got: got.Status, Want: tc.status, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}

			// AND: the non-Base fields are copied as expected.
			fieldTests = []test.FieldAssertion{
				{Name: "Method", Got: got.Method, Want: tc.lookup.Method, Mode: test.CompareEqual},
				{Name: "URL", Got: got.URL, Want: tc.lookup.URL, Mode: test.CompareEqual},
				{Name: "AllowInvalidCerts", Got: got.AllowInvalidCerts, Want: tc.lookup.AllowInvalidCerts, Mode: test.CompareDifferentPointer},
				{Name: "TargetHeader", Got: got.TargetHeader, Want: tc.lookup.TargetHeader, Mode: test.CompareEqual},
				{Name: "BasicAuth", Got: got.BasicAuth, Want: tc.lookup.BasicAuth, Mode: test.CompareDifferentPointer},
				{Name: "Headers", Got: &got.Headers, Want: &tc.lookup.Headers, Mode: test.CompareDifferentPointer},
				{Name: "Body", Got: got.Body, Want: tc.lookup.Body, Mode: test.CompareEqual},
				{Name: "JSON", Got: got.JSON, Want: tc.lookup.JSON, Mode: test.CompareEqual},
				{Name: "Regex", Got: got.Regex, Want: tc.lookup.Regex, Mode: test.CompareEqual},
				{Name: "RegexTemplate", Got: got.RegexTemplate, Want: tc.lookup.RegexTemplate, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}

			// AND: defaults pointers are shared.
			fieldTests = []test.FieldAssertion{
				{Name: "Defaults", Got: got.Defaults, Want: tc.lookup.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: tc.lookup.HardDefaults, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBasicAuth_Copy(t *testing.T) {
	// GIVEN: a BasicAuth.
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
			// WHEN: Copy is called on it.
			got := tc.auth.Copy()

			prefix := fmt.Sprintf("%s\nBasicAuth.Copy()", packageName)

			// THEN: nil should be returned when nil is copied.
			if tc.auth == nil {
				if got != nil {
					t.Errorf("%s returned non-nil for nil input", prefix)
				}
				return
			}
			// THEN: the address should be different.
			if got == tc.auth {
				t.Errorf(
					"%s returned the same address: %p",
					prefix, got,
				)
			}

			// AND: the values should be the same.
			if got.Username != tc.auth.Username {
				t.Errorf(
					"%s returned different .Username value\ngot:  %q\nwant: %q",
					prefix, got.Username, tc.auth.Username,
				)
			}
			if got.Password != tc.auth.Password {
				t.Errorf(
					"%s returned different .Password value\ngot:  %q\nwant: %q",
					prefix, got.Password, tc.auth.Password,
				)
			}
		})
	}
}

func TestLookup_InheritSecrets(t *testing.T) {
	// GIVEN: a Lookup with secrets, and another Lookup to inherit from.
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
					{OldIndex: test.Ptr(0)},
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
			name: "no inheritance when no matching headers",
			lookup: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
				},
			},
			other: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Other", Value: "secret"},
				},
			},
			secretRefs: &shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.Ptr(1)},
				},
			},
			want: &Lookup{
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
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
					{OldIndex: test.Ptr(1)},
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

			// WHEN: InheritSecrets is called.
			tc.lookup.InheritSecrets(tc.other, tc.secretRefs)

			// THEN: the Lookup inherits the secrets as expected.
			if got, want := tc.lookup.String(""), tc.want.String(""); got != want {
				t.Errorf(
					"%s\nLookup.InheritSecrets() stringified mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}
		})
	}
}
