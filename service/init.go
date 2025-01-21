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
	"strings"

	"github.com/release-argus/Argus/command"
	"github.com/release-argus/Argus/notify/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	latestver "github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
	"github.com/release-argus/Argus/webhook"
)

// LogInit for this package.
func LogInit(log *util.JLog) {
	jLog = log
	deployedver.LogInit(log)
	latestver.LogInit(log)
}

// ServiceInfo returns info about the service.
func (s *Service) ServiceInfo() util.ServiceInfo {
	return util.ServiceInfo{
		ID:            s.ID,
		Name:          s.Name,
		URL:           s.LatestVersion.ServiceURL(true),
		WebURL:        s.Status.GetWebURL(),
		LatestVersion: s.Status.LatestVersion(),
	}
}

// IconURL returns the URL Icon for the Service.
func (s *Service) IconURL() string {
	// Service.Icon
	if strings.HasPrefix(s.Dashboard.Icon, "http") {
		return s.Dashboard.Icon
	}

	var icon string
	//nolint:typecheck
	if s.Notify != nil {
		// Search for a web icon.
		for _, notify := range s.Notify {
			// `Params.Icon`
			icon = notify.GetParam("icon")
			if icon != "" && strings.HasPrefix(icon, "http") {
				return icon
			}
		}
	}

	return icon
}

// Init will initialise the Service metric.
func (s *Service) Init(
	defaults, hardDefaults *Defaults,

	rootNotifyConfig *shoutrrr.SliceDefaults,
	notifyDefaults, notifyHardDefaults *shoutrrr.SliceDefaults,

	rootWebHookConfig *webhook.SliceDefaults,
	webhookDefaults, webhookHardDefaults *webhook.Defaults,
) {
	// Status.
	s.Status.Init(
		len(s.Notify), len(s.Command), len(s.WebHook),
		&s.ID, &s.Name,
		&s.Dashboard.WebURL)

	// Service.
	s.Defaults = defaults
	s.HardDefaults = hardDefaults
	// Dashboard.
	s.Dashboard.Defaults = &s.Defaults.Dashboard
	s.Dashboard.HardDefaults = &s.HardDefaults.Dashboard
	// Options.
	s.Options.Defaults = &s.Defaults.Options
	s.Options.HardDefaults = &s.HardDefaults.Options

	// Notify/
	// use defaults?
	if len(s.Notify) == 0 && len(defaults.Notify) != 0 {
		s.Notify = make(shoutrrr.Slice, len(defaults.Notify))
		for key := range defaults.Notify {
			s.Notify[key] = &shoutrrr.Shoutrrr{}
		}
		s.notifyFromDefaults = true
	}
	s.Notify.Init(
		&s.Status,
		rootNotifyConfig, notifyDefaults, notifyHardDefaults)

	// Command.
	// use defaults?
	if len(s.Command) == 0 && len(defaults.Command) != 0 {
		s.Command = make(command.Slice, len(defaults.Command))
		copy(s.Command, defaults.Command)
		s.commandFromDefaults = true
	}
	if len(s.Command) != 0 {
		s.CommandController = &command.Controller{}
		s.CommandController.Init(
			&s.Status,
			&s.Command,
			&s.Notify,
			s.Options.GetIntervalPointer())
	}

	// WebHook.
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

	// LatestVersion.
	if s.LatestVersion != nil {
		s.LatestVersion.Init(
			&s.Options,
			&s.Status,
			&s.Defaults.LatestVersion, &s.HardDefaults.LatestVersion)
	}

	// DeployedVersionLookup.
	s.DeployedVersionLookup.Init(
		&s.Options,
		&s.Status,
		&s.Defaults.DeployedVersionLookup, &s.HardDefaults.DeployedVersionLookup)

}

// initMetrics will initialise the Prometheus metrics for the Service.
func (s *Service) initMetrics() {
	if s.LatestVersion != nil {
		s.LatestVersion.InitMetrics(s.LatestVersion)
	}
	s.DeployedVersionLookup.InitMetrics()
	s.Notify.InitMetrics()
	s.CommandController.InitMetrics()
	s.WebHook.InitMetrics()
	s.Status.InitMetrics()
	metric.ServiceCountCurrent.Add(1)
}

// deleteMetrics will delete the Prometheus metrics for the Service.
func (s *Service) deleteMetrics() {
	if s.LatestVersion != nil {
		s.LatestVersion.DeleteMetrics(s.LatestVersion)
	}
	s.DeployedVersionLookup.DeleteMetrics()
	s.Notify.DeleteMetrics()
	s.CommandController.DeleteMetrics()
	s.WebHook.DeleteMetrics()
	s.Status.DeleteMetrics()

	metric.ServiceCountCurrent.Add(-1)
}
