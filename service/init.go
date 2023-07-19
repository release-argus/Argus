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
	"strings"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
	deployedver.LogInit(log)
	latestver.LogInit(log)
}

// Init will initialise the Service metric.
func (s *Service) Init(
	defaults *Defaults,
	hardDefaults *Defaults,

	rootNotifyConfig *shoutrrr.SliceDefaults,
	notifyDefaults *shoutrrr.SliceDefaults,
	notifyHardDefaults *shoutrrr.SliceDefaults,

	rootWebHookConfig *webhook.SliceDefaults,
	webhookDefaults *webhook.WebHookDefaults,
	webhookHardDefaults *webhook.WebHookDefaults,
) {
	// Status
	s.Status.Init(
		len(s.Notify), len(s.Command), len(s.WebHook),
		&s.ID,
		&s.Dashboard.WebURL)

	// Service
	s.Defaults = defaults
	s.HardDefaults = hardDefaults
	// Dashbooard
	s.Dashboard.Defaults = &s.Defaults.Dashboard
	s.Dashboard.HardDefaults = &s.HardDefaults.Dashboard
	// Options
	s.Options.Defaults = &s.Defaults.Options
	s.Options.HardDefaults = &s.HardDefaults.Options

	// Notify
	// use defaults?
	if s.Notify == nil && len(defaults.Notify) != 0 {
		s.Notify = make(shoutrrr.Slice, len(defaults.Notify))
		for key := range defaults.Notify {
			s.Notify[key] = &shoutrrr.Shoutrrr{}
		}
		s.notifyFromDefaults = true
	}
	s.Notify.Init(
		&s.Status,
		rootNotifyConfig, notifyDefaults, notifyHardDefaults)

	// Command
	// use defaults?
	if len(s.Command) == 0 && len(defaults.Command) != 0 {
		s.Command = make(command.Slice, len(defaults.Command))
		copy(s.Command, defaults.Command)
		s.commandFromDefaults = true
	}
	//nolint:typecheck
	if s.Command != nil && len(s.Command) != 0 {
		s.CommandController = &command.Controller{}
		s.CommandController.Init(
			&s.Status,
			&s.Command,
			&s.Notify,
			s.Options.GetIntervalPointer())
	}

	// WebHook
	// use defaults?
	if s.WebHook == nil && len(defaults.WebHook) != 0 {
		s.WebHook = make(webhook.Slice, len(defaults.WebHook))
		for key := range defaults.WebHook {
			s.WebHook[key] = &webhook.WebHook{}
		}
		s.webhookFromDefaults = true
	}
	s.WebHook.Init(
		&s.Status,
		rootWebHookConfig, webhookDefaults, webhookHardDefaults,
		&s.Notify,
		s.Options.GetIntervalPointer())

	// LatestVersion
	s.LatestVersion.Init(
		&s.Defaults.LatestVersion, &s.HardDefaults.LatestVersion,
		&s.Status,
		&s.Options)

	// DeployedVersionLookup
	s.DeployedVersionLookup.Init(
		&s.Defaults.DeployedVersionLookup, &s.HardDefaults.DeployedVersionLookup,
		&s.Status,
		&s.Options)

}

// ServiceInfo returns info about the service.
func (s *Service) ServiceInfo() *util.ServiceInfo {
	return &util.ServiceInfo{
		ID:            s.ID,
		URL:           s.LatestVersion.ServiceURL(true),
		WebURL:        s.Status.GetWebURL(),
		LatestVersion: s.Status.LatestVersion(),
	}
}

// IconURL returns the URL Icon for the Service.
func (s *Service) IconURL() (icon string) {
	// Service.Icon
	if strings.HasPrefix(s.Dashboard.Icon, "http") {
		icon = s.Dashboard.Icon
		return
	}

	//nolint:typecheck
	if s.Notify != nil {
		for key := range s.Notify {
			// `Params.Icon`
			icon = s.Notify[key].GetParam("icon")
			if icon != "" && strings.HasPrefix(icon, "http") {
				return
			}
		}
	}

	return
}

// InitMetrics of the Service.
func (s *Service) InitMetrics() {
	s.LatestVersion.InitMetrics()
	s.DeployedVersionLookup.InitMetrics()
	s.Notify.InitMetrics()
	s.CommandController.InitMetrics()
	s.WebHook.InitMetrics()
	s.Status.InitMetrics()
}

// DeleteMetrics of the Service.
func (s *Service) DeleteMetrics() {
	s.LatestVersion.DeleteMetrics()
	s.DeployedVersionLookup.DeleteMetrics()
	s.Notify.DeleteMetrics()
	s.CommandController.DeleteMetrics()
	s.WebHook.DeleteMetrics()
	s.Status.DeleteMetrics()
}

// ResetMetrics of the Service.
func (s *Service) ResetMetrics() {
	s.DeleteMetrics()
	s.InitMetrics()
}
