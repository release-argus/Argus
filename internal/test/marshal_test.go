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
	"fmt"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestUnmarshal(t *testing.T) {
	// GIVEN: []byte and a format.
	tests := []struct {
		name     string
		data     []byte
		format   string
		want     string
		errRegex string
	}{
		{
			name:   "JSON/no data",
			data:   []byte{},
			format: "json",
			want:   "{}\n",
			errRegex: TrimYAML(`
				^jsontext:
					unexpected EOF$`,
			),
		},
		{
			name:     "YAML/no data",
			data:     []byte{},
			format:   "yaml",
			errRegex: `^$`,
			want:     "{}\n",
		},
		{
			name:     "unsupported format/no data",
			data:     []byte{},
			format:   "x",
			want:     "{}\n",
			errRegex: `^unsupported format: "x"$`,
		},
		{
			name: "JSON/valid data",
			data: []byte(TrimJSON(`{
				"string": "hi",
				"int": 6,
				"bool": true
			}`)),
			format: "json",
			want: TrimYAML(`
			string: hi
			int: 6
			bool: true
			`),
			errRegex: `^$`,
		},
		{
			name: "YAML/valid data",
			data: []byte(TrimYAML(`
				string: foo
				int: 42
				bool: true
			`)),
			format: "yaml",
			want: TrimYAML(`
			string: foo
			int: 42
			bool: true
			`),
			errRegex: `^$`,
		},
		{
			name:     "unsupported format/core",
			data:     []byte("{}"),
			format:   "txt",
			errRegex: `^unsupported format: "txt"$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Unmarshal is called on it.
			var target testStruct
			err := Unmarshal(tc.format, tc.data, &target)

			prefix := fmt.Sprintf(
				"%s\nUnmarshal(format=%s, data=%v)",
				packageName, tc.format, tc.data,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the result stringifies as expected.
			b, err := yaml.Marshal(target)
			if err != nil {
				t.Fatalf("%s: marshal error: %v", prefix, err)
			}
			if got := string(b); got != tc.want {
				t.Errorf(
					"%s stringified mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

type customUnmarshaler struct {
	got []byte
}

func (c *customUnmarshaler) UnmarshalJSON(data []byte) error {
	c.got = data
	return nil
}

func (c *customUnmarshaler) UnmarshalYAML(data []byte) error {
	c.got = data
	return nil
}

func TestUnmarshal__customUnmarshaler(t *testing.T) {
	// GIVEN: a struct implementing json.Unmarshaler and yaml.BytesUnmarshaler.
	tests := []struct {
		name   string
		format string
		data   []byte
	}{
		{
			name:   "json.Unmarshaler",
			format: "json",
			data:   []byte(`{"key":"value"}`),
		},
		{
			name:   "yaml.Unmarshaler",
			format: "yaml",
			data:   []byte("key: value\n"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Unmarshal is called with a custom unmarshaler target.
			target := &customUnmarshaler{}
			err := Unmarshal(tc.format, tc.data, target)

			prefix := fmt.Sprintf(
				"%s\nUnmarshal(format=%s, data=%v)",
				packageName, tc.format, tc.data,
			)

			// THEN: no error is returned.
			if err != nil {
				t.Fatalf(
					"%s unexpected error: %q",
					prefix, err,
				)
			}

			// AND: the custom unmarshaler received the raw data.
			if string(target.got) != string(tc.data) {
				t.Errorf(
					"%s custom unmarshaler data mismatch\ngot:  %q\nwant: %q",
					prefix, target.got, tc.data,
				)
			}
		})
	}
}
