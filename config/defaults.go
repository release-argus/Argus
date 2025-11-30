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
	"errors"
	"fmt"

	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
	"github.com/release-argus/Argus/webhook"
)

// Defaults for the other Structs.
type Defaults struct {
	Service service.Defaults           `json:"service,omitempty" yaml:"service,omitempty"`
	Notify  shoutrrr.ShoutrrrsDefaults `json:"notify,omitempty" yaml:"notify,omitempty"`
	WebHook webhook.Defaults           `json:"webhook,omitempty" yaml:"webhook,omitempty"`
}

// String returns a string representation of the Defaults.
func (d *Defaults) String(prefix string) string {
	if d == nil {
		return ""
	}
	return util.ToYAMLString(d, prefix)
}

// Default sets these Defaults to the default values.
func (d *Defaults) Default() {
	d.Service.Default()

	// Notify defaults.
	d.Notify.Default()

	// WebHook defaults.
	d.WebHook.Default()

	// Overwrite defaults with environment variables.
	d.MapEnvToStruct()

	// Notify Types.
	for notifyType, notify := range d.Notify {
		notify.Type = notifyType
	}
}

// MapEnvToStruct maps environment variables to this struct.
func (d *Defaults) MapEnvToStruct() {
	err := mapEnvToStruct(d, "", nil)
	if err == nil {
		// env vars parsed correctly, check the values are valid in the struct.
		if err = d.CheckValues(""); err != nil {
			err = convertToEnvErrors(err)
		}
	}

	if err != nil {
		logutil.Log.Fatal(
			"One or more 'ARGUS_' environment variables are invalid:\n"+err.Error(),
			logutil.LogFrom{}, true)
	}
}

// CheckValues validates the fields of the Defaults struct.
func (d *Defaults) CheckValues(prefix string) error {
	itemPrefix := prefix + "  "
	var errs []error

	// Service.
	util.AppendCheckError(&errs, prefix, "service",
		d.Service.CheckValues(itemPrefix))
	// Notify.
	util.AppendCheckError(&errs, prefix, "notify",
		d.Notify.CheckValues(itemPrefix))
	// WebHook.
	util.AppendCheckError(&errs, prefix, "webhook",
		d.WebHook.CheckValues(itemPrefix))

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// Print the defaults to the console with the given prefix.
func (d *Defaults) Print(prefix string) {
	itemPrefix := prefix + "  "
	str := d.String(itemPrefix)
	delim := "\n"
	if str == "{}\n" {
		delim = " "
	}
	fmt.Printf("%sdefaults:%s%s",
		prefix, delim, str)
}
