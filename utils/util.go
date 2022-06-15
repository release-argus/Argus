// Copyright [2022] [Argus]
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

package utils

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// Contains returns whether `s` contains `e``
func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// EvalNilPtr - Return the value of pointer if it's non-nil, otherwise nilValue.
func EvalNilPtr[T comparable](pointer *T, nilValue T) T {
	if pointer == nil {
		return nilValue
	}
	return *pointer
}

// PtrOrValueToPtr will take the pointer `a` and the value `b`, returning
// the first that isn't nil.
func PtrOrValueToPtr[T comparable](a *T, b T) *T {
	if a == nil {
		return &b
	}
	return a
}

// ValueIfNotNil will take the `check` pointer and return `value`
// when `check` is not nil.
func ValueIfNotNil[T comparable](check *T, value T) *T {
	if check == nil {
		return nil
	}
	return &value
}

// ValueIfNotDefault will take the `check` var and return `value`
// when `check` is not it's default.
func ValueIfNotDefault[T comparable](check T, value T) T {
	var fresh T
	if check == fresh {
		return check
	}
	return value
}

// ValueIfNotNil will take the `check` pointer and return the default
// value of that type if `check` is nil.
func DefaultIfNil[T comparable](check *T) T {
	if check == nil {
		var fresh T
		return fresh
	}
	return *check
}

// GetFirstNonNilPtr will return the first pointer in `pointers` that is not nil.
func GetFirstNonNilPtr[T comparable](pointers ...*T) *T {
	for _, pointer := range pointers {
		if pointer != nil {
			return pointer
		}
	}
	return nil
}

// GetFirstNonDefault will return the first var in `vars` that is not the default.
func GetFirstNonDefault[T comparable](vars ...T) T {
	var fresh T
	for _, v := range vars {
		if v != fresh {
			return v
		}
	}
	return fresh
}

// PrintlnIfNotDefault will print `msg` is `x` is not the default for that type.
func PrintlnIfNotDefault[T comparable](x T, msg string) {
	var fresh T
	if x != fresh {
		fmt.Println(msg)
	}
}

// PrintlnIfNotNil will print `msg` is `ptr` is not nil.
func PrintlnIfNotNil[T comparable](ptr *T, msg string) {
	if ptr != nil {
		fmt.Println(msg)
	}
}

// PrintlnIfNil will print `msg` is `ptr` is nil.
func PrintlnIfNil[T comparable](ptr *T, msg string) {
	if ptr == nil {
		fmt.Println(msg)
	}
}

// DefaultOrValue will return the default of `check` if it's nil, otherwise value
func DefaultOrValue[T comparable](check *T, value T) T {
	if check == nil {
		var fresh T
		return fresh
	}
	return value
}

// ErrorToString accounts for nil errors, returning an empty string for those
// and err.Error() for non-nil errors.
func ErrorToString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

const alphanumericLower = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandAlphaNumericLower will return a random alphanumeric (lowercase) string of length n.
func RandAlphaNumericLower(n int) string {
	return RandString(n, alphanumericLower)
}

const numeric = "0123456789"

// RandNumeric will return a random numeric string of length n.
func RandNumeric(n int) string {
	return RandString(n, numeric)
}

// RandString will make a random string of length n with alphabet.
func RandString(n int, alphabet string) string {
	b := make([]byte, n)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		b[i] = alphabet[int(n.Int64())]
	}
	return string(b)
}

// NormaliseNewlines all newlines in `data` to \n.
func NormaliseNewlines(data []byte) []byte {
	// replace CR LF \r\n (Windows) with LF \n (Unix)
	data = bytes.Replace(data, []byte{13, 10}, []byte{10}, -1)
	// replace CF \r (Mac) with LF \n (Unix)
	data = bytes.Replace(data, []byte{13}, []byte{10}, -1)

	return data
}

// CopyMap will return a copy of the map
func CopyMap[T, Y comparable](m map[T]Y) map[T]Y {
	m2 := make(map[T]Y, len(m))
	for key := range m {
		m2[key] = m[key]
	}
	return m2
}

// GetPortFromURL extracts PORT from https://HOST:PORT
// and uses http/https defaults if there is no port specified
//
// If none of the above returns anything, return defaultPort
func GetPortFromURL(url string, defaultPort string) (convertedPort string) {
	if strings.HasPrefix(url, "https") {
		convertedPort = "443"
		url = strings.TrimPrefix(url, "https://")
	} else {
		convertedPort = "80"
		url = strings.TrimPrefix(url, "http://")
	}

	url = strings.Split(url, "/")[0]
	if strings.Contains(url, ":") {
		convertedPort = strings.Split(url, ":")[1]
	}

	if convertedPort == "" {
		return defaultPort
	}
	return convertedPort
}
