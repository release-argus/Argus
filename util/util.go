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

// Package util provides utility functions for the Argus project.
package util

import (
	"bytes"
	"strings"

	"gopkg.in/yaml.v3"
)

// SecretValue defines the value used to represent a secret.
var SecretValue = "<secret>"

// Field is a helper struct for String() methods.
type Field struct {
	Name  string
	Value interface{}
}

// DereferenceOrNilValue will return the value of pointer if non-nil, otherwise nilValue.
func DereferenceOrNilValue[T comparable](pointer *T, nilValue T) T {
	if pointer == nil {
		return nilValue
	}
	return *pointer
}

// StringToBoolPtr will take a string and convert it to a boolean pointer.
//
//	"" => nil
//	"true" => true
//	"false" => false
func StringToBoolPtr(str string) *bool {
	if str == "" {
		return nil
	}
	val := str == "true"
	return &val
}

// ValueUnlessDefault returns `value` when `check` differs from the default value for its type,
// otherwise default.
func ValueUnlessDefault[T comparable](check T, value T) T {
	var fresh T
	if check == fresh {
		return check
	}
	return value
}

// ValueOrValue returns `first` if it differs from the default,
// otherwise `second`.
func ValueOrValue[T comparable](first T, second T) T {
	var fresh T
	if first != fresh {
		return first
	}
	return second
}

// DereferenceOrDefault returns the value of `check` if not nil,
// otherwise default for the type.
func DereferenceOrDefault[T comparable](check *T) T {
	if check == nil {
		return *new(T)
	}
	return *check
}

// DereferenceOrValue returns the default of `check` if nil,
// otherwise `value`.
func DereferenceOrValue[T comparable](check *T, value T) T {
	if check == nil {
		return *new(T)
	}
	return value
}

type customComparable interface {
	bool | int | map[string]string | string | uint8 | uint16
}

// FirstNonNilPtr will return the first non-nil pointer in `pointers`.
func FirstNonNilPtr[T customComparable](pointers ...*T) *T {
	for _, pointer := range pointers {
		if pointer != nil {
			return pointer
		}
	}
	return nil
}

// FirstNonDefault will return the first non-default var in `vars`.
func FirstNonDefault[T comparable](vars ...T) T {
	var fresh T
	for _, v := range vars {
		if v != fresh {
			return v
		}
	}
	return fresh
}

// PtrValueOrValue will return the value of `ptr` if non-nil, otherwise `fallback`.
func PtrValueOrValue[T comparable](ptr *T, fallback T) T {
	if ptr != nil {
		return *ptr
	}
	return fallback
}

// CopyPointer will return a pointer to a copy of the value of `ptr`.
func CopyPointer[T comparable](ptr *T) *T {
	if ptr == nil {
		return nil
	}

	val := *ptr
	return &val
}

// NormaliseNewlines all newlines in `data` to \n.
func NormaliseNewlines(data []byte) []byte {
	// replace CR LF \r\n (Windows) with LF \n (Unix).
	data = bytes.ReplaceAll(data, []byte{13, 10}, []byte{10})
	// replace CF \r (Mac) with LF \n (Unix).
	data = bytes.ReplaceAll(data, []byte{13}, []byte{10})

	return data
}

// CopySecretValues loops through 'fields' and replace values in 'to' of 'SecretValue' with values in 'from'.
// if non-empty.
func CopySecretValues(from, to map[string]string, fields []string) {
	for _, field := range fields {
		if to[field] == SecretValue && from[field] != "" {
			to[field] = from[field]
		}
	}
}

// ToYAMLString returns a YAML string representation of `input`.
func ToYAMLString(input interface{}, prefix string) string {
	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	defer enc.Close()

	err := enc.Encode(input)
	if err != nil {
		return ""
	}

	str := buf.String()

	// Add prefix to each line.
	if prefix != "" && str != "" && str != "{}\n" {
		str = strings.Replace(str, "\n", "\n"+prefix,
			strings.Count(str, "\n")-1)
		str = prefix + str
	}

	return str
}

// Indentation returns the indentation of a given line based on the specified indent size.
func Indentation(line string, indentSize uint8) string {
	indent := strings.Repeat(" ", int(indentSize))

	var count int
	for strings.HasPrefix(line, strings.Repeat(indent, count+1)) {
		count++
	}

	return strings.Repeat(indent, count)
}
