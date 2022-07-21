// Copyright [2022] [Argus]
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
	"github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filters"
	"github.com/release-argus/Argus/service/options"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

var (
	jLog *utils.JLog
)

// Slice is a slice mapping of Service.
type Slice map[string]*Service

// Service is a source to be serviceed and provides everything needed to extract
// the latest version from the URL provided.
type Service struct {
	ID                    string                   `yaml:"-"`                          // service_name
	Type                  string                   `yaml:"type,omitempty"`             // service_name
	Comment               *string                  `yaml:"comment,omitempty"`          // Comment on the Service
	Options               options.Options          `yaml:"options,omitempty"`          // Options to give the Service
	LatestVersion         latest_version.Lookup    `yaml:"latest_version,omitempty"`   // Vars to getting the latest version of the Service
	DeployedVersionLookup *deployed_version.Lookup `yaml:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version
	CommandController     *command.Controller      `yaml:"-"`                          // The controller for the OS Commands that tracks fails and has the announce channel
	Command               *command.Slice           `yaml:"command,omitempty"`          // OS Commands to run on new release
	WebHook               *webhook.Slice           `yaml:"webhook,omitempty"`          // Service-specific WebHook vars
	Notify                *shoutrrr.Slice          `yaml:"notify,omitempty"`           // Service-specific Shoutrrr vars
	Dashboard             DashboardOptions         `yaml:"dashboard,omitempty"`        // Options for the dashboard
	Status                *service_status.Status   `yaml:"-"`                          // Track the Status of this source (version and regex misses)
	HardDefaults          *Service                 `yaml:"-"`                          // Hardcoded default values
	Defaults              *Service                 `yaml:"-"`                          // Default values

	// TODO: Deprecate
	OldStatus          *service_status.OldStatus `yaml:"status,omitempty"`              // For moving version info to argus.db
	Active             *bool                     `yaml:"active,omitempty"`              // options.active
	Interval           *string                   `yaml:"interval,omitempty"`            // options.interval
	SemanticVersioning *bool                     `yaml:"semantic_versioning,omitempty"` // options.semantic_versioning
	URL                *string                   `yaml:"url,omitempty"`                 // latest_version.url
	AllowInvalidCerts  *bool                     `yaml:"allow_invalid_certs,omitempty"` // latest_version.allow_invalid_certs
	AccessToken        *string                   `yaml:"access_token,omitempty"`        // latest_version.access_token
	UsePreRelease      *bool                     `yaml:"use_prerelease,omitempty"`      // latest_version.use_prerelease
	URLCommands        *filters.URLCommandSlice  `yaml:"url_commands,omitempty"`        // latest_version.url_commands
	AutoApprove        *bool                     `yaml:"auto_approve,omitempty"`        // dashboard.auto_approve
	Icon               *string                   `yaml:"icon,omitempty"`                // dashboard.icon
	IconLinkTo         *string                   `yaml:"icon_link_to,omitempty"`        // dashboard.icon_link_to
	WebURL             *string                   `yaml:"web_url,omitempty"`             // dashboard.web_url
}

func (s *Service) convert() {
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
		s.Options.Interval = s.Interval
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
		s.LatestVersion.AccessToken = *s.AccessToken
		s.AccessToken = nil
	}
	if s.UsePreRelease != nil {
		didConvert = true
		s.LatestVersion.UsePreRelease = s.UsePreRelease
		s.UsePreRelease = nil
	}
	if s.URLCommands != nil {
		didConvert = true
		s.LatestVersion.URLCommands = s.URLCommands
		s.URLCommands = nil
	}
	s.Dashboard = DashboardOptions{}
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
		*s.Status.SaveChannel <- true
	}
}
