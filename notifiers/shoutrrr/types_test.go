// Copyright [2022] [Argus]
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

package shoutrrr

import (
	"strings"
	"testing"

	svcstatus "github.com/release-argus/Argus/service/status"
)

func TestShoutrrr_String(t *testing.T) {
	tests := map[string]struct {
		shoutrrr *Shoutrrr
		want     string
	}{
		"nil": {
			shoutrrr: nil,
			want:     "<nil>"},
		"empty": {
			shoutrrr: &Shoutrrr{},
			want:     "{}\n"},
		"all fields defined": {
			shoutrrr: &Shoutrrr{
				Type: "discord",
				ID:   "foo",
				Failed: &map[string]*bool{
					"foo": boolPtr(true)},
				ServiceStatus: &svcstatus.Status{
					LatestVersion: "1.2.3"},
				Options: map[string]string{
					"delay": "1h"},
				URLFields: map[string]string{
					"webhookid": "456"},
				Params: map[string]string{
					"title": "argus"},
				Main: &Shoutrrr{
					URLFields: map[string]string{
						"token": "bar"}},
				Defaults: &Shoutrrr{
					Options: map[string]string{
						"delay": "2h"}},
				HardDefaults: &Shoutrrr{
					Options: map[string]string{
						"delay": "3h"}},
			},
			want: `
type: discord
options:
    delay: 1h
url_fields:
    webhookid: "456"
params:
    title: argus
`},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Shoutrrr is stringified with String
			got := tc.shoutrrr.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}

func TestSlice_String(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice *Slice
		want  string
	}{
		"nil": {
			slice: nil,
			want:  "<nil>",
		},
		"empty": {
			slice: &Slice{},
			want:  "{}\n",
		},
		"one element": {
			slice: &Slice{
				"foo": &Shoutrrr{
					Type: "discord"}},
			want: `
foo:
    type: discord
`,
		},
		"multiple elements": {
			slice: &Slice{
				"foo": &Shoutrrr{
					Type: "discord"},
				"bar": &Shoutrrr{
					Type: "gotify"},
			},
			want: `
bar:
    type: gotify
foo:
    type: discord
`,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN the Slice is stringified with String
			got := tc.slice.String()

			// THEN the result is as expected
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Errorf("got:\n%q\nwant:\n%q",
					got, tc.want)
			}
		})
	}
}
