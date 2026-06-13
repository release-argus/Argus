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

package test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestAssertDecode(t *testing.T) {
	// GIVEN: a function to decode.
	f := func(format string, data []byte) (string, error) {
		if string(data) == "fail" {
			return "", errors.New("fail")
		}
		return string(data), nil
	}
	tests := []struct {
		name                                   string
		data                                   string
		stringify                              func(string) string
		want                                   string
		decodeErrRegex, testErrRegex           string
		decodeErrMismatch, stringifiedMismatch bool
	}{
		{
			name:           "success",
			data:           "hello",
			stringify:      genericStringify[string],
			want:           "hello\n",
			decodeErrRegex: `^$`,
			testErrRegex:   `^$`,
		},
		{
			name:           "fail that stringifies as expected and matches errRegex",
			data:           "fail",
			stringify:      genericStringify[string],
			want:           "",
			decodeErrRegex: `^fail$`,
			testErrRegex:   `^$`,
		},
		{
			name:              "errRegex mismatch",
			data:              "fail",
			stringify:         genericStringify[string],
			want:              "",
			decodeErrRegex:    `^regex unmatched$`,
			decodeErrMismatch: true,
			testErrRegex: TrimYAML(`
				test
				.* error mismatch`,
			),
		},
		{
			name:                "stringified want mismatch",
			data:                "hi",
			stringify:           genericStringify[string],
			want:                "abc",
			stringifiedMismatch: true,
			decodeErrRegex:      `^$`,
			testErrRegex: TrimYAML(`
				test
				.* stringified mismatch`,
			),
		},
		{
			name:           "stringify function missing",
			data:           "",
			stringify:      nil,
			want:           "",
			decodeErrRegex: `^$`,
			testErrRegex:   `stringify function is required`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format := "yaml"

			// WHEN: AssertDecode is called.
			decoded, decodeErr, testErr := AssertDecode(
				t,
				f,
				format,
				tc.data,
				tc.stringify,
				tc.want,
				tc.decodeErrRegex,
				packageName,
				"AssertDecode",
			)

			prefix := fmt.Sprintf(
				"%s\nAssertDecode(f, format=%q, data=%q)",
				packageName, format, tc.data,
			)

			// THEN: The result is as expected.
			if !(decoded == "" && tc.want == "") {
				gotBytes, _ := decode.Marshal("yaml", decoded)
				gotStr := string(gotBytes)
				stringifiedMatch := gotStr == tc.want
				if tc.stringifiedMismatch {
					if stringifiedMatch {
						t.Fatalf(
							"%s stringified result shouldn't match\ngot:  %q\nwant: NOT %q",
							prefix, gotStr, tc.want,
						)
					}
				} else if !stringifiedMatch {
					t.Fatalf(
						"%s stringified mismatch\ngot:  %q\nwant: %q",
						prefix, gotStr, tc.want,
					)
				}
			}

			// AND: The decode error is as expected.
			e := errfmt.FormatError(decodeErr)
			decodeErrMatch := regexp.MustCompile(tc.decodeErrRegex).MatchString(e)
			if tc.decodeErrMismatch {
				if decodeErrMatch {
					t.Fatalf(
						"%s decode error regex shouldn't match\ngot:  %q\nwant: NOT %q",
						prefix, e, tc.decodeErrRegex,
					)
				}
			} else if !decodeErrMatch {
				t.Fatalf(
					"%s decode error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.decodeErrRegex,
				)
			}

			// AND: The test error is as expected.
			e = errfmt.FormatError(testErr)
			if !regexp.MustCompile(tc.testErrRegex).MatchString(e) {
				t.Fatalf(
					"%s test error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.testErrRegex,
				)
			}
		})
	}
}

func TestAssertApplyOverrides(t *testing.T) {
	// GIVEN: a function to apply overrides to a struct.
	f := func(format string, data []byte, v *testStruct) (*testStruct, error) {
		if string(data) == "fail" {
			return nil, errors.New("fail")
		}

		if err := Unmarshal(format, data, v); err != nil {
			return nil, err
		}

		return v, nil
	}
	tests := []struct {
		name                                      string
		data                                      string
		stringify                                 func(*testStruct) string
		want                                      string
		overridesErrRegex, testErrRegex           string
		overridesErrMismatch, stringifiedMismatch bool
	}{
		{
			name: "success",
			data: TrimYAML(`
				string: hello
				int: 1
				bool: true
			`),
			stringify: genericStringify[*testStruct],
			want: TrimYAML(`
				string: hello
				int: 1
				bool: true
			`),
			overridesErrRegex: `^$`,
			testErrRegex:      `^$`,
		},
		{
			name:              "fail that stringifies as expected and matches errRegex",
			data:              "fail",
			stringify:         genericStringify[*testStruct],
			want:              "null\n",
			overridesErrRegex: `^fail$`,
			testErrRegex:      `^$`,
		},
		{
			name:                 "errRegex mismatch",
			data:                 "fail",
			stringify:            genericStringify[*testStruct],
			want:                 "",
			overridesErrRegex:    `^regex unmatched$`,
			overridesErrMismatch: true,
			testErrRegex: TrimYAML(`
				test
				.* error mismatch`,
			),
		},
		{
			name: "stringified want mismatch",
			data: TrimYAML(`
				string: hello
				int: 1
				bool: true
			`),
			stringify: genericStringify[*testStruct],
			want: TrimYAML(`
				string: bye
				int: 2
				bool: false
			`),
			stringifiedMismatch: true,
			overridesErrRegex:   `^$`,
			testErrRegex: TrimYAML(`
				test
				.* stringified mismatch`,
			),
		},
		{
			name:              "stringify function missing",
			data:              "",
			stringify:         nil,
			want:              "",
			overridesErrRegex: `^$`,
			testErrRegex:      `stringify function is required`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format := "yaml"
			v := testStruct{}

			// WHEN: AssertApplyOverrides is called.
			overridden, overridesErr, testErr := AssertApplyOverrides(
				t,
				&v,
				f,
				format,
				tc.data,
				tc.stringify,
				tc.want,
				tc.overridesErrRegex,
				true,
				packageName,
				"AssertApplyOverrides",
			)

			prefix := fmt.Sprintf(
				"%s\nAssertApplyOverrides(f, format=%q, data=%q)",
				packageName, format, tc.data,
			)

			// THEN: The test error is as expected.
			e := errfmt.FormatError(testErr)
			if !regexp.MustCompile(tc.testErrRegex).MatchString(e) {
				t.Fatalf(
					"%s test error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.testErrRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: The result is as expected.
			gotBytes, _ := decode.Marshal("yaml", overridden)
			gotStr := string(gotBytes)
			stringifiedMatch := gotStr == tc.want
			if tc.stringifiedMismatch {
				if stringifiedMatch {
					t.Fatalf(
						"%s stringified result shouldn't match\ngot:  %q\nwant: NOT %q",
						prefix, gotStr, tc.want,
					)
				}
			} else if !stringifiedMatch {
				t.Fatalf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, tc.want,
				)
			}

			// AND: The overrides error is as expected.
			e = errfmt.FormatError(overridesErr)
			overridesErrMatch := regexp.MustCompile(tc.overridesErrRegex).MatchString(e)
			if tc.overridesErrMismatch {
				if overridesErrMatch {
					t.Fatalf(
						"%s overrides error regex shouldn't match\ngot:  %q\nwant: NOT %q",
						prefix, e, tc.overridesErrRegex,
					)
				}
			} else if !overridesErrMatch {
				t.Fatalf(
					"%s overrides error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.overridesErrRegex,
				)
			}
		})
	}
}

func TestAssertApplyOverrides_AddressMatch(t *testing.T) {
	type tStruct interface {
		Foo() string
	}

	// GIVEN: a function to apply overrides to a struct.
	f := func(format string, data []byte, v tStruct) (tStruct, error) {
		if err := Unmarshal(format, data, v); err != nil {
			return nil, err
		}

		if v.Foo() == "same" {
			return v, nil
		}
		v = &testStruct{}
		err := Unmarshal(format, data, v)

		return v, err
	}
	tests := []struct {
		name                                      string
		data                                      string
		want                                      string
		sameAddress                               bool
		overridesErrRegex, testErrRegex           string
		overridesErrMismatch, stringifiedMismatch bool
	}{
		{
			name: "address unchanged",
			data: TrimYAML(`
				string: same
				int: 1
				bool: true
			`),
			want: TrimYAML(`
				string: same
				int: 1
				bool: true
			`),
			sameAddress:       true,
			overridesErrRegex: `^$`,
			testErrRegex:      `^$`,
		},
		{
			name: "want address change, but didn't",
			data: TrimYAML(`
				string: same
				int: 1
				bool: true
			`),
			want: TrimYAML(`
				string: same
				int: 1
				bool: true
			`),
			sameAddress:       false,
			overridesErrRegex: `^$`,
			testErrRegex: TrimYAML(`
				test
				.* pointer mismatch.*should have changed`,
			),
		},
		{
			name: "want same address, but changed",
			data: TrimYAML(`
				string: diff
				int: 1
				bool: true
			`),
			want: TrimYAML(`
				string: diff
				int: 1
				bool: true
			`),
			sameAddress:       true,
			overridesErrRegex: `^$`,
			testErrRegex: TrimYAML(`
				test
				.* pointer mismatch.*unexpected`,
			),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format := "yaml"
			var v tStruct
			v = &testStruct{}

			// WHEN: AssertApplyOverrides is called.
			overridden, overridesErr, testErr := AssertApplyOverrides(
				t,
				v,
				f,
				format,
				tc.data,
				genericStringify[tStruct],
				tc.want,
				tc.overridesErrRegex,
				tc.sameAddress,
				packageName,
				"AssertApplyOverrides",
			)

			prefix := fmt.Sprintf(
				"%s\nAssertApplyOverrides(f, format=%q, data=%q)",
				packageName, format, tc.data,
			)

			// THEN: The test error is as expected.
			e := errfmt.FormatError(testErr)
			if !regexp.MustCompile(tc.testErrRegex).MatchString(e) {
				t.Fatalf(
					"%s test error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.testErrRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: The result is as expected.
			gotBytes, _ := decode.Marshal("yaml", overridden)
			gotStr := string(gotBytes)
			stringifiedMatch := gotStr == tc.want
			if tc.stringifiedMismatch {
				if stringifiedMatch {
					t.Fatalf(
						"%s stringified result shouldn't match\ngot:  %q\nwant: NOT %q",
						prefix, gotStr, tc.want,
					)
				}
			} else if !stringifiedMatch {
				t.Fatalf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, gotStr, tc.want,
				)
			}

			// AND: The overrides error is as expected.
			e = errfmt.FormatError(overridesErr)
			overridesErrMatch := regexp.MustCompile(tc.overridesErrRegex).MatchString(e)
			if tc.overridesErrMismatch {
				if overridesErrMatch {
					t.Fatalf(
						"%s overrides error regex shouldn't match\ngot:  %q\nwant: NOT %q",
						prefix, e, tc.overridesErrRegex,
					)
				}
			} else if !overridesErrMatch {
				t.Fatalf(
					"%s overrides error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.overridesErrRegex,
				)
			}
		})
	}
}

func TestTestStruct_Foo(t *testing.T) {
	// GIVEN: a testStruct with a String.
	v := testStruct{
		String: "TestTestStruct_Foo",
	}
	// WHEN: Foo is called.
	got := v.Foo()
	// THEN: that String value is returned.
	if got != v.String {
		t.Errorf(
			"%s\ntestStruct.Foo() value mismatch\ngot:  %q\nwant: %q",
			packageName, got, v.String,
		)
	}
}
