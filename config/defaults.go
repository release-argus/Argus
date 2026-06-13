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

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/webhook"
)

// Defaults holds default configuration for services, notifiers, and webhooks.
type Defaults struct {
	Service service.Defaults           `json:"service,omitzero" yaml:"service,omitzero"`
	Notify  shoutrrr.ShoutrrrsDefaults `json:"notify,omitzero" yaml:"notify,omitzero"`
	WebHook webhook.Defaults           `json:"webhook,omitzero" yaml:"webhook,omitzero"`
}

// IsZero implements the yaml.IsZeroer interface.
func (d *Defaults) IsZero() bool {
	return d == nil || (d.Service.IsZero() && d.Notify.IsZero() && d.WebHook.IsZero())
}

// DecodeDefaults creates and returns new [Defaults] from format-encoded data.
func DecodeDefaults(format string, data []byte) (*Defaults, error) {
	var field Defaults
	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "defaults",
			Err: err,
		}
	}

	return &field, nil
}

// String returns a string representation of the receiver.
func (d *Defaults) String(prefix string) string {
	if d == nil {
		return ""
	}
	return decode.ToYAMLString(d, prefix)
}

// Default sets the values of the receiver to their default values.
func (d *Defaults) Default() bool {
	d.Service.Default()

	// Notify defaults.
	d.Notify.Default()

	// WebHook defaults.
	d.WebHook.Default()

	// Overwrite defaults with environment variables.
	if ok := d.MapEnvToStruct(); !ok {
		return false
	}

	// Notify Types.
	for notifyType, notify := range d.Notify {
		notify.Type = notifyType
	}

	return true
}

// SetDefaults assigns defaults to the receiver.
func (d *Defaults) SetDefaults(dflts *Defaults) {
	// TODO: Store HardDefaults inside Defaults.
	d.Service.SetDefaults(&dflts.Service)
	d.Notify.Init()
	dflts.Notify.Init()
}

// MapEnvToStruct maps environment variables to the receiver.
func (d *Defaults) MapEnvToStruct() bool {
	err := mapEnvToStruct(d, "", nil)
	if err == nil {
		// env vars parsed correctly, check the values are valid in the struct.
		// (ignore changed as we can't persist environment variable changes)
		if err, _ = d.CheckValues(); err != nil {
			err = convertToEnvErrors(err)
		}
	}

	if err != nil {
		logx.Fatal(
			"One or more 'ARGUS_' environment variables are invalid:\n"+err.Error(),
			logx.LogFrom{},
		)
		return false
	}
	return true
}

// CheckValues validates each entry and reports whether any values were modified.
func (d *Defaults) CheckValues() (error, bool) {
	var errs []error

	// Service.
	if err := d.Service.CheckValues(); err != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "service",
				Err: err,
			},
		)
	}
	// Notify.
	notifyErr, notifyChanged := d.Notify.CheckValues()
	if notifyErr != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "notify",
				Err: notifyErr,
			},
		)
	}
	// WebHook.
	webhookErr, webhookChanged := d.WebHook.CheckValues()
	if webhookErr != nil {
		errs = append(
			errs,
			&decode.KeyFieldError{
				Key: "webhook",
				Err: webhookErr,
			},
		)
	}

	if len(errs) == 0 {
		return nil, notifyChanged || webhookChanged
	}
	return errors.Join(errs...), false
}

// Print writes a YAML representation of the receiver to stdout.
func (d *Defaults) Print(prefix string) {
	if d.IsZero() {
		return
	}

	str := d.String(prefix + "  ")
	fmt.Printf(
		"%sdefaults:\n%s",
		prefix, str,
	)
}
