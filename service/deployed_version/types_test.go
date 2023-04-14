// Copyright [2023] [Argus]
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

	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
)

func TestLookup_String(t *testing.T) {
	tests := map[string]struct {
		lookup Lookup
		want   string
	}{
		"empty": {
			lookup: Lookup{},
			want:   "{}\n",
		},
		"filled": {
			lookup: Lookup{
				URL:               "https://example.com",
				AllowInvalidCerts: boolPtr(false),
				BasicAuth: &BasicAuth{
					Username: "user", Password: "pass"},
				Headers: []Header{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				JSON:  "value.version",
				Regex: "v([0-9.]+)",
				Options: &opt.Options{
					SemanticVersioning: boolPtr(true)},
				Status:       &svcstatus.Status{},
				Defaults:     &Lookup{AllowInvalidCerts: boolPtr(false)},
				HardDefaults: &Lookup{AllowInvalidCerts: boolPtr(false)}},
			want: `
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
json: value.version
regex: v([0-9.]+)
`,
		},
		"quotes otherwise invalid yaml strings": {
			lookup: Lookup{
				BasicAuth: &BasicAuth{
					Username: ">123", Password: "{pass}"}},
			want: `
basic_auth:
    username: '>123'
    password: '{pass}'
`},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Lookup is stringified with String
			got := tc.lookup.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
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
				Defaults: &Lookup{
					AllowInvalidCerts: boolPtr(false)}},
			b:    &Lookup{},
			want: true,
		},
		"hard_defaults ignored": {
			a: &Lookup{
				HardDefaults: &Lookup{
					AllowInvalidCerts: boolPtr(false)}},
			b:    &Lookup{},
			want: true,
		},
		"equal": {
			a: &Lookup{
				URL:               "https://example.com",
				AllowInvalidCerts: boolPtr(false),
				BasicAuth: &BasicAuth{
					Username: "user", Password: "pass"},
				Headers: []Header{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				JSON:  "value.version",
				Regex: "v([0-9.]+)",
				Options: &opt.Options{
					SemanticVersioning: boolPtr(true)},
				Defaults:     &Lookup{AllowInvalidCerts: boolPtr(false)},
				HardDefaults: &Lookup{AllowInvalidCerts: boolPtr(false)},
			},
			b: &Lookup{
				URL:               "https://example.com",
				AllowInvalidCerts: boolPtr(false),
				BasicAuth: &BasicAuth{
					Username: "user", Password: "pass"},
				Headers: []Header{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				JSON:  "value.version",
				Regex: "v([0-9.]+)",
				Options: &opt.Options{
					SemanticVersioning: boolPtr(true)},
				Defaults:     &Lookup{AllowInvalidCerts: boolPtr(false)},
				HardDefaults: &Lookup{AllowInvalidCerts: boolPtr(false)},
			},
			want: true,
		},
		"not equal": {
			a: &Lookup{
				URL:               "https://example.com",
				AllowInvalidCerts: boolPtr(false),
				BasicAuth: &BasicAuth{
					Username: "user", Password: "pass"},
				Headers: []Header{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				JSON:  "value.version",
				Regex: "v([0-9.]+)",
				Options: &opt.Options{
					SemanticVersioning: boolPtr(true)},
				Defaults:     &Lookup{AllowInvalidCerts: boolPtr(false)},
				HardDefaults: &Lookup{AllowInvalidCerts: boolPtr(false)},
			},
			b: &Lookup{
				URL:               "https://example.com/other",
				AllowInvalidCerts: boolPtr(false),
				BasicAuth: &BasicAuth{
					Username: "user", Password: "pass"},
				Headers: []Header{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				JSON:  "value.version",
				Regex: "v([0-9.]+)",
				Options: &opt.Options{
					SemanticVersioning: boolPtr(true)},
				Defaults:     &Lookup{AllowInvalidCerts: boolPtr(false)},
				HardDefaults: &Lookup{AllowInvalidCerts: boolPtr(false)},
			},
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
		name, tc := name, tc
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
