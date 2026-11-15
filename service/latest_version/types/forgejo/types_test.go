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
	"maps"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	lvbasetest "github.com/release-argus/Argus/service/latest_version/types/base/test"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestLookup_Marshal(t *testing.T) {
	lookup := testLookup(t, `
		type: forgejo
		host: https://codeberg.org
		url: owner/repo
		use_prerelease: true
		url_commands:
			- type: regex
				regex: '([0-9.]+)'
		require:
			regex_version: '^[0-9]'
		access_token: token
		allow_invalid_certs: true`)

	got := lookup.String("")

	wantOrder := []string{
		"type", "host", "url", "url_commands", "require",
		"access_token", "allow_invalid_certs", "use_prerelease",
	}
	var gotOrder []string
	keyRe := regexp.MustCompile(`^[a-z_]+:`)
	for line := range strings.SplitSeq(got, "\n") {
		matches := keyRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		gotOrder = append(gotOrder, strings.TrimSuffix(matches[0], ":"))
	}
	if !slices.Equal(gotOrder, wantOrder) {
		t.Fatalf(
			"Lookup.String() key order mismatch\ngot:  %v\nwant: %v\n\n%s",
			gotOrder, wantOrder, got,
		)
	}

	gotJSON := decode.ToJSONString(lookup)
	var at []int
	for _, key := range wantOrder {
		i := strings.Index(gotJSON, `"`+key+`":`)
		if i == -1 {
			var fields map[string]any
			if err := decode.Unmarshal("json", []byte(gotJSON), &fields); err != nil {
				t.Fatalf("decode.Unmarshal() failed: %v", err)
			}
			t.Fatalf(
				"decode.ToJSONString() key %q mismatch\ngot:  absent (keys=%q)\nwant: present",
				key, util.SortedKeys(fields),
			)
		}
		at = append(at, i)
	}
	if !slices.IsSorted(at) {
		t.Fatalf("decode.ToJSONString() key order mismatch\ngot:  %s\nwant: %v", gotJSON, wantOrder)
	}
}

func TestLookup_Marshal_dropsNothing(t *testing.T) {
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

	auxType := reflect.TypeFor[lookupMarshal]()
	got := make(map[string]bool, auxType.NumField())
	for field := range auxType.Fields() {
		tag, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
		got[tag] = true
	}

	for _, field := range want {
		if !got[field] {
			t.Fatalf(
				"lookupMarshal field %q mismatch\ngot:  absent (fields=%q)\nwant: present",
				field, slices.Sorted(maps.Keys(got)),
			)
		}
	}
}

func TestLookup_Unmarshal(t *testing.T) {
	lvCfg := lvbasetest.PlainDefaultsConfig(t)

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

			var err error
			if tc.format == "json" {
				err = lookup.UnmarshalJSON([]byte(tc.data))
			} else {
				err = lookup.UnmarshalYAML([]byte(tc.data))
			}

			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if lookup.Host != tc.wantHost {
				t.Fatalf("Lookup.Host mismatch\ngot:  %q\nwant: %q", lookup.Host, tc.wantHost)
			}
			if g, w := test.StringifyPtr(lookup.UsePreRelease), test.StringifyPtr(tc.wantUsePreRelease); g != w {
				t.Fatalf("Lookup.UsePreRelease mismatch\ngot:  %s\nwant: %s", g, w)
			}
		})
	}
}

func TestLookup_String(t *testing.T) {
	// Mixed case, so String() must keep the host exactly as written.
	want := test.TrimYAML(`
		type: forgejo
		host: HTTPS://CodeBerg.org
		url: owner/repo
		use_prerelease: true
	`)
	lookup := testLookup(t, want)

	got := lookup.String("")

	if got != want {
		t.Fatalf("Lookup.String() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestLookup_Copy(t *testing.T) {
	lvCfg := lvbasetest.PlainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	tests := []struct {
		name   string
		lookup *Lookup
		status *status.Status
	}{
		{
			name:   "nil",
			lookup: nil,
			status: nil,
		},
		{
			name: "filled",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: test
						host: https://codeberg.org
						url: owner/repo
						access_token: dummy-copy-token
						allow_invalid_certs: true
						use_prerelease: true
					`)),
					test.Must(t, func() (*opt.Options, error) {
						return opt.Decode(
							"yaml", []byte(test.TrimYAML(`
								active: false
								interval: 2s
								semantic_versioning: false
							`)),
							optCfg,
						)
					}),
					svcStatus,
					lvCfg,
				)
			}),
			status: test.Must(t, func() (*status.Status, error) {
				return statustest.New("yaml", nil)
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wantStr := decode.ToYAMLString(tc.lookup, "")

			gotInterface := tc.lookup.Copy(tc.status)

			if tc.lookup == nil {
				if gotInterface != nil {
					t.Errorf("copy mismatch\ngot:  non-nil (%T)\nwant: nil", gotInterface)
				}
				return
			}

			if gotInterface == nil {
				t.Fatal("copy mismatch\ngot:  nil\nwant: non-nil")
			}

			if gotInterface == tc.lookup {
				t.Fatalf(
					"copy aliases the original\ngot:  %p\nwant: NOT %p",
					gotInterface, tc.lookup,
				)
			}

			got, ok := gotInterface.(*Lookup)
			if !ok {
				t.Fatalf("copy type mismatch\ngot:  %T\nwant: *Lookup", gotInterface)
			}

			if gotStr := got.String(""); gotStr != wantStr {
				t.Fatalf("Lookup.String() mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Type", Got: got.Type, Want: tc.lookup.Type, Mode: test.CompareEqual},
				{Name: "URL", Got: got.URL, Want: tc.lookup.URL, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "AllowInvalidCerts", Got: got.AllowInvalidCerts, Want: tc.lookup.AllowInvalidCerts, Mode: test.CopiedPointerMode(tc.lookup.AllowInvalidCerts)},
				{Name: "Require", Got: got.Require, Want: tc.lookup.Require, Mode: test.CopiedPointerMode(tc.lookup.Require)},
				{Name: "Options", Got: got.Options, Want: tc.lookup.Options, Mode: test.CopiedPointerMode(tc.lookup.Options)},
				{Name: "Status", Got: got.Status, Want: tc.status, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "AccessToken", Got: got.AccessToken, Want: tc.lookup.AccessToken, Mode: test.CompareEqual},
				{Name: "UsePreRelease", Got: got.UsePreRelease, Want: tc.lookup.UsePreRelease, Mode: test.CopiedPointerMode(tc.lookup.UsePreRelease)},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Defaults", Got: got.Defaults, Want: tc.lookup.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: tc.lookup.HardDefaults, Mode: test.CompareSamePointer},
				{Name: "typeDefaults", Got: got.typeDefaults, Want: tc.lookup.typeDefaults, Mode: test.CompareSamePointer},
				{Name: "typeHardDefaults", Got: got.typeHardDefaults, Want: tc.lookup.typeHardDefaults, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestLookup_TypeDefaults(t *testing.T) {
	lookup := &Lookup{}
	defaults := &Defaults{}
	hardDefaults := &Defaults{}
	hardDefaults.Default()

	lookup.SetTypeDefaults(defaults, hardDefaults)
	gotDefaults, gotHardDefaults := lookup.GetTypeDefaults()

	if gotDefaults != defaults {
		t.Errorf(
			"Lookup.GetTypeDefaults() defaults mismatch\ngot:  %p\nwant: %p",
			gotDefaults, defaults,
		)
	}
	if gotHardDefaults != hardDefaults {
		t.Errorf(
			"Lookup.GetTypeDefaults() hardDefaults mismatch\ngot:  %p\nwant: %p",
			gotHardDefaults, hardDefaults,
		)
	}
}
