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

package manual

import (
	"fmt"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// ############
// # DECODING #
// ############

func TestLookup_UnmarshalJSON(t *testing.T) {
	// GIVEN: data to unmarshal into a Lookup.
	tests := []struct {
		name     string
		format   string
		data     []byte
		want     string
		errRegex string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     []byte{},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     []byte{},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: []byte(test.TrimJSON(`{
				"type": "manual",
				"version": "1.2.3"
			}`)),
			want: test.TrimYAML(`
				type: manual
				version: 1.2.3
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: []byte(test.TrimYAML(`
				type: manual
				version: "1.2.3"
			`)),
			want: test.TrimYAML(`
				type: manual
				version: 1.2.3
			`),
			errRegex: `^$`,
		},
		{
			name:   "JSON/invalid type - version",
			format: "json",
			data: []byte(test.TrimJSON(`{
				"type": "manual",
				"version": ["1.2.3"]
			}`)),
			want: test.TrimYAML(`
				type: manual
				version: "1.2.3"
			`),
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name:   "YAML/invalid type - version",
			format: "yaml",
			data:   []byte(`version: ["https://example.com"]`),
			errRegex: test.TrimYAML(`
				^[^\s]+ .*unmarshal.*
				\>  \d | .+$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: a Lookup.
			tgt := Lookup{}

			// WHEN: the data is unmarshaled into that Service.
			err := decode.Unmarshal(tc.format, tc.data, &tgt)

			prefix := fmt.Sprintf(
				"%s\nLookup Unmarshal(format=%q, data=%q)",
				packageName, tc.format, tc.data,
			)

			// THEN: The error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the Lookup is as expected.
			if got := tgt.String(""); got != tc.want {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestString(t *testing.T) {
	dvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	// GIVEN: a Lookup.
	lookup := testLookup(t, "")
	lookup.Status.SetDeployedVersion(
		lookup.Status.DeployedVersion(),
		time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
		false,
	)
	lookup.Status.ServiceInfo.ID = "TestString"
	tests := []struct {
		name   string
		lookup *Lookup
		want   string
	}{
		{
			name: "empty",
			lookup: test.Must(t, func() (*Lookup, error) {
				l := Lookup{}
				l.Status = &status.Status{}
				l.Status.ServiceInfo.ID = "empty"
				return &l, nil
			}),
			want: "{}\n",
		},
		{
			name: "version moved to Status",
			lookup: test.Must(t, func() (*Lookup, error) {
				options, _ := opt.Decode(
					"yaml", []byte("interval: 1h2m3s"),
					optCfg,
				)
				lookup, _ := Decode("yaml", []byte(test.TrimYAML(`
						type: manual
						version: 1.2.3
					`)),
					options,
					lookup.Status.Copy(true),
					dvCfg,
				)
				lookup.Query(false, logx.LogFrom{})
				return lookup, nil
			}),
			want: "type: manual\n",
		},
		{
			name: "filled",
			lookup: test.Must(t, func() (*Lookup, error) {
				options, _ := opt.Decode(
					"yaml", []byte("interval: 1h2m3s"),
					optCfg,
				)
				return Decode("yaml", []byte(test.TrimYAML(`
						type: manual
						version: 1.2.3
					`)),
					options,
					lookup.Status.Copy(true),
					dvCfg,
				)
			}),
			want: test.TrimYAML(`
				type: manual
				version: 1.2.3
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
	dvCfg := plainDefaultsConfig(t)
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
			name: "filled",
			lookup: test.Must(t, func() (*Lookup, error) {
				svcStatus, _ := statustest.New("yaml", nil)
				return Decode(
					"yaml", []byte(test.TrimYAML(`
						type: test
						version: foo
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
					dvCfg,
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
			fieldTests := []test.FieldAssertion{
				{Name: "Type", Got: got.Type, Want: tc.lookup.Type, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}

			// AND: copied pointers should be value-equal and non-aliased.
			fieldTests = []test.FieldAssertion{
				{Name: "Options", Got: got.Options, Want: tc.lookup.Options, Mode: test.CompareDifferentPointer},
				{Name: "Status", Got: got.Status, Want: tc.status, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}

			// AND: the non-Base fields are copied as expected.
			fieldTests = []test.FieldAssertion{
				{Name: "Version", Got: got.Version, Want: tc.lookup.Version, Mode: test.CompareEqual},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}

			// AND: defaults pointers are shared.
			fieldTests = []test.FieldAssertion{
				{Name: "Defaults", Got: got.Defaults, Want: tc.lookup.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: tc.lookup.HardDefaults, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
