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
	"testing"

	"github.com/release-argus/Argus/util"
)

func TestLookup_CheckValues(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		url        string
		json       string
		regex      string
		defaults   *LookupDefaults
		errRegex   string
		nilService bool
	}{
		"nil service": {
			errRegex:   `^$`,
			nilService: true,
		},
		"valid service": {
			errRegex: `^$`,
			url:      "https://example.com",
			regex:    "[0-9.]+",
			defaults: &LookupDefaults{},
		},
		"no url": {
			errRegex: `url: <required>`,
			url:      "",
			defaults: &LookupDefaults{},
		},
		"invalid json - string in square brackets": {
			errRegex: `json: .* <invalid>`,
			json:     "foo[bar]",
			defaults: &LookupDefaults{},
		},
		"invalid regex": {
			errRegex: `regex: .* <invalid>`,
			regex:    "[0-",
			defaults: &LookupDefaults{},
		},
		"all errs": {
			errRegex: `url: <required>`,
			url:      "",
			regex:    "[0-",
			defaults: &LookupDefaults{},
		},
		"no url doesnt fail for Lookup Defaults": {
			errRegex: `^$`,
			url:      "",
			defaults: nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{}
			lookup = testLookup()
			lookup.URL = tc.url
			lookup.JSON = tc.json
			lookup.Regex = tc.regex
			lookup.Defaults = nil
			if tc.defaults != nil {
				lookup.Defaults = tc.defaults
			}
			if tc.nilService {
				lookup = nil
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
		})
	}
}
