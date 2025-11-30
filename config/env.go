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

// Package config provides the configuration for Argus.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/release-argus/Argus/util"
)

// mapEnvToStruct maps environment variables to a struct.
func mapEnvToStruct(src any, prefix string, envVars []string) error {
	var errs []error
	srcV := reflect.ValueOf(src)
	if srcV.Kind() == reflect.Ptr {
		srcV = srcV.Elem()
	}

	// First call, get all matching env vars.
	if prefix == "" {
		prefix = "ARGUS_"
		// Extract ARGUS_* env vars.
		for _, envVar := range os.Environ() {
			// Skip empty env vars.
			if strings.HasPrefix(envVar, prefix) && strings.SplitN(envVar, "=", 2)[1] != "" {
				envVars = append(envVars, envVar)
			}
		}
		// Have no env vars to map.
		if len(envVars) == 0 {
			return nil
		}
	}

	for i := 0; i < srcV.NumField(); i++ {
		field := srcV.Field(i)
		fieldType := field.Type()
		kind := fieldType.Kind()
		// Get kind of this pointer.
		if kind == reflect.Ptr {
			// Skip nil pointers to non-comparable types.
			if !fieldType.Elem().Comparable() && field.IsNil() {
				continue
			}
			kind = fieldType.Elem().Kind()
		}

		// YAML tag of this field.
		srcT := reflect.TypeOf(src)
		if srcT.Kind() == reflect.Ptr {
			srcT = srcT.Elem()
		}
		fieldTag := srcT.Field(i).Tag.Get("yaml")
		fieldName := strings.Split(fieldTag, ",")[0]
		if fieldName == "" || fieldName == "-" {
			if fieldTag == ",inline" {
				if err := mapEnvToStruct(field.Addr().Interface(), prefix, envVars); err != nil {
					errs = append(errs, err)
				}
			}
			continue
		}
		fieldName = strings.ToUpper(prefix + fieldName)
		if !hasVarWithPrefix(fieldName, envVars) {
			continue
		}

		var err error
		switch kind {
		case reflect.Bool, reflect.Int, reflect.String, reflect.Uint8, reflect.Uint16:
			// Check if env var exists for this field.
			if envValueStr, exists := os.LookupEnv(fieldName); exists {
				// Get ENV var in correct type.
				switch kind {
				case reflect.Bool:
					err = setBoolField(field, envValueStr, fieldName)
				case reflect.Int:
					err = setIntField(field, envValueStr, fieldName)
				case reflect.String:
					err = setStringField(field, envValueStr)
				case reflect.Uint8:
					err = setUintField(field, envValueStr, fieldName, 8)
				case reflect.Uint16:
					err = setUintField(field, envValueStr, fieldName, 16)
				}
			}
		case reflect.Map:
			err = setMapFields(field, fieldName, envVars)
		case reflect.Struct:
			fieldName += "_"
			if field.Kind() == reflect.Ptr {
				// If nil, create a new instance of this struct.
				if field.IsNil() {
					field.Set(reflect.New(field.Type().Elem()))
				}

				err = mapEnvToStruct(field.Interface(), fieldName, envVars)
			} else {
				err = mapEnvToStruct(field.Addr().Interface(), fieldName, envVars)
			}
		}
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// hasVarWithPrefix returns whether any of the environment variables in the provided
// slice start with the specified prefix.
func hasVarWithPrefix(prefix string, envVars []string) bool {
	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, prefix) {
			return true
		}
	}

	return false
}

func setBoolField(field reflect.Value, value, envKey string) error {
	return setField(field, value, envKey, strconv.ParseBool, "(expected 'true' or 'false')")
}

func setIntField(field reflect.Value, value, envKey string) error {
	return setField(field, value, envKey, strconv.Atoi, "(expected an integer)")
}

func setStringField(field reflect.Value, value string) error {
	return setField(field, value, "", func(s string) (string, error) { return s, nil }, "")
}

func setUintField(field reflect.Value, value, envKey string, bitSize int) error {
	parser := func(s string) (uint16, error) {
		v, err := strconv.ParseUint(s, 10, bitSize)
		return uint16(v), err
	}
	return setField(field, value, envKey, parser,
		fmt.Sprintf("(expected an unsigned (non-negative) integer between 0 and %d)", math.MaxUint16))
}

// setField sets a given field's value by parsing a string using the provided parser function and validates the result.
// The field update depends on its kind (pointer or value). It returns an error if parsing or assignment fails.
func setField[T any](field reflect.Value, value, envKey string, parser func(string) (T, error), errorMsg string) error {
	parsedValue, err := parser(value)
	if err != nil {
		return fmt.Errorf("%s: %q <invalid> %s", envKey, value, errorMsg)
	}

	if field.Kind() == reflect.Ptr {
		field.Set(reflect.ValueOf(&parsedValue))
	} else {
		switch v := any(parsedValue).(type) {
		case bool:
			field.SetBool(v)
		case int:
			field.SetInt(int64(v))
		case string:
			field.SetString(v)
		case uint8:
			field.SetUint(uint64(v))
		case uint16:
			field.SetUint(uint64(v))
		}
	}

	return nil
}

func setMapFields(field reflect.Value, envKey string, envVars []string) error {
	// Notify maps.
	if strings.HasPrefix(envKey, "ARGUS_NOTIFY_") {
		for _, envVar := range envVars {
			if strings.HasPrefix(envVar, envKey) {
				// Get key and value.
				keyValue := strings.SplitN(envVar, "=", 2)

				// Remove fieldName from key (get key of map).
				// e.g. "ARGUS_NOTIFY_MATTERMOST_OPTIONS_MAX_TRIES=7"
				// = "max_tries=7"
				keyValue[0] = strings.ToLower(
					strings.Replace(keyValue[0], envKey+"_", "", 1))

				// Set value in map.
				field.SetMapIndex(reflect.ValueOf(keyValue[0]), reflect.ValueOf(keyValue[1]))
			}
		}

		return nil
	}

	var errs []error
	// Recurse into map.
	for _, key := range field.MapKeys() {
		if err := mapEnvToStruct(
			field.MapIndex(key).Interface(),
			fmt.Sprintf("%s_%s_",
				envKey, strings.ToUpper(key.String())),
			envVars); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// convertToEnvErrors converts the YAML struct errors to environment variable errors.
func convertToEnvErrors(errs error) error {
	if errs == nil {
		return nil
	}

	var newErrs []error
	basePrefix := []string{"ARGUS"}
	lines := strings.Split(errs.Error(), "\n")
	varLineRegex := regexp.MustCompile(`^(\s*)([^;]+):$`)
	valueRegex := regexp.MustCompile(`^\s*([^:]+): (.+)$`)
	currentIndent := -1
	for _, line := range lines {
		if varLineRegex.MatchString(line) {
			match := varLineRegex.FindStringSubmatch(line)
			indent := len(match[1]) / 2
			varName := strings.ToUpper(match[2])
			// Check whether indent matches the current indent.
			switch {
			case indent == currentIndent:
				basePrefix = basePrefix[:len(basePrefix)-1]
				basePrefix = append(basePrefix, varName)
			case indent > currentIndent:
				basePrefix = append(basePrefix, varName)
			default:
				basePrefix = basePrefix[:indent+1]
				basePrefix = append(basePrefix, varName)
			}
			currentIndent = indent
		} else {
			value := valueRegex.FindStringSubmatch(line)
			newErrs = append(newErrs,
				errors.New(strings.Join(basePrefix, "_")+
					fmt.Sprintf("_%s: %s",
						strings.ToUpper(value[1]), value[2])))
		}
	}

	return errors.Join(newErrs...)
}

// loadEnvFile loads environment variables from the specified file.
func loadEnvFile(filePath string) error {
	// Check if the file exists.
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	// Open the file.
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open env file %q:\n  %w",
			filePath, err)
	}
	defer file.Close()

	// Load environment variables from the file.
	return loadEnvFromReader(file)
}

// loadEnvFromReader loads environment variables from an io.Reader.
//
//	It reads the content line by line, ignoring empty lines and comments.
//	Both "key=value" and "export key=value" formats are supported.
//	Quoted values, using either single or double quotes, are also handled.
func loadEnvFromReader(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignore empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Get the key and its value.
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid env line: %q", line)
		}
		key := strings.TrimSpace(strings.TrimPrefix(parts[0], "export "))

		value := parts[1]
		// Handle quoted values.
		if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
			(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
			value = value[1 : len(value)-1]
		}
		// Handle nested environment variables.
		value = util.EvalEnvVars(value)

		// Set the environment variable.
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set env var %q: %w",
				key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file: %w", err)
	}

	return nil
}
