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

package decode

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

type testParent struct {
	A string    `yaml:"a"`
	B testChild `yaml:"b"`
}
type testChild struct {
	C string `yaml:"c"`
}

func TestNewYAMLEncoder_Indent(t *testing.T) {
	// GIVEN: an amount of spaces to indent with.
	tests := []struct {
		spaces int
	}{
		{spaces: 2},
		{spaces: 4},
		{spaces: 8},
		{spaces: 16},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("%d spaces", tc.spaces)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// AND: a NewYAMLEncoder created with this indent count.
			var buf bytes.Buffer
			enc := NewYAMLEncoder(&buf, tc.spaces)

			// AND: a struct to encode with it.
			txt := "hello"
			a := testParent{
				A: txt,
				B: testChild{
					C: txt,
				},
			}

			// WHEN: the struct is encoded.
			if err := enc.Encode(a); err != nil {
				t.Fatalf(
					"%s\nNewYAMLEncoder.Encode() error = %v",
					packageName, err,
				)
			}
			baseResult := "a: hello\nb:\n  c: hello\n"

			// THEN: the result is as expected.
			got := buf.String()
			want := strings.ReplaceAll(
				baseResult,
				"b:\n  ",
				"b:\n"+strings.Repeat(" ", tc.spaces),
			)
			if got != want {
				t.Errorf(
					"%s\nNewYAMLEncoder.Encode() mismatch\ngot:  %q\nwant: %q",
					packageName, got, want,
				)
			}
		})
	}
}

func TestNewYAMLEncoder_IndentSequence(t *testing.T) {
	// GIVEN: a YAML encoder.
	var buf bytes.Buffer
	enc := NewYAMLEncoder(&buf, 2)

	// WHEN: a sequence is encoded.
	seq := []string{"hello", "world"}
	if err := enc.Encode(seq); err != nil {
		t.Fatalf(
			"%s\nNewYAMLEncoder.Encode() error = %v",
			packageName, err,
		)
	}

	// THEN: the result is as expected.
	got := buf.String()
	want := "  - hello\n  - world\n"
	if got != want {
		t.Errorf(
			"%s\nNewYAMLEncoder.Encode() mismatch\ngot:  %q\nwant: %q",
			packageName, got, want,
		)
	}
}

func TestNewYAMLEncoder_UseLiteralStyleIfMultiline(t *testing.T) {
	// GIVEN: a YAML encoder.
	var buf bytes.Buffer
	enc := NewYAMLEncoder(&buf, 2)

	// WHEN: a multiline string is encoded.
	str := "hello\nworld"
	if err := enc.Encode(str); err != nil {
		t.Fatalf(
			"%s\nNewYAMLEncoder.Encode() error = %v",
			packageName, err,
		)
	}

	// THEN: the result is as expected.
	got := buf.String()
	want := "|-\n  hello\n  world\n"
	if got != want {
		t.Errorf(
			"%s\nNewYAMLEncoder.Encode() mismatch\ngot:  %q\nwant: %q",
			packageName, got, want,
		)
	}
}

func TestNewYAMLEncoder_UseSingleQuote(t *testing.T) {
	// GIVEN: a YAML encoder.
	var buf bytes.Buffer
	enc := NewYAMLEncoder(&buf, 2)

	// WHEN: a string of numbers is encoded.
	str := "123"
	if err := enc.Encode(str); err != nil {
		t.Fatalf(
			"%s\nNewYAMLEncoder.Encode() error = %v",
			packageName, err,
		)
	}

	// THEN: the result is as expected.
	got := buf.String()
	want := "'123'\n"
	if got != want {
		t.Errorf(
			"%s\nNewYAMLEncoder.Encode() mismatch\ngot:  %q\nwant: %q",
			packageName, got, want,
		)
	}
}

type customErrorMarshal struct{}

func (c customErrorMarshal) MarshalYAML() (any, error) {
	return nil, fmt.Errorf("intentional marshal error")
}

func TestToYAMLString(t *testing.T) {
	// GIVEN: a struct to print in YAML format.
	tests := []struct {
		name     string
		input    any
		wantYAML string
	}{
		{
			name:     "invalid data",
			input:    customErrorMarshal{},
			wantYAML: "",
		},
		{
			name:     "empty struct",
			input:    struct{}{},
			wantYAML: "{}\n",
		},
		{
			name: "nested struct",
			input: struct {
				Test struct {
					Foo string `yaml:"foo" json:"foo"`
				} `yaml:"test" json:"test"`
			}{
				Test: struct {
					Foo string `yaml:"foo" json:"foo"`
				}{
					Foo: "bar",
				},
			},
			wantYAML: "test:\n  foo: bar\n",
		},
		{
			name: "simple struct",
			input: struct {
				Test string `yaml:"test" json:"test"`
			}{
				Test: "test",
			},
			wantYAML: "test: test\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test.AssertStringWithPrefixes(
				t,
				packageName,
				func(prefix string) string {
					return ToYAMLString(tc.input, prefix)
				},
				tc.wantYAML,
			)
		})
	}
}
