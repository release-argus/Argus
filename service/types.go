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
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	dvbase "github.com/release-argus/Argus/service/deployed_version/types/base"
	latestver "github.com/release-argus/Argus/service/latest_version"
	lvbase "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/polymorphic"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

// DefaultsConfig pairs soft and hard service defaults.
type DefaultsConfig struct {
	Soft *Defaults
	Hard *Defaults
}

// Services is a string map of Service.
type Services map[string]*Service

// Service is a source to track latest and deployed versions of a service.
// It also has the ability to run commands, send notifications and send WebHooks on new releases.
type Service struct {
	ID                    string             `json:"-" yaml:"-"`                                                   // Key/Name of the Service.
	Name                  string             `json:"name,omitempty" yaml:"name,omitempty"`                         // Name of the Service.
	Comment               string             `json:"comment,omitempty" yaml:"comment,omitempty"`                   // Comment on the Service.
	Options               opt.Options        `json:"options,omitempty" yaml:"options,omitempty"`                   // Options to give the Service.
	LatestVersion         latestver.Lookup   `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`     // Vars to scrape the latest version of the Service.
	DeployedVersionLookup deployedver.Lookup `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"` // Vars to scrape the Service's current deployed version.

	Notify             shoutrrr.Shoutrrrs `json:"notify,omitempty" yaml:"notify,omitempty"` // Service-specific Shoutrrr vars.
	NotifyFromDefaults bool               `json:"-" yaml:"-"`

	CommandController   *command.Controller `json:"-" yaml:"-"`                                 // The controller for the OS Commands that tracks fails and has the announce channel.
	Command             command.Commands    `json:"command,omitempty" yaml:"command,omitempty"` // OS Commands to run on new release.
	CommandFromDefaults bool                `json:"-" yaml:"-"`

	WebHook             webhook.WebHooks `json:"webhook,omitempty" yaml:"webhook,omitempty"` // Service-specific WebHook vars.
	WebHookFromDefaults bool             `json:"-" yaml:"-"`

	Dashboard dashboard.Options `json:"dashboard,omitempty" yaml:"dashboard,omitempty"` // Options for the dashboard.

	Status status.Status `json:"-" yaml:"-"` // Track the Status of this source (version and regex misses).

	Defaults     *Defaults `json:"-" yaml:"-"` // Default values.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hardcoded default values.
}

// serviceMarshal is a marshal-only helper for [Service].
type serviceMarshal struct {
	Name                  string             `json:"name,omitempty" yaml:"name,omitempty"`
	Comment               string             `json:"comment,omitempty" yaml:"comment,omitempty"`
	Options               opt.Options        `json:"options,omitempty" yaml:"options,omitempty"`
	LatestVersion         latestver.Lookup   `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`
	DeployedVersionLookup deployedver.Lookup `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"`
	Notify                shoutrrr.Shoutrrrs `json:"notify,omitempty" yaml:"notify,omitempty"`
	Command               command.Commands   `json:"command,omitempty" yaml:"command,omitempty"`
	WebHook               webhook.WebHooks   `json:"webhook,omitempty" yaml:"webhook,omitempty"`
	Dashboard             dashboard.Options  `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`
}

// serviceDecode is an unmarshal-only helper for [Service].
type serviceDecode struct {
	Name      string             `json:"name,omitempty" yaml:"name,omitempty"`
	Comment   string             `json:"comment,omitempty" yaml:"comment,omitempty"`
	Options   opt.Options        `json:"options,omitempty" yaml:"options,omitempty"`
	Notify    shoutrrr.Shoutrrrs `json:"notify,omitempty" yaml:"notify,omitempty"`
	Command   command.Commands   `json:"command,omitempty" yaml:"command,omitempty"`
	WebHook   webhook.WebHooks   `json:"webhook,omitempty" yaml:"webhook,omitempty"`
	Dashboard dashboard.Options  `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface.
func (s *Service) MarshalJSON() ([]byte, error) {
	return s.marshal("json")
}

// MarshalYAML implements the yaml.Marshaler interface.
func (s *Service) MarshalYAML() ([]byte, error) {
	return s.marshal("yaml")
}

// marshal implements the format.Marshaler interface.
func (s *Service) marshal(format string) ([]byte, error) {
	aux := serviceMarshal{
		Name:                  s.Name,
		Comment:               s.Comment,
		Options:               s.Options,
		LatestVersion:         s.LatestVersion,
		DeployedVersionLookup: s.DeployedVersionLookup,
		Dashboard:             s.Dashboard,
	}

	if !s.NotifyFromDefaults {
		aux.Notify = s.Notify
	}
	if !s.CommandFromDefaults {
		aux.Command = s.Command
	}
	if !s.WebHookFromDefaults {
		aux.WebHook = s.WebHook
	}

	return decode.Marshal(format, aux) //nolint:wrapcheck
}

// UnmarshalJSON implements the json.Marshaler interface.
// Use DecodeService for a full unmarshal.
func (s *Service) UnmarshalJSON(data []byte) error {
	return s.unmarshal("json", data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// Use DecodeService for a full unmarshal.
func (s *Service) UnmarshalYAML(data []byte) error {
	return s.unmarshal("yaml", data)
}

// unmarshal implements the format.Unmarshaler interface.
func (s *Service) unmarshal(format string, data []byte) error {
	aux := serviceDecode{
		Name:      s.Name,
		Comment:   s.Comment,
		Options:   s.Options,
		Notify:    s.Notify,
		Command:   s.Command,
		WebHook:   s.WebHook,
		Dashboard: s.Dashboard,
	}

	// Unmarshal in the given format.
	if err := decode.Unmarshal(format, data, &aux); err != nil {
		return err //nolint:wrapcheck
	}
	s.Name = aux.Name
	s.Comment = aux.Comment
	s.Options = aux.Options
	s.Options.SetDefaults(
		&s.Defaults.Options,
		&s.HardDefaults.Options,
	)
	s.Notify = aux.Notify
	s.Command = aux.Command
	s.WebHook = aux.WebHook
	s.Dashboard = aux.Dashboard
	s.Dashboard.SetDefaults(
		&s.Defaults.Dashboard,
		&s.HardDefaults.Dashboard,
	)

	// LatestVersion.
	if err := s.unmarshalLatestVersion(format, data); err != nil {
		return err
	}

	// DeployedVersionLookup.
	if err := s.unmarshalDeployedVersion(format, data); err != nil {
		return err
	}

	return nil
}

// unmarshalLatestVersion implements the format.Unmarshaler interface for [Service.LatestVersion].
func (s *Service) unmarshalLatestVersion(format string, data []byte) error {
	// Extract.
	lvRaw, err := polymorphic.Extract(format, data, "latest_version")
	if err != nil {
		return &decode.KeyFieldError{
			Key: "latest_version",
			Err: err,
		}
	}
	if len(lvRaw) == 0 {
		return nil
	}

	// Overrides.
	s.LatestVersion, err = latestver.ApplyOverrides(
		format, lvRaw,
		s.LatestVersion,
		&s.Options,
		&s.Status,
		lvbase.DefaultsConfig{
			Soft: &s.Defaults.LatestVersion,
			Hard: &s.HardDefaults.LatestVersion,
		},
	)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if s.LatestVersion != nil {
		s.LatestVersion.SetStatus(&s.Status)
	}

	return nil
}

// unmarshalDeployedVersion implements the format.Unmarshaler interface for [Service.DeployedVersion].
func (s *Service) unmarshalDeployedVersion(format string, data []byte) error {
	// Extract.
	dvRaw, err := polymorphic.Extract(format, data, "deployed_version")
	if err != nil {
		return &decode.KeyFieldError{
			Key: "deployed_version",
			Err: err,
		}
	}
	if len(dvRaw) == 0 {
		return nil
	}

	// Overrides.
	s.DeployedVersionLookup, err = deployedver.ApplyOverrides(
		format, dvRaw,
		s.DeployedVersionLookup,
		&s.Options,
		&s.Status,
		dvbase.DefaultsConfig{
			Soft: &s.Defaults.DeployedVersionLookup,
			Hard: &s.HardDefaults.DeployedVersionLookup,
		},
	)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if s.DeployedVersionLookup != nil {
		s.DeployedVersionLookup.SetStatus(&s.Status)
	}

	return nil
}

// String returns a string representation of the receiver.
func (s *Service) String(prefix string) string {
	if s == nil {
		return ""
	}
	return decode.ToYAMLString(s, prefix)
}

// Summary returns a [ServiceSummary] for the receiver.
func (s *Service) Summary() *apitype.ServiceSummary {
	if s == nil {
		return nil
	}

	var latestVersionType string
	if s.LatestVersion != nil {
		latestVersionType = s.LatestVersion.GetType()
	}
	hasDeployedVersionLookup := s.DeployedVersionLookup != nil

	svcInfo := s.Status.GetServiceInfo()
	summary := &apitype.ServiceSummary{
		ID:                       s.ID,
		Name:                     util.PtrIfNotZero(s.Name),
		Active:                   s.Options.Active,
		Comment:                  util.PtrIfNotZero(svcInfo.Comment),
		Type:                     latestVersionType,
		WebURL:                   util.PtrIfNotZero(svcInfo.WebURL),
		Icon:                     util.PtrIfNotZero(svcInfo.Icon),
		IconLinkTo:               util.PtrIfNotZero(svcInfo.IconLinkTo),
		HasDeployedVersionLookup: &hasDeployedVersionLookup,
		Status: &apitype.Status{
			ApprovedVersion:          svcInfo.ApprovedVersion,
			DeployedVersion:          svcInfo.DeployedVersion,
			DeployedVersionTimestamp: s.Status.DeployedVersionTimestamp(),
			LatestVersion:            svcInfo.LatestVersion,
			LatestVersionTimestamp:   s.Status.LatestVersionTimestamp(),
			LastQueried:              s.Status.LastQueried(),
		},
	}

	// Tags.
	if len(s.Dashboard.Tags) != 0 {
		summary.Tags = &s.Dashboard.Tags
	}

	// Command.
	if len(s.Command) != 0 {
		commands := len(s.Command)
		summary.Command = &commands
	}
	// WebHook.
	if len(s.WebHook) != 0 {
		webhooks := len(s.WebHook)
		summary.WebHook = &webhooks
	}

	return summary
}

// UsingDefaults returns whether the receiver is using the Notifiers/Commands/WebHooks from Defaults.
func (s *Service) UsingDefaults() (bool, bool, bool) {
	if s == nil {
		return false, false, false
	}
	return s.NotifyFromDefaults, s.CommandFromDefaults, s.WebHookFromDefaults
}

// GetName returns the [Service.Name] || [Service.ID].
func (s *Service) GetName() string {
	return util.ValueOr(s.Name, s.ID)
}
