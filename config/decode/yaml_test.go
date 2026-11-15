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
	"errors"
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

func TestNewYAMLEncoder__indent(t *testing.T) {
	tests := []struct {
		name   string
		spaces int
	}{
		{name: "2 spaces", spaces: 2},
		{name: "4 spaces", spaces: 4},
		{name: "8 spaces", spaces: 8},
		{name: "16 spaces", spaces: 16},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			enc := NewYAMLEncoder(&buf, tc.spaces)

			txt := "hello"
			a := testParent{
				A: txt,
				B: testChild{
					C: txt,
				},
			}

			err := enc.Encode(a)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := buf.String()
			want := strings.ReplaceAll(
				"a: hello\nb:\n  c: hello\n",
				"b:\n  ",
				"b:\n"+strings.Repeat(" ", tc.spaces),
			)
			if got != want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

// testInterfaceMarshaler substitutes itself with another value to marshal
// (yaml.InterfaceMarshaler), like service.Service does.
type testInterfaceMarshaler struct{}

func (m testInterfaceMarshaler) MarshalYAML() (any, error) {
	return testParent{
		A: "hello",
		B: testChild{
			C: "hello",
		},
	}, nil
}

func TestNewYAMLEncoder__nestedMarshaler(t *testing.T) {
	tests := []struct {
		name   string
		spaces int
	}{
		{name: "2 spaces", spaces: 2},
		{name: "4 spaces", spaces: 4},
		{name: "8 spaces", spaces: 8},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			enc := NewYAMLEncoder(&buf, tc.spaces)

			a := struct {
				Service testInterfaceMarshaler `yaml:"service"`
			}{}

			err := enc.Encode(a)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			indent := strings.Repeat(" ", tc.spaces)
			got := buf.String()
			want := "service:\n" +
				indent + "a: hello\n" +
				indent + "b:\n" +
				indent + indent + "c: hello\n"
			if got != want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

func TestNewYAMLEncoder__indentSequence(t *testing.T) {
	var buf bytes.Buffer
	enc := NewYAMLEncoder(&buf, 2)

	seq := []string{"hello", "world"}
	err := enc.Encode(seq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	want := "  - hello\n  - world\n"
	if got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestNewYAMLEncoder__useLiteralStyleIfMultiline(t *testing.T) {
	var buf bytes.Buffer
	enc := NewYAMLEncoder(&buf, 2)

	str := "hello\nworld"
	err := enc.Encode(str)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	want := "|-\n  hello\n  world\n"
	if got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestNewYAMLEncoder__useSingleQuote(t *testing.T) {
	var buf bytes.Buffer
	enc := NewYAMLEncoder(&buf, 2)

	str := "123"
	err := enc.Encode(str)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	// Quoted so it is not read back as a number.
	want := "'123'\n"
	if got != want {
		t.Errorf("mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

type customErrorMarshal struct{}

func (c customErrorMarshal) MarshalYAML() (any, error) {
	return nil, errors.New("intentional marshal error")
}

func TestToYAMLString(t *testing.T) {
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
				"ToYAMLString",
				func(prefix string) string {
					return ToYAMLString(tc.input, prefix)
				},
				tc.wantYAML,
			)
		})
	}
}
