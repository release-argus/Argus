// Copyright [2023] [Argus]
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

package util

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Field is a helper struct for String() methods.
type Field struct {
	Name  string
	Value interface{}
}

// Contains returns whether `s` contains `e`
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
// `a` when it isn't nil.
func PtrOrValueToPtr[T comparable](a *T, b T) *T {
	if a == nil {
		return &b
	}
	return a
}

// StringToBoolPtr will take a string and convert it to a boolean pointer
//
// "" => nil
//
// "true" => true
//
// "false" => false
func StringToBoolPtr(str string) *bool {
	if str == "" {
		return nil
	}
	val := str == "true"
	return &val
}

// ValueIfNotNil will take the `check` pointer and return address of `value`
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

type customComparable interface {
	bool | int | map[string]string | string | uint
}

// FirstNonNilPtr will return the first pointer in `pointers` that is not nil.
func FirstNonNilPtr[T customComparable](pointers ...*T) *T {
	for _, pointer := range pointers {
		if pointer != nil {
			return pointer
		}
	}
	return nil
}

// FirstNonDefault will return the first var in `vars` that is not the default.
func FirstNonDefault[T comparable](vars ...T) T {
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

// PtrValueOrValue will return the value of `ptr` if it's non-nil, otherwise `fallback`.
func PtrValueOrValue[T comparable](ptr *T, fallback T) T {
	if ptr != nil {
		return *ptr
	}
	return fallback
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
	data = bytes.ReplaceAll(data, []byte{13, 10}, []byte{10})
	// replace CF \r (Mac) with LF \n (Unix)
	data = bytes.ReplaceAll(data, []byte{13}, []byte{10})

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

// LowercaseStringStringMap will convert all lowercase all keys in the map
func LowercaseStringStringMap(change *map[string]string) (lowercasedMap map[string]string) {
	lowercasedMap = make(map[string]string, len(*change))
	for i := range *change {
		lowercasedMap[strings.ToLower(i)] = (*change)[i]
	}
	return
}

// Sorted keys will return a sorted list of the keys in a map.
func SortedKeys[V any](m map[string]V) (keys []string) {
	keys = make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return
}

// StringToPointer will return a pointer to str, but nil if it's an empty string.
func StringToPointer(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

func BasicAuth(username string, password string) string {
	encode := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(encode))
}

func GetKeysFromJSON(data string) []string {
	return getKeysFromJSONBytes([]byte(data), "")
}

func getKeysFromJSONBytes(data []byte, prefix string) (keys []string) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		// Unmarshaling fail
		return []string{}
	}
	keys = make([]string, len(obj))

	// Iterate over the JSON object
	index := 0
	for key, value := range obj {
		// Add the key to the list
		fullKey := prefix + key
		keys[index] = fullKey
		index++

		// If value is a JSON object, recursively get its keys
		if bytes.HasPrefix(value, []byte("{")) {
			subKeys := getKeysFromJSONBytes(value, fullKey+".")
			keys = append(keys, subKeys...)
		}
	}
	// sort keys
	sort.Strings(keys)
	return
}

// ToYAMLString will return a YAML string representation of the interface.
func ToYAMLString(iface interface{}, prefix string) (str string) {
	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	defer enc.Close()

	err := enc.Encode(iface)
	if err != nil {
		return
	}
	str = buf.String()

	if prefix != "" && str != "" && str != "{}\n" {
		str = strings.Replace(str, "\n", "\n"+prefix,
			strings.Count(str, "\n")-1)
		str = prefix + str
	}

	return
}

// ToJSONString will return a JSON string representation of the interface.
func ToJSONString(iface interface{}) (str string) {
	bytes, err := json.Marshal(iface)
	if err != nil {
		return
	}
	str = string(bytes)

	return
}
