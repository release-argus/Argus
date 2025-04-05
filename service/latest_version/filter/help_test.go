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

//go:build unit || integration

package filter

import (
	"os"
	"testing"
	"time"

	"github.com/release-argus/Argus/command"
	logtest "github.com/release-argus/Argus/test/log"
)

var packageName = "latestver.filter"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	// Exit.
	os.Exit(exitCode)
}

func testURLCommandRegex() URLCommand {
	regex := "-([0-9.]+)-"
	index := 0
	return URLCommand{
		Type:  "regex",
		Regex: regex,
		Index: &index}
}

func testURLCommandRegexTemplate() URLCommand {
	regex := "-([0-9.]+)-"
	index := 0
	template := "_$1_"
	return URLCommand{
		Type:     "regex",
		Regex:    regex,
		Index:    &index,
		Template: template}
}

func testURLCommandReplace() URLCommand {
	old := "foo"
	new := "bar"
	return URLCommand{
		Type: "replace",
		Old:  old,
		New:  &new}
}

func testURLCommandSplit() URLCommand {
	text := "this"
	index := 1
	return URLCommand{
		Type:  "split",
		Text:  text,
		Index: &index}
}

func testRequire() Require {
	return Require{
		Command:      command.Command{"echo", "foo"},
		RegexContent: "bish",
		RegexVersion: "bash",
		Docker: NewDockerCheck(
			"ghcr",
			"releaseargus/argus",
			"latest",
			"", "", "", time.Now(), nil)}
}
