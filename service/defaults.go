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

// Package service provides the service functionality for Argus.
package service

import (
	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver_base "github.com/release-argus/Argus/service/deployed_version/types/base"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util/polymorphic"
)

// Defaults are the default values for a Service.
type Defaults struct {
	Options               opt.Defaults              `json:"options,omitzero" yaml:"options,omitzero"`                   // Options to give the Service.
	LatestVersion         latestver_base.Defaults   `json:"latest_version,omitzero" yaml:"latest_version,omitzero"`     // Vars to scrape the latest version of the Service.
	DeployedVersionLookup deployedver_base.Defaults `json:"deployed_version,omitzero" yaml:"deployed_version,omitzero"` // Vars to scrape the Service's current deployed version.
	Notify                map[string]struct{}       `json:"notify,omitempty" yaml:"notify,omitempty"`                   // Default Notifiers to give a Service.
	Command               command.Commands          `json:"command,omitempty" yaml:"command,omitempty"`                 // Default Commands to give a Service.
	WebHook               map[string]struct{}       `json:"webhook,omitempty" yaml:"webhook,omitempty"`                 // Default WebHooks to give a Service.
	Dashboard             dashboard.Defaults        `json:"dashboard,omitzero" yaml:"dashboard,omitzero"`               // Dashboard defaults.

	Status status.Defaults `json:"-" yaml:"-"` // Track the Status of this source (version and regex misses).
}

// IsZero implements the yaml.IsZeroer interface.
func (d Defaults) IsZero() bool {
	return d.Options.IsZero() &&
		d.LatestVersion.IsZero() &&
		d.DeployedVersionLookup.IsZero() &&
		len(d.Notify) == 0 &&
		len(d.Command) == 0 &&
		len(d.WebHook) == 0 &&
		d.Dashboard.IsZero()
}

// DefaultsDecode is an unmarshal-only helper for [Defaults].
type DefaultsDecode struct {
	Options               opt.Defaults              `json:"options,omitempty" yaml:"options,omitempty"`
	DeployedVersionLookup deployedver_base.Defaults `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"`
	Notify                map[string]struct{}       `json:"notify,omitempty" yaml:"notify,omitempty"`
	Command               command.Commands          `json:"command,omitempty" yaml:"command,omitempty"`
	WebHook               map[string]struct{}       `json:"webhook,omitempty" yaml:"webhook,omitempty"`
	Dashboard             dashboard.Defaults        `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Use [DecodeDefaults] for a complete Defaults.
func (d *Defaults) UnmarshalJSON(data []byte) error {
	return d.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use [DecodeDefaults] for a complete Defaults.
func (d *Defaults) UnmarshalYAML(data []byte) error {
	return d.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (d *Defaults) unmarshal(format string, data []byte) error {
	aux := DefaultsDecode{
		Options:               d.Options,
		DeployedVersionLookup: d.DeployedVersionLookup,
		Notify:                d.Notify,
		Command:               d.Command,
		WebHook:               d.WebHook,
		Dashboard:             d.Dashboard,
	}

	// Unmarshal using the provided function.
	if err := decode.Unmarshal(format, data, &aux); err != nil {
		return err //nolint:wrapcheck
	}
	d.Options = aux.Options
	d.DeployedVersionLookup = aux.DeployedVersionLookup
	d.Notify = aux.Notify
	d.Command = aux.Command
	d.WebHook = aux.WebHook
	d.Dashboard = aux.Dashboard

	// LatestVersion.
	if err := polymorphic.Unmarshal(format, data, "latest_version", &d.LatestVersion); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

// DecodeDefaults creates and returns new [Defaults] from format-encoded data.
func DecodeDefaults(format string, data []byte) (*Defaults, error) {
	var field Defaults
	if err := decode.Unmarshal(format, data, &field); err != nil {
		return nil, &decode.KeyFieldError{
			Key: "service",
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

// Default sets the receiver to the default values.
func (d *Defaults) Default() {
	// Service.Options
	d.Options.Default()

	// Service.LatestVersion
	d.LatestVersion.Default()

	// Service.DeployedVersionLookup
	d.DeployedVersionLookup.Default()

	// Service.Dashboard
	d.Dashboard.Default()

	d.Init()
}

// SetDefaults assigns defaults to the receiver.
func (d *Defaults) SetDefaults(dflts *Defaults) {
	d.LatestVersion.SetDefaults(&dflts.LatestVersion)

	// Options.
	d.LatestVersion.Options = &d.Options
	d.DeployedVersionLookup.Options = &d.Options
	dflts.LatestVersion.Options = &dflts.Options
	dflts.DeployedVersionLookup.Options = &dflts.Options
}

// Init wires the appropriate Defaults pointers between structs.
func (d *Defaults) Init() {
	d.LatestVersion.Options = &d.Options
	d.DeployedVersionLookup.Options = &d.Options
}
