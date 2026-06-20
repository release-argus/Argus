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

package decode

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// JSONMarshalOpts are the options used when encoding JSON via [Marshal].
var JSONMarshalOpts = []json.Options{
	json.Deterministic(true),
}

// ParseKeys returns the JSON keys in the string.
func ParseKeys(key string) ([]any, error) {
	// Split the key into individual components.
	// e.g. "foo.bar[1].bash" => ["foo", "bar", "1", "bash"]
	keyCount := strings.Count(key, ".") + strings.Count(key, "[")
	keys := make([]any, 0, keyCount+1)
	keyStrLength := len(key)
	i := 0

	for i < keyStrLength {
		switch key[i] {
		case '.':
			// Handle dot notation.
			i++
		case '[':
			// Handle array notation.
			i++
			start := i
			for i < keyStrLength && key[i] != ']' {
				i++
			}
			index := key[start:i]
			intIndex, err := strconv.Atoi(index)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to parse index %q in %q",
					index, key,
				)
			}

			keys = append(keys, intIndex)
			i++
		default:
			// Handle regular key.
			start := i
			for i < keyStrLength && key[i] != '.' && key[i] != '[' {
				i++
			}

			keys = append(keys, key[start:i])
		}
	}

	return keys, nil
}

// navigateJSON walks jsonData using dot and bracket notation in fullKey.
// It supports object lookups, array indexing (including negative indices),
// and returns the final scalar value as a string.
func navigateJSON(jsonData *any, fullKey string) (string, error) {
	if fullKey == "" {
		return "", errors.New("no key was given")
	}

	//nolint:errcheck // Verify in deployed_version.verify.CheckValues.
	keys, _ := ParseKeys(fullKey)
	keyCount := len(keys)
	parsedJSON := *jsonData
	for keyIndex := range keyCount {
		key := keys[keyIndex]
		moreKeys := keyIndex < keyCount-1

		switch value := parsedJSON.(type) {
		case map[string]any:
			// Object traversal requires a string key.
			keyStr, ok := key.(string)
			if !ok {
				return "", fmt.Errorf(
					"got a map, but the wanted key is not a string: %v (%T) at %v",
					key, key, parsedJSON,
				)
			}

			next, ok := value[keyStr]
			if !ok {
				return "", fmt.Errorf(
					"failed to find value for %q in %v",
					fullKey, *jsonData,
				)
			}

			// A null value is only valid if this is the final path element.
			if next == nil && moreKeys {
				return "", fmt.Errorf(
					"got null at %q while navigating %q",
					keyStr, fullKey,
				)
			}

			parsedJSON = next

		case []any:
			// Array traversal requires an integer index.
			index, ok := key.(int)
			if !ok {
				return "", fmt.Errorf(
					"got an array, but the key is not an integer index: %v at %v",
					key, parsedJSON,
				)
			}

			// Negative index.
			if index < 0 {
				index = len(value) + index
			}

			if index >= len(value) || index < 0 {
				return "", fmt.Errorf(
					"index %d (%s) out of range at %v",
					index, fullKey, parsedJSON,
				)
			}

			next := value[index]

			// A null value is only valid if this is the final path element.
			if next == nil && moreKeys {
				return "", fmt.Errorf(
					"got null at index %d while navigating %q",
					index, fullKey,
				)
			}

			parsedJSON = next

		case string, int, float32, float64, bool:
			// Primitive values cannot be traversed further.
			return "", fmt.Errorf(
				"failed to find key %q while navigating %q: %v is not an object or array",
				key, fullKey, parsedJSON,
			)

		case nil:
			return "", fmt.Errorf(
				"got null at %q while navigating %q",
				key, fullKey,
			)

		default:
			return "", fmt.Errorf(
				"got unsupported type %T at %q while navigating %q: %v",
				parsedJSON, key, fullKey, parsedJSON,
			)
		}
	}

	// A final null value is treated as an empty string with no error.
	if parsedJSON == nil {
		return "", nil
	}

	switch v := parsedJSON.(type) {
	case string, int, float32, float64, bool:
		return fmt.Sprint(v), nil
	}

	// Only scalar values can be returned.
	return "", fmt.Errorf(
		"failed to find value for %q in %v",
		fullKey, *jsonData,
	)
}

// GetValueByKey returns the value of the key in the JSON.
func GetValueByKey(body []byte, key string, jsonFrom string) (string, error) {
	// If the key is empty, return the body as a string.
	if key == "" {
		return string(body), nil
	}

	var jsonData any
	err := Unmarshal("json", body, &jsonData)
	// If the JSON proves invalid, return an error.
	if err != nil {
		return "", fmt.Errorf(
			"failed to unmarshal response from %q into JSON: %w",
			jsonFrom, err,
		)
	}

	value, err := navigateJSON(&jsonData, key)
	if err != nil {
		return "", fmt.Errorf("failed to navigate JSON: %w", err)
	}

	return value, nil
}

// ToJSONString converts input to its JSON string representation.
func ToJSONString(input any) string {
	bytes, err := Marshal("json", input)
	if err != nil {
		return ""
	}
	return string(bytes)
}
