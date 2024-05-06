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
	"github.com/release-argus/Argus/test"
)

func TestLookup_String(t *testing.T) {
	tests := map[string]struct {
		lookup *Lookup
		want   string
	}{
		"nil": {
			lookup: nil,
			want:   "",
		},
		"empty": {
			lookup: &Lookup{},
			want:   "{}",
		},
		"filled": {
			lookup: New(
				test.BoolPtr(false),
				&BasicAuth{
					Username: "user", Password: "pass"},
				test.StringPtr("body_here"),
				&[]Header{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				"value.version",
				"GET",
				opt.New(
					test.BoolPtr(true), "9m", test.BoolPtr(false),
					nil, nil),
				"v([0-9.]+)", test.StringPtr("$1"),
				&svcstatus.Status{},
				"https://example.com",
				NewDefaults(
					test.BoolPtr(false)),
				NewDefaults(
					test.BoolPtr(false))),
			want: `
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
regex_template: $1`,
		},
		"quotes otherwise invalid yaml strings": {
			lookup: New(
				nil,
				&BasicAuth{
					Username: ">123", Password: "{pass}"},
				nil, nil, "", "", nil, "", nil, &svcstatus.Status{}, "", nil, nil),
			want: `
basic_auth:
  username: '>123'
  password: '{pass}'`,
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

				// WHEN the Lookup is stringified with String
				got := tc.lookup.String(prefix)

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
			a: New(
				test.BoolPtr(false),
				&BasicAuth{
					Username: "user", Password: "pass"},
				test.StringPtr("body_here"),
				&[]Header{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				"value.version",
				"GET",
				opt.New(
					nil, "", nil,
					nil, nil),
				"v([0-9.]+)", test.StringPtr("$1"),
				&svcstatus.Status{},
				"https://example.com",
				NewDefaults(
					test.BoolPtr(false)),
				NewDefaults(
					test.BoolPtr(false))),
			b: New(
				test.BoolPtr(false),
				&BasicAuth{
					Username: "user", Password: "pass"},
				test.StringPtr("body_here"),
				&[]Header{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				"value.version",
				"GET",
				opt.New(
					nil, "", test.BoolPtr(true),
					nil, nil),
				"v([0-9.]+)", test.StringPtr("$1"),
				&svcstatus.Status{},
				"https://example.com",
				NewDefaults(
					test.BoolPtr(false)),
				NewDefaults(
					test.BoolPtr(false))),
			want: true,
		},
		"not equal": {
			a: New(
				test.BoolPtr(false),
				&BasicAuth{
					Username: "user", Password: "pass"},
				test.StringPtr("body_here"),
				&[]Header{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				"value.version",
				"GET",
				opt.New(
					nil, "", test.BoolPtr(true),
					nil, nil),
				"v([0-9.]+)", test.StringPtr("$1"),
				&svcstatus.Status{},
				"https://example.com",
				NewDefaults(
					test.BoolPtr(false)),
				NewDefaults(
					test.BoolPtr(false))),
			b: New(
				test.BoolPtr(false),
				&BasicAuth{
					Username: "user", Password: "pass"},
				test.StringPtr("body_here"),
				&[]Header{
					{Key: "X-Header", Value: "val"},
					{Key: "X-Another", Value: "val2"}},
				"value.version",
				"GET",
				opt.New(
					nil, "", test.BoolPtr(true),
					nil, nil),
				"v([0-9.]+)", test.StringPtr("$1"),
				&svcstatus.Status{},
				"https://example.com/other",
				NewDefaults(
					test.BoolPtr(false)),
				NewDefaults(
					test.BoolPtr(false))),
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
