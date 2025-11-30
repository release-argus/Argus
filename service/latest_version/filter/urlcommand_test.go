// Copyright [2025] [Argus]
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
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

func TestURLCommandSlice_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected URLCommands
		errRegex string
	}{
		"quoted JSON string": {
			input: `"[{\"type\":\"regex\",\"regex\":\"foo\",\"index\":1}]"`,
			expected: URLCommands{
				{Type: "regex", Regex: "foo", Index: test.IntPtr(1)},
			},
			errRegex: `^$`,
		},
		"invalid JSON - quoted JSON string": {
			input:    `"`,
			expected: nil,
			errRegex: `^unexpected end of JSON input$`,
		},
		"JSON - list": {
			input: test.TrimJSON(`[
				{"type":"regex", "regex":"foo", "index":1},
				{"type":"replace", "old":"bar", "new":"baz"}
			]`),
			expected: URLCommands{
				{Type: "regex", Regex: "foo", Index: test.IntPtr(1)},
				{Type: "replace", Old: "bar", New: test.StringPtr("baz")},
			},
			errRegex: `^$`,
		},
		"single URLCommand": {
			input: `{"type":"split","text":"-"}`,
			expected: URLCommands{
				{Type: "split", Text: "-"},
			},
			errRegex: `^$`,
		},
		"invalid JSON": {
			input:    `{"type":"regex","regex":"foo","index":}`,
			expected: nil,
			errRegex: `invalid character.*`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var slice URLCommands

			// WHEN UnmarshalJSON is called.
			err := slice.UnmarshalJSON([]byte(tc.input))

			// THEN the expected result is returned.
			if !util.RegexCheck(tc.errRegex, util.ErrorToString(err)) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, util.ErrorToString(err))
			}

			if len(slice) != len(tc.expected) {
				t.Fatalf("%s\nslice length mismatch\nwant: %d\ngot:  %d\n%#v",
					packageName,
					len(tc.expected), len(slice),
					slice)
			}

			for i := range tc.expected {
				if slice[i].Type != tc.expected[i].Type {
					t.Errorf("%s\nmismatch on Type\nwant: %q\ngot:  %q\n",
						packageName, tc.expected[i].Type, slice[i].Type)
				}
				if slice[i].Regex != tc.expected[i].Regex {
					t.Errorf("%s\nmismatch on Regex\nwant: %q\ngot:  %q\n",
						packageName, tc.expected[i].Regex, slice[i].Regex)
				}
				gotIndex := strings.ReplaceAll(fmt.Sprint(util.DereferenceOrValue(slice[i].Index, 999)),
					"999", "nil")
				wantIndex := strings.ReplaceAll(fmt.Sprint(util.DereferenceOrValue(tc.expected[i].Index, 999)),
					"999", "nil")
				if gotIndex != wantIndex {
					t.Errorf("%s\nmismatch on Index\nwant: %q\ngot:  %q\n",
						packageName, wantIndex, gotIndex)
				}
				if slice[i].Text != tc.expected[i].Text {
					t.Errorf("%s\nmismatch on Text\nwant: %q\ngot:  %q\n",
						packageName, tc.expected[i].Text, slice[i].Text)
				}
				if slice[i].Old != tc.expected[i].Old {
					t.Errorf("%s\nmismatch on Old\nwant: %q\ngot:  %q\n",
						packageName, tc.expected[i].Old, slice[i].Old)
				}
				gotNew := util.DereferenceOrDefault(slice[i].New)
				wantNew := util.DereferenceOrDefault(tc.expected[i].New)
				if gotNew != wantNew {
					t.Errorf("%s\nmismatch on New\nwant: %q\ngot:  %q\n",
						packageName, wantNew, gotNew)
				}
			}
		})
	}
}

func TestURLCommandSlice_UnmarshalYAML(t *testing.T) {
	// GIVEN a file to read a URLCommands.
	tests := map[string]struct {
		input    string
		slice    URLCommands
		errRegex string
	}{
		"invalid unmarshal": {
			input: test.TrimYAML(`
				type: regex
				regex: foo
				regex: foo
				index: 1
				text: hi
				old: was
				new: now
			`),
			errRegex: `mapping key .* already defined`,
		},
		"non-list URLCommand": {
			input: test.TrimYAML(`
				type: regex
				regex: foo
				index: 1
				text: hi
				old: was
				new: now
			`),
			slice: URLCommands{
				{Type: "regex",
					Regex: `foo`, Index: test.IntPtr(1),
					Text: "hi", Old: "was", New: test.StringPtr("now")}},
			errRegex: `^$`,
		},
		"list of URLCommands": {
			input: test.TrimYAML(`
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
			slice: URLCommands{
				{Type: "regex",
					Regex: `\"([0-9.+])\"`, Index: test.IntPtr(1)},
				{Type: "replace", Old: "foo", New: test.StringPtr("bar")},
				{Type: "split", Text: "abc", Index: test.IntPtr(2)}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			var slice URLCommands

			// WHEN Unmarshalled.
			err := yaml.Unmarshal([]byte(tc.input), &slice)

			// THEN the it errors when appropriate and unmarshals correctly into a list.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
			if len(slice) != len(tc.slice) {
				t.Fatalf("%s\nslice length mismatch\nwant: %d\ngot:  %d\n%#v",
					packageName,
					len(tc.slice), len(slice),
					slice)
			}
			for i := range tc.slice {
				if slice[i].Type != tc.slice[i].Type {
					t.Errorf("%s\nmismatch on Type\nwant: %q\ngot:  %q\n",
						packageName, tc.slice[i].Type, slice[i].Type)
				}
				if slice[i].Regex != tc.slice[i].Regex {
					t.Errorf("%s\nmismatch on Regex\nwant: %q\ngot:  %q\n",
						packageName, tc.slice[i].Regex, slice[i].Regex)
				}
				gotIndex := strings.ReplaceAll(fmt.Sprint(util.DereferenceOrValue(slice[i].Index, 999)),
					"999", "nil")
				wantIndex := strings.ReplaceAll(fmt.Sprint(util.DereferenceOrValue(tc.slice[i].Index, 999)),
					"999", "nil")
				if gotIndex != wantIndex {
					t.Errorf("%s\nmismatch on Index\nwant: %q\ngot:  %q\n",
						packageName, wantIndex, gotIndex)
				}
				if slice[i].Text != tc.slice[i].Text {
					t.Errorf("%s\nmismatch on Text\nwant: %q\ngot:  %q\n",
						packageName, tc.slice[i].Text, slice[i].Text)
				}
				if slice[i].Old != tc.slice[i].Old {
					t.Errorf("%s\nmismatch on Old\nwant: %q\ngot:  %q\n",
						packageName, tc.slice[i].Old, slice[i].Old)
				}
				gotNew := util.DereferenceOrDefault(slice[i].New)
				wantNew := util.DereferenceOrDefault(tc.slice[i].New)
				if gotNew != wantNew {
					t.Errorf("%s\nmismatch on New\nwant: %q\ngot:  %q\n",
						packageName, wantNew, gotNew)
				}
			}
		})
	}
}

func TestURLCommandSlice_String(t *testing.T) {
	// GIVEN a URLCommands.
	tests := map[string]struct {
		slice *URLCommands
		want  string
	}{
		"regex": {
			slice: &URLCommands{
				testURLCommandRegex()},
			want: test.TrimYAML(`
				- type: regex
					regex: -([0-9.]+)-
					index: 0
			`),
		},
		"regex (templated)": {
			slice: &URLCommands{
				testURLCommandRegexTemplate()},
			want: test.TrimYAML(`
				- type: regex
					regex: -([0-9.]+)-
					index: 0
					template: _$1_
			`),
		},
		"replace": {
			slice: &URLCommands{
				testURLCommandReplace()},
			want: test.TrimYAML(`
				- type: replace
					new: bar
					old: foo
			`),
		},
		"split": {
			slice: &URLCommands{
				testURLCommandSplit()},
			want: test.TrimYAML(`
				- type: split
					index: 1
					text: this
			`),
		},
		"all types": {
			slice: &URLCommands{
				testURLCommandRegex(),
				testURLCommandReplace(),
				testURLCommandSplit()},
			want: test.TrimYAML(`
				- type: regex
					regex: -([0-9.]+)-
					index: 0
				- type: replace
					new: bar
					old: foo
				- type: split
					index: 1
					text: this
			`),
		},
		"nil slice": {
			slice: nil,
			want:  "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// WHEN String is called on it.
			got := tc.slice.String()

			// THEN the expected string is returned.
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Fatalf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestURLCommand_String(t *testing.T) {
	// GIVEN a URLCommand.
	regex := testURLCommandRegex()
	replace := testURLCommandReplace()
	split := testURLCommandSplit()
	tests := map[string]struct {
		cmd  *URLCommand
		want string
	}{
		"regex": {
			cmd: &regex,
			want: test.TrimYAML(`
				type: regex
				regex: -([0-9.]+)-
				index: 0
			`),
		},
		"replace": {
			cmd: &replace,
			want: test.TrimYAML(`
				type: replace
				new: bar
				old: foo
			`),
		},
		"split": {
			cmd: &split,
			want: test.TrimYAML(`
				type: split
				index: 1
				text: this
			`),
		},
		"nil slice": {
			cmd:  nil,
			want: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN String is called on it.
			got := tc.cmd.String()

			// THEN the expected string is returned.
			tc.want = strings.TrimPrefix(tc.want, "\n")
			if got != tc.want {
				t.Fatalf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, got)
			}
		})
	}
}

func TestURLCommandSlice_GetVersions(t *testing.T) {
	// GIVEN a URLCommands.
	testText := "abc123-def456"
	tests := map[string]struct {
		slice        *URLCommands
		text         string
		wantVersions []string
		errRegex     string
	}{
		"empty slice": {
			slice:        &URLCommands{},
			text:         testText,
			wantVersions: []string{testText},
			errRegex:     `^$`,
		},
		"empty slice+text": {
			slice:        &URLCommands{},
			text:         "",
			wantVersions: nil,
			errRegex:     `^$`,
		},
		"single version - regex": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.IntPtr(1)}},
			text:         testText,
			wantVersions: []string{"def"},
			errRegex:     `^$`,
		},
		"single version - replace": {
			slice: &URLCommands{
				{Type: "replace", Old: "-", New: test.StringPtr(" ")}},
			text:         testText,
			wantVersions: []string{"abc123 def456"},
			errRegex:     `^$`,
		},
		"multiple versions - split": {
			slice: &URLCommands{
				{Type: "split", Text: "-"}},
			text:         testText,
			wantVersions: []string{"abc123", "def456"},
			errRegex:     `^$`,
		},
		"multiple versions - split fail": {
			slice: &URLCommands{
				{Type: "split", Text: "_"}},
			text:         testText,
			wantVersions: nil,
			errRegex:     `^split didn't find any "_" to split on$`,
		},
		"multiple versions - regex and split": {
			slice: &URLCommands{
				{Type: "split", Text: "-", Index: nil},
				{Type: "regex", Regex: `([a-z]+)[0-9]+`}},
			text:         testText,
			wantVersions: []string{"abc", "def"},
			errRegex:     `^$`,
		},
		"multiple versions - regex fail": {
			slice: &URLCommands{
				{Type: "split", Text: "-", Index: nil},
				{Type: "split", Text: "_", Index: test.IntPtr(0)}},
			text:         testText,
			wantVersions: nil,
			errRegex:     `^split didn't find any "_" to split on$`,
		},
		"regex doesn't match": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([h-z]+)[0-9]+`, Index: test.IntPtr(1)}},
			text:         testText,
			wantVersions: nil,
			errRegex:     `regex .* didn't return any matches on "` + testText + `"`,
		},
		"split index out of bounds": {
			slice: &URLCommands{
				{Type: "split", Text: "-", Index: test.IntPtr(2)}},
			text:         testText,
			wantVersions: nil,
			errRegex:     `split .* returned \d elements on "[^']+", but the index wants element number \d`,
		},
		"all types": {
			slice: &URLCommands{
				{Type: "split", Text: "-", Index: test.IntPtr(0)},
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.IntPtr(0)},
				{Type: "replace", Old: "b", New: test.StringPtr("a")},
				{Type: "replace", Old: "c", New: test.StringPtr("a")}},
			text:         testText,
			wantVersions: []string{"aaa"},
			errRegex:     `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN GetVersions is called on it.
			versions, err := tc.slice.GetVersions(tc.text, logutil.LogFrom{})

			// THEN the expected versions are returned.
			wantVersions := strings.Join(tc.wantVersions, "__")
			gotVersions := strings.Join(versions, "__")
			if gotVersions != wantVersions {
				t.Errorf("%s\nwant:\n%v\ngot:\n%v",
					packageName, tc.wantVersions, versions)
			}
			// AND the expected error is returned.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot: %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestURLCommandSlice_Run(t *testing.T) {
	// GIVEN a URLCommands and text to run it on.
	testText := "abc123-def456"
	tests := map[string]struct {
		slice    *URLCommands
		text     string
		want     []string
		errRegex string
	}{
		"nil slice": {
			slice:    nil,
			errRegex: `^$`,
			want:     nil,
		},
		"regex": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.IntPtr(1)}},
			errRegex: `^$`,
			want:     []string{"def"},
		},
		"regex with negative index": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.IntPtr(-1)}},
			errRegex: `^$`,
			want:     []string{"def"},
		},
		"regex doesn't match (gives text that didn't match)": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([h-z]+)[0-9]+`, Index: test.IntPtr(1)}},
			errRegex: `regex .* didn't return any matches on "` + testText + `"`,
			want:     nil,
		},
		"regex doesn't match (doesn't give text that didn't match as too long)": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([h-z]+)[0-9]+`, Index: test.IntPtr(1)}},
			errRegex: `regex .* didn't return any matches on "[^"]+"$`,
			text:     strings.Repeat("a123", 5),
			want:     nil,
		},
		"regex index out of bounds": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.IntPtr(2)}},
			errRegex: `regex .* returned \d elements on "[^']+", but the index wants element number \d`,
			want:     nil,
		},
		"regex with template": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)([0-9]+)`, Index: test.IntPtr(1), Template: "$1_$2"}},
			errRegex: `^$`,
			want:     []string{"def_456"},
		},
		"regex multiple matches": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`}},
			errRegex: `^$`,
			want:     []string{"abc", "def"},
		},
		"regex multiple matches - with template": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)([0-9])`, Template: "$1_$2"}},
			errRegex: `^$`,
			want:     []string{"abc_1", "def_4"},
		},
		"replace": {
			slice: &URLCommands{
				{Type: "replace", Old: "-", New: test.StringPtr(" ")}},
			errRegex: `^$`,
			want:     []string{"abc123 def456"},
		},
		"split": {
			slice: &URLCommands{
				{Type: "split", Text: "-", Index: test.IntPtr(-1)}},
			errRegex: `^$`,
			want:     []string{"def456"},
		},
		"split with negative index": {
			slice: &URLCommands{
				{Type: "split", Text: "-", Index: test.IntPtr(0)}},
			errRegex: `^$`,
			want:     []string{"abc123"},
		},
		"split on unknown text": {
			slice: &URLCommands{
				{Type: "split", Text: "7", Index: test.IntPtr(0)}},
			errRegex: `split didn't find any .* to split on`,
			want:     nil,
		},
		"split index out of bounds": {
			slice: &URLCommands{
				{Type: "split", Text: "-", Index: test.IntPtr(2)}},
			errRegex: `split .* returned \d elements on "[^']+", but the index wants element number \d`,
			want:     nil,
		},
		"split with no index": {
			text: "a-b-c-d",
			slice: &URLCommands{
				{Type: "split", Text: "-"}},
			errRegex: `^$`,
			want:     []string{"a", "b", "c", "d"},
		},
		"all types": {
			slice: &URLCommands{
				{Type: "regex", Regex: `([a-z]+)[0-9]+`, Index: test.IntPtr(1)},
				{Type: "replace", Old: "e", New: test.StringPtr("a")},
				{Type: "split", Text: "a", Index: test.IntPtr(1)}},
			errRegex: `^$`,
			want:     []string{"f"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			text := testText
			if tc.text != "" {
				text = tc.text
			}

			// WHEN run is called on it.
			versions, err := tc.slice.Run(text, logutil.LogFrom{})

			// THEN the expected text was returned.
			if !reflect.DeepEqual(tc.want, versions) {
				t.Errorf("%s\nwant: %q\ngot:  %q",
					packageName, tc.want, text)
			}
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
			}
		})
	}
}

func TestURLCommand_regex(t *testing.T) {
	type args struct {
		versionIndex int
		versions     []string
	}
	type wants struct {
		versions *[]string
		errRegex string
	}
	// GIVEN a URLCommand for regex.
	tests := map[string]struct {
		command URLCommand
		args    args
		want    wants
	}{
		"no matches": {
			command: URLCommand{
				Type:  "regex",
				Regex: `([h-z]+)[0-9]+`,
				Index: test.IntPtr(1),
			},
			args: args{
				versions: []string{
					"abc123-def456"},
				versionIndex: 0},
			want: wants{
				versions: nil,
				errRegex: `^regex "[^"]+" didn't return any matches`},
		},
		"matches with index": {
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
				Index: test.IntPtr(1),
			},
			args: args{
				versions: []string{
					"abc123-def456"},
				versionIndex: 0},
			want: wants{
				versions: &[]string{
					"def"},
				errRegex: `^$`},
		},
		"matches with negative index": {
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
				Index: test.IntPtr(-1),
			},
			args: args{
				versions: []string{
					"abc123-def456"},
				versionIndex: 0},
			want: wants{
				versions: &[]string{
					"def"},
				errRegex: `^$`},
		},
		"index out of range": {
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
				Index: test.IntPtr(2),
			},
			args: args{
				versions: []string{
					"abc123-def456"},
				versionIndex: 0},
			want: wants{
				versions: nil,
				errRegex: `^regex .* returned \d elements on .* but the index wants element number \d`},
		},
		"matches with template": {
			command: URLCommand{
				Type:     "regex",
				Regex:    `([a-z]+)([0-9]+)`,
				Index:    test.IntPtr(1),
				Template: "$1_$2",
			},
			args: args{
				versions: []string{
					"abc123-def456"},
				versionIndex: 0},
			want: wants{
				versions: &[]string{
					"def_456"},
				errRegex: `^$`},
		},
		"multiple matches without index": {
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
			},
			args: args{
				versions: []string{
					"abc123-def456"},
				versionIndex: 0},
			want: wants{
				versions: &[]string{
					"abc",
					"def"},
				errRegex: `^$`},
		},
		"multiple matches with template": {
			command: URLCommand{
				Type:     "regex",
				Regex:    `([a-z]+)([0-9]+)`,
				Template: "$1_$2",
			},
			args: args{
				versions: []string{
					"abc123-def456"},
				versionIndex: 0},
			want: wants{
				versions: &[]string{
					"abc_123",
					"def_456"},
				errRegex: `^$`},
		},
		"insert at beginning": {
			command: URLCommand{
				Type:  "regex",
				Regex: `[a-z]`,
			},
			args: args{
				versions: []string{
					"abc123",
					"def456"},
				versionIndex: 0},
			want: wants{
				versions: &[]string{
					"a",
					"b",
					"c",
					"def456"},
				errRegex: `^$`},
		},
		"insert at middle": {
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
			},
			args: args{
				versions: []string{
					"abc123",
					"def456",
					"ghi789"},
				versionIndex: 1},
			want: wants{
				versions: &[]string{
					"abc123",
					"def",
					"ghi789"},
				errRegex: `^$`},
		},
		"insert at end": {
			command: URLCommand{
				Type:  "regex",
				Regex: `[a-z]+([0-9]+)`,
			},
			args: args{
				versions: []string{
					"abc123",
					"def456",
					"ghi789"},
				versionIndex: 2},
			want: wants{
				versions: &[]string{
					"abc123",
					"def456",
					"789"},
				errRegex: `^$`},
		},
		"insert at specific position": {
			command: URLCommand{
				Type:  "regex",
				Regex: `([a-z]+)[0-9]+`,
			},
			args: args{
				versions: []string{
					"abc123",
					"def456",
					"ghi789",
					"jkl012"},
				versionIndex: 1},
			want: wants{
				versions: &[]string{
					"abc123",
					"def",
					"ghi789",
					"jkl012"},
				errRegex: `^$`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.want.versions == nil {
				argsVersions := util.CopyList(tc.args.versions)
				tc.want.versions = &argsVersions
			}

			// WHEN regex is called on it for the version at the given index.
			err := tc.command.regex(tc.args.versionIndex, &tc.args.versions, logutil.LogFrom{})

			// THEN the expected versions are returned.
			if !reflect.DeepEqual(*tc.want.versions, tc.args.versions) {
				t.Errorf("%s\nwant: %v\ngot:  %v",
					packageName, *tc.want.versions, tc.args.versions)
			}
			// AND the expected error is returned.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want.errRegex, e)
			}
		})
	}
}

func TestURLCommand_split(t *testing.T) {
	type args struct {
		versionIndex int
		versions     []string
	}
	type wants struct {
		versions *[]string
		errRegex string
	}
	// GIVEN a URLCommand for split.
	tests := map[string]struct {
		command URLCommand
		args    args
		want    wants
	}{
		"split with index": {
			command: URLCommand{
				Type:  "split",
				Text:  "-",
				Index: test.IntPtr(1),
			},
			args: args{
				versions: []string{
					"abc-def-ghi"},
				versionIndex: 0},
			want: wants{
				versions: &[]string{
					"def"},
				errRegex: `^$`},
		},
		"split with negative index": {
			command: URLCommand{
				Type:  "split",
				Text:  "-",
				Index: test.IntPtr(-1),
			},
			args: args{
				versions: []string{
					"abc-def-ghi"},
				versionIndex: 0},
			want: wants{
				versions: &[]string{
					"ghi"},
				errRegex: `^$`},
		},
		"split with no index": {
			command: URLCommand{
				Type: "split",
				Text: "-",
			},
			args: args{
				versions: []string{
					"abc-def-ghi"},
				versionIndex: 0},
			want: wants{
				versions: &[]string{
					"abc", "def", "ghi"},
				errRegex: `^$`},
		},
		"split index out of bounds": {
			command: URLCommand{
				Type:  "split",
				Text:  "-",
				Index: test.IntPtr(3),
			},
			args: args{
				versions: []string{
					"abc-def-ghi"},
				versionIndex: 0},
			want: wants{
				versions: nil,
				errRegex: `split .* returned \d elements on "[^']+", but the index wants element number \d`},
		},
		"split on unknown text": {
			command: URLCommand{
				Type:  "split",
				Text:  "_",
				Index: test.IntPtr(0),
			},
			args: args{
				versions: []string{
					"abc-def-ghi"},
				versionIndex: 0},
			want: wants{
				versions: nil,
				errRegex: `split didn't find any "_" to split on`},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.want.versions == nil {
				argsVersions := util.CopyList(tc.args.versions)
				tc.want.versions = &argsVersions
			}

			// WHEN split is called on it for the version at the given index.
			err := tc.command.split(tc.args.versionIndex, &tc.args.versions, logutil.LogFrom{})

			// THEN the expected versions are returned.
			if !reflect.DeepEqual(*tc.want.versions, tc.args.versions) {
				t.Errorf("%s\nwant: %v\ngot:  %v",
					packageName, *tc.want.versions, tc.args.versions)
			}
			// AND the expected error is returned.
			e := util.ErrorToString(err)
			if !util.RegexCheck(tc.want.errRegex, e) {
				t.Fatalf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.want.errRegex, e)
			}
		})
	}
}

func TestURLCommandSlice_CheckValues(t *testing.T) {
	// GIVEN a URLCommands.
	tests := map[string]struct {
		slice     *URLCommands
		wantSlice *URLCommands
		errRegex  string
	}{
		"nil slice": {
			slice:    nil,
			errRegex: `^$`,
		},
		"valid regex": {
			slice:    &URLCommands{testURLCommandRegex()},
			errRegex: `^$`,
		},
		"undefined regex": {
			slice: &URLCommands{
				{Type: "regex"}},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: regex
					regex: <required>.*$`),
		},
		"invalid regex": {
			slice: &URLCommands{
				{Type: "regex", Regex: `[0-`}},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: regex
					regex: .* <invalid>.*$`),
		},
		"valid regex with template": {
			slice:    &URLCommands{testURLCommandRegexTemplate()},
			errRegex: `^$`,
		},
		"valid regex with empty template": {
			slice: &URLCommands{
				{Type: "regex", Regex: `[0-]`, Template: ""}},
			wantSlice: &URLCommands{
				{Type: "regex", Regex: `[0-]`}},
			errRegex: `^$`,
		},
		"valid replace": {
			slice: &URLCommands{
				testURLCommandReplace()},
			errRegex: `^$`,
		},
		"invalid replace": {
			slice: &URLCommands{
				{Type: "replace"}},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: replace
					old: <required>.*$`),
		},
		"valid split": {
			slice: &URLCommands{
				testURLCommandSplit()},
			errRegex: `^$`,
		},
		"invalid split": {
			slice: &URLCommands{
				{Type: "split"}},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: split
					text: <required>`),
		},
		"invalid type": {
			slice: &URLCommands{
				{Type: "something"}},
			errRegex: test.TrimYAML(`
				^- item_0:
					type: .* <invalid>.*$`),
		},
		"valid all types": {
			slice: &URLCommands{
				testURLCommandRegex(),
				testURLCommandReplace(),
				testURLCommandSplit()},
			errRegex: `^$`,
		},
		"all possible errors": {
			slice: &URLCommands{
				{Type: "regex"}, {Type: "replace"},
				{Type: "split"},
				{Type: "something"}},
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
					type: "something" <invalid>.*$`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// WHEN CheckValues is called on it.
			err := tc.slice.CheckValues("")

			// THEN err is expected.
			e := util.ErrorToString(err)
			lines := strings.Split(e, "\n")
			wantLines := strings.Count(tc.errRegex, "\n")
			if wantLines > len(lines) {
				t.Fatalf("%s\nwant: %d lines of error:\n%q\ngot:  %d lines:\n%v\n\nstdout: %q",
					packageName, wantLines, tc.errRegex, len(lines), lines, e)
				return
			}
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf("%s\nerror mismatch\nwant: %q\ngot:  %q",
					packageName, tc.errRegex, e)
				return
			}

			// AND the slice is as expected.
			if tc.wantSlice != nil {
				strHave := tc.slice.String()
				strWant := tc.wantSlice.String()
				if strHave != strWant {
					t.Errorf("%s\nwant slice:\n%q\ngot  slice: %q",
						packageName, strWant, strHave)
				}
			}
		})
	}
}
