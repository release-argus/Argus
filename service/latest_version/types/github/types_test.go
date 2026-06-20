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

// Package github provides a github-based lookup type.
package github

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	ghtypes "github.com/release-argus/Argus/service/latest_version/types/github/api_type"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

// ############
// # DECODING #
// ############

func TestLookup_Unmarshal(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)
	// GIVEN: data to unmarshal into a Lookup.
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"url": "https://example.com",
				"allow_invalid_certs": true,
				"use_prerelease": true,
				"url_commands": [
					{
						"type": "regex",
						"regex": "foo"
					}
				],
				"require": {
					"regex_version": "v.+"
				}
			}`),
			want: test.TrimYAML(`
				require:
					regex_version: v.+
				use_prerelease: true
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				url: https://example.com
				allow_invalid_certs: true
				use_prerelease: true
				url_commands:
					- type: regex
						regex: foo
				require:
					regex_version: v.+
			`),
			want: test.TrimYAML(`
				require:
					regex_version: v.+
				use_prerelease: true
			`),
			errRegex: `^$`,
		},
		{
			name:     "JSON/invalid data types",
			format:   "json",
			data:     `{"use_prerelease": ["https://example.com"]}`,
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name:   "YAML/invalid data types",
			format: "yaml",
			data:   `use_prerelease: [https://example.com]`,
			errRegex: test.TrimYAML(`
				^[^\s]+ .*unmarshal.*
				[^\s]+.* use_prerelease:.*
				\s+\^$`,
			),
		},
		{
			name:     "JSON/invalid formatting",
			format:   "json",
			data:     `{"url": "https://example.com"`,
			errRegex: `unexpected`,
		},
		{
			name:   "Require - error",
			format: "yaml",
			data: test.TrimYAML(`
				url: owner/repo
				use_prerelease: true
				url_commands:
					- type: regex
						regex: foo
				require:
					regex_version: [ v.+ ]
			`),
			want: test.TrimYAML(`
				url: owner/repo
				url_commands:
					- type: regex
						regex: foo
				use_prerelease: true
			`),
			errRegex: test.TrimYAML(`
				^require:
					[^\s]+ .*unmarshal .*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					v := Lookup{
						Lookup: base.Lookup{
							Defaults:     lvCfg.Soft,
							HardDefaults: lvCfg.Hard,
						},
					}
					err := decode.Unmarshal(format, data, &v)
					return &v, err
				},
				tc.format, tc.data,
				func(t *Lookup) string { return t.String("") },
				tc.want,
				tc.errRegex,
				packageName,
				"Lookup",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestLookup_String(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	// GIVEN: a Lookup.
	tests := []struct {
		name   string
		lookup *Lookup
		want   string
	}{
		{
			name:   "empty",
			lookup: &Lookup{},
			want:   "{}\n",
		},
		{
			name: "filled",
			lookup: test.Must(t, func() (*Lookup, error) {
				options, _ := opt.Decode(
					"yaml", []byte("interval: 1h2m3s"),
					optCfg,
				)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						access_token: token
						require:
							regex_content: foo.tar.gz
						url: `+test.ArgusGitHubRepo+`
						url_commands:
							- type: regex
								regex: v([0-9.]+)
						use_prerelease: true
					`)),
					options,
					nil,
					lvCfg,
				)
			}),
			want: test.TrimYAML(`
				url: ` + test.ArgusGitHubRepo + `
				url_commands:
					- type: regex
						regex: v([0-9.]+)
				require:
					regex_content: foo.tar.gz
				access_token: token
				use_prerelease: true
			`),
		},
		{
			name: "quotes otherwise invalid YAML strings",
			lookup: test.Must(t, func() (*Lookup, error) {
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						access_token: ">123"
						url_commands:
							- type: regex
								regex: '{2}([0-9.]+)'
					`)),
					nil,
					nil,
					lvCfg,
				)
			}),
			want: test.TrimYAML(`
				url_commands:
					- type: regex
						regex: '{2}([0-9.]+)'
				access_token: '>123'
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				tc.lookup.String,
				tc.want,
			)
		})
	}
}

// #########
// # STATE #
// #########

func TestLookup_Copy(t *testing.T) {
	lvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	// GIVEN: a Lookup.
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
			name: "data",
			lookup: &Lookup{
				data: Data{
					eTag:    "foo",
					perPage: 1,
					releases: []ghtypes.Release{
						{URL: "example.com"},
					},
					tagFallback: true,
				},
				Lookup: base.Lookup{
					Status: test.Must(t, func() (*status.Status, error) {
						return statustest.New("yaml", nil)
					}),
				},
			},
			status: nil,
		},
		{
			name: "filled",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				lv, err := Decode(
					"yaml", []byte(test.TrimYAML(`
						type: test
						access_token: token
						use_prerelease: true
						require:
							regex_version: '^1.*'
							docker:
								image: foo
								tag: bar
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
				if err == nil {
					lv.data = Data{
						eTag:    "foo",
						perPage: 1,
						releases: []ghtypes.Release{
							{URL: "example.com"},
						},
						tagFallback: true,
					}
				}

				return lv, err
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

			// WHEN: Copy() is called on it.
			gotInterface := tc.lookup.Copy(tc.status)

			prefix := fmt.Sprintf(
				"%s\nLookup.Copy(status=%p)",
				packageName, tc.status,
			)

			// THEN: if nil was copied, we get nil.
			if tc.lookup == nil {
				if gotInterface != nil {
					t.Errorf(
						"%s of nil mismatch\ngot:  %v\nwant: nil",
						prefix, gotInterface,
					)
				}
				return
			}

			// AND: the copy is non-nil.
			if gotInterface == nil {
				t.Fatalf("%s got nil want non-nil", prefix)
			}

			// AND: the copy is distinct.
			if gotInterface == tc.lookup {
				t.Fatalf(
					"%s should return a distinct copy\ngot:  %p\nwant: %p",
					prefix, gotInterface, tc.lookup,
				)
			}

			// AND: the type is unchanged.
			got, ok := gotInterface.(*Lookup)
			if !ok {
				t.Fatalf(
					"%s type shouldn't have changed\ngot:  %T\nwant: Lookup",
					prefix, gotInterface,
				)
			}

			// AND: the copy unmarshals the same.
			if gotStr := got.String(""); gotStr != wantStr {
				t.Fatalf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, wantStr,
				)
			}

			// AND: the fields are copied as expected.
			err := []test.FieldAssertion{
				{Name: "Type", Got: got.Type, Want: tc.lookup.Type, Mode: test.CompareEqual},
				{Name: "URL", Got: got.URL, Want: tc.lookup.URL, Mode: test.CompareEqual},
				{Name: "URLCommands", Got: &got.URLCommands, Want: &tc.lookup.URLCommands, Mode: test.CompareDifferentPointer},
			}
			if testErr := test.AssertFields(t, err, prefix, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			// AND: copied pointers should be value-equal and non-aliased.
			err = []test.FieldAssertion{
				{Name: "Require", Got: got.Require, Want: tc.lookup.Require, Mode: test.CompareDifferentPointer},
				{Name: "Options", Got: got.Options, Want: tc.lookup.Options, Mode: test.CompareDifferentPointer},
				{Name: "Status", Got: got.Status, Want: tc.lookup.Status, Mode: test.CompareDifferentPointer},
			}
			if testErr := test.AssertFields(t, err, prefix, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			// AND: the non-Base fields are copied as expected.
			err = []test.FieldAssertion{
				{Name: "AccessToken", Got: got.AccessToken, Want: tc.lookup.AccessToken, Mode: test.CompareEqual},
				{Name: "UsePreRelease", Got: got.UsePreRelease, Want: tc.lookup.UsePreRelease, Mode: test.CompareDifferentPointer},
			}
			if testErr := test.AssertFields(t, err, prefix, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			// AND: defaults pointers are shared.
			err = []test.FieldAssertion{
				{Name: "Defaults", Got: got.Defaults, Want: tc.lookup.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: tc.lookup.HardDefaults, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, err, prefix, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
