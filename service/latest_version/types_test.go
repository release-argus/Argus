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

package latestver

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/latest_version/types/web"
)

func TestIsEqual(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	// GIVEN: two Lookups.
	tests := []struct {
		name string
		a, b Lookup
		want bool
	}{
		{
			name: "empty",
			a:    &web.Lookup{},
			b:    &web.Lookup{},
			want: true,
		},
		{
			name: "defaults ignored",
			a: &web.Lookup{
				Lookup: base.Lookup{
					Defaults: &base.Defaults{
						AllowInvalidCerts: test.Ptr(false),
					},
				},
			},
			b:    &web.Lookup{},
			want: true,
		},
		{
			name: "hard_defaults ignored",
			a: &web.Lookup{
				Lookup: base.Lookup{
					Defaults: &base.Defaults{
						AllowInvalidCerts: test.Ptr(false),
					},
				},
			},
			b:    &web.Lookup{},
			want: true,
		},
		{
			name: "equal - url",
			a: test.Must(t, func() (Lookup, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						url: https://example.com
						allow_invalid_certs: false
						url_commands:
							- type: split
								text: v
						require:
							regex_version: v([0-9.]+)
					`)),
					nil,
					nil,
					lvCfg,
				)
			}),
			b: test.Must(t, func() (Lookup, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						url: https://example.com
						allow_invalid_certs: false
						url_commands:
							- type: split
								text: v
						require:
							regex_version: v([0-9.]+)
					`)),
					nil,
					nil,
					lvCfg,
				)
			}),
			want: true,
		},
		{
			name: "equal - github",
			a: test.Must(t, func() (Lookup, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						access_token: token
						url_commands:
							- type: split
								text: v
						require:
							regex_version: v([0-9.]+)
					`)),
					nil,
					nil,
					lvCfg,
				)
			}),
			b: test.Must(t, func() (Lookup, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
						access_token: token
						url_commands:
							- type: split
								text: v
						require:
							regex_version: v([0-9.]+)
					`)),
					nil,
					nil,
					lvCfg,
				)
			}),
			want: true,
		},
		{
			name: "not equal",
			a: test.Must(t, func() (Lookup, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					nil,
					lvCfg,
				)
			}),
			b: test.Must(t, func() (Lookup, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: url
						url: release-argus/ARGUS
					`)),
					nil,
					nil,
					lvCfg,
				)
			}),
			want: false,
		},
		{
			name: "not equal with nil",
			a: test.Must(t, func() (Lookup, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: github
						url: `+test.ArgusGitHubRepo+`
					`)),
					nil,
					nil,
					lvCfg,
				)
			}),
			b:    nil,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the two Lookups are compared.
			got := IsEqual(tc.a, tc.b)

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nIsEqual() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}
