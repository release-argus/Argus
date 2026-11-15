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
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

type struct1 struct {
	Field1 string `json:"field_1" yaml:"field_1"`
	Field2 int    `json:"field_2" yaml:"field_2"`
}
type struct2 struct {
	Field1 string  `json:"field_1" yaml:"field_1"`
	Field2 int     `json:"field_2" yaml:"field_2"`
	Field3 struct1 `json:"field_3" yaml:"field_3"`
}

func testData() struct2 {
	return struct2{
		Field1: "test",
		Field2: 1,
		Field3: struct1{
			Field1: "test",
			Field2: 1,
		},
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     []byte
		key      string
		v        struct2
		want     string
		errRegex string
	}{
		{
			name:     "empty data",
			format:   "json",
			data:     []byte{},
			key:      "field1",
			v:        testData(),
			want:     decode.ToYAMLString(testData(), ""),
			errRegex: `^$`,
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     []byte("{}"),
			key:      "field1",
			v:        testData(),
			want:     decode.ToYAMLString(testData(), ""),
			errRegex: `^$`,
		},
		{
			name:   "Extract decode",
			format: "json",
			data:   []byte(`{"abc": "123"`),
			key:    "field1",
			v:      testData(),
			errRegex: test.TrimYAML(`
				^extract "field1":
					[^\s]+ unexpected EOF`,
			),
		},
		{
			name:   "unknown key",
			format: "yaml",
			data: []byte(test.TrimYAML(`
				foo:
					field_1: a
					field_2: 2
					field_3:
						field_1: b
						field_2: 2
			`)),
			key:      "bar",
			v:        testData(),
			want:     decode.ToYAMLString(testData(), ""),
			errRegex: `^$`,
		},
		{
			name:   "known key/valid",
			format: "yaml",
			data: []byte(test.TrimYAML(`
				foo:
					field_1: a
					field_2: 2
					field_3:
						field_1: b
						field_2: 2
			`)),
			key: "foo",
			v:   testData(),
			want: test.TrimYAML(`
				field_1: a
				field_2: 2
				field_3:
					field_1: b
					field_2: 2
			`),
			errRegex: `^$`,
		},
		{
			name:   "known key/invalid data type",
			format: "yaml",
			data: []byte(test.TrimYAML(`
				foo:
					field_1: [a]
			`)),
			key: "foo",
			v:   testData(),
			errRegex: test.TrimYAML(`
				^foo:
					[^\s]+ .*unmarshal`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := Unmarshal(tc.format, tc.data, tc.key, &tc.v)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if got := decode.ToYAMLString(tc.v, ""); got != tc.want {
				t.Errorf("decode.ToYAMLString() mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}
