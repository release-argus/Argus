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

// Package filter provides filtering for latest_version queries.
package filter

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
)

var urlCommandTypes = []string{"regex", "replace", "split"}

// URLCommands is a list of URLCommand that filter versions from the URL Content.
type URLCommands []URLCommand

// URLCommand is a command to filter versions from the URL body.
type URLCommand struct {
	Type     string `json:"type" yaml:"type"`                             // regex/replace/split.
	Regex    string `json:"regex,omitempty" yaml:"regex,omitempty"`       // regex: regexp.MustCompile(Regex).
	Text     string `json:"text,omitempty" yaml:"text,omitempty"`         // split: strings.Split(tgtString, "Text").
	Old      string `json:"old,omitempty" yaml:"old,omitempty"`           // replace: strings.ReplaceAll(tgtString, "Old", "New").
	New      string `json:"new,omitempty" yaml:"new,omitempty"`           // replace: strings.ReplaceAll(tgtString, "Old", "New").
	Index    *int   `json:"index,omitempty" yaml:"index,omitempty"`       // regex/split: re.FindAllString(URL_content, -1)[Index]  /  strings.Split("text")[Index].
	Template string `json:"template,omitempty" yaml:"template,omitempty"` // regex: template.
}

// ############
// # DECODING #
// ############

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// It supports both the canonical form:
//
//	[ { type: "split" } ]
//
// and shorthand:
//
//	{ type: "split" }
//
// The shorthand is converted to a single-element list.
func (s *URLCommands) UnmarshalJSON(data []byte) error {
	// Handle the case where data is a quoted JSON string (from web requests).
	if len(data) > 0 && data[0] == '"' && data[len(data)-1] == '"' {
		var jsonStr string
		if err := decode.Unmarshal("json", data, &jsonStr); err != nil {
			return err //nolint:wrapcheck
		}
		data = []byte(jsonStr)
	}

	return s.unmarshal("json", data)
}

// UnmarshalYAML implements yaml.Unmarshaler for URLCommands.
//
// It supports both the canonical form:
//
//	[ { type: "split" } ]
//
// and shorthand:
//
//	{ type: "split" }
//
// The shorthand is converted to a single-element list.
func (s *URLCommands) UnmarshalYAML(data []byte) error {
	return s.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (s *URLCommands) unmarshal(format string, data []byte) error {
	// Alias to avoid recursion.
	var multi []URLCommand
	if err := decode.Unmarshal(format, data, &multi); err == nil {
		*s = multi
		return nil
	}

	// Else, try to unmarshal as a single URLCommand.
	var single URLCommand
	err := decode.Unmarshal(format, data, &single)
	if err == nil {
		if !single.IsZero() {
			*s = []URLCommand{single}
		}
		return nil
	}

	return err //nolint:wrapcheck
}

// ##########
// # STATE #
// ##########

// IsZero implements the yaml.IsZeroer interface.
func (c *URLCommand) IsZero() bool {
	return c == nil || (c.Type == "" && c.Regex == "" &&
		c.Text == "" && c.Old == "" && c.New == "" &&
		c.Index == nil && c.Template == "")
}

// #############
// # STRINGIFY #
// #############

// String implements fmt.Stringer and returns a YAML representation.
func (s *URLCommands) String() string {
	if s == nil {
		return ""
	}
	return decode.ToYAMLString(s, "")
}

// String implements fmt.Stringer and returns a YAML representation.
func (c *URLCommand) String() string {
	if c == nil {
		return ""
	}
	return decode.ToYAMLString(c, "")
}

// ##############
// # VALIDATION #
// ##############

// CheckValues validates the fields of each [URLCommand].
func (s *URLCommands) CheckValues() error {
	if s == nil {
		return nil
	}

	var errs []error
	for index, urlCommand := range *s {
		if err := urlCommand.CheckValues(); err != nil {
			errs = append(
				errs,
				&decode.KeyFieldError{
					Key: fmt.Sprintf("- item_%d", index),
					Err: err,
				},
			)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// CheckValues validates the fields of the receiver.
func (c *URLCommand) CheckValues() error {
	if !util.Contains(urlCommandTypes, c.Type) {
		return polymorphic.InvalidTypeError{
			Key:     "type",
			Value:   c.Type,
			Allowed: urlCommandTypes,
		}
	}

	errs := []error{
		fmt.Errorf("type: %s", c.Type),
	}
	switch c.Type {
	case "regex":
		if c.Regex == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "regex",
					Description: "regular expression to use",
				},
			)
		} else {
			_, err := regexp.Compile(c.Regex)
			if err != nil {
				errs = append(
					errs,
					&decode.FieldError{
						Key:         "regex",
						Value:       c.Regex,
						Description: err.Error(),
					},
				)
			}
		}
	case "replace":
		if c.Old == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "old",
					Description: "text to replace",
				},
			)
		}
	case "split":
		if c.Text == "" {
			errs = append(
				errs,
				&decode.FieldError{
					Key:         "text",
					Description: "text to split on",
				},
			)
		}
	}

	if len(errs) == 1 {
		return nil
	}
	return errors.Join(errs...)
}

// ############
// # COMMANDS #
// ############

// GetVersions extracts version strings from text using the URLCommands.
func (s *URLCommands) GetVersions(text string, logFrom logx.LogFrom) ([]string, error) {
	// No URLCommands to run, so treat the text as a single version.
	if len(*s) == 0 {
		if text == "" {
			return nil, nil
		}
		return []string{text}, nil
	}
	return s.Run(text, logFrom)
}

// Run applies each URLCommand in sequence to text and returns the resulting version strings.
func (s *URLCommands) Run(text string, logFrom logx.LogFrom) ([]string, error) {
	if s == nil {
		return nil, nil
	}

	urlCommandLogFrom := logx.LogFrom{Primary: logFrom.Primary, Secondary: "url_commands"}
	versions := []string{text}
	var err error
	for _, urlCommand := range *s {
		versions, err = urlCommand.run(versions, urlCommandLogFrom)
		if err != nil {
			return nil, err
		}
	}
	return versions, nil
}

// run applies the URLCommand to each version in versions.
func (c *URLCommand) run(versions []string, logFrom logx.LogFrom) ([]string, error) {
	var err error

	for i, version := range versions {
		// Iterate through the commands to filter the text.
		if logx.IsLevel("DEBUG") {
			logx.Debug(
				fmt.Sprintf("Looking through:\n%q", version),
				logFrom,
				true,
			)
		}

		var msg string
		switch c.Type {
		case "split":
			msg = fmt.Sprintf(
				"Splitting on %q with index %d",
				c.Text, c.Index,
			)
			versions, err = c.split(i, versions, logFrom)
		case "replace":
			msg = fmt.Sprintf(
				"Replacing %q with %q",
				c.Old, c.New,
			)
			versions[i] = strings.ReplaceAll(version, c.Old, c.New)
		case "regex":
			msg = fmt.Sprintf("Regexing %q", c.Regex)
			if c.Template != "" {
				msg = fmt.Sprintf(
					"%s with template %q",
					msg, c.Template,
				)
			}
			versions, err = c.regex(i, versions, logFrom)
		}
		if err != nil {
			return nil, err
		}

		if logx.IsLevel("DEBUG") {
			msg = fmt.Sprintf(
				"%s\nResolved to %q",
				msg, version,
			)
			logx.Debug(msg, logFrom, true)
		}
	}
	return versions, nil
}

// regex applies the regex URLCommand to versions[versionIndex], expanding all matches.
func (c *URLCommand) regex(versionIndex int, versions []string, logFrom logx.LogFrom) ([]string, error) {
	re := regexp.MustCompile(c.Regex)

	version := versions[versionIndex]
	matches := re.FindAllStringSubmatch(version, -1)
	// No matches.
	if len(matches) == 0 {
		err := fmt.Errorf(
			"%s %q didn't return any matches on %q",
			c.Type, c.Regex, util.TruncateMessage(version, 50),
		)
		logx.Warn(err, logFrom, true)
		return nil, err
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
			err := fmt.Errorf(
				"%s (%s) returned %d elements on %q, but the index wants element number %d",
				c.Type, c.Regex, len(matches), version, index+1,
			)
			logx.Warn(err, logFrom, true)
			return nil, err
		}

		versions[len(versions)-1] = util.RegexTemplate(matches[index], c.Template)
		return versions, nil
	}

	// Add all subMatches to the versions list.
	subMatch := make([]string, len(matches))
	for i := range matches {
		subMatch[i] = util.RegexTemplate(matches[i], c.Template)
	}

	// Replace the current version in the list with the ordered subVersions.
	versions = util.SliceReplace(versions, versionIndex, subMatch)
	return versions, nil
}

// split applies the split URLCommand to the version at versionIndex, replacing it with the split result.
func (c *URLCommand) split(versionIndex int, versions []string, logFrom logx.LogFrom) ([]string, error) {
	texts, err := c.splitAllMatches(versions[versionIndex], logFrom)
	if err != nil {
		return nil, err
	}

	// If no index specified, replace versionIndex with the split text.
	if c.Index == nil {
		versions = util.SliceReplace(versions, versionIndex, texts)
		return versions, nil
	}

	index := *c.Index
	// Handle negative indices.
	if index < 0 {
		index = len(texts) + index
	}

	if (len(texts) - index) < 1 {
		err := fmt.Errorf(
			"%s (%q) returned %d elements on %q, but the index wants element number %d",
			c.Type, c.Text, len(texts), versions[versionIndex], index+1,
		)
		logx.Warn(err, logFrom, true)

		return nil, err
	}

	versions[versionIndex] = texts[index]
	return versions, nil
}

// splitAllMatches splits text on URLCommand.Text and returns all segments.
func (c *URLCommand) splitAllMatches(text string, logFrom logx.LogFrom) ([]string, error) {
	texts := strings.Split(text, c.Text)
	if len(texts) == 1 {
		err := fmt.Errorf(
			"%s didn't find any %q to split on",
			c.Type, c.Text,
		)
		logx.Warn(err, logFrom, true)

		return nil, err
	}
	return texts, nil
}
