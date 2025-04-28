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
)

func TestDefaults_String(t *testing.T) {
	// GIVEN a Defaults.
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
				"foo",
				"discord",
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

func TestSliceDefaults_String(t *testing.T) {
	// GIVEN a Slice.
	tests := map[string]struct {
		slice *SliceDefaults
		want  string
	}{
		"nil": {
			slice: nil,
			want:  "",
		},
		"empty": {
			slice: &SliceDefaults{},
			want:  "{}",
		},
		"one element": {
			slice: &SliceDefaults{
				"foo": NewDefaults(
					"discord",
					nil, nil, nil)},
			want: test.TrimYAML(`
				foo:
					type: discord`),
		},
		"multiple elements": {
			slice: &SliceDefaults{
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

				// WHEN the Slice is stringified with String.
				got := tc.slice.String(prefix)

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

func TestSlice_String(t *testing.T) {
	// GIVEN a Slice.
	tests := map[string]struct {
		slice *Slice
		want  string
	}{
		"nil": {
			slice: nil,
			want:  "",
		},
		"empty": {
			slice: &Slice{},
			want:  "{}",
		},
		"one element": {
			slice: &Slice{
				"foo": New(
					nil, "",
					"discord",
					nil, nil, nil, nil, nil, nil)},
			want: test.TrimYAML(`
				foo:
					type: discord`),
		},
		"multiple elements": {
			slice: &Slice{
				"foo": New(
					nil, "",
					"discord",
					nil, nil, nil, nil, nil, nil),
				"bar": New(
					nil, "",
					"gotify",
					nil, nil, nil, nil, nil, nil),
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

				// WHEN the Slice is stringified with String.
				got := tc.slice.String(prefix)

				// THEN the result is as expected.
				if got != want {
					t.Errorf("%s\n(prefix=%q) mismatch\nwant: %q\ngot:  %q",
						packageName, prefix, want, got)
				}
			}
		})
	}
}
