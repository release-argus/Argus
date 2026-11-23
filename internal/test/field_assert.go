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
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"time"
)

type CompareMode int

const (
	CompareEqual CompareMode = iota
	CompareNotEqual
	CompareSamePointer
	CompareDifferentPointer
	CompareNotNil
)

type FieldAssertion struct {
	Name, Target string
	Got          any
	Want         any
	Mode         CompareMode
	Check        func() (got, want any, ok bool)
}

// AssertFields checks every row and joins the mismatches, each named target.Name.
func AssertFields(
	t TB,
	fieldTests []FieldAssertion,
	target string,
) error {
	t.Helper()

	var errs []error
	for _, tc := range fieldTests {
		var got, want any
		var ok bool

		rowTarget := target
		if tc.Target != "" {
			rowTarget = tc.Target
		}
		field := tc.Name
		if rowTarget != "" {
			field = rowTarget + "." + tc.Name
		}
		subject := field + " mismatch"

		if tc.Check != nil {
			got, want, ok = tc.Check()
		} else {
			got, want = tc.Got, tc.Want
			if tc.Mode != CompareNotNil && panicsOnEqual(tc.Mode, got, want) {
				errs = append(errs, fmt.Errorf(
					"%s: %T is not comparable with ==, use Check",
					field, got,
				))
				continue
			}
			switch tc.Mode {
			case CompareNotEqual:
				ok = !equal(got, want)
			case CompareSamePointer:
				ok = got == want || (isNil(want) && isNil(got))
			case CompareDifferentPointer:
				ok = got != want && (!isNil(want) || !isNil(got))
			case CompareNotNil:
				ok = !isNil(got)
			default:
				ok = equal(got, want)
			}
		}
		if ok {
			continue
		}

		var err error
		switch tc.Mode {
		case CompareNotEqual:
			err = fmt.Errorf(
				"%s\ngot:  %s\nwant: NOT %s",
				subject, formatValue(got), formatValue(want),
			)
		case CompareSamePointer:
			err = fmt.Errorf(
				"%s\ngot:  %s\nwant: %s",
				subject, pointerString(got), pointerString(want),
			)
		case CompareDifferentPointer:
			err = fmt.Errorf(
				"%s\ngot:  %s\nwant: NOT %s",
				subject, pointerString(got), pointerString(want),
			)
		case CompareNotNil:
			err = fmt.Errorf("%s\ngot:  %s\nwant: non-nil", subject, formatValue(got))
		default:
			err = fmt.Errorf("%s\ngot:  %s\nwant: %s", subject, formatValue(got), formatValue(want))
		}
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CopiedPointerMode is the mode asserting a copy of want:
// [CompareSamePointer] when want is nil (as its copy must stay nil),
// otherwise [CompareDifferentPointer].
func CopiedPointerMode(want any) CompareMode {
	if isNil(want) {
		return CompareSamePointer
	}
	return CompareDifferentPointer
}

// equal compares with ==, except [time.Time] by [time.Time.Equal] and []byte by [bytes.Equal].
func equal(got, want any) bool {
	switch g := got.(type) {
	case time.Time:
		if w, isTime := want.(time.Time); isTime {
			return g.Equal(w)
		}
	case []byte:
		if w, isBytes := want.([]byte); isBytes {
			return bytes.Equal(g, w)
		}
	}
	return got == want
}

// panicsOnEqual reports whether comparing a and b under mode would panic.
func panicsOnEqual(mode CompareMode, a, b any) bool {
	if a == nil || b == nil || reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}
	if _, isBytes := a.([]byte); isBytes && (mode == CompareEqual || mode == CompareNotEqual) {
		return false
	}
	return !reflect.ValueOf(a).Comparable() || !reflect.ValueOf(b).Comparable()
}

// pointerString renders v's address when it has one.
func pointerString(v any) string {
	if !isNil(v) {
		if p, ok := addressOf(v); ok {
			return fmt.Sprintf("%#x", p)
		}
	}
	return formatValue(v)
}

// formatValue renders v quoted when its values are strings, and a nil as "nil".
func formatValue(v any) string {
	if isNil(v) {
		return "nil"
	}
	if holdsStrings(reflect.ValueOf(v)) {
		return fmt.Sprintf("%q", v)
	}
	return fmt.Sprintf("%+v", v)
}

func holdsStrings(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return true
	case reflect.Interface:
		return holdsStrings(v.Elem())
	case reflect.Array, reflect.Slice:
		for i := range v.Len() {
			if !holdsStrings(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map:
		for iter := v.MapRange(); iter.Next(); {
			if !holdsStrings(iter.Key()) || !holdsStrings(iter.Value()) {
				return false
			}
		}
		return true
	}
	return false
}
