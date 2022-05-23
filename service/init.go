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
	"fmt"
	"strings"
	"time"

	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
)

var (
	jLog *utils.JLog
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
		s.Status = &Status{}
	}
	if s.Status.Fails == nil {
		s.Status.Fails = &StatusFails{}
	}
	// Default LatestVersion to DeployedVersion
	if s.Status.LatestVersion == "" {
		s.Status.LatestVersion = s.Status.DeployedVersion
		s.Status.LatestVersionTimestamp = s.Status.DeployedVersionTimestamp
	}
	// Default DeployedVersion to LatestVersion
	if s.Status.DeployedVersion == "" {
		s.Status.DeployedVersion = s.Status.LatestVersion
		s.Status.DeployedVersionTimestamp = s.Status.LatestVersionTimestamp
	}

	s.Defaults = defaults
	s.HardDefaults = hardDefaults
	if s.DeployedVersionLookup != nil {
		s.DeployedVersionLookup.Defaults = defaults.DeployedVersionLookup
		s.DeployedVersionLookup.HardDefaults = hardDefaults.DeployedVersionLookup
	}

	ignoreMisses := s.GetIgnoreMisses()
	if ignoreMisses != nil {
		s.URLCommands.SetParentIgnoreMisses(ignoreMisses)
	}
}

// initMetrics will initialise the Prometheus metrics.
func (s *Service) initMetrics() {
	// ############
	// # Counters #
	// ############
	metrics.InitPrometheusCounterWithIDAndResult(metrics.QueryMetric, *(*s).ID, "SUCCESS")
	metrics.InitPrometheusCounterWithIDAndResult(metrics.QueryMetric, *(*s).ID, "FAIL")
}

// GetAccessToken will return the GitHub access token to use.
//
// `Service.AccessToken` > `Service.Defaults.Service.AccessToken`
func (s *Service) GetAccessToken() string {
	return utils.DefaultIfNil(utils.GetFirstNonNilPtr(s.AccessToken, s.Defaults.AccessToken, s.HardDefaults.AccessToken))
}

// GetAllowInvalidCerts returns whether invalid HTTPS certs are allowed.
func (s *Service) GetAllowInvalidCerts() bool {
	return *utils.GetFirstNonNilPtr(s.AllowInvalidCerts, s.Defaults.AllowInvalidCerts, s.HardDefaults.AllowInvalidCerts)
}

// GetAutoApprove returns whether new releases of this service should be auto-approved.
func (s *Service) GetAutoApprove() bool {
	return *utils.GetFirstNonNilPtr(s.AutoApprove, s.Defaults.AutoApprove, s.HardDefaults.AutoApprove)
}

// GetIgnoreMisses returns whether URL Command misses should be logged.
func (s *Service) GetIgnoreMisses() *bool {
	return utils.GetFirstNonNilPtr(s.IgnoreMisses, s.Defaults.IgnoreMisses, s.HardDefaults.IgnoreMisses)
}

// GetServiceInfo returns info about the service.
func (s *Service) GetServiceInfo() utils.ServiceInfo {
	return utils.ServiceInfo{
		ID:            *s.ID,
		URL:           s.GetServiceURL(true),
		WebURL:        s.GetWebURL(),
		LatestVersion: s.Status.LatestVersion,
	}
}

// GetServiceURL returns the service's URL (handles the github type where the URL
// may be `owner/repo`, adding the github.com prefix in that case).
func (s *Service) GetServiceURL(ignoreWebURL bool) string {
	if !ignoreWebURL && utils.GetFirstNonNilPtr(s.WebURL, s.Defaults.WebURL) != nil {
		// Don't use this template if `LatestVersion` hasn't been found and is used in `WebURL`.
		if s.Status.LatestVersion == "" {
			if !strings.Contains(*s.WebURL, "version") {
				return s.GetWebURL()
			}
		} else {
			return s.GetWebURL()
		}
	}

	serviceURL := *s.URL
	// GitHub service. Get the non-API URL.
	if *s.Type == "github" {
		// If it's "owner/repo" rather than a full path.
		if strings.Count(serviceURL, "/") == 1 {
			serviceURL = fmt.Sprintf("https://github.com/%s", serviceURL)
		}
	}
	return serviceURL
}

// GetIconURL returns the URL Icon for the Service.
func (s *Service) GetIconURL() *string {
	// Service.Icon
	if s.Icon != nil {
		return s.Icon
	}

	if s.Slack != nil {
		for key := range *s.Slack {
			// `Service.Slack.IconURL/IconEmoji`
			if (*s.Slack)[key].IconURL != nil {
				return (*s.Slack)[key].IconURL
			}
		}

		for key := range *s.Slack {
			// `Slack.IconURL/IconEmoji`
			if (*s.Slack)[key].Main.IconURL != nil {
				return (*s.Slack)[key].Main.IconURL
			}
		}

		for key := range *s.Slack {
			// `Default(*s.Slack)[key].Slack.IconURL/IconEmoji`
			if (*s.Slack)[key].Defaults.IconURL != nil {
				return (*s.Slack)[key].Defaults.IconURL
			}
		}
	}

	return nil
}

// GetInterval returns the interval between queries on this Service's version.
func (s *Service) GetInterval() string {
	return *utils.GetFirstNonNilPtr(s.Interval, s.Defaults.Interval, s.HardDefaults.Interval)
}

// GetIntervalDuration returns the interval between queries on this Service's version.
func (s *Service) GetIntervalDuration() time.Duration {
	d, _ := time.ParseDuration(s.GetInterval())
	return d
}

// GetUsePreRelease returns whether to use GitHub PreReleases.
func (s *Service) GetUsePreRelease() bool {
	return *utils.GetFirstNonNilPtr(s.UsePreRelease, s.Defaults.UsePreRelease, s.HardDefaults.UsePreRelease)
}

// GetSemanticVersioning returns whether semantic versioning is enabled for this Service.
func (s *Service) GetSemanticVersioning() bool {
	return *utils.GetFirstNonNilPtr(s.SemanticVersioning, s.Defaults.SemanticVersioning, s.HardDefaults.SemanticVersioning)
}

// GetRegexContent returns the wanted query body content regex.
func (s *Service) GetRegexContent() *string {
	return utils.GetFirstNonNilPtr(s.RegexContent, s.Defaults.RegexContent, s.HardDefaults.RegexContent)
}

// GetRegexVersion returns the wanted query version regex.
func (s *Service) GetRegexVersion() *string {
	return utils.GetFirstNonNilPtr(s.RegexVersion, s.Defaults.RegexVersion, s.HardDefaults.RegexVersion)
}

// GetWebURL returns the Web URL.
func (s *Service) GetWebURL() string {
	template := utils.GetFirstNonNilPtr(s.WebURL, s.Defaults.WebURL)
	if template == nil {
		return ""
	}

	return utils.TemplateString(*template, utils.ServiceInfo{LatestVersion: s.Status.LatestVersion})
}

// GetURL will ensure `url` is a valid GitHub API URL if `urlType` is 'github'
func GetURL(url string, urlType string) string {
	if urlType == "github" {
		// Convert "owner/repo" to the API path.
		if strings.Count(url, "/") == 1 {
			url = fmt.Sprintf("https://api.github.com/repos/%s/releases", url)
		}
	}
	return url
}
