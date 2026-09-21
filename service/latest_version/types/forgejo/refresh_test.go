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

// Package forgejo provides a Forgejo-based lookup type.
package forgejo

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/util"
)

func TestLookup_InheritSecrets(t *testing.T) {
	// GIVEN: a Lookup carrying edits, and the stored Lookup it is edited from.
	tests := []struct {
		name      string
		hosts     map[string]HostDefaults // Instances the defaults name.
		host      string
		token     string
		fromHost  string
		fromToken string
		fromOther bool // Inherit from a Lookup of another type.
		want      string
	}{
		{
			name:      "masked token on the same host inherits the stored one",
			host:      "https://codeberg.org",
			token:     util.SecretValue,
			fromHost:  "https://codeberg.org",
			fromToken: "stored-token",
			want:      "stored-token",
		},
		{
			name:      "masked token on an equivalent spelling of the host inherits the stored one",
			host:      "CODEBERG.org:443/",
			token:     util.SecretValue,
			fromHost:  "https://codeberg.org",
			fromToken: "stored-token",
			want:      "stored-token",
		},
		{
			name:      "masked token on a changed host is cleared",
			host:      "https://git.example.com",
			token:     util.SecretValue,
			fromHost:  "https://codeberg.org",
			fromToken: "stored-token",
			want:      "",
		},
		{
			name:      "masked placeholder on a value the service never had is cleared",
			host:      "https://codeberg.org",
			token:     util.SecretValue,
			fromHost:  "https://codeberg.org",
			fromToken: "",
			want:      "",
		},
		{
			name:      "masked token cannot be inherited from another lookup type",
			host:      "https://codeberg.org",
			token:     util.SecretValue,
			fromOther: true,
			want:      "",
		},
		{
			name:      "new token is kept",
			host:      "https://git.example.com",
			token:     "new-token",
			fromHost:  "https://codeberg.org",
			fromToken: "stored-token",
			want:      "new-token",
		},
		{
			name:      "emptied token is kept empty",
			host:      "https://codeberg.org",
			token:     "",
			fromHost:  "https://codeberg.org",
			fromToken: "stored-token",
			want:      "",
		},
		{
			name:      "masked token on the URL an entry names inherits the stored one",
			hosts:     map[string]HostDefaults{"Codeberg": {URL: "https://codeberg.org"}},
			host:      "https://codeberg.org",
			token:     util.SecretValue,
			fromHost:  "Codeberg",
			fromToken: "stored-token",
			want:      "stored-token",
		},
		{
			name:      "masked token on the entry naming the stored URL inherits the stored one",
			hosts:     map[string]HostDefaults{"Codeberg": {URL: "https://codeberg.org"}},
			host:      "Codeberg",
			token:     util.SecretValue,
			fromHost:  "https://codeberg.org",
			fromToken: "stored-token",
			want:      "stored-token",
		},
		{
			name:      "masked token on an entry naming another instance is cleared",
			hosts:     map[string]HostDefaults{"Elsewhere": {URL: "https://git.example.com"}},
			host:      "Elsewhere",
			token:     util.SecretValue,
			fromHost:  "https://codeberg.org",
			fromToken: "stored-token",
			want:      "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, test.TrimYAML(`
				host: `+tc.host+`
				url: owner/repo
			`))
			lookup.AccessToken = tc.token
			setHostDefaults(lookup, tc.hosts, nil)

			var from base.BaseInterface = testLookup(t, test.TrimYAML(`
				host: `+tc.fromHost+`
				url: owner/repo
			`))
			if fromLookup, ok := from.(*Lookup); ok {
				fromLookup.AccessToken = tc.fromToken
				setHostDefaults(fromLookup, tc.hosts, nil)
			}
			if tc.fromOther {
				from = &base.Lookup{URL: "owner/repo"}
			}

			// WHEN: InheritSecrets is called.
			lookup.InheritSecrets(from, nil)

			// THEN: the token is as expected.
			if lookup.AccessToken != tc.want {
				t.Fatalf(
					"%s\nLookup.InheritSecrets() access_token mismatch\ngot:  %q\nwant: %q",
					packageName, lookup.AccessToken, tc.want,
				)
			}

			// AND: a masked value never survives.
			if lookup.AccessToken == util.SecretValue {
				t.Fatalf(
					"%s\nLookup.InheritSecrets() left the masked placeholder in place",
					packageName,
				)
			}
		})
	}
}
