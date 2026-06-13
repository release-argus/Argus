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

// Package util provides utility functions for the Argus project.
package util

// SecretValue defines the value used to represent a secret.
const SecretValue = "<secret>"

// Field is a helper struct for String() methods.
type Field struct {
	Name  string
	Value any
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

// ValueUnlessDefault returns `value` when `condition` differs from the default value for its type,
// otherwise default.
func ValueUnlessDefault[T comparable](condition T, value T) T {
	var zero T
	if condition == zero {
		return condition
	}
	return value
}

// ValueOr returns `first` if it differs from the default,
// otherwise `second`.
func ValueOr[T comparable](first T, second T) T {
	var zero T
	if first != zero {
		return first
	}
	return second
}

// DerefOrZero returns the value of `ptr` if not nil,
// otherwise default for the type.
func DerefOrZero[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

// DerefOr returns the value of 'ptr' if non-nil,
// otherwise the 'fallback'.
func DerefOr[T any](ptr *T, fallback T) T {
	if ptr != nil {
		return *ptr
	}
	return fallback
}

// PtrIfNotZero returns a pointer to `v` if `v` is not the zero value.
// Otherwise it returns nil.
func PtrIfNotZero[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

// ClonePtr returns a pointer to a copy of the value of `ptr`.
func ClonePtr[T any](ptr *T) *T {
	if ptr == nil {
		return nil
	}

	val := *ptr
	return &val
}

// RestoreMaskedValues loops through 'fields' and replaces values in 'to' of 'SecretValue' with values in 'from'
// if non-empty.
func RestoreMaskedValues[K comparable](original, target map[K]string, fields []K) map[K]string {
	for _, field := range fields {
		if target[field] == SecretValue && original[field] != "" {
			target[field] = original[field]
		}
	}

	return target
}

// Indentation returns the indentation of a given line based on the specified indent size.
func Indentation(line string, indentSize uint8) string {
	if indentSize == 0 {
		return ""
	}

	prefix := 0

	for prefix < len(line) && line[prefix] == ' ' {
		prefix++
	}

	prefix = (prefix / int(indentSize)) * int(indentSize)

	return line[:prefix]
}

// TruncateMessage shortens a message to `maxLength` and appends "..." if it exceeds the limit.
func TruncateMessage(msg string, maxLength int) string {
	if len(msg) > maxLength {
		return msg[:maxLength] + "..."
	}
	return msg
}
