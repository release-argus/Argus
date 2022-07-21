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
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/release-argus/Argus/utils"
	"gopkg.in/yaml.v3"
)

func TestURLCommandPrintRegex(t *testing.T) {
	// GIVEN a Regex URLCommand
	command := testURLCommandRegex()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	command.Print("")

	// THEN 4 lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 4
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}

func TestURLCommandPrintReplace(t *testing.T) {
	// GIVEN a Replace URLCommand
	command := testURLCommandReplace()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	command.Print("")

	// THEN 3 lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 3
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}

func TestURLCommandPrintSplit(t *testing.T) {
	// GIVEN a Split URLCommand
	command := testURLCommandSplit()
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	command.Print("")

	// THEN 4 lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 4
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}

func TestURLCommandSlicePrintNil(t *testing.T) {
	// GIVEN a nil SLice
	var slice *URLCommandSlice
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	slice.Print("")

	// THEN 1 line is printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 1
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}

func TestURLCommandSlicePrintAllTypes(t *testing.T) {
	// GIVEN a URLCommandSlice containing each URLCommand type
	slice := URLCommandSlice{
		testURLCommandRegex(),
		testURLCommandReplace(),
		testURLCommandSplit(),
	}
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN Print is called on it
	slice.Print("")

	// THEN 11 lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 11
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s",
			want, got, out)
	}
}

func TestURLCommandSetParentIgnoreMissesWithNil(t *testing.T) {
	// GIVEN a nil slice and parentIgnoreMisses
	var slice *URLCommandSlice
	parentIgnoreMisses := false

	// WHEN SetParentIgnoreMisses is called on it
	slice.SetParentIgnoreMisses(&parentIgnoreMisses)

	// THEN nothing crashes
}

func TestURLCommandSetParentIgnoreMisses(t *testing.T) {
	// GIVEN a URLCommandSlice containing each URLCommand type and parentIgnoreMisses
	slice := URLCommandSlice{
		testURLCommandRegex(),
		testURLCommandReplace(),
		testURLCommandSplit(),
	}
	parentIgnoreMisses := false

	// WHEN SetParentIgnoreMisses is called on it
	slice.SetParentIgnoreMisses(&parentIgnoreMisses)

	// THEN all the URLCommandSlice is given parentIgnoreMisses
	for _, command := range slice {
		if command.ParentIgnoreMisses != &parentIgnoreMisses {
			t.Errorf("Command %v was not given ParentIgnoreMisses.\nGot %v, want %v",
				command, command.ParentIgnoreMisses, &parentIgnoreMisses)
		}
	}
}

func TestURLCommandGetIgnoreMisses(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN GetIgnoreMisses is called on it
	got := service.GetIgnoreMisses()

	// THEN IgnoreMisses is returned
	want := service.IgnoreMisses
	if got != want {
		t.Errorf("Got %v, want %v",
			got, want)
	}
}

func TestURLCommandSliceRunWithNil(t *testing.T) {
	// GIVEN a nil URLCommand URLCommandSlice
	var slice *URLCommandSlice

	// WHEN run is called on it
	_, err := slice.run("", utils.LogFrom{})

	// THEN err is nil
	if err != nil {
		t.Errorf("Should have nil, not %q",
			err.Error())
	}
}

func TestURLCommandSliceRunWithSuccess(t *testing.T) {
	// GIVEN a URLCommand URLCommandSlice that will pass
	jLog = utils.NewJLog("WARN", false)
	slice := URLCommandSlice{
		testURLCommandSplit(),
		testURLCommandRegex(),
		testURLCommandReplace(),
	}
	*slice[0].Text = ","
	*slice[1].Regex = "([a-z]+)1"
	slice[1].Index = 0
	*slice[2].Old = "c"
	*slice[2].New = "a"
	want := "aaa"

	// WHEN run is called on it
	text, err := slice.run("a0b1c2d3,a3bb2ccc1dddd0", utils.LogFrom{})

	// THEN the text was correctly extracted
	if err != nil {
		t.Errorf("Unexpected err %q",
			err.Error())
	}
	if want != text {
		t.Errorf("Should have got %q, not %q",
			want, text)
	}
}

func TestURLCommandSliceRunWithFail(t *testing.T) {
	// GIVEN a URLCommand URLCommandSlice that will pass
	jLog = utils.NewJLog("WARN", false)
	slice := URLCommandSlice{
		testURLCommandSplit(),
		testURLCommandRegex(),
		testURLCommandReplace(),
	}
	*slice[0].Text = ","
	*slice[1].Regex = "([e-z]+)[0-9]"
	slice[1].Index = 0
	*slice[2].Old = "c"
	*slice[2].New = "a"

	// WHEN run is called on it
	_, err := slice.run("a0b1c2d3,a3bb2ccc1dddd0", utils.LogFrom{})

	// THEN the text was correctly extracted
	if err == nil {
		t.Errorf("Should have failed and got an err, not %v",
			err)
	}
}

func TestURLCommandRunSplit(t *testing.T) {
	// GIVEN a Split URLCommand
	jLog = utils.NewJLog("WARN", false)
	command := testURLCommandSplit()

	// WHEN run is called on this
	text, err := command.run("youwantthissecondpart", utils.LogFrom{})

	// THEN the text is split correctly and returned
	want := "secondpart"
	if err != nil {
		t.Errorf("err should be nil, got %s",
			err.Error())
	}
	if want != text {
		t.Errorf("Want %q, got %q",
			want, text)
	}
}

func TestURLCommandRunReplace(t *testing.T) {
	// GIVEN a Replace URLCommand
	jLog = utils.NewJLog("WARN", false)
	command := testURLCommandReplace()

	// WHEN run is called on this
	text, err := command.run("iwantfoo", utils.LogFrom{})

	// THEN the text is replaced correctly and returned
	want := "iwantbar"
	if err != nil {
		t.Errorf("err should be nil, got %s",
			err.Error())
	}
	if want != text {
		t.Errorf("Want %q, got %q",
			want, text)
	}
}

func TestURLCommandRunRegex(t *testing.T) {
	// GIVEN a RegEx URLCommand
	jLog = utils.NewJLog("WARN", false)
	command := testURLCommandRegex()
	command.Index = 1

	// WHEN run is called on this
	text, err := command.run("-0-a-1.2.3-b-0-", utils.LogFrom{})

	// THEN the correct RegEx submatch is returned
	want := "1.2.3"
	if err != nil {
		t.Errorf("err should be nil, got %s",
			err.Error())
	}
	if want != text {
		t.Errorf("Want %q, got %q",
			want, text)
	}
}

func TestURLCommandRunRegexNegativeIndex(t *testing.T) {
	// GIVEN a RegEx URLCommand
	jLog = utils.NewJLog("WARN", false)
	command := testURLCommandRegex()
	command.Index = -1

	// WHEN run is called on this
	text, err := command.run("-0-a-1.2.3-b-4-", utils.LogFrom{})

	// THEN the correct RegEx submatch is returned
	want := "4"
	if err != nil {
		t.Errorf("err should be nil, got %q",
			err.Error())
	}
	if want != text {
		t.Errorf("Want %q, got %q",
			want, text)
	}
}

func TestURLCommandRunRegexOutOfRange(t *testing.T) {
	// GIVEN a RegEx URLCommand
	jLog = utils.NewJLog("WARN", false)
	command := testURLCommandRegex()
	command.Index = 10

	// WHEN run is called on this
	_, err := command.run("-0-a-1.2.3-b-0-", utils.LogFrom{})

	// THEN err is non-nil as Index is too big
	e := utils.ErrorToString(err)
	contain := " but the index wants element number "
	if !strings.Contains(e, contain) {
		t.Errorf("err should be non-nil and contain %q, got %q",
			contain, e)
	}
}

func TestURLCommandRunRegexThatDoesntMatch(t *testing.T) {
	// GIVEN a RegEx URLCommand
	jLog = utils.NewJLog("WARN", false)
	command := testURLCommandRegex()
	*command.Regex = "foo123"

	// WHEN run is called on this
	_, err := command.run("-0-a-1.2.3-b-0-", utils.LogFrom{})

	// THEN err is non-nil as Index is too big
	e := utils.ErrorToString(err)
	endswith := "didn't return any matches"
	if !strings.HasSuffix(e, endswith) {
		t.Errorf("err should be non-nil ending in %q, not %q",
			endswith, e)
	}
}

func TestURLCommandRunSplitNegativeIndex(t *testing.T) {
	// GIVEN a Split URLCommand
	jLog = utils.NewJLog("WARN", false)
	command := testURLCommandSplit()
	command.Index = -1

	// WHEN run is called on this
	text, err := command.run("youwantthissecondpartthisthirdpart", utils.LogFrom{})

	// THEN the text is split correctly and returned
	want := "thirdpart"
	if err != nil {
		t.Errorf("err should be nil, got %q",
			err.Error())
	}
	if want != text {
		t.Errorf("Want %q, got %q",
			want, text)
	}
}

func TestURLCommandRunSplitOutOfRange(t *testing.T) {
	// GIVEN a Split URLCommand
	jLog = utils.NewJLog("WARN", false)
	command := testURLCommandSplit()
	command.Index = 10

	// WHEN run is called on this
	_, err := command.run("youwantthissecondpartthisthirdpart", utils.LogFrom{})

	// THEN the text is split correctly and returned
	e := utils.ErrorToString(err)
	contain := " but the index wants element number "
	if !strings.Contains(e, contain) {
		t.Errorf("err should be non-nil and contain %q, got %q",
			contain, e)
	}
}

func TestURLCommandRunSplitThatDoesntMatch(t *testing.T) {
	// GIVEN a Split URLCommand
	jLog = utils.NewJLog("WARN", false)
	command := testURLCommandSplit()
	*command.Text = "unknown"

	// WHEN run is called on this
	_, err := command.run("youwantthissecondpartthisthirdpart", utils.LogFrom{})

	// THEN the text is split correctly and returned
	e := utils.ErrorToString(err)
	contain := "split didn't find any "
	if !strings.Contains(e, contain) {
		t.Errorf("err should be non-nil and contain %q, got %q",
			contain, e)
	}
}

func TestURLCommandCheckValuesSplitPass(t *testing.T) {
	// GIVEN a valid Split URLCommand
	command := testURLCommandSplit()

	// WHEN CheckValues is called on it
	err := command.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("%v should be valid, got \n%s",
			command, err.Error())
	}
}

func TestURLCommandCheckValuesSplitFail(t *testing.T) {
	// GIVEN an invalid Split URLCommand
	command := testURLCommandSplit()
	command.Text = nil

	// WHEN CheckValues is called on it
	err := command.CheckValues("")

	// THEN 2 errors were produced
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			command, wantCount, errCount, e)
	}
}

func TestURLCommandCheckValuesReplacePass(t *testing.T) {
	// GIVEN a valid Replace URLCommand
	command := testURLCommandReplace()

	// WHEN CheckValues is called on it
	err := command.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("%v should be valid, got \n%s",
			command, err.Error())
	}
}

func TestURLCommandCheckValuesReplaceFail(t *testing.T) {
	// GIVEN an invalid Replace URLCommand
	command := testURLCommandReplace()
	command.New = nil
	command.Old = nil

	// WHEN CheckValues is called on it
	err := command.CheckValues("")

	// THEN 3 errors were produced
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 3
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			command, wantCount, errCount, e)
	}
}

func TestURLCommandCheckValuesRegexPass(t *testing.T) {
	// GIVEN a valid Regex URLCommand
	command := testURLCommandRegex()

	// WHEN CheckValues is called on it
	err := command.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("%v should be valid, got \n%s",
			command, err.Error())
	}
}

func TestURLCommandCheckValuesRegexFail(t *testing.T) {
	// GIVEN an invalid Regex URLCommand
	command := testURLCommandRegex()
	command.Regex = nil

	// WHEN CheckValues is called on it
	err := command.CheckValues("")

	// THEN 2 errors were produced
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 2
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			command, wantCount, errCount, e)
	}
}

func TestURLCommandCheckValuesInvalidType(t *testing.T) {
	// GIVEN an unknown type URLCommand
	command := testURLCommandRegex()
	command.Type = "something"

	// WHEN CheckValues is called on it
	err := command.CheckValues("")

	// THEN 1 error was produced
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 1
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			command, wantCount, errCount, e)
	}
}

func TestCheckValuesWithNil(t *testing.T) {
	// GIVEN a nil URLCommand URLCommandSlice
	var slice *URLCommandSlice

	// WHEN CheckValues is called on it
	err := slice.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("err should be nil, not %q",
			err.Error())
	}
}

func TestCheckValuesPass(t *testing.T) {
	// GIVEN a URLCommand URLCommandSlice with every type
	slice := URLCommandSlice{
		testURLCommandRegex(),
		testURLCommandReplace(),
		testURLCommandSplit(),
	}

	// WHEN CheckValues is called on it
	err := slice.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("err should be nil, not %q",
			err.Error())
	}
}

func TestCheckValuesFail(t *testing.T) {
	// GIVEN a URLCommand URLCommandSlice with every type
	slice := URLCommandSlice{
		testURLCommandRegex(),
		testURLCommandRegex(),
		testURLCommandReplace(),
		testURLCommandReplace(),
		testURLCommandSplit(),
		testURLCommandSplit(),
	}
	slice[0].Regex = nil
	slice[2].New = nil
	slice[2].Old = nil
	slice[4].Text = nil

	// WHEN CheckValues is called on it
	err := slice.CheckValues("")

	// THEN 1 error was produced
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 11
	if errCount != wantCount {
		t.Errorf("%v contains 5 undefined vars, so should have %d errs, not %d!\nGot %s",
			slice, wantCount, errCount, e)
	}
}

func TestUnmarshalYAMLSingle(t *testing.T) {
	// GIVEN we've read a config file containg a single URLCommand not in list style
	var slice URLCommandSlice
	data, _ := os.ReadFile("../test/URLCommandSlice_single.yml")

	// WHEN Unmarshalled
	err := yaml.Unmarshal(data, &slice)

	// THEN the URLCommand is correctly unmarshalled
	wantType := "regex"
	wantRegex := "foo"
	wantIndex := 1
	wantText := "hi"
	wantOld := "was"
	wantNew := "now"
	wantIgnoreMisses := true
	if err != nil {
		t.Errorf("Unmarshal err'd %q",
			err.Error())
	}
	if len(slice) != 1 {
		t.Errorf("Expecting 1 URLCommand, got %d\n%v",
			len(slice), slice)
	}
	if slice[0].Type != wantType {
		t.Errorf("regex not unmarshalled to %s, got %v",
			wantType, *slice[0].Regex)
	}
	if *slice[0].Regex != wantRegex {
		t.Errorf("regex not unmarshalled to %s, got %v",
			wantRegex, *slice[0].Regex)
	}
	if slice[0].Index != wantIndex {
		t.Errorf("index not unmarshalled to %s, got %v",
			wantType, slice[0].Index)
	}
	if *slice[0].Text != wantText {
		t.Errorf("text not unmarshalled to %s, got %v",
			wantType, *slice[0].Text)
	}
	if *slice[0].Old != wantOld {
		t.Errorf("old not unmarshalled to %s, got %v",
			wantType, *slice[0].Old)
	}
	if *slice[0].New != wantNew {
		t.Errorf("new not unmarshalled to %s, got %v",
			wantType, *slice[0].New)
	}
	if *slice[0].IgnoreMisses != wantIgnoreMisses {
		t.Errorf("ignore_misses not unmarshalled to %s, got %v",
			wantType, *slice[0].IgnoreMisses)
	}
}

func TestUnmarshalYAMLMulti(t *testing.T) {
	// GIVEN we've read a config file containg a list of 2 URLCommands
	var slice URLCommandSlice
	data, _ := os.ReadFile("../test/URLCommandSlice_multi.yml")

	// WHEN Unmarshalled
	err := yaml.Unmarshal(data, &slice)

	// THEN the URLCommands are correctly unmarshalled
	wantType := "replace"
	if err != nil {
		t.Errorf("Unmarshal err'd %q",
			err.Error())
	}
	if len(slice) != 2 {
		t.Errorf("Expecting 1 URLCommand, got %d\n%v",
			len(slice), slice)
	}
	if slice[0].Type != wantType {
		t.Errorf("regex not unmarshalled to %s, got %v",
			wantType, *slice[0].Regex)
	}
}

func TestUnmarshalYAMLInvalid(t *testing.T) {
	// GIVEN we've read a config file containg invalid YAML
	var slice URLCommandSlice
	data, _ := os.ReadFile("../test/URLCommandSlice_invalid.yml")

	// WHEN Unmarshalled
	err := yaml.Unmarshal(data, &slice)

	// THEN the URLCommands fail to unmarshal
	if err == nil {
		t.Errorf("Unmarshal should've err'd, not %v",
			err)
	}
}
