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

package option

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
)

func TestDecodeDefaults(t *testing.T) {
	// GIVEN: data in a given format to Decode into Defaults.
	tests := []struct {
		name         string
		format, data string
		errRegex     string
		want         string
	}{
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"interval": "1h",
				"semantic_versioning": true
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				interval: 1h
				semantic_versioning: true
			`),
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				interval: 1h
				semantic_versioning: true
			`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				interval: 1h
				semantic_versioning: true
			`),
		},
		{
			name:   "invalid format",
			format: "invalid",
			data:   `{"foo": "bar"}`,
			errRegex: test.TrimYAML(`
				^options:
					unsupported format: "invalid"$`,
			),
		},
		{
			name:   "YAML/invalid data type",
			format: "yaml",
			data:   "interval: ['1h','2h']",
			errRegex: test.TrimYAML(`
				^options:
					[^\s]+] .*unmarshal.*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				DecodeDefaults,
				tc.format, tc.data,
				func(v *Defaults) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				packageName,
				"DecodeDefaults",
			); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	optCfg := plainDefaultsConfig(t)

	// GIVEN: data in a given format to Decode into Options.
	tests := []struct {
		name         string
		format, data string
		errRegex     string
		want         string
	}{
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
			data:     "",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"interval": "1h",
				"semantic_versioning": true
			}`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				interval: 1h
				semantic_versioning: true
			`),
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				interval: 1h
				semantic_versioning: true
			`),
			errRegex: `^$`,
			want: test.TrimYAML(`
				interval: 1h
				semantic_versioning: true
			`),
		},
		{
			name:   "invalid format",
			format: "invalid",
			data:   `{"foo": "bar"}`,
			errRegex: test.TrimYAML(`
				^options:
					unsupported format: "invalid"$`,
			),
		},
		{
			name:   "invalid data type",
			format: "yaml",
			data:   "interval: ['1h','2h']",
			errRegex: test.TrimYAML(`
				^options:
					[^\s]+] .*unmarshal.*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Decode is called with it.
			options, err, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*Options, error) {
					return Decode(
						format, data,
						optCfg,
					)
				},
				tc.format, tc.data,
				func(v *Options) string { return v.String() },
				tc.want,
				tc.errRegex,
				packageName,
				"Decode",
			)
			if testErr != nil {
				t.Fatal(testErr)
			}
			if err != nil || options == nil {
				return
			}

			prefix := fmt.Sprintf(
				"%s\nDecode(format=%q, data=%q)",
				packageName, tc.format, tc.data,
			)

			// THEN: It stringifies as expected.
			if got := options.String(); got != tc.want {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: Pointers are handed out to it correctly.
			fieldTests := []test.FieldAssertion{
				{Name: "Defaults", Got: options.Defaults, Want: optCfg.Soft, Mode: test.CompareSamePointer},
				{Name: "HardDefaults", Got: options.HardDefaults, Want: optCfg.Hard, Mode: test.CompareSamePointer},
			}
			if err := test.AssertFields(t, fieldTests, prefix, "Lookup"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
