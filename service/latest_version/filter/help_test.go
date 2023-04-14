// Copyright [2023] [Argus]
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
	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/util"
)

func stringPtr(val string) *string {
	return &val
}
func testLogging(level string) {
	jLog = util.NewJLog(level, false)
	LogInit(jLog)
}

func testURLCommandRegex() URLCommand {
	regex := "-([0-9.]+)-"
	index := 0
	return URLCommand{
		Type:  "regex",
		Regex: &regex,
		Index: index,
	}
}

func testURLCommandReplace() URLCommand {
	old := "foo"
	new := "bar"
	return URLCommand{
		Type: "replace",
		Old:  &old,
		New:  &new,
	}
}

func testURLCommandSplit() URLCommand {
	text := "this"
	index := 1
	return URLCommand{
		Type:  "split",
		Text:  &text,
		Index: index,
	}
}

func testRequire() Require {
	return Require{
		Command:      command.Command{"echo", "foo"},
		RegexContent: "bish",
		RegexVersion: "bash",
		Docker: &DockerCheck{
			Type:  "ghcr",
			Image: "releaseargus/argus",
			Tag:   "latest",
		},
	}
}
