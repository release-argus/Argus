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

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	dashtest "github.com/release-argus/Argus/service/dashboard/test"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
)

func TestLookup_DecodeSelf(t *testing.T) {
	dvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig(t)

	tests := []struct {
		name         string
		format, data string
		errRegex     string
		want         string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     ``,
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     `{}`,
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     ``,
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "JSON/invalid payload decode error",
			format:   "json",
			data:     `{`,
			errRegex: `unexpected`,
		},
		{
			name:     "JSON/invalid data types",
			format:   "json",
			data:     `{"version": ["v.1.2.3"]}`,
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name:   "filled/semantic version",
			format: "json",
			data: test.TrimJSON(`{
				"type": "manual",
				"version": "1.2.3"
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: manual
				version: 1.2.3
			`),
		},
		{
			name:   "filled/non-semantic version",
			format: "json",
			data: test.TrimJSON(`{
				"type": "manual",
				"version": "1.2.3 4"
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: manual
				version: 1.2.3 4
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := opttest.PlainOptions(t, optCfg)
			dash := dashtest.PlainOptions(t)
			svcStatus := status.New(
				nil, nil, nil,
				"",
				"", "",
				"", "", "",
				dash,
			)
			dvl := &Lookup{}
			dvl.Init(
				options,
				svcStatus,
				dvCfg,
			)

			lookup, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Lookup, error) {
					err := dvl.DecodeSelf(format, data)
					return dvl, err
				},
				tc.format,
				tc.data,
				func(v *Lookup) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				"Lookup.DecodeSelf",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: dvl.GetOptions(), Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: dvl.GetStatus(), Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: dvl.GetDefaults(), Want: dvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: dvl.GetHardDefaults(), Want: dvCfg.Hard, Mode: test.CompareSamePointer},
				{Name: "DeployedVersion()", Target: "Status", Got: dvl.Status.DeployedVersion(), Want: "", Mode: test.CompareEqual},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestLookup_ApplyOverrides(t *testing.T) {
	dvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig(t)

	type Args struct {
		format, data string
		target       *Lookup
	}

	tests := []struct {
		name     string
		args     Args
		errRegex string
		want     string
	}{
		{
			name: "empty data returns previous",
			args: Args{
				format: "json",
				data:   "",
				target: &Lookup{},
			},
			errRegex: `^$`,
		},
		{
			name: "invalid payload causes decode error",
			args: Args{
				format: "json",
				data:   `{`,
				target: &Lookup{},
			},
			errRegex: `unexpected`,
		},
		{
			name: "override error/base.Lookup",
			args: Args{
				format: "json",
				data:   `{"type": []}`,
				target: &Lookup{},
			},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "override error/Lookup",
			args: Args{
				format: "json",
				data:   `{"version": ["v1.2.3"]}`,
				target: &Lookup{},
			},
			errRegex: `^json: .*unmarshal.*$`,
		},
		{
			name: "filled",
			args: Args{
				format: "json",
				data: test.TrimJSON(`{
					"type": "manual",
					"version": "1.2.3"
				}`),
				target: &Lookup{
					Type:    "manual",
					Version: "0.0.0",
				},
			},
			errRegex: `^$`,
			want: test.TrimYAML(`
				type: manual
				version: 1.2.3
			`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			options := opttest.PlainOptions(t, optCfg)
			svcStatus := &status.Status{}
			tc.args.target.Init(options, svcStatus, dvCfg)
			if tc.want == "" {
				tc.want = decode.ToYAMLString(tc.args.target, "")
			}

			lookup, err, testErr := test.AssertApplyOverrides(
				t,
				tc.args.target,
				func(format string, data []byte, v *Lookup) (*Lookup, error) {
					err := v.ApplyOverrides(format, data)
					return v, err
				},
				tc.args.format, tc.args.data,
				func(v *Lookup) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				true,
				"Lookup.ApplyOverrides",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || lookup == nil {
				return
			}

			fieldTests := []test.FieldAssertion{
				{Name: "Options", Got: lookup.Options, Want: options, Mode: test.CompareSamePointer},
				{Name: "Status", Got: lookup.Status, Want: svcStatus, Mode: test.CompareSamePointer},
				{Name: "Defaults", Got: lookup.Defaults, Want: dvCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: lookup.HardDefaults, Want: dvCfg.Hard, Mode: test.CompareSamePointer},
			}
			if testErr := test.AssertFields(t, fieldTests, "Lookup"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}
