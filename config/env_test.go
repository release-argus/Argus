// Copyright [2024] [Argus]
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

package config

import (
	"errors"
	"os"
	"testing"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestMapEnvToStruct(t *testing.T) {
	// GIVEN a struct and a bunch of env vars
	test := map[string]struct {
		customStruct interface{}
		prefix       string
		env          map[string]string
		want         string
		errRegex     string
	}{
		"no ARGUS_ env vars": {
			env: map[string]string{
				"TEST_STRING": "foo",
				"TEST_INT":    "1"},
			customStruct: &struct {
				Test struct {
					String string `yaml:"string"`
					Int    int    `yaml:"int"`
				} `yaml:"test"`
			}{},
			want: "",
		},
		"nil non-comparable pointer": {
			env: map[string]string{
				"ARGUS_TEST_SLICE": "1,2,3",
				"ARGUS_TEST_MAP":   "foo:1,bar:2",
				"ARGUS_TEST_FUNC":  "func()",
				"ARGUS_TEST_INT":   "1"},
			customStruct: &struct {
				Test struct {
					PtrToSlice *[]int          `yaml:"slice"`
					PtrToMap   *map[string]int `yaml:"map"`
					PtrToFunc  *func()         `yaml:"func"`
					PtrToInt   *int            `yaml:"int"`
				} `yaml:"test"`
			}{},
			want: "",
		},
		"boolean": {
			env: map[string]string{
				"ARGUS_TEST_BOOLEAN_PTR0": "false",
				"ARGUS_TEST_BOOLEAN_PTR1": "f",
				"ARGUS_TEST_BOOLEAN_PTR2": "0",
				"ARGUS_TEST_BOOLEAN_PTR3": "",
				"ARGUS_TEST_BOOLEAN_VAL0": "true",
				"ARGUS_TEST_BOOLEAN_VAL1": "t",
				"ARGUS_TEST_BOOLEAN_VAL2": "1",
				"ARGUS_TEST_BOOLEAN_VAL3": ""},
			customStruct: &struct {
				Test struct {
					String struct {
						Ptr0 *bool `yaml:"ptr0"`
						Ptr1 *bool `yaml:"ptr1"`
						Ptr2 *bool `yaml:"ptr2"`
						Ptr3 *bool `yaml:"ptr3"`
						Val0 bool  `yaml:"val0"`
						Val1 bool  `yaml:"val1"`
						Val2 bool  `yaml:"val2"`
						Val3 bool  `yaml:"val3"`
					} `yaml:"boolean"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					boolean:
						ptr0: false
						ptr1: false
						ptr2: false
						ptr3: null
						val0: true
						val1: true
						val2: true
						val3: false
			`),
			errRegex: `^$`,
		},
		"integer": {
			env: map[string]string{
				"ARGUS_TEST_INTEGER_PTR0": "0",
				"ARGUS_TEST_INTEGER_PTR1": "1",
				"ARGUS_TEST_INTEGER_PTR2": "-1",
				"ARGUS_TEST_INTEGER_PTR3": "",
				"ARGUS_TEST_INTEGER_VAL0": "0",
				"ARGUS_TEST_INTEGER_VAL1": "1",
				"ARGUS_TEST_INTEGER_VAL2": "-1",
				"ARGUS_TEST_INTEGER_VAL3": ""},
			customStruct: &struct {
				Test struct {
					String struct {
						Ptr0 *int `yaml:"ptr0"`
						Ptr1 *int `yaml:"ptr1"`
						Ptr2 *int `yaml:"ptr2"`
						Ptr3 *int `yaml:"ptr3"`
						Val0 int  `yaml:"val0"`
						Val1 int  `yaml:"val1"`
						Val2 int  `yaml:"val2"`
						Val3 int  `yaml:"val3"`
					} `yaml:"integer"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					integer:
						ptr0: 0
						ptr1: 1
						ptr2: -1
						ptr3: null
						val0: 0
						val1: 1
						val2: -1
						val3: 0
			`),
			errRegex: `^$`,
		},
		"string": {
			env: map[string]string{
				"ARGUS_TEST_STRING_PTR0": "foo",
				"ARGUS_TEST_STRING_PTR1": "",
				"ARGUS_TEST_STRING_VAL":  "bar"},
			customStruct: &struct {
				Test struct {
					String struct {
						Ptr0 *string `yaml:"ptr0"`
						Ptr1 *string `yaml:"ptr1"`
						Ptr2 *string `yaml:"ptr2"`
						Val  string  `yaml:"val"`
					} `yaml:"string"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					string:
						ptr0: foo
						ptr1: null
						ptr2: null
						val: bar
			`),
		},
		"uint8": {
			env: map[string]string{
				"ARGUS_TEST_UNSIGNED_INTEGER_8_PTR0": "0",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_PTR1": "1",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_PTR2": "",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_VAL0": "0",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_VAL1": "1",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_VAL2": ""},
			customStruct: &struct {
				Test struct {
					String struct {
						Ptr0 *uint8 `yaml:"ptr0"`
						Ptr1 *uint8 `yaml:"ptr1"`
						Ptr2 *uint8 `yaml:"ptr2"`
						Val0 uint8  `yaml:"val0"`
						Val1 uint8  `yaml:"val1"`
						Val2 uint8  `yaml:"val2"`
					} `yaml:"unsigned_integer_8"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					unsigned_integer_8:
						ptr0: 0
						ptr1: 1
						ptr2: null
						val0: 0
						val1: 1
						val2: 0
			`),
			errRegex: `^$`,
		},
		"uint8 - invalid": {
			env: map[string]string{
				"ARGUS_TEST_UNSIGNED_INTEGER_8_PTR_INVALID": "1024",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_VAL_INVALID": "-1"},
			customStruct: &struct {
				Test struct {
					String struct {
						Ptr0 *uint8 `yaml:"ptr_invalid"`
						Val0 uint8  `yaml:"val_invalid"`
					} `yaml:"unsigned_integer_8"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					unsigned_integer_8:
						ptr_invalid: null
						val_invalid: 0
			`),
			errRegex: test.TrimYAML(`
				^ARGUS_TEST_UNSIGNED_INTEGER_8_PTR_INVALID: "1024" <invalid>.*
				ARGUS_TEST_UNSIGNED_INTEGER_8_VAL_INVALID: "-1" <invalid>.*$`),
		},
		"uint16": {
			env: map[string]string{
				"ARGUS_TEST_UNSIGNED_INTEGER_16_PTR0": "0",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_PTR1": "1",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_PTR2": "",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_VAL0": "0",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_VAL1": "1",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_VAL2": ""},
			customStruct: &struct {
				Test struct {
					String struct {
						Ptr0 *uint16 `yaml:"ptr0"`
						Ptr1 *uint16 `yaml:"ptr1"`
						Ptr2 *uint16 `yaml:"ptr2"`
						Val0 uint16  `yaml:"val0"`
						Val1 uint16  `yaml:"val1"`
						Val2 uint16  `yaml:"val2"`
					} `yaml:"unsigned_integer_16"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					unsigned_integer_16:
						ptr0: 0
						ptr1: 1
						ptr2: null
						val0: 0
						val1: 1
						val2: 0
			`),
			errRegex: `^$`,
		},
		"uint16 - invalid": {
			env: map[string]string{
				"ARGUS_TEST_UNSIGNED_INTEGER_16_PTR_INVALID": "65536",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_VAL_INVALID": "-1"},
			customStruct: &struct {
				Test struct {
					String struct {
						Ptr0 *uint16 `yaml:"ptr_invalid"`
						Val0 uint16  `yaml:"val_invalid"`
					} `yaml:"unsigned_integer_16"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					unsigned_integer_16:
						ptr_invalid: null
						val_invalid: 0
			`),
			errRegex: test.TrimYAML(`
				^ARGUS_TEST_UNSIGNED_INTEGER_16_PTR_INVALID: "65536" <invalid>.*
				ARGUS_TEST_UNSIGNED_INTEGER_16_VAL_INVALID: "-1" <invalid>.*$`),
		},
		"inline struct": {
			env: map[string]string{
				"ARGUS_TEST_INLINE_STRING": "foo"},
			customStruct: &struct {
				Test struct {
					Inline struct {
						String string `yaml:"string"`
					} `yaml:",inline"`
				} `yaml:"test_inline"`
			}{},
			want: test.TrimYAML(`
				test_inline:
					string: foo
			`),
			errRegex: `^$`,
		},
		"inline struct - error": {
			env: map[string]string{
				"ARGUS_TEST_INLINE_INT": "foo"},
			customStruct: &struct {
				Test struct {
					Inline struct {
						Integer int `yaml:"int"`
					} `yaml:",inline"`
				} `yaml:"test_inline"`
			}{},
			want: test.TrimYAML(`
				test_inline:
					int: 0
			`),
			errRegex: `^ARGUS_TEST_INLINE_INT: "foo" <invalid>.*$`,
		},
		"map - error": {
			env: map[string]string{
				"ARGUS_MAP_FOO_BOOL": "maybe"},
			customStruct: &struct {
				Map map[string]struct {
					Bool *bool `yaml:"bool"`
				} `yaml:"map"`
			}{
				Map: map[string]struct {
					Bool *bool `yaml:"bool"`
				}{
					"foo": {},
				}},
			errRegex: `ARGUS_MAP_FOO_BOOL: "maybe" <invalid>`,
		},
		"struct that was nil - error": {
			env: map[string]string{
				"ARGUS_STRUCT_BOOL": "sometimes"},
			customStruct: &struct {
				Struct *struct {
					Bool *bool `yaml:"bool"`
				} `yaml:"struct"`
			}{},
			errRegex: `ARGUS_STRUCT_BOOL: "sometimes" <invalid>`,
		},
	}

	for name, tc := range test {
		t.Run(name, func(t *testing.T) {

			for k, v := range tc.env {
				os.Setenv(k, v)
				t.Cleanup(func() { os.Unsetenv(k) })
			}

			// WHEN mapEnvToStruct is called on it
			err := mapEnvToStruct(tc.customStruct, tc.prefix, nil)

			// THEN any error is as expected
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) { // Expected a FATAL panic to be caught above
				t.Errorf("mapEnvToStruct() want error matching:\n%q\ngot:\n%q",
					tc.errRegex, e)
			}
			if tc.errRegex != "^$" {
				return
			}
			// AND the defaults are set to the appropriate env vars
			gotYAML := util.ToYAMLString(tc.customStruct, "")
			if tc.want != gotYAML {
				t.Errorf("mapEnvToStruct() mismatch\nwant:\n%q\ngot:\n%q",
					tc.want, gotYAML)
			}
		})
	}
}

func TestConvertToEnvErrors(t *testing.T) {
	tests := map[string]struct {
		input, expected error
	}{
		"nil error": {
			input:    nil,
			expected: nil,
		},
		"single error": {
			input: errors.New(test.TrimYAML(`
				service:
					options:
						interval: "10x" <invalid>`)),
			expected: errors.New(`ARGUS_SERVICE_OPTIONS_INTERVAL: "10x" <invalid>`),
		},
		"multiple errors": {
			input: errors.New(test.TrimYAML(`
				service:
					options:
						interval: "10x" <invalid>
					latest_version:
						require:
							docker:
								type: "pizza" <invalid>
				webhook:
					delay: "10y" <invalid>`)),
			expected: errors.Join(
				errors.New(`ARGUS_SERVICE_OPTIONS_INTERVAL: "10x" <invalid>`),
				errors.New(`ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE: "pizza" <invalid>`),
				errors.New(`ARGUS_WEBHOOK_DELAY: "10y" <invalid>`)),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN convertToEnvErrors is called with the input error
			got := convertToEnvErrors(tc.input)

			// THEN the result should match the expected error
			if got == nil && tc.expected == nil {
				return
			}
			if got == nil || tc.expected == nil || got.Error() != tc.expected.Error() {
				t.Errorf("convertToEnvErrors() got \n%q\nwant:\n%q", got, tc.expected)
			}
		})
	}
}
