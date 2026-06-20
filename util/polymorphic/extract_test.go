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
	"github.com/release-argus/Argus/util/errfmt"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestExtract(t *testing.T) {
	type Args struct {
		format string
		raw    []byte
		key    string
	}

	// GIVEN: a variety of raw input and a key.
	tests := []struct {
		name       string
		args       Args
		expected   []byte
		errorRegex string
	}{
		{
			name: "nil raw input",
			args: Args{
				format: "json",
				raw:    nil,
				key:    "test",
			},
			expected:   nil,
			errorRegex: `^$`,
		},
		{
			name: "key not present in raw input",
			args: Args{
				format: "json",
				raw:    []byte(`{"foo":"bar"}`),
				key:    "something",
			},
			expected:   nil,
			errorRegex: `^$`,
		},
		{
			name: "key present in JSON format",
			args: Args{
				format: "json",
				raw:    []byte(`{"key":"value"}`),
				key:    "key",
			},
			expected: func() []byte {
				b, _ := decode.Marshal("json", "value")
				return b
			}(),
			errorRegex: `^$`,
		},
		{
			name: "key present in YAML format",
			args: Args{
				format: "yaml",
				raw:    []byte(`{"key":"value"}`),
				key:    "key",
			},
			expected: func() []byte {
				b, _ := decode.Marshal("yaml", "value")
				return b
			}(),
			errorRegex: `^$`,
		},
		{
			name: "unsupported format",
			args: Args{
				format: "xml",
				raw:    []byte(`{"foo":"value"}`),
				key:    "key",
			},
			expected: nil,
			errorRegex: test.TrimYAML(`
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
			errorRegex: test.TrimYAML(`
				^extract "key":
					jsontext: .*
						invalid character .*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: Extract is called.
			result, err := Extract(tc.args.format, tc.args.raw, tc.args.key)

			prefix := fmt.Sprintf(
				"%s\nExtract(format=%q, data=%q, key=%q)",
				packageName, tc.args.format, tc.args.raw, tc.args.key,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errorRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %s\nwant: %s",
					prefix, e, tc.errorRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the result is as expected.
			if !util.AreSlicesEqual(tc.expected, result) {
				t.Errorf(
					"%s result mismatch\ngot:  %q\nwant: %q",
					prefix, result, tc.expected,
				)
			}
		})
	}
}

func TestExtract__marshalError(t *testing.T) {
	// GIVEN: a failing marshal function.
	original := marshalExtractSubtree
	customErr := fmt.Errorf("marshal failed")
	marshalExtractSubtree = func(format string, m any) ([]byte, error) {
		return nil, customErr
	}
	t.Cleanup(func() { marshalExtractSubtree = original })

	// AND: a key that is present to extract from valid data.
	key := "foo"
	data := []byte(`{"` + key + `":"value"}`)

	errRegex := key + ":\n  " + customErr.Error()

	// WHEN: Extract is called on this key.
	result, err := Extract("json", data, key)

	prefix := fmt.Sprintf(
		"%s\nExtract(format=%q, data=%q, key=%q)",
		packageName, "json", data, key,
	)

	// THEN: the marshal error is returned.
	e := errfmt.FormatError(err)
	if !util.RegexCheck(test.TrimYAML(errRegex), e) {
		t.Errorf(
			"%s error mismatch\ngot:  %q\nwant: %q",
			prefix, e, errRegex,
		)
	}

	// AND: nil is marshaled.
	if result != nil {
		t.Errorf(
			"%s result mismatch\ngot:  %v\nwant: nil",
			prefix, result,
		)
	}
}
