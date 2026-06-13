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
	"testing"

	"github.com/release-argus/Argus/util/errfmt"
)

func AssertUnmarshal[T any](
	t *testing.T,
	format string,
	data string,
	v *T,
	errRegex string,
	stringify func(*T) string,
	wantStr string,
	packageName, structName string,
) (unmarshalErr, testErr error) {
	t.Helper()

	prefix := fmt.Sprintf(
		"%s\n%s.Unmarshal(format=%q, data=%q)",
		packageName, structName, format, data,
	)

	// WHEN: that data is unmarshaled into T.
	unmarshalErr = Unmarshal(format, []byte(data), v)

	// THEN: The error is as expected.
	e := errfmt.FormatError(unmarshalErr)
	if !regexp.MustCompile(errRegex).MatchString(e) {
		testErr = fmt.Errorf(
			"%s error mismatch\ngot:  %q\nwant: %q",
			prefix, e, errRegex,
		)
	}
	if e != "" || testErr != nil {
		return
	}

	// AND: it marshals back as expected.
	var gotStr string
	if stringify == nil {
		return nil, fmt.Errorf("stringify function is required")
	} else {
		gotStr = stringify(v)
	}
	if gotStr != wantStr {
		testErr = fmt.Errorf(
			"%s stringified mismatch\ngot:  %q\nwant: %q",
			prefix, gotStr, wantStr,
		)
	}
	return
}
