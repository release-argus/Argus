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
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	lvbasetest "github.com/release-argus/Argus/service/latest_version/types/base/test"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

// ############
// # DECODING #
// ############

func TestLookup_Unmarshal(t *testing.T) {
	lvCfg := lvbasetest.PlainDefaultsConfig(t)

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
			name:   "Require/error",
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
						Defaults:     lvCfg.Soft,
						HardDefaults: lvCfg.Hard,
					}
					err := decode.Unmarshal(format, data, &v)
					return &v, err //nolint:wrapcheck
				},
				tc.format, tc.data,
				func(v *Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"Lookup.Unmarshal",
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
	lvCfg := lvbasetest.PlainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

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
				options, err := opt.Decode(
					"yaml", []byte("interval: 1h2m3s"),
					optCfg,
				)
				if err != nil {
					return nil, err
				}
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
				"Lookup.String",
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
			name: "data",
			lookup: &Lookup{
				apiRoot: "http://127.0.0.1:1/api",
				data: Data{
					etag:    "foo",
					perPage: 1,
					releases: []forgetypes.Release{
						{URL: "example.com"},
					},
					tagFallback: true,
				},
				Status: test.Must(t, func() (*status.Status, error) {
					return statustest.New("yaml", nil)
				}),
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
						etag:    "foo",
						perPage: 1,
						releases: []forgetypes.Release{
							{URL: "example.com"},
						},
						tagFallback: true,
					}
					lv.SetTypeDefaults(
						&Defaults{AccessToken: "soft"},
						&Defaults{AccessToken: "hard"},
					)
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

			gotInterface := tc.lookup.Copy(tc.status)

			if tc.lookup == nil {
				if gotInterface != nil {
					t.Errorf("copy mismatch\ngot:  %T\nwant: nil", gotInterface)
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
				{Name: "apiRoot", Got: got.apiRoot, Want: tc.lookup.apiRoot, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
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
				{Name: "data.etag", Got: got.data.etag, Want: tc.lookup.data.etag, Mode: test.CompareEqual},
				{Name: "data.perPage", Got: got.data.perPage, Want: tc.lookup.data.perPage, Mode: test.CompareEqual},
				{Name: "data.tagFallback", Got: got.data.tagFallback, Want: tc.lookup.data.tagFallback, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
			if testErr := test.AssertSlicesEqualFunc(
				t,
				got.data.releases,
				tc.lookup.data.releases,
				releaseEqual,
				"Lookup.data.releases",
			); testErr != nil {
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

// ############
// # DEFAULTS #
// ############

func TestLookup_TypeDefaults(t *testing.T) {
	tests := []struct {
		name         string
		defaults     *Defaults
		hardDefaults *Defaults
	}{
		{
			name:         "nil, nil",
			defaults:     nil,
			hardDefaults: nil,
		},
		{
			name:         "non-nil, non-nil",
			defaults:     &Defaults{AccessToken: "soft"},
			hardDefaults: &Defaults{AccessToken: "hard"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lookup := &Lookup{}

			lookup.SetTypeDefaults(tc.defaults, tc.hardDefaults)

			gotDefaults, gotHardDefaults := lookup.GetTypeDefaults()
			if gotDefaults != tc.defaults {
				t.Errorf(
					"Lookup.GetTypeDefaults() defaults mismatch\ngot:  %p\nwant: %p",
					gotDefaults, tc.defaults,
				)
			}
			if gotHardDefaults != tc.hardDefaults {
				t.Errorf(
					"Lookup.GetTypeDefaults() hardDefaults mismatch\ngot:  %p\nwant: %p",
					gotHardDefaults, tc.hardDefaults,
				)
			}
		})
	}
}
