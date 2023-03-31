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

package filter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/release-argus/Argus/util"
	"gopkg.in/yaml.v3"
)

// URLCommandSlice to be used to filter version from the URL Content.
type URLCommandSlice []URLCommand

// String returns a string representation of the URLCommandSlice.
func (s *URLCommandSlice) String() string {
	if s == nil {
		return ""
	}

	yamlBytes, _ := yaml.Marshal(s)
	return string(yamlBytes)
}

// URLCommand is a command to be ran to filter version from the URL body.
type URLCommand struct {
	Type  string  `yaml:"type" json:"type"`                       // regex/replace/split
	Regex *string `yaml:"regex,omitempty" json:"regex,omitempty"` // regex: regexp.MustCompile(Regex)
	Index int     `yaml:"index,omitempty" json:"index,omitempty"` // regex/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index]
	Text  *string `yaml:"text,omitempty" json:"text,omitempty"`   // split: strings.Split(tgtString, "Text")
	New   *string `yaml:"new,omitempty" json:"new,omitempty"`     // replace: strings.ReplaceAll(tgtString, "Old", "New")
	Old   *string `yaml:"old,omitempty" json:"old,omitempty"`     // replace: strings.ReplaceAll(tgtString, "Old", "New")
}

// String returns a string representation of the URLCommand.
func (c *URLCommand) String() string {
	if c == nil {
		return ""
	}

	yamlBytes, _ := yaml.Marshal(c)
	return string(yamlBytes)
}

// UnmarshalYAML allows handling of a dict as well as a list of dicts.
//
// It will convert a dict to a list of a dict.
//
// e.g.    URLCommandSlice: { type: "split" }
//
// becomes URLCommandSlice: [ { type: "split" } ]
func (s *URLCommandSlice) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var multi []URLCommand
	err = unmarshal(&multi)
	if err != nil {
		var single URLCommand
		err = unmarshal(&single)
		if err != nil {
			return
		}
		*s = []URLCommand{single}
	} else {
		*s = multi
	}
	return
}

// Print the URLCommand's in the URLCommandSlice.
func (s *URLCommandSlice) Print(prefix string) {
	if s == nil || len(*s) == 0 {
		return
	}
	fmt.Printf("%surl_commands:\n", prefix)

	for i := range *s {
		(*s)[i].Print(prefix + "  ")
	}
}

// Print will print the URLCommand.
func (c *URLCommand) Print(prefix string) {
	fmt.Printf("%s- type: %s\n", prefix, c.Type)
	switch c.Type {
	case "regex":
		fmt.Printf("%s  regex: %q\n", prefix, *c.Regex)
		util.PrintlnIfNotDefault(c.Index,
			fmt.Sprintf("%s  index: %d", prefix, c.Index))
	case "replace":
		fmt.Printf("%s  new: %q\n", prefix, *c.New)
		fmt.Printf("%s  old: %q\n", prefix, *c.Old)
	case "split":
		fmt.Printf("%s  text: %q\n", prefix, *c.Text)
		util.PrintlnIfNotDefault(c.Index,
			fmt.Sprintf("%s  index: %d", prefix, c.Index))
	}
}

// Run all of the URLCommand(s) in this URLCommandSlice.
func (s *URLCommandSlice) Run(text string, logFrom util.LogFrom) (string, error) {
	if s == nil {
		return text, nil
	}

	logFrom.Secondary = "url_commands"
	var err error
	for commandIndex := range *s {
		text, err = (*s)[commandIndex].run(text, &logFrom)
		if err != nil {
			return text, err
		}
	}
	return text, nil
}

// run this URLCommand on `text`
func (c *URLCommand) run(text string, logFrom *util.LogFrom) (string, error) {
	var err error
	// Iterate through the commands to filter the text.
	textBak := text
	msg := fmt.Sprintf("Looking through:\n%q", text)
	jLog.Debug(msg, *logFrom, true)

	switch c.Type {
	case "split":
		msg = fmt.Sprintf("Splitting on %q with index %d", *c.Text, c.Index)
		text, err = c.split(text, logFrom)
	case "replace":
		msg = fmt.Sprintf("Replacing %q with %q", *c.Old, *c.New)
		text = strings.ReplaceAll(text, *c.Old, *c.New)
	case "regex":
		msg = fmt.Sprintf("Regexing %q", *c.Regex)
		text, err = c.regex(text, logFrom)
	}
	if err != nil {
		return textBak, err
	}

	msg = fmt.Sprintf("%s\nResolved to %s", msg, text)
	jLog.Debug(msg, *logFrom, true)
	return text, err
}

// regex `text` with the URLCommand's regex.
func (c *URLCommand) regex(text string, logFrom *util.LogFrom) (string, error) {
	re := regexp.MustCompile(*c.Regex)

	index := c.Index
	texts := re.FindAllStringSubmatch(text, -1)
	// Handle negative indices.
	if index < 0 {
		index = len(texts) + c.Index
	}

	if len(texts) == 0 {
		err := fmt.Errorf("%s %q didn't return any matches",
			c.Type, *c.Regex)
		if len(text) < 20 {
			err = fmt.Errorf("%w on %q",
				err, text)
		}
		jLog.Warn(err, *logFrom, true)

		return text, err
	}

	if (len(texts) - index) < 1 {
		err := fmt.Errorf("%s (%s) returned %d elements but the index wants element number %d",
			c.Type, *c.Regex, len(texts), (index + 1))
		jLog.Warn(err, *logFrom, true)

		return text, err
	}

	return texts[index][len(texts[index])-1], nil
}

// split `text` with the URLCommand's text amd return the index specified.
func (c *URLCommand) split(text string, logFrom *util.LogFrom) (string, error) {
	texts := strings.Split(text, *c.Text)

	if len(texts) == 1 {
		err := fmt.Errorf("%s didn't find any %q to split on",
			c.Type, *c.Text)
		jLog.Warn(err, *logFrom, true)

		return text, err
	}

	index := c.Index
	// Handle negative indices.
	if index < 0 {
		index = len(texts) + index
	}

	if (len(texts) - index) < 1 {
		err := fmt.Errorf("%s (%s) returned %d elements but the index wants element number %d",
			c.Type, *c.Text, len(texts), (index + 1))
		jLog.Warn(err, *logFrom, true)

		return text, err
	}

	return texts[index], nil
}

// CheckValues of the URLCommand(s) in the URLCommandSlice.
func (s *URLCommandSlice) CheckValues(prefix string) (errs error) {
	if s == nil {
		return
	}

	for index := range *s {
		if err := (*s)[index].CheckValues(prefix + "    "); err != nil {
			errs = fmt.Errorf("%s%s  item_%d:\\%w",
				util.ErrorToString(errs), prefix, index, err)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%surl_commands:\\%s",
			prefix, util.ErrorToString(errs))
	}
	return
}

// CheckValues of the URLCommand.
func (c *URLCommand) CheckValues(prefix string) (errs error) {
	validType := true

	switch c.Type {
	case "regex":
		if c.Regex == nil {
			errs = fmt.Errorf("%s%sregex: <required> (regex to use)\\",
				util.ErrorToString(errs), prefix)
		} else {
			_, err := regexp.Compile(*c.Regex)
			if err != nil {
				errs = fmt.Errorf("%s%sregex: %q <invalid> (Invalid RegEx)\\",
					util.ErrorToString(errs), prefix, *c.Regex)
			}
		}
	case "replace":
		if c.New == nil {
			errs = fmt.Errorf("%s%snew: <required> (text you want to replace with)\\",
				util.ErrorToString(errs), prefix)
		}
		if c.Old == nil {
			errs = fmt.Errorf("%s%sold: <required> (text you want replaced)\\",
				util.ErrorToString(errs), prefix)
		}
	case "split":
		if c.Text == nil {
			errs = fmt.Errorf("%s%stext: <required> (text to split on)\\",
				util.ErrorToString(errs), prefix)
		}
	default:
		validType = false
		errs = fmt.Errorf("%s%stype: %q <invalid> is not a valid url_command (regex/replace/split)\\",
			util.ErrorToString(errs), prefix, c.Type)
	}

	if errs != nil && validType {
		errs = fmt.Errorf("%stype: %s\\%w",
			prefix, c.Type, errs)
	}
	return errs
}

// URLCommandsFromStr converts a JSON string to a URLCommandSlice.
func URLCommandsFromStr(jsonStr *string, defaults *URLCommandSlice, logFrom *util.LogFrom) (*URLCommandSlice, error) {
	// jsonStr == nil when it hasn't been changed, so just use defaults
	if jsonStr == nil {
		return defaults, nil
	}

	// Try and unmarshal this JSON string
	var urlCommands URLCommandSlice
	err := json.Unmarshal([]byte(*jsonStr), &urlCommands)
	// Ignore the JSON if it failed to unmarshal
	if err != nil {
		jLog.Error(fmt.Sprintf("Failed converting JSON - %q\n%s", *jsonStr, util.ErrorToString(err)),
			*logFrom, err != nil)
		return defaults, fmt.Errorf("failed converting JSON - %w", err)
	}

	// Check the URLCommands
	err = urlCommands.CheckValues("")
	if err != nil {
		return defaults, err
	}

	return &urlCommands, nil
}
