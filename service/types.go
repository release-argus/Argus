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
	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

var (
	jLog *util.JLog
)

// Slice is a slice mapping of Service.
type Slice map[string]*Service

// Defaults are the default values for a Service.
type Defaults struct {
	Options               opt.OptionsDefaults        `yaml:"options,omitempty" json:"options,omitempty"`                   // Options to give the Service
	LatestVersion         latestver.LookupDefaults   `yaml:"latest_version,omitempty" json:"latest_version,omitempty"`     // Vars to getting the latest version of the Service
	DeployedVersionLookup deployedver.LookupDefaults `yaml:"deployed_version,omitempty" json:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version
	Notify                map[string]struct{}        `yaml:"notify,omitempty" json:"notify,omitempty"`                     // Default Notify's to give the Service
	Command               command.Slice              `yaml:"command,omitempty" json:"command,omitempty"`                   // Default Command's to give the Service
	WebHook               map[string]struct{}        `yaml:"webhook,omitempty" json:"webhook,omitempty"`                   // Default WebHook's to give the Servic
	Dashboard             DashboardOptionsDefaults   `yaml:"dashboard,omitempty" json:"dashboard,omitempty"`               // Dashboard defaults

	Status svcstatus.StatusDefaults `yaml:"-" json:"-"` // Track the Status of this source (version and regex misses)
}

// Service is a source to be serviceed and provides everything needed to extract
// the latest version from the URL provided.
type Service struct {
	ID                    string              `yaml:"-" json:"name"`                                                // service_name
	Comment               string              `yaml:"comment,omitempty" json:"comment,omitempty"`                   // Comment on the Service
	Options               opt.Options         `yaml:"options,omitempty" json:"options,omitempty"`                   // Options to give the Service
	LatestVersion         latestver.Lookup    `yaml:"latest_version,omitempty" json:"latest_version,omitempty"`     // Vars to getting the latest version of the Service
	DeployedVersionLookup *deployedver.Lookup `yaml:"deployed_version,omitempty" json:"deployed_version,omitempty"` // Var to scrape the Service's current deployed version
	Notify                shoutrrr.Slice      `yaml:"notify,omitempty" json:"notify,omitempty"`                     // Service-specific Shoutrrr vars
	notifyFromDefaults    bool
	CommandController     *command.Controller `yaml:"-" json:"-"`                                 // The controller for the OS Commands that tracks fails and has the announce channel
	Command               command.Slice       `yaml:"command,omitempty" json:"command,omitempty"` // OS Commands to run on new release
	commandFromDefaults   bool
	WebHook               webhook.Slice `yaml:"webhook,omitempty" json:"webhook,omitempty"` // Service-specific WebHook vars
	webhookFromDefaults   bool
	Dashboard             DashboardOptions `yaml:"dashboard,omitempty" json:"dashboard,omitempty"` // Options for the dashboard

	Status svcstatus.Status `yaml:"-" json:"-"` // Track the Status of this source (version and regex misses)

	Defaults     *Defaults `yaml:"-" json:"-"` // Default values
	HardDefaults *Defaults `yaml:"-" json:"-"` // Hardcoded default values
}

// String returns a string representation of the Service.
func (s *Service) String(prefix string) (str string) {
	if s != nil {
		str = util.ToYAMLString(s, prefix)
	}
	return
}

// Summary returns a ServiceSummary for the Service.
func (s *Service) Summary() (summary *apitype.ServiceSummary) {
	if s == nil {
		return
	}

	icon := s.IconURL()
	hasDeployedVersionLookup := s.DeployedVersionLookup != nil
	commands := len(s.Command)
	webhooks := len(s.WebHook)
	summary = &apitype.ServiceSummary{
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
			ApprovedVersion:          s.Status.ApprovedVersion(),
			DeployedVersion:          s.Status.DeployedVersion(),
			DeployedVersionTimestamp: s.Status.DeployedVersionTimestamp(),
			LatestVersion:            s.Status.LatestVersion(),
			LatestVersionTimestamp:   s.Status.LatestVersionTimestamp(),
			LastQueried:              s.Status.LastQueried()}}
	return
}

// UsingDefaults returns whether the Service is using the Notify(s)/Command(s)/WebHook(s) from Defaults
func (s *Service) UsingDefaults() (bool, bool, bool) {
	if s == nil {
		return false, false, false
	}
	return s.notifyFromDefaults, s.commandFromDefaults, s.webhookFromDefaults
}
