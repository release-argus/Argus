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

package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

type testInterface interface{}

type testStruct struct {
	Int testInterface `yaml:"iface"`
}
type testStructChild struct {
	String string  `yaml:"string,omitzero"`
	Int    int     `yaml:"int,omitzero"`
	Float  float64 `yaml:"float,omitzero"`
}

func TestMigrateDeprecatedEnvVars(t *testing.T) {
	tests := []struct {
		name    string
		envVars []string
		want    map[string]string
	}{
		{
			name: "no deprecated vars present",
			envVars: []string{
				"ARGUS_FOO=bar",
			},
			want: map[string]string{
				"ARGUS_FOO": "bar",
			},
		},
		{
			name: "deprecated var present, canonical absent - renamed",
			envVars: []string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN=ghp_deprecated",
			},
			want: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_GITHUB_ACCESS_TOKEN": "ghp_deprecated",
			},
		},
		{
			name: "deprecated and canonical both present - canonical wins, deprecated dropped",
			envVars: []string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN=ghp_deprecated",
				"ARGUS_SERVICE_LATEST_VERSION_GITHUB_ACCESS_TOKEN=ghp_canonical",
			},
			want: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_GITHUB_ACCESS_TOKEN": "ghp_canonical",
			},
		},
		{
			name: "all exact-name deprecated vars are renamed",
			envVars: []string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN=ghp_deprecated",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE=true",
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS=true",
				"ARGUS_UNRELATED=untouched",
				"other=also_untouched",
			},
			want: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_GITHUB_ACCESS_TOKEN":     "ghp_deprecated",
				"ARGUS_SERVICE_LATEST_VERSION_GITHUB_USE_PRERELEASE":   "true",
				"ARGUS_SERVICE_LATEST_VERSION_URL_ALLOW_INVALID_CERTS": "true",
				"ARGUS_UNRELATED": "untouched",
				"other":           "also_untouched",
			},
		},
		{
			name: "deprecated 'require' prefix var is delegated to the prefix migration",
			envVars: []string{
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE=hub",
			},
			want: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_COMMON_REQUIRE_DOCKER_TYPE": "hub",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars.

			envVars := test.SplitEnvVars(tc.envVars)
			test.IsolateEnv(t, envVars)

			got := migrateDeprecatedEnvVars(tc.envVars)

			gotSorted, wantSorted := append([]string{}, got...), test.MapJoin(tc.want, "=")
			sort.Strings(gotSorted)
			if !slices.Equal(gotSorted, wantSorted) {
				t.Errorf("env vars mismatch\ngot:  %q\nwant: %q", gotSorted, wantSorted)
			}

			for k, v := range tc.want {
				if gotEnv := os.Getenv(k); gotEnv != v {
					t.Errorf("os.Getenv(%q) mismatch\ngot:  %q\nwant: %q", k, gotEnv, v)
				}
			}

			for k := range envVars {
				if _, kept := tc.want[k]; kept {
					continue
				}
				if v, ok := os.LookupEnv(k); ok {
					t.Errorf("os.LookupEnv(%q) mismatch\ngot:  present (%q)\nwant: absent", k, v)
				}
			}
		})
	}
}

func TestMigrateDeprecatedEnvVarPrefixes(t *testing.T) {
	tests := []struct {
		name    string
		envVars []string
		want    map[string]string
	}{
		{
			name: "no deprecated prefix present",
			envVars: []string{
				"ARGUS_FOO=bar",
			},
			want: map[string]string{
				"ARGUS_FOO": "bar",
			},
		},
		{
			name: "multiple sub-keys under the deprecated prefix are all renamed",
			envVars: []string{
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE=hub",
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_IMAGE=foo/bar",
				"UNRELATED=untouched",
			},
			want: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_COMMON_REQUIRE_DOCKER_TYPE":  "hub",
				"ARGUS_SERVICE_LATEST_VERSION_COMMON_REQUIRE_DOCKER_IMAGE": "foo/bar",
				"UNRELATED": "untouched",
			},
		},
		{
			name: "deprecated and canonical sub-key both present - canonical wins, deprecated dropped",
			envVars: []string{
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE=deprecated_value",
				"ARGUS_SERVICE_LATEST_VERSION_COMMON_REQUIRE_DOCKER_TYPE=canonical_value",
			},
			want: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_COMMON_REQUIRE_DOCKER_TYPE": "canonical_value",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars.

			envVars := test.SplitEnvVars(tc.envVars)
			test.IsolateEnv(t, envVars)

			got := migrateDeprecatedEnvVarPrefixes(tc.envVars)

			gotSorted, wantSorted := append([]string{}, got...), test.MapJoin(tc.want, "=")
			sort.Strings(gotSorted)
			if !slices.Equal(gotSorted, wantSorted) {
				t.Errorf("env vars mismatch\ngot:  %q\nwant: %q", gotSorted, wantSorted)
			}

			for k, v := range tc.want {
				if gotEnv := os.Getenv(k); gotEnv != v {
					t.Errorf("os.Getenv(%q) mismatch\ngot:  %q\nwant: %q", k, gotEnv, v)
				}
			}

			for k := range envVars {
				if _, kept := tc.want[k]; kept {
					continue
				}
				if v, ok := os.LookupEnv(k); ok {
					t.Errorf("os.LookupEnv(%q) mismatch\ngot:  present (%q)\nwant: absent", k, v)
				}
			}
		})
	}
}

// testMapEntry exercises map-of-struct env mapping: a nested struct, a pointer
// field, and a field the environment cannot name.
type testMapEntry struct {
	URL  string `yaml:"url,omitzero"`
	Auth struct {
		Token    string `yaml:"token,omitzero"`
		Insecure *bool  `yaml:"insecure,omitzero"`
	} `yaml:"auth,omitzero"`
	Internal int `yaml:"-"`
}

func TestMapEnvToStruct(t *testing.T) {
	tests := []struct {
		name         string
		customStruct any
		env          map[string]string
		marshalWant  bool
		want         string
		errRegex     string
	}{
		{
			name: "no env vars",
			env:  map[string]string{},
			customStruct: &struct {
				Test struct {
					String string `yaml:"string"`
					Int    int    `yaml:"int"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					string: ''
					int: 0
			`),
			errRegex: `^$`,
		},
		{
			name: "no ARGUS_ env vars",
			env: map[string]string{
				"TEST_STRING": "foo",
				"TEST_INT":    "1",
			},
			customStruct: &struct {
				Test struct {
					String string `yaml:"string"`
					Int    int    `yaml:"int"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					string: ''
					int: 0
			`),
			errRegex: `^$`,
		},
		{
			name: "nil non-comparable pointer",
			env: map[string]string{
				"ARGUS_TEST_SLICE": "1,2,3",
				"ARGUS_TEST_MAP":   "foo:1,bar:2",
				"ARGUS_TEST_FUNC":  "func()",
				"ARGUS_TEST_INT":   "1",
			},
			customStruct: &struct {
				Test struct {
					PtrToSlice *[]int          `yaml:"slice"`
					PtrToMap   *map[string]int `yaml:"map"`
					PtrToFunc  *func()         `yaml:"func"`
					PtrToInt   *int            `yaml:"int"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					slice: null
					map: null
					func: null
					int: 1
			`),
			errRegex: `^$`,
		},
		{
			name: "non-nil pointer to a string slice",
			env: map[string]string{
				"ARGUS_TEST_SLICE": "a, b,",
			},
			customStruct: func() any {
				custom := &struct {
					Test struct {
						Slice *[]string `yaml:"slice"`
					} `yaml:"test"`
				}{}
				custom.Test.Slice = &[]string{}
				return custom
			}(),
			want: test.TrimYAML(`
				test:
					slice:
						- a
						- b
			`),
			errRegex: `^$`,
		},
		{
			name: "ignore env vars under '-' tags",
			env: map[string]string{
				"ARGUS_TEST_ONE": "a",
			},
			customStruct: &struct {
				Test struct {
					String struct {
						Val3 bool `yaml:"one"`
					} `yaml:"-"`
				} `yaml:"test"`
			}{},
			want:     "test: {}\n",
			errRegex: `^$`,
		},
		{
			name: "boolean",
			env: map[string]string{
				"ARGUS_TEST_BOOLEAN_PTR0": "false",
				"ARGUS_TEST_BOOLEAN_PTR1": "f",
				"ARGUS_TEST_BOOLEAN_PTR2": "0",
				"ARGUS_TEST_BOOLEAN_PTR3": "",
				"ARGUS_TEST_BOOLEAN_VAL0": "true",
				"ARGUS_TEST_BOOLEAN_VAL1": "t",
				"ARGUS_TEST_BOOLEAN_VAL2": "1",
				"ARGUS_TEST_BOOLEAN_VAL3": "",
			},
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
		{
			name: "integer",
			env: map[string]string{
				"ARGUS_TEST_INTEGER_PTR0": "0",
				"ARGUS_TEST_INTEGER_PTR1": "1",
				"ARGUS_TEST_INTEGER_PTR2": "-1",
				"ARGUS_TEST_INTEGER_PTR3": "",
				"ARGUS_TEST_INTEGER_VAL0": "0",
				"ARGUS_TEST_INTEGER_VAL1": "1",
				"ARGUS_TEST_INTEGER_VAL2": "-1",
				"ARGUS_TEST_INTEGER_VAL3": "",
			},
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
		{
			name: "string",
			env: map[string]string{
				"ARGUS_TEST_STRING_PTR0": "foo",
				"ARGUS_TEST_STRING_PTR1": "",
				"ARGUS_TEST_STRING_VAL":  "bar",
			},
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
			errRegex: `^$`,
		},
		{
			name: "uint8/valid",
			env: map[string]string{
				"ARGUS_TEST_UNSIGNED_INTEGER_8_PTR0": "0",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_PTR1": "1",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_PTR2": "",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_VAL0": "0",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_VAL1": "1",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_VAL2": "",
			},
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
		{
			name: "uint8/invalid",
			env: map[string]string{
				"ARGUS_TEST_UNSIGNED_INTEGER_8_PTR_INVALID": "1024",
				"ARGUS_TEST_UNSIGNED_INTEGER_8_VAL_INVALID": "-1",
			},
			customStruct: &struct {
				Test struct {
					String struct {
						Ptr0 *uint8 `yaml:"ptr_invalid"`
						Val0 uint8  `yaml:"val_invalid"`
					} `yaml:"unsigned_integer_8"`
				} `yaml:"test"`
			}{},
			errRegex: test.TrimYAML(`
				^ARGUS_TEST_UNSIGNED_INTEGER_8_PTR_INVALID: "1024" <invalid>.*
				ARGUS_TEST_UNSIGNED_INTEGER_8_VAL_INVALID: "-1" <invalid>.*$`,
			),
		},
		{
			name: "uint16/valid",
			env: map[string]string{
				"ARGUS_TEST_UNSIGNED_INTEGER_16_PTR0": "0",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_PTR1": "1",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_PTR2": "",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_VAL0": "0",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_VAL1": "1",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_VAL2": "",
			},
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
		{
			name: "uint16/invalid",
			env: map[string]string{
				"ARGUS_TEST_UNSIGNED_INTEGER_16_PTR_INVALID": "65536",
				"ARGUS_TEST_UNSIGNED_INTEGER_16_VAL_INVALID": "-1",
			},
			customStruct: &struct {
				Test struct {
					String struct {
						Ptr0 *uint16 `yaml:"ptr_invalid"`
						Val0 uint16  `yaml:"val_invalid"`
					} `yaml:"unsigned_integer_16"`
				} `yaml:"test"`
			}{},
			errRegex: test.TrimYAML(`
				^ARGUS_TEST_UNSIGNED_INTEGER_16_PTR_INVALID: "65536" <invalid>.*
				ARGUS_TEST_UNSIGNED_INTEGER_16_VAL_INVALID: "-1" <invalid>.*$`,
			),
		},
		{
			name: "float/unsupported type",
			env: map[string]string{
				"ARGUS_TEST_FLOAT": "1.23",
			},
			customStruct: &struct {
				Test struct {
					Float *float64 `yaml:"float"`
				} `yaml:"test"`
			}{},
			errRegex: `unsupported env var kind on ARGUS_TEST_FLOAT: float64`,
		},
		{
			name: "string slice/valid",
			env: map[string]string{
				"ARGUS_TEST_SLICE": "alpha,beta",
			},
			customStruct: &struct {
				Test struct {
					Slice []string `yaml:"slice"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					slice:
						- alpha
						- beta
			`),
			errRegex: `^$`,
		},
		{
			name: "string slice/trims spaces and drops blanks",
			env: map[string]string{
				"ARGUS_TEST_SLICE": " alpha , ,beta, ",
			},
			customStruct: &struct {
				Test struct {
					Slice []string `yaml:"slice"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					slice:
						- alpha
						- beta
			`),
			errRegex: `^$`,
		},
		{
			name: "string slice/empty value gives an empty slice",
			env: map[string]string{
				"ARGUS_TEST_SLICE": "",
			},
			customStruct: &struct {
				Test struct {
					Slice []string `yaml:"slice"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					slice: []
			`),
			errRegex: `^$`,
		},
		{
			name: "string slice/unsupported element type",
			env: map[string]string{
				"ARGUS_TEST_SLICE": "1.23",
			},
			customStruct: &struct {
				Test struct {
					Slice []float64 `yaml:"slice"`
				} `yaml:"test"`
			}{},
			errRegex: `ARGUS_TEST_SLICE: unsupported slice element kind: float64`,
		},
		{
			name: "inline struct/valid",
			env: map[string]string{
				"ARGUS_TEST_INLINE_STRING": "foo",
			},
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
		{
			name: "inline struct/error",
			env: map[string]string{
				"ARGUS_TEST_INLINE_INT": "foo",
			},
			customStruct: &struct {
				Test struct {
					Inline struct {
						Integer int `yaml:"int"`
					} `yaml:",inline"`
				} `yaml:"test_inline"`
			}{},
			errRegex: `^ARGUS_TEST_INLINE_INT: "foo" <invalid>.*$`,
		},
		{
			name: "map/error",
			env: map[string]string{
				"ARGUS_MAP_FOO_BOOL": "maybe",
			},
			customStruct: &struct {
				Map map[string]struct {
					Bool *bool `yaml:"bool"`
				} `yaml:"map"`
			}{
				Map: map[string]struct {
					Bool *bool `yaml:"bool"`
				}{
					"foo": {},
				},
			},
			errRegex: `ARGUS_MAP_FOO_BOOL: "maybe" <invalid>`,
		},
		{
			name: "map/value-typed struct elements are written back",
			env: map[string]string{
				"ARGUS_MAP_FOO_STRING": "bar",
				"ARGUS_MAP_FOO_BOOL":   "true",
			},
			customStruct: &struct {
				Map map[string]struct {
					String string `yaml:"string"`
					Bool   *bool  `yaml:"bool"`
				} `yaml:"map"`
			}{
				Map: map[string]struct {
					String string `yaml:"string"`
					Bool   *bool  `yaml:"bool"`
				}{
					"foo": {},
				},
			},
			want: test.TrimYAML(`
				map:
					foo:
						string: bar
						bool: true
			`),
			errRegex: `^$`,
		},
		{
			name: "map/pointer-typed struct elements are written back",
			env: map[string]string{
				"ARGUS_MAP_FOO_STRING": "bar",
			},
			customStruct: &struct {
				Map map[string]*struct {
					String string `yaml:"string"`
				} `yaml:"map"`
			}{
				Map: map[string]*struct {
					String string `yaml:"string"`
				}{
					"foo": {},
				},
			},
			want: test.TrimYAML(`
				map:
					foo:
						string: bar
			`),
			errRegex: `^$`,
		},
		{
			name: "map of pointers/entries are created from their variables",
			env: map[string]string{
				"ARGUS_MAP_ABSENT_STRING": "bar",
			},
			customStruct: &struct {
				Map map[string]*struct {
					String string `yaml:"string"`
				} `yaml:"map"`
			}{
				Map: map[string]*struct {
					String string `yaml:"string"`
				}{
					"present": {},
				},
			},
			want: test.TrimYAML(`
				map:
					absent:
						string: bar
					present:
						string: ''
			`),
			errRegex: `^$`,
		},
		{
			name: "map of strings/the rest of the name is a lower-cased key",
			env: map[string]string{
				"ARGUS_TEST_OPTIONS_ABC":     "123",
				"ARGUS_TEST_OPTIONS_MAX_TRY": "7",
			},
			customStruct: &struct {
				Test struct {
					Options map[string]string `yaml:"options,omitempty"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					options:
						abc: '123'
						max_try: '7'
			`),
			errRegex: `^$`,
		},
		{
			name: "map of structs/entries are created from their variables",
			env: map[string]string{
				"ARGUS_TEST_HOST_CODEBERG_URL":          "https://codeberg.org",
				"ARGUS_TEST_HOST_CODEBERG_AUTH_TOKEN":   "cb-token",
				"ARGUS_TEST_HOST_MY_WORK_URL":           "https://git.example.com",
				"ARGUS_TEST_HOST_MY_WORK_AUTH_INSECURE": "true",
			},
			customStruct: &struct {
				Test struct {
					Host map[string]testMapEntry `yaml:"host,omitempty"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					host:
						codeberg:
							url: https://codeberg.org
							auth:
								token: cb-token
						my_work:
							url: https://git.example.com
							auth:
								insecure: true
			`),
			errRegex: `^$`,
		},
		{
			name: "map of structs/the longest field match names the key",
			env: map[string]string{
				"ARGUS_TEST_HOST_CODEBERG_AUTH_TOKEN": "cb-token",
			},
			customStruct: &struct {
				Test struct {
					Host map[string]struct {
						Auth struct {
							Token string `yaml:"token,omitzero"`
						} `yaml:"auth,omitzero"`
						Token string `yaml:"token,omitzero"`
					} `yaml:"host,omitempty"`
				} `yaml:"test"`
			}{},
			want: test.TrimYAML(`
				test:
					host:
						codeberg:
							auth:
								token: cb-token
			`),
			errRegex: `^$`,
		},
		{
			name: "map of structs/an existing entry is extended (whatever its case), not duplicated",
			env: map[string]string{
				"ARGUS_TEST_HOST_CODEBERG_AUTH_TOKEN": "cb-token",
			},
			customStruct: func() any {
				custom := &struct {
					Test struct {
						Host map[string]testMapEntry `yaml:"host,omitempty"`
					} `yaml:"test"`
				}{}
				custom.Test.Host = map[string]testMapEntry{
					"Codeberg": {URL: "https://codeberg.org"},
				}
				return custom
			}(),
			want: test.TrimYAML(`
				test:
					host:
						Codeberg:
							url: https://codeberg.org
							auth:
								token: cb-token
			`),
			errRegex: `^$`,
		},
		{
			name: "map of structs/a variable naming no known field creates nothing",
			env: map[string]string{
				"ARGUS_TEST_HOST_CODEBERG_USERNAME": "someone",
			},
			customStruct: &struct {
				Test struct {
					Host map[string]testMapEntry `yaml:"host,omitempty"`
				} `yaml:"test"`
			}{},
			want:     "test: {}\n",
			errRegex: `^$`,
		},
		{
			name: "map of non-structs/nothing is created for an unsupported element type",
			env: map[string]string{
				"ARGUS_TEST_MAP_FOO": "1",
			},
			customStruct: &struct {
				Test struct {
					Map map[string]int `yaml:"map,omitempty"`
				} `yaml:"test"`
			}{},
			want:     "test: {}\n",
			errRegex: `^$`,
		},
		{
			name: "map of structs/a map key with no fields creates nothing",
			env: map[string]string{
				"ARGUS_TEST_HOST_CODEBERG": "anything",
			},
			customStruct: &struct {
				Test struct {
					Host map[string]struct{} `yaml:"host,omitempty"`
				} `yaml:"test"`
			}{},
			want:     "test: {}\n",
			errRegex: `^$`,
		},
		{
			name: "map of structs/an invalid value is reported",
			env: map[string]string{
				"ARGUS_TEST_HOST_CODEBERG_AUTH_INSECURE": "bop",
			},
			customStruct: &struct {
				Test struct {
					Host map[string]testMapEntry `yaml:"host,omitempty"`
				} `yaml:"test"`
			}{},
			errRegex: `^ARGUS_TEST_HOST_CODEBERG_AUTH_INSECURE: "bop" <invalid> `,
		},
		{
			name: "map/nested value-typed struct elements are written back",
			env: map[string]string{
				"ARGUS_MAP_FOO_INNER_BAR_STRING": "baz",
			},
			customStruct: &struct {
				Map map[string]struct {
					Inner map[string]struct {
						String string `yaml:"string"`
					} `yaml:"inner"`
				} `yaml:"map"`
			}{
				Map: map[string]struct {
					Inner map[string]struct {
						String string `yaml:"string"`
					} `yaml:"inner"`
				}{
					"foo": {
						Inner: map[string]struct {
							String string `yaml:"string"`
						}{
							"bar": {},
						},
					},
				},
			},
			want: test.TrimYAML(`
				map:
					foo:
						inner:
							bar:
								string: baz
			`),
			errRegex: `^$`,
		},
		{
			name: "struct that was nil/error",
			env: map[string]string{
				"ARGUS_STRUCT_BOOL": "sometimes",
			},
			customStruct: &struct {
				Struct *struct {
					Bool *bool `yaml:"bool"`
				} `yaml:"struct"`
			}{},
			errRegex: `ARGUS_STRUCT_BOOL: "sometimes" <invalid>`,
		},
		{
			name: "interface/valid",
			env: map[string]string{
				"ARGUS_IFACE_STRING": "foo",
			},
			customStruct: &testStruct{
				Int: &testStructChild{},
			},
			marshalWant: true,
			want: test.TrimYAML(`
				iface:
					string: foo
			`),
			errRegex: `^$`,
		},
		{
			name: "interface/invalid structure",
			env: map[string]string{
				"ARGUS_IFACE_INT": "1.1",
			},
			customStruct: &testStruct{
				Int: &testStructChild{},
			},
			marshalWant: true,
			errRegex:    `^ARGUS_IFACE_INT: "1.1" <invalid>.*$`,
		},
		{
			name: "interface/nil struct not mapped",
			env: map[string]string{
				"ARGUS_IFACE_STRING": "foo",
				"ARGUS_IFACE_INT":    "1",
				"ARGUS_IFACE_FLOAT":  "2.3",
			},
			customStruct: &testStruct{
				Int: nil,
			},
			marshalWant: true,
			want:        "iface: null\n",
			errRegex:    `^$`,
		},
		{
			name: "map vars/ARGUS_NOTIFY_",
			env: map[string]string{
				"ARGUS_NOTIFY_DISCORD_OPTIONS_DELAY":        "2s",
				"ARGUS_NOTIFY_MATTERMOST_OPTIONS_MAX_TRIES": "7",
				"ARGUS_NOTIFY_MATTERMOST_URL_FIELDS_A":      "foo",
				"ARGUS_NOTIFY_MATTERMOST_PARAMS_A":          "bar",
			},
			customStruct: func() any {
				cfg := Config{
					Notify: shoutrrr.ShoutrrrsDefaults{},
				}

				defaults := shoutrrr.ShoutrrrsDefaults{}
				defaults.Default()
				for typ := range defaults {
					cfg.Notify[typ] = &shoutrrr.Defaults{}
					cfg.Notify[typ].InitMaps()
				}

				return &cfg
			}(),
			marshalWant: true,
			want: func() string {
				var str strings.Builder
				str.WriteString("notify:\n")
				defaults := shoutrrr.ShoutrrrsDefaults{}
				defaults.Default()
				types := util.SortedKeys(defaults)
				for _, typ := range types {
					str.WriteString("  ")
					str.WriteString(typ)
					str.WriteString(": {}\n")
				}

				val := str.String()
				d := strings.ReplaceAll(
					test.TrimYAML(`
						discord:
							options:
								delay: 2s`,
					),
					"\n", "\n  ",
				)
				val = strings.Replace(val, "discord: {}", d, 1)
				mm := strings.ReplaceAll(
					test.TrimYAML(`
						mattermost:
							options:
								max_tries: '7'
							url_fields:
								a: foo
							params:
								a: bar`,
					),
					"\n", "\n  ",
				)
				val = strings.Replace(val, "mattermost: {}", mm, 1)

				return val
			}(),
			errRegex: `^$`,
		},
		{
			name: "deprecated service.latest_version env vars are migrated before mapping",
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN":        "ghp_deprecated",
				"ARGUS_SERVICE_LATEST_VERSION_USE_PRERELEASE":      "true",
				"ARGUS_SERVICE_LATEST_VERSION_ALLOW_INVALID_CERTS": "true",
			},
			customStruct: &Defaults{},
			want: test.TrimYAML(`
				service:
					latest_version:
						github:
							access_token: ghp_deprecated
							use_prerelease: true
						url:
							allow_invalid_certs: true
			`),
			errRegex: `^$`,
		},
		{
			name: "canonical service.latest_version env vars need no migration",
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_GITHUB_ACCESS_TOKEN": "ghp_canonical",
			},
			customStruct: &Defaults{},
			want: test.TrimYAML(`
				service:
					latest_version:
						github:
							access_token: ghp_canonical
			`),
			errRegex: `^$`,
		},
		{
			name: "deprecated and canonical service.latest_version env vars both set, canonical wins",
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN":        "ghp_deprecated",
				"ARGUS_SERVICE_LATEST_VERSION_GITHUB_ACCESS_TOKEN": "ghp_canonical",
			},
			customStruct: &Defaults{},
			want: test.TrimYAML(`
				service:
					latest_version:
						github:
							access_token: ghp_canonical
			`),
			errRegex: `^$`,
		},
		{
			name: "deprecated service.latest_version.require env vars (prefix-mapped) are migrated before mapping",
			env: map[string]string{
				"ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE": "hub",
			},
			customStruct: &Defaults{},
			want: test.TrimYAML(`
				service:
					latest_version:
						common:
							require:
								docker:
									type: hub
			`),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars.

			test.IsolateEnv(t, tc.env)

			err := mapEnvToStruct(tc.customStruct, "", nil)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			gotYAML, producer := "", "decode.ToYAMLString()"
			if tc.marshalWant {
				producer = "decode.Marshal()"
				marshalled, err := decode.Marshal("yaml", tc.customStruct)
				if err != nil {
					t.Fatalf("decode.Marshal() failed: %v", err)
				}
				gotYAML = string(marshalled)
			} else {
				gotYAML = decode.ToYAMLString(tc.customStruct, "")
			}
			if gotYAML != tc.want {
				t.Errorf("%s mismatch\ngot:  %q\nwant: %q", producer, gotYAML, tc.want)
			}
		})
	}
}

func TestMapEnvToStruct_NoEnvVars(t *testing.T) {
	test.IsolateEnv(t, nil)
	// mapEnvToStruct collects every ARGUS-prefixed var, but IsolateEnv keeps ARGUS_TEST_*.
	for k := range test.SplitEnvVars(os.Environ()) {
		if strings.HasPrefix(k, "ARGUS") {
			_ = os.Unsetenv(k)
		}
	}

	var cfg Config

	err := mapEnvToStruct(&cfg, "", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	want := "{}\n"
	if got := decode.ToYAMLString(&cfg, ""); got != want {
		t.Errorf("decode.ToYAMLString() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestConvertToEnvErrors(t *testing.T) {
	tests := []struct {
		name        string
		input, want error
	}{
		{
			name:  "nil error",
			input: nil,
			want:  nil,
		},
		{
			name: "single error",
			input: errors.New(test.TrimYAML(`
				service:
					options:
						interval: "10x" <invalid>
			`)),
			want: errors.New(`ARGUS_SERVICE_OPTIONS_INTERVAL: "10x" <invalid>`),
		},
		{
			name: "multiple errors",
			input: errors.New(test.TrimYAML(`
				service:
					options:
						interval: "10x" <invalid>
					latest_version:
						require:
							docker:
								type: "pizza" <invalid>
				webhook:
					delay: "10y" <invalid>
			`)),
			want: errors.Join(
				errors.New(`ARGUS_SERVICE_OPTIONS_INTERVAL: "10x" <invalid>`),
				errors.New(`ARGUS_SERVICE_LATEST_VERSION_REQUIRE_DOCKER_TYPE: "pizza" <invalid>`),
				errors.New(`ARGUS_WEBHOOK_DELAY: "10y" <invalid>`),
			),
		},
		{
			name: "a line naming no field is kept whole",
			input: errors.New(test.TrimYAML(`
				service:
					latest_version:
						forgejo:
							host:
								something
			`)),
			want: errors.New(
				`ARGUS_SERVICE_LATEST_VERSION_FORGEJO_HOST: something`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := convertToEnvErrors(tc.input)

			if got == nil &&
				tc.want == nil {
				return
			}
			if got == nil ||
				tc.want == nil ||
				got.Error() != tc.want.Error() {
				t.Errorf("error mismatch\ngot:  %v\nwant: %v", got, tc.want)
			}
		})
	}
}

func TestLoadEnvFile(t *testing.T) {
	tests := []struct {
		name           string
		content        *string
		cannotReadFile bool
		want           map[string]string
		doNotWant      []string
		errRegex       string
	}{
		{
			name:     "no file",
			errRegex: `^$`,
		},
		{
			name:     "empty file",
			content:  new(""),
			want:     map[string]string{},
			errRegex: `^$`,
		},
		{
			name:           "cannot read file",
			content:        new("FOO=bar"),
			cannotReadFile: true,
			doNotWant:      []string{"FOO"},
			errRegex:       `failed to open env file `,
		},
		{
			name: "comments and empty lines",
			content: new(test.TrimYAML(`
				# comment

					# indented comment
				# comment=123
				FOO=bar
			`)),
			want: map[string]string{
				"FOO": "bar",
			},
			doNotWant: []string{"# comment", " comment", "comment"},
			errRegex:  `^$`,
		},
		{
			name: "basic key-value pairs",
			content: new(test.TrimYAML(`
				FOO=bar
				BAR=baz
			`)),
			want: map[string]string{
				"FOO": "bar",
				"BAR": "baz",
			},
			errRegex: `^$`,
		},
		{
			name: "export prefix",
			content: new(test.TrimYAML(`
				export FOO=bar
				export  BAR=test
				export=argus
			`)),
			want: map[string]string{
				"FOO":    "bar",
				"BAR":    "test",
				"export": "argus",
			},
			errRegex: `^$`,
		},
		{
			name: "quoted values",
			content: new(test.TrimYAML(`
				FOO="bar"
				BAR='123'
			`)),
			want: map[string]string{
				"FOO": "bar",
				"BAR": "123",
			},
			errRegex: `^$`,
		},
		{
			name: "unmatched quotes are kept",
			content: new(test.TrimYAML(`
				FOO="bar
				BAR=baz"
				BAZ='123
				QUX=123'
			`)),
			want: map[string]string{
				"FOO": `"bar`,
				"BAR": `baz"`,
				"BAZ": `'123`,
				"QUX": `123'`,
			},
			errRegex: `^$`,
		},
		{
			name: "env var expansion",
			content: new(test.TrimYAML(`
				FOO=bar
				BAR=${FOO}
			`)),
			want: map[string]string{
				"FOO": "bar",
				"BAR": "bar",
			},
			errRegex: `^$`,
		},
		{
			name: "invalid line format",
			content: new(test.TrimYAML(`
				FOO=bar
				invalid_line
			`)),
			errRegex: `invalid env line: "invalid_line"`,
		},
		{
			name: "invalid env var key",
			content: new(test.TrimYAML(`
				FOO=bar
				=baz
			`)),
			errRegex: `failed to set env var "":`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: sets env vars.

			tmpDir := t.TempDir()
			test.IsolateEnv(t, nil)

			filePath := filepath.Join(tmpDir, "nonexistent.env")
			if tc.content != nil {
				filePath = filepath.Join(tmpDir, ".env")
				if err := os.WriteFile(filePath, []byte(test.TrimYAML(*tc.content)), 0644); err != nil {
					t.Fatalf("os.WriteFile() failed: %v", err)
				}
			}
			if tc.cannotReadFile {
				if err := os.Chmod(filePath, 0000); err != nil {
					t.Fatalf("os.Chmod() failed: %v", err)
				}
			}

			err := loadEnvFile(filePath)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			for k, v := range tc.want {
				if got := os.Getenv(k); got != v {
					t.Errorf("os.Getenv(%q) mismatch\ngot:  %q\nwant: %q", k, got, v)
				}
			}
			for _, k := range tc.doNotWant {
				if got, ok := os.LookupEnv(k); ok {
					t.Errorf("os.LookupEnv(%q) mismatch\ngot:  present (%q)\nwant: absent", k, got)
				}
			}
		})
	}
}

type failingReader struct {
	r         io.Reader
	failAt    int
	readSoFar int
}

func (f *failingReader) Read(p []byte) (int, error) {
	if f.readSoFar >= f.failAt {
		return 0, errors.New("simulated read error")
	}
	n, err := f.r.Read(p)
	f.readSoFar += n
	return n, err
}

func TestLoadEnvFile_ReadError(t *testing.T) {
	reader := &failingReader{
		r:      strings.NewReader("FOO=bar\ntest=123\n"),
		failAt: 10,
	}
	test.IsolateEnv(t, nil)

	err := loadEnvFromReader(reader)
	if err == nil {
		t.Fatal("error mismatch\ngot:  nil\nwant: non-nil")
	}
}
