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
	"testing"
)

func TestLookup_GetAllowInvalidCerts(t *testing.T) {
	// GIVEN a Lookup
	tests := map[string]struct {
		allowInvalidCertsRoot        *bool
		allowInvalidCertsDefault     *bool
		allowInvalidCertsHardDefault *bool
		wantBool                     bool
	}{
		"root overrides all": {
			wantBool:                     true,
			allowInvalidCertsRoot:        boolPtr(true),
			allowInvalidCertsDefault:     boolPtr(false),
			allowInvalidCertsHardDefault: boolPtr(false)},
		"default overrides hardDefault": {
			wantBool:                     true,
			allowInvalidCertsRoot:        nil,
			allowInvalidCertsDefault:     boolPtr(true),
			allowInvalidCertsHardDefault: boolPtr(false)},
		"hardDefault is last resort": {
			wantBool:                     true,
			allowInvalidCertsRoot:        nil,
			allowInvalidCertsDefault:     nil,
			allowInvalidCertsHardDefault: boolPtr(true)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup()
			lookup.AllowInvalidCerts = tc.allowInvalidCertsRoot
			lookup.Defaults.AllowInvalidCerts = tc.allowInvalidCertsDefault
			lookup.HardDefaults.AllowInvalidCerts = tc.allowInvalidCertsHardDefault

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
