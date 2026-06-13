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
	"sync"

	"github.com/release-argus/Argus/config/decode"
	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/util/polymorphic"
	"github.com/release-argus/Argus/webhook"
)

// extractServiceSubtree extracts the service subtree from raw config bytes.
// (overridable for tests).
var extractServiceSubtree = polymorphic.Extract

// Config for Argus.
type Config struct {
	File         string                     `json:"-" yaml:"-"`                                   // Path to the config file (--config.file='').
	Settings     Settings                   `json:"settings,omitempty" yaml:"settings,omitempty"` // Settings for the program.
	HardDefaults Defaults                   `json:"-" yaml:"-"`                                   // Hardcoded default values for the various parameters.
	Defaults     Defaults                   `json:"defaults,omitempty" yaml:"defaults,omitempty"` // Default values for the various parameters.
	Notify       shoutrrr.ShoutrrrsDefaults `json:"notify,omitempty" yaml:"notify,omitempty"`     // Shoutrrr messages to send on a new release.
	WebHook      webhook.WebHooksDefaults   `json:"webhook,omitempty" yaml:"webhook,omitempty"`   // WebHooks to send on a new release.

	OrderMu sync.RWMutex     `json:"-" yaml:"-"`                                 // Mutex for the Order/Service slice.
	Order   []string         `json:"-" yaml:"-"`                                 // Ordered slice of all Service id's.
	Service service.Services `json:"service,omitempty" yaml:"service,omitempty"` // The services to monitor.

	DatabaseChannel chan dbtype.Message `json:"-" yaml:"-"` // Channel for broadcasts to the Database.
	SaveChannel     chan bool           `json:"-" yaml:"-"` // Channel for triggering a save of the config.
}

// ConfigDecode is an unmarshal-only helper for [Config].
type ConfigDecode struct {
	Settings Settings                   `json:"settings,omitempty" yaml:"settings,omitempty"`
	Notify   shoutrrr.ShoutrrrsDefaults `json:"notify,omitempty" yaml:"notify,omitempty"`
	WebHook  webhook.WebHooksDefaults   `json:"webhook,omitempty" yaml:"webhook,omitempty"`
	Defaults Defaults                   `json:"defaults,omitempty" yaml:"defaults,omitempty"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [Decode] for a full unmarshal.
func (c *Config) UnmarshalJSON(data []byte) error {
	return c.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [Decode] for a full unmarshal.
func (c *Config) UnmarshalYAML(data []byte) error {
	return c.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (c *Config) unmarshal(format string, data []byte) error {
	aux := ConfigDecode{
		Settings: c.Settings,
		Notify:   c.Notify,
		WebHook:  c.WebHook,
		Defaults: c.Defaults,
	}

	// Unmarshal in the given format.
	if err := decode.Unmarshal(format, data, &aux); err != nil {
		return err //nolint:wrapcheck
	}
	c.Settings = aux.Settings
	c.Notify = aux.Notify
	if c.Notify == nil {
		c.Notify = shoutrrr.ShoutrrrsDefaults{}
	}
	c.WebHook = aux.WebHook
	if c.WebHook == nil {
		c.WebHook = webhook.WebHooksDefaults{}
	}
	c.Defaults = aux.Defaults

	return nil
}

// Decode decodes YAML-encoded data into the receiver.
func (c *Config) Decode(raw []byte) error {
	if len(raw) == 0 {
		return nil
	}

	databaseChannel := make(chan dbtype.Message, 32)
	c.DatabaseChannel = databaseChannel
	c.HardDefaults.Service.Status.DatabaseChannel = c.DatabaseChannel

	saveChannel := make(chan bool, 32)
	c.SaveChannel = saveChannel
	c.HardDefaults.Service.Status.SaveChannel = c.SaveChannel

	c.OrderMu.Lock()
	defer c.OrderMu.Unlock()

	// Unmarshal static fields.
	if err := decode.Unmarshal("yaml", raw, c); err != nil {
		return err //nolint:wrapcheck
	}

	// Set defaults/hardDefaults.
	c.InitDefaults()

	// Extract "service" subtree.
	serviceRaw, err := extractServiceSubtree("yaml", raw, "service")
	if err != nil {
		return err
	}

	svcDefaults, notifyDefaults, webhookDefaults := c.GetDefaults()

	// Decode services.
	c.Service, err = service.DecodeServices(
		"yaml", serviceRaw,
		svcDefaults, notifyDefaults, webhookDefaults,
	)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// GetDefaults constructs and returns the default configurations for Service, Shoutrrr, and WebHook on the receiver.
func (c *Config) GetDefaults() (service.DefaultsConfig, shoutrrr.Config, webhook.Config) {
	svcDefaults := service.DefaultsConfig{
		Soft: &c.Defaults.Service,
		Hard: &c.HardDefaults.Service,
	}
	notifyDefaults := shoutrrr.Config{
		Root:         c.Notify,
		Defaults:     c.Defaults.Notify,
		HardDefaults: c.HardDefaults.Notify,
	}
	webhookDefaults := webhook.Config{
		Root:         c.WebHook,
		Defaults:     &c.Defaults.WebHook,
		HardDefaults: &c.HardDefaults.WebHook,
	}
	return svcDefaults, notifyDefaults, webhookDefaults
}
