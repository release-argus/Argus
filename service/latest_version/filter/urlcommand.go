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

// Package filter provides filtering for latest_version queries.
package filter

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

var urlCommandTypes = []string{"regex", "replace", "split"}

// URLCommandSlice is a list of URLCommand that filter versions from the URL Content.
type URLCommandSlice []URLCommand

// UnmarshalJSON allows handling of a dict as well as a list of dicts.
//
// It will convert a dict to a list of a dict.
//
//	e.g. URLCommandSlice: { "type": "split" }
//	becomes URLCommandSlice: [ { "type": "split" } ]
func (s *URLCommandSlice) UnmarshalJSON(data []byte) error {
	// Handle the case where data is a quoted JSON string. (from web requests).
	if len(data) > 0 && data[0] == '"' && data[len(data)-1] == '"' {
		var jsonStr string
		if err := json.Unmarshal(data, &jsonStr); err != nil {
			return err //nolint:wrapcheck
		}

		data = []byte(jsonStr)
	}

	return s.unmarshal(func(v interface{}) error {
		return json.Unmarshal(data, v)
	})
}

// UnmarshalYAML allows handling of a dict as well as a list of dicts.
//
// It will convert a dict to a list of a dict.
//
//	e.g. URLCommandSlice: { type: "split" }
//	becomes URLCommandSlice: [ { type: "split" } ]
func (s *URLCommandSlice) UnmarshalYAML(value *yaml.Node) error {
	return s.unmarshal(func(v interface{}) error {
		return value.Decode(v)
	})
}

// unmarshal will unmarshal the URLCommandSlice using the provided unmarshal function.
func (s *URLCommandSlice) unmarshal(unmarshalFunc func(interface{}) error) error {
	// Alias to avoid recursion.
	var multi []URLCommand
	if err := unmarshalFunc(&multi); err == nil {
		*s = multi
		return nil
	}

	// Else, try to unmarshal as a single URLCommand.
	var single URLCommand
	err := unmarshalFunc(&single)
	if err == nil {
		*s = []URLCommand{single}
		return nil
	}

	return err //nolint:wrapcheck
}

// String returns a string representation of the URLCommandSlice.
func (s *URLCommandSlice) String() string {
	if s == nil {
		return ""
	}
	return util.ToYAMLString(s, "")
}

// URLCommand is a command to filter versions from the URL body.
type URLCommand struct {
	Type     string  `json:"type" yaml:"type"`                             // regex/replace/split.
	Regex    string  `json:"regex,omitempty" yaml:"regex,omitempty"`       // regex: regexp.MustCompile(Regex).
	Index    *int    `json:"index,omitempty" yaml:"index,omitempty"`       // regex/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index].
	Template string  `json:"template,omitempty" yaml:"template,omitempty"` // regex: template.
	Text     string  `json:"text,omitempty" yaml:"text,omitempty"`         // split: strings.Split(tgtString, "Text").
	New      *string `json:"new,omitempty" yaml:"new,omitempty"`           // replace: strings.ReplaceAll(tgtString, "Old", "New").
	Old      string  `json:"old,omitempty" yaml:"old,omitempty"`           // replace: strings.ReplaceAll(tgtString, "Old", "New").
}

// String returns a string representation of the URLCommand.
func (c *URLCommand) String() string {
	if c == nil {
		return ""
	}
	return util.ToYAMLString(c, "")
}

// GetVersions from `text` using the URLCommands in this URLCommandSlice.
func (s *URLCommandSlice) GetVersions(text string, logFrom logutil.LogFrom) ([]string, error) {
	// No URLCommands to run, so treat the text as a single version.
	if len(*s) == 0 {
		if text == "" {
			return nil, nil
		}
		return []string{text}, nil
	}
	return s.Run(text, logFrom)
}

// Run all of the URLCommands in this URLCommandSlice on `text`.
func (s *URLCommandSlice) Run(text string, logFrom logutil.LogFrom) ([]string, error) {
	if s == nil {
		return nil, nil
	}

	urlCommandLogFrom := logutil.LogFrom{Primary: logFrom.Primary, Secondary: "url_commands"}
	versions := []string{text}
	for _, urlCommand := range *s {
		if err := urlCommand.run(&versions, urlCommandLogFrom); err != nil {
			return nil, err
		}
	}
	return versions, nil
}

// run this URLCommand on `text`.
func (c *URLCommand) run(versions *[]string, logFrom logutil.LogFrom) error {
	var err error

	for i, version := range *versions {
		// Iterate through the commands to filter the text.
		if logutil.Log.IsLevel("DEBUG") {
			logutil.Log.Debug(
				fmt.Sprintf("Looking through:\n%q", version),
				logFrom, true)
		}

		var msg string
		switch c.Type {
		case "split":
			msg = fmt.Sprintf("Splitting on %q with index %d",
				c.Text, c.Index)
			err = c.split(i, versions, logFrom)
		case "replace":
			msg = fmt.Sprintf("Replacing %q with %q",
				c.Old, *c.New)
			(*versions)[i] = strings.ReplaceAll(version, c.Old, *c.New)
		case "regex":
			msg = fmt.Sprintf("Regexing %q", c.Regex)
			if c.Template != "" {
				msg = fmt.Sprintf("%s with template %q",
					msg, c.Template)
			}
			err = c.regex(i, versions, logFrom)
		}
		if err != nil {
			return err
		}

		if logutil.Log.IsLevel("DEBUG") {
			msg = fmt.Sprintf("%s\nResolved to %q",
				msg, version)
			logutil.Log.Debug(msg, logFrom, true)
		}
	}
	return nil
}

// regex applies the URLCommands regex to `versions[versionIndex]`.
//
// Parameters:
//   - versionIndex: The index of the version in the `versions` slice to validate.
//   - versions: A pointer to the slice of version strings to regex.
//   - logFrom: Used for logging the source of the operation.
func (c *URLCommand) regex(versionIndex int, versions *[]string, logFrom logutil.LogFrom) error {
	re := regexp.MustCompile(c.Regex)

	version := (*versions)[versionIndex]
	matches := re.FindAllStringSubmatch(version, -1)
	// No matches.
	if len(matches) == 0 {
		err := fmt.Errorf("%s %q didn't return any matches on %q",
			c.Type, c.Regex, util.TruncateMessage(version, 50))
		logutil.Log.Warn(err, logFrom, true)
		return err
	}

	// Specific index requested.
	if c.Index != nil {
		index := *c.Index
		// Handle negative indices.
		if index < 0 {
			index = len(matches) + *c.Index
		}

		// Index out of range.
		if (len(matches) - index) < 1 {
			err := fmt.Errorf("%s (%s) returned %d elements on %q, but the index wants element number %d",
				c.Type, c.Regex, len(matches), version, index+1)
			logutil.Log.Warn(err, logFrom, true)
			return err
		}

		(*versions)[len(*versions)-1] = util.RegexTemplate(matches[index], c.Template)
		return nil
	}

	// Add all subMatches to the versions list.
	subMatch := make([]string, len(matches))
	for i := range matches {
		subMatch[i] = util.RegexTemplate(matches[i], c.Template)
	}

	// Replace the current version in the list with the ordered subVersions.
	util.ReplaceWithElements(versions, versionIndex, subMatch)
	return nil
}

// split processes the version string at `versions[versionIndex]` by splitting it
// using the URLCommands text pattern, and updates the element at `versionIndex`.
//
//   - If no `Index` is specified, the entire split result replaces the version string at `versionIndex`.
//   - If `Index` is specified, the element at that index replaces the current version string at `versionIndex`.
//   - Negative indices are supported, where `-1` refers to the last element.
//   - If the split result does not contain enough elements to retrieve the specified index, an error is returned.
//
// Parameters:
//   - versionIndex: The index of the version in the `versions` slice to process.
//   - versions: A pointer to the slice of version strings to modify.
//   - logFrom: Used for logging the source of the operation.
func (c *URLCommand) split(versionIndex int, versions *[]string, logFrom logutil.LogFrom) error {
	texts, err := c.splitAllMatches((*versions)[versionIndex], logFrom)
	if err != nil {
		return err
	}

	// If no index specified, replace versionIndex with the split text.
	if c.Index == nil {
		util.ReplaceWithElements(versions, versionIndex, texts)
		return nil
	}

	index := *c.Index
	// Handle negative indices.
	if index < 0 {
		index = len(texts) + index
	}

	if (len(texts) - index) < 1 {
		err := fmt.Errorf("%s (%s) returned %d elements on %q, but the index wants element number %d",
			c.Type, c.Text, len(texts), (*versions)[versionIndex], index+1)
		logutil.Log.Warn(err, logFrom, true)

		return err
	}

	(*versions)[versionIndex] = texts[index]
	return nil
}

// splitAllMatches will split `text` on the URLCommands text, and return all matches.
func (c *URLCommand) splitAllMatches(text string, logFrom logutil.LogFrom) ([]string, error) {
	texts := strings.Split(text, c.Text)
	if len(texts) == 1 {
		err := fmt.Errorf("%s didn't find any %q to split on",
			c.Type, c.Text)
		logutil.Log.Warn(err, logFrom, true)

		return nil, err
	}
	return texts, nil
}

// CheckValues validates the fields of each URLCommand in the URLCommandSlice.
func (s *URLCommandSlice) CheckValues(prefix string) error {
	if s == nil {
		return nil
	}

	itemPrefix := prefix + "  "
	var errs []error
	for index, urlCommand := range *s {
		util.AppendCheckError(&errs, prefix, fmt.Sprintf("- item_%d", index),
			urlCommand.CheckValues(itemPrefix))
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of the URLCommand struct.
func (c *URLCommand) CheckValues(prefix string) error {
	if !util.Contains(urlCommandTypes, c.Type) {
		return fmt.Errorf("%stype: %q <invalid> is not a valid url_command [regex, replace, split]",
			prefix, c.Type)
	}

	errs := []error{
		fmt.Errorf("%stype: %s",
			prefix, c.Type)}
	switch c.Type {
	case "regex":
		if c.Regex == "" {
			errs = append(errs, errors.New(prefix+
				"regex: <required> (regex to use)"))
		} else {
			_, err := regexp.Compile(c.Regex)
			if err != nil {
				errs = append(errs,
					fmt.Errorf("%sregex: %q <invalid> (Invalid RegEx)",
						prefix, c.Regex))
			}
		}
	case "replace":
		if c.Old == "" {
			errs = append(errs, errors.New(prefix+
				"old: <required> (text you want replaced)"))
		}
		if c.New == nil {
			c.New = new(string)
		}
	case "split":
		if c.Text == "" {
			errs = append(errs, errors.New(prefix+
				"text: <required> (text to split on)"))
		}
	}

	if len(errs) == 1 {
		return nil
	}
	return errors.Join(errs...)
}
