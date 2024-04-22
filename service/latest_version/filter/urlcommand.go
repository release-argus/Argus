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

package filter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/release-argus/Argus/util"
)

// URLCommandSlice to be used to filter version from the URL Content.
type URLCommandSlice []URLCommand

// String returns a string representation of the URLCommandSlice.
func (s *URLCommandSlice) String() (str string) {
	if s != nil {
		str = util.ToYAMLString(s, "")
	}
	return
}

// URLCommand is a command to be ran to filter version from the URL body.
type URLCommand struct {
	Type     string  `yaml:"type" json:"type"`                             // regex/replace/split
	Regex    *string `yaml:"regex,omitempty" json:"regex,omitempty"`       // regex: regexp.MustCompile(Regex)
	Index    int     `yaml:"index,omitempty" json:"index,omitempty"`       // regex/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index]
	Template *string `yaml:"template,omitempty" json:"template,omitempty"` // regex: template
	Text     *string `yaml:"text,omitempty" json:"text,omitempty"`         // split: strings.Split(tgtString, "Text")
	New      *string `yaml:"new,omitempty" json:"new,omitempty"`           // replace: strings.ReplaceAll(tgtString, "Old", "New")
	Old      *string `yaml:"old,omitempty" json:"old,omitempty"`           // replace: strings.ReplaceAll(tgtString, "Old", "New")
}

// String returns a string representation of the URLCommand.
func (c *URLCommand) String() (str string) {
	if c != nil {
		str = util.ToYAMLString(c, "")
	}
	return
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

// Run all of the URLCommand(s) in this URLCommandSlice.
func (s *URLCommandSlice) Run(text string, logFrom *util.LogFrom) (string, error) {
	if s == nil {
		return text, nil
	}

	urlCommandLogFrom := &util.LogFrom{Primary: logFrom.Primary, Secondary: "url_commands"}
	var err error
	for commandIndex := range *s {
		text, err = (*s)[commandIndex].run(text, urlCommandLogFrom)
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
	if jLog.IsLevel("DEBUG") {
		jLog.Debug(
			fmt.Sprintf("Looking through:\n%q", text),
			logFrom, true)
	}

	var msg string
	switch c.Type {
	case "split":
		msg = fmt.Sprintf("Splitting on %q with index %d", *c.Text, c.Index)
		text, err = c.split(text, logFrom)
	case "replace":
		msg = fmt.Sprintf("Replacing %q with %q", *c.Old, *c.New)
		text = strings.ReplaceAll(text, *c.Old, *c.New)
	case "regex":
		msg = fmt.Sprintf("Regexing %q", *c.Regex)
		if c.Template != nil {
			msg = fmt.Sprintf("%s with template %q", msg, *c.Template)
		}
		text, err = c.regex(text, logFrom)
	}
	if err != nil {
		return textBak, err
	}

	msg = fmt.Sprintf("%s\nResolved to %s", msg, text)
	if jLog.IsLevel("DEBUG") {
		jLog.Debug(msg, logFrom, true)
	}
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

	// No matches.
	if len(texts) == 0 {
		err := fmt.Errorf("%s %q didn't return any matches",
			c.Type, *c.Regex)
		if len(text) < 20 {
			err = fmt.Errorf("%w on %q",
				err, text)
		}
		jLog.Warn(err, logFrom, true)

		return text, err
	}
	// Index out of range.
	if (len(texts) - index) < 1 {
		err := fmt.Errorf("%s (%s) returned %d elements on %q, but the index wants element number %d",
			c.Type, *c.Regex, len(texts), text, (index + 1))
		jLog.Warn(err, logFrom, true)

		return text, err
	}

	regexMatches := texts[index]
	return util.RegexTemplate(regexMatches, c.Template), nil
}

// split `text` with the URLCommand's text amd return the index specified.
func (c *URLCommand) split(text string, logFrom *util.LogFrom) (string, error) {
	texts := strings.Split(text, *c.Text)

	if len(texts) == 1 {
		err := fmt.Errorf("%s didn't find any %q to split on",
			c.Type, *c.Text)
		jLog.Warn(err, logFrom, true)

		return text, err
	}

	index := c.Index
	// Handle negative indices.
	if index < 0 {
		index = len(texts) + index
	}

	if (len(texts) - index) < 1 {
		err := fmt.Errorf("%s (%s) returned %d elements on %q, but the index wants element number %d",
			c.Type, *c.Text, len(texts), text, (index + 1))
		jLog.Warn(err, logFrom, true)

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
		// Remove the template if it's empty
		if util.DefaultIfNil(c.Template) == "" {
			c.Template = nil
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
			logFrom, err != nil)
		return defaults, fmt.Errorf("failed converting JSON - %w", err)
	}

	// Check the URLCommands
	err = urlCommands.CheckValues("")
	if err != nil {
		return defaults, err
	}

	return &urlCommands, nil
}
