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

//go:build unit || integration

package test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/release-argus/Argus/util/errfmt"
)

// AssertDecode runs the given func, with the given data and asserts that the error is as expected,
// and that the struct produced stringifies as expected.
func AssertDecode[T any](
	t *testing.T,
	f func(format string, data []byte) (T, error),
	format, data string,
	stringify func(T) string,
	wantStr, errRegex string,
	packageName, funcName string,
) (decoded T, decodeErr, testErr error) {
	t.Helper()

	// WHEN: The decode function is called with it.
	v, err := f(format, []byte(data))

	prefix := fmt.Sprintf(
		"%s\n%s(format=%q, data=%q)",
		packageName, funcName, format, data,
	)

	// THEN: The error is as expected.
	e := errfmt.FormatError(err)
	if !regexp.MustCompile(errRegex).MatchString(e) {
		testErr = fmt.Errorf(
			"%s error mismatch\ngot:  %q\nwant: %q",
			prefix, e, errRegex,
		)
	}
	if e != "" {
		return v, err, testErr
	}

	// AND: The struct stringifies correctly.
	var gotStr string
	if stringify == nil {
		return v, nil, fmt.Errorf("stringify function is required")
	} else if !isNil(v) {
		gotStr = stringify(v)
	}
	if gotStr != wantStr {
		testErr = fmt.Errorf(
			"%s stringified mismatch\ngot:  %q\nwant: %q",
			prefix, gotStr, wantStr,
		)
	}
	return v, nil, testErr
}

func AssertApplyOverrides[T any](
	t *testing.T,
	target T,
	f func(format string, data []byte, target T) (T, error),
	format, data string,
	stringify func(T) string,
	wantStr, errRegex string,
	sameAddress bool,
	packageName, funcName string,
) (overridden T, overridesErr, testErr error) {
	t.Helper()

	// WHEN: ApplyOverrides is called with it.
	v, err := f(format, []byte(data), target)

	prefix := fmt.Sprintf(
		"%s\n%s(format=%q, data=%q)",
		packageName, funcName, format, data,
	)

	// THEN: The error is as expected.
	e := errfmt.FormatError(err)
	if !regexp.MustCompile(errRegex).MatchString(e) {
		testErr = fmt.Errorf(
			"%s error mismatch\ngot:  %q\nwant: %q",
			prefix, e, errRegex,
		)
	}
	if e != "" {
		return v, err, testErr
	}

	// AND: The struct stringifies correctly.
	var gotStr string
	if stringify == nil {
		return v, nil, fmt.Errorf("stringify function is required")
	} else if !isNil(v) {
		gotStr = stringify(v)
	}
	if gotStr != wantStr {
		testErr = fmt.Errorf(
			"%s stringified mismatch\ngot:  %q\nwant: %q",
			prefix, gotStr, wantStr,
		)
	}

	// AND: the returned struct points to the same address as the previous if non-nil previously.
	if !isNil(v) && !isNil(target) {
		gotPtr := reflect.ValueOf(v).Pointer()
		targetPtr := reflect.ValueOf(target).Pointer()
		gotSameAddress := gotPtr == targetPtr
		if gotSameAddress != sameAddress {
			msg := "%s pointer mismatch - address should have changed\ngot:  %v\nwant: %v"
			if sameAddress {
				msg = "%s pointer mismatch - address changed unexpectedly\ngot:  %v\nwant: %v"
			}
			testErr = fmt.Errorf(
				msg,
				prefix, gotPtr, targetPtr,
			)
		}
	}
	return v, nil, testErr
}
