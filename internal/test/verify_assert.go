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
	"regexp"
	"strings"

	"github.com/release-argus/Argus/util/errfmt"
)

// AssertCheckValuesWithErrorAndChanged runs the given function and asserts both error and changed flag.
func AssertCheckValuesWithErrorAndChanged(
	t TB,
	wantErrRegex string,
	wantChanged bool,
	checkValues func() (error, bool),
) (error, bool) {
	t.Helper()

	err, changed := checkValues()

	e := errfmt.FormatError(err)
	lines := strings.Split(e, "\n")
	gotLines := len(lines)
	wantLines := strings.Count(wantErrRegex, "\n") + 1
	if gotLines < wantLines {
		t.Fatalf(
			"error line count mismatch\ngot:  %d\nwant: >= %d\nerror:\n%q\nerrRegex:\n%q",
			gotLines, wantLines, e, wantErrRegex,
		)
	}

	if !regexp.MustCompile(wantErrRegex).MatchString(e) {
		t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, wantErrRegex)
	}

	if changed != wantChanged {
		t.Errorf("changed mismatch\ngot:  %t\nwant: %t", changed, wantChanged)
	}

	return err, changed
}

// AssertCheckValuesWithError runs the given function and asserts its error matches wantErrRegex.
func AssertCheckValuesWithError(
	t TB,
	wantErrRegex string,
	checkValues func() error,
) error {
	t.Helper()

	err, _ := AssertCheckValuesWithErrorAndChanged(
		t,
		wantErrRegex,
		false,
		func() (error, bool) {
			return checkValues(), false
		},
	)
	return err
}
