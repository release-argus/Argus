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
	"regexp"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestAssertDecode(t *testing.T) {
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
			testErrRegex:      `error mismatch`,
		},
		{
			name:                "stringified want mismatch",
			data:                "hi",
			stringify:           genericStringify[string],
			want:                "abc",
			stringifiedMismatch: true,
			decodeErrRegex:      `^$`,
			testErrRegex:        `^AssertDecode\(\) mismatch`,
		},
		{
			name:                "missing error and stringified want mismatch are both reported",
			data:                "hi",
			stringify:           genericStringify[string],
			want:                "abc",
			decodeErrRegex:      `^fail$`,
			decodeErrMismatch:   true,
			stringifiedMismatch: true,
			testErrRegex:        `^error mismatch\n(.*\n)+AssertDecode\(\) mismatch\n`,
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

			decoded, decodeErr, testErr := AssertDecode(
				t,
				f,
				format,
				tc.data,
				tc.stringify,
				tc.want,
				tc.decodeErrRegex,
				"AssertDecode",
			)

			if decoded != "" || tc.want != "" {
				gotBytes, err := decode.Marshal("yaml", decoded)
				if err != nil {
					t.Fatalf("decode.Marshal() failed: %v", err)
				}
				gotStr := string(gotBytes)
				stringifiedMatch := gotStr == tc.want
				if tc.stringifiedMismatch {
					if stringifiedMatch {
						t.Fatalf(
							"decode.Marshal() mismatch\ngot:  %q\nwant: NOT %q",
							gotStr, tc.want,
						)
					}
				} else if !stringifiedMatch {
					t.Fatalf("decode.Marshal() mismatch\ngot:  %q\nwant: %q", gotStr, tc.want)
				}
			}

			e := errfmt.FormatError(decodeErr)
			decodeErrMatch := regexp.MustCompile(tc.decodeErrRegex).MatchString(e)
			if tc.decodeErrMismatch {
				if decodeErrMatch {
					t.Fatalf("decode error mismatch\ngot:  %q\nwant: NOT %q", e, tc.decodeErrRegex)
				}
			} else if !decodeErrMatch {
				t.Fatalf("decode error mismatch\ngot:  %q\nwant: %q", e, tc.decodeErrRegex)
			}

			e = errfmt.FormatError(testErr)
			if !regexp.MustCompile(tc.testErrRegex).MatchString(e) {
				t.Fatalf("test error mismatch\ngot:  %q\nwant: %q", e, tc.testErrRegex)
			}
		})
	}
}

func TestAssertApplyOverrides(t *testing.T) {
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
		name                            string
		data                            string
		stringify                       func(*testStruct) string
		want                            string
		overridesErrRegex, testErrRegex string
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
			name:              "errRegex mismatch",
			data:              "fail",
			stringify:         genericStringify[*testStruct],
			want:              "",
			overridesErrRegex: `^regex unmatched$`,
			testErrRegex:      `error mismatch`,
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
			overridesErrRegex: `^$`,
			testErrRegex:      `^AssertApplyOverrides\(\) mismatch`,
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
				"AssertApplyOverrides",
			)

			e := errfmt.FormatError(testErr)
			if !regexp.MustCompile(tc.testErrRegex).MatchString(e) {
				t.Fatalf("test error mismatch\ngot:  %q\nwant: %q", e, tc.testErrRegex)
			}
			if testErr != nil {
				return
			}

			gotBytes, err := decode.Marshal("yaml", overridden)
			if err != nil {
				t.Fatalf("decode.Marshal() failed: %v", err)
			}
			if gotStr := string(gotBytes); gotStr != tc.want {
				t.Fatalf("decode.Marshal() mismatch\ngot:  %q\nwant: %q", gotStr, tc.want)
			}

			e = errfmt.FormatError(overridesErr)
			if !regexp.MustCompile(tc.overridesErrRegex).MatchString(e) {
				t.Fatalf("overrides error mismatch\ngot:  %q\nwant: %q", e, tc.overridesErrRegex)
			}
		})
	}
}

func TestAssertApplyOverrides__nonPointer(t *testing.T) {
	f := func(_ string, _ []byte, v testStruct) (testStruct, error) {
		return v, nil
	}
	tests := []struct {
		name         string
		want         string
		testErrRegex string
	}{
		{
			name:         "stringified want matches",
			want:         "string: hello\n",
			testErrRegex: `^AssertApplyOverrides\(\) address check needs a pointer-like type, got test\.testStruct$`,
		},
		{
			name: "stringified want mismatch",
			want: "string: bye\n",
			testErrRegex: TrimYAML(`
				^AssertApplyOverrides\(\) mismatch
				got:  \"string: hello\\n\"
				want: \"string: bye\\n\"
				AssertApplyOverrides\(\) address check needs a pointer-like type, got test\.testStruct$`),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, testErr := AssertApplyOverrides(
				t,
				testStruct{String: "hello"},
				f,
				"yaml",
				"",
				genericStringify[testStruct],
				tc.want,
				`^$`,
				true,
				"AssertApplyOverrides",
			)

			e := errfmt.FormatError(testErr)
			if !regexp.MustCompile(tc.testErrRegex).MatchString(e) {
				t.Errorf("test error mismatch\ngot:  %q\nwant: %q", e, tc.testErrRegex)
			}
		})
	}
}

func TestAssertApplyOverrides_AddressMatch(t *testing.T) {
	type tStruct interface {
		Foo() string
	}

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
		name                            string
		data                            string
		want                            string
		sameAddress                     bool
		overridesErrRegex, testErrRegex string
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
			testErrRegex:      `^AssertApplyOverrides\(\) address mismatch\ngot:  \S+\nwant: NOT \S+$`,
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
			testErrRegex:      `^AssertApplyOverrides\(\) address mismatch\ngot:  \S+\nwant: \S+$`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format := "yaml"
			var v tStruct = &testStruct{}

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
				"AssertApplyOverrides",
			)

			e := errfmt.FormatError(testErr)
			if !regexp.MustCompile(tc.testErrRegex).MatchString(e) {
				t.Fatalf("test error mismatch\ngot:  %q\nwant: %q", e, tc.testErrRegex)
			}
			if testErr != nil {
				return
			}

			gotBytes, err := decode.Marshal("yaml", overridden)
			if err != nil {
				t.Fatalf("decode.Marshal() failed: %v", err)
			}
			if gotStr := string(gotBytes); gotStr != tc.want {
				t.Fatalf("decode.Marshal() mismatch\ngot:  %q\nwant: %q", gotStr, tc.want)
			}

			e = errfmt.FormatError(overridesErr)
			if !regexp.MustCompile(tc.overridesErrRegex).MatchString(e) {
				t.Fatalf("overrides error mismatch\ngot:  %q\nwant: %q", e, tc.overridesErrRegex)
			}
		})
	}
}
