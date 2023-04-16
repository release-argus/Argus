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

	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	opt "github.com/release-argus/Argus/service/options"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
	"gopkg.in/yaml.v3"
)

// Defaults for the other Structs.
type Defaults struct {
	Service service.Service `yaml:"service,omitempty"`
	Notify  shoutrrr.Slice  `yaml:"notify,omitempty"`
	WebHook webhook.WebHook `yaml:"webhook,omitempty"`
}

// String returns a string representation of the Defaults.
func (d *Defaults) String() string {
	if d == nil {
		return "<nil>"
	}

	yamlBytes, _ := yaml.Marshal(d)
	return string(yamlBytes)
}

// SetDefaults (last resort vars).
func (d *Defaults) SetDefaults() {
	// Service defaults.
	serviceSemanticVersioning := true
	d.Service.Options = opt.Options{
		Interval:           "10m",
		SemanticVersioning: &serviceSemanticVersioning}
	serviceLatestVersionAllowInvalidCerts := false
	usePreRelease := false
	d.Service.LatestVersion = latestver.Lookup{
		AllowInvalidCerts: &serviceLatestVersionAllowInvalidCerts,
		UsePreRelease:     &usePreRelease}
	serviceDeployedVersionLookupAllowInvalidCerts := false
	d.Service.DeployedVersionLookup = &deployedver.Lookup{
		AllowInvalidCerts: &serviceDeployedVersionLookupAllowInvalidCerts}
	serviceAutoApprove := false
	d.Service.Dashboard = service.DashboardOptions{
		AutoApprove: &serviceAutoApprove}

	notifyDefaultOptions := map[string]string{
		"message":   "{{ service_id }} - {{ version }} released",
		"max_tries": "3",
		"delay":     "0s"}

	// Notify defaults.
	d.Notify = make(shoutrrr.Slice)
	d.Notify["discord"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params: types.Params{
			"username": "Argus"}}
	d.Notify["smtp"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		URLFields: map[string]string{
			"port": "25"},
		Params: types.Params{}}
	d.Notify["googlechat"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	d.Notify["gotify"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		URLFields: map[string]string{
			"port": "443"},
		Params: types.Params{
			"title":    "Argus",
			"priority": "0"}}
	d.Notify["ifttt"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params: types.Params{
			"title":             "Argus",
			"usemessageasvalue": "2",
			"usetitleasvalue":   "0"}}
	d.Notify["join"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	d.Notify["mattermost"] = &shoutrrr.Shoutrrr{
		Options: map[string]string{
			"message":   "<{{ service_url }}|{{ service_id }}> - {{ version }} released{% if web_url %} (<{{ web_url }}|changelog>){% endif %}",
			"max_tries": "3",
			"delay":     "0s"},
		URLFields: map[string]string{
			"username": "Argus",
			"port":     "443"}}
	d.Notify["matrix"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		URLFields: map[string]string{
			"port": "443"},
		Params: types.Params{}}
	d.Notify["opsgenie"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	d.Notify["pushbullet"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		URLFields: map[string]string{
			"port": "443"},
		Params: types.Params{
			"title": "Argus"}}
	d.Notify["pushover"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	d.Notify["rocketchat"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		URLFields: map[string]string{
			"port": "443"},
		Params: types.Params{}}
	d.Notify["slack"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params: types.Params{
			"botname": "Argus"}}
	d.Notify["teams"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	d.Notify["telegram"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params: types.Params{
			"notification": "yes",
			"preview":      "yes"}}
	d.Notify["zulip"] = &shoutrrr.Shoutrrr{
		Options: notifyDefaultOptions,
		Params:  types.Params{}}
	// InitMaps
	for _, notify := range d.Notify {
		notify.InitMaps()
	}

	// WebHook defaults.
	d.WebHook.Type = "github"
	d.WebHook.Delay = "0s"
	webhookMaxTries := uint(3)
	d.WebHook.MaxTries = &webhookMaxTries
	webhookAllowInvalidCerts := false
	d.WebHook.AllowInvalidCerts = &webhookAllowInvalidCerts
	webhookDesiredStatusCode := 0
	d.WebHook.DesiredStatusCode = &webhookDesiredStatusCode
	webhookSilentFails := false
	d.WebHook.SilentFails = &webhookSilentFails

	// Overwrite defaults with environment variables.
	err := d.MapEnvToStruct()
	jLog.Fatal(
		strings.ReplaceAll(util.ErrorToString(err), "\\", "\n"),
		util.LogFrom{},
		err != nil)
}

// MapEnvToStruct maps environment variables this struct.
func (d *Defaults) MapEnvToStruct() (err error) {
	err = mapEnvToStruct(d, "", nil)
	if err != nil {
		return
	}
	err = d.CheckValues()
	return
}
func mapEnvToStruct(src interface{}, prefix string, envVars *[]string) (err error) {
	srcV := reflect.ValueOf(src).Elem()
	if prefix == "" {
		prefix = "ARGUS_"
		// All env vars
		allEnvVars := os.Environ()
		envVarsTrimmed := make([]string, 0, len(allEnvVars))
		// TRIM non-ARGUS env vars
		for _, envVar := range allEnvVars {
			if strings.HasPrefix(envVar, prefix) {
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
		fieldName := srcT.Field(i).Tag.Get("yaml")
		if strings.Contains(fieldName, ",") {
			fieldName = strings.Split(fieldName, ",")[0]
		}
		if fieldName == "" || fieldName == "-" {
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
					}

					// UInt
				case reflect.Uint:
					envValue, err := strconv.ParseUint(envValueStr, 10, 64)
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
			if strings.HasPrefix(fieldName, "ARGUS_NOTIFY_") &&
				!strings.HasSuffix(fieldName, "ARGUS_NOTIFY_") {
				for _, envVar := range *envVars {
					if strings.HasPrefix(envVar, fieldName) {
						// Get key and value
						keyValue := strings.SplitN(envVar, "=", 2)
						// Remove fieldName from key
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
				err = mapEnvToStruct(srcV.Field(i).Interface(), fieldName, envVars)
			} else {
				err = mapEnvToStruct(srcV.Field(i).Addr().Interface(), fieldName, envVars)
			}
		}
	}
	return
}

// CheckValues are valid.
func (d *Defaults) CheckValues() (errs error) {
	prefix := "  "

	// Service
	if err := d.Service.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%sservice:\\%w",
			util.ErrorToString(errs), prefix, err)
	}

	// Notify
	for i := range d.Notify {
		// Remove the types since the key is the type
		d.Notify[i].Type = ""
	}
	if err := d.Notify.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%w",
			util.ErrorToString(errs), err)
	}

	// WebHook
	if err := d.WebHook.CheckValues(prefix); err != nil {
		errs = fmt.Errorf("%s%swebhook:\\%w",
			util.ErrorToString(errs), prefix, err)
	}

	return
}

// Print the defaults Strcut.
func (d *Defaults) Print() {
	fmt.Println("defaults:")

	// Service defaults.
	fmt.Println("  service:")
	d.Service.Print("    ")

	// Notify defaults.
	d.Notify.Print("  ")

	// WebHook defaults.
	fmt.Println("  webhook:")
	d.WebHook.Print("    ")
}
