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
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

func TestDecode(t *testing.T) {
	dvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	type Args struct {
		format string
		data   string
		lookup *Lookup
	}
	// GIVEN: data in a given format to Decode into a Lookup.
	tests := []struct {
		name     string
		args     Args
		want     string
		errRegex string
	}{
		{
			name: "JSON/empty",
			args: Args{
				format: "json",
				data:   `{}`,
			},
			want: "{}\n",
		},
		{
			name: "YAML/empty",
			args: Args{
				format: "yaml",
				data:   "",
			},
			errRegex: `^$`,
			want:     "",
		},
		{
			name: "JSON/invalid payload decode error",
			args: Args{
				format: "json",
				data:   `{`,
			},
			errRegex: `unexpected`,
		},
		{
			name: "JSON/invalid data type",
			args: Args{
				format: "json",
				data:   `{"type": ["url"]}`,
			},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "JSON/valid payload, no require",
			args: Args{
				format: "json",
				data:   `{"type": "url"}`,
			},
			errRegex: `^$`,
			want:     "type: url\n",
		},
		{
			name: "JSON/filled",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "foo",
					"url": "hi",
					"require": {
						"regex_version": "foo",
						"docker": {
							"image": "i",
							"tag": "t"
						}
					}
				}`),
			},
			errRegex: `^$`,
			want:     "type: foo\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// AND: Options + Status.
			options, _ := opt.Decode(
				"yaml", nil,
				optCfg,
			)
			svcStatus, _ := statustest.New("yaml", nil)

			// WHEN: Decode is called.
			lookup, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					return Decode(
						format, data,
						options,
						svcStatus,
						dvCfg,
					)
				},
				tc.args.format, tc.args.data,
				func(v *Lookup) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"Decode",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nDecode(format=%q, data=%q)",
				packageName, tc.args.format, tc.args.data,
			)

			// THEN: pointers are handed out as expected.
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: dvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: dvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestApplyOverrides(t *testing.T) {
	dvCfg := plainDefaultsConfig(t)
	optCfg := opttest.PlainDefaultsConfig(t)

	type Args struct {
		format, data string
		target       *Lookup
	}
	// GIVEN: data to unmarshal in a format onto a Require.
	tests := []struct {
		name           string
		args           Args
		want           string
		pointerChanged bool
		errRegex       string
	}{
		{
			name: "JSON/no data",
			args: Args{
				format: "json",
				data:   "",
				target: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "YAML/no data",
			args: Args{
				format: "yaml",
				data:   "",
				target: &Lookup{},
			},
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name: "JSON/new, filled",
			args: Args{
				format: "json",
				data:   `{"type": "foo"}`,
			},
			pointerChanged: true,
			want:           "type: foo\n",
			errRegex:       `^$`,
		},
		{
			name: "YAML/new, filled",
			args: Args{
				format: "yaml",
				data:   `type: foo`,
			},
			pointerChanged: true,
			want:           "type: foo\n",
			errRegex:       `^$`,
		},
		{
			name: "JSON/existing, filled",
			args: Args{
				format: "json",
				data:   `{"type": "bar"}`,
				target: &Lookup{
					Type: "foo",
				},
			},
			want:     "type: bar\n",
			errRegex: `^$`,
		},
		{
			name: "YAML/existing, filled",
			args: Args{
				format: "yaml",
				data:   `type: bar`,
				target: &Lookup{
					Type: "foo",
				},
			},
			want:     "type: bar\n",
			errRegex: `^$`,
		},
		{
			name: "JSON/existing, filled - type error",
			args: Args{
				format: "json",
				data:   `{"type": []}`,
				target: &Lookup{},
			},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "YAML/existing, filled - type error",
			args: Args{
				format: "yaml",
				data:   `type: []`,
				target: &Lookup{},
			},
			errRegex: `^[^\s]+ .*unmarshal.*`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options, _ := opt.Decode(
				"yaml", nil,
				optCfg,
			)
			svcStatus := status.New(
				nil, nil, nil,
				"",
				"", "",
				"", "",
				"",
				nil,
			)
			if tc.args.target != nil {
				tc.args.target.Init(options, svcStatus, dvCfg)
			}

			// WHEN: ApplyOverrides is called.
			lookup, err, testErr := test.AssertApplyOverrides(
				t,
				tc.args.target,
				func(format string, data []byte, v *Lookup) (*Lookup, error) {
					return ApplyOverrides(
						format, data,
						v,
						options,
						svcStatus,
						dvCfg,
					)
				},
				tc.args.format, tc.args.data,
				func(v *Lookup) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				true,
				packageName,
				"ApplyOverrides",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nApplyOverrides(format=%q, data=%q)",
				packageName, tc.args.format, tc.args.data,
			)

			// AND: pointers are handed out as expected.
			wantOptions := options
			wantStatus := svcStatus
			wantSoft := dvCfg.Soft
			wantHard := dvCfg.Hard
			if !tc.pointerChanged {
				wantOptions = lookup.Options
				wantStatus = lookup.Status
				wantSoft = lookup.Defaults
				wantHard = lookup.HardDefaults
			}
			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: wantOptions, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: wantStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: wantSoft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: wantHard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
