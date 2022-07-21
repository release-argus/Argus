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

	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

// Init will initialise the Service metrics.
func (s *Service) Init(
	log *utils.JLog,
	defaults *Service,
	hardDefaults *Service,
) {
	jLog = log

	s.Defaults = defaults
	s.Options.Defaults = &s.Defaults.Options
	s.HardDefaults = hardDefaults
	s.Options.HardDefaults = &s.HardDefaults.Options

	if s.Status == nil {
		s.Status = &service_status.Status{}
	}
	s.Status.ServiceID = &s.ID
	s.Status.WebURL = &s.Dashboard.WebURL

	s.DeployedVersionLookup.Init(jLog, &s.Defaults.DeployedVersionLookup, &s.HardDefaults.DeployedVersionLookup, &s.Status, &s.Options)
	s.LatestVersion.Init(jLog, &s.Defaults.LatestVersion, &s.HardDefaults.LatestVersion, &s.Status, &s.Options)
	s.convert()
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
		for key := range *s.Notify {
			// `Params.Icon`
			icon := (*s.Notify)[key].GetParam("icon")
			if icon != "" && strings.HasPrefix(icon, "http") {
				return icon
			}
		}
	}

	return ""
}
