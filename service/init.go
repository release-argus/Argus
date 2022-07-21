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
	"github.com/release-argus/Argus/web/metrics"
)

// Init will initialise the Service metrics.
func (s *Service) Init(
	log *utils.JLog,
	defaults *Service,
	hardDefaults *Service,
) {
	jLog = log
	s.initMetrics()
	if s.Status == nil {
		s.Status = &service_status.Status{}
	}
	if s.Status.Fails == nil {
		s.Status.Fails = &service_status.Fails{}
	}

	s.Defaults = defaults
	s.HardDefaults = hardDefaults
	if s.DeployedVersionLookup != nil {
		s.DeployedVersionLookup.Defaults = defaults.DeployedVersionLookup
		s.DeployedVersionLookup.HardDefaults = hardDefaults.DeployedVersionLookup
	}
}

// initMetrics will initialise the Prometheus metrics.
func (s *Service) initMetrics() {
	// ############
	// # Counters #
	// ############
	metrics.InitPrometheusCounterWithIDAndResult(metrics.QueryMetric, (*s).ID, "SUCCESS")
	metrics.InitPrometheusCounterWithIDAndResult(metrics.QueryMetric, (*s).ID, "FAIL")
}

// GetServiceInfo returns info about the service.
func (s *Service) GetServiceInfo() utils.ServiceInfo {
	return utils.ServiceInfo{
		ID:            s.ID,
		URL:           s.GetServiceURL(true),
		WebURL:        s.Dashboard.GetWebURL(s.Status.LatestVersion),
		LatestVersion: s.Status.LatestVersion,
	}
}

// GetServiceURL returns the service's URL (handles the github type where the URL
// may be `owner/repo`, adding the github.com prefix in that case).
func (s *Service) GetServiceURL(ignoreWebURL bool) string {
	if !ignoreWebURL && s.Dashboard.WebURL != "" {
		// Don't use this template if `LatestVersion` hasn't been found and is used in `WebURL`.
		if s.Status.LatestVersion == "" {
			if !strings.Contains(s.Dashboard.WebURL, "version") {
				return s.Dashboard.GetWebURL(s.Status.LatestVersion)
			}
		} else {
			return s.Dashboard.GetWebURL(s.Status.LatestVersion)
		}
	}

	serviceURL := s.LatestVersion.GetFriendlyURL()
	return serviceURL
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
