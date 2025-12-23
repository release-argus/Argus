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
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestLookup_allowInvalidCerts(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		rootValue, defaultValue, hardDefaultValue *bool
		want                                      bool
	}{
		"root overrides all": {
			want:             true,
			rootValue:        test.BoolPtr(true),
			defaultValue:     test.BoolPtr(false),
			hardDefaultValue: test.BoolPtr(false)},
		"default overrides hardDefault": {
			want:             true,
			defaultValue:     test.BoolPtr(true),
			hardDefaultValue: test.BoolPtr(false)},
		"hardDefault is last resort": {
			want:             true,
			hardDefaultValue: test.BoolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.AllowInvalidCerts = tc.rootValue
			lookup.Defaults.AllowInvalidCerts = tc.defaultValue
			lookup.HardDefaults.AllowInvalidCerts = tc.hardDefaultValue

			// WHEN allowInvalidCerts is called.
			got := lookup.allowInvalidCerts()

			// THEN the function returns the correct result.
			if got != tc.want {
				t.Errorf("%s\nwant: %t\ngot:  %t",
					packageName, tc.want, got)
			}
		})
	}
}

func TestLookup_url(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		env  map[string]string
		url  string
		want string
	}{
		"URL": {
			url:  "https://example.com",
			want: "https://example.com",
		},
		"URL from env": {
			env:  map[string]string{"TEST_LOOKUP__DV_GET_URL_ONE": "https://example.com"},
			url:  "${TEST_LOOKUP__DV_GET_URL_ONE}",
			want: "https://example.com",
		},
		"URL with env partial": {
			env:  map[string]string{"TEST_LOOKUP__DV_GET_URL_TWO": "example.com"},
			url:  "https://${TEST_LOOKUP__DV_GET_URL_TWO}",
			want: "https://example.com",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				_ = os.Setenv(k, v)
				t.Cleanup(func() { _ = os.Unsetenv(k) })
			}

			lookup := testLookup(false)
			lookup.URL = tc.url

			// WHEN url is called.
			got := lookup.url()

			// THEN the function returns the correct result.
			if got != tc.want {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestLookup_body(t *testing.T) {
	// GIVEN a Lookup.
	tests := map[string]struct {
		body string
		want io.Reader
	}{
		"empty body": {
			body: "",
			want: nil,
		},
		"non-empty body": {
			body: "test body",
			want: strings.NewReader("test body"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(false)
			lookup.Body = tc.body

			// WHEN body is called.
			got := lookup.body()

			// THEN the function returns the correct result.
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("%s\nwant: %v\ngot:  %v",
					packageName, tc.want, got)
			}
		})
	}
}
