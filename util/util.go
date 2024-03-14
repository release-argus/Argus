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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"sort"
	"strconv"
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

// FirstNonNilPtrWithEnv will return the first pointer in `pointers` that is not nil after evaluating any environment variables.
func FirstNonNilPtrWithEnv(pointers ...*string) *string {
	for _, pointer := range pointers {
		if pointer != nil {
			if val := EvalEnvVars(*pointer); val != *pointer {
				return &val
			}
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

// FirstNonDefaultWithEnv will return the first var in `vars` that is not an empty string after evaluating any environment variables. "" is returned if all are empty.
func FirstNonDefaultWithEnv(vars ...string) string {
	for _, v := range vars {
		if v = EvalEnvVars(v); v != "" {
			return v
		}
	}
	return ""
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

// CopyIfSecret will loop through 'fields' and replace values in 'to' of '<secret>' with values in 'from' if they are non-empty.
func CopyIfSecret(from, to map[string]string, fields []string) {
	for _, field := range fields {
		if to[field] == "<secret>" && from[field] != "" {
			to[field] = from[field]
		}
	}
}

// InitMap will initialise the map if it's nil.
func InitMap(m *map[string]string) {
	if *m == nil {
		*m = make(map[string]string)
	}
}

// CopyMap will return a copy of the map
func CopyMap[T, Y comparable](m map[T]Y) map[T]Y {
	m2 := make(map[T]Y, len(m))
	for key := range m {
		m2[key] = m[key]
	}
	return m2
}

// MergeMaps will merge `m2` into `m1` and any fields in `fields` that are '<secret>' will be replaced with the value in `m2`.
func MergeMaps(m1, m2 map[string]string, fields []string) (m3 map[string]string) {
	m3 = CopyMap(m1)
	for k, v := range m2 {
		m3[k] = v
	}
	CopyIfSecret(m1, m3, fields)
	return
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

// isHashed will return whether the string is a hashed value.
func isHashed(s string) bool {
	return RegexCheck("^h__[a-f0-9]{64}$", s)
}

// Hash will return the SHA256 hash of the string.
func hash(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}

// HashFromString will return the byte slice of the hash string.
func hashFromString(s string) []byte {
	hash, _ := hex.DecodeString(s)
	return hash
}

// GetHash will return the SHA256 hash of the string. If it's already hashed, the string hash is converted to a byte slice.
func GetHash(s string) [32]byte {
	if isHashed(s) {
		hash := hashFromString(s[3:])
		var hash32 [32]byte
		copy(hash32[:], hash)
		return hash32
	}
	return hash(s)
}

func FmtHash(h [32]byte) string {
	return fmt.Sprintf("h__%x", h)
}

// envVarRegex is a regular expression to match environment variables.
var envVarRegex = regexp.MustCompile(`\$\{([a-zA-Z]\w*)\}`)

// envReplaceFunc is a function to replace environment variables in a string.
func envReplaceFunc(match string) string {
	envVarName := match[2 : len(match)-1] // Remove the '${' and '}'.
	if value, ok := os.LookupEnv(envVarName); ok {
		return value
	}
	return match
}

// EvalEnvVars will evaluate the environment variables in the string.
func EvalEnvVars(input string) string {
	// May contain an environment variable.
	if strings.Contains(input, "${") {
		return envVarRegex.ReplaceAllStringFunc(input, envReplaceFunc)
	}
	// No environment variables.
	return input
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

// ParseKeys will return the JSON keys in the string.
func ParseKeys(key string) (keys []interface{}, err error) {
	// Split the key into individual components
	// Example: "foo.bar[1].bash" => ["foo", "bar", "1", "bash"]
	keyCount := strings.Count(key, ".") + strings.Count(key, "[")
	keys = make([]interface{}, 0, keyCount+1)
	keyStrLength := len(key)
	i := 0

	for i < keyStrLength {
		switch key[i] {
		case '.':
			// Handle dot notation
			i++
		case '[':
			// Handle array notation
			i++
			start := i
			for i < keyStrLength && key[i] != ']' {
				i++
			}
			index := key[start:i]
			var intIndex int
			intIndex, err = strconv.Atoi(index)
			if err != nil {
				err = fmt.Errorf("failed to parse index %q in %q",
					index, key)
				return
			}

			keys = append(keys, intIndex)
			i++
		default:
			// Handle regular key
			start := i
			for i < keyStrLength && key[i] != '.' && key[i] != '[' {
				i++
			}

			keys = append(keys, key[start:i])
		}
	}

	return
}

func navigateJSON(jsonData *interface{}, fullKey string) (jsonValue string, err error) {
	if fullKey == "" {
		return "", fmt.Errorf("no key was given to navigate the JSON")
	}
	//nolint:errcheck // Verify in deployed_version.verify.CheckValues
	keys, _ := ParseKeys(fullKey)
	keyCount := len(keys)
	keyIndex := 0
	parsedJSON := *jsonData
	for keyIndex < keyCount {
		key := keys[keyIndex]
		switch value := parsedJSON.(type) {
		// Regular key
		case map[string]interface{}:
			// Ensure key is a string
			keyStr, ok := key.(string)
			if !ok {
				err = fmt.Errorf("got a map, but the key is not a string: %q at %v",
					key, parsedJSON)
				return
			}
			parsedJSON = value[keyStr]
		// Array
		case []interface{}:
			// Parse the index from the key.
			index, ok := key.(int)
			fmt.Printf("index: %v, ok: %t, key: %v\n", index, ok, key)
			if !ok {
				err = fmt.Errorf("got an array, but the key is not an integer index: %q at %v",
					key, parsedJSON)
				return
			}
			// Negative index
			if index < 0 {
				index = len(value) + index
			}

			// Check if the index is out of range.
			if index >= len(value) || index < 0 {
				err = fmt.Errorf("index %d (%s) out of range at %v",
					index, fullKey, parsedJSON)
				return
			}

			parsedJSON = value[index]
		// If the value is a string, int, float32, or float64, we can't navigate further.
		case string, int, float32, float64:
			err = fmt.Errorf("got a value of %q at %q, but there are more keys to navigate: %s at %v",
				value, key, fullKey, parsedJSON)
			return
		}
		keyIndex++
	}

	// If type is string, int, float32, or float64, we've found the value.
	switch v := parsedJSON.(type) {
	case string, int, float32, float64:
		jsonValue = fmt.Sprint(v)
		return
	}

	// If we got here, we didn't get a value.
	err = fmt.Errorf("failed to find value for %q in %v",
		fullKey, *jsonData)
	return
}

// GetValueByKey will return the value of the key in the JSON.
func GetValueByKey(rawBody []byte, key string, jsonFrom string) (string, error) {
	// If the key is empty, return the stringified body.
	if key == "" {
		return string(rawBody), nil
	}

	var jsonData interface{}
	err := json.Unmarshal(rawBody, &jsonData)
	// If the JSON is invalid, return an error.
	if err != nil {
		err := fmt.Errorf("failed to unmarshal the following from %q into json: %q",
			jsonFrom, string(rawBody))
		return "", err
	}

	return navigateJSON(&jsonData, key)
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
