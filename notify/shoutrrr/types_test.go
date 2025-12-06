// Copyright [2025] [Argus]
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

	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestDefaults_String(t *testing.T) {
	// GIVEN Defaults.
	tests := map[string]struct {
		shoutrrr *Defaults
		want     string
	}{
		"nil": {
			shoutrrr: nil,
			want:     ""},
		"empty": {
			shoutrrr: &Defaults{},
			want:     "{}"},
		"all fields defined": {
			shoutrrr: NewDefaults(
				"discord",
				map[string]string{
					"delay": "1h"},
				map[string]string{
					"webhookid": "456"},
				map[string]string{
					"title": "argus"}),
			want: test.TrimYAML(`
				type: discord
				options:
					delay: 1h
				url_fields:
					webhookid: "456"
				params:
					title: argus`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Shoutrrr is stringified with String.
				got := tc.shoutrrr.String(prefix)

				// THEN the result is as expected when stringified.
				if got != want {
					t.Fatalf("%s\n(prefix=%q) mismatch\nwant: %q\ngot:  %q",
						packageName, prefix, want, got)
				}
			}
		})
	}
}

func TestShoutrrr_String(t *testing.T) {
	// GIVEN a Shoutrrr.
	tests := map[string]struct {
		shoutrrr      *Shoutrrr
		latestVersion string
		want          string
	}{
		"nil": {
			shoutrrr: nil,
			want:     ""},
		"empty": {
			shoutrrr: &Shoutrrr{},
			want:     "{}"},
		"all fields defined": {
			latestVersion: "1.2.3",
			shoutrrr: New(
				nil,
				"foo", "discord",
				map[string]string{
					"delay": "1h"},
				map[string]string{
					"webhookid": "456"},
				map[string]string{
					"title": "argus"},
				NewDefaults(
					"", nil,
					map[string]string{
						"token": "bar"},
					nil),
				NewDefaults(
					"",
					map[string]string{
						"delay": "2h"},
					nil, nil),
				NewDefaults(
					"",
					map[string]string{
						"delay": "3h"},
					nil, nil)),
			want: test.TrimYAML(`
				type: discord
				options:
					delay: 1h
				url_fields:
					webhookid: "456"
				params:
					title: argus`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.latestVersion != "" {
				if tc.shoutrrr.ServiceStatus == nil {
					tc.shoutrrr.ServiceStatus = status.New(
						nil, nil, nil,
						"",
						"", "",
						"", "",
						"",
						&dashboard.Options{})
				}
				tc.shoutrrr.ServiceStatus.SetLatestVersion(tc.latestVersion, "", false)
			}
			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := test.TrimYAML(tc.want)
				want = strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Shoutrrr is stringified with String.
				got := tc.shoutrrr.String(prefix)

				// THEN the result is as expected.
				if got != want {
					t.Errorf("%s\n(prefix=%q) mismatch\nwant: %q\ngot:  %q",
						packageName, prefix, want, got)
				}
			}
		})
	}
}

func TestShoutrrsDefaults_String(t *testing.T) {
	// GIVEN Shoutrrrs.
	tests := map[string]struct {
		shoutrrrsDefaults *ShoutrrrsDefaults
		want              string
	}{
		"nil": {
			shoutrrrsDefaults: nil,
			want:              "",
		},
		"empty": {
			shoutrrrsDefaults: &ShoutrrrsDefaults{},
			want:              "{}",
		},
		"one element": {
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"discord",
					nil, nil, nil)},
			want: test.TrimYAML(`
				foo:
					type: discord`),
		},
		"multiple elements": {
			shoutrrrsDefaults: &ShoutrrrsDefaults{
				"foo": NewDefaults(
					"discord",
					nil, nil, nil),
				"bar": NewDefaults(
					"gotify",
					nil, nil, nil),
			},
			want: test.TrimYAML(`
				bar:
					type: gotify
				foo:
					type: discord`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := test.TrimYAML(tc.want)
				want = strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Shoutrrrs is stringified with String.
				got := tc.shoutrrrsDefaults.String(prefix)

				// THEN the result is as expected.
				want = strings.TrimPrefix(want, "\n")
				if got != want {
					t.Fatalf("%s\n(prefix=%q) mismatch\nwant: %q\ngot:  %q",
						packageName, prefix, want, got)
				}
			}
		})
	}
}

func TestShoutrrrs_String(t *testing.T) {
	// GIVEN Shoutrrrs.
	tests := map[string]struct {
		shoutrrrs *Shoutrrrs
		want      string
	}{
		"nil": {
			shoutrrrs: nil,
			want:      "",
		},
		"empty": {
			shoutrrrs: &Shoutrrrs{},
			want:      "{}",
		},
		"one element": {
			shoutrrrs: &Shoutrrrs{
				"foo": New(
					nil,
					"", "discord",
					nil, nil, nil,
					nil, nil, nil)},
			want: test.TrimYAML(`
				foo:
					type: discord`),
		},
		"multiple elements": {
			shoutrrrs: &Shoutrrrs{
				"foo": New(
					nil,
					"", "discord",
					nil, nil, nil,
					nil, nil, nil),
				"bar": New(
					nil,
					"", "gotify",
					nil, nil, nil,
					nil, nil, nil),
			},
			want: test.TrimYAML(`
				bar:
					type: gotify
				foo:
					type: discord`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				want := test.TrimYAML(tc.want)
				want = strings.TrimPrefix(tc.want, "\n")
				if want != "" {
					if want != "{}" {
						want = prefix + strings.ReplaceAll(want, "\n", "\n"+prefix)
					}
					want += "\n"
				}

				// WHEN the Shoutrrrs is stringified with String.
				got := tc.shoutrrrs.String(prefix)

				// THEN the result is as expected.
				if got != want {
					t.Errorf("%s\n(prefix=%q) mismatch\nwant: %q\ngot:  %q",
						packageName, prefix, want, got)
				}
			}
		})
	}
}

func TestShoutrrrs_UnmarshalJSON(t *testing.T) {
	// GIVEN various JSON inputs to unmarshal into Shoutrrrs.
	tests := map[string]struct {
		json     string
		wantErr  string
		wantKeys map[string]string
	}{
		"valid array with two items": {
			json: test.TrimJSON(`[
				{"name": "a", "type": "slack"},
				{"name": "b", "type": "gotify"}
			]`),
			wantErr: `^$`,
			wantKeys: map[string]string{
				"a": "slack",
				"b": "gotify",
			},
		},
		"empty array becomes empty map": {
			json:     `[]`,
			wantErr:  `^$`,
			wantKeys: map[string]string{},
		},
		"null becomes empty map": {
			json:     `null`,
			wantErr:  `^$`,
			wantKeys: map[string]string{},
		},
		"duplicate ids - last wins": {
			json: test.TrimJSON(`[
				{"name": "dupe", "type": "slack"},
				{"name": "dupe", "type": "gotify"}
			]`),
			wantErr: `^$`,
			wantKeys: map[string]string{
				"dupe": "gotify",
			},
		},
		"invalid JSON": {
			json:    `{`,
			wantErr: `.+`,
		},
		"wrong shape (object instead of array)": {
			json: test.TrimJSON(`{
				"name": "a", "type": "slack"
			}`),
			wantErr: `json: cannot unmarshal object.+$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN unmarshaling JSON into a Shoutrrrs.
			var s Shoutrrrs
			err := s.UnmarshalJSON([]byte(tc.json))

			// THEN errors produced match the regex.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.wantErr, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot: %q", packageName, tc.wantErr, err.Error())
			}
			if e != "" {
				return
			}

			// AND map keys and types are as expected.
			if len(s) != len(tc.wantKeys) {
				t.Fatalf("%s\nlength mismatch\nwant: %d\ngot: %d", packageName, len(tc.wantKeys), len(s))
			}
			for id, wantType := range tc.wantKeys {
				got, ok := s[id]
				if !ok {
					t.Errorf("%s\nmissing key %q", packageName, id)
				}
				if got == nil {
					t.Errorf("%s\nvalue for key %q is nil", packageName, id)
				}
				if got.Type != wantType {
					t.Errorf("%s\nType mismatch for %q\nwant: %q\n got: %q", packageName, id, wantType, got.Type)
				}
				if got.ID != id {
					t.Errorf("%s\nID mismatch for key %q\nwant: %q\n got: %q", packageName, id, id, got.ID)
				}
			}
		})
	}
}

func TestShoutrrrs_MarshalJSON(t *testing.T) {
	// GIVEN various Shoutrrrs states to marshal.
	tests := map[string]struct {
		shoutrrrs *Shoutrrrs
		wantStr   string
	}{
		"nil map -> null": {
			shoutrrrs: nil,
			wantStr:   "null",
		},
		"empty map -> empty array": {
			shoutrrrs: &Shoutrrrs{},
			wantStr:   "[]",
		},
		"two items": {
			shoutrrrs: func() *Shoutrrrs {
				m := Shoutrrrs{
					"a": New(
						nil,
						"a", "slack",
						nil, nil, nil,
						nil, nil, nil),
					"b": New(nil,
						"b", "gotify",
						nil, nil, nil,
						nil, nil, nil),
				}
				return &m
			}(),
			wantStr: test.TrimJSON(`[
				{"type": "slack", "name": "a"},
				{"type": "gotify", "name": "b"}
			]`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN marshaling the Shoutrrrs.
			data, err := tc.shoutrrrs.MarshalJSON()
			if err != nil {
				t.Fatalf("%s\nMarshalJSON returned error: %v", packageName, err)
			}

			// THEN the result matches the expected JSON.
			dataStr := string(data)
			if dataStr != tc.wantStr {
				t.Errorf("%s\nJSON mismatch\nwant: %q\ngot:  %q",
					packageName, tc.wantStr, dataStr)
			}
		})
	}
}
