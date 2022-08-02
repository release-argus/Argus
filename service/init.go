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
	"github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

// Init will initialise the Service metrics.
func (s *Service) Init(
	log *utils.JLog,
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

	s.Status.Init(len(s.Notify), len(s.Command), len(s.WebHook), &s.ID, &s.Dashboard.WebURL)
	s.Defaults = defaults
	s.Dashboard.Defaults = &s.Defaults.Dashboard
	s.Options.Defaults = &s.Defaults.Options
	s.HardDefaults = hardDefaults
	s.Dashboard.HardDefaults = &s.HardDefaults.Dashboard
	s.Options.HardDefaults = &s.HardDefaults.Options

	s.Notify.Init(jLog, &s.Status, rootNotifyConfig, notifyDefaults, notifyHardDefaults)

	if s.Command != nil {
		s.CommandController = &command.Controller{}
		s.CommandController.Init(jLog, &s.Status, &s.Command, &s.Notify, s.Options.GetIntervalPointer())
	}

	s.WebHook.Init(jLog, &s.Status, rootWebHookConfig, webhookDefaults, webhookHardDefaults, &s.Notify, s.Options.GetIntervalPointer())

	s.LatestVersion.Init(jLog, &s.Defaults.LatestVersion, &s.HardDefaults.LatestVersion, &s.Status, &s.Options)
	if s.Defaults.DeployedVersionLookup == nil {
		s.Defaults.DeployedVersionLookup = &deployed_version.Lookup{}
	}
	s.DeployedVersionLookup.Init(jLog, s.Defaults.DeployedVersionLookup, s.HardDefaults.DeployedVersionLookup, &s.Status, &s.Options)
	s.Convert()
}

// GetServiceInfo returns info about the service.
func (s *Service) GetServiceInfo() utils.ServiceInfo {
	return utils.ServiceInfo{
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
