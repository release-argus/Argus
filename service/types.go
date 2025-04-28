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
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/notify/shoutrrr"
	"github.com/release-argus/Argus/service/dashboard"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	deployedver_base "github.com/release-argus/Argus/service/deployed_version/types/base"
	latestver "github.com/release-argus/Argus/service/latest_version"
	latestver_base "github.com/release-argus/Argus/service/latest_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

// Slice is a slice mapping of Service.
type Slice map[string]*Service

// UnmarshalJSON handles the unmarshalling of a Slice.
// It unmarshals the JSON data into a map of string keys to Service pointers,
// and then calls the giveIDs method to assign IDs to the services.
func (s *Slice) UnmarshalJSON(data []byte) error {
	var aux map[string]*Service

	if err := json.Unmarshal(data, &aux); err != nil {
		errStr := util.FormatUnmarshalError("json", err)
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
		return errors.New("failed to unmarshal service.Slice:\n  " + errStr)
	}
	*s = aux

	s.giveIDs()

	return nil
}

// UnmarshalJSON handles the unmarshalling of a Slice.
// It unmarshals the YAML data into a map of string keys to Service pointers,
// and then calls the giveIDs method to assign IDs to the services.
func (s *Slice) UnmarshalYAML(value *yaml.Node) error {
	var aux map[string]*Service

	// Unmarshal YAML data.
	if err := value.Decode(&aux); err != nil {
		errStr := util.FormatUnmarshalError("yaml", err)
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
		return errors.New("failed to unmarshal service.Slice:\n  " + errStr)
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
	Options               opt.Defaults              `json:"options,omitempty" yaml:"options,omitempty"`                   // Options to give the Service.
	LatestVersion         latestver_base.Defaults   `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`     // Vars to scrape the latest version of the Service.
	DeployedVersionLookup deployedver_base.Defaults `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"` // Vars to scrape the Service's current deployed version.
	Notify                map[string]struct{}       `json:"notify,omitempty" yaml:"notify,omitempty"`                     // Default Notifiers to give a Service.
	Command               command.Slice             `json:"command,omitempty" yaml:"command,omitempty"`                   // Default Commands to give a Service.
	WebHook               map[string]struct{}       `json:"webhook,omitempty" yaml:"webhook,omitempty"`                   // Default WebHooks to give a Service.
	Dashboard             dashboard.OptionsDefaults `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`               // Dashboard defaults.

	Status status.Defaults `json:"-" yaml:"-"` // Track the Status of this source (version and regex misses).
}

// Service is a source to track latest and deployed versions of a service.
// It also has the ability to run commands, send notifications and send WebHooks on new releases.
type Service struct {
	ID                    string             `json:"-" yaml:"-"`                               // Key/Name of the Service.
	Name                  string             `json:"-" yaml:"-"`                               // Name of the Service.
	marshalName           bool               ``                                                // Whether to marshal the Name.
	Comment               string             `json:"-" yaml:"-"`                               // Comment on the Service.
	Options               opt.Options        `json:"-" yaml:"-"`                               // Options to give the Service.
	LatestVersion         latestver.Lookup   `json:"-" yaml:"-"`                               // Vars to scrape the latest version of the Service.
	DeployedVersionLookup deployedver.Lookup `json:"-" yaml:"-"`                               // Vars to scrape the Service's current deployed version.
	Notify                shoutrrr.Slice     `json:"notify,omitempty" yaml:"notify,omitempty"` // Service-specific Shoutrrr vars.
	notifyFromDefaults    bool
	CommandController     *command.Controller `json:"-" yaml:"-"`                                 // The controller for the OS Commands that tracks fails and has the announce channel.
	Command               command.Slice       `json:"command,omitempty" yaml:"command,omitempty"` // OS Commands to run on new release.
	commandFromDefaults   bool
	WebHook               webhook.Slice `json:"webhook,omitempty" yaml:"webhook,omitempty"` // Service-specific WebHook vars.
	webhookFromDefaults   bool
	Dashboard             dashboard.Options `json:"dashboard,omitempty" yaml:"dashboard,omitempty"` // Options for the dashboard.

	Status status.Status `json:"-" yaml:"-"` // Track the Status of this source (version and regex misses).

	Defaults     *Defaults `json:"-" yaml:"-"` // Default values.
	HardDefaults *Defaults `json:"-" yaml:"-"` // Hardcoded default values.
}

// MarshalName returns whether the Name should be marshalled.
// (explicitly set in the config).
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

	serviceInfo := s.Status.GetServiceInfo()
	summary := &apitype.ServiceSummary{
		ID:                       s.ID,
		Active:                   s.Options.Active,
		Type:                     latestVersionType,
		HasDeployedVersionLookup: &hasDeployedVersionLookup,
		Status: &apitype.Status{
			ApprovedVersion:          serviceInfo.ApprovedVersion,
			DeployedVersion:          serviceInfo.DeployedVersion,
			DeployedVersionTimestamp: s.Status.DeployedVersionTimestamp(),
			LatestVersion:            serviceInfo.LatestVersion,
			LatestVersionTimestamp:   s.Status.LatestVersionTimestamp(),
			LastQueried:              s.Status.LastQueried()}}

	// Icon.
	if serviceInfo.Icon != "" {
		summary.Icon = &serviceInfo.Icon
	}
	// IconLinkTo.
	if serviceInfo.IconLinkTo != "" {
		summary.IconLinkTo = &serviceInfo.IconLinkTo
	}
	// WebURL.
	if serviceInfo.WebURL != "" {
		summary.WebURL = &serviceInfo.WebURL
	}

	// Name.
	if s.MarshalName() {
		summary.Name = &s.Name
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

// UsingDefaults returns whether the Service is using the Notify(s)/Command(s)/WebHook(s) from Defaults.
func (s *Service) UsingDefaults() (bool, bool, bool) {
	if s == nil {
		return false, false, false
	}
	return s.notifyFromDefaults, s.commandFromDefaults, s.webhookFromDefaults
}

// unmarshalVersionLookups handles the unmarshalling of LatestVersion and DeployedVersion fields.
func (s *Service) unmarshalVersionLookups(
	format string, // "json" | "yaml"
	latestVersion, deployedVersion any,
) error {
	// -- Dynamic LatestVersion type --
	if latestVersion != nil {
		lookupType, err := extractLookupType(
			format, latestVersion,
			s.LatestVersion)
		if err != nil {
			errStr := util.FormatUnmarshalError(format, err)
			errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
			return errors.New("failed to unmarshal service.Service.LatestVersion:\n  " + errStr)
		}
		s.LatestVersion, err = latestver.New(
			lookupType,
			format, latestVersion,
			nil,
			nil,
			nil, nil)
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	// -- Dynamic DeployedVersion type --
	if deployedVersion != nil {
		lookupType, err := extractLookupType(
			format, deployedVersion,
			s.DeployedVersionLookup)
		if err != nil {
			errStr := util.FormatUnmarshalError(format, err)
			errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
			return errors.New("failed to unmarshal service.Service.DeployedVersion:\n  " + errStr)
		}
		if format == "yaml" && lookupType == "" {
			// Default to url for YAML only
			lookupType = "url"
		}

		s.DeployedVersionLookup, err = deployedver.New(
			lookupType,
			format, deployedVersion,
			nil,
			nil,
			nil, nil)
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	return nil
}

type structWithGetType interface {
	GetType() string
}

// extractLookupType extracts the type field from the YAML,
// or uses the GetType from the struct if it's not in the YAML,
// and the struct is non-nil.
func extractLookupType(
	dataFormat string,
	data any,
	lookup structWithGetType,
) (string, error) {
	// Check for the type field in the YAML.
	var typeField struct {
		Type string `yaml:"type"`
	}
	var err error
	switch v := data.(type) {
	case *yaml.Node:
		err = v.Decode(&typeField)
	case json.RawMessage:
		err = json.Unmarshal(v, &typeField)
	}
	if err != nil {
		return "", fmt.Errorf("invalid %s:\n%s",
			dataFormat, strings.TrimPrefix(err.Error(), dataFormat+": "))
	}

	if typeField.Type != "" {
		return typeField.Type, nil
	}

	// If we don't have a type in the YAML, check if we already have a type in the struct.
	if lookup != nil {
		return lookup.GetType(), nil
	}

	// Invalid, but let the parent function handle it.
	return "", nil
}

// UnmarshalJSON handles the unmarshalling of a Service.
//
// This addresses the dynamic Latest/Deployed Version types.
func (s *Service) UnmarshalJSON(data []byte) error {
	// Alias to avoid recursion.
	type Alias Service
	aux := &struct {
		*Alias          `json:",inline"` // Embed the original struct.
		Name            *string          `json:"name,omitempty"`             // Name of the Service.
		Comment         *string          `json:"comment,omitempty"`          // Comment on the Service.
		Options         *opt.Options     `json:"options,omitempty"`          // Options to give the Service.
		LatestVersion   json.RawMessage  `json:"latest_version,omitempty"`   // Temp LatestVersion field to get Type.
		DeployedVersion json.RawMessage  `json:"deployed_version,omitempty"` // Temp DeployedVersion field to get Type.
	}{
		Alias:   (*Alias)(s),
		Name:    &s.Name,
		Comment: &s.Comment,
		Options: &s.Options,
	}

	// Unmarshal into aux to separate the latest_version field.
	if err := json.Unmarshal(data, &aux); err != nil {
		errStr := util.FormatUnmarshalError("json", err)
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
		return errors.New("failed to unmarshal service.Service:\n  " + errStr)
	}

	// Name.
	if s.Name != "" {
		s.marshalName = true
	}

	var latestVersionNode, deployedVersionNode any
	if aux.LatestVersion != nil {
		latestVersionNode = aux.LatestVersion
	}
	if aux.DeployedVersion != nil {
		deployedVersionNode = aux.DeployedVersion
	}

	return s.unmarshalVersionLookups(
		"json",
		latestVersionNode,
		deployedVersionNode)
}

// UnmarshalYAML handles the unmarshalling of a Service.
//
// This addresses the dynamic Latest/Deployed Version types.
func (s *Service) UnmarshalYAML(value *yaml.Node) error {
	// Alias to avoid recursion.
	type Alias Service
	aux := &struct {
		*Alias          `yaml:",inline"` // Embed the original struct.
		Name            *string          `yaml:"name,omitempty"`             // Name of the Service.
		Comment         *string          `yaml:"comment,omitempty"`          // Comment on the Service.
		Options         *opt.Options     `yaml:"options,omitempty"`          // Options to give the Service.
		LatestVersion   util.RawNode     `yaml:"latest_version,omitempty"`   // Temp LatestVersion field to get Type.
		DeployedVersion util.RawNode     `yaml:"deployed_version,omitempty"` // Temp DeployedVersion field to get Type.
	}{
		Alias:   (*Alias)(s),
		Name:    &s.Name,
		Comment: &s.Comment,
		Options: &s.Options,
	}

	// Unmarshal into aux to separate the latest_version field.
	if err := value.Decode(&aux); err != nil {
		errStr := util.FormatUnmarshalError("yaml", err)
		errStr = strings.ReplaceAll(errStr, "\n", "\n  ")
		return errors.New("failed to unmarshal service.Service:\n  " + errStr)
	}

	// Name.
	if s.Name != "" {
		s.marshalName = true
	}

	var latestVersionNode, deployedVersionNode any
	if aux.LatestVersion.Node != nil {
		latestVersionNode = aux.LatestVersion.Node
	}
	if aux.DeployedVersion.Node != nil {
		deployedVersionNode = aux.DeployedVersion.Node
	}

	return s.unmarshalVersionLookups(
		"yaml",
		latestVersionNode,
		deployedVersionNode)
}

// MarshalJSON handles the marshalling of a Service.
func (s *Service) MarshalJSON() ([]byte, error) {
	result, err := s.marshal(func(v any) (any, error) {
		return json.Marshal(v) //nolint:wrapcheck
	})
	return result.([]byte), err
}

// MarshalYAML handles the marshalling of a Service.
func (s *Service) MarshalYAML() (any, error) {
	return s.marshal(func(v any) (any, error) {
		return v, nil
	})
}

// marshal is a shared function for marshalling a Service.
func (s *Service) marshal(marshalFunc func(any) (any, error)) (any, error) {
	// Alias to avoid recursion.
	type Alias Service
	aux := &struct {
		Name            string                          `json:"name,omitempty" yaml:"name,omitempty"`                         // Name of the Service.
		Comment         string                          `json:"comment,omitempty" yaml:"comment,omitempty"`                   // Comment on the Service.
		Options         opt.Options                     `json:"options,omitempty" yaml:"options,omitempty"`                   // Options to give the Service.
		LatestVersion   latestver.Lookup                `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`     // Vars to scrape the latest version of the Service.
		DeployedVersion deployedver.Lookup              `json:"deployed_version,omitempty" yaml:"deployed_version,omitempty"` // Vars to scrape the deployed version of the Service.
		*Alias          `json:",inline" yaml:",inline"` // Embed the original struct.
	}{
		Name:            s.Name,
		Comment:         s.Comment,
		Options:         s.Options,
		LatestVersion:   s.LatestVersion,
		DeployedVersion: s.DeployedVersionLookup,
		Alias:           (*Alias)(s),
	}

	if !s.MarshalName() {
		aux.Name = ""
	}

	return marshalFunc(aux)
}
