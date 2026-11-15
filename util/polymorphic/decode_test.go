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
	"maps"
	"slices"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

type dummyType struct {
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
				"dummy": func() Inheritable { return &dummyType{Name: "dummy"} },
			},
		},
		{
			name: "dual constructors",
			constructors: map[string]func() Inheritable{
				"github": func() Inheritable { return &dummyType{Name: "github"} },
				"web":    func() Inheritable { return &dummyType{Name: "web"} },
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ToInheritableMap(tc.constructors)

			gotKeys := slices.Sorted(maps.Keys(got))
			wantKeys := slices.Sorted(maps.Keys(tc.constructors))
			if !slices.Equal(gotKeys, wantKeys) {
				t.Fatalf("keys mismatch\ngot:  %q\nwant: %q", gotKeys, wantKeys)
			}

			for key := range tc.constructors {
				var name string
				if d, ok := got[key]().(*dummyType); ok {
					name = d.Name
				}
				if name != key {
					t.Errorf("got[%q]().Name mismatch\ngot:  %q\nwant: %q", key, name, key)
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
				defaultType: "something",
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
			name: "no type, no previous, use defaultType",
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
			name: "no type, use previous",
			args: Args{
				format:      "json",
				data:        []byte(`{"name":"test"}`),
				previous:    &dummyType{},
				defaultType: "something",
			},
			want:     "test",
			errRegex: `^$`,
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
			want:     "hello",
			errRegex: `^$`,
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
			want:     "hello",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ResolveType(
				tc.args.format, tc.args.data,
				tc.args.previous,
				tc.args.defaultType,
			)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}

			if got != tc.want {
				t.Errorf("type mismatch\ngot:  %q\nwant: %q", got, tc.want)
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

			got, err := Construct(tc.typ, constructors)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				if got != nil {
					t.Errorf("result mismatch\ngot:  %T\nwant: nil", got)
				}
				return
			}

			if gotType := got.GetType(); gotType != tc.typ {
				t.Errorf("Inheritable.GetType() mismatch\ngot:  %q\nwant: %q", gotType, tc.typ)
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

	tests := []struct {
		name     string
		args     args
		wantType string
		errRegex string
	}{
		{
			name: "no data",
			args: args{
				format:      "json",
				data:        []byte{},
				previous:    &dummyType{},
				defaultType: "something",
			},
			errRegex: `^type: "test" <invalid>.*$`,
		},
		{
			name: "null data",
			args: args{
				format:      "json",
				data:        []byte("null"),
				previous:    &dummyType{},
				defaultType: "something",
			},
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
				defaultType: "something",
			},
			errRegex: `^type: "test" <invalid> .*$`,
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
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				if got != nil {
					t.Errorf("result mismatch\ngot:  %T\nwant: nil", got)
				}
				return
			}

			if gotType := got.GetType(); gotType != tc.wantType {
				t.Errorf("Inheritable.GetType() mismatch\ngot:  %q\nwant: %q", gotType, tc.wantType)
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

	tests := []struct {
		name        string
		args        args
		sameAddress bool
		want        string
		errRegex    string
	}{
		{
			name:        "empty overrides",
			sameAddress: true,
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
			name: "type change/unknown type",
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
			name: "type change/known type/invalid data type",
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
			name: "type change/known type/valid",
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
			name: "no previous, create new struct",
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
			name:        "type unchanged, all vars changed",
			sameAddress: true,
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

			if _, _, testErr := test.AssertApplyOverrides(
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
				tc.sameAddress,
				"ApplyOverrides",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}
