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
	"os"

	"github.com/release-argus/Argus/util"
	logutil "github.com/release-argus/Argus/util/log"
)

// CheckValues validates the fields of the Config struct.
func (c *Config) CheckValues() bool {
	var errs []error

	childPrefix := "  "

	// settings.
	util.AppendCheckError(&errs, "", "settings", c.Settings.CheckValues(childPrefix))
	// defaults.
	defaultsErr, defaultsChanged := c.Defaults.CheckValues(childPrefix)
	if defaultsChanged {
		c.SaveChannel <- true
	}
	util.AppendCheckError(&errs, "", "defaults", defaultsErr)
	// notify.
	notifyErr, notifyChanged := c.Notify.CheckValues(childPrefix)
	if notifyChanged {
		c.SaveChannel <- true
	}
	util.AppendCheckError(&errs, "", "notify", notifyErr)
	// webhook.
	webhookErr, webhookChanged := c.WebHook.CheckValues(childPrefix)
	if webhookChanged {
		c.SaveChannel <- true
	}
	util.AppendCheckError(&errs, "", "webhook", webhookErr)
	// service.
	util.AppendCheckError(&errs, "", "service",
		c.Service.CheckValues(childPrefix))

	// Combine all errors if any are present.
	if len(errs) > 0 {
		combinedErr := errors.Join(errs...)
		fmt.Println(combinedErr.Error())
		logutil.Log.Fatal("Config could not be parsed successfully.", logutil.LogFrom{})
		return false
	}
	return true
}

// Print the parsed config if *flag.
func (c *Config) Print(flag *bool) {
	if !*flag {
		return
	}

	c.Defaults.Print("")
	if len(c.Notify) > 0 {
		c.Notify.Print("")
		fmt.Println()
	}
	if len(c.WebHook) > 0 {
		c.WebHook.Print("")
		fmt.Println()
	}
	if len(c.Order) > 0 {
		c.Service.Print("", c.Order)
		fmt.Println()
	}
	if !logutil.Log.Testing {
		os.Exit(0)
	}
}
