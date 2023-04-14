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
	"fmt"
	"io"
	"os"
	"testing"
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
		name, tc := name, tc
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
			ptr: stringPtr("foo"), nilStr: "bar",
			want: "foo"},
	}

	for name, tc := range tests {
		name, tc := name, tc
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
			a: stringPtr("foo"), b: "bar",
			want: "foo"},
	}

	for name, tc := range tests {
		name, tc := name, tc
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
			check: stringPtr("foo"), value: "bar",
			want: stringPtr("bar")},
	}

	for name, tc := range tests {
		name, tc := name, tc
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
		name, tc := name, tc
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
			check: stringPtr("foo"),
			want:  "foo"},
	}

	for name, tc := range tests {
		name, tc := name, tc
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

func TestGetFirstNonNilPtr(t *testing.T) {
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
				stringPtr("bar")},
			wantIndex: 3,
		},
		"2 non-nil pointers": {
			pointers: []*string{
				stringPtr("foo"),
				nil,
				nil,
				stringPtr("bar")},
			wantIndex: 0,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN GetFirstNonNilPtr is run on a slice of pointers
			got := GetFirstNonNilPtr(tc.pointers...)

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
		name, tc := name, tc
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

func TestGetFirstNonDefault(t *testing.T) {
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN GetFirstNonDefault is run on a slice of slice
			got := GetFirstNonDefault(tc.slice...)

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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			// t.Parallel()

			msg := "var is not default from PrintlnIfNotDefault"
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN PrintlnIfNotDefault is called
			PrintlnIfNotDefault(tc.element, msg)

			// THEN the var is printed when it should be
			w.Close()
			out, _ := io.ReadAll(r)
			got := string(out)
			os.Stdout = stdout
			if !tc.didPrint {
				if got != "" {
					t.Fatalf("printed %q",
						got)
				}
				return
			}
			if got != msg+"\n" {
				t.Errorf("unexpected print %q",
					got)
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
			element: stringPtr("foo"), didPrint: true},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {

			msg := "var is not default from PrintlnIfNotNil"
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN PrintlnIfNotNil is called
			PrintlnIfNotNil(tc.element, msg)

			// THEN the var is printed when it should be
			w.Close()
			out, _ := io.ReadAll(r)
			got := string(out)
			os.Stdout = stdout
			if !tc.didPrint {
				if got != "" {
					t.Fatalf("printed %q",
						got)
				}
				return
			}
			if got != msg+"\n" {
				t.Errorf("unexpected print %q",
					got)
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
			element: stringPtr("foo"), didPrint: false},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {

			msg := "var is not default from PrintlnIfNil"
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN PrintlnIfNil is called
			PrintlnIfNil(tc.element, msg)

			// THEN the var is printed when it should be
			w.Close()
			out, _ := io.ReadAll(r)
			got := string(out)
			os.Stdout = stdout
			if !tc.didPrint {
				if got != "" {
					t.Fatalf("printed %q",
						got)
				}
				return
			}
			if got != msg+"\n" {
				t.Errorf("unexpected print %q",
					got)
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
			element: stringPtr("foo"), value: "bar", want: "bar"},
	}

	for name, tc := range tests {
		name, tc := name, tc
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

func TestGetValue(t *testing.T) {
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
			ptr:   stringPtr("foo"),
			value: "bar", want: "foo"},
		"nil bool pointer": {
			ptr:   (*bool)(nil),
			value: false, want: false},
		"non-nil bool pointer": {
			ptr:   boolPtr(true),
			value: false, want: true},
		"nil int pointer": {
			ptr:   (*int)(nil),
			value: 1, want: 1},
		"non-nil int pointer": {
			ptr:   intPtr(3),
			value: 2, want: 3},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN GetValue is called
			var got interface{}
			switch v := tc.ptr.(type) {
			case *string:
				got = GetValue(v, tc.value.(string))
			case *bool:
				got = GetValue(v, tc.value.(bool))
			case *int:
				got = GetValue(v, tc.value.(int))
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
		name, tc := name, tc
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
		name, tc := name, tc
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
		name, tc := name, tc
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
		name, tc := name, tc
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
		name, tc := name, tc
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
		name, tc := name, tc
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

func TestGetPortFromURL(t *testing.T) {
	// GIVEN different byte strings
	tests := map[string]struct {
		url         string
		defaultPort string
		want        string
	}{
		"http url": {
			url:         "http://example.com",
			defaultPort: "1",
			want:        "80"},
		"http url with port": {
			url:         "http://example.com:123",
			defaultPort: "1",
			want:        "123"},
		"https url": {
			url:         "https://example.com",
			defaultPort: "1",
			want:        "443"},
		"https url with port": {
			url:         "https://example.com:123",
			defaultPort: "1",
			want:        "123"},
		"no protocol url with port": {
			url:         "example.com:123",
			defaultPort: "1",
			want:        "123"},
		"no protocol url with no port": {
			url:         "example.com",
			defaultPort: "1",
			want:        "1"},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN GetPortFromURL is called
			got := GetPortFromURL(tc.url, tc.defaultPort)

			// THEN the port is extracted/defaulted correctly
			if got != tc.want {
				t.Errorf("port not extracted from %q correctly\nwant: %q\ngot:  %q",
					tc.url, tc.want, got)
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
		name, tc := name, tc
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
			input: "true", want: boolPtr(true)},
		"'false' gives false": {
			input: "false", want: boolPtr(false)},
		"'' gives nil": {
			input: "", want: nil},
	}

	for name, tc := range tests {
		name, tc := name, tc
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
		name, tc := name, tc
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
			want:  stringPtr("test")},
	}

	for name, tc := range tests {
		name, tc := name, tc
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
