// Copyright [2022] [Hymenaios]
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

package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hymenaios-io/Hymenaios/utils"
)

// URLCommandSlice is a slice of URLCommand to be used to filter version from the URL Content.
type URLCommandSlice []URLCommand

// URLCommand is a command to be ran to filter version from the URL body.
type URLCommand struct {
	Type               string  `yaml:"type"`                    // regex/regex_submatch/replace/split
	Regex              *string `yaml:"regex,omitempty"`         // regex/regex_submatch: regexp.MustCompile(Regex)
	Index              *int    `yaml:"index,omitempty"`         // regex_submatch/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index]
	Text               *string `yaml:"text,omitempty"`          // split:                strings.Split(tgtString, "Text")
	New                *string `yaml:"new,omitempty"`           // replace:              strings.ReplaceAll(tgtString, "Old", "New")
	Old                *string `yaml:"old,omitempty"`           // replace:              strings.ReplaceAll(tgtString, "Old", "New")
	IgnoreMisses       *bool   `yaml:"ignore_misses,omitempty"` // Ignore this command failing (e.g. split on text that doesn't exist)
	ParentIgnoreMisses *bool   `yaml:"-"`                       // IgnoreMisses, but from the parent Service (used as default)
}

// UnmarshalYAML allows handling of a dict as well as a list of dicts.
//
// It will convert a dict to a list of a dict.
//
// e.g.    URLCommandSlice: { type: "split" }
//
// becomes URLCommandSlice: [ { type: "split" } ]
func (c *URLCommandSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var multi []URLCommand
	err := unmarshal(&multi)
	if err != nil {
		var single URLCommand
		err := unmarshal(&single)
		if err != nil {
			return err
		}
		*c = []URLCommand{single}
	} else {
		*c = multi
	}
	return nil
}

// Print will print the URLCommand's in the URLCommandSlice.
func (c *URLCommandSlice) Print(prefix string) {
	if c == nil {
		fmt.Printf("%surl_commands: []\n", prefix)
		return
	}

	for _, command := range *c {
		command.Print(prefix)
	}
}

// Print will print the URLCommand.
func (c *URLCommand) Print(prefix string) {
	fmt.Printf("%s  - type: %s\n", prefix, c.Type)
	switch c.Type {
	case "regex":
		fmt.Printf("%s    regex: %q\n", prefix, *c.Regex)
		utils.PrintlnIfNotNil(c.GetIgnoreMisses(), fmt.Sprintf("%s    ignore_misses: %t", prefix, *c.GetIgnoreMisses()))
		fmt.Printf("%s    index: %d\n", prefix, *c.Index)
	case "regex_submatch":
		fmt.Printf("%s    regex: %q\n", prefix, *c.Regex)
		utils.PrintlnIfNotNil(c.GetIgnoreMisses(), fmt.Sprintf("%s    ignore_misses: %t", prefix, *c.GetIgnoreMisses()))
	case "replace":
		fmt.Printf("%s    new: %q\n", prefix, *c.New)
		fmt.Printf("%s    old: %q\n", prefix, *c.Old)
	case "split":
		fmt.Printf("%s    text: %q\n", prefix, *c.Text)
		fmt.Printf("%s    index: %d\n", prefix, *c.Index)
		utils.PrintlnIfNotNil(c.GetIgnoreMisses(), fmt.Sprintf("%s    ignore_misses: %t", prefix, *c.GetIgnoreMisses()))
	}
}

// SetParentIgnoreMisses will set ParentIgnoreMisses of each URLCommand in the Slice to ignore if it's nil in the slice.
func (c *URLCommandSlice) SetParentIgnoreMisses(ignore *bool) {
	if c == nil {
		return
	}

	for commandIndex := range *c {
		(*c)[commandIndex].ParentIgnoreMisses = ignore
	}
}

// GetIgnoreMisses will get the IgnoreMisses of this URLCommand, or ParentIgnoreMisses if that's nil.
func (c *URLCommand) GetIgnoreMisses() *bool {
	return utils.GetFirstNonNilPtr(c.IgnoreMisses, c.ParentIgnoreMisses)
}

// run will run all of the URLCommand(s) in this URLCommandSlice.
func (c *URLCommandSlice) run(text string, logFrom utils.LogFrom) (string, error) {
	if c == nil {
		return text, nil
	}

	logFrom.Secondary = "url_commands"
	var err error
	for commandIndex := range *c {
		text, err = (*c)[commandIndex].run(text, logFrom)
		if err != nil {
			return text, err
		}
	}
	return text, nil
}

// run will exectue this URLCommand on `text`
func (c *URLCommand) run(text string, logFrom utils.LogFrom) (string, error) {
	var err error
	// Iterate through the commands to filter the text.
	textBak := text
	msg := fmt.Sprintf("Looking through:\n%s", text)
	jLog.Debug(msg, logFrom, true)

	switch c.Type {
	case "split":
		text, err = c.split(text, logFrom)
	case "replace":
		text = strings.ReplaceAll(text, *c.Old, *c.New)
	case "regex", "regex_submatch":
		text, err = c.regex(text, logFrom)
	}
	if err != nil {
		return textBak, err
	}

	msg = fmt.Sprintf("Resolved to %s", text)
	jLog.Debug(msg, logFrom, true)
	return text, err
}

func (c *URLCommand) regex(text string, logFrom utils.LogFrom) (string, error) {
	re := regexp.MustCompile(*c.Regex)

	index := c.Index
	var texts []string
	switch c.Type {
	case "regex":
		texts = re.FindAllString(text, -1)
		// Handle negative indices.
		if *c.Index < 0 {
			*index = len(texts) + *c.Index
		}
	case "regex_submatch":
		texts = re.FindStringSubmatch(text)
		index = utils.PtrOrValueToPtr(c.Index, 1)
	}

	if len(texts) == 0 {
		err := fmt.Errorf("%s (%s) didn't return any matches", c.Type, *c.Regex)
		jLog.Warn(err, logFrom, !utils.EvalBoolPtr(c.GetIgnoreMisses(), false))

		return text, err
	}

	if (len(texts) - *index) < 1 {
		err := fmt.Errorf("%s (%s) returned %d elements but the index wants element number %d", c.Type, *c.Regex, len(texts), (*index + 1))
		jLog.Warn(err, logFrom, !utils.EvalBoolPtr(c.GetIgnoreMisses(), false))

		return text, err
	}

	return texts[*index], nil
}

func (c *URLCommand) split(text string, logFrom utils.LogFrom) (string, error) {
	texts := strings.Split(text, *c.Text)

	if len(texts) == 1 {
		err := fmt.Errorf("%s didn't find any %q to split on", c.Type, *c.Text)
		jLog.Warn(err, logFrom, !utils.EvalBoolPtr(c.GetIgnoreMisses(), false))

		return text, err
	}

	index := *c.Index
	// Handle negative indices.
	if index < 0 {
		index = len(texts) + index
	}

	if (len(texts) - index) < 1 {
		err := fmt.Errorf("%s (%s) returned %d elements but the index wants element number %d", c.Type, *c.Text, len(texts), (index + 1))
		jLog.Warn(err, logFrom, !utils.EvalBoolPtr(c.GetIgnoreMisses(), false))

		return text, err
	}

	return texts[index], nil
}

// CheckValues of the URLCommand(s) in the URLCommandSlice.
func (c *URLCommandSlice) CheckValues(prefix string) error {
	if c == nil {
		return nil
	}

	var errs error
	for index := range *c {
		if err := (*c)[index].CheckValues(prefix + "    "); err != nil {
			errs = fmt.Errorf("%s%s  item_%d:\\%w", utils.ErrorToString(errs), prefix, index, err)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%surl_commands:\\%s", prefix, utils.ErrorToString(errs))
	}
	return errs
}

// CheckValues of the URLCommand.
func (c *URLCommand) CheckValues(prefix string) (errs error) {
	validType := true

	switch c.Type {
	case "split":
		if c.Index == nil {
			errs = fmt.Errorf("%s%sindex: <required> (element of the split to take)\\", utils.ErrorToString(errs), prefix)
		} else if *c.Index < 0 {
			errs = fmt.Errorf("%s%sindex: <invalid> (indices must be non-negative)\\", utils.ErrorToString(errs), prefix)
		}
		if c.Text == nil {
			errs = fmt.Errorf("%s%stext: <required> (text to split on)\\", utils.ErrorToString(errs), prefix)
		}
	case "replace":
		if c.New == nil {
			errs = fmt.Errorf("%s%snew: <required> (text you want to replace with)\\", utils.ErrorToString(errs), prefix)
		}
		if c.Old == nil {
			errs = fmt.Errorf("%s%sold: <required> (text you want replaced)\\", utils.ErrorToString(errs), prefix)
		}
	case "regex":
		if c.Regex == nil {
			errs = fmt.Errorf("%s%sregex: <required> (regex to use)\\", utils.ErrorToString(errs), prefix)
		}
	case "regex_submatch":
		if c.Regex == nil {
			errs = fmt.Errorf("%s%sregex: <required> (regex to use)\\", utils.ErrorToString(errs), prefix)
		}
	default:
		validType = false
		errs = fmt.Errorf("%s%stype: <invalid> %q is not a valid url_command (split/replace/regex/regex_submatch)\\", utils.ErrorToString(errs), prefix, c.Type)
	}

	if errs != nil && validType {
		errs = fmt.Errorf("%stype: %s\\%s\\", prefix, c.Type, errs)
	}
	return
}
