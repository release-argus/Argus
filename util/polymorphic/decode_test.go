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

package polymorphic

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

type dummyType struct {
	Type string `json:"type" yaml:"type"`
	Name string `json:"name" yaml:"name"`
}

func (d dummyType) GetType() string {
	return "test"
}
func (d dummyType) ApplyOverrides(s string, bytes []byte) error {
	return nil
}
func (d dummyType) DecodeSelf(s string, bytes []byte) error {
	return nil
}

func TestToInheritableMap(t *testing.T) {
	// GIVEN: a map of constructors.
	tests := []struct {
		name         string
		constructors map[string]func() Inheritable
	}{
		{
			name:         "empty map",
			constructors: map[string]func() Inheritable{},
		},
		{
			name: "single constructor",
			constructors: map[string]func() Inheritable{
				"dummy": func() Inheritable { return &dummyType{} },
			},
		},
		{
			name: "dual constructors",
			constructors: map[string]func() Inheritable{
				"github": func() Inheritable { return &dummyType{} },
				"web":    func() Inheritable { return &dummyType{} },
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: ToInheritableMap is called.
			got := ToInheritableMap(tc.constructors)

			// THEN: we can construct all the keys in the map.
			for key := range tc.constructors {
				g := got[key]()
				if g == nil {
					t.Errorf(
						"%s\nToInheritableMap[%q] gave nil",
						packageName, key,
					)
				}
			}
		})
	}
}

func TestResolveType(t *testing.T) {
	type Args struct {
		format      string
		data        []byte
		previous    Inheritable
		defaultType string
	}
	// GIVEN: data to extract a type from, the default type, and an Inheritable to fallback on.
	tests := []struct {
		name     string
		args     Args
		want     string
		errRegex string
	}{
		{
			name: "empty data",
			args: Args{
				format:      "json",
				data:        []byte{},
				previous:    &dummyType{},
				defaultType: "test",
			},
			want:     "test",
			errRegex: `^$`,
		},
		{
			name: "fail to unmarshal",
			args: Args{
				format:      "json",
				data:        []byte(`invalid json`),
				previous:    &dummyType{},
				defaultType: "test",
			},
			errRegex: `invalid character`,
		},
		{
			name: "no type / previous - use defaultType",
			args: Args{
				format:      "json",
				data:        []byte{},
				previous:    nil,
				defaultType: "something",
			},
			want:     "something",
			errRegex: `^$`,
		},
		{
			name: "no type - use previous",
			args: Args{
				format:      "json",
				data:        []byte(`{"name":"test"}`),
				previous:    &dummyType{},
				defaultType: "something",
			},
			want: "test",
		},
		{
			name: "JSON/type present",
			args: Args{
				format: "json",
				data: []byte(test.TrimJSON(`{
					"type":"hello",
					"name":"test"
				}`)),
				previous:    &dummyType{},
				defaultType: "something",
			},
			want: "hello",
		},
		{
			name: "YAML/type present",
			args: Args{
				format: "yaml",
				data: []byte(test.TrimYAML(`
					type: hello
					name: test
				`)),
				previous:    &dummyType{},
				defaultType: "something",
			},
			want: "hello",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: ResolveType is called.
			got, err := ResolveType(
				tc.args.format, tc.args.data,
				tc.args.previous,
				tc.args.defaultType,
			)

			prefix := fmt.Sprintf("%s\nResolveType()", packageName)

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s value mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}

			// AND: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
		})
	}
}

type test1 struct {
	Name, Other string
}

func (t *test1) GetType() string                             { return "test1" }
func (t *test1) ApplyOverrides(s string, bytes []byte) error { return nil }
func (t *test1) DecodeSelf(s string, bytes []byte) error     { return nil }

type test2 struct {
	Name, Other string
}

func (t *test2) GetType() string                             { return "test2" }
func (t *test2) ApplyOverrides(s string, bytes []byte) error { return nil }
func (t *test2) DecodeSelf(s string, bytes []byte) error     { return nil }

type test3 struct {
	Name, Other string
}

func (t *test3) GetType() string                             { return "test3" }
func (t *test3) ApplyOverrides(s string, bytes []byte) error { return nil }
func (t *test3) DecodeSelf(s string, bytes []byte) error     { return nil }

var constructors = map[string]func() Inheritable{
	"test1": func() Inheritable { return &test1{} },
	"test2": func() Inheritable { return &test2{} },
	"test3": func() Inheritable { return &test3{} },
}

func TestConstruct(t *testing.T) {
	// GIVEN: a set of constructors and a type to construct.
	tests := []struct {
		name     string
		typ      string
		errRegex string
	}{
		{
			name:     "known type",
			typ:      "test1",
			errRegex: `^$`,
		},
		{
			name:     "unknown type",
			typ:      "test4",
			errRegex: `^type: "test4" <invalid> \(supported values = \['test1', 'test2', 'test3'\]\)$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Construct is called to produce this struct.
			got, err := Construct(tc.typ, constructors)

			prefix := fmt.Sprintf(
				"%s\nConstruct(%s)",
				packageName, tc.typ,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					packageName, e, tc.errRegex,
				)
			}
			if e != "" {
				if got != nil {
					t.Errorf(
						"%s expected nil to be constructed with an error (type=%T): %q",
						prefix, got, e,
					)
				}
				return
			}

			// AND: the result is as expected.
			if gotType := got.GetType(); gotType != tc.typ {
				t.Errorf(
					"%s type mismatch\ngot:  %q\nwant: %q",
					prefix, gotType, tc.typ,
				)
			}
		})
	}
}

func TestInstantiate(t *testing.T) {
	type args struct {
		format      string
		data        []byte
		previous    Inheritable
		defaultType string
	}
	// GIVEN: a type to instantiate into an object.
	tests := []struct {
		name     string
		args     args
		wantNil  bool
		wantType string
		errRegex string
	}{
		{
			name: "no data",
			args: args{
				format:      "json",
				data:        []byte{},
				previous:    &dummyType{},
				defaultType: "test",
			},
			wantNil:  true,
			errRegex: `^type: "test" <invalid>.*$`,
		},
		{
			name: "null data",
			args: args{
				format:      "json",
				data:        []byte("null"),
				previous:    &dummyType{},
				defaultType: "test1",
			},
			wantNil:  true,
			errRegex: `^type: "test" <invalid>.*$`,
		},
		{
			name: "invalid data",
			args: args{
				format: "json",
				data: []byte(test.TrimJSON(`{
					"type":"test1",
					"name":"test"
				`)),
				previous:    &dummyType{},
				defaultType: "test",
			},
			errRegex: `unexpected`,
		},
		{
			name: "type from data",
			args: args{
				format: "json",
				data: []byte(test.TrimJSON(`{
					"type":"test1",
					"name":"test"
				}`)),
				previous:    &dummyType{},
				defaultType: "test",
			},
			wantType: "test1",
			errRegex: `^$`,
		},
		{
			name: "type from previous",
			args: args{
				format:      "json",
				data:        []byte(`{"name":"test"}`),
				previous:    &dummyType{},
				defaultType: "test",
			},
			errRegex: `^type: "[^"]+" <invalid> .*$`,
		},
		{
			name: "type from defaultType",
			args: args{
				format:      "json",
				data:        []byte(`{"name":"test"}`),
				previous:    nil,
				defaultType: "test2",
			},
			wantType: "test2",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Instantiate(
				tc.args.format, tc.args.data,
				tc.args.previous,
				tc.args.defaultType,
				constructors,
			)

			prefix := fmt.Sprintf(
				"%s\nInstantiate(format=%q, data=%q)",
				packageName, tc.args.format, tc.args.data,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if tc.wantNil && got != nil {
				t.Fatalf(
					"%s expected nil to be instantiated:\ngot: %T",
					prefix, got,
				)
			}
			if e != "" || tc.wantNil {
				return
			}

			// AND: the struct type is as expected.
			if gotType := got.GetType(); gotType != tc.wantType {
				t.Errorf(
					"%s type mismatch\ngot:  %q\nwant: %q",
					prefix, gotType, tc.wantType,
				)
			}
		})
	}
}

func TestApplyOverrides(t *testing.T) {
	type args struct {
		format string
		data   string
		target Inheritable
	}
	// GIVEN: overrides to apply to a struct.
	tests := []struct {
		name     string
		args     args
		want     string
		errRegex string
	}{
		{
			name: "empty overrides",
			args: args{
				format: "yaml",
				data:   "",
				target: &test1{
					Name:  "a",
					Other: "b",
				},
			},
			want: test.TrimYAML(`
				name: a
				other: b
			`),
			errRegex: `^$`,
		},
		{
			name: "null data gives nil struct",
			args: args{
				format: "yaml",
				data:   "null",
				target: &test1{
					Name:  "a",
					Other: "b",
				},
			},
			want:     "",
			errRegex: `^$`,
		},
		{
			name: "fail to extract type",
			args: args{
				format: "yaml",
				data:   `{ abc: 123 `,
				target: &test1{
					Name:  "a",
					Other: "b",
				},
			},
			want:     "",
			errRegex: `^[^\s]+ could not find flow .*`,
		},
		{
			name: "type change - unknown type",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					type: test4
					name: a
					other: b
				`),
				target: &test1{
					Name:  "a",
					Other: "b",
				},
			},
			want:     "",
			errRegex: `^type: "[^"]+" <invalid> .*$`,
		},
		{
			name: "type change - known type - invalid data type",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					type: test3
					name: [c]
				`),
				target: &test1{
					Name: "c",
				},
			},
			want:     "",
			errRegex: `^[^\s]+ .*unmarshal`,
		},
		{
			name: "type change - known type",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					type: test3
					name: c
				`),
				target: &test1{
					Name: "c",
				},
			},
			want: test.TrimYAML(`
				name: c
				other: ''
			`),
			errRegex: `^$`,
		},
		{
			name: "no previous - create new struct",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					type: test1
					name: hi
					other: bye
				`),
			},
			want: test.TrimYAML(`
				name: hi
				other: bye
			`),
			errRegex: `^$`,
		},
		{
			name: "type unchanged - all vars changed",
			args: args{
				format: "yaml",
				data: test.TrimYAML(`
					name: hi
					other: bye
				`),
				target: &test1{
					Name:  "a",
					Other: "b",
				},
			},
			want: test.TrimYAML(`
				name: hi
				other: bye
			`),
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the ApplyOverrides is called.
			test.AssertApplyOverrides(
				t,
				tc.args.target,
				func(format string, data []byte, target Inheritable) (Inheritable, error) {
					return ApplyOverrides(
						format, data,
						target,
						"",
						constructors,
					)
				},
				tc.args.format, tc.args.data,
				func(v Inheritable) string { return decode.ToYAMLString(v, "") },
				tc.want,
				tc.errRegex,
				true,
				packageName,
				"ApplyOverrides",
			)
		})
	}
}
