// Copyright [2022] [Argus]
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

package utils

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestContainsTrue(t *testing.T) {
	// GIVEN a list of strings
	want := "argus"
	lst := []string{"hello", want, "foo"}

	// WHEN Contains is run on this list with a element inside it
	found := Contains(lst, want)

	// THEN true is returned
	if !found {
		t.Errorf("%q couldn't be found in %v. But it is there!", want, lst)
	}
}

func TestContainsFalse(t *testing.T) {
	// GIVEN a list of strings
	want := "test"
	lst := []string{"hello", "argus", "foo"}

	// WHEN Contains is run on this list with a element inside it
	found := Contains(lst, want)

	// THEN true is returned
	if found {
		t.Errorf("%q shouldn't have been found in %v!", want, lst)
	}
}

func TestEvalNilPtrWithNilPtr(t *testing.T) {
	// Given a nil pointer and nilValue to be returned
	var pointer *string
	nilValue := "argus"

	// WHEN EvalNilPtr is called with these values
	got := EvalNilPtr(pointer, nilValue)

	// THEN nilValue is returned
	if got != nilValue {
		t.Errorf("EvalNilPtr was called with a nil pointer (%v), so should have returned %q",
			pointer, nilValue)
	}
}

func TestEvalNilPtrWithNonNilPtr(t *testing.T) {
	// Given a nil pointer and nilValue to be returned
	str := "foo"
	nilValue := "argus"

	// WHEN EvalNilPtr is called with these values
	got := EvalNilPtr(&str, nilValue)

	// THEN nilValue is returned
	if got != str {
		t.Errorf("EvalNilPtr was called with a nil pointer (%v), so should have returned %q",
			&str, str)
	}
}

func TestPtrOrValueToPtrWithNilPtr(t *testing.T) {
	// GIVEN a nil pointer and a string
	var (
		pointer *string
		value   string = "argus"
	)

	// WHEN PtrOrValueToPtr is called with these values
	got := PtrOrValueToPtr(pointer, value)
	want := value

	// THEN a pointer to value is returned
	if *got != want {
		t.Errorf("PtrOrValueToPtr with %v and %q should have returned &%q but returned &%q",
			pointer, value, value, *got)
	}
}

func TestPtrOrValueToPtrWithNonNilPtr(t *testing.T) {
	// GIVEN a non-nil pointer and a string
	var (
		pointer string = "something"
		value   string = "argus"
	)

	// WHEN PtrOrValueToPtr is called with these values
	got := PtrOrValueToPtr(&pointer, value)

	// THEN the pointer is returned
	if *got != pointer {
		t.Errorf("PtrOrValueToPtr with %v and %q should have returned &%q but returned &%q",
			pointer, value, pointer, *got)
	}
}

func TestValueIfNotNilWithNilPtr(t *testing.T) {
	// Given a nil pointer and a value
	var pointer *string
	value := "argus"

	// WHEN ValueIfNotNil is called with this pointer
	got := ValueIfNotNil(pointer, value)

	// THEN nil is returned
	if got != nil {
		t.Errorf("ValueIfNotNil was called with a nil pointer (%v), so should have returned nil, not %q",
			pointer, *got)
	}
}

func TestValueIfNotNilWithNonNilPtr(t *testing.T) {
	// Given a nil pointer and a value
	str := "foo"
	value := "argus"

	// WHEN ValueIfNotNil is called with this pointer
	got := ValueIfNotNil(&str, value)

	// THEN nil is returned
	if *got != value {
		t.Errorf("ValueIfNotNil was called with a non-nil pointer &(%q), so should have returned a pointer to value, not %q",
			str, *got)
	}
}

func TestValueIfNotDefaultWithDefault(t *testing.T) {
	// Given a default comparable and a value
	str := ""
	value := "argus"

	// WHEN ValueIfNotDefault is called with these vars
	got := ValueIfNotDefault(str, value)

	// THEN value is returned
	if got != str {
		t.Errorf("ValueIfNotDefault was called with a default string so should have returned that default, not %q",
			got)
	}
}

func TestValueIfNotDefaultWithNonDefault(t *testing.T) {
	// GIVEN a default comparable and a value
	str := "test"
	value := "argus"

	// WHEN ValueIfNotDefault is called with these vars
	got := ValueIfNotDefault(str, value)

	// THEN value is returned
	if got != value {
		t.Errorf("ValueIfNotDefault was called with a default string so should have returned %q, not %q",
			str, got)
	}
}

func TestDefaultIfNilWithNil(t *testing.T) {
	// GIVEN a nil pointer to an int
	var pointer *int

	// WHEN DefaultIfNil is called with this nil pointer
	want := 0
	got := DefaultIfNil(pointer)

	// THEN the default of int (0) would be returned
	if got != want {
		t.Errorf("DefaultIfNil should have given %d but gave %d with a nil int pointer", want, got)
	}
}

func TestDefaultIfNilWithNonNil(t *testing.T) {
	// GIVEN a nil pointer to an int
	value := 1

	// WHEN DefaultIfNil is called with this nil pointer
	got := DefaultIfNil(&value)

	// THEN the default of int (0) would be returned
	if got != value {
		t.Errorf("DefaultIfNil should have given %d but gave %d with a nil int pointer", value, got)
	}
}

func TestGetFirstNonNilPtrWithAllNil(t *testing.T) {
	// GIVEN a bunch of nil pointers
	var (
		a *string
		b *string
		c *string
		d *string
	)

	// WHEN GetFirstNonNilPtr is called with these items
	var want *string
	got := GetFirstNonNilPtr(a, b, c, d)

	// THEN nil should be returned
	if got != want {
		t.Errorf("GetFirstNonNilPtr was given a list of nil's which should have returned nil, but returned %v", *got)
	}
}

func TestGetFirstNonNilPtrWithNonANil(t *testing.T) {
	// GIVEN a bunch of nil pointers
	var (
		a *string
		b string = "argus"
		c *string
		d *string
		e string = "foo"
	)

	// WHEN GetFirstNonNilPtr is called with these items
	want := "argus"
	got := GetFirstNonNilPtr(a, &b, c, d, &e)

	// THEN nil should be returned
	if *got != want {
		t.Errorf("GetFirstNonNilPtr was given a list of string pointers and should have returned the first non-nil, %s", want)
	}
}

func TestGetFirstNonDefaultWithAllDefault(t *testing.T) {
	// GIVEN a bunch of empty (default) strings
	var (
		a string
		b string
		c string
		d string
	)

	// WHEN GetFirstNonDefault is called with these items
	want := ""
	got := GetFirstNonDefault(a, b, c, d)

	// THEN the default string should be returned
	if got != want {
		t.Errorf("GetFirstNonDefault should have returned the empty string when given a list of empty strings. Got %s", got)
	}
}

func TestGetFirstNonDefaultWithNonDefault(t *testing.T) {
	// GIVEN a bunch of empty (default) strings
	var (
		a string
		b string = "argus"
		c string
		d string = "foo"
	)

	// WHEN GetFirstNonDefault is called with these items
	want := "argus"
	got := GetFirstNonDefault(a, b, c, d)

	// THEN the default string should be returned
	if got != want {
		t.Errorf("GetFirstNonDefault should have returned the first non-default (%s). Got %s", want, got)
	}
}

func TestPrintLnIfNotDefaultWithDefault(t *testing.T) {
	// GIVEN a default string
	str := ""

	// WHEN PrintlnIfNotDefault is called with this string
	PrintlnIfNotDefault(str, "SHOULDNT PRINT")

	// THEN it doesn't print
}

func TestPrintLnIfNotDefaultWithNonDefault(t *testing.T) {
	// GIVEN a non-default string
	str := "argus"

	// WHEN PrintlnIfNotDefault is called with this string
	PrintlnIfNotDefault(str, "SHOULD PRINT")

	// THEN it prints
}

func TestPrintLnIfNotNilWithNil(t *testing.T) {
	// GIVEN a nil pointer
	var pointer *string

	// WHEN PrintlnIfNotNil is called with this string
	PrintlnIfNotNil(pointer, "SHOULDNT PRINT")

	// THEN it doesn't print
}

func TestPrintLnIfNotNilWithNonNil(t *testing.T) {
	// GIVEN a non-default string
	str := "argus"

	// WHEN PrintlnIfNotNil is called with this string
	PrintlnIfNotNil(&str, "SHOULD PRINT")

	// THEN it prints
}

func TestPrintLnIfNilWithNil(t *testing.T) {
	// GIVEN a nil pointer
	var pointer *string

	// WHEN PrintlnIfNil is called with this string
	PrintlnIfNil(pointer, "SHOULDNT PRINT")

	// THEN it doesn't print
}

func TestPrintLnIfNilWithNonNil(t *testing.T) {
	// GIVEN a non-default string
	str := "argus"

	// WHEN PrintlnIfNil is called with this string
	PrintlnIfNil(&str, "SHOULD PRINT")

	// THEN it prints
}

func TestDefaultOrValueWithNil(t *testing.T) {
	// GIVEN a nil pointer and a string
	var pointer *string
	value := "argus"

	// WHEN DefaultOrValue is called with these vars
	got := DefaultOrValue(pointer, value)
	var want string

	// THEN the default string is returned
	if got != want {
		t.Errorf("DefaultOrValue should have returned %q when used with a nil, not %q", want, got)
	}
}

func TestDefaultOrValueWithNonNil(t *testing.T) {
	// GIVEN a nil pointer and a string
	pointer := "test"
	value := "argus"

	// WHEN DefaultOrValue is called with these vars
	got := DefaultOrValue(&pointer, value)

	// THEN value is returned
	if got != value {
		t.Errorf("DefaultOrValue should have returned %q when used with a nil, not %q", value, got)
	}
}

func TestErrorToStringWithNil(t *testing.T) {
	// GIVEN a nil error
	var err error

	// WHEN ErrorToString is called with it
	got := ErrorToString(err)
	want := ""

	// THEN an empty string is returned
	if got != want {
		t.Errorf("ErrorToString should have returned %q, but got %q", want, got)
	}
}

func TestErrorToStringWithErr(t *testing.T) {
	// GIVEN a nil error
	err := fmt.Errorf("something")

	// WHEN ErrorToString is called with it
	got := ErrorToString(err)
	want := err.Error()

	// THEN an empty string is returned
	if got != want {
		t.Errorf("ErrorToString should have returned %q, but got %q", want, got)
	}
}

func TestRandAlphaNumericLower(t *testing.T) {
	// GIVEN an alphanumeric string of length 10 is desired
	n := 10

	// WHEN RandAlphaNumericLower is called
	got := RandAlphaNumericLower(n)

	// THEN we got an alphanumeric string of length 10
	re := "^[0-9a-z]{10}$"
	regex := regexp.MustCompile(re)
	match := regex.MatchString(got)
	if !match {
		t.Errorf("%q RegEx didn't match an alphanumeric produced from RandAlphaNumericLower. Got %q", re, got)
	}
}

func TestRandNumeric(t *testing.T) {
	// GIVEN an alphanumeric string of length 10 is desired
	n := 10

	// WHEN TestRandNumeric is called with this length
	got := RandNumeric(n)

	// THEN we got an alphanumeric string of length 10
	re := "^[0-9]{10}$"
	regex := regexp.MustCompile(re)
	match := regex.MatchString(got)
	if !match {
		t.Errorf("%q RegEx didn't match a numeric produced from RandNumeric. Got %q", re, got)
	}
}

func TestRandString(t *testing.T) {
	// GIVEN a string of only a's and b's of length 20 is wanted
	alphabet := "ab"
	n := 20

	// WHEN RandString is called with these values
	got := RandString(n, alphabet)

	// THEN we got an alphanumeric string of length 10
	re := "^[ab]{20}$"
	regex := regexp.MustCompile(re)
	match := regex.MatchString(got)
	if !match {
		t.Errorf("%q RegEx didn't match a string produced from RandString. Got %q", re, got)
	}
}

func TestNormaliseNewlinesMac(t *testing.T) {
	// GIVEN a byte string with Mac newlines
	str := "hello\nargus\r"

	// WHEN normalised with NormaliseNewlines
	got := NormaliseNewlines([]byte(str))
	want := "hello\nargus\n"

	// THEN the Mac newlines are normalised to \n
	if string(got) != want {
		t.Errorf("Mac newlines were not normalised from %q to %q. Got %q", str, want, string(got))
	}
}

func TestNormaliseNewlinesWindows(t *testing.T) {
	// GIVEN a byte string with Windows newlines
	str := "hello\nargus\r\n"

	// WHEN normalised with NormaliseNewlines
	got := NormaliseNewlines([]byte(str))
	want := "hello\nargus\n"

	// THEN the Windows newlines are normalised to \n
	if string(got) != want {
		t.Errorf("Windows newlines were not normalised from %q to %q. Got %q", str, want, string(got))
	}
}

func TestCopyMap(t *testing.T) {
	// GIVEN a string map
	original := map[string]int{
		"a": 0,
		"b": 1,
		"c": 2,
	}

	// WHEN the map is copied
	copy := CopyMap(original)

	// THEN the map is a copy of the original one
	if len(copy) != len(original) {
		t.Errorf("CopyMap did not return an identical copy, length differed. Got %v from %v", copy, original)
	}
	for key := range copy {
		if original[key] != copy[key] {
			t.Errorf("CopyMap did not return an identical copy, map differed. Got %v from %v", copy, original)
		}
	}
	if &original == &copy {
		t.Error("CopyMap didn't copy the map, it's got the same address")
	}
}

func TestGetPortFromURLWithNoPortOrProto(t *testing.T) {
	// GIVEN a url with no protocl or port
	url := "example.com"

	// GIVEN GetPortFromURL is called on this string
	defaultPort := "1"
	got := GetPortFromURL(url, defaultPort)

	// THEN defaultPort is retunred
	if got != defaultPort {
		t.Errorf("GetPortFromURL shouldn't have found a port in %q. Got %q, want %q (defaultPort)", url, got, defaultPort)
	}
}

func TestGetPortFromURLWithPort(t *testing.T) {
	// GIVEN a url with a port specified
	port := "30"
	url := fmt.Sprintf("example.com:%s/hello", port)

	// GIVEN GetPortFromURL is called on this string
	got := GetPortFromURL(url, "1")

	// THEN defaultPort is retunred
	if got != port {
		t.Errorf("GetPortFromURL should have got the port from %q. Got %q, want %q", url, got, port)
	}
}

func TestGetPortFromURLWithProtoAndPort(t *testing.T) {
	// GIVEN a url with protocol and port specified
	port := "30"
	url := fmt.Sprintf("https://example.com:%s/hello", port)

	// GIVEN GetPortFromURL is called on this string
	defaultPort := "1"
	got := GetPortFromURL(url, defaultPort)

	// THEN the specified port is retunred
	if got != port {
		t.Errorf("GetPortFromURL should have got the port from %q. Got %q, want %q", url, got, port)
	}
}

func TestGetPortFromURLWithProtoHTTP(t *testing.T) {
	// GIVEN a url with protocol and port specified
	url := "http://example.com/hello"

	// GIVEN GetPortFromURL is called on this string
	defaultPort := "1"
	got := GetPortFromURL(url, defaultPort)
	want := "80"

	// THEN default protocol port is retunred
	if got != want {
		t.Errorf("GetPortFromURL should have got the port from the 'http://' in %q. Got %q, want %q", url, got, want)
	}
}

func TestGetPortFromURLWithProtoHTTPS(t *testing.T) {
	// GIVEN a url with protocol and port specified
	url := "https://example.com/hello"

	// WHEN GetPortFromURL is called on this string
	got := GetPortFromURL(url, "1")
	want := "443"

	// THEN defaultPort is retunred
	if got != want {
		t.Errorf("GetPortFromURL should have got the port from the 'https://' in %q. Got %q, want %q", url, got, want)
	}
}

func TestLowercaseStringStringMap(t *testing.T) {
	// GIVEN a [string]string map
	had := map[string]string{
		"fOO":   "1",
		"OTHER": "2",
	}

	// WHEN LowercaseStringStringMap is called on it
	got := LowercaseStringStringMap(&had)

	// THEN all the keys in the map are lowercased
	for key := range got {
		want := strings.ToLower(key)
		if key != want {
			t.Errorf("Keys were not lowercased, got %s, want %s", key, want)
		}
	}
}
