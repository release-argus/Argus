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

// Package config provides the configuration for Argus.
package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/util/errfmt"
)

// exitAfterPrint terminates the process after printing config
// (overridable for tests).
var exitAfterPrint = os.Exit

// CheckValues validates the fields of the receiver.
func (c *Config) CheckValues() bool {
	var errs []error

	// settings.
	if err := c.Settings.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "settings",
				Err: err,
			},
		)
	}
	// defaults.
	defaultsErr, defaultsChanged := c.Defaults.CheckValues()
	if defaultsErr != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "defaults",
				Err: defaultsErr,
			},
		)
	}
	// notify.
	notifyErr, notifyChanged := c.Notify.CheckValues()
	if notifyErr != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "notify",
				Err: notifyErr,
			},
		)
	}
	// webhook.
	webhookErr, webhookChanged := c.WebHook.CheckValues()
	if webhookErr != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "webhook",
				Err: webhookErr,
			},
		)
	}
	// service.
	serviceErr, serviceChanged := c.Service.CheckValues()
	if serviceErr != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "service",
				Err: serviceErr,
			},
		)
	}

	// Combine all errors if any are present.
	if len(errs) > 0 {
		combinedErr := errors.Join(errs...)
		fmt.Println(errfmt.FormatError(combinedErr))
		logx.Fatal("Config could not be parsed successfully.", logx.LogFrom{})
		return false
	}

	// All files valid, queue save if changed.
	if defaultsChanged || notifyChanged || webhookChanged || serviceChanged {
		c.SaveChannel <- true
	}
	return true
}

// Print writes the parsed configuration to stdout when flag is true, then exits.
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

	exitAfterPrint(0)
}
