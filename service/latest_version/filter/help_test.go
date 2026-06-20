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

//go:build unit || integration

package filter

import (
	"fmt"
	"os"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
)

var packageName = "latestver.filter"

func TestMain(m *testing.M) {
	// Log.
	logtest.InitLog()

	// Run other tests.
	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Printf("%s\nexit code channel not empty", packageName)
		exitCode = 1
	}

	// Exit.
	os.Exit(exitCode)
}

// plainDefaults returns plain defaults and hardDefaults for testing.
func plainDefaults(t *testing.T) (*RequireDefaults, *RequireDefaults) {
	t.Helper()

	defaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults, _ := DecodeDefaults("yaml", nil)
	hardDefaults.Default()

	defaults.SetDefaults(hardDefaults)

	return defaults, hardDefaults
}

func testURLCommandRegex() URLCommand {
	regex := "-([0-9.]+)-"
	index := 0

	return URLCommand{
		Type:  "regex",
		Regex: regex,
		Index: &index,
	}
}

func testURLCommandRegexTemplate() URLCommand {
	regex := "-([0-9.]+)-"
	index := 0
	template := "_$1_"

	return URLCommand{
		Type:     "regex",
		Regex:    regex,
		Index:    &index,
		Template: template,
	}
}

func testURLCommandReplace() URLCommand {
	oldText := "foo"
	newText := "bar"

	return URLCommand{
		Type: "replace",
		Old:  oldText,
		New:  newText,
	}
}

func testURLCommandSplit() URLCommand {
	text := "this"
	index := 1

	return URLCommand{
		Type:  "split",
		Text:  text,
		Index: &index,
	}
}
