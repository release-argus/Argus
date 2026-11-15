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

package web

import (
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	lvbasetest "github.com/release-argus/Argus/service/latest_version/types/base/test"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/shared"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
)

// ############
// # DECODING #
// ############

func TestLookup_Unmarshal(t *testing.T) {
	lvCfg := lvbasetest.PlainDefaultsConfig()

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
				"type": "web",
				"allow_invalid_certs": true,
				"headers": [
					{"key": "X-Something", "value": "foo"}
				],
				"require": {
					"regex_version": "v.+"
				},
				"url": "https://example.com",
				"url_commands": [
					{"type": "regex", "regex": "foo"}
				],
				"use_prerelease": true
			}`),
			want: test.TrimYAML(`
				require:
					regex_version: v.+
				allow_invalid_certs: true
				headers:
					- key: X-Something
						value: foo
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				type: web
				allow_invalid_certs: true
				headers:
					- key: X-Something
						value: foo
				require:
					regex_version: v.+
				url: https://example.com
				url_commands:
					- type: regex
						regex: foo
			`),
			want: test.TrimYAML(`
				require:
					regex_version: v.+
				allow_invalid_certs: true
				headers:
					- key: X-Something
						value: foo
			`),
			errRegex: `^$`,
		},
		{
			name:     "JSON/invalid data types",
			format:   "json",
			data:     `{"allow_invalid_certs": maybe}`,
			errRegex: `^jsontext: invalid character`,
		},
		{
			name:   "YAML/invalid data types",
			format: "yaml",
			data:   `allow_invalid_certs: maybe`,
			errRegex: test.TrimYAML(`
				^[^\s]+ .*unmarshal.*
				[^\s]+.* allow_invalid_certs:.*
				\s+\^$`,
			),
		},
		{
			name:     "JSON/invalid formatting",
			format:   "json",
			data:     `{"url": "https://example.com`,
			errRegex: `unexpected`,
		},
		{
			name:   "Require/error",
			format: "yaml",
			data: test.TrimYAML(`
				url: https://example.com
				allow_invalid_certs: true
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
					return &v, err
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
	tests := []struct {
		name   string
		lookup *Lookup
		want   string
	}{
		{
			name:   "nil",
			lookup: nil,
			want:   "null\n",
		},
		{
			name:   "empty",
			lookup: &Lookup{},
			want:   "{}\n",
		},
		{
			name: "filled",
			lookup: &Lookup{
				Type: "test",
				URL:  "https://example.com",
				URLCommands: filter.URLCommands{
					{Type: "regex", Regex: "abc"},
				},
				Require: &filter.Require{
					RegexVersion: "def",
				},
				AllowInvalidCerts: new(true),
				Headers: shared.Headers{
					{Key: "X-Test-1", Value: "foo"},
					{Key: "X-Test-2", Value: "bar"},
				},
			},
			want: test.TrimYAML(`
				type: test
				url: https://example.com
				url_commands:
					- type: regex
						regex: abc
				require:
					regex_version: def
				allow_invalid_certs: true
				headers:
					- key: X-Test-1
						value: foo
					- key: X-Test-2
						value: bar
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
	lvCfg := lvbasetest.PlainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig()

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
			name: "headers",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						headers:
							- key: X-Something
								value: foo
					`)),
					nil,
					svcStatus,
					lvCfg,
				)
			}),
			status: nil,
		},
		{
			name: "filled",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				lv, err := Decode(
					"yaml", []byte(test.TrimYAML(`
						type: test
						url: https://example.com
						url_commands:
							- type: regex
								regex: v.*
						require:
							regex_version: '^1.*'
							docker:
								image: foo
								tag: bar

						allow_invalid_certs: true
						headers:
							- key: X-Something
								value: foo
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
					lv.SetTypeDefaults(
						&Defaults{AllowInvalidCerts: new(true)},
						&Defaults{AllowInvalidCerts: new(false)},
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
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Require", Got: got.Require, Want: tc.lookup.Require, Mode: test.CopiedPointerMode(tc.lookup.Require)},
				{Name: "Options", Got: got.Options, Want: tc.lookup.Options, Mode: test.CopiedPointerMode(tc.lookup.Options)},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "AllowInvalidCerts", Got: got.AllowInvalidCerts, Want: tc.lookup.AllowInvalidCerts, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
			if len(got.Headers) > 0 && &got.Headers[0] == &tc.lookup.Headers[0] {
				t.Fatalf(
					"Lookup.Headers: copy aliases the original\ngot:  %p\nwant: NOT %p",
					&got.Headers[0], &tc.lookup.Headers[0],
				)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Status", Got: got.Status, Want: tc.status, Mode: test.CompareSamePointer},
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

func TestLookup_InheritSecrets(t *testing.T) {
	tests := []struct {
		name             string
		lookup, previous *Lookup
		secretRefs       *shared.VSecretRef
		want             string
	}{
		{
			name: "secrets to undefined vars",
			lookup: &Lookup{
				Require: &filter.Require{
					Docker: &docker.GHCRRegistry{
						Auth: &docker.GHCRAuth{
							Token: util.SecretValue,
						},
					},
				},
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
				},
			},
			previous: &Lookup{},
			secretRefs: &shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(0)},
				},
			},
			want: test.TrimYAML(`
				require:
					docker:
						auth:
							token: ` + util.SecretValue + `
				headers:
					- key: X-Test
						value: ` + util.SecretValue + `
			`),
		},
		{
			name: "secrets inherited",
			lookup: &Lookup{
				Require: &filter.Require{
					Docker: &docker.GHCRRegistry{
						Auth: &docker.GHCRAuth{
							Token: util.SecretValue,
						},
					},
				},
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
				},
			},
			previous: &Lookup{
				Require: &filter.Require{
					Docker: &docker.GHCRRegistry{
						Auth: &docker.GHCRAuth{
							Token: "123",
						},
					},
				},
				Headers: shared.Headers{
					{Key: "X-Test-A", Value: "A"},
					{Key: "X-Test-B", Value: "B"},
					{Key: "X-Test-C", Value: "C"},
				},
			},
			secretRefs: &shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: new(1)},
				},
			},
			want: test.TrimYAML(`
				require:
					docker:
						auth:
							token: '123'
				headers:
					- key: X-Test
						value: B
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.lookup.InheritSecrets(tc.previous, tc.secretRefs)

			if got := tc.lookup.String(""); got != tc.want {
				t.Errorf("Lookup.String() mismatch\ngot:  %q\nwant: %q", got, tc.want)
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
			defaults:     &Defaults{AllowInvalidCerts: new(true)},
			hardDefaults: &Defaults{AllowInvalidCerts: new(false)},
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
