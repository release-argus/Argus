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
	"strings"
	"testing"

	"github.com/release-argus/Argus/util/errfmt"
)

func TestAssertMapEqualFunc(t *testing.T) {
	caseInsensitive := strings.EqualFold

	tests := []struct {
		name     string
		a        map[string]string
		b        map[string]string
		eq       func(string, string) bool
		target   string
		errRegex string
	}{
		{
			name:     "equal/non-empty",
			a:        map[string]string{"foo": "bar"},
			b:        map[string]string{"foo": "bar"},
			eq:       caseInsensitive,
			errRegex: `^$`,
		},
		{
			name:     "equal/both nil",
			a:        nil,
			b:        nil,
			eq:       caseInsensitive,
			errRegex: `^$`,
		},
		{
			name:     "equal/nil and empty",
			a:        nil,
			b:        map[string]string{},
			eq:       caseInsensitive,
			errRegex: `^$`,
		},
		{
			name:     "equal by eq, not by ==",
			a:        map[string]string{"foo": "BAR"},
			b:        map[string]string{"foo": "bar"},
			eq:       caseInsensitive,
			errRegex: `^$`,
		},
		{
			name:     "length mismatch/no target",
			a:        map[string]string{"foo": "bar"},
			b:        map[string]string{"foo": "bar", "baz": "qux"},
			eq:       caseInsensitive,
			errRegex: `^length mismatch\ngot:  1 \(map\["foo":"bar"\]\)\nwant: 2 \(map\["baz":"qux" "foo":"bar"\]\)$`,
		},
		{
			name:     "length mismatch/with target",
			a:        map[string]string{"foo": "bar"},
			b:        map[string]string{"foo": "bar", "baz": "qux"},
			eq:       caseInsensitive,
			target:   "abc",
			errRegex: `^abc length mismatch\n`,
		},
		{
			name:     "key absent",
			a:        map[string]string{"foo": "bar"},
			b:        map[string]string{"baz": "qux"},
			eq:       caseInsensitive,
			target:   "abc",
			errRegex: `^abc\["baz"\] mismatch\ngot:  absent \(map\["foo":"bar"\]\)\nwant: "qux"$`,
		},
		{
			name:     "value mismatch/1 key",
			a:        map[string]string{"foo": "bar"},
			b:        map[string]string{"foo": "baz"},
			eq:       caseInsensitive,
			target:   "abc",
			errRegex: `^abc\["foo"\] mismatch\ngot:  "bar"\nwant: "baz"$`,
		},
		{
			name:     "value mismatch/every key reported, sorted",
			a:        map[string]string{"b": "x", "a": "x"},
			b:        map[string]string{"b": "y", "a": "y"},
			eq:       caseInsensitive,
			errRegex: `^\["a"\] mismatch\ngot:  "x"\nwant: "y"\n\["b"\] mismatch\ngot:  "x"\nwant: "y"$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := AssertMapEqualFunc(t, tc.a, tc.b, tc.eq, tc.target)
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}

func TestAssertMapEqual(t *testing.T) {
	tests := []struct {
		name     string
		a, b     map[string]any
		target   string
		errRegex string
	}{
		{
			name:     "equal",
			a:        map[string]any{"foo": "bar"},
			b:        map[string]any{"foo": "bar"},
			errRegex: `^$`,
		},
		{
			name: "length mismatch",
			a:    map[string]any{"foo": "bar"},
			b:    map[string]any{"foo": "bar", "baz": "qux"},
			errRegex: TrimYAML(`
				^length mismatch
				got:  1 \(map\["foo":"bar"\]\)
				want: 2 \(map\["baz":"qux" "foo":"bar"\]\)$`),
		},
		{
			name: "length mismatch/nil map",
			a:    nil,
			b:    map[string]any{"foo": "bar"},
			errRegex: TrimYAML(`
				^length mismatch
				got:  0 \(nil\)
				want: 1 \(map\["foo":"bar"\]\)$`),
		},
		{
			name:   "error with target",
			a:      map[string]any{"foo": "bar"},
			b:      map[string]any{"foo": "bar", "baz": "qux"},
			target: "abc",
			errRegex: TrimYAML(`
				^abc length mismatch
				got:  1 \(map\["foo":"bar"\]\)
				want: 2 \(map\["baz":"qux" "foo":"bar"\]\)$`),
		},
		{
			name: "key absent",
			a:    map[string]any{"foo": "bar"},
			b:    map[string]any{"baz": "qux"},
			errRegex: TrimYAML(`
				^\["baz"\] mismatch
				got:  absent \(map\["foo":"bar"\]\)
				want: "qux"$`),
		},
		{
			name: "key mismatch",
			a:    map[string]any{"foo": "bar"},
			b:    map[string]any{"foo": "baz"},
			errRegex: TrimYAML(`
				^\["foo"\] mismatch
				got:  "bar"
				want: "baz"$`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := AssertMapEqual(t, tc.a, tc.b, tc.target)
			e := errfmt.FormatError(err)
			if !regexp.MustCompile(tc.errRegex).MatchString(e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
		})
	}
}
