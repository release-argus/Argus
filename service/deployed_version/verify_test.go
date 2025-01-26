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
	"net/http"
	"strings"
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		method, url          string
		body                 string
		json                 string
		regex, regexTemplate string
		defaults             *Defaults
		errRegex             string
		nilService           bool
	}{
		"nil service": {
			errRegex:   `^$`,
			nilService: true,
		},
		"valid service": {
			errRegex: `^$`,
			method:   http.MethodGet,
			url:      "https://example.com",
			regex:    `[0-9.]+`,
			defaults: &Defaults{},
		},
		"method - empty string": {
			errRegex: `^$`,
			method:   "",
			url:      "https://example.com",
		},
		"method - invalid": {
			errRegex: `method: "[^"]+" <invalid>`,
			method:   "FOO",
			url:      "https://example.com",
		},
		"method - valid": {
			errRegex: `^$`,
			method:   http.MethodGet,
			url:      "https://example.com",
		},
		"method - case insensitive": {
			errRegex: `^$`,
			method:   "gEt",
			url:      "https://example.com",
		},
		"url - empty string": {
			errRegex: `url: <required>`,
			method:   http.MethodGet,
			url:      "",
			defaults: &Defaults{},
		},
		"body - removed for GET": {
			errRegex: `^$`,
			method:   http.MethodGet,
			url:      "https://example.com",
			body:     "foo",
			defaults: &Defaults{},
		},
		"body - not removed for POST": {
			errRegex: `^$`,
			method:   http.MethodPost,
			url:      "https://example.com",
			body:     "foo",
			defaults: &Defaults{},
		},
		"json - invalid, string in square brackets": {
			errRegex: `json: .* <invalid>`,
			method:   http.MethodGet,
			json:     "foo[bar]",
			defaults: &Defaults{},
		},
		"regex - invalid": {
			errRegex: `regex: .* <invalid>`,
			method:   http.MethodGet,
			regex:    `[0-`,
			defaults: &Defaults{},
		},
		"regexTemplate - with no regex": {
			method:        http.MethodGet,
			url:           "https://example.com",
			errRegex:      `^$`,
			regexTemplate: "$1.$2.$3",
			defaults:      &Defaults{},
		},
		"all errs": {
			errRegex: `url: <required>`,
			method:   http.MethodGet,
			url:      "",
			regex:    `[0-`,
			defaults: &Defaults{},
		},
		"no url doesn't fail for Lookup Defaults": {
			errRegex: `^$`,
			method:   http.MethodGet,
			url:      "",
			defaults: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{}
			lookup = testLookup()
			lookup.Method = tc.method
			lookup.URL = tc.url
			lookup.Body = tc.body
			lookup.JSON = tc.json
			lookup.Regex = tc.regex
			lookup.RegexTemplate = tc.regexTemplate
			lookup.Defaults = nil
			if tc.defaults != nil {
				lookup.Defaults = tc.defaults
			}
			var hadBody string
			if tc.nilService {
				lookup = nil
			} else {
				hadBody = lookup.Body
			}

			// WHEN CheckValues is called
			err := lookup.CheckValues("")

			// THEN it errors when expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			if lookup == nil {
				return
			}
			// AND RegexTemplate is empty when Regex is empty
			if lookup.RegexTemplate != "" && lookup.Regex == "" {
				t.Fatalf("RegexTemplate should be nil when Regex is empty")
			}
			// AND Body is empty when Method is GET
			if lookup.Method == http.MethodGet && lookup.Body != "" {
				t.Fatalf("Body should be nil when Method is GET")
			}
			// AND Body is kept when Method is POST
			if lookup.Method == http.MethodPost && hadBody != "" && lookup.Body == "" {
				t.Fatalf("Body should be kept when Method is POST")
			}
			// AND Method is uppercased
			wantMethod := strings.ToUpper(tc.method)
			if wantMethod == "" {
				wantMethod = http.MethodGet
			}
			if lookup.Method != wantMethod {
				t.Fatalf("Method should be uppercased:\nwant: %q\ngot:  %q",
					wantMethod, lookup.Method)
			}
		})
	}
}
