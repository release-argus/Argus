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

package filter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

// ############
// # DECODING #
// ############

func TestURLCommands_Unmarshal(t *testing.T) {
	// GIVEN: data in a given format to unmarshal into URLCommands.
	tests := []struct {
		name         string
		format, data string
		want         string
		expected     URLCommands
		errRegex     string
	}{
		{
			name:   "JSON/empty",
			format: "json",
			data:   "",
			want:   "",
			errRegex: test.TrimYAML(`
				^jsontext:
					unexpected EOF$`,
			),
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			want:     "[]\n",
			errRegex: `^$`,
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			want:     "[]\n",
			errRegex: `^$`,
		},
		{
			name:     "JSON/invalid",
			format:   "json",
			data:     `"`,
			expected: nil,
			errRegex: `unexpected`,
		},
		{
			name:     "YAML/invalid",
			format:   "yaml",
			data:     `"`,
			expected: nil,
			errRegex: `^[^\s]+ could not find end character.*`,
		},
		{
			name:   "JSON/list",
			format: "json",
			data: test.TrimJSON(`[
				{"type": "regex", "regex": "foo", "index": 1},
				{"type": "replace", "old": "bar", "new": "baz"},
				{"type": "split", "text": "abc", "index": 2}
			]`),
			want: "  " + strings.ReplaceAll(
				test.TrimYAML(`
					- type: regex
						regex: foo
						index: 1
					- type: replace
						old: bar
						new: baz
					- type: split
						text: abc
						index: 2`,
				),
				"\n", "\n  ",
			) + "\n",
			errRegex: `^$`,
		},
		{
			name:   "YAML/list",
			format: "yaml",
			data: test.TrimYAML(`
				- type: regex
					regex: \"([0-9.+])\"
					index: 1
				- type: replace
					old: foo
					new: bar
				- type: split
					text: abc
					index: 2
			`),
			errRegex: `^$`,
			want: "  " + strings.ReplaceAll(
				test.TrimYAML(`
					- type: regex
						regex: '\"([0-9.+])\"'
						index: 1
					- type: replace
						old: foo
						new: bar
					- type: split
						text: abc
						index: 2`,
				),
				"\n", "\n  ",
			) + "\n",
		},
		{
			name:   "JSON/single URLCommand",
			format: "json",
			data: test.TrimJSON(`{
				"type": "regex",
				"regex": "foo",
				"index": 1,
				"text": "hi",
				"old": "was",
				"new": "now"
			}`),
			want: "  " + strings.ReplaceAll(
				test.TrimYAML(`
					- type: regex
						regex: foo
						text: hi
						old: was
						new: now
						index: 1`,
				),
				"\n", "\n  ",
			) + "\n",
			errRegex: `^$`,
		},
		{
			name:   "YAML/single URLCommand",
			format: "yaml",
			data: test.TrimYAML(`
				type: regex
				regex: foo
				index: 1
				text: hi
				old: was
				new: now
			`),
			want: "  " + strings.ReplaceAll(
				test.TrimYAML(`
					- type: regex
						regex: foo
						text: hi
						old: was
						new: now
						index: 1`,
				),
				"\n", "\n  ",
			) + "\n",
			errRegex: `^$`,
		},
		{
			name:   "JSON/quoted string",
			format: "json",
			data:   `"[{\"type\":\"regex\",\"regex\":\"foo\",\"index\":1}]"`,
			want: "  " + strings.ReplaceAll(
				test.TrimYAML(`
					- type: regex
						regex: foo
						index: 1`,
				),
				"\n", "\n  ",
			) + "\n",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, _, testErr := test.AssertDecode(
				t,
				func(format string, data []byte) (*URLCommands, error) {
					var zero URLCommands
					err := decode.Unmarshal(format, data, &zero)
					return &zero, err
				},
				tc.format, tc.data,
				func(v *URLCommands) string { return v.String() },
				tc.want,
				tc.errRegex,
				packageName,
				"URLCommands",
			); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// ##########
// # STATE #
// ##########

func TestURLCommand_IsZero(t *testing.T) {
	// GIVEN: a URLCommand struct.
	tests := []struct {
		name string
		data URLCommand
		want bool
	}{
		{
			name: "empty",
			data: URLCommand{},
			want: true,
		},
		{
			name: "non-empty/Type",
			data: URLCommand{
				Type: "regex",
			},
			want: false,
		},
		{
			name: "non-empty/Regex",
			data: URLCommand{
				Regex: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/Text",
			data: URLCommand{
				Text: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/Old",
			data: URLCommand{
				Old: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/New",
			data: URLCommand{
				New: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/Index",
			data: URLCommand{
				Index: test.Ptr(3),
			},
			want: false,
		},
		{
			name: "non-empty/Template",
			data: URLCommand{
				Template: "foo",
			},
			want: false,
		},
		{
			name: "non-empty/regex without template",
			data: URLCommand{
				Type:  "regex",
				Regex: "foo",
				Text:  "v$0",
			},
			want: false,
		},
		{
			name: "non-empty/replace",
			data: URLCommand{
				Type: "replace",
				Old:  "foo",
				New:  "bar",
			},
			want: false,
		},
		{
			name: "non-empty/split",
			data: URLCommand{
				Type:  "split",
				Text:  "abc",
				Index: test.Ptr(2),
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: IsZero is called on it.
			got := tc.data.IsZero()

			// THEN: the result is as expected.
			if got != tc.want {
				t.Errorf(
					"%s\nURLCommand.IsZero() value mismatch\ngot:  %t\nwant: %t",
					packageName, got, tc.want,
				)
			}
		})
	}
}

// #############
// # STRINGIFY #
// #############

func TestURLCommands_String(t *testing.T) {
	// GIVEN: a URLCommands.
	tests := []struct {
		name        string
		urlCommands *URLCommands
		want        string
	}{
		{
			name: "regex",
			urlCommands: &URLCommands{
				testURLCommandRegex(),
			},
			want: strings.TrimPrefix(
				test.TrimYAML(`
					seq:
						- type: regex
							regex: -([0-9.]+)-
							index: 0
			`),
				"seq:\n",
			),
		},
		{
			name: "regex (templated)",
			urlCommands: &URLCommands{
				testURLCommandRegexTemplate(),
			},
			want: strings.TrimPrefix(
				test.TrimYAML(`
					seq:
						- type: regex
							regex: -([0-9.]+)-
							index: 0
							template: _$1_
				`),
				"seq:\n",
			),
		},
		{
			name: "replace",
			urlCommands: &URLCommands{
				testURLCommandReplace(),
			},
			want: strings.TrimPrefix(
				test.TrimYAML(`
					seq:
						- type: replace
							old: foo
							new: bar
				`),
				"seq:\n",
			),
		},
		{
			name: "split",
			urlCommands: &URLCommands{
				testURLCommandSplit(),
			},
			want: strings.TrimPrefix(
				test.TrimYAML(`
					seq:
						- type: split
							text: this
							index: 1
				`),
				"seq:\n",
			),
		},
		{
			name: "all types",
			urlCommands: &URLCommands{
				testURLCommandRegex(),
				testURLCommandReplace(),
				testURLCommandSplit(),
			},
			want: strings.TrimPrefix(
				test.TrimYAML(`
					seq:
						- type: regex
							regex: -([0-9.]+)-
							index: 0
						- type: replace
							old: foo
							new: bar
						- type: split
							text: this
							index: 1
				`),
				"seq:\n",
			),
		},
		{
			name:        "nil slice",
			urlCommands: nil,
			want:        "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			// WHEN: String is called on it.
			got := tc.urlCommands.String()

			// THEN: the expected string is returned.
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Fatalf(
					"%s\nURLCommands.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

func TestURLCommand_String(t *testing.T) {
	// GIVEN: a URLCommand.
	regex := testURLCommandRegex()
	replace := testURLCommandReplace()
	split := testURLCommandSplit()
	tests := []struct {
		name string
		cmd  *URLCommand
		want string
	}{
		{
			name: "regex",
			cmd:  &regex,
			want: test.TrimYAML(`
				type: regex
				regex: -([0-9.]+)-
				index: 0
			`),
		},
		{
			name: "replace",
			cmd:  &replace,
			want: test.TrimYAML(`
				type: replace
				old: foo
				new: bar
			`),
		},
		{
			name: "split",
			cmd:  &split,
			want: test.TrimYAML(`
				type: split
				text: this
				index: 1
			`),
		},
		{
			name: "nil slice",
			cmd:  nil,
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: String is called on it.
			got := tc.cmd.String()

			// THEN: the expected string is returned.
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Fatalf(
					"%s\nURLCommand.String() value mismatch\ngot:  %q\nwant: %q",
					packageName, got, tc.want,
				)
			}
		})
	}
}

// ##############
// # VALIDATION #
// ##############

func TestURLCommands_CheckValues(t *testing.T) {
	// GIVEN: a URLCommands.
	tests := []struct {
		name     string
		input    *URLCommands
		errRegex string
	}{
		{
			name:     "nil slice",
			input:    (*URLCommands)(nil),
			errRegex: `^$`,
		},
		{
			name: "valid regex",
			input: &URLCommands{
				testURLCommandRegex(),
			},
			errRegex: `^$`,
		},
		{
			name: "undefined regex",
			input: &URLCommands{
				{Type: "regex"},
			},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: regex
					regex: <required>.*$`,
			),
		},
		{
			name: "invalid regex",
			input: &URLCommands{
				{Type: "regex", Regex: `[0-`},
			},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: regex
					regex: .* <invalid> \(error parsing regexp.*\)$`,
			),
		},
		{
			name: "valid regex with template",
			input: &URLCommands{
				testURLCommandRegexTemplate(),
			},
			errRegex: `^$`,
		},
		{
			name: "valid regex with empty template",
			input: &URLCommands{
				{Type: "regex", Regex: `[0-]`, Template: ""},
			},
			errRegex: `^$`,
		},
		{
			name: "valid replace",
			input: &URLCommands{
				testURLCommandReplace(),
			},
			errRegex: `^$`,
		},
		{
			name: "invalid replace",
			input: &URLCommands{
				{Type: "replace"},
			},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: replace
					old: <required>.*$`,
			),
		},
		{
			name: "valid split",
			input: &URLCommands{
				testURLCommandSplit(),
			},
			errRegex: `^$`,
		},
		{
			name: "invalid split",
			input: &URLCommands{
				{Type: "split"},
			},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: split
					text: <required>`,
			),
		},
		{
			name: "invalid type",
			input: &URLCommands{
				{Type: "something"},
			},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: .* <invalid>.*$`,
			),
		},
		{
			name: "valid all types",
			input: &URLCommands{
				testURLCommandRegex(),
				testURLCommandReplace(),
				testURLCommandSplit(),
			},
			errRegex: `^$`,
		},
		{
			name: "all possible errors",
			input: &URLCommands{
				{Type: "regex"},
				{Type: "replace"},
				{Type: "split"},
				{Type: "something"},
			},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: regex
					regex: <required>.*
				- item_1:
					type: replace
					old: <required>.*
				- item_2:
					type: split
					text: <required>.*
				- item_3:
					type: "something" <invalid>.*$`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// CheckValues() shouldn't change the URLCommands.
			want := tc.input.String()

			_ = test.AssertCheckValuesWithError(
				t,
				packageName,
				tc.errRegex,
				tc.input.CheckValues,
			)

			// AND: the urlCommands is unchanged.
			if got := tc.input.String(); got != want {
				t.Errorf(
					"%s\nURLCommands.CheckValues() changed the slice unexpectedly:\ngot  %q\nwant: %q",
					packageName, got, want,
				)
			}
		})
	}
}

// ############
// # COMMANDS #
// ############

func TestURLCommands_GetVersions(t *testing.T) {
	// GIVEN: a URLCommands.
	testText := "abc123-def456"
	tests := []struct {
		name         string
		urlCommands  *URLCommands
		text         string
		wantVersions []string
		errRegex     string
	}{
		{
			name:         "empty slice",
			urlCommands:  &URLCommands{},
			text:         testText,
			wantVersions: []string{testText},
			errRegex:     `^$`,
		},
		{
			name:         "empty slice+text",
			urlCommands:  &URLCommands{},
			text:         "",
			wantVersions: nil,
			errRegex:     `^$`,
		},
		{
			name: "single version/regex",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.Ptr(1)},
			},
			text:         testText,
			wantVersions: []string{"def"},
			errRegex:     `^$`,
		},
		{
			name: "single version/replace",
			urlCommands: &URLCommands{
				{Type: "replace", Old: "-", New: " "},
			},
			text:         testText,
			wantVersions: []string{"abc123 def456"},
			errRegex:     `^$`,
		},
		{
			name: "multiple versions/split",
			urlCommands: &URLCommands{
				{Type: "split", Text: "-"},
			},
			text:         testText,
			wantVersions: []string{"abc123", "def456"},
			errRegex:     `^$`,
		},
		{
			name: "multiple versions/split fail",
			urlCommands: &URLCommands{
				{Type: "split", Text: "_"},
			},
			text:         testText,
			wantVersions: nil,
			errRegex:     `^split didn't find any "_" to split on$`,
		},
		{
			name: "multiple versions/regex and split",
			urlCommands: &URLCommands{
				{Type: "split", Text: "-", Index: nil},
				{Type: "regex", Regex: `([a-z]+)[0-9]+`},
			},
			text:         testText,
			wantVersions: []string{"abc", "def"},
			errRegex:     `^$`,
		},
		{
			name: "multiple versions/regex fail",
			urlCommands: &URLCommands{
				{Type: "split", Text: "-", Index: nil},
				{Type: "split", Text: "_", Index: test.Ptr(0)},
			},
			text:         testText,
			wantVersions: nil,
			errRegex:     `^split didn't find any "_" to split on$`,
		},
		{
			name: "regex doesn't match",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([h-z]+)[0-9]+`, Index: test.Ptr(1)},
			},
			text:         testText,
			wantVersions: nil,
			errRegex:     `regex .* didn't return any matches on "` + testText + `"`,
		},
		{
			name: "split index out of bounds",
			urlCommands: &URLCommands{
				{Type: "split", Text: "-", Index: test.Ptr(2)},
			},
			text:         testText,
			wantVersions: nil,
			errRegex:     `split .* returned \d elements on "[^']+", but the index wants element number \d`,
		},
		{
			name: "all types",
			urlCommands: &URLCommands{
				{Type: "split", Text: "-", Index: test.Ptr(0)},
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.Ptr(0)},
				{Type: "replace", Old: "b", New: "a"},
				{Type: "replace", Old: "c", New: "a"},
			},
			text:         testText,
			wantVersions: []string{"aaa"},
			errRegex:     `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: GetVersions is called on it.
			versions, err := tc.urlCommands.GetVersions(tc.text, logx.LogFrom{})

			prefix := fmt.Sprintf(
				"%s\nURLCommands.GetVersions(%v)",
				packageName, tc.text,
			)

			// THEN: the expected versions are returned.
			wantVersions := strings.Join(tc.wantVersions, "__")
			gotVersions := strings.Join(versions, "__")
			if gotVersions != wantVersions {
				t.Errorf(
					"%s result mismatch\ngot:\n%v\nwant:\n%v",
					prefix, tc.wantVersions, gotVersions,
				)
			}

			// AND: the expected error is returned.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
		})
	}
}

func TestURLCommands_Run(t *testing.T) {
	// GIVEN: a URLCommands and text to run it on.
	testText := "abc123-def456"
	tests := []struct {
		name        string
		urlCommands *URLCommands
		text        string
		want        []string
		errRegex    string
	}{
		{
			name:        "nil slice",
			urlCommands: nil,
			errRegex:    `^$`,
			want:        nil,
		},
		{
			name: "regex/standard",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.Ptr(1)},
			},
			errRegex: `^$`,
			want:     []string{"def"},
		},
		{
			name: "regex/negative index",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.Ptr(-1)},
			},
			errRegex: `^$`,
			want:     []string{"def"},
		},
		{
			name: "regex/no match/gives text that didn't match",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([h-z]+)[0-9]+`, Index: test.Ptr(1)},
			},
			errRegex: `regex .* didn't return any matches on "` + testText + `"`,
			want:     nil,
		},
		{
			name: "regex/no match/doesn't give text that didn't match as too long",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([h-z]+)[0-9]+`, Index: test.Ptr(1)},
			},
			errRegex: `regex .* didn't return any matches on "[^"]+"$`,
			text:     strings.Repeat("a123", 5),
			want:     nil,
		},
		{
			name: "regex/index out of bounds",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.Ptr(2)},
			},
			errRegex: `regex .* returned \d elements on "[^']+", but the index wants element number \d`,
			want:     nil,
		},
		{
			name: "regex/with template",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)([0-9]+)`, Index: test.Ptr(1), Template: "$1_$2"},
			},
			errRegex: `^$`,
			want:     []string{"def_456"},
		},
		{
			name: "regex/multiple matches/no template",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`},
			},
			errRegex: `^$`,
			want:     []string{"abc", "def"},
		},
		{
			name: "regex/multiple matches/with template",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)([0-9])`, Template: "$1_$2"},
			},
			errRegex: `^$`,
			want:     []string{"abc_1", "def_4"},
		},
		{
			name: "replace",
			urlCommands: &URLCommands{
				{Type: "replace", Old: "-", New: " "},
			},
			errRegex: `^$`,
			want:     []string{"abc123 def456"},
		},
		{
			name: "split/standard",
			urlCommands: &URLCommands{
				{Type: "split", Text: "-", Index: test.Ptr(-1)},
			},
			errRegex: `^$`,
			want:     []string{"def456"},
		},
		{
			name: "split/negative index",
			urlCommands: &URLCommands{
				{Type: "split", Text: "-", Index: test.Ptr(0)},
			},
			errRegex: `^$`,
			want:     []string{"abc123"},
		},
		{
			name: "split/unknown text",
			urlCommands: &URLCommands{
				{Type: "split", Text: "7", Index: test.Ptr(0)},
			},
			errRegex: `split didn't find any .* to split on`,
			want:     nil,
		},
		{
			name: "split/index out of bounds",
			urlCommands: &URLCommands{
				{Type: "split", Text: "-", Index: test.Ptr(2)},
			},
			errRegex: `split .* returned \d elements on "[^']+", but the index wants element number \d`,
			want:     nil,
		},
		{
			name: "split/no index",
			text: "a-b-c-d",
			urlCommands: &URLCommands{
				{Type: "split", Text: "-"},
			},
			errRegex: `^$`,
			want:     []string{"a", "b", "c", "d"},
		},
		{
			name: "all types",
			urlCommands: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.Ptr(1)},
				{Type: "replace", Old: "e", New: "a"},
				{Type: "split", Text: "a", Index: test.Ptr(1)},
			},
			errRegex: `^$`,
			want:     []string{"f"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			text := testText
			if tc.text != "" {
				text = tc.text
			}

			// WHEN: run is called on it.
			versions, err := tc.urlCommands.Run(text, logx.LogFrom{})

			prefix := fmt.Sprintf(
				"%s\nURLCommands.Run(%q)",
				packageName, tc.name,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}

			// AND: the expected text was returned.
			if !util.AreSlicesEqual(versions, tc.want) {
				t.Errorf(
					"%s versions mismatch\ngot:  %q\nwant: %q",
					prefix, text, tc.want,
				)
			}
		})
	}
}

func TestURLCommand_Regex(t *testing.T) {
	type args struct {
		versionIndex int
		versions     []string
	}
	type wants struct {
		versions *[]string
		errRegex string
	}
	// GIVEN: a URLCommand for regex.
	tests := []struct {
		name    string
		command URLCommand
		args    args
		want    wants
	}{
		{
			name: "no matches",
			command: URLCommand{
				Type:  "regex",
				Regex: `([h-z]+)[0-9]+`,
				Index: test.Ptr(1),
			},
			args: args{
				versions: []string{
					"abc123-def456",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{},
				errRegex: `^regex "[^"]+" didn't return any matches`,
			},
		},
		{
			name: "matches with index",
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
				Index: test.Ptr(1),
			},
			args: args{
				versions: []string{
					"abc123-def456",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{
					"def",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "matches with negative index",
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
				Index: test.Ptr(-1),
			},
			args: args{
				versions: []string{
					"abc123-def456",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{
					"def",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "index out of range",
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
				Index: test.Ptr(2),
			},
			args: args{
				versions: []string{
					"abc123-def456",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{},
				errRegex: `^regex .* returned \d elements on .* but the index wants element number \d`,
			},
		},
		{
			name: "matches with template",
			command: URLCommand{
				Type:     "regex",
				Regex:    `([a-z]+)([0-9]+)`,
				Index:    test.Ptr(1),
				Template: "$1_$2",
			},
			args: args{
				versions: []string{
					"abc123-def456",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{
					"def_456",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "multiple matches without index",
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
			},
			args: args{
				versions: []string{
					"abc123-def456",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{
					"abc",
					"def",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "multiple matches with template",
			command: URLCommand{
				Type:     "regex",
				Regex:    `([a-z]+)([0-9]+)`,
				Template: "$1_$2",
			},
			args: args{
				versions: []string{
					"abc123-def456",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{
					"abc_123",
					"def_456",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "insert at beginning",
			command: URLCommand{
				Type:  "regex",
				Regex: `[a-z]`,
			},
			args: args{
				versions: []string{
					"abc123",
					"def456",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{
					"a",
					"b",
					"c",
					"def456",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "insert at middle",
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
			},
			args: args{
				versions: []string{
					"abc123",
					"def456",
					"ghi789",
				},
				versionIndex: 1,
			},
			want: wants{
				versions: &[]string{
					"abc123",
					"def",
					"ghi789",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "insert at end",
			command: URLCommand{
				Type:  "regex",
				Regex: `[a-z]+([0-9]+)`,
			},
			args: args{
				versions: []string{
					"abc123",
					"def456",
					"ghi789",
				},
				versionIndex: 2,
			},
			want: wants{
				versions: &[]string{
					"abc123",
					"def456",
					"789",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "insert at specific position",
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
			},
			args: args{
				versions: []string{
					"abc123",
					"def456",
					"ghi789",
					"jkl012",
				},
				versionIndex: 1,
			},
			want: wants{
				versions: &[]string{
					"abc123",
					"def",
					"ghi789",
					"jkl012",
				},
				errRegex: `^$`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.want.versions == nil {
				argsVersions := util.CopySlice(tc.args.versions)
				tc.want.versions = &argsVersions
			}

			// WHEN: regex is called on it for the version at the given index.
			versions, err := tc.command.regex(
				tc.args.versionIndex,
				tc.args.versions,
				logx.LogFrom{},
			)

			prefix := fmt.Sprintf(
				"%s\nURLCommands.regex(regex=%q, index=%d, versions=%q) ",
				packageName, tc.command.Regex, tc.args.versionIndex, tc.want.versions,
			)

			// THEN: the expected versions are returned.
			if !util.AreSlicesEqual(versions, *tc.want.versions) {
				t.Errorf(
					"%s result mismatch\ngot:  %v\nwant: %v",
					prefix, versions, *tc.want.versions,
				)
			}

			// AND: the expected error is returned.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, tc.want.errRegex, e,
				)
			}
		})
	}
}

func TestURLCommand_Split(t *testing.T) {
	type args struct {
		versionIndex int
		versions     []string
	}
	type wants struct {
		versions *[]string
		errRegex string
	}
	// GIVEN: a URLCommand for split.
	tests := []struct {
		name    string
		command URLCommand
		args    args
		want    wants
	}{
		{
			name: "split with index",
			command: URLCommand{
				Type:  "split",
				Text:  "-",
				Index: test.Ptr(1),
			},
			args: args{
				versions: []string{
					"abc-def-ghi",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{
					"def",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "split with negative index",
			command: URLCommand{
				Type:  "split",
				Text:  "-",
				Index: test.Ptr(-1),
			},
			args: args{
				versions: []string{
					"abc-def-ghi",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{
					"ghi",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "split with no index",
			command: URLCommand{
				Type: "split",
				Text: "-",
			},
			args: args{
				versions: []string{
					"abc-def-ghi",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{
					"abc", "def", "ghi",
				},
				errRegex: `^$`,
			},
		},
		{
			name: "split index out of bounds",
			command: URLCommand{
				Type:  "split",
				Text:  "-",
				Index: test.Ptr(3),
			},
			args: args{
				versions: []string{
					"abc-def-ghi",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{},
				errRegex: `split .* returned \d elements on "[^']+", but the index wants element number \d`,
			},
		},
		{
			name: "split on unknown text",
			command: URLCommand{
				Type:  "split",
				Text:  "_",
				Index: test.Ptr(0),
			},
			args: args{
				versions: []string{
					"abc-def-ghi",
				},
				versionIndex: 0,
			},
			want: wants{
				versions: &[]string{},
				errRegex: `split didn't find any "_" to split on`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.want.versions == nil {
				argsVersions := util.CopySlice(tc.args.versions)
				tc.want.versions = &argsVersions
			}

			// WHEN: split is called on it for the version at the given index.
			versions, err := tc.command.split(
				tc.args.versionIndex,
				tc.args.versions,
				logx.LogFrom{},
			)

			prefix := fmt.Sprintf(
				"%s\nURLCommands.split(text=%q, index=%d, versions=%q)",
				packageName, tc.command.Text, tc.args.versionIndex, tc.want.versions,
			)

			// THEN: the expected versions are returned.
			if !util.AreSlicesEqual(versions, *tc.want.versions) {
				t.Errorf(
					"%s result mismatch\ngot:  %v\nwant: %v",
					prefix, versions, *tc.want.versions,
				)
			}

			// AND: the expected error is returned.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Fatalf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.want.errRegex,
				)
			}
		})
	}
}
