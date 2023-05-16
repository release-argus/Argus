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

package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

// Defaults for the other Structs.
type Defaults struct {
	Service service.Defaults        `yaml:"service,omitempty" json:"service,omitempty"`
	Notify  shoutrrr.SliceDefaults  `yaml:"notify,omitempty" json:"notify,omitempty"`
	WebHook webhook.WebHookDefaults `yaml:"webhook,omitempty" json:"webhook,omitempty"`
}

// String returns a string representation of the Defaults.
func (d *Defaults) String(prefix string) (str string) {
	if d != nil {
		str = util.ToYAMLString(d, prefix)
	}
	return
}

// SetDefaults (last resort vars).
func (d *Defaults) SetDefaults() {
	d.Service.SetDefaults()

	// Notify defaults
	d.Notify.SetDefaults()

	// WebHook defaults.
	d.WebHook.SetDefaults()

	// Overwrite defaults with environment variables.
	d.MapEnvToStruct()

	// Notify Types
	for notifyType, notify := range d.Notify {
		notify.Type = notifyType
	}
}

// MapEnvToStruct maps environment variables to this struct.
func (d *Defaults) MapEnvToStruct() {
	err := mapEnvToStruct(d, "", nil)
	if err == nil {
		err = d.CheckValues("")
	}
	if err != nil {
		jLog.Fatal(
			"One or more 'ARGUS_' environment variables are incorrect:\n"+
				strings.ReplaceAll(util.ErrorToString(err), "\\", "\n"),
			util.LogFrom{}, true)
	}
}

// mapEnvToStruct maps environment variables to a struct.
func mapEnvToStruct(src interface{}, prefix string, envVars *[]string) (err error) {
	srcV := reflect.ValueOf(src).Elem()
	// First call, get all matching env vars
	if prefix == "" {
		prefix = "ARGUS_"
		// All env vars
		allEnvVars := os.Environ()
		envVarsTrimmed := make([]string, 0, len(allEnvVars))
		// TRIM non-ARGUS env vars
		for _, envVar := range allEnvVars {
			parts := strings.SplitN(envVar, "=", 2)
			// Skip empty env vars
			if strings.HasPrefix(envVar, prefix) && parts[1] != "" {
				envVarsTrimmed = append(envVarsTrimmed, envVar)
			}
		}
		// Have no pointers to map
		if len(envVarsTrimmed) == 0 {
			return
		}
		envVars = &envVarsTrimmed
	}

	for i := 0; i < srcV.NumField(); i++ {
		fieldType := srcV.Field(i).Type()
		kind := fieldType.Kind()
		// Get kind of this pointer
		if kind == reflect.Ptr {
			// Skip nil pointers to non-comparable types
			if !fieldType.Elem().Comparable() && srcV.Field(i).IsNil() {
				continue
			}
			kind = fieldType.Elem().Kind()
		}

		// YAML tag of this field
		srcT := reflect.TypeOf(src).Elem()
		fieldTag := srcT.Field(i).Tag.Get("yaml")
		fieldName := strings.Split(fieldTag, ",")[0]
		if fieldName == "" || fieldName == "-" {
			if fieldTag == ",inline" {
				err = mapEnvToStruct(srcV.Field(i).Addr().Interface(), prefix, envVars)
				if err != nil {
					return
				}
			}
			continue
		}
		fieldName = strings.ToUpper(prefix + fieldName)
		hasEnvVar := false
		for _, envVar := range *envVars {
			if strings.HasPrefix(envVar, fieldName) {
				hasEnvVar = true
				break
			}
		}
		if !hasEnvVar {
			continue
		}

		switch kind {
		case reflect.Bool, reflect.Int, reflect.String, reflect.Uint:
			// Check if env var exists for this field
			if envValueStr, exists := os.LookupEnv(fieldName); exists {
				isPointer := fieldType.Kind() == reflect.Ptr
				// Get ENV var in correct type
				switch kind {
				// Boolean
				case reflect.Bool:
					envValue, err := strconv.ParseBool(envValueStr)
					if err != nil {
						return fmt.Errorf("invalid bool for %s: %q", fieldName, envValueStr)
					}
					// All are pointers to distinguish between undefined
					if isPointer {
						srcV.Field(i).Set(reflect.ValueOf(&envValue))
					}

					// Integer
				case reflect.Int:
					envValue, err := strconv.Atoi(envValueStr)
					if err != nil {
						return fmt.Errorf("invalid integer for %s: %q", fieldName, envValueStr)
					}
					// All are pointers to distinguish between undefined
					if isPointer {
						srcV.Field(i).Set(reflect.ValueOf(&envValue))
					}

					// String
				case reflect.String:
					envValue := envValueStr
					if !isPointer {
						srcV.Field(i).SetString(envValue)
					} else {
						srcV.Field(i).Set(reflect.ValueOf(&envValue))
					}

					// UInt
				case reflect.Uint:
					envValue, err := strconv.ParseUint(envValueStr, 10, 32)
					if err != nil {
						return fmt.Errorf("invalid uint for %s: %q", fieldName, envValueStr)
					}
					// All are pointers to distinguish between undefined
					if isPointer {
						uInt := uint(envValue)
						srcV.Field(i).Set(reflect.ValueOf(&uInt))
					}
				}
			}
		case reflect.Map:
			// Notify maps
			if strings.HasPrefix(fieldName, "ARGUS_NOTIFY_") {
				for _, envVar := range *envVars {
					if strings.HasPrefix(envVar, fieldName) {
						// Get key and value
						keyValue := strings.SplitN(envVar, "=", 2)

						// Remove fieldName from key (get key of map)
						// e.g."ARGUS_NOTIFY_MATTERMOST_OPTIONS_MAX_TRIES=7"
						// -> "max_tries=7"
						keyValue[0] = strings.ToLower(
							strings.Replace(keyValue[0], fieldName+"_", "", 1))

						// Set value in map
						srcV.Field(i).SetMapIndex(reflect.ValueOf(keyValue[0]), reflect.ValueOf(keyValue[1]))
					}
				}
				continue
			}
			// Recurse into map
			for _, key := range srcV.Field(i).MapKeys() {
				err = mapEnvToStruct(
					srcV.Field(i).MapIndex(key).Interface(),
					fmt.Sprintf("%s_%s_", fieldName, strings.ToUpper(key.String())),
					envVars)
			}
		case reflect.Struct:
			fieldName += "_"
			if fieldType.Kind() == reflect.Ptr {
				// If it's nil, create a new instance of this struct
				if srcV.Field(i).IsNil() {
					srcV.Field(i).Set(reflect.New(fieldType.Elem()))
				}

				err = mapEnvToStruct(srcV.Field(i).Interface(), fieldName, envVars)
			} else {
				err = mapEnvToStruct(srcV.Field(i).Addr().Interface(), fieldName, envVars)
			}
		}
	}
	return
}

// CheckValues are valid.
func (d *Defaults) CheckValues(prefix string) (errs error) {
	itemPrefix := prefix + "  "

	// Service
	if err := d.Service.CheckValues(itemPrefix); err != nil {
		errs = fmt.Errorf("%s%sservice:\\%w",
			util.ErrorToString(errs), prefix, err)
	}

	// Notify
	if err := d.Notify.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), err)
	}

	// WebHook
	if err := d.WebHook.CheckValues(itemPrefix); err != nil {
		errs = fmt.Errorf("%s%swebhook:\\%w",
			util.ErrorToString(errs), prefix, err)
	}

	return
}

// Print the defaults Strcut.
func (d *Defaults) Print(prefix string) {
	itemPrefix := prefix + "  "
	str := d.String(itemPrefix)
	delim := "\n"
	if str == "{}\n" {
		delim = " "
	}
	fmt.Printf("%sdefaults:%s%s", prefix, delim, str)
}
