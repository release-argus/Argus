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

	"github.com/release-argus/Argus/util/errfmt"
)

func TestAssertMapEqual(t *testing.T) {
	// GIVEN: two maps to compare.
	tests := []struct {
		name     string
		a, b     map[string]interface{}
		want     bool
		errRegex string
	}{
		{
			name:     "equal",
			a:        map[string]interface{}{"foo": "bar"},
			b:        map[string]interface{}{"foo": "bar"},
			want:     true,
			errRegex: `^$`,
		},
		{
			name:     "length mismatch",
			a:        map[string]interface{}{"foo": "bar"},
			b:        map[string]interface{}{"foo": "bar", "baz": "qux"},
			errRegex: `length mismatch`,
		},
		{
			name:     "key missing",
			a:        map[string]interface{}{"foo": "bar"},
			b:        map[string]interface{}{"baz": "qux"},
			errRegex: `\["baz"\] missing`,
		},
		{
			name:     "key mismatch",
			a:        map[string]interface{}{"foo": "bar"},
			b:        map[string]interface{}{"foo": "baz"},
			errRegex: `\["foo"\] mismatch`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: AssertMapEqual is called.
			err := AssertMapEqual(t, tc.a, tc.b, "TestAssertMapEqual", "target")

			// THEN: The error is as expected.
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Fatalf("%s\nAssertMapEqual(a=%v, b=%v) error mismatch\ngot:  %q\nwant: %q",
					packageName, tc.a, tc.b,
					e, tc.errRegex,
				)
			}
		})
	}
}
