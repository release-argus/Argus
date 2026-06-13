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

package shoutrrr

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
)

func TestMapStringStringOmitNull_Unmarshal(t *testing.T) {
	// GIVEN: data in a given format to unmarshal into a MapStringStringOmitNull.
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "JSON/null value omitted",
			format:   "json",
			data:     `{"a": null}`,
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/null value omitted",
			format:   "yaml",
			data:     `a: null`,
			want:     "{}\n",
			errRegex: `^$`,
		},
		{
			name:     "JSON/single kv",
			format:   "json",
			data:     `{"foo":"bar"}`,
			want:     "foo: bar\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/single kv",
			format:   "yaml",
			data:     "foo: bar",
			want:     "foo: bar\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/multiple with null and empty string omitted",
			format: "json",
			data: test.TrimJSON(`{
				"a": "1",
				"b": null,
				"c": ""
			}`),
			want: test.TrimYAML(`
				a: '1'
				c: ''
			`),
			errRegex: `^$`,
		},
		{
			name:   "YAML/multiple with null and empty string omitted",
			format: "yaml",
			data: test.TrimYAML(`
				a: 1
				b: null
				c: ''
			`),
			want: test.TrimYAML(`
				a: '1'
				c: ''
			`),
			errRegex: `^$`,
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     "[]",
			errRegex: `json: .* unmarshal .*$`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     `foo`,
			errRegex: `^[^\s]+ string was used`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v MapStringStringOmitNull
			if _, testErr := test.AssertUnmarshal(
				t,
				tc.format, tc.data,
				&v,
				tc.errRegex,
				func(v *MapStringStringOmitNull) string { return decode.ToYAMLString(v, "") },
				tc.want,
				packageName,
				"MapStringStringOmitNull",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestMapStringStringOmitNull_Copy(t *testing.T) {
	type wantCase struct {
		name  string
		input MapStringStringOmitNull
		// Want the copy to be this after initial copy.
		wantCopyInitial map[string]string

		// Mutations to verify that Copy doesn't alias the underlying map.
		mutateOriginal map[string]string
		wantCopyAfter  map[string]string
		// Mutations to verify that Copy can be mutated.
		mutateCopy  map[string]string
		wantCopyEnd map[string]string
		// Want the original to only be mutated if mutateOriginal is set.
		wantOrigEnd map[string]string
	}

	tests := []wantCase{
		{
			name:  "nil receiver",
			input: MapStringStringOmitNull(nil),

			wantCopyInitial: map[string]string{},
			mutateCopy:      map[string]string{"a": "b"},

			wantCopyEnd: map[string]string{"a": "b"},

			wantOrigEnd: map[string]string{},
		},
		{
			name:            "empty map",
			input:           MapStringStringOmitNull{},
			wantCopyInitial: map[string]string{},

			mutateCopy:  map[string]string{"a": "b"},
			wantCopyEnd: map[string]string{"a": "b"},

			wantOrigEnd: map[string]string{},
		},
		{
			name: "non-empty map independent",
			input: MapStringStringOmitNull{
				"a": "b",
				"c": "d",
			},
			wantCopyInitial: map[string]string{
				"a": "b",
				"c": "d",
			},

			mutateOriginal: map[string]string{"a": "z"},
			wantCopyAfter: map[string]string{
				"a": "b",
				"c": "d",
			},

			mutateCopy: map[string]string{"c": "y"},
			wantCopyEnd: map[string]string{
				"a": "b",
				"c": "y",
			},

			wantOrigEnd: map[string]string{
				"a": "z",
				"c": "d",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orig := tc.input
			origNil := orig == nil

			// WHEN: Copy is called.
			inputCopy := orig.Copy()

			// THEN: the copy should not be nil.
			if inputCopy == nil {
				t.Fatalf("%s\nCopy() returned nil pointer", packageName)
			}
			if *inputCopy == nil {
				t.Fatalf("%s\nCopy() returned nil map", packageName)
			}

			// AND: the copy should be a copy of the original.
			if testErr := test.AssertMapEqual(
				t,
				*inputCopy,
				tc.wantCopyInitial,
				fmt.Sprintf("%s\nCopy()", packageName),
				"",
			); testErr != nil {
				t.Error(testErr)
			}

			if tc.mutateOriginal != nil {
				// WHEN: the original is mutated.
				for k, v := range tc.mutateOriginal {
					orig[k] = v
				}
				// THEN: the copy should be unchanged.
				if testErr := test.AssertMapEqual(
					t,
					*inputCopy,
					tc.wantCopyAfter,
					fmt.Sprintf("%s\nCopy() mismatch after mutating original", packageName),
					"result",
				); testErr != nil {
					t.Error(testErr)
				}
			}

			if tc.mutateCopy != nil {
				// WHEN: the copy is mutated.
				for k, v := range tc.mutateCopy {
					(*inputCopy)[k] = v
				}
				// THEN: the original should be the mutated value.
				if testErr := test.AssertMapEqual(
					t,
					*inputCopy,
					tc.wantCopyEnd,
					fmt.Sprintf("%s\nCopy() mismatch after mutating result", packageName),
					"input",
				); testErr != nil {
					t.Error(testErr)
				}
			}

			// AND: the original should remain nil if it was nil before.
			if origNil && orig != nil {
				t.Fatalf(
					"%s\noriginal map should still be nil; got len=%d",
					packageName, len(orig),
				)
			}

			// AND: the original should be unchanged after all mutations.
			if testErr := test.AssertMapEqual(
				t,
				orig,
				tc.wantOrigEnd,
				fmt.Sprintf("%s\nCopy() mismatch after all mutations -", packageName),
				"original",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
