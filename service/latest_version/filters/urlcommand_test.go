// Copyright [2022] [Argus]
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

package filters

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/release-argus/Argus/utils"
	"gopkg.in/yaml.v3"
)

func TestURLCommandSliceInit(t *testing.T) {
	// GIVEN URLCommandSlice and a JLog
	var slice URLCommandSlice
	newJLog := utils.NewJLog("WARN", false)

	// WHEN Init is called with it
	slice.Init(newJLog)

	// THEN the global JLog is set to its address
	if jLog != newJLog {
		t.Fatalf("JLog should have been initialised to the one we called Init with")
	}
}

func TestURLCommandSlicePrint(t *testing.T) {
	// GIVEN a URLCommandSlice
	tests := map[string]struct {
		slice *URLCommandSlice
		lines int
	}{
		"regex":     {slice: &URLCommandSlice{testURLCommandRegex()}, lines: 3},
		"replace":   {slice: &URLCommandSlice{testURLCommandReplace()}, lines: 4},
		"split":     {slice: &URLCommandSlice{testURLCommandSplit()}, lines: 4},
		"all types": {slice: &URLCommandSlice{testURLCommandRegex(), testURLCommandReplace(), testURLCommandSplit()}, lines: 9},
		"nil slice": {slice: nil, lines: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// WHEN Print is called on it
			tc.slice.Print("")

			// THEN the expected number of lines are printed
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = stdout
			got := strings.Count(string(out), "\n")
			if got != tc.lines {
				t.Errorf("Print should have given %d lines, but gave %d\n%s",
					tc.lines, got, out)
			}
		})
	}
}

func TestURLCommandSliceRun(t *testing.T) {
	// GIVEN a URLCommandSlice
	testLogging("WARN")
	testText := "abc123-def456"
	tests := map[string]struct {
		slice    *URLCommandSlice
		text     string
		want     string
		errRegex string
	}{
		"nil slice":                 {slice: nil, errRegex: "^$", want: testText},
		"regex":                     {slice: &URLCommandSlice{{Type: "regex", Regex: stringPtr("([a-z]+)[0-9]+"), Index: 1}}, errRegex: "^$", want: "def"},
		"regex with negative index": {slice: &URLCommandSlice{{Type: "regex", Regex: stringPtr("([a-z]+)[0-9]+"), Index: -1}}, errRegex: "^$", want: "def"},
		"regex doesn't match (gives text that didn't match)": {slice: &URLCommandSlice{{Type: "regex", Regex: stringPtr("([h-z]+)[0-9]+"), Index: 1}},
			errRegex: "regex .* didn't return any matches on .*" + testText, want: testText},
		"regex doesn't match (doesn't give text that didn't match)": {slice: &URLCommandSlice{{Type: "regex", Regex: stringPtr("([h-z]+)[0-9]+"), Index: 1}},
			errRegex: "regex .* didn't return any matches$", text: strings.Repeat("a123", 5), want: "a123a123a123a123a123"},
		"regex index out of bounds": {slice: &URLCommandSlice{{Type: "regex", Regex: stringPtr("([a-z]+)[0-9]+"), Index: 2}},
			errRegex: `regex .* returned \d elements but the index wants element number \d`, want: testText},
		"replace":                   {slice: &URLCommandSlice{{Type: "replace", Old: stringPtr("-"), New: stringPtr(" ")}}, errRegex: "^$", want: "abc123 def456"},
		"split":                     {slice: &URLCommandSlice{{Type: "split", Text: stringPtr("-"), Index: -1}}, errRegex: "^$", want: "def456"},
		"split with negative index": {slice: &URLCommandSlice{{Type: "split", Text: stringPtr("-"), Index: 0}}, errRegex: "^$", want: "abc123"},
		"split on unknown text": {slice: &URLCommandSlice{{Type: "split", Text: stringPtr("7"), Index: 0}},
			errRegex: "split didn't find any .* to split on", want: testText},
		"split index out of bounds": {slice: &URLCommandSlice{{Type: "split", Text: stringPtr("-"), Index: 2}},
			errRegex: `split .* returned \d elements but the index wants element number \d`, want: testText},
		"all types": {slice: &URLCommandSlice{{Type: "regex", Regex: stringPtr("([a-z]+)[0-9]+"), Index: 1},
			{Type: "replace", Old: stringPtr("e"), New: stringPtr("a")},
			{Type: "split", Text: stringPtr("a"), Index: 1}}, errRegex: "^$", want: "f"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN run is called on it
			text := testText
			if tc.text != "" {
				text = tc.text
			}
			text, err := tc.slice.Run(text, utils.LogFrom{})

			// THEN the expected text was returned
			if tc.want != text {
				t.Errorf("Should have got %q, not %q",
					tc.want, text)
			}
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
		})
	}
}

func TestURLCommandCheckValues(t *testing.T) {
	// GIVEN a URLCommandSlice
	tests := map[string]struct {
		slice    *URLCommandSlice
		errRegex []string
	}{
		"nil slice":       {slice: nil, errRegex: []string{`^$`}},
		"valid regex":     {slice: &URLCommandSlice{testURLCommandRegex()}, errRegex: []string{`^$`}},
		"undefined regex": {slice: &URLCommandSlice{{Type: "regex"}}, errRegex: []string{`^url_commands:$`, `^  item_0:$`, `^    regex: <required>`}},
		"invalid regex":   {slice: &URLCommandSlice{{Type: "regex", Regex: stringPtr("[0-")}}, errRegex: []string{`^    regex: .* <invalid>`}},
		"valid replace":   {slice: &URLCommandSlice{testURLCommandReplace()}, errRegex: []string{`^$`}},
		"invalid replace": {slice: &URLCommandSlice{{Type: "replace"}}, errRegex: []string{`^    new: <required>`, `^    old: <required>`}},
		"valid split":     {slice: &URLCommandSlice{testURLCommandSplit()}, errRegex: []string{`^$`}},
		"invalid split":   {slice: &URLCommandSlice{{Type: "split"}}, errRegex: []string{`^    text: <required>`}},
		"invalid type":    {slice: &URLCommandSlice{{Type: "something"}}, errRegex: []string{`^    type: .* <invalid>`}},
		"valid all types": {slice: &URLCommandSlice{testURLCommandRegex(), testURLCommandReplace(), testURLCommandSplit()}, errRegex: []string{`^$`}},
		"all possible errors": {slice: &URLCommandSlice{{Type: "regex"}, {Type: "replace"}, {Type: "split"}, {Type: "something"}},
			errRegex: []string{`^url_commands:$`, `^  item_0:$`, `^    regex: <required>`, `^    new: <required>`, `^    old: <required>`, `^    text: <required>`, `^    type: .* <invalid>`}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// WHEN CheckValues is called on it
			err := tc.slice.CheckValues("")

			// THEN err is expected
			e := utils.ErrorToString(err)
			lines := strings.Split(e, `\`)
			for i := range tc.errRegex {
				re := regexp.MustCompile(tc.errRegex[i])
				found := false
				for j := range lines {
					match := re.MatchString(lines[j])
					if match {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("want match for: %q\ngot:  %q",
						tc.errRegex[i], strings.ReplaceAll(e, `\`, "\n"))
				}
			}
		})
	}
}

func TestUnmarshalYAMLSingle(t *testing.T) {
	// GIVEN a file to read a URLCommandSlice
	tests := map[string]struct {
		file     string
		slice    URLCommandSlice
		errRegex string
	}{
		"invalid unmarshal": {file: "../../../test/urlcommandslice_invalid.yml", errRegex: "mapping key .* already defined"},
		"non-list URLCommand": {file: "../../../test/urlcommandslice_single.yml", errRegex: "^$",
			slice: URLCommandSlice{{Type: "regex", Regex: stringPtr("foo"), Index: 1, Text: stringPtr("hi"),
				Old: stringPtr("was"), New: stringPtr("now")}}},
		"list of URLCommands": {file: "../../../test/urlcommandslice_multi.yml", errRegex: "^$",
			slice: URLCommandSlice{{Type: "regex", Regex: stringPtr(`\"([0-9.+])\"`), Index: 1}, {Type: "replace", Old: stringPtr("foo"), New: stringPtr("bar")},
				{Type: "split", Text: stringPtr("abc"), Index: 2}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var slice URLCommandSlice
			data, _ := os.ReadFile(tc.file)

			// WHEN Unmarshalled
			err := yaml.Unmarshal(data, &slice)

			// THEN the it errs when appropriate and unmarshals correctly into a list
			e := utils.ErrorToString(err)
			re := regexp.MustCompile(tc.errRegex)
			match := re.MatchString(e)
			if !match {
				t.Fatalf("want match for %q\nnot: %q",
					tc.errRegex, e)
			}
			if len(slice) != len(tc.slice) {
				t.Fatalf("got a slice of length %d. want %d\n%#v",
					len(slice), len(tc.slice), slice)
			}
			for i := range tc.slice {
				if slice[i].Type != tc.slice[i].Type {
					t.Errorf("wrong Type:\nwant: %q\ngot:  %q\n",
						tc.slice[i].Type, slice[i].Type)
				}
				if utils.DefaultIfNil(slice[i].Regex) != utils.DefaultIfNil(tc.slice[i].Regex) {
					t.Errorf("wrong Regex:\nwant: %q\ngot:  %q\n",
						utils.DefaultIfNil(tc.slice[i].Regex), utils.DefaultIfNil(slice[i].Regex))
				}
				if slice[i].Index != tc.slice[i].Index {
					t.Errorf("wrong Index:\nwant: %q\ngot:  %q\n",
						tc.slice[i].Index, slice[i].Index)
				}
				if utils.DefaultIfNil(slice[i].Text) != utils.DefaultIfNil(tc.slice[i].Text) {
					t.Errorf("wrong Text:\nwant: %q\ngot:  %q\n",
						utils.DefaultIfNil(tc.slice[i].Text), utils.DefaultIfNil(slice[i].Text))
				}
				if utils.DefaultIfNil(slice[i].Old) != utils.DefaultIfNil(tc.slice[i].Old) {
					t.Errorf("wrong Old:\nwant: %q\ngot:  %q\n",
						utils.DefaultIfNil(tc.slice[i].Old), utils.DefaultIfNil(slice[i].Old))
				}
				if utils.DefaultIfNil(slice[i].New) != utils.DefaultIfNil(tc.slice[i].New) {
					t.Errorf("wrong New:\nwant: %q\ngot:  %q\n",
						utils.DefaultIfNil(tc.slice[i].New), utils.DefaultIfNil(slice[i].New))
				}
			}
		})
	}
}
