// Copyright [2024] [Argus]
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
	"io"
	"os"
	"strings"
	"sync"
)

var StdoutMutex sync.Mutex // Only one test should write to stdout at a time
func CaptureStdout() func() string {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	StdoutMutex.Lock()
	return func() string {
		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = stdout
		StdoutMutex.Unlock()
		return string(out)
	}
}

// BoolPtr returns a pointer to the given boolean value
func BoolPtr(val bool) *bool {
	return &val
}

// IntPtr returns a pointer to the given integer value
func IntPtr(val int) *int {
	return &val
}

// StringPtr returns a pointer to the given string value
func StringPtr(val string) *string {
	return &val
}

// UIntPtr returns a pointer to the given unsigned integer value
func UIntPtr(val int) *uint {
	converted := uint(val)
	return &converted
}

// StringifyPtr returns a string representation of the given pointer
func StringifyPtr[T comparable](ptr *T) string {
	str := "nil"
	if ptr != nil {
		str = fmt.Sprint(*ptr)
	}
	return str
}

// TrimJSON removes unnecessary whitespace from a JSON string
func TrimJSON(str string) string {
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, `": `, `":`)
	str = strings.ReplaceAll(str, `", `, `",`)
	str = strings.ReplaceAll(str, `, "`, `,"`)
	return str
}

// Combinations generates all possible combinations of the given input
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
