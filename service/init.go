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
	defaults *ServiceDefaults,
	hardDefaults *ServiceDefaults,

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
	s.Notify.Init(
		&s.Status,
		rootNotifyConfig, notifyDefaults, notifyHardDefaults)

	// Command
	//nolint:typecheck
	if s.Command != nil {
		s.CommandController = &command.Controller{}
		s.CommandController.Init(
			&s.Status,
			&s.Command,
			&s.Notify,
			s.Options.GetIntervalPointer())
	}

	// WebHook
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

// GetServiceInfo returns info about the service.
func (s *Service) GetServiceInfo() *util.ServiceInfo {
	return &util.ServiceInfo{
		ID:            s.ID,
		URL:           s.LatestVersion.GetServiceURL(true),
		WebURL:        s.Status.GetWebURL(),
		LatestVersion: s.Status.GetLatestVersion(),
	}
}

// GetIconURL returns the URL Icon for the Service.
func (s *Service) GetIconURL() (icon string) {
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
}

// DeleteMetrics of the Service.
func (s *Service) DeleteMetrics() {
	s.LatestVersion.DeleteMetrics()
	s.DeployedVersionLookup.DeleteMetrics()
	s.Notify.DeleteMetrics()
	s.CommandController.DeleteMetrics()
	s.WebHook.DeleteMetrics()
}

// ResetMetrics of the Service.
func (s *Service) ResetMetrics() {
	s.DeleteMetrics()
	s.InitMetrics()
}
