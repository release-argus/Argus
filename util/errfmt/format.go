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

// Package errfmt formats error chains for display to users.
package errfmt

import (
	"errors"
	"strings"
)

// FormatError renders err and its wrapped errors as an indented multi-line string.
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	lines := make([]string, 0, 4)
	lines = appendFormattedErrorLines(err, lines, 0)
	return strings.Join(lines, "\n")
}

// appendFormattedErrorLines appends formatted lines for err and its unwrap chain to lines.
func appendFormattedErrorLines(err error, lines []string, indents int) []string {
	if err == nil {
		return lines
	}

	if errs, ok := err.(interface{ Unwrap() []error }); ok {
		for _, child := range errs.Unwrap() {
			lines = appendFormattedErrorLines(
				child,
				lines,
				indents,
			)
		}
		return lines
	}

	child := errors.Unwrap(err)

	msg := err.Error()
	if child != nil {
		childMsg := " " + child.Error()
		msg = strings.TrimSuffix(msg, childMsg)
	}

	indentation := strings.Repeat("  ", indents)
	msg = strings.TrimSuffix(msg, "\n")
	lines = append(
		lines,
		indentation+strings.ReplaceAll(
			msg,
			"\n",
			"\n"+indentation,
		),
	)

	if child != nil {
		lines = appendFormattedErrorLines(
			child,
			lines,
			indents+1,
		)
	}
	return lines
}
