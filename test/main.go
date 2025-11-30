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

//go:build unit || integration

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

var StdoutMutex sync.Mutex // Only one test should write to stdout at a time.
// CaptureStdout temporarily captures all output written to the standard output
// and returns a function that, when called, restores the original standard output and
// returns the captured output as a string.
func CaptureStdout() func() string {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	StdoutMutex.Lock()

	var buf bytes.Buffer
	done := make(chan struct{})

	// Drain the pipe until closed.
	go func() {
		io.Copy(&buf, r)
		close(done)
	}()

	return func() string {
		w.Close()
		<-done

		os.Stdout = stdout
		StdoutMutex.Unlock()
		return buf.String()
	}
}

// BoolPtr returns a pointer to the given boolean value.
func BoolPtr(val bool) *bool { return &val }

// IntPtr returns a pointer to the given integer value.
func IntPtr(val int) *int { return &val }

// StringPtr returns a pointer to the given string value.
func StringPtr(val string) *string { return &val }

// StringSlicePtr returns a pointer to the given string slice.
func StringSlicePtr(val []string) *[]string { return &val }

// UInt8Ptr returns a pointer to the given unsigned integer value.
func UInt8Ptr(val int) *uint8 {
	converted := uint8(val)
	return &converted
}

// UInt16Ptr returns a pointer to the given unsigned integer value.
func UInt16Ptr(val int) *uint16 {
	converted := uint16(val)
	return &converted
}

// StringifyPtr returns a string representation of the given pointer.
func StringifyPtr[T comparable](ptr *T) string {
	str := "<nil>"
	if ptr != nil {
		str = fmt.Sprint(*ptr)
	}
	return str
}

// TrimJSON removes unnecessary whitespace from a JSON string.
func TrimJSON(str string) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(str)); err != nil {
		// Return original string if invalid JSON.
		return str
	}
	return buf.String()
}

// TrimYAML removes unnecessary whitespace from a YAML string.
// and converts leading tabs to spaces.
func TrimYAML(str string) string {
	return normaliseLeadingWhitespace(str, "\n")
}

// FlattenMultilineString takes in a string with \n and \t characters, and returns a string with those characters
// replaced with spaces.
func FlattenMultilineString(str string) string {
	return normaliseLeadingWhitespace(str, " ")
}

// normaliseLeadingWhitespace normalises the leading whitespace of each line in the given string.
// It trims any leading newline character from the input string, splits the string into lines,
// and processes each line to remove or replace leading whitespace. It then joins the lines back
// together with the given `joinWith` string.
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

// Combinations generates all possible combinations of the given input.
func Combinations[T comparable](input []T) [][]T {
	var result [][]T

	var generate func(index int, current []T)
	generate = func(index int, current []T) {
		if index == len(input) {
			return
		}

		for i := index; i < len(input); i++ {
			newCombination := append([]T{}, current...)
			newCombination = append(newCombination, input[i])
			result = append(result, newCombination)
			generate(i+1, newCombination)
		}
	}

	generate(0, []T{})
	return result
}

// Indent returns a string with lines indented by the given amount of spaces.
func Indent(str string, indent int) string {
	lines := strings.Split(str, "\n")

	return strings.Join(lines, "\n"+strings.Repeat(" ", indent))
}

// IgnoreError calls the given function and returns the result, ignoring any error.
func IgnoreError[T any](t *testing.T, fn func() (T, error)) T {
	result, err := fn()
	if err != nil {
		var bare T
		t.Logf("unexpected error: %v", err)
		return bare
	}

	return result
}

func YAMLToNode(t *testing.T, yamlStr string) (*yaml.Node, error) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlStr), &node)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func EqualSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
