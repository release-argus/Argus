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
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_Marshal(t *testing.T) {
	// GIVEN: a Lookup with every marshalled field set.
	lookup := testLookup(t, `
		type: forgejo
		host: https://codeberg.org
		url: owner/repo
		use_prerelease: true
		url_commands:
			- type: regex
				regex: '([0-9.]+)'
		require:
			regex_version: '^[0-9]'`)

	// WHEN: it is marshalled.
	got := lookup.String("")

	// THEN: 'host' sits with the 'url' it completes, after 'type'.
	wantOrder := []string{"type", "host", "url", "url_commands", "require", "use_prerelease"}
	var gotOrder []string
	for line := range strings.SplitSeq(got, "\n") {
		matches := regexp.MustCompile(`^[a-z_]+:`).FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		gotOrder = append(gotOrder, strings.TrimSuffix(matches[0], ":"))
	}
	if !slices.Equal(gotOrder, wantOrder) {
		t.Fatalf(
			"%s\nLookup.String() key order mismatch\ngot:  %v\nwant: %v\n\n%s",
			packageName, gotOrder, wantOrder, got,
		)
	}

	// AND: the JSON the API serves carries the same order.
	gotJSON := decode.ToJSONString(lookup)
	var at []int
	for _, key := range wantOrder {
		i := strings.Index(gotJSON, `"`+key+`":`)
		if i == -1 {
			t.Fatalf(
				"%s\nLookup.MarshalJSON() dropped %q\ngot:  %s",
				packageName, key, gotJSON,
			)
		}
		at = append(at, i)
	}
	if !slices.IsSorted(at) {
		t.Fatalf(
			"%s\nLookup.MarshalJSON() key order mismatch\nwant: %v\ngot:  %s",
			packageName, wantOrder, gotJSON,
		)
	}
}

func TestLookup_Marshal_dropsNothing(t *testing.T) {
	// GIVEN: the fields [Lookup] marshals, taken from the struct.
	want := make([]string, 0, 8)
	for _, typ := range []reflect.Type{
		reflect.TypeFor[base.Lookup](),
		reflect.TypeFor[Lookup](),
	} {
		for field := range typ.Fields() {
			tag, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
			if tag == "" || tag == "-" {
				continue
			}
			want = append(want, tag)
		}
	}

	// WHEN: the marshal-only helper is inspected.
	auxType := reflect.TypeFor[lookupMarshal]()
	got := make(map[string]bool, auxType.NumField())
	for field := range auxType.Fields() {
		tag, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
		got[tag] = true
	}

	// THEN: it carries every field the Lookup marshals.
	for _, field := range want {
		if !got[field] {
			t.Fatalf(
				"%s\nlookupMarshal is missing %q, so it would be dropped on save\nhas: %v",
				packageName, field, got,
			)
		}
	}
}

func TestLookup_Unmarshal(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)

	// GIVEN: a Lookup with values already on it, and encoded data to unmarshal over it.
	tests := []struct {
		name              string
		format            string
		data              string
		wantHost          string
		wantUsePreRelease *bool
		errRegex          string
	}{
		{
			name:   "yaml/both fields replaced",
			format: "yaml",
			data: test.TrimYAML(`
				host: https://gitea.com
				use_prerelease: false
			`),
			wantHost:          "https://gitea.com",
			wantUsePreRelease: new(false),
			errRegex:          `^$`,
		},
		{
			name:   "json/both fields replaced",
			format: "json",
			data: test.TrimJSON(`{
				"host":"https://gitea.com",
				"use_prerelease":false
			}`),
			wantHost:          "https://gitea.com",
			wantUsePreRelease: new(false),
			errRegex:          `^$`,
		},
		{
			name:              "yaml/omitted fields are kept",
			format:            "yaml",
			data:              "url: owner/repo",
			wantHost:          "https://codeberg.org",
			wantUsePreRelease: new(true),
			errRegex:          `^$`,
		},
		{
			name:              "yaml/no data leaves the receiver untouched",
			format:            "yaml",
			data:              "",
			wantHost:          "https://codeberg.org",
			wantUsePreRelease: new(true),
			errRegex:          `^$`,
		},
		{
			name:     "yaml/invalid",
			format:   "yaml",
			data:     "host: [",
			errRegex: `sequence end token ']' not found`,
		},
		{
			name:     "json/invalid",
			format:   "json",
			data:     `{"host":`,
			errRegex: `unexpected EOF|invalid`,
		},
		{
			name:   "yaml/require that cannot be decoded",
			format: "yaml",
			data:   "require: 123",
			errRegex: test.TrimYAML(`
				^require:
					extract "docker":`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{
				Host:          "https://codeberg.org",
				UsePreRelease: new(true),
				Defaults:      lvCfg.Soft,
				HardDefaults:  lvCfg.Hard,
			}

			// WHEN: it is unmarshalled over.
			var err error
			if tc.format == "json" {
				err = lookup.UnmarshalJSON([]byte(tc.data))
			} else {
				err = lookup.UnmarshalYAML([]byte(tc.data))
			}

			prefix := fmt.Sprintf(
				"%s\nLookup.Unmarshal%s(%q)",
				packageName, strings.ToUpper(tc.format), tc.data,
			)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if err != nil {
				return
			}

			// AND: the fields are as expected.
			if lookup.Host != tc.wantHost {
				t.Fatalf(
					"%s host mismatch\ngot:  %q\nwant: %q",
					prefix, lookup.Host, tc.wantHost,
				)
			}
			if g, w := test.StringifyPtr(lookup.UsePreRelease), test.StringifyPtr(tc.wantUsePreRelease); g != w {
				t.Fatalf(
					"%s use_prerelease mismatch\ngot:  %s\nwant: %s",
					prefix, g, w,
				)
			}
		})
	}
}

func TestLookup_String(t *testing.T) {
	// GIVEN: a Lookup.
	want := test.TrimYAML(`
		type: forgejo
		host: HTTPS://CodeBerg.org
		url: owner/repo
		use_prerelease: true
	`)
	lookup := testLookup(t, want)

	// WHEN: it is stringified.
	got := lookup.String("")

	// THEN: it round-trips the host exactly as it was written.
	if got != want {
		t.Fatalf(
			"%s\nLookup.String() mismatch\ngot:  %q\nwant: %q",
			packageName, got, want,
		)
	}
}

func TestLookup_Clone(t *testing.T) {
	// GIVEN: a Lookup.
	lookup := testLookup(t, `
		type: forgejo
		host: https://codeberg.org
		url: owner/repo
		use_prerelease: true`,
	)

	// WHEN: it is cloned.
	got := lookup.Clone(lookup.Status)

	// THEN: the clone matches.
	if g, w := got.String(""), lookup.String(""); g != w {
		t.Fatalf(
			"%s\nLookup.Clone() mismatch\ngot:  %q\nwant: %q",
			packageName, g, w,
		)
	}

	// AND: use_prerelease is a copy, not the same pointer.
	if got.UsePreRelease == lookup.UsePreRelease {
		t.Fatalf("%s\nLookup.Clone() shares the use_prerelease pointer", packageName)
	}

	// AND: changing the clone leaves the original unchanged.
	got.Host = "https://gitea.com"
	if want := "https://codeberg.org"; lookup.Host != want {
		t.Fatalf(
			"%s\nLookup.Clone() shares the host\ngot:  %q\nwant: %q",
			packageName, lookup.Host, want,
		)
	}
}

func TestLookup_Clone__Nil(t *testing.T) {
	// GIVEN: a nil Lookup.
	var lookup *Lookup

	// WHEN: it is cloned/copied.
	// THEN: nothing comes back, and nothing panics.
	if got := lookup.Clone(nil); got != nil {
		t.Fatalf(
			"%s\nLookup.Clone() on nil mismatch\ngot:  %v\nwant: nil",
			packageName, got,
		)
	}
	if got := lookup.Copy(nil); got != nil {
		t.Fatalf(
			"%s\nLookup.Copy() on nil mismatch\ngot:  %v\nwant: nil",
			packageName, got,
		)
	}
}

func TestLookup_Copy(t *testing.T) {
	// GIVEN: a Lookup.
	lookup := testLookup(t, `
		type: forgejo
		host: https://codeberg.org
		url: owner/repo
		use_prerelease: true`,
	)

	// WHEN: it is copied as a base.Interface.
	got := lookup.Copy(lookup.Status)

	// THEN: the copy matches, and is a distinct Lookup.
	if g, w := got.String(""), lookup.String(""); g != w {
		t.Fatalf(
			"%s\nLookup.Copy() mismatch\ngot:  %q\nwant: %q",
			packageName, g, w,
		)
	}
	copied, ok := got.(*Lookup)
	if !ok {
		t.Fatalf(
			"%s\nLookup.Copy() type mismatch\ngot:  %T\nwant: *Lookup",
			packageName, got,
		)
	}
	if copied == lookup {
		t.Fatalf(
			"%s\nLookup.Copy() returned the receiver rather than a copy",
			packageName,
		)
	}
}

func TestLookup_TypeDefaults(t *testing.T) {
	// GIVEN: a Lookup and a pair of type-specific Defaults.
	lookup := &Lookup{}
	defaults := &Defaults{}
	hardDefaults := &Defaults{}
	hardDefaults.Default()

	// WHEN: they are assigned.
	lookup.SetTypeDefaults(defaults, hardDefaults)
	gotDefaults, gotHardDefaults := lookup.GetTypeDefaults()

	// THEN: they come back unchanged.
	if gotDefaults != defaults || gotHardDefaults != hardDefaults {
		t.Fatalf(
			"%s\nLookup.GetTypeDefaults() mismatch\ngot:  (%p, %p)\nwant: (%p, %p)",
			packageName,
			gotDefaults, gotHardDefaults,
			defaults, hardDefaults,
		)
	}
}
