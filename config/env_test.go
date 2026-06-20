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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/util/errfmt"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

type testInterface interface{}

type testStruct struct {
	Int testInterface `yaml:"iface"`
}
type testStructChild struct {
	String string  `yaml:"string,omitempty"`
	Int    int     `yaml:"int,omitempty"`
	Float  float64 `yaml:"float,omitempty"`
}

func TestMapEnvToStruct(t *testing.T) {
	// GIVEN: a struct and a bunch of env vars.
	tests := []struct {
		name                string
		customStruct        any
		customStructBuilder *func() any
		prefix              string
		env                 map[string]string
		marshalWant         bool
		want                string
		errRegex            string
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
			want: "",
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
			want: "",
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
			want: "",
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
			want: test.TrimYAML(`
				test:
					unsigned_integer_8:
						ptr_invalid: null
						val_invalid: 0
			`),
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
			want: test.TrimYAML(`
				test:
					unsigned_integer_16:
						ptr_invalid: null
						val_invalid: 0
			`),
			errRegex: test.TrimYAML(`
				^ARGUS_TEST_UNSIGNED_INTEGER_16_PTR_INVALID: "65536" <invalid>.*
				ARGUS_TEST_UNSIGNED_INTEGER_16_VAL_INVALID: "-1" <invalid>.*$`,
			),
		},
		{
			name: "float - unsupported type",
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
			want: test.TrimYAML(`
				test_inline:
					int: 0
			`),
			errRegex: `^ARGUS_TEST_INLINE_INT: "foo" <invalid>.*$`,
		},
		{
			name: "map - error",
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
			name: "struct that was nil - error",
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
			want:        `{}`,
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
			name: "map vars - ARGUS_NOTIFY_",
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
					str.WriteString(fmt.Sprintf("  %s: {}\n", typ))
				}

				val := str.String()
				// Discord.
				d := strings.ReplaceAll(
					test.TrimYAML(`
						discord:
							options:
								delay: 2s`,
					),
					"\n", "\n  ",
				)
				val = strings.Replace(val, "discord: {}", d, 1)
				// MatterMost.
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing some env vars.

			test.SetEnv(t, tc.env)

			// WHEN: mapEnvToStruct is called on it.
			err := mapEnvToStruct(tc.customStruct, tc.prefix, nil)

			// THEN: any error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) { // Expected a FATAL panic to be caught above.
				t.Errorf(
					"%s\nMapEnvToStruct() error mismatch\ngot:  %q\nwant: %q",
					packageName, e, tc.errRegex,
				)
			}
			if tc.errRegex != "^$" {
				return
			}

			// AND: the defaults are set to the appropriate env vars.
			var gotYAML string
			if tc.marshalWant {
				got, _ := decode.Marshal("yaml", tc.customStruct)
				gotYAML = string(got)
			} else {
				gotYAML = decode.ToYAMLString(tc.customStruct, "")
			}
			if gotYAML != tc.want {
				t.Errorf(
					"%s\nMapEnvToStruct() stringified mismatch\ngot:  %q\nwant: %q",
					packageName, gotYAML, tc.want,
				)
			}
		})
	}
}

func TestMapEnvToStruct_NoEnvVars(t *testing.T) {
	// GIVEN: no env vars are set.
	envVars := os.Environ()
	originalEnv := make(map[string]string, len(envVars))
	for _, kv := range envVars {
		vals := strings.SplitN(kv, "=", 2)
		k, v := vals[0], vals[1]
		originalEnv[k] = v
		os.Unsetenv(k)
	}
	t.Logf(
		"%s\noriginal env vars pre-TestMapEnvToStruct_NoEnvVars (%d):\n%+v",
		packageName, len(originalEnv), originalEnv,
	)
	t.Cleanup(func() {
		for k, v := range originalEnv {
			os.Setenv(k, v)
		}
	})

	// AND: a config to map no env vars onto.
	var cfg Config

	// WHEN: mapEnvToStruct is called on it.
	err := mapEnvToStruct(&cfg, "", nil)

	prefix := fmt.Sprintf("%s\nmapEnvToStruct()", packageName)

	// THEN: any error is as expected.
	e := errfmt.FormatError(err)
	if !util.RegexCheck(`^$`, e) { // Expected a FATAL panic to be caught above.
		t.Errorf(
			"%s error mismatch\ngot:  %q\nwant: nil",
			prefix, e,
		)
	}

	// AND: the Config stringifies empty.
	want := "{}\n"
	if got := decode.ToYAMLString(&cfg, ""); got != want {
		t.Errorf(
			"%s with no env vars, Config mismatch\ngot:  %q\nwant: %q",
			prefix, got, want,
		)
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN: convertToEnvErrors is called with the input error.
			got := convertToEnvErrors(tc.input)

			// THEN: the result should match the expected error.
			if got == nil &&
				tc.want == nil {
				return
			}
			if got == nil ||
				tc.want == nil ||
				got.Error() != tc.want.Error() {
				t.Errorf(
					"%s\nconvertToEnvErrors() error mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestLoadEnvFile(t *testing.T) {
	// GIVEN: a file of environment variables.
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
			content:  test.Ptr(""),
			want:     map[string]string{},
			errRegex: "^$",
		},
		{
			name:           "cannot read file",
			content:        test.Ptr("FOO=bar"),
			cannotReadFile: true,
			doNotWant:      []string{"FOO"},
			errRegex:       `failed to open env file `,
		},
		{
			name: "comments and empty lines",
			content: test.Ptr(test.TrimYAML(`
				# comment

					# indented comment
				# comment=123
				FOO=bar
			`)),
			want: map[string]string{
				"FOO": "bar",
			},
			doNotWant: []string{"# comment", " comment", "comment"},
			errRegex:  "^$",
		},
		{
			name: "basic key-value pairs",
			content: test.Ptr(test.TrimYAML(`
				FOO=bar
				BAR=baz
			`)),
			want: map[string]string{
				"FOO": "bar",
				"BAR": "baz",
			},
			errRegex: "^$",
		},
		{
			name: "export prefix",
			content: test.Ptr(test.TrimYAML(`
				export FOO=bar
				export  BAR=test
				export=argus
			`)),
			want: map[string]string{
				"FOO":    "bar",
				"BAR":    "test",
				"export": "argus",
			},
			errRegex: "^$",
		},
		{
			name: "quoted values",
			content: test.Ptr(test.TrimYAML(`
				FOO="bar"
				BAR='123'
			`)),
			want: map[string]string{
				"FOO": "bar",
				"BAR": "123",
			},
			errRegex: "^$",
		},
		{
			name: "env var expansion",
			content: test.Ptr(test.TrimYAML(`
				FOO=bar
				BAR=${FOO}
			`)),
			want: map[string]string{
				"FOO": "bar",
				"BAR": "bar",
			},
			errRegex: "^$",
		},
		{
			name: "invalid line format",
			content: test.Ptr(test.TrimYAML(`
				FOO=bar
				invalid_line
			`)),
			errRegex: `invalid env line: "invalid_line"`,
		},
		{
			name: "invalid env var key",
			content: test.Ptr(test.TrimYAML(`
				FOO=bar
				=baz
			`)),
			errRegex: `failed to set env var "":`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're sharing some env vars.
			tmpDir := t.TempDir()

			// Create env file if content provided.
			filePath := filepath.Join(tmpDir, "nonexistent.env")
			if tc.content != nil {
				filePath = filepath.Join(tmpDir, ".env")
				if err := os.WriteFile(filePath, []byte(test.TrimYAML(*tc.content)), 0644); err != nil {
					t.Fatalf(
						"%s\n failed to create test file: %v",
						packageName, err,
					)
				}
			}
			if tc.cannotReadFile {
				_ = os.Chmod(filePath, 0000)
			}

			// WHEN: loadEnvFile is called.
			err := loadEnvFile(filePath)

			prefix := fmt.Sprintf("%s\nLoadEnvFile()", packageName)

			// THEN: any error matches expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: environment variables are set as expected.
			if tc.errRegex == "^$" {
				// Verify the expected env vars are set.
				for k, v := range tc.want {
					if got := os.Getenv(k); got != v {
						t.Errorf(
							"%s env var %q mismatch\ngot:  %q\nwant: %q",
							prefix, k,
							got, v,
						)
					}
				}
				// Verify unexpected env vars are not set.
				want := ""
				for _, k := range tc.doNotWant {
					if got := os.Getenv(k); got != want {
						t.Errorf(
							"%s\nenv var %q should not be set\ngot:  %q\nwant: %q",
							prefix, k,
							got, want,
						)
					}
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
	// GIVEN: a reader that fails after a certain number of bytes.
	content := "FOO=bar\ntest=123\n"
	reader := &failingReader{
		r:      strings.NewReader(content),
		failAt: 10, // fail after 10 bytes.
	}

	// WHEN: LoadEnvFile is called with the failing reader.
	err := loadEnvFromReader(reader)

	// THEN: an error is returned.
	if err == nil {
		t.Fatalf(
			"%s\nerror mismatch after LoadEnvFile that should fail\ngot:  nil\nwant: error",
			packageName,
		)
	}
}
