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

package base

import (
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

func TestLookup_Copy(t *testing.T) {
	lvCfg := plainDefaultsConfig()
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
			name: "require",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
							require:
								regex_version: '^1.*'
						`)),
					nil,
					svcStatus,
					lvCfg,
				)
			}),
			status: nil,
		},
		{
			name: "require.docker",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
							require:
								regex_version: '^1.*'
								docker:
									image: foo
									tag: bar
						`)),
					nil,
					svcStatus,
					lvCfg,
				)
			}),
			status: nil,
		},
		{
			name: "options",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte("type: test"),
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

			if gotStr := decode.ToYAMLString(got, ""); gotStr != wantStr {
				t.Errorf("Lookup mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
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
				{Name: "Status", Got: got.Status, Want: tc.status, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: got.Defaults, Want: tc.lookup.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: tc.lookup.HardDefaults, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
