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

import "strings"

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

	msg := err.Error()
	children := unwrapAll(err)
	childMsgs := make([]string, len(children))
	for i, child := range children {
		childMsgs[i] = child.Error()
	}

	switch {
	// errors.Join, or a bare %w, adds no text of its own.
	case len(children) > 0 && msg == strings.Join(childMsgs, "\n"):
		for _, child := range children {
			lines = appendFormattedErrorLines(child, lines, indents)
		}
		return lines
	case len(children) == 1 && strings.HasSuffix(msg, " "+childMsgs[0]):
		lines = appendMessage(lines, strings.TrimSuffix(msg, " "+childMsgs[0]), indents)
		return appendFormattedErrorLines(children[0], lines, indents+1)
	// A leading wrapped error heads its own line, and the text after it goes underneath.
	case len(children) == 2 && strings.HasPrefix(msg, childMsgs[0]+": "):
		lines = appendMessage(lines, childMsgs[0]+":", indents)
		return appendMessage(lines, strings.TrimPrefix(msg, childMsgs[0]+": "), indents+1)
	}

	lines = appendMessage(lines, msg, indents)
	return lines
}

// unwrapAll returns the errors that err wraps directly.
func unwrapAll(err error) []error {
	switch e := err.(type) { //nolint:errorlint // Only err's own Unwrap, not its chain.
	case interface{ Unwrap() []error }:
		return e.Unwrap()
	case interface{ Unwrap() error }:
		if child := e.Unwrap(); child != nil {
			return []error{child}
		}
	}
	return nil
}

// appendMessage appends msg to lines, indenting every line of it.
func appendMessage(lines []string, msg string, indents int) []string {
	indentation := strings.Repeat("  ", indents)
	msg = strings.TrimSuffix(msg, "\n")
	return append(lines, indentation+strings.ReplaceAll(msg, "\n", "\n"+indentation))
}
