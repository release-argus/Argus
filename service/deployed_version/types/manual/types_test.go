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
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

// ############
// # DECODING #
// ############

func TestLookup_Unmarshal(t *testing.T) {
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:   "JSON/empty",
			format: "json",
			data:   "",
			errRegex: test.TrimYAML(`
				^jsontext:
					unexpected EOF$`,
			),
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
				"type": "manual",
				"version": "1.2.3"
			}`),
			want: test.TrimYAML(`
				type: manual
				version: 1.2.3
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				type: manual
				version: "1.2.3"
			`),
			want: test.TrimYAML(`
				type: manual
				version: 1.2.3
			`),
			errRegex: `^$`,
		},
		{
			name:   "JSON/invalid type - version",
			format: "json",
			data: test.TrimJSON(`{
				"type": "manual",
				"version": ["1.2.3"]
			}`),
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name:   "YAML/invalid type - version",
			format: "yaml",
			data:   `version: ["https://example.com"]`,
			errRegex: test.TrimYAML(`
				^[^\s]+ .*unmarshal.*
				\>  \d | .+$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					var zero Lookup
					err := decode.Unmarshal(format, data, &zero)
					return &zero, err
				},
				tc.format, tc.data,
				func(v *Lookup) string { return v.String("") },
				tc.want,
				tc.errRegex,
				"Lookup.Unmarshal",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

// #########
// # STATE #
// #########

func TestLookup_Copy(t *testing.T) {
	dvCfg := plainDefaultsConfig()
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

			gotInterface := tc.lookup.Copy(tc.status)

			if tc.lookup == nil {
				if gotInterface != nil {
					t.Error("Lookup.Copy() mismatch\ngot:  non-nil\nwant: nil")
				}
				return
			}

			if gotInterface == nil {
				t.Fatal("Lookup.Copy() mismatch\ngot:  nil\nwant: non-nil")
			}

			if gotInterface == tc.lookup {
				t.Fatalf(
					"copy aliases the original\ngot:  %p\nwant: NOT %p",
					gotInterface, tc.lookup,
				)
			}

			got, ok := gotInterface.(*Lookup)
			if !ok {
				t.Fatalf("type mismatch\ngot:  %T\nwant: *Lookup", gotInterface)
			}

			if gotStr := got.String(""); gotStr != wantStr {
				t.Fatalf("Lookup.String() mismatch\ngot:  %q\nwant: %q", gotStr, wantStr)
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Type", Got: got.Type, Want: tc.lookup.Type, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Options", Got: got.Options, Want: tc.lookup.Options, Mode: test.CopiedPointerMode(tc.lookup.Options)},
				{Name: "Status", Got: got.Status, Want: tc.status, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Version", Got: got.Version, Want: tc.lookup.Version, Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}

			fieldTests = []test.FieldAssertion{
				{Name: "Defaults", Got: got.Defaults, Want: tc.lookup.Defaults, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: got.HardDefaults, Want: tc.lookup.HardDefaults, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestString(t *testing.T) {
	dvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig()

	lookup := testLookup(t, "")
	lookup.Status.SetDeployedVersion(
		lookup.Status.DeployedVersion(),
		time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
		false,
	)
	lookup.Status.ServiceInfo.ID = t.Name()
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
				options, err := opt.Decode(
					"yaml", []byte("interval: 1h2m3s"),
					optCfg,
				)
				if err != nil {
					return nil, err
				}
				l, err := Decode("yaml", []byte(test.TrimYAML(`
						type: manual
						version: 1.2.3
					`)),
					options,
					lookup.Status.Copy(true),
					dvCfg,
				)
				if err != nil {
					return nil, err
				}
				return l, l.Query(false, logx.LogFrom{})
			}),
			want: "type: manual\n",
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
				"Lookup.String",
				tc.lookup.String,
				tc.want,
			)
		})
	}
}
