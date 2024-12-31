// Copyright [2024] [Argus]
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

package deployedver

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestLookup_GetAllowInvalidCerts(t *testing.T) {
	// GIVEN a Lookup
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

			lookup := testLookup()
			lookup.AllowInvalidCerts = tc.rootValue
			lookup.Defaults.AllowInvalidCerts = tc.defaultValue
			lookup.HardDefaults.AllowInvalidCerts = tc.hardDefaultValue

			// WHEN GetAllowInvalidCerts is called
			got := lookup.GetAllowInvalidCerts()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %t\ngot:  %t",
					tc.want, got)
			}
		})
	}
}

func TestLookup_GetURL(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		env  map[string]string
		url  string
		want string
	}{
		"returns URL": {
			url:  "https://example.com",
			want: "https://example.com",
		},
		"returns URL from env": {
			env:  map[string]string{"TEST_LOOKUP__DV_GET_URL_ONE": "https://example.com"},
			url:  "${TEST_LOOKUP__DV_GET_URL_ONE}",
			want: "https://example.com",
		},
		"returns URL partially from env": {
			env:  map[string]string{"TEST_LOOKUP__DV_GET_URL_TWO": "example.com"},
			url:  "https://${TEST_LOOKUP__DV_GET_URL_TWO}",
			want: "https://example.com",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}

			lookup := testLookup()
			lookup.URL = tc.url

			// WHEN GetURL is called
			got := lookup.GetURL()

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestLookup_GetBody(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		body string
		want io.Reader
	}{
		"empty body": {
			body: "",
			want: nil,
		},
		"non-empty body": {
			body: ("test body"),
			want: strings.NewReader("test body"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup()
			lookup.Body = tc.body

			// WHEN GetBody is called
			got := lookup.GetBody()

			// THEN the function returns the correct result
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("want: %v\ngot:  %v",
					tc.want, got)
			}
		})
	}
}
