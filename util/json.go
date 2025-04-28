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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ParseKeys will return the JSON keys in the string.
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
				return nil, fmt.Errorf("failed to parse index %q in %q",
					index, key)
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

// ToJSONString converts `input` to its JSON string representation.
func ToJSONString(input any) string {
	bytes, err := json.Marshal(input)
	if err != nil {
		return ""
	}
	return string(bytes)
}

// navigateJSON will navigate the JSON object to find the value of the key.
func navigateJSON(jsonData *any, fullKey string) (string, error) {
	if fullKey == "" {
		return "", errors.New("no key was given to navigate the JSON")
	}
	//nolint:errcheck // Verify in deployed_version.verify.CheckValues.
	keys, _ := ParseKeys(fullKey)
	keyCount := len(keys)
	keyIndex := 0
	parsedJSON := *jsonData
	for keyIndex < keyCount {
		key := keys[keyIndex]
		switch value := parsedJSON.(type) {
		// Regular key.
		case map[string]any:
			// Ensure key represents a string.
			keyStr, ok := key.(string)
			if !ok {
				return "", fmt.Errorf("got a map, but the key is not a string: %q at %v",
					key, parsedJSON)
			}
			parsedJSON = value[keyStr]
		// Array.
		case []any:
			// Parse the index from the key.
			index, ok := key.(int)
			fmt.Printf("index: %v, ok: %t, key: %v\n",
				index, ok, key)
			if !ok {
				return "", fmt.Errorf("got an array, but the key is not an integer index: %q at %v",
					key, parsedJSON)
			}
			// Negative index.
			if index < 0 {
				index = len(value) + index
			}

			// Check if the index falls out of range.
			if index >= len(value) || index < 0 {
				return "", fmt.Errorf("index %d (%s) out of range at %v",
					index, fullKey, parsedJSON)
			}

			parsedJSON = value[index]
		// If the value is a string, int, float32, or float64, we can't navigate further.
		case string, int, float32, float64:
			return "", fmt.Errorf("got a value of %q at %q, but there are more keys to navigate: %s at %v",
				value, key, fullKey, parsedJSON)
		}
		keyIndex++
	}

	// If type is string, int, float32, or float64, we have found the value.
	switch v := parsedJSON.(type) {
	case string, int, float32, float64:
		return fmt.Sprint(v), nil
	}

	// If we got here, we didn't get a value.
	return "", fmt.Errorf("failed to find value for %q in %v",
		fullKey, *jsonData)
}

// GetValueByKey will return the value of the key in the JSON.
func GetValueByKey(body []byte, key string, jsonFrom string) (string, error) {
	// If the key is empty, return the body as a string.
	if key == "" {
		return string(body), nil
	}

	var jsonData any
	err := json.Unmarshal(body, &jsonData)
	// If the JSON proves invalid, return an error.
	if err != nil {
		err := fmt.Errorf("failed to unmarshal the following from %q into json: %q",
			jsonFrom, TruncateMessage(string(body), 250))
		return "", err
	}

	return navigateJSON(&jsonData, key)
}
