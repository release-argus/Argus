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

// Package service provides the service functionality for Argus.
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/notify/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
	"gopkg.in/yaml.v3"
)

var (
	jLog *util.JLog
)

// Slice is a slice mapping of Service.
type Slice map[string]*Service

// UnmarshalJSON handles the unmarshalling of a Slice.
func (s *Slice) UnmarshalJSON(data []byte) error {
	var aux map[string]*Service

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal Slice:\n%w", err)
	}
	*s = aux

	s.giveIDs()

	return nil
}

// UnmarshalYAML handles the unmarshalling of a Slice.
func (s *Slice) UnmarshalYAML(value *yaml.Node) error {
	var aux map[string]*Service

	// Unmarshal YAML data.
	if err := value.Decode(&aux); err != nil {
		return fmt.Errorf("failed to unmarshal Slice:\n%w", err)
	}
	*s = aux

	s.giveIDs()

	return nil
}

// giveIDs gives the Services their IDs if they don't have one.
func (s *Slice) giveIDs() {
	for id, service := range *s {
		// Remove the service if nil.
		if service == nil {
			delete(*s, id)
			continue
		}

		service.ID = id
		// Default Name to ID.
		if service.Name == "" {
			service.Name = id
		}
	}
}

// Defaults are the default values for a Service.
type Defaults struct {
	Options               opt.Defaults             `yaml:"options,omitempty" json:"options,omitempty"`                   // Options to give the Service.
	LatestVersion         latestver_base.Defaults  `yaml:"latest_version,omitempty" json:"latest_version,omitempty"`     // Vars to scrape the latest version of the Service.
	DeployedVersionLookup deployedver.Defaults     `yaml:"deployed_version,omitempty" json:"deployed_version,omitempty"` // Vars to scrape the Service's current deployed version.
	Notify                map[string]struct{}      `yaml:"notify,omitempty" json:"notify,omitempty"`                     // Default Notifiers to give a Service.
	Command               command.Slice            `yaml:"command,omitempty" json:"command,omitempty"`                   // Default Commands to give a Service.
	WebHook               map[string]struct{}      `yaml:"webhook,omitempty" json:"webhook,omitempty"`                   // Default WebHooks to give a Service.
	Dashboard             DashboardOptionsDefaults `yaml:"dashboard,omitempty" json:"dashboard,omitempty"`               // Dashboard defaults.

	Status status.Defaults `yaml:"-" json:"-"` // Track the Status of this source (version and regex misses).
}

// Service is a source to track latest and deployed versions of a service.
// It also has the ability to run commands, send notifications and send WebHooks on new releases.
type Service struct {
	ID                    string              `yaml:"-" json:"-"`                                                   // Key/Name of the Service.
	Name                  string              `yaml:"-" json:"-"`                                                   // Name of the Service.
	marshalName           bool                ``                                                                    // Whether to marshal the Name.
	Comment               string              `yaml:"-" json:"-"`                                                   // Comment on the Service.
	Options               opt.Options         `yaml:"-" json:"-"`                                                   // Options to give the Service.
	LatestVersion         latestver.Lookup    `yaml:"-" json:"-"`                                                   // Vars to scrape the latest version of the Service.
	DeployedVersionLookup *deployedver.Lookup `yaml:"deployed_version,omitempty" json:"deployed_version,omitempty"` // Vars to scrape the Service's current deployed version.
	Notify                shoutrrr.Slice      `yaml:"notify,omitempty" json:"notify,omitempty"`                     // Service-specific Shoutrrr vars.
	notifyFromDefaults    bool
	CommandController     *command.Controller `yaml:"-" json:"-"`                                 // The controller for the OS Commands that tracks fails and has the announce channel.
	Command               command.Slice       `yaml:"command,omitempty" json:"command,omitempty"` // OS Commands to run on new release.
	commandFromDefaults   bool
	WebHook               webhook.Slice `yaml:"webhook,omitempty" json:"webhook,omitempty"` // Service-specific WebHook vars.
	webhookFromDefaults   bool
	Dashboard             DashboardOptions `yaml:"dashboard,omitempty" json:"dashboard,omitempty"` // Options for the dashboard.

	Status status.Status `yaml:"-" json:"-"` // Track the Status of this source (version and regex misses).

	Defaults     *Defaults `yaml:"-" json:"-"` // Default values.
	HardDefaults *Defaults `yaml:"-" json:"-"` // Hardcoded default values.
}

// MarshalName returns whether the Name should be marshaled.
// (explicitly set in the config)
func (s *Service) MarshalName() bool {
	return s.marshalName
}

// String returns a string representation of the Service.
func (s *Service) String(prefix string) string {
	if s == nil {
		return ""
	}
	return util.ToYAMLString(s, prefix)
}

// Summary returns a ServiceSummary for the Service.
func (s *Service) Summary() *apitype.ServiceSummary {
	if s == nil {
		return nil
	}

	var latestVersionType string
	if s.LatestVersion != nil {
		latestVersionType = s.LatestVersion.GetType()
	}
	hasDeployedVersionLookup := s.DeployedVersionLookup != nil
	commands := len(s.Command)
	webhooks := len(s.WebHook)

	summary := &apitype.ServiceSummary{
		ID:                       s.ID,
		Active:                   s.Options.Active,
		Type:                     latestVersionType,
		WebURL:                   s.Status.GetWebURL(),
		Icon:                     s.IconURL(),
		IconLinkTo:               s.Dashboard.IconLinkTo,
		HasDeployedVersionLookup: &hasDeployedVersionLookup,
		Command:                  commands,
		WebHook:                  webhooks,
		Status: &apitype.Status{
			ApprovedVersion:          s.Status.ApprovedVersion(),
			DeployedVersion:          s.Status.DeployedVersion(),
			DeployedVersionTimestamp: s.Status.DeployedVersionTimestamp(),
			LatestVersion:            s.Status.LatestVersion(),
			LatestVersionTimestamp:   s.Status.LatestVersionTimestamp(),
			LastQueried:              s.Status.LastQueried()}}

	// Name
	if s.MarshalName() {
		summary.Name = &s.Name
	}

	return summary
}

// UsingDefaults returns whether the Service is using the Notify(s)/Command(s)/WebHook(s) from Defaults.
func (s *Service) UsingDefaults() (bool, bool, bool) {
	if s == nil {
		return false, false, false
	}
	return s.notifyFromDefaults, s.commandFromDefaults, s.webhookFromDefaults
}

// UnmarshalJSON handles the unmarshalling of a Service.
//
// This addresses the dynamic LatestVersion type.
func (s *Service) UnmarshalJSON(data []byte) error {
	// Alias to avoid recursion.
	type Alias Service
	aux := &struct {
		*Alias        `json:",inline"`
		Name          *string         `json:"name,omitempty"`           // Name of the Service.
		Comment       *string         `json:"comment,omitempty"`        // Comment on the Service.
		Options       *opt.Options    `json:"options,omitempty"`        // Options to give the Service.
		LatestVersion json.RawMessage `json:"latest_version,omitempty"` // Temp LatestVersion field to get Type.
	}{
		Alias:   (*Alias)(s),
		Name:    &s.Name,
		Comment: &s.Comment,
		Options: &s.Options,
	}

	// Unmarshal into aux to separate the latest_version field.
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal Service:\n%w", err)
	}

	// Name
	if s.Name != "" {
		s.marshalName = true
	}

	// -- Dynamic LatestVersion type --
	if aux.LatestVersion == nil {
		return nil
	}

	// Check for the type field in the JSON.
	var typeField struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(aux.LatestVersion, &typeField); err != nil {
		return fmt.Errorf("error in latest_version field:\ntype: <invalid> (%s)",
			strings.Replace(err.Error(), "json: ", "", 1))
	}

	var err error
	// If we don't have a type in the JSON, check if we already have a type in the struct.
	if typeField.Type == "" {
		// Have a LatestVersion struct, so unmarshal into it.
		if s.LatestVersion != nil {
			err = json.Unmarshal(aux.LatestVersion, s.LatestVersion)
		} else
		// No LatestVersion struct and the type remains absent.
		{
			err = fmt.Errorf("type: <required> [%s]",
				strings.Join(latestver.PossibleTypes, ", "))
		}
		if err != nil {
			return fmt.Errorf("error in latest_version field:\n%w", err)
		}
	} else
	// We have a type in the JSON, so we can unmarshal it.
	{
		if s.LatestVersion, err = latestver.UnmarshalJSON(aux.LatestVersion); err != nil {
			return fmt.Errorf("error in latest_version field:\n%w", err)
		}
	}

	return nil
}

// MarshalJSON handles the marshalling of a Service.
//
// (dynamic typing).
func (s *Service) MarshalJSON() ([]byte, error) {
	// Alias to avoid recursion.
	type Alias Service
	aux := &struct {
		Name          string           `json:"name,omitempty"`           // Name of the Service.
		Comment       string           `json:"comment,omitempty"`        // Comment on the Service.
		Options       opt.Options      `json:"options,omitempty"`        // Options to give the Service.
		LatestVersion latestver.Lookup `json:"latest_version,omitempty"` // Vars to getting the latest version of the Service.
		*Alias        `json:",inline"`
	}{
		Name:          s.Name,
		Comment:       s.Comment,
		Options:       s.Options,
		LatestVersion: s.LatestVersion,
		Alias:         (*Alias)(s),
	}

	if !s.MarshalName() {
		aux.Name = ""
	}

	return json.Marshal(aux) //nolint:wrapcheck
}

// UnmarshalYAML handles the unmarshalling of a Service.
//
// (dynamic typing).
func (s *Service) UnmarshalYAML(value *yaml.Node) error {
	// Alias to avoid recursion.
	type Alias Service
	aux := &struct {
		*Alias        `yaml:",inline"`
		Name          *string      `yaml:"name,omitempty"`           // Name of the Service.
		Comment       *string      `yaml:"comment,omitempty"`        // Comment on the Service.
		Options       *opt.Options `yaml:"options,omitempty"`        // Options to give the Service.
		LatestVersion RawNode      `yaml:"latest_version,omitempty"` // Temp LatestVersion field to get Type.
	}{
		Alias:   (*Alias)(s),
		Name:    &s.Name,
		Comment: &s.Comment,
		Options: &s.Options,
	}

	// Unmarshal into aux to separate the latest_version field.
	if err := value.Decode(&aux); err != nil {
		return fmt.Errorf("failed to unmarshal Service:\n%w", err)
	}

	// Name
	if s.Name != "" {
		s.marshalName = true
	}

	// -- Dynamic LatestVersion type --
	if aux.LatestVersion.Node == nil {
		return nil
	}

	// Check for the type field in the YAML.
	var typeField struct {
		Type string `json:"type"`
	}
	if err := aux.LatestVersion.Decode(&typeField); err != nil {
		return fmt.Errorf("error in latest_version field:\ntype: <invalid> (%q)",
			strings.Replace(err.Error(), "yaml: unmarshal errors:\n  ", "", 1))
	}

	var err error
	// If we don't have a type in the YAML, check if we already have a type in the struct.
	if typeField.Type == "" {
		// Have a LatestVersion struct, so unmarshal into it.
		if s.LatestVersion != nil {
			err = aux.LatestVersion.Decode(s.LatestVersion)
		} else
		// No LatestVersion struct and the type remains absent, so error.
		{
			err = fmt.Errorf("type: <required> [%s]",
				strings.Join(latestver.PossibleTypes, ", "))
		}
		if err != nil {
			return fmt.Errorf("error in latest_version field:\n%w", err)
		}
	} else
	// We have a type in the YAML, so we can unmarshal it.
	{
		// Validate the type and create the appropriate Lookup instance.
		if s.LatestVersion, err = latestver.New(
			typeField.Type,
			"yaml", aux.LatestVersion.Node,
			nil,
			nil,
			nil, nil); err != nil {
			return fmt.Errorf("error in latest_version field:\n%w", err)
		}
	}

	return nil
}

// RawNode is a struct that holds a *yaml.Node.
type RawNode struct{ *yaml.Node }

// UnmarshalYAML handles the unmarshalling of a RawNode.
func (n *RawNode) UnmarshalYAML(node *yaml.Node) error {
	n.Node = node
	return nil
}

// MarshalYAML handles the marshalling of a Service.
//
// (dynamic typing).
func (s *Service) MarshalYAML() (interface{}, error) {
	// Alias to avoid recursion.
	type Alias Service
	aux := &struct {
		Name          string           `yaml:"name,omitempty"`           // Name of the Service.
		Comment       string           `yaml:"comment,omitempty"`        // Comment on the Service.
		Options       opt.Options      `yaml:"options,omitempty"`        // Options to give the Service.
		LatestVersion latestver.Lookup `yaml:"latest_version,omitempty"` // Vars to getting the latest version of the Service.
		*Alias        `yaml:",inline"`
	}{
		Name:          s.Name,
		Comment:       s.Comment,
		Options:       s.Options,
		LatestVersion: s.LatestVersion,
		Alias:         (*Alias)(s),
	}

	if !s.MarshalName() {
		aux.Name = ""
	}

	return aux, nil
}
