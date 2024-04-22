// Copyright [2023] [Argus]
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

package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/test"
)

func TestContains(t *testing.T) {
	// GIVEN lists of strings
	tests := map[string]struct {
		list        []string
		contain     string
		doesContain bool
	}{
		"[]string does contain": {
			list:    []string{"hello", "hi", "hiya"},
			contain: "hi", doesContain: true},
		"[]string does not contain": {
			list:    []string{"hello", "hi", "hiya"},
			contain: "howdy", doesContain: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Contains is run on this list with a element inside it
			var found bool
			found = Contains(tc.list, tc.contain)

			// THEN true is returned if it does contain the item
			if found != tc.doesContain {
				t.Errorf("want Contains=%t, got Contains=%t",
					found, tc.doesContain)
			}
		})
	}
}

func TestEvalNilPtr(t *testing.T) {
	// GIVEN lists of strings
	tests := map[string]struct {
		ptr    *string
		nilStr string
		want   string
	}{
		"nil *string": {
			ptr: nil, nilStr: "bar",
			want: "bar"},
		"non-nil *string": {
			ptr: test.StringPtr("foo"), nilStr: "bar",
			want: "foo"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN EvalNilPtr is run on a pointer
			got := EvalNilPtr(tc.ptr, tc.nilStr)

			// THEN the correct value is returned
			if got != tc.want {
				t.Errorf("want: %s\ngot:  %s",
					tc.want, got)
			}
		})
	}
}

func TestPtrOrValueToPtr(t *testing.T) {
	// GIVEN a pointer and a value
	tests := map[string]struct {
		a    *string
		b    string
		want string
	}{
		"nil `a` pointer": {
			a: nil, b: "bar",
			want: "bar"},
		"non-nil `a` pointer": {
			a: test.StringPtr("foo"), b: "bar",
			want: "foo"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN PtrOrValueToPtr is run on pointer and a value
			got := PtrOrValueToPtr(tc.a, tc.b)

			// THEN the correct value is returned
			if *got != tc.want {
				t.Errorf("want: %s\ngot:  %s",
					tc.want, *got)
			}
		})
	}
}

func TestValueIfNotNil(t *testing.T) {
	// GIVEN a value to check and a value we want when it's not nil
	tests := map[string]struct {
		check *string
		value string
		want  *string
	}{
		"nil `check` pointer": {
			check: nil, value: "foo",
			want: nil},
		"non-nil `check` pointer": {
			check: test.StringPtr("foo"), value: "bar",
			want: test.StringPtr("bar")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ValueIfNotNil is run on pointer and a value
			got := ValueIfNotNil(tc.check, tc.value)

			// THEN the correct value is returned
			if tc.want == nil {
				if got != nil {
					t.Errorf("want: %v\ngot:  &%q",
						tc.want, *got)
				}
				return
			}
			if got == nil {
				t.Errorf("want: %q\ngot:  &%v",
					*tc.want, got)
			}
			if *got != *tc.want {
				t.Errorf("want: %q\ngot:  %q",
					*tc.want, *got)
			}
		})
	}
}

func TestValueIfNotDefault(t *testing.T) {
	// GIVEN a value to check and a value we want when it's not default
	tests := map[string]struct {
		check string
		value string
		want  string
	}{
		"default `check` value": {
			check: "", value: "foo",
			want: ""},
		"non-default `check` value": {
			check: "foo", value: "bar",
			want: "bar"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ValueIfNotDefault is run on pointer and a value
			got := ValueIfNotDefault(tc.check, tc.value)

			// THEN the correct value is returned
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestDefaultIfNil(t *testing.T) {
	// GIVEN a value to check and a value we want when it's nil
	tests := map[string]struct {
		check *string
		value string
		want  string
	}{
		"nil `check` pointer": {
			check: nil,
			want:  ""},
		"non-nil `check` pointer": {
			check: test.StringPtr("foo"),
			want:  "foo"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN DefaultIfNil is run on pointer and a value
			got := DefaultIfNil(tc.check)

			// THEN the correct value is returned
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestFirstNonNilPtr(t *testing.T) {
	// GIVEN a bunch of pointers
	tests := map[string]struct {
		pointers  []*string
		allNil    bool
		wantIndex int
	}{
		"no pointers": {
			pointers: []*string{},
			allNil:   true,
		},
		"all nil pointers": {
			pointers: []*string{
				nil,
				nil,
				nil,
				nil},
			allNil: true,
		},
		"1 non-nil pointer": {
			pointers: []*string{
				nil,
				nil,
				nil,
				test.StringPtr("bar")},
			wantIndex: 3,
		},
		"2 non-nil pointers": {
			pointers: []*string{
				test.StringPtr("foo"),
				nil,
				nil,
				test.StringPtr("bar")},
			wantIndex: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN FirstNonNilPtr is run on a slice of pointers
			got := FirstNonNilPtr(tc.pointers...)

			// THEN the correct pointer (or nil) is returned
			if tc.allNil {
				if got != nil {
					t.Fatalf("got:  %v\nfrom: %v",
						got, tc.pointers)
				}
				return
			}
			if got != tc.pointers[tc.wantIndex] {
				t.Errorf("want: %v\ngot:  %v",
					tc.pointers[tc.wantIndex], got)
			}
		})
	}
}

func TestFirstNonNilPtrWithEnv(t *testing.T) {
	// GIVEN a bunch of pointers
	tests := map[string]struct {
		env         map[string]string
		pointers    []*string
		allNil      bool
		wantIndex   int
		wantText    string
		diffAddress bool
	}{
		"no pointers": {
			pointers: []*string{},
			allNil:   true,
		},
		"all nil pointers": {
			pointers: []*string{
				nil,
				nil,
				nil,
				nil},
			allNil: true,
		},
		"1 non-nil pointer": {
			pointers: []*string{
				nil,
				nil,
				nil,
				test.StringPtr("bar")},
			wantIndex: 3,
		},
		"1 non-nil pointer (env var)": {
			env: map[string]string{"TESTFIRSTNONNILPTRWITHENV_ONE": "bar"},
			pointers: []*string{
				nil,
				nil,
				nil,
				test.StringPtr("${TESTFIRSTNONNILPTRWITHENV_ONE}")},
			wantIndex:   3,
			diffAddress: true,
			wantText:    "bar",
		},
		"1 non-nil pointer (env var partial)": {
			env: map[string]string{"TESTFIRSTNONNILPTRWITHENV_TWO": "bar"},
			pointers: []*string{
				nil,
				nil,
				nil,
				test.StringPtr("foo${TESTFIRSTNONNILPTRWITHENV_TWO}")},
			wantIndex:   3,
			diffAddress: true,
			wantText:    "foobar",
		},
		"1 non-nil pointer (empty env var)": {
			env: map[string]string{"TESTFIRSTNONNILPTRWITHENV_THREE": ""},
			pointers: []*string{
				nil,
				nil,
				nil,
				test.StringPtr("${TESTFIRSTNONNILPTRWITHENV_THREE}")},
			wantIndex:   3,
			diffAddress: true,
			wantText:    "",
		},
		"1 non-nil pointer (unset env var)": {
			pointers: []*string{
				nil,
				nil,
				nil,
				test.StringPtr("${TESTFIRSTNONNILPTRWITHENV_UNSET}")},
			wantIndex:   3,
			diffAddress: false,
			wantText:    "${TESTFIRSTNONNILPTRWITHENV_UNSET}",
		},
		"2 non-nil pointers": {
			pointers: []*string{
				test.StringPtr("foo"),
				nil,
				nil,
				test.StringPtr("bar")},
			wantIndex: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// WHEN FirstNonNilPtrWithEnv is run on a slice of pointers
			got := FirstNonNilPtrWithEnv(tc.pointers...)

			// THEN the correct pointer (or nil) is returned
			if tc.allNil {
				if got != nil {
					t.Fatalf("got:  %v\nfrom: %v",
						got, tc.pointers)
				}
				return
			}
			// Addresses should be the same (unless we're using an env var)
			if got != tc.pointers[tc.wantIndex] && !tc.diffAddress {
				t.Errorf("want: %v\ngot:  %v",
					tc.pointers[tc.wantIndex], got)
				// Addresses should only be the same
			} else if got == tc.pointers[tc.wantIndex] {
				// IF we're using an env var
				if tc.diffAddress {
					t.Errorf("addresses of pointers should differ (%v, %v)",
						tc.pointers[tc.wantIndex], got)
				}
				// Should have what the env var is set to
			} else if *got != tc.wantText {
				t.Errorf("want: %v\ngot:  %v",
					tc.wantText, *got)
			}
		})
	}
}

func TestValueIfTrue(t *testing.T) {
	// GIVEN lists of strings
	tests := map[string]struct {
		list        []string
		contain     string
		doesContain bool
	}{
		"[]string does contain": {
			list:    []string{"hello", "hi", "hiya"},
			contain: "hi", doesContain: true},
		"[]string does not contain": {
			list:    []string{"hello", "hi", "hiya"},
			contain: "howdy", doesContain: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN Contains is run on this list with a element inside it
			var found bool
			found = Contains(tc.list, tc.contain)

			// THEN true is returned if it does contain the item
			if found != tc.doesContain {
				t.Errorf("want Contains=%t, got Contains=%t",
					found, tc.doesContain)
			}
		})
	}
}

func TestFirstNonDefault(t *testing.T) {
	// GIVEN a bunch of comparables
	tests := map[string]struct {
		slice      []string
		allDefault bool
		wantIndex  int
	}{
		"no vars": {
			slice:      []string{},
			allDefault: true,
		},
		"all default vars": {
			slice: []string{
				"",
				"",
				"",
				""},
			allDefault: true,
		},
		"1 non-default var": {
			slice: []string{
				"",
				"",
				"",
				"bar"},
			wantIndex: 3,
		},
		"2 non-default vars": {
			slice: []string{
				"foo",
				"",
				"",
				"bar"},
			wantIndex: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN FirstNonDefault is run on a slice of slice
			got := FirstNonDefault(tc.slice...)

			// THEN the correct var (or "") is returned
			if tc.allDefault {
				if got != "" {
					t.Fatalf("got:  %v\nfrom: %v",
						got, tc.slice)
				}
				return
			}
			if got != tc.slice[tc.wantIndex] {
				t.Errorf("want: %v\ngot:  %v",
					tc.slice[tc.wantIndex], got)
			}
		})
	}
}

func TestFirstNonDefaultWithEnv(t *testing.T) {
	// GIVEN a bunch of comparables
	tests := map[string]struct {
		env         map[string]string
		slice       []string
		allDefault  bool
		wantIndex   int
		wantText    string
		diffAddress bool
	}{
		"no vars": {
			slice:      []string{},
			allDefault: true,
		},
		"all default vars": {
			slice: []string{
				"",
				"",
				"",
				""},
			allDefault: true,
		},
		"1 non-default var": {
			slice: []string{
				"",
				"",
				"",
				"bar"},
			wantIndex: 3,
		},
		"1 non-default var (env var)": {
			env: map[string]string{"TESTFIRSTNONDEFAULTWITHENV_ONE": "bar"},
			slice: []string{
				"",
				"",
				"",
				"${TESTFIRSTNONDEFAULTWITHENV_ONE}"},
			wantIndex:   3,
			wantText:    "bar",
			diffAddress: true,
		},
		"1 non-default var (env var partial)": {
			env: map[string]string{"TESTFIRSTNONDEFAULTWITHENV_TWO": "bar"},
			slice: []string{
				"",
				"",
				"",
				"foo${TESTFIRSTNONDEFAULTWITHENV_TWO}"},
			wantIndex:   3,
			wantText:    "foobar",
			diffAddress: true,
		},
		"2 non-default vars": {
			slice: []string{
				"foo",
				"",
				"",
				"bar"},
			wantIndex: 0,
		},
		"2 non-default vars (empty env vars ignored)": {
			env: map[string]string{
				"TESTFIRSTNONDEFAULTWITHENV_THREE": "",
				"TESTFIRSTNONDEFAULTWITHENV_FOUR":  "bar"},
			slice: []string{
				"${TESTFIRSTNONDEFAULTWITHENV_THREE}",
				"${TESTFIRSTNONDEFAULTWITHENV_UNSET}",
				"",
				"${TESTFIRSTNONDEFAULTWITHENV_FOUR}"},
			wantIndex:   3,
			wantText:    "${TESTFIRSTNONDEFAULTWITHENV_UNSET}",
			diffAddress: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// WHEN FirstNonDefaultWithEnv is run on a slice of slice
			got := FirstNonDefaultWithEnv(tc.slice...)

			// THEN the correct var (or "") is returned
			if tc.allDefault {
				if got != "" {
					t.Fatalf("got:  %v\nfrom: %v",
						got, tc.slice)
				}
				return
			}
			// Addresses should be the same (unless we're using an env var)
			if got != tc.slice[tc.wantIndex] && !tc.diffAddress {
				t.Errorf("want: %v\ngot:  %v",
					tc.slice[tc.wantIndex], got)
				// Addresses should only be the same
			} else if got == tc.slice[tc.wantIndex] {
				// IF we're using an env var
				if tc.diffAddress {
					t.Errorf("addresses of pointers should differ (%v, %v)",
						tc.slice[tc.wantIndex], got)
				}
				// Should have what the env var is set to
			} else if got != tc.wantText {
				t.Errorf("want: %v\ngot:  %v",
					tc.wantText, got)
			}
		})
	}
}

func TestPrintlnIfNotDefault(t *testing.T) {
	// GIVEN a bunch of comparables
	tests := map[string]struct {
		element  string
		didPrint bool
	}{
		"default var": {
			element: "", didPrint: false},
		"non-default var": {
			element: "foo", didPrint: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			msg := "var is not default from PrintlnIfNotDefault"

			// WHEN PrintlnIfNotDefault is called
			PrintlnIfNotDefault(tc.element, msg)

			// THEN the var is printed when it should be
			stdout := releaseStdout()
			if !tc.didPrint {
				if stdout != "" {
					t.Fatalf("printed %q",
						stdout)
				}
				return
			}
			if stdout != msg+"\n" {
				t.Errorf("unexpected print %q",
					stdout)
			}
		})
	}
}

func TestPrintlnIfNotNil(t *testing.T) {
	// GIVEN a bunch of comparables
	tests := map[string]struct {
		element  *string
		didPrint bool
	}{
		"nil pointer": {
			element: nil, didPrint: false},
		"non-nil pointer": {
			element: test.StringPtr("foo"), didPrint: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			msg := "var is not default from PrintlnIfNotNil"

			// WHEN PrintlnIfNotNil is called
			PrintlnIfNotNil(tc.element, msg)

			// THEN the var is printed when it should be
			stdout := releaseStdout()
			if !tc.didPrint {
				if stdout != "" {
					t.Fatalf("printed %q",
						stdout)
				}
				return
			}
			if stdout != msg+"\n" {
				t.Errorf("unexpected print %q",
					stdout)
			}
		})
	}
}

func TestPrintlnIfNil(t *testing.T) {
	// GIVEN a bunch of comparables
	tests := map[string]struct {
		element  *string
		didPrint bool
	}{
		"nil pointer": {
			element: nil, didPrint: true},
		"non-nil pointer": {
			element: test.StringPtr("foo"), didPrint: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			msg := "var is not default from PrintlnIfNil"

			// WHEN PrintlnIfNil is called
			PrintlnIfNil(tc.element, msg)

			// THEN the var is printed when it should be
			stdout := releaseStdout()
			if !tc.didPrint {
				if stdout != "" {
					t.Fatalf("printed %q",
						stdout)
				}
				return
			}
			if stdout != msg+"\n" {
				t.Errorf("unexpected print %q",
					stdout)
			}
		})
	}
}

func TestDefaultOrValue(t *testing.T) {
	// GIVEN a bunch of comparables
	tests := map[string]struct {
		element *string
		value   string
		want    string
	}{
		"nil pointer": {
			element: nil, want: ""},
		"non-nil pointer": {
			element: test.StringPtr("foo"), value: "bar", want: "bar"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN DefaultOrValue is called
			got := DefaultOrValue(tc.element, tc.value)

			// THEN the var is printed when it should be
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestPtrValueOrValue(t *testing.T) {
	// GIVEN a bunch of comparables pointers and values
	tests := map[string]struct {
		ptr   interface{}
		value interface{}
		want  interface{}
	}{
		"nil string pointer": {
			ptr:   (*string)(nil),
			value: "argus", want: "argus"},
		"non-nil string pointer": {
			ptr:   test.StringPtr("foo"),
			value: "bar", want: "foo"},
		"nil bool pointer": {
			ptr:   (*bool)(nil),
			value: false, want: false},
		"non-nil bool pointer": {
			ptr:   test.BoolPtr(true),
			value: false, want: true},
		"nil int pointer": {
			ptr:   (*int)(nil),
			value: 1, want: 1},
		"non-nil int pointer": {
			ptr:   test.IntPtr(3),
			value: 2, want: 3},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN PtrValueOrValue is called
			var got interface{}
			switch v := tc.ptr.(type) {
			case *string:
				got = PtrValueOrValue(v, tc.value.(string))
			case *bool:
				got = PtrValueOrValue(v, tc.value.(bool))
			case *int:
				got = PtrValueOrValue(v, tc.value.(int))
			}

			// THEN the pointer is returned if it's nil, otherwise the value
			if got != tc.want {
				t.Errorf("\nwant: %v\ngot:  %v", tc.want, got)
			}
		})
	}
}

func TestErrorToString(t *testing.T) {
	// GIVEN a bunch of comparables
	tests := map[string]struct {
		err  error
		want string
	}{
		"nil error": {
			err: nil, want: ""},
		"non-nil error": {
			err: fmt.Errorf("test error"), want: "test error"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ErrorToString is called
			got := ErrorToString(tc.err)

			// THEN the var is printed when it should be
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestRandString(t *testing.T) {
	// GIVEN different size strings are wanted with different alphabets
	tests := map[string]struct {
		wanted   int
		alphabet string
	}{
		"length 1 string, length 1 alphabet": {
			wanted: 1, alphabet: "a"},
		"length 2, length 1 alphabet": {
			wanted: 2, alphabet: "b"},
		"length 3, length 1 alphabet": {
			wanted: 3, alphabet: "c"},
		"length 10, length 1 alphabet": {
			wanted: 10, alphabet: "d"},
		"length 10, length 5 alphabet": {
			wanted: 10, alphabet: "abcde"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN RandString is called
			got := RandString(tc.wanted, tc.alphabet)

			// THEN we get a random alphabet string of the specified length
			if len(got) != tc.wanted {
				t.Errorf("got length %d. wanted %d",
					tc.wanted, len(got))
			}
			charactersVerified := 0
			for charactersVerified != tc.wanted {
				var characters []string
				for i := range tc.alphabet {
					characters = append(characters, string(tc.alphabet[i]))
				}

				for i := range characters {
					if got == characters[i] {
						RemoveIndex(&characters, i)
						break
					}
				}
				charactersVerified++
			}
		})
	}
}

func TestRandAlphaNumericLower(t *testing.T) {
	// GIVEN different size strings are wanted
	tests := map[string]struct {
		wanted int
	}{
		"length 1": {
			wanted: 1},
		"length 2": {
			wanted: 2},
		"length 3": {
			wanted: 3},
		"length 10": {
			wanted: 10},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN RandAlphaNumericLower is called
			got := RandAlphaNumericLower(tc.wanted)

			// THEN we get a random alphanumeric string of the specified length
			if len(got) != tc.wanted {
				t.Errorf("got length %d. wanted %d",
					tc.wanted, len(got))
			}
			charactersVerified := 0
			for charactersVerified != tc.wanted {
				var characters []string
				for i := range alphanumericLower {
					characters = append(characters, string(alphanumericLower[i]))
				}

				for i := range characters {
					if got == characters[i] {
						RemoveIndex(&characters, i)
						break
					}
				}
				charactersVerified++
			}
		})
	}
}

func TestRandNumeric(t *testing.T) {
	// GIVEN different size strings are wanted
	tests := map[string]struct {
		wanted int
	}{
		"length 1": {
			wanted: 1},
		"length 2": {
			wanted: 2},
		"length 3": {
			wanted: 3},
		"length 10": {
			wanted: 10},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN RandNumeric is called
			got := RandNumeric(tc.wanted)

			// THEN we get a random numeric string of the specified length
			if len(got) != tc.wanted {
				t.Errorf("got length %d. wanted %d",
					tc.wanted, len(got))
			}
			charactersVerified := 0
			for charactersVerified != tc.wanted {
				var characters []string
				for i := range numeric {
					characters = append(characters, string(numeric[i]))
				}

				for i := range characters {
					if got == characters[i] {
						RemoveIndex(&characters, i)
						break
					}
				}
				charactersVerified++
			}
		})
	}
}

func TestNormaliseNewlines(t *testing.T) {
	// GIVEN different byte strings
	tests := map[string]struct {
		input []byte
		want  []byte
	}{
		"string with no newlines": {
			input: []byte("hello there"),
			want:  []byte("hello there")},
		"string with linux newlines": {
			input: []byte("hello\nthere"),
			want:  []byte("hello\nthere")},
		"string with multiple linux newlines": {
			input: []byte("hello\nthere\n"),
			want:  []byte("hello\nthere\n")},
		"string with windows newlines": {
			input: []byte("hello\r\nthere"),
			want:  []byte("hello\nthere")},
		"string with multiple windows newlines": {
			input: []byte("hello\r\nthere\r\n"),
			want:  []byte("hello\nthere\n")},
		"string with mac newlines": {
			input: []byte("hello\r\nthere"),
			want:  []byte("hello\nthere")},
		"string with multiple mac newlines": {
			input: []byte("hello\r\nthere\r\n"),
			want:  []byte("hello\nthere\n")},
		"string with multiple mac and windows newlines": {
			input: []byte("\rhello\r\nthere\r\n. hi\r"),
			want:  []byte("\nhello\nthere\n. hi\n")},
		"string with multiple mac, windows and linux newlines": {
			input: []byte("\rhello\r\nthere\r\n. hi\r. foo\nbar\n"),
			want:  []byte("\nhello\nthere\n. hi\n. foo\nbar\n")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN NormaliseNewlines is called
			got := NormaliseNewlines(tc.input)

			// THEN the newlines are normalised correctly
			if string(got) != string(tc.want) {
				t.Errorf("want: %q\ngot:  %q",
					string(tc.want), string(got))
			}
		})
	}
}

func TestCopyIfSecret(t *testing.T) {
	// GIVEN maps with secrets to be copied
	tests := map[string]struct {
		input    map[string]string
		copyFrom map[string]string
		want     map[string]string
		fields   []string
	}{
		"empty map": {
			input: map[string]string{},
			copyFrom: map[string]string{
				"foo": "bar"},
			want:   map[string]string{},
			fields: []string{"foo"},
		},
		"copy only '<secret>'s in fields": {
			input: map[string]string{
				"test": "<secret>",
				"foo":  "<secret>"},
			copyFrom: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "<secret>"},
			fields: []string{"test"},
		},
		"copy only '<secret>'s in fields that also exist in from": {
			input: map[string]string{
				"test": "<secret>",
				"foo":  "<secret>",
				"bar":  "<secret>"},
			copyFrom: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "<secret>",
				"bar":  "<secret>"},
			fields: []string{"test", "bar"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CopyIfSecret is called
			CopyIfSecret(tc.copyFrom, tc.input, tc.fields)

			// THEN the secrets are copied correctly
			if len(tc.input) != len(tc.want) {
				t.Fatalf("want: %v\ngot:  %v",
					tc.want, tc.input)
			}
			for i := range tc.input {
				if tc.input[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, tc.input)
				}
			}
		})
	}
}

func TestInitMap(t *testing.T) {
	// GIVEN a map
	tests := map[string]struct {
		input map[string]string
	}{
		"nil map": {
			input: nil,
		},
		"empty map": {
			input: map[string]string{},
		},
		"non-empty map": {
			input: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"non-empty map with same keys but differing case": {
			input: map[string]string{
				"test": "123",
				"tESt": "bar"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			had := CopyMap(tc.input)

			// WHEN InitMap is called
			InitMap(&tc.input)

			// THEN the map is initialised correctly
			if tc.input == nil {
				t.Fatalf("map is nil")
			}
			// AND any values inside haven't changed
			if len(tc.input) != len(had) {
				t.Fatalf("want: %v\ngot:  %v",
					had, tc.input)
			}
			for i := range tc.input {
				if tc.input[i] != had[i] {
					t.Fatalf("want: %v\ngot:  %v",
						had, tc.input)
				}
			}
		})
	}

}

func TestCopyMap(t *testing.T) {
	// GIVEN different byte strings
	tests := map[string]struct {
		input map[string]string
		want  map[string]string
	}{
		"empty map": {
			input: map[string]string{},
			want:  map[string]string{},
		},
		"non-empty map": {
			input: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"non-empty map with same keys but differing case": {
			input: map[string]string{
				"test": "123",
				"tESt": "bar"},
			want: map[string]string{
				"test": "123",
				"tESt": "bar"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CopyMap is called
			got := CopyMap(tc.input)

			// THEN the map is copied correctly
			if &got == &tc.want {
				t.Error("map wasn't copied, they have the same addresses")
			}
			if len(got) != len(tc.want) {
				t.Fatalf("want: %v\ngot:  %v",
					tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, got)
				}
			}
		})
	}
}

func TestMergeMaps(t *testing.T) {
	// GIVEN two maps and a list of fields that may contain secrets
	tests := map[string]struct {
		base      map[string]string
		overrides map[string]string
		fields    []string
		want      map[string]string
	}{
		"empty maps": {
			base:      map[string]string{},
			overrides: map[string]string{},
			want:      map[string]string{},
		},
		"nil maps": {
			base:      nil,
			overrides: nil,
			want:      map[string]string{},
		},
		"empty base map": {
			base: map[string]string{},
			overrides: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"nil base map": {
			base: nil,
			overrides: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"empty overrides map": {
			base: map[string]string{
				"test": "123",
				"foo":  "bar"},
			overrides: map[string]string{},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"nil overrides map": {
			base: map[string]string{
				"test": "123",
				"foo":  "bar"},
			overrides: nil,
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"non-empty maps": {
			base: map[string]string{
				"test":      "123",
				"foo":       "bar",
				"bish":      "bash",
				"something": "else"},
			overrides: map[string]string{
				"test": "456",
				"foo":  "baz",
				"bish": ""},
			want: map[string]string{
				"test":      "456",
				"foo":       "baz",
				"bish":      "",
				"something": "else"},
		},
		"ref secret in base map": {
			base: map[string]string{
				"test": "123"},
			overrides: map[string]string{
				"test": "<secret>"},
			want: map[string]string{
				"test": "123"},
			fields: []string{"test"},
		},
		"ref secret in base map, secret not found/empty": {
			base: map[string]string{
				"foo": ""},
			overrides: map[string]string{
				"foo":  "<secret>",
				"test": "<secret>"},
			want: map[string]string{
				"foo":  "<secret>",
				"test": "<secret>"},
			fields: []string{"foo", "test"},
		},
		"secret not in fields": {
			base: map[string]string{
				"test": "123",
				"foo":  "bar"},
			overrides: map[string]string{
				"test": "<secret>",
				"foo":  "<secret>"},
			want: map[string]string{
				"foo":  "<secret>",
				"test": "<secret>"},
			fields: []string{"other"},
		},
		"non-empty maps with secrets": {
			base: map[string]string{
				"test":      "123",
				"foo":       "bar",
				"bish":      "bash",
				"something": "else",
				"nothing":   ""},
			overrides: map[string]string{
				"test":    "456",
				"foo":     "<secret>",
				"bish":    "<secret>",
				"nothing": "<secret>"},
			want: map[string]string{
				"test":      "456",
				"foo":       "<secret>",
				"bish":      "bash",
				"something": "else",
				"nothing":   "<secret>"},
			fields: []string{"test", "bish"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN MergeMaps is called
			got := MergeMaps(tc.base, tc.overrides, tc.fields)

			// THEN the maps are merged correctly
			if len(got) != len(tc.want) {
				t.Fatalf("want: %v\ngot:  %v",
					tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, got)
				}
			}
		})
	}
}

func TestLowercaseStringStringMap(t *testing.T) {
	// GIVEN different byte strings
	tests := map[string]struct {
		input map[string]string
		want  map[string]string
	}{
		"empty map": {
			input: map[string]string{},
			want:  map[string]string{},
		},
		"lower-cased map": {
			input: map[string]string{
				"test": "123",
				"foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"lower-cased map with mixed-cased values": {
			input: map[string]string{
				"test": "123",
				"foo":  "bAr"},
			want: map[string]string{
				"test": "123",
				"foo":  "bAr"},
		},
		"upper-cased map": {
			input: map[string]string{
				"TEST": "123",
				"FOO":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
		"upper-cased map with mixed-case values": {
			input: map[string]string{
				"TEST": "123",
				"FOO":  "bAr"},
			want: map[string]string{
				"test": "123",
				"foo":  "bAr"},
		},
		"mixed-case map": {
			input: map[string]string{
				"tESt": "123",
				"Foo":  "bar"},
			want: map[string]string{
				"test": "123",
				"foo":  "bar"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN LowercaseStringStringMap is called
			got := LowercaseStringStringMap(&tc.input)

			// THEN the map keys are lower-cased correctly
			if len(got) != len(tc.want) {
				t.Fatalf("want: %v\ngot:  %v",
					tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, got)
				}
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	// GIVEN a string map
	strMap := map[string]int{
		"a": 0,
		"b": 0,
		"c": 0,
		"d": 0,
		"e": 0,
		"f": 0,
		"g": 0,
		"h": 0,
		"i": 0,
		"j": 0,
		"k": 0,
		"l": 0,
		"m": 0,
		"n": 0,
		"o": 0,
		"p": 0,
		"q": 0,
		"r": 0,
		"s": 0,
		"t": 0,
		"u": 0,
		"v": 0,
		"w": 0,
		"x": 0,
		"z": 0}

	// WHEN SortedKeys is called on it
	sorted := SortedKeys(strMap)

	// THEN the keys of the map are returned alphabetically sorted
	want := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
		"f",
		"g",
		"h",
		"i",
		"j",
		"k",
		"l",
		"m",
		"n",
		"o",
		"p",
		"q",
		"r",
		"s",
		"t",
		"u",
		"v",
		"w",
		"x",
		"z"}
	for i := 1; i < 1000; i++ { // repeat due to random ordering
		for i := range sorted {
			if sorted[i] != want[i] {
				t.Errorf("want index=%d to be %q, not %q\nwant: %v\ngot:  %v",
					i, want[i], sorted[i], want, sorted)
			}
		}
	}
}

func TestStringToBoolPtr(t *testing.T) {
	// GIVEN a string
	tests := map[string]struct {
		input string
		want  *bool
	}{
		"'true' gives true": {
			input: "true", want: test.BoolPtr(true)},
		"'false' gives false": {
			input: "false", want: test.BoolPtr(false)},
		"'' gives nil": {
			input: "", want: nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN StringToBoolPtr is called
			got := StringToBoolPtr(tc.input)

			// THEN the string is converted to a bool pointer
			if got == tc.want {
				return
			}
			// One of them is nil, but the other is not
			if (got == nil && tc.want != nil) || (tc.want == nil && got != nil) {
				t.Errorf("want: %v\ngot:  %v",
					tc.want, got)
			}
			// Not the same bool value
			if *got != *tc.want {
				t.Errorf("want: %v\ngot:  %v",
					tc.want, got)
			}
		})
	}
}

func TestEnvReplaceFunc(t *testing.T) {
	// GIVEN an env var that may or may not exist
	envVarNameBase := "TESTENVREPLACEFUNC"
	strBase := "${" + envVarNameBase + "}"
	tests := map[string]struct {
		envVar *string
		want   string
	}{
		"undefined env var": {
			want: strBase,
		},
		"empty env var": {
			envVar: test.StringPtr(""),
			want:   "",
		},
		"non-empty env var": {
			envVar: test.StringPtr("bar"),
			want:   "bar",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			str := strBase
			if tc.envVar != nil {
				envVarName := envVarNameBase + strings.ReplaceAll(name, " ", "_")
				str = "${" + envVarName + "}"
				os.Setenv(envVarName, *tc.envVar)
				defer os.Unsetenv(envVarName)
			}

			// WHEN EnvReplaceFunc is called
			got := envReplaceFunc(str)

			// THEN the string is evaluated correctly
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestEvalEnvVars(t *testing.T) {
	// GIVEN a string
	tests := map[string]struct {
		input string
		env   map[string]string
		want  string
	}{
		"no env vars": {
			input: "hello there ${not an env var}",
			want:  "hello there ${not an env var}",
		},
		"1 env var": {
			env:   map[string]string{"TESTEVALENVVARS_ONE": "bar"},
			input: "hello there ${TESTEVALENVVARS_ONE}",
			want:  "hello there bar",
		},
		"2 env vars": {
			env: map[string]string{
				"TESTEVALENVVARS_TWO":   "bar",
				"TESTEVALENVVARS_THREE": "baz"},
			input: "hello there ${TESTEVALENVVARS_TWO} ${TESTEVALENVVARS_THREE}",
			want:  "hello there bar baz",
		},
		"unset env var": {
			input: "hello there ${TESTEVALENVVARS_UNSET}",
			want:  "hello there ${TESTEVALENVVARS_UNSET}",
		},
		"empty env var": {
			env:   map[string]string{"TESTEVALENVVARS_FOUR": ""},
			input: "hello there ${TESTEVALENVVARS_FOUR}",
			want:  "hello there ",
		},
		"nested env vars not evaluated": {
			env: map[string]string{
				"TESTEVALENVVARS_FIVE":  "bar",
				"TESTEVALENVVARS_SIX":   "${TESTEVALENVVARS_SEVEN}",
				"TESTEVALENVVARS_SEVEN": "qux"},
			input: "hello there ${TESTEVALENVVARS_FIVE} ${TESTEVALENVVARS_SIX}",
			want:  "hello there bar ${TESTEVALENVVARS_SEVEN}",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// WHEN EvalEnvVars is called
			got := EvalEnvVars(tc.input)

			// THEN the string is evaluated correctly
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestGetKeysFromJSON(t *testing.T) {
	// GIVEN a JSON string
	tests := map[string]struct {
		input string
		want  []string
	}{
		"empty string": {
			input: "",
			want:  []string{},
		},
		"empty object": {
			input: "{}",
			want:  []string{},
		},
		"empty array": {
			input: "[]",
			want:  []string{},
		},
		"1 key": {
			input: `{"test": 123}`,
			want:  []string{"test"},
		},
		"2 keys": {
			input: `{"test": 123, "foo": "bar"}`,
			want: []string{
				"foo",
				"test"},
		},
		"nested keys": {
			input: `{"test": 123, "foo": {"bar": "baz"}}`,
			want: []string{
				"foo",
				"foo.bar",
				"test"},
		},
		"array keys": {
			input: `{"test": 123, "foo": ["bar", "baz"]}`,
			want: []string{
				"foo",
				"test"},
		},
		"nested array keys": {
			input: `{"test": 123, "foo": ["bar", {"baz": "bish"}]}`,
			want: []string{
				"foo",
				"test"},
		},
		"array of objects": {
			input: `{"test": 123, "foo": [{"bar": "baz"}, {"bish": "bash"}]}`,
			want: []string{
				"foo",
				"test"},
		},
		"array of arrays": {
			input: `{"test": 123, "foo": [["bar", "baz"], ["bish", "bash"]]}`,
			want: []string{
				"foo",
				"test"},
		},
		"array of arrays of objects": {
			input: `{"test": 123, "foo": [[{"bar": "baz"}, {"bish": "bash"}], [{"bash": "quuz"}, {"corge": "grault"}]]}`,
			want: []string{
				"foo",
				"test"},
		},
		"nested objects": {
			input: `{"test": 123, "foo": {"bar": {"baz": "qux"}}}`,
			want: []string{
				"foo",
				"foo.bar",
				"foo.bar.baz",
				"test"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN getKeysFromJSONBytes is called
			got := getKeysFromJSONBytes([]byte(tc.input), "")

			// THEN the keys are returned correctly
			if len(got) != len(tc.want) {
				t.Fatalf("want: %v\ngot:  %v",
					tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("want: %v\ngot:  %v",
						tc.want, got)
				}
			}

			// AND GetKeysFromJSON gets the same result
			gotOther := GetKeysFromJSON(tc.input)
			if len(got) != len(gotOther) {
				t.Fatalf("want: %v\ngot:  %v",
					got, gotOther)
			}
			for i := range got {
				if got[i] != gotOther[i] {
					t.Fatalf("want: %v\ngot:  %v",
						got, gotOther)
				}
			}
		})
	}
}

func TestBasicAuth(t *testing.T) {
	// GIVEN a username and password
	username := "test"
	password := "123"

	// WHEN BasicAuth is called with this
	got := BasicAuth(username, password)

	// THEN username:password is returned in base64
	want := "dGVzdDoxMjM="
	if want != got {
		t.Errorf("Failed encoding\nwant: %q\ngot:  %q",
			want, got)
	}
}

func TestIsHashed(t *testing.T) {
	// GIVEN a string
	tests := map[string]struct {
		input string
		want  bool
	}{
		"empty string": {
			input: "",
			want:  false,
		},
		"non-hashed string": {
			input: "h__foo",
			want:  false,
		},
		"hashed string": {
			input: fmt.Sprintf("h__%x", sha256.Sum256([]byte("foo"))),
			want:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN isHashed is called on it
			got := isHashed(tc.input)

			// THEN the hash is detected correctly
			if got != tc.want {
				t.Errorf("want: %v\ngot:  %v",
					tc.want, got)
			}
		})
	}
}

func TestHash(t *testing.T) {
	// GIVEN a string
	tests := map[string]struct {
		input string
		want  [32]byte
	}{
		"empty string": {
			input: "",
			want:  sha256.Sum256([]byte("")),
		},
		"non-empty string": {
			input: "foo",
			want:  sha256.Sum256([]byte("foo")),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN hash is called on it
			got := hash(tc.input)

			// THEN the string is hashed correctly
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
		})
	}
}

func TestHashFromString(t *testing.T) {
	// GIVEN a string that contains a hash
	tests := map[string]string{
		"empty string":     "",
		"non-empty string": "foobar",
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := sha256.Sum256([]byte(tc))
			input := fmt.Sprintf("h__%x", want)

			// WHEN hashFromString is called on it
			got := hashFromString(input[3:])

			// THEN the string is hashed correctly
			var got32 [32]byte
			copy(got32[:], got[:])
			if got32 != want {
				t.Errorf("want: %q\ngot:  %q",
					want, got32)
			}
		})
	}
}

func TestGetHash(t *testing.T) {
	// GIVEN a string that may or may not be hashed
	tests := map[string]struct {
		input         string
		alreadyHashed bool
	}{
		"empty string": {
			input: "",
		},
		"non-empty string": {
			input: "foo",
		},
		"hashed string": {
			input:         fmt.Sprintf("h__%x", sha256.Sum256([]byte("foo"))),
			alreadyHashed: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := tc.input
			if !tc.alreadyHashed {
				want = FmtHash(sha256.Sum256([]byte(tc.input)))
			}

			// WHEN GetHash is called on it
			got := GetHash(tc.input)

			// THEN the string is hashed correctly
			gotHash := FmtHash(got)
			if gotHash != want {
				t.Errorf("want: %q\ngot:  %q",
					want, gotHash)
			}
		})
	}
}

func TestFmtHash(t *testing.T) {
	// GIVEN a hash
	hash := sha256.Sum256([]byte("foo"))

	// WHEN FmtHash is called on it
	got := FmtHash(hash)

	// THEN the hash is formatted correctly
	want := fmt.Sprintf("h__%x", hash)
	if got != want {
		t.Errorf("want: %q\ngot:  %q",
			want, got)
	}
}

func TestStringToPointer(t *testing.T) {
	// GIVEN a string
	tests := map[string]struct {
		input string
		want  *string
	}{
		"empty string": {
			input: "",
			want:  nil},
		"non-empty string": {
			input: "test",
			want:  test.StringPtr("test")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN StringToPointer is called on it
			got := StringToPointer(tc.input)

			// THEN the string is converted to a pointer
			// AND the empty string is converted to nil
			if got == tc.want {
				return
			}
			// AND other values become pointers to the string
			if *got != *tc.want {
				t.Errorf("want: %q\ngot:  %q",
					*tc.want, *got)
			}
		})
	}
}

func TestParseKeys(t *testing.T) {
	// GIVEN a JSON key string
	tests := map[string]struct {
		input    string
		want     []interface{}
		errRegex string
	}{
		"empty string": {
			input: "",
			want:  []interface{}{},
		},
		"single key": {
			input: "foo",
			want:  []interface{}{"foo"},
		},
		"multiple keys": {
			input: "foo.bar",
			want:  []interface{}{"foo", "bar"},
		},
		"multiple keys with array": {
			input: "foo.bar[1]",
			want:  []interface{}{"foo", "bar", 1},
		},
		"multiple keys with array of objects": {
			input: "foo.bar[1].baz",
			want:  []interface{}{"foo", "bar", 1, "baz"},
		},
		"multiple keys with array of arrays": {
			input: "foo.bar[1][2]",
			want:  []interface{}{"foo", "bar", 1, 2},
		},
		"multiple keys with array of arrays of objects": {
			input: "foo.bar[1][2].baz",
			want:  []interface{}{"foo", "bar", 1, 2, "baz"},
		},
		"multiple keys with array of arrays of objects with array": {
			input: "foo.bar[1][2].baz[3]",
			want:  []interface{}{"foo", "bar", 1, 2, "baz", 3},
		},
		"non-int index": {
			input:    "foo.bar[1.1][2].baz[3]",
			want:     []interface{}{"foo", "bar"},
			errRegex: `failed to parse index "1.1" in `,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN ParseKeys is called
			got, err := ParseKeys(tc.input)

			// THEN the keys are returned correctly
			if len(got) != len(tc.want) {
				t.Fatalf("different amount of keys returned\nwant: %v\ngot:  %v",
					tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("wrong key - %d\nwant: %v\ngot:  %v",
						i, tc.want, got)
				}
			}
			// AND the error is returned correctly
			if tc.errRegex == "" {
				tc.errRegex = `^$`
			}
			e := ErrorToString(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Errorf("want error matching %q, got %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestNavigateJSON(t *testing.T) {
	// GIVEN a JSON string
	tests := map[string]struct {
		input    string
		key      string
		want     string
		errRegex string
	}{
		"empty key": {
			input:    `{ "foo": "bar" }`,
			key:      "",
			errRegex: "no key was given",
		},
		"object not found": {
			input:    "{}",
			key:      "foo",
			errRegex: `failed to find value for "[^"]+" in `,
		},
		"simple JSON": {
			input: `{"foo": "bar"}`,
			key:   "foo",
			want:  "bar",
		},
		"multi-level JSON": {
			input: `{"foo": {"bar": "baz"}}`,
			key:   "foo.bar",
			want:  "baz",
		},
		"multi-level JSON with array": {
			input: `{"foo": {"bar": ["baz", "bish"]}}`,
			key:   "foo.bar[1]",
			want:  "bish",
		},
		"multi-level JSON with array of objects": {
			input: `{"foo": {"bar": [{"baz": "bish"}, {"bash": "quuz"}]}}`,
			key:   "foo.bar[1].bash",
			want:  "quuz",
		},
		"multi-level JSON with array of arrays": {
			input: `{"foo": {"bar": [["baz", "bish"], ["bash", "quuz"]]}}`,
			key:   "foo.bar[1][1]",
			want:  "quuz",
		},
		"negative index": {
			input: `{"foo": {"bar": [["baz", "bish"], ["bash", "quuz"]]}}`,
			key:   "foo.bar[-1][1]",
			want:  "quuz",
		},
		"fail: index of map": {
			input:    `{"foo": {"bar": {"baz": "bish"}}}`,
			key:      "foo.bar[1]",
			errRegex: "got a map, but the key is not a string",
		},
		"fail: non-int index": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "quuz"]]}}`,
			key:      "foo.bar.bar",
			errRegex: "got an array, but the key is not an integer index",
		},
		"fail: index out of range": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "quuz"]]}}`,
			key:      "foo.bar[1][2]",
			errRegex: `index \d \([^)]+\) out of range`,
		},
		"fail: index out of range (negative)": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "quuz"]]}}`,
			key:      "foo.bar[-4][3]",
			errRegex: `index -\d \([^)]+\) out of range`,
		},
		"fail: got value instead of object": {
			input:    `{"foo": {"bar": "baz"}}`,
			key:      "foo.bar.baz",
			errRegex: `got a value of "[^"]+" at "[^"]+", but there are more keys to navigate`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var jsonData interface{}
			err := json.Unmarshal([]byte(tc.input), &jsonData)

			// WHEN navigateJSON is called
			got, err := navigateJSON(&jsonData, tc.key)

			// THEN the value is returned correctly
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
			// AND the error is returned correctly
			if tc.errRegex == "" {
				tc.errRegex = `^$`
			}
			e := ErrorToString(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Errorf("want error matching %q, got %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestGetValueByKey(t *testing.T) {
	// GIVEN a JSON string
	tests := map[string]struct {
		input    string
		key      string
		want     string
		errRegex string
	}{
		"fail unmarshal": {
			input:    "{",
			key:      "foo",
			errRegex: "failed to unmarshal the following from",
		},
		"empty key": {
			input: `{ "foo": "bar" }`,
			key:   "",
			want:  "__root",
		},
		"object not found": {
			input:    "{}",
			key:      "foo",
			errRegex: `failed to find value for "[^"]+" in `,
		},
		"simple JSON": {
			input: `{"foo": "bar"}`,
			key:   "foo",
			want:  "bar",
		},
		"multi-level JSON": {
			input: `{"foo": {"bar": "baz"}}`,
			key:   "foo.bar",
			want:  "baz",
		},
		"negative index": {
			input: `{"foo": {"bar": [["baz", "bish"], ["bash", "quuz"]]}}`,
			key:   "foo.bar[-1][1]",
			want:  "quuz",
		},
		"fail: index out of range": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "quuz"]]}}`,
			key:      "foo.bar[1][2]",
			errRegex: `index \d \([^)]+\) out of range`,
		},
		"fail: index out of range (negative)": {
			input:    `{"foo": {"bar": [["baz", "bish"], ["bash", "quuz"]]}}`,
			key:      "foo.bar[-4][3]",
			errRegex: `index -\d \([^)]+\) out of range`,
		},
		"fail: got value instead of object": {
			input:    `{"foo": {"bar": "baz"}}`,
			key:      "foo.bar.baz",
			errRegex: `got a value of "[^"]+" at "[^"]+", but there are more keys to navigate`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN GetValueByKey is called
			got, err := GetValueByKey([]byte(tc.input), tc.key, "https://release-argus.com")

			// THEN the value is returned correctly
			tc.want = strings.ReplaceAll(tc.want, "__root", tc.input)
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %q",
					tc.want, got)
			}
			// AND the error is returned correctly
			if tc.errRegex == "" {
				tc.errRegex = `^$`
			}
			e := ErrorToString(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Errorf("want error matching %q, got %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestTo____String(t *testing.T) {
	// GIVEN a struct to print in YAML format
	tests := map[string]struct {
		input    interface{}
		wantYAML string
		wantJSON string
	}{
		"empty struct": {
			input:    struct{}{},
			wantYAML: "{}",
			wantJSON: "{}",
		},
		"simple struct": {
			input: struct {
				Test string `yaml:"test" json:"test"`
			}{
				Test: "test"},
			wantYAML: "test: test",
			wantJSON: `{"test":"test"}`,
		},
		"nested struct": {
			input: struct {
				Test struct {
					Foo string `yaml:"foo" json:"foo"`
				} `yaml:"test" json:"test"`
			}{
				Test: struct {
					Foo string `yaml:"foo" json:"foo"`
				}{
					Foo: "bar"}},
			wantYAML: "test:\n  foo: bar",
			wantJSON: `{"test":{"foo":"bar"}}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			prefixes := []string{"", " ", "  ", "    ", "- "}
			for _, prefix := range prefixes {
				wantYAML := strings.TrimPrefix(tc.wantYAML, "\n")
				if wantYAML != "" {
					if wantYAML != "{}" {
						wantYAML = prefix + strings.ReplaceAll(wantYAML, "\n", "\n"+prefix)
					}
					wantYAML += "\n"
				}

				// WHEN ToYAMLString is called
				gotYAML := ToYAMLString(tc.input, prefix)

				// THEN the struct is printed in YAML format
				if gotYAML != wantYAML {
					t.Fatalf("(prefix=%q) want:\n%q\ngot:\n%q",
						prefix, wantYAML, gotYAML)
				}
			}

			// WHEN ToJSONString is called
			gotJSON := ToJSONString(tc.input)

			// THEN the struct is printed in JSON format
			if gotJSON != tc.wantJSON {
				t.Fatalf("want:\n%q\ngot:\n%q",
					tc.wantJSON, gotJSON)
			}
		})
	}
}
