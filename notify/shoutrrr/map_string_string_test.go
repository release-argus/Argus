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
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
)

func TestMapStringStringOmitNull_Unmarshal(t *testing.T) {
	tests := []struct {
		name         string
		format, data string
		want         string
		errRegex     string
	}{
		{
			name:   "JSON/empty",
			format: "json",
			data:   "",
			errRegex: test.TrimYAML(`
				^jsontext:
					unexpected EOF$`,
			),
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
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

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*MapStringStringOmitNull, error) {
					var zero MapStringStringOmitNull
					err := decode.Unmarshal(format, data, &zero)
					return &zero, err
				},
				tc.format, tc.data,
				func(v *MapStringStringOmitNull) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				"MapStringStringOmitNull.Unmarshal",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

func TestMapStringStringOmitNull_Copy(t *testing.T) {
	tests := []struct {
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
	}{
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

			inputCopy := orig.Copy()

			// util.CopyMap always makes a map, even for a nil receiver.
			if inputCopy == nil {
				t.Fatal("copy mismatch\ngot:  nil\nwant: non-nil")
			}
			if *inputCopy == nil {
				t.Fatal("copy map mismatch\ngot:  nil\nwant: non-nil")
			}

			if testErr := test.AssertMapEqual(
				t,
				*inputCopy,
				tc.wantCopyInitial,
				"copy",
			); testErr != nil {
				t.Errorf("after Copy: %v", testErr)
			}

			if tc.mutateOriginal != nil {
				for k, v := range tc.mutateOriginal {
					orig[k] = v
				}

				if testErr := test.AssertMapEqual(
					t,
					*inputCopy,
					tc.wantCopyAfter,
					"copy",
				); testErr != nil {
					t.Errorf("mutate original: %v", testErr)
				}
			}

			if tc.mutateCopy != nil {
				for k, v := range tc.mutateCopy {
					(*inputCopy)[k] = v
				}

				if testErr := test.AssertMapEqual(
					t,
					*inputCopy,
					tc.wantCopyEnd,
					"copy",
				); testErr != nil {
					t.Errorf("mutate copy: %v", testErr)
				}
			}

			if testErr := test.AssertMapEqual(
				t,
				orig,
				tc.wantOrigEnd,
				"original",
			); testErr != nil {
				t.Errorf("end: %v", testErr)
			}
		})
	}
}
