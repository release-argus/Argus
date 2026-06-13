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

// Package base provides the base struct for deployed_version lookups.
package base

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
)

// ############
// # DECODING #
// ############

func TestDecodeDefaults(t *testing.T) {
	// GIVEN: data in a given format to Decode into Defaults.
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
			errRegex: "^$",
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"type": "url",
				"allow_invalid_certs": true,
				"method": "GET"
			}`),
			want: test.TrimYAML(`
				type: url
				allow_invalid_certs: true
				method: GET
			`),
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				type: url
				allow_invalid_certs: true
				method: GET
			`),
			want: test.TrimYAML(`
				type: url
				allow_invalid_certs: true
				method: GET
			`),
		},
		{
			name:   "JSON/invalid format",
			format: "json",
			data:   `{"allow_invalid_certs": "true"}`,
			errRegex: test.TrimYAML(`
				^deployed_version:
					json: .*unmarshal.*$`,
			),
		},
		{
			name:   "YAML/invalid deployed_version data types",
			format: "yaml",
			data:   `allow_invalid_certs: maybe`,
			errRegex: test.TrimYAML(`
				^deployed_version:
					[^\s]+.* string .*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, testErr := test.AssertDecode(
				t,
				DecodeDefaults,
				tc.format, tc.data,
				func(v *Defaults) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"DecodeDefaults",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

// #########
// # STATE #
// #########

func TestDefaults_IsZero(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name string
		opt  *Defaults
		want bool
	}{
		{
			name: "empty",
			opt:  &Defaults{},
			want: true,
		},
		{
			name: "non-empty Type",
			opt: &Defaults{
				Type: "url",
			},
			want: false,
		},
		{
			name: "non-empty AllowInvalidCerts",
			opt: &Defaults{
				AllowInvalidCerts: test.Ptr(true),
			},
			want: false,
		},
		{
			name: "non-empty Method",
			opt: &Defaults{
				Method: http.MethodPost,
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called.
			got := tc.opt.IsZero()

			// THEN: it should return the expected value.
			if got != tc.want {
				t.Errorf(
					"%s\nDefaults.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

// ########
// # INIT #
// ########

func TestDefaults_Default(t *testing.T) {
	// GIVEN: Defaults.
	defaults := Defaults{}

	// WHEN: Default is called.
	defaults.Default()

	// THEN: it should set the defaults.
	if defaults.AllowInvalidCerts == nil {
		t.Errorf("%s\nDefaults.Default() .AllowInvalidCerts not set\ngot:  nil\nwant: non-nil", packageName)
	}
}

// ##############
// # VALIDATION #
// ##############

func TestDefaults_CheckValues(t *testing.T) {
	// GIVEN: Defaults.
	tests := []struct {
		name       string
		method     string
		errRegex   string
		wantMethod string
	}{
		{
			name:       "empty method - no error",
			method:     "",
			errRegex:   `^$`,
			wantMethod: "",
		},
		{
			name:       "valid lowercase method - uppercased and ok",
			method:     "post",
			errRegex:   `^$`,
			wantMethod: http.MethodPost,
		},
		{
			name:       "valid uppercase method - unchanged and ok",
			method:     "GET",
			errRegex:   `^$`,
			wantMethod: http.MethodGet,
		},
		{
			name:   "unsupported method",
			method: http.MethodDelete,
			errRegex: fmt.Sprintf(
				`^method: "%s" <invalid> .*%s.*$`,
				http.MethodDelete, http.MethodGet,
			),
			wantMethod: http.MethodDelete,
		},
		{
			name:   "invalid method",
			method: "foo",
			errRegex: fmt.Sprintf(
				`^method: "FOO" <invalid> .*%s.*$`,
				http.MethodPost,
			),
			wantMethod: "FOO",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := Defaults{Method: tc.method}

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				input.CheckValues,
			)

			// AND: Method is uppercased/unchanged as expected.
			if input.Method != tc.wantMethod {
				t.Errorf(
					"%s\nDefaults.CheckValues() .Method mismatch\ngot:  %q\nwant: %q",
					packageName, input.Method, tc.wantMethod,
				)
			}
		})
	}
}
