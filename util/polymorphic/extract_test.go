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
	"errors"
	"testing"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestExtract(t *testing.T) {
	type Args struct {
		format string
		raw    []byte
		key    string
	}

	tests := []struct {
		name     string
		args     Args
		expected []byte
		errRegex string
	}{
		{
			name: "nil raw input",
			args: Args{
				format: "json",
				raw:    nil,
				key:    "test",
			},
			expected: nil,
			errRegex: `^$`,
		},
		{
			name: "key not present in raw input",
			args: Args{
				format: "json",
				raw:    []byte(`{"foo":"bar"}`),
				key:    "something",
			},
			expected: nil,
			errRegex: `^$`,
		},
		{
			name: "key present in JSON format",
			args: Args{
				format: "json",
				raw:    []byte(`{"key":"value"}`),
				key:    "key",
			},
			expected: []byte(`"value"`),
			errRegex: `^$`,
		},
		{
			name: "key present in YAML format",
			args: Args{
				format: "yaml",
				raw:    []byte(`{"key":"value"}`),
				key:    "key",
			},
			expected: []byte("value\n"),
			errRegex: `^$`,
		},
		{
			name: "unsupported format",
			args: Args{
				format: "xml",
				raw:    []byte(`{"foo":"value"}`),
				key:    "key",
			},
			expected: nil,
			errRegex: test.TrimYAML(`
				^extract "key":
					unsupported format: "xml"$`,
			),
		},
		{
			name: "error unmarshaling raw input",
			args: Args{
				format: "json",
				raw:    []byte(`{foo:value}`),
				key:    "key",
			},
			expected: nil,
			errRegex: test.TrimYAML(`
				^extract "key":
					jsontext: .*
						invalid character .*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := Extract(tc.args.format, tc.args.raw, tc.args.key)
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, tc.errRegex)
			}
			if err != nil {
				return
			}

			if !util.AreSlicesEqual(tc.expected, result) {
				t.Errorf("result mismatch\ngot:  %q\nwant: %q", result, tc.expected)
			}
		})
	}
}

func TestExtract__marshalError(t *testing.T) {
	original := marshalExtractSubtree
	customErr := errors.New("marshal failed")
	marshalExtractSubtree = func(format string, m any) ([]byte, error) {
		return nil, customErr
	}
	t.Cleanup(func() { marshalExtractSubtree = original })

	key := "foo"
	data := []byte(`{"` + key + `":"value"}`)

	errRegex := test.TrimYAML(key + ":\n  " + customErr.Error())

	result, err := Extract("json", data, key)
	e := errfmt.FormatError(err)
	if !util.RegexCheck(errRegex, e) {
		t.Errorf("error mismatch\ngot:  %q\nwant: %q", e, errRegex)
	}

	if result != nil {
		t.Errorf("result mismatch\ngot:  %q\nwant: nil", result)
	}
}
