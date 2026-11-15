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

package github

import (
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestDefaults_IsZero(t *testing.T) {
	tests := []struct {
		name string
		data *Defaults
		want bool
	}{
		{
			name: "empty",
			data: &Defaults{},
			want: true,
		},
		{
			name: "non-empty/AccessToken",
			data: &Defaults{
				AccessToken: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/UsePreRelease",
			data: &Defaults{
				UsePreRelease: new(true),
			},
			want: false,
		},
		{
			name: "non-empty/all",
			data: &Defaults{
				AccessToken:   "foo",
				UsePreRelease: new(true),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.data.IsZero()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %t\nwant: %t", got, tc.want)
			}
		})
	}
}

func TestDefaults_Default(t *testing.T) {
	defaults := Defaults{}
	want := Defaults{
		UsePreRelease: new(false),
	}

	defaults.Default()

	if defaults.AccessToken != want.AccessToken {
		t.Errorf(
			"Defaults.AccessToken mismatch\ngot:  %q\nwant: %q",
			defaults.AccessToken, want.AccessToken,
		)
	}
	if defaults.UsePreRelease == nil || *defaults.UsePreRelease != *want.UsePreRelease {
		t.Errorf(
			"Defaults.UsePreRelease mismatch\ngot:  %s\nwant: %s",
			test.StringifyPtr(defaults.UsePreRelease), test.StringifyPtr(want.UsePreRelease),
		)
	}
}
