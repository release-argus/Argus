// Copyright [2023] [Argus]
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

package service

import (
	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filter"
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
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

// Service is a source to be serviceed and provides everything needed to extract
// the latest version from the URL provided.
type Service struct {
	ID                    string              `yaml:"-" json:"name"`                                                // service_name
	Comment               string              `yaml:"comment,omitempty" json:"comment,omitempty"`                   // Comment on the Service
	Options               opt.Options         `yaml:"options,omitempty" json:"options,omitempty"`                   // Options to give the Service
	LatestVersion         latestver.Lookup    `yaml:"latest_version,omitempty" json:"latest_version,omitempty"`     // Vars to getting the latest version of the Service
	DeployedVersionLookup *deployedver.Lookup `yaml:"deployed_version,omitempty" json:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version
	Notify                shoutrrr.Slice      `yaml:"notify,omitempty" json:"notify,omitempty"`                     // Service-specific Shoutrrr vars
	CommandController     *command.Controller `yaml:"-" json:"-"`                                                   // The controller for the OS Commands that tracks fails and has the announce channel
	Command               command.Slice       `yaml:"command,omitempty" json:"command,omitempty"`                   // OS Commands to run on new release
	WebHook               webhook.Slice       `yaml:"webhook,omitempty" json:"webhook,omitempty"`                   // Service-specific WebHook vars
	Dashboard             DashboardOptions    `yaml:"dashboard,omitempty" json:"dashboard,omitempty"`               // Options for the dashboard
	Status                svcstatus.Status    `yaml:"-" json:"-"`                                                   // Track the Status of this source (version and regex misses)
	Defaults              *Service            `yaml:"-" json:"-"`                                                   // Default values
	HardDefaults          *Service            `yaml:"-" json:"-"`                                                   // Hardcoded default values

	// TODO: Deprecate
	OldStatus          *svcstatus.OldStatus    `yaml:"status,omitempty" json:"-"`              // For moving version info to argus.db
	Type               string                  `yaml:"type,omitempty" json:"-"`                // DEPRECATED - use latestver.type
	Active             *bool                   `yaml:"active,omitempty" json:"-"`              // DEPRECATED - use option.active
	Interval           *string                 `yaml:"interval,omitempty" json:"-"`            // DEPRECATED - use option.interval
	SemanticVersioning *bool                   `yaml:"semantic_versioning,omitempty" json:"-"` // DEPRECATED - use option.semantic_versioning
	URL                *string                 `yaml:"url,omitempty" json:"-"`                 // DEPRECATED - use latestver.url
	AllowInvalidCerts  *bool                   `yaml:"allow_invalid_certs,omitempty" json:"-"` // DEPRECATED - use latestver.allow_invalid_certs
	AccessToken        *string                 `yaml:"access_token,omitempty" json:"-"`        // DEPRECATED - use latestver.access_token
	UsePreRelease      *bool                   `yaml:"use_prerelease,omitempty" json:"-"`      // DEPRECATED - use latestver.use_prerelease
	URLCommands        *filter.URLCommandSlice `yaml:"url_commands,omitempty" json:"-"`        // DEPRECATED - use latestver.url_commands
	AutoApprove        *bool                   `yaml:"auto_approve,omitempty" json:"-"`        // DEPRECATED - use dashboard.auto_approve
	Icon               *string                 `yaml:"icon,omitempty" json:"-"`                // DEPRECATED - use dashboard.icon
	IconLinkTo         *string                 `yaml:"icon_link_to,omitempty" json:"-"`        // DEPRECATED - use dashboard.icon_link_to
	WebURL             *string                 `yaml:"web_url,omitempty" json:"-"`             // DEPRECATED - use dashboard.web_url
}

// String returns a string representation of the Service.
func (s *Service) String() string {
	if s == nil {
		return "<nil>"
	}
	yamlBytes, _ := yaml.Marshal(s)
	return string(yamlBytes)
}

func (s *Service) Convert() {
	didConvert := false
	if s.Type != "" {
		didConvert = true
		s.LatestVersion.Type = s.Type
		s.Type = ""
	}
	if s.Active != nil {
		didConvert = true
		s.Options.Active = s.Active
		s.Active = nil
	}
	if s.Interval != nil {
		didConvert = true
		s.Options.Interval = *s.Interval
		s.Interval = nil
	}
	if s.SemanticVersioning != nil {
		didConvert = true
		s.Options.SemanticVersioning = s.SemanticVersioning
		s.SemanticVersioning = nil
	}
	if s.URL != nil {
		didConvert = true
		s.LatestVersion.URL = *s.URL
		s.URL = nil
	}
	if s.AllowInvalidCerts != nil {
		didConvert = true
		s.LatestVersion.AllowInvalidCerts = s.AllowInvalidCerts
		s.AllowInvalidCerts = nil
	}
	if s.AccessToken != nil {
		didConvert = true
		s.LatestVersion.AccessToken = s.AccessToken
		s.AccessToken = nil
	}
	if s.UsePreRelease != nil {
		didConvert = true
		s.LatestVersion.UsePreRelease = s.UsePreRelease
		s.UsePreRelease = nil
	}
	if s.URLCommands != nil {
		didConvert = true
		s.LatestVersion.URLCommands = *s.URLCommands
		s.URLCommands = nil
	}
	if s.AutoApprove != nil {
		didConvert = true
		s.Dashboard.AutoApprove = s.AutoApprove
		s.AutoApprove = nil
	}
	if s.Icon != nil {
		didConvert = true
		s.Dashboard.Icon = *s.Icon
		s.Icon = nil
	}
	if s.IconLinkTo != nil {
		didConvert = true
		s.Dashboard.IconLinkTo = *s.IconLinkTo
		s.IconLinkTo = nil
	}
	if s.WebURL != nil {
		didConvert = true
		s.Dashboard.WebURL = *s.WebURL
		s.WebURL = nil
	}

	if didConvert {
		s.Status.SendSave()
	}
}

// Summary returns a ServiceSummary for the Service.
func (s *Service) Summary() *apitype.ServiceSummary {
	icon := s.GetIconURL()
	hasDeployedVersionLookup := s.DeployedVersionLookup != nil
	commands := len(s.Command)
	webhooks := len(s.WebHook)
	return &apitype.ServiceSummary{
		ID:                       s.ID,
		Active:                   s.Options.Active,
		Type:                     &s.LatestVersion.Type,
		WebURL:                   s.Status.GetWebURL(),
		Icon:                     &icon,
		IconLinkTo:               &s.Dashboard.IconLinkTo,
		HasDeployedVersionLookup: &hasDeployedVersionLookup,
		Command:                  &commands,
		WebHook:                  &webhooks,
		Status: &apitype.Status{
			ApprovedVersion:          s.Status.GetApprovedVersion(),
			DeployedVersion:          s.Status.GetDeployedVersion(),
			DeployedVersionTimestamp: s.Status.GetDeployedVersionTimestamp(),
			LatestVersion:            s.Status.GetLatestVersion(),
			LatestVersionTimestamp:   s.Status.GetLatestVersionTimestamp(),
			LastQueried:              s.Status.GetLastQueried()}}
}
