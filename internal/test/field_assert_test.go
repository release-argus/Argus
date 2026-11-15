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
	"regexp"
	"testing"
	"time"

	"github.com/release-argus/Argus/util/errfmt"
)

func TestAssertFields(t *testing.T) {
	val := "test"
	ptr := &val
	val2 := "test2"
	ptr2 := &val2
	instant := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		fieldTest      FieldAssertion
		errRegex       string
		targetOverride *string
	}{
		{
			name:      "CompareEqual success",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareEqual},
			errRegex:  `^$`,
		},
		{
			name:      "CompareEqual mismatch",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val2, Mode: CompareEqual},
			errRegex:  `^target\.NAME mismatch\ngot:  [^\s]+.*\nwant: [^\s]+.*$`,
		},
		{
			name:      "CompareEqual uncomparable type",
			fieldTest: FieldAssertion{Name: "NAME", Got: []string{"a"}, Want: []string{"a"}, Mode: CompareEqual},
			errRegex:  `^target\.NAME: \[\]string is not comparable with ==, use Check$`,
		},
		{
			name:      "CompareEqual success/time.Time by Equal",
			fieldTest: FieldAssertion{Name: "NAME", Got: instant, Want: instant.In(time.FixedZone("X", 3600)), Mode: CompareEqual},
			errRegex:  `^$`,
		},
		{
			name:      "CompareEqual mismatch/time.Time",
			fieldTest: FieldAssertion{Name: "NAME", Got: instant, Want: instant.Add(time.Second), Mode: CompareEqual},
			errRegex:  `^target\.NAME mismatch\ngot:  .+\nwant: .+$`,
		},
		{
			name:      "CompareEqual success/[]byte by bytes.Equal",
			fieldTest: FieldAssertion{Name: "NAME", Got: []byte("ab"), Want: []byte("ab"), Mode: CompareEqual},
			errRegex:  `^$`,
		},
		{
			name:      "CompareEqual mismatch/[]byte",
			fieldTest: FieldAssertion{Name: "NAME", Got: []byte("ab"), Want: []byte("ac"), Mode: CompareEqual},
			errRegex:  `^target\.NAME mismatch\ngot:  .+\nwant: .+$`,
		},
		{
			name:      "CompareNotEqual mismatch/[]byte",
			fieldTest: FieldAssertion{Name: "NAME", Got: []byte("ab"), Want: []byte("ab"), Mode: CompareNotEqual},
			errRegex:  `^target\.NAME mismatch\ngot:  .+\nwant: NOT .+$`,
		},
		{
			name:      "CompareSamePointer uncomparable type/[]byte",
			fieldTest: FieldAssertion{Name: "NAME", Got: []byte("ab"), Want: []byte("ab"), Mode: CompareSamePointer},
			errRegex:  `^target\.NAME: \[\]uint8 is not comparable with ==, use Check$`,
		},
		{
			name:      "CompareNotEqual success",
			fieldTest: FieldAssertion{Name: "NAME", Got: "abc", Want: "123", Mode: CompareNotEqual},
			errRegex:  `^$`,
		},
		{
			name:      "CompareNotEqual mismatch",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareNotEqual},
			errRegex:  `^target\.NAME mismatch\ngot:  [^\s]+.*\nwant: NOT [^\s]+.*$`,
		},
		{
			name:      "CompareSamePointer success",
			fieldTest: FieldAssertion{Name: "NAME", Got: ptr, Want: ptr, Mode: CompareSamePointer},
			errRegex:  `^$`,
		},
		{
			name:      "CompareSamePointer mismatch",
			fieldTest: FieldAssertion{Name: "NAME", Got: ptr, Want: ptr2, Mode: CompareSamePointer},
			errRegex:  `^target\.NAME mismatch\ngot:  [^\s]+.*\nwant: [^\s]+.*$`,
		},
		{
			name:      "CompareSamePointer success/both nil",
			fieldTest: FieldAssertion{Name: "NAME", Got: (*string)(nil), Want: (*string)(nil), Mode: CompareSamePointer},
			errRegex:  `^$`,
		},
		{
			name:      "CompareDifferentPointer success",
			fieldTest: FieldAssertion{Name: "NAME", Got: ptr, Want: ptr2, Mode: CompareDifferentPointer},
			errRegex:  `^$`,
		},
		{
			name:      "CompareDifferentPointer mismatch",
			fieldTest: FieldAssertion{Name: "NAME", Got: ptr, Want: ptr, Mode: CompareDifferentPointer},
			errRegex:  `^target\.NAME mismatch\ngot:  [^\s]+.*\nwant: NOT [^\s]+.*$`,
		},
		{
			name:      "CompareDifferentPointer mismatch/both nil",
			fieldTest: FieldAssertion{Name: "NAME", Got: (*string)(nil), Want: (*string)(nil), Mode: CompareDifferentPointer},
			errRegex:  `^target\.NAME mismatch\ngot:  [^\s]+.*\nwant: NOT [^\s]+.*$`,
		},
		{
			name:      "CompareDifferentPointer success/nil want",
			fieldTest: FieldAssertion{Name: "NAME", Got: ptr, Want: (*string)(nil), Mode: CompareDifferentPointer},
			errRegex:  `^$`,
		},
		{
			name:      "CompareNotNil success",
			fieldTest: FieldAssertion{Name: "NAME", Got: ptr, Mode: CompareNotNil},
			errRegex:  `^$`,
		},
		{
			name:      "CompareNotNil mismatch/untyped nil",
			fieldTest: FieldAssertion{Name: "NAME", Mode: CompareNotNil},
			errRegex:  `^target\.NAME mismatch\ngot:  nil\nwant: non-nil$`,
		},
		{
			name:      "CompareNotNil mismatch/typed nil pointer",
			fieldTest: FieldAssertion{Name: "NAME", Got: (*string)(nil), Mode: CompareNotNil},
			errRegex:  `^target\.NAME mismatch\ngot:  nil\nwant: non-nil$`,
		},
		{
			name:      "CompareNotNil mismatch/nil map",
			fieldTest: FieldAssertion{Name: "NAME", Got: map[string]string(nil), Mode: CompareNotNil},
			errRegex:  `^target\.NAME mismatch\ngot:  nil\nwant: non-nil$`,
		},
		{
			name: ".Check() success",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareEqual, Check: func() (any, any, bool) {
				return val, val, true
			}},
			errRegex: `^$`,
		},
		{
			name: ".Check() mismatch",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareEqual, Check: func() (any, any, bool) {
				return val, val2, false
			}},
			errRegex: `target\.NAME mismatch\ngot:  .*\nwant: .*`,
		},
		{
			name: "no target param",
			fieldTest: FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareEqual, Check: func() (any, any, bool) {
				return val, val2, false
			}},
			errRegex:       `^NAME mismatch\ngot:  .*\nwant: .*`,
			targetOverride: new(""),
		},
		{
			name: "FieldAssertion.Target overrides target param",
			fieldTest: FieldAssertion{
				Name:   "NAME",
				Got:    val,
				Want:   val2,
				Mode:   CompareEqual,
				Target: "custom_target",
			},
			errRegex: `custom_target\.NAME mismatch\ngot:  .*\nwant: .*`,
		},
		{
			name:           "CompareEqual mismatch/strings quoted",
			fieldTest:      FieldAssertion{Name: "NAME", Got: "", Want: " ", Mode: CompareEqual},
			errRegex:       `^NAME mismatch\ngot:  ""\nwant: " "$`,
			targetOverride: new(""),
		},
		{
			name:           "CompareEqual mismatch/nil rendered as nil",
			fieldTest:      FieldAssertion{Name: "NAME", Got: nil, Want: "x", Mode: CompareEqual},
			errRegex:       `^NAME mismatch\ngot:  nil\nwant: "x"$`,
			targetOverride: new(""),
		},
		{
			name:           "CompareNotEqual mismatch/no target",
			fieldTest:      FieldAssertion{Name: "NAME", Got: val, Want: val, Mode: CompareNotEqual},
			errRegex:       `^NAME mismatch\ngot:  .*\nwant: NOT .*`,
			targetOverride: new(""),
		},
		{
			name:           "CompareSamePointer mismatch/no target",
			fieldTest:      FieldAssertion{Name: "NAME", Got: ptr, Want: ptr2, Mode: CompareSamePointer},
			errRegex:       `^NAME mismatch\ngot:  .*\nwant: .*`,
			targetOverride: new(""),
		},
		{
			name:           "CompareDifferentPointer mismatch/no target",
			fieldTest:      FieldAssertion{Name: "NAME", Got: ptr, Want: ptr, Mode: CompareDifferentPointer},
			errRegex:       `^NAME mismatch\ngot:  .*\nwant: NOT .*`,
			targetOverride: new(""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := "target"
			if tc.targetOverride != nil {
				target = *tc.targetOverride
			}
			fieldTests := []FieldAssertion{tc.fieldTest}

			err := AssertFields(
				t,
				fieldTests,
				target,
			)
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestAssertFields__rowTargetDoesNotLeak(t *testing.T) {
	fieldTests := []FieldAssertion{
		{Name: "FIRST", Got: 1, Want: 2, Mode: CompareEqual, Target: "custom_target"},
		{Name: "SECOND", Got: 1, Want: 2, Mode: CompareEqual},
	}

	err := AssertFields(t, fieldTests, "target")
	e := errfmt.FormatError(err)
	want := `(?m)^target\.SECOND mismatch$`
	if !regexp.MustCompile(want).MatchString(e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, want)
	}
}

func TestCopiedPointerMode(t *testing.T) {
	val := "test"

	tests := []struct {
		name string
		want any
		mode CompareMode
	}{
		{name: "untyped nil", want: nil, mode: CompareSamePointer},
		{name: "typed nil pointer", want: (*string)(nil), mode: CompareSamePointer},
		{name: "nil map", want: map[string]string(nil), mode: CompareSamePointer},
		{name: "non-nil pointer", want: &val, mode: CompareDifferentPointer},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := CopiedPointerMode(tc.want)

			if got != tc.mode {
				t.Errorf("mismatch\ngot:  %d\nwant: %d", got, tc.mode)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want string
	}{
		{name: "nil", v: nil, want: "nil"},
		{name: "string", v: "a", want: `"a"`},
		{name: "string slice", v: []string{"a", "b"}, want: `["a" "b"]`},
		{name: "string map", v: map[string]string{"k": "v"}, want: `map["k":"v"]`},
		{name: "map with a non-string key", v: map[int]string{1: "a"}, want: "map[1:a]"},
		{name: "map with a non-string value", v: map[string]int{"k": 1}, want: "map[k:1]"},
		{name: "slice with a non-string element", v: []any{"a", 1}, want: "[a 1]"},
		{name: "int", v: 1, want: "1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := formatValue(tc.v)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
