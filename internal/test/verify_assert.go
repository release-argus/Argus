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
	"regexp"
	"strings"

	"github.com/release-argus/Argus/util/errfmt"
)

// AssertCheckValuesWithErrorAndChanged runs the given function and asserts both error and changed flag.
//
// GIVEN: a function that returns (error, bool)
//
// WHEN: the function is called.
//
// THEN: 'error' matches the supplied regex and 'changed' flag matches expectation.
func AssertCheckValuesWithErrorAndChanged(
	t tLogger,
	packageName string,
	wantErrRegex string,
	wantChanged bool,
	checkValues func() (error, bool),
) (error, bool) {
	t.Helper()

	// Apply prefix to regex.
	errRegex := wantErrRegex

	// WHEN: the function is called.
	err, changed := checkValues()

	prefix := fmt.Sprintf("%s\nCheckValues()", packageName)

	// THEN: The number of lines of error matches expected.
	e := errfmt.FormatError(err)
	lines := strings.Split(e, "\n")
	gotLines := len(lines)
	wantLines := strings.Count(errRegex, "\n") + 1
	if gotLines < wantLines {
		t.Fatalf(
			"%s error stdout line count mismatch\ngot:  %d\nwant: %d\nstdout:\n%q\nerrRegex:\n%q",
			prefix,
			gotLines, wantLines,
			e, errRegex,
		)
	}

	// AND: error matches regex.
	regex := regexp.MustCompile(errRegex)
	if !regex.MatchString(e) {
		t.Fatalf(
			"%s error mismatch\ngot:  %q\nwant: %q",
			prefix, e, errRegex,
		)
	}

	// AND: changed flag matches expectation.
	if changed != wantChanged {
		t.Errorf(
			"%s changed flag mismatch\ngot:  %t\nwant: %t",
			prefix, changed, wantChanged,
		)
	}

	return err, changed
}

// AssertCheckValuesWithError runs the given function and asserts errors match.
//
// GIVEN: a function that returns an error
//
// WHEN: this function is called.
//
// THEN: error matches prefixed regex.
func AssertCheckValuesWithError(
	t tLogger,
	packageName string,
	wantErrRegex string,
	checkValues func() error,
) error {
	t.Helper()

	err, _ := AssertCheckValuesWithErrorAndChanged(
		t,
		packageName,
		wantErrRegex,
		false,
		func() (error, bool) {
			return checkValues(), false
		},
	)
	return err
}
