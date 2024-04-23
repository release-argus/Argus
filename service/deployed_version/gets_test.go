// Copyright [2023] [Argus]
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
	"os"
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestLookup_GetAllowInvalidCerts(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		root        *bool
		dfault      *bool
		hardDefault *bool
		wantBool    bool
	}{
		"root overrides all": {
			wantBool:    true,
			root:        test.BoolPtr(true),
			dfault:      test.BoolPtr(false),
			hardDefault: test.BoolPtr(false)},
		"default overrides hardDefault": {
			wantBool:    true,
			dfault:      test.BoolPtr(true),
			hardDefault: test.BoolPtr(false)},
		"hardDefault is last resort": {
			wantBool:    true,
			hardDefault: test.BoolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup()
			lookup.AllowInvalidCerts = tc.root
			lookup.Defaults.AllowInvalidCerts = tc.dfault
			lookup.HardDefaults.AllowInvalidCerts = tc.hardDefault

			// WHEN GetAllowInvalidCerts is called
			got := lookup.GetAllowInvalidCerts()

			// THEN the function returns the correct result
			if got != tc.wantBool {
				t.Errorf("want: %t\ngot:  %t",
					tc.wantBool, got)
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
			env:  map[string]string{"TESTLOOKUP_DV_GETURL_ONE": "https://example.com"},
			url:  "${TESTLOOKUP_DV_GETURL_ONE}",
			want: "https://example.com",
		},
		"returns URL partially from env": {
			env:  map[string]string{"TESTLOOKUP_DV_GETURL_TWO": "example.com"},
			url:  "https://${TESTLOOKUP_DV_GETURL_TWO}",
			want: "https://example.com",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
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
