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
	"strings"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	deployedver "github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/webhook"
)

// Init will initialise the Service metric.
func (s *Service) Init(
	log *util.JLog,
	defaults *Service,
	hardDefaults *Service,
	rootNotifyConfig *shoutrrr.Slice,
	notifyDefaults *shoutrrr.Slice,
	notifyHardDefaults *shoutrrr.Slice,
	rootWebHookConfig *webhook.Slice,
	webhookDefaults *webhook.WebHook,
	webhookHardDefaults *webhook.WebHook,
) {
	jLog = log

	// Status
	s.Status.Init(
		len(s.Notify),
		len(s.Command),
		len(s.WebHook),
		&s.ID,
		&s.Dashboard.WebURL)

	s.Defaults = defaults
	s.Dashboard.Defaults = &s.Defaults.Dashboard
	s.Options.Defaults = &s.Defaults.Options
	s.HardDefaults = hardDefaults
	s.Dashboard.HardDefaults = &s.HardDefaults.Dashboard
	s.Options.HardDefaults = &s.HardDefaults.Options

	// Notify
	s.Notify.Init(
		jLog,
		&s.Status,
		rootNotifyConfig,
		notifyDefaults,
		notifyHardDefaults)

	// Command
	if s.Command != nil {
		s.CommandController = &command.Controller{}
		s.CommandController.Init(
			jLog,
			&s.Status,
			&s.Command,
			&s.Notify,
			s.Options.GetIntervalPointer())
	}

	// WebHook
	s.WebHook.Init(
		jLog,
		&s.Status,
		rootWebHookConfig,
		webhookDefaults,
		webhookHardDefaults,
		&s.Notify,
		s.Options.GetIntervalPointer())

	// LatestVersion
	s.LatestVersion.Init(
		jLog,
		&s.Defaults.LatestVersion,
		&s.HardDefaults.LatestVersion,
		&s.Status,
		&s.Options)

	// DeployedVersionLookup
	if s.Defaults.DeployedVersionLookup == nil {
		s.Defaults.DeployedVersionLookup = &deployedver.Lookup{}
	}
	s.DeployedVersionLookup.Init(
		jLog,
		s.Defaults.DeployedVersionLookup,
		s.HardDefaults.DeployedVersionLookup,
		&s.Status,
		&s.Options)

	// Convert from old format
	s.Convert()
}

// GetServiceInfo returns info about the service.
func (s *Service) GetServiceInfo() *util.ServiceInfo {
	return &util.ServiceInfo{
		ID:            s.ID,
		URL:           s.LatestVersion.GetServiceURL(true),
		WebURL:        s.Status.GetWebURL(),
		LatestVersion: s.Status.LatestVersion,
	}
}

// GetIconURL returns the URL Icon for the Service.
func (s *Service) GetIconURL() string {
	// Service.Icon
	if strings.HasPrefix(s.Dashboard.Icon, "http") {
		return s.Dashboard.Icon
	}

	if s.Notify != nil {
		for key := range s.Notify {
			// `Params.Icon`
			icon := s.Notify[key].GetParam("icon")
			if icon != "" && strings.HasPrefix(icon, "http") {
				return icon
			}
		}
	}

	return ""
}

// InitMetrics  of the Service.
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
