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

// Package config provides the configuration for Argus.
package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/release-argus/Argus/util"
)

// CheckValues validates the fields of the Config struct.
func (c *Config) CheckValues() {
	var errs []error
	c.Settings.CheckValues()

	// defaults.
	util.AppendCheckError(&errs, "", "defaults", c.Defaults.CheckValues("  "))
	// notify.
	util.AppendCheckError(&errs, "", "notify", c.Notify.CheckValues("  "))
	// webhook.
	util.AppendCheckError(&errs, "", "webhook", c.WebHook.CheckValues("  "))
	// service.
	util.AppendCheckError(&errs, "", "service", c.Service.CheckValues("  "))

	// Combine all errors if any are present.
	if len(errs) > 0 {
		combinedErr := errors.Join(errs...)
		fmt.Println(combinedErr.Error())
		jLog.Fatal("Config could not be parsed successfully.", util.LogFrom{}, true)
	}
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
	if !jLog.Testing {
		os.Exit(0)
	}
}
