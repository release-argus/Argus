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

package url_command

import (
	"github.com/release-argus/Argus/utils"
)

var (
	jLog *utils.JLog
)

// Slice of URLCommand to be used to filter version from the URL Content.
type Slice []URLCommand

// URLCommand is a command to be ran to filter version from the URL body.
type URLCommand struct {
	Type               string  `yaml:"type"`                    // regex/replace/split
	Regex              *string `yaml:"regex,omitempty"`         // regex: regexp.MustCompile(Regex)
	Index              int     `yaml:"index,omitempty"`         // regex/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index]
	Text               *string `yaml:"text,omitempty"`          // split:                strings.Split(tgtString, "Text")
	New                *string `yaml:"new,omitempty"`           // replace:              strings.ReplaceAll(tgtString, "Old", "New")
	Old                *string `yaml:"old,omitempty"`           // replace:              strings.ReplaceAll(tgtString, "Old", "New")
	IgnoreMisses       *bool   `yaml:"ignore_misses,omitempty"` // Ignore this command failing (e.g. split on text that doesn't exist)
	ParentIgnoreMisses *bool   `yaml:"-"`                       // IgnoreMisses, but from the parent Service (used as default)
}
