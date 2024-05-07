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
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		method        string
		url           string
		body          *string
		json          string
		regex         string
		regexTemplate *string
		defaults      *LookupDefaults
		errRegex      string
		nilService    bool
	}{
		"nil service": {
			errRegex:   `^$`,
			nilService: true,
		},
		"valid service": {
			errRegex: `^$`,
			method:   "GET",
			url:      "https://example.com",
			regex:    "[0-9.]+",
			defaults: &LookupDefaults{},
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
			method:   "GET",
			url:      "https://example.com",
		},
		"method - case insensitive": {
			errRegex: `^$`,
			method:   "gEt",
			url:      "https://example.com",
		},
		"url - empty string": {
			errRegex: `url: <required>`,
			method:   "GET",
			url:      "",
			defaults: &LookupDefaults{},
		},
		"body - removed for GET": {
			errRegex: `^$`,
			method:   "GET",
			url:      "https://example.com",
			body:     test.StringPtr("foo"),
			defaults: &LookupDefaults{},
		},
		"body - not removed for POST": {
			errRegex: `^$`,
			method:   "POST",
			url:      "https://example.com",
			body:     test.StringPtr("foo"),
			defaults: &LookupDefaults{},
		},
		"json - invalid, string in square brackets": {
			errRegex: `json: .* <invalid>`,
			method:   "GET",
			json:     "foo[bar]",
			defaults: &LookupDefaults{},
		},
		"regex - invalid": {
			errRegex: `regex: .* <invalid>`,
			method:   "GET",
			regex:    "[0-",
			defaults: &LookupDefaults{},
		},
		"regexTemplate - with no regex": {
			method:        "GET",
			url:           "https://example.com",
			errRegex:      `^$`,
			regexTemplate: test.StringPtr("$1.$2.$3"),
			defaults:      &LookupDefaults{},
		},
		"all errs": {
			errRegex: `url: <required>`,
			method:   "GET",
			url:      "",
			regex:    "[0-",
			defaults: &LookupDefaults{},
		},
		"no url doesn't fail for Lookup Defaults": {
			errRegex: `^$`,
			method:   "GET",
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
			var hadBody *string
			if tc.nilService {
				lookup = nil
			} else {
				hadBody = lookup.Body
			}

			// WHEN CheckValues is called
			err := lookup.CheckValues("")

			// THEN it err's when expected
			e := util.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			if lookup == nil {
				return
			}
			// AND RegexTemplate is nil when Regex is empty
			if lookup.RegexTemplate != nil && lookup.Regex == "" {
				t.Fatalf("RegexTemplate should be nil when Regex is empty")
			}
			// AND Body is nil when Method is GET
			if lookup.Method == "GET" && lookup.Body != nil {
				t.Fatalf("Body should be nil when Method is GET")
			}
			// AND Body is kept when Method is POST
			if lookup.Method == "POST" && hadBody != nil && lookup.Body == nil {
				t.Fatalf("Body should be kept when Method is POST")
			}
			// AND Method is uppercased
			wantMethod := strings.ToUpper(tc.method)
			if wantMethod == "" {
				wantMethod = "GET"
			}
			if lookup.Method != wantMethod {
				t.Fatalf("Method should be uppercased:\nwant: %q\ngot:  %q",
					wantMethod, lookup.Method)
			}
		})
	}
}
