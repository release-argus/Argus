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

package test

import (
	"fmt"
	"regexp"
	"strings"
)

// RegexBracketEscaper is a replacer to escape regex brackets.
var RegexBracketEscaper = strings.NewReplacer(
	"[", `\[`,
	"]", `\]`,
	"(", `\(`,
	")", `\)`,
)

// Indent returns a string with lines indented by the given number of spaces.
func Indent(str string, indent int) string {
	lines := strings.Split(str, "\n")

	return strings.Join(lines, "\n"+strings.Repeat(" ", indent))
}

// TrimJSON removes unnecessary whitespace from a JSON string.
func TrimJSON(str string) string {
	replacer := strings.NewReplacer(
		"\n", "",
		"\t", "",
		`": `, `":`,
		`", `, `",`,
		`, "`, `,"`,
	)
	return replacer.Replace(strings.TrimSpace(str))
}

// TrimYAML removes unnecessary whitespace from a YAML string and converts leading tabs to spaces.
func TrimYAML(str string) string {
	return normaliseLeadingWhitespace(str, "\n")
}

type tLogger interface {
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Helper()
}

type FakeT struct {
	Errors []string
}

func (f *FakeT) Errorf(format string, args ...any) {
	f.Errors = append(f.Errors, fmt.Sprintf(format, args...))
}

func (f *FakeT) Fatalf(format string, args ...any) {
	f.Errors = append(f.Errors, fmt.Sprintf(format, args...))
}

func (f *FakeT) Helper() {
	// No-op.
}

// addPrefix adds the given prefix to each line of the input string.
func addPrefix(str, prefix string) string {
	str = strings.TrimPrefix(str, "\n")

	if str == "" {
		return ""
	}

	if str != "{}\n" {
		count := strings.Count(str, "\n")
		if strings.HasSuffix(str, "\n") {
			count--
		}
		str = prefix + strings.Replace(str, "\n", "\n"+prefix, count)
	}

	return str
}

// normaliseLeadingWhitespace strips leading whitespace from the first line of str and re-indents subsequent lines, joining with joinWith.
func normaliseLeadingWhitespace(str string, joinWith string) string {
	str = strings.TrimPrefix(str, "\n")
	lines := strings.Split(str, "\n")
	leadingWhitespaceRegEx := regexp.MustCompile(`^(\s*)`)
	fullWhitespaceRegEx := regexp.MustCompile(`^\s*$`)
	var whitespacePrefix string
	for i := range lines {
		if i != 0 {
			// Remove whitespacePrefix from the beginning of the line.
			lines[i] = strings.TrimPrefix(lines[i], whitespacePrefix)
		}

		leadingWhitespace := leadingWhitespaceRegEx.FindString(lines[i])

		if i == 0 {
			whitespacePrefix = leadingWhitespace
			lines[i] = strings.Replace(lines[i], leadingWhitespace, "", 1)
		} else if leadingWhitespace != "" && strings.Contains(leadingWhitespace, "\t") {
			// Empty the line if it contains only whitespace.
			if fullWhitespaceRegEx.MatchString(lines[i]) {
				lines[i] = ""
			} else {
				// Convert leading tabs to spaces.
				newWhitespace := strings.ReplaceAll(leadingWhitespace, "\t", "  ")
				lines[i] = strings.Replace(lines[i], leadingWhitespace, newWhitespace, 1)
			}
		}
	}

	return strings.Join(lines, joinWith)
}
