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
	"errors"
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
	call string,
) (decoded T, decodeErr, testErr error) {
	t.Helper()

	t.Logf("calling %s(format=%q, data=%q)", call, format, data)
	v, err := f(format, []byte(data))
	e := errfmt.FormatError(err)
	if !regexp.MustCompile(errRegex).MatchString(e) {
		testErr = fmt.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
	if err != nil {
		return v, err, testErr
	}

	var gotStr string
	if stringify == nil {
		return v, nil, errors.Join(testErr, errors.New("stringify function is required"))
	} else if !isNil(v) {
		gotStr = stringify(v)
	}
	if gotStr != wantStr {
		testErr = errors.Join(
			testErr,
			fmt.Errorf("%s() mismatch\ngot:  %q\nwant: %q", call, gotStr, wantStr),
		)
	}
	return v, nil, testErr
}

// AssertApplyOverrides is [AssertDecode] onto target, also asserting sameAddress of the result.
func AssertApplyOverrides[T any](
	t *testing.T,
	target T,
	f func(format string, data []byte, target T) (T, error),
	format, data string,
	stringify func(T) string,
	wantStr, errRegex string,
	sameAddress bool,
	call string,
) (overridden T, overridesErr, testErr error) {
	t.Helper()

	t.Logf("calling %s(format=%q, data=%q)", call, format, data)
	v, err := f(format, []byte(data), target)
	e := errfmt.FormatError(err)
	if !regexp.MustCompile(errRegex).MatchString(e) {
		testErr = fmt.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}
	if err != nil {
		return v, err, testErr
	}

	var gotStr string
	if stringify == nil {
		return v, nil, errors.Join(testErr, errors.New("stringify function is required"))
	} else if !isNil(v) {
		gotStr = stringify(v)
	}
	if gotStr != wantStr {
		testErr = errors.Join(
			testErr,
			fmt.Errorf("%s() mismatch\ngot:  %q\nwant: %q", call, gotStr, wantStr),
		)
	}

	if !isNil(v) && !isNil(target) {
		gotPtr, gotOK := addressOf(v)
		targetPtr, targetOK := addressOf(target)
		if !gotOK || !targetOK {
			return v, nil, errors.Join(
				testErr,
				fmt.Errorf("%s() address check needs a pointer-like type, got %T", call, v),
			)
		}
		gotSameAddress := gotPtr == targetPtr
		if gotSameAddress != sameAddress {
			if sameAddress {
				testErr = errors.Join(testErr, fmt.Errorf(
					"%s() address mismatch\ngot:  %#x\nwant: %#x",
					call, gotPtr, targetPtr,
				))
			} else {
				testErr = errors.Join(testErr, fmt.Errorf(
					"%s() address mismatch\ngot:  %#x\nwant: NOT %#x",
					call, gotPtr, targetPtr,
				))
			}
		}
	}
	return v, nil, testErr
}

// addressOf returns v's address, and false when v's kind has none.
func addressOf(v any) (uintptr, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map,
		reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.Pointer(), true
	}
	return 0, false
}
