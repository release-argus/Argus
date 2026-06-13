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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/latest_version/filter"
	"github.com/release-argus/Argus/service/latest_version/filter/docker"
	"github.com/release-argus/Argus/service/latest_version/types/base"
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
			name:   "Require - error",
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
			want: test.TrimYAML(`
				url: https://example.com
				url_commands:
					- type: regex
						regex: foo
				allow_invalid_certs: true
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

			v := Lookup{
				Lookup: base.Lookup{
					Defaults:     lvCfg.Soft,
					HardDefaults: lvCfg.Hard,
				},
			}
			if _, testErr := test.AssertUnmarshal(
				t,
				tc.format, tc.data,
				&v,
				tc.errRegex,
				func(t *Lookup) string { return t.String("") },
				tc.want,
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
	// GIVEN: a Lookup.
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
				Lookup: base.Lookup{
					Type: "test",
					URL:  "https://example.com",
					URLCommands: filter.URLCommands{
						{Type: "regex", Regex: "abc"},
					},
					Require: &filter.Require{
						RegexVersion: "def",
					},
				},
				AllowInvalidCerts: test.Ptr(true),
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
				return Decode(
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
				{Name: "AllowInvalidCerts", Got: got.AllowInvalidCerts, Want: tc.lookup.AllowInvalidCerts, Mode: test.CompareSamePointer},
				{Name: "Headers", Got: &got.Headers, Want: &tc.lookup.Headers, Mode: test.CompareDifferentPointer},
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

func TestLookup_InheritSecrets(t *testing.T) {
	// GIVEN: a Lookup, a 'previous' Lookup, and secret refs.
	tests := []struct {
		name             string
		lookup, previous *Lookup
		secretRefs       *shared.VSecretRef
		want             string
	}{
		{
			name: "secrets to undefined vars",
			lookup: &Lookup{
				Lookup: base.Lookup{
					Require: &filter.Require{
						Docker: &docker.GHCRRegistry{
							CommonRegistry: docker.CommonRegistry{
								Auth: &docker.GHCRAuth{
									GHCRAuthDefaults: docker.GHCRAuthDefaults{
										Token: util.SecretValue,
									},
								},
							},
						},
					},
				},
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
				},
			},
			previous: &Lookup{
				Lookup: base.Lookup{},
			},
			secretRefs: &shared.VSecretRef{
				Headers: []shared.OldIntIndex{
					{OldIndex: test.Ptr(0)},
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
				Lookup: base.Lookup{
					Require: &filter.Require{
						Docker: &docker.GHCRRegistry{
							CommonRegistry: docker.CommonRegistry{
								Auth: &docker.GHCRAuth{
									GHCRAuthDefaults: docker.GHCRAuthDefaults{
										Token: util.SecretValue,
									},
								},
							},
						},
					},
				},
				Headers: shared.Headers{
					{Key: "X-Test", Value: util.SecretValue},
				},
			},
			previous: &Lookup{
				Lookup: base.Lookup{
					Require: &filter.Require{
						Docker: &docker.GHCRRegistry{
							CommonRegistry: docker.CommonRegistry{
								Auth: &docker.GHCRAuth{
									GHCRAuthDefaults: docker.GHCRAuthDefaults{
										Token: "123",
									},
								},
							},
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
					{OldIndex: test.Ptr(1)},
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

			// WHEN: InheritSecrets is called.
			tc.lookup.InheritSecrets(tc.previous, tc.secretRefs)

			// THEN: the expected secrets are inherited.
			if got := tc.lookup.String(""); got != tc.want {
				t.Errorf("%s\nLookup.InheritSecrets(secrets=%+v) mismatch\ngot:  %q\nwant: %q",
					packageName, tc.secretRefs, got, tc.want,
				)
			}
		})
	}
}
